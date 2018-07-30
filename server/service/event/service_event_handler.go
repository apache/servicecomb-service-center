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
	"github.com/apache/incubator-servicecomb-service-center/pkg/util"
	"github.com/apache/incubator-servicecomb-service-center/server/core/backend"
	pb "github.com/apache/incubator-servicecomb-service-center/server/core/proto"
	"github.com/apache/incubator-servicecomb-service-center/server/service/cache"
	"github.com/apache/incubator-servicecomb-service-center/server/service/metrics"
	serviceUtil "github.com/apache/incubator-servicecomb-service-center/server/service/util"
	"golang.org/x/net/context"
	"strings"
)

type ServiceEventHandler struct {
}

func (h *ServiceEventHandler) Type() backend.StoreType {
	return backend.SERVICE
}

func (h *ServiceEventHandler) OnEvent(evt backend.KvEvent) {
	ms := evt.KV.Value.(*pb.MicroService)
	_, domainProject := backend.GetInfoFromSvcKV(evt.KV)
	fn, fv := getFramework(ms)

	switch evt.Type {
	case pb.EVT_INIT, pb.EVT_CREATE:
		metrics.ReportServices(fn, fv, 1)

		newDomain := domainProject[:strings.Index(domainProject, "/")]
		newProject := domainProject[strings.Index(domainProject, "/")+1:]
		err := serviceUtil.NewDomainProject(context.Background(), newDomain, newProject)
		if err != nil {
			util.Logger().Errorf(err, "new domain(%s) or project(%s) failed", newDomain, newProject)
		}
	case pb.EVT_DELETE:
		metrics.ReportServices(fn, fv, -1)
	default:
	}

	if evt.Type == pb.EVT_INIT {
		return
	}

	util.Logger().Infof("caught [%s] service %s/%s/%s event",
		evt.Type, ms.AppId, ms.ServiceName, ms.Version)

	// cache
	providerKey := pb.MicroServiceToKey(domainProject, ms)
	cache.FindInstances.Remove(providerKey)
}

func getFramework(ms *pb.MicroService) (string, string) {
	if ms.Framework != nil && len(ms.Framework.Name) > 0 {
		version := ms.Framework.Version
		if len(ms.Framework.Version) == 0 {
			version = "UNKNOWN"
		}
		return ms.Framework.Name, version
	}
	return "UNKNOWN", "UNKNOWN"
}

func NewServiceEventHandler() *ServiceEventHandler {
	return &ServiceEventHandler{}
}
