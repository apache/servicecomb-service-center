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
	"github.com/apache/servicecomb-service-center/pkg/gopool"
	"github.com/apache/servicecomb-service-center/pkg/model"
	"github.com/apache/servicecomb-service-center/server/core/backend"
	"github.com/apache/servicecomb-service-center/server/plugin/discovery"
)

func (ds *DataSource) DumpCache(ctx context.Context, cache *model.Cache) {
	gopool.New(ctx, gopool.Configure().Workers(2)).
		Do(func(_ context.Context) { setValue(backend.Store().Service(), &cache.Microservices) }).
		Do(func(_ context.Context) { setValue(backend.Store().ServiceIndex(), &cache.Indexes) }).
		Do(func(_ context.Context) { setValue(backend.Store().ServiceAlias(), &cache.Aliases) }).
		Do(func(_ context.Context) { setValue(backend.Store().ServiceTag(), &cache.Tags) }).
		Do(func(_ context.Context) { setValue(backend.Store().RuleIndex(), &cache.RuleIndexes) }).
		Do(func(_ context.Context) { setValue(backend.Store().Rule(), &cache.Rules) }).
		Do(func(_ context.Context) { setValue(backend.Store().DependencyRule(), &cache.DependencyRules) }).
		Do(func(_ context.Context) { setValue(backend.Store().SchemaSummary(), &cache.Summaries) }).
		Do(func(_ context.Context) { setValue(backend.Store().Instance(), &cache.Instances) }).
		Done()
}

func setValue(e discovery.Adaptor, setter model.Setter) {
	e.Cache().ForEach(func(k string, kv *discovery.KeyValue) (next bool) {
		setter.SetValue(&model.KV{
			Key:         k,
			Rev:         kv.ModRevision,
			Value:       kv.Value,
			ClusterName: kv.ClusterName,
		})
		return true
	})
}
