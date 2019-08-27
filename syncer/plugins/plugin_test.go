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

	pb "github.com/apache/servicecomb-service-center/syncer/proto"
)

type mockPlugin struct{}

func newAdaptor() PluginInstance { return &mockAdaptor{} }

type mockAdaptor struct{}

func (*mockAdaptor) New(opts ...SCConfigOption) (Servicecenter, error) {
	return &mockRepository{}, nil
}

type mockRepository struct{}

func (r *mockRepository) GetAll(ctx context.Context) (data *pb.SyncData, err error) { return }

func (r *mockRepository) CreateService(ctx context.Context, domainProject string, service *pb.SyncService) (str string, err error) {
	return
}

func (r *mockRepository) DeleteService(ctx context.Context, domainProject, serviceId string) (err error) {
	return
}

func (r *mockRepository) ServiceExistence(ctx context.Context, domainProject string, service *pb.SyncService) (str string, err error) {
	return
}

func (r *mockRepository) RegisterInstance(ctx context.Context, domainProject, serviceId string, instance *pb.SyncInstance) (str string, err error) {
	return
}

func (r *mockRepository) UnregisterInstance(ctx context.Context, domainProject, serviceId, instanceId string) (err error) {
	return
}

func (r *mockRepository) Heartbeat(ctx context.Context, domainProject, serviceId, instanceId string) (err error) {
	return
}

func TestManager_New(t *testing.T) {
	pm := Plugins()

	notfound := PluginType(999)
	t.Log(notfound.String())

	p := pm.Get(notfound, BUILDIN)
	if p != nil {
		t.Fatalf("get %s %s failed", notfound, BUILDIN)
	}

	instanceNil := pm.Instance(PluginServicecenter)
	if instanceNil != pm.Instance(PluginServicecenter) {
		t.Fatalf("instance storage plugin: %s failed", PluginServicecenter)
	}

	getNil := pm.Get(PluginServicecenter, BUILDIN)
	if getNil != nil {
		t.Fatalf("get %s %s failed", PluginServicecenter, BUILDIN)
	}

	RegisterPlugin(&Plugin{Kind: PluginServicecenter, Name: "mock", New: newAdaptor})
	SetPluginConfig(PluginServicecenter.String(), "mock")

	repositoryInstance := pm.Instance(PluginServicecenter)
	if repositoryInstance != pm.Instance(PluginServicecenter) {
		t.Fatalf("instance storage plugin: %s failed", PluginServicecenter)
	}
	pm.Servicecenter()

	RegisterPlugin(&Plugin{Kind: notfound, Name: "mock", New: func() PluginInstance { return &mockPlugin{} }})

	LoadPlugins()

	DynamicPluginFunc(notfound, "mock")
}
