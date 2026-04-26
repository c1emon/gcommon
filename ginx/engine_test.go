package ginx

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestEngineBuilderMiddlewareOrder(t *testing.T) {
	gin.SetMode(gin.TestMode)
	order := make([]string, 0, 4)

	m1 := func(c *gin.Context) {
		order = append(order, "m1-before")
		c.Next()
		order = append(order, "m1-after")
	}
	m2 := func(c *gin.Context) {
		order = append(order, "m2-before")
		c.Next()
		order = append(order, "m2-after")
	}

	eng := NewEngineBuilder().Use(m1, m2).Build()
	eng.GET("/", func(c *gin.Context) {
		order = append(order, "handler")
		c.Status(http.StatusNoContent)
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	eng.ServeHTTP(rec, req)

	want := []string{"m1-before", "m2-before", "handler", "m2-after", "m1-after"}
	if len(order) != len(want) {
		t.Fatalf("unexpected order length: got %d want %d (%v)", len(order), len(want), order)
	}
	for i := range want {
		if order[i] != want[i] {
			t.Fatalf("unexpected order at %d: got %s want %s (%v)", i, order[i], want[i], order)
		}
	}
}

func TestNewDefaultEngineIncludesErrorResponder(t *testing.T) {
	gin.SetMode(gin.TestMode)
	eng := NewDefaultEngine(DefaultEngineConfig{
		Logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
	})
	eng.GET("/err", func(c *gin.Context) {
		_ = c.Error(assertErr{})
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/err", nil)
	eng.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

type assertErr struct{}

func (assertErr) Error() string { return "assert error" }
