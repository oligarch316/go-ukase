package ukdec

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/oligarch316/go-ukase/ukcore"
	"github.com/oligarch316/go-ukase/ukcore/ukspec"
)

type Decoder struct {
	config Config
	input  ukcore.Input
}

func NewDecoder(input ukcore.Input, opts ...Option) *Decoder {
	return &Decoder{
		config: newConfig(opts),
		input:  input,
	}
}

func (d *Decoder) Decode(params any) error {
	paramsVal, err := d.loadValue(params)
	if err != nil {
		return err
	}

	spec, err := d.loadSpec(paramsVal)
	if err != nil {
		return err
	}

	for _, flag := range d.input.Flags {
		if err := d.decodeFlag(paramsVal, spec, flag); err != nil {
			return err
		}
	}

	return d.decodeArgs(paramsVal, spec, d.input.Arguments)
}

func (Decoder) loadValue(v any) (ukcore.ParamsValue, error) {
	paramsVal, err := ukcore.NewParamsValue(v)
	if err != nil {
		tmpVal := reflect.ValueOf(v)
		err = decodeErr(tmpVal).params(err)
	}

	return paramsVal, err
}

func (d Decoder) loadSpec(paramsVal ukcore.ParamsValue) (ukspec.Parameters, error) {
	spec, err := ukspec.NewParameters(paramsVal.Type(), d.config.Spec...)
	if err != nil {
		// TODO: Wrap error appropriately
	}

	return spec, err
}

func (d Decoder) decodeFlag(paramsVal ukcore.ParamsValue, spec ukspec.Parameters, flag ukcore.Flag) error {
	flagSpec, ok := spec.LookupFlag(flag.Name)
	if !ok {
		return decodeErr(paramsVal.Value).flagName(flag)
	}

	fieldVal := paramsVal.EnsureFieldByIndex(flagSpec.FieldIndex)

	if err := decode(fieldVal, flag.Value); err != nil {
		if errors.Is(err, errUnsupportedKind) {
			return decodeErr(paramsVal.Value).field(fieldVal, flagSpec.FieldName, err)
		}

		return decodeErr(paramsVal.Value).flagValue(flag, err)
	}

	return nil
}

func (d Decoder) decodeArgs(paramsVal ukcore.ParamsValue, spec ukspec.Parameters, args []string) error {
	for pos, arg := range args {
		argSpec, ok := spec.LookupArgument(pos)
		if !ok {
			return fmt.Errorf("[TODO decodeArgs] invalid argument position '%d'", pos)
		}

		fieldVal := paramsVal.EnsureFieldByIndex(argSpec.FieldIndex)

		if err := decode(fieldVal, arg); err != nil {
			return fmt.Errorf("[TODO decodeArgs] decode error: %w", err)
		}
	}

	return nil
}
