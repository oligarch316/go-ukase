package ukinit

import "github.com/oligarch316/go-ukase/ukspec"

var defaultConfig = Config{
	// TODO
	// Event: func(any) {},
}

type Option interface{ UkaseApplyInit(*Config) }

type Config struct {
	// TODO: Document
	Spec []ukspec.Option

	// TODO
	// Event func(any)
}

func newConfig(opts []Option) Config {
	config := defaultConfig
	for _, opt := range opts {
		opt.UkaseApplyInit(&config)
	}
	return config
}
