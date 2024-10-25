package ukspec

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/oligarch316/ukase/internal/ierror"
)

// =============================================================================
// Argument
// =============================================================================

type Argument struct {
	FieldType  reflect.Type
	FieldName  string
	FieldIndex []int

	Position ArgumentPosition
}

func (a Argument) String() string { return fmt.Sprintf("%s (%s)", a.FieldName, a.Position) }

func loadArgument(s *state, sField reflect.StructField, tag []byte, index int) error {
	s.Config.Log.Debug("loading argument field", "type", sField.Type, "name", sField.Name)

	if !sField.IsExported() {
		err := ierror.NewD("not exported")
		return InvalidFieldError{Trail: s.Scope.Trail, Field: sField, err: err}
	}

	argument := Argument{
		FieldType:  sField.Type,
		FieldName:  sField.Name,
		FieldIndex: append(s.Scope.FieldIndex, index),
	}

	if err := argument.Position.UnmarshalText(tag); err != nil {
		return InvalidFieldError{Trail: s.Scope.Trail, Field: sField, err: err}
	}

	return s.InsertArgument(argument)
}

// =============================================================================
// ArgumentPosition
// =============================================================================

type ArgumentPosition struct{ Low, High *uint }

func (ap ArgumentPosition) String() string {
	switch lowEmpty, highEmpty := ap.Low == nil, ap.High == nil; {
	case lowEmpty && highEmpty:
		return ":"
	case lowEmpty:
		return fmt.Sprintf(":%d", *ap.High)
	case highEmpty:
		return fmt.Sprintf("%d:", *ap.Low)
	default:
		return fmt.Sprintf("%d:%d", *ap.Low, *ap.High)
	}
}

func (ap ArgumentPosition) MarshalText() ([]byte, error) {
	str, err := ap.String(), ap.validate()
	return []byte(str), err
}

func (ap *ArgumentPosition) UnmarshalText(text []byte) error {
	tag := bytes.TrimSpace(text)
	if len(tag) == 0 {
		return ierror.NewD("argument position empty")
	}

	validate := func(errs ...error) error {
		for _, err := range errs {
			if err != nil {
				return err
			}
		}
		return ap.validate()
	}

	atoui := func(s string) (uint, error) {
		switch ui64, err := strconv.ParseUint(s, 10, 0); {
		case err == nil:
			return uint(ui64), nil
		case errors.Is(err, strconv.ErrRange):
			return 0, ierror.FmtD("argument position index '%s' is out of range", s)
		case errors.Is(err, strconv.ErrSyntax):
			return 0, ierror.FmtD("argument position index '%s' exhibits invalid uint syntax", s)
		default:
			// INTERNAL:
			// • ❬range/syntax error❭ ⇒ handled in preceding cases
			// • ❬base/bitSize error❭ ⇒ arguments to `ParseUint` should be valid
			return 0, ierror.FmtI("argument position index '%s' triggers unexpected parse error", s)
		}
	}

	lowTag, highTag, sepFound := strings.Cut(string(tag), ":")
	if !sepFound {
		// "#" ⇒ [#,#+1)
		low, lowErr := atoui(lowTag)
		high := low + 1

		ap.Low, ap.High = &low, &high
		return validate(lowErr)
	}

	switch lowEmpty, highEmpty := lowTag == "", highTag == ""; {
	case lowEmpty && highEmpty:
		// ":" ⇒ [-∞,∞]
		ap.Low, ap.High = nil, nil
		return validate()
	case lowEmpty:
		// ":#" ⇒ [-∞,#)
		high, highErr := atoui(highTag)
		ap.Low, ap.High = nil, &high
		return validate(highErr)
	case highEmpty:
		// "#:" ⇒ [#,∞]
		low, lowErr := atoui(lowTag)
		ap.Low, ap.High = &low, nil
		return validate(lowErr)
	}

	// "#:#" ⇒ [#,#)
	low, lowErr := atoui(lowTag)
	high, highErr := atoui(highTag)

	ap.Low, ap.High = &low, &high
	return validate(lowErr, highErr)
}

func (ap ArgumentPosition) validate() error {
	switch {
	case ap.High == nil:
		return nil
	case ap.Low == nil && *ap.High != 0:
		return nil
	case ap.Low != nil && *ap.High > *ap.Low:
		return nil
	default:
		return ierror.FmtD("argument position '%s' describes a nonsensical range", ap)
	}
}
