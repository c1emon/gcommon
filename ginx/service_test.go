package ginx

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func TestNewHTTPServiceNameDefault(t *testing.T) {
	svc := NewHTTPService(HTTPServiceConfig{
		Server: &http.Server{
			Addr:    "127.0.0.1:18080",
			Handler: http.NewServeMux(),
		},
	})
	if svc.Name() != "http@127.0.0.1:18080" {
		t.Fatalf("unexpected default name: %s", svc.Name())
	}
}

func TestHTTPServiceStartInvalidAddr(t *testing.T) {
	svc := NewHTTPService(HTTPServiceConfig{
		Server: &http.Server{
			Addr:    "bad::addr",
			Handler: http.NewServeMux(),
		},
	})

	if err := svc.Start(); err == nil {
		t.Fatal("expected start error for invalid address")
	}
}

func TestHTTPServiceStartAndStop(t *testing.T) {
	svc := NewHTTPService(HTTPServiceConfig{
		Name: "test-http",
		Server: &http.Server{
			Addr:    "127.0.0.1:0",
			Handler: http.NewServeMux(),
		},
	})

	if err := svc.Start(); err != nil {
		t.Fatalf("start service: %v", err)
	}

	timeoutCtx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := svc.Stop(timeoutCtx); err != nil {
		t.Fatalf("stop service: %v", err)
	}
}

func TestNewGinService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	svc := NewGinService(GinServiceConfig{
		Name:   "engine-http",
		Addr:   "127.0.0.1:18082",
		Engine: engine,
	})

	if svc.Name() != "engine-http" {
		t.Fatalf("unexpected service name: %s", svc.Name())
	}
	if svc.server == nil {
		t.Fatal("server should not be nil")
	}
	if svc.server.Addr != "127.0.0.1:18082" {
		t.Fatalf("unexpected addr: %s", svc.server.Addr)
	}
	if svc.server.Handler != engine {
		t.Fatal("server handler should be the provided gin engine")
	}
}
