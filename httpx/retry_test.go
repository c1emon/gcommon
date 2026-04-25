package httpx

import (
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func TestRetry_retries5xx(t *testing.T) {
	var n atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		c := n.Add(1)
		if c < 3 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}))
	t.Cleanup(srv.Close)

	m := NewManager()
	c := m.Register("c",
		WithBaseURL(srv.URL),
		WithRetry(RetryPolicy{
			Enabled:    true,
			MaxRetries: 5,
			MinBackoff: time.Millisecond,
			MaxBackoff: 10 * time.Millisecond,
		}),
	)

	resp, err := c.R().Get("/")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("want 200, got %d", resp.StatusCode)
	}
	if n.Load() != 3 {
		t.Fatalf("want 3 server hits, got %d", n.Load())
	}
}
