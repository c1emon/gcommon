package tree_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/c1emon/gcommon/reader/tree"
)

func Test_a(t *testing.T) {
	root := tree.NewNode("root")
	root.AddChild(tree.NewNode("child0"))
	root.AddChild(tree.NewNode("child1"))
	root.AddChild(tree.NewNode("child2"))
	child3 := tree.NewNode("child3")
	root.AddChild(child3)

	child3.AddChild(tree.NewNode("child3-0"))
	child3.AddChild(tree.NewNode("child3-1"))
	child3.AddChild(tree.NewNode("child3-2"))

	v, err := json.MarshalIndent(root.ToBiNode().ToNode(), "", "  ")
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println(string(v))
}
