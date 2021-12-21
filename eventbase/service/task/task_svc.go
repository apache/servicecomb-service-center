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

	"github.com/go-chassis/cari/sync"

	"github.com/apache/servicecomb-service-center/eventbase/datasource"
	"github.com/apache/servicecomb-service-center/eventbase/request"
)

func Delete(ctx context.Context, tasks ...*sync.Task) error {
	return datasource.GetTaskDao().Delete(ctx, tasks...)
}

func Update(ctx context.Context, task *sync.Task) error {
	return datasource.GetTaskDao().Update(ctx, task)
}

func List(ctx context.Context, request *request.ListTaskRequest) ([]*sync.Task, error) {
	opts := []datasource.TaskFindOption{
		datasource.WithDomain(request.Domain),
		datasource.WithProject(request.Project),
		datasource.WithAction(request.TaskAction),
		datasource.WithDataType(request.TaskDataType),
		datasource.WithStatus(request.TaskStatus),
	}
	tasks, err := datasource.GetTaskDao().List(ctx, opts...)
	if err != nil {
		return nil, err
	}
	return tasks, nil
}
