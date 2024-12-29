package util

type Builder[T any] interface {
	Build() (T, error)
}
