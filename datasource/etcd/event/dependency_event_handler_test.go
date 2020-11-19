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

package event

import "testing"

func TestDependencyEventHandler_Handle(t *testing.T) {
	h := &DependencyEventHandler{}
	r := h.backoff(nil, 0)
	if r != 1 {
		t.Fatalf("TestDependencyEventHandler_Handle failed, %v", r)
	}
	cb := 0
	r = h.backoff(func() { cb = 1 }, 1)
	if r != 2 || cb != 1 {
		t.Fatalf("TestDependencyEventHandler_Handle failed, %v", r)
	}
}
