package optional

import (
	"encoding/json"
	"errors"
)

// ErrNotPresent is returned by Optional.Get when no value is set.
var ErrNotPresent = errors.New("optional: value not present")

// Optional is an optional T.
type Optional[T any] struct {
	value *T
}

// New returns an Optional holding a copy of v.
func New[T any](v T) Optional[T] {
	return Optional[T]{&v}
}

// NewNil returns an Optional with no value (absent).
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

// Clear removes the value; the optional becomes absent.
func (s *Optional[T]) Clear() {
	s.value = nil
}

// ToPtr returns a *T of the value or nil if not present.
func (s Optional[T]) ToPtr() *T {
	if !s.Present() {
		return nil
	}
	v := *s.value
	return &v
}

// Get returns the T value or [ErrNotPresent] if not present.
func (s Optional[T]) Get() (T, error) {
	if !s.Present() {
		var zero T
		return zero, ErrNotPresent
	}
	return *s.value, nil
}

// MustGet returns the T value or panics if not present.
func (s Optional[T]) MustGet() T {
	if !s.Present() {
		panic("optional: value not present")
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
	var probe any
	if err := json.Unmarshal(data, &probe); err != nil {
		return err
	}
	if probe == nil {
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
