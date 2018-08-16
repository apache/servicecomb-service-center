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
	"errors"
	"math/rand"
	"testing"
)

func TestConcurrentMap(t *testing.T) {
	cm := ConcurrentMap{}
	s := cm.Size()
	if s != 0 {
		t.Fatalf("TestConcurrentMap Size failed.")
	}
	v, b := cm.Get("a")
	if b || v != nil {
		t.Fatalf("TestConcurrentMap Get a not exist item failed.")
	}
	cm.Put("a", "1")
	v, b = cm.Get("a")
	if !b || v.(string) != "1" {
		t.Fatalf("TestConcurrentMap Get an exist item failed.")
	}
	cm.Put("a", "2")
	v, b = cm.Get("a")
	if v.(string) != "2" {
		t.Fatalf("TestConcurrentMap Put an item again failed.")
	}
	cm.PutIfAbsent("b", "1")
	cm.PutIfAbsent("a", "3")
	v, b = cm.Get("a")
	if !b || v.(string) != "2" {
		t.Fatalf("TestConcurrentMap Get an item after PutIfAbsent failed.")
	}
	cm.Remove("a")
	v, b = cm.Get("a")
	if b || v != nil {
		t.Fatalf("TestConcurrentMap Get an item after Remove failed.")
	}
	s = cm.Size()
	if s != 1 { // only 'b' is left
		t.Fatalf("TestConcurrentMap Size after Put failed.")
	}
	cm.Clear()
	s = cm.Size()
	if s != 0 {
		t.Fatalf("TestConcurrentMap Size after Clear failed.")
	}
}

func TestConcurrentMap_ForEach(t *testing.T) {
	l := 0
	cm := ConcurrentMap{}
	cm.ForEach(func(item MapItem) bool {
		l++
		return true
	})
	if l != 0 {
		t.Fatalf("TestConcurrentMap_ForEach failed.")
	}
	for i := 0; i < 1000; i++ {
		cm.Put(i, i)
	}
	cm.ForEach(func(item MapItem) bool {
		l++
		cm.Remove(item.Key)
		return true
	})
	if l != 1000 || cm.Size() != 0 {
		t.Fatalf("TestConcurrentMap_ForEach does not empty failed.")
	}
}

func TestConcurrentMap_Fetch(t *testing.T) {
	cm := ConcurrentMap{}
	v, err := cm.Fetch("a", func() (interface{}, error) {
		return "a", nil
	})
	if err != nil || v != "a" {
		t.Fatalf("TestConcurrentMap_Fetch failed.")
	}
	v, err = cm.Fetch("a", func() (interface{}, error) {
		return "b", nil
	})
	if err != nil || v != "a" {
		t.Fatalf("TestConcurrentMap_Fetch failed.")
	}
	v, err = cm.Fetch("b", func() (interface{}, error) {
		return nil, errors.New("err")
	})
	if err == nil || v != nil {
		t.Fatalf("TestConcurrentMap_Fetch failed.")
	}
}

func BenchmarkConcurrentMap_Get(b *testing.B) {
	var v interface{}
	cm := ConcurrentMap{}
	cm.Put("a", "1")
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			v, _ = cm.Get("a")
		}
	})
	b.ReportAllocs()
	// go1.9- 20000000	        95.8 ns/op	       0 B/op	       0 allocs/op
	// go1.9+ 50000000	        30.2 ns/op	       0 B/op	       0 allocs/op
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
	// go1.9- 3000000	       424 ns/op	      32 B/op	       2 allocs/op
	// go1.9+ 5000000	       333 ns/op	      16 B/op	       1 allocs/op
}

func BenchmarkConcurrentMap_PutAndGet(b *testing.B) {
	cm := &ConcurrentMap{}
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			i := rand.Intn(10)
			cm.Put(i, i)
			_, _ = cm.Get(i)
		}
	})
	b.ReportAllocs()
	// go1.9- 5000000	       300 ns/op	      32 B/op	       2 allocs/op
	// go1.9+ 5000000	       294 ns/op	      30 B/op	       2 allocs/op
}

func BenchmarkConcurrentMap_ForEach(b *testing.B) {
	cm := ConcurrentMap{}
	for i := 0; i < 100; i++ {
		cm.Put(i, i)
	}
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			cm.ForEach(func(item MapItem) bool {
				return true
			})
		}
	})
	b.ReportAllocs()
	// go1.9- 1000000	      1096 ns/op	    3200 B/op	       1 allocs/op
	// go1.9+ 3000000	       394 ns/op	       0 B/op	       0 allocs/op
}

func BenchmarkConcurrentMap_PutAndForEach(b *testing.B) {
	cm := ConcurrentMap{}
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			i := rand.Intn(10)
			cm.ForEach(func(item MapItem) bool {
				return true
			})
			cm.Put(i, i)
		}
	})
	b.ReportAllocs()
	// go1.9- 2000000	       747 ns/op	     336 B/op	       3 allocs/op
	// go1.9+ 5000000	       301 ns/op	      30 B/op	       2 allocs/op
}

func BenchmarkConcurrentMap_Fetch(b *testing.B) {
	cm := ConcurrentMap{}
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			i := rand.Intn(10)
			_, _ = cm.Fetch(i, func() (interface{}, error) {
				return i, nil
			})
		}
	})
	b.ReportAllocs()
	// go1.9- 5000000	       274 ns/op	       8 B/op	       1 allocs/op
	// go1.9+ 5000000	       277 ns/op	       7 B/op	       0 allocs/op
}

func BenchmarkConcurrentMap_PutAndFetch(b *testing.B) {
	cm := ConcurrentMap{}
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			i := rand.Intn(10)
			_, _ = cm.Fetch(i, func() (interface{}, error) {
				return i, nil
			})
			cm.Put(i, i)
		}
	})
	b.ReportAllocs()
	// go1.9- 5000000	       346 ns/op	      24 B/op	       3 allocs/op
	// go1.9+ 5000000	       305 ns/op	      37 B/op	       3 allocs/op
}
