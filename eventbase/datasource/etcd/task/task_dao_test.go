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

	"github.com/go-chassis/cari/db"
	"github.com/go-chassis/cari/sync"
	"github.com/stretchr/testify/assert"

	"servicecomb-service-center/eventbase/datasource"
	"servicecomb-service-center/eventbase/datasource/etcd"
	"servicecomb-service-center/eventbase/test"
)

var ds datasource.DataSource

func init() {
	cfg := &db.Config{
		Kind: test.Etcd,
		URI:  test.EtcdURI,
	}
	ds, _ = etcd.NewDatasource(cfg)
}

func TestTask(t *testing.T) {
	var (
		task = sync.Task{
			TaskID:    "30b93187-2a38-49e3-ae99-1961b28329b0",
			Action:    "create",
			DataType:  "config",
			Domain:    "default",
			Project:   "default",
			Timestamp: 1638171566,
			Status:    "pending"}
		taskTwo = sync.Task{
			TaskID:    "40b93187-2a38-49e3-ae99-1961b28329b0",
			Action:    "update",
			DataType:  "config",
			Domain:    "default",
			Project:   "default",
			Timestamp: 1638171567,
			Status:    "done"}
		taskThree = sync.Task{
			TaskID:    "50b93187-2a38-49e3-ae99-1961b28329b0",
			Action:    "update",
			DataType:  "config",
			Domain:    "default",
			Project:   "default",
			Timestamp: 1638171568,
			Status:    "pending"}
	)

	t.Run("create task", func(t *testing.T) {
		t.Run("create a task should pass", func(t *testing.T) {
			_, err := ds.TaskDao().Create(context.Background(), &task)
			assert.NoError(t, err)
		})

		t.Run("create a same task should fail", func(t *testing.T) {
			_, err := ds.TaskDao().Create(context.Background(), &task)
			assert.NotNil(t, err)
		})

		t.Run("create taskTwo and taskThree should pass", func(t *testing.T) {
			_, err := ds.TaskDao().Create(context.Background(), &taskTwo)
			assert.NoError(t, err)
			_, err = ds.TaskDao().Create(context.Background(), &taskThree)
			assert.NoError(t, err)
		})
	})

	t.Run("update task", func(t *testing.T) {
		t.Run("update a existing task should pass", func(t *testing.T) {
			task.Status = "done"
			err := ds.TaskDao().Update(context.Background(), &task)
			assert.NoError(t, err)
		})

		t.Run("update a not existing task should fail", func(t *testing.T) {
			notExistTask := sync.Task{
				TaskID:    "not-exist",
				Action:    "create",
				DataType:  "config",
				Domain:    "default",
				Project:   "default",
				Timestamp: 1638171568,
				Status:    "pending",
			}
			err := ds.TaskDao().Update(context.Background(), &notExistTask)
			assert.NotNil(t, err)
		})
	})

	t.Run("list task", func(t *testing.T) {
		t.Run("list task with action ,dataType and status should pass", func(t *testing.T) {
			opts := []datasource.TaskFindOption{
				datasource.WithAction(task.Action),
				datasource.WithDataType(task.DataType),
				datasource.WithStatus(task.Status),
			}
			tasks, err := ds.TaskDao().List(context.Background(), task.Domain, task.Project, opts...)
			assert.NoError(t, err)
			assert.Equal(t, 1, len(tasks))
		})

		t.Run("list task without action ,dataType and status should pass", func(t *testing.T) {
			tasks, err := ds.TaskDao().List(context.Background(), "default", "default")
			assert.NoError(t, err)
			assert.Equal(t, 3, len(tasks))
			assert.Equal(t, tasks[0].Timestamp, task.Timestamp)
			assert.Equal(t, tasks[1].Timestamp, taskTwo.Timestamp)
			assert.Equal(t, tasks[2].Timestamp, taskThree.Timestamp)
		})

	})

	t.Run("delete task", func(t *testing.T) {
		t.Run("delete tasks should pass", func(t *testing.T) {
			err := ds.TaskDao().Delete(context.Background(), []*sync.Task{&task, &taskTwo, &taskThree}...)
			assert.NoError(t, err)
		})
	})
}
