package test

import (
	"context"
	"net/http"
	"os"
	"testing"

	"github.com/c1emon/gcommon/ginx"
	"github.com/c1emon/gcommon/logx"
	"github.com/c1emon/gcommon/server"
	"github.com/c1emon/gcommon/service"
	"github.com/gin-gonic/gin"
)

// Test_http_svc is a manual smoke test: it blocks on server.Run until a signal.
// Run with: INTEGRATION=1 go test -v ./test -run Test_http_svc
func Test_http_svc(t *testing.T) {
	if os.Getenv("INTEGRATION") != "1" {
		t.Skip("set INTEGRATION=1 to run this blocking smoke test")
	}

	logger := logx.NewLogger(logx.Config{
		Format: logx.FormatText,
	})

	repo := service.NewServiceRepo()

	r := ginx.NewDefaultEngine(ginx.DefaultEngineConfig{Logger: logger})

	r.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "Welcome Gin Server")
	})

	repo.Register(ginx.NewHTTPService(ginx.HTTPServiceConfig{
		Logger: logger,
		Server: &http.Server{
			Addr:    ":8080",
			Handler: r,
		},
	}))

	srv, err := server.New(repo, logger)
	if err != nil {
		t.Error(err)
		return
	}

	go func() {
		srv.ListenToSystemSignals(context.Background())
	}()
	srv.Run()
}
