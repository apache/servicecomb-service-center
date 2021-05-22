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
	"io/ioutil"
	"net/http"

	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/rest"
	"github.com/apache/servicecomb-service-center/pkg/util"
	"github.com/apache/servicecomb-service-center/server/core"
	"github.com/apache/servicecomb-service-center/server/rest/controller"
	pb "github.com/go-chassis/cari/discovery"
)

type DependencyService struct {
}

func (s *DependencyService) URLPatterns() []rest.Route {
	return []rest.Route{
		{Method: rest.HTTPMethodPost, Path: "/v4/:project/registry/dependencies", Func: s.AddDependenciesForMicroServices},
		{Method: rest.HTTPMethodPut, Path: "/v4/:project/registry/dependencies", Func: s.CreateDependenciesForMicroServices},
		{Method: rest.HTTPMethodGet, Path: "/v4/:project/registry/microservices/:consumerId/providers", Func: s.GetConProDependencies},
		{Method: rest.HTTPMethodGet, Path: "/v4/:project/registry/microservices/:providerId/consumers", Func: s.GetProConDependencies},
	}
}

//Deprecated
func (s *DependencyService) AddDependenciesForMicroServices(w http.ResponseWriter, r *http.Request) {
	requestBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Error("read body failed", err)
		controller.WriteError(w, pb.ErrInvalidParams, err.Error())
		return
	}
	request := &pb.AddDependenciesRequest{}
	err = json.Unmarshal(requestBody, request)
	if err != nil {
		log.Errorf(err, "invalid json: %s", util.BytesToStringWithNoCopy(requestBody))
		controller.WriteError(w, pb.ErrInvalidParams, err.Error())
		return
	}

	resp, err := core.ServiceAPI.AddDependenciesForMicroServices(r.Context(), request)
	if err != nil {
		controller.WriteError(w, pb.ErrInternal, err.Error())
	}
	w.Header().Add("Deprecation", "version=\"v4\"")
	controller.WriteResponse(w, r, resp.Response, nil)
}

//Deprecated
func (s *DependencyService) CreateDependenciesForMicroServices(w http.ResponseWriter, r *http.Request) {
	requestBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Error("read body failed", err)
		controller.WriteError(w, pb.ErrInvalidParams, err.Error())
		return
	}
	request := &pb.CreateDependenciesRequest{}
	err = json.Unmarshal(requestBody, request)
	if err != nil {
		log.Errorf(err, "invalid json: %s", util.BytesToStringWithNoCopy(requestBody))
		controller.WriteError(w, pb.ErrInvalidParams, err.Error())
		return
	}

	resp, err := core.ServiceAPI.CreateDependenciesForMicroServices(r.Context(), request)
	if err != nil {
		controller.WriteError(w, pb.ErrInternal, err.Error())
	}
	w.Header().Add("Deprecation", "version=\"v4\"")
	controller.WriteResponse(w, r, resp.Response, nil)
}

func (s *DependencyService) GetConProDependencies(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	request := &pb.GetDependenciesRequest{
		ServiceId:  query.Get(":consumerId"),
		SameDomain: query.Get("sameDomain") == "1",
		NoSelf:     query.Get("noSelf") == "1",
	}
	resp, _ := core.ServiceAPI.GetConsumerDependencies(r.Context(), request)
	controller.WriteResponse(w, r, resp.Response, resp)
}

func (s *DependencyService) GetProConDependencies(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	request := &pb.GetDependenciesRequest{
		ServiceId:  query.Get(":providerId"),
		SameDomain: query.Get("sameDomain") == "1",
		NoSelf:     query.Get("noSelf") == "1",
	}
	resp, _ := core.ServiceAPI.GetProviderDependencies(r.Context(), request)
	controller.WriteResponse(w, r, resp.Response, resp)
}
