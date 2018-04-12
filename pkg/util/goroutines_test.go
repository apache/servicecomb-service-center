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
	"golang.org/x/net/context"
	"sync"
	"testing"
	"time"
)

func TestGoRoutine_Do(t *testing.T) {
	test1 := NewGo(context.Background())
	defer test1.Close(true)
	stopCh1 := make(chan struct{})
	test1.Do(func(ctx context.Context) {
		defer close(stopCh1)
		select {
		case <-ctx.Done():
			fail(t, "ctx should not be done.")
		case <-time.After(time.Second):
		}
	})
	<-stopCh1

	ctx, cancel := context.WithCancel(context.Background())
	test2 := NewGo(ctx)
	defer test2.Close(true)
	stopCh2 := make(chan struct{})
	test2.Do(func(ctx context.Context) {
		defer close(stopCh2)
		select {
		case <-ctx.Done():
		case <-time.After(time.Second):
			fail(t, "time out to wait stopCh2 close.")
		}
	})
	cancel()
	<-stopCh2

	ctx, _ = context.WithTimeout(context.Background(), 0)
	test3 := NewGo(ctx)
	defer test3.Close(true)
	stopCh3 := make(chan struct{})
	test3.Do(func(ctx context.Context) {
		defer close(stopCh3)
		select {
		case <-ctx.Done():
		case <-time.After(time.Second):
			fail(t, "time out to wait ctx done.")
		}
	})
	<-stopCh3
}

func TestGoRoutine_Wait(t *testing.T) {
	var mux sync.Mutex
	MAX := 10
	resultArr := make([]int, 0, MAX)
	test := NewGo(context.Background())
	for i := 0; i < MAX; i++ {
		func(i int) {
			test.Do(func(ctx context.Context) {
				select {
				case <-ctx.Done():
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
	test := NewGo(context.Background())
	test.Do(func(ctx context.Context) {
		select {
		case <-ctx.Done():
		case <-time.After(time.Second):
			fail(t, "time out to wait ctx close.")
		}
	})
	test.Close(true)
	test.Close(true)
}

func TestGo(t *testing.T) {
	Go(func(ctx context.Context) {
		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(time.Second):
			}
		}
	})
	Go(func(ctx context.Context) {
		var a *int
		fmt.Println(*a)
	})
	GoCloseAndWait()
}
