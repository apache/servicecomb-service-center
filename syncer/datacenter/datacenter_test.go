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
package datacenter

import (
	"context"
	"errors"
	"testing"

	"github.com/apache/servicecomb-service-center/server/core/proto"
	"github.com/apache/servicecomb-service-center/syncer/config"
	"github.com/apache/servicecomb-service-center/syncer/plugins"
	pb "github.com/apache/servicecomb-service-center/syncer/proto"
	"github.com/apache/servicecomb-service-center/syncer/test/dcmock"
)

func TestNewDataCenter(t *testing.T) {
	defer func() {
		err := recover()
		if err != nil {
			t.Log(err)
		}
	}()
	_, err := NewDataCenter([]string{"127.0.0.1:30100"})
	if err != nil {
		t.Log(err)
	}

	_, err = NewDataCenter([]string{"127.0.0.1:30100"})
	if err != nil {
		t.Fatal(err)
		return
	}
}

func TestOnEvent(t *testing.T) {
	conf := config.DefaultConfig()
	conf.DatacenterPlugin = dcmock.PluginName
	initPlugin(conf)
	dc, err := NewDataCenter([]string{"http://127.0.0.1:30100"})
	if err != nil {
		t.Fatal(err)
		return
	}

	mapping := pb.SyncMapping{}

	dcmock.SetGetAll(func(ctx context.Context) (data *pb.SyncData, e error) {
		return nil, errors.New("test error")
	})

	data, err := dc.GetSyncData(mapping)
	if err != nil {
		t.Log(err)
	}

	dcmock.SetGetAll(nil)

	data, err = dc.GetSyncData(mapping)
	if err != nil {
		t.Fatal(err)
		return
	}

	_, err = dc.SetSyncData(data, mapping)
	if err != nil {
		t.Fatal(err)
		return
	}

	dcmock.SetRegisterInstance(func(ctx context.Context, domainProject, serviceId string, instance *proto.MicroServiceInstance) (s string, e error) {
		return "", errors.New("test error")
	})

	_, err = dc.SetSyncData(data, mapping)
	if err != nil {
		t.Log(err)
	}

	dcmock.SetRegisterInstance(nil)
	_, err = dc.SetSyncData(data, mapping)
	if err != nil {
		t.Log(err)
	}

	_, err = dc.SetSyncData(data, mapping)
	if err != nil {
		t.Log(err)
	}

	dcmock.SetHeartbeat(func(ctx context.Context, domainProject, serviceId, instanceId string) error {
		return errors.New("test error")
	})
	_, err = dc.SetSyncData(data, mapping)
	if err != nil {
		t.Log(err)
	}
}

func TestOnEventWrongData(t *testing.T) {
	conf := config.DefaultConfig()
	initPlugin(conf)
	dc, err := NewDataCenter([]string{"127.0.0.2:30100"})
	if err != nil {
		t.Fatal(err)
		return
	}

	mapping := pb.SyncMapping{}
	data, err := dc.GetSyncData(mapping)
	if err != nil {
		t.Fatal(err)
		return
	}

	_, err = dc.SetSyncData(data, mapping)
	if err != nil {
		t.Fatal(err)
		return
	}
}

func initPlugin(conf *config.Config) {
	plugins.SetPluginConfig(plugins.PluginDatacenter.String(), conf.DatacenterPlugin)
}
