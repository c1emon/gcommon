package ginx

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"

	"github.com/c1emon/gcommon/logx"
	"github.com/c1emon/gcommon/service"
	"github.com/gin-gonic/gin"
)

var _ service.Service = &HTTPService{}

// HTTPService adapts an http.Server to the service.Service lifecycle.
type HTTPService struct {
	name    string
	server  *http.Server
	logger  *slog.Logger
	serveCh chan error
}

// Name returns the service name shown in lifecycle logs.
func (s *HTTPService) Name() string {
	return s.name
}

// ServeErrors exposes asynchronous Serve errors after Start returns.
func (s *HTTPService) ServeErrors() <-chan error {
	return s.serveCh
}

// Start begins listening and serving in a background goroutine.
func (s *HTTPService) Start() error {
	ln, err := net.Listen("tcp", s.server.Addr)
	if err != nil {
		return err
	}
	go func() {
		if err := s.server.Serve(ln); err != nil && !errors.Is(err, http.ErrServerClosed) {
			select {
			case s.serveCh <- err:
			default:
			}
			s.logger.Error("http service serve failed", slog.String("service", s.name), logx.Err(err))
		}
	}()
	return nil
}

// Stop gracefully shuts the underlying http.Server down.
func (s *HTTPService) Stop(timeOutCtx context.Context) error {
	return s.server.Shutdown(timeOutCtx)
}

// HTTPServiceConfig configures NewHTTPService.
type HTTPServiceConfig struct {
	Name   string
	Server *http.Server
	Logger *slog.Logger
}

// NewHTTPService builds an HTTPService from a preconfigured *http.Server.
// If Name is empty it defaults to "http@<addr>".
// Logger must be non-nil.
func NewHTTPService(cfg HTTPServiceConfig) *HTTPService {
	server := cfg.Server
	if server == nil {
		server = &http.Server{}
	}
	if cfg.Logger == nil {
		panic("ginx: HTTPServiceConfig.Logger must be non-nil")
	}
	logger := cfg.Logger
	name := cfg.Name
	if name == "" {
		name = fmt.Sprintf("http@%s", server.Addr)
	}
	return &HTTPService{
		name:    name,
		server:  server,
		logger:  logger,
		serveCh: make(chan error, 1),
	}
}

// GinServiceConfig configures NewGinService.
type GinServiceConfig struct {
	Name   string
	Addr   string
	Engine *gin.Engine
	Logger *slog.Logger
}

// NewGinService builds an HTTPService from a gin engine and listen address.
// Logger must be non-nil.
func NewGinService(cfg GinServiceConfig) *HTTPService {
	return NewHTTPService(HTTPServiceConfig{
		Name:   cfg.Name,
		Logger: cfg.Logger,
		Server: &http.Server{
			Addr:    cfg.Addr,
			Handler: cfg.Engine,
		},
	})
}
