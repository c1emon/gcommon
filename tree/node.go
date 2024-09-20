package tree

type NodeLike[T any] interface {
	Node[T] | BiNode[T]
}

type Node[T any] struct {
	Children []*Node[T] `json:"children,omitempty"`
	Data     T          `json:"data,omitempty"`
}

func (n *Node[T]) AddChild(child *Node[T]) {
	if n.Children == nil {
		n.Children = []*Node[T]{}
	}
	n.Children = append(n.Children, child)
}

func (n *Node[T]) Iter() func(func(*Node[T]) bool) {
	return NewNodeIter(n)
}

func NewNode[T any](data T) *Node[T] {
	return &Node[T]{
		Data: data,
	}
}

type BiNode[T any] struct {
	Parent   *BiNode[T]   `json:"-"`
	IsRoot   bool         `json:"-"`
	Children []*BiNode[T] `json:"children,omitempty"`
	Data     T            `json:"data,omitempty"`
}

func (n *BiNode[T]) AddChild(child *BiNode[T]) {
	if n.Children == nil {
		n.Children = []*BiNode[T]{}
	}
	child.Parent = n
	child.IsRoot = false
	n.Children = append(n.Children, child)
}

func NewBiNode[T any](data T) *BiNode[T] {
	return &BiNode[T]{
		Data: data,
	}
}

func (n *BiNode[T]) Iter() func(func(*BiNode[T]) bool) {
	return NewBiNodeIter(n)
}

func fromNode[T any](n *Node[T]) *BiNode[T] {
	parent := &BiNode[T]{
		IsRoot: false,
		Data:   n.Data,
	}
	if n.Children != nil {
		parent.Children = []*BiNode[T]{}

		for _, child := range n.Children {
			biChild := fromNode(child)
			biChild.Parent = parent
			parent.Children = append(parent.Children, biChild)
		}
	}

	return parent
}

func FromRootNode[T any](n *Node[T]) *BiNode[T] {
	root := &BiNode[T]{
		IsRoot: true,
		Data:   n.Data,
	}

	if n.Children != nil {
		root.Children = []*BiNode[T]{}

		for _, child := range n.Children {
			biChild := fromNode(child)
			biChild.Parent = root
			root.Children = append(root.Children, biChild)
		}
	}

	return root
}
