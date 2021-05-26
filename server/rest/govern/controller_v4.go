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

	"strings"

	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/rest"
	"github.com/apache/servicecomb-service-center/pkg/util"
	"github.com/apache/servicecomb-service-center/server/core"
	pb "github.com/go-chassis/cari/discovery"
)

// Service 治理相关接口服务
type ResourceV4 struct {
	//
}

// URLPatterns 路由
func (governService *ResourceV4) URLPatterns() []rest.Route {
	return []rest.Route{
		{Method: http.MethodGet, Path: "/v4/:project/govern/microservices/:serviceId", Func: governService.GetServiceDetail},
		{Method: http.MethodGet, Path: "/v4/:project/govern/relations", Func: governService.GetGraph},
		{Method: http.MethodGet, Path: "/v4/:project/govern/microservices", Func: governService.GetAllServicesInfo},
		{Method: http.MethodGet, Path: "/v4/:project/govern/apps", Func: governService.GetAllApplications},
		{Method: http.MethodGet, Path: "/v4/:project/govern/statistics", Func: governService.GetAllServicesStatistics},
	}
}

// GetGraph 获取依赖连接图详细依赖关系
func (governService *ResourceV4) GetGraph(w http.ResponseWriter, r *http.Request) {
	var (
		graph      Graph
		withShared = util.StringTRUE(r.URL.Query().Get("withShared"))
	)
	request := &pb.GetServicesRequest{}
	ctx := r.Context()
	domainProject := util.ParseDomainProject(ctx)

	resp, err := core.ServiceAPI.GetServices(ctx, request)
	if err != nil {
		rest.WriteError(w, pb.ErrInternal, err.Error())
		return
	}
	services := resp.Services
	if len(services) <= 0 {
		return
	}
	nodes := make([]Node, 0, len(services))
	for _, service := range services {
		if governService.isSkipped(withShared, domainProject, service) {
			continue
		}

		var node Node
		node.Name = service.ServiceName
		node.ID = service.ServiceId
		node.AppID = service.AppId
		node.Version = service.Version
		nodes = append(nodes, node)

		proRequest := &pb.GetDependenciesRequest{
			ServiceId:  service.ServiceId,
			SameDomain: true,
			NoSelf:     true,
		}
		proResp, err := core.ServiceAPI.GetConsumerDependencies(ctx, proRequest)
		if err != nil {
			log.Errorf(err, "get service[%s/%s/%s/%s]'s providers failed",
				service.Environment, service.AppId, service.ServiceName, service.Version)
			rest.WriteResponse(w, r, proResp.Response, nil)
			return
		}

		providers := proResp.Providers
		lines := governService.genLinesFromNode(withShared, domainProject, node, providers)
		graph.Lines = append(graph.Lines, lines...)
	}
	graph.Nodes = nodes
	rest.WriteResponse(w, r, nil, graph)
}

func (governService *ResourceV4) genLinesFromNode(withShared bool, domainProject string, node Node, providers []*pb.MicroService) []Line {
	lines := make([]Line, 0)
	for _, child := range providers {
		if child == nil {
			continue
		}

		if node.ID == child.ServiceId {
			continue
		}
		if governService.isSkipped(withShared, domainProject, child) {
			continue
		}
		line := Line{}
		line.From = node
		line.To.Name = child.ServiceName
		line.To.ID = child.ServiceId
		lines = append(lines, line)
	}
	return lines
}

func (governService *ResourceV4) isSkipped(withShared bool, domainProject string, service *pb.MicroService) bool {
	return !withShared && core.IsGlobal(pb.MicroServiceToKey(domainProject, service))
}

// GetServiceDetail 查询服务详细信息
func (governService *ResourceV4) GetServiceDetail(w http.ResponseWriter, r *http.Request) {
	serviceID := r.URL.Query().Get(":serviceId")
	request := &pb.GetServiceRequest{
		ServiceId: serviceID,
	}
	ctx := r.Context()
	resp, _ := ServiceAPI.GetServiceDetail(ctx, request)
	rest.WriteResponse(w, r, resp.Response, resp)
}

func (governService *ResourceV4) GetAllServicesInfo(w http.ResponseWriter, r *http.Request) {
	request := &pb.GetServicesInfoRequest{}
	ctx := r.Context()
	query := r.URL.Query()
	optsStr := query.Get("options")
	request.Options = strings.Split(optsStr, ",")
	request.AppId = query.Get("appId")
	request.ServiceName = query.Get("serviceName")
	request.Environment = query.Get("env")
	request.WithShared = util.StringTRUE(query.Get("withShared"))
	countOnly := query.Get("countOnly")
	if countOnly != "0" && countOnly != "1" && strings.TrimSpace(countOnly) != "" {
		rest.WriteError(w, pb.ErrInvalidParams, "parameter countOnly must be 1 or 0")
		return
	}
	if countOnly == "1" {
		request.CountOnly = true
	}
	resp, _ := ServiceAPI.GetServicesInfo(ctx, request)
	rest.WriteResponse(w, r, resp.Response, resp)
}

func (governService *ResourceV4) GetAllServicesStatistics(w http.ResponseWriter, r *http.Request) {
	request := &pb.GetServicesRequest{}
	ctx := r.Context()
	resp, _ := ServiceAPI.GetServicesStatistics(ctx, request)
	rest.WriteResponse(w, r, resp.Response, resp)
}

func (governService *ResourceV4) GetAllApplications(w http.ResponseWriter, r *http.Request) {
	request := &pb.GetAppsRequest{}
	ctx := r.Context()
	query := r.URL.Query()
	request.Environment = query.Get("env")
	request.WithShared = util.StringTRUE(query.Get("withShared"))
	resp, _ := ServiceAPI.GetApplications(ctx, request)
	rest.WriteResponse(w, r, resp.Response, resp)
}
