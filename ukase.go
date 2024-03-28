package ukase

import (
	"context"

	"github.com/oligarch316/go-ukase/ukcore"
	"github.com/oligarch316/go-ukase/ukenc"
)

type Handler[Params any] func(context.Context, Params) error

func New() *ukcore.Mux { return ukcore.NewMux() }

func MustRegister[Params any](m *ukcore.Mux, h Handler[Params], defaults Params, target ...string) {
	if err := Register(m, h, defaults, target...); err != nil {
		panic(err)
	}
}

func Register[Params any](m *ukcore.Mux, h Handler[Params], defaults Params, target ...string) error {
	handleInput := func(ctx context.Context, input ukcore.Input) error {
		decoder := ukenc.NewDecoder(input)
		params := defaults

		if err := decoder.Decode(&params); err != nil {
			return err
		}

		return h(ctx, params)
	}

	return m.Handle(handleInput, defaults, target...)
}
