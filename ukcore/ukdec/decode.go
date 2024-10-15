package ukdec

import (
	"fmt"
	"reflect"

	"github.com/oligarch316/go-ukase/internal/ierror"
	"github.com/oligarch316/go-ukase/internal/ireflect"
	"github.com/oligarch316/go-ukase/ukcore"
	"github.com/oligarch316/go-ukase/ukcore/ukspec"
)

type Decoder struct {
	config Config
	input  ukcore.Input
}

func NewDecoder(input ukcore.Input, opts ...Option) *Decoder {
	config := newConfig(opts)
	return &Decoder{config: config, input: input}
}

func (d *Decoder) Decode(params any) error {
	d.config.Log.Info("decoding parameters", "type", fmt.Sprintf("%T", params))

	paramsVal, err := ireflect.NewParametersValue(params)
	if err != nil {
		return InvalidParametersError{Type: reflect.TypeOf(params), err: err}
	}

	paramsSpec, err := ukspec.NewParameters(paramsVal.Type(), d.config.Spec...)
	if err != nil {
		return err
	}

	if err := d.decodeFlags(paramsVal, paramsSpec, d.input.Flags); err != nil {
		return err
	}

	return d.decodeArguments(paramsVal, paramsSpec, d.input.Arguments)
}

func (d Decoder) decodeFlags(paramsVal ireflect.ParametersValue, paramsSpec ukspec.Parameters, flags []ukcore.Flag) error {
	for _, flag := range flags {
		flagSpec, ok := paramsSpec.LookupFlag(flag.Name)
		if !ok {
			err := ierror.FmtU("unknown flag name '%s' (%s)", flag.Name, flag.Value)
			return UnknownFieldError[ukcore.Flag]{Input: flag, err: err}
		}

		d.config.Log.Debug("decoding flag field", "type", flagSpec.FieldType, "name", flagSpec.FieldName)
		fieldVal := paramsVal.EnsureFieldByIndex(flagSpec.FieldIndex)

		if err := decode(fieldVal, flag.Value); err != nil {
			return fmt.Errorf("[TODO decodeFlags] decodeField error: %w", err)
		}
	}

	return nil
}

func (d Decoder) decodeArguments(paramsVal ireflect.ParametersValue, paramsSpec ukspec.Parameters, args []ukcore.Argument) error {
	for _, arg := range args {
		argSpec, ok := paramsSpec.LookupArgument(arg.Position)
		if !ok {
			err := ierror.FmtU("unknown argument position '%d' (%s)", arg.Position, arg.Value)
			return UnknownFieldError[ukcore.Argument]{Input: arg, err: err}
		}

		d.config.Log.Debug("decoding argument field", "type", argSpec.FieldType, "name", argSpec.FieldName)
		fieldVal := paramsVal.EnsureFieldByIndex(argSpec.FieldIndex)

		if err := decode(fieldVal, arg.Value); err != nil {
			return fmt.Errorf("[TODO decodeArguments] decodeField error: %w", err)
		}
	}

	return nil
}
