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
	"context"
	"os"
	"testing"

	"github.com/apache/servicecomb-service-center/pkg/registry"
	"github.com/apache/servicecomb-service-center/server/core/proto"

	"github.com/astaxie/beego"
	"github.com/stretchr/testify/assert"
)

func TestPrepareSelfRegistration(t *testing.T) {
	beego.BConfig.RunMode = "dev"
	prepareSelfRegistration()
	if Service == nil || Service.Environment != "development" {
		t.Fatalf("TestPrepareSelfRegistration faild, %v", Service)
	}

	beego.BConfig.RunMode = "prod"
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
}

func TestSetSharedMode(t *testing.T) {
	SetSharedMode()
	if IsShared(&registry.MicroServiceKey{}) {
		t.Fatalf("TestSetSharedMode failed")
	}
	if IsShared(&registry.MicroServiceKey{Tenant: "default"}) {
		t.Fatalf("TestSetSharedMode failed")
	}
	if IsShared(&registry.MicroServiceKey{Tenant: "default/default"}) {
		t.Fatalf("TestSetSharedMode failed")
	}
	if IsShared(&registry.MicroServiceKey{Tenant: "default/default", AppId: "default"}) {
		t.Fatalf("TestSetSharedMode failed")
	}

	os.Setenv("CSE_SHARED_SERVICES", "shared")
	SetSharedMode()
	if IsShared(&registry.MicroServiceKey{Tenant: "default/default", AppId: "default", ServiceName: "no-shared"}) {
		t.Fatalf("TestSetSharedMode failed")
	}
	if !IsShared(&registry.MicroServiceKey{Tenant: "default/default", AppId: "default", ServiceName: "shared"}) {
		t.Fatalf("TestSetSharedMode failed")
	}
	if !IsShared(&registry.MicroServiceKey{Tenant: "default/default", AppId: "default", Alias: "shared"}) {
		t.Fatalf("TestSetSharedMode failed")
	}
}

func TestRemoveSharedServices(t *testing.T) {
	testSharedServiceName := "TestRemoveSharedServices"
	err := os.Setenv("CSE_SHARED_SERVICES", testSharedServiceName)
	SetSharedMode()
	assert.NoError(t, err)
	sharedSvc := &registry.MicroService{
		AppId:       RegistryAppID,
		ServiceName: testSharedServiceName,
	}
	noneSharedSvc := &registry.MicroService{
		AppId:       RegistryAppID,
		ServiceName: "a",
	}

	services := []*registry.MicroService{sharedSvc, noneSharedSvc, sharedSvc, noneSharedSvc}
	assert.True(t, hasSharedService(services))
	t.Run("withShared: true, should not remove shared services", func(t *testing.T) {
		s := RemoveSharedServices(true, RegistryDomainProject, services)
		assert.Equal(t, 4, len(s))
	})
	t.Run("withShared: false, but not default domain project, "+
		"should not remove shared services", func(t *testing.T) {
		s := RemoveSharedServices(false, "a/a", services)
		assert.Equal(t, 4, len(s))
	})
	t.Run("withShared: false, default domain project, "+
		"should remove shared services", func(t *testing.T) {
		s := RemoveSharedServices(false, RegistryDomainProject, services)
		assert.Equal(t, 2, len(s))
		assert.False(t, hasSharedService(s))
	})
	t.Run("remove shared services", func(t *testing.T) {
		t.Run("is shared service: [no, no, no]", func(t *testing.T) {
			services = []*registry.MicroService{noneSharedSvc, noneSharedSvc, noneSharedSvc}
			s := RemoveSharedServices(false, RegistryDomainProject, services)
			assert.Equal(t, 3, len(s))
		})
		t.Run("is shared service: [yes]", func(t *testing.T) {
			services = []*registry.MicroService{sharedSvc}
			s := RemoveSharedServices(false, RegistryDomainProject, services)
			assert.Equal(t, 0, len(s))
			assert.False(t, hasSharedService(s))
		})
		t.Run("is shared service: [yes, yes, yes]", func(t *testing.T) {
			services = []*registry.MicroService{sharedSvc, sharedSvc, sharedSvc}
			s := RemoveSharedServices(false, RegistryDomainProject, services)
			assert.Equal(t, 0, len(s))
		})
		t.Run("is shared service: [yes, yes, yes, no, no]", func(t *testing.T) {
			services = []*registry.MicroService{sharedSvc, sharedSvc, sharedSvc, noneSharedSvc, noneSharedSvc}
			s := RemoveSharedServices(false, RegistryDomainProject, services)
			assert.Equal(t, 2, len(s))
			assert.False(t, hasSharedService(s))
		})
		t.Run("is shared service: [no, no, yes, yes]", func(t *testing.T) {
			services = []*registry.MicroService{noneSharedSvc, noneSharedSvc, sharedSvc, sharedSvc}
			s := RemoveSharedServices(false, RegistryDomainProject, services)
			assert.Equal(t, 2, len(s))
			assert.False(t, hasSharedService(s))
		})
		t.Run("is shared service: [yes, no, no, yes, no, yes, yes]", func(t *testing.T) {
			services = []*registry.MicroService{sharedSvc, noneSharedSvc, noneSharedSvc, sharedSvc, noneSharedSvc, sharedSvc, sharedSvc}
			s := RemoveSharedServices(false, RegistryDomainProject, services)
			assert.Equal(t, 3, len(s))
			assert.False(t, hasSharedService(s))
		})
	})
}

func hasSharedService(services []*registry.MicroService) bool {
	for _, s := range services {
		if IsShared(proto.MicroServiceToKey(RegistryDomainProject, s)) {
			return true
		}
	}
	return false
}
