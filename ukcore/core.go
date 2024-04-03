package ukcore

import "context"

type Target []string

type Flag struct{ Name, Value string }

type Input struct {
	Target Target
	Flags  []Flag
	Args   []string
}

type Executor interface {
	Execute(context.Context, Input) error
}

type Handler func(context.Context, Input) error

func (h Handler) Execute(ctx context.Context, input Input) error { return h(ctx, input) }

type Command struct {
	Executor
	DefaultParams any
}

func (c Command) Register(mux *Mux, target ...string) error {
	return mux.Register(c.Executor, c.DefaultParams, target...)
}
