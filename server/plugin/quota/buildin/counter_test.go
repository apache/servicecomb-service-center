// Licensed to the Apache Software Foundation (ASF) under one or more
// contributor license agreements.  See the NOTICE file distributed with
// this work for additional information regarding copyright ownership.
// The ASF licenses this file to You under the Apache License, Version 2.0
// (the "License"); you may not use this file except in compliance with
// the License.  You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package buildin

import (
	kv "github.com/apache/servicecomb-service-center/datasource/etcd/pkg/kv"
	"testing"
)

func TestGlobalCounter_OnCreate(t *testing.T) {
	var counter GlobalCounter
	counter.OnCreate(kv.SERVICE, "a/b")
	counter.OnCreate(kv.ServiceIndex, "a/b")
	counter.OnCreate(kv.INSTANCE, "a/b")
	counter.OnCreate(kv.ServiceIndex, "a/b")
	counter.OnCreate(kv.INSTANCE, "a/b")
	if counter.ServiceCount != 2 || counter.InstanceCount != 2 {
		t.Fatal("TestGlobalCounter_OnCreate failed", counter)
	}
}

func TestGlobalCounter_OnDelete(t *testing.T) {
	var counter GlobalCounter
	counter.OnDelete(kv.SERVICE, "a/b")
	counter.OnDelete(kv.ServiceIndex, "a/b")
	counter.OnDelete(kv.INSTANCE, "a/b")
	if counter.ServiceCount != 0 || counter.InstanceCount != 0 {
		t.Fatal("TestGlobalCounter_OnDelete failed", counter)
	}
	counter.OnCreate(kv.ServiceIndex, "a/b")
	counter.OnCreate(kv.INSTANCE, "a/b")
	counter.OnDelete(kv.ServiceIndex, "a/b")
	counter.OnDelete(kv.INSTANCE, "a/b")
	if counter.ServiceCount != 0 || counter.InstanceCount != 0 {
		t.Fatal("TestGlobalCounter_OnDelete failed", counter)
	}
}
