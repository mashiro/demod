package demod

import (
	"context"
	"fmt"
	"log/slog"
)

type moduleHandler struct {
	inner slog.Handler
}

func NewModuleHandler(inner slog.Handler) slog.Handler {
	return &moduleHandler{inner: inner}
}

func (h *moduleHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.inner.Enabled(ctx, level)
}

func (h *moduleHandler) Handle(ctx context.Context, r slog.Record) error {
	var module string
	var kept []slog.Attr
	r.Attrs(func(a slog.Attr) bool {
		if a.Key == "module" {
			module = a.Value.String()
		} else {
			kept = append(kept, a)
		}
		return true
	})

	nr := slog.NewRecord(r.Time, r.Level, r.Message, r.PC)
	if module != "" {
		nr.Message = fmt.Sprintf("[%s] %s", module, r.Message)
	}
	nr.AddAttrs(kept...)
	return h.inner.Handle(ctx, nr)
}

func (h *moduleHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &moduleHandler{inner: h.inner.WithAttrs(attrs)}
}

func (h *moduleHandler) WithGroup(name string) slog.Handler {
	return &moduleHandler{inner: h.inner.WithGroup(name)}
}
