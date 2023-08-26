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

package disco

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/go-chassis/go-chassis/v2/pkg/codec"

	"github.com/apache/servicecomb-service-center/datasource"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/rest"
	"github.com/apache/servicecomb-service-center/pkg/util"
	discosvc "github.com/apache/servicecomb-service-center/server/service/disco"
	pb "github.com/go-chassis/cari/discovery"
)

type InstanceResource struct {
	//
}

func (s *InstanceResource) URLPatterns() []rest.Route {
	return []rest.Route{
		{Method: http.MethodGet, Path: "/v4/:project/registry/instances", Func: s.FindInstances},
		{Method: http.MethodPost, Path: "/v4/:project/registry/instances/action", Func: s.InstancesAction},
		{Method: http.MethodGet, Path: "/v4/:project/registry/microservices/:serviceId/instances", Func: s.ListInstance},
		{Method: http.MethodGet, Path: "/v4/:project/registry/microservices/:serviceId/instances/:instanceId", Func: s.GetInstance},
		{Method: http.MethodPost, Path: "/v4/:project/registry/microservices/:serviceId/instances",
			Func: s.LegacyRegisterInstance},
		{Method: http.MethodDelete, Path: "/v4/:project/registry/microservices/:serviceId/instances/:instanceId", Func: s.UnregisterInstance},
		{Method: http.MethodPut, Path: "/v4/:project/registry/microservices/:serviceId/instances/:instanceId/properties", Func: s.PutInstanceProperties},
		{Method: http.MethodPut, Path: "/v4/:project/registry/microservices/:serviceId/instances/:instanceId/status", Func: s.PutInstanceStatus},
		{Method: http.MethodPut, Path: "/v4/:project/registry/microservices/:serviceId/instances/:instanceId/heartbeat",
			Func: s.SendHeartbeat},
		{Method: http.MethodPut, Path: "/v4/:project/registry/heartbeats", Func: s.SendManyHeartbeat},
		{Method: http.MethodPut, Path: "/v4/:project/registry/instances/status", Func: s.UpdateManyInstanceStatus},
	}
}
func (s *InstanceResource) LegacyRegisterInstance(w http.ResponseWriter, r *http.Request) {
	message, err := io.ReadAll(r.Body)
	if err != nil {
		log.Error("read body failed", err)
		rest.WriteError(w, pb.ErrInvalidParams, err.Error())
		return
	}

	request := &pb.RegisterInstanceRequest{}
	err = codec.Decode(message, request)
	if err != nil {
		log.Error(fmt.Sprintf("invalid json: %s", util.BytesToStringWithNoCopy(message)), err)
		rest.WriteError(w, pb.ErrInvalidParams, "Unmarshal error")
		return
	}
	if request.Instance != nil {
		request.Instance.ServiceId = r.URL.Query().Get(":serviceId")
	}

	resp, err := discosvc.RegisterInstance(r.Context(), request)
	if err != nil {
		log.Error("register instance failed", err)
		rest.WriteServiceError(w, err)
		return
	}
	rest.WriteResponse(w, r, nil, resp)
}

func (s *InstanceResource) SendHeartbeat(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	request := &pb.HeartbeatRequest{
		ServiceId:  query.Get(":serviceId"),
		InstanceId: query.Get(":instanceId"),
	}
	err := discosvc.SendHeartbeat(r.Context(), request)
	if err != nil {
		rest.WriteServiceError(w, err)
		return
	}
	rest.WriteResponse(w, r, nil, nil)
}

func (s *InstanceResource) SendManyHeartbeat(w http.ResponseWriter, r *http.Request) {
	message, err := io.ReadAll(r.Body)
	if err != nil {
		log.Error("read body failed", err)
		rest.WriteError(w, pb.ErrInvalidParams, err.Error())
		return
	}

	request := &pb.HeartbeatSetRequest{}
	err = codec.Decode(message, request)
	if err != nil {
		log.Error(fmt.Sprintf("invalid json: %s", util.BytesToStringWithNoCopy(message)), err)
		rest.WriteError(w, pb.ErrInvalidParams, "Unmarshal error")
		return
	}
	resp, err := discosvc.SendManyHeartbeat(r.Context(), request)
	if err != nil {
		rest.WriteServiceError(w, err)
		return
	}
	rest.WriteResponse(w, r, nil, resp)
}

func (s *InstanceResource) UnregisterInstance(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	request := &pb.UnregisterInstanceRequest{
		ServiceId:  query.Get(":serviceId"),
		InstanceId: query.Get(":instanceId"),
	}
	err := discosvc.UnregisterInstance(r.Context(), request)
	if err != nil {
		log.Error("unregister instance failed", err)
		rest.WriteServiceError(w, err)
		return
	}
	rest.WriteResponse(w, r, nil, nil)
}

func (s *InstanceResource) FindInstances(w http.ResponseWriter, r *http.Request) {
	var ids []string
	query := r.URL.Query()
	keys := query.Get("tags")
	if len(keys) > 0 {
		ids = strings.Split(keys, ",")
	}
	serviceName := query.Get("serviceName")
	request := &pb.FindInstancesRequest{
		ConsumerServiceId: r.Header.Get("X-ConsumerId"),
		AppId:             query.Get("appId"),
		ServiceName:       serviceName,
		Alias:             serviceName,
		Environment:       query.Get("env"),
		Tags:              ids,
	}

	ctx := util.SetTargetDomainProject(r.Context(), r.Header.Get("X-Domain-Name"), query.Get(":project"))

	resp, err := discosvc.FindInstances(ctx, request)
	if err != nil {
		log.Error("find instances failed", err)
		rest.WriteServiceError(w, err)
		return
	}

	iv, _ := ctx.Value(util.CtxRequestRevision).(string)
	ov, _ := ctx.Value(util.CtxResponseRevision).(string)
	w.Header().Set(util.HeaderRev, ov)
	if len(iv) > 0 && iv == ov {
		w.WriteHeader(http.StatusNotModified)
		return
	}
	rest.WriteResponse(w, r, nil, resp)
}

