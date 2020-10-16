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
package cache

import (
	"testing"
)

type mockEventHandler struct {
	MockType Type
	Evt      KvEvent
}

func (h *mockEventHandler) Type() Type {
	return h.MockType
}
func (h *mockEventHandler) OnEvent(evt KvEvent) {
	h.Evt = evt
}

func TestAddEventHandler(t *testing.T) {
	h := &mockEventHandler{MockType: 0}
	evt := KvEvent{Revision: 1}

	// case: add
	proxy := EventProxy(0)
	if nil == proxy {
		t.Fatalf("TestAddEventHandler failed")
	}
	cfg := Configure()
	if EventProxy(0).InjectConfig(cfg) != cfg {
		t.Fatalf("TestAddEventHandler failed")
	}

	// case: normal
	AddEventHandler(h)
	proxy.OnEvent(evt)
	if h.Evt != evt {
		t.Fatalf("TestAddEventHandler failed")
	}

	AddEventHandleFunc(0, func(e KvEvent) {
		if e != evt {
			t.Fatalf("TestAddEventHandler failed")
		}
	})
	proxy.OnEvent(evt)
}
