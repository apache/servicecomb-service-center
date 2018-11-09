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

	"github.com/apache/servicecomb-service-center/pkg/rest"
	"github.com/apache/servicecomb-service-center/server/admin/model"
	"github.com/apache/servicecomb-service-center/server/rest/controller"
)

// AdminService 治理相关接口服务
type AdminServiceControllerV4 struct {
}

// URLPatterns 路由
func (ctrl *AdminServiceControllerV4) URLPatterns() []rest.Route {
	return []rest.Route{
		{rest.HTTP_METHOD_GET, "/v4/:project/admin/dump", ctrl.Dump},
		{rest.HTTP_METHOD_GET, "/v4/:project/admin/clusters", ctrl.Clusters},
	}
}

func (ctrl *AdminServiceControllerV4) Dump(w http.ResponseWriter, r *http.Request) {
	request := &model.DumpRequest{}
	ctx := r.Context()
	resp, _ := AdminServiceAPI.Dump(ctx, request)

	respInternal := resp.Response
	resp.Response = nil
	controller.WriteResponse(w, respInternal, resp)
}

func (ctrl *AdminServiceControllerV4) Clusters(w http.ResponseWriter, r *http.Request) {
	request := &model.ClustersRequest{}
	ctx := r.Context()
	resp, _ := AdminServiceAPI.Clusters(ctx, request)

	respInternal := resp.Response
	resp.Response = nil
	controller.WriteResponse(w, respInternal, resp)
}
