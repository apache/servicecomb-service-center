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
package kv

import (
	"fmt"
	"github.com/apache/servicecomb-service-center/datasource/etcd/pkg/sd"
	pb "github.com/apache/servicecomb-service-center/pkg/registry"
	"github.com/apache/servicecomb-service-center/pkg/util"
	"testing"
	"time"
)

type mockCache struct {
	c map[string]*sd.KeyValue
}

func (n *mockCache) Name() string              { return "mock" }
func (n *mockCache) Size() int                 { return 0 }
func (n *mockCache) Get(k string) *sd.KeyValue { return nil }
func (n *mockCache) GetAll(arr *[]*sd.KeyValue) (i int) {
	for range n.c {
		i++
	}
	return i
}
func (n *mockCache) GetPrefix(prefix string, arr *[]*sd.KeyValue) int        { return 0 }
func (n *mockCache) ForEach(iter func(k string, v *sd.KeyValue) (next bool)) {}
func (n *mockCache) Put(k string, v *sd.KeyValue)                            { n.c[k] = v }

func TestInstanceEventDeferHandler_OnCondition(t *testing.T) {
	iedh := &InstanceEventDeferHandler{
		Percent: 0,
	}

	if iedh.OnCondition(nil, nil) {
		t.Fatalf(`TestInstanceEventDeferHandler_OnCondition with 0%% failed`)
	}

	iedh.Percent = 0.01
	if !iedh.OnCondition(nil, nil) {
		t.Fatalf(`TestInstanceEventDeferHandler_OnCondition with 1%% failed`)
	}
}

func TestInstanceEventDeferHandler_HandleChan(t *testing.T) {
	b := &pb.MicroServiceInstance{
		HealthCheck: &pb.HealthCheck{
			Interval: 3,
			Times:    0,
		},
	}
	kv1 := &sd.KeyValue{
		Key:   util.StringToBytesWithNoCopy("/1"),
		Value: b,
	}
	kv2 := &sd.KeyValue{
		Key:   util.StringToBytesWithNoCopy("/2"),
		Value: b,
	}
	kv3 := &sd.KeyValue{
		Key:   util.StringToBytesWithNoCopy("/3"),
		Value: b,
	}
	kv4 := &sd.KeyValue{
		Key:   util.StringToBytesWithNoCopy("/4"),
		Value: b,
	}
	kv5 := &sd.KeyValue{
		Key:   util.StringToBytesWithNoCopy("/5"),
		Value: b,
	}
	kv6 := &sd.KeyValue{
		Key:   util.StringToBytesWithNoCopy("/6"),
		Value: b,
	}

	c := &mockCache{c: make(map[string]*sd.KeyValue)}
	c.Put("/1", kv1)
	evts0 := []sd.KvEvent{
		{
			Type: pb.EVT_DELETE,
			KV:   kv1,
		},
	}

	iedh := &InstanceEventDeferHandler{
		Percent: 1,
	}
	iedh.OnCondition(c, evts0)
	select {
	case evt := <-iedh.HandleChan():
		if string(evt.KV.Key) != "/1" || evt.Type != pb.EVT_DELETE {
			t.Fatalf(`TestInstanceEventDeferHandler_HandleChan DELETE failed`)
		}
	case <-time.After(deferCheckWindow + time.Second):
		t.Fatalf(`TestInstanceEventDeferHandler_HandleChan DELETE timed out`)
	}

	c.Put("/1", kv1)
	c.Put("/2", kv2)
	c.Put("/3", kv3)
	c.Put("/4", kv4)
	c.Put("/5", kv5)
	c.Put("/6", kv6)

	evts1 := []sd.KvEvent{
		{
			Type: pb.EVT_CREATE,
			KV:   kv1,
		},
		{
			Type: pb.EVT_UPDATE,
			KV:   kv1,
		},
	}
	evts2 := []sd.KvEvent{
		{
			Type: pb.EVT_DELETE,
			KV:   kv2,
		},
		{
			Type: pb.EVT_DELETE,
			KV:   kv3,
		},
		{
			Type: pb.EVT_DELETE,
			KV:   kv4,
		},
		{
			Type: pb.EVT_DELETE,
			KV:   kv5,
		},
		{
			Type: pb.EVT_DELETE,
			KV:   kv6,
		},
	}
	evts3 := []sd.KvEvent{
		{
			Type: pb.EVT_CREATE,
			KV:   kv2,
		},
		{
			Type: pb.EVT_UPDATE,
			KV:   kv4,
		},
		{
			Type: pb.EVT_UPDATE,
			KV:   kv5,
		},
		{
			Type: pb.EVT_CREATE,
			KV:   kv6,
		},
	}

	iedh.Percent = 0.01
	iedh.OnCondition(c, evts1)
	iedh.OnCondition(c, evts2)
	iedh.OnCondition(c, evts3)

	getEvents(t, iedh)

	iedh.Percent = 0.9
	iedh.OnCondition(c, evts1)
	iedh.OnCondition(c, evts2)
	iedh.OnCondition(c, evts3)

	getEvents(t, iedh)
}

func getEvents(t *testing.T, iedh *InstanceEventDeferHandler) {
	fmt.Println(time.Now())
	c := time.After(3500 * time.Millisecond)
	var evt3 *sd.KvEvent
	for {
		select {
		case evt := <-iedh.HandleChan():
			fmt.Println(time.Now(), evt.Type, string(evt.KV.Key))
			if string(evt.KV.Key) == "/3" {
				evt3 = &evt
				if iedh.Percent == 0.01 && evt.Type == pb.EVT_DELETE {
					t.Fatalf(`TestInstanceEventDeferHandler_HandleChan with 1%% failed`)
				}
			}
			continue
		case <-c:
			if iedh.Percent == 0.9 && evt3 == nil {
				t.Fatalf(`TestInstanceEventDeferHandler_HandleChan with 90%% failed`)
			}
		}
		break
	}
}

func TestConvert(t *testing.T) {
	value := sd.NewKeyValue()
	_, ok := value.Value.(*pb.MicroServiceInstance)
	if ok {
		t.Fatal("TestConvert failed")
	}

	var inst *pb.MicroServiceInstance = nil
	value.Value = inst
	_, ok = value.Value.(*pb.MicroServiceInstance)
	if !ok {
		t.Fatal("TestConvert failed")
	}
}
