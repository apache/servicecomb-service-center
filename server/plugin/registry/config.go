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
	"strings"
	"sync"
	"time"

	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/types"
	"github.com/apache/servicecomb-service-center/server/core"
	"github.com/astaxie/beego"
)

var (
	defaultRegistryConfig Config
	configOnce            sync.Once
)

type Config struct {
	SslEnabled       bool           `json:"-"`
	ManagerAddress   string         `json:"manageAddress,omitempty"`
	ClusterName      string         `json:"manageName,omitempty"`
	ClusterAddresses string         `json:"manageClusters,omitempty"` // the raw string of cluster configuration
	Clusters         types.Clusters `json:"-"`                        // parsed from ClusterAddresses
	ClusterCheck     bool           `json:"checkCluster,omitempty"`
	DialTimeout      time.Duration  `json:"connectTimeout"`
	RequestTimeOut   time.Duration  `json:"registryTimeout"`
	AutoSyncInterval time.Duration  `json:"autoSyncInterval"`
}

//InitClusterInfo re-org address info with node name
func (c *Config) InitClusterInfo() {
	c.Clusters = make(types.Clusters)
	// sc-0=http(s)://host1:port1,http(s)://host2:port2,sc-1=http(s)://host3:port3
	kvs := strings.Split(c.ClusterAddresses, "=")
	if l := len(kvs); l >= 2 {
		var (
			names []string
			addrs [][]string
		)
		for i := 0; i < l; i++ {
			ss := strings.Split(kvs[i], ",")
			sl := len(ss)
			if i != l-1 {
				names = append(names, ss[sl-1])
			}
			if i != 0 {
				if sl > 1 && i != l-1 {
					addrs = append(addrs, ss[:sl-1])
				} else {
					addrs = append(addrs, ss)
				}
			}
		}
		for i, name := range names {
			c.Clusters[name] = addrs[i]
		}
		if len(c.ManagerAddress) > 0 {
			c.Clusters[c.ClusterName] = strings.Split(c.ManagerAddress, ",")
		}
	}
	if len(c.Clusters) == 0 {
		c.Clusters[c.ClusterName] = strings.Split(c.ClusterAddresses, ",")
	}
}

// RegistryAddresses return the address of current SC's registry backend
func (c *Config) RegistryAddresses() []string {
	return c.Clusters[c.ClusterName]
}

func Configuration() *Config {
	configOnce.Do(func() {
		var err error
		defaultRegistryConfig.ClusterName = beego.AppConfig.DefaultString("manager_name", DefaultClusterName)
		defaultRegistryConfig.ManagerAddress = beego.AppConfig.String("manager_addr")
		defaultRegistryConfig.ClusterAddresses = beego.AppConfig.DefaultString("manager_cluster", "http://127.0.0.1:2379")
		defaultRegistryConfig.ClusterCheck = beego.AppConfig.DefaultBool("registry_check_cluster", true)
		defaultRegistryConfig.InitClusterInfo()

		registryAddresses := strings.Join(defaultRegistryConfig.RegistryAddresses(), ",")
		defaultRegistryConfig.SslEnabled = core.ServerInfo.Config.SslEnabled &&
			strings.Contains(strings.ToLower(registryAddresses), "https://")

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
		defaultRegistryConfig.AutoSyncInterval, err = time.ParseDuration(core.ServerInfo.Config.AutoSyncInterval)
		if err != nil {
			log.Errorf(err, "auto_sync_interval is invalid")
		}
	})
	return &defaultRegistryConfig
}
