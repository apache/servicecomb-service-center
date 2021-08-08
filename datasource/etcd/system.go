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
	"github.com/apache/servicecomb-service-center/datasource/etcd/sd"
	"github.com/apache/servicecomb-service-center/datasource/etcd/state"
	"github.com/apache/servicecomb-service-center/datasource/etcd/state/kvstore"
	"github.com/apache/servicecomb-service-center/pkg/dump"
	"github.com/apache/servicecomb-service-center/pkg/etcdsync"
	"github.com/apache/servicecomb-service-center/pkg/goutil"
	"github.com/go-chassis/foundation/gopool"
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
	goutil.New(gopool.Configure().WithContext(ctx).Workers(10)).
		Do(func(_ context.Context) { setValue(sd.Service(), &cache.Microservices) }).
		Do(func(_ context.Context) { setValue(sd.ServiceIndex(), &cache.Indexes) }).
		Do(func(_ context.Context) { setValue(sd.ServiceAlias(), &cache.Aliases) }).
		Do(func(_ context.Context) { setValue(sd.ServiceTag(), &cache.Tags) }).
		Do(func(_ context.Context) { setValue(sd.DependencyRule(), &cache.DependencyRules) }).
		Do(func(_ context.Context) { setValue(sd.SchemaSummary(), &cache.Summaries) }).
		Do(func(_ context.Context) { setValue(sd.Instance(), &cache.Instances) }).
		Done()
	return &cache
}

func setValue(e state.State, setter dump.Setter) {
	e.Cache().ForEach(func(k string, kv *kvstore.KeyValue) (next bool) {
		setter.SetValue(&dump.KV{
			Key:         k,
			Rev:         kv.ModRevision,
			Value:       kv.Value,
			ClusterName: kv.ClusterName,
		})
		return true
	})
}
