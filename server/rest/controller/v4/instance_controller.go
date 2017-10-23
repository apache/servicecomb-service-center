//Copyright 2017 Huawei Technologies Co., Ltd
//
//Licensed under the Apache License, Version 2.0 (the "License");
//you may not use this file except in compliance with the License.
//You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
//Unless required by applicable law or agreed to in writing, software
//distributed under the License is distributed on an "AS IS" BASIS,
//WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//See the License for the specific language governing permissions and
//limitations under the License.
package v4

import (
	"encoding/json"
	"github.com/ServiceComb/service-center/pkg/rest"
	"github.com/ServiceComb/service-center/pkg/util"
	"github.com/ServiceComb/service-center/server/core"
	pb "github.com/ServiceComb/service-center/server/core/proto"
	"github.com/ServiceComb/service-center/server/rest/controller"
	"io/ioutil"
	"net/http"
	"strings"
)

type MicroServiceInstanceService struct {
	//
}

func (this *MicroServiceInstanceService) URLPatterns() []rest.Route {
	return []rest.Route{
		{rest.HTTP_METHOD_GET, "/v4/:domain/registry/instances", this.FindInstances},
		{rest.HTTP_METHOD_GET, "/v4/:domain/registry/microservices/:serviceId/instances", this.GetInstances},
		{rest.HTTP_METHOD_GET, "/v4/:domain/registry/microservices/:serviceId/instances/:instanceId", this.GetOneInstance},
		{rest.HTTP_METHOD_POST, "/v4/:domain/registry/microservices/:serviceId/instances", this.RegisterInstance},
		{rest.HTTP_METHOD_DELETE, "/v4/:domain/registry/microservices/:serviceId/instances/:instanceId", this.UnregisterInstance},
		{rest.HTTP_METHOD_PUT, "/v4/:domain/registry/microservices/:serviceId/instances/:instanceId/properties", this.UpdateMetadata},
		{rest.HTTP_METHOD_PUT, "/v4/:domain/registry/microservices/:serviceId/instances/:instanceId/status", this.UpdateStatus},
		{rest.HTTP_METHOD_PUT, "/v4/:domain/registry/microservices/:serviceId/instances/:instanceId/heartbeat", this.Heartbeat},
		{rest.HTTP_METHOD_PUT, "/v4/:domain/registry/heartbeats", this.HeartbeatSet},
	}
}
func (this *MicroServiceInstanceService) RegisterInstance(w http.ResponseWriter, r *http.Request) {
	message, err := ioutil.ReadAll(r.Body)
	if err != nil {
		util.Logger().Error("register instance failed, body err", err)
		controller.WriteText(http.StatusBadRequest, err.Error(), w)
		return
	}

	request := &pb.RegisterInstanceRequest{}
	err = json.Unmarshal(message, request)
	if err != nil {
		util.Logger().Error("register instance failed, Unmarshal error", err)
		controller.WriteText(http.StatusInternalServerError, "Unmarshal error", w)
		return
	}
	if request.GetInstance() != nil {
		request.Instance.ServiceId = r.URL.Query().Get(":serviceId")
	}

	resp, err := core.InstanceAPI.Register(r.Context(), request)
	respInternal := resp.Response
	resp.Response = nil
	controller.WriteJsonResponse(respInternal, resp, err, w)
}

//TODO 什么样的服务允许更新服务心跳，只能是本服务才可以更新自己，如何屏蔽其他服务伪造的心跳更新？
func (this *MicroServiceInstanceService) Heartbeat(w http.ResponseWriter, r *http.Request) {
	request := &pb.HeartbeatRequest{
		ServiceId:  r.URL.Query().Get(":serviceId"),
		InstanceId: r.URL.Query().Get(":instanceId"),
	}
	resp, err := core.InstanceAPI.Heartbeat(r.Context(), request)
	controller.WriteTextResponse(resp.GetResponse(), err, "", w)
}

func (this *MicroServiceInstanceService) HeartbeatSet(w http.ResponseWriter, r *http.Request) {
	message, err := ioutil.ReadAll(r.Body)
	if err != nil {
		util.Logger().Error("register instance failed, body err", err)
		controller.WriteText(http.StatusBadRequest, err.Error(), w)
		return
	}

	request := &pb.HeartbeatSetRequest{}
	err = json.Unmarshal(message, request)
	if err != nil {
		util.Logger().Error("register instance failed, Unmarshal error", err)
		controller.WriteText(http.StatusInternalServerError, "Unmarshal error", w)
		return
	}
	resp, _ := core.InstanceAPI.HeartbeatSet(r.Context(), request)

	if resp.Response.Code == pb.Response_SUCCESS {
		controller.WriteText(http.StatusOK, "", w)
		return
	}
	if resp.Instances == nil || len(resp.Instances) == 0 {
		controller.WriteText(http.StatusBadRequest, resp.Response.Message, w)
		return
	}
	resp.Response = nil
	objJson, err := json.Marshal(resp)
	if err != nil {
		controller.WriteText(http.StatusInternalServerError, err.Error(), w)
		return
	}
	controller.WriteJson(http.StatusBadRequest, objJson, w)
	return
}

func (this *MicroServiceInstanceService) UnregisterInstance(w http.ResponseWriter, r *http.Request) {
	request := &pb.UnregisterInstanceRequest{
		ServiceId:  r.URL.Query().Get(":serviceId"),
		InstanceId: r.URL.Query().Get(":instanceId"),
	}
	resp, err := core.InstanceAPI.Unregister(r.Context(), request)
	controller.WriteTextResponse(resp.GetResponse(), err, "", w)
}

