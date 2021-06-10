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

	discosvc "github.com/apache/servicecomb-service-center/server/service/disco"

	"github.com/apache/servicecomb-service-center/pkg/rest"
	"github.com/apache/servicecomb-service-center/version"
	pb "github.com/go-chassis/cari/discovery"
)

var (
	versionJSONCache []byte
	versionResp      *pb.Response
	parseVersionOnce sync.Once
)

const APIVersion = "4.0.0"

type Result struct {
	*version.Set
	APIVersion string `json:"apiVersion"`
}

type MainService struct {
	//
}

func (s *MainService) URLPatterns() []rest.Route {
	return []rest.Route{
		{Method: http.MethodGet, Path: "/v4/:project/registry/version", Func: s.GetVersion},
		{Method: http.MethodGet, Path: "/v4/:project/registry/health", Func: s.ClusterHealth},
	}
}

func (s *MainService) ClusterHealth(w http.ResponseWriter, r *http.Request) {
	resp, _ := discosvc.ClusterHealth(r.Context())
	rest.WriteResponse(w, r, resp.Response, resp)
}

func (s *MainService) GetVersion(w http.ResponseWriter, r *http.Request) {
	parseVersionOnce.Do(func() {
		result := Result{
			version.Ver(),
			APIVersion,
		}
		versionJSONCache, _ = json.Marshal(result)
		versionResp = pb.CreateResponse(pb.ResponseSuccess, "get version successfully")
	})
	rest.WriteResponse(w, r, versionResp, versionJSONCache)
}
