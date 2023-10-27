package ginx

import (
	"github.com/c1emon/gcommon/logx"
	"github.com/gin-gonic/gin"
)

func LogrusLogger(logger logx.Logger) gin.HandlerFunc {

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
			logger.Info("[%s %d] %s", method, status, uri)
		case status >= 400 && status < 500 && len(c.Errors) > 0:
			logger.Warn("[%s %d] %s: %s", method, status, uri, c.Errors[0].Error())
		case status >= 500 && status < 600 && len(c.Errors) > 0:
			logger.Error("[%s %d] %s: %s", method, status, uri, c.Errors[0].Error())
		default:
			logger.Error("[%s %d] %s: %s\n%+v", method, status, uri, "unknown status", c.Errors)
		}

	}
}
