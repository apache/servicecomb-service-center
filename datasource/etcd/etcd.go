// Licensed to the Apache Software Foundation (ASF) under one or more
// contributor license agreements.  See the NOTICE file distributed with
// this work for additional information regarding copyright ownership.
// The ASF licenses this file to You under the Apache License, Version 2.0
// (the "License"); you may not use this file except in compliance with
// the License.  You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package etcd

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/apache/servicecomb-service-center/datasource"
	"github.com/apache/servicecomb-service-center/datasource/etcd/event"
	"github.com/apache/servicecomb-service-center/datasource/etcd/mux"
	"github.com/apache/servicecomb-service-center/datasource/etcd/sd"
	"github.com/apache/servicecomb-service-center/datasource/etcd/state"
	tracer "github.com/apache/servicecomb-service-center/datasource/etcd/tracing"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/server/config"
	"github.com/go-chassis/foundation/gopool"
	"github.com/little-cui/etcdadpt"
	"github.com/little-cui/etcdadpt/middleware/tracing"
)

var clustersIndex = make(map[string]int)

func init() {
	datasource.Install("etcd", NewDataSource)
	datasource.Install("embeded_etcd", NewDataSource) //TODO remove misspell in future
	datasource.Install("embedded_etcd", NewDataSource)

	sd.RegisterInnerTypes()
}

type DataSource struct {
	Options *datasource.Options

	metadataManager datasource.MetadataManager
	sysManager      datasource.SystemManager
	depManager      datasource.DependencyManager
	scManager       datasource.SCManager
	metricsManager  datasource.MetricsManager
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

func NewDataSource(opts datasource.Options) (datasource.DataSource, error) {
	log.Warn("data source enable etcd mode")

	etcdCfg := Configuration()
	etcdCfg.Kind = opts.Kind
	etcdCfg.Logger = opts.Logger
	etcdCfg.SslEnabled = opts.SslEnabled
	etcdCfg.TLSConfig = opts.TLSConfig
	etcdCfg.ConnectedFunc = opts.ConnectedFunc
	etcdCfg.ErrorFunc = opts.ErrorFunc
	opts.Config = etcdCfg
	opts.Config.Init()
	opts.SslEnabled = opts.SslEnabled && strings.Contains(strings.ToLower(opts.ClusterAddresses), "https://")
	inst := &DataSource{
		Options: &opts,
	}
	if err := inst.initialize(); err != nil {
		return nil, err
	}
	inst.metadataManager = &MetadataManager{
		SchemaNotEditable:  opts.SchemaNotEditable,
		InstanceTTL:        opts.InstanceTTL,
		InstanceProperties: opts.InstanceProperties,
	}
	inst.sysManager = newSysManager()
	inst.depManager = &DepManager{}
	inst.scManager = &SCManager{}
	inst.metricsManager = &MetricsManager{}
	return inst, nil
}

func (ds *DataSource) initialize() error {
	// Wait for kv store ready
	ds.initKvStore()
	// Compact
	ds.autoCompact()
	return nil
}

func (ds *DataSource) initClustersIndex() {
	clusterMap, err := etcdadpt.ListCluster(context.Background())
	if err != nil {
		log.Fatal("init clusters index failed", err)
	}
	var clusters []string
	for name := range clusterMap {
		clusters = append(clusters, name)
	}
	sort.Strings(clusters)
	for i, name := range clusters {
		clustersIndex[name] = i
	}
}

func (ds *DataSource) initPlugins() {
	// registry
	tracing.Register(tracer.New())
	err := etcdadpt.Init(ds.Options.Config)
	if err != nil {
		log.Fatal("client init failed", err)
	}
	// clusters
	ds.initClustersIndex()

	// discovery
	kind := config.GetString("discovery.kind", "etcd", config.WithStandby("discovery_plugin"))
	err = state.Init(state.Config{
		Kind:        kind,
		ClusterName: ds.Options.ClusterName,
		Logger:      ds.Options.Logger,
		EnableCache: ds.Options.EnableCache,
	})
	if err != nil {
		log.Fatal("sd init failed", err)
	}
}

func (ds *DataSource) initKvStore() {
	// init client/sd plugins
	ds.initPlugins()
	// Add events handlers
	event.Initialize()
}

func (ds *DataSource) autoCompact() {
	delta := ds.Options.CompactIndexDelta
	interval := ds.Options.CompactInterval
	if delta <= 0 || interval == 0 {
		return
	}
	gopool.Go(func(ctx context.Context) {
		log.Info(fmt.Sprintf("enabled the automatic compact mechanism, compact once every %s, reserve %d", interval, delta))
		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(interval):
				lock, err := mux.Try(mux.GlobalLock)
				if err != nil {
					log.Error("can not compact backend by this service center instance now", err)
					continue
				}

				err = etcdadpt.Instance().Compact(ctx, delta)
				if err != nil {
					log.Error("", err)
				}

				if err := lock.Unlock(); err != nil {
					log.Error("", err)
				}
			}
		}
	})
}
