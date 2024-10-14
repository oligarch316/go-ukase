package ukcore

import "context"

type Exec func(context.Context, Input) error

type Input struct {
	Program   string
	Target    []string
	Arguments []string
	Flags     []Flag
}

type Flag struct{ Name, Value string }
