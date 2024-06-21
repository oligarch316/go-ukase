package ukcli

import (
	"context"
	"reflect"

	"github.com/oligarch316/go-ukase/ukcore"
	"github.com/oligarch316/go-ukase/ukcore/ukinit"
)

// =============================================================================
// Directive
// =============================================================================

type Directive interface {
	UkaseRegister(State) error
}

type directive func(State) error

func (d directive) UkaseRegister(state State) error { return d(state) }

// =============================================================================
// Rule
// =============================================================================

type Rule[Params any] func(*Params)

func NewRule[Params any](rule func(*Params)) Rule[Params] {
	return Rule[Params](rule)
}

func (r Rule[Params]) UkaseRegister(state State) error {
	state.RegisterRule(ukinit.NewRule(r))
	return nil
}

// =============================================================================
// Info
// =============================================================================

type Info struct{ Value any }

func NewInfo(info any) Info {
	return Info{Value: info}
}

func (i Info) Bind(target ...string) Directive {
	dir := func(s State) error { return s.RegisterInfo(i.Value, target...) }
	return directive(dir)
}

// =============================================================================
// Exec
// =============================================================================

type Exec[Params any] func(context.Context, Input) error

func (e Exec[Params]) Bind(target ...string) Directive {
	dir := func(s State) error { return e.register(s, target) }
	return directive(dir)
}

func (e Exec[Params]) register(state State, target []string) error {
	t := reflect.TypeFor[Params]()

	spec, err := state.loadSpec(t)
	if err != nil {
		return err
	}

	exec := func(ctx context.Context, in ukcore.Input) error {
		return e(ctx, newInput(in, state))
	}

	return state.RegisterExec(exec, spec, target...)
}
