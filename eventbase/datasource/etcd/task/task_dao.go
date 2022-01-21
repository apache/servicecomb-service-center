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
	"encoding/json"

	"github.com/apache/servicecomb-service-center/eventbase/datasource"
	"github.com/apache/servicecomb-service-center/eventbase/datasource/etcd/key"

	"github.com/go-chassis/cari/sync"
	"github.com/go-chassis/openlog"
	"github.com/little-cui/etcdadpt"
)

type Dao struct {
}

func (d *Dao) Create(ctx context.Context, task *sync.Task) (*sync.Task, error) {
	taskBytes, err := json.Marshal(task)
	if err != nil {
		return nil, err
	}
	ok, err := etcdadpt.InsertBytes(ctx, key.TaskKey(task.Domain, task.Project, task.ID, task.Timestamp), taskBytes)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, datasource.ErrTaskAlreadyExists
	}
	return task, nil
}

func (d *Dao) Update(ctx context.Context, task *sync.Task) error {
	keyTask := key.TaskKey(task.Domain, task.Project, task.ID, task.Timestamp)
	resp, err := etcdadpt.Get(ctx, keyTask)
	if err != nil {
		return err
	}
	if resp == nil {
		return datasource.ErrTaskNotExists
	}
	var dbTask sync.Task
	err = json.Unmarshal(resp.Value, &dbTask)
	if err != nil {
		return err
	}
	dbTask.Status = task.Status

	taskBytes, err := json.Marshal(dbTask)
	if err != nil {
		return err
	}
	return etcdadpt.PutBytes(ctx, keyTask, taskBytes)
}

func (d *Dao) Delete(ctx context.Context, tasks ...*sync.Task) error {
	delOptions := make([]etcdadpt.OpOptions, len(tasks))
	for i, task := range tasks {
		delOptions[i] = etcdadpt.OpDel(etcdadpt.WithStrKey(key.TaskKey(task.Domain, task.Project, task.ID, task.Timestamp)))
	}
	err := etcdadpt.Txn(ctx, delOptions)
	if err != nil {
		return err
	}
	return nil
}

func (d Dao) List(ctx context.Context, options ...datasource.TaskFindOption) ([]*sync.Task, error) {
	opts := datasource.NewTaskFindOptions()
	for _, o := range options {
		o(&opts)
	}
	tasks := make([]*sync.Task, 0)
	kvs, _, err := etcdadpt.List(ctx, key.TaskList(opts.Domain, opts.Project))
	if err != nil {
		return tasks, err
	}
	for _, kv := range kvs {
		task := sync.Task{}
		err := json.Unmarshal(kv.Value, &task)
		if err != nil {
			datasource.Logger().Error("unmarshal task failed", openlog.WithErr(err))
			continue
		}
		if !filterMatch(&task, opts) {
			continue
		}
		tasks = append(tasks, &task)
	}
	return tasks, nil
}

func filterMatch(task *sync.Task, options datasource.TaskFindOptions) bool {
	if options.Action != "" && task.Action != options.Action {
		return false
	}
	if options.ResourceType != "" && task.ResourceType != options.ResourceType {
		return false
	}
	if options.Status != "" && task.Status != options.Status {
		return false
	}
	return true
}
