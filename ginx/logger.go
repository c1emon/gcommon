package ginx

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func LogrusLogger(logger *logrus.Logger) gin.HandlerFunc {

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
			logger.Infof("[%s %d] %s", method, status, uri)
		case status >= 400 && status < 500 && len(c.Errors) > 0:
			logger.Warnf("[%s %d] %s: %s", method, status, uri, c.Errors[0].Error())
		case status >= 500 && status < 600 && len(c.Errors) > 0:
			logger.Errorf("[%s %d] %s: %s", method, status, uri, c.Errors[0].Error())
		default:
			logger.Errorf("[%s %d] %s: %s\n%+v", method, status, uri, "unknown status", c.Errors)
		}

	}
}
