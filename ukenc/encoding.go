package ukenc

import (
	"errors"
	"reflect"

	"github.com/oligarch316/go-ukase/ukcore"
)

type Decoder struct{ input ukcore.Input }

func NewDecoder(input ukcore.Input) *Decoder { return &Decoder{input: input} }

func (d *Decoder) Decode(params any) error {
	val, err := d.loadValue(params)
	if err != nil {
		return err
	}

	info, err := d.loadInfo(val)
	if err != nil {
		return err
	}

	for _, flag := range d.input.Flags {
		if err := d.decodeFlag(val, info, flag); err != nil {
			return err
		}
	}

	return d.decodeArgs(val, info, d.input.Args)
}

func (Decoder) loadValue(v any) (reflect.Value, error) {
	val := reflect.ValueOf(v)
	switch {
	case val.Kind() != reflect.Pointer:
		return val, decodeErr(val).params("destination is not a pointer")
	case val.IsNil():
		return val, decodeErr(val).params("destination is a nil pointer")
	}

	elem := val.Elem()
	if elem.Kind() != reflect.Struct {
		return elem, decodeErr(elem).params("destination does not point to a struct")
	}

	return elem, nil
}

func (Decoder) loadInfo(structVal reflect.Value) (ukcore.ParamsInfo, error) {
	// TODO: Consider caching params info

	info, err := ukcore.NewParamsInfo(structVal.Type())
	if err != nil {
		// TODO: Wrap error appropriately
	}

	return info, err
}

func (Decoder) decodeFlag(structVal reflect.Value, info ukcore.ParamsInfo, flag ukcore.Flag) error {
	flagInfo, ok := info.Flags[flag.Name]
	if !ok {
		return decodeErr(structVal).flagName(flag)
	}

	// TODO: Do I need to handle possibly "stepping through a nil pointer"? (ugh)
	fieldVal := structVal.FieldByIndex(flagInfo.FieldIndex)

	if err := decode(fieldVal, flag.Value); err != nil {
		if errors.Is(err, errUnsupportedKind) {
			return decodeErr(structVal).field(fieldVal, flagInfo.FieldName, err)
		}

		return decodeErr(structVal).flagValue(flag, err)
	}

	return nil
}

func (Decoder) decodeArgs(structVal reflect.Value, info ukcore.ParamsInfo, args []string) error {
	switch {
	case len(args) == 0:
		return nil
	case info.Args == nil:
		return errors.New("[TODO decodeArgs] info.Args is nil")
	}

	// TODO: Do I need to handle possibly "stepping through a nil pointer"? (ugh)
	argsVal := structVal.FieldByIndex(info.Args.FieldIndex)

	for _, arg := range args {
		if err := decode(argsVal, arg); err != nil {
			return err
		}
	}

	return nil
}
