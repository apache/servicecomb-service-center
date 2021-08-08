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
	"errors"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

type testTask struct {
	done      context.CancelFunc
	returnErr string
	sleep     time.Duration
}

func (tt *testTask) Key() string {
	return "test"
}

func (tt *testTask) Err() error {
	if len(tt.returnErr) == 0 {
		return nil
	}
	return errors.New(tt.returnErr)
}

func (tt *testTask) Do(ctx context.Context) error {
	if tt.done != nil {
		defer tt.done()
	}
	wait := tt.sleep
	if wait == 0 {
		<-time.After(time.Second)
	} else {
		<-time.After(wait)
	}
	return tt.Err()
}

func TestBaseAsyncTasker_AddTask(t *testing.T) {
	at := NewTaskService()
	at.Run()
	<-at.Ready()
	defer at.Stop()

	ctx, cancel := context.WithCancel(context.Background())
	err := at.Add(ctx, nil)
	assert.Error(t, err)
	cancel()

	testCtx1, testC1 := context.WithCancel(context.Background())
	err = at.Add(testCtx1, &testTask{
		done:      testC1,
		returnErr: "test1",
	})
	assert.Error(t, testCtx1.Err())
	assert.Equal(t, "test1", err.Error())

	lt, _ := at.LatestHandled("test")
	assert.Equal(t, "test1", lt.Err().Error())

	testCtx2, testC2 := context.WithCancel(context.Background())
	err = at.Add(testCtx2, &testTask{
		done:      testC2,
		returnErr: "test2",
	})
	assert.Equal(t, "test1", err.Error())

	<-testCtx2.Done()
	// pkg/task/executor.go:53
	<-time.After(time.Millisecond)
	lt, _ = at.LatestHandled("test")
	assert.Equal(t, "test2", lt.Err().Error())
}

func TestBaseAsyncTasker_Stop(t *testing.T) {
	at := NewTaskService()
	at.Stop()
	at.Run()

	_, cancel := context.WithCancel(context.Background())
	err := at.Add(context.Background(), &testTask{
		done: cancel,
	})
	if err != nil {
		t.Fatalf("add task should be ok")
	}
	_, cancel = context.WithCancel(context.Background())
	err = at.Add(context.Background(), &testTask{
		done:  cancel,
		sleep: 3 * time.Second,
	})
	if err != nil {
		t.Fatalf("add task should be ok")
	}
	<-time.After(time.Second)
	at.Stop()

	err = at.Add(context.Background(), &testTask{})
	if err != nil {
		t.Fatalf("add task should be ok when Tasker is stopped")
	}

	at.Stop()
}
