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
	pb "github.com/ServiceComb/service-center/server/core/proto"
	"github.com/ServiceComb/service-center/server/core/registry"
	"github.com/ServiceComb/service-center/server/core/registry/store"
	serviceUtil "github.com/ServiceComb/service-center/server/service/util"
	"golang.org/x/net/context"
	"strings"
)

type DomainProjectAsyncTask struct {
	key string
	err error

	Domain  string
	Project string
}

func (apt *DomainProjectAsyncTask) Key() string {
	return apt.key
}

func (apt *DomainProjectAsyncTask) Err() error {
	return apt.err
}

func (apt *DomainProjectAsyncTask) Do(ctx context.Context) error {
	defer store.AsyncTaskService().DeferRemove(apt.Key())

	err := serviceUtil.NewDomainProject(ctx, apt.Domain, apt.Project)
	if err != nil {
		util.Logger().Errorf(err, "new domain(%s) or project(%s) failed", apt.Domain, apt.Project)
		return err
	}
	util.Logger().Infof("new domain(%s) and project(%s)", apt.Domain, apt.Project)
	return nil
}

type ServiceEventHandler struct {
}

func (h *ServiceEventHandler) Type() store.StoreType {
	return store.SERVICE
}

func (h *ServiceEventHandler) OnEvent(evt *store.KvEvent) {
	action := evt.Action
	if action == pb.EVT_DELETE || action == pb.EVT_UPDATE {
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
		ok, err := serviceUtil.ProjectExist(context.Background(), newDomain, newProject, registry.WithCacheOnly())
		if err != nil {
			util.Logger().Errorf(err, "find project %s/%s file failed", newDomain, newProject)
			return
		}
		if ok {
			return
		}
		store.AsyncTaskService().Add(context.Background(), NewDomainProjectAsyncTask(newDomain, newProject))
	}
}

func NewServiceEventHandler() *ServiceEventHandler {
	return &ServiceEventHandler{}
}

func NewDomainProjectAsyncTask(domain, project string) *DomainProjectAsyncTask {
	return &DomainProjectAsyncTask{
		key:     "DomainProjectAsyncTask_" + domain + "/" + project,
		Domain:  domain,
		Project: project,
	}
}
