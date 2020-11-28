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
	"sync"
	"time"

	"github.com/apache/servicecomb-service-center/datasource/etcd/client"
	"github.com/apache/servicecomb-service-center/server/config"
)

var (
	defaultRegistryConfig client.Config
	configOnce            sync.Once
)

func Configuration() *client.Config {
	configOnce.Do(func() {
		defaultRegistryConfig.ClusterName = config.GetString("registry.etcd.cluster.name", client.DefaultClusterName, config.WithStandby("manager_name"))
		defaultRegistryConfig.ManagerAddress = config.GetString("registry.etcd.cluster.managerEndpoints", "", config.WithStandby("manager_addr"))
		defaultRegistryConfig.ClusterAddresses = config.GetString("registry.etcd.cluster.endpoints", "http://127.0.0.1:2379", config.WithStandby("manager_cluster"))
		defaultRegistryConfig.InitClusterInfo()

		defaultRegistryConfig.DialTimeout = config.GetDuration("registry.etcd.connect.timeout", client.DefaultDialTimeout, config.WithStandby("connect_timeout"))
		defaultRegistryConfig.RequestTimeOut = config.GetDuration("registry.etcd.request.timeout", client.DefaultRequestTimeout, config.WithStandby("registry_timeout"))
		defaultRegistryConfig.AutoSyncInterval = config.GetDuration("registry.etcd.autoSyncInterval", 30*time.Second, config.WithStandby("auto_sync_interval"))
	})
	return &defaultRegistryConfig
}

func WithTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, defaultRegistryConfig.RequestTimeOut)
}
