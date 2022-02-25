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

	"github.com/apache/servicecomb-service-center/eventbase/model"
	servicetask "github.com/apache/servicecomb-service-center/eventbase/service/task"
	carisync "github.com/go-chassis/cari/sync"
)

func ListTask(ctx context.Context) ([]*carisync.Task, error) {
	return servicetask.List(ctx, &model.ListTaskRequest{})
}

type syncTasks []*carisync.Task

func (s syncTasks) Len() int {
	return len(s)
}

func (s syncTasks) Less(i, j int) bool {
	return s[i].Timestamp <= s[j].Timestamp
}

func (s syncTasks) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
