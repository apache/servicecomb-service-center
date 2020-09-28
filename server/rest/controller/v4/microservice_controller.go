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
	"github.com/apache/servicecomb-service-center/pkg/log"
	pb "github.com/apache/servicecomb-service-center/pkg/registry"
	"github.com/apache/servicecomb-service-center/pkg/rest"
	"github.com/apache/servicecomb-service-center/pkg/util"
	"github.com/apache/servicecomb-service-center/server/core"
	"github.com/apache/servicecomb-service-center/server/rest/controller"
	scerr "github.com/apache/servicecomb-service-center/server/scerror"
	"io/ioutil"
	"net/http"
)

var trueOrFalse = map[string]bool{"true": true, "false": false, "1": true, "0": false}

type MicroServiceService struct {
	//
}

func (s *MicroServiceService) URLPatterns() []rest.Route {
	return []rest.Route{
		{Method: rest.HTTPMethodGet, Path: "/v4/:project/registry/existence", Func: s.GetExistence},
		{Method: rest.HTTPMethodGet, Path: "/v4/:project/registry/microservices", Func: s.GetServices},
		{Method: rest.HTTPMethodGet, Path: "/v4/:project/registry/microservices/:serviceId", Func: s.GetServiceOne},
		{Method: rest.HTTPMethodPost, Path: "/v4/:project/registry/microservices", Func: s.Register},
		{Method: rest.HTTPMethodPut, Path: "/v4/:project/registry/microservices/:serviceId/properties", Func: s.Update},
		{Method: rest.HTTPMethodDelete, Path: "/v4/:project/registry/microservices/:serviceId", Func: s.Unregister},
		{Method: rest.HTTPMethodDelete, Path: "/v4/:project/registry/microservices", Func: s.UnregisterServices},
	}
}

func (s *MicroServiceService) Register(w http.ResponseWriter, r *http.Request) {
	message, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Error("read body failed", err)
		controller.WriteError(w, scerr.ErrInvalidParams, err.Error())
		return
	}
	var request pb.CreateServiceRequest
	err = json.Unmarshal(message, &request)
	if err != nil {
		log.Errorf(err, "invalid json: %s", util.BytesToStringWithNoCopy(message))
		controller.WriteError(w, scerr.ErrInvalidParams, err.Error())
		return
	}
	resp, err := core.ServiceAPI.Create(r.Context(), &request)
	if err != nil {
		log.Errorf(err, "create service failed")
		controller.WriteError(w, scerr.ErrInternal, err.Error())
		return
	}
	respInternal := resp.Response
	resp.Response = nil
	controller.WriteResponse(w, respInternal, resp)
}

func (s *MicroServiceService) Update(w http.ResponseWriter, r *http.Request) {
	message, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Error("read body failed", err)
		controller.WriteError(w, scerr.ErrInvalidParams, err.Error())
		return
	}
	request := &pb.UpdateServicePropsRequest{
		ServiceId: r.URL.Query().Get(":serviceId"),
	}
	err = json.Unmarshal(message, request)
	if err != nil {
		log.Errorf(err, "invalid json: %s", util.BytesToStringWithNoCopy(message))
		controller.WriteError(w, scerr.ErrInvalidParams, err.Error())
		return
	}
	resp, err := core.ServiceAPI.UpdateProperties(r.Context(), request)
	if err != nil {
		log.Errorf(err, "can not update service")
		controller.WriteError(w, scerr.ErrInternal, "can not update service")
		return
	}
	controller.WriteResponse(w, resp.Response, nil)
}

func (s *MicroServiceService) Unregister(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	serviceID := query.Get(":serviceId")
	force := query.Get("force")

	b, ok := trueOrFalse[force]
	if force != "" && !ok {
		controller.WriteError(w, scerr.ErrInvalidParams, "parameter force must be false or true")
		return
	}

	request := &pb.DeleteServiceRequest{
		ServiceId: serviceID,
		Force:     b,
	}
	resp, err := core.ServiceAPI.Delete(r.Context(), request)
	if err != nil {
		log.Errorf(err, "delete service[%s] failed", serviceID)
		controller.WriteError(w, scerr.ErrInternal, "delete service failed")
		return
	}
	controller.WriteResponse(w, resp.Response, nil)
}

func (s *MicroServiceService) GetServices(w http.ResponseWriter, r *http.Request) {
	request := &pb.GetServicesRequest{}
	resp, err := core.ServiceAPI.GetServices(r.Context(), request)
	if err != nil {
		log.Errorf(err, "get services failed")
		controller.WriteError(w, scerr.ErrInternal, err.Error())
		return
	}
	respInternal := resp.Response
	resp.Response = nil
	controller.WriteResponse(w, respInternal, resp)
}

func (s *MicroServiceService) GetExistence(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	request := &pb.GetExistenceRequest{
		Type:        query.Get("type"),
		Environment: query.Get("env"),
		AppId:       query.Get("appId"),
		ServiceName: query.Get("serviceName"),
		Version:     query.Get("version"),
		ServiceId:   query.Get("serviceId"),
		SchemaId:    query.Get("schemaId"),
	}
	resp, err := core.ServiceAPI.Exist(r.Context(), request)
	if err != nil {
		log.Errorf(err, "check service existence failed")
		controller.WriteError(w, scerr.ErrInternal, "check service existence failed")
		return
	}
	w.Header().Add("X-Schema-Summary", resp.Summary)
	respInternal := resp.Response
	resp.Response = nil
	resp.Summary = ""
	controller.WriteResponse(w, respInternal, resp)
}

func (s *MicroServiceService) GetServiceOne(w http.ResponseWriter, r *http.Request) {
	request := &pb.GetServiceRequest{
		ServiceId: r.URL.Query().Get(":serviceId"),
	}
	resp, err := core.ServiceAPI.GetOne(r.Context(), request)
	if err != nil {
		log.Errorf(err, "get service[%s] failed", request.ServiceId)
		controller.WriteError(w, scerr.ErrInternal, "get service failed")
		return
	}
	respInternal := resp.Response
	resp.Response = nil
	controller.WriteResponse(w, respInternal, resp)
}

func (s *MicroServiceService) UnregisterServices(w http.ResponseWriter, r *http.Request) {
	message, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Error("read body failed", err)
		controller.WriteError(w, scerr.ErrInvalidParams, err.Error())
		return
	}

	request := &pb.DelServicesRequest{}

	err = json.Unmarshal(message, request)
	if err != nil {
		log.Errorf(err, "invalid json: %s", util.BytesToStringWithNoCopy(message))
		controller.WriteError(w, scerr.ErrInvalidParams, err.Error())
		return
	}

	resp, err := core.ServiceAPI.DeleteServices(r.Context(), request)
	if err != nil {
		log.Errorf(err, "delete services failed")
		controller.WriteError(w, scerr.ErrInternal, "delete services failed")
		return
	}
	respInternal := resp.Response
	resp.Response = nil
	controller.WriteResponse(w, respInternal, resp)
}
