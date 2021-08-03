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
	"context"

	"github.com/apache/servicecomb-service-center/datasource"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/util"
	"github.com/apache/servicecomb-service-center/server/core"
	"github.com/go-chassis/cari/discovery"
)

//Node 节点信息
type Node struct {
	ID       string   `json:"id"`
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

func Draw(ctx context.Context, withShared bool) (*Graph, error) {
	var graph Graph

	resp, err := core.ServiceAPI.GetServices(ctx, &discovery.GetServicesRequest{})
	if err != nil {
		return nil, discovery.NewError(discovery.ErrInternal, err.Error())
	}
	services := resp.Services
	if len(services) <= 0 {
		return &graph, nil
	}

	domainProject := util.ParseDomainProject(ctx)
	nodes := make([]Node, 0, len(services))
	for _, service := range services {
		if isSkipped(withShared, domainProject, service) {
			continue
		}

		var node Node
		node.Name = service.ServiceName
		node.ID = service.ServiceId
		node.AppID = service.AppId
		node.Version = service.Version
		nodes = append(nodes, node)

		proRequest := &discovery.GetDependenciesRequest{
			ServiceId:  service.ServiceId,
			SameDomain: true,
			NoSelf:     true,
		}
		proResp, err := core.ServiceAPI.GetConsumerDependencies(ctx, proRequest)
		if err != nil {
			log.Errorf(err, "get service[%s/%s/%s/%s]'s providers failed",
				service.Environment, service.AppId, service.ServiceName, service.Version)
			return nil, err
		}

		providers := proResp.Providers
		lines := genLinesFromNode(withShared, domainProject, node, providers)
		graph.Lines = append(graph.Lines, lines...)
	}
	graph.Nodes = nodes
	return &graph, nil
}

func genLinesFromNode(withShared bool, domainProject string, node Node, providers []*discovery.MicroService) []Line {
	lines := make([]Line, 0)
	for _, child := range providers {
		if child == nil {
			continue
		}

		if node.ID == child.ServiceId {
			continue
		}
		if isSkipped(withShared, domainProject, child) {
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

func isSkipped(withShared bool, domainProject string, service *discovery.MicroService) bool {
	return !withShared && datasource.IsGlobal(discovery.MicroServiceToKey(domainProject, service))
}
