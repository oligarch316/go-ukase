package ukspec

// =============================================================================
// Config
// =============================================================================

var defaultConfig = Config{
	ElideBoolType:          true,
	ElideIsBoolFlag:        false,
	ElideDefaultConsumable: defaultConsumable,
}

type Option interface{ UkaseApplySpec(*Config) }

type Config struct {
	// TODO: Document
	ElideBoolType bool

	// TODO: Document
	ElideIsBoolFlag bool

	// TODO: Document
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

	return func(s string) (ok bool) { _, ok = set[s]; return }
}
