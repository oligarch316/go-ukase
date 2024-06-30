package ukopt

import (
	"github.com/oligarch316/go-ukase"
	"github.com/oligarch316/go-ukase/ukcli"
)

// =============================================================================
// General
// =============================================================================

var (
	_ ukcli.Option = CLI(nil)
	_ ukase.Option = CLI(nil)
)

type CLI func(*ukcli.Config)

func (o CLI) UkaseApplyCLI(c *ukcli.Config) { o(c) }
func (o CLI) UkaseApplyApp(c *ukase.Config) { c.CLI = append(c.CLI, o) }

// =============================================================================
// Specific
// =============================================================================

func CLIMiddleware(middleware func(ukcli.State) ukcli.State) CLI {
	return func(c *ukcli.Config) { c.Middleware = append(c.Middleware, middleware) }
}
