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
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/apache/servicecomb-service-center/eventbase/service/task"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/util"
	v1sync "github.com/apache/servicecomb-service-center/syncer/api/v1"
	"github.com/apache/servicecomb-service-center/syncer/metrics"
	"github.com/apache/servicecomb-service-center/syncer/service/event"
	"github.com/apache/servicecomb-service-center/syncer/service/replicator/resource"

	carisync "github.com/go-chassis/cari/sync"
	"github.com/go-chassis/foundation/gopool"
)

const (
	defaultInternal = 2 * time.Second

	heartbeatInternal = 15 * time.Second
	taskTTL           = 30
	taskName          = "load--handle-task"
)

func Work() {
	work()
}

func work() {
	dl := DistributedLock{
		key:               taskName,
		heartbeatDuration: heartbeatInternal,
		ttl:               taskTTL,
		do: func(ctx context.Context) {
			m := NewManager()
			m.LoadAndHandleTask(ctx)
			m.UpdateResultTask(ctx)
		},
	}
	dl.LockDo()
}

// Manager defines task manager, transfer task to event, and send event to event manager
type Manager interface {
	LoadAndHandleTask(ctx context.Context)
	UpdateResultTask(ctx context.Context)
}

type ManagerOption func(*managerOptions)

type managerOptions struct {
	internal    time.Duration
	operator    Operator
	eventSender event.Sender
}

func toManagerOptions(os ...ManagerOption) *managerOptions {
	mo := new(managerOptions)
	mo.internal = defaultInternal
	mo.eventSender = event.GetManager()

	for _, o := range os {
		o(mo)
	}
	return mo
}

func ManagerInternal(i time.Duration) ManagerOption {
	return func(mo *managerOptions) {
		mo.internal = i
	}
}

func EventSender(e event.Sender) ManagerOption {
	return func(options *managerOptions) {
		options.eventSender = e
	}
}

func ManagerOperator(l Operator) ManagerOption {
	return func(mo *managerOptions) {
		mo.operator = l
	}
}

func NewManager(os ...ManagerOption) Manager {
	m := &manager{
		toHandleTasks: make([]*carisync.Task, 0, 10),
		result:        make(chan *event.Result, 1000),
	}

	mo := toManagerOptions(os...)
	if mo.operator == nil {
		mo.operator = m
	}

	m.internal = mo.internal
	m.operator = mo.operator
	m.eventSender = mo.eventSender
	return m
}

type manager struct {
	internal time.Duration
	ticker   *time.Ticker

	toHandleTasks []*carisync.Task

	isClosing bool
	result    chan *event.Result
	cache     sync.Map

	operator    Operator
	eventSender event.Sender
}

// Operator define task operator, to list tasks and delete task
type Operator interface {
	ListTasks(ctx context.Context) ([]*carisync.Task, error)
	DeleteTask(ctx context.Context, t *carisync.Task) error
}

func (m *manager) LoadAndHandleTask(ctx context.Context) {
	gopool.Go(func(goctx context.Context) {
		m.ticker = time.NewTicker(m.internal)
		for {
			select {
			case _, ok := <-m.ticker.C:
				if !ok {
					log.Info("ticker is closed")
					return
				}

				ts, err := m.operator.ListTasks(ctx)
				if err != nil {
					log.Error("load task failed", err)
					continue
				}

				m.handleTasks(ts)
			case <-goctx.Done():
				m.Close()
				return
			case <-ctx.Done():
				m.Close()
				return
			}
		}
	})
}

func (m *manager) Close() {
	m.ticker.Stop()
}

func (m *manager) ListTasks(ctx context.Context) ([]*carisync.Task, error) {
	tasks, err := ListTask(ctx)
	if err != nil {
		return nil, err
	}

	metrics.PendingTaskSet(int64(len(tasks)))

	noHandleTasks := make([]*carisync.Task, 0, len(tasks))
	skipTaskIDs := make([]string, 0, len(tasks))
	allTaskIDs := make([]string, 0, len(tasks))
	for _, t := range tasks {
		log.Info(fmt.Sprintf("list task id: %v", t.ID))
		_, ok := m.cache.Load(t.ID)
		if ok {
			skipTaskIDs = append(skipTaskIDs, t.ID)
			continue
		}
		m.cache.Store(t.ID, t)
		allTaskIDs = append(allTaskIDs, t.ID)
		noHandleTasks = append(noHandleTasks, t)
		log.Info(fmt.Sprintf("no handle task id: %v", t.ID))
	}
	allTaskIDs = append(allTaskIDs, skipTaskIDs...)
	m.cache.Range(func(key, value any) bool {
		needDelete := true
		for _, tID := range allTaskIDs {
			if tID == key {
				needDelete = false
				break
			}
		}
		if needDelete {
			m.cache.Delete(key)
		}
		return true
	})

	log.Info(fmt.Sprintf("load task raw count %d, to handle count %d, skip ids %v",
		len(tasks), len(noHandleTasks), skipTaskIDs))

	return noHandleTasks, nil
}

func (m *manager) DeleteTask(ctx context.Context, t *carisync.Task) error {
	return task.Delete(ctx, t)
}

func (m *manager) UpdateResultTask(ctx context.Context) {
	gopool.Go(func(goctx context.Context) {
		log.Info("start updateTasks task")
		for {
			select {
			case res := <-m.result:
				if m.isClosing {
					m.closeUpdateTasks()
				}

				m.handleResult(res)
			case <-ctx.Done():
				m.isClosing = true

			case <-goctx.Done():
				log.Info("updateTasks exit")
				return
			}
		}
	})
}

func (m *manager) closeUpdateTasks() {
	c := 0
	m.cache.Range(func(_, _ interface{}) bool {
		c++
		return true
	})

	if c != 0 {
		return
	}

	close(m.result)
}

func (m *manager) handleResult(res *event.Result) {
	if res.Error != nil || res.Data.Code == resource.Fail {
		log.Error(fmt.Sprintf("get task %s result, return error", res.ID), res.Error)
		m.cache.Range(func(key, value interface{}) bool {
			m.cache.Delete(key)
			return true
		})
		return
	}

	if res.Data == nil {
		log.Info("result data is empty")
		return
	}

	log.Info(fmt.Sprintf("key: %s,result: %v", res.ID, res.Data))

	t, ok := m.cache.Load(res.ID)
	if !ok {
		return
	}

	code := res.Data.Code
	if code != resource.Fail {
		tk := t.(*carisync.Task)
		err := m.operator.DeleteTask(context.TODO(), tk)
		if err != nil {
			log.Error("delete task failed", err)
		}
	}
}

func (m *manager) handleTasks(sts syncTasks) {
	sort.Sort(sts)

	for _, st := range sts {
		m.eventSender.Send(toEvent(st, m.result))
	}
}

func toEvent(task *carisync.Task, result chan<- *event.Result) *event.Event {
	ops := task.Opts
	if len(ops) == 0 {
		ops = make(map[string]string, 2)
	}

	ops[string(util.CtxDomain)] = task.Domain
	ops[string(util.CtxProject)] = task.Project
	return &event.Event{
		Event: &v1sync.Event{
			Id:        task.ID,
			Action:    task.Action,
			Subject:   task.ResourceType,
			Opts:      ops,
			Value:     task.Resource,
			Timestamp: task.Timestamp,
		},
		CanNotAbandon: true,
		Result:        result,
	}
}
