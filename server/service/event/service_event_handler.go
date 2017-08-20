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
	pb "github.com/ServiceComb/service-center/server/core/proto"
	"github.com/ServiceComb/service-center/server/core/registry/store"
	"github.com/ServiceComb/service-center/server/service/dependency"
	"github.com/ServiceComb/service-center/util"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"golang.org/x/net/context"
	"time"
)

const DEFAULT_TASK_TIMOUT = 1 * time.Minute

type UpdateDependencyAsyncTask struct {
	key string
	kv  *mvccpb.KeyValue
	err error
}

func (at *UpdateDependencyAsyncTask) Key() string {
	return at.key
}

func (at *UpdateDependencyAsyncTask) Err() error {
	return at.err
}

func (at *UpdateDependencyAsyncTask) Do(ctx context.Context) (err error) {
	for {
		err = at.do(ctx)
		if err == nil {
			return
		}

		select {
		case <-ctx.Done():
			return
		default:
		}
	}
}

func (at *UpdateDependencyAsyncTask) do(ctx context.Context) (err error) {
	defer func() { at.err = err }()

	serviceId, tenantProject, _ := pb.GetInfoFromSvcKV(at.kv)

	var service pb.MicroService
	err = json.Unmarshal(at.kv.Value, &service)
	if err != nil {
		util.LOGGER.Errorf(err, "create microservice: unmarshal service %s file failed", serviceId)
		return
	}
	//创建服务间的依赖
	serviceFlag := util.StringJoin([]string{service.AppId, service.ServiceName, service.Version}, "/")
	util.LOGGER.Infof("create microservice: add dependency for %s(%s)", serviceId, serviceFlag)
	err = dependency.UpdateAsProviderDependency(context.Background(), serviceId, &pb.MicroServiceKey{
		AppId:       service.AppId,
		ServiceName: service.ServiceName,
		Alias:       service.Alias,
		Version:     service.Version,
		Tenant:      tenantProject,
	})
	if err != nil {
		util.LOGGER.Errorf(err, "create microservice: update dependency as provider %s(%s) failed",
			serviceId, serviceFlag)
	}
	return
}

type ServiceEventHandler struct {
}

func (h *ServiceEventHandler) Type() store.StoreType {
	return store.SERVICE
}

func (h *ServiceEventHandler) OnEvent(evt *store.KvEvent) {
	kv := evt.KV
	action := evt.Action

	serviceId, _, data := pb.GetInfoFromSvcKV(kv)
	if data == nil {
		util.LOGGER.Errorf(nil, "unmarshal service file failed, service %s [%s] event, data is nil",
			serviceId, action)
		return
	}

	util.LOGGER.Infof("caught service %s [%s] event", serviceId, action)

	key := util.BytesToStringWithNoCopy(kv.Key)

	switch evt.Action {
	case pb.EVT_UPDATE:
		return
	case pb.EVT_DELETE:
		store.Store().AsyncTasker().RemoveTask(key)
		return
	case pb.EVT_CREATE:
		ctx, _ := context.WithTimeout(context.Background(), DEFAULT_TASK_TIMOUT)
		store.Store().AsyncTasker().AddTask(ctx, &UpdateDependencyAsyncTask{
			key: key,
			kv:  kv,
		})
	}
}

func NewServiceEventHandler() *ServiceEventHandler {
	return &ServiceEventHandler{}
}
