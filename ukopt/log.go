package ukopt

import (
	"log/slog"

	"github.com/oligarch316/go-ukase"
	"github.com/oligarch316/go-ukase/ukcore"
	"github.com/oligarch316/go-ukase/ukhelpx"
	"github.com/oligarch316/go-ukase/ukreflect/ukenc"
	"github.com/oligarch316/go-ukase/ukreflect/ukinit"
	"github.com/oligarch316/go-ukase/ukspec"
)

const logKey = "ukase"

type Logger struct{ logger *slog.Logger }

func Log(l *slog.Logger) Logger { return Logger{l} }

func (l Logger) with(name string) *slog.Logger { return l.logger.With(logKey, name) }

func (l Logger) UkaseApply(c *ukase.Config) {
	c.Core = append(c.Core, l)
	c.Enc = append(c.Enc, l)
	c.Init = append(c.Init, l)
	c.Spec = append(c.Spec, l)
}

func (l Logger) UkaseApplyCore(c *ukcore.Config)  { c.Log = l.with("core") }
func (l Logger) UkaseApplyEnc(c *ukenc.Config)    { /* TODO */ }
func (l Logger) UkaseApplyHelp(c *ukhelpx.Config) { c.Log = l.with("help") }
func (l Logger) UkaseApplyInit(c *ukinit.Config)  { /* TODO */ }
func (l Logger) UkaseApplySpec(c *ukspec.Config)  { /* TODO */ }
