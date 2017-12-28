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
	"strings"
)

type MicroServiceInstanceService struct {
	//
}

func (this *MicroServiceInstanceService) URLPatterns() []rest.Route {
	return []rest.Route{
		{rest.HTTP_METHOD_GET, "/v4/:project/registry/instances", this.FindInstances},
		{rest.HTTP_METHOD_GET, "/v4/:project/registry/microservices/:serviceId/instances", this.GetInstances},
		{rest.HTTP_METHOD_GET, "/v4/:project/registry/microservices/:serviceId/instances/:instanceId", this.GetOneInstance},
		{rest.HTTP_METHOD_POST, "/v4/:project/registry/microservices/:serviceId/instances", this.RegisterInstance},
		{rest.HTTP_METHOD_DELETE, "/v4/:project/registry/microservices/:serviceId/instances/:instanceId", this.UnregisterInstance},
		{rest.HTTP_METHOD_PUT, "/v4/:project/registry/microservices/:serviceId/instances/:instanceId/properties", this.UpdateMetadata},
		{rest.HTTP_METHOD_PUT, "/v4/:project/registry/microservices/:serviceId/instances/:instanceId/status", this.UpdateStatus},
		{rest.HTTP_METHOD_PUT, "/v4/:project/registry/microservices/:serviceId/instances/:instanceId/heartbeat", this.Heartbeat},
		{rest.HTTP_METHOD_PUT, "/v4/:project/registry/heartbeats", this.HeartbeatSet},
	}
}
func (this *MicroServiceInstanceService) RegisterInstance(w http.ResponseWriter, r *http.Request) {
	message, err := ioutil.ReadAll(r.Body)
	if err != nil {
		util.Logger().Error("register instance failed, body err", err)
		controller.WriteError(w, scerr.ErrInvalidParams, err.Error())
		return
	}

	request := &pb.RegisterInstanceRequest{}
	err = json.Unmarshal(message, request)
	if err != nil {
		util.Logger().Error("register instance failed, Unmarshal error", err)
		controller.WriteError(w, scerr.ErrInvalidParams, "Unmarshal error")
		return
	}
	if request.GetInstance() != nil {
		request.Instance.ServiceId = r.URL.Query().Get(":serviceId")
	}

	resp, err := core.InstanceAPI.Register(r.Context(), request)
	respInternal := resp.Response
	resp.Response = nil
	controller.WriteResponse(w, respInternal, resp)
}

//TODO 什么样的服务允许更新服务心跳，只能是本服务才可以更新自己，如何屏蔽其他服务伪造的心跳更新？
func (this *MicroServiceInstanceService) Heartbeat(w http.ResponseWriter, r *http.Request) {
	request := &pb.HeartbeatRequest{
		ServiceId:  r.URL.Query().Get(":serviceId"),
		InstanceId: r.URL.Query().Get(":instanceId"),
	}
	resp, _ := core.InstanceAPI.Heartbeat(r.Context(), request)
	controller.WriteResponse(w, resp.Response, nil)
}

func (this *MicroServiceInstanceService) HeartbeatSet(w http.ResponseWriter, r *http.Request) {
	message, err := ioutil.ReadAll(r.Body)
	if err != nil {
		util.Logger().Error("register instance failed, body err", err)
		controller.WriteError(w, scerr.ErrInvalidParams, err.Error())
		return
	}

	request := &pb.HeartbeatSetRequest{}
	err = json.Unmarshal(message, request)
	if err != nil {
		util.Logger().Error("register instance failed, Unmarshal error", err)
		controller.WriteError(w, scerr.ErrInvalidParams, "Unmarshal error")
		return
	}
	resp, _ := core.InstanceAPI.HeartbeatSet(r.Context(), request)

	if resp.Response.Code == pb.Response_SUCCESS {
		controller.WriteJsonObject(w, nil)
		return
	}
	respInternal := resp.Response
	resp.Response = nil
	controller.WriteResponse(w, respInternal, resp)
	return
}

