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

package servicecenter

import (
	"context"
	"fmt"
	"sync"
	"time"

	pb "github.com/go-chassis/cari/discovery"

	"github.com/apache/servicecomb-service-center/datasource/etcd"
	"github.com/apache/servicecomb-service-center/datasource/etcd/path"
	"github.com/apache/servicecomb-service-center/datasource/etcd/sd"
	"github.com/apache/servicecomb-service-center/datasource/etcd/state"
	"github.com/apache/servicecomb-service-center/datasource/etcd/state/kvstore"
	"github.com/apache/servicecomb-service-center/pkg/dump"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/util"
	"github.com/apache/servicecomb-service-center/server/alarm"
	"github.com/go-chassis/foundation/gopool"
)

var (
	syncer     *Syncer
	syncerOnce sync.Once
)

type Syncer struct {
	Client *SCClientAggregate

	cachers map[kvstore.Type]*Cacher
}

func (c *Syncer) Initialize() {
	c.cachers = make(map[kvstore.Type]*Cacher)
	c.Client = GetOrCreateSCClient()
}

func (c *Syncer) Sync(ctx context.Context) {
	cache, errs := c.Client.GetScCache(ctx)
	if len(errs) > 0 {
		err := fmt.Errorf("%v", errs)
		log.Error("Sync catches errors", err)
		err = alarm.Raise(alarm.IDBackendConnectionRefuse,
			alarm.AdditionalContext(err.Error()))
		if err != nil {
			log.Error("", err)
		}
		if cache == nil {
			return
		}
	}
	err := alarm.Clear(alarm.IDBackendConnectionRefuse)
	if err != nil {
		log.Error("", err)
	}
	// microservice
	serviceCacher, ok := c.cachers[sd.TypeService]
	if ok {
		c.check(serviceCacher, &cache.Microservices, errs)
	}
	indexCacher, ok := c.cachers[sd.TypeServiceIndex]
	if ok {
		c.checkWithConflictHandleFunc(indexCacher, &cache.Indexes, errs, c.logConflictFunc)
	}
	aliasCacher, ok := c.cachers[sd.TypeServiceAlias]
	if ok {
		c.checkWithConflictHandleFunc(aliasCacher, &cache.Aliases, errs, c.logConflictFunc)
	}
	// microservice meta
	tagCacher, ok := c.cachers[sd.TypeServiceTag]
	if ok {
		c.check(tagCacher, &cache.Tags, errs)
	}
	depRuleCacher, ok := c.cachers[sd.TypeDependencyRule]
	if ok {
		c.check(depRuleCacher, &cache.DependencyRules, errs)
	}
	schemaSummaryCacher, ok := c.cachers[sd.TypeSchemaSummary]
	if ok {
		c.check(schemaSummaryCacher, &cache.Summaries, errs)
	}
	// instance
	instCacher, ok := c.cachers[sd.TypeInstance]
	if ok {
		c.check(instCacher, &cache.Instances, errs)
	}
}

func (c *Syncer) check(local *Cacher, remote dump.Getter, skipClusters map[string]error) {
	c.checkWithConflictHandleFunc(local, remote, skipClusters, c.skipHandleFunc)
}

