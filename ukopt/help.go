package ukopt

import (
	"github.com/oligarch316/go-ukase"
	"github.com/oligarch316/go-ukase/ukmeta"
	"github.com/oligarch316/go-ukase/ukmeta/ukhelp"
)

// =============================================================================
// General
// =============================================================================

var (
	_ ukhelp.Option = Help(nil)
	_ ukase.Option  = Help(nil)
)

type Help func(*ukhelp.Config)

func (o Help) UkaseApplyHelp(c *ukhelp.Config) { o(c) }
func (o Help) UkaseApplyApp(c *ukase.Config)   { c.Help = append(c.Help, o) }

// =============================================================================
// Specific
// =============================================================================

func HelpInfo(info any) Help {
	return func(c *ukhelp.Config) { c.Info = info }
}

func HelpEncode(encode func(in ukmeta.Input) (any, error)) Help {
	return func(c *ukhelp.Config) { c.Encode = encode }
}
