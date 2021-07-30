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

package etcd

import (
	"context"

	"github.com/apache/servicecomb-service-center/datasource"
	"github.com/apache/servicecomb-service-center/datasource/etcd/kv"
	"github.com/apache/servicecomb-service-center/datasource/etcd/sd"
	"github.com/apache/servicecomb-service-center/pkg/dump"
	"github.com/apache/servicecomb-service-center/pkg/etcdsync"
	"github.com/apache/servicecomb-service-center/pkg/gopool"
)

type SysManager struct {
	locks map[string]*etcdsync.DLock
}

func newSysManager() datasource.SystemManager {
	inst := &SysManager{
		locks: make(map[string]*etcdsync.DLock),
	}
	return inst
}
func (sm *SysManager) DumpCache(ctx context.Context) *dump.Cache {
	var cache dump.Cache
	gopool.New(ctx, gopool.Configure().Workers(2)).
		Do(func(_ context.Context) { setValue(kv.Store().Service(), &cache.Microservices) }).
		Do(func(_ context.Context) { setValue(kv.Store().ServiceIndex(), &cache.Indexes) }).
		Do(func(_ context.Context) { setValue(kv.Store().ServiceAlias(), &cache.Aliases) }).
		Do(func(_ context.Context) { setValue(kv.Store().ServiceTag(), &cache.Tags) }).
		Do(func(_ context.Context) { setValue(kv.Store().RuleIndex(), &cache.RuleIndexes) }).
		Do(func(_ context.Context) { setValue(kv.Store().Rule(), &cache.Rules) }).
		Do(func(_ context.Context) { setValue(kv.Store().DependencyRule(), &cache.DependencyRules) }).
		Do(func(_ context.Context) { setValue(kv.Store().SchemaSummary(), &cache.Summaries) }).
		Do(func(_ context.Context) { setValue(kv.Store().Instance(), &cache.Instances) }).
		Done()
	return &cache
}

func setValue(e sd.Adaptor, setter dump.Setter) {
	e.Cache().ForEach(func(k string, kv *sd.KeyValue) (next bool) {
		setter.SetValue(&dump.KV{
			Key:         k,
			Rev:         kv.ModRevision,
			Value:       kv.Value,
			ClusterName: kv.ClusterName,
		})
		return true
	})
}
