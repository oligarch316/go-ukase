package ukhelp

import (
	"context"

	"github.com/oligarch316/ukase/ukcli"
	"github.com/oligarch316/ukase/ukmeta"
)

type Builder struct{ config Config }

func NewBuilder(opts ...Option) Builder {
	config := newConfig(opts)
	return Builder{config: config}
}

func (b Builder) Auto(name string) func(ukcli.State) ukcli.State {
	builder := ukmeta.NewBuilder(b.Build)
	return builder.Auto(name)
}

func (b Builder) Build(refTarget ...string) (ukcli.Exec[struct{}], any) {
	exec := func(ctx context.Context, in ukcli.Input) error {
		helpInput, err := b.config.Prepare(in, refTarget)
		if err != nil {
			return err
		}

		helpData, err := b.config.Encode(helpInput)
		if err != nil {
			return err
		}

		return b.config.Render(ctx, helpData)
	}

	return exec, b.config.Info
}
