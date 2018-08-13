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
package buffer

import (
	"strings"
	"testing"
)

func TestNewPool(t *testing.T) {
	p := NewPool(10)
	b := p.Get()
	if b == nil {
		t.Fatalf("TestNewPool falied")
	}
	b.WriteString("a")
	if b.String() != "a" {
		t.Fatalf("TestNewPool falied")
	}
	p.Put(b)
	b = p.Get()
	if b == nil || b.Len() != 0 {
		t.Fatalf("TestNewPool falied")
	}
}

func BenchmarkNewPool(b *testing.B) {
	p := NewPool(4 * 10)
	s := strings.Repeat("a", 4*1024)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			buf := p.Get()
			buf.WriteString(s)
			_ = buf.String()
			p.Put(buf)
		}
	})
	b.ReportAllocs()
	// 2000000	       872 ns/op	    4098 B/op	       1 allocs/op
}
