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
	"net/http"
	"sync"

	pb "github.com/apache/servicecomb-service-center/pkg/registry"
	"github.com/apache/servicecomb-service-center/pkg/rest"
	"github.com/apache/servicecomb-service-center/server/core"
	"github.com/apache/servicecomb-service-center/server/core/proto"
	"github.com/apache/servicecomb-service-center/server/rest/controller"
	"github.com/apache/servicecomb-service-center/version"
)

var (
	versionJSONCache []byte
	versionResp      *pb.Response
	parseVersionOnce sync.Once
)

const APIVersion = "4.0.0"

type Result struct {
	*version.Set
	APIVersion string              `json:"apiVersion"`
	Config     *proto.ServerConfig `json:"config,omitempty"`
}

type MainService struct {
	//
}

func (s *MainService) URLPatterns() []rest.Route {
	return []rest.Route{
		{Method: rest.HTTPMethodGet, Path: "/v4/:project/registry/version", Func: s.GetVersion},
		{Method: rest.HTTPMethodGet, Path: "/v4/:project/registry/health", Func: s.ClusterHealth},
	}
}

func (s *MainService) ClusterHealth(w http.ResponseWriter, r *http.Request) {
	resp, _ := core.InstanceAPI.ClusterHealth(r.Context())
	respInternal := resp.Response
	resp.Response = nil
	controller.WriteResponse(w, r, respInternal, resp)
}

func (s *MainService) GetVersion(w http.ResponseWriter, r *http.Request) {
	parseVersionOnce.Do(func() {
		result := Result{
			version.Ver(),
			APIVersion,
			&core.ServerInfo.Config,
		}
		versionJSONCache, _ = json.Marshal(result)
		versionResp = proto.CreateResponse(proto.Response_SUCCESS, "get version successfully")
	})
	controller.WriteJSONIfSuccess(w, versionResp, versionJSONCache)
}
