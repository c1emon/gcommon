package logx

import (
	"bytes"
	"log/slog"
	"strings"
	"testing"
)

func TestInitAndDefault(t *testing.T) {
	var out bytes.Buffer
	previous := slog.Default()
	t.Cleanup(func() {
		slog.SetDefault(previous)
		defaultMu.Lock()
		defaultLogger = nil
		defaultMu.Unlock()
	})

	Init(Config{
		Format: FormatJSON,
		Output: &out,
	})

	Default().Info("default log", "key", "value")

	logLine := out.String()
	if !strings.Contains(logLine, `"msg":"default log"`) {
		t.Fatalf("expected message in output, got: %s", logLine)
	}
	if !strings.Contains(logLine, `"key":"value"`) {
		t.Fatalf("expected attr in output, got: %s", logLine)
	}
}
