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
	"github.com/ServiceComb/service-center/pkg/util"
	"github.com/ServiceComb/service-center/server/core/backend/store"
	pb "github.com/ServiceComb/service-center/server/core/proto"
	serviceUtil "github.com/ServiceComb/service-center/server/service/util"
	"golang.org/x/net/context"
	"strings"
)

type ServiceEventHandler struct {
}

func (h *ServiceEventHandler) Type() store.StoreType {
	return store.SERVICE
}

func (h *ServiceEventHandler) OnEvent(evt *store.KvEvent) {
	action := evt.Action
	if action != pb.EVT_CREATE && action != pb.EVT_INIT {
		return
	}

	kv := evt.KV
	serviceId, domainProject, data := pb.GetInfoFromSvcKV(kv)
	if data == nil {
		util.Logger().Errorf(nil,
			"unmarshal service file failed, service %s [%s] event, data is nil",
			serviceId, action)
		return
	}

	switch action {
	case pb.EVT_CREATE, pb.EVT_INIT:
		newDomain := domainProject[:strings.Index(domainProject, "/")]
		newProject := domainProject[strings.Index(domainProject, "/")+1:]
		err := serviceUtil.NewDomainProject(context.Background(), newDomain, newProject)
		if err != nil {
			util.Logger().Errorf(err, "new domain(%s) or project(%s) failed", newDomain, newProject)
			return
		}
	}
}

func NewServiceEventHandler() *ServiceEventHandler {
	return &ServiceEventHandler{}
}
