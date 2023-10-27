package util

type Option[T any] interface {
	Apply(*T)
}

type FuncOption[T any] struct {
	f func(*T)
}

func (fo *FuncOption[T]) Apply(optVal *T) {
	fo.f(optVal)
}

func WrapFuncOption[T any](f func(*T)) *FuncOption[T] {
	return &FuncOption[T]{f: f}
}
