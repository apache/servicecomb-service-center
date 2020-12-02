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
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"

	"errors"

	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/syncer/pkg/utils"
	"gopkg.in/yaml.v2"
	"k8s.io/apimachinery/pkg/util/uuid"
)

// DefaultConfig returns the default config
func DefaultConfig() *Config {
	hostname, err := os.Hostname()
	if err != nil {
		log.Errorf(err, "Error determining hostname: %s", err)
		hostname = string(uuid.NewUUID())
	}

	return &Config{
		Mode:    ModeSingle,
		Node:    hostname,
		DataDir: defaultDataDir + hostname,
		Listener: Listener{
			BindAddr: "0.0.0.0:" + strconv.Itoa(defaultBindPort),
			RPCAddr:  "0.0.0.0:" + strconv.Itoa(defaultRPCPort),
			PeerAddr: "127.0.0.1:" + strconv.Itoa(defaultPeerPort),
		},
		Task: Task{
			Kind: "ticker",
			Params: []Label{
				{
					Key:   defaultTaskKey,
					Value: defaultTaskValue,
				},
			},
		},
		Registry: Registry{
			Address: "http://127.0.0.1:30100",
			Plugin:  defaultDCPluginName,
		},
	}
}

// LoadConfig loads configuration from file
func LoadConfig(filepath string) (*Config, error) {
	if filepath == "" {
		return nil, nil
	}
	if !(utils.IsFileExist(filepath)) {
		err := errors.New("file is not exist")
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

func (c *Config) GetTLSConfig(name string) *TLSConfig {
	return findInTLSConfigs(c.TLSConfigs, name)
}

func pathFromSSLEnvOrDefault(server, path string) string {
	env := os.Getenv(defaultEnvSSLRoot)
	if len(env) == 0 {
		wd, _ := os.Getwd()
		return filepath.Join(wd, defaultCertsDir, server, path)
	}
	return os.ExpandEnv(filepath.Join("$"+defaultEnvSSLRoot, server, path))
}
