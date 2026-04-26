package httphealth

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/c1emon/gcommon/health"
)

func TestHandler_okJSON(t *testing.T) {
	h, err := Handler(health.Config{ServiceName: "api", Version: "v1"})
	if err != nil {
		t.Fatal(err)
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status: %d", rec.Code)
	}
	body, _ := io.ReadAll(rec.Body)
	var payload struct {
		Status    string `json:"status"`
		Component struct {
			Name    string `json:"name"`
			Version string `json:"version"`
		} `json:"component"`
		System struct {
			Version string `json:"version"`
		} `json:"system"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		t.Fatalf("json: %v", err)
	}
	if payload.Status != "OK" {
		t.Fatalf("status field: %q", payload.Status)
	}
	if payload.Component.Name != "api" || payload.Component.Version != "v1" {
		t.Fatalf("component: %+v", payload.Component)
	}
	if payload.System.Version == "" {
		t.Fatal("expected non-empty Go runtime version in system block")
	}
}

func TestHandler_requiresServiceName(t *testing.T) {
	_, err := Handler(health.Config{})
	if err == nil {
		t.Fatal("expected error for empty ServiceName")
	}
}
