package ukinitopt

import (
	"github.com/oligarch316/go-ukase"
	"github.com/oligarch316/go-ukase/ukreflect/ukinit"
)

type ForceCustomInit bool

func (o ForceCustomInit) UkaseApplyInit(c *ukinit.Config) { c.ForceCustomInit = bool(o) }
func (o ForceCustomInit) UkaseApply(c *ukase.Config)      { c.Init = append(c.Init, o) }
