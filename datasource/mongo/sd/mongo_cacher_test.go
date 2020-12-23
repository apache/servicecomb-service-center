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
	"testing"

	"github.com/apache/servicecomb-service-center/pkg/gopool"
	pb "github.com/go-chassis/cari/discovery"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
)

func TestNewMongoCacher(t *testing.T) {
	mockMongoCache := NewMongoCache("test", DefaultOptions())
	w := &mockWatcher{}
	lw := &mockListWatch{
		Bus: make(chan *MongoListWatchResponse, 100),
	}
	w.lw = lw

	cr := &MongoCacher{
		Options:   DefaultOptions(),
		ready:     make(chan struct{}),
		lw:        lw,
		goroutine: gopool.New(context.Background()),
		cache:     mockMongoCache,
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
		Options: DefaultOptions().SetTable(instance).
			SetEventFunc(func(e MongoEvent) {
				evt = e
			}),
		ready:     make(chan struct{}),
		lw:        lw,
		goroutine: gopool.New(context.Background()),
		cache:     mockMongoCache,
	}
	lw.Watcher = w

	mockDocumentID := "5fcf2f1a4ea1e6d2f4c61d47"
	mockBusinessID := "95cd8dbf3e8411eb92d8fa163e8a81c9"
	mockResumeToken, _ :=
		bson.Marshal(bson.M{"_data": "825FDB4272000000012B022C0100296E5A10043E2D15AC82D9484C8090E68AF36FED2A46645F696400645FD76265066A6D2DF2AAC8D80004"})

	var mongoInfos []MongoInfo

	data := MongoInfo{DocumentID: mockDocumentID, BusinessID: mockBusinessID,
		Value: Instance{Domain: "default", Project: "default",
			InstanceInfo: &pb.MicroServiceInstance{InstanceId: mockBusinessID, ModTimestamp: "100000"}}}

	mongoInfos = append(mongoInfos, data)

	test := &MongoListWatchResponse{
		OperationType: insertOp,
		Infos:         mongoInfos,
	}

	// case: list 1 resp and watch 0 event
	t.Run("case list: first time list init cache", func(t *testing.T) {
		// case: resume token is nil, first time list event is init
		cr.cache.Remove(mockBusinessID)
		cr.cache.RemoveDocumentID(mockDocumentID)
		lw.resumeToken = nil
		lw.ListResponse = test
		lw.Bus <- nil

		cr.refresh(ctx)

		// check ready
		assert.Equal(t, true, cr.IsReady())

		//check config
		assert.Equal(t, instance, cr.Options.Key)

		// check event
		if evt.Type != pb.EVT_INIT || evt.DocumentID != mockDocumentID || evt.BusinessID != mockBusinessID ||
			evt.Value.(Instance).Domain != "default" || evt.Value.(Instance).Project != "default" {
			t.Fatalf("TestNewMongoCacher failed, %v", evt)
		}

		// check cache
		cache := cr.cache.Get(mockBusinessID)
		if cache == nil || cache.(Instance).Domain != "default" || cache.(Instance).Project != "default" ||
			cache.(Instance).InstanceInfo.ModTimestamp != "100000" {
			t.Fatalf("TestNewMongoCacher failed")
		}
	})

	t.Run("case list: re-list and updateOp cache", func(t *testing.T) {
		// case: re-list and should be no event
		lw.Bus <- nil
		evt.Value = nil
		cr.refresh(ctx)

		//check events
		assert.Equal(t, nil, evt.Value)

		// check cache
		cache := cr.cache.Get(mockBusinessID)
		if cache == nil || cache.(Instance).Domain != "default" || cache.(Instance).Project != "default" ||
			cache.(Instance).InstanceInfo.ModTimestamp != "100000" {
			t.Fatalf("TestNewMongoCacher failed")
		}

		// prepare updateOp data
		dataUpdate := MongoInfo{DocumentID: mockDocumentID, BusinessID: mockBusinessID,
			Value: Instance{Domain: "default", Project: "default",
				InstanceInfo: &pb.MicroServiceInstance{InstanceId: mockBusinessID, HostName: "test", ModTimestamp: "100001"}}}

		var mongoUpdateInfos []MongoInfo
		mongoUpdateInfos = append(mongoUpdateInfos, dataUpdate)
		testUpdate := &MongoListWatchResponse{
			OperationType: insertOp,
			Infos:         mongoUpdateInfos,
		}

		lw.ListResponse = testUpdate
		lw.resumeToken = mockResumeToken

		// case: re-list and over no event times, and then event should be updateOp
		for i := 0; i < DefaultForceListInterval; i++ {
			lw.Bus <- nil
		}

		evt.Value = nil
		for i := 0; i < DefaultForceListInterval; i++ {
			cr.refresh(ctx)
		}
		// check event
		if evt.Type != pb.EVT_UPDATE || evt.DocumentID != mockDocumentID || evt.BusinessID != mockBusinessID ||
			evt.Value.(Instance).Domain != "default" || evt.Value.(Instance).Project != "default" ||
			evt.Value.(Instance).InstanceInfo.ModTimestamp != "100001" {
			t.Fatalf("TestNewMongoCacher failed, %v", evt)
		}
		// check cache
		cache = cr.cache.Get(mockBusinessID)
		if cache == nil || cache.(Instance).Domain != "default" || cache.(Instance).Project != "default" ||
			cache.(Instance).InstanceInfo.ModTimestamp != "100001" {
			t.Fatalf("TestNewMongoCacher failed")
		}
	})

	t.Run("case list: no infos list and delete cache", func(t *testing.T) {
		// case: no infos list, event should be delete
		lw.ListResponse = &MongoListWatchResponse{}
		for i := 0; i < DefaultForceListInterval; i++ {
			lw.Bus <- nil
		}
		evt.Value = nil
		for i := 0; i < DefaultForceListInterval; i++ {
			cr.refresh(ctx)
		}

		// check event
		assert.Equal(t, pb.EVT_DELETE, evt.Type)

		// check cache
		cache := cr.cache.Get(mockBusinessID)
		assert.Equal(t, nil, cache)
	})

	t.Run("case list: mark cache dirty and reset cache", func(t *testing.T) {
		lw.ListResponse = test
		lw.Bus <- nil
		lw.resumeToken = nil
		evt.Value = nil

		cr.cache.MarkDirty()
		cr.refresh(ctx)

		// check event
		if evt.Value != nil {
			t.Fatalf("TestNewMongoCacher failed, %v", evt)
		}

		// check cache
		cache := cr.cache.Get(mockBusinessID)
		if cache == nil || cache.(Instance).Domain != "default" || cache.(Instance).Project != "default" {
			t.Fatalf("TestNewMongoCacher failed:%s", test.OperationType)
		}
	})

	t.Run("case watch: caught create event", func(t *testing.T) {
		cr.cache.Remove(mockBusinessID)
		cr.cache.RemoveDocumentID(mockDocumentID)
		lw.Bus <- test
		lw.Bus <- nil
		lw.resumeToken = mockResumeToken

		cr.refresh(ctx)

		// check event
		if evt.Type != pb.EVT_CREATE || evt.DocumentID != mockDocumentID || evt.BusinessID != mockBusinessID {
			t.Fatalf("TestNewMongoCacher failed, %v", evt)
		}
		// check cache
		cache := cr.cache.Get(mockBusinessID)
		if cache == nil || cache.(Instance).Domain != "default" || cache.(Instance).Project != "default" {
			t.Fatalf("TestNewMongoCacher failed")
		}
	})

	t.Run("case watch: caught updateOp event", func(t *testing.T) {
		// prepare updateOp data
		dataUpdate := MongoInfo{DocumentID: mockDocumentID, BusinessID: mockBusinessID,
			Value: Instance{Domain: "default", Project: "default",
				InstanceInfo: &pb.MicroServiceInstance{InstanceId: mockBusinessID, HostName: "test", ModTimestamp: "100001"}}}

		var mongoUpdateInfos []MongoInfo
		mongoUpdateInfos = append(mongoUpdateInfos, dataUpdate)
		testUpdate := &MongoListWatchResponse{
			OperationType: updateOp,
			Infos:         mongoUpdateInfos,
		}
		lw.Bus <- testUpdate
		lw.Bus <- nil

		cr.refresh(ctx)
		// check event
		if evt.Type != pb.EVT_UPDATE || evt.DocumentID != mockDocumentID || evt.BusinessID != mockBusinessID ||
			evt.Value.(Instance).Domain != "default" || evt.Value.(Instance).Project != "default" || evt.Value.(Instance).InstanceInfo.ModTimestamp != "100001" {
			t.Fatalf("TestNewMongoCacher failed, %v", evt)
		}
		// check cache
		cache := cr.cache.Get(mockBusinessID)
		if cache == nil || cache.(Instance).Domain != "default" || cache.(Instance).Project != "default" || cache.(Instance).InstanceInfo.ModTimestamp != "100001" {
			t.Fatalf("TestNewMongoCacher failed")
		}
	})

	t.Run("case watch: caught delete event but value is nil", func(t *testing.T) {
		test.OperationType = deleteOp
		lw.Bus <- test
		lw.Bus <- nil

		cr.refresh(ctx)
		// check event
		if evt.Type != pb.EVT_DELETE || evt.DocumentID != mockDocumentID || evt.BusinessID != mockBusinessID ||
			evt.Value.(Instance).Domain != "default" || evt.Value.(Instance).Project != "default" || evt.Value.(Instance).InstanceInfo.ModTimestamp != "100001" {
			t.Fatalf("TestNewMongoCacher failed, %v", evt)
		}
		// check cache
		cache := cr.cache.Get(mockBusinessID)
		if cache != nil {
			t.Fatalf("TestNewMongoCacher failed")
		}

	})
}

func TestMongoCacher_Run(t *testing.T) {

	w := &mockWatcher{}
	lw := &mockListWatch{
		Bus: make(chan *MongoListWatchResponse, 100),
	}

	mockMongoCache := NewMongoCache("test", DefaultOptions())
	cr := &MongoCacher{
		Options:   DefaultOptions().SetTable(instance),
		ready:     make(chan struct{}),
		lw:        lw,
		goroutine: gopool.New(context.Background()),
		cache:     mockMongoCache,
	}
	lw.Watcher = w

	cr.Run()

	// check cache
	cache := cr.cache
	assert.NotNil(t, cache)

	cr.Stop()
}
