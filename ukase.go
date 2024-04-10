package ukase

import (
	"context"

	"github.com/oligarch316/go-ukase/ukcore"
	"github.com/oligarch316/go-ukase/ukreflect/ukenc"
	"github.com/oligarch316/go-ukase/ukreflect/ukinit"
	"github.com/oligarch316/go-ukase/ukspec"
)

// =============================================================================
// Config
// =============================================================================

var defaultConfig = Config{}

type Option interface{ UkaseApply(*Config) }

type Config struct {
	Core []ukcore.Option
	Enc  []ukenc.Option
	Init []ukinit.Option
}

func newConfig(opts []Option) Config {
	config := defaultConfig
	for _, opt := range opts {
		opt.UkaseApply(&config)
	}
	return config
}

// =============================================================================
// Runtime
// =============================================================================

type Runtime struct {
	config  Config
	mux     *ukcore.Mux
	ruleSet *ukinit.RuleSet
}

func New(opts ...Option) *Runtime {
	config := newConfig(opts)

	return &Runtime{
		config:  config,
		mux:     ukcore.New(config.Core...),
		ruleSet: ukinit.New(config.Init...),
	}
}

func Default[Params any](runtime *Runtime, rules ...func(*Params)) {
	ukinit.Register(runtime.ruleSet, rules...)
}

func Command[Params any](runtime *Runtime, handler Handler[Params], target ...string) error {
	spec, err := ukspec.For[Params]()
	if err != nil {
		return err
	}

	cmd := command[Params]{runtime: runtime, handler: handler}
	return runtime.mux.Register(cmd, spec, target...)
}

func (r *Runtime) Execute(ctx context.Context, values []string) error {
	return r.mux.Execute(ctx, values)
}

// =============================================================================
// Command
// =============================================================================

type Handler[Params any] func(context.Context, Params) error

type command[Params any] struct {
	runtime *Runtime
	handler Handler[Params]
}

func (c command[Params]) Execute(ctx context.Context, input ukcore.Input) error {
	params, err := ukinit.Create[Params](c.runtime.ruleSet)
	if err != nil {
		return err
	}

	decoder := ukenc.NewDecoder(input, c.runtime.config.Enc...)
	if err := decoder.Decode(&params); err != nil {
		return err
	}

	return c.handler(ctx, params)
}
