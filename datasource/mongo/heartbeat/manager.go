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

package heartbeat

import (
	"fmt"
	"github.com/apache/servicecomb-service-center/pkg/log"
)

type healthCheckEngine func(opts Options) (HealthCheck, error)

var (
	plugins             = make(map[ImplName]healthCheckEngine)
	healthCheckInstance HealthCheck
)

// load plugins configuration into plugins
func Install(pluginImplName string, engineFunc healthCheckEngine) {
	plugins[ImplName(pluginImplName)] = engineFunc
}

// construct plugin instance
func Init(opts Options) error {
	inst, err := New(opts)
	if err != nil {
		return err
	}
	healthCheckInstance = inst
	log.Info(fmt.Sprintf("healthcheck plugin [%s] enabled", opts.PluginImplName))
	return nil
}

func New(opts Options) (HealthCheck, error) {
	if opts.PluginImplName == "" {
		return nil, fmt.Errorf("plugin implement name is nil")
	}
	f, ok := plugins[opts.PluginImplName]
	if !ok {
		return nil, fmt.Errorf("plugin implement not supported [#{opts.PluginImplName}]")
	}
	return f(opts)
}

// Instance is the instance of HealthCheck
func Instance() HealthCheck {
	return healthCheckInstance
}
