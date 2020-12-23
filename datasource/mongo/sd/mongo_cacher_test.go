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

package sd

import (
	"context"
	"fmt"
	"github.com/apache/servicecomb-service-center/datasource/sdcommon"
	"testing"

	"github.com/apache/servicecomb-service-center/pkg/gopool"
	pb "github.com/go-chassis/cari/discovery"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
)

type MockListWatch struct {
	ListResponse  *sdcommon.ListWatchResp
	WatchResponse *sdcommon.ListWatchResp
	resumeToken   bson.Raw
}

func (lw *MockListWatch) List(sdcommon.ListWatchConfig) (*sdcommon.ListWatchResp, error) {
	if lw.ListResponse == nil {
		return nil, fmt.Errorf("list error")
	}
	return lw.ListResponse, nil
}

func (lw *MockListWatch) DoWatch(ctx context.Context, f func(*sdcommon.ListWatchResp)) error {
	if lw.WatchResponse == nil {
		return fmt.Errorf("error")
	}

	f(lw.WatchResponse)
	<-ctx.Done()
	return nil
}

func (lw *MockListWatch) EventBus(op sdcommon.ListWatchConfig) *sdcommon.EventBus {
	return sdcommon.NewEventBus(lw, op)
}

func TestNewMongoCacher(t *testing.T) {
	mockMongoCache := NewMongoCache("test", DefaultOptions())
	lw := &MockListWatch{}

	cr := &MongoCacher{
		Options:     DefaultOptions(),
		ready:       make(chan struct{}),
		isFirstTime: true,
		lw:          lw,
		goroutine:   gopool.New(context.Background()),
		cache:       mockMongoCache,
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// case: cause list internal error before initialized
	t.Run("case list: internal error before initialized", func(t *testing.T) {
		cr.refresh(ctx)
		if cr.IsReady() {
			t.Fatalf("TestNewKvCacher failed")
		}

	})

	// prepare mock data
	var evt MongoEvent
	cr = &MongoCacher{
		Options:   DefaultOptions().SetTable(instance),
		ready:     make(chan struct{}),
		lw:        lw,
		goroutine: gopool.New(context.Background()),
		cache:     mockMongoCache,
	}

	EventProxy(instance).AddHandleFunc(func(e MongoEvent) {
		evt = e
	})

	mockDocumentID := "5fcf2f1a4ea1e6d2f4c61d47"
	mockResourceID := "95cd8dbf3e8411eb92d8fa163e8a81c9"
	mockResumeToken, _ :=
		bson.Marshal(bson.M{"_data": "825FDB4272000000012B022C0100296E5A10043E2D15AC82D9484C8090E68AF36FED2A46645F696400645FD76265066A6D2DF2AAC8D80004"})

	var resources []*sdcommon.Resource
	resource := &sdcommon.Resource{Key: mockResourceID, DocumentID: mockDocumentID, Value: Instance{Domain: "default", Project: "default",
		InstanceInfo: &pb.MicroServiceInstance{InstanceId: mockResourceID, ModTimestamp: "100000"}}}
	resources = append(resources, resource)
	test := &sdcommon.ListWatchResp{
		Action:    sdcommon.ActionCreate,
		Resources: resources,
	}

	evtExpect := MongoEvent{
		DocumentID: mockDocumentID,
		ResourceID: mockResourceID,
		Value:      resource.Value,
		Type:       pb.EVT_INIT,
	}

	// case: list 1 resp and watch 0 event
	t.Run("case list: first time list init cache", func(t *testing.T) {
		// case: resume token is nil, first time list event is init
		cr.isFirstTime = true
		cr.cache.Remove(mockResourceID)
		cr.cache.RemoveDocumentID(mockDocumentID)

		lw.ListResponse = test
		lw.resumeToken = nil
		lw.WatchResponse = nil

		cr.refresh(ctx)

		// check ready
		assert.Equal(t, true, cr.IsReady())

		//check config
		assert.Equal(t, instance, cr.Options.Key)

		// check event
		assert.Equal(t, evtExpect, evt)

		// check cache
		cache := cr.cache.Get(mockResourceID)
		assert.Equal(t, resource.Value, cache)
	})

	t.Run("case list: re-list and updateOp cache", func(t *testing.T) {
		// case: re-list and should be no event
		lw.WatchResponse = nil
		evt.Value = nil
		cr.refresh(ctx)

		//check events
		assert.Nil(t, evt.Value)

		// check cache
		cache := cr.cache.Get(mockResourceID)
		assert.Equal(t, resource.Value, cache)

		// prepare updateOp data
		dataUpdate := &sdcommon.Resource{Key: mockResourceID, DocumentID: mockDocumentID,
			Value: Instance{Domain: "default", Project: "default",
				InstanceInfo: &pb.MicroServiceInstance{InstanceId: mockResourceID, HostName: "test", ModTimestamp: "100001"}}}

		var mongoUpdateResources []*sdcommon.Resource
		mongoUpdateResources = append(mongoUpdateResources, dataUpdate)
		testUpdate := &sdcommon.ListWatchResp{
			Action:    sdcommon.ActionUpdate,
			Resources: mongoUpdateResources,
		}

		lw.ListResponse = testUpdate
		lw.resumeToken = mockResumeToken

		// case: re-list and over no event times, and then event should be updateOp
		for i := 0; i < sdcommon.DefaultForceListInterval; i++ {
			lw.WatchResponse = nil
		}

		evt.Value = nil
		for i := 0; i < sdcommon.DefaultForceListInterval; i++ {
			cr.refresh(ctx)
		}
		// check event
		evtExpect.Type = pb.EVT_UPDATE
		evtExpect.Value = dataUpdate.Value
		assert.Equal(t, evtExpect, evt)

		// check cache
		cache = cr.cache.Get(mockResourceID)
		assert.Equal(t, dataUpdate.Value, cache)
	})

	t.Run("case list: no infos list and delete cache", func(t *testing.T) {
		// case: no infos list, event should be delete
		lw.ListResponse = &sdcommon.ListWatchResp{}
		for i := 0; i < sdcommon.DefaultForceListInterval; i++ {
			lw.WatchResponse = nil
		}
		evt.Value = nil
		for i := 0; i < sdcommon.DefaultForceListInterval; i++ {
			cr.refresh(ctx)
		}

		// check event
		evtExpect.Type = pb.EVT_DELETE
		assert.Equal(t, evtExpect, evt)

		// check cache
		cache := cr.cache.Get(mockResourceID)
		assert.Nil(t, nil, cache)
	})

	t.Run("case list: mark cache dirty and reset cache", func(t *testing.T) {
		lw.ListResponse = test
		cr.isFirstTime = true
		evt.Value = nil

		cr.cache.MarkDirty()
		cr.refresh(ctx)

		// check event
		if evt.Value != nil {
			t.Fatalf("TestNewMongoCacher failed, %v", evt)
		}

		// check cache
		cache := cr.cache.Get(mockResourceID)
		assert.Equal(t, resource.Value, cache)
	})

	t.Run("case watch: caught create event", func(t *testing.T) {
		cr.cache.Remove(mockResourceID)
		cr.cache.RemoveDocumentID(mockDocumentID)
		lw.WatchResponse = test
		lw.ListResponse = nil
		lw.resumeToken = mockResumeToken

		cr.refresh(ctx)

		// check event
		evtExpect.Type = pb.EVT_CREATE
		evtExpect.Value = resource.Value
		assert.Equal(t, evtExpect, evt)

		// check cache
		cache := cr.cache.Get(mockResourceID)
		assert.Equal(t, resource.Value, cache)

	})

	t.Run("case watch: caught updateOp event", func(t *testing.T) {
		// prepare updateOp data
		dataUpdate := &sdcommon.Resource{Key: mockResourceID, DocumentID: mockDocumentID,
			Value: Instance{Domain: "default", Project: "default",
				InstanceInfo: &pb.MicroServiceInstance{InstanceId: mockResourceID, HostName: "test", ModTimestamp: "100001"}}}

		var mongoUpdateResources []*sdcommon.Resource
		mongoUpdateResources = append(mongoUpdateResources, dataUpdate)
		testUpdate := &sdcommon.ListWatchResp{
			Action:    sdcommon.ActionUpdate,
			Resources: mongoUpdateResources,
		}
		lw.WatchResponse = testUpdate
		lw.ListResponse = nil

		cr.refresh(ctx)

		// check event
		evtExpect.Type = pb.EVT_UPDATE
		evtExpect.Value = dataUpdate.Value
		assert.Equal(t, evtExpect, evt)

		// check cache
		cache := cr.cache.Get(mockResourceID)
		assert.Equal(t, dataUpdate.Value, cache)
	})

	t.Run("case watch: caught delete event but value is nil", func(t *testing.T) {
		test.Action = sdcommon.ActionDelete
		lw.WatchResponse = test
		lw.ListResponse = nil

		cr.refresh(ctx)

		// check event
		evtExpect.Type = pb.EVT_DELETE
		assert.Equal(t, evtExpect, evt)

		// check cache
		cache := cr.cache.Get(mockResourceID)
		assert.Nil(t, cache)

	})
}

func TestMongoCacher_Run(t *testing.T) {
	lw := &MockListWatch{}

	mockMongoCache := NewMongoCache("test", DefaultOptions())
	cr := &MongoCacher{
		Options:   DefaultOptions().SetTable(instance),
		ready:     make(chan struct{}),
		lw:        lw,
		goroutine: gopool.New(context.Background()),
		cache:     mockMongoCache,
	}

	cr.Run()

	// check cache
	cache := cr.cache
	assert.NotNil(t, cache)

	cr.Stop()
}
