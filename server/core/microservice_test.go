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

package core

import (
	"github.com/apache/incubator-servicecomb-service-center/server/core/proto"
	"github.com/apache/incubator-servicecomb-service-center/version"
	"golang.org/x/net/context"
	"os"
	"testing"
)

func TestPrepareSelfRegistration(t *testing.T) {
	version.Ver().RunMode = "dev"
	prepareSelfRegistration()
	if Service == nil || Service.Environment != "development" {
		t.Fatalf("TestPrepareSelfRegistration faild, %v", Service)
	}

	version.Ver().RunMode = "prod"
	prepareSelfRegistration()
	if Service == nil || Service.AppId != "default" || Service.ServiceName != "SERVICECENTER" ||
		Service.Environment != "production" || Service.Properties["allowCrossApp"] != "true" {
		t.Fatalf("TestPrepareSelfRegistration faild, %v", Service)
	}

	if Instance == nil || Instance.Status != "UP" {
		t.Fatalf("TestPrepareSelfRegistration faild, %v", Instance)
	}

	if IsSCInstance(context.Background()) {
		t.Fatalf("TestPrepareSelfRegistration faild")
	}

	exist := GetExistenceRequest()
	if exist == nil || exist.Environment != "production" || exist.ServiceName != "SERVICECENTER" ||
		exist.AppId != "default" {
		t.Fatalf("TestPrepareSelfRegistration faild, %v", exist)
	}

	instance := RegisterInstanceRequest("test", []string{"a"})
	if instance == nil || instance.Instance.HostName != "test" || instance.Instance.Endpoints[0] != "a" {
		t.Fatalf("TestPrepareSelfRegistration faild, %v", instance)
	}
}

func TestSetSharedMode(t *testing.T) {
	SetSharedMode()
	if IsShared(&proto.MicroServiceKey{}) {
		t.Fatalf("TestSetSharedMode failed")
	}
	if IsShared(&proto.MicroServiceKey{Tenant: "default"}) {
		t.Fatalf("TestSetSharedMode failed")
	}
	if IsShared(&proto.MicroServiceKey{Tenant: "default/default"}) {
		t.Fatalf("TestSetSharedMode failed")
	}
	if IsShared(&proto.MicroServiceKey{Tenant: "default/default", AppId: "default"}) {
		t.Fatalf("TestSetSharedMode failed")
	}

	os.Setenv("CSE_SHARED_SERVICES", "shared")
	SetSharedMode()
	if IsShared(&proto.MicroServiceKey{Tenant: "default/default", AppId: "default", ServiceName: "no-shared"}) {
		t.Fatalf("TestSetSharedMode failed")
	}
	if !IsShared(&proto.MicroServiceKey{Tenant: "default/default", AppId: "default", ServiceName: "shared"}) {
		t.Fatalf("TestSetSharedMode failed")
	}
}
