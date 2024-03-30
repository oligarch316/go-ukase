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

	if err := decode(fieldVal, flag.Value); err != nil {
		if errors.Is(err, errUnsupportedKind) {
			return decodeErr(structVal).field(fieldVal, flagInfo.FieldName, err)
		}

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

var errUnsupportedKind = errors.New("unsupported kind")

var kindDecoders map[reflect.Kind]func(reflect.Value, string) error

func init() {
	kindDecoders = map[reflect.Kind]func(reflect.Value, string) error{
		// Indirect
		reflect.Interface: decodeInterface,
		reflect.Pointer:   decodePointer,

		// Collection
		reflect.Slice: decodeSlice,

		// Basic
		reflect.Bool:       decodeBool,
		reflect.Int:        decodeInt,
		reflect.Int8:       decodeInt,
		reflect.Int16:      decodeInt,
		reflect.Int32:      decodeInt,
		reflect.Int64:      decodeInt,
		reflect.Uint:       decodeUint,
		reflect.Uint8:      decodeUint,
		reflect.Uint16:     decodeUint,
		reflect.Uint32:     decodeUint,
		reflect.Uint64:     decodeUint,
		reflect.Float32:    decodeFloat,
		reflect.Float64:    decodeFloat,
		reflect.Complex64:  decodeComplex,
		reflect.Complex128: decodeComplex,
		reflect.String:     decodeString,
	}
}

// TODO:
// End users should not need to actually decode a value into a field to discover
// that field's kind isn't supported. Exposing the kind map or a Validate()
// check against it seems sane. However, we don't want to import it into
// ukcore during ParamInfo construction because "sanctity of the dependency graph".

func decode(dst reflect.Value, src string) error {
	kindDecoder, ok := kindDecoders[dst.Kind()]
	if !ok {
		return fmt.Errorf("%w '%s'", errUnsupportedKind, dst.Kind())
	}

	return kindDecoder(dst, src)
}

// =============================================================================
// Kind› Indirect
// =============================================================================

func decodeInterface(dst reflect.Value, src string) error {
	// NOTE:
	// We could inspect any existing value present in `dst` and match it's kind
	// This seems footgun/surprise inducing and so we'll eschew that for now

	dst.Set(reflect.ValueOf(src))
	return nil
}

func decodePointer(dst reflect.Value, src string) error {
	elemType := dst.Type().Elem()

	val := reflect.New(elemType)
	if err := decode(val.Elem(), src); err != nil {
		return err
	}

	dst.Set(val)
	return nil
}

// =============================================================================
// Kind› Collection
// =============================================================================

// TODO:
// Reminder, in builtin there's `type byte = uint8`
// Possibly want special case behavior for Slice<uint8>

func decodeSlice(dstl reflect.Value, src string) error {
	return errors.New("[TODO decodeFlagSlice] not yet implemented")
}

// =============================================================================
// Kind› Basic
// =============================================================================

func decodeBool(dst reflect.Value, src string) error {
	boolVal, err := strconv.ParseBool(src)
	if err != nil {
		return err
	}

	dst.SetBool(boolVal)
	return nil
}

func decodeInt(dst reflect.Value, src string) error {
	intVal, err := strconv.ParseInt(src, 10, dst.Type().Bits())
	if err != nil {
		return err
	}

	dst.SetInt(intVal)
	return nil
}

func decodeUint(dst reflect.Value, src string) error {
	uintVal, err := strconv.ParseUint(src, 10, dst.Type().Bits())
	if err != nil {
		return err
	}

	dst.SetUint(uintVal)
	return nil
}

func decodeFloat(dst reflect.Value, src string) error {
	floatVal, err := strconv.ParseFloat(src, dst.Type().Bits())
	if err != nil {
		return err
	}

	dst.SetFloat(floatVal)
	return nil
}

func decodeComplex(dst reflect.Value, src string) error {
	complexVal, err := strconv.ParseComplex(src, dst.Type().Bits())
	if err != nil {
		return err
	}

	dst.SetComplex(complexVal)
	return nil
}

func decodeString(dst reflect.Value, src string) error {
	dst.SetString(src)
	return nil
}
