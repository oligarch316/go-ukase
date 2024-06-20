package ukase

import (
	"context"
	"reflect"

	"github.com/oligarch316/go-ukase/ukcore"
	"github.com/oligarch316/go-ukase/ukreflect/ukenc"
	"github.com/oligarch316/go-ukase/ukreflect/ukinit"
	"github.com/oligarch316/go-ukase/ukspec"
)

// =============================================================================
// Config
// =============================================================================

var defaultConfig = Config{}

type Option interface{ UkaseApply(*Config) }

type Config struct {
	// TODO: Document
	Core []ukcore.Option

	// TODO: Document
	Enc []ukenc.Option

	// TODO: Document
	Init []ukinit.Option

	// TODO: Document
	Spec []ukspec.Option

	// TODO: Document
	Middleware []func(State) State
}

func newConfig(opts []Option) Config {
	config := defaultConfig
	for _, opt := range opts {
		opt.UkaseApply(&config)
	}
	return config
}

// =============================================================================
// Runtime
// =============================================================================

type Runtime struct {
	config     Config
	directives []Directive
}

func New(opts ...Option) *Runtime {
	return &Runtime{config: newConfig(opts)}
}

func (r *Runtime) Add(directives ...Directive) {
	r.directives = append(r.directives, directives...)
}

func (r *Runtime) Execute(ctx context.Context, values []string) error {
	state := newState(r.config)

	if err := r.prepare(state); err != nil {
		return err
	}

	return state.mux.Execute(ctx, values)
}

func (r *Runtime) prepare(state State) error {
	for _, middleware := range r.config.Middleware {
		state = middleware(state)
	}

	for _, d := range r.directives {
		if err := d.UkaseRegister(state); err != nil {
			return err
		}
	}

	return nil
}

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
	df := func(state State) error { return state.RegisterInfo(i.Value, target) }
	return directive(df)
}

// =============================================================================
// Exec
// =============================================================================

type Exec[Params any] func(context.Context, Input) error

func (e Exec[Params]) Bind(target ...string) Directive {
	df := func(state State) error { return e.register(state, target) }
	return directive(df)
}

func (e Exec[Params]) register(state State, target []string) error {
	t := reflect.TypeFor[Params]()

	spec, err := state.spec(t)
	if err != nil {
		return err
	}

	exec := func(ctx context.Context, in ukcore.Input) error {
		return e(ctx, newInput(in, state))
	}

	return state.RegisterExec(exec, spec, target)
}

// =============================================================================
// Handler
// =============================================================================

type Handler[Params any] func(context.Context, Params) error

func NewHandler[Params any](handler func(context.Context, Params) error) Handler[Params] {
	return Handler[Params](handler)
}

func (h Handler[Params]) Bind(target ...string) Directive {
	exec := Exec[Params](h.exec)
	return exec.Bind(target...)
}

func (h Handler[Params]) exec(ctx context.Context, input Input) error {
	var params Params

	if err := input.Initialize(&params); err != nil {
		return err
	}

	if err := input.Decode(&params); err != nil {
		return err
	}

	return h(ctx, params)
}
