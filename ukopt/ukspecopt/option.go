package ukspecopt

import (
	"strconv"

	"github.com/oligarch316/go-ukase"
	"github.com/oligarch316/go-ukase/ukreflect/ukenc"
	"github.com/oligarch316/go-ukase/ukreflect/ukinit"
	"github.com/oligarch316/go-ukase/ukspec"
)

type ElideBoolType bool

func (o ElideBoolType) UkaseApplySpec(c *ukspec.Config) { c.ElideBoolType = bool(o) }
func (o ElideBoolType) UkaseApplyEnc(c *ukenc.Config)   { c.Spec = append(c.Spec, o) }
func (o ElideBoolType) UkaseApplyInit(c *ukinit.Config) { c.Spec = append(c.Spec, o) }
func (o ElideBoolType) UkaseApply(c *ukase.Config) {
	c.Enc = append(c.Enc, o)
	c.Init = append(c.Init, o)
	c.Spec = append(c.Spec, o)
}

type ElideIsBoolFlag bool

func (o ElideIsBoolFlag) UkaseApplySpec(c *ukspec.Config) { c.ElideIsBoolFlag = bool(o) }
func (o ElideIsBoolFlag) UkaseApplyEnc(c *ukenc.Config)   { c.Spec = append(c.Spec, o) }
func (o ElideIsBoolFlag) UkaseApplyInit(c *ukinit.Config) { c.Spec = append(c.Spec, o) }
func (o ElideIsBoolFlag) UkaseApply(c *ukase.Config) {
	c.Enc = append(c.Enc, o)
	c.Init = append(c.Init, o)
	c.Spec = append(c.Spec, o)
}

type ElideDefaultConsumable func(string) bool

func (o ElideDefaultConsumable) UkaseApplySpec(c *ukspec.Config) { c.ElideDefaultConsumable = o }
func (o ElideDefaultConsumable) UkaseApplyEnc(c *ukenc.Config)   { c.Spec = append(c.Spec, o) }
func (o ElideDefaultConsumable) UkaseApplyInit(c *ukinit.Config) { c.Spec = append(c.Spec, o) }
func (o ElideDefaultConsumable) UkaseApply(c *ukase.Config) {
	c.Enc = append(c.Enc, o)
	c.Init = append(c.Init, o)
	c.Spec = append(c.Spec, o)
}

func ElideDefaultConsumableSet(valid ...string) ElideDefaultConsumable {
	return ukspec.ConsumableSet(valid...)
}

var ElideDefaultMinimal = ElideDefaultConsumableSet("true", "false")
var ElideDefaultParsable = ElideDefaultConsumable(elideParsable)

func elideParsable(text string) bool {
	_, err := strconv.ParseBool(text)
	return err == nil
}
