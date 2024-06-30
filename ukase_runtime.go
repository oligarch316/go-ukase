//go:build !ukgen

package ukase

import (
	"github.com/oligarch316/go-ukase/ukcli"
	"github.com/oligarch316/go-ukase/ukmeta/ukhelp"
)

type appConfig struct{ Config }

func (ac appConfig) UkaseApplyCLI(c *ukcli.Config) {
	ac.cliApplyHelpAuto(c)
	ac.cliApplyUser(c)
}

func (ac appConfig) cliApplyHelpAuto(c *ukcli.Config) {
	helpBuilder := ukhelp.NewBuilder(ac.Help...)
	helpAuto := helpBuilder.Auto(ac.HelpCommand)

	c.Middleware = append(c.Middleware, helpAuto)
}

func (ac appConfig) cliApplyUser(c *ukcli.Config) {
	for _, opt := range ac.CLI {
		opt.UkaseApplyCLI(c)
	}
}
