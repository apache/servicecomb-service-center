// Licensed to the Apache Software Foundation (ASF) under one or more
// contributor license agreements.  See the NOTICE file distributed with
// this work for additional information regarding copyright ownership.
// The ASF licenses this file to You under the Apache License, Version 2.0
// (the "License"); you may not use this file except in compliance with
// the License.  You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package adaptor

import (
	"github.com/apache/incubator-servicecomb-service-center/server/core"
	pb "github.com/apache/incubator-servicecomb-service-center/server/core/proto"
	"github.com/apache/incubator-servicecomb-service-center/server/plugin/pkg/discovery"
	"k8s.io/api/core/v1"
)

type ServiceIndexCacher struct {
	*discovery.CommonCacher
}

// onServiceEvent is the method to refresh service cache
func (c *ServiceIndexCacher) onServiceEvent(evt K8sEvent) {
	svc := evt.Object.(*v1.Service)
	domainProject := Kubernetes().GetDomainProject()
	serviceId := string(svc.UID)
	indexKey := core.GenerateServiceIndexKey(generateServiceKey(domainProject, svc))

	if !CanRegisterService(svc) {
		kv := c.Cache().Get(indexKey)
		if kv != nil {
			c.Notify(pb.EVT_DELETE, indexKey, kv)
		}
		return
	}

	switch evt.EventType {
	case pb.EVT_CREATE:
		kv := AsKeyValue(indexKey, serviceId, svc.ResourceVersion)
		c.Notify(evt.EventType, indexKey, kv)
	case pb.EVT_UPDATE:
	case pb.EVT_DELETE:
		kv := c.Cache().Get(indexKey)
		if kv != nil {
			c.Notify(evt.EventType, indexKey, kv)
		}
	}
}

func NewServiceIndexCacher(c *discovery.CommonCacher) (si *ServiceIndexCacher) {
	si = &ServiceIndexCacher{CommonCacher: c}
	Kubernetes().AppendEventFunc(TypeService, si.onServiceEvent)
	return
}
