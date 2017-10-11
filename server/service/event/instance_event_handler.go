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
package event

import (
	"encoding/json"
	"github.com/ServiceComb/service-center/pkg/util"
	pb "github.com/ServiceComb/service-center/server/core/proto"
	"github.com/ServiceComb/service-center/server/core/registry/store"
	nf "github.com/ServiceComb/service-center/server/service/notification"
	serviceUtil "github.com/ServiceComb/service-center/server/service/util"
	"golang.org/x/net/context"
)

type InstanceEventHandler struct {
}

func (h *InstanceEventHandler) Type() store.StoreType {
	return store.INSTANCE
}

func (h *InstanceEventHandler) OnEvent(evt *store.KvEvent) {
	kv := evt.KV
	action := evt.Action
	providerId, providerInstanceId, tenantProject, data := pb.GetInfoFromInstKV(kv)
	if data == nil {
		util.Logger().Errorf(nil,
			"unmarshal provider service instance file failed, instance %s/%s [%s] event, data is nil",
			providerId, providerInstanceId, action)
		return
	}

	if nf.GetNotifyService().Closed() {
		util.Logger().Warnf(nil, "caught instance %s/%s [%s] event, but notify service is closed",
			providerId, providerInstanceId, action)
		return
	}
	util.Logger().Infof("caught instance %s/%s [%s] event", providerId, providerInstanceId, action)

	var instance pb.MicroServiceInstance
	err := json.Unmarshal(data, &instance)
	if err != nil {
		util.Logger().Errorf(err, "unmarshal provider service instance %s/%s file failed",
			providerId, providerInstanceId)
		return
	}
	// 查询服务版本信息
	ms, err := serviceUtil.GetServiceInCache(context.Background(), tenantProject, providerId)
	if ms == nil {
		util.Logger().Errorf(err, "get provider service %s/%s id in cache failed",
			providerId, providerInstanceId)
		return
	}

	// 查询所有consumer
	consumerIds, _, err := serviceUtil.GetConsumerIds(context.Background(), tenantProject, ms)
	if err != nil {
		util.Logger().Errorf(err, "query service %s consumers failed", providerId)
		return
	}

	nf.PublishInstanceEvent(tenantProject, action, &pb.MicroServiceKey{
		AppId:       ms.AppId,
		ServiceName: ms.ServiceName,
		Version:     ms.Version,
	}, &instance, evt.Revision, consumerIds)
}

func NewInstanceEventHandler() *InstanceEventHandler {
	return &InstanceEventHandler{}
}
