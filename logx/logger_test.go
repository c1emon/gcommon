package logx

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"testing"
)

func TestNewLoggerWithJSONAndReplaceAttr(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer
	logger := NewLogger(Config{
		Format:    FormatJSON,
		Output:    &out,
		Level:     slog.LevelInfo,
		AddSource: true,
		FixedAttrs: []slog.Attr{
			slog.String("service", "gcommon"),
		},
		ReplaceAttr: func(_ []string, a slog.Attr) slog.Attr {
			if a.Key == "token" {
				return slog.String("token", "***")
			}
			return a
		},
	})

	logger.InfoContext(context.Background(), "login", "token", "raw-token")

	logLine := out.String()
	if !strings.Contains(logLine, `"msg":"login"`) {
		t.Fatalf("expected message in output, got: %s", logLine)
	}
	if !strings.Contains(logLine, `"token":"***"`) {
		t.Fatalf("expected masked token in output, got: %s", logLine)
	}
	if !strings.Contains(logLine, `"service":"gcommon"`) {
		t.Fatalf("expected fixed attr in output, got: %s", logLine)
	}
	if !strings.Contains(logLine, `"source"`) {
		t.Fatalf("expected source field when AddSource is enabled, got: %s", logLine)
	}
}
