package tree

import "github.com/c1emon/gcommon/stack"

type nodeIter[T any] struct {
	stack *stack.Stack[*Node[T]]
}

func NewNodeIter[T any](root *Node[T]) func(func(*Node[T]) bool) {
	stack := stack.NewStack[*Node[T]]()
	stack.Push(root)

	i := &nodeIter[T]{stack}

	return func(yield func(*Node[T]) bool) {
		for !i.stack.Empty() {
			node := i.stack.Pop()
			if node.Children != nil {
				for _, child := range node.Children {
					i.stack.Push(child)
				}
			}
			if !yield(node) {
				return
			}
		}
	}
}

type biNodeIter[T any] struct {
	stack *stack.Stack[*BiNode[T]]
}

func NewBiNodeIter[T any](root *BiNode[T]) func(func(*BiNode[T]) bool) {
	stack := stack.NewStack[*BiNode[T]]()
	stack.Push(root)

	i := &biNodeIter[T]{stack}

	return func(yield func(*BiNode[T]) bool) {
		for !i.stack.Empty() {
			node := i.stack.Pop()
			if node.Children != nil {
				for _, child := range node.Children {
					i.stack.Push(child)
				}
			}
			if !yield(node) {
				return
			}
		}
	}
}
