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

package buildin

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chassis/cari/discovery"
	rbacmodel "github.com/go-chassis/cari/rbac"

	"github.com/apache/servicecomb-service-center/datasource"
	"github.com/apache/servicecomb-service-center/pkg/rest"
	"github.com/apache/servicecomb-service-center/server/plugin/auth"
	"github.com/apache/servicecomb-service-center/server/service/rbac"
)

const (
	LabelEnvironment = "environment"
	LabelAppID       = "appId"
	LabelServiceName = "serviceName"
	QueryEnv         = "env"
)

var (
	// APIServiceExistence Apply by service key or serviceId
	// - /v4/:project/registry/existence?env=xxx&appId=xxx&serviceName=xxx
	// - /v4/:project/registry/existence?serviceId=xxx&schemaId=xxx
	APIServiceExistence = "/v4/:project/registry/existence"
	// APIServicesList Method GET: apply all by optional service key
	// Method POST or DELETE: apply by request body
	APIServicesList     = "/v4/:project/registry/microservices"
	APIServiceInfo      = "/v4/:project/registry/microservices/:serviceId"
	APIProConDependency = "/v4/:project/registry/microservices/:providerId/consumers"
	APIConProDependency = "/v4/:project/registry/microservices/:consumerId/providers"
	// APIDiscovery Apply by service key
	APIDiscovery = "/v4/:project/registry/instances"
	// APIBatchDiscovery Apply by request body
	APIBatchDiscovery = "/v4/:project/registry/instances/action"
	// APIHeartbeats Apply by request body
	APIHeartbeats = "/v4/:project/registry/heartbeats"
	// APIGovServicesList Apply by optional service key
	// - /v4/:project/govern/microservices?appId=xxx
	// Apply all:
	// - /v4/:project/govern/microservices?options=statistics
	// - /v4/:project/govern/microservices/statistics
	APIGovServicesList = "/v4/:project/govern/microservices"
	APIGovServiceInfo  = "/v4/:project/govern/microservices/:serviceId"
)

func init() {
	RegisterParseFunc(APIServiceInfo, ByServiceID)
	RegisterParseFunc(APIGovServiceInfo, ByServiceID)
	RegisterParseFunc(APIProConDependency, func(r *http.Request) (*auth.ResourceScope, error) {
		return fromQueryKey(r, ":providerId")
	})
	RegisterParseFunc(APIConProDependency, func(r *http.Request) (*auth.ResourceScope, error) {
		return fromQueryKey(r, ":consumerId")
	})
	RegisterParseFunc(APIDiscovery, ByServiceKey)
	RegisterParseFunc(APIServiceExistence, ByServiceKey)
	RegisterParseFunc(APIGovServicesList, ApplyAll)
	RegisterParseFunc(APIServicesList, ByRequestBody)
	RegisterParseFunc(APIBatchDiscovery, ByDiscoveryRequestBody)
	RegisterParseFunc(APIHeartbeats, ByHeartbeatRequestBody)
}

func ByServiceID(r *http.Request) (*auth.ResourceScope, error) {
	return fromQueryKey(r, ":serviceId")
}
func fromQueryKey(r *http.Request, queryKey string) (*auth.ResourceScope, error) {
	ctx := r.Context()
	apiPath, ok := ctx.Value(rest.CtxMatchPattern).(string)
	if !ok {
		return nil, ErrCtxMatchPatternNotFound
	}
	serviceID := r.URL.Query().Get(queryKey)
	labels, err := serviceIDToLabels(ctx, serviceID)
	if err != nil {
		return nil, err
	}
	return &auth.ResourceScope{
		Type:   rbacmodel.GetResource(apiPath),
		Labels: labels,
		Verb:   rbac.MethodToVerbs[r.Method],
	}, nil
}
func serviceIDToLabels(ctx context.Context, serviceID string) ([]map[string]string, error) {
	service, err := datasource.GetMetadataManager().GetService(ctx, &discovery.GetServiceRequest{ServiceId: serviceID})
	if err != nil {
		return nil, err
	}
	return []map[string]string{{
		LabelEnvironment: service.Environment,
		LabelAppID:       service.AppId,
		LabelServiceName: service.ServiceName,
	}}, nil
}

