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

func TestRegisterType(t *testing.T) {
	id := RegisterType("a", 0)
	if id.String() != "a" || id.QueueSize() != DefaultQueueSize {
		t.Fatal("TestRegisterType failed", id.String(), id.QueueSize())
	}
	id = RegisterType("b", 1)
	if id.String() != "b" || id.QueueSize() != 1 {
		t.Fatal("TestRegisterType failed", id.String(), id.QueueSize())
	}
	id = Type(999)
	if id.String() != "Type999" || id.QueueSize() != DefaultQueueSize {
		t.Fatal("TestRegisterType failed", id.String(), id.QueueSize())
	}
	if INNER.String() != "INNER" || INNER.QueueSize() != DefaultQueueSize {
		t.Fatal("TestRegisterType failed", id.String(), id.QueueSize())
	}
}
