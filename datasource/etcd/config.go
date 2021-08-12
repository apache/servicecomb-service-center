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
	"time"

	"github.com/apache/servicecomb-service-center/server/config"
	"github.com/little-cui/etcdadpt"
)

func Configuration() etcdadpt.Config {
	cfg := etcdadpt.Config{}

	// cluster configs
	cfg.ClusterName = config.GetString("registry.etcd.cluster.name", etcdadpt.DefaultClusterName, config.WithStandby("manager_name"))
	cfg.ManagerAddress = config.GetString("registry.etcd.cluster.managerEndpoints", "", config.WithStandby("manager_addr"))
	cfg.ClusterAddresses = config.GetString("registry.etcd.cluster.endpoints", "http://127.0.0.1:2379", config.WithStandby("manager_cluster"))

	// connection configs
	cfg.DialTimeout = config.GetDuration("registry.etcd.connect.timeout", etcdadpt.DefaultDialTimeout, config.WithStandby("connect_timeout"))
	cfg.RequestTimeOut = config.GetDuration("registry.etcd.request.timeout", etcdadpt.DefaultRequestTimeout, config.WithStandby("registry_timeout"))
	cfg.AutoSyncInterval = config.GetDuration("registry.etcd.autoSyncInterval", 30*time.Second, config.WithStandby("auto_sync_interval"))

	// compaction configs
	cfg.CompactIndexDelta = config.GetInt64("registry.etcd.compact.indexDelta", 100, config.WithStandby("compact_index_delta"))
	cfg.CompactInterval = config.GetDuration("registry.etcd.compact.interval", 12*time.Hour, config.WithStandby("compact_interval"))
	return cfg
}
