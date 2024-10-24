package itest

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/oligarch316/go-ukase/internal/ierror"
	"gotest.tools/v3/assert"
	"gotest.tools/v3/assert/cmp"
)

type Runner[T any] func(subtest T) (name string, comparison cmp.Comparison)

func (r Runner[T]) Run(t *testing.T, subtests ...T) {
	for _, subtest := range subtests {
		name, comparison := r(subtest)
		t.Run(name, func(t *testing.T) { assert.Check(t, comparison) })
	}
}

// =============================================================================
// Message Formatting
// =============================================================================

type failureFormat []string

func (ff failureFormat) Result(a ...any) cmp.Result {
	format := strings.Join(ff, "\n")
	message := fmt.Sprintf(format, a...)
	return cmp.ResultFailure(message)
}

// =============================================================================
// Combination Comparison
// =============================================================================

func CmpSequence(seq ...cmp.Comparison) cmp.Comparison {
	return func() cmp.Result {
		for _, item := range seq {
			if result := item(); !result.Success() {
				return result
			}
		}

		return cmp.ResultSuccess
	}
}

// =============================================================================
// Error Comparisons
// =============================================================================

// -----------------------------------------------------------------------------
// Error Comparisons› Basic
// -----------------------------------------------------------------------------

func CmpErrorAs[Expected error](actual error) cmp.Comparison {
	failFormat := failureFormat{
		"",
		"unexpected error type",
		"  actual:   %T",
		"  expected: %s",
		"",
		"  message: %s",
		"",
	}

	return func() cmp.Result {
		if expected := new(Expected); errors.As(actual, expected) {
			return cmp.ResultSuccess
		}

		return failFormat.Result(
			actual,
			reflect.TypeFor[Expected](),
			actual,
		)
	}
}

func CmpErrorIs(actual, expected error) cmp.Comparison {
	failFormat := failureFormat{
		"",
		"unexpected error value",
		"  actual:   %s",
		"  expected: %s",
		"",
		"  type   : %T",
		"",
	}

	return func() cmp.Result {
		if errors.Is(actual, expected) {
			return cmp.ResultSuccess
		}

		return failFormat.Result(
			actual,
			expected,
			actual,
		)
	}
}

// -----------------------------------------------------------------------------
// Error Comparisons› Severity
// -----------------------------------------------------------------------------

func CmpErrorAsI[Expected error](actual error) cmp.Comparison {
	cmpAs := CmpErrorAs[Expected](actual)
	cmpIs := CmpErrorIsI(actual)
	return CmpSequence(cmpAs, cmpIs)
}

func CmpErrorAsD[Expected error](actual error) cmp.Comparison {
	cmpAs := CmpErrorAs[Expected](actual)
	cmpIs := CmpErrorIsD(actual)
	return CmpSequence(cmpAs, cmpIs)
}

func CmpErrorAsU[Expected error](actual error) cmp.Comparison {
	cmpAs := CmpErrorAs[Expected](actual)
	cmpIs := CmpErrorIsU(actual)
	return CmpSequence(cmpAs, cmpIs)
}

var CmpErrorIsI = errorSeverity{I: true}.compareIs
var CmpErrorIsD = errorSeverity{D: true}.compareIs
var CmpErrorIsU = errorSeverity{U: true}.compareIs

type errorSeverity struct{ I, D, U bool }

func (eSev errorSeverity) compareIs(actual error) cmp.Comparison {
	failFormat := failureFormat{
		"",
		"unexpected error severity",
		"  actual:   %-9t %-10t %t",
		"  expected: %-9t %-10t %t",
		"            ──────────────────────────",
		"            Internal  Developer  User",
		"",
		"  type:    %T",
		"  message: %s",
		"",
	}

	return func() cmp.Result {
		aSev := errorSeverity{
			I: errors.Is(actual, ierror.ErrInternal),
			D: errors.Is(actual, ierror.ErrDeveloper),
			U: errors.Is(actual, ierror.ErrUser),
		}

		if aSev == eSev {
			return cmp.ResultSuccess
		}

		return failFormat.Result(
			aSev.I, aSev.D, aSev.U,
			eSev.I, eSev.D, eSev.U,
			actual,
			actual,
		)
	}
}
