package ginx

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/c1emon/gcommon/v2/errorx"
	"github.com/gin-gonic/gin"
)

type contextCaptureHandler struct {
	mu    sync.Mutex
	seen  any
	attrs map[string]any
}

func (h *contextCaptureHandler) Enabled(context.Context, slog.Level) bool { return true }
func (h *contextCaptureHandler) Handle(ctx context.Context, rec slog.Record) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.seen = ctx.Value("req-key")
	h.attrs = map[string]any{}
	rec.Attrs(func(a slog.Attr) bool {
		h.attrs[a.Key] = attrValue(a.Value)
		return true
	})
	return nil
}
func (h *contextCaptureHandler) WithAttrs([]slog.Attr) slog.Handler { return h }
func (h *contextCaptureHandler) WithGroup(string) slog.Handler      { return h }

func attrValue(v slog.Value) any {
	switch v.Kind() {
	case slog.KindString:
		return v.String()
	case slog.KindInt64:
		return v.Int64()
	default:
		return v.Any()
	}
}

func TestErrorResponderHttpErrorData(t *testing.T) {
	gin.SetMode(gin.TestMode)
	eng := NewBareEngine()
	eng.Use(ErrorResponder())
	eng.GET("/data", func(c *gin.Context) {
		_ = c.Error(errorx.NewHttpError(http.StatusTeapot, 777, "teapot", map[string]any{
			"kind": "brew",
		}))
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/data", nil)
	eng.ServeHTTP(rec, req)

	if rec.Code != http.StatusTeapot {
		t.Fatalf("expected status %d, got %d", http.StatusTeapot, rec.Code)
	}

	body, err := io.ReadAll(rec.Body)
	if err != nil {
		t.Fatalf("read response body: %v", err)
	}
	var payload map[string]any
	if err := json.Unmarshal(body, &payload); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if int(payload["code"].(float64)) != 777 {
		t.Fatalf("expected code 777, got %v", payload["code"])
	}
	data := payload["data"].(map[string]any)
	if data["kind"] != "brew" {
		t.Fatalf("expected data.kind=brew, got %v", data["kind"])
	}
}

func TestRecoveryWritesInternalError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	eng := NewBareEngine()
	lg := slog.New(slog.NewTextHandler(io.Discard, nil))
	eng.Use(ErrorResponder(), Recovery(lg))
	eng.GET("/panic", func(c *gin.Context) {
		panic("boom")
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/panic", nil)
	eng.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, rec.Code)
	}
}

func TestLoggerUsesRequestContext(t *testing.T) {
	gin.SetMode(gin.TestMode)
	capture := &contextCaptureHandler{}
	logger := slog.New(capture)

	eng := NewBareEngine()
	eng.Use(Logger(logger))
	eng.GET("/ok", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/ok", nil)
	req = req.WithContext(context.WithValue(req.Context(), "req-key", "req-value"))
	eng.ServeHTTP(rec, req)

	if capture.seen != "req-value" {
		t.Fatalf("expected logger context value req-value, got %v", capture.seen)
	}
}

func TestLoggerDetailAttrsByLevel(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tests := []struct {
		name            string
		level           slog.Level
		expectDetailKey bool
	}{
		{name: "info-no-detail", level: slog.LevelInfo, expectDetailKey: false},
		{name: "debug-with-detail", level: slog.LevelDebug, expectDetailKey: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			capture := &contextCaptureHandler{}
			logger := slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{
				Level: tt.level,
			}))
			logger = slog.New(&levelAwareCapture{
				level:    tt.level,
				delegate: capture,
			})

			eng := NewBareEngine()
			eng.Use(Logger(logger))
			eng.POST("/ok", func(c *gin.Context) {
				c.Header("X-Resp-Test", "resp-ok")
				body, _ := io.ReadAll(c.Request.Body)
				c.String(http.StatusOK, "echo:"+string(body))
			})

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/ok", strings.NewReader("hello-debug"))
			req.Header.Set("X-Req-Test", "req-ok")
			eng.ServeHTTP(rec, req)

			_, hasDetail := capture.attrs["user_agent"]
			if hasDetail != tt.expectDetailKey {
				t.Fatalf("detail attrs mismatch, got has user_agent=%v", hasDetail)
			}
			_, hasReqBody := capture.attrs["request_body"]
			if hasReqBody != tt.expectDetailKey {
				t.Fatalf("request body detail mismatch, got has request_body=%v", hasReqBody)
			}
			_, hasRespBody := capture.attrs["response_body"]
			if hasRespBody != tt.expectDetailKey {
				t.Fatalf("response body detail mismatch, got has response_body=%v", hasRespBody)
			}
			_, hasReqHeaders := capture.attrs["request_headers"]
			if hasReqHeaders != tt.expectDetailKey {
				t.Fatalf("request header detail mismatch, got has request_headers=%v", hasReqHeaders)
			}
			_, hasRespHeaders := capture.attrs["response_headers"]
			if hasRespHeaders != tt.expectDetailKey {
				t.Fatalf("response header detail mismatch, got has response_headers=%v", hasRespHeaders)
			}
			if tt.expectDetailKey {
				if capture.attrs["request_body"] != "hello-debug" {
					t.Fatalf("unexpected request_body: %v", capture.attrs["request_body"])
				}
				if capture.attrs["response_body"] != "echo:hello-debug" {
					t.Fatalf("unexpected response_body: %v", capture.attrs["response_body"])
				}
				reqHeaders, ok := capture.attrs["request_headers"].(map[string][]string)
				if !ok {
					t.Fatalf("request_headers missing or type mismatch: %T", capture.attrs["request_headers"])
				}
				if got := reqHeaders["X-Req-Test"]; len(got) != 1 || got[0] != "req-ok" {
					t.Fatalf("unexpected request header value: %v", got)
				}
				respHeaders, ok := capture.attrs["response_headers"].(map[string][]string)
				if !ok {
					t.Fatalf("response_headers missing or type mismatch: %T", capture.attrs["response_headers"])
				}
				if got := respHeaders["X-Resp-Test"]; len(got) != 1 || got[0] != "resp-ok" {
					t.Fatalf("unexpected response header value: %v", got)
				}
			}
		})
	}
}

