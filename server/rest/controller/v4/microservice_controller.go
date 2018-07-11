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
	"github.com/apache/incubator-servicecomb-service-center/pkg/rest"
	"github.com/apache/incubator-servicecomb-service-center/pkg/util"
	"github.com/apache/incubator-servicecomb-service-center/server/core"
	pb "github.com/apache/incubator-servicecomb-service-center/server/core/proto"
	scerr "github.com/apache/incubator-servicecomb-service-center/server/error"
	"github.com/apache/incubator-servicecomb-service-center/server/rest/controller"
	"io/ioutil"
	"net/http"
)

var trueOrFalse = map[string]bool{"true": true, "false": false, "1": true, "0": false}

type MicroServiceService struct {
	//
}

func (this *MicroServiceService) URLPatterns() []rest.Route {
	return []rest.Route{
		{rest.HTTP_METHOD_GET, "/v4/:project/registry/existence", this.GetExistence},
		{rest.HTTP_METHOD_GET, "/v4/:project/registry/microservices", this.GetServices},
		{rest.HTTP_METHOD_GET, "/v4/:project/registry/microservices/:serviceId", this.GetServiceOne},
		{rest.HTTP_METHOD_POST, "/v4/:project/registry/microservices", this.Register},
		{rest.HTTP_METHOD_PUT, "/v4/:project/registry/microservices/:serviceId/properties", this.Update},
		{rest.HTTP_METHOD_DELETE, "/v4/:project/registry/microservices/:serviceId", this.Unregister},
		{rest.HTTP_METHOD_DELETE, "/v4/:project/registry/microservices", this.UnregisterServices},
	}
}

func (this *MicroServiceService) Register(w http.ResponseWriter, r *http.Request) {
	message, err := ioutil.ReadAll(r.Body)
	if err != nil {
		util.Logger().Error("body err", err)
		controller.WriteError(w, scerr.ErrInvalidParams, err.Error())
		return
	}
	var request pb.CreateServiceRequest
	err = json.Unmarshal(message, &request)
	if err != nil {
		util.Logger().Error("Unmarshal error", err)
		controller.WriteError(w, scerr.ErrInvalidParams, err.Error())
		return
	}
	resp, err := core.ServiceAPI.Create(r.Context(), &request)
	respInternal := resp.Response
	resp.Response = nil
	controller.WriteResponse(w, respInternal, resp)
}

func (this *MicroServiceService) Update(w http.ResponseWriter, r *http.Request) {
	message, err := ioutil.ReadAll(r.Body)
	if err != nil {
		util.Logger().Error("body err", err)
		controller.WriteError(w, scerr.ErrInvalidParams, err.Error())
		return
	}
	request := &pb.UpdateServicePropsRequest{
		ServiceId: r.URL.Query().Get(":serviceId"),
	}
	err = json.Unmarshal(message, request)
	if err != nil {
		util.Logger().Error("Unmarshal error", err)
		controller.WriteError(w, scerr.ErrInvalidParams, err.Error())
		return
	}
	resp, err := core.ServiceAPI.UpdateProperties(r.Context(), request)
	controller.WriteResponse(w, resp.Response, nil)
}

func (this *MicroServiceService) Unregister(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	serviceId := query.Get(":serviceId")
	force := query.Get("force")

	b, ok := trueOrFalse[force]
	if force != "" && !ok {
		controller.WriteError(w, scerr.ErrInvalidParams, "parameter force must be false or true")
		return
	}

	request := &pb.DeleteServiceRequest{
		ServiceId: serviceId,
		Force:     b,
	}
	resp, _ := core.ServiceAPI.Delete(r.Context(), request)
	controller.WriteResponse(w, resp.Response, nil)
}

func (this *MicroServiceService) GetServices(w http.ResponseWriter, r *http.Request) {
	request := &pb.GetServicesRequest{}
	resp, _ := core.ServiceAPI.GetServices(r.Context(), request)
	respInternal := resp.Response
	resp.Response = nil
	controller.WriteResponse(w, respInternal, resp)
}

func (this *MicroServiceService) GetExistence(w http.ResponseWriter, r *http.Request) {
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
	resp, _ := core.ServiceAPI.Exist(r.Context(), request)
	w.Header().Add("X-Schema-Summary", resp.Summary)
	respInternal := resp.Response
	resp.Response = nil
	resp.Summary = ""
	controller.WriteResponse(w, respInternal, resp)
}

func (this *MicroServiceService) GetServiceOne(w http.ResponseWriter, r *http.Request) {
	request := &pb.GetServiceRequest{
		ServiceId: r.URL.Query().Get(":serviceId"),
	}
	resp, _ := core.ServiceAPI.GetOne(r.Context(), request)
	respInternal := resp.Response
	resp.Response = nil
	controller.WriteResponse(w, respInternal, resp)
}

func (this *MicroServiceService) UnregisterServices(w http.ResponseWriter, r *http.Request) {
	request_body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		util.Logger().Error("body ,err", err)
		controller.WriteError(w, scerr.ErrInvalidParams, err.Error())
		return
	}

	request := &pb.DelServicesRequest{}

	err = json.Unmarshal(request_body, request)
	if err != nil {
		util.Logger().Error("unmarshal ,err ", err)
		controller.WriteError(w, scerr.ErrInvalidParams, err.Error())
		return
	}

	resp, err := core.ServiceAPI.DeleteServices(r.Context(), request)
	respInternal := resp.Response
	resp.Response = nil
	controller.WriteResponse(w, respInternal, resp)
}
