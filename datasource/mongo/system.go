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
	"strings"

	"github.com/apache/servicecomb-service-center/datasource"
	"github.com/apache/servicecomb-service-center/datasource/mongo/sd"
	"github.com/apache/servicecomb-service-center/pkg/dump"
	"github.com/apache/servicecomb-service-center/pkg/gopool"
)

const (
	DumpServicePrefix  = "/cse-sr/ms/files"
	DumpInstancePrefix = "/cse-sr/inst/files"
)

func (ds *DataSource) DumpCache(ctx context.Context, cache *dump.Cache) {
	gopool.New(ctx, gopool.Configure().Workers(2)).
		Do(func(_ context.Context) { setServiceValue(sd.Store().Service(), &cache.Microservices) }).
		Do(func(_ context.Context) { setInstanceValue(sd.Store().Instance(), &cache.Instances) }).
		Done()
}

func (ds *DataSource) DLock(ctx context.Context, request *datasource.DLockRequest) error {
	return nil
}

func (ds *DataSource) DUnlock(ctx context.Context, request *datasource.DUnlockRequest) error {
	return nil
}

func setServiceValue(e *sd.MongoCacher, setter dump.Setter) {
	e.Cache().ForEach(func(k string, kv interface{}) (next bool) {
		setter.SetValue(&dump.KV{
			Key:   constructKey(DumpServicePrefix, kv.(sd.Service).Domain, kv.(sd.Service).Project, k),
			Value: kv.(sd.Service).ServiceInfo,
		})
		return true
	})
}

func setInstanceValue(e *sd.MongoCacher, setter dump.Setter) {
	e.Cache().ForEach(func(k string, kv interface{}) (next bool) {
		setter.SetValue(&dump.KV{
			Key:   constructKey(DumpInstancePrefix, kv.(sd.Instance).Domain, kv.(sd.Instance).Project, kv.(sd.Instance).InstanceInfo.ServiceId, k),
			Value: kv.(sd.Instance).InstanceInfo,
		})
		return true
	})
}

func constructKey(strs ...string) string {
	return strings.Join(strs, "/")
}
