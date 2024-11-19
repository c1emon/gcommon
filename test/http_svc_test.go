package test

import (
	"context"
	"log/slog"
	"net/http"
	"testing"

	"github.com/c1emon/gcommon/ginx"
	"github.com/c1emon/gcommon/httpx"
	"github.com/c1emon/gcommon/logx"
	"github.com/c1emon/gcommon/server"
	"github.com/c1emon/gcommon/service"
	"github.com/gin-gonic/gin"
)

func Test_http_svc(t *testing.T) {

	handler := logx.NewConsoleSlogHandler()
	logger := slog.New(handler)

	repo := service.NewServiceRepo()

	r := ginx.New(
		ginx.WithMiddleware(ginx.Recovery(logger)),
		ginx.WithMiddleware(ginx.Logger(logger)),
	)

	r.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "Welcome Gin Server")
	})

	repo.Registe(httpx.NewHttpService(":8080", r))

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
