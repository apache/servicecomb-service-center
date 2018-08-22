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
	"github.com/apache/incubator-servicecomb-service-center/pkg/log"
	"github.com/apache/incubator-servicecomb-service-center/pkg/rest"
	"github.com/apache/incubator-servicecomb-service-center/server/core"
	pb "github.com/apache/incubator-servicecomb-service-center/server/core/proto"
	scerr "github.com/apache/incubator-servicecomb-service-center/server/error"
	"github.com/apache/incubator-servicecomb-service-center/server/rest/controller"
	"io/ioutil"
	"net/http"
)

type DependencyService struct {
	//
}

func (this *DependencyService) URLPatterns() []rest.Route {
	return []rest.Route{
		{rest.HTTP_METHOD_POST, "/v4/:project/registry/dependencies", this.AddDependenciesForMicroServices},
		{rest.HTTP_METHOD_PUT, "/v4/:project/registry/dependencies", this.CreateDependenciesForMicroServices},
		{rest.HTTP_METHOD_GET, "/v4/:project/registry/microservices/:consumerId/providers", this.GetConProDependencies},
		{rest.HTTP_METHOD_GET, "/v4/:project/registry/microservices/:providerId/consumers", this.GetProConDependencies},
	}
}

func (this *DependencyService) AddDependenciesForMicroServices(w http.ResponseWriter, r *http.Request) {
	requestBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Error("body err", err)
		controller.WriteError(w, scerr.ErrInvalidParams, err.Error())
		return
	}
	request := &pb.AddDependenciesRequest{}
	err = json.Unmarshal(requestBody, request)
	if err != nil {
		log.Error("Invalid json", err)
		controller.WriteError(w, scerr.ErrInvalidParams, err.Error())
		return
	}

	resp, err := core.ServiceAPI.AddDependenciesForMicroServices(r.Context(), request)
	controller.WriteResponse(w, resp.Response, nil)
}

func (this *DependencyService) CreateDependenciesForMicroServices(w http.ResponseWriter, r *http.Request) {
	requestBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Error("body err", err)
		controller.WriteError(w, scerr.ErrInvalidParams, err.Error())
		return
	}
	request := &pb.CreateDependenciesRequest{}
	err = json.Unmarshal(requestBody, request)
	if err != nil {
		log.Error("Invalid json", err)
		controller.WriteError(w, scerr.ErrInvalidParams, err.Error())
		return
	}

	resp, err := core.ServiceAPI.CreateDependenciesForMicroServices(r.Context(), request)
	controller.WriteResponse(w, resp.Response, nil)
}

func (this *DependencyService) GetConProDependencies(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	request := &pb.GetDependenciesRequest{
		ServiceId:  query.Get(":consumerId"),
		SameDomain: query.Get("sameDomain") == "1",
		NoSelf:     query.Get("noSelf") == "1",
	}
	resp, _ := core.ServiceAPI.GetConsumerDependencies(r.Context(), request)
	respInternal := resp.Response
	resp.Response = nil
	controller.WriteResponse(w, respInternal, resp)
}

func (this *DependencyService) GetProConDependencies(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	request := &pb.GetDependenciesRequest{
		ServiceId:  query.Get(":providerId"),
		SameDomain: query.Get("sameDomain") == "1",
		NoSelf:     query.Get("noSelf") == "1",
	}
	resp, _ := core.ServiceAPI.GetProviderDependencies(r.Context(), request)
	respInternal := resp.Response
	resp.Response = nil
	controller.WriteResponse(w, respInternal, resp)
}
