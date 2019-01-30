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

package backoff

import (
	"errors"
	"testing"
)

func TestDelay(t *testing.T) {
	if nil == DelayIn(0, func() error {
		t.Fatal("TestDelay failed")
		return nil
	}) {
		t.Fatal("TestDelay failed")
	}

	var times = 0
	if nil != DelayIn(2, func() error {
		times++
		return nil
	}) || times > 1 {
		t.Fatal("TestDelay failed")
	}

	times = 0
	if nil == DelayIn(2, func() error {
		times++
		return errors.New("error")
	}) || times != 2 {
		t.Fatal("TestDelay failed")
	}

	if nil != Delay(func() error {
		return nil
	}) {
		t.Fatal("TestDelay failed")
	}
}
