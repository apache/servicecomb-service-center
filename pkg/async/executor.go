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
package async

import (
	"errors"
	"github.com/apache/incubator-servicecomb-service-center/pkg/util"
	"golang.org/x/net/context"
)

type Executor struct {
	pool       *util.GoRoutine
	tasks      *util.UniQueue
	latestTask Task
}

func (s *Executor) AddTask(task Task) (err error) {
	if task == nil {
		return errors.New("invalid parameters")
	}

	err = s.tasks.Put(task)
	if err != nil {
		return
	}
	return s.latestTask.Err()
}

func (s *Executor) Execute() {
	select {
	case task, ok := <-s.tasks.Chan():
		if !ok {
			return
		}
		s.pool.Do(func(ctx context.Context) {
			at := task.(Task)
			at.Do(ctx)
			s.latestTask = at
		})
	default:
	}
}

func (s *Executor) Close() {
	s.tasks.Close()
}

func NewExecutor(pool *util.GoRoutine, task Task) *Executor {
	return &Executor{
		pool:       pool,
		tasks:      util.NewUniQueue(),
		latestTask: task,
	}
}
