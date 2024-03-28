package ukenc

import (
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

func (d *Decoder) Decode(v any) error {
	val := reflect.ValueOf(v)
	switch {
	case val.Kind() != reflect.Pointer:
		return errors.New("[TODO Decode] val is non-pointer")
	case val.IsNil():
		return errors.New("[TODO Decode] val is nil pointer")
	}

	elem := val.Elem()
	if elem.Kind() != reflect.Struct {
		return errors.New("[TODO Decode] elem is non-struct")
	}

	info, err := ukcore.NewParamsInfo(elem.Type())
	if err != nil {
		return err
	}

	// TODO: Decode arguments

	return d.decodeFlags(elem, info)
}

func (d *Decoder) DecodeArgs(v any) error {
	return errors.New("[TODO DecodeArgs] not yet implemented")
}

func (d *Decoder) DecodeFlags(v any) error {
	val := reflect.ValueOf(v)
	switch {
	case val.Kind() != reflect.Pointer:
		return errors.New("[TODO DecodeFlags] val is non-pointer")
	case val.IsNil():
		return errors.New("[TODO DecodeFlags] val is nil pointer")
	}

	elem := val.Elem()
	if elem.Kind() != reflect.Struct {
		return errors.New("[TODO DecodeFlags] elem is non-struct")
	}

	info, err := ukcore.NewParamsInfo(elem.Type())
	if err != nil {
		return err
	}

	return d.decodeFlags(elem, info)
}

func (d *Decoder) decodeFlags(val reflect.Value, info ukcore.ParamsInfo) error {
	for _, flag := range d.input.Flags {
		flagInfo, ok := info.Flags[flag.Name]
		if !ok {
			return errors.New("[TODO decodeFlags] got a flag missing from struct info")
		}

		// TODO: Do I need to handle possibly "stepping through a nil pointer"? (ugh)
		fieldVal := val.FieldByIndex(flagInfo.Index)
		if err := decodeFlag(fieldVal, flag); err != nil {
			return err
		}
	}

	return nil
}

// =============================================================================
// Flag
// =============================================================================

var flagDecoders = map[reflect.Kind]func(reflect.Value, ukcore.Flag) error{
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

func decodeFlag(fieldVal reflect.Value, flag ukcore.Flag) error {
	// TODO: Handle "custom" fields

	kind := fieldVal.Kind()

	if decoder, ok := flagDecoders[kind]; ok {
		return decoder(fieldVal, flag)
	}

	return fmt.Errorf("[TODO decodeFlag] got bad field kind (%s)", kind)
}

// =============================================================================
// Flag› Custom
// =============================================================================

// TODO: Handle interface{ Set(string) error }
// TODO: Handle encoding.TextUnmarshaler

// =============================================================================
// Flag› Indirect
// =============================================================================

func decodeFlagInterface(val reflect.Value, flag ukcore.Flag) error {
	return errors.New("[TODO decodeFlagInterface] not yet implemented")
}

func decodeFlagPointer(val reflect.Value, flag ukcore.Flag) error {
	return errors.New("[TODO decodeFlagPointer] not yet implemented")
}

// =============================================================================
// Flag› Collection
// =============================================================================

// TODO:
// Reminder, in builtin there's `type byte = uint8`
// Possibly wan't special case behavior for Slice<uint8>

func decodeFlagSlice(val reflect.Value, flag ukcore.Flag) error {
	return errors.New("[TODO decodeFlagSlice] not yet implemented")
}

// =============================================================================
// Flag› Basic
// =============================================================================

func decodeFlagBool(val reflect.Value, flag ukcore.Flag) error {
	boolVal, err := strconv.ParseBool(flag.Value)
	if err != nil {
		// TODO: Better error information
		return err
	}

	val.SetBool(boolVal)
	return nil
}

func decodeFlagInt(val reflect.Value, flag ukcore.Flag) error {
	intVal, err := strconv.ParseInt(flag.Value, 10, val.Type().Bits())
	if err != nil {
		// TODO: Better error information
		return err
	}

	val.SetInt(intVal)
	return nil
}

func decodeFlagUint(val reflect.Value, flag ukcore.Flag) error {
	uintVal, err := strconv.ParseUint(flag.Value, 10, val.Type().Bits())
	if err != nil {
		// TODO: Better error information
		return err
	}

	val.SetUint(uintVal)
	return nil
}

func decodeFlagFloat(val reflect.Value, flag ukcore.Flag) error {
	floatVal, err := strconv.ParseFloat(flag.Value, val.Type().Bits())
	if err != nil {
		// TODO: Better error information
		return err
	}

	val.SetFloat(floatVal)
	return nil
}

func decodeFlagComplex(val reflect.Value, flag ukcore.Flag) error {
	complexVal, err := strconv.ParseComplex(flag.Value, val.Type().Bits())
	if err != nil {
		// TODO: Better error information
		return err
	}

	val.SetComplex(complexVal)
	return nil
}

func decodeFlagString(val reflect.Value, flag ukcore.Flag) error {
	val.SetString(flag.Value)
	return nil
}
