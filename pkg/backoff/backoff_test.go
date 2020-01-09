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
package backoff

import (
	"testing"
	"time"
)

func TestPowerBackoff_Delay(t *testing.T) {
	i, m := time.Second, 30*time.Second
	b := &PowerBackoff{
		MaxDelay:  30 * time.Second,
		InitDelay: 1 * time.Second,
		Factor:    1.6,
	}
	r := b.Delay(-1)
	if r != i {
		t.Fatalf("TestPowerBackoff_Delay -1 failed, result is %s", r)
	}
	r = b.Delay(0)
	if r != i {
		t.Fatalf("TestPowerBackoff_Delay 0 failed, result is %s", r)
	}
	r = b.Delay(1)
	if r != 1600*time.Millisecond {
		t.Fatalf("TestPowerBackoff_Delay 1 failed, result is %s", r)
	}
	r = b.Delay(4)
	if r != 6553600*time.Microsecond {
		t.Fatalf("TestPowerBackoff_Delay 4 failed, result is %s", r)
	}
	r = b.Delay(8)
	if r != m {
		t.Fatalf("TestPowerBackoff_Delay 8 failed, result is %s", r)
	}
}

func TestGetBackoff(t *testing.T) {
	if GetBackoff() == nil {
		t.Fatalf("TestGetBackoff failed")
	}
}
