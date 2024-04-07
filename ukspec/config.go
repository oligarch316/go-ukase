package ukspec

// =============================================================================
// Option
// =============================================================================

type Option interface{ UkaseApplySpec(*Config) }

// =============================================================================
// Config
// =============================================================================

var defaultConfig = Config{
	ElideBoolType:          true,
	ElideIsBoolFlag:        false,
	ElideDefaultConsumable: defaultConsumable,
}

type Config struct {
	ElideBoolType          bool
	ElideIsBoolFlag        bool
	ElideDefaultConsumable func(string) bool
}

func newConfig(opts []Option) Config {
	config := defaultConfig
	for _, opt := range opts {
		opt.UkaseApplySpec(&config)
	}
	return config
}

// =============================================================================
// Consumable
// =============================================================================

var defaultConsumable = ConsumableSet(
	"true", "false",
	"True", "False",
	"TRUE", "FALSE",
)

func ConsumableSet(valid ...string) func(string) bool {
	set := make(map[string]struct{})
	for _, item := range valid {
		set[item] = struct{}{}
	}

	return func(text string) bool { _, ok := set[text]; return ok }
}
