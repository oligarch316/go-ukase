package ukopt

import (
	"github.com/oligarch316/go-ukase"
	"github.com/oligarch316/go-ukase/ukcore"
	"github.com/oligarch316/go-ukase/ukreflect/ukenc"
	"github.com/oligarch316/go-ukase/ukreflect/ukinit"
	"github.com/oligarch316/go-ukase/ukspec"
)

var (
	// Valid `ukase` options
	_ ukase.Option = Core(nil)
	_ ukase.Option = Enc(nil)
	_ ukase.Option = Init(nil)
	_ ukase.Option = Spec(nil)

	// Valid `ukcore` options
	_ ukcore.Option = Core(nil)

	// Valid `ukenc` options
	_ ukenc.Option = Enc(nil)
	_ ukenc.Option = Spec(nil)

	// Valid `ukinit` options
	_ ukinit.Option = Init(nil)
	_ ukinit.Option = Spec(nil)

	// Valie `ukspec` options
	_ ukspec.Option = Spec(nil)
)

type (
	Core func(*ukcore.Config)
	Enc  func(*ukenc.Config)
	Init func(*ukinit.Config)
	Spec func(*ukspec.Config)
)

func (o Core) UkaseApplyCore(c *ukcore.Config) { o(c) }
func (o Core) UkaseApply(c *ukase.Config)      { c.Core = append(c.Core, o) }

func (o Enc) UkaseApplyEnc(c *ukenc.Config) { o(c) }
func (o Enc) UkaseApply(c *ukase.Config)    { c.Enc = append(c.Enc, o) }

func (o Init) UkaseApplyInit(c *ukinit.Config) { o(c) }
func (o Init) UkaseApply(c *ukase.Config)      { c.Init = append(c.Init, o) }

func (o Spec) UkaseApplySpec(c *ukspec.Config) { o(c) }
func (o Spec) UkaseApplyEnc(c *ukenc.Config)   { c.Spec = append(c.Spec, o) }
func (o Spec) UkaseApplyInit(c *ukinit.Config) { c.Spec = append(c.Spec, o) }
func (o Spec) UkaseApply(c *ukase.Config) {
	c.Enc = append(c.Enc, o)
	c.Init = append(c.Init, o)
}
