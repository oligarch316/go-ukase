package ukreflect

import (
	"errors"
	"reflect"
)

// LoadValueOf TODO: Document
func LoadValueOf(v any) (reflect.Value, error) {
	val := reflect.ValueOf(v)
	switch {
	case val.Kind() != reflect.Pointer:
		return val, errors.New("[TODO LoadValueOf] destination is not a pointer")
	case val.IsNil():
		return val, errors.New("[TODO LoadValueOf] destination is a nil pointer")
	}

	elem := val.Elem()
	if elem.Kind() != reflect.Struct {
		return elem, errors.New("[TODO LoadValueOf] destination does not point to a struct")
	}

	return elem, nil
}

// LoadFieldByIndex TODO: Document
func LoadFieldByIndex(val reflect.Value, index []int) reflect.Value {
	for _, i := range index {
		if val.Kind() == reflect.Pointer {
			if val.IsZero() {
				newVal := reflect.New(val.Type().Elem())
				val.Set(newVal)
			}

			val = val.Elem()
		}

		val = val.Field(i)
	}

	return val
}
