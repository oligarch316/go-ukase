package ukcli

import (
	"context"
	"reflect"

	"github.com/oligarch316/ukase/ukcore"
	"github.com/oligarch316/ukase/ukcore/ukinit"
)

var directiveNoop directiveFunc = func(State) error { return nil }

// =============================================================================
// Directive
// =============================================================================

type Directive interface{ UkaseRegister(State) error }

type directiveFunc func(State) error

func NewDirective(directive func(State) error) Directive { return directiveFunc(directive) }

func (df directiveFunc) UkaseRegister(state State) error { return df(state) }

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
	return directiveFunc(dir)
}

// =============================================================================
// Exec
// =============================================================================

type Exec[Params any] func(context.Context, Input) error

func (e Exec[Params]) Bind(target ...string) Directive {
	if e == nil {
		return directiveNoop
	}

	dir := func(s State) error { return e.register(s, target) }
	return directiveFunc(dir)
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

// =============================================================================
// Handler
// =============================================================================

type Handler[Params any] func(context.Context, Params) error

func NewHandler[Params any](handler func(context.Context, Params) error) Handler[Params] {
	return Handler[Params](handler)
}

func (h Handler[Params]) Bind(target ...string) Directive {
	if h == nil {
		return directiveNoop
	}

	exec := Exec[Params](h.exec)
	return exec.Bind(target...)
}

func (h Handler[Params]) exec(ctx context.Context, in Input) error {
	var params Params

	if err := in.Initialize(&params); err != nil {
		return err
	}

	if err := in.Decode(&params); err != nil {
		return err
	}

	return h(ctx, params)
}
