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
package config

import (
	"fmt"
	"os"

	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/syncer/pkg/utils"
	"github.com/apache/servicecomb-service-center/syncer/plugins/servicecenter"
	"github.com/apache/servicecomb-service-center/syncer/serf"
)

var (
	DefaultMode           = "service-center"
	DefaultDCPort         = 30100
	DefaultTickerInterval = 30
)

// Config is the configuration that can be set for Syncer. Some of these
// configurations are exposed as command-line flags.
type Config struct {
	// Wraps the serf config
	*serf.Config
	LogFile string `yaml:"log_file"`

	// Mode is the type of servicecenter, currently supports "service-center"
	Mode string `yaml:"mode"`

	// SCAddr servicecenter address, which is the service registry address.
	// Cluster mode is supported, and multiple addresses are separated by an English ",".
	SCAddr string `yaml:"dc_addr"`

	// JoinAddr The management address of one gossip pool member.
	JoinAddr          string `yaml:"join_addr"`
	TickerInterval    int    `yaml:"ticker_interval"`
	Profile           string `yaml:"profile"`
	EnableCompression bool   `yaml:"enable_compression"`
	AutoSync          bool   `yaml:"auto_sync"`
	ServicecenterPlugin  string `json:"servicecenter_plugin"`
}

// DefaultConfig returns the default config
func DefaultConfig() *Config {
	serfConf := serf.DefaultConfig()
	hostname, err := os.Hostname()
	if err != nil {
		log.Errorf(err, "Error determining hostname: %s", err)
		return nil
	}
	serfConf.NodeName = hostname
	return &Config{
		Mode:             DefaultMode,
		SCAddr:           fmt.Sprintf("127.0.0.1:%d", DefaultDCPort),
		TickerInterval:   DefaultTickerInterval,
		Config:           serfConf,
		ServicecenterPlugin: servicecenter.PluginName,
	}
}

// Verification Provide config verification
func (c *Config) Verification() error {
	ip, port, err := utils.SplitHostPort(c.BindAddr, serf.DefaultBindPort)
	if err != nil {
		return err
	}
	if ip == "127.0.0.1" {
		c.BindAddr = fmt.Sprintf("0.0.0.0:%d", port)
	}

	ip, port, err = utils.SplitHostPort(c.RPCAddr, serf.DefaultRPCPort)
	if err != nil {
		return err
	}
	c.RPCPort = port
	if ip == "127.0.0.1" {
		c.RPCAddr = fmt.Sprintf("0.0.0.0:%d", c.RPCPort)
	}
	return nil
}
