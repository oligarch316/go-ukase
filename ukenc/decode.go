package ukenc

import (
	"encoding"
	"errors"
	"fmt"
	"reflect"
	"strconv"

	"github.com/oligarch316/go-ukase/ukcore"
)

// TODO:
// I _maybe_ want to share some cache from reflect.Type -> ukcore.ParamsInfo

type Decoder struct{ input ukcore.Input }

func NewDecoder(input ukcore.Input) *Decoder { return &Decoder{input: input} }

func (d *Decoder) Decode(params any) error {
	val := reflect.ValueOf(params)
	switch {
	case val.Kind() != reflect.Pointer:
		return decodeErr(val).params("destination is not a pointer")
	case val.IsNil():
		return decodeErr(val).params("destination is a nil pointer")
	}

	elem := val.Elem()
	if elem.Kind() != reflect.Struct {
		return decodeErr(elem).params("destination does not point to a struct")
	}

	info, err := ukcore.NewParamsInfo(elem.Type())
	if err != nil {
		// TODO: Wrap this appropriately
		return err
	}

	for _, flag := range d.input.Flags {
		if err := d.decodeFlag(elem, info, flag); err != nil {
			return err
		}
	}

	return d.decodeArgs(elem, info)
}

func (d *Decoder) decodeFlag(structVal reflect.Value, info ukcore.ParamsInfo, flag ukcore.Flag) error {
	flagInfo, ok := info.Flags[flag.Name]
	if !ok {
		return decodeErr(structVal).flagName(flag)
	}

	// TODO: Do I need to handle possibly "stepping through a nil pointer"? (ugh)
	fieldVal := structVal.FieldByIndex(flagInfo.FieldIndex)

	if unmarshaler, ok := loadTextUnmarshaler(fieldVal); ok {
		text := []byte(flag.Value)

		if err := unmarshaler.UnmarshalText(text); err != nil {
			return decodeErr(structVal).flagValue(flag, err)
		}

		return nil
	}

	kindDecoder, ok := kindDecoders[fieldVal.Kind()]
	if !ok {
		message := fmt.Sprintf("unsupported kind '%s'", fieldVal.Kind())
		return decodeErr(structVal).field(fieldVal, flagInfo.FieldName, message)
	}

	if err := kindDecoder(fieldVal, flag); err != nil {
		return decodeErr(structVal).flagValue(flag, err)
	}

	return nil
}

func (d *Decoder) decodeArgs(structVal reflect.Value, info ukcore.ParamsInfo) error {
	// NOTE: Don't forget that (as it stands) input.Args may be nil!

	// TODO
	return nil
}

// =============================================================================
// Interface
// =============================================================================

var typeTextUnmarshaler = reflect.TypeOf((*encoding.TextUnmarshaler)(nil)).Elem()

func loadTextUnmarshaler(fieldVal reflect.Value) (encoding.TextUnmarshaler, bool) {
	fieldType := fieldVal.Type()

	if fieldType.Implements(typeTextUnmarshaler) {
		unmarshaler, ok := fieldVal.Interface().(encoding.TextUnmarshaler)
		return unmarshaler, ok
	}

	if reflect.PointerTo(fieldType).Implements(typeTextUnmarshaler) {
		if !fieldVal.CanAddr() {
			// This should never be the case as `fieldVal` is expected to be a
			// field of an addressable (provided via pointer) struct.
			// Still, safety first
			return nil, false
		}

		unmarshaler, ok := fieldVal.Addr().Interface().(encoding.TextUnmarshaler)
		return unmarshaler, ok
	}

	return nil, false
}

// =============================================================================
// Kind
// =============================================================================

var kindDecoders = map[reflect.Kind]func(reflect.Value, ukcore.Flag) error{
	// Indirect
	reflect.Interface: decodeFlagInterface,
	reflect.Pointer:   decodeFlagPointer,

	// Collection
	reflect.Slice: decodeFlagSlice,

	// Basic
	reflect.Bool:       decodeFlagBool,
	reflect.Int:        decodeFlagInt,
	reflect.Int8:       decodeFlagInt,
	reflect.Int16:      decodeFlagInt,
	reflect.Int32:      decodeFlagInt,
	reflect.Int64:      decodeFlagInt,
	reflect.Uint:       decodeFlagUint,
	reflect.Uint8:      decodeFlagUint,
	reflect.Uint16:     decodeFlagUint,
	reflect.Uint32:     decodeFlagUint,
	reflect.Uint64:     decodeFlagUint,
	reflect.Float32:    decodeFlagFloat,
	reflect.Float64:    decodeFlagFloat,
	reflect.Complex64:  decodeFlagComplex,
	reflect.Complex128: decodeFlagComplex,
	reflect.String:     decodeFlagString,
}

// =============================================================================
// Kind› Indirect
// =============================================================================

func decodeFlagInterface(val reflect.Value, flag ukcore.Flag) error {
	return errors.New("[TODO decodeFlagInterface] not yet implemented")
}

func decodeFlagPointer(val reflect.Value, flag ukcore.Flag) error {
	return errors.New("[TODO decodeFlagPointer] not yet implemented")
}

// =============================================================================
// Kind› Collection
// =============================================================================

// TODO:
// Reminder, in builtin there's `type byte = uint8`
// Possibly want special case behavior for Slice<uint8>

func decodeFlagSlice(val reflect.Value, flag ukcore.Flag) error {
	return errors.New("[TODO decodeFlagSlice] not yet implemented")
}

// =============================================================================
// Kind› Basic
// =============================================================================

func decodeFlagBool(val reflect.Value, flag ukcore.Flag) error {
	boolVal, err := strconv.ParseBool(flag.Value)
	if err != nil {
		return err
	}

	val.SetBool(boolVal)
	return nil
}

func decodeFlagInt(val reflect.Value, flag ukcore.Flag) error {
	intVal, err := strconv.ParseInt(flag.Value, 10, val.Type().Bits())
	if err != nil {
		return err
	}

	val.SetInt(intVal)
	return nil
}

func decodeFlagUint(val reflect.Value, flag ukcore.Flag) error {
	uintVal, err := strconv.ParseUint(flag.Value, 10, val.Type().Bits())
	if err != nil {
		return err
	}

	val.SetUint(uintVal)
	return nil
}

func decodeFlagFloat(val reflect.Value, flag ukcore.Flag) error {
	floatVal, err := strconv.ParseFloat(flag.Value, val.Type().Bits())
	if err != nil {
		return err
	}

	val.SetFloat(floatVal)
	return nil
}

func decodeFlagComplex(val reflect.Value, flag ukcore.Flag) error {
	complexVal, err := strconv.ParseComplex(flag.Value, val.Type().Bits())
	if err != nil {
		return err
	}

	val.SetComplex(complexVal)
	return nil
}

func decodeFlagString(val reflect.Value, flag ukcore.Flag) error {
	val.SetString(flag.Value)
	return nil
}
