package ukgen

import (
	"log/slog"
	"reflect"

	"github.com/oligarch316/go-ukase/internal/ilog"
	"github.com/oligarch316/go-ukase/ukcli/ukinfo"
	"github.com/oligarch316/go-ukase/ukcore/ukspec"
)

// =============================================================================
// Config
// =============================================================================

type Option interface{ UkaseApplyGen(*Config) }

type Config struct {
	Log   *slog.Logger
	Spec  []ukspec.Option
	Names ConfigNames
	Types ConfigTypes
}

type ConfigNames struct {
	Package            string
	EncoderConstructor string
	EncoderDefault     string
	EncoderType        string
	ParameterTypes     map[reflect.Type]string
}

type ConfigTypes struct {
	ArgumentInfo reflect.Type
	FlagInfo     reflect.Type
}

func newConfig(opts []Option) Config {
	config := cfgDefault
	for _, opt := range opts {
		opt.UkaseApplyGen(&config)
	}
	return config
}

// =============================================================================
// Defaults
// =============================================================================

var cfgDefault = Config{
	Log:   ilog.Discard,
	Spec:  nil,
	Names: cfgDefaultNames,
	Types: cfgDefaultTypes,
}

var cfgDefaultNames = ConfigNames{
	Package:            "ukdoc",
	EncoderConstructor: "NewHelpEncoder",
	EncoderDefault:     "EncodeHelp",
	EncoderType:        "HelpEncoder",
	ParameterTypes:     make(map[reflect.Type]string),
}

var cfgDefaultTypes = ConfigTypes{
	ArgumentInfo: reflect.TypeFor[ukinfo.Any](),
	FlagInfo:     reflect.TypeFor[ukinfo.Any](),
}
