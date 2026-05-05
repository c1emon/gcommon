package httpx

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestBrowser_chromeSetsUserAgent(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(r.Header.Get("User-Agent")))
	}))
	t.Cleanup(srv.Close)

	f := NewClientFactory()
	f.RegisterProfile("c", WithBaseURL(srv.URL), WithBrowser(BrowserChrome))
	c := f.MustNewClient("c")
	resp, err := c.R().Get("/")
	if err != nil {
		t.Fatal(err)
	}
	ua := string(resp.Bytes())
	if !strings.Contains(strings.ToLower(ua), "chrome") {
		t.Fatalf("expected Chrome-like User-Agent, got %q", ua)
	}
}

func TestBrowser_profileWinsOverUserAgent(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(r.Header.Get("User-Agent")))
	}))
	t.Cleanup(srv.Close)

	f := NewClientFactory()
	f.RegisterProfile("c", WithBaseURL(srv.URL), WithBrowser(BrowserChrome), WithUserAgent("custom-ua-should-not-win"))
	c := f.MustNewClient("c")
	resp, err := c.R().Get("/")
	if err != nil {
		t.Fatal(err)
	}
	ua := string(resp.Bytes())
	if strings.Contains(ua, "custom-ua-should-not-win") {
		t.Fatalf("custom UA should be ignored when profile is set, got %q", ua)
	}
}
