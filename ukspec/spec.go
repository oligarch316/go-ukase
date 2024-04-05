package ukspec

import "reflect"

// =============================================================================
// Config
// =============================================================================

var defaultConfig = Config{
	Elide: ElideConfig{
		AllowBoolType:   true,
		AllowIsBoolFlag: false,
		DecideDefault:   NewDecideSet("true", "True", "TRUE", "false", "False", "FALSE"),
	},
}

type Option interface{ UkaseApplySpec(*Config) }

type Config struct {
	Elide ElideConfig
}

func newConfig(opts []Option) Config {
	config := defaultConfig
	for _, opt := range opts {
		opt.UkaseApplySpec(&config)
	}
	return config
}

// =============================================================================
// Sugar
// =============================================================================

func Create[T any](opts ...Option) (Params, error) {
	var tmp [0]T
	t := reflect.TypeOf(tmp).Elem()
	return New(t, opts...)
}

func Parse(v any, opts ...Option) (Params, error) {
	t := reflect.TypeOf(v)
	return New(t, opts...)
}
