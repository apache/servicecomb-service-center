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
	"context"
	"encoding/json"
	"fmt"
	"github.com/apache/servicecomb-service-center/datasource/etcd/client"
	"github.com/apache/servicecomb-service-center/datasource/etcd/sd"
	"github.com/apache/servicecomb-service-center/pkg/gopool"
	pb "github.com/apache/servicecomb-service-center/pkg/registry"
	"github.com/apache/servicecomb-service-center/pkg/util"
	"github.com/apache/servicecomb-service-center/server/core/proto"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"math/rand"
	"strconv"
	"testing"
)

type mockCache struct {
	Key string
	KV  *sd.KeyValue
}

func (n *mockCache) Name() string                   { return "NULL" }
func (n *mockCache) Size() int                      { return 0 }
func (n *mockCache) GetAll(arr *[]*sd.KeyValue) int { return 0 }
func (n *mockCache) Get(k string) *sd.KeyValue {
	if k == n.Key {
		return n.KV
	}
	return nil
}
func (n *mockCache) GetPrefix(prefix string, arr *[]*sd.KeyValue) int {
	if prefix == n.Key {
		if arr != nil {
			*arr = append(*arr, n.KV)
		}
		return 1
	}
	return 0
}
func (n *mockCache) ForEach(iter func(k string, v *sd.KeyValue) (next bool)) {
	iter(n.Key, n.KV)
}
func (n *mockCache) Put(k string, v *sd.KeyValue) {
	n.Key = k
	n.KV = v
}
func (n *mockCache) Remove(k string) {
	if k == n.Key {
		n.Key = ""
		n.KV = nil
	}
}
func (n *mockCache) MarkDirty()  {}
func (n *mockCache) Dirty() bool { return false }
func (n *mockCache) Clear()      {}

