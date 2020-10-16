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
	"github.com/apache/servicecomb-service-center/datasource/etcd/cache"
	"github.com/apache/servicecomb-service-center/pkg/registry"
	"github.com/apache/servicecomb-service-center/server/core"
	"k8s.io/api/core/v1"
)

type ServiceIndexCacher struct {
	*cache.CommonCacher
}

// onServiceEvent is the method to refresh service cache
func (c *ServiceIndexCacher) onServiceEvent(evt K8sEvent) {
	svc := evt.Object.(*v1.Service)
	domainProject := Kubernetes().GetDomainProject()
	indexKey := core.GenerateServiceIndexKey(generateServiceKey(domainProject, svc))
	serviceID := generateServiceID(domainProject, svc)

	if !ShouldRegisterService(svc) {
		kv := c.Cache().Get(indexKey)
		if kv != nil {
			c.Notify(registry.EVT_DELETE, indexKey, kv)
		}
		return
	}

	switch evt.EventType {
	case registry.EVT_CREATE:
		kv := AsKeyValue(indexKey, serviceID, svc.ResourceVersion)
		c.Notify(evt.EventType, indexKey, kv)
	case registry.EVT_UPDATE:
	case registry.EVT_DELETE:
		kv := c.Cache().Get(indexKey)
		if kv != nil {
			c.Notify(evt.EventType, indexKey, kv)
		}
	}
}

func NewServiceIndexCacher(c *cache.CommonCacher) (si *ServiceIndexCacher) {
	si = &ServiceIndexCacher{CommonCacher: c}
	Kubernetes().AppendEventFunc(TypeService, si.onServiceEvent)
	return
}
