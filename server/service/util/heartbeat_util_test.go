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
package util_test

import (
	"fmt"
	serviceUtil "github.com/ServiceComb/service-center/server/service/util"
	"golang.org/x/net/context"
	"testing"
)

func TestHeartbeatUtil(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.FailNow()
		}
	}()
	serviceUtil.HeartbeatUtil(context.Background(), "", "", "")
}

func TestKeepAliveLease(t *testing.T) {
	_, err := serviceUtil.KeepAliveLease(context.Background(), "", "", "", -1)
	if err == nil {
		fmt.Printf("KeepAliveLease -1 failed")
		t.FailNow()
	}

	_, err = serviceUtil.KeepAliveLease(context.Background(), "", "", "", 0)
	if err == nil {
		fmt.Printf("KeepAliveLease failed")
		t.FailNow()
	}
}
