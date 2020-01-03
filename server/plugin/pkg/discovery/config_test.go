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

package discovery

import (
	pb "github.com/apache/servicecomb-service-center/server/core/proto"
	"testing"
	"time"
)

type mockDeferHandler struct {
}

func (m *mockDeferHandler) OnCondition(CacheReader, []KvEvent) bool {
	return false
}
func (m *mockDeferHandler) HandleChan() <-chan KvEvent {
	return nil
}
func (m *mockDeferHandler) Reset() bool {
	return false
}

func TestConfigure(t *testing.T) {
	cfg := Configure()
	if cfg == nil {
		t.Fatalf("TestConfigure failed")
	}
	cfg.WithPrefix("/test")
	if cfg.Key != "/test" {
		t.Fatalf("TestConfigure failed")
	}
	cfg.WithTimeout(2 * time.Second)
	if cfg.Timeout != 2*time.Second {
		t.Fatalf("TestConfigure failed")
	}
	cfg.WithInitSize(1)
	if cfg.InitSize != 1 {
		t.Fatalf("TestConfigure failed")
	}
	cfg.WithPeriod(3 * time.Second)
	if cfg.Period != 3*time.Second {
		t.Fatalf("TestConfigure failed")
	}
	cfg.WithDeferHandler(&mockDeferHandler{})
	if cfg.DeferHandler == nil {
		t.Fatalf("TestConfigure failed")
	}
	i := 0
	cfg.WithEventFunc(func(evt KvEvent) {
		i++
	})
	cfg.AppendEventFunc(func(evt KvEvent) {
		i += 2
	})
	cfg.OnEvent(KvEvent{})
	if i != 3 {
		t.Fatalf("TestConfigure failed")
	}
	cfg.WithParser(pb.MapParser)
	if cfg.Parser != pb.MapParser {
		t.Fatalf("TestConfigure failed")
	}
	if cfg.String() != "{key: /test, timeout: 2s, period: 3s}" {
		t.Fatalf("TestConfigure failed")
	}
}
