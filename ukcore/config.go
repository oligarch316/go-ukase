package ukcore

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/oligarch316/go-ukase/internal"
	"github.com/oligarch316/go-ukase/ukspec"
)

// =============================================================================
// Config
// =============================================================================

type Option interface{ UkaseApplyCore(*Config) }

type Config struct {
	// TODO: Document
	Log *slog.Logger

	// TODO: Document
	ExecUnspecified Exec

	// TODO: Document
	ExecConflict func(original, update ukspec.Params) (overwrite bool, err error)

	// TODO: Document
	InfoConflict func(original, update any) (overwrite bool, err error)

	// TODO: Document
	FlagConflict func(original, update ukspec.Flag) error
}

func newConfig(opts []Option) Config {
	config := cfgDefault
	for _, opt := range opts {
		opt.UkaseApplyCore(&config)
	}
	return config
}

// =============================================================================
// Defaults
// =============================================================================

var cfgDefault = Config{
	Log:             internal.LogDiscard,
	ExecUnspecified: cfgExecUnspecified,
	ExecConflict:    cfgExecConflict,
	InfoConflict:    cfgInfoConflict,
	FlagConflict:    cfgFlagConflict,
}

func cfgExecUnspecified(_ context.Context, i Input) error {
	return fmt.Errorf("unspecified target '%s'", strings.Join(i.Target, " "))
}

func cfgExecConflict(_, _ ukspec.Params) (bool, error) {
	return false, errors.New("exec already exists")
}

func cfgInfoConflict(_, _ any) (bool, error) {
	return false, errors.New("info already exists")
}

func cfgFlagConflict(o, u ukspec.Flag) error {
	if o.Elide.Allow != u.Elide.Allow {
		return fmt.Errorf("incompatible elide behavior '%t' and '%t'", o.Elide.Allow, u.Elide.Allow)
	}

	return nil
}
