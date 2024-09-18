package httpx

import (
	"fmt"
	"time"
)

type Response[T any] struct {
	Code int    `json:"code"`
	Msg  string `json:"msg,omitempty"`
	Ts   int64  `json:"ts"`
	Data T      `json:"data,omitempty"`
}

func (r *Response[T]) WithError(e error) *Response[T] {
	r.Msg = e.Error()
	return r
}

func NewResponse[T any](c int) *Response[T] {
	return &Response[T]{Code: c, Ts: time.Now().Unix()}
}

func ResponseOK[T any]() *Response[T] {
	return NewResponse[T](0)
}

func ResponseBadParam[T any](param, reason string) *Response[T] {
	resp := NewResponse[T](1001)
	resp.Msg = fmt.Sprintf("bad param [%s]: %s", param, reason)
	return resp
}

func ResponseNotFound[T any](id string) *Response[T] {
	resp := NewResponse[T](1002)
	resp.Msg = fmt.Sprintf("[%s] not found", id)
	return resp
}

func ResponseDuplicateKey[T any](key string) *Response[T] {
	resp := NewResponse[T](1003)
	resp.Msg = fmt.Sprintf("duplicate key [%s]", key)
	return resp
}

func ResponseNotAllowed[T any](res string) *Response[T] {
	resp := NewResponse[T](1004)
	resp.Msg = fmt.Sprintf("[%s] not allowed", res)
	return resp
}

type PaginationResponse[T any] struct {
	*Response[[]T]
	*Pagination
}

func WarpPagination[T any](resp *Response[[]T]) *PaginationResponse[T] {
	return &PaginationResponse[T]{
		Response: resp,
	}
}
