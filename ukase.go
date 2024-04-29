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
	Core []ukcore.Option
	Enc  []ukenc.Option
	Init []ukinit.Option
	Spec []ukspec.Option
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
	state := state{
		config:  r.config,
		mux:     ukcore.New(r.config.Core...),
		ruleSet: ukinit.NewRuleSet(r.config.Init...),
	}

	for _, directive := range r.directives {
		if err := directive.UkaseRegister(&state); err != nil {
			return err
		}
	}

	return state.mux.Execute(ctx, values)
}

// =============================================================================
// Scope
// =============================================================================

type scopedState struct {
	State
	target []string
}

func (ss scopedState) registerRule(rule ukinit.Rule) {
	// TODO: Sequester rules added to a scope to that scope and it's children?

	ss.State.registerRule(rule)
}

func (ss scopedState) registerExec(exec ukcore.Exec, spec ukspec.Params, target []string) error {
	target = append(ss.target, target...)
	return ss.State.registerExec(exec, spec, target)
}

type Scope struct {
	target     []string
	directives []Directive
}

func NewScope(target ...string) *Scope {
	return &Scope{target: target}
}

func (s *Scope) Add(directives ...Directive) *Scope {
	s.directives = append(s.directives, directives...)
	return s
}

func (s *Scope) UkaseRegister(state State) error {
	childState := scopedState{State: state, target: s.target}

	for _, directive := range s.directives {
		if err := directive.UkaseRegister(childState); err != nil {
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
	state.registerRule(ukinit.NewRule(r))
	return nil
}

// =============================================================================
// Exec
// =============================================================================

type Input struct {
	ukcore.Input
	state State
}

func (i Input) Initialize(v any) error { return i.state.execInit(v) }
func (i Input) Decode(v any) error     { return i.state.execDecode(i.Input, v) }

type Exec[Params any] func(context.Context, Input) error

func NewExec[Params any](exec func(context.Context, Input) error) Exec[Params] {
	return Exec[Params](exec)
}

func (e Exec[Params]) Command(target ...string) Directive {
	d := func(state State) error { return e.register(state, target) }
	return directive(d)
}

func (e Exec[Params]) register(state State, target []string) error {
	t := reflect.TypeFor[Params]()

	spec, err := state.execSpec(t)
	if err != nil {
		return err
	}

	exec := func(ctx context.Context, coreInput ukcore.Input) error {
		input := Input{Input: coreInput, state: state}
		return e(ctx, input)
	}

	return state.registerExec(exec, spec, target)
}

// =============================================================================
// Handler
// =============================================================================

type Handler[Params any] func(context.Context, Params) error

func NewHandler[Params any](handler func(context.Context, Params) error) Handler[Params] {
	return Handler[Params](handler)
}

func (h Handler[Params]) Command(target ...string) Directive {
	var exec Exec[Params] = h.exec
	return exec.Command(target...)
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
