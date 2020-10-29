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
	"github.com/apache/servicecomb-service-center/datasource"
	"github.com/apache/servicecomb-service-center/datasource/etcd/client"
	"github.com/apache/servicecomb-service-center/datasource/etcd/job"
	"github.com/apache/servicecomb-service-center/datasource/etcd/kv"
	"github.com/apache/servicecomb-service-center/datasource/etcd/mux"
	"github.com/apache/servicecomb-service-center/datasource/etcd/sd"
	"github.com/apache/servicecomb-service-center/pkg/gopool"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/server/config"
	"time"
)

func init() {
	datasource.Install("etcd", NewDataSource)
	datasource.Install("embeded_etcd", NewDataSource)
}

type DataSource struct {
	// SchemaEditable determines whether schema modification is allowed for
	SchemaEditable bool
	// TTL options
	ttlFromEnv int64
	// Compact options
	CompactIndexDelta int64
	CompactInterval   time.Duration
}

func NewDataSource(opts datasource.Options) (datasource.DataSource, error) {
	// TODO: construct a reasonable DataSource instance
	log.Warnf("dependency data source enable etcd mode")

	inst := &DataSource{
		SchemaEditable:    opts.SchemaEditable,
		ttlFromEnv:        opts.TTL,
		CompactInterval:   opts.CompactInterval,
		CompactIndexDelta: opts.CompactIndexDelta,
	}
	// TODO: deal with exception
	if err := inst.initialize(); err != nil {
		return nil, err
	}
	return inst, nil
}

func (ds *DataSource) initialize() error {
	// TODO: init dependency members
	// init client/sd plugins
	ds.initPlugins()
	// Wait for kv store ready
	ds.initKvStore()
	// Compact
	ds.autoCompact()
	// Jobs
	job.ClearNoInstanceServices()
	return nil
}

func (ds *DataSource) initPlugins() {
	kind := config.GetString("registry.kind", "", config.WithStandby("registry_plugin"))
	err := client.Init(client.Options{PluginImplName: client.ImplName(kind)})
	if err != nil {
		log.Fatalf(err, "client init failed")
	}
	kind = config.GetString("discovery.kind", "", config.WithStandby("discovery_plugin"))
	err = sd.Init(sd.Options{PluginImplName: sd.ImplName(kind)})
	if err != nil {
		log.Fatalf(err, "sd init failed")
	}
}

func (ds *DataSource) initKvStore() {
	kv.Store().Run()
	<-kv.Store().Ready()
}

func (ds *DataSource) autoCompact() {
	delta := ds.CompactIndexDelta
	interval := ds.CompactInterval
	if delta <= 0 || interval == 0 {
		return
	}
	gopool.Go(func(ctx context.Context) {
		log.Infof("enabled the automatic compact mechanism, compact once every %s, reserve %d", interval, delta)
		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(interval):
				lock, err := mux.Try(mux.GlobalLock)
				if err != nil {
					log.Errorf(err, "can not compact backend by this service center instance now")
					continue
				}

				err = client.Instance().Compact(ctx, delta)
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
