package ukase

import (
	"reflect"

	"github.com/oligarch316/go-ukase/ukcore"
	"github.com/oligarch316/go-ukase/ukreflect/ukenc"
	"github.com/oligarch316/go-ukase/ukreflect/ukinit"
	"github.com/oligarch316/go-ukase/ukspec"
)

// =============================================================================
// Directive
// =============================================================================

var _ Directive = directive(nil)

type Directive interface {
	UkaseRegister(State) error
}

type directive func(State) error

func (d directive) UkaseRegister(state State) error { return d(state) }

// =============================================================================
// State
// =============================================================================

var _ State = (*state)(nil)

type State interface {
	decode(ukcore.Input, any) error
	initialize(any) error
	meta(target []string) (ukcore.Meta, error)
	spec(t reflect.Type) (ukspec.Params, error)

	RegisterExec(exec ukcore.Exec, spec ukspec.Params, target []string) error
	RegisterInfo(info any, target []string) error
	RegisterRule(rule ukinit.Rule)
}

type state struct {
	config  Config
	mux     *ukcore.Mux
	ruleSet *ukinit.RuleSet
}

func newState(config Config) *state {
	return &state{
		config:  config,
		mux:     ukcore.New(config.Core...),
		ruleSet: ukinit.NewRuleSet(config.Init...),
	}
}

func (s *state) decode(input ukcore.Input, v any) error {
	decoder := ukenc.NewDecoder(input, s.config.Enc...)
	return decoder.Decode(v)
}

func (s *state) initialize(v any) error {
	spec, err := ukspec.Of(v, s.config.Spec...)
	if err != nil {
		return err
	}

	return s.ruleSet.Process(spec, v)
}

func (s *state) meta(target []string) (ukcore.Meta, error) {
	return s.mux.Meta(target...)
}

func (s *state) spec(t reflect.Type) (ukspec.Params, error) {
	return ukspec.New(t, s.config.Spec...)
}

func (s *state) RegisterExec(exec ukcore.Exec, spec ukspec.Params, target []string) error {
	return s.mux.RegisterExec(exec, spec, target...)
}

func (s *state) RegisterInfo(info any, target []string) error {
	return s.mux.RegisterInfo(info, target...)
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
	Meta(target []string) (ukcore.Meta, error)
}

type input struct {
	core  ukcore.Input
	state State
}

func newInput(core ukcore.Input, state State) input {
	return input{core: core, state: state}
}

func (i input) Core() ukcore.Input                        { return i.core }
func (i input) Decode(v any) error                        { return i.state.decode(i.core, v) }
func (i input) Initialize(v any) error                    { return i.state.initialize(v) }
func (i input) Meta(target []string) (ukcore.Meta, error) { return i.state.meta(target) }
