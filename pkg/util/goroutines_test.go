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
package util

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestGoRoutine_Init(t *testing.T) {
	var test GoRoutine
	stopCh1 := make(chan struct{})
	defer close(stopCh1)
	stopCh2 := make(chan struct{})
	defer close(stopCh2)

	test.Init(stopCh1)
	c := test.StopCh()
	if c != stopCh1 {
		fail(t, "init GoRoutine failed.")
	}

	test.Init(stopCh2)
	c = test.StopCh()
	if c == stopCh2 {
		fail(t, "init GoRoutine twice.")
	}
}

func TestGoRoutine_Do(t *testing.T) {
	var test1 GoRoutine
	stopCh := make(chan struct{})
	test1.Init(make(chan struct{}))
	test1.Do(func(neverStopCh <-chan struct{}) {
		defer close(stopCh)
		select {
		case <-neverStopCh:
			fail(t, "neverStopCh should not be closed.")
		case <-time.After(time.Second):
		}
	})
	<-stopCh

	var test2 GoRoutine
	stopCh1 := make(chan struct{})
	stopCh2 := make(chan struct{})
	test2.Init(stopCh1)
	test2.Do(func(stopCh <-chan struct{}) {
		defer close(stopCh2)
		select {
		case <-stopCh:
		case <-time.After(time.Second):
			fail(t, "time out to wait stopCh1 close.")
		}
	})
	close(stopCh1)
	<-stopCh2
}

func TestGoRoutine_Wait(t *testing.T) {
	var test GoRoutine
	var mux sync.Mutex
	MAX := 10
	resultArr := make([]int, 0, MAX)
	test.Init(make(chan struct{}))
	for i := 0; i < MAX; i++ {
		func(i int) {
			test.Do(func(neverStopCh <-chan struct{}) {
				select {
				case <-neverStopCh:
				case <-time.After(time.Second):
					mux.Lock()
					resultArr = append(resultArr, i)
					fmt.Printf("goroutine %d finish.\n", i)
					mux.Unlock()
				}

			})
		}(i)
	}
	fmt.Println("waiting for all goroutines finish.")
	test.Wait()
	fmt.Println(resultArr)
	if len(resultArr) != MAX {
		fail(t, "fail to wait all goroutines finish.")
	}
}

func TestGoRoutine_Close(t *testing.T) {
	var test GoRoutine
	test.Init(make(chan struct{}))
	test.Do(func(stopCh <-chan struct{}) {
		select {
		case <-stopCh:
		case <-time.After(time.Second):
			fail(t, "time out to wait stopCh close.")
		}
	})
	test.Close(true)
	test.Close(true)
}

func TestGo(t *testing.T) {
	GoInit()
	Go(func(stopCh <-chan struct{}) {
		for {
			select {
			case <-stopCh:
				return
			case <-time.After(time.Second):
			}
		}
	})
	Go(func(stopCh <-chan struct{}) {
		var a *int
		fmt.Println(*a)
	})
	GoCloseAndWait()
}

func TestNewGo(t *testing.T) {
	g := NewGo(make(chan struct{}))
	defer g.Close(true)
}
