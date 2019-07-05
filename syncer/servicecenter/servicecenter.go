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
	"errors"

	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/syncer/plugins"
	pb "github.com/apache/servicecomb-service-center/syncer/proto"
	"github.com/apache/servicecomb-service-center/syncer/servicecenter/storage"
	"github.com/coreos/etcd/clientv3"
)

// Store interface of servicecenter
type Servicecenter interface {
	SetStorage(storage *clientv3.Client)
	FlushData()
	Registry(clusterName string, data *pb.SyncData)
	Discovery() *pb.SyncData
}

type servicecenter struct {
	servicecenter plugins.Servicecenter
	storage       storage.Storage
}

// NewServicecenter new store with endpoints
func NewServicecenter(endpoints []string) (Servicecenter, error) {
	dc, err := plugins.Plugins().Servicecenter().New(endpoints)
	if err != nil {
		return nil, err
	}

	return &servicecenter{
		servicecenter: dc,
	}, nil
}

func (s *servicecenter) SetStorage(client *clientv3.Client) {
	s.storage = storage.NewStorage(client)
}

// FlushData flush data to servicecenter, update mapping data
func (s *servicecenter) FlushData() {
	data, err := s.servicecenter.GetAll(context.Background())
	if err != nil {
		log.Errorf(err, "Syncer discover instances failed")
		return
	}

	maps := s.storage.GetMaps()

	data, maps = s.exclude(data, maps)
	s.storage.UpdateData(data)
	s.storage.UpdateMaps(maps)
}

// Registry registry data to the servicecenter, update mapping data
func (s *servicecenter) Registry(clusterName string, data *pb.SyncData) {
	mapping := s.storage.GetMapByCluster(clusterName)
	for _, inst := range data.Instances {
		svc := searchService(inst, data.Services)
		if svc == nil {
			err := errors.New("service does not exist")
			log.Errorf(err, "servicecenter.Registry, serviceID = %s, instanceId = %s", inst.ServiceId, inst.InstanceId)
			continue
		}

		// If the svc is in the mapping, just do nothing, if not, created it in servicecenter and get the new serviceID
		svcID := s.createService(svc)
		log.Debugf("create service success orgServiceID= %s, curServiceID = %s", inst.ServiceId, svcID)

		// If inst is in the mapping, just heart beat it in servicecenter
		log.Debugf("trying to do registration of instance, instanceID = %s", inst.InstanceId)
		if s.heartbeatInstances(mapping, inst) {
			continue
		}

		// If inst is not in the mapping, that is because this the first time syncer get the instance data
		// in this case, we should registry it to the servicecenter and get the new instanceID
		item := &pb.MappingEntry{
			ClusterName:   clusterName,
			DomainProject: svc.DomainProject,
			OrgServiceID:  svc.ServiceId,
			OrgInstanceID: inst.InstanceId,
			CurServiceID:  svcID,
			CurInstanceID: s.registryInstances(svc.DomainProject, svcID, inst),
		}

		// Use new serviceID and instanceID to update mapping data in this servicecenter
		if item.CurInstanceID != "" {
			mapping = append(mapping, item)
		}
	}
	// UnRegistry instances that is not in the data which means the instance in the mapping is no longer actived
	mapping = s.unRegistryInstances(data, mapping)
	// Update mapping data of the cluster to the storage of the servicecenter
	s.storage.UpdateMapByCluster(clusterName, mapping)
}

// Discovery discovery data from storage
func (s *servicecenter) Discovery() *pb.SyncData {
	return s.storage.GetData()
}
