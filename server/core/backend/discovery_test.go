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
	"context"
	"errors"
	"github.com/apache/servicecomb-service-center/pkg/task"
	"github.com/apache/servicecomb-service-center/server/plugin/pkg/discovery"
	"github.com/apache/servicecomb-service-center/server/plugin/pkg/registry"
	"testing"
)

var closedCh = make(chan struct{})

func init() {
	close(closedCh)
}

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

	tt := NewLeaseAsyncTask(registry.OpGet())
	tt.TTL = 1
	s.taskService = &mockAsyncTaskService{Task: tt}

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

	s.Stop()
}

type extend struct {
	evts []discovery.KvEvent
	cfg  *discovery.Config
}

func (e *extend) Name() string {
	return "test"
}

func (e *extend) Config() *discovery.Config {
	return e.cfg
}

func TestInstallType(t *testing.T) {
	s := &KvStore{}
	s.Initialize()

	// case: normal
	e := &extend{cfg: discovery.Configure()}
	e.cfg.WithPrefix("/test").WithEventFunc(func(evt discovery.KvEvent) {
		e.evts = append(e.evts, evt)
	})
	id, err := s.Install(e)
	if err != nil {
		t.Fatal(err)
	}
	if id == discovery.TypeError {
		t.Fatal(err)
	}
	if id.String() != "test" {
		t.Fatalf("TestInstallType failed")
	}

	// case: inject config
	itf, _ := s.AddOns[id]
	cfg := itf.(AddOn).Config()
	if cfg == nil || cfg.OnEvent == nil {
		t.Fatal("installType fail", err)
	}

	cfg.OnEvent(discovery.KvEvent{Revision: 1})
	if s.rev != 1 || len(e.evts) != 1 {
		t.Fatalf("TestInstallType failed")
	}

	// case: install again
	cfg = discovery.Configure().WithPrefix("/test")
	id, err = s.Install(NewAddOn("test", cfg))
	if id != discovery.TypeError || err == nil {
		t.Fatal("installType fail", err)
	}
}

func TestNewAddOn(t *testing.T) {
	s := &KvStore{}
	s.Initialize()

	id, err := s.Install(NewAddOn("TestNewAddOn", nil))
	if id != discovery.TypeError || err == nil {
		t.Fatalf("TestNewAddOn failed")
	}
	id, err = s.Install(NewAddOn("", discovery.Configure()))
	if id != discovery.TypeError || err == nil {
		t.Fatalf("TestNewAddOn failed")
	}
	id, err = s.Install(nil)
	if id != discovery.TypeError || err == nil {
		t.Fatalf("TestNewAddOn failed")
	}
	id, err = s.Install(NewAddOn("TestNewAddOn", discovery.Configure()))
	if id == discovery.TypeError || err != nil {
		t.Fatalf("TestNewAddOn failed")
	}
	_, err = s.Install(NewAddOn("TestNewAddOn", discovery.Configure()))
	if err == nil {
		t.Fatalf("TestNewAddOn failed")
	}
}
