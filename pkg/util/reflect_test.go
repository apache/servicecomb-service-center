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

type testStru struct {
	f1 int
	f2 string
	f3 testField
	f4 *testField
}

type testField struct {
}

func TestLoadStruct(t *testing.T) {
	obj1 := testStru{}
	v := Reflect(obj1)
	if v.Type.String() != "util.testStru" {
		t.Fatalf("TestLoadStruct failed, %s != 'testStru'", v.Type.String())
	}
	if len(v.Fields) != 4 {
		t.Fatalf("TestLoadStruct failed, wrong count of fields")
	}
	if v.Name() != "pkg/util.testStru" || v.FullName != "github.com/apache/incubator-servicecomb-service-center/pkg/util.testStru" {
		t.Fatalf("TestLoadStruct failed")
	}
	for _, f := range v.Fields {
		fmt.Println(f.Name, f.Type.String())
	}

	obj2 := testStru{}
	v1 := Reflect(obj2)
	if v1.FullName != v.FullName {
		t.Fatalf("TestLoadStruct failed")
	}
	v2 := Reflect(&obj2)
	if v2.FullName != v.FullName {
		t.Fatalf("TestLoadStruct failed")
	}
	v = Reflect(nil)
	if v.FullName != "" {
		t.Fatalf("TestLoadStruct failed")
	}

	if FuncName(TestLoadStruct) != "pkg/util.TestLoadStruct" {
		t.Fatalf("TestLoadStruct failed")
	}
	f := TestLoadStruct
	if FormatFuncName(FuncName(f)) != "TestLoadStruct" {
		t.Fatalf("TestLoadStruct failed")
	}
}

func TestSizeof(t *testing.T) {
	s := &S{}
	if Sizeof(s) != 8+152 {
		t.Fatalf("TestSizeof failed")
	}
}

func BenchmarkLoadStruct(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			Reflect(testStru{})
		}
	})
	b.ReportAllocs()
	// 20000000	        86.9 ns/op	      32 B/op	       1 allocs/op
}

type S struct {
	a  int              // 8
	s  string           // 16
	p  *S               // 8
	m  map[int32]uint32 // 8
	u  []uint64         // 24
	ua [8]uint64        // 64
	ch chan int         // 8
	i  interface{}      // 16
}

func BenchmarkSizeof(b *testing.B) {
	s := &S{}
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			Sizeof(S{
				p: s,
				i: s,
			})
		}
	})
	b.ReportAllocs()
	// 2000000	       650 ns/op	     160 B/op	       1 allocs/op
}
