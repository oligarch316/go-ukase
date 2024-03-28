package ukcore

import (
	"context"
	"errors"
	"fmt"
)

type muxNode struct {
	handler Handler
	info    ParamsInfo
}

type Mux struct {
	children map[string]*Mux
	node     *muxNode
}

func NewMux() *Mux { return &Mux{children: make(map[string]*Mux)} }

func (m *Mux) MustHandle(handler Handler, defaultParams any, target ...string) {
	if err := m.Handle(handler, defaultParams, target...); err != nil {
		panic(err)
	}
}

func (m *Mux) Handle(handler Handler, defaultParams any, target ...string) error {
	info, err := ParamsInfoOf(defaultParams)
	if err != nil {
		return err
	}

	current := m

	for _, childName := range target {
		child, ok := current.children[childName]
		if !ok {
			child = NewMux()
			current.children[childName] = child
		}

		current = child
	}

	if current.node != nil {
		// TODO: if !opts.AllowOverwrite -> return error
	}

	current.node = &muxNode{handler: handler, info: info}
	return nil
}

func (m *Mux) Run(ctx context.Context, values []string) error {
	if len(values) == 0 {
		return errors.New("[TODO Run] empty vales")
	}

	// Consume the 1st value (program name) as the root command
	target := Target{values[0]}
	parser := NewParser(values[1:])

	// Set up
	input := Input{Target: target}
	current := m

	for {
		// ----- Flags
		flags, err := parser.ParseFlags(current.info())
		if err != nil {
			return err
		}

		input.Flags = append(input.Flags, flags...)

		// ----- Subcommands
		token := parser.ParseToken()

		if token.Kind == KindDelim || token.Kind == KindEOF {
			break
		}

		if token.Kind != KindString {
			return fmt.Errorf("[TODO Run] <INTERNAL> got an unexpected token kind (%s)", token.Kind)
		}

		child, ok := current.children[token.Value]
		if !ok {
			input.Args = append(input.Args, token.Value)
			break
		}

		input.Target = append(input.Target, token.Value)
		current = child
	}

	// Consume any remaining values as arguments
	input.Args = append(input.Args, parser.Flush()...)

	// Delegate to handler
	return current.handler()(ctx, input)
}

func (m Mux) info() ParamsInfo {
	if m.node != nil {
		return m.node.info
	}

	return EmptyParamsInfo
}

func (m Mux) handler() Handler {
	if m.node != nil {
		return m.node.handler
	}

	// TODO: Parameterize
	return func(context.Context, Input) error {
		return errors.New("TODO: default handler stuffz")
	}
}
