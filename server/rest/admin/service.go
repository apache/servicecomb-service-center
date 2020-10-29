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

package admin

import (
	"context"
	"github.com/apache/servicecomb-service-center/datasource"
	"github.com/apache/servicecomb-service-center/pkg/dump"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/registry"
	"github.com/apache/servicecomb-service-center/pkg/util"
	"github.com/apache/servicecomb-service-center/server/alarm"
	"github.com/apache/servicecomb-service-center/server/core"
	scerr "github.com/apache/servicecomb-service-center/server/scerror"
	"github.com/apache/servicecomb-service-center/version"
	"github.com/astaxie/beego"
	"os"
	"strings"
)

var (
	AdminServiceAPI = &Service{}
	configs         map[string]string
	environments    = make(map[string]string)
)

func init() {
	// cache envs
	for _, kv := range os.Environ() {
		arr := strings.Split(kv, "=")
		environments[arr[0]] = arr[1]
	}

	// cache configs
	configs, _ = beego.AppConfig.GetSection("default")
	if section, err := beego.AppConfig.GetSection(beego.BConfig.RunMode); err == nil {
		for k, v := range section {
			configs[k] = v
		}
	}
}

type Service struct {
}

func (service *Service) Dump(ctx context.Context, in *dump.Request) (*dump.Response, error) {
	domainProject := util.ParseDomainProject(ctx)

	if !core.IsDefaultDomainProject(domainProject) {
		return &dump.Response{
			Response: registry.CreateResponse(scerr.ErrForbidden, "Required admin permission"),
		}, nil
	}

	resp := &dump.Response{
		Response: registry.CreateResponse(registry.ResponseSuccess, "Admin dump successfully"),
	}

	if len(in.Options) == 0 {
		service.dump(ctx, "cache", resp)
		return resp, nil
	}

	options := make(map[string]struct{}, len(in.Options))
	for _, option := range in.Options {
		if option == "all" {
			service.dump(ctx, "all", resp)
			return resp, nil
		}
		options[option] = struct{}{}
	}
	for option := range options {
		service.dump(ctx, option, resp)
	}
	return resp, nil
}

func (service *Service) dump(ctx context.Context, option string, resp *dump.Response) {
	switch option {
	case "info":
		resp.Info = version.Ver()
	case "config":
		resp.AppConfig = configs
	case "env":
		resp.Environments = environments
	case "cache":
		var cache dump.Cache
		datasource.Instance().DumpCache(ctx, &cache)
		resp.Cache = &cache
	case "all":
		service.dump(ctx, "info", resp)
		service.dump(ctx, "config", resp)
		service.dump(ctx, "env", resp)
		service.dump(ctx, "cache", resp)
	}
}

func (service *Service) Clusters(ctx context.Context, in *dump.ClustersRequest) (*dump.ClustersResponse, error) {
	clusters, err := datasource.Instance().GetClusters(ctx)
	if err != nil {
		return nil, err
	}
	return &dump.ClustersResponse{
		Clusters: clusters,
	}, nil
}

func (service *Service) AlarmList(ctx context.Context, in *dump.AlarmListRequest) (*dump.AlarmListResponse, error) {
	return &dump.AlarmListResponse{
		Alarms: alarm.ListAll(),
	}, nil
}

func (service *Service) ClearAlarm(ctx context.Context, in *dump.ClearAlarmRequest) (*dump.ClearAlarmResponse, error) {
	alarm.ClearAll()
	log.Infof("service center alarms are cleared")
	return &dump.ClearAlarmResponse{}, nil
}
