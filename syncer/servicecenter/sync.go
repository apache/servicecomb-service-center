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
	"fmt"

	"github.com/apache/servicecomb-service-center/pkg/log"
	pb "github.com/apache/servicecomb-service-center/syncer/proto"
)

// Send an instance heartbeat if the instance has already been registered
func (s *servicecenter) heartbeatInstances(mapping pb.SyncMapping, instance *pb.SyncInstance) bool {
	index := mapping.OriginIndex(instance.InstanceId)
	if index == -1 {
		return false
	}

	item := mapping[index]
	err := s.servicecenter.Heartbeat(context.Background(), item.DomainProject, item.CurServiceID, item.CurInstanceID)
	if err != nil {
		log.Errorf(err, "Servicecenter heartbeat instance failed")
	}
	log.Debugf("Instance %s is already exist, sent heartbeat to service-center", item.OrgInstanceID)
	return true
}

func (s *servicecenter) createService(service *pb.SyncService) (string, error) {
	ctx := context.Background()
	serviceID, err := s.servicecenter.ServiceExistence(ctx, service.DomainProject, service)
	if err != nil {
		log.Warn(fmt.Sprintf("get service existence failed %s", err))
	}

	if serviceID == "" {
		serviceID, err = s.servicecenter.CreateService(ctx, service.DomainProject, service)
		if err != nil {
			log.Error("create service failed", err)
			return "", err
		}
		log.Debug(fmt.Sprintf("create service successful, serviceID = %s", serviceID))
	} else {
		log.Debug(fmt.Sprintf("service already exists, serviceID = %s", serviceID))
	}
	return serviceID, nil
}

func (s *servicecenter) registryInstances(domainProject, serviceID string, instance *pb.SyncInstance) string {
	instanceID, err := s.servicecenter.RegisterInstance(context.Background(), domainProject, serviceID, instance)
	if err != nil {
		log.Errorf(err, "Servicecenter registry instance failed")
		return ""
	}
	log.Debugf("Registered instance successful, instanceID = %s", instanceID)
	return instanceID
}

// DeleteInstances Unregister instances of mapping table that has been unregistered from other servicecenter
func (s *servicecenter) unRegistryInstances(data *pb.SyncData, mapping pb.SyncMapping) pb.SyncMapping {
	ctx := context.Background()
	nm := make(pb.SyncMapping, 0, len(mapping))
next:
	for _, val := range mapping {
		for _, inst := range data.Instances {
			if val.OrgInstanceID == inst.InstanceId {
				nm = append(nm, val)
				continue next
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
