package httpx

import (
	"log/slog"
	"time"

	"github.com/c1emon/gcommon/util"
	"github.com/imroc/req/v3"
)

type Options struct {
}

func WithBaseUrl(url string) *util.FuncOption[Client] {
	return util.WrapFuncOption(func(client *Client) {
		client.SetBaseURL(url)
	})
}

func WithUA(ua string) *util.FuncOption[Client] {
	return util.WrapFuncOption(func(client *Client) {
		client.SetUserAgent(ua)
	})
}

func WithHeader(key, val string) *util.FuncOption[Client] {
	return util.WrapFuncOption(func(client *Client) {
		client.SetCommonHeader(key, val)
	})
}

func WithTimeOut(t time.Duration) *util.FuncOption[Client] {
	return util.WrapFuncOption(func(client *Client) {
		client.SetTimeout(t)
	})
}

// TODO: adapt for slog
func WithLogger(l *slog.Logger) *util.FuncOption[Client] {
	return util.WrapFuncOption(func(client *Client) {
		client.SetLogger(nil)
	})
}

func WithReqInterceptor(i ReqInterceptor) *util.FuncOption[Client] {
	return util.WrapFuncOption(func(client *Client) {
		client.OnBeforeRequest(func(client *req.Client, req *req.Request) error {
			return i(&Client{client}, &Request{req})
		})
	})
}

func WithRespInterceptor(i RespInterceptor) *util.FuncOption[Client] {
	return util.WrapFuncOption(func(client *Client) {
		client.OnAfterResponse(func(client *req.Client, req *req.Response) error {
			return i(&Client{client}, &Response{req})
		})
	})
}
