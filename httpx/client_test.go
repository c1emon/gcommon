package httpx_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/c1emon/gcommon/httpx"
)

func TestManagerRegister_clientIsolation(t *testing.T) {
	a := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("a"))
	}))
	b := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("b"))
	}))
	t.Cleanup(a.Close)
	t.Cleanup(b.Close)

	m := httpx.NewManager()
	ca := m.Register("a", httpx.WithBaseURL(a.URL))
	cb := m.Register("b", httpx.WithBaseURL(b.URL))

	ra, err := ca.R().Get("/")
	if err != nil {
		t.Fatal(err)
	}
	bodyA, _ := io.ReadAll(ra.Body)
	if string(bodyA) != "a" {
		t.Fatalf("client a: got %q", bodyA)
	}

	rb, err := cb.R().Get("/")
	if err != nil {
		t.Fatal(err)
	}
	bodyB, _ := io.ReadAll(rb.Body)
	if string(bodyB) != "b" {
		t.Fatalf("client b: got %q", bodyB)
	}
}
