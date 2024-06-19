package internal

import (
	"context"
	"log/slog"
)

var LogDiscard = slog.New(logHandlerDiscard{})

type logHandlerDiscard struct{}

func (logHandlerDiscard) Enabled(context.Context, slog.Level) bool  { return false }
func (logHandlerDiscard) Handle(context.Context, slog.Record) error { return nil }
func (lhd logHandlerDiscard) WithAttrs([]slog.Attr) slog.Handler    { return lhd }
func (lhd logHandlerDiscard) WithGroup(string) slog.Handler         { return lhd }
