package ginx

import (
	"errors"
	"net/http"

	"github.com/c1emon/gcommon/v2/errorx"
	"github.com/c1emon/gcommon/v2/vo"
	"github.com/gin-gonic/gin"
)

const requestErrorKey = "_ginx_request_error"

type requestError struct {
	err    error
	status int
	code   int
	msg    string
	data   any
}

func setRequestError(c *gin.Context, reqErr requestError) {
	c.Set(requestErrorKey, reqErr)
}

func getRequestError(c *gin.Context) (requestError, bool) {
	raw, ok := c.Get(requestErrorKey)
	if !ok {
		return requestError{}, false
	}
	reqErr, ok := raw.(requestError)
	if !ok {
		return requestError{}, false
	}
	return reqErr, true
}

func normalizeError(err error) requestError {
	var he *errorx.HttpError
	if errors.As(err, &he) {
		return requestError{
			err:    err,
			status: he.HttpStatus(),
			code:   he.Code(),
			msg:    he.Error(),
			data:   he.Data(),
		}
	}

	var ce *errorx.CommonError
	if errors.As(err, &ce) {
		status := http.StatusBadRequest
		if errors.Is(err, errorx.ErrInternal) {
			status = http.StatusInternalServerError
		}
		return requestError{
			err:    err,
			status: status,
			code:   ce.Code(),
			msg:    ce.Error(),
		}
	}

	return requestError{
		err:    err,
		status: http.StatusBadRequest,
		code:   -1,
		msg:    err.Error(),
	}
}

// ErrorResponder writes one JSON error response after the handler chain.
func ErrorResponder() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if c.Writer.Written() {
			return
		}

		if reqErr, ok := getRequestError(c); ok {
			if reqErr.data != nil {
				c.JSON(reqErr.status, vo.NewResult(reqErr.code, reqErr.msg, reqErr.data))
				return
			}
			c.JSON(reqErr.status, vo.NewMsgResult(reqErr.code, reqErr.msg))
			return
		}

		if len(c.Errors) == 0 {
			return
		}
		last := c.Errors.Last().Err
		reqErr := normalizeError(last)
		setRequestError(c, reqErr)
		if reqErr.data != nil {
			c.JSON(reqErr.status, vo.NewResult(reqErr.code, reqErr.msg, reqErr.data))
			return
		}
		c.JSON(reqErr.status, vo.NewMsgResult(reqErr.code, reqErr.msg))
	}
}
