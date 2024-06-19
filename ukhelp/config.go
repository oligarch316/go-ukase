package ukhelp

import (
	"context"
	_ "embed"
	"io"
	"log/slog"
	"os"
	"strings"

	"github.com/oligarch316/go-ukase/internal"
)

// =============================================================================
// Config
// =============================================================================

type Option interface{ UkaseApplyHelp(*Config) }

type Config struct {
	// TODO: Document
	Log *slog.Logger

	// TODO: Document
	Out io.Writer

	// TODO: Document
	Encode func(context.Context, Input) (any, error)

	// TODO: Document
	Render func(io.Writer, any) error
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
	Log:    internal.LogDiscard,
	Out:    os.Stdout,
	Encode: cfgEncode,
	Render: cfgTemplate.Render,
}

var cfgRenderFuncs renderFuncs

//go:embed help.tmpl
var cfgTemplateText string

var cfgTemplate = TemplateRenderer{
	Name: "help",
	Text: cfgTemplateText,
	Funcs: map[string]any{
		"describeArgument":    cfgRenderFuncs.describeArgument,
		"describeCommand":     cfgRenderFuncs.describeCommand,
		"describeFlag":        cfgRenderFuncs.describeFlag,
		"describeSubcommand":  cfgRenderFuncs.describeSubcommand,
		"hasArguments":        cfgRenderFuncs.hasArguments,
		"hasFlags":            cfgRenderFuncs.hasFlags,
		"hasSubcommands":      cfgRenderFuncs.hasSubcommands,
		"hasCommandExec":      cfgRenderFuncs.hasCommandExec,
		"hasCommandTarget":    cfgRenderFuncs.hasCommandTarget,
		"hasUsage":            cfgRenderFuncs.hasUsage,
		"labelArgument":       cfgRenderFuncs.labelArgument,
		"labelFlag":           cfgRenderFuncs.labelFlag,
		"labelSubcommand":     cfgRenderFuncs.labelSubcommand,
		"maxLabelArguments":   cfgRenderFuncs.maxLabelArguments,
		"maxLabelFlags":       cfgRenderFuncs.maxLabelFlags,
		"maxLabelSubcommands": cfgRenderFuncs.maxLabelSubcommands,
		"stringsJoin":         strings.Join,
	},
}

func cfgEncode(_ context.Context, input Input) (any, error) {
	return Encode(input), nil
}
