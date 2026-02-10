package demod

import (
	"bytes"
	"encoding/json"
	"log/slog"
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
