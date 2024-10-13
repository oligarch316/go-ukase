package ukspec

import (
	"fmt"
	"reflect"
	"strings"
	"unicode/utf8"

	"github.com/oligarch316/go-ukase/internal/ierror"
)

// =============================================================================
// Flag
// =============================================================================

type Flag struct {
	FieldType  reflect.Type
	FieldName  string
	FieldIndex []int

	Elide FlagElide
	Names FlagNames
}

func (f Flag) String() string { return fmt.Sprintf("%s (%s)", f.FieldName, f.Names) }

func loadFlag(s *state, sField reflect.StructField, tag []byte, index int) error {
	s.Config.Log.Debug("loading flag field", "type", sField.Type, "name", sField.Name)

	if !sField.IsExported() {
		err := ierror.NewD("not exported")
		return InvalidFieldError{Trail: s.Scope.Trail, Field: sField, err: err}
	}

	flag := Flag{
		FieldType:  sField.Type,
		FieldName:  sField.Name,
		FieldIndex: append(s.Scope.FieldIndex, index),
		Elide:      newFlagElide(s.Config, sField),
	}

	if err := flag.Names.UnmarshalText(tag); err != nil {
		return InvalidFieldError{Trail: s.Scope.Trail, Field: sField, err: err}
	}

	for i, name := range flag.Names {
		flag.Names[i] = s.Scope.Prefix.String() + name
	}

	return s.InsertFlag(flag)
}

// =============================================================================
// FlagElide
// =============================================================================

type FlagElide struct {
	Allow      bool
	Consumable func(string) bool
}

func newFlagElide(config Config, sField reflect.StructField) FlagElide {
	type decider interface{ UkaseElide(string) bool }
	type allower interface{ UkaseElide() bool }
	type isBooler interface{ IsBoolFlag() bool }

	zero := reflect.New(sField.Type).Interface()

	if x, ok := zero.(decider); ok {
		return FlagElide{Allow: true, Consumable: x.UkaseElide}
	}

	if x, ok := zero.(allower); ok {
		return FlagElide{Allow: x.UkaseElide(), Consumable: config.ElideConsumable}
	}

	if x, ok := zero.(isBooler); config.ElideAllowIsBoolFlag && ok {
		return FlagElide{Allow: x.IsBoolFlag(), Consumable: config.ElideConsumable}
	}

	switch zero.(type) {
	case *bool, **bool:
		return FlagElide{Allow: config.ElideAllowBoolType, Consumable: config.ElideConsumable}
	}

	return FlagElide{Allow: false, Consumable: config.ElideConsumable}
}

// =============================================================================
// FlagNames
// =============================================================================

type FlagNames []string

func (fn FlagNames) String() string { return strings.Join(fn, " ") }

func (fn FlagNames) MarshalText() ([]byte, error) {
	str, err := fn.String(), fn.validate()
	return []byte(str), err
}

func (fn *FlagNames) UnmarshalText(text []byte) error {
	*fn = strings.Fields(string(text))
	return fn.validate()
}

func (fn FlagNames) validate() error {
	if len(fn) == 0 {
		return ierror.NewD("flag names empty")
	}

	for _, name := range fn {
		switch r, _ := utf8.DecodeRuneInString(name); r {
		case '-':
			return ierror.FmtD("flag name '%s' begins with reserved '-' character", name)
		case utf8.RuneError:
			// INTERNAL:
			// • ❬empty❭   ⇒ `strings.Fields` should never produce this
			// • ❬invalid❭ ⇒ go source should be strictly utf8 encoded
			return ierror.FmtI("flag name '%s' begins with unparsable character", name)
		}
	}

	return nil
}
