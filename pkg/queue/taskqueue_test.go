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
	"golang.org/x/net/context"
	"testing"
)

type mockWorker struct {
	Object chan interface{}
}

func (h *mockWorker) Handle(ctx context.Context, obj interface{}) {
	h.Object <- obj
}

func TestNewEventQueue(t *testing.T) {
	h := &mockWorker{make(chan interface{}, 1)}

	q := NewTaskQueue(0)
	q.AddWorker(h)

	q.Do(context.Background(), Task{Object: 1})
	if <-h.Object != 1 {
		t.Fatalf("TestNewEventQueue failed")
	}

	q.Do(context.Background(), Task{Object: 11, Async: true})
	if <-h.Object != 11 {
		t.Fatalf("TestNewEventQueue failed")
	}

	q.Run()
	q.Add(Task{Object: 2})
	if <-h.Object != 2 {
		t.Fatalf("TestNewEventQueue failed")
	}

	q.Add(Task{Object: 22, Async: true})
	if <-h.Object != 22 {
		t.Fatalf("TestNewEventQueue failed")
	}
	q.Stop()
	q.Add(Task{Object: 3})
}
