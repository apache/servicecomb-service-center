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
	"strings"

	"github.com/apache/servicecomb-service-center/pkg/log"
	pb "github.com/apache/servicecomb-service-center/pkg/registry"
	"github.com/apache/servicecomb-service-center/pkg/rest"
	"github.com/apache/servicecomb-service-center/pkg/util"
	"github.com/apache/servicecomb-service-center/server/core"
	"github.com/apache/servicecomb-service-center/server/core/proto"
	"github.com/apache/servicecomb-service-center/server/rest/controller"
	scerr "github.com/apache/servicecomb-service-center/server/scerror"
)

type MicroServiceInstanceService struct {
	//
}

func (s *MicroServiceInstanceService) URLPatterns() []rest.Route {
	return []rest.Route{
		{Method: rest.HTTPMethodGet, Path: "/v4/:project/registry/instances", Func: s.FindInstances},
		{Method: rest.HTTPMethodPost, Path: "/v4/:project/registry/instances/action", Func: s.InstancesAction},
		{Method: rest.HTTPMethodGet, Path: "/v4/:project/registry/microservices/:serviceId/instances", Func: s.GetInstances},
		{Method: rest.HTTPMethodGet, Path: "/v4/:project/registry/microservices/:serviceId/instances/:instanceId", Func: s.GetOneInstance},
		{Method: rest.HTTPMethodPost, Path: "/v4/:project/registry/microservices/:serviceId/instances", Func: s.RegisterInstance},
		{Method: rest.HTTPMethodDelete, Path: "/v4/:project/registry/microservices/:serviceId/instances/:instanceId", Func: s.UnregisterInstance},
		{Method: rest.HTTPMethodPut, Path: "/v4/:project/registry/microservices/:serviceId/instances/:instanceId/properties", Func: s.UpdateMetadata},
		{Method: rest.HTTPMethodPut, Path: "/v4/:project/registry/microservices/:serviceId/instances/:instanceId/status", Func: s.UpdateStatus},
		{Method: rest.HTTPMethodPut, Path: "/v4/:project/registry/microservices/:serviceId/instances/:instanceId/heartbeat", Func: s.Heartbeat},
		{Method: rest.HTTPMethodPut, Path: "/v4/:project/registry/heartbeats", Func: s.HeartbeatSet},
	}
}
func (s *MicroServiceInstanceService) RegisterInstance(w http.ResponseWriter, r *http.Request) {
	message, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Error("read body failed", err)
		controller.WriteError(w, scerr.ErrInvalidParams, err.Error())
		return
	}

	request := &pb.RegisterInstanceRequest{}
	err = json.Unmarshal(message, request)
	if err != nil {
		log.Errorf(err, "invalid json: %s", util.BytesToStringWithNoCopy(message))
		controller.WriteError(w, scerr.ErrInvalidParams, "Unmarshal error")
		return
	}
	if request.Instance != nil {
		request.Instance.ServiceId = r.URL.Query().Get(":serviceId")
	}

	resp, err := core.InstanceAPI.Register(r.Context(), request)
	if err != nil {
		log.Errorf(err, "register instance failed")
		controller.WriteError(w, scerr.ErrInternal, "register instance failed")
		return
	}
	respInternal := resp.Response
	resp.Response = nil
	controller.WriteResponse(w, r, respInternal, resp)
}

//TODO 什么样的服务允许更新服务心跳，只能是本服务才可以更新自己，如何屏蔽其他服务伪造的心跳更新？
func (s *MicroServiceInstanceService) Heartbeat(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	request := &pb.HeartbeatRequest{
		ServiceId:  query.Get(":serviceId"),
		InstanceId: query.Get(":instanceId"),
	}
	resp, _ := core.InstanceAPI.Heartbeat(r.Context(), request)
	controller.WriteResponse(w, r, resp.Response, nil)
}

func (s *MicroServiceInstanceService) HeartbeatSet(w http.ResponseWriter, r *http.Request) {
	message, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Error("read body failed", err)
		controller.WriteError(w, scerr.ErrInvalidParams, err.Error())
		return
	}

	request := &pb.HeartbeatSetRequest{}
	err = json.Unmarshal(message, request)
	if err != nil {
		log.Errorf(err, "invalid json: %s", util.BytesToStringWithNoCopy(message))
		controller.WriteError(w, scerr.ErrInvalidParams, "Unmarshal error")
		return
	}
	resp, _ := core.InstanceAPI.HeartbeatSet(r.Context(), request)

	if resp.Response.GetCode() == proto.Response_SUCCESS {
		controller.WriteResponse(w, r, nil, nil)
		return
	}
	respInternal := resp.Response
	resp.Response = nil
	controller.WriteResponse(w, r, respInternal, resp)
}

func (s *MicroServiceInstanceService) UnregisterInstance(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	request := &pb.UnregisterInstanceRequest{
		ServiceId:  query.Get(":serviceId"),
		InstanceId: query.Get(":instanceId"),
	}
	resp, _ := core.InstanceAPI.Unregister(r.Context(), request)
	controller.WriteResponse(w, r, resp.Response, nil)
}

