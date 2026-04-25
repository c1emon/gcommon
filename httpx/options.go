package httpx

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/c1emon/gcommon/util"
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

func WithBaseUrl(url string) util.Option[Client] {
	return util.WrapFuncOption(func(client *Client) {
		client.SetBaseURL(url)
	})
}

func WithUA(ua string) util.Option[Client] {
	return util.WrapFuncOption(func(client *Client) {
		client.SetUserAgent(ua)
	})
}

func WithHeader(key, val string) util.Option[Client] {
	return util.WrapFuncOption(func(client *Client) {
		client.SetCommonHeader(key, val)
	})
}

func WithTimeOut(t time.Duration) util.Option[Client] {
	return util.WrapFuncOption(func(client *Client) {
		client.SetTimeout(t)
	})
}

// WithLogger wires req's internal logging to slog. Pass nil to disable req logs.
func WithLogger(l *slog.Logger) util.Option[Client] {
	return util.WrapFuncOption(func(client *Client) {
		if l == nil {
			client.SetLogger(nil)
			return
		}
		client.SetLogger(reqSlogLogger{log: l})
	})
}

func WithReqInterceptor(i ReqInterceptor) util.Option[Client] {
	return util.WrapFuncOption(func(client *Client) {
		client.OnBeforeRequest(func(client *req.Client, req *req.Request) error {
			return i(&Client{client}, &Request{req})
		})
	})
}

func WithRespInterceptor(i RespInterceptor) util.Option[Client] {
	return util.WrapFuncOption(func(client *Client) {
		client.OnAfterResponse(func(client *req.Client, req *req.Response) error {
			return i(&Client{client}, &Response{req})
		})
	})
}
