package ukcli

import (
	"log/slog"

	"github.com/oligarch316/ukase/internal/ilog"
	"github.com/oligarch316/ukase/ukcore/ukdec"
	"github.com/oligarch316/ukase/ukcore/ukexec"
	"github.com/oligarch316/ukase/ukcore/ukinit"
	"github.com/oligarch316/ukase/ukcore/ukspec"
)

// =============================================================================
// Config
// =============================================================================

type Option interface{ UkaseApplyCLI(*Config) }

type Config struct {
	// TODO: Document
	Log *slog.Logger

	// TODO: Document
	Exec []ukexec.Option

	// TODO: Document
	Decode []ukdec.Option

	// TODO: Document
	Init []ukinit.Option

	// TODO: Document
	Spec []ukspec.Option

	// TODO: Document
	Middleware []func(State) State
}

func newConfig(opts []Option) Config {
	config := cfgDefault
	for _, opt := range opts {
		opt.UkaseApplyCLI(&config)
	}
	return config
}

// =============================================================================
// Defaults
// =============================================================================

var cfgDefault = Config{
	Log:        ilog.Discard,
	Exec:       nil,
	Decode:     nil,
	Init:       nil,
	Spec:       nil,
	Middleware: nil,
}
