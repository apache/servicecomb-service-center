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
package notification

import (
	"encoding/json"
	apt "github.com/ServiceComb/service-center/server/core"
	pb "github.com/ServiceComb/service-center/server/core/proto"
	"github.com/ServiceComb/service-center/server/core/registry/store"
	"github.com/ServiceComb/service-center/server/service/dependency"
	"github.com/ServiceComb/service-center/server/service/microservice"
	"github.com/ServiceComb/service-center/util"
	"golang.org/x/net/context"
	"strings"
)

type InstanceEventHandler struct {
	service *NotifyService
}

func (h *InstanceEventHandler) Type() store.StoreType {
	return store.INSTANCE
}

func (h *InstanceEventHandler) OnEvent(evt *store.KvEvent) {
	kv := evt.KV
	action := evt.Action
	providerId, providerInstanceId, tenantProject, data := pb.GetInfoFromInstKV(kv)
	if data == nil {
		util.LOGGER.Errorf(nil,
			"unmarshal provider service instance file failed, instance %s/%s [%s] event, data is nil",
			providerId, providerInstanceId, action)
		return
	}

	if h.service.Closed() {
		util.LOGGER.Warnf(nil, "caught instance %s/%s [%s] event, but notify service is closed",
			providerId, providerInstanceId, action)
		return
	}
	util.LOGGER.Infof("caught instance %s/%s [%s] event",
		providerId, providerInstanceId, action)

	var instance pb.MicroServiceInstance
	err := json.Unmarshal(data, &instance)
	if err != nil {
		util.LOGGER.Errorf(err, "unmarshal provider service instance %s/%s file failed",
			providerId, providerInstanceId)
		return
	}
	// 查询服务版本信息
	ms, err := microservice.GetServiceInCache(context.Background(), tenantProject, providerId)
	if ms == nil {
		util.LOGGER.Errorf(err, "get provider service %s/%s id in cache failed",
			providerId, providerInstanceId)
		return
	}

	// 查询所有consumer
	Kvs, err := dependency.GetConsumersInCache(context.Background(), tenantProject, providerId)
	if err != nil {
		util.LOGGER.Errorf(err, "query service %s consumers failed", providerId)
		return
	}

	response := &pb.WatchInstanceResponse{
		Response: pb.CreateResponse(pb.Response_SUCCESS, "Watch instance successfully."),
		Action:   string(action),
		Key: &pb.MicroServiceKey{
			AppId:       ms.AppId,
			ServiceName: ms.ServiceName,
			Version:     ms.Version,
		},
		Instance: &instance,
	}
	for _, dependence := range Kvs {
		consumer := util.BytesToStringWithNoCopy(dependence.Key)
		consumer = consumer[strings.LastIndex(consumer, "/")+1:]
		job := NewWatchJob(INSTANCE, consumer, apt.GetInstanceRootKey(tenantProject)+"/",
			evt.Revision, response)
		util.LOGGER.Debugf("publish event to notify service, %v", job)

		// TODO add超时怎么处理？
		h.service.AddJob(job)
	}
}

func NewInstanceEventHandler(s *NotifyService) EventHandler {
	return &InstanceEventHandler{
		service: s,
	}
}
