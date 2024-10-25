package ukdec

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/oligarch316/ukase/internal/ierror"
)

var (
	ErrInvalidField = errors.New("invalid field error")
	ErrUnknownField = errors.New("unknown field error")
)

type InvalidParametersError struct {
	Type reflect.Type
	err  error
}

type InvalidFieldError[S any] struct {
	Source      S
	Destination reflect.Type
	err         error
}

type UnknownFieldError[S any] struct {
	Source S
	err    error
}

var errIsTagged = ierror.IsTaggedFunc(ierror.ErrDec)

func (e InvalidParametersError) Is(t error) bool { return errIsTagged(t) }
func (e InvalidFieldError[S]) Is(t error) bool   { return errIsTagged(t, ErrInvalidField) }
func (e UnknownFieldError[S]) Is(t error) bool   { return errIsTagged(t, ErrUnknownField) }

func (e InvalidParametersError) Unwrap() error { return e.err }
func (e InvalidFieldError[S]) Unwrap() error   { return e.err }
func (e UnknownFieldError[S]) Unwrap() error   { return e.err }

func (e InvalidParametersError) Error() string {
	return fmt.Sprintf("invalid parameters '%s': %s", e.Type, e.err)
}

func (e InvalidFieldError[S]) Error() string { return e.err.Error() }
func (e UnknownFieldError[S]) Error() string { return e.err.Error() }
