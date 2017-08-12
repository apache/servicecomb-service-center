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
	"github.com/ServiceComb/service-center/util"
	"golang.org/x/net/context"
	"sync"
	"time"
)

const (
	DEFAULT_MAX_TASK_COUNT        = 1000
	DEFAULT_REMOVE_TASKS_INTERVAL = 30 * time.Second
)

type AsyncTask interface {
	Key() string
	Do(ctx context.Context) error
	Err() error
}

type AsyncTasker interface {
	AddTask(ctx context.Context, task AsyncTask) error
	RemoveTask(key string) error
	LatestHandled(key string) (AsyncTask, error)
	Run()
	Stop()
	Ready() <-chan struct{}
}

type BaseAsyncTasker struct {
	queues      map[string]*util.UniQueue
	latestTasks map[string]AsyncTask
	removeTasks map[string]struct{}
	goroutine   *util.GoRoutine
	queueLock   sync.RWMutex
	ready       chan struct{}
	isClose     bool
}

func (lat *BaseAsyncTasker) AddTask(ctx context.Context, task AsyncTask) error {
	if task == nil || ctx == nil {
		return errors.New("invalid parameters")
	}
	lat.queueLock.RLock()
	queue, ok := lat.queues[task.Key()]
	latestTask := lat.latestTasks[task.Key()]
	lat.queueLock.RUnlock()
	if !ok {
		lat.queueLock.Lock()
		queue, ok = lat.queues[task.Key()]
		if !ok {
			queue = util.NewUniQueue()
			lat.queues[task.Key()] = queue
			latestTask = task
			lat.latestTasks[task.Key()] = latestTask
		}
		lat.queueLock.Unlock()
	}

	if !ok || lat.isClose {
		// do immediately at first time
		return task.Do(ctx)
	}

	err := queue.Put(ctx, task)
	if err != nil {
		return err
	}
	util.LOGGER.Debugf("add task done! key is %s", task.Key())

	handled, err := lat.LatestHandled(task.Key())
	if err != nil {
		return err
	}
	return handled.Err()
}

func (lat *BaseAsyncTasker) RemoveTask(key string) error {
	lat.queueLock.Lock()
	if lat.isClose {
		lat.queueLock.Unlock()
		return errors.New("AsyncTasker is stopped")
	}
	lat.removeTasks[key] = struct{}{}
	lat.queueLock.Unlock()
	return nil
}

func (lat *BaseAsyncTasker) removeTask(key string) {
	lat.queueLock.Lock()
	delete(lat.queues, key)
	delete(lat.latestTasks, key)
	delete(lat.removeTasks, key)
	lat.queueLock.Unlock()
	util.LOGGER.Debugf("remove task, key is %s", key)
}

func (lat *BaseAsyncTasker) LatestHandled(key string) (AsyncTask, error) {
	lat.queueLock.RLock()
	at, ok := lat.latestTasks[key]
	lat.queueLock.RUnlock()
	if !ok {
		return nil, errors.New("expired behavior")
	}
	return at, nil
}

func (lat *BaseAsyncTasker) schedule(stopCh <-chan struct{}) {
	util.SafeCloseChan(lat.ready)
	ready := make(chan AsyncTask, DEFAULT_MAX_TASK_COUNT)
	defer func() {
		close(ready)
		util.LOGGER.Debugf("AsyncTasker is ready to stop")
	}()
	for {
		select {
		case <-stopCh:
			util.LOGGER.Debugf("scheduler exited for AsyncTasker is stopped")
			return
		default:
			go lat.collectReadyTasks(ready)

			ctx, _ := context.WithTimeout(context.Background(), time.Second)
			select {
			case task := <-ready:
				lat.scheduleTask(task.(AsyncTask))
				lat.scheduleReadyTasks(ready)
			case <-ctx.Done():
				util.LOGGER.Debugf("timed out to collect ready tasks")
			}
		}
	}
}

