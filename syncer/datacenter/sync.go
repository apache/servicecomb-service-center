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
	pb "github.com/apache/servicecomb-service-center/syncer/proto"
)

// sync synchronize information from other datacenters
func (s *store) sync(data *pb.SyncData, curMapping pb.SyncMapping) (pb.SyncMapping, error) {
	orgMapping, curMapping := s.syncServiceInstances(data.Services, curMapping)
	return s.deleteInstances(orgMapping, curMapping), nil
}

// syncServiceInstances register instances from other datacenter
func (s *store) syncServiceInstances(services []*pb.SyncService, curMapping pb.SyncMapping) (pb.SyncMapping, pb.SyncMapping) {
	var err error
	ctx := context.Background()
	orgMapping := make([]*pb.MappingItem, 0, 10)

	for _, svc := range services {
		curServiceID := ""
		for _, inst := range svc.Instances {
			var item *pb.MappingItem

			// Send an instance heartbeat if the instance has already been registered
			if index := curMapping.OriginIndex(inst.InstanceId); index != -1 {
				item = curMapping[index]
				curServiceID = item.CurServiceID
				err = s.datacenter.Heartbeat(ctx, item.DomainProject, item.CurServiceID, item.CurInstanceID)
				if err != nil {
					log.Errorf(err, "Syncer heartbeat instance failed")
				} else {
					orgMapping = append(orgMapping, item)
				}
				continue
			}

			// Create microservice if the service to which the instance belongs does not exist
			if curServiceID == "" {
				if curServiceID, _ = s.datacenter.ServiceExistence(ctx, svc.DomainProject, svc.Service); curServiceID == "" {
					svc.Service.ServiceId = ""
					curServiceID, err = s.datacenter.CreateService(ctx, svc.DomainProject, svc.Service)
					if err != nil {
						log.Errorf(err, "Syncer create service failed")
						continue
					}
				}
			}
			// Register instance information when the instance does not exist
			item = &pb.MappingItem{
				DomainProject: svc.DomainProject,
				OrgServiceID:  inst.ServiceId,
				OrgInstanceID: inst.InstanceId,
				CurServiceID:  curServiceID,
			}
			inst.ServiceId = curServiceID
			inst.InstanceId = ""
			curInstanceID, err := s.datacenter.RegisterInstance(ctx, svc.DomainProject, curServiceID, inst)
			if err != nil {
				log.Errorf(err, "Syncer create service failed")
				continue
			}

			item.CurInstanceID = curInstanceID
			orgMapping = append(orgMapping, item)
			curMapping = append(curMapping, item)
		}
	}
	return orgMapping, curMapping
}

// deleteInstances Unregister instances of mapping table that has been unregistered from other datacenter
func (s *store) deleteInstances(orgMapping, curMapping pb.SyncMapping) pb.SyncMapping {
	l := 0
	ol := len(orgMapping)
	cl := len(curMapping)
	if ol < cl {
		l = cl - ol
	} else {
		l = ol - cl
	}
	if l == 0 {
		return curMapping
	}

	ctx := context.Background()
	nm := make(pb.SyncMapping, 0, l)
	for _, val := range curMapping {
		if index := orgMapping.CurrentIndex(val.CurInstanceID); index != -1 {
			nm = append(nm, val)
			continue
		}

		err := s.datacenter.UnregisterInstance(ctx, val.DomainProject, val.CurServiceID, val.CurInstanceID)
		if err != nil {
			log.Errorf(err, "Syncer delete service failed")
		}
	}
	return nm
}
