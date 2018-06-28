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
	"sync/atomic"
	"time"
)

const (
	maxExecutorCount       = 1000
	removeExecutorInterval = 30 * time.Second
	initExecutorTTL        = 4
	executeInterval        = 1 * time.Second
	compactTimes           = 2
)

type executorWithTTL struct {
	*Executor

	TTL int64
}

type TaskService struct {
	executors map[string]*executorWithTTL
	goroutine *util.GoRoutine
	lock      sync.RWMutex
	ready     chan struct{}
	isClose   bool
}

func (lat *TaskService) getOrNewExecutor(task Task) (s *Executor, isNew bool) {
	var (
		ok  bool
		key = task.Key()
	)

	lat.lock.RLock()
	se, ok := lat.executors[key]
	lat.lock.RUnlock()
	if !ok {
		lat.lock.Lock()
		se, ok = lat.executors[key]
		if !ok {
			isNew = true
			se = &executorWithTTL{
				Executor: NewExecutor(lat.goroutine, task),
				TTL:      initExecutorTTL,
			}
			lat.executors[key] = se
		}
		lat.lock.Unlock()
	}
	atomic.StoreInt64(&se.TTL, initExecutorTTL)
	return se.Executor, isNew
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

func (lat *TaskService) removeExecutor(key string) {
	if s, ok := lat.executors[key]; ok {
		s.Close()
		delete(lat.executors, key)
	}
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
	ticker := time.NewTicker(removeExecutorInterval)
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
			slice := make([]*executorWithTTL, 0, l)
			for _, s := range lat.executors {
				slice = append(slice, s)
			}
			lat.lock.RUnlock()

			for _, s := range slice {
				s.Execute() // non-blocked
			}
		case <-ticker.C:
			timer.Stop()

			lat.lock.RLock()
			l := len(lat.executors)
			if l > max {
				max = l
			}

			removes := make([]string, 0, l)
			for key, se := range lat.executors {
				if atomic.AddInt64(&se.TTL, -1) == 0 {
					removes = append(removes, key)
				}
			}
			lat.lock.RUnlock()

			if len(removes) == 0 {
				continue
			}

			lat.lock.Lock()
			for _, key := range removes {
				lat.removeExecutor(key)
			}

			l = len(lat.executors)
			if max > maxExecutorCount && max > l*compactTimes {
				lat.renew()
				max = l
			}
			lat.lock.Unlock()

			util.Logger().Debugf("daemon thread completed, %d executor(s) removed", len(removes))
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
	newExecutor := make(map[string]*executorWithTTL, maxExecutorCount)
	for k, e := range lat.executors {
		newExecutor[k] = e
	}
	lat.executors = newExecutor
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
