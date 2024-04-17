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

	Hooks []func(*Runtime, []string)
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

type Record struct {
	Exec   ukcore.Exec
	Spec   ukspec.Params
	Target []string
}

type Runtime struct {
	config  Config
	mux     *ukcore.Mux
	records []Record
	ruleSet *ukinit.RuleSet
}

func New(opts ...Option) *Runtime {
	config := newConfig(opts)

	return &Runtime{
		config:  config,
		mux:     ukcore.New(config.Core...),
		ruleSet: ukinit.NewRuleSet(config.Init...),
	}
}

func (r *Runtime) register(record Record) error {
	r.records = append(r.records, record)
	return r.mux.Register(record.Exec, record.Spec, record.Target...)
}

func (r *Runtime) Records() []Record { return r.records }

func (r *Runtime) Execute(ctx context.Context, values []string) error {
	for _, hook := range r.config.Hooks {
		hook(r, values)
	}

	return r.mux.Execute(ctx, values)
}

// =============================================================================
// Rule
// =============================================================================

type Rule interface{ Register(runtime *Runtime) }

func NewRule[Params any](op func(*Params)) Rule {
	return rule[Params](op)
}

func AddRule[Params any](runtime *Runtime, op func(*Params)) {
	NewRule(op).Register(runtime)
}

type rule[Params any] func(*Params)

func (r rule[Params]) Register(runtime *Runtime) {
	ukinit.NewRule(r).Register(runtime.ruleSet)
}

// =============================================================================
// Command
// =============================================================================

type Command interface {
	Register(runtime *Runtime, target ...string) error
}

func NewCommand[Params any](handler func(context.Context, Params) error) Command {
	return command[Params](handler)
}

func AddCommand[Params any](runtime *Runtime, handler func(context.Context, Params) error, target ...string) error {
	return NewCommand(handler).Register(runtime, target...)
}

type command[Params any] func(context.Context, Params) error

func (c command[Params]) Register(runtime *Runtime, target ...string) error {
	spec, err := ukspec.For[Params]()
	if err != nil {
		return err
	}

	exec := exec[Params]{runtime: runtime, handler: c}
	record := Record{Exec: exec, Spec: spec, Target: target}
	return runtime.register(record)
}

// =============================================================================
// Exec
// =============================================================================

type exec[Params any] struct {
	runtime *Runtime
	handler func(context.Context, Params) error
}

func (e exec[Params]) Execute(ctx context.Context, input ukcore.Input) error {
	params, err := ukinit.For[Params](e.runtime.ruleSet)
	if err != nil {
		// TODO:
		// Config/Option for control?
		// Or at least discernable error type
		return err
	}

	decoder := ukenc.NewDecoder(input, e.runtime.config.Enc...)
	if err := decoder.Decode(&params); err != nil {
		// TODO:
		// Config/Option for control?
		// Or at least discernable error type
		return err
	}

	return e.handler(ctx, params)
}
