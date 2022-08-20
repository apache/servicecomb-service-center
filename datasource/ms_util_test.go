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

package datasource_test

import (
	"testing"

	"github.com/apache/servicecomb-service-center/datasource"

	"github.com/go-chassis/cari/discovery"
	"github.com/stretchr/testify/assert"
)

func TestSetDefault(t *testing.T) {
	service := &discovery.MicroService{}
	datasource.SetServiceDefaultValue(service)
	if len(service.Level) == 0 ||
		len(service.Status) == 0 {
		t.Fatalf(`TestSetDefault failed`)
	}
}

func TestRemoveGlobalServices(t *testing.T) {
	testGlobalServiceName := "RemoveGlobalServices"
	datasource.RegisterGlobalService(testGlobalServiceName)
	globalSvc := &discovery.MicroService{
		AppId:       datasource.RegistryAppID,
		ServiceName: testGlobalServiceName,
	}
	noneGlobalSvc := &discovery.MicroService{
		AppId:       datasource.RegistryAppID,
		ServiceName: "a",
	}
	//global Global
	services := []*discovery.MicroService{globalSvc, noneGlobalSvc, globalSvc, noneGlobalSvc}
	assert.True(t, hasGlobalService(services))
	t.Run("withShared: true, should not remove global services", func(t *testing.T) {
		s := datasource.RemoveGlobalServices(true, datasource.RegistryDomainProject, services)
		assert.Equal(t, 4, len(s))
	})
	t.Run("withShared: false, but not default domain project, "+
		"should not remove global services", func(t *testing.T) {
		s := datasource.RemoveGlobalServices(false, "a/a", services)
		assert.Equal(t, 4, len(s))
	})
	t.Run("withShared: false, default domain project, "+
		"should remove global services", func(t *testing.T) {
		s := datasource.RemoveGlobalServices(false, datasource.RegistryDomainProject, services)
		assert.Equal(t, 2, len(s))
		assert.False(t, hasGlobalService(s))
	})
	t.Run("remove global services", func(t *testing.T) {
		t.Run("is global service: [no, no, no]", func(t *testing.T) {
			services = []*discovery.MicroService{noneGlobalSvc, noneGlobalSvc, noneGlobalSvc}
			s := datasource.RemoveGlobalServices(false, datasource.RegistryDomainProject, services)
			assert.Equal(t, 3, len(s))
		})
		t.Run("is global service: [yes]", func(t *testing.T) {
			services = []*discovery.MicroService{globalSvc}
			s := datasource.RemoveGlobalServices(false, datasource.RegistryDomainProject, services)
			assert.Equal(t, 0, len(s))
			assert.False(t, hasGlobalService(s))
		})
		t.Run("is global service: [yes, yes, yes]", func(t *testing.T) {
			services = []*discovery.MicroService{globalSvc, globalSvc, globalSvc}
			s := datasource.RemoveGlobalServices(false, datasource.RegistryDomainProject, services)
			assert.Equal(t, 0, len(s))
		})
		t.Run("is global service: [yes, yes, yes, no, no]", func(t *testing.T) {
			services = []*discovery.MicroService{globalSvc, globalSvc, globalSvc, noneGlobalSvc, noneGlobalSvc}
			s := datasource.RemoveGlobalServices(false, datasource.RegistryDomainProject, services)
			assert.Equal(t, 2, len(s))
			assert.False(t, hasGlobalService(s))
		})
		t.Run("is global service: [no, no, yes, yes]", func(t *testing.T) {
			services = []*discovery.MicroService{noneGlobalSvc, noneGlobalSvc, globalSvc, globalSvc}
			s := datasource.RemoveGlobalServices(false, datasource.RegistryDomainProject, services)
			assert.Equal(t, 2, len(s))
			assert.False(t, hasGlobalService(s))
		})
		t.Run("is global service: [no, no, yes, yes, no, no]", func(t *testing.T) {
			services = []*discovery.MicroService{noneGlobalSvc, noneGlobalSvc, globalSvc, globalSvc, noneGlobalSvc, noneGlobalSvc}
			s := datasource.RemoveGlobalServices(false, datasource.RegistryDomainProject, services)
			assert.Equal(t, 4, len(s))
			assert.False(t, hasGlobalService(s))
		})
		t.Run("is global service: [yes, no, no, yes, no, yes, yes]", func(t *testing.T) {
			services = []*discovery.MicroService{globalSvc, noneGlobalSvc, noneGlobalSvc, globalSvc, noneGlobalSvc, globalSvc, globalSvc}
			s := datasource.RemoveGlobalServices(false, datasource.RegistryDomainProject, services)
			assert.Equal(t, 3, len(s))
			assert.False(t, hasGlobalService(s))
		})
	})
}

func hasGlobalService(services []*discovery.MicroService) bool {
	for _, s := range services {
		if datasource.IsGlobal(discovery.MicroServiceToKey(datasource.RegistryDomainProject, s)) {
			return true
		}
	}
	return false
}
