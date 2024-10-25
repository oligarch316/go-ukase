package ukopt

import (
	"github.com/oligarch316/ukase"
	"github.com/oligarch316/ukase/ukcli"
	"github.com/oligarch316/ukase/ukcore/ukdec"
	"github.com/oligarch316/ukase/ukcore/ukinit"
	"github.com/oligarch316/ukase/ukcore/ukspec"
	"github.com/oligarch316/ukase/ukmeta/ukgen"
)

// =============================================================================
// General
// =============================================================================

var (
	_ ukspec.Option = Spec(nil)
	_ ukdec.Option  = Spec(nil)
	_ ukgen.Option  = Spec(nil)
	_ ukinit.Option = Spec(nil)
	_ ukcli.Option  = Spec(nil)
	_ ukase.Option  = Spec(nil)
)

type Spec func(*ukspec.Config)

func (o Spec) UkaseApplySpec(c *ukspec.Config) { o(c) }
func (o Spec) UkaseApplyDec(c *ukdec.Config)   { c.Spec = append(c.Spec, o) }
func (o Spec) UkaseApplyGen(c *ukgen.Config)   { c.Spec = append(c.Spec, o) }
func (o Spec) UkaseApplyInit(c *ukinit.Config) { c.Spec = append(c.Spec, o) }
func (o Spec) UkaseApplyApp(c *ukase.Config)   { c.CLI = append(c.CLI, o) }

func (o Spec) UkaseApplyCLI(c *ukcli.Config) {
	c.Spec = append(c.Spec, o)
	c.Decode = append(c.Decode, o)
	c.Init = append(c.Init, o)
}

// =============================================================================
// Specific
// =============================================================================

// TODO:

// func SpecElideBoolType(allow bool) Spec {
// 	return func(c *ukspec.Config) { c.ElideAllowBoolType = allow }
// }

// func SpecElideIsBoolFlag(allow bool) Spec {
// 	return func(c *ukspec.Config) { c.ElideAllowIsBoolFlag = allow }
// }

// func SpecElideConsumable(consumable func(string) bool) Spec {
// 	return func(c *ukspec.Config) { c.ElideConsumable = consumable }
// }

// func SpecElideConsumableSet(valid ...string) Spec {
// 	consumable := internal.ConsumableSet(valid...)
// 	return SpecElideConsumable(consumable)
// }
