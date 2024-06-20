package ukenc

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/oligarch316/go-ukase/ukcore"
)

var ErrDecode = errors.New("decode error")

type errorDecode struct{ ParamsType reflect.Type }

func decodeErr(paramsVal reflect.Value) errorDecode {
	return errorDecode{ParamsType: paramsVal.Type()}
}

func (errorDecode) Is(target error) bool { return target == ErrDecode }

type ErrorDecodeParams struct {
	errorDecode
	err error
}

func (ed errorDecode) params(err error) ErrorDecodeParams {
	return ErrorDecodeParams{errorDecode: ed, err: err}
}

func (edtt ErrorDecodeParams) Unwrap() error { return edtt.err }

func (edtt ErrorDecodeParams) Error() string {
	return fmt.Sprintf("invalid parameters (%s): %s", edtt.ParamsType, edtt.err)
}

type ErrorDecodeField struct {
	errorDecode
	FieldType reflect.Type
	FieldName string
	err       error
}

func (ed errorDecode) field(field reflect.Value, fieldName string, err error) ErrorDecodeField {
	return ErrorDecodeField{
		errorDecode: ed,
		FieldType:   field.Type(),
		FieldName:   fieldName,
		err:         err,
	}
}

func (edft ErrorDecodeField) Unwrap() error { return edft.err }

func (edft ErrorDecodeField) Error() string {
	return fmt.Sprintf("invalid parameters field '%s' (%s): %s", edft.FieldName, edft.FieldType, edft.err)
}

type ErrorDecodeFlagName struct {
	errorDecode
	Flag ukcore.InputFlag
}

func (ed errorDecode) flagName(flag ukcore.InputFlag) ErrorDecodeFlagName {
	return ErrorDecodeFlagName{errorDecode: ed, Flag: flag}
}

func (edfn ErrorDecodeFlagName) Error() string {
	return fmt.Sprintf("invalid flag name '%s'", edfn.Flag.Name)
}

type ErrorDecodeFlagValue struct {
	errorDecode
	Flag ukcore.InputFlag
	err  error
}

func (ed errorDecode) flagValue(flag ukcore.InputFlag, err error) ErrorDecodeFlagValue {
	return ErrorDecodeFlagValue{errorDecode: ed, err: err, Flag: flag}
}

func (edfv ErrorDecodeFlagValue) Unwrap() error { return edfv.err }

func (edfv ErrorDecodeFlagValue) Error() string {
	return fmt.Sprintf("invalid value for flag '%s': %s", edfv.Flag.Name, edfv.err)
}
