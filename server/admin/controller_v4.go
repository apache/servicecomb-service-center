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
	model2 "github.com/apache/servicecomb-service-center/pkg/model"
	"net/http"

	"github.com/apache/servicecomb-service-center/pkg/rest"
	"github.com/apache/servicecomb-service-center/server/rest/controller"
	"strings"
)

// AdminService 治理相关接口服务
type AdminServiceControllerV4 struct {
}

// URLPatterns 路由
func (ctrl *AdminServiceControllerV4) URLPatterns() []rest.Route {
	return []rest.Route{
		{Method: rest.HTTP_METHOD_GET, Path: "/v4/:project/admin/alarms", Func: ctrl.AlarmList},
		{Method: rest.HTTP_METHOD_DELETE, Path: "/v4/:project/admin/alarms", Func: ctrl.ClearAlarm},
		{Method: rest.HTTP_METHOD_GET, Path: "/v4/:project/admin/dump", Func: ctrl.Dump},
		{Method: rest.HTTP_METHOD_GET, Path: "/v4/:project/admin/clusters", Func: ctrl.Clusters},
	}
}

func (ctrl *AdminServiceControllerV4) Dump(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	var options []string
	if s := strings.TrimSpace(query.Get("options")); len(s) > 0 {
		options = strings.Split(s, ",")
	}
	request := &model2.DumpRequest{
		Options: options,
	}
	ctx := r.Context()
	resp, _ := AdminServiceAPI.Dump(ctx, request)

	respInternal := resp.Response
	resp.Response = nil
	controller.WriteResponse(w, respInternal, resp)
}

func (ctrl *AdminServiceControllerV4) Clusters(w http.ResponseWriter, r *http.Request) {
	request := &model2.ClustersRequest{}
	ctx := r.Context()
	resp, _ := AdminServiceAPI.Clusters(ctx, request)

	respInternal := resp.Response
	resp.Response = nil
	controller.WriteResponse(w, respInternal, resp)
}

func (ctrl *AdminServiceControllerV4) AlarmList(w http.ResponseWriter, r *http.Request) {
	request := &model2.AlarmListRequest{}
	ctx := r.Context()
	resp, _ := AdminServiceAPI.AlarmList(ctx, request)

	respInternal := resp.Response
	resp.Response = nil
	controller.WriteResponse(w, respInternal, resp)
}

func (ctrl *AdminServiceControllerV4) ClearAlarm(w http.ResponseWriter, r *http.Request) {
	request := &model2.ClearAlarmRequest{}
	ctx := r.Context()
	resp, _ := AdminServiceAPI.ClearAlarm(ctx, request)
	controller.WriteResponse(w, resp.Response, nil)
}
