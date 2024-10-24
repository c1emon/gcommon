package errorx

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
