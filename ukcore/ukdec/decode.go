package ukdec

import (
	"fmt"
	"log/slog"
	"reflect"

	"github.com/oligarch316/go-ukase/internal/ierror"
	"github.com/oligarch316/go-ukase/internal/ireflect"
	"github.com/oligarch316/go-ukase/ukcore"
	"github.com/oligarch316/go-ukase/ukcore/ukspec"
)

// =============================================================================
// Convenience
// =============================================================================

func DecodeFor[Params any](input ukcore.Input, opts ...Option) (Params, error) {
	params := new(Params)
	err := Decode(input, params, opts...)
	return *params, err
}

func Decode(input ukcore.Input, params any, opts ...Option) error {
	return NewDecoder(input, opts...).Decode(params)
}

// =============================================================================
// Decoder
// =============================================================================

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
			return UnknownFieldError[ukcore.Flag]{Source: flag, err: err}
		}

		fieldVal := paramsVal.EnsureFieldByIndex(flagSpec.FieldIndex)

		d.config.Log.Debug("decoding flag field",
			slog.Group("input", "name", flag.Name, "value", flag.Value),
			slog.Group("spec", "type", flagSpec.FieldType, "name", flagSpec.FieldName),
		)

		if err := decodeField(fieldVal, flag.Value); err != nil {
			return InvalidFieldError[ukcore.Flag]{Source: flag, Destination: fieldVal.Type(), err: err}
		}
	}

	return nil
}

func (d Decoder) decodeArguments(paramsVal ireflect.ParametersValue, paramsSpec ukspec.Parameters, args []ukcore.Argument) error {
	for _, arg := range args {
		argSpec, ok := paramsSpec.LookupArgument(arg.Position)
		if !ok {
			err := ierror.FmtU("unknown argument position '%d' (%s)", arg.Position, arg.Value)
			return UnknownFieldError[ukcore.Argument]{Source: arg, err: err}
		}

		fieldVal := paramsVal.EnsureFieldByIndex(argSpec.FieldIndex)

		d.config.Log.Debug("decoding argument field",
			slog.Group("input", "position", arg.Position, "value", arg.Value),
			slog.Group("spec", "type", argSpec.FieldType, "name", argSpec.FieldName),
		)

		if err := decodeField(fieldVal, arg.Value); err != nil {
			return InvalidFieldError[ukcore.Argument]{Source: arg, Destination: fieldVal.Type(), err: err}
		}
	}

	return nil
}
