package demod

import (
	"context"
	"fmt"
	"log/slog"
)

type moduleHandler struct {
	inner  slog.Handler
	module string
}

func NewModuleHandler(inner slog.Handler) slog.Handler {
	return &moduleHandler{inner: inner}
}

// WithModule returns a new logger with the module name set.
// If the logger uses a moduleHandler, the module name is displayed as a prefix (e.g. "[name] msg").
// Otherwise, it falls back to logger.With("module", name).
func WithModule(logger *slog.Logger, name string) *slog.Logger {
	if mh, ok := logger.Handler().(*moduleHandler); ok {
		return slog.New(&moduleHandler{inner: mh.inner, module: name})
	}
	return logger.With("module", name)
}

func (h *moduleHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.inner.Enabled(ctx, level)
}

func (h *moduleHandler) Handle(ctx context.Context, r slog.Record) error {
	if h.module == "" {
		return h.inner.Handle(ctx, r)
	}
	nr := slog.NewRecord(r.Time, r.Level, fmt.Sprintf("[%s] %s", h.module, r.Message), r.PC)
	r.Attrs(func(a slog.Attr) bool {
		nr.AddAttrs(a)
		return true
	})
	return h.inner.Handle(ctx, nr)
}

func (h *moduleHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &moduleHandler{inner: h.inner.WithAttrs(attrs), module: h.module}
}

func (h *moduleHandler) WithGroup(name string) slog.Handler {
	return &moduleHandler{inner: h.inner.WithGroup(name), module: h.module}
}
