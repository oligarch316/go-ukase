package ukcore

import (
	"context"
	"errors"
	"fmt"

	"github.com/oligarch316/go-ukase/ukspec"
)

// =============================================================================
// Input
// =============================================================================

type Flag struct{ Name, Value string }

type Input struct {
	Program string
	Target  []string
	Args    []string
	Flags   []Flag
}

// =============================================================================
// Data
// =============================================================================

type Exec func(context.Context, Input) error

type Meta interface {
	Info() (any, bool)
	Spec() (ukspec.Params, bool)
	Children() map[string]Meta
}

// =============================================================================
// Mux
// =============================================================================

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

func (m *Mux) RegisterExec(exec Exec, spec ukspec.Params, target ...string) error {
	if err := m.copyFlags(spec.Flags); err != nil {
		return err
	}

	node := m.root

	for _, childName := range target {
		child, ok := node.children[childName]
		if !ok {
			child = newMuxNode()
			node.children[childName] = child
		}

		child.copyFlags(spec.Flags)
		node = child
	}

	if node.exec != nil && !m.config.MuxOverwrite {
		return fmt.Errorf("[TODO RegisterExec] overwrite not allowed on target %v", target)
	}

	node.exec, node.spec = exec, spec
	return nil
}

func (m *Mux) RegisterInfo(info any, target ...string) error {
	node := m.root

	for _, childName := range target {
		child, ok := node.children[childName]
		if !ok {
			child = newMuxNode()
			node.children[childName] = child
		}

		node = child
	}

	// TODO: Separate config setting for this?
	if node.info != nil && !m.config.MuxOverwrite {
		return fmt.Errorf("[TODO RegisterInfo] overwrite not allowed on target %v", target)
	}

	node.info = info
	return nil
}

func (m *Mux) Lookup(target ...string) (Meta, error) {
	node := m.root

	for _, childName := range target {
		child, ok := node.children[childName]
		if !ok {
			return nil, fmt.Errorf("[TODO Lookup] go an unknown name: %s", childName)
		}

		node = child
	}

	return node, nil
}

func (m *Mux) Execute(ctx context.Context, values []string) error {
	if len(values) == 0 {
		return errors.New("[TODO Execute] empty values")
	}

	// Set up
	programName, parser := values[0], Parser(values[1:])
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

		if token.Kind == KindDelim || token.Kind == KindEOF {
			break
		}

		if token.Kind != KindString {
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
	input.Args = append(input.Args, parser.Flush()...)

	// Exec
	if node.exec != nil {
		return node.exec(ctx, input)
	}

	return m.config.ExecDefault(ctx, input)
}

func (m *Mux) copyFlags(flags map[string]ukspec.Flag) error {
	checkLevel := m.config.FlagCheck

	for name, spec := range flags {
		extant, exists := m.root.flags[name]
		if !exists {
			m.root.flags[name] = spec
			continue
		}

		elideMatch := extant.Elide.Allow == spec.Elide.Allow
		typeMatch := extant.Type == spec.Type

		switch {
		case checkLevel == FlagCheckElide && !elideMatch:
			return errors.New("[TODO copyFlags] flag elide conflict")
		case checkLevel == FlagCheckType && !typeMatch:
			return errors.New("[TODO copyFlags] flag type conflict")
		}

		m.root.flags[name] = spec
	}

	return nil
}

// =============================================================================
// Mux Node
// =============================================================================

type muxNode struct {
	exec Exec
	spec ukspec.Params
	info any

	flags    map[string]ukspec.Flag
	children map[string]*muxNode
}

func newMuxNode() *muxNode {
	return &muxNode{
		flags:    make(map[string]ukspec.Flag),
		children: make(map[string]*muxNode),
	}
}

func (mn *muxNode) Info() (any, bool)           { return mn.info, mn.info != nil }
func (mn *muxNode) Spec() (ukspec.Params, bool) { return mn.spec, mn.spec.Type != nil }

func (mn *muxNode) Children() map[string]Meta {
	res := make(map[string]Meta)

	for childName, child := range mn.children {
		res[childName] = child
	}

	return res
}

func (mn *muxNode) copyFlags(flags map[string]ukspec.Flag) {
	for name, spec := range flags {
		mn.flags[name] = spec
	}
}
