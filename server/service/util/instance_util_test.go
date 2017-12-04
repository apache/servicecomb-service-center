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
	"github.com/ServiceComb/service-center/pkg/util"
	"github.com/ServiceComb/service-center/server/core/proto"
	"golang.org/x/net/context"
	"testing"
)

func TestGetLeaseId(t *testing.T) {

	_, err := GetLeaseId(util.SetContext(context.Background(), "cacheOnly", "1"), "", "", "")
	if err != nil {
		fmt.Printf(`GetLeaseId WithCacheOnly failed`)
		t.FailNow()
	}

	_, err = GetLeaseId(context.Background(), "", "", "")
	if err == nil {
		fmt.Printf(`GetLeaseId failed`)
		t.FailNow()
	}
}

func TestGetInstance(t *testing.T) {
	_, err := GetInstance(util.SetContext(context.Background(), "cacheOnly", "1"), "", "", "")
	if err != nil {
		fmt.Printf(`GetInstance WithCacheOnly failed`)
		t.FailNow()
	}

	_, err = GetInstance(context.Background(), "", "", "")
	if err == nil {
		fmt.Printf(`GetInstance failed`)
		t.FailNow()
	}

	_, err = GetAllInstancesOfOneService(util.SetContext(context.Background(), "cacheOnly", "1"), "", "")
	if err != nil {
		fmt.Printf(`GetAllInstancesOfOneService WithCacheOnly failed`)
		t.FailNow()
	}

	_, err = GetAllInstancesOfOneService(context.Background(), "", "")
	if err == nil {
		fmt.Printf(`GetAllInstancesOfOneService failed`)
		t.FailNow()
	}

	QueryAllProvidersIntances(context.Background(), "")

	_, err = queryServiceInstancesKvs(context.Background(), "", 0)
	if err == nil {
		fmt.Printf(`queryServiceInstancesKvs failed`)
		t.FailNow()
	}
}

func TestInstanceExist(t *testing.T) {
	_, err := InstanceExist(util.SetContext(context.Background(), "cacheOnly", "1"), "", "", "")
	if err != nil {
		fmt.Printf(`InstanceExist WithCacheOnly failed`)
		t.FailNow()
	}

	_, err = InstanceExist(context.Background(), "", "", "")
	if err == nil {
		fmt.Printf(`InstanceExist failed`)
		t.FailNow()
	}
}

func TestCheckEndPoints(t *testing.T) {
	_, _, err := CheckEndPoints(context.Background(), &proto.RegisterInstanceRequest{
		Instance: &proto.MicroServiceInstance{
			ServiceId: "a",
		},
	})
	if err == nil {
		fmt.Printf(`CheckEndPoints failed`)
		t.FailNow()
	}

	b := isContain([]string{"a"}, "a")
	if !b {
		fmt.Printf(`isContain contain failed`)
		t.FailNow()
	}

	b = isContain([]string{}, "a")
	if b {
		fmt.Printf(`isContain empty failed`)
		t.FailNow()
	}
}

func TestDeleteServiceAllInstances(t *testing.T) {
	err := DeleteServiceAllInstances(context.Background(), "")
	if err == nil {
		fmt.Printf(`DeleteServiceAllInstances failed`)
		t.FailNow()
	}
}

func TestParseEndpointValue(t *testing.T) {
	epv := ParseEndpointValue([]byte("x/y"))
	if epv.serviceId != "x" || epv.instanceId != "y" {
		fmt.Printf(`ParseEndpointValue failed`)
		t.FailNow()
	}
}
