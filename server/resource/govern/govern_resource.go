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

package govern

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/apache/servicecomb-service-center/pkg/rest"
	"github.com/apache/servicecomb-service-center/pkg/util"
	governsvc "github.com/apache/servicecomb-service-center/server/service/govern"
	pb "github.com/go-chassis/cari/discovery"
	"github.com/go-chassis/foundation/stringutil"
)

// Service 治理相关接口服务
type Resource struct {
	//
}

// URLPatterns 路由
func (res *Resource) URLPatterns() []rest.Route {
	return []rest.Route{
		{Method: http.MethodGet, Path: "/v4/:project/govern/microservices/:serviceId", Func: res.GetService},
		{Method: http.MethodGet, Path: "/v4/:project/govern/relations", Func: res.Draw},
		{Method: http.MethodGet, Path: "/v4/:project/govern/microservices", Func: res.ListService},
		{Method: http.MethodGet, Path: "/v4/:project/govern/apps", Func: res.ListApp},
		{Method: http.MethodGet, Path: "/v4/:project/govern/statistics", Func: res.GetOverview},
	}
}

// Draw 获取依赖连接图详细依赖关系
func (res *Resource) Draw(w http.ResponseWriter, r *http.Request) {
	graph, err := governsvc.Draw(r.Context(), util.StringTRUE(r.URL.Query().Get("withShared")))
	if err != nil {
		rest.WriteServiceError(w, err)
		return
	}
	rest.WriteResponse(w, r, nil, graph)
}

// GetService 查询服务详细信息
func (res *Resource) GetService(w http.ResponseWriter, r *http.Request) {
	serviceID := r.URL.Query().Get(":serviceId")
	request := &pb.GetServiceRequest{
		ServiceId: serviceID,
	}
	ctx := r.Context()
	serviceDetail, err := governsvc.GetServiceDetail(ctx, request)
	if err != nil {
		rest.WriteServiceError(w, err)
		return
	}
	rest.WriteResponse(w, r, nil, &pb.GetServiceDetailResponse{Service: serviceDetail})
}

func (res *Resource) ListService(w http.ResponseWriter, r *http.Request) {
	request := &pb.GetServicesInfoRequest{}
	ctx := r.Context()
	query := r.URL.Query()
	optsStr := query.Get("options")
	request.Options = strings.Split(optsStr, ",")
	request.AppId = query.Get("appId")
	request.ServiceName = query.Get("serviceName")
	request.Environment = query.Get("env")
	request.WithShared = util.StringTRUE(query.Get("withShared"))
	request.Properties = ParseProperties(query, "property")
	countOnly := query.Get("countOnly")
	if countOnly != "0" && countOnly != "1" && strings.TrimSpace(countOnly) != "" {
		rest.WriteError(w, pb.ErrInvalidParams, "parameter countOnly must be 1 or 0")
		return
	}
	if countOnly == "1" {
		request.CountOnly = true
	}
	resp, err := governsvc.ListServiceDetail(ctx, request)
	if err != nil {
		rest.WriteServiceError(w, err)
		return
	}
	rest.WriteResponse(w, r, nil, resp)
}

func ParseProperties(query url.Values, key string) map[string]string {
	propertyList := query[key]
	properties := make(map[string]string, len(propertyList))
	for _, kv := range propertyList {
		if !strings.Contains(kv, ":") {
			properties[kv] = ""
			continue
		}

		k, v := stringutil.SplitToTwo(kv, ":")
		properties[k] = v
	}
	return properties
}

func (res *Resource) GetOverview(w http.ResponseWriter, r *http.Request) {
	st, err := governsvc.GetOverview(r.Context(), &pb.GetServicesRequest{})
	if err != nil {
		rest.WriteServiceError(w, err)
		return
	}
	rest.WriteResponse(w, r, nil, &pb.GetServicesInfoStatisticsResponse{Statistics: st})
}

func (res *Resource) ListApp(w http.ResponseWriter, r *http.Request) {
	request := &pb.GetAppsRequest{}
	ctx := r.Context()
	query := r.URL.Query()
	request.Environment = query.Get("env")
	request.WithShared = util.StringTRUE(query.Get("withShared"))
	resp, err := governsvc.ListApp(ctx, request)
	if err != nil {
		rest.WriteServiceError(w, err)
		return
	}
	rest.WriteResponse(w, r, nil, resp)
}
