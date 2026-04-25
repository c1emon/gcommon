package httpx

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"golang.org/x/time/rate"
)

func TestLimiter_globalAndClientAllowRequests(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)

	gl := rate.NewLimiter(rate.Every(time.Millisecond), 32)
	cl := rate.NewLimiter(rate.Every(time.Millisecond), 32)

	m := NewManager(WithGlobalLimiter(gl))
	c := m.Register("c", WithBaseURL(srv.URL), WithLimiter(cl))

	for range 5 {
		_, err := c.R().Get("/")
		if err != nil {
			t.Fatal(err)
		}
	}
}
