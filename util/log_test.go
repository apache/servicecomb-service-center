//Copyright 2017 Huawei Technologies Co., Ltd
//
//Licensed under the Apache License, Version 2.0 (the "License");
//you may not use this file except in compliance with the License.
//You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
//Unless required by applicable law or agreed to in writing, software
//distributed under the License is distributed on an "AS IS" BASIS,
//WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//See the License for the specific language governing permissions and
//limitations under the License.
package util

import (
	"fmt"
	"testing"
)

func TestLogger(t *testing.T) {
	SetCustomLoggerName("Not Exist", "testDefaultLOGGER")
	l := Logger()
	if l != LOGGER {
		fmt.Println("should equal to LOGGER")
		t.FailNow()
	}
	SetCustomLoggerName("TestLogger", "testFuncName")
	l = Logger()
	if l == LOGGER || l == nil {
		fmt.Println("should create a new instance for 'TestLogger'")
		t.FailNow()
	}
	s := Logger()
	if l != s {
		fmt.Println("should be the same logger")
		t.FailNow()
	}
	SetCustomLoggerName("github.com/ServiceComb/service-center/util", "testPkgPath")
	l = Logger()
	if l == LOGGER || l == nil {
		fmt.Println("should create a new instance for 'util'")
		t.FailNow()
	}
	// l.Infof("OK")
}
