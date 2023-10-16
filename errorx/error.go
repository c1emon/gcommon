package errorx

import (
	"github.com/pkg/errors"
)

var (
	ErrInternal            = NewCommonError(1000, 502, "internal error")
	ErrUnknown             = NewCommonError(1001, 500, "unknown error")
	ErrIllegalParam        = NewCommonError(1002, 400, "illegal param")
	ErrDuplicateKey        = NewCommonError(1003, 400, "duplicate key")
	ErrResourceUnavailable = NewCommonError(1004, 400, "resource unavailable")
	ErrResourceNotFound    = NewCommonError(1005, 400, "resource not found")
)

type ErrorX interface {
	error
	Code() int
	HttpStatus() int
}

type CommonError struct {
	code       int
	httpStatus int
	message    string
}

func NewCommonError(code, httpStatus int, message string) ErrorX {
	return &CommonError{code: code, httpStatus: httpStatus, message: message}
}

func (e CommonError) Error() string {
	return e.message
}

func (e CommonError) Code() int {
	return e.code
}

func (e CommonError) HttpStatus() int {
	return e.httpStatus
}

func (e CommonError) Is(err error) bool {
	if x, ok := errors.Cause(err).(ErrorX); ok {
		return e.Code() == x.Code()
	}
	return false
}

func From(err error) ErrorX {
	if err == nil {
		return nil
	}

	for _, parser := range Parsers.Iter() {
		if parser.Support(err) {
			return parser.Parse(err)
		}
	}

	return ErrUnknown
}
