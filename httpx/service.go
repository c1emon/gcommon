package httpx

import (
	"context"
	"errors"
	"net"
	"net/http"

	"github.com/c1emon/gcommon/service"
)

var _ service.Service = &HttpService{}

type HttpService struct {
	srv *http.Server
}

func (HttpService) Name() string {
	return "HttpService"
}

func (s *HttpService) Start() error {
	ln, err := net.Listen("tcp", s.srv.Addr)
	if err != nil {
		return err
	}
	go func() {
		if err := s.srv.Serve(ln); err != nil && !errors.Is(err, http.ErrServerClosed) {
			// Serve errors after Start returned cannot be surfaced via Start(); callers may log via metrics.
			_ = err
		}
	}()
	return nil
}
func (s *HttpService) Stop(timeOutCtx context.Context) error {
	return s.srv.Shutdown(timeOutCtx)
}

func NewHttpService(endpoint string, handler http.Handler) *HttpService {
	return &HttpService{
		srv: &http.Server{
			Addr:    endpoint,
			Handler: handler,
		},
	}
}
