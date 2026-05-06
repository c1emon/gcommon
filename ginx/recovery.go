package ginx

import (
	"errors"
	"fmt"
	"log/slog"
	"net"
	"os"
	"strings"

	"github.com/c1emon/gcommon/logx/v2"
	"github.com/c1emon/gcommon/v2/errorx"
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
// logger must be non-nil.
func Recovery(logger *slog.Logger) gin.HandlerFunc {
	if logger == nil {
		panic("ginx: Recovery called with nil *slog.Logger")
	}
	return func(c *gin.Context) {
		defer func() {
			if v := recover(); v != nil {
				ctx := c.Request.Context()
				panicErr := panicAsError(v)
				// Check for a broken connection, as it is not really a
				// condition that warrants a panic stack trace.
				if ne, ok := v.(*net.OpError); ok {
					var se *os.SyscallError
					if errors.As(ne, &se) {
						errStr := strings.ToLower(se.Error())
						if strings.Contains(errStr, "broken pipe") || strings.Contains(errStr, "connection reset by peer") {
							_ = c.Error(panicErr)
							setRequestError(c, normalizeError(panicErr))
							c.Abort()
							logger.WarnContext(ctx, "gin recovery", logx.Err(panicErr))
							return
						}
					}
				}
				logger.ErrorContext(ctx, "gin panic", logx.Err(panicErr))
				_ = c.Error(errorx.ErrInternal)
				setRequestError(c, normalizeError(errorx.ErrInternal))
				c.Abort()
			}
		}()
		c.Next()
	}
}
