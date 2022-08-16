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
	"fmt"
	"io"
	"net/http"

	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/rest"
	"github.com/apache/servicecomb-service-center/pkg/util"
	discosvc "github.com/apache/servicecomb-service-center/server/service/disco"
	pb "github.com/go-chassis/cari/discovery"
)

type DependencyService struct {
}

func (s *DependencyService) URLPatterns() []rest.Route {
	return []rest.Route{
		{Method: http.MethodPost, Path: "/v4/:project/registry/dependencies", Func: s.AddDependencies},
		{Method: http.MethodPut, Path: "/v4/:project/registry/dependencies", Func: s.PutDependencies},
		{Method: http.MethodGet, Path: "/v4/:project/registry/microservices/:consumerId/providers", Func: s.ListProviders},
		{Method: http.MethodGet, Path: "/v4/:project/registry/microservices/:providerId/consumers", Func: s.ListConsumers},
	}
}

// Deprecated
func (s *DependencyService) AddDependencies(w http.ResponseWriter, r *http.Request) {
	requestBody, err := io.ReadAll(r.Body)
	if err != nil {
		log.Error("read body failed", err)
		rest.WriteError(w, pb.ErrInvalidParams, err.Error())
		return
	}
	request := &pb.AddDependenciesRequest{}
	err = json.Unmarshal(requestBody, request)
	if err != nil {
		log.Error(fmt.Sprintf("invalid json: %s", util.BytesToStringWithNoCopy(requestBody)), err)
		rest.WriteError(w, pb.ErrInvalidParams, err.Error())
		return
	}

	err = discosvc.AddDependencies(r.Context(), request)
	if err != nil {
		rest.WriteServiceError(w, err)
		return
	}
	w.Header().Add("Deprecation", "version=\"v4\"")
	rest.WriteResponse(w, r, nil, nil)
}

// Deprecated
func (s *DependencyService) PutDependencies(w http.ResponseWriter, r *http.Request) {
	requestBody, err := io.ReadAll(r.Body)
	if err != nil {
		log.Error("read body failed", err)
		rest.WriteError(w, pb.ErrInvalidParams, err.Error())
		return
	}
	request := &pb.CreateDependenciesRequest{}
	err = json.Unmarshal(requestBody, request)
	if err != nil {
		log.Error(fmt.Sprintf("invalid json: %s", util.BytesToStringWithNoCopy(requestBody)), err)
		rest.WriteError(w, pb.ErrInvalidParams, err.Error())
		return
	}

	err = discosvc.PutDependencies(r.Context(), request)
	if err != nil {
		rest.WriteServiceError(w, err)
		return
	}
	w.Header().Add("Deprecation", "version=\"v4\"")
	rest.WriteResponse(w, r, nil, nil)
}

func (s *DependencyService) ListProviders(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	request := &pb.GetDependenciesRequest{
		ServiceId:  query.Get(":consumerId"),
		SameDomain: query.Get("sameDomain") == "1",
		NoSelf:     query.Get("noSelf") == "1",
	}
	resp, err := discosvc.ListProviders(r.Context(), request)
	if err != nil {
		rest.WriteServiceError(w, err)
		return
	}
	rest.WriteResponse(w, r, nil, resp)
}

func (s *DependencyService) ListConsumers(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	request := &pb.GetDependenciesRequest{
		ServiceId:  query.Get(":providerId"),
		SameDomain: query.Get("sameDomain") == "1",
		NoSelf:     query.Get("noSelf") == "1",
	}
	resp, err := discosvc.ListConsumers(r.Context(), request)
	if err != nil {
		rest.WriteServiceError(w, err)
		return
	}
	rest.WriteResponse(w, r, nil, resp)
}
