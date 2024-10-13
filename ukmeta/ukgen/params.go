package ukgen

import (
	"cmp"
	"errors"
	"reflect"
	"slices"

	"github.com/oligarch316/go-ukase/ukcore/ukspec"
)

// =============================================================================
// Data
// =============================================================================

type paramsData struct {
	Type      paramsTypeData
	Inlines   []paramsInlineData
	Flags     []paramsFlagData
	Arguments []paramsArgumentData
}

type paramsInlineData struct {
	FieldName  string
	FieldIndex int
	FieldType  paramsTypeData
}

type paramsFlagData struct {
	FieldName  string
	FieldIndex int
}

type paramsArgumentData struct {
	FieldName  string
	FieldIndex int
}

type paramsTypeData struct{ Source, Sink typeData }

// =============================================================================
// Load
// =============================================================================

type paramsStore map[reflect.Type]ukspec.Parameters

func newParamsStore() paramsStore { return make(paramsStore) }

func (g *Generator) loadParams(spec ukspec.Parameters) error {
	if _, exists := g.params[spec.Type]; exists {
		return nil
	}

	g.params[spec.Type] = spec

	for _, inlineSpec := range spec.Inlines {
		if err := g.loadParamsInline(inlineSpec); err != nil {
			return err
		}
	}

	return nil
}

func (g *Generator) loadParamsInline(inlineSpec ukspec.Inline) error {
	paramsSpec, err := ukspec.NewParameters(inlineSpec.FieldType, g.config.Spec...)
	if err != nil {
		return err
	}

	return g.loadParams(paramsSpec)
}

// =============================================================================
// Generate
// =============================================================================

func (g *Generator) generateParams() ([]paramsData, error) {
	var list []paramsData

	for _, spec := range g.params {
		typeData, valid, err := g.generateParamsType(spec)
		switch {
		case err != nil:
			return nil, err
		case !valid:
			continue
		}

		inlinesData, err := g.generateParamsInlines(spec)
		if err != nil {
			return nil, err
		}

		flagsData := g.generateParamsFlags(spec)
		argumentsData := g.generateParamsArguments(spec)

		item := paramsData{
			Type:      typeData,
			Inlines:   inlinesData,
			Flags:     flagsData,
			Arguments: argumentsData,
		}

		list = append(list, item)
	}

	// Sort by lexicographic (sink) type name
	compare := func(a, b paramsData) int { return cmp.Compare(a.Type.Sink.TypeName, b.Type.Sink.TypeName) }
	slices.SortFunc(list, compare)

	return list, nil
}

func (g *Generator) generateParamsType(spec ukspec.Parameters) (paramsTypeData, bool, error) {
	source, err := g.loadImport(spec.Type)
	if err != nil {
		return paramsTypeData{}, true, err
	}

	if source.TypeName == "" {
		// Parameter (source) type name is anonymous ⇒ ignore
		return paramsTypeData{}, false, nil
	}

	if source.PackageName == "main" {
		// Parameter (source) type defined in "main" package ⇒ error
		return paramsTypeData{}, true, errors.New("[TODO validateParamsType] got a type from 'main'")
	}

	sink := g.parseParamsSinkType(spec.Type, source)

	if _, reserved := g.reservedNames()[sink.TypeName]; reserved {
		// Parameter (sink) type name is reserved ⇒ error
		return paramsTypeData{}, true, errors.New("[TODO validateParamsType] reserved name conflict")
	}

	data := paramsTypeData{Source: source, Sink: sink}
	return data, true, nil
}

func (g *Generator) generateParamsInlines(spec ukspec.Parameters) ([]paramsInlineData, error) {
	var list []paramsInlineData

	for _, inlineSpec := range spec.Inlines {
		fieldIndex, valid := g.parseParamsFieldIndex(inlineSpec.FieldIndex)
		if !valid {
			continue
		}

		fieldSourceType, err := g.loadImport(inlineSpec.FieldType)
		if err != nil {
			return nil, err
		}

		fieldSinkType := g.parseParamsSinkType(inlineSpec.FieldType, fieldSourceType)

		item := paramsInlineData{
			FieldName:  inlineSpec.FieldName,
			FieldIndex: fieldIndex,
			FieldType: paramsTypeData{
				Source: fieldSourceType,
				Sink:   fieldSinkType,
			},
		}

		list = append(list, item)
	}

	// Sort by lexicographic order of field name
	compare := func(a, b paramsInlineData) int { return cmp.Compare(a.FieldName, b.FieldName) }
	slices.SortFunc(list, compare)

	return list, nil
}

func (g *Generator) generateParamsFlags(spec ukspec.Parameters) []paramsFlagData {
	var list []paramsFlagData

	for _, flagSpec := range spec.Flags {
		fieldIndex, valid := g.parseParamsFieldIndex(flagSpec.FieldIndex)
		if !valid {
			continue
		}

		item := paramsFlagData{FieldName: flagSpec.FieldName, FieldIndex: fieldIndex}
		list = append(list, item)
	}

	// Sort by lexicographic order of field name
	compare := func(a, b paramsFlagData) int { return cmp.Compare(a.FieldName, b.FieldName) }
	slices.SortFunc(list, compare)

	return list
}

func (g *Generator) generateParamsArguments(spec ukspec.Parameters) []paramsArgumentData {
	var list []paramsArgumentData

	for _, argumentSpec := range spec.Arguments {
		fieldIndex, valid := g.parseParamsFieldIndex(argumentSpec.FieldIndex)
		if !valid {
			continue
		}

		item := paramsArgumentData{FieldName: argumentSpec.FieldName, FieldIndex: fieldIndex}
		list = append(list, item)
	}

	// Sort by lexicographic order of field name
	compare := func(a, b paramsArgumentData) int { return cmp.Compare(a.FieldName, b.FieldName) }
	slices.SortFunc(list, compare)

	return list
}

// =============================================================================
// Utility
// =============================================================================

func (Generator) parseParamsFieldIndex(index []int) (int, bool) {
	if len(index) == 1 {
		return index[0], true
	}

	return 0, false
}

func (g Generator) parseParamsSinkType(sourceType reflect.Type, sourceData typeData) typeData {
	data := typeData{TypeName: sourceData.TypeName}

	if name, exists := g.config.Names.ParameterTypes[sourceType]; exists {
		data.TypeName = name
	}

	return data
}
