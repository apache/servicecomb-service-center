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

package task_test

import (
	"context"
	"testing"

	"github.com/go-chassis/cari/discovery"
	"github.com/go-chassis/cari/sync"
	"github.com/stretchr/testify/assert"

	"github.com/apache/servicecomb-service-center/eventbase/datasource"
	"github.com/apache/servicecomb-service-center/eventbase/model"
	"github.com/apache/servicecomb-service-center/eventbase/service/task"
	"github.com/apache/servicecomb-service-center/eventbase/test"
)

func init() {
	err := datasource.Init(&datasource.Config{
		Kind:   test.DBKind,
		Logger: nil,
	})
	if err != nil {
		panic(err)
	}
}

func TestTaskService(t *testing.T) {
	taskOne, _ := sync.NewTask("task", "task", sync.CreateAction, "service",
		discovery.MicroService{
			ServiceId:   "123",
			AppId:       "appId1",
			ServiceName: "svc1",
			Version:     "1.0",
		})
	taskTwo, _ := sync.NewTask("task", "task", sync.UpdateAction, "service",
		discovery.MicroService{
			ServiceId:   "456",
			AppId:       "appId2",
			ServiceName: "svc2",
			Version:     "1.0",
		})
	taskThree, _ := sync.NewTask("task", "task", sync.DeleteAction, "service",
		discovery.MicroService{
			ServiceId:   "789",
			AppId:       "appId3",
			ServiceName: "svc3",
			Version:     "1.0",
		})
	t.Run("to create three tasks for next delete update and list operations, should pass", func(t *testing.T) {
		_, err := datasource.GetDataSource().TaskDao().Create(context.Background(), taskOne)
		assert.Nil(t, err)
		_, err = datasource.GetDataSource().TaskDao().Create(context.Background(), taskTwo)
		assert.Nil(t, err)
		_, err = datasource.GetDataSource().TaskDao().Create(context.Background(), taskThree)
		assert.Nil(t, err)
	})

	t.Run("list task service", func(t *testing.T) {
		t.Run("list task with default domain and default project should pass", func(t *testing.T) {
			listReq := model.ListTaskRequest{
				Domain:  "task",
				Project: "task",
			}
			tasks, err := task.List(context.Background(), &listReq)
			assert.Nil(t, err)
			assert.Equal(t, 3, len(tasks))
		})
	})

	t.Run("update task service", func(t *testing.T) {
		t.Run("set the status of the taskOne to done should pass", func(t *testing.T) {
			taskOne.Status = sync.DoneStatus
			err := task.Update(context.Background(), taskOne)
			assert.Nil(t, err)
			listReq := model.ListTaskRequest{
				Domain:       taskOne.Domain,
				Project:      taskOne.Project,
				Action:       taskOne.Action,
				ResourceType: taskOne.ResourceType,
				Status:       taskOne.Status,
			}
			tasks, err := task.List(context.Background(), &listReq)
			assert.Nil(t, err)
			assert.Equal(t, 1, len(tasks))
		})
	})

	t.Run("delete task service", func(t *testing.T) {
		t.Run("delete all tasks in default domain and default project should pass", func(t *testing.T) {
			listReq := model.ListTaskRequest{
				Domain:  "task",
				Project: "task",
			}
			tasks, err := task.List(context.Background(), &listReq)
			assert.Nil(t, err)
			assert.Equal(t, 3, len(tasks))
			err = task.Delete(context.Background(), tasks...)
			assert.Nil(t, err)
			dTasks, err := task.List(context.Background(), &listReq)
			assert.Nil(t, err)
			assert.Equal(t, 0, len(dTasks))
		})
	})
}
