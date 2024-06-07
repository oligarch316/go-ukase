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

type Directive interface {
	UkaseRegister(State) error
}

type DirectiveFunc func(State) error

func (df DirectiveFunc) UkaseRegister(state State) error { return df(state) }

// =============================================================================
// State
// =============================================================================

type State interface {
	decode(ukcore.Input, any) error
	initialize(any) error

	loadMeta(target []string) (ukcore.Meta, error)
	loadSpec(t reflect.Type) (ukspec.Params, error)

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

func (s *state) loadMeta(target []string) (ukcore.Meta, error) {
	return s.mux.Meta(target...)
}

func (s *state) loadSpec(t reflect.Type) (ukspec.Params, error) {
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

type Input struct {
	ukcore.Input
	state State
}

func (i Input) Decode(v any) error                        { return i.state.decode(i.Input, v) }
func (i Input) Initialize(v any) error                    { return i.state.initialize(v) }
func (i Input) Meta(target []string) (ukcore.Meta, error) { return i.state.loadMeta(target) }
