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
	"testing"

	"errors"
	"fmt"
	"github.com/ServiceComb/service-center/pkg/lager"
	"github.com/ServiceComb/service-center/pkg/util"
	"golang.org/x/net/context"
	"time"
)

func init() {
	util.InitLogger("async_task_test", &lager.Config{
		LoggerLevel:   "DEBUG",
		LoggerFile:    "",
		EnableRsyslog: false,
		LogFormatText: true,
		EnableStdOut:  true,
	})
}

func fail(t *testing.T, format string, args ...interface{}) {
	fmt.Printf(format, args...)
	fmt.Println()
	t.FailNow()
}

type testTask struct {
	done   context.CancelFunc
	test   string
	result bool
	wait   time.Duration
}

func (tt *testTask) Key() string {
	return "test"
}

func (tt *testTask) Err() error {
	if tt.result {
		return nil
	}
	return errors.New(tt.test)
}

func (tt *testTask) Do(ctx context.Context) error {
	if tt.done != nil {
		defer tt.done()
	}
	wait := tt.wait
	if wait == 0 {
		<-time.After(time.Second)
	} else {
		<-time.After(wait)
	}
	return tt.Err()
}

func TestBaseAsyncTasker_AddTask(t *testing.T) {
	at := NewAsyncTaskService()
	at.Run()
	defer at.Stop()

	ctx, cancel := context.WithCancel(context.Background())
	err := at.Add(ctx, nil)
	if err == nil {
		fail(t, "add nil task should be error")
	}
	cancel()

	testCtx1, testC1 := context.WithCancel(context.Background())
	err = at.Add(testCtx1, &testTask{
		done: testC1,
		test: "test1",
	})
	if testCtx1.Err() == nil || err == nil || err.Error() != "test1" {
		fail(t, "first time add task should be sync")
	}
	lt, _ := at.LatestHandled("test")
	if lt.Err().Error() != "test1" {
		fail(t, "should get first handled task 'test1'")
	}

	testCtx2, testC2 := context.WithTimeout(context.Background(), 3*time.Second)
	err = at.Add(testCtx2, &testTask{
		done: testC2,
		test: "test2",
	})
	if err.Error() != "test1" {
		fail(t, "second time add task should return prev result")
	}
	<-testCtx2.Done()
	lt, _ = at.LatestHandled("test")
	if lt.Err().Error() != "test2" {
		fail(t, "should get second handled task 'test2'")
	}
}

func TestBaseAsyncTasker_Stop(t *testing.T) {
	at := NewAsyncTaskService()
	at.Stop()
	at.Run()

	_, cancel := context.WithCancel(context.Background())
	err := at.Add(context.Background(), &testTask{
		done:   cancel,
		test:   "test stop",
		result: true,
	})
	if err != nil {
		fail(t, "add task should be ok")
	}
	_, cancel = context.WithCancel(context.Background())
	err = at.Add(context.Background(), &testTask{
		done:   cancel,
		test:   "test stop",
		wait:   3 * time.Second,
		result: true,
	})
	if err != nil {
		fail(t, "add task should be ok")
	}
	<-time.After(time.Second)
	at.Stop()

	err = at.Add(context.Background(), &testTask{result: true})
	if err != nil {
		fail(t, "add task should be ok when Tasker is stopped")
	}

	at.Stop()
}

func TestBaseAsyncTasker_RemoveTask(t *testing.T) {
	at := NewAsyncTaskService()
	at.Run()

	err := at.DeferRemove("test")
	if err != nil {
		fail(t, "remove task should be ok")
	}
	_, cancel := context.WithCancel(context.Background())
	err = at.Add(context.Background(), &testTask{
		done:   cancel,
		test:   "test remove task",
		result: true,
		wait:   33 * time.Second,
	})
	if err != nil {
		fail(t, "add task should be ok")
	}
	fmt.Println("OK")

	err = at.DeferRemove("test")
	if err != nil {
		fail(t, "remove task should be ok")
	}
	at.Stop()

	err = at.DeferRemove("test")
	if err == nil {
		fail(t, "remove task should be error when Tasker is stopped")
	}
}
