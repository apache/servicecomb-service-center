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

package k8s

import (
	"github.com/apache/incubator-servicecomb-service-center/server/core"
	pb "github.com/apache/incubator-servicecomb-service-center/server/core/proto"
	"k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ServiceCacher struct {
	*K8sCacher
}

// onServiceEvent is the method to refresh service cache
func (c *ServiceCacher) onServiceEvent(evt K8sEvent) {
	svc := evt.Object.(*v1.Service)
	if svc.Namespace == meta.NamespaceSystem {
		return
	}

	domainProject := Kubernetes().GetDomainProject()
	serviceId := string(svc.UID)
	key := core.GenerateServiceKey(domainProject, serviceId)

	switch evt.EventType {
	case pb.EVT_CREATE, pb.EVT_UPDATE:
		ms := FromK8sService(svc)
		kv := AsKeyValue(key, ms, svc.ResourceVersion)
		c.Notify(evt.EventType, key, kv)
	case pb.EVT_DELETE:
		// service
		kv := c.Cache().Get(key)
		if kv != nil {
			c.Notify(evt.EventType, key, kv)
		}
	}
}

func NewServiceCacher(c *K8sCacher) (s *ServiceCacher) {
	s = &ServiceCacher{K8sCacher: c}
	Kubernetes().AppendEventFunc(TypeService, s.onServiceEvent)
	return
}
