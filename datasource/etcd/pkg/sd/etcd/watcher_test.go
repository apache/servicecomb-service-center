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
	"fmt"
	"github.com/apache/servicecomb-service-center/datasource/etcd/client"
	"testing"
	"time"
)

type mockWatcher struct {
	lw *mockListWatch
}

func (w *mockWatcher) EventBus() <-chan *client.PluginResponse {
	return w.lw.Bus
}
func (w *mockWatcher) Stop() {
}

type mockListWatch struct {
	ListResponse *client.PluginResponse
	Bus          chan *client.PluginResponse
	Watcher      Watcher
	Rev          int64
}

func (lw *mockListWatch) List(op ListWatchConfig) (*client.PluginResponse, error) {
	if lw.ListResponse == nil {
		return nil, fmt.Errorf("error")
	}
	lw.Rev = lw.ListResponse.Revision
	return lw.ListResponse, nil
}
func (lw *mockListWatch) DoWatch(ctx context.Context, f func(*client.PluginResponse)) error {
	if lw.ListResponse == nil {
		return fmt.Errorf("error")
	}
	if len(lw.ListResponse.Kvs) > 0 {
		lw.Rev = lw.ListResponse.Kvs[0].ModRevision
	}
	f(lw.ListResponse)
	<-ctx.Done()
	return nil
}
func (lw *mockListWatch) Watch(op ListWatchConfig) Watcher {
	return lw.Watcher
}
func (lw *mockListWatch) Revision() int64 {
	return lw.Rev
}

func TestInnerWatcher_EventBus(t *testing.T) {
	w := newInnerWatcher(&mockListWatch{}, ListWatchConfig{Timeout: time.Second, Context: context.Background()})
	resp := <-w.EventBus()
	if resp != nil {
		t.Fatalf("TestInnerWatcher_EventBus failed")
	}
	w.Stop()

	test := &client.PluginResponse{
		Action: client.ActionPut,
	}
	w = newInnerWatcher(&mockListWatch{ListResponse: test}, ListWatchConfig{Timeout: time.Second, Context: context.Background()})
	resp = <-w.EventBus()
	if resp == nil || resp.Action != client.ActionPut {
		t.Fatalf("TestInnerWatcher_EventBus failed")
	}
	w.Stop()
}
