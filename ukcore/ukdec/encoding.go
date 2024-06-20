package ukdec

import (
	"errors"
	"reflect"

	"github.com/oligarch316/go-ukase/ukcore/ukexec"
	"github.com/oligarch316/go-ukase/ukcore/ukspec"
	"github.com/oligarch316/go-ukase/ukreflect"
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
	val, err := d.loadValue(params)
	if err != nil {
		return err
	}

	spec, err := d.loadSpec(val)
	if err != nil {
		return err
	}

	for _, flag := range d.input.Flags {
		if err := d.decodeFlag(val, spec, flag); err != nil {
			return err
		}
	}

	return d.decodeArgs(val, spec, d.input.Args)
}

func (Decoder) loadValue(v any) (reflect.Value, error) {
	val, err := ukreflect.LoadValueOf(v)
	if err != nil {
		val = reflect.ValueOf(v)
		err = decodeErr(val).params(err)
	}

	return val, err
}

func (d Decoder) loadSpec(structVal reflect.Value) (ukspec.Params, error) {
	spec, err := ukspec.New(structVal.Type(), d.config.Spec...)
	if err != nil {
		// TODO: Wrap error appropriately
	}

	return spec, err
}

func (d Decoder) decodeFlag(structVal reflect.Value, spec ukspec.Params, flag ukexec.InputFlag) error {
	flagSpec, ok := spec.FlagIndex[flag.Name]
	if !ok {
		return decodeErr(structVal).flagName(flag)
	}

	fieldVal := ukreflect.LoadFieldByIndex(structVal, flagSpec.FieldIndex)

	if err := decode(fieldVal, flag.Value); err != nil {
		if errors.Is(err, errUnsupportedKind) {
			return decodeErr(structVal).field(fieldVal, flagSpec.FieldName, err)
		}

		return decodeErr(structVal).flagValue(flag, err)
	}

	return nil
}

func (d Decoder) decodeArgs(structVal reflect.Value, spec ukspec.Params, args []string) error {
	switch {
	case len(args) == 0:
		return nil
	case spec.Args == nil:
		return errors.New("[TODO decodeArgs] have args but not spec")
	}

	argsVal := ukreflect.LoadFieldByIndex(structVal, spec.Args.FieldIndex)

	for _, arg := range args {
		if err := decode(argsVal, arg); err != nil {
			return err
		}
	}

	return nil
}
