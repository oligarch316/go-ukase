package ukhelp

import (
	"cmp"
	"slices"

	"github.com/oligarch316/go-ukase/ukcore/ukspec"
	"github.com/oligarch316/go-ukase/ukmeta"
)

type Encoder[T any] func(info any) (description T, err error)

func NewEncoder[T any](encodeDescription func(any) (T, error)) Encoder[T] {
	return Encoder[T](encodeDescription)
}

// =============================================================================
// Encode
// =============================================================================

func (e Encoder[T]) Encode(in ukmeta.Input) (Output[T], error) {
	command, err := e.EncodeCommand(in)
	if err != nil {
		return Output[T]{}, err
	}

	subcommands, err := e.EncodeSubcommands(in)
	if err != nil {
		return Output[T]{}, err
	}

	flags, err := e.EncodeFlags(in)
	if err != nil {
		return Output[T]{}, err
	}

	arguments, err := e.EncodeArguments(in)
	if err != nil {
		return Output[T]{}, err
	}

	output := Output[T]{
		Command:     command,
		Subcommands: subcommands,
		Flags:       flags,
		Arguments:   arguments,
	}

	return output, nil
}

func (e Encoder[T]) EncodeCommand(in ukmeta.Input) (o OutputCommand[T], err error) {
	core, reference := in.Core(), in.MetaReference()

	o.Description, err = e(reference.Info)
	o.Program = core.Program
	o.Target = reference.Target
	o.Exec = reference.Exec

	return
}

func (e Encoder[T]) EncodeSubcommands(in ukmeta.Input) ([]OutputSubcommand[T], error) {
	var list []OutputSubcommand[T]

	for name, meta := range in.MetaReference().Children() {
		description, err := e(meta.Info)
		if err != nil {
			return nil, err
		}

		item := OutputSubcommand[T]{Description: description, Name: name}
		list = append(list, item)
	}

	e.SortSubcommands(list)
	return list, nil
}

func (e Encoder[T]) EncodeFlags(in ukmeta.Input) ([]OutputFlag[T], error) {
	var list []OutputFlag[T]

	for _, spec := range in.MetaReference().Spec.Flags {
		names := slices.Clone(spec.FlagNames)
		e.SortFlagNames(names)

		item := OutputFlag[T]{Names: names}
		list = append(list, item)
	}

	e.SortFlags(list)
	return list, nil
}

func (e Encoder[T]) EncodeArguments(in ukmeta.Input) ([]OutputArgument[T], error) {
	var list []OutputArgument[T]

	// >===== TODO
	var argumentSpecs []ukspec.Args
	if tmp := in.MetaReference().Spec.Args; tmp != nil {
		argumentSpecs = []ukspec.Args{*tmp}
	}
	// <=====

	for range argumentSpecs {
		// TODO
		indexStart, indexEnd := -1, -1

		item := OutputArgument[T]{IndexStart: indexStart, IndexEnd: indexEnd}
		list = append(list, item)
	}

	e.SortArguments(list)
	return list, nil
}

// =============================================================================
// Sort
// =============================================================================

func (e Encoder[T]) SortSubcommands(list []OutputSubcommand[T]) {
	// Sort by lexicographic order of subcommand name
	compare := func(a, b OutputSubcommand[T]) int { return cmp.Compare(a.Name, b.Name) }
	slices.SortFunc(list, compare)
}

func (e Encoder[T]) SortFlags(list []OutputFlag[T]) {
	// Sort by lexicographic order of 1st flag name
	compare := func(a, b OutputFlag[T]) int { return cmp.Compare(a.Names[0], b.Names[0]) }
	slices.SortFunc(list, compare)
}

func (e Encoder[T]) SortArguments(list []OutputArgument[T]) {
	// Sort by position
	// TODO
}

func (e Encoder[T]) SortFlagNames(list []string) {
	// Sort by length of name
	compare := func(a, b string) int { return len(a) - len(b) }
	slices.SortFunc(list, compare)
}
