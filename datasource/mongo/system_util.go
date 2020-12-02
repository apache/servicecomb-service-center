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

	"github.com/apache/servicecomb-service-center/datasource/mongo/client"
	"github.com/apache/servicecomb-service-center/pkg/dump"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"go.mongodb.org/mongo-driver/bson"

	"strings"
)

const (
	DumpServicePrefix  = "/cse-sr/ms/files"
	DumpInstancePrefix = "/cse-sr/inst/files"
)

func SetDumpServices(ctx context.Context, cache *dump.Cache) bool {
	res, err := client.GetMongoClient().Find(ctx, CollectionService, bson.M{})
	if err != nil {
		log.Error("SetDumpServices access mongo failed", err)
		return false
	}
	for res.Next(ctx) {
		var tmp *Service
		err := res.Decode(&tmp)
		if err != nil {
			log.Error("SetDumpServices decode failed", err)
			return false
		}
		cache.Microservices = append(cache.Microservices, &dump.Microservice{
			KV: &dump.KV{
				Key:   generatorDumpKey(DumpServicePrefix, tmp.Domain, tmp.Project, tmp.ServiceInfo.ServiceId),
				Rev:   0,
				Value: tmp.ServiceInfo,
			},
			Value: tmp.ServiceInfo,
		})
	}
	return true
}

func SetDumpInstances(ctx context.Context, cache *dump.Cache) bool {
	res, err := client.GetMongoClient().Find(ctx, CollectionInstance, bson.M{})
	if err != nil {
		log.Error("SetDumpInstances access mongo failed", err)
		return false
	}
	for res.Next(ctx) {
		var tmp *Instance
		err := res.Decode(&tmp)
		if err != nil {
			log.Error("SetDumpInstances decode failed", err)
			return false
		}
		cache.Instances = append(cache.Instances, &dump.Instance{
			KV: &dump.KV{
				Key:   generatorDumpKey(DumpInstancePrefix, tmp.Domain, tmp.Project, tmp.InstanceInfo.ServiceId, tmp.InstanceInfo.InstanceId),
				Value: tmp.InstanceInfo,
			},
			Value: tmp.InstanceInfo,
		})
	}
	return true
}

func generatorDumpKey(strs ...string) string {
	return strings.Join(strs, "/")
}
