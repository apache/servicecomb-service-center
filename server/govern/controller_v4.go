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

	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/rest"
	"github.com/apache/servicecomb-service-center/pkg/util"
	"github.com/apache/servicecomb-service-center/server/core"
	pb "github.com/apache/servicecomb-service-center/server/core/proto"
	"github.com/apache/servicecomb-service-center/server/rest/controller"
	scerr "github.com/apache/servicecomb-service-center/server/scerror"
	"strings"
)

// GovernService 治理相关接口服务
type GovernServiceControllerV4 struct {
	//
}

// URLPatterns 路由
func (governService *GovernServiceControllerV4) URLPatterns() []rest.Route {
	return []rest.Route{
		{rest.HTTP_METHOD_GET, "/v4/:project/govern/microservices/:serviceId", governService.GetServiceDetail},
		{rest.HTTP_METHOD_GET, "/v4/:project/govern/relations", governService.GetGraph},
		{rest.HTTP_METHOD_GET, "/v4/:project/govern/microservices", governService.GetAllServicesInfo},
		{rest.HTTP_METHOD_GET, "/v4/:project/govern/apps", governService.GetAllApplications},
	}
}

// GetGraph 获取依赖连接图详细依赖关系
func (governService *GovernServiceControllerV4) GetGraph(w http.ResponseWriter, r *http.Request) {
	var (
		graph      Graph
		withShared = util.StringTRUE(r.URL.Query().Get("withShared"))
	)
	request := &pb.GetServicesRequest{}
	ctx := r.Context()
	domainProject := util.ParseDomainProject(ctx)

	resp, err := core.ServiceAPI.GetServices(ctx, request)
	if err != nil {
		controller.WriteError(w, scerr.ErrInternal, err.Error())
		return
	}
	services := resp.GetServices()
	if len(services) <= 0 {
		return
	}
	nodes := make([]Node, 0, len(services))
	for _, service := range services {
		if !withShared && core.IsShared(pb.MicroServiceToKey(domainProject, service)) {
			continue
		}

		var node Node
		node.Name = service.ServiceName
		node.Id = service.ServiceId
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
			controller.WriteResponse(w, proResp.Response, nil)
			return
		}

		providers := proResp.Providers
		countInner := len(providers)
		if countInner <= 0 {
			continue
		}
		for _, child := range providers {
			if child == nil {
				continue
			}

			if service.ServiceId == child.ServiceId {
				continue
			}
			line := Line{}
			line.From = node
			line.To.Name = child.ServiceName
			line.To.Id = child.ServiceId
			graph.Lines = append(graph.Lines, line)
		}
	}
	graph.Nodes = nodes
	controller.WriteResponse(w, nil, graph)
}

// GetServiceDetail 查询服务详细信息
func (governService *GovernServiceControllerV4) GetServiceDetail(w http.ResponseWriter, r *http.Request) {
	serviceID := r.URL.Query().Get(":serviceId")
	request := &pb.GetServiceRequest{
		ServiceId: serviceID,
	}
	ctx := r.Context()
	resp, _ := GovernServiceAPI.GetServiceDetail(ctx, request)

	respInternal := resp.Response
	resp.Response = nil
	controller.WriteResponse(w, respInternal, resp)
}

func (governService *GovernServiceControllerV4) GetAllServicesInfo(w http.ResponseWriter, r *http.Request) {
	request := &pb.GetServicesInfoRequest{}
	ctx := r.Context()
	query := r.URL.Query()
	optsStr := query.Get("options")
	request.Options = strings.Split(optsStr, ",")
	request.AppId = query.Get("appId")
	request.ServiceName = query.Get("serviceName")
	request.WithShared = util.StringTRUE(query.Get("withShared"))
	countOnly := query.Get("countOnly")
	if countOnly != "0" && countOnly != "1" && strings.TrimSpace(countOnly) != "" {
		controller.WriteError(w, scerr.ErrInvalidParams, "parameter countOnly must be 1 or 0")
		return
	}
	if countOnly == "1" {
		request.CountOnly = true
	}
	resp, _ := GovernServiceAPI.GetServicesInfo(ctx, request)

	respInternal := resp.Response
	resp.Response = nil
	controller.WriteResponse(w, respInternal, resp)
}

func (governService *GovernServiceControllerV4) GetAllApplications(w http.ResponseWriter, r *http.Request) {
	request := &pb.GetAppsRequest{}
	ctx := r.Context()
	query := r.URL.Query()
	request.Environment = query.Get("env")
	request.WithShared = util.StringTRUE(query.Get("withShared"))
	resp, _ := GovernServiceAPI.GetApplications(ctx, request)

	respInternal := resp.Response
	resp.Response = nil
	controller.WriteResponse(w, respInternal, resp)
}