func TestNewKvCacher(t *testing.T) {
	w := &mockWatcher{}
	lw := &mockListWatch{
		Bus: make(chan *client.PluginResponse, 100),
	}
	w.lw = lw

	cr := &KvCacher{
		Cfg:       sd.Configure(),
		ready:     make(chan struct{}),
		lw:        lw,
		goroutine: gopool.New(context.Background()),
		cache:     &mockCache{},
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// case: cause list internal error before initialized
	cr.refresh(ctx)
	if cr.IsReady() {
		t.Fatalf("TestNewKvCacher failed")
	}

	// prepare data
	var evt sd.KvEvent
	cr = &KvCacher{
		Cfg: sd.Configure().
			WithEventFunc(func(e sd.KvEvent) {
				evt = e
			}),
		ready:     make(chan struct{}),
		lw:        lw,
		goroutine: gopool.New(context.Background()),
		cache:     &mockCache{},
	}

	lw.Watcher = w
	data := &mvccpb.KeyValue{Key: []byte("ka"), Value: []byte("va"), Version: 1, ModRevision: 2}
	test := &client.PluginResponse{
		Action:   client.Put,
		Revision: 3,
		Kvs:      []*mvccpb.KeyValue{data}}

	// case: list 1 resp and watch 0 event
	cr.cache.Remove("ka")
	lw.ListResponse = test
	lw.Bus <- nil

	cr.refresh(ctx)
	// check ready
	if !cr.IsReady() {
		t.Fatalf("TestNewKvCacher failed")
	}
	// check event
	if evt.Type != pb.EVT_INIT || evt.Revision != 3 || evt.KV.ModRevision != 2 || string(evt.KV.Key) != "ka" || string(evt.KV.Value.([]byte)) != "va" {
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
	for i := 0; i < DefaultForceListInterval; i++ {
		lw.Bus <- nil
	}
	evt.KV = nil
	for i := 0; i < DefaultForceListInterval; i++ {
		cr.refresh(ctx)
	}
	// check event
	if evt.Type != pb.EVT_UPDATE || evt.Revision != 4 || evt.KV.ModRevision != 3 || string(evt.KV.Key) != "ka" || string(evt.KV.Value.([]byte)) != "va" {
		t.Fatalf("TestNewKvCacher failed, %v", evt)
	}
	// check cache
	kv = cr.cache.Get("ka")
	if kv == nil || kv.ModRevision != 3 || string(kv.Key) != "ka" || string(kv.Value.([]byte)) != "va" {
		t.Fatalf("TestNewKvCacher failed")
	}

	lw.ListResponse = &client.PluginResponse{Revision: 5}
	for i := 0; i < DefaultForceListInterval; i++ {
		lw.Bus <- nil
	}
	evt.KV = nil
	for i := 0; i < DefaultForceListInterval; i++ {
		cr.refresh(ctx)
	}
	// check event
	if evt.Type != pb.EVT_DELETE || evt.Revision != 5 || evt.KV.ModRevision != 3 || string(evt.KV.Key) != "ka" || string(evt.KV.Value.([]byte)) != "va" {
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
	if evt.Type != pb.EVT_CREATE || evt.Revision != 5 || evt.KV.ModRevision != 5 || string(evt.KV.Key) != "ka" || string(evt.KV.Value.([]byte)) != "va" {
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
	if evt.Type != pb.EVT_UPDATE || evt.Revision != 6 || evt.KV.ModRevision != 6 || string(evt.KV.Key) != "ka" || string(evt.KV.Value.([]byte)) != "va" {
		t.Fatalf("TestNewKvCacher failed, %v", evt)
	}
	// check cache
	kv = cr.cache.Get("ka")
	if kv == nil || kv.ModRevision != 6 || string(kv.Key) != "ka" || string(kv.Value.([]byte)) != "va" {
		t.Fatalf("TestNewKvCacher failed")
	}

	test.Revision = 8
	test.Action = client.Delete
	data.Version = 0
	data.ModRevision = 6
	lw.Bus <- test
	lw.Bus <- nil

	cr.refresh(ctx)
	// check event
	if evt.Type != pb.EVT_DELETE || evt.Revision != 6 || evt.KV.ModRevision != 6 || string(evt.KV.Key) != "ka" || string(evt.KV.Value.([]byte)) != "va" {
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
	if evt.Type != pb.EVT_CREATE || evt.Revision != 9 || evt.KV.ModRevision != 1 || string(evt.KV.Key) != "ka" || string(evt.KV.Value.([]byte)) != "va" {
		t.Fatalf("TestNewKvCacher failed, %v", evt)
	}
	// check cache
	kv = cr.cache.Get("ka")
	if kv == nil || kv.ModRevision != 1 || string(kv.Key) != "ka" || string(kv.Value.([]byte)) != "va" {
		t.Fatalf("TestNewKvCacher failed")
	}

	// case: caught delete event but value is nil
	test.Revision = 10
	test.Action = client.Delete
	data.Version = 0
	data.ModRevision = 1
	data.Value = nil
	lw.Bus <- test
	lw.Bus <- nil

	cr.refresh(ctx)
	data.Value = []byte("va")
	// check event
	if evt.Type != pb.EVT_DELETE || evt.Revision != 1 || evt.KV.ModRevision != 1 || string(evt.KV.Key) != "ka" || string(evt.KV.Value.([]byte)) != "va" {
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
	old := *cr.Cfg
	cr.Cfg.WithParser(proto.MapParser)
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
	var evts = make(map[string]sd.KvEvent)
	lw.Rev = 0
	test.Revision = 3
	test.Kvs = nil
	for i := 0; i < eventBlockSize+1; i++ {
		kv := *data
		kv.Key = []byte(strconv.Itoa(i))
		kv.Value = []byte(strconv.Itoa(i))
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
	cr.Cfg.WithEventFunc(func(evt sd.KvEvent) {
		evts[string(evt.KV.Key)] = evt
	})
	cr.refresh(ctx)
	*cr.Cfg = old
	// check all events
	for i := 0; i < eventBlockSize+1; i++ {
		s := strconv.Itoa(i)
		if evt, ok := evts[s]; !ok || evt.Type != pb.EVT_CREATE || evt.KV.ModRevision != int64(i) || string(evt.KV.Value.([]byte)) != s {
			t.Fatalf("TestNewKvCacher failed, %v", evt)
		}
		delete(evts, s)
	}
	evt = evts[string(data.Key)]
	if len(evts) != 1 || evt.Type != pb.EVT_CREATE || evt.Revision != 3 || evt.KV.ModRevision != 2 {
		t.Fatalf("TestNewKvCacher failed, %v %v", evts, evt)
	}
	delete(evts, string(data.Key))

	// case: cacher is ready and the next list failed, prevent to watch with rev = 0
	if !cr.IsReady() {
		t.Fatalf("TestNewKvCacher failed")
	}
	lw.Rev = 0            // watch failed
	lw.ListResponse = nil // the next list
	lw.Bus <- test
	old = *cr.Cfg
	cr.Cfg.WithEventFunc(func(evt sd.KvEvent) {
		t.Fatalf("TestNewKvCacher failed, %v", evt)
	})
	cr.refresh(ctx)
	*cr.Cfg = old
	<-lw.Bus
}

func BenchmarkFilter(b *testing.B) {
	inst := &pb.MicroServiceInstance{
		HealthCheck: &pb.HealthCheck{
			Interval: 4,
			Times:    0,
		},
	}
	v, _ := json.Marshal(inst)

	cfg := sd.Configure().WithParser(proto.InstanceParser)

	n := 300 * 1000 // 30w
	cache := sd.NewKvCache("test", cfg)
	items := make([]*mvccpb.KeyValue, 0, n)
	for ; n > 0; n-- {
		k := fmt.Sprintf("/%d", n)
		if n <= 10*1000 {
			// create
			items = append(items, &mvccpb.KeyValue{
				Key:         util.StringToBytesWithNoCopy(k),
				Value:       v,
				ModRevision: int64(rand.Int()),
			})
		} else if n > 100*1000 && n <= 20*1000 {
			// update
			cache.Put(k, &sd.KeyValue{
				Key:         util.StringToBytesWithNoCopy(k),
				Value:       inst,
				ModRevision: 1,
			})
			items = append(items, &mvccpb.KeyValue{
				Key:         util.StringToBytesWithNoCopy(k),
				Value:       v,
				ModRevision: int64(rand.Int()),
			})
		} else {
			// delete
			cache.Put(k, &sd.KeyValue{
				Key:         util.StringToBytesWithNoCopy(k),
				Value:       inst,
				ModRevision: 1,
			})
		}
	}
	cacher := &KvCacher{Cfg: cfg}
	cacher.cache = cache

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = cacher.filter(1, items)
	}
	b.ReportAllocs()

	// TODO bad performance!!!
	//10	 167974203 ns/op	92637508 B/op	   80028 allocs/op
}
