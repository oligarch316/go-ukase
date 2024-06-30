package ukopt

import (
	"github.com/oligarch316/go-ukase"
	"github.com/oligarch316/go-ukase/ukcli"
	"github.com/oligarch316/go-ukase/ukcore/ukinit"
)

// =============================================================================
// General
// =============================================================================

var (
	_ ukinit.Option = Init(nil)
	_ ukcli.Option  = Init(nil)
	_ ukase.Option  = Init(nil)
)

type Init func(*ukinit.Config)

func (o Init) UkaseApplyInit(c *ukinit.Config) { o(c) }
func (o Init) UkaseApplyCLI(c *ukcli.Config)   { c.Init = append(c.Init, o) }
func (o Init) UkaseApplyApp(c *ukase.Config)   { c.CLI = append(c.CLI, o) }

// =============================================================================
// Specific
// =============================================================================
