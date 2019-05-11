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
package datacenter

import (
	"context"
	"github.com/apache/servicecomb-service-center/pkg/log"

	scpb "github.com/apache/servicecomb-service-center/server/core/proto"
	pb "github.com/apache/servicecomb-service-center/syncer/proto"
)

// sync synchronize information from other datacenters
func (s *store) sync(nodeData *pb.NodeDataInfo) {
	allInstances, mapping := s.syncServiceInstances(nodeData.DataInfo.Services, s.cache.GetSyncMapping(nodeData.NodeName))
	mapping = s.deleteInstances(allInstances, mapping)
	s.cache.SaveSyncMapping(nodeData.NodeName, mapping)
}

// syncServiceInstances register instances from other datacenter
func (s *store) syncServiceInstances(services []*pb.SyncService, mapping pb.SyncMapping) ([]*scpb.MicroServiceInstance, pb.SyncMapping) {
	var err error
	ctx := context.Background()
	allInstances := make([]*scpb.MicroServiceInstance, 0, 10)
	for _, svc := range services {
		serviceID := ""
		for _, inst := range svc.Instances {
			allInstances = append(allInstances, inst)

			// Send an instance heartbeat if the instance has already been registered
			syncKey, ok := mapping[inst.InstanceId]
			if ok && syncKey.InstanceID != ""{
				err = s.repo.Heartbeat(ctx, svc.DomainProject, syncKey.ServiceID, syncKey.InstanceID)
				if err != nil {
					log.Errorf(err, "Syncer heartbeat instance failed")
				}
				continue
			}

			// Create microservice if the service to which the instance belongs does not exist
			if serviceID == "" {
				if serviceID, _ = s.repo.ServiceExistence(ctx, svc.DomainProject, svc.Service); serviceID == "" {
					serviceID, err = s.repo.CreateService(ctx, svc.DomainProject, svc.Service)
					if err != nil {
						log.Errorf(err, "Syncer create service failed")
					}
				}
			}

			// Register instance information when the instance does not exist
			instanceID, err := s.repo.RegisterInstance(ctx, svc.DomainProject, serviceID, inst)
			if err != nil {
				log.Errorf(err, "Syncer create service failed")
				continue
			}

			mapping[inst.InstanceId] = &pb.SyncServiceKey{
				DomainProject: svc.DomainProject,
				ServiceID:     serviceID,
				InstanceID:    instanceID,
			}
		}
	}
	return allInstances, mapping
}

// deleteInstances Unregister instances of mapping table that has been unregistered from other datacenter
func (s *store) deleteInstances(ins []*scpb.MicroServiceInstance, mapping pb.SyncMapping) pb.SyncMapping {
	ctx := context.Background()
	nm := make(pb.SyncMapping)
	for key, val := range mapping {
		skip := false
		for _, instance := range ins {
			skip = instance.InstanceId == key
			if skip {
				break
			}
		}
		if skip {
			nm[key] = val
			continue
		}

		err := s.repo.UnregisterInstance(ctx, val.DomainProject, val.ServiceID, val.InstanceID)
		if err != nil {
			log.Errorf(err,"Syncer delete service failed")
		}
	}
	return nm
}
