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
package queue

import (
	"fmt"
	"golang.org/x/net/context"
	"sync/atomic"
	"testing"
	"time"
)

func TestNewUniQueue(t *testing.T) {
	uq := NewUniQueue()
	if uq == nil {
		t.Fatalf("NewUniQueue should return ok")
	}
}

func TestUniQueue_Close(t *testing.T) {
	uq := NewUniQueue()
	err := uq.Put("abc")
	if err != nil {
		t.Fatalf("NewUniQueue should return ok")
	}

	uq.Close()

	item := uq.Get(context.Background())
	if item != nil {
		t.Fatalf("Get expect '%v' to 'nil' when queue closed", item)
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
		t.Fatalf("Get should be timed out, result: %v", item)
	}

	err := uq.Put("abc")
	if err != nil {
		t.Fatalf("Put('abc') should be ok")
	}
	err = uq.Put("efg")
	if err != nil {
		t.Fatalf("Put('efg') should be ok")
	}

	ctx, _ = context.WithTimeout(context.Background(), time.Second)
	item = uq.Get(ctx)
	if item == nil || item.(string) != "efg" {
		t.Fatalf("Get expect '%v' to 'efg'", item)
	}

	err = uq.Put("abc")
	if err != nil {
		t.Fatalf("Put('abc') should be ok")
	}
	err = uq.Put("efg")
	if err != nil {
		t.Fatalf("Put('efg') should be ok")
	}

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		item = <-uq.Chan()
		if item == nil || item.(string) != "efg" {
			t.Fatalf("Get expect '%v' to 'efg'", item)
		}
		cancel()
	}()
	<-ctx.Done()
}

func TestUniQueue_Put(t *testing.T) {
	uq := NewUniQueue()

	err := uq.Put(1)
	if err != nil {
		t.Fatalf("Put(1) should be ok")
	}
	uq.Close()
	err = uq.Put(2)
	if err == nil {
		t.Fatalf("Put(2) should return 'channel is closed' error")
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
			err := uq.Put(atomic.AddInt32(&g, 1))
			if err != nil {
				b.FailNow()
			}
		}
	})
	uq.Close()
	<-closed
	b.ReportAllocs()
	// 5000000	       389 ns/op	       4 B/op	       1 allocs/op
}
