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

	"github.com/go-chassis/cari/dlock"

	"github.com/apache/servicecomb-service-center/datasource/rbac"
	"github.com/apache/servicecomb-service-center/datasource/schema"
	"github.com/apache/servicecomb-service-center/eventbase/datasource"
	"github.com/apache/servicecomb-service-center/pkg/log"
)

type dataSourceEngine func(opts Options) (DataSource, error)

var (
	plugins        = make(map[string]dataSourceEngine)
	dataSourceInst DataSource
)

// load plugins configuration into plugins
func Install(pluginImplName string, engineFunc dataSourceEngine) {
	plugins[pluginImplName] = engineFunc
}

// Init construct storage plugin instance
// invoked by sc main process
func Init(opts Options) error {
	if opts.Kind == "" {
		return nil
	}

	err := initDatasource(opts)
	if err != nil {
		return err
	}
	err = schema.Init(schema.Options{Kind: opts.Kind})
	if err != nil {
		return err
	}
	err = rbac.Init(rbac.Options{Kind: opts.Kind})
	if err != nil {
		return err
	}
	err = dlock.Init(dlock.Options{
		Kind: opts.Kind,
	})
	if err != nil {
		return err
	}
	// init eventbase
	err = datasource.Init(&datasource.Config{
		Kind:   opts.Kind,
		Logger: log.Logger,
	})
	return err
}

func initDatasource(opts Options) error {
	dataSourceEngine, ok := plugins[opts.Kind]
	if !ok {
		return fmt.Errorf("plugin implement not supported [%s]", opts.Kind)
	}
	var err error
	dataSourceInst, err = dataSourceEngine(opts)
	if err != nil {
		return err
	}
	log.Info(fmt.Sprintf("datasource plugin [%s] enabled", opts.Kind))
	return nil
}

func GetSCManager() SCManager {
	return dataSourceInst.SCManager()
}
func GetMetadataManager() MetadataManager {
	return dataSourceInst.MetadataManager()
}
func GetSystemManager() SystemManager {
	return dataSourceInst.SystemManager()
}
func GetDependencyManager() DependencyManager {
	return dataSourceInst.DependencyManager()
}
func GetMetricsManager() MetricsManager {
	return dataSourceInst.MetricsManager()
}
