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

package lb

import (
	"testing"
)

func TestNewRoundRobinLB(t *testing.T) {
	lb := NewRoundRobinLB(nil)
	if lb.Next() != "" {
		t.Fatalf("TestNewRoundRobinLB failed")
	}
	lb = NewRoundRobinLB([]string{"1"})
	if lb.Next() != "1" {
		t.Fatalf("TestNewRoundRobinLB failed")
	}
	if lb.Next() != "1" {
		t.Fatalf("TestNewRoundRobinLB failed")
	}
	lb = NewRoundRobinLB([]string{"1", "2"})
	if lb.Next() != "1" {
		t.Fatalf("TestNewRoundRobinLB failed")
	}
	if lb.Next() != "2" {
		t.Fatalf("TestNewRoundRobinLB failed")
	}
	if lb.Next() != "1" {
		t.Fatalf("TestNewRoundRobinLB failed")
	}
}

func BenchmarkNewRoundLB(b *testing.B) {
	lb := NewRoundRobinLB([]string{"1", "2", "3"})
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = lb.Next()
		}
	})
}
