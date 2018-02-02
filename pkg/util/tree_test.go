package util

import (
	"fmt"
	"reflect"
	"testing"
)

func TestTree(t *testing.T) {
	compareFunc := func(node *Node, addRes interface{}) bool {
		k := addRes.(int)
		kCompare := node.Res.(int)
		if k > kCompare {
			return false
		}
		return true
	}
	testSlice := []int{6, 3, 7, 2, 4, 5}
	targetSlice := []int{2, 3, 4, 5, 6, 7}
	slice := testSlice[:0]
	handle := func(res interface{}) error {
		slice = append(slice, res.(int))
		return nil
	}

	testTree := NewTree(compareFunc)

	for _, v := range testSlice {
		testTree.AddNode(v)
	}

	testTree.InOrderTraversal(testTree.GetRoot(), handle)
	if !reflect.DeepEqual(slice, targetSlice) {
		fmt.Printf(`TestTree failed`)
		t.FailNow()
	}
}
