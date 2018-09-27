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

package registry

import (
	"github.com/apache/incubator-servicecomb-service-center/pkg/log"
	"github.com/apache/incubator-servicecomb-service-center/server/core"
	"github.com/astaxie/beego"
	"strings"
	"sync"
	"time"
)

var (
	defaultRegistryConfig Config
	configOnce            sync.Once
)

type Config struct {
	SslEnabled       bool
	EmbedMode        string
	ManagerAddress   string
	ClusterName      string
	ClusterAddresses string
	DialTimeout      time.Duration
	RequestTimeOut   time.Duration
	AutoSyncInterval time.Duration
}

func (c *Config) Clusters() map[string]string {
	clusters := make(map[string]string)
	kvs := strings.Split(c.ClusterAddresses, ",")
	for _, cluster := range kvs {
		// sc-0=http(s)://host:port
		names := strings.Split(cluster, "=")
		if len(names) != 2 {
			continue
		}
		clusters[names[0]] = names[1]
	}
	if len(clusters) == 0 {
		clusters[c.ClusterName] = c.ClusterAddresses
	}
	return clusters
}

func (c *Config) ClusterAddress() string {
	return c.Clusters()[c.ClusterName]
}

func Configuration() *Config {
	configOnce.Do(func() {
		var err error

		defaultRegistryConfig.ClusterName = beego.AppConfig.String("manager_name")
		defaultRegistryConfig.ManagerAddress = beego.AppConfig.String("manager_addr")
		defaultRegistryConfig.ClusterAddresses = beego.AppConfig.DefaultString("manager_cluster", "http://127.0.0.1:2379")
		defaultRegistryConfig.DialTimeout, err = time.ParseDuration(beego.AppConfig.DefaultString("registry_timeout", "30s"))
		if err != nil {
			log.Errorf(err, "connect_timeout is invalid, use default time %s", defaultDialTimeout)
			defaultRegistryConfig.DialTimeout = defaultDialTimeout
		}
		defaultRegistryConfig.RequestTimeOut, err = time.ParseDuration(beego.AppConfig.DefaultString("registry_timeout", "30s"))
		if err != nil {
			log.Errorf(err, "registry_timeout is invalid, use default time %s", defaultRequestTimeout)
			defaultRegistryConfig.RequestTimeOut = defaultRequestTimeout
		}
		defaultRegistryConfig.SslEnabled = core.ServerInfo.Config.SslEnabled &&
			strings.Index(strings.ToLower(defaultRegistryConfig.ClusterAddresses), "https://") >= 0
		defaultRegistryConfig.AutoSyncInterval, err = time.ParseDuration(core.ServerInfo.Config.AutoSyncInterval)
		if err != nil {
			log.Errorf(err, "auto_sync_interval is invalid")
		}
	})
	return &defaultRegistryConfig
}
