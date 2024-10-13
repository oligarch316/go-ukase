package ukhelp

import "github.com/oligarch316/go-ukase/ukcore/ukspec"

type Output[T any] struct {
	Command     OutputCommand[T]
	Subcommands []OutputSubcommand[T]
	Flags       []OutputFlag[T]
	Arguments   []OutputArgument[T]
}

type OutputCommand[T any] struct {
	Description T
	Program     string
	Target      []string
	Exec        bool
}

type OutputSubcommand[T any] struct {
	Description T
	Name        string
}

type OutputFlag[T any] struct {
	Description T
	Names       ukspec.FlagNames
}

type OutputArgument[T any] struct {
	Description T
	Position    ukspec.ArgumentPosition
}
