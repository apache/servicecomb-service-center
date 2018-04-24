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
package util

import (
	"github.com/apache/incubator-servicecomb-service-center/pkg/util"
	"github.com/apache/incubator-servicecomb-service-center/server/core/proto"
	"golang.org/x/net/context"
	"testing"
)

func TestGetLeaseId(t *testing.T) {

	_, err := GetLeaseId(util.SetContext(context.Background(), "cacheOnly", "1"), "", "", "")
	if err != nil {
		t.Fatalf(`GetLeaseId WithCacheOnly failed`)
	}

	_, err = GetLeaseId(context.Background(), "", "", "")
	if err == nil {
		t.Fatalf(`GetLeaseId failed`)
	}
}

func TestGetInstance(t *testing.T) {
	_, err := GetInstance(util.SetContext(context.Background(), "cacheOnly", "1"), "", "", "")
	if err != nil {
		t.Fatalf(`GetInstance WithCacheOnly failed`)
	}

	_, err = GetInstance(context.Background(), "", "", "")
	if err == nil {
		t.Fatalf(`GetInstance failed`)
	}

	_, err = GetAllInstancesOfOneService(util.SetContext(context.Background(), "cacheOnly", "1"), "", "")
	if err != nil {
		t.Fatalf(`GetAllInstancesOfOneService WithCacheOnly failed`)
	}

	_, err = GetAllInstancesOfOneService(context.Background(), "", "")
	if err == nil {
		t.Fatalf(`GetAllInstancesOfOneService failed`)
	}

	QueryAllProvidersInstances(context.Background(), "")

	_, err = queryServiceInstancesKvs(context.Background(), "", 0)
	if err == nil {
		t.Fatalf(`queryServiceInstancesKvs failed`)
	}
}

func TestInstanceExist(t *testing.T) {
	_, err := InstanceExist(util.SetContext(context.Background(), "cacheOnly", "1"), "", "", "")
	if err != nil {
		t.Fatalf(`InstanceExist WithCacheOnly failed`)
	}

	_, err = InstanceExist(context.Background(), "", "", "")
	if err == nil {
		t.Fatalf(`InstanceExist failed`)
	}
}

func TestCheckEndPoints(t *testing.T) {
	_, err := CheckEndPoints(context.Background(), &proto.MicroServiceInstance{
		ServiceId: "a",
	})
	if err == nil {
		t.Fatalf(`CheckEndPoints failed`)
	}
}

func TestDeleteServiceAllInstances(t *testing.T) {
	err := DeleteServiceAllInstances(context.Background(), "")
	if err == nil {
		t.Fatalf(`DeleteServiceAllInstances failed`)
	}
}

func TestParseEndpointValue(t *testing.T) {
	epv := ParseEndpointIndexValue([]byte("x/y"))
	if epv.serviceId != "x" || epv.instanceId != "y" {
		t.Fatalf(`ParseEndpointIndexValue failed`)
	}
}

func TestGetInstanceCountOfOneService(t *testing.T) {
	_, err := GetInstanceCountOfOneService(context.Background(), "", "")
	if err == nil {
		t.Fatalf(`GetInstanceCountOfOneService failed`)
	}
}
