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
	"github.com/apache/servicecomb-service-center/syncer/notify"
	"github.com/apache/servicecomb-service-center/syncer/pkg/events"
	"github.com/apache/servicecomb-service-center/syncer/plugins"
	pb "github.com/apache/servicecomb-service-center/syncer/proto"
	"github.com/apache/servicecomb-service-center/syncer/test/repository/testmock"
)

func TestNewStore(t *testing.T) {
	defer func() {
		err := recover()
		if err != nil {
			t.Log(err)
		}
	}()
	_, err := NewStore([]string{"127.0.0.1:30100"})
	if err != nil {
		t.Fatal(err)
		return
	}
}

func TestOnEvent(t *testing.T) {
	conf := config.DefaultConfig()
	conf.RepositoryPlugin = testmock.PluginName
	initPlugin(conf)
	store, err := NewStore([]string{"http://127.0.0.1:30100"})
	if err != nil {
		t.Fatal(err)
		return
	}
	ctx := context.Background()
	store.OnEvent(events.NewContextEvent("test-event", ctx))
	store.OnEvent(events.NewContextEvent(notify.EventTicker, ctx))

	data := store.LocalInfo()
	ctx = context.WithValue(ctx, notify.EventPullBySerf, &pb.NodeDataInfo{NodeName: "testNode", DataInfo: data})

	testmock.SetRegisterInstance(func(ctx context.Context, domainProject, serviceId string, instance *proto.MicroServiceInstance) (s string, e error) {
		return "", errors.New("test error")
	})
	store.OnEvent(events.NewContextEvent(notify.EventPullBySerf, ctx))

	testmock.SetRegisterInstance(nil)
	store.OnEvent(events.NewContextEvent(notify.EventPullBySerf, ctx))

	store.OnEvent(events.NewContextEvent(notify.EventPullBySerf, ctx))
	testmock.SetHeartbeat(func(ctx context.Context, domainProject, serviceId, instanceId string) error {
		return errors.New("test error")
	})
	store.OnEvent(events.NewContextEvent(notify.EventPullBySerf, ctx))
}

func TestOnEventWrongData(t *testing.T) {
	initPlugin(config.DefaultConfig())
	store, err := NewStore([]string{"127.0.0.2:30100"})
	if err != nil {
		t.Fatal(err)
		return
	}
	ctx := context.Background()
	store.OnEvent(events.NewContextEvent(notify.EventTicker, ctx))
	store.OnEvent(events.NewContextEvent(notify.EventPullBySerf, ctx))
}

func initPlugin(conf *config.Config) {
	plugins.SetPluginConfig(plugins.PluginStorage.String(), conf.StoragePlugin)
	plugins.SetPluginConfig(plugins.PluginRepository.String(), conf.RepositoryPlugin)
}
