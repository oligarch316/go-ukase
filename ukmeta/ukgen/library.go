package ukgen

import (
	"fmt"
	"reflect"
	"slices"
	"strconv"
	"sync"

	"github.com/oligarch316/ukase/ukmeta"
	"github.com/oligarch316/ukase/ukmeta/ukhelp"
)

// =============================================================================
// Parameter Type Mapping
// =============================================================================

type ParamsMap map[reflect.Type]reflect.Type

func ParamsMapAdd[K, V any](paramsMap ParamsMap) {
	key := reflect.TypeFor[K]()
	val := reflect.TypeFor[V]()
	paramsMap[key] = val
}

func (pm ParamsMap) NewInput(in ukmeta.Input) Input { return newInput(in, pm) }

// =============================================================================
// Field Index Mapping
// =============================================================================

type fieldIndexCache map[reflect.Type]map[int]int

func (fic fieldIndexCache) load(t reflect.Type) (map[int]int, error) {
	if indexMap, ok := fic[t]; ok {
		return indexMap, nil
	}

	indexMap, err := fic.build(t)
	if err != nil {
		return nil, err
	}

	fic[t] = indexMap
	return indexMap, nil
}

func (fieldIndexCache) build(t reflect.Type) (map[int]int, error) {
	if t.Kind() != reflect.Struct {
		return nil, fmt.Errorf("[TODO fieldIndexCache.build] got an invalid kind '%s'", t.Kind())
	}

	indexMap := make(map[int]int)

	for sinkIdx := 0; sinkIdx < t.NumField(); sinkIdx++ {
		sinkField := t.Field(sinkIdx)

		sinkTag, ok := sinkField.Tag.Lookup(tagKeyIndex)
		if !ok {
			continue
		}

		sourceIdx, err := strconv.ParseInt(sinkTag, 10, 32)
		if err != nil {
			return nil, err
		}

		indexMap[int(sourceIdx)] = sinkIdx
	}

	return indexMap, nil
}

// =============================================================================
// Input
// =============================================================================

var _ Input = input{}

type Input interface {
	ukmeta.Input

	MetaInfo(index []int) (any, error)
}

type input struct {
	ukmeta.Input

	indexCache fieldIndexCache
	loadInfo   func() (reflect.Value, error)
}

func newInput(in ukmeta.Input, paramsMap ParamsMap) input {
	loadInfo := func() (reflect.Value, error) {
		sourceType := in.MetaReference().Spec.Type

		sinkType, ok := paramsMap[sourceType]
		if !ok {
			return reflect.Value{}, fmt.Errorf("[TODO loadInfo] unknown source type '%s'", sourceType)
		}

		ptrVal := reflect.New(sinkType)
		err := in.Initialize(ptrVal.Interface())
		return ptrVal.Elem(), err
	}

	return input{
		Input:      in,
		indexCache: make(fieldIndexCache),
		loadInfo:   sync.OnceValues(loadInfo),
	}
}

func (i input) MetaInfo(index []int) (any, error) {
	sinkVal, err := i.loadInfo()
	if err != nil {
		return nil, err
	}

	for _, sourceIdx := range index {
		sinkType := sinkVal.Type()

		indexMap, err := i.indexCache.load(sinkType)
		if err != nil {
			return nil, err
		}

		sinkIdx, ok := indexMap[sourceIdx]
		if !ok {
			return nil, fmt.Errorf("[TODO MetaInfo] unknown source index '%d'", sourceIdx)
		}

		sinkVal = sinkVal.Field(sinkIdx)
	}

	return sinkVal.Interface(), nil
}

// =============================================================================
// Encoder
// =============================================================================

type Encoder[T any] func(info any) (description T, err error)

func (e Encoder[T]) super() ukhelp.Encoder[T] { return ukhelp.Encoder[T](e) }

func (e Encoder[T]) Encode(in Input) (ukhelp.Output[T], error) {
	command, err := e.EncodeCommand(in)
	if err != nil {
		return ukhelp.Output[T]{}, err
	}

	subcommands, err := e.EncodeSubcommands(in)
	if err != nil {
		return ukhelp.Output[T]{}, err
	}

	flags, err := e.EncodeFlags(in)
	if err != nil {
		return ukhelp.Output[T]{}, err
	}

	arguments, err := e.EncodeArguments(in)
	if err != nil {
		return ukhelp.Output[T]{}, err
	}

	output := ukhelp.Output[T]{
		Command:     command,
		Subcommands: subcommands,
		Flags:       flags,
		Arguments:   arguments,
	}

	return output, nil
}

func (e Encoder[T]) EncodeCommand(in Input) (ukhelp.OutputCommand[T], error) {
	return e.super().EncodeCommand(in)
}

func (e Encoder[T]) EncodeSubcommands(in Input) ([]ukhelp.OutputSubcommand[T], error) {
	return e.super().EncodeSubcommands(in)
}

func (e Encoder[T]) EncodeFlags(in Input) ([]ukhelp.OutputFlag[T], error) {
	var list []ukhelp.OutputFlag[T]

	super := e.super()

	for _, spec := range in.MetaReference().Spec.Flags {
		info, err := in.MetaInfo(spec.FieldIndex)
		if err != nil {
			return nil, err
		}

		description, err := e(info)
		if err != nil {
			return nil, err
		}

		names := slices.Clone(spec.Names)
		super.SortFlagNames(names)

		item := ukhelp.OutputFlag[T]{Description: description, Names: names}
		list = append(list, item)
	}

	super.SortFlags(list)
	return list, nil
}

func (e Encoder[T]) EncodeArguments(in Input) ([]ukhelp.OutputArgument[T], error) {
	var list []ukhelp.OutputArgument[T]

	super := e.super()

	for _, spec := range in.MetaReference().Spec.Arguments {
		info, err := in.MetaInfo(spec.FieldIndex)
		if err != nil {
			return nil, err
		}

		description, err := e(info)
		if err != nil {
			return nil, err
		}

		item := ukhelp.OutputArgument[T]{Description: description, Position: spec.Position}
		list = append(list, item)
	}

	super.SortArguments(list)
	return list, nil
}
