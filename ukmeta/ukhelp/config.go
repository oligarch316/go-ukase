package ukhelp

import (
	_ "embed"

	"context"
	"os"

	"github.com/oligarch316/ukase/ukcli"
	"github.com/oligarch316/ukase/ukcli/ukinfo"
	"github.com/oligarch316/ukase/ukmeta"
)

// =============================================================================
// Config
// =============================================================================

type Option interface{ UkaseApplyHelp(*Config) }

type Config struct {
	Info    any
	Prepare func(in ukcli.Input, refTarget []string) (ukmeta.Input, error)
	Encode  func(in ukmeta.Input) (any, error)
	Render  func(ctx context.Context, data any) error
}

func newConfig(opts []Option) Config {
	config := cfgDefault
	for _, opt := range opts {
		opt.UkaseApplyHelp(&config)
	}
	return config
}

// =============================================================================
// Defaults
// =============================================================================

var cfgDefault = Config{
	Info:    "Show help information",
	Prepare: cfgPrepare,
	Encode:  cfgEncode,
	Render:  cfgRender,
}

func cfgPrepare(in ukcli.Input, refTarget []string) (ukmeta.Input, error) {
	// TODO:
	// Since adding the `ukcore.Argument` type with included `Position` field,
	// add some rigor (docs or otherwise) to assumption they come in sorted order.

	for _, arg := range in.Core().Arguments {
		refTarget = append(refTarget, arg.Value)
	}

	return ukmeta.NewInput(in, refTarget...)
}

func cfgEncode(in ukmeta.Input) (any, error) {
	encoder := NewEncoder(ukinfo.Encode)
	return encoder.Encode(in)
}

func cfgRender(ctx context.Context, data any) error {
	return cfgTemplate.Render(data)
}

var cfgTemplate = TemplateRenderer{
	Name:  "help",
	Text:  cfgTemplateText,
	Out:   os.Stdout,
	Funcs: cfgRenderFuncs.Map(),
}

//go:embed render.tmpl
var cfgTemplateText string

var cfgRenderFuncs = NewRenderFuncs(ukinfo.Render)
