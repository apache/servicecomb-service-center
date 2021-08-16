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

package counter

import (
	"testing"

	registry2 "github.com/apache/servicecomb-service-center/pkg/registry"
	"github.com/apache/servicecomb-service-center/server/core"
	"github.com/apache/servicecomb-service-center/server/core/backend"
	"github.com/apache/servicecomb-service-center/server/plugin/discovery"
)

type mockCounter struct {
	ServiceCount  int64
	InstanceCount int64
}

func (c *mockCounter) OnCreate(t discovery.Type, domainProject string) {
	switch t {
	case backend.ServiceIndex:
		c.ServiceCount++
	case backend.INSTANCE:
		c.InstanceCount++
	default:
		panic("error")
	}
}

func (c *mockCounter) OnDelete(t discovery.Type, domainProject string) {
	switch t {
	case backend.ServiceIndex:
		c.ServiceCount--
	case backend.INSTANCE:
		c.InstanceCount--
	default:
		panic("error")
	}
}

func TestNewServiceIndexEventHandler(t *testing.T) {

	var counter = mockCounter{}
	RegisterCounter(&counter)
	h := NewServiceIndexEventHandler()

	cases := []discovery.KvEvent{
		{
			Type: registry2.EVT_INIT,
			KV: &discovery.KeyValue{
				Key: []byte(core.GenerateServiceIndexKey(&registry2.MicroServiceKey{
					Tenant:      core.RegistryDomainProject,
					Project:     "",
					AppId:       core.RegistryAppID,
					ServiceName: core.RegistryServiceName,
					Version:     "e",
					Environment: "f",
					Alias:       "g",
				})),
				Value: "1",
			},
		},
		{
			Type: registry2.EVT_UPDATE,
			KV: &discovery.KeyValue{
				Key: []byte(core.GenerateServiceIndexKey(&registry2.MicroServiceKey{
					Tenant:      core.RegistryDomainProject,
					Project:     "",
					AppId:       core.RegistryAppID,
					ServiceName: core.RegistryServiceName,
					Version:     "e",
					Environment: "f",
					Alias:       "g",
				})),
				Value: "1",
			},
		},
		{
			Type: registry2.EVT_DELETE,
			KV: &discovery.KeyValue{
				Key: []byte(core.GenerateServiceIndexKey(&registry2.MicroServiceKey{
					Tenant:      core.RegistryDomainProject,
					Project:     "",
					AppId:       core.RegistryAppID,
					ServiceName: core.RegistryServiceName,
					Version:     "e",
					Environment: "f",
					Alias:       "g",
				})),
				Value: "1",
			},
		},
		{
			Type: registry2.EVT_CREATE,
			KV: &discovery.KeyValue{
				Key: []byte(core.GenerateServiceIndexKey(&registry2.MicroServiceKey{
					Tenant:      core.RegistryDomainProject,
					Project:     "",
					AppId:       core.RegistryAppID,
					ServiceName: core.RegistryServiceName,
					Version:     "e",
					Environment: "f",
					Alias:       "g",
				})),
				Value: "1",
			},
		},
		{
			Type: registry2.EVT_INIT,
			KV: &discovery.KeyValue{
				Key: []byte(core.GenerateServiceIndexKey(&registry2.MicroServiceKey{
					Tenant:      "a/b",
					Project:     "",
					AppId:       "c",
					ServiceName: "d",
					Version:     "e",
					Environment: "f",
					Alias:       "g",
				})),
				Value: "1",
			},
		},
		{
			Type: registry2.EVT_DELETE,
			KV: &discovery.KeyValue{
				Key: []byte(core.GenerateServiceIndexKey(&registry2.MicroServiceKey{
					Tenant:      "a/b",
					Project:     "",
					AppId:       "c",
					ServiceName: "d",
					Version:     "e",
					Environment: "f",
					Alias:       "g",
				})),
				Value: "1",
			},
		},
		{
			Type: registry2.EVT_UPDATE,
			KV: &discovery.KeyValue{
				Key: []byte(core.GenerateServiceIndexKey(&registry2.MicroServiceKey{
					Tenant:      "a/b",
					Project:     "",
					AppId:       "c",
					ServiceName: "d",
					Version:     "e",
					Environment: "f",
					Alias:       "g",
				})),
				Value: "1",
			},
		},
		{
			Type: registry2.EVT_CREATE,
			KV: &discovery.KeyValue{
				Key: []byte(core.GenerateServiceIndexKey(&registry2.MicroServiceKey{
					Tenant:      "a/b",
					Project:     "",
					AppId:       "c",
					ServiceName: "d",
					Version:     "e",
					Environment: "f",
					Alias:       "g",
				})),
				Value: "1",
			},
		},
	}

	for _, evt := range cases {
		h.OnEvent(evt)
	}
	if counter.ServiceCount != 1 || counter.InstanceCount != 0 {
		t.Fatal("TestNewServiceIndexEventHandler failed", counter)
	}
}

func TestNewInstanceEventHandler(t *testing.T) {
	var counter = mockCounter{}
	RegisterCounter(&counter)
	h := NewInstanceEventHandler()
	SharedServiceIds.Put(core.RegistryDomainProject+core.SPLIT+"2", struct{}{})
	cases := []discovery.KvEvent{
		{
			Type: registry2.EVT_INIT,
			KV: &discovery.KeyValue{
				Key:   []byte(core.GenerateInstanceKey(core.RegistryDomainProject, "2", "1")),
				Value: nil,
			},
		},
		{
			Type: registry2.EVT_UPDATE,
			KV: &discovery.KeyValue{
				Key:   []byte(core.GenerateInstanceKey(core.RegistryDomainProject, "2", "1")),
				Value: nil,
			},
		},
		{
			Type: registry2.EVT_CREATE,
			KV: &discovery.KeyValue{
				Key:   []byte(core.GenerateInstanceKey(core.RegistryDomainProject, "2", "1")),
				Value: nil,
			},
		},
		{
			Type: registry2.EVT_DELETE,
			KV: &discovery.KeyValue{
				Key:   []byte(core.GenerateInstanceKey(core.RegistryDomainProject, "2", "1")),
				Value: nil,
			},
		},
		{
			Type: registry2.EVT_INIT,
			KV: &discovery.KeyValue{
				Key:   []byte(core.GenerateInstanceKey("a/b", "1", "1")),
				Value: nil,
			},
		},
		{
			Type: registry2.EVT_DELETE,
			KV: &discovery.KeyValue{
				Key:   []byte(core.GenerateInstanceKey("a/b", "1", "1")),
				Value: nil,
			},
		},
		{
			Type: registry2.EVT_UPDATE,
			KV: &discovery.KeyValue{
				Key:   []byte(core.GenerateInstanceKey("a/b", "1", "1")),
				Value: nil,
			},
		},
		{
			Type: registry2.EVT_CREATE,
			KV: &discovery.KeyValue{
				Key:   []byte(core.GenerateInstanceKey("a/b", "1", "1")),
				Value: nil,
			},
		},
	}

	for _, evt := range cases {
		h.OnEvent(evt)
	}
	if counter.InstanceCount != 1 || counter.ServiceCount != 0 {
		t.Fatal("TestNewServiceIndexEventHandler failed", counter)
	}
}
