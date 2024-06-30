package ukdec

import "github.com/oligarch316/go-ukase/ukcore/ukspec"

var defaultConfig = Config{}

type Option interface{ UkaseApplyDec(*Config) }

type Config struct {
	// TODO: Document
	Spec []ukspec.Option
}

func newConfig(opts []Option) Config {
	config := defaultConfig
	for _, opt := range opts {
		opt.UkaseApplyDec(&config)
	}
	return config
}
