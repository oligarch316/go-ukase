package ukhelp

import (
	_ "embed"
	"html/template"
	"io"
	"os"
)

// =============================================================================
// Config
// =============================================================================

var defaultConfig = Config{
	Out:      os.Stdout,
	Template: defaultTemplate,
}

type Option interface{ UkaseApplyHelp(*Config) }

type Config struct {
	Out      io.Writer
	Template func(target []string) (*template.Template, error)
}

func newConfig(opts []Option) Config {
	config := defaultConfig
	for _, opt := range opts {
		opt.UkaseApplyHelp(&config)
	}
	return config
}

// =============================================================================
// Template Default
// =============================================================================

//go:embed help.tmpl
var helpTmpl string

func defaultTemplate(_ []string) (*template.Template, error) {
	return template.New("help").Parse(helpTmpl)
}
