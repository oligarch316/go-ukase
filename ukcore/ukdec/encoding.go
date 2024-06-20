package ukdec

import (
	"errors"
	"reflect"

	"github.com/oligarch316/go-ukase/ukcore"
	"github.com/oligarch316/go-ukase/ukcore/ukexec"
	"github.com/oligarch316/go-ukase/ukcore/ukspec"
)

type Decoder struct {
	config Config
	input  ukexec.Input
}

func NewDecoder(input ukexec.Input, opts ...Option) *Decoder {
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

	return d.decodeArgs(paramsVal, spec, d.input.Args)
}

func (Decoder) loadValue(v any) (ukcore.ParamsValue, error) {
	paramsVal, err := ukcore.NewParamsValue(v)
	if err != nil {
		tmpVal := reflect.ValueOf(v)
		err = decodeErr(tmpVal).params(err)
	}

	return paramsVal, err
}

func (d Decoder) loadSpec(paramsVal ukcore.ParamsValue) (ukspec.Params, error) {
	spec, err := ukspec.New(paramsVal.Type(), d.config.Spec...)
	if err != nil {
		// TODO: Wrap error appropriately
	}

	return spec, err
}

func (d Decoder) decodeFlag(paramsVal ukcore.ParamsValue, spec ukspec.Params, flag ukexec.InputFlag) error {
	flagSpec, ok := spec.FlagIndex[flag.Name]
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

func (d Decoder) decodeArgs(paramsVal ukcore.ParamsValue, spec ukspec.Params, args []string) error {
	switch {
	case len(args) == 0:
		return nil
	case spec.Args == nil:
		return errors.New("[TODO decodeArgs] have args but not spec")
	}

	argsVal := paramsVal.EnsureFieldByIndex(spec.Args.FieldIndex)

	for _, arg := range args {
		if err := decode(argsVal, arg); err != nil {
			return err
		}
	}

	return nil
}
