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
	scerr "github.com/apache/incubator-servicecomb-service-center/server/error"
	"golang.org/x/net/context"
	"sync"
)

var AdminServiceAPI = &AdminService{}

type AdminService struct {
}

func (service *AdminService) Dump(ctx context.Context, in *model.DumpRequest) (*model.DumpResponse, error) {

	domainProject := util.ParseDomainProject(ctx)
	var cache model.Cache

	if !core.IsDefaultDomainProject(domainProject) {
		return &model.DumpResponse{
			Response: pb.CreateResponse(scerr.ErrForbidden, "Required admin permission"),
			Cache:    cache,
		}, nil
	}

	service.dumpAll(&cache)

	return &model.DumpResponse{
		Response: pb.CreateResponse(pb.Response_SUCCESS, "Admin dump successfully"),
		Cache:    cache,
	}, nil
}

func (service *AdminService) dumpAll(cache *model.Cache) {
	var wg sync.WaitGroup

	service.parallel(&wg, func() {
		backend.Store().Service().Cacher().Cache().ForEach(func(k string, v *backend.KeyValue) (next bool) {
			cache.Microservices = append(cache.Microservices, model.Microservice{
				KV:    model.KV{Key: k, Rev: v.ModRevision},
				Value: v.Value.(*pb.MicroService),
			})
			return true
		})
	})
	service.parallel(&wg, func() {
		backend.Store().ServiceIndex().Cacher().Cache().ForEach(func(k string, v *backend.KeyValue) (next bool) {
			cache.Indexes = append(cache.Indexes, model.MicroserviceIndex{
				KV:    model.KV{Key: k, Rev: v.ModRevision},
				Value: v.Value.(string),
			})
			return true
		})
	})
	service.parallel(&wg, func() {
		backend.Store().ServiceAlias().Cacher().Cache().ForEach(func(k string, v *backend.KeyValue) (next bool) {
			cache.Aliases = append(cache.Aliases, model.MicroserviceAlias{
				KV:    model.KV{Key: k, Rev: v.ModRevision},
				Value: v.Value.(string),
			})
			return true
		})
	})
	service.parallel(&wg, func() {
		backend.Store().ServiceTag().Cacher().Cache().ForEach(func(k string, v *backend.KeyValue) (next bool) {
			cache.Tags = append(cache.Tags, model.Tag{
				KV:    model.KV{Key: k, Rev: v.ModRevision},
				Value: v.Value.(map[string]string),
			})
			return true
		})
	})
	service.parallel(&wg, func() {
		backend.Store().Rule().Cacher().Cache().ForEach(func(k string, v *backend.KeyValue) (next bool) {
			cache.Rules = append(cache.Rules, model.MicroServiceRule{
				KV:    model.KV{Key: k, Rev: v.ModRevision},
				Value: v.Value.(*pb.ServiceRule),
			})
			return true
		})
		backend.Store().DependencyRule().Cacher().Cache().ForEach(func(k string, v *backend.KeyValue) (next bool) {
			cache.DependencyRules = append(cache.DependencyRules, model.MicroServiceDependencyRule{
				KV:    model.KV{Key: k, Rev: v.ModRevision},
				Value: v.Value.(*pb.MicroServiceDependency),
			})
			return true
		})
	})
	service.parallel(&wg, func() {
		backend.Store().SchemaSummary().Cacher().Cache().ForEach(func(k string, v *backend.KeyValue) (next bool) {
			cache.Summaries = append(cache.Summaries, model.Summary{
				KV:    model.KV{Key: k, Rev: v.ModRevision},
				Value: v.Value.(string),
			})
			return true
		})
	})
	service.parallel(&wg, func() {
		backend.Store().Instance().Cacher().Cache().ForEach(func(k string, v *backend.KeyValue) (next bool) {
			cache.Instances = append(cache.Instances, model.Instance{
				KV:    model.KV{Key: k, Rev: v.ModRevision},
				Value: v.Value.(*pb.MicroServiceInstance),
			})
			return true
		})
	})

	wg.Wait()
}

func (service *AdminService) parallel(wg *sync.WaitGroup, f func()) {
	wg.Add(1)
	util.Go(func(_ context.Context) {
		f()
		wg.Done()
	})
}
