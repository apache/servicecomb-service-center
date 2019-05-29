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
package servicecenter

import (
	"context"

	"github.com/apache/servicecomb-service-center/pkg/log"
	scpb "github.com/apache/servicecomb-service-center/server/core/proto"
	pb "github.com/apache/servicecomb-service-center/syncer/proto"
)

// Send an instance heartbeat if the instance has already been registered
func (s *servicecenter) heartbeatInstances(mapping pb.SyncMapping, instance *scpb.MicroServiceInstance) bool {
	index := mapping.OriginIndex(instance.InstanceId)
	if index == -1 {
		return false
	}

	item := mapping[index]
	err := s.servicecenter.Heartbeat(context.Background(), item.DomainProject, item.CurServiceID, item.CurInstanceID)
	if err != nil {
		log.Errorf(err, "Servicecenter heartbeat instance failed")
	}
	log.Debugf("Instance %s is already exist, sent heartbeat to service-center")
	instance.InstanceId = item.CurInstanceID
	return true
}

func (s *servicecenter) createService(service *pb.SyncService) string {
	ctx := context.Background()
	serviceID, _ := s.servicecenter.ServiceExistence(ctx, service.DomainProject, service.Service)
	if serviceID != "" {
		return serviceID
	}
	service.Service.ServiceId = ""
	serviceID, err := s.servicecenter.CreateService(ctx, service.DomainProject, service.Service)
	if err != nil {
		log.Errorf(err, "Servicecenter create service failed")
		return ""
	}
	log.Debugf("Create service successful, serviceID = %s", serviceID)
	service.Service.ServiceId = serviceID
	return serviceID
}

func (s *servicecenter) registryInstances(domainProject, serviceId string, instance *scpb.MicroServiceInstance) string {
	instance.ServiceId = serviceId
	instance.InstanceId = ""
	instanceID, err := s.servicecenter.RegisterInstance(context.Background(), domainProject, serviceId, instance)
	if err != nil {
		log.Errorf(err, "Servicecenter registry instance failed")
		return ""
	}
	log.Debugf("Registered instance successful, instanceID = %s", instanceID)
	instance.InstanceId = instanceID
	return instanceID
}

// DeleteInstances Unregister instances of mapping table that has been unregistered from other servicecenter
func (s *servicecenter) unRegistryInstances(data *pb.SyncData, mapping pb.SyncMapping) pb.SyncMapping {
	ctx := context.Background()
	nm := make(pb.SyncMapping, 0, len(mapping))
next:
	for _, val := range mapping {
		for _, svc := range data.Services {
			for _, inst := range svc.Instances {
				if val.CurInstanceID == inst.InstanceId {
					nm = append(nm, val)
					continue next
				}
			}
		}

		err := s.servicecenter.UnregisterInstance(ctx, val.DomainProject, val.CurServiceID, val.CurInstanceID)
		if err != nil {
			log.Errorf(err, "Servicecenter delete instance failed")
		}
		log.Debugf("Unregistered instance, InstanceID = %s", val.CurInstanceID)
	}
	return nm
}
