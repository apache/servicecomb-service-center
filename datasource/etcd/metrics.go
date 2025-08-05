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
	"fmt"

	"github.com/go-chassis/cari/discovery"
	"github.com/go-chassis/etcdadpt"

	"github.com/apache/servicecomb-service-center/datasource"
	"github.com/apache/servicecomb-service-center/datasource/etcd/path"
	"github.com/apache/servicecomb-service-center/datasource/etcd/sd"
	"github.com/apache/servicecomb-service-center/pkg/log"
)

type MetricsManager struct {
}

var (
	defaultLabel = datasource.MetricsLabels{
		Domain:           datasource.RegistryDomain,
		Project:          datasource.RegistryProject,
		Framework:        discovery.Unknown,
		FrameworkVersion: discovery.Unknown,
	}
)

func (m *MetricsManager) Report(ctx context.Context, r datasource.MetricsReporter) error {
	reportDomains(ctx, r)
	reportServices(ctx, r)
	reportSchemas(ctx, r)
	return nil
}

func reportDomains(ctx context.Context, r datasource.MetricsReporter) {
	key := path.GenerateDomainKey("")
	domainsResp, err := sd.Domain().Search(ctx,
		etcdadpt.WithCacheOnly(), etcdadpt.WithCountOnly(),
		etcdadpt.WithStrKey(key),
		etcdadpt.WithPrefix())
	if err != nil {
		log.Error("query all domains failed", err)
		return
	}
	r.DomainAdd(float64(domainsResp.Count))
}

func reportSchemas(ctx context.Context, r datasource.MetricsReporter) {
	key := path.GetServiceSchemaSummaryRootKey("")
	schemaKeysResp, err := sd.SchemaSummary().Search(ctx,
		etcdadpt.WithCacheOnly(), etcdadpt.WithKeyOnly(),
		etcdadpt.WithStrKey(key),
		etcdadpt.WithPrefix())
	if err != nil {
		log.Error("query all schemas failed", err)
		return
	}
	isSchemaMetricSet := false
	for _, keyValue := range schemaKeysResp.Kvs {
		domainProject, _, _ := path.GetInfoFromSchemaSummaryKV(keyValue.Key)
		domain, project := path.SplitDomainProject(domainProject)
		labels := datasource.MetricsLabels{
			Domain:  domain,
			Project: project,
		}
		r.SchemaAdd(1, labels)
		isSchemaMetricSet = true
	}
	if !isSchemaMetricSet {
		r.SchemaAdd(0, defaultLabel)
	}
}

func reportServices(ctx context.Context, r datasource.MetricsReporter) {
	key := path.GetServiceRootKey("")
	servicesResp, err := sd.Service().Search(ctx,
		etcdadpt.WithCacheOnly(),
		etcdadpt.WithStrKey(key),
		etcdadpt.WithPrefix())
	if err != nil {
		log.Error("query all microservices failed", err)
		return
	}
	isMetricSet := false
	for _, keyValue := range servicesResp.Kvs {
		service := keyValue.Value.(*discovery.MicroService)
		_, domainProject := path.GetInfoFromSvcKV(keyValue.Key)
		if datasource.IsGlobal(discovery.MicroServiceToKey(domainProject, service)) {
			continue
		}
		domain, project := path.SplitDomainProject(domainProject)
		frameworkName, frameworkVersion := discovery.ToFrameworkLabel(service)
		labels := datasource.MetricsLabels{
			Domain:           domain,
			Project:          project,
			Framework:        frameworkName,
			FrameworkVersion: frameworkVersion,
		}
		r.ServiceAdd(1, labels)
		instanceCount := getInstanceCount4Service(ctx, domainProject, service)
		r.FrameworkSet(labels)
		r.InstanceAdd(float64(instanceCount), labels)
		isMetricSet = true
	}
	// 0也应该是有效的指标，无指标是异常场景
	if !isMetricSet {
		r.ServiceAdd(0, defaultLabel)
		r.SetFrameworkValue(0, defaultLabel)
		r.InstanceAdd(0, defaultLabel)
	}
}

func getInstanceCount4Service(ctx context.Context, domainProject string, service *discovery.MicroService) (instanceCount int64) {
	instancesResp, err := sd.Instance().Search(ctx,
		etcdadpt.WithCacheOnly(), etcdadpt.WithCountOnly(),
		etcdadpt.WithStrKey(path.GenerateInstanceKey(domainProject, service.ServiceId, "")),
		etcdadpt.WithPrefix())
	if err != nil {
		log.Error(fmt.Sprintf("query microservice %s isntances failed", service.ServiceId), err)
		return
	}
	return instancesResp.Count
}
