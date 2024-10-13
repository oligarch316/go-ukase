package ilog

import (
	"context"
	"log/slog"
)

var Discard = slog.New(handlerDiscard{})

type handlerDiscard struct{}

func (handlerDiscard) Enabled(context.Context, slog.Level) bool  { return false }
func (handlerDiscard) Handle(context.Context, slog.Record) error { return nil }
func (hd handlerDiscard) WithAttrs([]slog.Attr) slog.Handler     { return hd }
func (hd handlerDiscard) WithGroup(string) slog.Handler          { return hd }
