package ukinit

import "github.com/oligarch316/ukase/ukcore/ukspec"

var defaultConfig = Config{
	// TODO
}

type Option interface{ UkaseApplyInit(*Config) }

type Config struct {
	// TODO: Document
	Spec []ukspec.Option
}

func newConfig(opts []Option) Config {
	config := defaultConfig
	for _, opt := range opts {
		opt.UkaseApplyInit(&config)
	}
	return config
}
