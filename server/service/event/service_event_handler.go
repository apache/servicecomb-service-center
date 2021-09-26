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
	"strings"

	"github.com/apache/servicecomb-service-center/pkg/log"
	pb "github.com/apache/servicecomb-service-center/pkg/registry"
	"github.com/apache/servicecomb-service-center/server/core"
	"github.com/apache/servicecomb-service-center/server/core/backend"
	"github.com/apache/servicecomb-service-center/server/core/proto"
	"github.com/apache/servicecomb-service-center/server/plugin/discovery"
	"github.com/apache/servicecomb-service-center/server/service/cache"
	serviceUtil "github.com/apache/servicecomb-service-center/server/service/util"
)

// ServiceEventHandler is the handler to handle:
// 1. report service metrics
// 2. save the new domain & project mapping
// 3. reset the find instance cache
type ServiceEventHandler struct {
}

func (h *ServiceEventHandler) Type() discovery.Type {
	return backend.SERVICE
}

func (h *ServiceEventHandler) OnEvent(evt discovery.KvEvent) {
	ms := evt.KV.Value.(*pb.MicroService)
	_, domainProject := core.GetInfoFromSvcKV(evt.KV.Key)

	switch evt.Type {
	case pb.EVT_INIT, pb.EVT_CREATE:
		idx := strings.Index(domainProject, "/")
		newDomain := domainProject[:idx]
		newProject := domainProject[idx+1:]
		err := serviceUtil.NewDomainProject(context.Background(), newDomain, newProject)
		if err != nil {
			log.Errorf(err, "new domain[%s] or project[%s] failed", newDomain, newProject)
		}
	}

	if evt.Type == pb.EVT_INIT {
		return
	}

	log.Infof("caught [%s] service[%s][%s/%s/%s/%s] event",
		evt.Type, ms.ServiceId, ms.Environment, ms.AppId, ms.ServiceName, ms.Version)

	// cache
	providerKey := proto.MicroServiceToKey(domainProject, ms)
	cache.FindInstances.Remove(providerKey)
}

func NewServiceEventHandler() *ServiceEventHandler {
	return &ServiceEventHandler{}
}
