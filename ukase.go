package ukase

import (
	"context"

	"github.com/oligarch316/go-ukase/ukcore"
	"github.com/oligarch316/go-ukase/ukenc"
)

// =============================================================================
// Mux
// =============================================================================

func MustMux(opts ...func(*ukcore.MuxConfig)) *ukcore.Mux {
	mux, err := NewMux(opts...)
	if err != nil {
		panic(err)
	}

	return mux
}

func NewMux(opts ...func(*ukcore.MuxConfig)) (*ukcore.Mux, error) {
	// TODO:
	// Create a "help" command, set MuxConfig.DefaultCommand = <help>

	return ukcore.NewMux(opts...)
}

// =============================================================================
// Command
// =============================================================================

type Handler[Params any] func(context.Context, Params) error

type Command[Params any] struct {
	Handler  Handler[Params]
	Defaults Params
}

func NewCommand[Params any](handler Handler[Params], overrides ...func(*Params)) Command[Params] {
	defaults := initialize[Params](overrides...)
	return Command[Params]{Handler: handler, Defaults: defaults}
}

func (c Command[Params]) Execute(ctx context.Context, input ukcore.Input) error {
	// TODO:
	// Do we need/want to worry about deep/shallow copy issues here?
	// If so, 1st thought is to store opts in the Command struct, then do initialization here
	// Removes the ability to do the following tho, so trying to avoid:
	//   myCommand := NewCommand(myFunc)
	//   myCommand.Defaults.ParamA = "abc"
	//   myCommand.Defaults.ParamB = "xyz"

	params := c.Defaults
	decoder := ukenc.NewDecoder(input)

	if err := decoder.Decode(&params); err != nil {
		return err
	}

	return c.Handler(ctx, params)
}

func (c Command[Params]) MustRegister(mux *ukcore.Mux, target ...string) {
	if err := c.Register(mux, target...); err != nil {
		panic(err)
	}
}

func (c Command[Params]) Register(mux *ukcore.Mux, target ...string) error {
	return mux.Register(c, c.Defaults, target...)
}

func initialize[Params any](overrides ...func(*Params)) Params {
	type Initializer interface{ InitUkaseParams() }

	params := new(Params)
	if initializer, ok := (any)(params).(Initializer); ok {
		initializer.InitUkaseParams()
	}

	for _, override := range overrides {
		override(params)
	}

	return *params
}
