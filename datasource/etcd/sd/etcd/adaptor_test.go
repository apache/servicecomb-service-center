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

package etcd_test

import (
	_ "github.com/apache/servicecomb-service-center/test"

	"github.com/apache/servicecomb-service-center/datasource/etcd/sd"

	"testing"

	. "github.com/apache/servicecomb-service-center/datasource/etcd/sd/etcd"
	"github.com/apache/servicecomb-service-center/server/config"
)

func TestNewKvEntity(t *testing.T) {
	config.Server.Config.EnableCache = false
	i := NewEtcdAdaptor("a", sd.Configure().WithInitSize(1))
	if _, ok := i.Indexer.(*Indexer); !ok {
		t.Fatalf("TestNewIndexer failed")
	}
	config.Server.Config.EnableCache = true

	i.Run()
	<-i.Ready()
	i.Stop()

	i = NewEtcdAdaptor("a", sd.Configure().WithInitSize(0))
	if _, ok := i.Indexer.(*Indexer); !ok {
		t.Fatalf("TestNewIndexer failed")
	}

	i = NewEtcdAdaptor("a", sd.Configure())
	if _, ok := i.Indexer.(*CacheIndexer); !ok {
		t.Fatalf("TestNewIndexer failed")
	}
	if _, ok := i.Cacher.(*KvCacher); !ok {
		t.Fatalf("TestNewIndexer failed")
	}
}

func TestNewRepository(t *testing.T) {
	repo := NewRepository(sd.Options{})
	i := repo.(*Repository).New(0, sd.Configure())
	if _, ok := i.(*Adaptor); !ok {
		t.Fatalf("TestNewIndexer failed")
	}
}
