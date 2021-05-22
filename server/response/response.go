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

package response

import (
	"github.com/go-chassis/cari/discovery"
)

func init() {
	RegisterFilter("/v4/:project/registry/microservices", MicroserviceListFilter)
	//response.RegisterFilter("/v4/:project/registry/instances", )
	RegisterFilter("/v4/:project/registry/microservices/:providerId/consumers", ProvidersListFilter)
	RegisterFilter("/v4/:project/registry/microservices/:consumerId/providers", ConsumersListFilter)
	// control panel apis
	RegisterFilter("/v4/:project/govern/microservices", MicroServiceInfoListFilter)
	RegisterFilter("/v4/:project/govern/apps", AppIDListFilter)
}

func MicroserviceListFilter(obj interface{}, labels []map[string]string) interface{} {
	servicesResponse, ok := obj.(*discovery.GetServicesResponse)
	if !ok {
		return obj
	}
	servicesResponse.Services = filterMicroservices(servicesResponse.Services, labels)
	return servicesResponse
}

func filterMicroservices(sources []*discovery.MicroService, labelsList []map[string]string) []*discovery.MicroService {
	var services []*discovery.MicroService
	for _, service := range sources {
		for _, labels := range labelsList {
			if env, ok := labels["environment"]; ok && service.Environment != env {
				continue
			}
			if app, ok := labels["appId"]; ok && service.AppId != app {
				continue
			}
			if name, ok := labels["serviceName"]; ok && service.ServiceName != name {
				continue
			}
			services = append(services, service)
			break
		}
	}
	return services
}

func ProvidersListFilter(obj interface{}, labels []map[string]string) interface{} {
	servicesResponse, ok := obj.(*discovery.GetConDependenciesResponse)
	if !ok {
		return obj
	}
	servicesResponse.Providers = filterMicroservices(servicesResponse.Providers, labels)
	return servicesResponse
}

func ConsumersListFilter(obj interface{}, labels []map[string]string) interface{} {
	servicesResponse, ok := obj.(*discovery.GetProDependenciesResponse)
	if !ok {
		return obj
	}
	servicesResponse.Consumers = filterMicroservices(servicesResponse.Consumers, labels)
	return servicesResponse
}

func MicroServiceInfoListFilter(obj interface{}, labelsList []map[string]string) interface{} {
	servicesResponse, ok := obj.(*discovery.GetServicesInfoResponse)
	if !ok {
		return obj
	}
	var services []*discovery.ServiceDetail
	for _, service := range servicesResponse.AllServicesDetail {
		for _, labels := range labelsList {
			if env, ok := labels["environment"]; ok && service.MicroService.Environment != env {
				continue
			}
			if app, ok := labels["appId"]; ok && service.MicroService.AppId != app {
				continue
			}
			if name, ok := labels["serviceName"]; ok && service.MicroService.ServiceName != name {
				continue
			}
			services = append(services, service)
			break
		}
	}
	servicesResponse.AllServicesDetail = services
	return servicesResponse
}

func AppIDListFilter(obj interface{}, labelsList []map[string]string) interface{} {
	appsResponse, ok := obj.(*discovery.GetAppsResponse)
	if !ok {
		return obj
	}
	var apps []string
	for _, appID := range appsResponse.AppIds {
		for _, labels := range labelsList {
			if app, ok := labels["appId"]; ok && appID != app {
				continue
			}
			apps = append(apps, appID)
			break
		}
	}
	appsResponse.AppIds = apps
	return appsResponse
}
