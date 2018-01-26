package tree

import (
	"fmt"
	"testing"
	"reflect"
	"errors"
)

func TestTree(t *testing.T) {
	testTree := new(Tree)
	root := testTree.Root
	testSlice := []int{6,3,7,2,4,5}
	targetSlice := []int{2,3,4,5,6,7}
	compareFunc := func(addRes interface{}, node *Node) bool {
		k := addRes.(int)
		kCompare := node.Res.(int)
		if k > kCompare {
			return false
		}
		return true
	}
	for _, v := range testSlice {
		root = testTree.AddNode(root, v, compareFunc)
	}
	slice := testSlice[:0]

	handle := func(res interface{}) error {
		slice = append(slice, res.(int))
		return nil
	}

	testTree.MidOderTraversal(root, handle)
	if !reflect.DeepEqual(slice, targetSlice) {
		fmt.Printf(`TestTree failed`)
		t.FailNow()
	}

	handleFailFast := func(res interface{}) error {
		return errors.New("test fail fast")
	}
	slice = testSlice[:0]
	testTree.MidOderTraversalFailFast(root, handleFailFast)
	if len(slice) != 0 {
		fmt.Printf(`MidOderTraversalFailFast fail fast failed`)
		t.FailNow()
	}
}