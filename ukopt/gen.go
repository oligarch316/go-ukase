package ukopt

import (
	"github.com/oligarch316/go-ukase"
	"github.com/oligarch316/go-ukase/ukmeta/ukgen"
)

// =============================================================================
// General
// =============================================================================

var (
	_ ukgen.Option = Gen(nil)
	_ ukase.Option = Gen(nil)
)

type Gen func(*ukgen.Config)

func (o Gen) UkaseApplyGen(c *ukgen.Config) { o(c) }
func (o Gen) UkaseApplyApp(c *ukase.Config) { c.Gen = append(c.Gen, o) }

// =============================================================================
// Specific
// =============================================================================
