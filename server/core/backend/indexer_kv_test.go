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
package backend

import (
	"github.com/apache/incubator-servicecomb-service-center/pkg/util"
	"github.com/apache/incubator-servicecomb-service-center/server/infra/registry"
	"golang.org/x/net/context"
	"testing"
)

func TestCacheIndexer_Search(t *testing.T) {
	indexer := &CacheIndexer{
		baseIndexer: &baseIndexer{
			Cfg: DefaultConfig().WithPrefix("/a/"),
		},
		prefixIndex: map[string]map[string]*KeyValue{},
	}

	v1 := &KeyValue{Version: 1}
	v2 := &KeyValue{Version: 2}
	v3 := &KeyValue{Version: 3}
	indexer.addPrefixKey("/a/b/c/", "/a/b/c/1", v1)
	indexer.addPrefixKey("/a/b/c/", "/a/b/c/2", v2)
	indexer.addPrefixKey("/a/f/g/", "/a/f/g/3", v3)

	arr := []*KeyValue{}
	c := indexer.getPrefixKey(&arr, "/a/b")
	if c != 0 {
		t.Fatalf("TestCacheIndexer getPrefixKey failed, %d", c)
	}

	arr = []*KeyValue{}
	c = indexer.getPrefixKey(&arr, "/a/b/")
	if c != 2 {
		t.Fatalf("TestCacheIndexer getPrefixKey failed, %d", c)
	}

	if arr[0].Version != 1 && arr[1].Version != 1 {
		t.Fatalf("TestCacheIndexer getPrefixKey failed, %v", arr)
	}

	arr = []*KeyValue{}
	c = indexer.getPrefixKey(&arr, "/a/")
	if c != 3 {
		t.Fatalf("TestCacheIndexer getPrefixKey failed, %d", c)
	}

	indexer.deletePrefixKey("/a/b/c/", "/a/b/c/2")
	arr = []*KeyValue{}
	c = indexer.getPrefixKey(&arr, "/a/b/")
	if c != 1 {
		t.Fatalf("TestCacheIndexer getPrefixKey failed, %d", c)
	}

	indexer.deletePrefixKey("/a/b/c/", "/a/b/c/2")

	resp := indexer.searchByPrefix(registry.PluginOp{
		Key:       []byte("/a/"),
		CountOnly: true,
	})
	if resp.Count != 2 || len(resp.Kvs) > 0 {
		t.Fatalf("TestCacheIndexer searchByPrefix failed, %v", resp)
	}

	resp = indexer.searchByPrefix(registry.PluginOp{
		Key: []byte("/a/"),
	})
	if resp.Count != 2 || len(resp.Kvs) != 2 {
		t.Fatalf("TestCacheIndexer searchByPrefix failed, %v", resp)
	}

	resp, err := indexer.Search(util.SetContext(context.Background(), CTX_CACHEONLY, "1"),
		registry.WithKey([]byte("/a/b/c/1")))
	if err != nil || resp.Count != 1 || resp.Kvs[0].Version != 1 {
		t.Fatalf("TestCacheIndexer Search failed, %v", err)
	}
}
