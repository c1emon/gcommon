package httpx

import (
	"context"
	"errors"
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
	var err error
	go func() {
		err = s.srv.ListenAndServe()
	}()
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err
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
