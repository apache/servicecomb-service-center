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
	"net/http"

	"github.com/apache/servicecomb-service-center/pkg/dump"

	"strings"

	"github.com/apache/servicecomb-service-center/pkg/rest"
)

// Service 治理相关接口服务
type ControllerV4 struct {
}

// URLPatterns 路由
func (ctrl *ControllerV4) URLPatterns() []rest.Route {
	return []rest.Route{
		{Method: http.MethodGet, Path: "/v4/:project/admin/alarms", Func: ctrl.AlarmList},
		{Method: http.MethodDelete, Path: "/v4/:project/admin/alarms", Func: ctrl.ClearAlarm},
		{Method: http.MethodGet, Path: "/v4/:project/admin/dump", Func: ctrl.Dump},
		{Method: http.MethodGet, Path: "/v4/:project/admin/clusters", Func: ctrl.Clusters},
	}
}

func (ctrl *ControllerV4) Dump(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	var options []string
	if s := strings.TrimSpace(query.Get("options")); len(s) > 0 {
		options = strings.Split(s, ",")
	}
	request := &dump.Request{
		Options: options,
	}
	ctx := r.Context()
	resp, _ := AdminServiceAPI.Dump(ctx, request)
	rest.WriteResponse(w, r, resp.Response, resp)
}

func (ctrl *ControllerV4) Clusters(w http.ResponseWriter, r *http.Request) {
	request := &dump.ClustersRequest{}
	ctx := r.Context()
	resp, _ := AdminServiceAPI.Clusters(ctx, request)
	rest.WriteResponse(w, r, resp.Response, resp)
}

func (ctrl *ControllerV4) AlarmList(w http.ResponseWriter, r *http.Request) {
	request := &dump.AlarmListRequest{}
	ctx := r.Context()
	resp, _ := AdminServiceAPI.AlarmList(ctx, request)
	rest.WriteResponse(w, r, resp.Response, resp)
}

func (ctrl *ControllerV4) ClearAlarm(w http.ResponseWriter, r *http.Request) {
	request := &dump.ClearAlarmRequest{}
	ctx := r.Context()
	resp, _ := AdminServiceAPI.ClearAlarm(ctx, request)
	rest.WriteResponse(w, r, resp.Response, nil)
}
