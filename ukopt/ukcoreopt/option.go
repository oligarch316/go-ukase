package ukcoreopt

import (
	"github.com/oligarch316/go-ukase"
	"github.com/oligarch316/go-ukase/ukcore"
	"github.com/oligarch316/go-ukase/ukopt"
)

func ExecDefault(exec ukcore.Exec) ukopt.Core {
	return func(c *ukcore.Config) { c.ExecDefault = exec }
}

type MuxOverwrite bool

func (o MuxOverwrite) UkaseApplyCore(c *ukcore.Config) { c.MuxOverwrite = bool(o) }
func (o MuxOverwrite) UkaseApply(c *ukase.Config)      { c.Core = append(c.Core, o) }

type FlagCheck ukcore.FlagCheckLevel

func (o FlagCheck) UkaseApplyCore(c *ukcore.Config) { c.FlagCheck = ukcore.FlagCheckLevel(o) }
func (o FlagCheck) UkaseApply(c *ukase.Config)      { c.Core = append(c.Core, o) }
