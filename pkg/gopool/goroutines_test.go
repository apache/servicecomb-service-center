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
package gopool

import (
	"fmt"
	"golang.org/x/net/context"
	"sync"
	"testing"
	"time"
)

func TestGoRoutine_Do(t *testing.T) {
	test1 := New(context.Background())
	defer test1.Close(true)
	stopCh1 := make(chan struct{})
	test1.Do(func(ctx context.Context) {
		defer close(stopCh1)
		select {
		case <-ctx.Done():
			t.Fatalf("ctx should not be done.")
		case <-time.After(time.Second):
		}
	})
	<-stopCh1

	ctx, cancel := context.WithCancel(context.Background())
	test2 := New(ctx)
	defer test2.Close(true)
	stopCh2 := make(chan struct{})
	test2.Do(func(ctx context.Context) {
		defer close(stopCh2)
		select {
		case <-ctx.Done():
		case <-time.After(time.Second):
			t.Fatalf("time out to wait stopCh2 close.")
		}
	})
	cancel()
	<-stopCh2
}

func TestGoRoutine_Wait(t *testing.T) {
	var mux sync.Mutex
	MAX := 10
	resultArr := make([]int, 0, MAX)
	test := New(context.Background(), Configure().Idle(time.Second).Workers(5))
	for i := 0; i < MAX; i++ {
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
	}
	<-time.After(time.Second)
	fmt.Println("waiting for all goroutines finish.")
	test.Done()
	fmt.Println(resultArr)
	if len(resultArr) != MAX {
		t.Fatalf("fail to wait all goroutines finish. expected %d to %d", len(resultArr), MAX)
	}
}

func TestGoRoutine_Exception1(t *testing.T) {
	test := New(context.Background())
	test.Do(func(ctx context.Context) {
		select {
		case <-ctx.Done():
		case <-time.After(time.Second):
			t.Fatalf("time out to wait ctx close.")
		}
	})
	test.Close(false)
	test.Do(func(ctx context.Context) {
		select {
		case <-ctx.Done():
			t.Fatalf("can not execute f after closed.")
		}
	})
	test.Done()
	test.Close(true)
}

func TestGoRoutine_Exception2(t *testing.T) {
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
	CloseAndWait()
}
