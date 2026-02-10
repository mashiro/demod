package demod

import (
	"context"
	"hash/fnv"
	"log/slog"
)

// ANSI color codes for module name coloring.
var moduleColors = []string{
	"\033[31m", // Red
	"\033[32m", // Green
	"\033[33m", // Yellow
	"\033[34m", // Blue
	"\033[35m", // Magenta
	"\033[36m", // Cyan
	"\033[91m", // Bright Red
	"\033[92m", // Bright Green
	"\033[93m", // Bright Yellow
	"\033[94m", // Bright Blue
	"\033[95m", // Bright Magenta
	"\033[96m", // Bright Cyan
}

func colorize(s, color string, noColor bool) string {
	if noColor {
		return s
	}
	return color + s + "\033[0m"
}

func moduleColor(name string) string {
	h := fnv.New32a()
	h.Write([]byte(name))
	return moduleColors[h.Sum32()%uint32(len(moduleColors))]
}

type moduleHandler struct {
	inner   slog.Handler
	module  string
	noColor bool
}

func NewModuleHandler(inner slog.Handler, noColor bool) slog.Handler {
	return &moduleHandler{inner: inner, noColor: noColor}
}

// WithModule returns a new logger with the module name set.
// If the logger uses a moduleHandler, the module name is displayed as a prefix (e.g. "[name] msg").
// Otherwise, it falls back to logger.With("module", name).
func WithModule(logger *slog.Logger, name string) *slog.Logger {
	if mh, ok := logger.Handler().(*moduleHandler); ok {
		return slog.New(&moduleHandler{inner: mh.inner, module: name, noColor: mh.noColor})
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
	prefix := colorize("["+h.module+"]", moduleColor(h.module), h.noColor)
	nr := slog.NewRecord(r.Time, r.Level, prefix+" "+r.Message, r.PC)
	r.Attrs(func(a slog.Attr) bool {
		nr.AddAttrs(a)
		return true
	})
	return h.inner.Handle(ctx, nr)
}

func (h *moduleHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &moduleHandler{inner: h.inner.WithAttrs(attrs), module: h.module, noColor: h.noColor}
}

func (h *moduleHandler) WithGroup(name string) slog.Handler {
	return &moduleHandler{inner: h.inner.WithGroup(name), module: h.module, noColor: h.noColor}
}
