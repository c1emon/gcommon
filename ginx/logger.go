package ginx

import (
	"fmt"
	"log/slog"

	"github.com/gin-gonic/gin"
)

func LogrusLogger(logger *slog.Logger) gin.HandlerFunc {

	return func(c *gin.Context) {

		// Start timer
		//start := time.Now()
		method := c.Request.Method
		uri := c.Request.RequestURI

		// Process request
		c.Next()
		//latency := time.Now().Sub(start)
		status := c.Writer.Status()

		switch {
		case status >= 100 && status < 400:
			logger.Info(fmt.Sprintf("[%s %d] %s", method, status, uri))
		case status >= 400 && status < 500 && len(c.Errors) > 0:
			logger.Warn(fmt.Sprintf("[%s %d] %s: %s", method, status, uri, c.Errors[0].Error()))
		case status >= 500 && status < 600 && len(c.Errors) > 0:
			logger.Error(fmt.Sprintf("[%s %d] %s: %s", method, status, uri, c.Errors[0].Error()))
		default:
			logger.Error(fmt.Sprintf("[%s %d] %s: %s\n%+v", method, status, uri, "unknown status", c.Errors))
		}

	}
}
