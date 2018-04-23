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
package store

import (
	"encoding/json"
	"fmt"
	"github.com/apache/incubator-servicecomb-service-center/pkg/util"
	pb "github.com/apache/incubator-servicecomb-service-center/server/core/proto"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"math/rand"
	"testing"
	"time"
)

func TestInstanceEventDeferHandler_OnCondition(t *testing.T) {
	iedh := &InstanceEventDeferHandler{
		Percent: 0,
	}

	if iedh.OnCondition(nil, nil) {
		fmt.Printf(`TestInstanceEventDeferHandler_OnCondition with 0%% failed`)
		t.FailNow()
	}

	iedh.Percent = 0.01
	if !iedh.OnCondition(nil, nil) {
		fmt.Printf(`TestInstanceEventDeferHandler_OnCondition with 1%% failed`)
		t.FailNow()
	}
}

func TestInstanceEventDeferHandler_HandleChan(t *testing.T) {
	inst := &pb.MicroServiceInstance{
		HealthCheck: &pb.HealthCheck{
			Interval: 4,
			Times:    0,
		},
	}
	b, _ := json.Marshal(inst)
	kv1 := &mvccpb.KeyValue{
		Key:   util.StringToBytesWithNoCopy("/1"),
		Value: b,
	}
	kv2 := &mvccpb.KeyValue{
		Key:   util.StringToBytesWithNoCopy("/2"),
		Value: b,
	}
	kv3 := &mvccpb.KeyValue{
		Key:   util.StringToBytesWithNoCopy("/3"),
		Value: b,
	}
	kv4 := &mvccpb.KeyValue{
		Key:   util.StringToBytesWithNoCopy("/4"),
		Value: b,
	}
	kv5 := &mvccpb.KeyValue{
		Key:   util.StringToBytesWithNoCopy("/5"),
		Value: b,
	}
	kv6 := &mvccpb.KeyValue{
		Key:   util.StringToBytesWithNoCopy("/6"),
		Value: b,
	}

	cache := NewKvCache(nil, 1)
	cache.store["/1"] = kv1
	cache.store["/2"] = kv2
	cache.store["/3"] = kv3
	cache.store["/4"] = kv4
	cache.store["/5"] = kv5
	cache.store["/6"] = kv6

	evts1 := []KvEvent{
		{
			Type:   pb.EVT_CREATE,
			Object: kv1,
		},
		{
			Type:   pb.EVT_UPDATE,
			Object: kv1,
		},
	}
	evts2 := []KvEvent{
		{
			Type:   pb.EVT_DELETE,
			Object: kv2,
		},
		{
			Type:   pb.EVT_DELETE,
			Object: kv3,
		},
		{
			Type:   pb.EVT_DELETE,
			Object: kv4,
		},
		{
			Type:   pb.EVT_DELETE,
			Object: kv5,
		},
		{
			Type:   pb.EVT_DELETE,
			Object: kv6,
		},
	}
	evts3 := []KvEvent{
		{
			Type:   pb.EVT_CREATE,
			Object: kv2,
		},
		{
			Type:   pb.EVT_UPDATE,
			Object: kv4,
		},
		{
			Type:   pb.EVT_UPDATE,
			Object: kv5,
		},
		{
			Type:   pb.EVT_CREATE,
			Object: kv6,
		},
	}

	iedh := &InstanceEventDeferHandler{
		Percent: 0.01,
	}

	iedh.OnCondition(cache, evts1)
	iedh.OnCondition(cache, evts2)
	iedh.OnCondition(cache, evts3)

	getEvents(t, iedh)

	iedh.Percent = 0.8
	iedh.OnCondition(cache, evts1)
	iedh.OnCondition(cache, evts2)
	iedh.OnCondition(cache, evts3)

	getEvents(t, iedh)
}

func getEvents(t *testing.T, iedh *InstanceEventDeferHandler) {
	fmt.Println(time.Now())
	c := time.After(3 * time.Second)
	var evt3 *KvEvent
	for {
		select {
		case evt := <-iedh.HandleChan():
			fmt.Println(time.Now(), evt)
			if string(evt.Object.(*mvccpb.KeyValue).Key) == "/3" {
				evt3 = &evt
				if iedh.Percent == 0.01 && evt.Type == pb.EVT_DELETE {
					fmt.Printf(`TestInstanceEventDeferHandler_HandleChan with 1%% failed`)
					t.FailNow()
				}
			}
			continue
		case <-c:
			if iedh.Percent == 0.8 && evt3 == nil {
				fmt.Printf(`TestInstanceEventDeferHandler_HandleChan with 80%% failed`)
				t.FailNow()
			}
		}
		break
	}
}

func BenchmarkFilter(b *testing.B) {
	inst := &pb.MicroServiceInstance{
		HealthCheck: &pb.HealthCheck{
			Interval: 4,
			Times:    0,
		},
	}
	v, _ := json.Marshal(inst)

	cacher := &KvCacher{}

	n := 300 * 1000 // 30w
	cache := NewKvCache(cacher, n)
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
			cache.store[k] = &mvccpb.KeyValue{
				Key:         util.StringToBytesWithNoCopy(k),
				Value:       v,
				ModRevision: 1,
			}
			items = append(items, &mvccpb.KeyValue{
				Key:         util.StringToBytesWithNoCopy(k),
				Value:       v,
				ModRevision: int64(rand.Int()),
			})
		} else {
			// delete
			cache.store[k] = &mvccpb.KeyValue{
				Key:         util.StringToBytesWithNoCopy(k),
				Value:       v,
				ModRevision: 1,
			}
		}
	}
	cacher.cache = cache

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cacher.filter(1, items)
	}
	b.ReportAllocs()

	// TODO bad performance!!!
	//10	 120612060 ns/op	37128035 B/op	     134 allocs/op
	//
}
