package ukmeta

import (
	"reflect"
	"sync"

	"github.com/oligarch316/go-ukase/ukcli"
	"github.com/oligarch316/go-ukase/ukcore/ukexec"
)

var _ Input = input{}

type Input interface {
	ukcli.Input

	MetaReference() Reference
	MetaDefault(index []int) (any, error)
}

type Reference struct {
	ukexec.Meta
	Target []string
}

type input struct {
	ukcli.Input

	loadDefaults func() (reflect.Value, error)
	reference    Reference
}

func NewInput(in ukcli.Input, refTarget ...string) (Input, error) {
	refMeta, err := in.Lookup(refTarget...)
	if err != nil {
		return nil, err
	}

	loadDefaults := func() (reflect.Value, error) {
		ptrVal := reflect.New(refMeta.Spec.Type)
		err := in.Initialize(ptrVal.Interface())
		return ptrVal.Elem(), err
	}

	input := input{
		Input:        in,
		loadDefaults: sync.OnceValues(loadDefaults),
		reference:    Reference{Meta: refMeta, Target: refTarget},
	}

	return input, nil
}

func (i input) MetaReference() Reference { return i.reference }

func (i input) MetaDefault(index []int) (any, error) {
	defaultsVal, err := i.loadDefaults()
	if err != nil {
		return nil, err
	}

	fieldVal, err := defaultsVal.FieldByIndexErr(index)
	if err != nil {
		return nil, err
	}

	return fieldVal.Interface(), nil
}
