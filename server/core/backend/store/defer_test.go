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

	cache := NewKvCache(nil, 1)
	cache.store["/1"] = kv1
	cache.store["/2"] = kv2
	cache.store["/3"] = kv3

	evts1 := []*Event{
		{
			Type:   pb.EVT_CREATE,
			Object: kv1,
		},
		{
			Type:   pb.EVT_UPDATE,
			Object: kv1,
		},
	}
	evts2 := []*Event{
		{
			Type:   pb.EVT_DELETE,
			Object: kv2,
		},
		{
			Type:   pb.EVT_DELETE,
			Object: kv3,
		},
	}
	evts3 := []*Event{
		{
			Type:   pb.EVT_CREATE,
			Object: kv2,
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
	var evt3 *Event
	for {
		select {
		case evt := <-iedh.HandleChan():
			fmt.Println(time.Now(), evt)
			if string(evt.Object.(*mvccpb.KeyValue).Key) == "/3" {
				evt3 = evt
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
