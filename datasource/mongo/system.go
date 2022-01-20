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
	"context"

	"github.com/go-chassis/foundation/gopool"
	"github.com/patrickmn/go-cache"

	"github.com/apache/servicecomb-service-center/datasource"
	"github.com/apache/servicecomb-service-center/datasource/mongo/model"
	"github.com/apache/servicecomb-service-center/datasource/mongo/sd"
	"github.com/apache/servicecomb-service-center/pkg/dump"
	"github.com/apache/servicecomb-service-center/pkg/goutil"
	"github.com/apache/servicecomb-service-center/pkg/util"
)

type SysManager struct {
}

func (ds *SysManager) DumpCache(ctx context.Context) *dump.Cache {
	var cache dump.Cache
	goutil.New(gopool.Configure().Workers(2).WithContext(ctx)).
		Do(func(_ context.Context) { setServiceValue(sd.Store().Service(), &cache.Microservices) }).
		Do(func(_ context.Context) { setInstanceValue(sd.Store().Instance(), &cache.Instances) }).
		Done()
	return &cache
}

func setServiceValue(e *sd.MongoCacher, setter dump.Setter) {
	e.Cache().ForEach(func(k string, kv interface{}) (next bool) {
		service := kv.(cache.Item).Object.(model.Service)
		setter.SetValue(&dump.KV{
			Key: util.StringJoin([]string{datasource.ServiceKeyPrefix, service.Domain, service.Project, k},
				datasource.SPLIT),
			Value: service.Service,
		})
		return true
	})
}

func setInstanceValue(e *sd.MongoCacher, setter dump.Setter) {
	e.Cache().ForEach(func(k string, kv interface{}) (next bool) {
		instance := kv.(cache.Item).Object.(model.Instance)
		setter.SetValue(&dump.KV{
			Key: util.StringJoin([]string{datasource.InstanceKeyPrefix, instance.Domain, instance.Project,
				instance.Instance.ServiceId, k}, datasource.SPLIT),
			Value: instance.Instance,
		})
		return true
	})
}
