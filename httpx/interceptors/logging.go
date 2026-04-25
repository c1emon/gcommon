package interceptors

import (
	"context"
	"log/slog"

	"github.com/imroc/req/v3"
)

// RequestLogger logs request method and URL.
func RequestLogger(log *slog.Logger) req.RequestMiddleware {
	return func(_ *req.Client, r *req.Request) error {
		ctx := r.Context()
		urlStr := r.RawURL
		if urlStr == "" && r.URL != nil {
			urlStr = r.URL.String()
		}
		log.InfoContext(ctx, "http request", slog.String("method", r.Method), slog.String("url", urlStr))
		return nil
	}
}

// ResponseLogger logs response HTTP status.
func ResponseLogger(log *slog.Logger) req.ResponseMiddleware {
	return func(_ *req.Client, r *req.Response) error {
		ctx := context.Background()
		if r.Request != nil {
			ctx = r.Request.Context()
		}
		code := 0
		if r.Response != nil {
			code = r.StatusCode
		}
		log.InfoContext(ctx, "http response", slog.Int("status", code))
		return nil
	}
}
