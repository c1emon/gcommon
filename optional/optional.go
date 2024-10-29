package optional

import (
	"encoding/json"
	"errors"
)

// Optional is an optional T.
type Optional[T any] struct {
	value *T
}

// New creates an optional.String from a T.
func New[T any](v T) Optional[T] {
	return Optional[T]{&v}
}

// NewNil creates an nil optional.String from a T.
func NewNil[T any]() Optional[T] {
	return Optional[T]{}
}

// NewFromPtr creates an optional[T] from a T pointer.
func NewFromPtr[T any](v *T) Optional[T] {
	if v == nil {
		return NewNil[T]()
	}
	return New(*v)
}

// Set sets the T value.
func (s *Optional[T]) Set(v T) {
	s.value = &v
}

// ToPtr returns a *T of the value or nil if not present.
func (s Optional[T]) ToPtr() *T {
	if !s.Present() {
		return nil
	}
	v := *s.value
	return &v
}

// Get returns the T value or an error if not present.
func (s Optional[T]) Get() (T, error) {
	if !s.Present() {
		var zero T
		return zero, errors.New("value not present")
	}
	return *s.value, nil
}

// MustGet returns the T value or panics if not present.
func (s Optional[T]) MustGet() T {
	if !s.Present() {
		panic("value not present")
	}
	return *s.value
}

// Present returns whether or not the value is present.
func (s Optional[T]) Present() bool {
	return s.value != nil
}

// OrElse returns the T value or a default value if the value is not present.
func (s Optional[T]) OrElse(v T) T {
	if s.Present() {
		return *s.value
	}
	return v
}

// If calls the function f with the value if the value is present.
func (s Optional[T]) If(fn func(T)) {
	if s.Present() {
		fn(*s.value)
	}
}

func (s Optional[T]) MarshalJSON() ([]byte, error) {
	if s.Present() {
		return json.Marshal(s.value)
	}
	return json.Marshal(nil)
}

func (s *Optional[T]) UnmarshalJSON(data []byte) error {

	if string(data) == "null" {
		s.value = nil
		return nil
	}

	var value T

	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}

	s.value = &value
	return nil
}
