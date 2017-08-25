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

type AsyncTasker struct {
	queues      map[string]*util.UniQueue
	latestTasks map[string]AsyncTask
	removeTasks map[string]struct{}
	goroutine   *util.GoRoutine
	queueLock   sync.RWMutex
	ready       chan struct{}
	isClose     bool
}

func (lat *AsyncTasker) getOrNewQueue(key string, task AsyncTask) (*util.UniQueue, bool) {
	lat.queueLock.RLock()
	queue, ok := lat.queues[key]
	_, remove := lat.removeTasks[key]
	lat.queueLock.RUnlock()
	if !ok {
		lat.queueLock.Lock()
		queue, ok = lat.queues[key]
		if !ok {
			queue = util.NewUniQueue()
			lat.queues[key] = queue
			lat.latestTasks[key] = task
		}
		lat.queueLock.Unlock()
	}
	if remove && ok {
		lat.queueLock.Lock()
		_, remove = lat.removeTasks[key]
		if remove {
			delete(lat.removeTasks, key)
		}
		lat.queueLock.Unlock()
	}
	return queue, !ok
}

func (lat *AsyncTasker) AddTask(ctx context.Context, task AsyncTask) error {
	if task == nil || ctx == nil {
		return errors.New("invalid parameters")
	}

	queue, isNew := lat.getOrNewQueue(task.Key(), task)
	if isNew || lat.isClose {
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

func (lat *AsyncTasker) DeferRemoveTask(key string) error {
	lat.queueLock.Lock()
	if lat.isClose {
		lat.queueLock.Unlock()
		return errors.New("AsyncTasker is stopped")
	}
	_, exist := lat.queues[key]
	if !exist {
		lat.queueLock.Unlock()
		return nil
	}
	lat.removeTasks[key] = struct{}{}
	lat.queueLock.Unlock()
	return nil
}

func (lat *AsyncTasker) removeTask(key string) {
	lat.queueLock.Lock()
	delete(lat.queues, key)
	delete(lat.latestTasks, key)
	delete(lat.removeTasks, key)
	lat.queueLock.Unlock()
	util.LOGGER.Debugf("remove task, key is %s", key)
}

func (lat *AsyncTasker) LatestHandled(key string) (AsyncTask, error) {
	lat.queueLock.RLock()
	at, ok := lat.latestTasks[key]
	lat.queueLock.RUnlock()
	if !ok {
		return nil, errors.New("expired behavior")
	}
	return at, nil
}

func (lat *AsyncTasker) schedule(stopCh <-chan struct{}) {
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
			}
		}
	}
}

func (lat *AsyncTasker) daemonRemoveTask(stopCh <-chan struct{}) {
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

func (lat *AsyncTasker) Run() {
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

func (lat *AsyncTasker) scheduleReadyTasks(ready <-chan AsyncTask) {
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

func (lat *AsyncTasker) collectReadyTasks(ready chan<- AsyncTask) {
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
		}
	}
	lat.queueLock.RUnlock()
}

func (lat *AsyncTasker) scheduleTask(at AsyncTask) {
	util.LOGGER.Debugf("start to run task, key is %s", at.Key())
	lat.goroutine.Do(func(stopCh <-chan struct{}) {
		ctx, cancel := context.WithCancel(context.Background())
		go func() {
			defer cancel()

			lat.queueLock.RLock()
			_, ok := lat.latestTasks[at.Key()]
			lat.queueLock.RUnlock()
			if !ok {
				util.LOGGER.Debugf("task is removed, key is %s", at.Key())
				return
			}

			at.Do(ctx)

			lat.UpdateLatestTask(at)
		}()
		select {
		case <-ctx.Done():
			util.LOGGER.Debugf("finish to handle task, key is %s, result: %s", at.Key(), at.Err())
		case <-stopCh:
			cancel()
			util.LOGGER.Debugf("cancelled task for AsyncTasker is stopped, key is %s", at.Key())
		}
	})
}

func (lat *AsyncTasker) UpdateLatestTask(at AsyncTask) {
	lat.queueLock.RLock()
	_, ok := lat.latestTasks[at.Key()]
	lat.queueLock.RUnlock()
	if !ok {
		return
	}

	lat.queueLock.Lock()
	_, ok = lat.latestTasks[at.Key()]
	if ok {
		lat.latestTasks[at.Key()] = at
	}
	lat.queueLock.Unlock()
}

func (lat *AsyncTasker) Stop() {
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

func (lat *AsyncTasker) Ready() <-chan struct{} {
	return lat.ready
}

func NewAsyncTasker() *AsyncTasker {
	return &AsyncTasker{
		latestTasks: make(map[string]AsyncTask),
		queues:      make(map[string]*util.UniQueue),
		removeTasks: make(map[string]struct{}),
		goroutine:   util.NewGo(make(chan struct{})),
		ready:       make(chan struct{}),
		isClose:     true,
	}
}
