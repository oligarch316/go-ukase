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
	ExecDefault:  new(execDefault),
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

type execDefault []string

func (ed execDefault) Error() string {
	return fmt.Sprintf("unspecified target '%s'", strings.Join(ed, "."))
}

func (execDefault) Execute(_ context.Context, input Input) error {
	return execDefault(input.Target)
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
