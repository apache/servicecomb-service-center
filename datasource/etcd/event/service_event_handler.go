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

package event

import (
	"context"
	"fmt"

	"github.com/apache/servicecomb-service-center/datasource"
	"github.com/apache/servicecomb-service-center/datasource/etcd/cache"
	"github.com/apache/servicecomb-service-center/datasource/etcd/path"
	"github.com/apache/servicecomb-service-center/datasource/etcd/sd"
	"github.com/apache/servicecomb-service-center/datasource/etcd/state/kvstore"
	serviceUtil "github.com/apache/servicecomb-service-center/datasource/etcd/util"
	"github.com/apache/servicecomb-service-center/pkg/log"
	pb "github.com/go-chassis/cari/discovery"
)

// ServiceEventHandler is the handler to handle:
// 2. save the new domain & project mapping
// 3. reset the find instance cache
type ServiceEventHandler struct {
}

func (h *ServiceEventHandler) Type() kvstore.Type {
	return sd.TypeService
}

func (h *ServiceEventHandler) OnEvent(evt kvstore.Event) {
	ms, ok := evt.KV.Value.(*pb.MicroService)
	if !ok {
		log.Error("failed to assert MicroService", datasource.ErrAssertFail)
		return
	}
	_, domainProject := path.GetInfoFromSvcKV(evt.KV.Key)

	switch evt.Type {
	case pb.EVT_INIT, pb.EVT_CREATE:
		newDomain, newProject := path.SplitDomainProject(domainProject)
		err := serviceUtil.NewDomainProject(context.Background(), newDomain, newProject)
		if err != nil {
			log.Error(fmt.Sprintf("new domain[%s] or project[%s] failed", newDomain, newProject), err)
		}
	}

	if evt.Type == pb.EVT_INIT {
		return
	}

	log.Info(fmt.Sprintf("caught [%s] service[%s][%s/%s/%s/%s] event",
		evt.Type, ms.ServiceId, ms.Environment, ms.AppId, ms.ServiceName, ms.Version))

	// cache
	providerKey := pb.MicroServiceToKey(domainProject, ms)
	cache.FindInstances.Remove(providerKey)
}

func NewServiceEventHandler() *ServiceEventHandler {
	return &ServiceEventHandler{}
}