func (s *InstanceResource) InstancesAction(w http.ResponseWriter, r *http.Request) {
	message, err := io.ReadAll(r.Body)
	if err != nil {
		log.Error("read body failed", err)
		rest.WriteError(w, pb.ErrInvalidParams, err.Error())
		return
	}
	query := r.URL.Query()
	action := query.Get("type")
	switch action {
	case "query":
		findManyInstances(w, r, message)
	default:
		err = fmt.Errorf("Invalid action: %s", action)
		log.Error("invalid request", err)
		rest.WriteError(w, pb.ErrInvalidParams, err.Error())
	}
}

func findManyInstances(w http.ResponseWriter, r *http.Request, body []byte) {
	request := &pb.BatchFindInstancesRequest{}
	err := codec.Decode(body, request)
	if err != nil {
		log.Error(fmt.Sprintf("invalid json: %s", util.BytesToStringWithNoCopy(body)), err)
		rest.WriteError(w, pb.ErrInvalidParams, "Unmarshal error")
		return
	}
	request.ConsumerServiceId = r.Header.Get("X-ConsumerId")

	ctx := util.SetTargetDomainProject(r.Context(), r.Header.Get("X-Domain-Name"), r.URL.Query().Get(":project"))
	resp, err := discosvc.FindManyInstances(ctx, request)
	if err != nil {
		log.Error("find many instances failed", err)
		rest.WriteServiceError(w, err)
		return
	}
	rest.WriteResponse(w, r, resp.Response, resp)
}

func (s *InstanceResource) GetInstance(w http.ResponseWriter, r *http.Request) {
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

	resp, err := discosvc.GetInstance(r.Context(), request)
	if err != nil {
		log.Error("get instance failed", err)
		rest.WriteServiceError(w, err)
		return
	}

	iv, _ := r.Context().Value(util.CtxRequestRevision).(string)
	ov, _ := r.Context().Value(util.CtxResponseRevision).(string)
	w.Header().Set(util.HeaderRev, ov)
	if len(iv) > 0 && iv == ov {
		w.WriteHeader(http.StatusNotModified)
		return
	}
	rest.WriteResponse(w, r, nil, resp)
}

func (s *InstanceResource) ListInstance(w http.ResponseWriter, r *http.Request) {
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
	resp, err := discosvc.ListInstance(r.Context(), request)
	if err != nil {
		log.Error("list instance failed", err)
		rest.WriteServiceError(w, err)
		return
	}

	iv, _ := r.Context().Value(util.CtxRequestRevision).(string)
	ov, _ := r.Context().Value(util.CtxResponseRevision).(string)
	w.Header().Set(util.HeaderRev, ov)
	if len(iv) > 0 && iv == ov {
		w.WriteHeader(http.StatusNotModified)
		return
	}
	rest.WriteResponse(w, r, nil, resp)
}

func (s *InstanceResource) PutInstanceStatus(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	status := query.Get("value")
	request := &pb.UpdateInstanceStatusRequest{
		ServiceId:  query.Get(":serviceId"),
		InstanceId: query.Get(":instanceId"),
		Status:     status,
	}
	err := discosvc.PutInstanceStatus(r.Context(), request)
	if err != nil {
		log.Error("update instance status failed", err)
		rest.WriteServiceError(w, err)
		return
	}
	rest.WriteResponse(w, r, nil, nil)
}

func (s *InstanceResource) PutInstanceProperties(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	message, err := io.ReadAll(r.Body)
	if err != nil {
		log.Error("read body failed", err)
		rest.WriteError(w, pb.ErrInvalidParams, err.Error())
		return
	}
	request := &pb.UpdateInstancePropsRequest{
		ServiceId:  query.Get(":serviceId"),
		InstanceId: query.Get(":instanceId"),
	}
	err = codec.Decode(message, request)
	if err != nil {
		log.Error(fmt.Sprintf("invalid json: %s", util.BytesToStringWithNoCopy(message)), err)
		rest.WriteError(w, pb.ErrInvalidParams, "Unmarshal error")
		return
	}
	err = discosvc.PutInstanceProperties(r.Context(), request)
	if err != nil {
		log.Error("can not update instance properties", err)
		rest.WriteServiceError(w, err)
		return
	}
	rest.WriteResponse(w, r, nil, nil)
}

func (s *InstanceResource) UpdateManyInstanceStatus(w http.ResponseWriter, r *http.Request) {
	request := &UpdateManyInstanceStatusRequest{}
	message, _ := io.ReadAll(r.Body)
	err := codec.Decode(message, request)
	if err != nil {
		log.Error(fmt.Sprintf("invalid json: %s", util.BytesToStringWithNoCopy(message)), err)
		rest.WriteError(w, pb.ErrInvalidParams, "Unmarshal error")
		return
	}
	err = discosvc.UpdateManyInstanceStatus(r.Context(), &request.Matches, request.Status)
	if err != nil {
		log.Error("can not update instance properties", err)
		rest.WriteServiceError(w, err)
		return
	}
	rest.WriteResponse(w, r, nil, nil)
}

type UpdateManyInstanceStatusRequest struct {
	Matches datasource.MatchPolicy `json:"matches,omitempty"`
	Status  string                 `json:"status,omitempty"`
}
