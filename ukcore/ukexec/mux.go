package ukexec

import (
	"context"
	"errors"
	"fmt"

	"github.com/oligarch316/go-ukase/ukcore"
	"github.com/oligarch316/go-ukase/ukcore/ukspec"
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
	m.config.Log.Debug(
		"registering exec",
		"target", target,
		"paramsType", spec.Type,
	)

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
	m.config.Log.Debug(
		"registering info",
		"target", target,
		"infoType", fmt.Sprintf("%T", info),
	)

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
			return Meta{}, fmt.Errorf("invalid target '%s': %w", target, ErrTargetNotExist)
		}

		node = child
	}

	return newMeta(node), nil
}

func (m *Mux) Execute(ctx context.Context, values []string) error {
	parser := newParser(values)

	program, ok := parser.ConsumeValue()
	if !ok {
		return ErrorParse{err: ErrMissingProgram}
	}

	input := ukcore.Input{Program: program}
	node := m.root

	for {
		// Consume all flags for the current node
		flags, err := parser.ConsumeFlags(node.flags)
		if err != nil {
			return ErrorParse{Target: input.Target, Position: parser.Position, err: err}
		}

		input.Flags = append(input.Flags, flags...)

		// Consume the next token of kind ...
		token := parser.ConsumeToken()

		// ... ❬Delim❭ or ❬EOF❭ ⇒ break out to argument parsing
		if token.Kind == kindDelim || token.Kind == kindEOF {
			break
		}

		// ... non-subcommand ⇒ set as 1st argument and break out to argument parsing
		child, ok := node.children[token.Value]
		if !ok {
			input.Arguments = append(input.Arguments, token.Value)
			break
		}

		// ... subcommand ⇒ append command name to target and continue
		input.Target = append(input.Target, token.Value)
		node = child
	}

	// All remaining unconsumed values are treated as arguments
	input.Arguments = append(input.Arguments, parser.Values...)

	m.config.Log.Info("executing", "target", input.Target)

	if node.exec == nil {
		return m.config.ExecUnspecified(ctx, input)
	}

	return node.exec(ctx, input)
}
