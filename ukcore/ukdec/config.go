package ukdec

import (
	"log/slog"

	"github.com/oligarch316/ukase/internal/ilog"
	"github.com/oligarch316/ukase/ukcore/ukspec"
)

// =============================================================================
// Config
// =============================================================================

type Option interface{ UkaseApplyDec(*Config) }

type Config struct {
	// TODO: Document
	Log *slog.Logger

	// TODO: Document
	Spec []ukspec.Option
}

func newConfig(opts []Option) Config {
	config := cfgDefault
	for _, opt := range opts {
		opt.UkaseApplyDec(&config)
	}
	return config
}

// =============================================================================
// Defaults
// =============================================================================

var cfgDefault = Config{
	Log:  ilog.Discard,
	Spec: nil,
}
