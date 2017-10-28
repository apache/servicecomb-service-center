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
	"fmt"
	"github.com/ServiceComb/service-center/pkg/util"
	pb "github.com/ServiceComb/service-center/server/core/proto"
	"github.com/ServiceComb/service-center/server/core/registry"
	"github.com/ServiceComb/service-center/server/core/registry/store"
	nf "github.com/ServiceComb/service-center/server/service/notification"
	serviceUtil "github.com/ServiceComb/service-center/server/service/util"
	"golang.org/x/net/context"
	"strings"
)

type DomainAsyncTask struct {
	key string
	err error

	Domain string
}

func (apt *DomainAsyncTask) Key() string {
	return apt.key
}

func (apt *DomainAsyncTask) Do(ctx context.Context) error {
	defer store.Store().AsyncTasker().DeferRemoveTask(apt.Key())

	ok, err := serviceUtil.DomainExist(ctx, apt.Domain)
	if err != nil {
		util.Logger().Errorf(err, "find domain %s file failed", apt.Domain)
		return err
	}
	if ok {
		return nil
	}

	err = serviceUtil.NewDomain(ctx, apt.Domain)
	if err != nil {
		util.Logger().Errorf(err, "new domain %s file failed", apt.Domain)
		return err
	}
	return nil
}

func (apt *DomainAsyncTask) Err() error {
	return apt.err
}

func (apt *DomainAsyncTask) publish(ctx context.Context, tenant, consumerId string, rev int64) error {
	consumer, err := serviceUtil.GetService(ctx, tenant, consumerId)
	if err != nil {
		util.Logger().Errorf(err, "get comsumer for publish event %s failed", consumerId)
		return err
	}
	if consumer == nil {
		consumerTmp, found := serviceUtil.MsCache().Get(consumerId)
		if !found {
			util.Logger().Errorf(nil, "service not exist, %s", consumerId)
			return fmt.Errorf("service not exist, %s", consumerId)
		}
		consumer = consumerTmp.(*pb.MicroService)
	}
	providerIds, err := serviceUtil.GetProvidersInCache(ctx, tenant, consumerId, consumer)
	if err != nil {
		util.Logger().Errorf(err, "get provider services by consumer %s failed", consumerId)
		return err
	}

	for _, providerId := range providerIds {
		provider, err := serviceUtil.GetService(ctx, tenant, providerId)
		if provider == nil {
			util.Logger().Warnf(err, "get service %s file failed", providerId)
			continue
		}
		nf.PublishInstanceEvent(tenant, pb.EVT_EXPIRE,
			&pb.MicroServiceKey{
				AppId:       provider.AppId,
				ServiceName: provider.ServiceName,
				Version:     provider.Version,
			}, nil, rev, []string{consumerId})
	}
	return nil
}

type ServiceEventHandler struct {
}

func (h *ServiceEventHandler) Type() store.StoreType {
	return store.SERVICE
}

func (h *ServiceEventHandler) OnEvent(evt *store.KvEvent) {
	kv := evt.KV
	action := evt.Action
	serviceId, tenantProject, data := pb.GetInfoFromSvcKV(kv)
	if data == nil {
		util.Logger().Errorf(nil,
			"unmarshal service file failed, service %s [%s] event, data is nil",
			serviceId, action)
		return
	}

	if nf.GetNotifyService().Closed() {
		util.Logger().Warnf(nil, "caught service %s [%s] event, but notify service is closed",
			serviceId, action)
		return
	}

	switch action {
	case pb.EVT_CREATE:
		newDomain := tenantProject[:strings.Index(tenantProject, "/")]
		ok, err := serviceUtil.DomainExist(context.Background(), newDomain, registry.WithCacheOnly())
		if err != nil {
			util.Logger().Errorf(err, "find domain %s file failed", newDomain)
			return
		}
		if ok {
			return
		}
		store.Store().AsyncTasker().AddTask(context.Background(), NewDomainAsyncTask(newDomain))
	}
}

func NewServiceEventHandler() *ServiceEventHandler {
	return &ServiceEventHandler{}
}

func NewDomainAsyncTask(domain string) *DomainAsyncTask {
	return &DomainAsyncTask{
		key:    "DomainAsyncTask_" + domain,
		Domain: domain,
	}
}
