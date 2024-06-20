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

var _ Input = input{}

type InputReference struct {
	ukcore.Meta
	Target ukcore.InputTarget
}

type Input interface {
	ukase.Input

	Reference() InputReference
	DefaultByIndex(index []int) (any, error)
}

type input struct {
	ukase.Input

	reference    InputReference
	loadDefaults func() (reflect.Value, error)
}

func newInput(ref []string, in ukase.Input) (input, error) {
	ref = append(ref, in.Core().Args...)

	meta, err := in.Meta(ref)
	if err != nil {
		return input{}, err
	}

	reference := InputReference{Meta: meta, Target: ref}

	loadDefaults := func() (reflect.Value, error) {
		ptrVal := reflect.New(meta.Spec.Type)
		err := in.Initialize(ptrVal.Interface())
		return ptrVal.Elem(), err
	}

	input := input{
		Input:        in,
		reference:    reference,
		loadDefaults: sync.OnceValues(loadDefaults),
	}

	return input, nil
}

func (i input) Reference() InputReference { return i.reference }

func (i input) DefaultByIndex(index []int) (any, error) {
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

func Encode(in Input) Output {
	return Output{
		Command:     EncodeCommand(in),
		Subcommands: EncodeSubcommands(in),
		Arguments:   EncodeArguments(in),
		Flags:       EncodeFlags(in),
	}
}

type OutputCommand struct {
	Description OutputDescription
	Program     string
	Target      []string
	Exec        bool
}

func EncodeCommand(in Input) OutputCommand {
	reference := in.Reference()

	return OutputCommand{
		Description: EncodeDescription(reference.Info),
		Program:     in.Core().Program,
		Target:      reference.Target,
		Exec:        reference.Exec,
	}
}

type OutputSubcommand struct {
	Description OutputDescription
	Name        string
}

func EncodeSubcommands(in Input) []OutputSubcommand {
	var list []OutputSubcommand

	for name, meta := range in.Reference().Children() {
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

func EncodeArguments(in Input) []OutputArgument {
	var list []OutputArgument

	// TODO
	var argSpecs []ukspec.Args
	if tmp := in.Reference().Spec.Args; tmp != nil {
		argSpecs = []ukspec.Args{*tmp}
	}

	for range argSpecs {
		// TODO
		position := OutputArgumentPosition{Start: -1, End: -1}
		item := OutputArgument{Position: position}
		list = append(list, item)
	}

	// TODO: Sort by position

	return list
}

type OutputArgumentPosition struct{ Start, End int }

type OutputFlag struct {
	Description OutputDescription
	Names       []string
}

func EncodeFlags(in Input) []OutputFlag {
	var list []OutputFlag

	for _, spec := range in.Reference().Spec.Flags {
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
