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
package cache

import (
	"golang.org/x/net/context"
	"testing"
	"time"
)

type level1 struct {
}

func (l *level1) Name(ctx context.Context) string {
	return ctx.Value("key1").(string)
}

func (l *level1) Init(ctx context.Context, parent *Node) (node *Node, err error) {
	p := l.Name(ctx)
	if p == "1" {
		node = NewNode()
		node.Cache.Set("a", "a")
		node.Cache.Set("b", "b")
		node.Cache.Set("c", "c")
	}
	return
}

type level2 struct {
	changed string
}

func (l *level2) Name(ctx context.Context) string {
	return ctx.Value("key2").(string)
}

func (l *level2) Init(ctx context.Context, parent *Node) (node *Node, err error) {
	if parent == nil {
		return
	}

	p := l.Name(ctx)
	node = NewNode()
	if len(l.changed) == 0 {
		i := parent.Cache.Get(p)
		node.Cache.Set("aa", i)
	} else {
		node.Cache.Set("aa", l.changed)
	}

	return
}

func TestTree_Get(t *testing.T) {
	tree := NewTree(Configure().WithMaxSize(1).WithTTL(time.Second))
	node, err := tree.Get(context.Background())
	if node != nil || err != nil {
		t.Fatalf("TestTree_Get failed")
	}

	l2 := &level2{}
	tree.AddFilter(&level1{}, l2)

	ctx := context.WithValue(context.WithValue(context.Background(), "key1", ""), "key2", "")
	node, err = tree.Get(ctx)
	if node != nil || err != nil {
		t.Fatalf("TestTree_Get failed")
	}

	node, err = tree.Get(context.WithValue(context.WithValue(ctx, "key1", "1"), "key2", "b"))
	if node == nil || err != nil {
		t.Fatalf("TestTree_Get failed")
	}
	if node.Cache.Get("aa") != "b" {
		t.Fatalf("TestTree_Get failed")
	}

	node, err = tree.Get(context.WithValue(context.WithValue(ctx, "key1", "1"), "key2", "a"))
	if node == nil || err != nil {
		t.Fatalf("TestTree_Get failed")
	}
	if node.Cache.Get("aa") != "a" {
		t.Fatalf("TestTree_Get failed")
	}

	node, err = tree.Get(context.WithValue(context.WithValue(ctx, "key1", "1"), "key2", "b"))
	if node == nil || err != nil {
		t.Fatalf("TestTree_Get failed")
	}
	if node.Cache.Get("aa") != "b" {
		t.Fatalf("TestTree_Get failed")
	}

	tree.Remove(context.WithValue(ctx, "key1", "1"))

	l2.changed = "changed"
	node, err = tree.Get(context.WithValue(context.WithValue(ctx, "key1", "1"), "key2", "b"))
	if node == nil || err != nil {
		t.Fatalf("TestTree_Get failed")
	}
	if node.Cache.Get("aa") != "changed" {
		t.Fatalf("TestTree_Get failed")
	}

	l2.changed = ""
	node, err = tree.Simulate(context.WithValue(context.WithValue(ctx, "key1", "1"), "key2", "b"))
	if node == nil || err != nil {
		t.Fatalf("TestTree_Get failed")
	}
	if node.Cache.Get("aa") != "b" {
		t.Fatalf("TestTree_Get failed")
	}
}