func TestLoggerBodyTruncationAtDebug(t *testing.T) {
	gin.SetMode(gin.TestMode)
	capture := &contextCaptureHandler{}
	logger := slog.New(&levelAwareCapture{
		level:    slog.LevelDebug,
		delegate: capture,
	})

	eng := NewBareEngine()
	eng.Use(Logger(logger))
	eng.POST("/truncate", func(c *gin.Context) {
		body, _ := io.ReadAll(c.Request.Body)
		c.String(http.StatusOK, "resp:"+string(body))
	})

	longReq := strings.Repeat("a", maxLoggedBodyBytes+128)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/truncate", strings.NewReader(longReq))
	eng.ServeHTTP(rec, req)

	reqBody, ok := capture.attrs["request_body"].(string)
	if !ok {
		t.Fatalf("request_body missing or not string: %v", capture.attrs["request_body"])
	}
	if len(reqBody) != maxLoggedBodyBytes {
		t.Fatalf("unexpected request_body length: got %d want %d", len(reqBody), maxLoggedBodyBytes)
	}
	reqTruncated, ok := capture.attrs["request_body_truncated"].(bool)
	if !ok || !reqTruncated {
		t.Fatalf("expected request_body_truncated=true, got %v", capture.attrs["request_body_truncated"])
	}

	respBody, ok := capture.attrs["response_body"].(string)
	if !ok {
		t.Fatalf("response_body missing or not string: %v", capture.attrs["response_body"])
	}
	if len(respBody) != maxLoggedBodyBytes {
		t.Fatalf("unexpected response_body length: got %d want %d", len(respBody), maxLoggedBodyBytes)
	}
	respTruncated, ok := capture.attrs["response_body_truncated"].(bool)
	if !ok || !respTruncated {
		t.Fatalf("expected response_body_truncated=true, got %v", capture.attrs["response_body_truncated"])
	}
}

type levelAwareCapture struct {
	level    slog.Level
	delegate *contextCaptureHandler
}

func (h *levelAwareCapture) Enabled(_ context.Context, l slog.Level) bool { return l >= h.level }
func (h *levelAwareCapture) Handle(ctx context.Context, rec slog.Record) error {
	return h.delegate.Handle(ctx, rec)
}
func (h *levelAwareCapture) WithAttrs(_ []slog.Attr) slog.Handler { return h }
func (h *levelAwareCapture) WithGroup(_ string) slog.Handler      { return h }
