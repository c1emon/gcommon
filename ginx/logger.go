package ginx

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"time"

	"github.com/c1emon/gcommon/logx/v2"
	"github.com/gin-gonic/gin"
)

// maxLoggedBodyBytes limits request/response body logging at debug level.
const maxLoggedBodyBytes = 4096

// Logger emits one structured log line per request after the handler returns.
// logger must be non-nil.
func Logger(logger *slog.Logger) gin.HandlerFunc {
	if logger == nil {
		panic("ginx: Logger called with nil *slog.Logger")
	}
	return func(c *gin.Context) {
		start := time.Now()
		ctx := c.Request.Context()
		debugEnabled := logger.Enabled(ctx, slog.LevelDebug)
		method := c.Request.Method
		path := c.Request.URL.Path
		if path == "" {
			path = c.Request.RequestURI
		}

		var requestBody []byte
		var requestHeaders map[string][]string
		var responseCapture *bodyCaptureWriter
		if debugEnabled {
			requestBody = captureRequestBody(c)
			requestHeaders = cloneHeader(c.Request.Header)
			responseCapture = &bodyCaptureWriter{ResponseWriter: c.Writer}
			c.Writer = responseCapture
		}

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()
		attrs := []any{
			logx.Method(method),
			logx.Status(status),
			logx.Path(path),
			logx.DurationMs(latency.Milliseconds()),
		}
		if reqErr, ok := getRequestError(c); ok && reqErr.err != nil {
			attrs = append(attrs, logx.Err(reqErr.err))
		}
		attrs = appendDetailAttrs(logger, ctx, c, attrs)
		if debugEnabled {
			reqBody, reqTruncated := summarizeBody(requestBody)
			respBody, respTruncated := summarizeBody(responseCapture.body.Bytes())
			attrs = append(attrs,
				slog.Any("request_headers", requestHeaders),
				slog.String("request_body", reqBody),
				slog.Bool("request_body_truncated", reqTruncated),
				slog.Any("response_headers", cloneHeader(c.Writer.Header())),
				slog.String("response_body", respBody),
				slog.Bool("response_body_truncated", respTruncated),
			)
		}

		switch {
		case status >= 100 && status < 400:
			logger.InfoContext(ctx, "http request", attrs...)
		case status >= 400 && status < 500:
			logger.WarnContext(ctx, "http request", attrs...)
		case status >= 500 && status < 600:
			logger.ErrorContext(ctx, "http request", attrs...)
		default:
			logger.ErrorContext(ctx, "http request", append(attrs, slog.String("note", "unusual status"))...)
		}
	}
}

func captureRequestBody(c *gin.Context) []byte {
	if c.Request == nil || c.Request.Body == nil {
		return nil
	}
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.Request.Body = io.NopCloser(bytes.NewReader(nil))
		return nil
	}
	c.Request.Body = io.NopCloser(bytes.NewReader(body))
	return body
}

func summarizeBody(body []byte) (string, bool) {
	if len(body) <= maxLoggedBodyBytes {
		return string(body), false
	}
	return string(body[:maxLoggedBodyBytes]), true
}

func cloneHeader(h map[string][]string) map[string][]string {
	if h == nil {
		return map[string][]string{}
	}
	dst := make(map[string][]string, len(h))
	for k, vals := range h {
		cloneVals := make([]string, len(vals))
		copy(cloneVals, vals)
		dst[k] = cloneVals
	}
	return dst
}

func appendDetailAttrs(logger *slog.Logger, ctx context.Context, c *gin.Context, attrs []any) []any {
	if !logger.Enabled(ctx, slog.LevelDebug) {
		return attrs
	}
	return append(attrs,
		logx.ClientIP(c.ClientIP()),
		slog.String("user_agent", c.Request.UserAgent()),
		slog.Int("response_bytes", c.Writer.Size()),
		slog.String("route", c.FullPath()),
	)
}

type bodyCaptureWriter struct {
	gin.ResponseWriter
	body bytes.Buffer
}

func (w *bodyCaptureWriter) Write(data []byte) (int, error) {
	_, _ = w.body.Write(data)
	return w.ResponseWriter.Write(data)
}

func (w *bodyCaptureWriter) WriteString(s string) (int, error) {
	_, _ = w.body.WriteString(s)
	return w.ResponseWriter.WriteString(s)
}
