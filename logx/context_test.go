package logx

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"testing"
)

func TestFromContext(t *testing.T) {
	var out bytes.Buffer
	previous := slog.Default()
	slog.SetDefault(NewLogger(Config{
		Format: FormatJSON,
		Output: &out,
	}))
	t.Cleanup(func() {
		slog.SetDefault(previous)
	})

	ctx := WithAttrs(context.Background(), TraceID("trace-1"), UserID("user-1"))
	FromContext(ctx).InfoContext(ctx, "ctx log")

	logLine := out.String()
	if !strings.Contains(logLine, `"msg":"ctx log"`) {
		t.Fatalf("expected message in output, got: %s", logLine)
	}
	if !strings.Contains(logLine, `"trace_id":"trace-1"`) {
		t.Fatalf("expected trace_id in output, got: %s", logLine)
	}
	if !strings.Contains(logLine, `"user_id":"user-1"`) {
		t.Fatalf("expected user_id in output, got: %s", logLine)
	}
}

func TestDefaultLoggerInjectsContextAttrs(t *testing.T) {
	var out bytes.Buffer
	previous := slog.Default()
	slog.SetDefault(NewLogger(Config{
		Format: FormatJSON,
		Output: &out,
	}))
	t.Cleanup(func() {
		slog.SetDefault(previous)
	})

	ctx := WithAttrs(context.Background(), TraceID("trace-2"))
	Default().InfoContext(ctx, "direct log")

	logLine := out.String()
	if !strings.Contains(logLine, `"msg":"direct log"`) {
		t.Fatalf("expected message in output, got: %s", logLine)
	}
	if !strings.Contains(logLine, `"trace_id":"trace-2"`) {
		t.Fatalf("expected trace_id in output, got: %s", logLine)
	}
}
