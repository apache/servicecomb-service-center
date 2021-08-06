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
	"math/rand"
	"strconv"
	"testing"

	pb "github.com/go-chassis/cari/discovery"

	"github.com/apache/servicecomb-service-center/datasource/etcd/sd"
	"github.com/apache/servicecomb-service-center/datasource/etcd/value"
	"github.com/apache/servicecomb-service-center/datasource/sdcommon"
	"github.com/apache/servicecomb-service-center/pkg/util"
	"github.com/go-chassis/foundation/gopool"
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

type mockListWatch struct {
	ListResponse  *sdcommon.ListWatchResp
	WatchResponse *sdcommon.ListWatchResp
	Rev           int64
}

func (lw *mockListWatch) List(sdcommon.ListWatchConfig) (*sdcommon.ListWatchResp, error) {
	if lw.ListResponse == nil {
		return nil, fmt.Errorf("list error")
	}
	lw.Rev = lw.ListResponse.Revision
	return lw.ListResponse, nil
}

func (lw *mockListWatch) DoWatch(ctx context.Context, f func(*sdcommon.ListWatchResp)) error {
	if lw.WatchResponse == nil {
		return fmt.Errorf("error")
	}
	if len(lw.WatchResponse.Resources) > 0 {
		lw.Rev = lw.WatchResponse.Resources[0].ModRevision
	}
	f(lw.WatchResponse)
	<-ctx.Done()
	return nil
}

func (lw *mockListWatch) EventBus(op sdcommon.ListWatchConfig) *sdcommon.EventBus {
	return sdcommon.NewEventBus(lw, op)
}

func (lw *mockListWatch) Revision() int64 {
	return lw.Rev
}

func TestNewKvCacher(t *testing.T) {
	lw := &mockListWatch{}

	cr := &KvCacher{
		Cfg:       sd.Configure(),
		ready:     make(chan struct{}),
		lw:        lw,
		goroutine: gopool.New(),
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
		goroutine: gopool.New(),
		cache:     &mockCache{},
	}

	data := &sdcommon.Resource{Key: "ka", Value: []byte("va"), Version: 1, ModRevision: 2}

	test := &sdcommon.ListWatchResp{
		Action:    sdcommon.ActionPUT,
		Revision:  3,
		Resources: []*sdcommon.Resource{data},
	}
	// case: list 1 resp and watch 0 event
	cr.cache.Remove("ka")
	lw.ListResponse = test
	lw.WatchResponse = nil

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
	lw.WatchResponse = nil
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
	for i := 0; i < sdcommon.DefaultForceListInterval; i++ {
		lw.WatchResponse = nil
	}
	evt.KV = nil
	for i := 0; i < sdcommon.DefaultForceListInterval; i++ {
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

	lw.ListResponse = &sdcommon.ListWatchResp{Revision: 5}
	for i := 0; i < sdcommon.DefaultForceListInterval; i++ {
		lw.WatchResponse = nil
	}
	evt.KV = nil
	for i := 0; i < sdcommon.DefaultForceListInterval; i++ {
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
	lw.WatchResponse = test
	lw.ListResponse = nil

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
	lw.WatchResponse = test
	lw.ListResponse = nil

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
	test.Action = sdcommon.ActionDelete
	data.Version = 0
	data.ModRevision = 6
	lw.WatchResponse = test
	lw.ListResponse = nil

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
	lw.WatchResponse = nil
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
	test.Action = sdcommon.ActionDelete
	data.Version = 0
	data.ModRevision = 1
	data.Value = nil
	lw.WatchResponse = test
	lw.ListResponse = nil

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
	lw.WatchResponse = nil
	evt.KV = nil
	old := *cr.Cfg
	cr.Cfg.WithParser(value.MapParser)
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
	lw.WatchResponse = test

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
	test.Resources = nil
	for i := 0; i < sdcommon.EventBlockSize+1; i++ {
		kv := *data
		kv.Key = strconv.Itoa(i)
		kv.Value = []byte(strconv.Itoa(i))
		kv.Version = int64(i)
		kv.ModRevision = int64(i)
		test.Resources = append(test.Resources, &kv)
	}
	data.ModRevision = 2
	test.Resources = append(test.Resources, data)
	lw.ListResponse = test
	lw.WatchResponse = nil
	evt.KV = nil
	old = *cr.Cfg
	cr.Cfg.WithEventFunc(func(evt sd.KvEvent) {
		evts[string(evt.KV.Key)] = evt
	})
	cr.refresh(ctx)
	*cr.Cfg = old
	// check all events
	for i := 0; i < sdcommon.EventBlockSize+1; i++ {
		s := strconv.Itoa(i)
		if evt, ok := evts[s]; !ok || evt.Type != pb.EVT_CREATE || evt.KV.ModRevision != int64(i) || string(evt.KV.Value.([]byte)) != s {
			t.Fatalf("TestNewKvCacher failed, %v", evt)
		}
		delete(evts, s)
	}
	evt = evts[data.Key]
	if len(evts) != 1 || evt.Type != pb.EVT_CREATE || evt.Revision != 3 || evt.KV.ModRevision != 2 {
		t.Fatalf("TestNewKvCacher failed, %v %v", evts, evt)
	}
	delete(evts, data.Key)

	// case: cacher is ready and the next list failed, prevent to watch with rev = 0
	if !cr.IsReady() {
		t.Fatalf("TestNewKvCacher failed")
	}
	lw.Rev = 0            // watch failed
	lw.ListResponse = nil // the next list
	lw.WatchResponse = test
	old = *cr.Cfg
	cr.Cfg.WithEventFunc(func(evt sd.KvEvent) {
		t.Fatalf("TestNewKvCacher failed, %v", evt)
	})
	cr.refresh(ctx)
	*cr.Cfg = old
}

func BenchmarkFilter(b *testing.B) {
	inst := &pb.MicroServiceInstance{
		HealthCheck: &pb.HealthCheck{
			Interval: 4,
			Times:    0,
		},
	}
	v, _ := json.Marshal(inst)

	cfg := sd.Configure().WithParser(value.InstanceParser)

	n := 300 * 1000 // 30w
	cache := sd.NewKvCache("test", cfg)
	items := make([]*sdcommon.Resource, 0, n)
	for ; n > 0; n-- {
		k := fmt.Sprintf("/%d", n)
		if n <= 10*1000 {
			// create
			items = append(items, &sdcommon.Resource{
				Key:         k,
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
			items = append(items, &sdcommon.Resource{
				Key:         k,
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
