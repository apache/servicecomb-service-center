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
package etcd

import (
	"github.com/apache/servicecomb-service-center/datasource/etcd/cache"
	"github.com/apache/servicecomb-service-center/server/core"
	"testing"
)

func TestNewKvEntity(t *testing.T) {
	core.ServerInfo.Config.EnableCache = false
	i := NewEtcdAdaptor("a", cache.Configure().WithInitSize(1))
	if _, ok := i.Indexer.(*Indexer); !ok {
		t.Fatalf("TestNewIndexer failed")
	}
	core.ServerInfo.Config.EnableCache = true

	i.Run()
	<-i.Ready()
	i.Stop()

	i = NewEtcdAdaptor("a", cache.Configure().WithInitSize(0))
	if _, ok := i.Indexer.(*Indexer); !ok {
		t.Fatalf("TestNewIndexer failed")
	}

	i = NewEtcdAdaptor("a", cache.Configure())
	if _, ok := i.Indexer.(*CacheIndexer); !ok {
		t.Fatalf("TestNewIndexer failed")
	}
	if _, ok := i.Cacher.(*KvCacher); !ok {
		t.Fatalf("TestNewIndexer failed")
	}
}

func TestNewRepository(t *testing.T) {
	repo := NewRepository()
	i := repo.(*Repository).New(0, cache.Configure())
	if _, ok := i.(*Adaptor); !ok {
		t.Fatalf("TestNewIndexer failed")
	}
}