func (s *MicroServiceInstanceService) FindInstances(w http.ResponseWriter, r *http.Request) {
	var ids []string
	query := r.URL.Query()
	keys := query.Get("tags")
	if len(keys) > 0 {
		ids = strings.Split(keys, ",")
	}
	request := &pb.FindInstancesRequest{
		ConsumerServiceId: r.Header.Get("X-ConsumerId"),
		AppId:             query.Get("appId"),
		ServiceName:       query.Get("serviceName"),
		VersionRule:       query.Get("version"),
		Environment:       query.Get("env"),
		Tags:              ids,
	}

	ctx := util.SetTargetDomainProject(r.Context(), r.Header.Get("X-Domain-Name"), query.Get(":project"))

	resp, _ := core.InstanceAPI.Find(ctx, request)
	respInternal := resp.Response
	resp.Response = nil

	iv, _ := ctx.Value(util.CtxRequestRevision).(string)
	ov, _ := ctx.Value(util.CtxResponseRevision).(string)
	w.Header().Set(util.HeaderRev, ov)
	if len(iv) > 0 && iv == ov {
		w.WriteHeader(http.StatusNotModified)
		return
	}

	controller.WriteResponse(w, r, respInternal, resp)
}

func (s *MicroServiceInstanceService) InstancesAction(w http.ResponseWriter, r *http.Request) {
	message, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Error("read body failed", err)
		controller.WriteError(w, scerr.ErrInvalidParams, err.Error())
		return
	}
	query := r.URL.Query()
	action := query.Get("type")
	switch action {
	case "query":
		request := &pb.BatchFindInstancesRequest{}
		err = json.Unmarshal(message, request)
		if err != nil {
			log.Errorf(err, "invalid json: %s", util.BytesToStringWithNoCopy(message))
			controller.WriteError(w, scerr.ErrInvalidParams, "Unmarshal error")
			return
		}
		request.ConsumerServiceId = r.Header.Get("X-ConsumerId")
		ctx := util.SetTargetDomainProject(r.Context(), r.Header.Get("X-Domain-Name"), r.URL.Query().Get(":project"))
		resp, _ := core.InstanceAPI.BatchFind(ctx, request)
		respInternal := resp.Response
		resp.Response = nil
		controller.WriteResponse(w, r, respInternal, resp)
	default:
		err = fmt.Errorf("Invalid action: %s", action)
		log.Errorf(err, "invalid request")
		controller.WriteError(w, scerr.ErrInvalidParams, err.Error())
	}
}

func (s *MicroServiceInstanceService) GetOneInstance(w http.ResponseWriter, r *http.Request) {
	var ids []string
	query := r.URL.Query()
	keys := query.Get("tags")
	if len(keys) > 0 {
		ids = strings.Split(keys, ",")
	}
	request := &pb.GetOneInstanceRequest{
		ConsumerServiceId:  r.Header.Get("X-ConsumerId"),
		ProviderServiceId:  query.Get(":serviceId"),
		ProviderInstanceId: query.Get(":instanceId"),
		Tags:               ids,
	}

	resp, _ := core.InstanceAPI.GetOneInstance(r.Context(), request)
	respInternal := resp.Response
	resp.Response = nil

	iv, _ := r.Context().Value(util.CtxRequestRevision).(string)
	ov, _ := r.Context().Value(util.CtxResponseRevision).(string)
	w.Header().Set(util.HeaderRev, ov)
	if len(iv) > 0 && iv == ov {
		w.WriteHeader(http.StatusNotModified)
		return
	}
	controller.WriteResponse(w, r, respInternal, resp)
}

func (s *MicroServiceInstanceService) GetInstances(w http.ResponseWriter, r *http.Request) {
	var ids []string
	query := r.URL.Query()
	keys := query.Get("tags")
	if len(keys) > 0 {
		ids = strings.Split(keys, ",")
	}
	request := &pb.GetInstancesRequest{
		ConsumerServiceId: r.Header.Get("X-ConsumerId"),
		ProviderServiceId: query.Get(":serviceId"),
		Tags:              ids,
	}
	resp, _ := core.InstanceAPI.GetInstances(r.Context(), request)
	respInternal := resp.Response
	resp.Response = nil

	iv, _ := r.Context().Value(util.CtxRequestRevision).(string)
	ov, _ := r.Context().Value(util.CtxResponseRevision).(string)
	w.Header().Set(util.HeaderRev, ov)
	if len(iv) > 0 && iv == ov {
		w.WriteHeader(http.StatusNotModified)
		return
	}
	controller.WriteResponse(w, r, respInternal, resp)
}

func (s *MicroServiceInstanceService) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	status := query.Get("value")
	request := &pb.UpdateInstanceStatusRequest{
		ServiceId:  query.Get(":serviceId"),
		InstanceId: query.Get(":instanceId"),
		Status:     status,
	}
	resp, _ := core.InstanceAPI.UpdateStatus(r.Context(), request)
	controller.WriteResponse(w, r, resp.Response, nil)
}

func (s *MicroServiceInstanceService) UpdateMetadata(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	message, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Error("read body failed", err)
		controller.WriteError(w, scerr.ErrInvalidParams, err.Error())
		return
	}
	request := &pb.UpdateInstancePropsRequest{
		ServiceId:  query.Get(":serviceId"),
		InstanceId: query.Get(":instanceId"),
	}
	err = json.Unmarshal(message, request)
	if err != nil {
		log.Errorf(err, "invalid json: %s", util.BytesToStringWithNoCopy(message))
		controller.WriteError(w, scerr.ErrInvalidParams, "Unmarshal error")
		return
	}
	resp, err := core.InstanceAPI.UpdateInstanceProperties(r.Context(), request)
	if err != nil {
		log.Errorf(err, "can not update instance")
		controller.WriteError(w, scerr.ErrInternal, "can not update instance")
	}
	controller.WriteResponse(w, r, resp.Response, nil)
}
