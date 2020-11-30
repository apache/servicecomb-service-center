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

package health

import (
	"testing"
	"time"

	"github.com/apache/servicecomb-service-center/server/alarm"
	"github.com/apache/servicecomb-service-center/server/notify"
)

func TestDefaultHealthChecker_Healthy(t *testing.T) {
	notify.Center().Start()

	// normal case
	var hc DefaultHealthChecker
	if err := hc.Healthy(); err != nil {
		t.Fatal("TestDefaultHealthChecker_Healthy failed", err)
	}
	alarm.Raise(alarm.IDBackendConnectionRefuse, alarm.AdditionalContext("a"))
	time.Sleep(time.Second)
	if err := hc.Healthy(); err == nil || err.Error() != "a" {
		t.Fatal("TestDefaultHealthChecker_Healthy failed", err)
	}
	alarm.Clear(alarm.IDBackendConnectionRefuse)
	time.Sleep(time.Second)
	if err := hc.Healthy(); err != nil {
		t.Fatal("TestDefaultHealthChecker_Healthy failed", err)
	}

	// set global hc
	if GlobalHealthChecker() == &hc {
		t.Fatal("TestDefaultHealthChecker_Healthy failed")
	}

	SetGlobalHealthChecker(&hc)

	if GlobalHealthChecker() != &hc {
		t.Fatal("TestDefaultHealthChecker_Healthy failed")
	}
}