func (lat *BaseAsyncTasker) daemonRemoveTask(stopCh <-chan struct{}) {
	for {
		select {
		case <-stopCh:
			util.LOGGER.Debugf("daemon remove task exited for AsyncTasker is stopped")
			return
		case <-time.After(DEFAULT_REMOVE_TASKS_INTERVAL):
			if lat.isClose {
				return
			}
			removes := make([]string, 0, len(lat.removeTasks))
			lat.queueLock.RLock()
			for key := range lat.removeTasks {
				removes = append(removes, key)
			}
			lat.queueLock.RUnlock()
			for _, key := range removes {
				lat.removeTask(key)
			}
			util.LOGGER.Debugf("daemon remove task is running, %d removed", len(removes))
		}
	}
}

func (lat *BaseAsyncTasker) Run() {
	lat.queueLock.Lock()
	if !lat.isClose {
		lat.queueLock.Unlock()
		return
	}
	lat.isClose = false
	lat.queueLock.Unlock()
	lat.goroutine.Do(lat.schedule)
	lat.goroutine.Do(lat.daemonRemoveTask)
}

func (lat *BaseAsyncTasker) scheduleReadyTasks(ready <-chan AsyncTask) {
	ctx, _ := context.WithTimeout(context.Background(), time.Second)
	for {
		select {
		case task := <-ready:
			lat.scheduleTask(task.(AsyncTask))
		case <-ctx.Done():
			return
		}
	}
}

func (lat *BaseAsyncTasker) collectReadyTasks(ready chan<- AsyncTask) {
	lat.queueLock.RLock()
	for key, queue := range lat.queues {
		select {
		case task, ok := <-queue.Chan():
			util.LOGGER.Debugf("get task in queue[%v], key is %s", ok, key)
			if !ok {
				continue
			}
			ready <- task.(AsyncTask) // will block when a lot of tasks coming in.
		default:
			util.LOGGER.Debugf("no task in queue, key is %s", key)
		}
	}
	lat.queueLock.RUnlock()
}

func (lat *BaseAsyncTasker) scheduleTask(at AsyncTask) {
	util.LOGGER.Debugf("start to run task, key is %s", at.Key())
	lat.goroutine.Do(func(stopCh <-chan struct{}) {
		ctx, cancel := context.WithCancel(context.Background())
		go func() {
			defer cancel()

			lat.queueLock.RLock()
			_, ok := lat.latestTasks[at.Key()]
			if !ok {
				lat.queueLock.RUnlock()
				util.LOGGER.Debugf("task is removed, key is %s", at.Key())
				return
			}
			lat.queueLock.RUnlock()

			at.Do(ctx)

			lat.queueLock.Lock()
			lat.latestTasks[at.Key()] = at
			lat.queueLock.Unlock()
		}()
		select {
		case <-ctx.Done():
			util.LOGGER.Debugf("finish to handle task, key is %s", at.Key())
		case <-stopCh:
			cancel()
			util.LOGGER.Debugf("cancelled task for AsyncTasker is stopped, key is %s", at.Key())
		}
	})
}

func (lat *BaseAsyncTasker) Stop() {
	lat.queueLock.Lock()
	if lat.isClose {
		lat.queueLock.Unlock()
		return
	}
	lat.isClose = true
	for key, queue := range lat.queues {
		queue.Close()
		delete(lat.queues, key)
		delete(lat.latestTasks, key)
	}
	for key := range lat.removeTasks {
		delete(lat.removeTasks, key)
	}
	lat.queueLock.Unlock()
	lat.goroutine.Close(true)

	util.SafeCloseChan(lat.ready)

	util.LOGGER.Debugf("AsyncTasker is stopped")
}

func (lat *BaseAsyncTasker) Ready() <-chan struct{} {
	return lat.ready
}

func NewAsyncTasker() AsyncTasker {
	return &BaseAsyncTasker{
		latestTasks: make(map[string]AsyncTask),
		queues:      make(map[string]*util.UniQueue),
		removeTasks: make(map[string]struct{}),
		goroutine:   util.NewGo(make(chan struct{})),
		ready:       make(chan struct{}),
		isClose:     true,
	}
}
