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

package sys

import (
	"encoding/json"
	"net/http"
	"sync"

	"github.com/apache/servicecomb-service-center/pkg/rest"
	discosvc "github.com/apache/servicecomb-service-center/server/service/disco"
	"github.com/apache/servicecomb-service-center/version"
)

var (
	versionJSONCache []byte
	parseVersionOnce sync.Once
)

const APIVersion = "4.0.0"

type VersionResponse struct {
	*version.Set
	APIVersion string `json:"apiVersion"`
}

type Resource struct {
	//
}

func (s *Resource) URLPatterns() []rest.Route {
	return []rest.Route{
		{Method: http.MethodGet, Path: "/v4/:project/registry/version", Func: s.GetVersion},
		{Method: http.MethodGet, Path: "/v4/:project/registry/health", Func: s.ClusterHealth},
	}
}

func (s *Resource) ClusterHealth(w http.ResponseWriter, r *http.Request) {
	resp, err := discosvc.ClusterHealth(r.Context())
	if err != nil {
		rest.WriteServiceError(w, err)
		return
	}
	rest.WriteResponse(w, r, nil, resp)
}

func (s *Resource) GetVersion(w http.ResponseWriter, r *http.Request) {
	parseVersionOnce.Do(func() {
		result := VersionResponse{
			version.Ver(),
			APIVersion,
		}
		versionJSONCache, _ = json.Marshal(result)
	})
	rest.WriteResponse(w, r, nil, versionJSONCache)
}
