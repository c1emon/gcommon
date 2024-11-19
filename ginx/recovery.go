package ginx

import (
	"log/slog"
	"net"
	"os"
	"strings"

	"github.com/c1emon/gcommon/errorx"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

func Recovery(logger *slog.Logger) gin.HandlerFunc {

	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// Check for a broken connection, as it is not really a
				// condition that warrants a panic stack trace.
				if ne, ok := err.(*net.OpError); ok {
					var se *os.SyscallError
					if errors.As(ne, &se) {
						errStr := strings.ToLower(se.Error())
						if strings.Contains(errStr, "broken pipe") || strings.Contains(errStr, "connection reset by peer") {
							c.Error(err.(error))
							c.Abort()
							logger.Warn("gin recovery", "error", err)
							return
						}
					}
				}
				//TODO: add stack trace
				logger.Error("gin panic", "error", err)
				c.Error(errorx.ErrInternal)

			}

		}()

		c.Next()

	}
}
