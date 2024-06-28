package ukhelp

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
	Names       []string
}

type OutputArgument[T any] struct {
	Description T
	IndexStart  int
	IndexEnd    int
}
