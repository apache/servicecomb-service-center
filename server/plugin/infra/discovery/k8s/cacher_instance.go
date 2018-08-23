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
	"github.com/apache/incubator-servicecomb-service-center/pkg/util"
	"github.com/apache/incubator-servicecomb-service-center/server/core"
	pb "github.com/apache/incubator-servicecomb-service-center/server/core/proto"
	"github.com/apache/incubator-servicecomb-service-center/server/infra/discovery"
	"k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"reflect"
	"strconv"
)

type InstanceCacher struct {
	*K8sCacher
}

// onServiceEvent is the method to refresh service cache
func (c *InstanceCacher) onServiceEvent(evt K8sEvent) {
	svc := evt.Object.(*v1.Service)
	if svc.Namespace == meta.NamespaceSystem {
		return
	}

	domainProject := Kubernetes().GetDomainProject()
	serviceId := string(svc.UID)
	instKey := core.GenerateInstanceKey(domainProject, serviceId, "")

	switch evt.EventType {
	case pb.EVT_DELETE:
		// instances
		var kvs []*discovery.KeyValue
		c.cache.GetPrefix(instKey, &kvs)
		for _, kv := range kvs {
			key := util.BytesToStringWithNoCopy(kv.Key)
			c.notify(pb.EVT_DELETE, key, kv)
		}
	}
}

func (c *InstanceCacher) getInstances(serviceId string) (m map[string]*discovery.KeyValue) {
	var arr []*discovery.KeyValue
	key := core.GenerateInstanceKey(Kubernetes().GetDomainProject(), serviceId, "")
	if l := c.cache.GetPrefix(key, &arr); l > 0 {
		m = make(map[string]*discovery.KeyValue, l)
		for _, kv := range arr {
			m[util.BytesToStringWithNoCopy(kv.Key)] = kv
		}
	}
	return
}

// onEndpointsEvent is the method to refresh instance cache
func (c *InstanceCacher) onEndpointsEvent(evt K8sEvent) {
	ep := evt.Object.(*v1.Endpoints)
	if ep.Namespace == meta.NamespaceSystem {
		return
	}

	svc := Kubernetes().GetService(ep.Namespace, ep.Name)
	if svc == nil {
		return
	}

	serviceId := string(svc.UID)
	oldKvs := c.getInstances(serviceId)
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
				}
				inst.DataCenterInfo.Region, inst.DataCenterInfo.AvailableZone = getRegionAZ(node)
				inst.ModTimestamp = inst.Timestamp
				for _, port := range ss.Ports {
					inst.Endpoints = append(inst.Endpoints, generateEndpoint(ea.IP, port))
				}

				old := c.cache.Get(key)
				kv := FromInstance(key, inst)
				newKvs[key] = kv

				if old == nil {
					c.notify(pb.EVT_CREATE, key, kv)
				} else if !reflect.DeepEqual(old, kv) {
					c.notify(pb.EVT_UPDATE, key, kv)
				}
			case pb.EVT_DELETE:
			}
		}
	}
	for k, v := range oldKvs {
		if _, ok := newKvs[k]; !ok {
			c.notify(pb.EVT_DELETE, k, v)
		}
	}
}

func (c *InstanceCacher) notify(action pb.EventType, key string, kv *discovery.KeyValue) {
	switch action {
	case pb.EVT_DELETE:
		c.cache.Remove(key)
	default:
		c.cache.Put(key, kv)
	}
	c.OnEvents(discovery.KvEvent{Type: action, KV: kv})
}

func NewInstanceCacher(c *K8sCacher) (i *InstanceCacher) {
	i = &InstanceCacher{K8sCacher: c}
	Kubernetes().AppendEventFunc(TypeService, i.onServiceEvent)
	Kubernetes().AppendEventFunc(TypeEndpoint, i.onEndpointsEvent)
	return
}
