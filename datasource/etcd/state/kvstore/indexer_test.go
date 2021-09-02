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

package kvstore

import (
	"context"
	"testing"

	"github.com/little-cui/etcdadpt"
)

type mockCache struct {
	Key string
	KV  *KeyValue
}

func (n *mockCache) Name() string                { return "NULL" }
func (n *mockCache) Size() int                   { return 0 }
func (n *mockCache) GetAll(arr *[]*KeyValue) int { return 0 }
func (n *mockCache) Get(k string) *KeyValue {
	if k == n.Key {
		return n.KV
	}
	return nil
}
func (n *mockCache) GetPrefix(prefix string, arr *[]*KeyValue) int {
	if prefix == n.Key {
		if arr != nil {
			*arr = append(*arr, n.KV)
		}
		return 1
	}
	return 0
}
func (n *mockCache) ForEach(iter func(k string, v *KeyValue) (next bool)) {}
func (n *mockCache) Put(k string, v *KeyValue) {
	n.Key = k
	n.KV = v
}

func TestCacheIndexer_Search(t *testing.T) {
	c := &mockCache{}
	i := NewCacheIndexer(c)

	// not match cache
	c.Put("ka", &KeyValue{Key: []byte("ka"), Value: []byte("va"), Version: 1, ModRevision: 1})
	resp, err := i.Search(context.Background(), etcdadpt.WithStrKey("/a"))
	if err != nil || resp == nil || resp.Count != 0 {
		t.Fatalf("TestEtcdIndexer_Search failed, %v, %v", err, resp)
	}
	resp, err = i.Search(context.Background(), etcdadpt.WithStrKey("/a"), etcdadpt.WithPrefix())
	if err != nil || resp == nil || resp.Count != 0 {
		t.Fatalf("TestEtcdIndexer_Search failed, %v, %v", err, resp)
	}
	resp, err = i.Search(context.Background(), etcdadpt.WithStrKey("/a"), etcdadpt.WithCountOnly())
	if err != nil || resp == nil || resp.Count != 0 {
		t.Fatalf("TestEtcdIndexer_Search failed, %v, %v", err, resp)
	}

	// case: use cache index
	c.Put("/a", &KeyValue{Key: []byte("/a"), Value: []byte("va"), Version: 1, ModRevision: 1})

	// exact match
	resp, err = i.Search(context.Background(), etcdadpt.WithStrKey("/a"))
	if err != nil || resp == nil || resp.Count != 1 || string(resp.Kvs[0].Value.([]byte)) != "va" {
		t.Fatalf("TestEtcdIndexer_Search failed, %v, %v", err, resp)
	}
	resp, err = i.Search(context.Background(), etcdadpt.WithStrKey("/a"), etcdadpt.WithCountOnly())
	if err != nil || resp == nil || resp.Count != 1 || len(resp.Kvs) != 0 {
		t.Fatalf("TestEtcdIndexer_Search failed, %v, %v", err, resp)
	}

	// prefix match
	resp, err = i.Search(context.Background(), etcdadpt.WithStrKey("/a"), etcdadpt.WithPrefix())
	if err != nil || resp == nil || resp.Count != 1 || string(resp.Kvs[0].Value.([]byte)) != "va" {
		t.Fatalf("TestEtcdIndexer_Search failed, %v, %v", err, resp)
	}
	resp, err = i.Search(context.Background(), etcdadpt.WithStrKey("/a"), etcdadpt.WithPrefix(), etcdadpt.WithCountOnly())
	if err != nil || resp == nil || resp.Count != 1 || len(resp.Kvs) != 0 {
		t.Fatalf("TestEtcdIndexer_Search failed, %v, %v", err, resp)
	}
}
