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
	FieldType reflect.Type
	FieldName string
	message   string
}

func (ed errorDecode) field(field reflect.Value, fieldName, message string) ErrorDecodeField {
	return ErrorDecodeField{
		errorDecode: ed,
		FieldType:   field.Type(),
		FieldName:   fieldName,
		message:     message,
	}
}

func (edft ErrorDecodeField) Error() string {
	return fmt.Sprintf("invalid field '%s' (%s): %s", edft.FieldName, edft.FieldType, edft.message)
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
