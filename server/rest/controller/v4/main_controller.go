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
package v4

import (
	"encoding/json"
	"github.com/ServiceComb/service-center/pkg/rest"
	"github.com/ServiceComb/service-center/pkg/util"
	"github.com/ServiceComb/service-center/server/core"
	pb "github.com/ServiceComb/service-center/server/core/proto"
	scerr "github.com/ServiceComb/service-center/server/error"
	"github.com/ServiceComb/service-center/server/rest/controller"
	"github.com/ServiceComb/service-center/version"
	"net/http"
)

const API_VERSION = "4.0.0"

type Result struct {
	*version.VersionSet
	ApiVersion string           `json:"apiVersion"`
	Config     *pb.ServerConfig `json:"config,omitempty"`
}

type MainService struct {
	//
}

func init() {
}

func (this *MainService) URLPatterns() []rest.Route {
	return []rest.Route{
		{rest.HTTP_METHOD_GET, "/v4/:domain/registry/version", this.GetVersion},
		{rest.HTTP_METHOD_GET, "/v4/:domain/registry/health", this.ClusterHealth},
	}
}

func (this *MainService) ClusterHealth(w http.ResponseWriter, r *http.Request) {
	resp, err := core.InstanceAPI.ClusterHealth(r.Context())
	if err != nil {
		util.Logger().Error("health check failed", err)
		controller.WriteError(w, scerr.ErrInternal, "health check failed")
		return
	}

	respInternal := resp.Response
	resp.Response = nil
	controller.WriteResponse(w, respInternal, resp)
}

func (this *MainService) GetVersion(w http.ResponseWriter, r *http.Request) {
	result := Result{
		version.Ver(),
		API_VERSION,
		core.ServerInfo.Config,
	}
	resultJSON, _ := json.Marshal(result)
	w.Header().Set("Content-Type", "application/json;charset=utf-8")
	w.Write(resultJSON)
}
