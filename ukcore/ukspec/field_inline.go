package ukspec

import (
	"fmt"
	"reflect"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/oligarch316/go-ukase/internal/ierror"
)

// =============================================================================
// Inline
// =============================================================================

type Inline struct {
	FieldType  reflect.Type
	FieldName  string
	FieldIndex []int

	Prefix InlinePrefix
}

func (i Inline) String() string {
	if i.Prefix == "" {
		return i.FieldName
	}
	return fmt.Sprintf("%s (%s)", i.FieldName, i.Prefix)
}

func loadInline(s *state, sField reflect.StructField, tag []byte, index int) error {
	s.Config.Log.Debug("loading inline field", "type", sField.Type, "name", sField.Name)

	if !sField.IsExported() {
		err := ierror.NewD("not exported")
		return InvalidFieldError{Trail: s.Scope.Trail, Field: sField, err: err}
	}

	inlineType := sField.Type
	for inlineType.Kind() == reflect.Pointer {
		inlineType = inlineType.Elem()
	}

	if inlineType.Kind() != reflect.Struct {
		err := ierror.NewD("inline not a struct or struct pointer")
		return InvalidFieldError{Trail: s.Scope.Trail, Field: sField, err: err}
	}

	inline := Inline{
		FieldType:  inlineType,
		FieldName:  sField.Name,
		FieldIndex: append(s.Scope.FieldIndex, index),
	}

	if err := inline.Prefix.UnmarshalText(tag); err != nil {
		return InvalidFieldError{Trail: s.Scope.Trail, Field: sField, err: err}
	}

	inline.Prefix = s.Scope.Prefix + inline.Prefix

	return s.InsertInline(inline)
}

// =============================================================================
// InlinePrefix
// =============================================================================

type InlinePrefix string

func (ip InlinePrefix) String() string { return string(ip) }

func (ip InlinePrefix) MarshalText() ([]byte, error) {
	str, err := ip.String(), ip.validate()
	return []byte(str), err
}

func (ip *InlinePrefix) UnmarshalText(text []byte) error {
	*ip = InlinePrefix(text)
	return ip.validate()
}

func (ip InlinePrefix) validate() error {
	s := string(ip)

	if strings.ContainsFunc(s, unicode.IsSpace) {
		return ierror.FmtD("inline prefix '%s' contains whitespace", s)
	}

	switch r, size := utf8.DecodeRuneInString(s); {
	case size == 0:
		return nil
	case r == '-':
		return ierror.FmtD("inline prefix '%s' begins with reserved '-' character", s)
	case r == utf8.RuneError:
		// INTERNAL:
		// • ❬empty❭   ⇒ handled in preceding case
		// • ❬invalid❭ ⇒ go source should be strictly utf8 encoded
		return ierror.FmtI("inline prefix '%s' begins with unparsable character", s)
	default:
		return nil
	}
}
