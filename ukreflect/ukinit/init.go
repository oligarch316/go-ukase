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

func For[T any](ruleSet *RuleSet) (T, error) {
	var data T

	spec, err := ukspec.Of(data)
	if err != nil {
		return data, err
	}

	return data, ruleSet.Process(spec, &data)
}

func (rs *RuleSet) Process(spec ukspec.Params, v any) error {
	// TODO: This is kinda ugly
	seed := ukspec.Inline{
		Type:    spec.Type,
		Inlines: spec.Inlines,
	}

	return rs.process(seed, v)
}

func (rs *RuleSet) process(spec ukspec.Inline, v any) error {
	if rs.config.ForceCustomInit {
		return rs.processForced(spec, v)
	}

	return rs.processUnforced(spec, v, false)
}

func (rs *RuleSet) processForced(spec ukspec.Inline, v any) error {
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

		err = rs.processForced(fieldSpec, fieldVal.Interface())
		if err != nil {
			return err
		}
	}

	// TODO: Documentation comment
	rs.processCustom(v)

	// TODO: Documentation comment
	return rs.processRule(spec, v)
}

func (rs *RuleSet) processUnforced(spec ukspec.Inline, v any, customComplete bool) error {
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

		err = rs.processUnforced(fieldSpec, fieldVal.Interface(), customComplete)
		if err != nil {
			return err
		}
	}

	// TODO: Documentation comment
	return rs.processRule(spec, v)
}

func (*RuleSet) processCustom(v any) bool {
	type Custom interface{ UkaseInit() }

	if custom, ok := v.(Custom); ok {
		custom.UkaseInit()
		return true
	}

	return false
}

func (rs *RuleSet) processRule(spec ukspec.Inline, v any) error {
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
