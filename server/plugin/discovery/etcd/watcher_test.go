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
	"testing"
	"time"

	"github.com/apache/servicecomb-service-center/server/plugin/registry"
)

type mockWatcher struct {
	lw *mockListWatch
}

func (w *mockWatcher) EventBus() <-chan *registry.PluginResponse {
	return w.lw.Bus
}
func (w *mockWatcher) Stop() {
}

type mockListWatch struct {
	ListResponse *registry.PluginResponse
	Bus          chan *registry.PluginResponse
	Watcher      Watcher
	Rev          int64
}

func (lw *mockListWatch) List(op ListWatchConfig) (*registry.PluginResponse, error) {
	if lw.ListResponse == nil {
		return nil, fmt.Errorf("error")
	}
	lw.Rev = lw.ListResponse.Revision
	return lw.ListResponse, nil
}
func (lw *mockListWatch) DoWatch(ctx context.Context, f func(*registry.PluginResponse)) error {
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

	test := &registry.PluginResponse{
		Action: registry.Put,
	}
	w = newInnerWatcher(&mockListWatch{ListResponse: test}, ListWatchConfig{Timeout: time.Second, Context: context.Background()})
	resp = <-w.EventBus()
	if resp == nil || resp.Action != registry.Put {
		t.Fatalf("TestInnerWatcher_EventBus failed")
	}
	w.Stop()
}
