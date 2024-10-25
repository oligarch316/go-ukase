package ukspec

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/oligarch316/ukase/internal/ierror"
)

var ErrConflict = errors.New("conflict error")

type ConflictError[T fmt.Stringer] struct {
	Trail            []Inline
	Original, Update T
	err              error
}

type InvalidFieldError struct {
	Trail []Inline
	Field reflect.StructField
	err   error
}

type InvalidParametersError struct {
	Type reflect.Type
	err  error
}

var errIsTagged = ierror.IsTaggedFunc(ierror.ErrSpec)

func (ConflictError[T]) Is(t error) bool       { return errIsTagged(t, ErrConflict) }
func (InvalidFieldError) Is(t error) bool      { return errIsTagged(t) }
func (InvalidParametersError) Is(t error) bool { return errIsTagged(t) }

func (e ConflictError[T]) Unwrap() error       { return e.err }
func (e InvalidFieldError) Unwrap() error      { return e.err }
func (e InvalidParametersError) Unwrap() error { return e.err }

func (e ConflictError[T]) Error() string {
	return fmt.Sprintf("conflict between '%s' and '%s': %s", e.Update, e.Original, e.err)
}

func (e InvalidFieldError) Error() string {
	return fmt.Sprintf("invalid field '%s': %s", e.Field.Name, e.err)
}

func (e InvalidParametersError) Error() string {
	return fmt.Sprintf("invalid parameters '%s': %s", e.Type, e.err)
}
