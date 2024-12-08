package tree_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/c1emon/gcommon/reader/tree"
)

func build() *tree.Node[string] {
	root := tree.NewNode("root")
	root.AddChild(tree.NewNode("child0"))
	root.AddChild(tree.NewNode("child1"))
	child2 := tree.NewNode("child2")
	root.AddChild(child2)
	child3 := tree.NewNode("child3")
	root.AddChild(child3)

	child2.AddChild(tree.NewNode("child2-0"))
	child2.AddChild(tree.NewNode("child2-1"))

	child3.AddChild(tree.NewNode("child3-0"))
	child3.AddChild(tree.NewNode("child3-1"))
	child3.AddChild(tree.NewNode("child3-2"))

	return root
}

func Test_map(t *testing.T) {

	v, err := json.MarshalIndent(build().ToBiNode().ToNode(), "", "  ")
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println(string(v))
}

func Test_iter(t *testing.T) {
	root := build().ToBiNode()
	for node := range root.Iter() {
		t.Log(node.Data)
	}
}
