package ukhelp

import (
	"context"
	_ "embed"
	"os"

	"github.com/oligarch316/go-ukase/ukcli"
	"github.com/oligarch316/go-ukase/ukcli/ukinfo"
	"github.com/oligarch316/go-ukase/ukmeta"
)

// =============================================================================
// Config
// =============================================================================

type Option interface{ UkaseApplyHelp(*Config) }

type Config struct {
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
	Prepare: cfgPrepare,
	Encode:  cfgEncode,
	Render:  cfgRender,
}

func cfgPrepare(in ukcli.Input, refTarget []string) (ukmeta.Input, error) {
	refTarget = append(refTarget, in.Core().Arguments...)
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
