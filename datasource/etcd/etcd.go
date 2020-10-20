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
	"errors"
	"github.com/apache/servicecomb-service-center/datasource"
	"github.com/apache/servicecomb-service-center/datasource/etcd/mux"
	"github.com/apache/servicecomb-service-center/datasource/etcd/registry"
	"github.com/apache/servicecomb-service-center/pkg/gopool"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"time"
)

// TODO: define error with names here

var ErrNotUnique = errors.New("kv result is not unique")

func init() {
	// TODO: set logger
	// TODO: register storage plugin to plugin manager
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

func NewDataSource(opts datasource.Options) *DataSource {
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
		return inst
	}
	return inst
}

func (ds *DataSource) initialize() error {
	// TODO: init dependency members
	ds.autoCompact()
	return nil
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

				err = registry.Instance().Compact(ctx, delta)
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
