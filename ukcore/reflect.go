package ukcore

import (
	"errors"
	"reflect"
)

type ParamsValue struct{ reflect.Value }

func NewParamsValue(v any) (ParamsValue, error) {
	val := reflect.ValueOf(v)
	switch {
	case val.Kind() != reflect.Pointer:
		return ParamsValue{}, errors.New("[TODO NewParamsValue] v is not a pointer")
	case val.IsNil():
		return ParamsValue{}, errors.New("[TODO NewParamsValue] v is a nil pointer")
	}

	elem := val.Elem()
	if elem.Kind() != reflect.Struct {
		return ParamsValue{}, errors.New("[TODO NewParamsValue] v is not a struct pointer")
	}

	return ParamsValue{Value: elem}, nil
}

func (pv ParamsValue) EnsureFieldByIndex(index []int) reflect.Value {
	val := pv.Value

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
