//go:build ukgen

package ukase

import (
	"github.com/oligarch316/go-ukase/ukcli"
	"github.com/oligarch316/go-ukase/ukcore"
	"github.com/oligarch316/go-ukase/ukcore/ukspec"
	"github.com/oligarch316/go-ukase/ukmeta/ukgen"
)

type appConfig struct{ Config }

func (ac appConfig) UkaseApplyCLI(c *ukcli.Config) {
	ac.cliApplyGen(c)
	ac.cliApplyUser(c)
}

func (ac appConfig) cliApplyGen(c *ukcli.Config) {
	generator := ukgen.NewGenerator(ac.Gen...)

	gentimeMiddleware := func(s ukcli.State) ukcli.State {
		err := generator.Bind().UkaseRegister(s)
		return gentimeState{State: s, registerErr: err}
	}

	c.Middleware = append(c.Middleware, gentimeMiddleware, generator.Middleware)
}

func (ac appConfig) cliApplyUser(c *ukcli.Config) {
	for _, opt := range ac.CLI {
		opt.UkaseApplyCLI(c)
	}
}

type gentimeState struct {
	ukcli.State
	registerErr error
}

func (gs gentimeState) RegisterExec(_ ukcore.Exec, _ ukspec.Parameters, _ ...string) error {
	return gs.registerErr
}
