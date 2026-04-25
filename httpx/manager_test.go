package httpx

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestManager_globalHeaderAndOverride(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(r.Header.Get("X-Trace")))
	}))
	t.Cleanup(srv.Close)

	m := NewManager(
		WithGlobalHeader("X-Trace", "global"),
	)
	c := m.Register("c", WithBaseURL(srv.URL), WithHeader("X-Trace", "client"))
	resp, err := c.R().Get("/")
	if err != nil {
		t.Fatal(err)
	}
	b, _ := io.ReadAll(resp.Body)
	if string(b) != "client" {
		t.Fatalf("want client header to override global, got %q", b)
	}
}

func TestDefaultManager_initAndGet(t *testing.T) {
	defaultManagerMu.Lock()
	prev := defaultManager
	defaultManagerMu.Unlock()
	t.Cleanup(func() {
		defaultManagerMu.Lock()
		defaultManager = prev
		defaultManagerMu.Unlock()
	})

	InitDefaultManager(WithGlobalHeader("X-Default-Test", "1"))
	got := GetDefaultManager()
	if got == nil {
		t.Fatal("GetDefaultManager returned nil after InitDefaultManager")
	}
	got.Register("smoke", WithBaseURL("https://example.com"))
	if _, ok := got.Client("smoke"); !ok {
		t.Fatal("expected registered client on default manager")
	}

	InitDefaultManager()
	if got2 := GetDefaultManager(); got2 == got {
		t.Fatal("expected InitDefaultManager to replace default manager instance")
	}
}
