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
package admin

import (
	"github.com/apache/incubator-servicecomb-service-center/pkg/util"
	"github.com/apache/incubator-servicecomb-service-center/server/admin/model"
	"github.com/apache/incubator-servicecomb-service-center/server/core"
	"github.com/apache/incubator-servicecomb-service-center/server/core/backend"
	pb "github.com/apache/incubator-servicecomb-service-center/server/core/proto"
	"golang.org/x/net/context"
)

var AdminServiceAPI = &AdminService{}

type AdminService struct {
}

func (service *AdminService) Dump(ctx context.Context, in *model.DumpRequest) (*model.DumpResponse, error) {

	domainProject := util.ParseDomainProject(ctx)
	cache := make(model.Cache)

	if core.IsDefaultDomainProject(domainProject) {
		service.dumpAll(cache)
	} else {
		service.dumpDomainProject(cache, domainProject)
	}

	return &model.DumpResponse{
		Response: pb.CreateResponse(pb.Response_SUCCESS, "Admin dump successfully."),
		Cache:    cache,
	}, nil
}

func (service *AdminService) dumpAll(cache model.Cache) {
	for id, indexer := range backend.Store().Entities() {
		var kvs []model.KV
		indexer.Cacher().Cache().ForEach(func(k string, v *backend.KeyValue) (next bool) {
			kvs = append(kvs, model.KV{
				Key: k, Value: v.Value, Rev: v.ModRevision,
			})
			return true
		})
		if len(kvs) == 0 {
			continue
		}
		cache[id.String()] = kvs
	}
}

func (service *AdminService) dumpDomainProject(cache model.Cache, domainProject string) {
	for id, indexer := range backend.Store().Entities() {
		cfg := indexer.Cacher().Config()
		if cfg == nil {
			continue
		}

		var kvs []model.KV
		var arr []*backend.KeyValue
		prefix := cfg.Prefix + core.SPLIT + domainProject + core.SPLIT
		indexer.Cacher().Cache().GetAll(prefix, &arr)
		for _, kv := range arr {
			kvs = append(kvs, model.KV{
				Key: util.BytesToStringWithNoCopy(kv.Key), Value: kv.Value, Rev: kv.ModRevision,
			})
		}
		if len(kvs) == 0 {
			continue
		}
		cache[id.String()] = kvs
	}
}
