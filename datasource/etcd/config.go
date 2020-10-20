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

package etcd

import (
	"context"
	"github.com/apache/servicecomb-service-center/datasource/etcd/registry"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/server/core"
	"github.com/astaxie/beego"
	"strings"
	"sync"
	"time"
)

var (
	defaultRegistryConfig registry.Config
	configOnce            sync.Once
)

func Configuration() *registry.Config {
	configOnce.Do(func() {
		var err error
		defaultRegistryConfig.ClusterName = beego.AppConfig.DefaultString("manager_name", registry.DefaultClusterName)
		defaultRegistryConfig.ManagerAddress = beego.AppConfig.String("manager_addr")
		defaultRegistryConfig.ClusterAddresses = beego.AppConfig.DefaultString("manager_cluster", "http://127.0.0.1:2379")
		defaultRegistryConfig.InitClusterInfo()

		registryAddresses := strings.Join(defaultRegistryConfig.RegistryAddresses(), ",")
		defaultRegistryConfig.SslEnabled = core.ServerInfo.Config.SslEnabled &&
			strings.Contains(strings.ToLower(registryAddresses), "https://")

		defaultRegistryConfig.DialTimeout, err = time.ParseDuration(beego.AppConfig.DefaultString("connect_timeout", "10s"))
		if err != nil {
			log.Errorf(err, "connect_timeout is invalid, use default time %s", registry.DefaultDialTimeout)
			defaultRegistryConfig.DialTimeout = registry.DefaultDialTimeout
		}
		defaultRegistryConfig.RequestTimeOut, err = time.ParseDuration(beego.AppConfig.DefaultString("registry_timeout", "30s"))
		if err != nil {
			log.Errorf(err, "registry_timeout is invalid, use default time %s", registry.DefaultRequestTimeout)
			defaultRegistryConfig.RequestTimeOut = registry.DefaultRequestTimeout
		}
		defaultRegistryConfig.AutoSyncInterval, err = time.ParseDuration(core.ServerInfo.Config.AutoSyncInterval)
		if err != nil {
			log.Errorf(err, "auto_sync_interval is invalid")
		}

		core.ServerInfo.Config.Plugins.Object("discovery").
			Set("config", defaultRegistryConfig)
	})
	return &defaultRegistryConfig
}

func WithTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, defaultRegistryConfig.RequestTimeOut)
}
