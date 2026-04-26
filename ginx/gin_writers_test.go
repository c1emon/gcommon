package ginx

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestSetGinSlogWriters_emitsPerLine(t *testing.T) {
	oldOut := gin.DefaultWriter
	oldErr := gin.DefaultErrorWriter
	t.Cleanup(func() {
		gin.DefaultWriter = oldOut
		gin.DefaultErrorWriter = oldErr
	})

	var buf bytes.Buffer
	log := slog.New(slog.NewJSONHandler(&buf, nil))
	SetGinSlogWriters(log)

	_, _ = io.WriteString(gin.DefaultWriter, "out-a\nout-b")
	_, _ = io.WriteString(gin.DefaultWriter, "\n")
	_, _ = io.WriteString(gin.DefaultErrorWriter, "err-line\n")

	dec := json.NewDecoder(&buf)
	var rec1, rec2, rec3 map[string]any
	if err := dec.Decode(&rec1); err != nil {
		t.Fatalf("decode 1: %v", err)
	}
	if err := dec.Decode(&rec2); err != nil {
		t.Fatalf("decode 2: %v", err)
	}
	if err := dec.Decode(&rec3); err != nil {
		t.Fatalf("decode 3: %v", err)
	}
	if got := rec1["message"]; got != "out-a" {
		t.Fatalf("rec1 message: got %v want out-a", got)
	}
	if got := rec2["message"]; got != "out-b" {
		t.Fatalf("rec2 message: got %v want out-b", got)
	}
	if got := rec3["message"]; got != "err-line" {
		t.Fatalf("rec3 message: got %v want err-line", got)
	}
	if rec1["level"] != "INFO" {
		t.Fatalf("expected INFO for DefaultWriter line, got %v", rec1["level"])
	}
	if rec3["level"] != "ERROR" {
		t.Fatalf("expected ERROR for DefaultErrorWriter line, got %v", rec3["level"])
	}
}

func TestNewDefaultEngine_setsGinWriters(t *testing.T) {
	oldOut := gin.DefaultWriter
	oldErr := gin.DefaultErrorWriter
	t.Cleanup(func() {
		gin.DefaultWriter = oldOut
		gin.DefaultErrorWriter = oldErr
	})

	discard := slog.New(slog.NewTextHandler(io.Discard, nil))
	_ = NewDefaultEngine(DefaultEngineConfig{Logger: discard})

	if _, ok := gin.DefaultWriter.(*slogLineWriter); !ok {
		t.Fatalf("expected gin.DefaultWriter to be *slogLineWriter, got %T", gin.DefaultWriter)
	}
	if _, ok := gin.DefaultErrorWriter.(*slogLineWriter); !ok {
		t.Fatalf("expected gin.DefaultErrorWriter to be *slogLineWriter, got %T", gin.DefaultErrorWriter)
	}
}
