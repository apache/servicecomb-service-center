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

package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/apache/servicecomb-service-center/datasource"
	"github.com/go-chassis/cari/discovery"
)

var (
	APIServiceExistence = "/v4/:project/registry/existence"
	APIServicesList     = "/v4/:project/registry/microservices"
	APIServiceInfo      = "/v4/:project/registry/microservices/:serviceId"

	APIProConDependency = "/v4/:project/registry/microservices/:providerId/consumers"
	APIConProDependency = "/v4/:project/registry/microservices/:consumerId/providers"

	APIInstancesList = "/v4/:project/registry/instances"
	APIHeartbeats    = "/v4/:project/registry/heartbeats"

	APIGovServicesList = "/v4/:project/govern/microservices"
	APIGovServiceInfo  = "/v4/:project/govern/microservices/:serviceId"
)

func init() {
	RegisterParseFunc(APIServiceInfo, serviceIdFunc)
	RegisterParseFunc(APIGovServiceInfo, serviceIdFunc)
	RegisterParseFunc(APIProConDependency, func(r *http.Request) ([]map[string]string, error) {
		return fromQueryKey(r, ":providerId")
	})
	RegisterParseFunc(APIConProDependency, func(r *http.Request) ([]map[string]string, error) {
		return fromQueryKey(r, ":consumerId")
	})
	RegisterParseFunc(APIInstancesList, serviceKeyFunc)
	RegisterParseFunc(APIServiceExistence, serviceKeyFunc)
	RegisterParseFunc(APIGovServicesList, serviceKeyFunc)
	RegisterParseFunc(APIServicesList, serviceInfoFunc)
}

func serviceIdFunc(r *http.Request) ([]map[string]string, error) {
	return fromQueryKey(r, ":serviceId")
}

func fromQueryKey(r *http.Request, queryKey string) ([]map[string]string, error) {
	ctx := r.Context()
	serviceId := r.URL.Query().Get(queryKey)
	return serviceIdToLabels(ctx, serviceId)
}

func serviceIdToLabels(ctx context.Context, serviceId string) ([]map[string]string, error) {
	response, err := datasource.Instance().GetService(ctx, &discovery.GetServiceRequest{ServiceId: serviceId})
	if err != nil {
		return nil, err
	}

	service := response.Service
	if service == nil {
		return nil, fmt.Errorf("resource %s not found", serviceId)
	}

	return []map[string]string{{
		"environment": service.Environment,
		"appId":       service.AppId,
		"serviceName": service.ServiceName,
	}}, nil
}

func serviceKeyFunc(r *http.Request) ([]map[string]string, error) {
	query := r.URL.Query()
	serviceId := query.Get("serviceId")
	if len(serviceId) > 0 {
		return serviceIdToLabels(r.Context(), serviceId)
	}

	env := query.Get("env")
	appID := query.Get("appId")
	serviceName := query.Get("serviceName")

	if len(env) == 0 && len(appID) == 0 && len(serviceName) == 0 {
		return ApplyAll(r)
	}

	return []map[string]string{{
		"environment": env,
		"appId":       appID,
		"serviceName": serviceName,
	}}, nil
}

func serviceInfoFunc(r *http.Request) ([]map[string]string, error) {
	if r.Method == http.MethodGet {
		return ApplyAll(r)
	}
	if r.Method == http.MethodDelete {
		return deleteServicesToLabels(r)
	}
	return createServiceToLabels(r)
}

func createServiceToLabels(r *http.Request) ([]map[string]string, error) {
	message, err := ReadBody(r)
	if err != nil {
		return nil, fmt.Errorf("read request body failed")
	}

	request := &discovery.CreateServiceRequest{}
	err = json.Unmarshal(message, request)
	if err != nil {
		return nil, fmt.Errorf("unmarshal CreateServiceRequest failed")
	}

	service := request.Service
	if service == nil {
		return nil, fmt.Errorf("invalid CreateServiceRequest")
	}

	return []map[string]string{{
		"environment": service.Environment,
		"appId":       service.AppId,
		"serviceName": service.ServiceName,
	}}, nil
}

func deleteServicesToLabels(r *http.Request) ([]map[string]string, error) {
	message, err := ReadBody(r)
	if err != nil {
		return nil, fmt.Errorf("read request body failed")
	}

	request := &discovery.DelServicesRequest{}

	err = json.Unmarshal(message, request)
	if err != nil {
		return nil, fmt.Errorf("unmarshal DelServicesRequest failed")
	}

	ctx := r.Context()
	var labels []map[string]string
	for _, serviceId := range request.ServiceIds {
		ls, err := serviceIdToLabels(ctx, serviceId)
		if err != nil {
			return nil, err
		}
		labels = append(labels, ls...)
	}
	return labels, nil
}

func ReadBody(r *http.Request) ([]byte, error) {
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	r.Body = ioutil.NopCloser(bytes.NewReader(data))
	return data, nil
}
