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
package servicecenter_test

import (
	"context"
	"errors"
	"testing"

	"github.com/apache/servicecomb-service-center/server/core/proto"
	"github.com/apache/servicecomb-service-center/syncer/config"
	"github.com/apache/servicecomb-service-center/syncer/pkg/mock/mockplugin"
	"github.com/apache/servicecomb-service-center/syncer/pkg/mock/mocksotrage"
	"github.com/apache/servicecomb-service-center/syncer/plugins"
	pb "github.com/apache/servicecomb-service-center/syncer/proto"
	"github.com/apache/servicecomb-service-center/syncer/servicecenter"
)

func TestNewServicecenter(t *testing.T) {
	defer func() {
		err := recover()
		if err != nil {
			t.Log(err)
		}
	}()
	_, err := servicecenter.NewServicecenter([]string{"127.0.0.1:30100"})
	if err != nil {
		t.Log(err)
	}

	_, err = servicecenter.NewServicecenter([]string{"127.0.0.1:30100"})
	if err != nil {
		t.Fatal(err)
		return
	}
}

func TestOnEvent(t *testing.T) {
	conf := config.DefaultConfig()
	conf.ServicecenterPlugin = mockplugin.PluginName
	initPlugin(conf)
	dc, err := servicecenter.NewServicecenter([]string{"http://127.0.0.1:30100"})
	if err != nil {
		t.Fatal(err)
		return
	}
	dc.SetStorage(mocksotrage.New())

	mockplugin.SetGetAll(func(ctx context.Context) (data *pb.SyncData, e error) {
		return nil, errors.New("test error")
	})

	dc.FlushData()
	data := dc.Discovery()
	if err != nil {
		t.Log(err)
	}

	mockplugin.SetGetAll(nil)

	dc.FlushData()
	data = dc.Discovery()
	if err != nil {
		t.Fatal(err)
		return
	}

	clusterName := "test_node"
	dc.Registry(clusterName, data)

	mockplugin.SetGetAll(mockplugin.NewGetAll)
	dc.FlushData()
	newData := dc.Discovery()
	if err != nil {
		t.Fatal(err)
		return
	}

	dc.Registry(clusterName, newData)

	mockplugin.SetRegisterInstance(func(ctx context.Context, domainProject, serviceId string, instance *proto.MicroServiceInstance) (s string, e error) {
		return "", errors.New("test error")
	})

	dc.Registry(clusterName, data)

	mockplugin.SetRegisterInstance(nil)

	dc.Registry(clusterName, data)

	dc.Registry(clusterName, data)

	mockplugin.SetHeartbeat(func(ctx context.Context, domainProject, serviceId, instanceId string) error {
		return errors.New("test error")
	})

	dc.Registry(clusterName, data)
}

func initPlugin(conf *config.Config) {
	plugins.SetPluginConfig(plugins.PluginServicecenter.String(), conf.ServicecenterPlugin)
}
