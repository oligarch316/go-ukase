package ierror

import (
	"errors"
	"fmt"
)

// =============================================================================
// Library
// =============================================================================

var ErrAny = errors.New("ukase error")

func IsTagged(target error, tags ...error) bool {
	for _, tag := range tags {
		if target == tag {
			return true
		}
	}
	return false
}

func IsTaggedFunc(static ...error) func(error, ...error) bool {
	return func(target error, tags ...error) bool {
		return IsTagged(target, append(static, tags...)...)
	}
}

// =============================================================================
// Package
// =============================================================================

var (
	ErrDec  = errors.New("ukdec error")
	ErrExec = errors.New("ukexec error")
	ErrInit = errors.New("ukinit error")
	ErrSpec = errors.New("ukspec error")
)

// =============================================================================
// Severity
// =============================================================================

var (
	ErrInternal  = errors.New("internal error")
	ErrDeveloper = errors.New("developer error")
	ErrUser      = errors.New("user error")
)

type internalError struct{ error }
type developerError struct{ error }
type userError struct{ error }

func (internalError) Is(target error) bool  { return IsTagged(target, ErrAny, ErrInternal) }
func (developerError) Is(target error) bool { return IsTagged(target, ErrAny, ErrDeveloper) }
func (userError) Is(target error) bool      { return IsTagged(target, ErrAny, ErrUser) }

func (e internalError) Unwrap() error  { return e.error }
func (e developerError) Unwrap() error { return e.error }
func (e userError) Unwrap() error      { return e.error }

func I(err error) error { return internalError{err} }
func D(err error) error { return developerError{err} }
func U(err error) error { return userError{err} }

func NewI(text string) error { return I(errors.New(text)) }
func NewD(text string) error { return D(errors.New(text)) }
func NewU(text string) error { return U(errors.New(text)) }

func FmtI(format string, a ...any) error { return I(fmt.Errorf(format, a...)) }
func FmtD(format string, a ...any) error { return D(fmt.Errorf(format, a...)) }
func FmtU(format string, a ...any) error { return U(fmt.Errorf(format, a...)) }
