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
	"github.com/apache/servicecomb-service-center/syncer/plugins"
	pb "github.com/apache/servicecomb-service-center/syncer/proto"
)

// Store interface of datacenter
type DataCenter interface {
	FlushData()
	Registry(nodeName string, data *pb.SyncData)
	Discovery() *pb.SyncData
}

type datacenter struct {
	datacenter plugins.Datacenter
	storage    Storage
}

type Storage interface {
	GetData() (data *pb.SyncData)
	UpdateData(data *pb.SyncData)
	GetMaps() (maps pb.SyncMapping)
	UpdateMaps(maps pb.SyncMapping)
	GetMapByNode(nodeName string) (mapping pb.SyncMapping)
	UpdateMapByNode(nodeName string, mapping pb.SyncMapping)
}

// NewStore new store with endpoints
func NewDataCenter(endpoints []string, storage Storage) (DataCenter, error) {
	dc, err := plugins.Plugins().Datacenter().New(endpoints)
	if err != nil {
		return nil, err
	}

	return &datacenter{
		datacenter: dc,
		storage:    storage,
	}, nil
}

// FlushData flush data to datacenter, update mapping data
func (s *datacenter) FlushData() {
	data, err := s.datacenter.GetAll(context.Background())
	if err != nil {
		log.Errorf(err, "Syncer discover instances failed")
	}

	maps := s.storage.GetMaps()

	data, maps = s.exclude(data, maps)
	s.storage.UpdateData(data)
	s.storage.UpdateMaps(maps)
}

// Registry registry data to the datacenter, update mapping data
func (s *datacenter) Registry(nodeName string, data *pb.SyncData) {
	mapping := s.storage.GetMapByNode(nodeName)
	for _, svc := range data.Services {
		log.Debugf("trying to do registration of service, serviceID = %s", svc.Service.ServiceId)
		// If the svc is in the mapping, just do nothing, if not, created it in datacenter and get the new serviceID
		svcID := s.createService(svc)
		for _, inst := range svc.Instances {
			// If inst is in the mapping, just heart beat it in datacenter
			log.Debugf("trying to do registration of instance, instanceID = %s", inst.InstanceId)
			if s.heartbeatInstances(mapping, inst) {
				continue
			}

			// If inst is not in the mapping, that is because this the first time syncer get the instance data
			// in this case, we should registry it to the datacenter and get the new instanceID
			item := &pb.MappingItem{
				DomainProject: svc.DomainProject,
				OrgServiceID:  inst.ServiceId,
				OrgInstanceID: inst.InstanceId,
				CurServiceID:  svcID,
				NodeName:      nodeName,
			}
			item.CurInstanceID = s.registryInstances(svc.DomainProject, svcID, inst)

			// Use new serviceID and instanceID to update mapping data in this datacenter
			if item.CurInstanceID != "" {
				mapping = append(mapping, item)
			}
		}
	}
	// UnRegistry instances that is not in the data which means the instance in the mapping is no longer actived
	mapping = s.unRegistryInstances(data, mapping)
	// Update mapping data of the node to the storage of the datacenter
	s.storage.UpdateMapByNode(nodeName, mapping)
}

// Discovery discovery data from storage
func (s *datacenter) Discovery() *pb.SyncData {
	return s.storage.GetData()
}
