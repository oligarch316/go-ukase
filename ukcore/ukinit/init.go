package ukinit

import (
	"errors"
	"fmt"
	"reflect"
	"slices"

	"github.com/oligarch316/ukase/internal/ireflect"
	"github.com/oligarch316/ukase/ukcore/ukspec"
)

type Rule interface {
	Register(*RuleSet)
	apply(any) error
}

func NewRule[T any](f func(*T)) Rule { return rule[T](f) }

type rule[T any] func(*T)

func (r rule[T]) Register(ruleSet *RuleSet) {
	t := reflect.TypeFor[T]()
	ruleSet.rules[t] = append(ruleSet.rules[t], r)
}

func (r rule[T]) apply(v any) error {
	if vt, ok := v.(*T); ok {
		r(vt)
		return nil
	}

	return fmt.Errorf("[TODO apply] <INTERNAL> v is not the correct type, expected: %T, actual: %T", new(T), v)
}

type custom interface{ UkaseInit() }

var typeCustom = reflect.TypeFor[custom]()

type RuleSet struct {
	config Config
	rules  map[reflect.Type][]Rule
}

func NewRuleSet(opts ...Option) *RuleSet {
	return &RuleSet{
		config: newConfig(opts),
		rules:  make(map[reflect.Type][]Rule),
	}
}

func (rs *RuleSet) Process(spec ukspec.Parameters, v any) error {
	paramsVal, err := ireflect.NewParametersValue(v)
	if err != nil {
		return err
	}

	inlines := slices.Clone(spec.Inlines)
	slices.SortStableFunc(inlines, rs.orderInline)

	for _, inlineSpec := range inlines {
		inlineVal, err := rs.loadInline(paramsVal, inlineSpec.FieldIndex)
		if err != nil {
			return err
		}

		if err := rs.processValue(inlineVal); err != nil {
			return err
		}
	}

	return rs.processValue(paramsVal.Addr())
}

func (rs RuleSet) processValue(val reflect.Value) error {
	valType := val.Type()

	customTrigger := valType.Implements(typeCustom)
	rules, rulesTrigger := rs.rules[valType.Elem()]

	if !customTrigger && !rulesTrigger {
		return nil
	}

	v := val.Interface()

	if customTrigger {
		v.(custom).UkaseInit()
	}

	if !rulesTrigger {
		return nil
	}

	for _, rule := range rules {
		if err := rule.apply(v); err != nil {
			return err
		}
	}

	return nil
}

func (RuleSet) loadInline(paramsVal ireflect.ParametersValue, index []int) (reflect.Value, error) {
	// Load the relevant field. Intermediate fields constructed automatically.
	inlineVal := paramsVal.EnsureFieldByIndex(index)

	// Ensure the result is a pointer, using `.Addr()` if necessary
	if inlineVal.Kind() != reflect.Pointer {
		if !inlineVal.CanAddr() {
			return inlineVal, errors.New("[TODO loadInline] <INTERNAL> CanAddr() is false for inline field")
		}

		inlineVal = inlineVal.Addr()
	}

	// Ensure the (pointer) result is non-zero
	if inlineVal.IsZero() {
		elemType := inlineVal.Type().Elem()
		inlineVal.Set(reflect.New(elemType))
	}

	return inlineVal, nil
}

func (RuleSet) orderInline(a, b ukspec.Inline) int {
	// TODO: Document bottom-up (DFS-esq) intention

	switch tierA, tierB := len(a.FieldIndex), len(b.FieldIndex); {
	case tierA > tierB:
		return -1
	case tierA < tierB:
		return 1
	default:
		return 0
	}
}
