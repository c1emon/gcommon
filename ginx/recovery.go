package ginx

import (
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"strings"

	"github.com/c1emon/gcommon/errorx"
	"github.com/c1emon/gcommon/vo"
	"github.com/gin-gonic/gin"
)

func panicAsError(v any) error {
	if e, ok := v.(error); ok {
		return e
	}
	return fmt.Errorf("%v", v)
}

// Recovery catches panics in downstream handlers, logs them, and sends a response.
// Broken pipe / connection reset is logged at warn without a JSON body.
func Recovery(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if v := recover(); v != nil {
				// Check for a broken connection, as it is not really a
				// condition that warrants a panic stack trace.
				if ne, ok := v.(*net.OpError); ok {
					var se *os.SyscallError
					if errors.As(ne, &se) {
						errStr := strings.ToLower(se.Error())
						if strings.Contains(errStr, "broken pipe") || strings.Contains(errStr, "connection reset by peer") {
							_ = c.Error(panicAsError(v))
							c.Abort()
							logger.Warn("gin recovery", "error", v)
							return
						}
					}
				}
				logger.Error("gin panic", "error", v)
				_ = c.Error(errorx.ErrInternal)
				c.AbortWithStatusJSON(http.StatusInternalServerError,
					vo.NewMsgResult(errorx.ErrInternal.Code(), errorx.ErrInternal.Error()))
			}
		}()
		c.Next()
	}
}
