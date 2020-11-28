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

package client

import (
	"context"
	"errors"
	"testing"

	"github.com/apache/servicecomb-service-center/pkg/task"
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

func TestKeepAlive(t *testing.T) {
	tt := NewLeaseAsyncTask(OpGet())
	tt.TTL = 1
	task.RegisterService(&mockAsyncTaskService{Task: tt})

	// KeepAlive case: add task error
	ttl, err := KeepAlive(context.Background(), WithKey([]byte("error")))
	if err == nil || ttl > 0 {
		t.Fatalf("TestStore failed")
	}

	// KeepAlive case: get last task error
	tt.key = "LeaseAsyncTask_a"
	ttl, err = KeepAlive(context.Background(), WithKey([]byte("b")))
	if err == nil || ttl > 0 {
		t.Fatalf("TestStore failed")
	}

	// KeepAlive case: get last task error
	tt.key = "LeaseAsyncTask_a"
	ttl, err = KeepAlive(context.Background(), WithKey([]byte("a")))
	if err != nil || ttl != 1 {
		t.Fatalf("TestStore failed")
	}
}
