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
package serf

import (
	"fmt"
	"strconv"

	"github.com/apache/servicecomb-service-center/syncer/pkg/utils"
	"github.com/hashicorp/memberlist"
	"github.com/hashicorp/serf/cmd/serf/command/agent"
	"github.com/hashicorp/serf/serf"
)

const (
	DefaultBindPort    = 30190
	DefaultRPCPort     = 30191
	DefaultClusterPort = 30192
	ModeSingle         = "single"
	ModeCluster        = "cluster"
	retryMaxAttempts   = 3
	groupExpect        = 3
	tagKeyClusterName  = "syncer-cluster-name"
	TagKeyClusterPort  = "syncer-cluster-port"
)

// DefaultConfig default config
func DefaultConfig() *Config {
	agentConf := agent.DefaultConfig()
	agentConf.BindAddr = fmt.Sprintf("0.0.0.0:%d", DefaultBindPort)
	agentConf.RPCAddr = fmt.Sprintf("0.0.0.0:%d", DefaultRPCPort)
	return &Config{
		Mode:        ModeSingle,
		Config:      agentConf,
		ClusterPort: DefaultClusterPort,
	}
}

// Config struct
type Config struct {
	// config from serf agent
	*agent.Config
	Mode string `json:"mode"`

	// name to group members into cluster
	ClusterName string `json:"cluster_name"`

	// port to communicate between cluster members
	ClusterPort int `yaml:"cluster_port"`
	RPCPort     int `yaml:"-"`
}

// readConfigFile reads configuration from config file
func (c *Config) readConfigFile(filepath string) error {
	if filepath != "" {
		// todo:
	}
	return nil
}

// convertToSerf convert Config to serf.Config
func (c *Config) convertToSerf() (*serf.Config, error) {
	serfConf := serf.DefaultConfig()

	bindIP, bindPort, err := utils.SplitHostPort(c.BindAddr, DefaultBindPort)
	if err != nil {
		return nil, fmt.Errorf("invalid bind address: %s", err)
	}

	switch c.Profile {
	case "lan":
		serfConf.MemberlistConfig = memberlist.DefaultLANConfig()
	case "wan":
		serfConf.MemberlistConfig = memberlist.DefaultWANConfig()
	case "local":
		serfConf.MemberlistConfig = memberlist.DefaultLocalConfig()
	default:
		serfConf.MemberlistConfig = memberlist.DefaultLANConfig()
	}

	serfConf.MemberlistConfig.BindAddr = bindIP
	serfConf.MemberlistConfig.BindPort = bindPort
	serfConf.NodeName = c.NodeName

	if c.ClusterName != "" {
		serfConf.Tags = map[string]string{tagKeyClusterName: c.ClusterName}
		serfConf.Tags[TagKeyClusterPort] = strconv.Itoa(c.ClusterPort)
	}

	if c.Mode == ModeCluster && c.RetryMaxAttempts <= 0 {
		c.RetryMaxAttempts = retryMaxAttempts
	}
	return serfConf, nil
}
