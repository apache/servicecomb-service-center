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
	"github.com/apache/servicecomb-service-center/datasource/etcd/path"
	"github.com/apache/servicecomb-service-center/datasource/etcd/sd"
	pb "github.com/go-chassis/cari/discovery"
	"k8s.io/api/core/v1"
)

type ServiceCacher struct {
	*sd.CommonCacher
}

// onServiceEvent is the method to refresh service cache
func (c *ServiceCacher) onServiceEvent(evt K8sEvent) {
	svc := evt.Object.(*v1.Service)
	domainProject := Kubernetes().GetDomainProject()
	serviceID := generateServiceID(domainProject, svc)
	key := path.GenerateServiceKey(domainProject, serviceID)

	if !ShouldRegisterService(svc) {
		kv := c.Cache().Get(key)
		if kv != nil {
			c.Notify(pb.EVT_DELETE, key, kv)
		}
		return
	}

	switch evt.EventType {
	case pb.EVT_CREATE, pb.EVT_UPDATE:
		ms := FromK8sService(domainProject, svc)
		kv := AsKeyValue(key, ms, svc.ResourceVersion)
		if c.Cache().Get(key) == nil {
			evt.EventType = pb.EVT_CREATE
		}
		c.Notify(evt.EventType, key, kv)
	case pb.EVT_DELETE:
		// service
		kv := c.Cache().Get(key)
		if kv != nil {
			c.Notify(evt.EventType, key, kv)
		}
	}
}

func NewServiceCacher(c *sd.CommonCacher) (s *ServiceCacher) {
	s = &ServiceCacher{CommonCacher: c}
	Kubernetes().AppendEventFunc(TypeService, s.onServiceEvent)
	return
}
