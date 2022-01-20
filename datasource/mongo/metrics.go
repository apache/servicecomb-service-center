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
	"fmt"

	"github.com/go-chassis/cari/discovery"

	"github.com/apache/servicecomb-service-center/datasource"
	"github.com/apache/servicecomb-service-center/datasource/mongo/dao"
	"github.com/apache/servicecomb-service-center/datasource/mongo/model"
	mutil "github.com/apache/servicecomb-service-center/datasource/mongo/util"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/util"
)

var globalServiceNames []string

func init() {
	for name := range datasource.GlobalServiceNames {
		globalServiceNames = append(globalServiceNames, name)
	}
}

type MetricsManager struct {
}

func (m *MetricsManager) Report(ctx context.Context, r datasource.MetricsReporter) error {
	reportDomains(ctx, r)
	reportServices(ctx, r)
	return nil
}

func reportDomains(ctx context.Context, r datasource.MetricsReporter) {
	count, err := dao.CountDomain(ctx)
	if err != nil {
		log.Error("count domains failed", err)
		return
	}
	r.DomainAdd(float64(count))
}

func reportServices(ctx context.Context, r datasource.MetricsReporter) {
	services, err := dao.GetServices(ctx, mutil.NewFilter(mutil.NotGlobal()))
	if err != nil {
		log.Error("query all services failed", err)
		return
	}
	for _, service := range services {
		frameworkName, frameworkVersion := discovery.ToFrameworkLabel(service.Service)
		labels := datasource.MetricsLabels{
			Domain:           service.Domain,
			Project:          service.Project,
			Framework:        frameworkName,
			FrameworkVersion: frameworkVersion,
		}
		r.ServiceAdd(1, labels)
		r.FrameworkSet(labels)

		reportInstances(ctx, r, service)
		reportSchemas(ctx, r, service)
	}
}

func reportInstances(ctx context.Context, r datasource.MetricsReporter, service *model.Service) {
	serviceID := service.Service.ServiceId
	count, err := dao.CountInstance(ctx, mutil.NewFilter(mutil.InstanceServiceID(serviceID)))
	if err != nil {
		log.Error(fmt.Sprintf("count service %s instance failed", serviceID), err)
		return
	}
	frameworkName, frameworkVersion := discovery.ToFrameworkLabel(service.Service)
	r.InstanceAdd(float64(count), datasource.MetricsLabels{
		Domain:           service.Domain,
		Project:          service.Project,
		Framework:        frameworkName,
		FrameworkVersion: frameworkVersion,
	})
}

func reportSchemas(ctx context.Context, r datasource.MetricsReporter, service *model.Service) {
	serviceID := service.Service.ServiceId
	count, err := dao.CountSchema(util.SetDomainProject(ctx, service.Domain, service.Project), serviceID)
	if err != nil {
		log.Error(fmt.Sprintf("count service %s schema failed", serviceID), err)
		return
	}
	r.SchemaAdd(float64(count), datasource.MetricsLabels{
		Domain:  service.Domain,
		Project: service.Project,
	})
}
