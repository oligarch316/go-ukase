package ukopt

import (
	"github.com/oligarch316/go-ukase"
	"github.com/oligarch316/go-ukase/ukcli"
	"github.com/oligarch316/go-ukase/ukcore/ukdec"
)

// =============================================================================
// General
// =============================================================================

var (
	_ ukdec.Option = Dec(nil)
	_ ukcli.Option = Dec(nil)
	_ ukase.Option = Dec(nil)
)

type Dec func(*ukdec.Config)

func (o Dec) UkaseApplyDec(c *ukdec.Config) { o(c) }
func (o Dec) UkaseApplyCLI(c *ukcli.Config) { c.Decode = append(c.Decode, o) }
func (o Dec) UkaseApplyApp(c *ukase.Config) { c.CLI = append(c.CLI, o) }

// =============================================================================
// Specific
// =============================================================================
