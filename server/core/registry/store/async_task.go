//Copyright 2017 Huawei Technologies Co., Ltd
//
//Licensed under the Apache License, Version 2.0 (the "License");
//you may not use this file except in compliance with the License.
//You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
//Unless required by applicable law or agreed to in writing, software
//distributed under the License is distributed on an "AS IS" BASIS,
//WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//See the License for the specific language governing permissions and
//limitations under the License.
package store

import (
	"errors"
	"github.com/ServiceComb/service-center/pkg/util"
	"golang.org/x/net/context"
	"sync"
	"time"
)

const (
	DEFAULT_MAX_SCHEDULE_COUNT    = 10
	DEFAULT_REMOVE_TASKS_INTERVAL = 30 * time.Second
)

type AsyncTask interface {
	Key() string
	Do(ctx context.Context) error
	Err() error
}

type scheduler struct {
	queue      *util.UniQueue
	latestTask AsyncTask
	once       sync.Once
}

func (s *scheduler) AddTask(ctx context.Context, task AsyncTask) (err error) {
	if task == nil || ctx == nil {
		return errors.New("invalid parameters")
	}

	s.once.Do(func() {
		go s.do()
	})

	err = s.queue.Put(ctx, task)
	if err != nil {
		return
	}
	return s.latestTask.Err()
}

func (s *scheduler) do() {
	for {
		select {
		case task, ok := <-s.queue.Chan():
			if !ok {
				util.Logger().Warnf(nil, "scheduler is closed, key is %s", s.latestTask.Key())
				return
			}
			at := task.(AsyncTask)
			at.Do(context.Background())
			s.latestTask = at
		}
	}
}

func (s *scheduler) Close() {
	s.queue.Close()
}

type AsyncTasker struct {
	schedules   map[string]*scheduler
	removeTasks map[string]struct{}
	goroutine   *util.GoRoutine
	lock        sync.RWMutex
	ready       chan struct{}
	isClose     bool
}

func (lat *AsyncTasker) getOrNewScheduler(task AsyncTask) (s *scheduler, isNew bool) {
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
			s = &scheduler{
				queue:      util.NewUniQueue(),
				latestTask: task,
			}
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

func (lat *AsyncTasker) AddTask(ctx context.Context, task AsyncTask) error {
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

func (lat *AsyncTasker) DeferRemoveTask(key string) error {
	lat.lock.Lock()
	if lat.isClose {
		lat.lock.Unlock()
		return errors.New("AsyncTasker is stopped")
	}
	_, exist := lat.schedules[key]
	if !exist {
		lat.lock.Unlock()
		util.Logger().Warnf(nil, "an unused scheduler will be removed, key is %s", key)
		return nil
	}
	lat.removeTasks[key] = struct{}{}
	lat.lock.Unlock()
	return nil
}

func (lat *AsyncTasker) removeScheduler(key string) {
	if s, ok := lat.schedules[key]; ok {
		s.Close()
		delete(lat.schedules, key)
	}
	delete(lat.removeTasks, key)
	util.Logger().Debugf("remove scheduler, key is %s", key)
}

func (lat *AsyncTasker) LatestHandled(key string) (AsyncTask, error) {
	lat.lock.RLock()
	s, ok := lat.schedules[key]
	lat.lock.RUnlock()
	if !ok {
		return nil, errors.New("expired behavior")
	}
	return s.latestTask, nil
}

func (lat *AsyncTasker) daemon(stopCh <-chan struct{}) {
	util.SafeCloseChan(lat.ready)
	for {
		select {
		case <-stopCh:
			util.Logger().Debugf("daemon thread exited for AsyncTasker is stopped")
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
			lat.lock.Unlock()
			if l > 0 {
				util.Logger().Infof("daemon thread completed, %d scheduler(s) removed", l)
			}
		}
	}
}

func (lat *AsyncTasker) Run() {
	lat.lock.Lock()
	if !lat.isClose {
		lat.lock.Unlock()
		return
	}
	lat.isClose = false
	lat.lock.Unlock()
	lat.goroutine.Do(lat.daemon)
}

func (lat *AsyncTasker) Stop() {
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

	util.Logger().Debugf("AsyncTasker is stopped")
}

func (lat *AsyncTasker) Ready() <-chan struct{} {
	return lat.ready
}

func NewAsyncTasker() *AsyncTasker {
	return &AsyncTasker{
		schedules:   make(map[string]*scheduler, DEFAULT_MAX_SCHEDULE_COUNT),
		removeTasks: make(map[string]struct{}, DEFAULT_MAX_SCHEDULE_COUNT),
		goroutine:   util.NewGo(make(chan struct{})),
		ready:       make(chan struct{}),
		isClose:     true,
	}
}
