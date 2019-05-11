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
//package plugins_test
package plugins

import (
	"context"
	"testing"

	scpb "github.com/apache/servicecomb-service-center/server/core/proto"
	"github.com/apache/servicecomb-service-center/syncer/plugins/repository"
	pb "github.com/apache/servicecomb-service-center/syncer/proto"
)

type mockPlugin struct{}

func mewMockStorage() PluginInstance { return &mockStorage{} }

type mockStorage struct{}

func (r *mockStorage) Stop() {}

func (r *mockStorage) SaveSyncData(data *pb.SyncData) {}

func (r *mockStorage) GetSyncData() (data *pb.SyncData) { return }

func (r *mockStorage) SaveSyncMapping(nodeName string, mapping pb.SyncMapping) {}

func (r *mockStorage) GetSyncMapping(nodeName string) (mapping pb.SyncMapping) { return }

func (r *mockStorage) GetAllMapping() (mapping pb.SyncMapping) { return }

func newAdaptor() PluginInstance { return &mockAdaptor{} }

type mockAdaptor struct{}

func (*mockAdaptor) New(endpoints []string) (repository.Repository, error) {
	return &mockRepository{}, nil
}

type mockRepository struct{}

func (r *mockRepository) GetAll(ctx context.Context) (data *pb.SyncData, err error) { return }

func (r *mockRepository) CreateService(ctx context.Context, domainProject string, service *scpb.MicroService) (str string, err error) {
	return
}

func (r *mockRepository) DeleteService(ctx context.Context, domainProject, serviceId string) (err error) {
	return
}

func (r *mockRepository) ServiceExistence(ctx context.Context, domainProject string, service *scpb.MicroService) (str string, err error) {
	return
}

func (r *mockRepository) RegisterInstance(ctx context.Context, domainProject, serviceId string, instance *scpb.MicroServiceInstance) (str string, err error) {
	return
}

func (r *mockRepository) UnregisterInstance(ctx context.Context, domainProject, serviceId, instanceId string) (err error) {
	return
}

func (r *mockRepository) DiscoveryInstances(ctx context.Context, domainProject, consumerId, providerAppId, providerServiceName, providerVersionRule string) (list []*scpb.MicroServiceInstance, err error) {
	return
}

func (r *mockRepository) Heartbeat(ctx context.Context, domainProject, serviceId, instanceId string) (err error) {
	return
}

func TestManager_New(t *testing.T) {
	pm := Plugins()

	notfound := PluginType(999)
	notfound.String()

	p := pm.Get(notfound, BUILDIN)
	if p != nil {
		t.Fatalf("get %s %s failed", PluginStorage, BUILDIN)
	}

	instanceNil := pm.Instance(PluginStorage)
	if instanceNil != pm.Instance(PluginStorage) {
		t.Fatalf("instance storage plugin: %s failed", PluginRepository)
	}

	getNil := pm.Get(PluginStorage, BUILDIN)
	if getNil != nil {
		t.Fatalf("get %s %s failed", PluginStorage, BUILDIN)
	}

	RegisterPlugin(&Plugin{Kind: PluginStorage, Name: "mock", New: mewMockStorage})
	SetPluginConfig(PluginStorage.String(), "mock")

	storageInstance := pm.Instance(PluginStorage)
	if storageInstance != pm.Instance(PluginStorage) {
		t.Fatalf("instance storage plugin: %s failed", PluginStorage)
	}
	pm.Storage()

	RegisterPlugin(&Plugin{Kind: PluginRepository, Name: "mock", New: newAdaptor})
	SetPluginConfig(PluginRepository.String(), "mock")

	repositoryInstance := pm.Instance(PluginRepository)
	if repositoryInstance != pm.Instance(PluginRepository) {
		t.Fatalf("instance storage plugin: %s failed", PluginRepository)
	}
	pm.Repository()

	RegisterPlugin(&Plugin{Kind: notfound, Name: "mock", New: func() PluginInstance { return &mockPlugin{} }})

	LoadPlugins()

	DynamicPluginFunc(notfound, "mock")
}
