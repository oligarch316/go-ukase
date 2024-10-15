package ukdec

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/oligarch316/go-ukase/internal/ierror"
)

var (
	ErrInvalidField = errors.New("invalid field error")
	ErrUnknownField = errors.New("unknown field error")
)

type InvalidParametersError struct {
	Type reflect.Type
	err  error
}

type InvalidFieldError[T any] struct {
	err error
}

type UnknownFieldError[T any] struct {
	Input T
	err   error
}

var errIsTagged = ierror.IsTaggedFunc(ierror.ErrDec)

func (e InvalidParametersError) Is(t error) bool { return errIsTagged(t) }
func (e InvalidFieldError[T]) Is(t error) bool   { return errIsTagged(t, ErrInvalidField) }
func (e UnknownFieldError[T]) Is(t error) bool   { return errIsTagged(t, ErrUnknownField) }

func (e InvalidParametersError) Unwrap() error { return e.err }
func (e InvalidFieldError[T]) Unwrap() error   { return e.err }
func (e UnknownFieldError[T]) Unwrap() error   { return e.err }

func (e InvalidParametersError) Error() string {
	return fmt.Sprintf("invalid parameters '%s': %s", e.Type, e.err)
}

func (e InvalidFieldError[T]) Error() string { return e.err.Error() }
func (e UnknownFieldError[T]) Error() string { return e.err.Error() }
