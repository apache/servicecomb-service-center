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

package client

import (
	"github.com/apache/servicecomb-service-center/pkg/cluster"
	"strings"
	"time"
)

type Config struct {
	SslEnabled       bool             `json:"-"`
	ManagerAddress   string           `json:"manageAddress,omitempty"`
	ClusterName      string           `json:"manageName,omitempty"`
	ClusterAddresses string           `json:"manageClusters,omitempty"` // the raw string of cluster configuration
	Clusters         cluster.Clusters `json:"-"`                        // parsed from ClusterAddresses
	DialTimeout      time.Duration    `json:"connectTimeout"`
	RequestTimeOut   time.Duration    `json:"registryTimeout"`
	AutoSyncInterval time.Duration    `json:"autoSyncInterval"`
}

//InitClusterInfo re-org address info with node name
func (c *Config) InitClusterInfo() {
	c.Clusters = make(cluster.Clusters)
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

// RegistryAddresses return the address of current SC's cache cache
func (c *Config) RegistryAddresses() []string {
	return c.Clusters[c.ClusterName]
}
