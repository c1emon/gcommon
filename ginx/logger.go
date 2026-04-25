package ginx

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
)

// Logger emits one structured log line per request after the handler returns.
func Logger(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		method := c.Request.Method
		path := c.Request.URL.Path
		if path == "" {
			path = c.Request.RequestURI
		}

		c.Next()

		status := c.Writer.Status()
		latency := time.Since(start)
		attrs := []any{
			slog.String("method", method),
			slog.Int("status", status),
			slog.String("path", path),
			slog.Duration("latency", latency),
		}

		switch {
		case status >= 100 && status < 400:
			logger.Info("http request", attrs...)
		case status >= 400 && status < 500:
			if len(c.Errors) > 0 {
				attrs = append(attrs, slog.String("err", c.Errors.Last().Error()))
			}
			logger.Warn("http request", attrs...)
		case status >= 500 && status < 600:
			if len(c.Errors) > 0 {
				attrs = append(attrs, slog.String("err", c.Errors.Last().Error()))
			}
			logger.Error("http request", attrs...)
		default:
			if len(c.Errors) > 0 {
				attrs = append(attrs, slog.String("err", c.Errors.Last().Error()))
			}
			logger.Error("http request", append(attrs, slog.String("note", "unusual status"))...)
		}
	}
}
