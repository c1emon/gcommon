package interceptors

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/imroc/req/v3"
)

func TestError_StrictSkipsNonJSONContentType(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		_, _ = w.Write([]byte(`{"code":400,"msg":"bad"}`))
	}))
	t.Cleanup(srv.Close)

	c := req.C().OnAfterResponse(Error(true))
	_, err := c.R().Get(srv.URL)
	if err != nil {
		t.Fatalf("strict mode should skip non-json content type: %v", err)
	}
}

func TestError_StrictParsesJSONContentType(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		_, _ = w.Write([]byte(`{"code":400,"msg":"bad"}`))
	}))
	t.Cleanup(srv.Close)

	c := req.C().OnAfterResponse(Error(true))
	_, err := c.R().Get(srv.URL)
	if err == nil {
		t.Fatal("want http error for non-zero code with json content type")
	}
}

func TestError_NonStrictParsesWithoutJSONContentType(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		_, _ = w.Write([]byte(`{"code":400,"msg":"bad"}`))
	}))
	t.Cleanup(srv.Close)

	c := req.C().OnAfterResponse(Error(false))
	_, err := c.R().Get(srv.URL)
	if err == nil {
		t.Fatal("want http error in non-strict mode even if content type is text/plain")
	}
}
