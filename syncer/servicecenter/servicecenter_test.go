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
	"strconv"
	"testing"

	"github.com/apache/servicecomb-service-center/syncer/config"
	"github.com/apache/servicecomb-service-center/syncer/pkg/mock/mockplugin"
	"github.com/apache/servicecomb-service-center/syncer/pkg/mock/mocksotrage"
	"github.com/apache/servicecomb-service-center/syncer/plugins"
	pb "github.com/apache/servicecomb-service-center/syncer/proto"
	"github.com/apache/servicecomb-service-center/syncer/servicecenter"
	"github.com/stretchr/testify/assert"
)

func TestNewServicecenter(t *testing.T) {
	defer func() {
		err := recover()
		if err != nil {
			t.Log(err)
		}
	}()
	_, err := servicecenter.NewServicecenter(
		plugins.WithEndpoints([]string{"127.0.0.1:30100"}))
	if err != nil {
		t.Log(err)
	}

	_, err = servicecenter.NewServicecenter(
		plugins.WithEndpoints([]string{"127.0.0.1:30100"}))
	if err != nil {
		t.Fatal(err)
		return
	}
}

func TestOnEvent(t *testing.T) {
	conf := config.DefaultConfig()
	conf.Registry.Plugin = mockplugin.PluginName
	initPlugin(conf)
	dc, err := servicecenter.NewServicecenter(
		plugins.WithEndpoints([]string{"127.0.0.1:30100"}))
	if err != nil {
		t.Fatal(err)
		return
	}
	mockServer, err := mocksotrage.NewKVServer()
	if err != nil {
		t.Fatal(err)
		return
	}
	defer mockServer.Stop()
	dc.SetStorageEngine(mockServer.Storage())

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

	clusterName := "test_cluster"
	dc.Registry(clusterName, data)

	mockplugin.SetGetAll(mockplugin.NewGetAll)
	dc.FlushData()
	newData := dc.Discovery()
	if err != nil {
		t.Fatal(err)
		return
	}

	dc.Registry(clusterName, newData)

	mockplugin.SetRegisterInstance(func(ctx context.Context, domainProject, serviceId string, instance *pb.SyncInstance) (s string, e error) {
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

	dc.IncrementRegistry(clusterName, data)
}

func initPlugin(conf *config.Config) {
	plugins.SetPluginConfig(plugins.PluginServicecenter.String(), conf.Registry.Plugin)
}

func TestServicecenter_IncrementRegistry(t *testing.T) {
	conf := config.DefaultConfig()
	conf.Registry.Plugin = mockplugin.PluginName
	initPlugin(conf)
	s, err := servicecenter.NewServicecenter(
		plugins.WithEndpoints([]string{"127.0.0.1:30100"}))
	if err != nil {
		t.Fatal(err)
		return
	}
	mockServer, err := mocksotrage.NewKVServer()
	if err != nil {
		t.Fatal(err)
		return
	}
	defer mockServer.Stop()
	s.SetStorageEngine(mockServer.Storage())

	mockplugin.SetGetAll(func(ctx context.Context) (data *pb.SyncData, e error) {
		return nil, errors.New("test error")
	})

	t.Run("IncrementRegistry without service", func(t *testing.T) {
		data := new(pb.SyncData)
		s.IncrementRegistry("cluster1", data)
		assert.Error(t, errors.New("run IncrementRegistry error"), "IncrementRegistry error without service")
	})

	t.Run("IncrementRegistry when cluster1", func(t *testing.T) {
		data := dataCreate()
		s.IncrementRegistry("cluster1", &data)
	})
}

func dataCreate() pb.SyncData {

	status := []pb.SyncService_Status{pb.SyncService_UNKNOWN, pb.SyncService_UP, pb.SyncService_DOWN}
	syncServiceArr := []*pb.SyncService{}
	for i := 0; i < 10; i++ {
		var ss = new(pb.SyncService)
		ss.App = "serviceApp" + strconv.FormatInt(int64(i), 10)
		ss.DomainProject = "default"
		ss.Environment = "env"
		ss.Name = "service" + strconv.FormatInt(int64(i/2), 10)
		ss.PluginName = "plugin" + strconv.FormatInt(int64(i/2), 10)
		ss.ServiceId = "a59f99611a6945677a21f28c0aeb05abb" + strconv.FormatInt(int64(i/2), 10)
		ss.Status = status[i%3]
		ss.Version = "1.0.0"
		syncServiceArr = append(syncServiceArr, ss)
	}

	insStatus := []pb.SyncInstance_Status{pb.SyncInstance_UNKNOWN, pb.SyncInstance_UP, pb.SyncInstance_STARTING, pb.SyncInstance_DOWN, pb.SyncInstance_OUTOFSERVICE}
	healthCheckModes := []pb.HealthCheck_Modes{pb.HealthCheck_UNKNOWN, pb.HealthCheck_PUSH, pb.HealthCheck_PULL}
	syncInstanceArr := []*pb.SyncInstance{}
	bytes := [][]byte{{'C', 'R', 'E', 'A', 'T', 'E'}, {'D', 'E', 'L', 'E', 'T', 'E'}}
	for i := 0; i < 11; i++ {
		healthCheck := pb.HealthCheck{
			Mode:     healthCheckModes[i%3],
			Interval: 30,
			Times:    30,
		}
		var num = i % 2
		if i == 10 {
			num = 1
		}
		exception := pb.Expansion{
			Kind:  "action",
			Bytes: bytes[num],
		}
		var is = new(pb.SyncInstance)
		is.HostName = "provider_demo" + strconv.FormatInt(int64(i), 10)
		is.Endpoints = []string{"rest://127.0.0.1:8080"}
		is.InstanceId = "5e1140fc232111eb9bb600acc8c56b5b" + strconv.FormatInt(int64(i/2), 10)
		is.HealthCheck = &healthCheck
		is.PluginName = "plugin" + strconv.FormatInt(int64(i/2), 10)
		if i == 10 {
			is.ServiceId = "a59f99611a6945677a21f28c0aeb05abb" + strconv.FormatInt(int64(i), 10)
		} else {
			is.ServiceId = "a59f99611a6945677a21f28c0aeb05abb" + strconv.FormatInt(int64(i/2), 10)
		}
		is.Status = insStatus[i%5]
		is.Version = "1.0.0"
		exceptions := append(pb.Expansions{}, &exception)
		is.Expansions = exceptions
		syncInstanceArr = append(syncInstanceArr, is)
	}

	data := pb.SyncData{
		Services:  syncServiceArr,
		Instances: syncInstanceArr,
	}
	return data
}
