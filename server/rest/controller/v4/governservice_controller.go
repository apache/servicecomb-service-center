//Copyright 2017 Huawei Technologies Co., Ltd
//
//Licensed under the Apache License, Version 2.0 (the "License");
//you may not use this file except in compliance with the License.
//You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
//Unless required by applicable law or agreed to in writing, software
//distributed under the License is distributed on an "AS IS" BASIS,
//WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//See the License for the specific language governing permissions and
//limitations under the License.
package v4

import (
	"net/http"

	"github.com/ServiceComb/service-center/pkg/util"

	"github.com/ServiceComb/service-center/pkg/rest"
	"github.com/ServiceComb/service-center/server/core"
	pb "github.com/ServiceComb/service-center/server/core/proto"
	"github.com/ServiceComb/service-center/server/rest/controller"
	"strings"
)

// GovernService 治理相关接口服务
type GovernService struct {
	//
}

// URLPatterns 路由
func (governService *GovernService) URLPatterns() []rest.Route {
	return []rest.Route{
		{rest.HTTP_METHOD_GET, "/v4/:domain/govern/microservices/:serviceId", governService.GetServiceDetail},
		{rest.HTTP_METHOD_GET, "/v4/:domain/govern/relations", governService.GetGraph},
		{rest.HTTP_METHOD_GET, "/v4/:domain/govern/microservices", governService.GetAllServicesInfo},
	}
}

//Node 节点信息
type Node struct {
	Id       string   `json:"id"`
	Name     string   `json:"name"`
	AppID    string   `json:"appId"`
	Version  string   `json:"version"`
	Type     string   `json:"type"`
	Color    string   `json:"color"`
	Position string   `json:"position"`
	Visits   []string `json:"-"`
}

//Line 连接线信息
type Line struct {
	From        Node   `json:"from"`
	To          Node   `json:"to"`
	Type        string `json:"type"`
	Color       string `json:"color"`
	Description string `json:"descriptor"`
}

//Circle 环信息
type Circle struct {
	Nodes []Node `json:"nodes"`
}

//Graph 图全集信息
type Graph struct {
	Nodes   []Node   `json:"nodes"`
	Lines   []Line   `json:"lines"`
	Circles []Circle `json:"circles"`
	Visits  []string `json:"-"`
}

// GetGraph 获取依赖连接图详细依赖关系
func (governService *GovernService) GetGraph(w http.ResponseWriter, r *http.Request) {
	var graph Graph
	noCache := r.URL.Query().Get("noCache")
	if noCache != "0" && noCache != "1" && strings.TrimSpace(noCache) != "" {
		controller.WriteText(http.StatusBadRequest, "parameter noCache must be 1 or 0", w)
		return
	}
	request := &pb.GetServicesRequest{
		NoCache: noCache == "1",
	}
	ctx := r.Context()
	resp, err := core.ServiceAPI.GetServices(ctx, request)
	if err != nil {
		controller.WriteText(http.StatusInternalServerError, err.Error(), w)
		return
	}
	services := resp.GetServices()
	if len(services) <= 0 {
		return
	}
	nodes := make([]Node, len(services))
	for index, service := range services {
		nodes[index].Name = service.ServiceName
		nodes[index].Id = service.ServiceId
		nodes[index].AppID = service.AppId
		nodes[index].Version = service.Version

		proRequest := &pb.GetDependenciesRequest{
			ServiceId: service.ServiceId,
		}
		proResp, err := core.ServiceAPI.GetConsumerDependencies(ctx, proRequest)
		if err != nil {
			util.Logger().Error("get Dependency failed.", err)
			controller.WriteText(http.StatusInternalServerError, "get Dependency failed", w)
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
			line.From = nodes[index]
			line.To.Name = child.ServiceName
			line.To.Id = child.ServiceId
			graph.Lines = append(graph.Lines, line)
		}
	}
	graph.Nodes = nodes
	controller.WriteJsonObject(http.StatusOK, graph, w)
}

// GetServiceDetail 查询服务详细信息
func (governService *GovernService) GetServiceDetail(w http.ResponseWriter, r *http.Request) {
	serviceID := r.URL.Query().Get(":serviceId")
	noCache := r.URL.Query().Get("noCache")
	if noCache != "0" && noCache != "1" && strings.TrimSpace(noCache) != "" {
		controller.WriteText(http.StatusBadRequest, "parameter noCache must be 1 or 0", w)
		return
	}
	request := &pb.GetServiceRequest{
		ServiceId: serviceID,
		NoCache:   noCache == "1",
	}
	ctx := r.Context()
	resp, err := core.GovernServiceAPI.GetServiceDetail(ctx, request)

	respInternal := resp.Response
	resp.Response = nil
	controller.WriteJsonResponse(respInternal, resp, err, w)
}

func (governService *GovernService) GetAllServicesInfo(w http.ResponseWriter, r *http.Request) {
	noCache := r.URL.Query().Get("noCache")
	if noCache != "0" && noCache != "1" && strings.TrimSpace(noCache) != "" {
		controller.WriteText(http.StatusBadRequest, "parameter noCache must be 1 or 0", w)
		return
	}
	request := &pb.GetServicesInfoRequest{
		NoCache: noCache == "1",
	}
	ctx := r.Context()
	optsStr := r.URL.Query().Get("options")
	request.Options = strings.Split(optsStr, ",")
	resp, err := core.GovernServiceAPI.GetServicesInfo(ctx, request)

	respInternal := resp.Response
	resp.Response = nil
	controller.WriteJsonResponse(respInternal, resp, err, w)
}
