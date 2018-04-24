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
	DEFAULT_MAX_SCHEDULE_COUNT    = 1000
	DEFAULT_REMOVE_TASKS_INTERVAL = 30 * time.Second
)

type Task interface {
	Key() string
	Do(ctx context.Context) error
	Err() error
}

type scheduler struct {
	queue      *util.UniQueue
	latestTask Task
	once       sync.Once
	goroutine  *util.GoRoutine
}

func (s *scheduler) AddTask(ctx context.Context, task Task) (err error) {
	if task == nil || ctx == nil {
		return errors.New("invalid parameters")
	}

	s.once.Do(func() {
		s.goroutine.Do(s.do)
	})

	err = s.queue.Put(ctx, task)
	if err != nil {
		return
	}
	return s.latestTask.Err()
}

func (s *scheduler) do(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case task, ok := <-s.queue.Chan():
			if !ok {
				return
			}
			at := task.(Task)
			at.Do(ctx)
			s.latestTask = at
		}
	}
}

func (s *scheduler) Close() {
	s.queue.Close()
	s.goroutine.Close(true)
}

func newScheduler(task Task) *scheduler {
	return &scheduler{
		queue:      util.NewUniQueue(),
		latestTask: task,
		goroutine:  util.NewGo(context.Background()),
	}
}

type TaskService struct {
	schedules   map[string]*scheduler
	removeTasks map[string]struct{}
	goroutine   *util.GoRoutine
	lock        sync.RWMutex
	ready       chan struct{}
	isClose     bool
}

func (lat *TaskService) getOrNewScheduler(task Task) (s *scheduler, isNew bool) {
	var (
		ok  bool
		key = task.Key()
	)

	lat.lock.RLock()
	s, ok = lat.schedules[key]
	_, remove := lat.removeTasks[key]
	lat.lock.RUnlock()
	if !ok {
		lat.lock.Lock()
		s, ok = lat.schedules[key]
		if !ok {
			isNew = true
			s = newScheduler(task)
			lat.schedules[key] = s
		}
		lat.lock.Unlock()
	}
	if remove && ok {
		lat.lock.Lock()
		_, remove = lat.removeTasks[key]
		if remove {
			delete(lat.removeTasks, key)
		}
		lat.lock.Unlock()
	}
	return
}

func (lat *TaskService) Add(ctx context.Context, task Task) error {
	if task == nil || ctx == nil {
		return errors.New("invalid parameters")
	}

	s, isNew := lat.getOrNewScheduler(task)
	if isNew {
		// do immediately at first time
		return task.Do(ctx)
	}
	return s.AddTask(ctx, task)
}

func (lat *TaskService) DeferRemove(key string) error {
	lat.lock.Lock()
	if lat.isClose {
		lat.lock.Unlock()
		return errors.New("TaskService is stopped")
	}
	_, exist := lat.schedules[key]
	if !exist {
		lat.lock.Unlock()
		return nil
	}
	lat.removeTasks[key] = struct{}{}
	lat.lock.Unlock()
	return nil
}

func (lat *TaskService) removeScheduler(key string) {
	if s, ok := lat.schedules[key]; ok {
		s.Close()
		delete(lat.schedules, key)
	}
	delete(lat.removeTasks, key)
	util.Logger().Debugf("remove scheduler, key is %s", key)
}

func (lat *TaskService) LatestHandled(key string) (Task, error) {
	lat.lock.RLock()
	s, ok := lat.schedules[key]
	lat.lock.RUnlock()
	if !ok {
		return nil, errors.New("expired behavior")
	}
	return s.latestTask, nil
}

func (lat *TaskService) daemon(ctx context.Context) {
	util.SafeCloseChan(lat.ready)
	for {
		select {
		case <-ctx.Done():
			util.Logger().Debugf("daemon thread exited for TaskService is stopped")
			return
		case <-time.After(DEFAULT_REMOVE_TASKS_INTERVAL):
			if lat.isClose {
				return
			}
			lat.lock.Lock()
			l := len(lat.removeTasks)
			for key := range lat.removeTasks {
				lat.removeScheduler(key)
			}

			if l > DEFAULT_MAX_SCHEDULE_COUNT {
				lat.renew()
			}
			lat.lock.Unlock()
			if l > 0 {
				util.Logger().Infof("daemon thread completed, %d scheduler(s) removed", l)
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

	for key := range lat.schedules {
		lat.removeScheduler(key)
	}

	lat.lock.Unlock()

	lat.goroutine.Close(true)

	util.SafeCloseChan(lat.ready)
}

func (lat *TaskService) Ready() <-chan struct{} {
	return lat.ready
}

func (lat *TaskService) renew() {
	lat.schedules = make(map[string]*scheduler, DEFAULT_MAX_SCHEDULE_COUNT)
	lat.removeTasks = make(map[string]struct{}, DEFAULT_MAX_SCHEDULE_COUNT)
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
