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
package metrics

import (
	"os"
	"testing"
	"time"

	"github.com/astaxie/beego"
)

func TestInstanceName(t *testing.T) {
	beego.AppConfig.Set("rpcaddr", "a")
	beego.AppConfig.Set("rpcport", "b")
	i := InstanceName()
	if i != "a:b" {
		t.Fatalf("TestInstanceName failed")
	}
	// case: initialize only one time
	beego.AppConfig.Set("httpaddr", "c")
	beego.AppConfig.Set("httpaddr", "d")
	i = InstanceName()
	if i != "a:b" {
		t.Fatalf("TestInstanceName failed")
	}
}

func TestPeriod(t *testing.T) {
	if getPeriod() != 30*time.Second {
		t.Fatalf("TestPeriod failed")
	}
	os.Setenv("METRICS_INTERVAL", time.Millisecond.String())
	if getPeriod() != 30*time.Second {
		t.Fatalf("TestPeriod failed")
	}
	os.Setenv("METRICS_INTERVAL", "err")
	if getPeriod() != 30*time.Second {
		t.Fatalf("TestPeriod failed")
	}
	os.Setenv("METRICS_INTERVAL", time.Second.String())
	if getPeriod() != time.Second {
		t.Fatalf("TestPeriod failed")
	}
}
