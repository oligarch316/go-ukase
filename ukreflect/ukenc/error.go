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
	message string
}

func (ed errorDecode) params(message string) ErrorDecodeParams {
	return ErrorDecodeParams{errorDecode: ed, message: message}
}

func (edtt ErrorDecodeParams) Error() string {
	return fmt.Sprintf("invalid parameters (%s): %s", edtt.ParamsType, edtt.message)
}

type ErrorDecodeField struct {
	errorDecode
	error
	FieldType reflect.Type
	FieldName string
}

func (ed errorDecode) field(field reflect.Value, fieldName string, err error) ErrorDecodeField {
	return ErrorDecodeField{
		errorDecode: ed,
		error:       err,
		FieldType:   field.Type(),
		FieldName:   fieldName,
	}
}

func (edft ErrorDecodeField) Unwrap() error { return edft.error }

func (edft ErrorDecodeField) Error() string {
	return fmt.Sprintf("invalid parameters field '%s' (%s): %s", edft.FieldName, edft.FieldType, edft.error)
}

type ErrorDecodeFlagName struct {
	errorDecode
	Flag ukcore.Flag
}

func (ed errorDecode) flagName(flag ukcore.Flag) ErrorDecodeFlagName {
	return ErrorDecodeFlagName{errorDecode: ed, Flag: flag}
}

func (edfn ErrorDecodeFlagName) Error() string {
	return fmt.Sprintf("invalid flag name '%s'", edfn.Flag.Name)
}

type ErrorDecodeFlagValue struct {
	errorDecode
	error
	Flag ukcore.Flag
}

func (ed errorDecode) flagValue(flag ukcore.Flag, err error) ErrorDecodeFlagValue {
	return ErrorDecodeFlagValue{errorDecode: ed, error: err, Flag: flag}
}

func (edfv ErrorDecodeFlagValue) Unwrap() error { return edfv.error }

func (edfv ErrorDecodeFlagValue) Error() string {
	return fmt.Sprintf("invalid value for flag '%s': %s", edfv.Flag.Name, edfv.error)
}
