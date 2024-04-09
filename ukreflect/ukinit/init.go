package ukinit

import (
	"errors"
	"reflect"

	"github.com/oligarch316/go-ukase/ukreflect"
	"github.com/oligarch316/go-ukase/ukspec"
)

type Rule interface{ apply(any) error }

type ruleList[T any] []func(*T)

func (rl ruleList[T]) apply(v any) error {
	if target, ok := v.(*T); ok {
		for _, f := range rl {
			f(target)
		}
		return nil
	}

	return errors.New("[TODO apply] <INTERNAL> v is not the correct type")
}

type RuleSet map[reflect.Type]Rule

func New() RuleSet { return make(RuleSet) }

func Register[T any](set RuleSet, fs ...func(*T)) {
	t := reflect.TypeFor[T]()

	if x, exists := set[t]; exists {
		fs = append(x.(ruleList[T]), fs...)
	}

	set[t] = ruleList[T](fs)
}

func Create[T any](set RuleSet) (T, error) {
	var target T

	spec, err := ukspec.Of(target)
	if err != nil {
		return target, err
	}

	// TODO: This is kinda ugly
	seed := ukspec.Inline{
		Type:    spec.Type,
		Inlines: spec.Inlines,
	}

	if err := set.process(&target, seed, false); err != nil {
		return target, err
	}

	return target, nil
}

func (rs RuleSet) process(v any, spec ukspec.Inline, customComplete bool) error {
	// TODO: Documentation comment
	customComplete = customComplete || rs.processCustom(v)

	// TODO: Documentation comment
	val, err := ukreflect.LoadValueOf(v)
	if err != nil {
		return err
	}

	for _, fieldSpec := range spec.Inlines {
		fieldVal := ukreflect.LoadFieldByIndex(val, fieldSpec.FieldIndex)

		if fieldVal.Kind() != reflect.Pointer {
			if !fieldVal.CanAddr() {
				return errors.New("[TODO process] <INTERNAL> CanAddr() is false for inline field")
			}

			fieldVal = fieldVal.Addr()
		}

		if err := rs.process(fieldVal.Interface(), fieldSpec, customComplete); err != nil {
			return err
		}
	}

	// TODO: Documentation comment
	return rs.processRule(v, spec)
}

func (rs RuleSet) processCustom(v any) bool {
	type Custom interface{ UkaseInit() }

	if custom, ok := v.(Custom); ok {
		custom.UkaseInit()
		return true
	}

	return false
}

func (rs RuleSet) processRule(v any, spec ukspec.Inline) error {
	if rule, ok := rs[spec.Type]; ok {
		return rule.apply(v)
	}

	return nil
}
