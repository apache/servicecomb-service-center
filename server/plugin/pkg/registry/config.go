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
	ClusterAddresses string   // the raw string of cluster configuration
	Clusters         Clusters // parsed from ClusterAddresses
	DialTimeout      time.Duration
	RequestTimeOut   time.Duration
	AutoSyncInterval time.Duration
}

func (c *Config) InitClusters() {
	c.Clusters = make(Clusters)
	kvs := strings.Split(c.ClusterAddresses, ",")
	for _, cluster := range kvs {
		// sc-0=http(s)://host1:port1|http(s)://host2:port2
		names := strings.Split(cluster, "=")
		if len(names) != 2 {
			continue
		}
		c.Clusters[names[0]] = strings.Split(names[1], "|")
	}
	if len(c.Clusters) == 0 {
		c.Clusters[c.ClusterName] = []string{c.ClusterAddresses}
	}
}

// ClusterAddress return the address of current SC's registry backend
func (c *Config) ClusterAddress() string {
	return c.Clusters[c.ClusterName][0]
}

func Configuration() *Config {
	configOnce.Do(func() {
		var err error
		defaultRegistryConfig.ClusterName = beego.AppConfig.DefaultString("manager_name", "default")
		defaultRegistryConfig.ManagerAddress = beego.AppConfig.String("manager_addr")
		defaultRegistryConfig.ClusterAddresses = beego.AppConfig.DefaultString("manager_cluster", "http://127.0.0.1:2379")
		defaultRegistryConfig.DialTimeout, err = time.ParseDuration(beego.AppConfig.DefaultString("connect_timeout", "10s"))
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
		defaultRegistryConfig.InitClusters()
	})
	return &defaultRegistryConfig
}
