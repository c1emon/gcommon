package stack

import "container/list"

type Stack[T any] struct {
	list *list.List
}

func NewStack[T any]() *Stack[T] {
	list := list.New()
	return &Stack[T]{list}
}

func (stack *Stack[T]) Push(value T) {
	stack.list.PushBack(value)
}

func (stack *Stack[T]) Pop() T {
	e := stack.list.Back()
	if e != nil {
		stack.list.Remove(e)
		return e.Value.(T)
	}
	var zero T
	return zero
}

func (stack *Stack[T]) Peak() T {
	e := stack.list.Back()
	if e != nil {
		return e.Value.(T)
	}
	var zero T
	return zero
}

func (stack *Stack[T]) Len() int {
	return stack.list.Len()
}

func (stack *Stack[T]) Empty() bool {
	return stack.list.Len() == 0
}
