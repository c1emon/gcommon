package errorx

var (
	ErrHttpInternal            = NewHttpError(400, 1000, "internal error", nil)
	ErrHttpUnknown             = NewHttpError(400, 1001, "unknown error", nil)
	ErrHttpIllegalParam        = NewHttpError(400, 1002, "illegal param", nil)
	ErrHttpDuplicateKey        = NewHttpError(400, 1003, "duplicate key", nil)
	ErrHttpResourceUnavailable = NewHttpError(400, 1004, "resource unavailable", nil)
	ErrHttpResourceNotFound    = NewHttpError(400, 1005, "resource not found", nil)
)

type HttpError struct {
	*CommonError
	httpCode int
	data     any
}

func (e HttpError) HttpStatus() int {
	return e.httpCode
}

func (e HttpError) Data() any {
	return e.data
}

func NewHttpError(httpCode, code int, message string, data any) *HttpError {
	return &HttpError{CommonError: &CommonError{code: code, message: message}, httpCode: httpCode, data: data}
}
