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
	"sync"
	"time"
)

const (
	maxExecutorCount    = 1000
	removeTasksInterval = 30 * time.Second
	executeInterval     = 100 * time.Millisecond
	compactTimes        = 2
)

type TaskService struct {
	executors         map[string]*Executor
	removingExecutors map[string]struct{}
	goroutine         *util.GoRoutine
	lock              sync.RWMutex
	ready             chan struct{}
	isClose           bool
}

func (lat *TaskService) getOrNewExecutor(task Task) (s *Executor, isNew bool) {
	var (
		ok  bool
		key = task.Key()
	)

	lat.lock.RLock()
	s, ok = lat.executors[key]
	_, remove := lat.removingExecutors[key]
	lat.lock.RUnlock()
	if !ok {
		lat.lock.Lock()
		s, ok = lat.executors[key]
		if !ok {
			isNew = true
			s = NewExecutor(lat.goroutine, task)
			lat.executors[key] = s
		}
		lat.lock.Unlock()
	}
	if remove && ok {
		lat.lock.Lock()
		_, remove = lat.removingExecutors[key]
		if remove {
			delete(lat.removingExecutors, key)
		}
		lat.lock.Unlock()
	}
	return
}

func (lat *TaskService) Add(ctx context.Context, task Task) error {
	if task == nil || ctx == nil {
		return errors.New("invalid parameters")
	}

	s, isNew := lat.getOrNewExecutor(task)
	if isNew {
		// do immediately at first time
		return task.Do(ctx)
	}
	return s.AddTask(task)
}

func (lat *TaskService) DeferRemove(key string) error {
	lat.lock.Lock()
	if lat.isClose {
		lat.lock.Unlock()
		return errors.New("TaskService is stopped")
	}
	_, exist := lat.executors[key]
	if !exist {
		lat.lock.Unlock()
		return nil
	}
	lat.removingExecutors[key] = struct{}{}
	lat.lock.Unlock()
	return nil
}

func (lat *TaskService) removeExecutor(key string) {
	if s, ok := lat.executors[key]; ok {
		s.Close()
		delete(lat.executors, key)
	}
	delete(lat.removingExecutors, key)
	util.Logger().Debugf("remove executor, key is %s", key)
}

func (lat *TaskService) LatestHandled(key string) (Task, error) {
	lat.lock.RLock()
	s, ok := lat.executors[key]
	lat.lock.RUnlock()
	if !ok {
		return nil, errors.New("expired behavior")
	}
	return s.latestTask, nil
}

func (lat *TaskService) daemon(ctx context.Context) {
	util.SafeCloseChan(lat.ready)
	ticker := time.NewTicker(removeTasksInterval)
	max := 0
	for {
		timer := time.NewTimer(executeInterval)
		select {
		case <-ctx.Done():
			timer.Stop()

			util.Logger().Debugf("daemon thread exited for TaskService is stopped")
			return
		case <-timer.C:
			lat.lock.RLock()
			l := len(lat.executors)
			slice := make([]*Executor, 0, l)
			for _, s := range lat.executors {
				slice = append(slice, s)
			}
			lat.lock.RUnlock()

			for _, s := range slice {
				s.Execute() // non-blocked
			}
		case <-ticker.C:
			timer.Stop()

			lat.lock.Lock()
			l, rl := len(lat.executors), len(lat.removingExecutors)
			if l > max {
				max = l
			}

			for key := range lat.removingExecutors {
				lat.removeExecutor(key)
			}

			l = len(lat.executors)
			if max > maxExecutorCount && max > l*compactTimes {
				lat.renew()
				max = l
			}
			lat.lock.Unlock()

			if rl > 0 {
				util.Logger().Debugf("daemon thread completed, %d executor(s) removed", rl)
			}
		}
	}
}

func (lat *TaskService) Run() {
	lat.lock.Lock()
	if !lat.isClose {
		lat.lock.Unlock()
		return
	}
	lat.isClose = false
	lat.lock.Unlock()
	lat.goroutine.Do(lat.daemon)
}

func (lat *TaskService) Stop() {
	lat.lock.Lock()
	if lat.isClose {
		lat.lock.Unlock()
		return
	}
	lat.isClose = true

	for key := range lat.executors {
		lat.removeExecutor(key)
	}

	lat.lock.Unlock()

	lat.goroutine.Close(true)

	util.SafeCloseChan(lat.ready)
}

func (lat *TaskService) Ready() <-chan struct{} {
	return lat.ready
}

func (lat *TaskService) renew() {
	newExecutor := make(map[string]*Executor, maxExecutorCount)
	for k, e := range lat.executors {
		newExecutor[k] = e
	}
	lat.executors = newExecutor
	lat.removingExecutors = make(map[string]struct{}, maxExecutorCount)
}

func NewTaskService() (lat *TaskService) {
	lat = &TaskService{
		goroutine: util.NewGo(context.Background()),
		ready:     make(chan struct{}),
		isClose:   true,
	}
	lat.renew()
	return
}