func ByServiceKey(r *http.Request) (*auth.ResourceScope, error) {
	query := r.URL.Query()

	if _, ok := query["serviceId"]; ok {
		return fromQueryKey(r, "serviceId")
	}

	apiPath, ok := r.Context().Value(rest.CtxMatchPattern).(string)
	if !ok {
		return nil, ErrCtxMatchPatternNotFound
	}

	return &auth.ResourceScope{
		Type: rbacmodel.GetResource(apiPath),
		Labels: []map[string]string{{
			LabelEnvironment: query.Get(QueryEnv),
			LabelAppID:       query.Get(LabelAppID),
			LabelServiceName: query.Get(LabelServiceName),
		}},
		Verb: rbac.MethodToVerbs[r.Method],
	}, nil
}

func ByRequestBody(r *http.Request) (*auth.ResourceScope, error) {
	if r.Method == http.MethodGet {
		// get or list by query string
		return ApplyAll(r)
	}
	return fromRequestBody(r)
}
func fromRequestBody(r *http.Request) (*auth.ResourceScope, error) {
	apiPath, ok := r.Context().Value(rest.CtxMatchPattern).(string)
	if !ok {
		return nil, ErrCtxMatchPatternNotFound
	}

	var (
		labels []map[string]string
		err    error
	)
	if r.Method == http.MethodDelete {
		// batch delete
		labels, err = deleteServicesToLabels(r)
	} else {
		// create service
		labels, err = createServiceToLabels(r)
	}
	if err != nil {
		return nil, err
	}

	return &auth.ResourceScope{
		Type:   rbacmodel.GetResource(apiPath),
		Labels: labels,
		Verb:   rbac.MethodToVerbs[r.Method],
	}, nil
}
func createServiceToLabels(r *http.Request) ([]map[string]string, error) {
	message, err := rest.ReadBody(r)
	if err != nil {
		return nil, err
	}

	request := &discovery.CreateServiceRequest{}
	err = json.Unmarshal(message, request)
	if err != nil {
		return nil, err
	}

	service := request.Service
	if service == nil {
		return nil, fmt.Errorf("invalid CreateServiceRequest")
	}

	return []map[string]string{{
		LabelEnvironment: service.Environment,
		LabelAppID:       service.AppId,
		LabelServiceName: service.ServiceName,
	}}, nil
}
func deleteServicesToLabels(r *http.Request) ([]map[string]string, error) {
	message, err := rest.ReadBody(r)
	if err != nil {
		return nil, err
	}

	request := &discovery.DelServicesRequest{}

	err = json.Unmarshal(message, request)
	if err != nil {
		return nil, err
	}

	ctx := r.Context()
	var labels []map[string]string
	for _, serviceID := range request.ServiceIds {
		ls, err := serviceIDToLabels(ctx, serviceID)
		if err != nil {
			return nil, err
		}
		labels = append(labels, ls...)
	}
	return labels, nil
}

func ByDiscoveryRequestBody(r *http.Request) (*auth.ResourceScope, error) {
	apiPath, ok := r.Context().Value(rest.CtxMatchPattern).(string)
	if !ok {
		return nil, ErrCtxMatchPatternNotFound
	}

	message, err := rest.ReadBody(r)
	if err != nil {
		return nil, err
	}

	request := &discovery.BatchFindInstancesRequest{}

	err = json.Unmarshal(message, request)
	if err != nil {
		return nil, err
	}

	ctx := r.Context()
	var labels []map[string]string
	for _, it := range request.Instances {
		ls, err := serviceIDToLabels(ctx, it.Instance.ServiceId)
		if err != nil {
			return nil, err
		}
		labels = append(labels, ls...)
	}

	return &auth.ResourceScope{
		Type:   rbacmodel.GetResource(apiPath),
		Labels: labels,
		Verb:   "get",
	}, nil
}

func ByHeartbeatRequestBody(r *http.Request) (*auth.ResourceScope, error) {
	apiPath, ok := r.Context().Value(rest.CtxMatchPattern).(string)
	if !ok {
		return nil, ErrCtxMatchPatternNotFound
	}

	message, err := rest.ReadBody(r)
	if err != nil {
		return nil, err
	}

	request := &discovery.HeartbeatSetRequest{}

	err = json.Unmarshal(message, request)
	if err != nil {
		return nil, err
	}

	ctx := r.Context()
	var labels []map[string]string
	for _, instance := range request.Instances {
		ls, err := serviceIDToLabels(ctx, instance.ServiceId)
		if err != nil {
			return nil, err
		}
		labels = append(labels, ls...)
	}

	return &auth.ResourceScope{
		Type:   rbacmodel.GetResource(apiPath),
		Labels: labels,
		Verb:   "update",
	}, nil
}
