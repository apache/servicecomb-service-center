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
	"path/filepath"
	"reflect"

	"github.com/go-chassis/go-archaius"

	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/util"
)

var config Config

type Config struct {
	Sync *Sync `yaml:"sync"`
}

type Sync struct {
	EnableOnStart bool    `yaml:"enableOnStart"`
	Peers         []*Peer `yaml:"peers"`
}

type Peer struct {
	Name      string   `yaml:"name"`
	Kind      string   `yaml:"kind"`
	Endpoints []string `yaml:"endpoints"`
	Mode      []string `yaml:"mode"`
}

func Init() (error, bool) {
	err := archaius.AddFile(filepath.Join(util.GetAppRoot(), "conf", "syncer", "syncer.yaml"))
	if err != nil {
		log.Warn(fmt.Sprintf("can not add syncer config file source, error: %s", err))
		return err, false
	}

	err, isRefresh := Reload()
	if err != nil {
		log.Fatal("reload syncer configs failed", err)
		return err, false
	}
	return nil, isRefresh
}

// Reload all configurations
func Reload() (error, bool) {
	oldConfig := config
	err := archaius.UnmarshalConfig(&config)
	if err != nil {
		return err, false
	}
	if !reflect.DeepEqual(oldConfig, config) {
		return nil, true
	}
	return nil, false
}

// GetConfig return the syncer full configurations
func GetConfig() Config {
	return config
}

// SetConfig for UT
func SetConfig(c Config) {
	config = c
}
