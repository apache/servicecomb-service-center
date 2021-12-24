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
	"io/ioutil"
	"net/http"

	"github.com/apache/servicecomb-service-center/datasource"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/rest"
	"github.com/apache/servicecomb-service-center/pkg/util"
	"github.com/apache/servicecomb-service-center/server/core"
	discosvc "github.com/apache/servicecomb-service-center/server/service/disco"
	pb "github.com/go-chassis/cari/discovery"
)

var trueOrFalse = map[string]bool{"true": true, "false": false, "1": true, "0": false}

type MicroServiceService struct {
	//
}

func (s *MicroServiceService) URLPatterns() []rest.Route {
	return []rest.Route{
		{Method: http.MethodGet, Path: "/v4/:project/registry/existence", Func: s.GetExistence},
		{Method: http.MethodGet, Path: "/v4/:project/registry/microservices", Func: s.GetServices},
		{Method: http.MethodGet, Path: "/v4/:project/registry/microservices/:serviceId", Func: s.GetServiceOne},
		{Method: http.MethodPost, Path: "/v4/:project/registry/microservices", Func: s.Register},
		{Method: http.MethodPut, Path: "/v4/:project/registry/microservices/:serviceId/properties", Func: s.Update},
		{Method: http.MethodDelete, Path: "/v4/:project/registry/microservices/:serviceId", Func: s.Unregister},
		{Method: http.MethodDelete, Path: "/v4/:project/registry/microservices", Func: s.UnregisterServices},
	}
}

func (s *MicroServiceService) Register(w http.ResponseWriter, r *http.Request) {
	message, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Error("read body failed", err)
		rest.WriteError(w, pb.ErrInvalidParams, err.Error())
		return
	}
	var request pb.CreateServiceRequest
	err = json.Unmarshal(message, &request)
	if err != nil {
		log.Error(fmt.Sprintf("invalid json: %s", util.BytesToStringWithNoCopy(message)), err)
		rest.WriteError(w, pb.ErrInvalidParams, err.Error())
		return
	}
	resp, err := core.ServiceAPI.Create(r.Context(), &request)
	if err != nil {
		log.Error("create service failed", err)
		rest.WriteError(w, pb.ErrInternal, err.Error())
		return
	}
	rest.WriteResponse(w, r, resp.Response, resp)
}

func (s *MicroServiceService) Update(w http.ResponseWriter, r *http.Request) {
	message, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Error("read body failed", err)
		rest.WriteError(w, pb.ErrInvalidParams, err.Error())
		return
	}
	request := &pb.UpdateServicePropsRequest{
		ServiceId: r.URL.Query().Get(":serviceId"),
	}
	err = json.Unmarshal(message, request)
	if err != nil {
		log.Error(fmt.Sprintf("invalid json: %s", util.BytesToStringWithNoCopy(message)), err)
		rest.WriteError(w, pb.ErrInvalidParams, err.Error())
		return
	}
	resp, err := core.ServiceAPI.UpdateProperties(r.Context(), request)
	if err != nil {
		log.Error("can not update service", err)
		rest.WriteError(w, pb.ErrInternal, "can not update service")
		return
	}
	rest.WriteResponse(w, r, resp.Response, nil)
}

func (s *MicroServiceService) Unregister(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	serviceID := query.Get(":serviceId")
	force := query.Get("force")

	b, ok := trueOrFalse[force]
	if force != "" && !ok {
		rest.WriteError(w, pb.ErrInvalidParams, "parameter force must be false or true")
		return
	}

	request := &pb.DeleteServiceRequest{
		ServiceId: serviceID,
		Force:     b,
	}
	resp, err := core.ServiceAPI.Delete(r.Context(), request)
	if err != nil {
		log.Error(fmt.Sprintf("delete service[%s] failed", serviceID), err)
		rest.WriteError(w, pb.ErrInternal, "delete service failed")
		return
	}
	rest.WriteResponse(w, r, resp.Response, nil)
}

func (s *MicroServiceService) GetServices(w http.ResponseWriter, r *http.Request) {
	request := &pb.GetServicesRequest{
		WithShared: util.StringTRUE(r.URL.Query().Get("withShared")),
	}
	resp, err := core.ServiceAPI.GetServices(r.Context(), request)
	if err != nil {
		log.Error("get services failed", err)
		rest.WriteError(w, pb.ErrInternal, err.Error())
		return
	}
	rest.WriteResponse(w, r, resp.Response, resp)
}

func (s *MicroServiceService) GetExistence(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	checkType := query.Get("type")
	request := &pb.GetExistenceRequest{
		Type:        checkType,
		Environment: query.Get("env"),
		AppId:       query.Get("appId"),
		ServiceName: query.Get("serviceName"),
		Version:     query.Get("version"),
		ServiceId:   query.Get("serviceId"),
		SchemaId:    query.Get("schemaId"),
	}
	resp, err := core.ServiceAPI.Exist(r.Context(), request)
	if err != nil {
		log.Error("check resource existence failed", err)
		rest.WriteServiceError(w, err)
		return
	}
	if checkType == datasource.ExistTypeSchema {
		w.Header().Add("X-Schema-Summary", resp.Summary)
		resp.Summary = ""
	}
	rest.WriteResponse(w, r, resp.Response, resp)
}

func (s *MicroServiceService) GetServiceOne(w http.ResponseWriter, r *http.Request) {
	request := &pb.GetServiceRequest{
		ServiceId: r.URL.Query().Get(":serviceId"),
	}
	service, err := discosvc.GetService(r.Context(), request)
	if err != nil {
		log.Error(fmt.Sprintf("get service[%s] failed", request.ServiceId), err)
		rest.WriteServiceError(w, err)
		return
	}
	rest.WriteResponse(w, r, nil, &pb.GetServiceResponse{Service: service})
}

func (s *MicroServiceService) UnregisterServices(w http.ResponseWriter, r *http.Request) {
	message, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Error("read body failed", err)
		rest.WriteError(w, pb.ErrInvalidParams, err.Error())
		return
	}

	request := &pb.DelServicesRequest{}

	err = json.Unmarshal(message, request)
	if err != nil {
		log.Error(fmt.Sprintf("invalid json: %s", util.BytesToStringWithNoCopy(message)), err)
		rest.WriteError(w, pb.ErrInvalidParams, err.Error())
		return
	}

	resp, err := core.ServiceAPI.DeleteServices(r.Context(), request)
	if err != nil {
		log.Error("delete services failed", err)
		rest.WriteError(w, pb.ErrInternal, "delete services failed")
		return
	}
	rest.WriteResponse(w, r, resp.Response, resp)
}
