package ukopt

import (
	"github.com/oligarch316/ukase"
	"github.com/oligarch316/ukase/ukcli"
	"github.com/oligarch316/ukase/ukcore/ukexec"
)

// =============================================================================
// General
// =============================================================================

var (
	_ ukexec.Option = Exec(nil)
	_ ukcli.Option  = Exec(nil)
	_ ukase.Option  = Exec(nil)
)

type Exec func(*ukexec.Config)

func (o Exec) UkaseApplyExec(c *ukexec.Config) { o(c) }
func (o Exec) UkaseApplyCLI(c *ukcli.Config)   { c.Exec = append(c.Exec, o) }
func (o Exec) UkaseApplyApp(c *ukase.Config)   { c.CLI = append(c.CLI, o) }

// =============================================================================
// Specific
// =============================================================================
