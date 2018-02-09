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
	"testing"
)

func TestConcurrentMap(t *testing.T) {
	cm := &ConcurrentMap{}
	v, b := cm.Get("a")
	if b || v != nil {
		fmt.Println("TestConcurrentMap Get a not exist item failed.")
		t.Fail()
	}
	v = cm.Put("a", "1")
	if v != nil {
		fmt.Println("TestConcurrentMap Put a new item failed.")
		t.Fail()
	}
	v, b = cm.Get("a")
	if !b || v.(string) != "1" {
		fmt.Println("TestConcurrentMap Get an exist item failed.")
		t.Fail()
	}
	v = cm.Put("a", "2")
	if v.(string) != "1" {
		fmt.Println("TestConcurrentMap Put an item again failed.")
		t.Fail()
	}
	v = cm.PutIfAbsent("b", "1")
	if v != nil {
		fmt.Println("TestConcurrentMap PutIfAbsent a not exist item failed.")
		t.Fail()
	}
	v = cm.PutIfAbsent("a", "3")
	if v.(string) != "2" {
		fmt.Println("TestConcurrentMap PutIfAbsent an item failed.")
		t.Fail()
	}
	v, b = cm.Get("a")
	if !b || v.(string) != "2" {
		fmt.Println("TestConcurrentMap Get an item after PutIfAbsent failed.")
		t.Fail()
	}
}

func TestNewConcurrentMap(t *testing.T) {
	cm := NewConcurrentMap(100)
	if cm == nil {
		fmt.Println("TestNewConcurrentMap failed.")
		t.Fail()
	}
}

func BenchmarkConcurrentMap_Get(b *testing.B) {
	var v interface{}
	cm := &ConcurrentMap{}
	cm.Put("a", "1")
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			v, _ = cm.Get("a")
		}
	})
	b.ReportAllocs()
	// 20000000	        88.7 ns/op	       0 B/op	       0 allocs/op
}

func BenchmarkConcurrentMap_Put(b *testing.B) {
	cm := &ConcurrentMap{}
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			cm.Put("a", "1")
		}
	})
	b.ReportAllocs()
	// 3000000	       420 ns/op	      32 B/op	       2 allocs/op
}

func BenchmarkConcurrentMap_PutAndGet(b *testing.B) {
	var v interface{}
	cm := &ConcurrentMap{}
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			cm.Put("a", "1")
			v, _ = cm.Get("a")
		}
	})
	b.ReportAllocs()
	// 3000000	       560 ns/op	      32 B/op	       2 allocs/op
}
