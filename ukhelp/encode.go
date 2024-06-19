package ukhelp

import (
	"cmp"
	"reflect"
	"slices"
	"sync"

	"github.com/oligarch316/go-ukase"
	"github.com/oligarch316/go-ukase/ukcore"
	"github.com/oligarch316/go-ukase/ukspec"
)

// =============================================================================
// Input
// =============================================================================

type Input struct {
	ukase.Input
	Meta      ukcore.Meta
	Reference ukcore.InputTarget

	loadDefaults func() (reflect.Value, error)
}

func newInput(ref []string, in ukase.Input) (Input, error) {
	reference := append(ref, in.Args...)

	meta, err := in.Meta(reference)
	if err != nil {
		return Input{}, err
	}

	loadDefaults := func() (reflect.Value, error) {
		ptrVal := reflect.New(meta.Spec.Type)
		err := in.Initialize(ptrVal.Interface())
		return ptrVal.Elem(), err
	}

	input := Input{
		Input:        in,
		Meta:         meta,
		Reference:    reference,
		loadDefaults: sync.OnceValues(loadDefaults),
	}

	return input, nil
}

func (i Input) DefaultByIndex(index []int) (any, error) {
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

// =============================================================================
// Output
// =============================================================================

type Output struct {
	Command     OutputCommand
	Subcommands []OutputSubcommand
	Arguments   []OutputArgument
	Flags       []OutputFlag
}

func Encode(input Input) Output {
	return Output{
		Command:     EncodeCommand(input),
		Subcommands: EncodeSubcommands(input),
		Arguments:   EncodeArguments(input),
		Flags:       EncodeFlags(input),
	}
}

type OutputCommand struct {
	Description OutputDescription
	Program     string
	Target      []string
	Exec        bool
}

func EncodeCommand(input Input) OutputCommand {
	return OutputCommand{
		Description: EncodeDescription(input.Meta.Info),
		Program:     input.Program,
		Target:      input.Reference,
		Exec:        input.Meta.Exec,
	}
}

type OutputSubcommand struct {
	Description OutputDescription
	Name        string
}

func EncodeSubcommands(input Input) []OutputSubcommand {
	var list []OutputSubcommand

	for name, meta := range input.Meta.Children() {
		description := EncodeDescription(meta.Info)
		item := OutputSubcommand{Description: description, Name: name}
		list = append(list, item)
	}

	// Sort by lexicographic order of subcommand name
	compare := func(a, b OutputSubcommand) int { return cmp.Compare(a.Name, b.Name) }
	slices.SortFunc(list, compare)

	return list
}

type OutputArgument struct {
	Description OutputDescription
	Position    OutputArgumentPosition
}

func EncodeArguments(input Input) []OutputArgument {
	var list []OutputArgument

	// TODO
	var argSpecs []ukspec.Args
	if input.Meta.Spec.Args != nil {
		argSpecs = []ukspec.Args{*input.Meta.Spec.Args}
	}

	for range argSpecs {
		// TODO
		position := OutputArgumentPosition{Start: -1, End: -1}
		item := OutputArgument{Position: position}
		list = append(list, item)
	}

	return list
}

type OutputArgumentPosition struct{ Start, End int }

type OutputFlag struct {
	Description OutputDescription
	Names       []string
}

func EncodeFlags(input Input) []OutputFlag {
	var list []OutputFlag

	for _, spec := range input.Meta.Spec.Flags {
		names := slices.Clone(spec.FlagNames)

		// Sort by length of flag name
		compare := func(a, b string) int { return len(a) - len(b) }
		slices.SortFunc(names, compare)

		item := OutputFlag{Names: names}
		list = append(list, item)
	}

	// Sort by lexicographic order of 1st flag name
	compare := func(a, b OutputFlag) int { return cmp.Compare(a.Names[0], b.Names[0]) }
	slices.SortFunc(list, compare)

	return list
}

type OutputDescription struct{ Long, Short string }

func EncodeDescription(info any) (o OutputDescription) {
	type describer interface{ UkaseDescribe() OutputDescription }

	switch infoT := info.(type) {
	case string:
		return OutputDescription{Short: infoT}
	case OutputDescription:
		return infoT
	case describer:
		return infoT.UkaseDescribe()
	default:
		return OutputDescription{}
	}
}
