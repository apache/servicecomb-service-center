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
	"github.com/apache/servicecomb-service-center/datasource/etcd/kv"
	"github.com/apache/servicecomb-service-center/datasource/etcd/mux"
	"github.com/apache/servicecomb-service-center/datasource/etcd/sd"
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
}

type DataSource struct {
	Options *datasource.Options

	accountLockManager datasource.AccountLockManager
	accountManager     datasource.AccountManager
	metadataManager    datasource.MetadataManager
	roleManager        datasource.RoleManager
	sysManager         datasource.SystemManager
	depManager         datasource.DependencyManager
	scManager          datasource.SCManager
	metricsManager     datasource.MetricsManager
}

func (ds *DataSource) AccountLockManager() datasource.AccountLockManager {
	return ds.accountLockManager
}

func (ds *DataSource) SystemManager() datasource.SystemManager {
	return ds.sysManager
}

func (ds *DataSource) AccountManager() datasource.AccountManager {
	return ds.accountManager
}

func (ds *DataSource) RoleManager() datasource.RoleManager {
	return ds.roleManager
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
	// TODO: construct a reasonable DataSource instance
	log.Warn("data source enable etcd mode")

	if len(opts.Config.Kind) == 0 {
		opts.Config = Configuration()
	}
	opts.Config.Init()
	opts.SslEnabled = opts.SslEnabled && strings.Contains(strings.ToLower(opts.ClusterAddresses), "https://")
	if opts.ReleaseAccountAfter == 0 {
		opts.ReleaseAccountAfter = 15 * time.Minute
	}
	inst := &DataSource{
		Options: &opts,
	}
	if err := inst.initialize(); err != nil {
		return nil, err
	}
	inst.accountManager = &AccountManager{}
	inst.accountLockManager = NewAccountLockManager(opts.ReleaseAccountAfter)
	inst.roleManager = &RoleManager{}
	inst.metadataManager = newMetadataManager(opts.SchemaNotEditable, opts.InstanceTTL)
	inst.sysManager = newSysManager()
	inst.depManager = &DepManager{}
	inst.scManager = &SCManager{}
	inst.metricsManager = &MetricsManager{}
	return inst, nil
}

func (ds *DataSource) initialize() error {
	// init client/sd plugins
	ds.initPlugins()
	// Add events handlers
	event.Initialize()
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

	// discovery
	kind := config.GetString("discovery.kind", "", config.WithStandby("discovery_plugin"))
	err = sd.Init(sd.Options{Kind: sd.Kind(kind)})
	if err != nil {
		log.Fatal("sd init failed", err)
	}

	// clusters
	ds.initClustersIndex()
}

func (ds *DataSource) initKvStore() {
	kv.Store().Run()
	<-kv.Store().Ready()
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
