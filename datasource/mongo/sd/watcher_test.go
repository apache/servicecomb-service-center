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
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

type mockWatcher struct {
	lw *mockListWatch
}

func (w *mockWatcher) WatchResponse() <-chan *MongoListWatchResponse {
	return w.lw.Bus
}
func (w *mockWatcher) Stop() {
}

type mockListWatch struct {
	ListResponse *MongoListWatchResponse
	Bus          chan *MongoListWatchResponse
	Watcher      Watcher
	resumeToken  bson.Raw
}

func (lw *mockListWatch) List(ListWatchConfig) (*MongoListWatchResponse, error) {
	if lw.ListResponse == nil {
		return nil, fmt.Errorf("error")
	}
	return lw.ListResponse, nil
}

func (lw *mockListWatch) DoWatch(ctx context.Context, f func(*MongoListWatchResponse)) error {
	if lw.ListResponse == nil {
		return fmt.Errorf("error")
	}

	f(lw.ListResponse)
	<-ctx.Done()
	return nil
}
func (lw *mockListWatch) Watch(ListWatchConfig) Watcher {
	return lw.Watcher
}

func (lw *mockListWatch) ResumeToken() bson.Raw {
	return lw.resumeToken
}

func TestInnerWatcher_EventBus(t *testing.T) {
	w := newInnerWatcher(&mockListWatch{}, ListWatchConfig{Timeout: time.Second, Context: context.Background()})
	resp := <-w.WatchResponse()
	if resp != nil {
		t.Fatal("TestInnerWatcher_EventBus failed")
	}
	w.Stop()

	test := &MongoListWatchResponse{
		OperationType: updateOp,
	}
	w = newInnerWatcher(&mockListWatch{ListResponse: test}, ListWatchConfig{Timeout: time.Second, Context: context.Background()})
	resp = <-w.WatchResponse()
	if resp == nil || resp.OperationType != updateOp {
		t.Fatal("TestInnerWatcher_EventBus failed")
	}
	w.Stop()
}
