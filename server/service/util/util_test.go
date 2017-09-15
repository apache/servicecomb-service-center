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
	serviceUtil "github.com/ServiceComb/service-center/server/service/util"
	"testing"
)

func TestWithNoCache(t *testing.T) {
	op := serviceUtil.WithNoCache(false)
	if op == nil {
		t.FailNow()
	}
	r := op()
	if r != nil {
		t.FailNow()
	}

	op = serviceUtil.WithNoCache(true)
	if op == nil {
		t.FailNow()
	}
	r = op()
	if r == nil {
		t.FailNow()
	}
}

func TestQueryOptions(t *testing.T) {
	opts := serviceUtil.QueryOptions()
	if opts != nil {
		t.FailNow()
	}

	opts = serviceUtil.QueryOptions(serviceUtil.WithNoCache(true))
	if len(opts) != 1 {
		t.FailNow()
	}
}
