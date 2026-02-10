package demod

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"strings"
	"testing"
)

type logEntry struct {
	Level  string `json:"level"`
	Msg    string `json:"msg"`
	Module string `json:"module,omitempty"`
	Key    string `json:"key,omitempty"`
}

func newTestLogger(buf *bytes.Buffer) *slog.Logger {
	return slog.New(NewModuleHandler(
		slog.NewJSONHandler(buf, &slog.HandlerOptions{Level: slog.LevelDebug}),
		true,
	))
}

func parseLogEntry(t *testing.T, buf *bytes.Buffer) logEntry {
	t.Helper()
	var entry logEntry
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("failed to parse log entry: %v\nbody: %s", err, buf.String())
	}
	return entry
}

func TestModuleHandler_WithoutModule(t *testing.T) {
	var buf bytes.Buffer
	logger := newTestLogger(&buf)

	logger.Info("hello", "key", "value")

	entry := parseLogEntry(t, &buf)
	if entry.Msg != "hello" {
		t.Errorf("msg = %q, want %q", entry.Msg, "hello")
	}
	if entry.Key != "value" {
		t.Errorf("key = %q, want %q", entry.Key, "value")
	}
}

func TestModuleHandler_WithModule(t *testing.T) {
	var buf bytes.Buffer
	logger := newTestLogger(&buf)
	logger = WithModule(logger, "mymod")

	logger.Info("cloning")

	entry := parseLogEntry(t, &buf)
	if entry.Msg != "[mymod] cloning" {
		t.Errorf("msg = %q, want %q", entry.Msg, "[mymod] cloning")
	}
}

func TestModuleHandler_WithModuleAndAttrs(t *testing.T) {
	var buf bytes.Buffer
	logger := newTestLogger(&buf)
	logger = WithModule(logger, "mymod")

	logger.Info("checkout", "key", "value")

	entry := parseLogEntry(t, &buf)
	if entry.Msg != "[mymod] checkout" {
		t.Errorf("msg = %q, want %q", entry.Msg, "[mymod] checkout")
	}
	if entry.Key != "value" {
		t.Errorf("key = %q, want %q", entry.Key, "value")
	}
}

func TestWithModule_Fallback(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))
	logger = WithModule(logger, "mymod")

	logger.Info("hello")

	entry := parseLogEntry(t, &buf)
	if entry.Msg != "hello" {
		t.Errorf("msg = %q, want %q", entry.Msg, "hello")
	}
	if entry.Module != "mymod" {
		t.Errorf("module = %q, want %q", entry.Module, "mymod")
	}
}

func TestModuleHandler_DebugLevel(t *testing.T) {
	var buf bytes.Buffer
	logger := newTestLogger(&buf)
	logger = WithModule(logger, "mymod")

	logger.Debug("exec", "key", "value")

	entry := parseLogEntry(t, &buf)
	if entry.Msg != "[mymod] exec" {
		t.Errorf("msg = %q, want %q", entry.Msg, "[mymod] exec")
	}
}

func TestColorize(t *testing.T) {
	t.Run("noColor", func(t *testing.T) {
		got := colorize("hello", "\033[31m", true)
		if got != "hello" {
			t.Errorf("got %q, want %q", got, "hello")
		}
	})
	t.Run("color", func(t *testing.T) {
		got := colorize("hello", "\033[31m", false)
		want := "\033[31mhello\033[0m"
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})
}

func TestModuleColor_Deterministic(t *testing.T) {
	c1 := moduleColor("foo")
	c2 := moduleColor("foo")
	if c1 != c2 {
		t.Errorf("same name produced different colors: %q vs %q", c1, c2)
	}
}

func TestModuleColor_DifferentNames(t *testing.T) {
	colors := make(map[string]bool)
	names := []string{"aaa", "bbb", "ccc", "ddd", "eee", "fff", "ggg", "hhh"}
	for _, name := range names {
		colors[moduleColor(name)] = true
	}
	if len(colors) < 2 {
		t.Errorf("expected multiple distinct colors, got %d", len(colors))
	}
}

func TestModuleHandler_WithColor(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(NewModuleHandler(
		slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}),
		false,
	))
	logger = WithModule(logger, "mymod")

	logger.Info("cloning")

	entry := parseLogEntry(t, &buf)
	if !strings.Contains(entry.Msg, "[mymod]") {
		t.Errorf("msg should contain [mymod], got %q", entry.Msg)
	}
	if !strings.Contains(entry.Msg, "\033[") {
		t.Errorf("msg should contain ANSI escape, got %q", entry.Msg)
	}
	if !strings.Contains(entry.Msg, "\033[0m") {
		t.Errorf("msg should contain ANSI reset, got %q", entry.Msg)
	}
	if !strings.HasSuffix(entry.Msg, "cloning") {
		t.Errorf("msg should end with message text, got %q", entry.Msg)
	}
}
