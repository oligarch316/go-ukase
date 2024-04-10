package ukinit

import (
	"errors"
	"reflect"

	"github.com/oligarch316/go-ukase/ukreflect"
	"github.com/oligarch316/go-ukase/ukspec"
)

type rule interface{ apply(any) error }

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

type RuleSet struct {
	config Config
	rules  map[reflect.Type]rule
}

func New(opts ...Option) *RuleSet {
	return &RuleSet{
		config: newConfig(opts),
		rules:  make(map[reflect.Type]rule),
	}
}

func Register[T any](set *RuleSet, fs ...func(*T)) {
	t := reflect.TypeFor[T]()

	if x, exists := set.rules[t]; exists {
		fs = append(x.(ruleList[T]), fs...)
	}

	set.rules[t] = ruleList[T](fs)
}

func Create[T any](set *RuleSet) (T, error) {
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
		fieldVal, err := rs.loadField(val, fieldSpec.FieldIndex)
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
		fieldVal, err := rs.loadField(val, fieldSpec.FieldIndex)
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
	if rule, ok := rs.rules[spec.Type]; ok {
		return rule.apply(v)
	}

	return nil
}

func (*RuleSet) loadField(val reflect.Value, index []int) (reflect.Value, error) {
	fieldVal := ukreflect.LoadFieldByIndex(val, index)

	if fieldVal.Kind() != reflect.Pointer {
		if !fieldVal.CanAddr() {
			return fieldVal, errors.New("[TODO loadField] <INTERNAL> CanAddr() is false for inline field")
		}

		fieldVal = fieldVal.Addr()
	}

	return fieldVal, nil
}
