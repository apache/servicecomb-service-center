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
	"fmt"
	"github.com/apache/incubator-servicecomb-service-center/pkg/util"
	"github.com/apache/incubator-servicecomb-service-center/server/core/proto"
	"github.com/apache/incubator-servicecomb-service-center/server/infra/registry"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"golang.org/x/net/context"
	"testing"
)

type mockCache struct {
	*nullCache
	Key string
	KV  *KeyValue
}

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
func (n *mockCache) ForEach(iter func(k string, v *KeyValue) (next bool)) {
	iter(n.Key, n.KV)
}
func (n *mockCache) Put(k string, v *KeyValue) {
	n.Key = k
	n.KV = v
}
func (n *mockCache) Remove(k string) {
	if k == n.Key {
		n.Key = ""
		n.KV = nil
	}
}

func TestNewKvCacher(t *testing.T) {
	w := &mockWatcher{}
	lw := &mockListWatch{
		Bus: make(chan *registry.PluginResponse, 100),
	}
	w.lw = lw

	cr := &KvCacher{
		Cfg:       Configure(),
		ready:     make(chan struct{}),
		lw:        lw,
		goroutine: util.NewGo(context.Background()),
		cache:     &mockCache{},
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// cause internal error
	cr.refresh(ctx)

	<-cr.Ready()
	if !cr.IsReady() {
		t.Fatalf("TestNewKvCacher failed")
	}

	// normal
	var evt KvEvent
	cr = &KvCacher{
		Cfg: Configure().
			WithNoEventPeriods(0).
			WithEventFunc(func(e KvEvent) {
				evt = e
			}),
		ready:     make(chan struct{}),
		lw:        lw,
		goroutine: util.NewGo(context.Background()),
		cache:     &mockCache{},
	}

	lw.Watcher = w
	data := &mvccpb.KeyValue{Key: []byte("ka"), Value: []byte("va"), Version: 1, ModRevision: 2}
	test := &registry.PluginResponse{
		Action:   registry.Put,
		Revision: 3,
		Kvs:      []*mvccpb.KeyValue{data}}

	// case: list 1 resp and watch 0 event
	cr.cache.Remove("ka")
	lw.ListResponse = test
	lw.Bus <- nil

	cr.refresh(ctx)
	// check event
	if evt.Type != proto.EVT_INIT || evt.Revision != 3 || evt.KV.ModRevision != 2 || string(evt.KV.Key) != "ka" || string(evt.KV.Value.([]byte)) != "va" {
		t.Fatalf("TestNewKvCacher failed, %v", evt)
	}
	// check cache
	kv := cr.cache.Get("ka")
	if kv == nil || kv.ModRevision != 2 || string(kv.Key) != "ka" || string(kv.Value.([]byte)) != "va" {
		t.Fatalf("TestNewKvCacher failed")
	}

	test.Revision = 4
	data.ModRevision = 3

	// case: re-list and should be no event
	lw.Bus <- nil
	evt.KV = nil
	cr.refresh(ctx)
	if evt.KV != nil {
		t.Fatalf("TestNewKvCacher failed")
	}
	// check cache
	kv = cr.cache.Get("ka")
	if kv == nil || kv.ModRevision != 2 || string(kv.Key) != "ka" || string(kv.Value.([]byte)) != "va" {
		t.Fatalf("TestNewKvCacher failed")
	}

	// case re-list and over no event times
	lw.Bus <- nil
	evt.KV = nil
	old := *cr.Cfg
	cr.Cfg.WithNoEventPeriods(1)
	cr.refresh(ctx)
	// check event
	if evt.Type != proto.EVT_UPDATE || evt.Revision != 4 || evt.KV.ModRevision != 3 || string(evt.KV.Key) != "ka" || string(evt.KV.Value.([]byte)) != "va" {
		t.Fatalf("TestNewKvCacher failed, %v", evt)
	}
	// check cache
	kv = cr.cache.Get("ka")
	if kv == nil || kv.ModRevision != 3 || string(kv.Key) != "ka" || string(kv.Value.([]byte)) != "va" {
		t.Fatalf("TestNewKvCacher failed")
	}

	lw.ListResponse = &registry.PluginResponse{Revision: 5}
	lw.Bus <- nil
	evt.KV = nil
	cr.refresh(ctx)
	*cr.Cfg = old
	// check event
	if evt.Type != proto.EVT_DELETE || evt.Revision != 5 || evt.KV.ModRevision != 3 || string(evt.KV.Key) != "ka" || string(evt.KV.Value.([]byte)) != "va" {
		t.Fatalf("TestNewKvCacher failed, %v", evt)
	}
	// check cache
	kv = cr.cache.Get("ka")
	if kv != nil {
		t.Fatalf("TestNewKvCacher failed")
	}

	// case: no list and watch 1 event
	test.Revision = 6
	data.ModRevision = 5
	lw.Bus <- test
	lw.Bus <- nil

	cr.refresh(ctx)
	// check event
	if evt.Type != proto.EVT_CREATE || evt.Revision != 5 || evt.KV.ModRevision != 5 || string(evt.KV.Key) != "ka" || string(evt.KV.Value.([]byte)) != "va" {
		t.Fatalf("TestNewKvCacher failed, %v", evt)
	}
	// check cache
	kv = cr.cache.Get("ka")
	if kv == nil || kv.ModRevision != 5 || string(kv.Key) != "ka" || string(kv.Value.([]byte)) != "va" {
		t.Fatalf("TestNewKvCacher failed")
	}

	test.Revision = 7
	data.Version = 2
	data.ModRevision = 6
	lw.Bus <- test
	lw.Bus <- nil

	cr.refresh(ctx)
	// check event
	if evt.Type != proto.EVT_UPDATE || evt.Revision != 6 || evt.KV.ModRevision != 6 || string(evt.KV.Key) != "ka" || string(evt.KV.Value.([]byte)) != "va" {
		t.Fatalf("TestNewKvCacher failed, %v", evt)
	}
	// check cache
	kv = cr.cache.Get("ka")
	if kv == nil || kv.ModRevision != 6 || string(kv.Key) != "ka" || string(kv.Value.([]byte)) != "va" {
		t.Fatalf("TestNewKvCacher failed")
	}

	test.Revision = 8
	test.Action = registry.Delete
	data.Version = 0
	data.ModRevision = 6
	lw.Bus <- test
	lw.Bus <- nil

	cr.refresh(ctx)
	// check event
	if evt.Type != proto.EVT_DELETE || evt.Revision != 6 || evt.KV.ModRevision != 6 || string(evt.KV.Key) != "ka" || string(evt.KV.Value.([]byte)) != "va" {
		t.Fatalf("TestNewKvCacher failed, %v", evt)
	}
	// check cache
	kv = cr.cache.Get("ka")
	if kv != nil {
		t.Fatalf("TestNewKvCacher failed")
	}

	// case: list 1 item and watch no event
	lw.Rev = 0
	test.Revision = 9
	data.Version = 1
	data.ModRevision = 1
	lw.ListResponse = test
	lw.Bus <- nil
	evt.KV = nil
	cr.refresh(ctx)
	// check event
	if evt.Type != proto.EVT_CREATE || evt.Revision != 9 || evt.KV.ModRevision != 1 || string(evt.KV.Key) != "ka" || string(evt.KV.Value.([]byte)) != "va" {
		t.Fatalf("TestNewKvCacher failed, %v", evt)
	}
	// check cache
	kv = cr.cache.Get("ka")
	if kv == nil || kv.ModRevision != 1 || string(kv.Key) != "ka" || string(kv.Value.([]byte)) != "va" {
		t.Fatalf("TestNewKvCacher failed")
	}

	// case: caught delete event but value is nil
	test.Revision = 10
	test.Action = registry.Delete
	data.Version = 0
	data.ModRevision = 1
	data.Value = nil
	lw.Bus <- test
	lw.Bus <- nil

	cr.refresh(ctx)
	data.Value = []byte("va")
	// check event
	if evt.Type != proto.EVT_DELETE || evt.Revision != 1 || evt.KV.ModRevision != 1 || string(evt.KV.Key) != "ka" || string(evt.KV.Value.([]byte)) != "va" {
		t.Fatalf("TestNewKvCacher failed, %v", evt)
	}
	// check cache
	kv = cr.cache.Get("ka")
	if kv != nil {
		t.Fatalf("TestNewKvCacher failed")
	}

	// case: parse failed
	lw.Rev = 0
	test.Revision = 1
	data.Version = 1
	data.ModRevision = 1
	lw.ListResponse = test
	lw.Bus <- nil
	evt.KV = nil
	old = *cr.Cfg
	cr.Cfg.WithParser(MapParser)
	cr.refresh(ctx)
	// check event
	if evt.KV != nil {
		t.Fatalf("TestNewKvCacher failed, %v", evt)
	}
	// check cache
	kv = cr.cache.Get("ka")
	if kv != nil {
		t.Fatalf("TestNewKvCacher failed")
	}

	lw.ListResponse = test
	lw.Bus <- test
	lw.Bus <- nil
	cr.refresh(ctx)
	*cr.Cfg = old
	// check event
	if evt.KV != nil {
		t.Fatalf("TestNewKvCacher failed, %v", evt)
	}
	// check cache
	kv = cr.cache.Get("ka")
	if kv != nil {
		t.Fatalf("TestNewKvCacher failed")
	}

	// case: list result is too large
	var evts = make(map[string]KvEvent)
	lw.Rev = 0
	test.Revision = 3
	test.Kvs = nil
	for i := 0; i < eventBlockSize+1; i++ {
		kv := *data
		kv.Key = []byte(fmt.Sprint(i))
		kv.Value = []byte(fmt.Sprint(i))
		kv.Version = int64(i)
		kv.ModRevision = int64(i)
		test.Kvs = append(test.Kvs, &kv)
	}
	data.ModRevision = 2
	test.Kvs = append(test.Kvs, data)
	lw.ListResponse = test
	lw.Bus <- nil
	evt.KV = nil
	old = *cr.Cfg
	cr.Cfg.WithEventFunc(func(evt KvEvent) {
		evts[string(evt.KV.Key)] = evt
	})
	cr.refresh(ctx)
	// check all events
	for i := 0; i < eventBlockSize+1; i++ {
		s := fmt.Sprint(i)
		if evt, ok := evts[s]; !ok || evt.Type != proto.EVT_CREATE || evt.KV.ModRevision != int64(i) || string(evt.KV.Value.([]byte)) != s {
			t.Fatalf("TestNewKvCacher failed, %v", evt)
		}
		delete(evts, s)
	}
	evt = evts[string(data.Key)]
	if len(evts) != 1 || evt.Type != proto.EVT_CREATE || evt.Revision != 3 || evt.KV.ModRevision != 2 {
		t.Fatalf("TestNewKvCacher failed, %v %v", evts, evt)
	}
	delete(evts, string(data.Key))

}
