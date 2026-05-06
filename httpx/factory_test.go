package httpx

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClientFactory_profileHeaderAndInstanceOverride(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(r.Header.Get("X-Trace")))
	}))
	t.Cleanup(srv.Close)

	f := NewClientFactory(WithGlobalHeader("X-Trace", "global"))
	f.RegisterProfile("api", WithBaseURL(srv.URL), WithHeader("X-Trace", "profile"))

	c, ok := f.NewClient("api", WithHeader("X-Trace", "instance"))
	if !ok {
		t.Fatal("expected profile to exist")
	}
	resp, err := c.R().Get("/")
	if err != nil {
		t.Fatal(err)
	}
	b, _ := io.ReadAll(resp.Body)
	if string(b) != "instance" {
		t.Fatalf("want instance header to override profile/global, got %q", b)
	}
}

func TestClientFactory_newClientReturnsIndependentInstances(t *testing.T) {
	a := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("a"))
	}))
	t.Cleanup(a.Close)

	f := NewClientFactory()
	f.RegisterProfile("api", WithBaseURL(a.URL))

	first := f.MustNewClient("api")
	second := f.MustNewClient("api")
	if first == second {
		t.Fatal("MustNewClient returned the same wrapper instance twice")
	}
	if first.Client == second.Client {
		t.Fatal("MustNewClient returned wrappers around the same req client")
	}
}

func TestClientFactory_unknownProfile(t *testing.T) {
	f := NewClientFactory()
	if c, ok := f.NewClient("missing"); ok || c != nil {
		t.Fatalf("NewClient missing profile: got client=%v ok=%v", c, ok)
	}
}

func TestDefaultClientFactory_initAndGet(t *testing.T) {
	defaultClientFactoryMu.Lock()
	prev := defaultClientFactory
	defaultClientFactoryMu.Unlock()
	t.Cleanup(func() {
		defaultClientFactoryMu.Lock()
		defaultClientFactory = prev
		defaultClientFactoryMu.Unlock()
	})

	InitDefaultClientFactory(WithGlobalHeader("X-Default-Test", "1"))
	got := GetDefaultClientFactory()
	if got == nil {
		t.Fatal("GetDefaultClientFactory returned nil after InitDefaultClientFactory")
	}
	got.RegisterProfile("smoke", WithBaseURL("https://example.com"))
	if names := got.ProfileNames(); len(names) != 1 || names[0] != "smoke" {
		t.Fatalf("expected one registered profile, got %+v", names)
	}

	InitDefaultClientFactory()
	if got2 := GetDefaultClientFactory(); got2 == got {
		t.Fatal("expected InitDefaultClientFactory to replace default factory instance")
	}
}
