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
	"github.com/ServiceComb/service-center/util"
	"golang.org/x/net/context"
	"sync"
	"time"
)

const DEFAULT_MAX_TASK_COUNT = 1000

type AsyncTask interface {
	Key() string
	Do(ctx context.Context) error
	Err() error
}

type AsyncTasker interface {
	AddTask(ctx context.Context, task AsyncTask) error
	LatestHandled(key string) AsyncTask
	Run()
	Stop()
}

type BaseAsyncTasker struct {
	queues      map[string]*util.UniQueue
	latestTasks map[string]AsyncTask
	firstReadys map[string]chan struct{}
	goroutine   *util.GoRoutine
	queueLock   sync.RWMutex
}

func (lat *BaseAsyncTasker) AddTask(ctx context.Context, task AsyncTask) error {
	lat.queueLock.RLock()
	queue, ok := lat.queues[task.Key()]
	latestTask := lat.latestTasks[task.Key()]
	firstReady := lat.firstReadys[task.Key()]
	lat.queueLock.RUnlock()
	if !ok {
		lat.queueLock.Lock()
		queue, ok = lat.queues[task.Key()]
		if !ok {
			queue = util.NewUniQueue()
			lat.queues[task.Key()] = queue
			latestTask = task
			lat.latestTasks[task.Key()] = latestTask
			firstReady = make(chan struct{})
			lat.firstReadys[task.Key()] = firstReady
		}
		lat.queueLock.Unlock()
	}

	err := queue.Put(ctx, task)
	if err != nil {
		return err
	}
	<-firstReady // sync firstly

	handled := lat.LatestHandled(task.Key())
	if handled != nil {
		return nil
	}
	return handled.Err()
}

func (lat *BaseAsyncTasker) LatestHandled(key string) AsyncTask {
	lat.queueLock.RLock()
	defer lat.queueLock.RUnlock()
	return lat.latestTasks[key]
}

func (lat *BaseAsyncTasker) Run() {
	lat.goroutine.Do(func(stopCh <-chan struct{}) {
		ready := make(chan AsyncTask, DEFAULT_MAX_TASK_COUNT)
		for {
			select {
			case <-stopCh:
				util.LOGGER.Debugf("BaseAsyncTasker is stopped")
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
		close(ready)
		util.LOGGER.Debugf("BaseAsyncTasker is ready to stop")
	})
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
			util.LOGGER.Debugf("get task in queue %v, key %s", ok, key)
			if !ok {
				continue
			}
			ready <- task.(AsyncTask) // will block when a lot of tasks coming in.
		default:
			util.LOGGER.Debugf("no task in queue, key %s", key)
		}
	}
	lat.queueLock.RUnlock()
}

func (lat *BaseAsyncTasker) scheduleTask(at AsyncTask) {
	util.LOGGER.Debugf("start to run task, key %s", at.Key())
	lat.goroutine.Do(func(stopCh <-chan struct{}) {
		ctx, cancel := context.WithCancel(context.Background())
		go func() {
			defer cancel()
			at.Do(ctx)

			lat.queueLock.Lock()
			lat.latestTasks[at.Key()] = at
			ready, ok := lat.firstReadys[at.Key()]
			if !ok {
				lat.queueLock.Unlock()
				return
			}
			lat.closeCh(ready)
			lat.queueLock.Unlock()
		}()
		select {
		case <-ctx.Done():
		case <-stopCh:
			cancel()
		}
		util.LOGGER.Debugf("finish to handle task %s", at.Key())
	})
}

func (lat *BaseAsyncTasker) closeCh(c chan struct{}) {
	select {
	case <-c:
	default:
		close(c)
	}
}

func (lat *BaseAsyncTasker) Stop() {
	lat.queueLock.Lock()
	for key, queue := range lat.queues {
		queue.Close()
		delete(lat.queues, key)

		c := lat.firstReadys[key]
		lat.closeCh(c)
		delete(lat.firstReadys, key)
		delete(lat.latestTasks, key)
	}
	lat.queueLock.Unlock()
	lat.goroutine.Close(true)
}

func NewAsyncTasker() AsyncTasker {
	return &BaseAsyncTasker{
		latestTasks: make(map[string]AsyncTask),
		queues:      make(map[string]*util.UniQueue),
		firstReadys: make(map[string]chan struct{}),
		goroutine:   util.NewGo(make(chan struct{})),
	}
}
