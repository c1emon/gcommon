package httpx

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestErrorInterceptor_bodyStillReadable(t *testing.T) {
	const payload = `{"code":0,"msg":"ok","ts":1}`

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(payload))
	}))
	t.Cleanup(srv.Close)

	f := NewClientFactory()
	f.RegisterProfile("t")
	c := f.MustNewClient("t")

	resp, err := c.R().Get(srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	if string(resp.Bytes()) != payload {
		t.Fatalf("caller Bytes(): got %q, want %q", string(resp.Bytes()), payload)
	}
}

func TestErrorInterceptor_secondHook_seesBody(t *testing.T) {
	const payload = `{"code":0,"msg":"ok","ts":1}`

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(payload))
	}))
	t.Cleanup(srv.Close)

	var second []byte
	f := NewClientFactory()
	f.RegisterProfile("t",
		WithRespInterceptor(func(_ *Client, resp *Response) error {
			b, err := resp.ToBytes()
			if err != nil {
				return err
			}
			second = b
			return nil
		}),
	)
	c := f.MustNewClient("t")

	_, err := c.R().Get(srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	if string(second) != payload {
		t.Fatalf("second hook body: got %q, want %q", string(second), payload)
	}
}

func TestErrorInterceptor_businessError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"code":400,"msg":"bad","ts":1}`))
	}))
	t.Cleanup(srv.Close)

	f := NewClientFactory()
	f.RegisterProfile("t", WithBusinessError())
	c := f.MustNewClient("t")
	_, err := c.R().Get(srv.URL)
	if err == nil {
		t.Fatal("want error for non-zero code")
	}
}

func TestErrorInterceptor_strictContentType_plainTextSkipped(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		_, _ = w.Write([]byte(`{"code":400,"msg":"bad","ts":1}`))
	}))
	t.Cleanup(srv.Close)

	f := NewClientFactory(WithGlobalBusinessError(), WithGlobalStrictJSONContentType())
	f.RegisterProfile("t")
	c := f.MustNewClient("t")
	_, err := c.R().Get(srv.URL)
	if err != nil {
		t.Fatalf("strict mode should skip non-json content-type, got %v", err)
	}
}

func TestErrorInterceptor_strictContentType_applicationJSONParsed(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		_, _ = w.Write([]byte(`{"code":400,"msg":"bad","ts":1}`))
	}))
	t.Cleanup(srv.Close)

	f := NewClientFactory(WithGlobalBusinessError(), WithGlobalStrictJSONContentType())
	f.RegisterProfile("t")
	c := f.MustNewClient("t")
	_, err := c.R().Get(srv.URL)
	if err == nil {
		t.Fatal("want error for json content-type with non-zero code")
	}
}

func TestBusinessError_defaultDoesNotMapEnvelope(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"code":400,"msg":"third-party","ts":1}`))
	}))
	t.Cleanup(srv.Close)

	f := NewClientFactory()
	f.RegisterProfile("third-party")
	resp, err := f.MustNewClient("third-party").R().Get(srv.URL)
	if err != nil {
		t.Fatalf("default client should not map third-party envelope: %v", err)
	}
	if string(resp.Bytes()) != `{"code":400,"msg":"third-party","ts":1}` {
		t.Fatalf("response body changed: %q", string(resp.Bytes()))
	}
}

func TestBusinessError_clientOptInMapsEnvelope(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"code":400,"msg":"bad","ts":1}`))
	}))
	t.Cleanup(srv.Close)

	f := NewClientFactory()
	f.RegisterProfile("biz", WithBusinessError())
	_, err := f.MustNewClient("biz").R().Get(srv.URL)
	if err == nil {
		t.Fatal("want business error when WithBusinessError is enabled")
	}
}

func TestBusinessError_globalEnableClientDisable(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"code":123,"msg":"x"}`))
	}))
	t.Cleanup(srv.Close)

	f := NewClientFactory(WithGlobalBusinessError())
	f.RegisterProfile("raw", DisableBusinessError())
	resp, err := f.MustNewClient("raw").R().Get(srv.URL)
	if err != nil {
		t.Fatalf("DisableBusinessError should leave response raw: %v", err)
	}
	if string(resp.Bytes()) != `{"code":123,"msg":"x"}` {
		t.Fatalf("body changed: %q", string(resp.Bytes()))
	}
}
