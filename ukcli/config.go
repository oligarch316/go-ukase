package ukcli

import (
	"github.com/oligarch316/go-ukase/ukcore/ukdec"
	"github.com/oligarch316/go-ukase/ukcore/ukexec"
	"github.com/oligarch316/go-ukase/ukcore/ukinit"
	"github.com/oligarch316/go-ukase/ukcore/ukspec"
)

// =============================================================================
// Config
// =============================================================================

type Option interface{ UkaseApply(*Config) }

type Config struct {
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
		opt.UkaseApply(&config)
	}
	return config
}

// =============================================================================
// Defaults
// =============================================================================

var cfgDefault = Config{}
