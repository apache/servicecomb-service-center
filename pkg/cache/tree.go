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
	"github.com/karlseguin/ccache"
	"golang.org/x/net/context"
	"sync"
)

type Tree struct {
	Config  *Config
	nodes   []*ccache.Cache
	filters []Filter
	lock    sync.RWMutex
}

func (t *Tree) AddFilter(f ...Filter) *Tree {
	t.filters = append(t.filters, f...)
	return t
}

func (t *Tree) Get(ctx context.Context) (node *Node, err error) {
	var parent *Node
	for i := range t.filters {
		if parent, err = t.getOrCreateNode(ctx, i, parent); parent == nil || err != nil {
			break
		}
	}
	node = parent
	return
}

func (t *Tree) Remove(ctx context.Context) {
	if len(t.filters) == 0 {
		return
	}

	filter := t.filters[0]
	if parent := t.getNode(0, filter.Name(ctx)); parent != nil {
		t.lock.Lock()
		t.remove(parent)
		t.lock.Unlock()
	}
}

func (t *Tree) remove(parent *Node) {
	for _, child := range parent.Childs {
		t.remove(child)
	}
	t.nodes[parent.Level].Delete(parent.Name)
}

func (t *Tree) getOrCreateNode(ctx context.Context, idx int, parent *Node) (node *Node, err error) {
	if len(t.filters) <= idx {
		return
	}

	filter := t.filters[idx]
	name := t.nodeFullName(filter.Name(ctx), parent)

	t.lock.RLock()
	if node = t.getNode(idx, name); node != nil {
		t.lock.RUnlock()
		return
	}
	t.lock.RUnlock()

	t.lock.Lock()
	if len(t.nodes) <= idx {
		t.nodes = append(t.nodes, ccache.New(ccache.Configure().MaxSize(t.Config.MaxSize())))
	}
	if node = t.getNode(idx, name); node != nil {
		t.lock.Unlock()
		return
	}

	node, err = t.createNode(ctx, idx, name, parent)
	if node == nil {
		t.lock.Unlock()
		return
	}
	t.nodes[idx].Set(name, node, t.Config.TTL())
	t.lock.Unlock()
	return
}

func (t *Tree) nodeFullName(name string, parent *Node) string {
	if parent != nil {
		name = parent.Name + "." + name
	}
	return name
}

func (t *Tree) getNode(idx int, name string) *Node {
	if len(t.nodes) <= idx {
		return nil
	}
	item := t.nodes[idx].Get(name)
	if item != nil && !item.Expired() {
		return item.Value().(*Node)
	}
	return nil
}

func (t *Tree) createNode(ctx context.Context, idx int, name string, parent *Node) (node *Node, err error) {
	node, err = t.filters[idx].Init(ctx, parent)
	if node == nil {
		return
	}
	node.Name = name
	node.Tree = t
	node.Level = idx
	if parent != nil {
		parent.Childs = append(parent.Childs, node)
	}
	return
}

// Simulate will create a temp node from tree
func (t *Tree) Simulate(ctx context.Context) (node *Node, err error) {
	var parent *Node
	for idx, filter := range t.filters {
		name := t.nodeFullName(filter.Name(ctx), parent)
		node, err = t.createNode(ctx, idx, name, parent)
		if node == nil {
			return
		}
		parent = node
	}
	return
}

func NewTree(cfg *Config) *Tree {
	return &Tree{Config: cfg}
}
