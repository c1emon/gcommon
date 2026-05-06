package healthgin

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/c1emon/gcommon/health/v2"
	"github.com/gin-gonic/gin"
)

func TestHandler_ginJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, err := Handler(health.Config{ServiceName: "api", Version: "v2"})
	if err != nil {
		t.Fatal(err)
	}

	eng := gin.New()
	eng.GET("/live", h)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/live", nil)
	eng.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status: %d", rec.Code)
	}
	body, _ := io.ReadAll(rec.Body)
	var payload struct {
		Component struct {
			Name    string `json:"name"`
			Version string `json:"version"`
		} `json:"component"`
		System struct {
			GoroutinesCount int `json:"goroutines_count"`
		} `json:"system"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		t.Fatalf("json: %v", err)
	}
	if payload.Component.Name != "api" || payload.Component.Version != "v2" {
		t.Fatalf("component: %+v", payload.Component)
	}
	if payload.System.GoroutinesCount < 1 {
		t.Fatalf("unexpected goroutines_count: %d", payload.System.GoroutinesCount)
	}
}
