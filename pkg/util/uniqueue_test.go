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
	"math"
	"sync/atomic"
	"testing"
	"time"
)

func TestNewUniQueue(t *testing.T) {
	_, err := newUniQueue(0)
	if err == nil {
		fail(t, "newUniQueue(0) should return error")
	}
	_, err = newUniQueue(math.MaxInt32)
	if err == nil {
		fail(t, "newUniQueue(math.MaxInt32) should return error")
	}
	uq, err := newUniQueue(1)
	if err != nil || uq == nil {
		fail(t, "newUniQueue(1) should return ok")
	}
	uq = NewUniQueue()
	if uq == nil {
		fail(t, "NewUniQueue should return ok")
	}
}

func TestUniQueue_Close(t *testing.T) {
	uq := NewUniQueue()
	err := uq.Put(context.Background(), "abc")
	if err != nil {
		fail(t, "NewUniQueue should return ok")
	}

	uq.Close()

	item := uq.Get(context.Background())
	if item != nil {
		fail(t, "Get expect '%v' to 'nil' when queue closed", item)
	}

	uq.Close()
}

func TestUniQueue_Get(t *testing.T) {
	uq := NewUniQueue()
	defer uq.Close()

	ctx, _ := context.WithTimeout(context.Background(), time.Second)
	start := time.Now()
	item := uq.Get(ctx)
	if time.Now().Sub(start) < time.Second || item != nil {
		fail(t, "Get should be timed out, result: %v", item)
	}

	err := uq.Put(context.Background(), "abc")
	if err != nil {
		fail(t, "Put('abc') should be ok")
	}
	err = uq.Put(context.Background(), "efg")
	if err != nil {
		fail(t, "Put('efg') should be ok")
	}

	time.Sleep(time.Second)

	ctx, _ = context.WithTimeout(context.Background(), time.Second)
	item = uq.Get(ctx)
	if item == nil || item.(string) != "efg" {
		fail(t, "Get expect '%v' to 'efg'", item)
	}
}

func TestUniQueue_Put(t *testing.T) {
	uq, err := newUniQueue(1)

	ctx, cancel := context.WithCancel(context.Background())
	err = uq.Put(ctx, 1)
	if err != nil {
		fail(t, "Put(1) should be ok")
	}
	cancel()
	err = uq.Put(ctx, 2)
	if err == nil {
		fail(t, "Put(2) should return 'timed out' error ")
	}
	uq.Close()
	err = uq.Put(context.Background(), 3)
	if err == nil {
		fail(t, "Put(3) should return 'channel is closed' error")
	}
}

func BenchmarkUniQueue_Get(b *testing.B) {
	var g int32 = 0
	closed := make(chan struct{})
	uq := NewUniQueue()
	go func() {
		for {
			item := uq.Get(context.Background())
			if item == nil {
				fmt.Println("Parallel:", b.N)
				close(closed)
				return
			}
			<-time.After(50 * time.Millisecond)
			fmt.Println(item.(int32))
		}
	}()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			err := uq.Put(context.Background(), atomic.AddInt32(&g, 1))
			if err != nil {
				b.FailNow()
			}
		}
	})
	uq.Close()
	<-closed
	b.ReportAllocs()
}
