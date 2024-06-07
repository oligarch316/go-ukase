package ukcore

import (
	"context"
	"errors"
	"fmt"

	"github.com/oligarch316/go-ukase/ukspec"
)

type Mux struct {
	config Config
	root   *muxNode
}

func New(opts ...Option) *Mux {
	return &Mux{
		config: newConfig(opts),
		root:   newMuxNode(),
	}
}

type muxNode struct {
	exec Exec
	info any
	spec *ukspec.Params

	children map[string]*muxNode
	flags    map[string]ukspec.Flag
}

func newMuxNode() *muxNode {
	return &muxNode{
		children: make(map[string]*muxNode),
		flags:    make(map[string]ukspec.Flag),
	}
}

// =============================================================================
// Write
// =============================================================================

func (m *Mux) RegisterExec(exec Exec, spec ukspec.Params, target ...string) error {
	if err := m.validateFlags(m.root, target, spec.FlagIndex); err != nil {
		return err
	}

	node := m.root
	m.updateFlags(node, spec.FlagIndex)

	for _, name := range target {
		child, ok := node.children[name]
		if !ok {
			child = newMuxNode()
			node.children[name] = child
		}

		node = child
		m.updateFlags(node, spec.FlagIndex)
	}

	return m.updateExec(node, target, exec, spec)
}

func (m *Mux) RegisterInfo(info any, target ...string) error {
	node := m.root

	for _, name := range target {
		child, ok := node.children[name]
		if !ok {
			child = newMuxNode()
			node.children[name] = child
		}

		node = child
	}

	return m.updateInfo(node, target, info)
}

func (m *Mux) updateExec(node *muxNode, target []string, exec Exec, spec ukspec.Params) error {
	if node.spec == nil {
		node.exec, node.spec = exec, &spec
		return nil
	}

	overwrite, err := m.config.ExecConflict(*node.spec, spec)
	if err != nil {
		return ErrorExecConflict{Target: target, Original: *node.spec, Update: spec, err: err}
	}

	if overwrite {
		node.exec, node.spec = exec, &spec
	}

	return nil
}

func (m *Mux) updateInfo(node *muxNode, target []string, info any) error {
	if node.info == nil {
		node.info = info
		return nil
	}

	overwrite, err := m.config.InfoConflict(node.info, info)
	if err != nil {
		return ErrorInfoConflict{Target: target, Original: node.info, Update: info, err: err}
	}

	if overwrite {
		node.info = info
	}

	return nil
}

func (Mux) updateFlags(node *muxNode, updates map[string]ukspec.Flag) {
	for name, update := range updates {
		node.flags[name] = update
	}
}

func (m *Mux) validateFlags(node *muxNode, target []string, flags map[string]ukspec.Flag) error {
	var errs []error

	for name, update := range flags {
		original, conflict := node.flags[name]
		if !conflict {
			continue
		}

		if err := m.config.FlagConflict(original, update); err != nil {
			err = ErrorFlagConflict{Target: target, Name: name, Original: original, Update: update, err: err}
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}

// =============================================================================
// Read
// =============================================================================

func (m *Mux) Meta(target ...string) (Meta, error) {
	node := m.root

	for _, name := range target {
		child, ok := node.children[name]
		if !ok {
			return Meta{}, fmt.Errorf("invalid target '%s': %w", InputTarget(target), ErrTargetNotExist)
		}

		node = child
	}

	return newMeta(node), nil
}

func (m *Mux) Execute(ctx context.Context, values []string) error {
	if len(values) == 0 {
		return ErrEmptyValues
	}

	// TODO: Track the current value index for error purposes

	// Set up
	programName, parser := values[0], parser(values[1:])
	input := Input{Program: programName}
	node := m.root

	for {
		// Flags
		flags, err := parser.ConsumeFlags(node.flags)
		if err != nil {
			return err
		}

		input.Flags = append(input.Flags, flags...)

		// Subcommands
		token := parser.ConsumeToken()

		if token.Kind == kindDelim || token.Kind == kindEOF {
			break
		}

		if token.Kind != kindString {
			return fmt.Errorf("[TODO Execute] <INTERNAL> got an unexpected token kind (%s)", token.Kind)
		}

		child, ok := node.children[token.Value]
		if !ok {
			input.Args = append(input.Args, token.Value)
			break
		}

		input.Target = append(input.Target, token.Value)
		node = child
	}

	// Args
	input.Args = append(input.Args, parser...)

	// Exec
	if node.exec != nil {
		return node.exec(ctx, input)
	}

	return m.config.ExecUnspecified(ctx, input)
}
