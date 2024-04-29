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

type directive func(State) error

func (d directive) UkaseRegister(state State) error { return d(state) }

// =============================================================================
// State
// =============================================================================

type State interface {
	// TODO: Probably want all of these methods exported/accessible

	execSpec(reflect.Type) (ukspec.Params, error)
	execInit(any) error
	execDecode(ukcore.Input, any) error

	registerRule(ukinit.Rule)
	registerExec(ukcore.Exec, ukspec.Params, []string) error
}

type state struct {
	config  Config
	mux     *ukcore.Mux
	ruleSet *ukinit.RuleSet
}

func (s *state) execSpec(t reflect.Type) (ukspec.Params, error) {
	return ukspec.New(t, s.config.Spec...)
}

func (s *state) execInit(v any) error {
	spec, err := ukspec.Of(v, s.config.Spec...)
	if err != nil {
		return err
	}

	return s.ruleSet.Process(spec, v)
}

func (s *state) execDecode(input ukcore.Input, v any) error {
	decoder := ukenc.NewDecoder(input, s.config.Enc...)
	return decoder.Decode(v)
}

func (s *state) registerRule(rule ukinit.Rule) { rule.Register(s.ruleSet) }

func (s *state) registerExec(exec ukcore.Exec, spec ukspec.Params, target []string) error {
	return s.mux.Register(exec, spec, target...)
}
