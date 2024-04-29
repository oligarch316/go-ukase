package ukcore

import (
	"context"
	"fmt"
	"strings"
)

// =============================================================================
// Config
// =============================================================================

var defaultConfig = Config{
	ExecDefault:  execDefault,
	FlagCheck:    FlagCheckElide,
	MuxOverwrite: false,
}

type Option interface{ UkaseApplyCore(*Config) }

type Config struct {
	// TODO: Document
	ExecDefault Exec

	// TODO: Document
	FlagCheck FlagCheckLevel

	// TODO: Document
	MuxOverwrite bool
}

func newConfig(opts []Option) Config {
	config := defaultConfig
	for _, opt := range opts {
		opt.UkaseApplyCore(&config)
	}
	return config
}

// =============================================================================
// Exec Default
// =============================================================================

func execDefault(_ context.Context, input Input) error {
	return fmt.Errorf("unspecified target '%s'", strings.Join(input.Target, " "))
}

// =============================================================================
// Flag Check Level
// =============================================================================

type FlagCheckLevel int

const (
	FlagCheckNone FlagCheckLevel = iota
	FlagCheckElide
	FlagCheckType
)

var flagCheckLevelToString = map[FlagCheckLevel]string{
	FlagCheckNone:  "none",
	FlagCheckElide: "elide",
	FlagCheckType:  "type",
}

func (fcl FlagCheckLevel) String() string {
	if str, ok := flagCheckLevelToString[fcl]; ok {
		return str
	}
	return fmt.Sprintf("unknown(%d)", fcl)
}
