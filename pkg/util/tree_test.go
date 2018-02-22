/*
 * Licensed to the Apache Software Foundation (ASF) under one or more
 * contributor license agreements.  See the NOTICE file distributed with
 * this work for additional information regarding copyright ownership.
 * The ASF licenses this file to You under the Apache License, Version 2.0
 * (the "License"); you may not use this file except in compliance with
 * the License.  You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
package util

import (
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
		fail(t, `TestTree failed`)
	}
}
