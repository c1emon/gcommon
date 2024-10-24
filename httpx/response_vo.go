package httpx

import (
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/c1emon/gcommon/errorx"
)

type ResponseVO[T any] struct {
	Code int    `json:"code"`
	Msg  string `json:"msg,omitempty"`
	Ts   int64  `json:"ts"`
	Data T      `json:"data,omitempty"`
}

func (r *ResponseVO[T]) WithError(e error) *ResponseVO[T] {
	r.Msg = e.Error()
	return r
}

func (r *ResponseVO[T]) HasError() bool {
	return r.Code != 0
}

func NewResponse[T any](c int) *ResponseVO[T] {
	return &ResponseVO[T]{Code: c, Ts: time.Now().Unix()}
}

func ResponseOK[T any]() *ResponseVO[T] {
	return NewResponse[T](0)
}

func ResponseBadParam[T any](param, reason string) *ResponseVO[T] {
	resp := NewResponse[T](1001)
	resp.Msg = fmt.Sprintf("bad param [%s]: %s", param, reason)
	return resp
}

func ResponseNotFound[T any](id string) *ResponseVO[T] {
	resp := NewResponse[T](1002)
	resp.Msg = fmt.Sprintf("[%s] not found", id)
	return resp
}

func ResponseDuplicateKey[T any](key string) *ResponseVO[T] {
	resp := NewResponse[T](1003)
	resp.Msg = fmt.Sprintf("duplicate key [%s]", key)
	return resp
}

func ResponseNotAllowed[T any](res string) *ResponseVO[T] {
	resp := NewResponse[T](1004)
	resp.Msg = fmt.Sprintf("[%s] not allowed", res)
	return resp
}

type PaginationResponse[T any] struct {
	*ResponseVO[[]T]
	*Pagination
}

func WarpPagination[T any](resp *ResponseVO[[]T]) *PaginationResponse[T] {
	return &PaginationResponse[T]{
		ResponseVO: resp,
	}
}

func ErrorInterceptor(client *Client, resp *Response) error {

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return errorx.NewIOError(err)
	}

	r := &ResponseVO[any]{}
	err = json.Unmarshal(b, r)
	if err != nil {
		return errorx.NewJsonError(err)
	}

	if r.HasError() {
		return errorx.NewHttpError(resp.StatusCode, r.Code, r.Msg, r.Data)
	}

	return nil
}
