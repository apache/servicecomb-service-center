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

package mongo

import (
	"fmt"

	"github.com/go-chassis/cari/db"
	dconfig "github.com/go-chassis/cari/db/config"

	"github.com/apache/servicecomb-service-center/datasource"
	"github.com/apache/servicecomb-service-center/datasource/mongo/heartbeat"
	"github.com/apache/servicecomb-service-center/datasource/mongo/sd"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/server/config"
	"github.com/apache/servicecomb-service-center/server/plugin/security/tlsconf"
)

const defaultExpireTime = 300
const defaultPoolSize = 1000

func init() {
	datasource.Install("mongo", NewDataSource)
}

type DataSource struct {
	metadataManager datasource.MetadataManager
	sysManager      datasource.SystemManager
	depManager      datasource.DependencyManager
	scManager       datasource.SCManager
	metricsManager  datasource.MetricsManager
	syncManager     datasource.SyncManager
}

func (ds *DataSource) SystemManager() datasource.SystemManager {
	return ds.sysManager
}

func (ds *DataSource) DependencyManager() datasource.DependencyManager {
	return ds.depManager
}

func (ds *DataSource) MetadataManager() datasource.MetadataManager {
	return ds.metadataManager
}

func (ds *DataSource) SCManager() datasource.SCManager {
	return ds.scManager
}

func (ds *DataSource) MetricsManager() datasource.MetricsManager {
	return ds.metricsManager
}

func (ds *DataSource) SyncManager() datasource.SyncManager {
	return ds.syncManager
}

func NewDataSource(opts datasource.Options) (datasource.DataSource, error) {
	// TODO: construct a reasonable DataSource instance
	inst := &DataSource{}
	// TODO: deal with exception
	if err := inst.initialize(); err != nil {
		return nil, err
	}
	inst.scManager = &SCManager{}
	inst.depManager = &DepManager{}
	inst.sysManager = &SysManager{}
	inst.metadataManager = &MetadataManager{
		InstanceTTL:        opts.InstanceTTL,
		InstanceProperties: opts.InstanceProperties,
	}
	inst.metricsManager = &MetricsManager{}
	inst.syncManager = &SyncManager{}
	return inst, nil
}

func (ds *DataSource) initialize() error {
	var err error
	// init heartbeat plugins
	err = ds.initPlugins()
	if err != nil {
		return err
	}
	// init mongo client
	err = ds.initClient()
	if err != nil {
		return err
	}
	// create db index and validator
	ensureDB()

	// if fast register enabled, init fast register service
	initFastRegister()

	// init cache
	ds.initStore()
	return nil
}

func (ds *DataSource) initPlugins() error {
	kind := config.GetString("heartbeat.kind", "cache")
	err := heartbeat.Init(heartbeat.Options{PluginImplName: heartbeat.ImplName(kind)})
	if err != nil {
		log.Fatal("heartbeat init failed", err)
		return err
	}
	return nil
}

func (ds *DataSource) initClient() error {
	cfg := dconfig.Config{Kind: "mongo"}
	cfg.URI = config.GetString("registry.mongo.cluster.uri", "mongodb://localhost:27017",
		config.WithStandby("manager_cluster"))
	cfg.SSLEnabled = config.GetBool("ssl.enable", false)
	cfg.Logger = log.Logger
	if cfg.SSLEnabled {
		tlsConfig, err := tlsconf.ClientConfig()
		if err != nil {
			log.Fatal("get datasource tlsConfig failed", err)
			return err
		}
		cfg.TLSConfig = tlsConfig
	}
	poolSize := config.GetInt("registry.mongo.cluster.poolSize", defaultPoolSize)
	if poolSize <= 0 {
		log.Warn(fmt.Sprintf("mongo cluster poolSize[%d] is too small, set to default size", poolSize))
		poolSize = defaultPoolSize
	}
	cfg.PoolSize = poolSize
	return db.Init(&cfg)
}

func (ds *DataSource) initStore() {
	if !config.GetRegistry().EnableCache {
		log.Debug("cache is disabled")
		return
	}
	sd.Store().Run()
	<-sd.Store().Ready()
}

func initFastRegister() {
	fastRegConfig := FastRegConfiguration()

	if fastRegConfig.QueueSize > 0 {
		fastRegisterService := NewFastRegisterInstanceService()
		SetFastRegisterInstanceService(fastRegisterService)

		NewRegisterTimeTask().Start()
	}
}
