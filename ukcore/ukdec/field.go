package ukdec

import (
	"encoding"
	"reflect"
	"strconv"

	"github.com/oligarch316/go-ukase/internal/ierror"
)

// =============================================================================
// Field
// › Dispatch entrypoint
// › Indirect and container decode logic recurses back here
// =============================================================================

func decodeField(dst reflect.Value, src string) error {
	if complete, err := decodeFieldIndirect(dst, src); complete {
		return err
	}

	if complete, err := decodeFieldCustom(dst, src); complete {
		return err
	}

	if complete, err := decodeFieldDirect(dst, src); complete {
		return err
	}

	// TODO:
	// We should error on unsupported field kinds during spec creation
	// When/if that happens, this can become an internal error
	return ierror.FmtD("unsupported destination kind '%s'", dst.Kind())
}

// =============================================================================
// Indirect Field
// › Handles interfaces and pointers
// =============================================================================

func decodeFieldIndirect(dst reflect.Value, src string) (bool, error) {
	switch dst.Kind() {
	case reflect.Interface:
		return true, decodeFieldInterface(dst, src)
	case reflect.Pointer:
		return true, decodeFieldPointer(dst, src)
	default:
		return false, nil
	}
}

func decodeFieldInterface(dst reflect.Value, src string) error {
	// Interface already contains a concrete type+value
	// ⇒ Copy that type+value to attain "settability"
	// ⇒ Decode into this copy and set `dst` on success
	if !dst.IsZero() {
		elemOld := dst.Elem()

		elemNew := reflect.New(elemOld.Type()).Elem()
		elemNew.Set(elemOld)

		if err := decodeField(elemNew, src); err != nil {
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
	return ierror.NewD("interface destination neither contains a non-zero value nor is string-assignable")
}

func decodeFieldPointer(dst reflect.Value, src string) error {
	if !dst.IsZero() {
		// Why bother with this?
		// ⇒ See tests 'DecodeBaroque/pointer->interface->…'
		return decodeField(dst.Elem(), src)
	}

	elemType := dst.Type().Elem()

	val := reflect.New(elemType)
	if err := decodeField(val.Elem(), src); err != nil {
		return err
	}

	dst.Set(val)
	return nil
}

// =============================================================================
// Custom Field
// › Handles encoding.TextUnmarshaler implementations
// =============================================================================

var typeTextUnmarshaler = reflect.TypeFor[encoding.TextUnmarshaler]()

func decodeFieldCustom(dst reflect.Value, src string) (bool, error) {
	unmarshaler, ok := loadTextUnmarshaler(dst)
	if !ok {
		return false, nil
	}

	if err := unmarshaler.UnmarshalText([]byte(src)); err != nil {
		return true, ierror.U(err)
	}

	return true, nil
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
// Direct Field
// › Handles slices
// › Handles "basic" types (bool, numeric, string)
//
// TODO:
// Move slices into their own "Container Type" section?
// Esp. if support for arrays (or even structs) is added
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

func decodeFieldDirect(dst reflect.Value, src string) (bool, error) {
	kind := dst.Kind()

	if kind == reflect.Slice {
		return true, decodeSlice(dst, src)
	}

	if decodeBasic, ok := basicDecoders[kind]; ok {
		// TODO:
		// The "user" errors coming from here are not "pretty" enough for actual end-users
		// example > strconv.ParseBool: parsing "ipsum": invalid syntax
		return true, decodeBasic(dst, src)
	}

	return false, nil
}

func decodeSlice(dst reflect.Value, src string) error {
	elemType := dst.Type().Elem()
	elemVal := reflect.New(elemType).Elem()

	if err := decodeField(elemVal, src); err != nil {
		return err
	}

	val := reflect.Append(dst, elemVal)
	dst.Set(val)
	return nil
}

func decodeBool(dst reflect.Value, src string) error {
	boolVal, err := strconv.ParseBool(src)
	if err != nil {
		return ierror.U(err)
	}

	dst.SetBool(boolVal)
	return nil
}

func decodeInt(dst reflect.Value, src string) error {
	intVal, err := strconv.ParseInt(src, 10, dst.Type().Bits())
	if err != nil {
		return ierror.U(err)
	}

	dst.SetInt(intVal)
	return nil
}

func decodeUint(dst reflect.Value, src string) error {
	uintVal, err := strconv.ParseUint(src, 10, dst.Type().Bits())
	if err != nil {
		return ierror.U(err)
	}

	dst.SetUint(uintVal)
	return nil
}

func decodeFloat(dst reflect.Value, src string) error {
	floatVal, err := strconv.ParseFloat(src, dst.Type().Bits())
	if err != nil {
		return ierror.U(err)
	}

	dst.SetFloat(floatVal)
	return nil
}

func decodeComplex(dst reflect.Value, src string) error {
	complexVal, err := strconv.ParseComplex(src, dst.Type().Bits())
	if err != nil {
		return ierror.U(err)
	}

	dst.SetComplex(complexVal)
	return nil
}

func decodeString(dst reflect.Value, src string) error {
	dst.SetString(src)
	return nil
}
