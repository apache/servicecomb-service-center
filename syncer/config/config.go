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
	"crypto/md5"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/syncer/etcd"
	"github.com/apache/servicecomb-service-center/syncer/pkg/utils"
	_ "github.com/apache/servicecomb-service-center/syncer/plugins/eureka"
	"github.com/apache/servicecomb-service-center/syncer/plugins/servicecenter"
	"github.com/apache/servicecomb-service-center/syncer/serf"
	"gopkg.in/yaml.v2"
)

var (
	DefaultDCPort         = 30100
	DefaultClusterPort    = 30192
	DefaultTickerInterval = 30
	DefaultConfigPath     = "./conf/config.yaml"
)

// Config is the configuration that can be set for Syncer. Some of these
// configurations are exposed as command-line flags.
type Config struct {
	// Wraps the serf config
	*serf.Config

	// Wraps the etcd config
	Etcd    *etcd.Config
	LogFile string `yaml:"log_file"`

	// SCAddr servicecenter address, which is the service registry address.
	// Cluster mode is supported, and multiple addresses are separated by an English ",".
	SCAddr string `yaml:"dc_addr"`

	// JoinAddr The management address of one gossip pool member.
	JoinAddr            string     `yaml:"join_addr"`
	TickerInterval      int        `yaml:"ticker_interval"`
	Profile             string     `yaml:"profile"`
	EnableCompression   bool       `yaml:"enable_compression"`
	AutoSync            bool       `yaml:"auto_sync"`
	TLSConfig           *TLSConfig `yaml:"tls_config"`
	ServicecenterPlugin string     `yaml:"servicecenter_plugin"`
}

// DefaultConfig returns the default config
func DefaultConfig() *Config {
	serfConf := serf.DefaultConfig()
	etcdConf := etcd.DefaultConfig()
	hostname, err := os.Hostname()
	if err != nil {
		log.Errorf(err, "Error determining hostname: %s", err)
		return nil
	}
	serfConf.NodeName = hostname
	etcdConf.Name = hostname
	return &Config{
		SCAddr:              fmt.Sprintf("127.0.0.1:%d", DefaultDCPort),
		TickerInterval:      DefaultTickerInterval,
		Config:              serfConf,
		Etcd:                etcdConf,
		TLSConfig:           DefaultTLSConfig(),
		ServicecenterPlugin: servicecenter.PluginName,
	}
}

// LoadConfig loads configuration from file
func LoadConfig(filepath string) (*Config, error) {
	if filepath == "" {
		filepath = DefaultConfigPath
	}
	if !(utils.IsFileExist(filepath)) {
		err := fmt.Errorf("file is not exist")
		log.Errorf(err, "Load config from %s failed", filepath)
		return nil, err
	}

	byteArr, err := ioutil.ReadFile(filepath)
	if err != nil {
		log.Errorf(err, "Load config from %s failed", filepath)
		return nil, err
	}

	conf := &Config{}
	err = yaml.Unmarshal(byteArr, conf)
	if err != nil {
		log.Errorf(err, "Unmarshal config file failed, content is %s", byteArr)
		return nil, err
	}
	return conf, nil
}

// Merge other configuration into the current configuration
func (c *Config) Merge(other *Config) {
	c.TLSConfig.Merge("", other.TLSConfig)
}

// Verify Provide config verification
func (c *Config) Verify() error {
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

	if c.JoinAddr != "" {
		c.RetryJoin = strings.Split(c.JoinAddr, ",")
	}

	if c.ClusterName == "" {
		c.ClusterName = fmt.Sprintf("%x", md5.Sum([]byte(c.SCAddr)))
	}

	c.TLSEnabled = c.TLSConfig.Enabled

	c.Etcd.SetName(c.NodeName)
	return nil
}