func (c *Syncer) checkWithConflictHandleFunc(local *Cacher, remote dump.Getter, skipClusters map[string]error,
	conflictHandleFunc func(origin *dump.KV, conflict dump.Getter, index int)) {
	exists := make(map[string]*dump.KV)
	remote.ForEach(func(i int, v *dump.KV) bool {
		// because the result of the remote return may contain the same data as
		// the local cache of the current SC. So we need to ignore it and
		// prevent the aggregation result from increasing.
		if v.ClusterName == state.Configuration().ClusterName {
			return true
		}
		if kv, ok := exists[v.Key]; ok {
			conflictHandleFunc(kv, remote, i)
			return true
		}
		exists[v.Key] = v
		old := local.Cache().Get(v.Key)
		newKv := &kvstore.KeyValue{
			Key:         util.StringToBytesWithNoCopy(v.Key),
			Value:       v.Value,
			ModRevision: v.Rev,
			ClusterName: v.ClusterName,
		}
		switch {
		case old == nil:
			newKv.Version = 1
			newKv.CreateRevision = v.Rev
			local.Notify(pb.EVT_CREATE, v.Key, newKv)
		case old.ModRevision != v.Rev:
			// if connect to some cluster failed, then skip to notify changes
			// of these clusters to prevent publish the wrong changes events of kvs.
			if err, ok := skipClusters[old.ClusterName]; ok {
				log.Error(fmt.Sprintf("cluster[%s] temporarily unavailable, skip cluster[%s] event %s %s",
					old.ClusterName, v.ClusterName, pb.EVT_UPDATE, v.Key), err)
				break
			}
			newKv.Version = 1 + old.Version
			newKv.CreateRevision = old.CreateRevision
			local.Notify(pb.EVT_UPDATE, v.Key, newKv)
		}
		return true
	})

	var deletes []*kvstore.KeyValue
	local.Cache().ForEach(func(key string, v *kvstore.KeyValue) (next bool) {
		var exist bool
		remote.ForEach(func(_ int, v *dump.KV) bool {
			if v.ClusterName == state.Configuration().ClusterName {
				return true
			}
			exist = v.Key == key
			return !exist
		})
		if !exist {
			if err, ok := skipClusters[v.ClusterName]; ok {
				log.Error(fmt.Sprintf("cluster[%s] temporarily unavailable, skip event %s %s",
					v.ClusterName, pb.EVT_DELETE, v.Key), err)
				return true
			}
			deletes = append(deletes, v)
		}
		return true
	})
	for _, v := range deletes {
		local.Notify(pb.EVT_DELETE, util.BytesToStringWithNoCopy(v.Key), v)
	}
}

func (c *Syncer) skipHandleFunc(origin *dump.KV, conflict dump.Getter, index int) {
}

func (c *Syncer) logConflictFunc(origin *dump.KV, conflict dump.Getter, index int) {
	switch conflict.(type) {
	case *dump.MicroserviceIndexSlice:
		slice := conflict.(*dump.MicroserviceIndexSlice)
		keyValue := (*slice)[index]
		if serviceID := origin.Value.(string); keyValue.Value != serviceID {
			key := path.GetInfoFromSvcIndexKV(util.StringToBytesWithNoCopy(keyValue.Key))
			log.Warn(fmt.Sprintf("conflict! can not merge microservice index[%s][%s][%s/%s/%s/%s], found one[%s] in cluster[%s]",
				keyValue.ClusterName, keyValue.Value, key.Environment, key.AppId, key.ServiceName, key.Version,
				serviceID, origin.ClusterName))
		}
	case *dump.MicroserviceAliasSlice:
		slice := conflict.(*dump.MicroserviceAliasSlice)
		keyValue := (*slice)[index]
		if serviceID := origin.Value.(string); keyValue.Value != serviceID {
			key := path.GetInfoFromSvcAliasKV(util.StringToBytesWithNoCopy(keyValue.Key))
			log.Warn(fmt.Sprintf("conflict! can not merge microservice alias[%s][%s][%s/%s/%s/%s], found one[%s] in cluster[%s]",
				keyValue.ClusterName, keyValue.Value, key.Environment, key.AppId, key.ServiceName, key.Version,
				serviceID, origin.ClusterName))
		}
	}
}

func (c *Syncer) loop(ctx context.Context) {
	cfg := etcd.Configuration()
	select {
	case <-ctx.Done():
	case <-time.After(minWaitInterval):
		c.Sync(ctx)
		d := cfg.AutoSyncInterval
		if d == 0 {
			return
		}
	loop:
		for {
			select {
			case <-ctx.Done():
				break loop
			case <-time.After(d):
				// TODO support watching sc
				c.Sync(ctx)
			}
		}
	}

	log.Debug("service center clusters syncer is stopped")
}

// unsafe
func (c *Syncer) AddCacher(t kvstore.Type, cacher *Cacher) {
	c.cachers[t] = cacher
}

func (c *Syncer) Run() {
	c.Initialize()
	gopool.Go(c.loop)
}

func GetOrCreateSyncer() *Syncer {
	syncerOnce.Do(func() {
		syncer = &Syncer{}
		syncer.Run()
	})
	return syncer
}
