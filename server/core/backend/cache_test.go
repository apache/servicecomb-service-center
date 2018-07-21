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
	"encoding/json"
	"fmt"
	"github.com/apache/incubator-servicecomb-service-center/pkg/util"
	pb "github.com/apache/incubator-servicecomb-service-center/server/core/proto"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"math/rand"
	"testing"
)

func BenchmarkFilter(b *testing.B) {
	inst := &pb.MicroServiceInstance{
		HealthCheck: &pb.HealthCheck{
			Interval: 4,
			Times:    0,
		},
	}
	v, _ := json.Marshal(inst)

	cfg := Configure().WithParser(InstanceParser)

	n := 300 * 1000 // 30w
	cache := NewKvCache("test", cfg)
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
			cache.Put(k, &KeyValue{
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
			cache.Put(k, &KeyValue{
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
		cacher.filter(1, items)
	}
	b.ReportAllocs()

	// TODO bad performance!!!
	//20	  82367261 ns/op	37964987 B/op	   80132 allocs/op
}

func TestKvCache_Get(t *testing.T) {
	c := NewKvCache("test", Configure())
	c.Put("", &KeyValue{Version: 1})
	c.Put("/", &KeyValue{Version: 1})
	c.Put("/a/b/c/d/e/1", &KeyValue{Version: 1})
	c.Put("/a/b/c/d/e/2", &KeyValue{Version: 2})
	c.Put("/a/b/d/d/f/3", &KeyValue{Version: 3})
	c.Put("/a/b/e/d/g/4", &KeyValue{Version: 4})

	if s := c.Size(); s == 0 {
		t.Fatalf("TestKvCache Size() failed, %d", s)
	}

	if l := c.GetAll(nil); l != 4 {
		t.Fatalf("TestKvCache GetAll() failed, %d", l)
	}

	if kv := c.Get("/a/b/c/d/e/2"); kv == nil || kv.Version != 2 {
		t.Fatalf("TestKvCache Get() failed, %v", kv)
	}

	if l := c.GetPrefix("/", nil); l != 4 {
		t.Fatalf("TestKvCache GetPrefix() failed, %d", l)
	}

	var arr []*KeyValue
	if l := c.GetPrefix("/a/b/c/", &arr); l != 2 || (arr[0].Version != 1 && arr[1].Version != 1) {
		t.Fatalf("TestKvCache GetPrefix() failed, %d, %v", l, arr)
	}

	l, b := -1, false
	c.ForEach(func(k string, v *KeyValue) (next bool) {
		next = false
		l++
		return
	})
	if l != 0 {
		t.Fatalf("TestKvCache ForEach() failed, %d", l)
	}
	c.ForEach(func(k string, v *KeyValue) (next bool) {
		next = true
		l++
		if v.Version == 4 {
			b = true
		}
		return
	})
	if l != 4 || !b {
		t.Fatalf("TestKvCache ForEach() failed, %d, %v", l, b)
	}

	c.Remove("")
	c.Remove("/")
	c.Remove("/a/b/c/d/e/2")
	c.Remove("/a/b/d/d/f/3")
	if l := c.GetAll(nil); l != 2 {
		t.Fatalf("TestKvCache GetAll() failed, %d", l)
	}

	c.Put("/a/b/c/d/e/1", &KeyValue{Version: 2})
	if kv := c.Get("/a/b/c/d/e/1"); kv == nil || kv.Version != 2 {
		t.Fatalf("TestKvCache Put() failed, %v", kv)
	}
}
