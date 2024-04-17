package ukinit

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/oligarch316/go-ukase/ukreflect"
	"github.com/oligarch316/go-ukase/ukspec"
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

func For[T any](set *RuleSet) (T, error) {
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

	if err := set.process(&target, seed); err != nil {
		return target, err
	}

	return target, nil
}

func (rs *RuleSet) process(v any, spec ukspec.Inline) error {
	if rs.config.ForceCustomInit {
		return rs.processForced(v, spec)
	}

	return rs.processUnforced(v, spec, false)
}

func (rs *RuleSet) processForced(v any, spec ukspec.Inline) error {
	val, err := ukreflect.LoadValueOf(v)
	if err != nil {
		return err
	}

	// TODO: Documentation comment
	for _, fieldSpec := range spec.Inlines {
		fieldVal, err := rs.loadInline(val, fieldSpec.FieldIndex)
		if err != nil {
			return err
		}

		err = rs.processForced(fieldVal.Interface(), fieldSpec)
		if err != nil {
			return err
		}
	}

	// TODO: Documentation comment
	rs.processCustom(v)

	// TODO: Documentation comment
	return rs.processRule(v, spec)
}

func (rs *RuleSet) processUnforced(v any, spec ukspec.Inline, customComplete bool) error {
	// TODO: Documentation comment
	customComplete = customComplete || rs.processCustom(v)

	val, err := ukreflect.LoadValueOf(v)
	if err != nil {
		return err
	}

	// TODO: Documentation comment
	for _, fieldSpec := range spec.Inlines {
		fieldVal, err := rs.loadInline(val, fieldSpec.FieldIndex)
		if err != nil {
			return err
		}

		err = rs.processUnforced(fieldVal.Interface(), fieldSpec, customComplete)
		if err != nil {
			return err
		}
	}

	// TODO: Documentation comment
	return rs.processRule(v, spec)
}

func (*RuleSet) processCustom(v any) bool {
	type Custom interface{ UkaseInit() }

	if custom, ok := v.(Custom); ok {
		custom.UkaseInit()
		return true
	}

	return false
}

func (rs *RuleSet) processRule(v any, spec ukspec.Inline) error {
	rules, ok := rs.rules[spec.Type]
	if !ok {
		return nil
	}

	for _, rule := range rules {
		if err := rule.apply(v); err != nil {
			return err
		}
	}

	return nil
}

func (*RuleSet) loadInline(val reflect.Value, index []int) (reflect.Value, error) {
	// TODO: Document
	index = index[len(index)-1:]

	fieldVal := ukreflect.LoadFieldByIndex(val, index)

	// TODO: Document
	if fieldVal.Kind() != reflect.Pointer {
		if !fieldVal.CanAddr() {
			return fieldVal, errors.New("[TODO loadField] <INTERNAL> CanAddr() is false for inline field")
		}

		fieldVal = fieldVal.Addr()
	}

	// TODO: Document
	if fieldVal.IsZero() {
		elemType := fieldVal.Type().Elem()
		fieldVal.Set(reflect.New(elemType))
	}

	return fieldVal, nil
}
