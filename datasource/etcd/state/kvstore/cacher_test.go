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
	"testing"

	"github.com/go-chassis/cari/discovery"
)

func TestNewCommonCacher(t *testing.T) {
	var e Event
	cfg := NewOptions()
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

	cacher.Notify(discovery.EVT_CREATE, "/a", &KeyValue{Version: 1})
	if e.Type == discovery.EVT_CREATE || cache.Get("/a").Version != 1 {
		t.Fatalf("TestNewCommonCacher failed")
	}
	cfg.WithEventFunc(func(evt Event) {
		e = evt
	})
	cacher.Notify(discovery.EVT_CREATE, "/a", &KeyValue{Version: 1})
	if e.Type != discovery.EVT_CREATE || cache.Get("/a").Version != 1 {
		t.Fatalf("TestNewCommonCacher failed")
	}
	cacher.Notify(discovery.EVT_DELETE, "/a", &KeyValue{Version: 1})
	if e.Type != discovery.EVT_DELETE || cache.Get("/a") != nil {
		t.Fatalf("TestNewCommonCacher failed")
	}
	cacher.Stop()
}
