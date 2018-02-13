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
	v := LoadStruct(obj1)
	if v.Type.String() != "util.testStru" {
		fail(t, "TestLoadStruct failed, %s != 'testStru'", v.Type.String())
	}
	if len(v.Fields) != 4 {
		fail(t, "TestLoadStruct failed, wrong count of fields")
	}
	for _, f := range v.Fields {
		fmt.Println(f.Name, f.Type.String())
	}

	obj2 := testStru{}
	v = LoadStruct(obj2)
}

func BenchmarkLoadStruct(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			obj := testStru{}
			LoadStruct(obj)
		}
	})
	b.ReportAllocs()
	// 20000000	        86.9 ns/op	      32 B/op	       1 allocs/op
}
