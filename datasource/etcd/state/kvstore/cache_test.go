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

package kvstore

import (
	"testing"
)

func TestNullCache_Name(t *testing.T) {
	NullCache.Put("", nil)
	if NullCache.Name() != "NULL" {
		t.Fatalf("TestNullCache_Name failed")
	}
	if NullCache.Size() != 0 {
		t.Fatalf("TestNullCache_Name failed")
	}
	if NullCache.Get("") != nil {
		t.Fatalf("TestNullCache_Name failed")
	}
	if NullCache.GetAll(nil) != 0 {
		t.Fatalf("TestNullCache_Name failed")
	}
	if NullCache.GetPrefix("", nil) != 0 {
		t.Fatalf("TestNullCache_Name failed")
	}
	NullCache.ForEach(func(k string, v *KeyValue) (next bool) {
		t.Fatalf("TestNullCache_Name failed")
		return false
	})
	NullCache.Remove("")

	if NullCacher.Cache() != NullCache {
		t.Fatalf("TestNullCache_Name failed")
	}
}

func TestKvCache_Get(t *testing.T) {
	c := NewKvCache("test", NewOptions())
	c.Put("", &KeyValue{Version: 1})
	c.Put("/", &KeyValue{Version: 1})
	c.Put("/a/b/c/d/e/1", &KeyValue{Version: 1})
	c.Put("/a/b/c/d/e/2", &KeyValue{Version: 2})
	c.Put("/a/b/d/d/f/3", &KeyValue{Version: 3})
	c.Put("/a/b/e/d/g/4", &KeyValue{Version: 4})

	if s := c.Size(); s == 0 {
		t.Fatalf("TestKvCache Size() failed, %d", s)
	}

	if l := c.GetAll(nil); l != 4 {
		t.Fatalf("TestKvCache GetAll() failed, %d", l)
	}

	if kv := c.Get("/a/b/c/d/e/2"); kv == nil || kv.Version != 2 {
		t.Fatalf("TestKvCache Get() failed, %v", kv)
	}

	if l := c.GetPrefix("/", nil); l != 4 {
		t.Fatalf("TestKvCache GetPrefix() failed, %d", l)
	}

	var arr []*KeyValue
	if l := c.GetPrefix("/a/b/c/", &arr); l != 2 || (arr[0].Version != 1 && arr[1].Version != 1) {
		t.Fatalf("TestKvCache GetPrefix() failed, %d, %v", l, arr)
	}

	l, b := -1, false
	c.ForEach(func(k string, v *KeyValue) (next bool) {
		next = false
		l++
		return
	})
	if l != 0 {
		t.Fatalf("TestKvCache ForEach() failed, %d", l)
	}
	c.ForEach(func(k string, v *KeyValue) (next bool) {
		next = true
		l++
		if v.Version == 4 {
			b = true
		}
		return
	})
	if l != 4 || !b {
		t.Fatalf("TestKvCache ForEach() failed, %d, %v", l, b)
	}

	c.Remove("")
	c.Remove("/")
	c.Remove("/a/b/c/d/e/2")
	c.Remove("/a/b/d/d/f/3")
	if l := c.GetAll(nil); l != 2 {
		t.Fatalf("TestKvCache GetAll() failed, %d", l)
	}

	c.Put("/a/b/c/d/e/1", &KeyValue{Version: 2})
	if kv := c.Get("/a/b/c/d/e/1"); kv == nil || kv.Version != 2 {
		t.Fatalf("TestKvCache Put() failed, %v", kv)
	}

	c.Put("/x//p/*", &KeyValue{Version: 1})
	c.Put("/x//p/app/name/version1", &KeyValue{Version: 2})
	c.Put("/x//p/app/name/version2", &KeyValue{Version: 3})
	n := c.GetPrefix("/x//p/", nil)
	if n != 3 {
		t.Fatalf("TestKvCache Put() failed, %v", n)
	}
	kvs := make([]*KeyValue, 0, n)
	c.GetPrefix("/x//p/", &kvs)
	if len(kvs) != 3 {
		t.Fatalf("TestKvCache Put() failed, %v", kvs)
	}
	for _, kv := range kvs {
		if kv.Version != 1 && kv.Version != 2 && kv.Version != 3 {
			t.Fatalf("TestKvCache Put() failed, %v", kvs)
		}
	}
}

func BenchmarkKvCache_GetAll1(b *testing.B) {
	c := NewKvCache("test", NewOptions())
	c.Put("/a/b/c/d/e/1", &KeyValue{Version: 1})
	c.Put("/a/b/c/d/e/2", &KeyValue{Version: 2})
	c.Put("/a/b/d/d/f/3", &KeyValue{Version: 3})
	c.Put("/a/b/e/d/g/4", &KeyValue{Version: 4})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = c.GetAll(nil)
	}
	b.ReportAllocs()
	// 1000000	      1269 ns/op	       0 B/op	       0 allocs/op
}

func BenchmarkKvCache_GetAll2(b *testing.B) {
	c := NewKvCache("test", NewOptions())
	c.Put("/a/b/c/d/e/1", &KeyValue{Version: 1})
	c.Put("/a/b/c/d/e/2", &KeyValue{Version: 2})
	c.Put("/a/b/d/d/f/3", &KeyValue{Version: 3})
	c.Put("/a/b/e/d/g/4", &KeyValue{Version: 4})
	b.ResetTimer()
	var arr []*KeyValue
	for i := 0; i < b.N; i++ {
		_ = c.GetAll(&arr)
	}
	b.ReportAllocs()
	// 1000000	      2784 ns/op	     173 B/op	       0 allocs/op
}
