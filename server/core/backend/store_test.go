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
	"errors"
	"github.com/apache/incubator-servicecomb-service-center/pkg/task"
	"github.com/apache/incubator-servicecomb-service-center/server/core"
	"github.com/apache/incubator-servicecomb-service-center/server/infra/registry"
	"golang.org/x/net/context"
	"testing"
)

type mockAsyncTaskService struct {
	Task task.Task
}

func (a *mockAsyncTaskService) Run()                   {}
func (a *mockAsyncTaskService) Stop()                  {}
func (a *mockAsyncTaskService) Ready() <-chan struct{} { return closedCh }
func (a *mockAsyncTaskService) Add(ctx context.Context, task task.Task) error {
	switch task.Key() {
	case "LeaseAsyncTask_error":
		return errors.New("error")
	}
	return nil
}
func (a *mockAsyncTaskService) LatestHandled(key string) (task.Task, error) {
	switch a.Task.Key() {
	case key:
		return a.Task, nil
	}
	return nil, errors.New("error")
}

func TestStore(t *testing.T) {
	s := &KvStore{}
	s.Initialize()
	e := s.Entities(999)
	if e != DefaultKvEntity() {
		t.Fatalf("TestStore failed")
	}

	core.ServerInfo.Config.EnableCache = false

	s.store(context.Background())

	tt := NewLeaseAsyncTask(registry.OpGet())
	tt.TTL = 1
	s.taskService = &mockAsyncTaskService{Task: tt}

	<-s.Ready()

	// KeepAlive case: add task error
	ttl, err := s.KeepAlive(context.Background(), registry.WithKey([]byte("error")))
	if err == nil || ttl > 0 {
		t.Fatalf("TestStore failed")
	}

	// KeepAlive case: get last task error
	tt.key = "LeaseAsyncTask_a"
	ttl, err = s.KeepAlive(context.Background(), registry.WithKey([]byte("b")))
	if err == nil || ttl > 0 {
		t.Fatalf("TestStore failed")
	}

	// KeepAlive case: get last task error
	tt.key = "LeaseAsyncTask_a"
	ttl, err = s.KeepAlive(context.Background(), registry.WithKey([]byte("a")))
	if err != nil || ttl != 1 {
		t.Fatalf("TestStore failed")
	}

	core.ServerInfo.Config.EnableCache = true

	s.Stop()
}

type extend struct {
	evts []KvEvent
}

func (e *extend) Name() string {
	return "test"
}

func (e *extend) Config() *Config {
	return Configure().WithPrefix("/test").WithEventFunc(func(evt KvEvent) {
		e.evts = append(e.evts, evt)
	})
}

func TestInstallType(t *testing.T) {
	s := &KvStore{}
	s.Initialize()

	// case: normal
	e := &extend{}
	id, err := s.Install(e)
	if err != nil {
		t.Fatal(err)
	}
	if id == TypeError {
		t.Fatal(err)
	}
	if id.String() != "test" {
		t.Fatalf("TestInstallType failed")
	}
	if TypeConfig[id] == nil || TypeConfig[id].OnEvent == nil || EventProxy(id) == nil {
		t.Fatal("installType fail", err)
	}

	// case: test event
	h := &mockEventHandler{StoreType: id}
	AddEventHandler(h)
	TypeConfig[id].OnEvent(KvEvent{Revision: 1})
	if len(e.evts) != 1 || e.evts[0].Revision != 1 || h.Evt.Revision != 1 || s.rev != 1 {
		t.Fatalf("TestInstallType failed")
	}

	// case: install again
	cfg := Configure().WithPrefix("/test")
	id, err = s.Install(NewExtension("test", cfg))
	if id != TypeError || err == nil {
		t.Fatal("installType fail", err)
	}
}
