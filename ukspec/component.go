package ukspec

import (
	"errors"
	"fmt"
	"reflect"
	"unicode/utf8"
)

// =============================================================================
// Args
// =============================================================================

type Args struct {
	Type       reflect.Type
	FieldName  string
	FieldIndex []int
}

func newArgs(field reflect.StructField, index []int) (Args, error) {
	if !field.IsExported() {
		return Args{}, errors.New("[TODO newArgs] not exported")
	}

	args := Args{
		Type:       field.Type,
		FieldName:  field.Name,
		FieldIndex: index,
	}

	return args, nil
}

// =============================================================================
// Inline
// =============================================================================

type Inline struct {
	Type       reflect.Type
	FieldName  string
	FieldIndex []int
}

func newInline(field reflect.StructField, index []int) (Inline, error) {
	if !field.IsExported() {
		return Inline{}, errors.New("[TODO newInline] not exported")
	}

	fieldType := field.Type
	if fieldType.Kind() == reflect.Pointer {
		fieldType = fieldType.Elem()
	}

	if fieldType.Kind() != reflect.Struct {
		return Inline{}, errors.New("[TODO newInline] not a struct (or struct pointer)")
	}

	inline := Inline{
		Type:       fieldType,
		FieldName:  field.Name,
		FieldIndex: index,
	}

	return inline, nil
}

// =============================================================================
// Flag
// =============================================================================

type Flag struct {
	Type       reflect.Type
	Elide      Elide
	FieldName  string
	FieldIndex []int
	FlagNames  []string
}

func newFlag(config Config, field reflect.StructField, names []string, index []int) (Flag, error) {
	if !field.IsExported() {
		return Flag{}, errors.New("[TODO newFlag] not exported")
	}

	for _, name := range names {
		switch r, _ := utf8.DecodeRuneInString(name); r {
		case utf8.RuneError:
			return Flag{}, fmt.Errorf("[TODO newFlag] flag name '%s' gave utf8.RuneError", name)
		case '-':
			return Flag{}, fmt.Errorf("[TODO newFlag] flag name '%s' beings with '-'", name)
		}
	}

	elide := newElide(config, field)
	flag := Flag{
		Type:       field.Type,
		Elide:      elide,
		FieldName:  field.Name,
		FieldIndex: index,
		FlagNames:  names,
	}

	return flag, nil
}

// =============================================================================
// Elide
// =============================================================================

type Elide struct {
	Allow      bool
	Consumable func(string) bool
}

func newElide(config Config, field reflect.StructField) Elide {
	type decider interface{ UkaseElide(string) bool }
	type allower interface{ UkaseElide() bool }
	type isBooler interface{ IsBoolFlag() bool }

	elide := Elide{Allow: false, Consumable: config.ElideDefaultConsumable}
	zero := reflect.New(field.Type).Interface()

	if x, ok := zero.(decider); ok {
		elide.Allow, elide.Consumable = true, x.UkaseElide
		return elide
	}

	if x, ok := zero.(allower); ok {
		elide.Allow = x.UkaseElide()
		return elide
	}

	if config.ElideIsBoolFlag {
		if x, ok := zero.(isBooler); ok {
			elide.Allow = x.IsBoolFlag()
			return elide
		}
	}

	if config.ElideBoolType {
		switch zero.(type) {
		case *bool, **bool:
			elide.Allow = true
			return elide
		}
	}

	return elide
}
