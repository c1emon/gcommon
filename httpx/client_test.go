package httpx_test

import (
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/c1emon/gcommon/httpx/v2"
)

func TestClientFactory_clientIsolation(t *testing.T) {
	a := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("a"))
	}))
	b := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("b"))
	}))
	t.Cleanup(a.Close)
	t.Cleanup(b.Close)

	f := httpx.NewClientFactory()
	f.RegisterProfile("a", httpx.WithBaseURL(a.URL))
	f.RegisterProfile("b", httpx.WithBaseURL(b.URL))
	ca := f.MustNewClient("a")
	cb := f.MustNewClient("b")

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

func TestClientFactory_cookieJarFactoryCreatesIndependentClients(t *testing.T) {
	f := httpx.NewClientFactory()
	created := 0
	f.RegisterProfile("session", httpx.WithCookieJarFactory(func() *cookiejar.Jar {
		created++
		jar, err := cookiejar.New(nil)
		if err != nil {
			t.Fatal(err)
		}
		return jar
	}))

	first := f.MustNewClient("session")
	second := f.MustNewClient("session")
	if created != 2 {
		t.Fatalf("expected cookie jar factory to run once per client, got %d", created)
	}

	firstJar := first.GetClient().Jar
	secondJar := second.GetClient().Jar
	if firstJar == nil || secondJar == nil {
		t.Fatalf("expected both clients to have cookie jars, got first=%v second=%v", firstJar, secondJar)
	}
	if firstJar == secondJar {
		t.Fatal("expected independent cookie jars")
	}

	u, err := url.Parse("https://example.com/")
	if err != nil {
		t.Fatal(err)
	}
	firstJar.SetCookies(u, []*http.Cookie{{Name: "sid", Value: "first"}})

	firstCookies, err := first.GetCookies(u.String())
	if err != nil {
		t.Fatal(err)
	}
	secondCookies, err := second.GetCookies(u.String())
	if err != nil {
		t.Fatal(err)
	}
	if len(firstCookies) != 1 || firstCookies[0].Name != "sid" || firstCookies[0].Value != "first" {
		t.Fatalf("unexpected first client cookies: %+v", firstCookies)
	}
	if len(secondCookies) != 0 {
		t.Fatalf("expected second client cookies to remain isolated, got %+v", secondCookies)
	}
}

func TestClientClone_preservesNameAndIsUnmanagedVariant(t *testing.T) {
	f := httpx.NewClientFactory()
	f.RegisterProfile("session")

	original := f.MustNewClient("session")
	clone := original.Clone()
	if clone == nil {
		t.Fatal("Clone returned nil")
	}
	if clone == original {
		t.Fatal("Clone returned the same wrapper")
	}
	if clone.Name() != original.Name() {
		t.Fatalf("expected clone name %q, got %q", original.Name(), clone.Name())
	}
	if clone.Client == original.Client {
		t.Fatal("Clone returned wrapper around the same req client")
	}

	clone.SetCookieJar(nil)
	if original.GetClient().Jar == nil {
		t.Fatal("expected clone cookie jar changes not to affect original")
	}

	var nilClient *httpx.Client
	if nilClient.Clone() != nil {
		t.Fatal("nil client clone should return nil")
	}
	if got := (&httpx.Client{}).Clone(); got == nil {
		t.Fatal("client with nil req client should still clone to an httpx wrapper")
	}
}
