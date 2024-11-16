package tree

type mapper[T, S any] struct {
	fn func(T) S
}

func (m mapper[T, S]) mapChild(n *Node[T]) *Node[S] {
	parent := &Node[S]{
		Data: m.fn(n.Data),
	}
	if n.Children != nil {
		parent.Children = []*Node[S]{}

		for _, child := range n.Children {
			parent.Children = append(parent.Children, m.mapChild(child))
		}
	}

	return parent
}

func (m mapper[T, S]) mapRoot(root *Node[T]) *Node[S] {
	newRoot := &Node[S]{
		Data: m.fn(root.Data),
	}

	if root.Children != nil {
		newRoot.Children = []*Node[S]{}

		for _, child := range root.Children {
			newRoot.Children = append(newRoot.Children, m.mapChild(child))
		}
	}

	return newRoot
}

func (m mapper[T, S]) Do(src *Node[T]) *Node[S] {
	return m.mapRoot(src)
}

func NewMapper[T, S any](fn func(T) S) *mapper[T, S] {
	return &mapper[T, S]{
		fn: fn,
	}
}
