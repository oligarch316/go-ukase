package ukase

import (
	"context"

	"github.com/oligarch316/go-ukase/ukcore"
	"github.com/oligarch316/go-ukase/ukreflect/ukenc"
	"github.com/oligarch316/go-ukase/ukreflect/ukinit"
	"github.com/oligarch316/go-ukase/ukspec"
)

var defaultConfig = Config{}

type Option interface{ UkaseApply(*Config) }

type Config struct {
	Core []ukcore.Option
	Spec []ukspec.Option
}

func newConfig(opts []Option) Config {
	config := defaultConfig
	for _, opt := range opts {
		opt.UkaseApply(&config)
	}
	return config
}

type Runtime struct {
	config Config
	mux    *ukcore.Mux
	rules  ukinit.RuleSet
}

func New(opts ...Option) *Runtime {
	config := newConfig(opts)

	return &Runtime{
		config: config,
		mux:    ukcore.New(config.Core...),
		rules:  ukinit.New(),
	}
}

func (r *Runtime) Execute(ctx context.Context, values []string) error {
	return r.mux.Execute(ctx, values)
}

func Default[Params any](runtime *Runtime, defaults ...func(*Params)) {
	ukinit.Register(runtime.rules, defaults...)
}

func Command[Params any](runtime *Runtime, handler func(context.Context, Params) error, target ...string) error {
	spec, err := ukspec.For[Params]()
	if err != nil {
		return err
	}

	cmd := command[Params]{runtime: runtime, handler: handler}
	return runtime.mux.Register(cmd, spec, target...)
}

type command[Params any] struct {
	runtime *Runtime
	handler func(context.Context, Params) error
}

func (c command[Params]) Execute(ctx context.Context, input ukcore.Input) error {
	params, err := ukinit.Create[Params](c.runtime.rules)
	if err != nil {
		return err
	}

	if err := ukenc.NewDecoder(input).Decode(&params); err != nil {
		return err
	}

	return c.handler(ctx, params)
}
