// Licensed to the Apache Software Foundation (ASF) under one or more
// contributor license agreements.  See the NOTICE file distributed with
// this work for additional information regarding copyright ownership.
// The ASF licenses this file to You under the Apache License, Version 2.0
// (the "License"); you may not use this file except in compliance with
// the License.  You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package discovery

import (
	"github.com/apache/servicecomb-service-center/pkg/registry"
	"testing"
)

func TestNewCommonCacher(t *testing.T) {
	var e KvEvent
	cfg := Configure()
	cache := NewKvCache("test", cfg)
	cacher := NewCommonCacher(cfg, cache)

	if cacher.Cache() != cache {
		t.Fatalf("TestNewCommonCacher failed")
	}
	select {
	case <-cacher.Ready():
		t.Fatalf("TestNewCommonCacher failed")
	default:
		cacher.Run()
	}
	select {
	case <-cacher.Ready():
	default:
		t.Fatalf("TestNewCommonCacher failed")
	}

	cacher.Notify(registry.EVT_CREATE, "/a", &KeyValue{Version: 1})
	if e.Type == registry.EVT_CREATE || cache.Get("/a").Version != 1 {
		t.Fatalf("TestNewCommonCacher failed")
	}
	cfg.WithEventFunc(func(evt KvEvent) {
		e = evt
	})
	cacher.Notify(registry.EVT_CREATE, "/a", &KeyValue{Version: 1})
	if e.Type != registry.EVT_CREATE || cache.Get("/a").Version != 1 {
		t.Fatalf("TestNewCommonCacher failed")
	}
	cacher.Notify(registry.EVT_DELETE, "/a", &KeyValue{Version: 1})
	if e.Type != registry.EVT_DELETE || cache.Get("/a") != nil {
		t.Fatalf("TestNewCommonCacher failed")
	}
	cacher.Stop()
}
