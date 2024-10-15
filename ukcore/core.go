package ukcore

import "context"

type Exec func(context.Context, Input) error

type Input struct {
	Program   string
	Target    []string
	Arguments []Argument
	Flags     []Flag
}

type Flag struct {
	Name  string
	Value string
}

type Argument struct {
	Position int
	Value    string
}
