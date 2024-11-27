package httpx

import (
	"time"

	"github.com/c1emon/gcommon/optional"
)

type ResultTypes[T any] interface {
	MsgResult | Result[T] | PageResult[T]
}

type MsgResult struct {
	Code int    `json:"code"`
	Msg  string `json:"msg,omitempty"`
	Ts   int64  `json:"ts"`
}

type Result[T any] struct {
	MsgResult
	Data optional.Optional[T] `json:"data,omitempty"`
}

func (r *Result[T]) HasError() bool {
	return r.Code != 0
}

func NewEmptyResult[T any]() *Result[T] {
	return &Result[T]{
		MsgResult: MsgResult{
			Code: 0,
			Msg:  "",
			Ts:   time.Now().Unix(),
		},
		Data: optional.NewNil[T]()}
}

func NewResult[T any](c int, msg string, data T) *Result[T] {
	return &Result[T]{
		MsgResult: MsgResult{
			Code: c,
			Msg:  msg,
			Ts:   time.Now().Unix(),
		},
		Data: optional.New(data)}
}

func NewMsgResult(c int, msg string) *MsgResult {
	return &MsgResult{
		Code: c,
		Msg:  msg,
		Ts:   time.Now().Unix(),
	}
}

func NewResultMsgOK[T any](msg string, data T) *Result[T] {
	return NewResult(0, msg, data)
}

func NewResultOK[T any](data T) *Result[T] {
	return NewResult(0, "ok", data)
}

// func ResponseBadParam[T any](param, reason string) *Result[T] {
// 	resp := NewResult[T](1001)
// 	resp.Msg = fmt.Sprintf("bad param [%s]: %s", param, reason)
// 	return resp
// }

// func ResponseNotFound[T any](id string) *Result[T] {
// 	resp := NewResult[T](1002)
// 	resp.Msg = fmt.Sprintf("[%s] not found", id)
// 	return resp
// }

// func ResponseDuplicateKey[T any](key string) *Result[T] {
// 	resp := NewResult[T](1003)
// 	resp.Msg = fmt.Sprintf("duplicate key [%s]", key)
// 	return resp
// }

// func ResponseNotAllowed[T any](res string) *Result[T] {
// 	resp := NewResult[T](1004)
// 	resp.Msg = fmt.Sprintf("[%s] not allowed", res)
// 	return resp
// }

type PageResult[T any] struct {
	Result[[]T]
	Pagination
}

func NewEmptyPageResult[T any]() *PageResult[T] {
	return &PageResult[T]{Result: *NewEmptyResult[[]T](), Pagination: Pagination{}}
}

func NewPageResult[T any](c int, msg string, datas []T) *PageResult[T] {
	return &PageResult[T]{Result: *NewResult[[]T](c, msg, datas), Pagination: Pagination{}}
}

func WarpPagination[T any](res Result[[]T]) *PageResult[T] {
	return &PageResult[T]{
		Result: res,
	}
}
