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
	"errors"
	"github.com/karlseguin/ccache"
	"golang.org/x/net/context"
	"sync"
)

var errNilNode = errors.New("nil node")

type Tree struct {
	Config  *Config
	roots   *ccache.Cache
	filters []Filter
	lock    sync.RWMutex
}

func (t *Tree) AddFilter(fs ...Filter) *Tree {
	t.lock.Lock()
	for _, f := range fs {
		t.filters = append(t.filters, f)
	}
	t.lock.Unlock()
	return t
}

func (t *Tree) Get(ctx context.Context, ops ...Option) (node *Node, err error) {
	var op Option
	if len(ops) > 0 {
		op = ops[0]
	}

	var (
		parent *Node
		i      int
	)

	if !op.NoCache {
		if parent, err = t.getOrCreateRoot(ctx); parent == nil {
			return
		}
		i++
		// parent may be a temp root in concurrent scene
	}

	for ; i < len(t.filters); i++ {
		if op.Level > 0 && op.Level == i {
			break
		}
		if parent, err = t.getOrCreateNode(ctx, i, parent); parent == nil {
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

	t.roots.Delete(t.filters[0].Name(ctx))
}

func (t *Tree) getOrCreateRoot(ctx context.Context) (node *Node, err error) {
	if len(t.filters) == 0 {
		return
	}

	filter := t.filters[0]
	name := filter.Name(ctx)
	item, err := t.roots.Fetch(name, t.Config.TTL(), func() (interface{}, error) {
		node, err := t.getOrCreateNode(ctx, 0, nil)
		if err != nil {
			return nil, err
		}
		if node == nil {
			return nil, errNilNode
		}
		return node, nil
	})
	switch err {
	case nil:
		node = item.Value().(*Node)
	case errNilNode:
		err = nil
	}
	return
}

func (t *Tree) getOrCreateNode(ctx context.Context, idx int, parent *Node) (node *Node, err error) {
	filter := t.filters[idx]
	name := t.nodeFullName(filter.Name(ctx), parent)

	if parent == nil {
		// new a temp node
		return t.createNode(ctx, idx, name, parent)
	}

	item, err := parent.Childs.Fetch(name, func() (interface{}, error) {
		node, err := t.createNode(ctx, idx, name, parent)
		if err != nil {
			return nil, err
		}
		if node == nil {
			return nil, errNilNode
		}
		return node, nil
	})
	switch err {
	case nil:
		node = item.(*Node)
	case errNilNode:
		err = nil
	}
	return
}

func (t *Tree) nodeFullName(name string, parent *Node) string {
	if parent != nil {
		name = parent.Name + "." + name
	}
	return name
}

func (t *Tree) createNode(ctx context.Context, idx int, name string, parent *Node) (node *Node, err error) {
	node, err = t.filters[idx].Init(ctx, parent)
	if node == nil {
		return
	}
	node.Name = name
	node.Tree = t
	node.Level = idx
	return
}

func NewTree(cfg *Config) *Tree {
	return &Tree{
		Config: cfg,
		roots:  ccache.New(ccache.Configure().MaxSize(cfg.MaxSize())),
	}
}
