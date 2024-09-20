package tree

import (
	"fmt"
)

type TreeReadFn[T, S any] func(T) ([]T, []S, error)
type ChildAddedFn[S any] func(child *BiNode[S])

type Reader[T, S any] struct {
	fn           TreeReadFn[T, S]
	childAddedFn ChildAddedFn[S]
	maxDepth     int
}

func (r *Reader[T, S]) SetMaxDepth(depth int) {
	r.maxDepth = depth
}

func (r *Reader[T, S]) SetChildAddedFn(fn ChildAddedFn[S]) {
	r.childAddedFn = fn
}

func (r *Reader[T, S]) read(parent *BiNode[S], arg T, depth int) error {
	if r.maxDepth > 0 && depth >= r.maxDepth {
		return fmt.Errorf("reach max depth")
	}

	args, datas, err := r.fn(arg)
	if err != nil {
		return err
	}
	if len(args) != len(datas) {
		return fmt.Errorf("TreeReadFn returned length mismatch")
	}
	if len(args) == 0 {
		return nil
	}

	for i := range len(args) {
		child := NewBiNode(datas[i])
		parent.AddChild(child)
		if r.childAddedFn != nil {
			r.childAddedFn(child)
		}
		err := r.read(child, args[i], depth+1)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *Reader[T, S]) Read(args T) (*BiNode[S], error) {
	root := &BiNode[S]{IsRoot: true}
	return root, r.read(root, args, 0)
}

func NewTreeReader[T, S any](fn TreeReadFn[T, S]) *Reader[T, S] {
	return &Reader[T, S]{fn: fn, maxDepth: 0}
}
