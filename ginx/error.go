package ginx

import (
	"net/http"

	"github.com/c1emon/gcommon/errorx"
	"github.com/c1emon/gcommon/httpx"
	"github.com/gin-gonic/gin"
)

func ErrorHandler() gin.HandlerFunc {

	return func(c *gin.Context) {

		// Process request
		c.Next()
		for _, e := range c.Errors {
			if err, ok := e.Err.(errorx.ErrorX); ok {
				c.JSON(err.HttpStatus(), httpx.NewResponse[any](err.Code()).WithError(err))
			} else {
				c.JSON(http.StatusBadRequest, httpx.NewResponse[any](1001).WithError(e))
			}
			return
		}

	}
}
