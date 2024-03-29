package ukcore

import (
	"errors"
	"fmt"
	"reflect"
)

// =============================================================================
// Params
// =============================================================================

var ErrParams = errors.New("invalid parameters type")

type errorParams struct{ TypeName string }

func (errorParams) Is(target error) bool { return target == ErrParams }

type ErrorParamsKind struct {
	errorParams
	Kind reflect.Kind
}

func (ep errorParams) kind(k reflect.Kind) ErrorParamsKind {
	return ErrorParamsKind{errorParams: ep, Kind: k}
}

func (epk ErrorParamsKind) Error() string {
	return fmt.Sprintf(
		"parameters type '%s' is of invalid kind '%s'",
		epk.TypeName,
		epk.Kind,
	)
}

type ErrorParamsArgsConflict struct {
	errorParams
	FieldNameOne, FieldNameTwo string
}

func (ep errorParams) argsConflict(one, two string) ErrorParamsArgsConflict {
	return ErrorParamsArgsConflict{
		errorParams:  ep,
		FieldNameOne: one,
		FieldNameTwo: two,
	}
}

func (epac ErrorParamsArgsConflict) Error() string {
	return fmt.Sprintf(
		"parameters type '%s' contains conflicting '%s' fields '%s' and '%s'",
		epac.TypeName,
		paramsTagArgs,
		epac.FieldNameOne,
		epac.FieldNameTwo,
	)
}

type ErrorParamsFlagConflict struct {
	errorParams
	FieldNameOne, FieldNameTwo string
}

func (ep errorParams) flagConflict(one, two string) ErrorParamsFlagConflict {
	return ErrorParamsFlagConflict{
		errorParams:  ep,
		FieldNameOne: one,
		FieldNameTwo: two,
	}
}

func (epfc ErrorParamsFlagConflict) Error() string {
	return fmt.Sprintf(
		"parameters type '%s' contains conflicting '%s' fields '%s' and '%s'",
		epfc.TypeName,
		paramsTagFlag,
		epfc.FieldNameOne,
		epfc.FieldNameTwo,
	)
}

// =============================================================================
// Parse
// =============================================================================

// TODO

// =============================================================================
// Mux
// =============================================================================

// TODO
