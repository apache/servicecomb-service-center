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
	"github.com/apache/servicecomb-service-center/pkg/plugin"
	"k8s.io/kube-openapi/pkg/validation/spec"
)

const (
	EnvironmentDev  = "dev"
	EnvironmentProd = "prod"
)

// AppConfig is yaml file struct
type AppConfig struct {
	Gov    *Gov          `yaml:"gov"`
	Server *ServerConfig `yaml:"server"`
}
type Gov struct {
	DistMap    map[string]DistributorOptions `json:"-" yaml:",inline"`
	MatchGroup *Policy                       `json:"match-group" yaml:"-"`
	Policies   Policies                      `yaml:"-"`
}
type Policies map[string]Policy
type Policy struct {
	ValidationSpec *spec.Schema
}
type DistributorOptions struct {
	Name     string `yaml:"name"`
	Type     string `yaml:"type"`
	Endpoint string `yaml:"endpoint"`
}

// GetImplName return the impl name
func (c *AppConfig) GetImplName(kind plugin.Kind) string {
	return GetString(kind.String()+".kind", plugin.Buildin, WithStandby(kind.String()+"_plugin"))
}
func (c *AppConfig) GetPluginDir() string {
	return c.Server.Config.PluginsDir
}
