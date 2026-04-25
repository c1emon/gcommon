package httpx

import (
	"fmt"
	"log/slog"

	"github.com/imroc/req/v3"
)

type reqSlogLogger struct {
	log *slog.Logger
}

func (s reqSlogLogger) Errorf(format string, v ...any) {
	s.log.Error(fmt.Sprintf(format, v...))
}

func (s reqSlogLogger) Warnf(format string, v ...any) {
	s.log.Warn(fmt.Sprintf(format, v...))
}

func (s reqSlogLogger) Debugf(format string, v ...any) {
	s.log.Debug(fmt.Sprintf(format, v...))
}

var _ req.Logger = reqSlogLogger{}
