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

package task

import (
	"context"
	"testing"

	"github.com/apache/servicecomb-service-center/syncer/service/event"

	"github.com/go-chassis/cari/sync"
	"github.com/stretchr/testify/assert"
)

func TestNewManager(t *testing.T) {
	receiver := make(chan struct{}, 1)
	fs := &mockSender{
		events:  make(map[string]*event.Event),
		receive: receiver,
	}
	ctx := context.TODO()
	m := NewManager(
		ManagerOperator(&mockOperator{
			tasks: map[string]*sync.Task{
				"xxx1": {
					ID:           "xxx1",
					ResourceType: "demo",
					Action:       "create",
					Timestamp:    0,
					Status:       "pending",
				},
			},
		}),
		ManagerInternal(defaultInternal),
		EventSender(fs))

	m.LoadAndHandleTask(ctx)
	m.UpdateResultTask(ctx)
	<-receiver
	assert.Equal(t, 1, len(fs.events))
}

type mockOperator struct {
	tasks map[string]*sync.Task
}

func (f *mockOperator) ListTasks(_ context.Context) ([]*sync.Task, error) {
	result := make([]*sync.Task, 0, len(f.tasks))
	for _, task := range f.tasks {
		result = append(result, task)
	}
	return result, nil
}

func (f *mockOperator) DeleteTask(_ context.Context, t *sync.Task) error {
	delete(f.tasks, t.ID)
	return nil
}

type mockSender struct {
	events  map[string]*event.Event
	receive chan struct{}
}

func (f *mockSender) Send(et *event.Event) {
	f.events[et.Id] = et
	f.receive <- struct{}{}
}