func (this *MicroServiceInstanceService) FindInstances(w http.ResponseWriter, r *http.Request) {
	var ids []string
	keys := r.URL.Query().Get("tags")
	if len(keys) > 0 {
		ids = strings.Split(keys, ",")
	}
	noCache := r.URL.Query().Get("noCache")
	if noCache != "0" && noCache != "1" && strings.TrimSpace(noCache) != "" {
		controller.WriteText(http.StatusBadRequest, "parameter noCache must be 1 or 0", w)
		return
	}
	request := &pb.FindInstancesRequest{
		ConsumerServiceId: r.Header.Get("X-ConsumerId"),
		AppId:             r.URL.Query().Get("appId"),
		ServiceName:       r.URL.Query().Get("serviceName"),
		VersionRule:       r.URL.Query().Get("version"),
		Env:               r.URL.Query().Get("env"),
		Tags:              ids,
		NoCache:           noCache == "1",
	}
	resp, err := core.InstanceAPI.Find(r.Context(), request)
	if err != nil {
		controller.WriteText(http.StatusInternalServerError, err.Error(), w)
		return
	}
	if resp.GetResponse().Code != pb.Response_SUCCESS {
		controller.WriteText(http.StatusBadRequest, resp.GetResponse().Message, w)
		return
	}
	resp.Response = nil
	controller.WriteJsonObject(http.StatusOK, resp, w)
}

func (this *MicroServiceInstanceService) GetOneInstance(w http.ResponseWriter, r *http.Request) {
	var ids []string
	keys := r.URL.Query().Get("tags")
	if len(keys) > 0 {
		ids = strings.Split(keys, ",")
	}
	noCache := r.URL.Query().Get("noCache")
	if noCache != "0" && noCache != "1" && strings.TrimSpace(noCache) != "" {
		controller.WriteText(http.StatusBadRequest, "parameter noCache must be 1 or 0", w)
		return
	}
	request := &pb.GetOneInstanceRequest{
		ConsumerServiceId:  r.Header.Get("X-ConsumerId"),
		ProviderServiceId:  r.URL.Query().Get(":serviceId"),
		ProviderInstanceId: r.URL.Query().Get(":instanceId"),
		Tags:               ids,
		Env:                r.URL.Query().Get("env"),
		NoCache:            noCache == "1",
	}
	resp, err := core.InstanceAPI.GetOneInstance(r.Context(), request)
	if err != nil {
		controller.WriteText(http.StatusInternalServerError, err.Error(), w)
		return
	}
	if resp.GetResponse().Code != pb.Response_SUCCESS {
		controller.WriteText(http.StatusBadRequest, resp.GetResponse().Message, w)
		return
	}
	resp.Response = nil
	controller.WriteJsonObject(http.StatusOK, resp, w)
}

func (this *MicroServiceInstanceService) GetInstances(w http.ResponseWriter, r *http.Request) {
	var ids []string
	keys := r.URL.Query().Get("tags")
	if len(keys) > 0 {
		ids = strings.Split(keys, ",")
	}
	noCache := r.URL.Query().Get("noCache")
	if noCache != "0" && noCache != "1" && strings.TrimSpace(noCache) != "" {
		controller.WriteText(http.StatusBadRequest, "parameter noCache must be 1 or 0", w)
		return
	}
	request := &pb.GetInstancesRequest{
		ConsumerServiceId: r.Header.Get("X-ConsumerId"),
		ProviderServiceId: r.URL.Query().Get(":serviceId"),
		Tags:              ids,
		Env:               r.URL.Query().Get("env"),
		NoCache:           noCache == "1",
	}
	resp, err := core.InstanceAPI.GetInstances(r.Context(), request)
	if err != nil {
		controller.WriteText(http.StatusInternalServerError, err.Error(), w)
		return
	}
	if resp.GetResponse().Code != pb.Response_SUCCESS {
		controller.WriteText(http.StatusBadRequest, resp.GetResponse().Message, w)
		return
	}
	resp.Response = nil
	controller.WriteJsonObject(http.StatusOK, resp, w)
}

func (this *MicroServiceInstanceService) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("value")
	request := &pb.UpdateInstanceStatusRequest{
		ServiceId:  r.URL.Query().Get(":serviceId"),
		InstanceId: r.URL.Query().Get(":instanceId"),
		Status:     status,
	}
	resp, err := core.InstanceAPI.UpdateStatus(r.Context(), request)
	controller.WriteTextResponse(resp.GetResponse(), err, "", w)
}

func (this *MicroServiceInstanceService) UpdateMetadata(w http.ResponseWriter, r *http.Request) {
	message, err := ioutil.ReadAll(r.Body)
	if err != nil {
		util.Logger().Error("body err", err)
		controller.WriteText(http.StatusBadRequest, err.Error(), w)
		return
	}
	request := &pb.UpdateInstancePropsRequest{
		ServiceId:  r.URL.Query().Get(":serviceId"),
		InstanceId: r.URL.Query().Get(":instanceId"),
	}
	err = json.Unmarshal(message, request)
	if err != nil {
		util.Logger().Error("Unmarshal error", err)
		controller.WriteText(http.StatusInternalServerError, "Unmarshal error", w)
		return
	}
	resp, err := core.InstanceAPI.UpdateInstanceProperties(r.Context(), request)
	controller.WriteTextResponse(resp.GetResponse(), err, "", w)
}
