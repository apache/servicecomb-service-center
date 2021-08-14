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
	"fmt"
	"sync"
	"time"

	"github.com/apache/servicecomb-service-center/pkg/model"
	proto "github.com/apache/servicecomb-service-center/pkg/registry"

	"github.com/apache/servicecomb-service-center/pkg/gopool"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/util"
	"github.com/apache/servicecomb-service-center/server/alarm"
	"github.com/apache/servicecomb-service-center/server/core"
	"github.com/apache/servicecomb-service-center/server/core/backend"
	"github.com/apache/servicecomb-service-center/server/plugin/discovery"
	"github.com/apache/servicecomb-service-center/server/plugin/registry"

	"context"
)

var (
	syncer     *Syncer
	syncerOnce sync.Once
)

type Syncer struct {
	Client *SCClientAggregate

	cachers map[discovery.Type]*Cacher
}

func (c *Syncer) Initialize() {
	c.cachers = make(map[discovery.Type]*Cacher)
	c.Client = GetOrCreateSCClient()
}

func (c *Syncer) Sync(ctx context.Context) {
	cache, errs := c.Client.GetScCache(ctx)
	if len(errs) > 0 {
		err := fmt.Errorf("%v", errs)
		log.Errorf(err, "Sync catches errors")
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
	serviceCacher, ok := c.cachers[backend.SERVICE]
	if ok {
		c.check(serviceCacher, &cache.Microservices, errs)
	}
	indexCacher, ok := c.cachers[backend.ServiceIndex]
	if ok {
		c.checkWithConflictHandleFunc(indexCacher, &cache.Indexes, errs, c.logConflictFunc)
	}
	aliasCacher, ok := c.cachers[backend.ServiceAlias]
	if ok {
		c.checkWithConflictHandleFunc(aliasCacher, &cache.Aliases, errs, c.logConflictFunc)
	}
	// microservice meta
	tagCacher, ok := c.cachers[backend.ServiceTag]
	if ok {
		c.check(tagCacher, &cache.Tags, errs)
	}
	ruleCacher, ok := c.cachers[backend.RULE]
	if ok {
		c.check(ruleCacher, &cache.Rules, errs)
	}
	ruleIndexCacher, ok := c.cachers[backend.RuleIndex]
	if ok {
		c.check(ruleIndexCacher, &cache.RuleIndexes, errs)
	}
	depRuleCacher, ok := c.cachers[backend.DependencyRule]
	if ok {
		c.check(depRuleCacher, &cache.DependencyRules, errs)
	}
	schemaSummaryCacher, ok := c.cachers[backend.SchemaSummary]
	if ok {
		c.check(schemaSummaryCacher, &cache.Summaries, errs)
	}
	// instance
	instCacher, ok := c.cachers[backend.INSTANCE]
	if ok {
		c.check(instCacher, &cache.Instances, errs)
	}
}

func (c *Syncer) check(local *Cacher, remote model.Getter, skipClusters map[string]error) {
	c.checkWithConflictHandleFunc(local, remote, skipClusters, c.skipHandleFunc)
}

func (c *Syncer) checkWithConflictHandleFunc(local *Cacher, remote model.Getter, skipClusters map[string]error,
	conflictHandleFunc func(origin *model.KV, conflict model.Getter, index int)) {
	exists := make(map[string]*model.KV)
	remote.ForEach(func(i int, v *model.KV) bool {
		// because the result of the remote return may contain the same data as
		// the local cache of the current SC. So we need to ignore it and
		// prevent the aggregation result from increasing.
		if v.ClusterName == registry.Configuration().ClusterName {
			return true
		}
		if kv, ok := exists[v.Key]; ok {
			conflictHandleFunc(kv, remote, i)
			return true
		}
		exists[v.Key] = v
		old := local.Cache().Get(v.Key)
		newKv := &discovery.KeyValue{
			Key:         util.StringToBytesWithNoCopy(v.Key),
			Value:       v.Value,
			ModRevision: v.Rev,
			ClusterName: v.ClusterName,
		}
		switch {
		case old == nil:
			newKv.Version = 1
			newKv.CreateRevision = v.Rev
			local.Notify(proto.EVT_CREATE, v.Key, newKv)
		case old.ModRevision != v.Rev:
			// if connect to some cluster failed, then skip to notify changes
			// of these clusters to prevent publish the wrong changes events of kvs.
			if err, ok := skipClusters[old.ClusterName]; ok {
				log.Errorf(err, "cluster[%s] temporarily unavailable, skip cluster[%s] event %s %s",
					old.ClusterName, v.ClusterName, proto.EVT_UPDATE, v.Key)
				break
			}
			newKv.Version = 1 + old.Version
			newKv.CreateRevision = old.CreateRevision
			local.Notify(proto.EVT_UPDATE, v.Key, newKv)
		}
		return true
	})

	var deletes []*discovery.KeyValue
	local.Cache().ForEach(func(key string, v *discovery.KeyValue) (next bool) {
		var exist bool
		remote.ForEach(func(_ int, v *model.KV) bool {
			if v.ClusterName == registry.Configuration().ClusterName {
				return true
			}
			exist = v.Key == key
			return !exist
		})
		if !exist {
			if err, ok := skipClusters[v.ClusterName]; ok {
				log.Errorf(err, "cluster[%s] temporarily unavailable, skip event %s %s",
					v.ClusterName, proto.EVT_DELETE, v.Key)
				return true
			}
			deletes = append(deletes, v)
		}
		return true
	})
	for _, v := range deletes {
		local.Notify(proto.EVT_DELETE, util.BytesToStringWithNoCopy(v.Key), v)
	}
}

func (c *Syncer) skipHandleFunc(origin *model.KV, conflict model.Getter, index int) {
}

func (c *Syncer) logConflictFunc(origin *model.KV, conflict model.Getter, index int) {
	switch conflict.(type) {
	case *model.MicroserviceIndexSlice:
		slice := conflict.(*model.MicroserviceIndexSlice)
		kv := (*slice)[index]
		if serviceID := origin.Value.(string); kv.Value != serviceID {
			key := core.GetInfoFromSvcIndexKV(util.StringToBytesWithNoCopy(kv.Key))
			log.Warnf("conflict! can not merge microservice index[%s][%s][%s/%s/%s/%s], found one[%s] in cluster[%s]",
				kv.ClusterName, kv.Value, key.Environment, key.AppId, key.ServiceName, key.Version,
				serviceID, origin.ClusterName)
		}
	case *model.MicroserviceAliasSlice:
		slice := conflict.(*model.MicroserviceAliasSlice)
		kv := (*slice)[index]
		if serviceID := origin.Value.(string); kv.Value != serviceID {
			key := core.GetInfoFromSvcAliasKV(util.StringToBytesWithNoCopy(kv.Key))
			log.Warnf("conflict! can not merge microservice alias[%s][%s][%s/%s/%s/%s], found one[%s] in cluster[%s]",
				kv.ClusterName, kv.Value, key.Environment, key.AppId, key.ServiceName, key.Version,
				serviceID, origin.ClusterName)
		}
	}
}

func (c *Syncer) loop(ctx context.Context) {
	select {
	case <-ctx.Done():
	case <-time.After(minWaitInterval):
		c.Sync(ctx)
		d := registry.Configuration().AutoSyncInterval
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
func (c *Syncer) AddCacher(t discovery.Type, cacher *Cacher) {
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
