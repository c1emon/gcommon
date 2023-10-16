package httpx

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
	return NewResponse(0)
}
