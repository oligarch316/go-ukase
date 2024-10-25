package ukase

import (
	"context"
	"errors"
	"log/slog"
	"os"

	"github.com/oligarch316/ukase/internal/ilog"
	"github.com/oligarch316/ukase/ukcli"
	"github.com/oligarch316/ukase/ukmeta/ukgen"
	"github.com/oligarch316/ukase/ukmeta/ukhelp"
)

// =============================================================================
// Config
// =============================================================================

var cfgDefault = Config{
	Log:            ilog.Discard,
	HelpCommand:    "help",
	InputProgram:   os.Args[0],
	InputArguments: os.Args[1:],
	CLI:            nil,
	Help:           nil,
	Gen:            nil,
}

type Option interface{ UkaseApplyApp(*Config) }

type Config struct {
	// TODO: Document
	Log *slog.Logger

	// TODO: Document
	HelpCommand string

	// TODO: Document
	InputProgram string

	// TODO: Document
	InputArguments []string

	// TODO: Document
	CLI []ukcli.Option

	// TODO: Document
	Help []ukhelp.Option

	// TODO: Document
	Gen []ukgen.Option
}

func newConfig(opts []Option) appConfig {
	config := cfgDefault
	for _, opt := range opts {
		opt.UkaseApplyApp(&config)
	}
	return appConfig{Config: config}
}

// =============================================================================
// Application
// =============================================================================

type Application struct {
	config  appConfig
	runtime *ukcli.Runtime
}

func NewApplication(opts ...Option) *Application {
	config := newConfig(opts)
	runtime := ukcli.NewRuntime(config)

	return &Application{config: config, runtime: runtime}
}

func (a *Application) Add(directives ...ukcli.Directive) {
	a.runtime.Add(directives...)
}

func (a *Application) Run(ctx context.Context) error {
	values := []string{a.config.InputProgram}
	values = append(values, a.config.InputArguments...)

	return a.runtime.Execute(ctx, values)
}

// =============================================================================
// Directive› Command
// =============================================================================

var NoHandler ukcli.Handler[struct{}] = nil
var NoInfo any = nil

type Command[Params any] struct {
	Handler ukcli.Handler[Params]
	Info    ukcli.Info
}

func NewCommand[Params any](handler func(context.Context, Params) error, info any) Command[Params] {
	return Command[Params]{
		Handler: ukcli.NewHandler(handler),
		Info:    ukcli.NewInfo(info),
	}
}

func NewRoot[Params any](handler func(context.Context, Params) error, info any) ukcli.Directive {
	command := NewCommand(handler, info)
	return command.Bind()
}

func (c Command[Params]) Bind(target ...string) ukcli.Directive {
	return ukcli.NewDirective(func(s ukcli.State) error {
		errHandler := c.Handler.Bind(target...).UkaseRegister(s)
		errInfo := c.Info.Bind(target...).UkaseRegister(s)
		return errors.Join(errHandler, errInfo)
	})
}

// =============================================================================
// Directive› Rule
// =============================================================================

func NewRule[Params any](rule func(*Params)) ukcli.Rule[Params] {
	return ukcli.NewRule(rule)
}
