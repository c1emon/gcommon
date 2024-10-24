package errorx

import "fmt"

type IOError struct {
	*CommonError
}

func NewIOError(err error) *IOError {
	e := &IOError{CommonError: NewCommonError(1029, "io error")}
	e.WithCause(err)
	return e
}

type JsonError struct {
	*CommonError
}

func NewJsonError(err error) *JsonError {
	e := &JsonError{CommonError: NewCommonError(1049, "json error")}
	e.WithCause(err)
	return e
}

type FieldError struct {
	*CommonError
	key string
}

func NewFieldError(key string) *FieldError {
	return &FieldError{CommonError: NewCommonError(1059, "field error"), key: key}
}

type FieldNotFoundError struct {
	*FieldError
}

func NewFieldNotFoundError(key string) *FieldNotFoundError {
	e := &FieldNotFoundError{FieldError: NewFieldError(key)}
	e.message = fmt.Sprintf("%s not found", key)
	return e
}
