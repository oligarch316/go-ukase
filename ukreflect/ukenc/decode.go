package ukenc

import (
	"encoding"
	"errors"
	"fmt"
	"reflect"
	"strconv"
)

// TODO:
// End users should not need to actually decode a value into a field to discover
// that field's kind isn't supported. Exposing the kind map or a Validate()
// check against it seems sane. However, we don't want to import it into
// ukcore during ParamInfo construction because "sanctity of the dependency graph".

// TODO: Make this more general (invalid field) to support the interface TODO error
var errUnsupportedKind = errors.New("unsupported kind")

func decode(dst reflect.Value, src string) error {
	if complete, err := decodeIndirect(dst, src); complete {
		return err
	}

	if complete, err := decodeCustom(dst, src); complete {
		return err
	}

	if complete, err := decodeDirect(dst, src); complete {
		return err
	}

	return fmt.Errorf("%w '%s'", errUnsupportedKind, dst.Kind())
}

// =============================================================================
// Indirect
// =============================================================================

func decodeIndirect(dst reflect.Value, src string) (bool, error) {
	switch dst.Kind() {
	case reflect.Interface:
		return true, decodeInterface(dst, src)
	case reflect.Pointer:
		return true, decodePointer(dst, src)
	default:
		return false, nil
	}
}

func decodeInterface(dst reflect.Value, src string) error {
	// Interface already contains a concrete type+value
	// ⇒ Copy that type+value to attain "settability"
	// ⇒ Decode into this copy and set `dst` on success
	if !dst.IsZero() {
		elemOld := dst.Elem()

		elemNew := reflect.New(elemOld.Type()).Elem()
		elemNew.Set(elemOld)

		if err := decode(elemNew, src); err != nil {
			return err
		}

		dst.Set(elemNew)
		return nil
	}

	// Interface contains no concrete type+value
	// ⇒ Choose the simple string type as a sane default
	// ⇒ Confirm assignability and set `dst` on success
	srcVal := reflect.ValueOf(src)
	if srcVal.Type().AssignableTo(dst.Type()) {
		dst.Set(srcVal)
		return nil
	}

	// Interface won't accept a simple string
	// ⇒ A reference value is required in this case, so fail
	return errors.New("[TODO decodeInterface] reference value is required here")
}

func decodePointer(dst reflect.Value, src string) error {
	if !dst.IsZero() {
		// Why bother with this?
		// ⇒ See tests 'DecodeBaroque/pointer->interface->…'
		return decode(dst.Elem(), src)
	}

	elemType := dst.Type().Elem()

	val := reflect.New(elemType)
	if err := decode(val.Elem(), src); err != nil {
		return err
	}

	dst.Set(val)
	return nil
}

// =============================================================================
// Custom
// =============================================================================

var typeTextUnmarshaler = reflect.TypeOf((*encoding.TextUnmarshaler)(nil)).Elem()

func decodeCustom(dst reflect.Value, src string) (bool, error) {
	if unmarshaler, ok := loadTextUnmarshaler(dst); ok {
		return true, unmarshaler.UnmarshalText([]byte(src))
	}

	return false, nil
}

func loadTextUnmarshaler(val reflect.Value) (encoding.TextUnmarshaler, bool) {
	typ := val.Type()

	if typ.Implements(typeTextUnmarshaler) {
		unmarshaler, ok := val.Interface().(encoding.TextUnmarshaler)
		return unmarshaler, ok
	}

	if reflect.PointerTo(typ).Implements(typeTextUnmarshaler) {
		if !val.CanAddr() {
			// This should never be the case as `val` is expected to be a field
			// of an addressable (provided via pointer) struct.
			// Still, safety first
			return nil, false
		}

		unmarshaler, ok := val.Addr().Interface().(encoding.TextUnmarshaler)
		return unmarshaler, ok
	}

	return nil, false
}

// =============================================================================
// Direct
// =============================================================================

var basicDecoders = map[reflect.Kind]func(reflect.Value, string) error{
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

func decodeDirect(dst reflect.Value, src string) (bool, error) {
	kind := dst.Kind()

	if kind == reflect.Slice {
		return true, decodeSlice(dst, src)
	}

	if decodeBasic, ok := basicDecoders[kind]; ok {
		return true, decodeBasic(dst, src)
	}

	return false, nil
}

func decodeSlice(dst reflect.Value, src string) error {
	elemType := dst.Type().Elem()
	elemVal := reflect.New(elemType).Elem()

	if err := decode(elemVal, src); err != nil {
		return err
	}

	val := reflect.Append(dst, elemVal)
	dst.Set(val)
	return nil
}

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