func (this *MicroServiceInstanceService) UnregisterInstance(w http.ResponseWriter, r *http.Request) {
	request := &pb.UnregisterInstanceRequest{
		ServiceId:  r.URL.Query().Get(":serviceId"),
		InstanceId: r.URL.Query().Get(":instanceId"),
	}
	resp, _ := core.InstanceAPI.Unregister(r.Context(), request)
	controller.WriteResponse(w, resp.Response, nil)
}

func (this *MicroServiceInstanceService) FindInstances(w http.ResponseWriter, r *http.Request) {
	var ids []string
	keys := r.URL.Query().Get("tags")
	if len(keys) > 0 {
		ids = strings.Split(keys, ",")
	}
	request := &pb.FindInstancesRequest{
		ConsumerServiceId: r.Header.Get("X-ConsumerId"),
		AppId:             r.URL.Query().Get("appId"),
		ServiceName:       r.URL.Query().Get("serviceName"),
		VersionRule:       r.URL.Query().Get("version"),
		Tags:              ids,
	}

	util.SetTargetDomainProject(r.Context(), r.Header.Get("X-Domain-Name"), r.URL.Query().Get(":project"))

	resp, _ := core.InstanceAPI.Find(r.Context(), request)
	respInternal := resp.Response
	resp.Response = nil
	controller.WriteResponse(w, respInternal, resp)
}

func (this *MicroServiceInstanceService) GetOneInstance(w http.ResponseWriter, r *http.Request) {
	var ids []string
	keys := r.URL.Query().Get("tags")
	if len(keys) > 0 {
		ids = strings.Split(keys, ",")
	}
	request := &pb.GetOneInstanceRequest{
		ConsumerServiceId:  r.Header.Get("X-ConsumerId"),
		ProviderServiceId:  r.URL.Query().Get(":serviceId"),
		ProviderInstanceId: r.URL.Query().Get(":instanceId"),
		Tags:               ids,
	}
	resp, _ := core.InstanceAPI.GetOneInstance(r.Context(), request)
	respInternal := resp.Response
	resp.Response = nil
	controller.WriteResponse(w, respInternal, resp)
}

func (this *MicroServiceInstanceService) GetInstances(w http.ResponseWriter, r *http.Request) {
	var ids []string
	keys := r.URL.Query().Get("tags")
	if len(keys) > 0 {
		ids = strings.Split(keys, ",")
	}
	request := &pb.GetInstancesRequest{
		ConsumerServiceId: r.Header.Get("X-ConsumerId"),
		ProviderServiceId: r.URL.Query().Get(":serviceId"),
		Tags:              ids,
	}
	resp, _ := core.InstanceAPI.GetInstances(r.Context(), request)
	respInternal := resp.Response
	resp.Response = nil
	controller.WriteResponse(w, respInternal, resp)
}

func (this *MicroServiceInstanceService) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("value")
	request := &pb.UpdateInstanceStatusRequest{
		ServiceId:  r.URL.Query().Get(":serviceId"),
		InstanceId: r.URL.Query().Get(":instanceId"),
		Status:     status,
	}
	resp, _ := core.InstanceAPI.UpdateStatus(r.Context(), request)
	controller.WriteResponse(w, resp.Response, nil)
}

func (this *MicroServiceInstanceService) UpdateMetadata(w http.ResponseWriter, r *http.Request) {
	message, err := ioutil.ReadAll(r.Body)
	if err != nil {
		util.Logger().Error("body err", err)
		controller.WriteError(w, scerr.ErrInvalidParams, err.Error())
		return
	}
	request := &pb.UpdateInstancePropsRequest{
		ServiceId:  r.URL.Query().Get(":serviceId"),
		InstanceId: r.URL.Query().Get(":instanceId"),
	}
	err = json.Unmarshal(message, request)
	if err != nil {
		util.Logger().Error("Unmarshal error", err)
		controller.WriteError(w, scerr.ErrInvalidParams, "Unmarshal error")
		return
	}
	resp, err := core.InstanceAPI.UpdateInstanceProperties(r.Context(), request)
	controller.WriteResponse(w, resp.Response, nil)
}
