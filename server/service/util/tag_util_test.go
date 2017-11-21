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
	"github.com/ServiceComb/service-center/pkg/util"
	serviceUtil "github.com/ServiceComb/service-center/server/service/util"
	"golang.org/x/net/context"
	"testing"
)

func TestAddTagIntoETCD(t *testing.T) {
	err := serviceUtil.AddTagIntoETCD(context.Background(), "", "", map[string]string{"a": "1"})
	if err == nil {
		fmt.Printf(`AddTagIntoETCD with {"a": "1"} tags failed`)
		t.FailNow()
	}
}

func TestGetTagsUtils(t *testing.T) {
	_, err := serviceUtil.GetTagsUtils(util.SetContext(context.Background(), "cacheOnly", "1"), "", "")
	if err != nil {
		fmt.Printf(`GetTagsUtils WithCacheOnly failed`)
		t.FailNow()
	}

	_, err = serviceUtil.GetTagsUtils(context.Background(), "", "")
	if err == nil {
		fmt.Printf(`GetTagsUtils failed`)
		t.FailNow()
	}
}
