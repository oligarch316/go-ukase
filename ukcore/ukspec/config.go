package ukspec

import (
	"log/slog"

	"github.com/oligarch316/go-ukase/internal/ilog"
	"github.com/oligarch316/go-ukase/internal/ispec"
)

// =============================================================================
// Config
// =============================================================================

type Option interface{ UkaseApplySpec(*Config) }

type Config struct {
	// TODO: Document
	Log *slog.Logger

	// TODO: Document
	ElideAllowBoolType bool

	// TODO: Document
	ElideAllowIsBoolFlag bool

	// TODO: Document
	ElideConsumable func(string) bool
}

func newConfig(opts []Option) Config {
	config := cfgDefault
	for _, opt := range opts {
		opt.UkaseApplySpec(&config)
	}
	return config
}

// =============================================================================
// Defaults
// =============================================================================

var cfgDefault = Config{
	Log:                  ilog.Discard,
	ElideAllowBoolType:   true,
	ElideAllowIsBoolFlag: false,
	ElideConsumable:      cfgElideConsumable,
}

var cfgElideConsumable = ispec.ConsumableSet(
	"true", "false",
	"True", "False",
	"TRUE", "FALSE",
)
