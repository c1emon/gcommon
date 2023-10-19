package httpx

import "fmt"

type Response struct {
	Code    int    `json:"code"`
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
	Data    any    `json:"data,omitempty"`
	Pagination
}

func (r *Response) WithMessage(msg string) *Response {
	r.Message = msg
	return r
}

func (r *Response) WithError(err string) *Response {
	r.Error = err
	return r
}

func (r *Response) WithData(data any) *Response {
	r.Data = data
	return r
}

func (r *Response) WithPagination(pagination *Pagination) *Response {
	r.Pagination = *pagination
	return r
}

func NewResponse(c int) *Response {
	return &Response{Code: c}
}

func ResponseOK() *Response {
	return NewResponse(200)
}

func ResponseBadParam(param, reason string) *Response {
	return NewResponse(1001).WithMessage(fmt.Sprintf("bad param [%s]: %s", param, reason))
}

func ResponseNotFound(id string) *Response {
	return NewResponse(1002).WithMessage(fmt.Sprintf("[%s] not found", id))
}

func ResponseDuplicateKey(key string) *Response {
	return NewResponse(1003).WithMessage(fmt.Sprintf("duplicate key [%s]", key))
}

func ResponseNotAllowed(res string) *Response {
	return NewResponse(1004).WithMessage(fmt.Sprintf("[%s] not allowed", res))
}
