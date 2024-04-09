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
	Target, Args []string
	Flags        []Flag
}

type Exec interface {
	Execute(context.Context, Input) error
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

func (m *Mux) Register(exec Exec, spec ukspec.Params, target ...string) error {
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
		return errors.New("[TODO Register] overwrite not allowed")
	}

	node.exec = exec
	return nil
}

func (m *Mux) Execute(ctx context.Context, values []string) error {
	if len(values) == 0 {
		return errors.New("[TODO Execute] empty values")
	}

	// Set up
	programName, parser := values[0], Parser(values[1:])
	input := Input{Target: []string{programName}}
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
		return node.exec.Execute(ctx, input)
	}

	return m.config.ExecDefault.Execute(ctx, input)
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
	exec     Exec
	flags    map[string]ukspec.Flag
	children map[string]*muxNode
}

func newMuxNode() *muxNode {
	return &muxNode{
		flags:    make(map[string]ukspec.Flag),
		children: make(map[string]*muxNode),
	}
}

func (mn *muxNode) copyFlags(flags map[string]ukspec.Flag) {
	for name, spec := range flags {
		mn.flags[name] = spec
	}
}
