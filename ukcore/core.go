package ukcore

import "context"

type Target []string

type Flag struct{ Name, Value string }

type Input struct {
	Target Target
	Flags  []Flag
	Args   []string
}

type Handler func(context.Context, Input) error
