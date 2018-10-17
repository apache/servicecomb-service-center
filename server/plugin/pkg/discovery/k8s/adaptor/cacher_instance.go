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
	"github.com/apache/incubator-servicecomb-service-center/pkg/util"
	"github.com/apache/incubator-servicecomb-service-center/server/core"
	pb "github.com/apache/incubator-servicecomb-service-center/server/core/proto"
	"github.com/apache/incubator-servicecomb-service-center/server/plugin/pkg/discovery"
	"k8s.io/api/core/v1"
	"reflect"
	"strconv"
)

type InstanceCacher struct {
	*discovery.CommonCacher
}

// onServiceEvent is the method to refresh service cache
func (c *InstanceCacher) onServiceEvent(evt K8sEvent) {
	svc := evt.Object.(*v1.Service)
	domainProject := Kubernetes().GetDomainProject()
	serviceId := string(svc.UID)

	switch evt.EventType {
	case pb.EVT_DELETE:
		c.deleteInstances(domainProject, serviceId)
	case pb.EVT_UPDATE:
		if !ShouldRegisterService(svc) {
			c.deleteInstances(domainProject, serviceId)
			return
		}
		ep := Kubernetes().GetEndpoints(svc.Namespace, svc.Name)
		c.onEndpointsEvent(K8sEvent{pb.EVT_CREATE, ep, nil})
	}
}

func (c *InstanceCacher) getInstances(domainProject, serviceId string) (m map[string]*discovery.KeyValue) {
	var arr []*discovery.KeyValue
	key := core.GenerateInstanceKey(domainProject, serviceId, "")
	if l := c.Cache().GetPrefix(key, &arr); l > 0 {
		m = make(map[string]*discovery.KeyValue, l)
		for _, kv := range arr {
			m[util.BytesToStringWithNoCopy(kv.Key)] = kv
		}
	}
	return
}

func (c *InstanceCacher) deleteInstances(domainProject, serviceId string) {
	var kvs []*discovery.KeyValue
	c.Cache().GetPrefix(core.GenerateInstanceKey(domainProject, serviceId, ""), &kvs)
	for _, kv := range kvs {
		key := util.BytesToStringWithNoCopy(kv.Key)
		c.Notify(pb.EVT_DELETE, key, kv)
	}
}

// onEndpointsEvent is the method to refresh instance cache
func (c *InstanceCacher) onEndpointsEvent(evt K8sEvent) {
	ep := evt.Object.(*v1.Endpoints)
	svc := Kubernetes().GetService(ep.Namespace, ep.Name)
	if svc == nil || !ShouldRegisterService(svc) {
		return
	}

	serviceId := string(svc.UID)
	domainProject := Kubernetes().GetDomainProject()
	oldKvs := c.getInstances(domainProject, serviceId)
	newKvs := make(map[string]*discovery.KeyValue)
	for _, ss := range ep.Subsets {
		for _, ea := range ss.Addresses {
			pod := Kubernetes().GetPodByIP(ea.IP)
			if pod == nil {
				continue
			}

			instanceId := string(pod.UID)
			key := core.GenerateInstanceKey(Kubernetes().GetDomainProject(), serviceId, instanceId)
			switch evt.EventType {
			case pb.EVT_CREATE, pb.EVT_UPDATE:
				if pod.Status.Phase != v1.PodRunning {
					continue
				}

				node := Kubernetes().GetNodeByPod(pod)
				if node == nil {
					continue
				}

				inst := &pb.MicroServiceInstance{
					InstanceId:     instanceId,
					ServiceId:      serviceId,
					HostName:       pod.Name,
					Status:         pb.MSI_UP,
					DataCenterInfo: &pb.DataCenterInfo{},
					Timestamp:      strconv.FormatInt(pod.CreationTimestamp.Unix(), 10),
					Version:        getLabel(svc.Labels, LabelVersion, pb.VERSION),
					Properties: map[string]string{
						PropNodeIP: pod.Status.HostIP,
					},
				}
				inst.DataCenterInfo.Region, inst.DataCenterInfo.AvailableZone = getRegionAZ(node)
				inst.ModTimestamp = inst.Timestamp
				for _, port := range ss.Ports {
					inst.Endpoints = append(inst.Endpoints, generateEndpoint(ea.IP, port))
				}

				old := c.Cache().Get(key)
				kv := AsKeyValue(key, inst, pod.ResourceVersion)
				newKvs[key] = kv

				if old == nil {
					c.Notify(pb.EVT_CREATE, key, kv)
				} else if !reflect.DeepEqual(old, kv) {
					c.Notify(pb.EVT_UPDATE, key, kv)
				}
			case pb.EVT_DELETE:
			}
		}
	}
	for k, v := range oldKvs {
		if _, ok := newKvs[k]; !ok {
			c.Notify(pb.EVT_DELETE, k, v)
		}
	}
}

func NewInstanceCacher(c *discovery.CommonCacher) (i *InstanceCacher) {
	i = &InstanceCacher{CommonCacher: c}
	Kubernetes().AppendEventFunc(TypeService, i.onServiceEvent)
	Kubernetes().AppendEventFunc(TypeEndpoint, i.onEndpointsEvent)
	return
}
