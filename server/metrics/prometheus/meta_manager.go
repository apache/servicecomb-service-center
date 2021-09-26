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

package prometheus

import (
	"context"
	"fmt"

	"github.com/apache/servicecomb-service-center/pkg/log"
	pb "github.com/apache/servicecomb-service-center/pkg/registry"
	"github.com/apache/servicecomb-service-center/server/core"
	"github.com/apache/servicecomb-service-center/server/core/backend"
	"github.com/apache/servicecomb-service-center/server/core/proto"
	"github.com/apache/servicecomb-service-center/server/plugin/registry"
)

type MetricsManager struct {
}

func (m *MetricsManager) Report(ctx context.Context, r *MetaReporter) error {
	reportDomains(ctx, r)
	reportServices(ctx, r)
	reportSchemas(ctx, r)
	return nil
}

func reportDomains(ctx context.Context, r *MetaReporter) {
	key := core.GenerateDomainKey("")
	domainsResp, err := backend.Store().Domain().Search(ctx,
		registry.WithCacheOnly(), registry.WithCountOnly(),
		registry.WithStrKey(key),
		registry.WithPrefix())
	if err != nil {
		log.Error("query all domains failed", err)
		return
	}
	r.DomainAdd(float64(domainsResp.Count))
}

func reportSchemas(ctx context.Context, r *MetaReporter) {
	key := core.GetServiceSchemaSummaryRootKey("")
	schemaKeysResp, err := backend.Store().SchemaSummary().Search(ctx,
		registry.WithCacheOnly(), registry.WithKeyOnly(),
		registry.WithStrKey(key),
		registry.WithPrefix())
	if err != nil {
		log.Error("query all schemas failed", err)
		return
	}
	for _, keyValue := range schemaKeysResp.Kvs {
		domainProject, _, _ := core.GetInfoFromSchemaSummaryKV(keyValue.Key)
		domain, project := core.SplitDomainProject(domainProject)
		labels := MetricsLabels{
			Domain:  domain,
			Project: project,
		}
		r.SchemaAdd(1, labels)
	}
}

func reportServices(ctx context.Context, r *MetaReporter) {
	key := core.GetServiceRootKey("")
	servicesResp, err := backend.Store().Service().Search(ctx,
		registry.WithCacheOnly(),
		registry.WithStrKey(key),
		registry.WithPrefix())
	if err != nil {
		log.Error("query all microservices failed", err)
		return
	}
	for _, keyValue := range servicesResp.Kvs {
		service := keyValue.Value.(*pb.MicroService)
		_, domainProject := core.GetInfoFromSvcKV(keyValue.Key)
		if core.IsShared(proto.MicroServiceToKey(domainProject, service)) {
			continue
		}
		domain, project := core.SplitDomainProject(domainProject)
		frameworkName, frameworkVersion := pb.ToFrameworkLabel(service)
		labels := MetricsLabels{
			Domain:           domain,
			Project:          project,
			Framework:        frameworkName,
			FrameworkVersion: frameworkVersion,
		}
		r.ServiceAdd(1, labels)

		reportInstances(ctx, r, domainProject, service)
	}
}

func reportInstances(ctx context.Context, r *MetaReporter, domainProject string, service *pb.MicroService) {
	instancesResp, err := backend.Store().Instance().Search(ctx,
		registry.WithCacheOnly(), registry.WithCountOnly(),
		registry.WithStrKey(core.GenerateInstanceKey(domainProject, service.ServiceId, "")),
		registry.WithPrefix())
	if err != nil {
		log.Error(fmt.Sprintf("query microservice %s isntances failed", service.ServiceId), err)
		return
	}
	if instancesResp.Count == 0 {
		return
	}
	count := float64(instancesResp.Count)
	domain, project := core.SplitDomainProject(domainProject)
	frameworkName, frameworkVersion := pb.ToFrameworkLabel(service)
	labels := MetricsLabels{
		Domain:           domain,
		Project:          project,
		Framework:        frameworkName,
		FrameworkVersion: frameworkVersion,
	}
	r.FrameworkSet(labels)
	r.InstanceAdd(count, labels)
}
