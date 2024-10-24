package errorx

import (
	"github.com/pkg/errors"
)

var (
	ErrInternal            = NewCommonError(1000, "internal error")
	ErrUnknown             = NewCommonError(1001, "unknown error")
	ErrIllegalParam        = NewCommonError(1002, "illegal param")
	ErrDuplicateKey        = NewCommonError(1003, "duplicate key")
	ErrResourceUnavailable = NewCommonError(1004, "resource unavailable")
	ErrResourceNotFound    = NewCommonError(1005, "resource not found")

	ErrIO = NewCommonError(1005, "resource not found")
)

// type ErrorX interface {
// 	error
// 	Code() int
// 	Unwrap() error
// }

type CommonError struct {
	error
	code    int
	message string
	// err      error
	metadata map[string]any
}

func (e CommonError) Error() string {
	return e.message
}

func (e CommonError) Code() int {
	return e.code
}

func (e CommonError) Unwrap() error {
	return e.error
}

func (e CommonError) Is(err error) bool {
	if x, ok := errors.Cause(err).(CommonError); ok {
		return e.Code() == x.Code()
	}
	return false
}

func (e *CommonError) WithCause(err error) *CommonError {
	e.error = err
	return e
}

func (e *CommonError) SetMetadata(key string, val any) {
	if val == nil {
		delete(e.metadata, key)
	}
	e.metadata[key] = val
}

func (e *CommonError) GetMetadata(key string) any {
	return e.metadata[key]
}

func NewCommonError(code int, message string) *CommonError {
	return &CommonError{code: code, message: message, metadata: make(map[string]any)}
}
