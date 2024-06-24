package ukcli

import (
	"context"
	"reflect"

	"github.com/oligarch316/go-ukase/ukcore"
	"github.com/oligarch316/go-ukase/ukcore/ukdec"
	"github.com/oligarch316/go-ukase/ukcore/ukexec"
	"github.com/oligarch316/go-ukase/ukcore/ukinit"
	"github.com/oligarch316/go-ukase/ukcore/ukspec"
)

// =============================================================================
// Runtime
// =============================================================================

type Runtime struct {
	config     Config
	directives []Directive
}

func NewRuntime(opts ...Option) *Runtime {
	config := newConfig(opts)
	return &Runtime{config: config}
}

func (r *Runtime) Add(directives ...Directive) {
	r.directives = append(r.directives, directives...)
}

func (r *Runtime) Execute(ctx context.Context, values []string) error {
	state := newState(r.config)

	if err := r.prepare(state); err != nil {
		return err
	}

	return state.execMux.Execute(ctx, values)
}

func (r *Runtime) prepare(state State) error {
	for _, middleware := range r.config.Middleware {
		state = middleware(state)
	}

	for _, dir := range r.directives {
		if err := dir.UkaseRegister(state); err != nil {
			return err
		}
	}

	return nil
}

// =============================================================================
// State
// =============================================================================

var _ State = (*state)(nil)

type State interface {
	// Execution time utilities
	loadMeta(target []string) (ukexec.Meta, error)
	loadSpec(t reflect.Type) (ukspec.Params, error)
	runDecode(ukcore.Input, any) error
	runInit(any) error

	// Registration time utilities
	RegisterExec(exec ukcore.Exec, spec ukspec.Params, target ...string) error
	RegisterInfo(info any, target ...string) error
	RegisterRule(rule ukinit.Rule)
}

type state struct {
	config  Config
	execMux *ukexec.Mux
	ruleSet *ukinit.RuleSet
}

func newState(config Config) *state {
	return &state{
		config:  config,
		execMux: ukexec.New(config.Exec...),
		ruleSet: ukinit.NewRuleSet(config.Init...),
	}
}

func (s *state) loadMeta(target []string) (ukexec.Meta, error) {
	return s.execMux.Meta(target...)
}

func (s *state) loadSpec(t reflect.Type) (ukspec.Params, error) {
	return ukspec.New(t, s.config.Spec...)
}

func (s *state) runDecode(i ukcore.Input, v any) error {
	decoder := ukdec.NewDecoder(i, s.config.Decode...)
	return decoder.Decode(v)
}

func (s *state) runInit(v any) error {
	spec, err := ukspec.Of(v, s.config.Spec...)
	if err != nil {
		return err
	}

	return s.ruleSet.Process(spec, v)
}

func (s *state) RegisterExec(exec ukcore.Exec, spec ukspec.Params, target ...string) error {
	return s.execMux.RegisterExec(exec, spec, target...)
}

func (s *state) RegisterInfo(info any, target ...string) error {
	return s.execMux.RegisterInfo(info, target...)
}

func (s *state) RegisterRule(rule ukinit.Rule) {
	rule.Register(s.ruleSet)
}

// =============================================================================
// Input
// =============================================================================

var _ Input = input{}

type Input interface {
	Core() ukcore.Input
	Decode(any) error
	Initialize(any) error
	Lookup(target ...string) (ukexec.Meta, error)
}

type input struct {
	core  ukcore.Input
	state State
}

func newInput(core ukcore.Input, state State) input {
	return input{core: core, state: state}
}

func (i input) Core() ukcore.Input                      { return i.core }
func (i input) Decode(v any) error                      { return i.state.runDecode(i.core, v) }
func (i input) Initialize(v any) error                  { return i.state.runInit(v) }
func (i input) Lookup(t ...string) (ukexec.Meta, error) { return i.state.loadMeta(t) }
