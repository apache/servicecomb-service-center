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
	"errors"
	"github.com/apache/servicecomb-service-center/pkg/gopool"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/util"
	"golang.org/x/net/context"
	"sync"
	"sync/atomic"
	"time"
)

const (
	initExecutorCount      = 1000
	removeExecutorInterval = 30 * time.Second
	initExecutorTTL        = 4
	executeInterval        = 1 * time.Second
	compactTimes           = 2
)

type executorWithTTL struct {
	*Executor
	TTL int64
}

type AsyncTaskService struct {
	executors map[string]*executorWithTTL
	goroutine *gopool.Pool
	lock      sync.RWMutex
	ready     chan struct{}
	isClose   bool
}

func (lat *AsyncTaskService) getOrNewExecutor(task Task) (s *Executor, isNew bool) {
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

func (lat *AsyncTaskService) Add(ctx context.Context, task Task) error {
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

func (lat *AsyncTaskService) removeExecutor(key string) {
	if s, ok := lat.executors[key]; ok {
		s.Close()
		delete(lat.executors, key)
	}
}

func (lat *AsyncTaskService) LatestHandled(key string) (Task, error) {
	lat.lock.RLock()
	s, ok := lat.executors[key]
	lat.lock.RUnlock()
	if !ok {
		return nil, errors.New("expired behavior")
	}
	return s.latestTask, nil
}

func (lat *AsyncTaskService) daemon(ctx context.Context) {
	util.SafeCloseChan(lat.ready)
	ticker := time.NewTicker(removeExecutorInterval)
	max := 0
	timer := time.NewTimer(executeInterval)
	defer timer.Stop()
	for {
		select {
		case <-ctx.Done():
			log.Debugf("daemon thread exited for AsyncTaskService stopped")
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

			timer.Reset(executeInterval)
		case <-ticker.C:
			util.ResetTimer(timer, executeInterval)

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
			if max > initExecutorCount && max > l*compactTimes {
				lat.renew()
				max = l
			}
			lat.lock.Unlock()

			log.Debugf("daemon thread completed, %d executor(s) removed", len(removes))
		}
	}
}

func (lat *AsyncTaskService) Run() {
	lat.lock.Lock()
	if !lat.isClose {
		lat.lock.Unlock()
		return
	}
	lat.isClose = false
	lat.lock.Unlock()
	lat.goroutine.Do(lat.daemon)
}

func (lat *AsyncTaskService) Stop() {
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

func (lat *AsyncTaskService) Ready() <-chan struct{} {
	return lat.ready
}

func (lat *AsyncTaskService) renew() {
	newExecutor := make(map[string]*executorWithTTL)
	for k, e := range lat.executors {
		newExecutor[k] = e
	}
	lat.executors = newExecutor
}

func NewTaskService() TaskService {
	lat := &AsyncTaskService{
		goroutine: gopool.New(context.Background()),
		ready:     make(chan struct{}),
		isClose:   true,
	}
	lat.renew()
	return lat
}
