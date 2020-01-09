// Licensed to the Apache Software Foundation (ASF) under one or more
// contributor license agreements.  See the NOTICE file distributed with
// this work for additional information regarding copyright ownership.
// The ASF licenses this file to You under the Apache License, Version 2.0
// (the "License"); you may not use this file except in compliance with
// the License.  You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package queue

import (
	"github.com/apache/servicecomb-service-center/pkg/gopool"
	"golang.org/x/net/context"
)

const (
	eventQueueSize = 1000
)

type Worker interface {
	Handle(ctx context.Context, obj interface{})
}

type Task struct {
	Object interface{}
	// Async can let workers handle this task concurrently, but
	// it will make this task unordered
	Async bool
}

type TaskQueue struct {
	Workers []Worker

	taskCh    chan Task
	goroutine *gopool.Pool
}

// AddWorker is the method to add Worker
func (q *TaskQueue) AddWorker(w Worker) {
	q.Workers = append(q.Workers, w)
}

// Add is the method to add task in queue, one task will be handled by all workers
func (q *TaskQueue) Add(t Task) {
	q.taskCh <- t
}

func (q *TaskQueue) dispatch(ctx context.Context, w Worker, obj interface{}) {
	w.Handle(ctx, obj)
}

// Do is the method to trigger workers handle the task immediately
func (q *TaskQueue) Do(ctx context.Context, task Task) {
	if task.Async {
		for _, w := range q.Workers {
			q.goroutine.Do(func(ctx context.Context) {
				q.dispatch(ctx, w, task.Object)
			})
		}
		return
	}
	for _, w := range q.Workers {
		q.dispatch(ctx, w, task.Object)
	}
}

// Run is the method to start a goroutine to pull and handle tasks from queue
func (q *TaskQueue) Run() {
	q.goroutine.Do(func(ctx context.Context) {
		for {
			select {
			case <-ctx.Done():
				return
			case task := <-q.taskCh:
				q.Do(ctx, task)
			}
		}
	})
}

// Stop is the method to stop the workers gracefully
func (q *TaskQueue) Stop() {
	q.goroutine.Close(true)
}

func NewTaskQueue(size int) *TaskQueue {
	if size <= 0 {
		size = eventQueueSize
	}
	return &TaskQueue{
		taskCh:    make(chan Task, size),
		goroutine: gopool.New(context.Background()),
	}
}
