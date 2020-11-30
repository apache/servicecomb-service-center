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

package datasource

import (
	"fmt"

	"github.com/apache/servicecomb-service-center/pkg/log"
)

type dataSourceEngine func(opts Options) (DataSource, error)

var (
	plugins        = make(map[ImplName]dataSourceEngine)
	dataSourceInst DataSource
)

// load plugins configuration into plugins
func Install(pluginImplName string, engineFunc dataSourceEngine) {
	plugins[ImplName(pluginImplName)] = engineFunc
}

// construct storage plugin instance
// invoked by sc main process
/* Usage:
 * interval, err := time.ParseDuration(core.ServerInfo.Config.CompactInterval)
 * if err != nil {
 * 	log.Errorf(err, "invalid compact interval %s, reset to default interval 12h", core.ServerInfo.Config.CompactInterval)
 * 	interval = 12 * time.Hour
 * }
 * Init(Options{
 * 	CompactIndexDelta:    core.ServerInfo.Config.CompactIndexDelta,
 * 	CompactInterval:      interval,
 * })
 */
func Init(opts Options) error {
	if opts.PluginImplName == "" {
		return nil
	}

	dataSourceEngine, ok := plugins[opts.PluginImplName]
	if !ok {
		return fmt.Errorf("plugin implement not supported [%s]", opts.PluginImplName)
	}
	var err error
	dataSourceInst, err = dataSourceEngine(opts)
	if err != nil {
		return err
	}
	log.Info(fmt.Sprintf("datasource plugin [%s] enabled", opts.PluginImplName))
	return nil
}

// Instance is the instance of DataSource
func Instance() DataSource {
	return dataSourceInst
}
