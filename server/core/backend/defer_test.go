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
	pb "github.com/apache/incubator-servicecomb-service-center/server/core/proto"
	"testing"
	"time"
)

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
	kv1 := &KeyValue{
		Key:   util.StringToBytesWithNoCopy("/1"),
		Value: b,
	}
	kv2 := &KeyValue{
		Key:   util.StringToBytesWithNoCopy("/2"),
		Value: b,
	}
	kv3 := &KeyValue{
		Key:   util.StringToBytesWithNoCopy("/3"),
		Value: b,
	}
	kv4 := &KeyValue{
		Key:   util.StringToBytesWithNoCopy("/4"),
		Value: b,
	}
	kv5 := &KeyValue{
		Key:   util.StringToBytesWithNoCopy("/5"),
		Value: b,
	}
	kv6 := &KeyValue{
		Key:   util.StringToBytesWithNoCopy("/6"),
		Value: b,
	}

	cache := NewKvCache("test", Configure())
	cache.Put("/1", kv1)
	cache.Put("/2", kv2)
	cache.Put("/3", kv3)
	cache.Put("/4", kv4)
	cache.Put("/5", kv5)
	cache.Put("/6", kv6)

	evts1 := []KvEvent{
		{
			Type: pb.EVT_CREATE,
			KV:   kv1,
		},
		{
			Type: pb.EVT_UPDATE,
			KV:   kv1,
		},
	}
	evts2 := []KvEvent{
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
	evts3 := []KvEvent{
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

	iedh := &InstanceEventDeferHandler{
		Percent: 0.01,
	}

	iedh.OnCondition(cache, evts1)
	iedh.OnCondition(cache, evts2)
	iedh.OnCondition(cache, evts3)

	getEvents(t, iedh)

	iedh.Percent = 0.9
	iedh.OnCondition(cache, evts1)
	iedh.OnCondition(cache, evts2)
	iedh.OnCondition(cache, evts3)

	getEvents(t, iedh)
}

func getEvents(t *testing.T, iedh *InstanceEventDeferHandler) {
	fmt.Println(time.Now())
	c := time.After(3500 * time.Millisecond)
	var evt3 *KvEvent
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
