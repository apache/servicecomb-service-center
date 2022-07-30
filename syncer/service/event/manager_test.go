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

package event_test

import (
	"context"
	"testing"
	"time"

	_ "github.com/apache/servicecomb-service-center/test"

	v1sync "github.com/apache/servicecomb-service-center/syncer/api/v1"
	"github.com/apache/servicecomb-service-center/syncer/service/event"
	"github.com/apache/servicecomb-service-center/syncer/service/replicator/resource"
	"github.com/stretchr/testify/assert"
)

func TestWork(t *testing.T) {
	r := &mockReplicator{
		results: &v1sync.Results{
			Results: map[string]*v1sync.Result{
				"xxx1": {
					Code:    resource.Success,
					Message: "ok",
				},
				"xxx2": {
					Code:    resource.Success,
					Message: "ok",
				},
			},
		},
		err: nil,
	}
	event.Work()
	em := event.GetManager().(*event.ManagerImpl)
	em.Replicator = r

	var f *forkResources
	resource.RegisterResources("fork", func(event *v1sync.Event) resource.Resource {
		f = &forkResources{
			loadCurrentResourceResult: nil,
			needOperateResult:         nil,
			operateResult:             resource.SuccessResult(),
			result:                    make(chan struct{}),
		}
		return f
	})
	event.Send(&event.Event{
		Event: &v1sync.Event{
			Id:        "xxx1",
			Action:    "create",
			Subject:   "fork",
			Opts:      nil,
			Timestamp: v1sync.Timestamp(),
		},
	})

	ctx := context.TODO()
	event.Publish(ctx, "create", "fork", map[string]string{
		"hello": "world",
	})
	result := make(chan *event.Result, 1)
	event.Send(&event.Event{
		Event: &v1sync.Event{
			Id:        "xxx2",
			Action:    "create",
			Subject:   "fork",
			Opts:      nil,
			Timestamp: v1sync.Timestamp(),
		},
		Result: result,
	})

	data := <-result
	if assert.NotNil(t, data.Data) {
		assert.Equal(t, resource.Success, data.Data.Code)
		assert.Equal(t, "ok", data.Data.Message)
	}
	time.Sleep(time.Second)
	if assert.NotNil(t, f) {
		<-f.result
	}
}

type mockReplicator struct {
	results *v1sync.Results
	err     error
}

func (f *mockReplicator) Replicate(_ context.Context, _ *v1sync.EventList) (*v1sync.Results, error) {
	return f.results, f.err
}

func (f *mockReplicator) Persist(_ context.Context, _ *v1sync.EventList) []*resource.Result {
	return nil
}

type forkResources struct {
	loadCurrentResourceResult *resource.Result
	needOperateResult         *resource.Result
	operateResult             *resource.Result
	result                    chan struct{}
}

func (f *forkResources) LoadCurrentResource(_ context.Context) *resource.Result {
	return f.loadCurrentResourceResult
}

func (f *forkResources) NeedOperate(_ context.Context) *resource.Result {
	return f.needOperateResult
}

func (f *forkResources) Operate(_ context.Context) *resource.Result {
	return f.operateResult
}

func (f forkResources) FailHandle(_ context.Context, _ int32) (*v1sync.Event, error) {
	f.result <- struct{}{}
	return nil, nil
}

func (f forkResources) CanDrop() bool {
	return true
}

func TestNewManager(t *testing.T) {
	nm := event.NewManager(event.ManagerInternal(event.DefaultInternal), event.Replicator(new(mockReplicator)))
	assert.NotNil(t, nm)
}
