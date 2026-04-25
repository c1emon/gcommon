package ginx

import (
	"errors"
	"net/http"

	"github.com/c1emon/gcommon/errorx"
	"github.com/c1emon/gcommon/httpx"
	"github.com/gin-gonic/gin"
)

// ErrorHandler writes one JSON error response after the handler chain, preferring
// the most recent [errorx.HttpError] in the gin error list.
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		if len(c.Errors) == 0 {
			return
		}
		for i := len(c.Errors) - 1; i >= 0; i-- {
			e := c.Errors[i]
			var he *errorx.HttpError
			if errors.As(e.Err, &he) {
				c.JSON(he.HttpStatus(), httpx.NewMsgResult(he.Code(), e.Error()))
				return
			}
		}
		last := c.Errors.Last()
		c.JSON(http.StatusBadRequest, httpx.NewMsgResult(-1, last.Error()))
	}
}
