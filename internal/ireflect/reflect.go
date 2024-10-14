package ireflect

import (
	"reflect"

	"github.com/oligarch316/go-ukase/internal/ierror"
)

type ParametersValue struct{ reflect.Value }

func NewParametersValue(v any) (ParametersValue, error) {
	val := reflect.ValueOf(v)
	switch {
	case val.Kind() != reflect.Pointer:
		return ParametersValue{}, ierror.NewD("value is not a pointer")
	case val.IsNil():
		return ParametersValue{}, ierror.NewD("value is a nil pointer")
	}

	elem := val.Elem()
	if elem.Kind() != reflect.Struct {
		return ParametersValue{}, ierror.NewD("value does not point to a struct")
	}

	return ParametersValue{Value: elem}, nil
}

func (pv ParametersValue) EnsureFieldByIndex(index []int) reflect.Value {
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
