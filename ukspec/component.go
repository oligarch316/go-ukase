package ukspec

import (
	"errors"
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
// Flag
// =============================================================================

type Flag struct {
	Type       reflect.Type
	Elide      Elide
	FlagName   string
	FieldName  string
	FieldIndex []int
}

func newFlag(config Config, field reflect.StructField, name string, index []int) (Flag, error) {
	if !field.IsExported() {
		return Flag{}, errors.New("[TODO newFlag] not exported")
	}

	switch r, _ := utf8.DecodeRuneInString(name); r {
	case utf8.RuneError:
		return Flag{}, errors.New("[TODO newFlag] name gave utf8.RuneError")
	case '-':
		return Flag{}, errors.New("[TODO newFlag] name begins with '-'")
	}

	elide := newElide(config.Elide, field)
	flag := Flag{
		Type:       field.Type,
		Elide:      elide,
		FlagName:   name,
		FieldName:  field.Name,
		FieldIndex: index,
	}

	return flag, nil
}

// =============================================================================
// Inline
// =============================================================================

type Inline struct {
	Type       reflect.Type
	FieldName  string
	FieldIndex []int
	Inlines    map[reflect.Type]Inline
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
		Inlines:    make(map[reflect.Type]Inline),
	}

	return inline, nil
}
