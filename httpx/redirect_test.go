package httpx

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRedirectPolicy_profileNoRedirectReadsLocation(t *testing.T) {
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("target"))
	}))
	t.Cleanup(target.Close)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Location", target.URL+"/next")
		w.WriteHeader(http.StatusFound)
	}))
	t.Cleanup(srv.Close)

	f := NewClientFactory()
	f.RegisterProfile("no-redirect", WithRedirectPolicy(NoRedirectPolicy()))
	resp, err := f.MustNewClient("no-redirect").R().Get(srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusFound {
		t.Fatalf("want 302 without following redirect, got %d", resp.StatusCode)
	}
	if got := resp.Header.Get("Location"); got != target.URL+"/next" {
		t.Fatalf("Location: got %q", got)
	}
}

func TestRedirectPolicy_defaultFollowsRedirect(t *testing.T) {
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("target"))
	}))
	t.Cleanup(target.Close)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Location", target.URL+"/next")
		w.WriteHeader(http.StatusFound)
	}))
	t.Cleanup(srv.Close)

	f := NewClientFactory()
	f.RegisterProfile("default")
	resp, err := f.MustNewClient("default").R().Get(srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("want followed redirect 200, got %d", resp.StatusCode)
	}
	if string(resp.Bytes()) != "target" {
		t.Fatalf("body: got %q", string(resp.Bytes()))
	}
}

func TestRedirectPolicy_cloneDoesNotMutateSessionClient(t *testing.T) {
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("target"))
	}))
	t.Cleanup(target.Close)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Location", target.URL+"/next")
		w.WriteHeader(http.StatusFound)
	}))
	t.Cleanup(srv.Close)

	f := NewClientFactory()
	f.RegisterProfile("browser")
	session := f.MustNewClient("browser")
	noRedirect := session.Clone().SetRedirectPolicy(NoRedirectPolicy())

	cloneResp, err := noRedirect.R().Get(srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	if cloneResp.StatusCode != http.StatusFound {
		t.Fatalf("clone should not follow redirect, got %d", cloneResp.StatusCode)
	}

	sessionResp, err := session.R().Get(srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	if sessionResp.StatusCode != http.StatusOK {
		t.Fatalf("session should still follow redirect, got %d", sessionResp.StatusCode)
	}
}
