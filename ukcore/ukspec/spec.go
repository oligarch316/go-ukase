package ukspec

import (
	"reflect"
	"slices"

	"github.com/oligarch316/go-ukase/internal/ierror"
	"github.com/oligarch316/go-ukase/internal/ispec"
)

// =============================================================================
// Parameters
// =============================================================================

type Parameters struct {
	Type reflect.Type

	Arguments []Argument
	Flags     []Flag
	Inlines   []Inline

	flagNames map[string]Flag
}

func ParametersFor[T any](opts ...Option) (Parameters, error) {
	t := reflect.TypeFor[T]()
	return NewParameters(t, opts...)
}

func ParametersOf(v any, opts ...Option) (Parameters, error) {
	t := reflect.TypeOf(v)
	return NewParameters(t, opts...)
}

func NewParameters(t reflect.Type, opts ...Option) (Parameters, error) {
	config := newConfig(opts)
	config.Log.Info("loading parameters", "type", t)

	paramsType := t
	for paramsType.Kind() == reflect.Pointer {
		paramsType = paramsType.Elem()
	}

	if paramsType.Kind() != reflect.Struct {
		err := ierror.NewD("not a struct or struct pointer")
		return Parameters{}, InvalidParametersError{Type: t, err: err}
	}

	s := newState(config, paramsType)
	for s.Shift() {
		if err := loadStruct(s); err != nil {
			return Parameters{}, err
		}
	}

	params := Parameters{
		Type:      paramsType,
		Arguments: s.argumentList,
		Flags:     s.flagList,
		Inlines:   s.inlineList,
		flagNames: s.flagMap,
	}

	return params, nil
}

// =============================================================================
// Parameters› Lookup
// =============================================================================

func (p Parameters) LookupArgument(position int) (Argument, bool) {
	match := func(a Argument) bool {
		lOk := a.Position.Low == nil || position >= int(*a.Position.Low)
		hOk := a.Position.High == nil || position < int(*a.Position.High)
		return lOk && hOk
	}

	if idx := slices.IndexFunc(p.Arguments, match); idx != -1 {
		return p.Arguments[idx], true
	}

	return Argument{}, false
}

func (p Parameters) LookupFlag(name string) (Flag, bool) {
	flag, ok := p.flagNames[name]
	return flag, ok
}

// =============================================================================
// Parameters› Load
// =============================================================================

func loadStruct(s *state) error {
	s.Config.Log.Debug("loading struct", "type", s.Scope.FieldType)

	for i := 0; i < s.Scope.FieldType.NumField(); i++ {
		if err := loadField(s, i); err != nil {
			return err
		}
	}

	return nil
}

func loadField(s *state, index int) error {
	sField := s.Scope.FieldType.Field(index)

	// Argument
	if tag, ok := sField.Tag.Lookup(ispec.TagKeyArguments); ok {
		return loadArgument(s, sField, []byte(tag), index)
	}

	// Flag
	if tag, ok := sField.Tag.Lookup(ispec.TagKeyFlag); ok {
		return loadFlag(s, sField, []byte(tag), index)
	}

	// Inline
	if tag, ok := sField.Tag.Lookup(ispec.TagKeyInline); ok {
		return loadInline(s, sField, []byte(tag), index)
	}

	// Untagged ⇒ ignore
	return nil
}
