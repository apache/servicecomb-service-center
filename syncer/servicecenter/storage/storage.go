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

package storage

import (
	"context"
	"sync"

	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/util"
	pb "github.com/apache/servicecomb-service-center/syncer/proto"
	"github.com/coreos/etcd/clientv3"
	"github.com/gogo/protobuf/proto"
)

var (
	// mappingsKey the key of instances mapping in etcd
	mappingsKey = "/syncer/v1/mappings"
	// servicesKey the key of service in etcd
	servicesKey = "/syncer/v1/services"
	// instancesKey the key of instance in etcd
	instancesKey = "/syncer/v1/instances"
)

type Storage interface {
	GetData() (data *pb.SyncData)
	UpdateData(data *pb.SyncData)
	GetMaps() (maps pb.SyncMapping)
	UpdateMaps(maps pb.SyncMapping)
	GetMapByCluster(clusterName string) (mapping pb.SyncMapping)
	UpdateMapByCluster(clusterName string, mapping pb.SyncMapping)
}

type storage struct {
	client *clientv3.Client
	data   *pb.SyncData
	lock   sync.RWMutex
}

func NewStorage(client *clientv3.Client) Storage {
	storage := &storage{
		client: client,
		data:   &pb.SyncData{},
	}
	return storage
}

// getPrefixKey Get data from etcd based on the prefix key
func (s *storage) getPrefixKey(prefix string, handler func(key, val []byte) (next bool)) {
	resp, err := s.client.Get(context.Background(), prefix, clientv3.WithPrefix())
	if err != nil {
		log.Errorf(err, "Get mapping from etcd failed: %s", err)
		return
	}

	for _, kv := range resp.Kvs {
		if !handler(kv.Key, kv.Value) {
			break
		}
	}
}

// UpdateData Update data to storage
func (s *storage) UpdateData(data *pb.SyncData) {
	services := s.GetServices()
	s.UpdateServices(data.Services)
	s.compareAndDeleteServices(services, data.Services)

	instances := s.GetInstances()
	s.UpdateInstances(data.Instances)
	s.compareAndDeleteInstances(instances, data.Instances)
}

// GetData Get data from storage
func (s *storage) GetData() (data *pb.SyncData) {
	data = &pb.SyncData{
		Services:  s.GetServices(),
		Instances: s.GetInstances(),
	}
	return
}

func (s *storage) compareAndDeleteServices(origin, renew []*pb.SyncService) {
	expires := make([]*pb.SyncService, 0, len(origin))
next:
	for _, item := range origin {
		for _, service := range renew {
			if service.ServiceId == item.ServiceId {
				continue next
			}
		}
		expires = append(expires, item)
	}
	s.DeleteServices(expires)
}

func (s *storage) compareAndDeleteInstances(origin, renew []*pb.SyncInstance) {
	expires := make([]*pb.SyncInstance, 0, len(origin))
next:
	for _, item := range origin {
		for _, service := range renew {
			if service.InstanceId == item.InstanceId {
				continue next
			}
		}
		expires = append(expires, item)
	}
	s.DeleteInstances(expires)
}

// cleanExpired clean the expired instances in the maps
func (s *storage) cleanExpired(sources pb.SyncMapping, actives pb.SyncMapping) {
next:
	for _, entry := range sources {
		for _, act := range actives {
			if act.CurInstanceID == entry.CurInstanceID {
				continue next
			}
		}
		key := mappingsKey + "/" + entry.ClusterName + "/" + entry.OrgInstanceID
		if _, err := s.client.Delete(context.Background(), key); err != nil {
			log.Errorf(err, "Delete instance clusterName=%s instanceID=%s failed", entry.ClusterName, entry.OrgInstanceID)
		}

		s.deleteInstance(entry.CurInstanceID)
	}
}

// UpdateServices Update services to storage
func (s *storage) UpdateServices(services []*pb.SyncService) {
	for _, val := range services {
		key := servicesKey + "/" + val.ServiceId
		data, err := proto.Marshal(val)
		if err != nil {
			log.Errorf(err, "Proto marshal failed: %s", err)
			continue
		}
		_, err = s.client.Put(context.Background(), key, util.BytesToStringWithNoCopy(data))
		if err != nil {
			log.Errorf(err, "Save service to etcd failed: %s", err)
		}
	}
}

// GetServices Get services from storage
func (s *storage) GetServices() (services []*pb.SyncService) {
	services = make([]*pb.SyncService, 0, 10)
	s.getPrefixKey(servicesKey, func(key, val []byte) (next bool) {
		next = true
		item := &pb.SyncService{}
		if err := proto.Unmarshal(val, item); err != nil {
			log.Errorf(err, "Proto unmarshal '%s' failed: %s", val, err)
			return
		}
		services = append(services, item)
		return
	})
	return
}

// DeleteServices Delete services from storage
func (s *storage) DeleteServices(services []*pb.SyncService) {
	for _, val := range services {
		s.deleteService(val.ServiceId)
	}
}

// DeleteServices Delete services from storage
func (s *storage) deleteService(serviceId string) {
	key := servicesKey + "/" + serviceId
	_, err := s.client.Delete(context.Background(), key)
	if err != nil {
		log.Errorf(err, "Delete service from etcd failed: %s", err)
	}
}

// UpdateInstances Update instances to storage
func (s *storage) UpdateInstances(instances []*pb.SyncInstance) {
	for _, val := range instances {
		key := instancesKey + "/" + val.InstanceId
		data, err := proto.Marshal(val)
		if err != nil {
			log.Errorf(err, "Proto marshal failed: %s", err)
			continue
		}
		_, err = s.client.Put(context.Background(), key, util.BytesToStringWithNoCopy(data))
		if err != nil {
			log.Errorf(err, "Save instance to etcd failed: %s", err)
		}
	}
}

// GetInstances Get instances from storage
func (s *storage) GetInstances() (instances []*pb.SyncInstance) {
	instances = make([]*pb.SyncInstance, 0, 10)
	s.getPrefixKey(instancesKey, func(key, val []byte) (next bool) {
		next = true
		item := &pb.SyncInstance{}
		if err := proto.Unmarshal(val, item); err != nil {
			log.Errorf(err, "Proto unmarshal '%s' failed: %s", val, err)
			return
		}
		instances = append(instances, item)
		return
	})
	return
}

// DeleteInstances Delete instances from storage
func (s *storage) DeleteInstances(instances []*pb.SyncInstance) {
	for _, val := range instances {
		s.deleteInstance(val.InstanceId)
	}
}

func (s *storage) deleteInstance(instanceID string) {
	key := instancesKey + "/" + instanceID
	_, err := s.client.Delete(context.Background(), key)
	if err != nil {
		log.Errorf(err, "Delete instance from etcd failed: %s", err)
	}
}

// UpdateMapByCluster update map to storage by clusterName of other cluster
func (s *storage) UpdateMapByCluster(clusterName string, mapping pb.SyncMapping) {
	newMaps := make(pb.SyncMapping, 0, len(mapping))
	for _, val := range mapping {
		key := mappingsKey + "/" + clusterName + "/" + val.OrgInstanceID
		data, err := proto.Marshal(val)
		if err != nil {
			log.Errorf(err, "Proto marshal failed: %s", err)
			continue
		}
		_, err = s.client.Put(context.Background(), key, util.BytesToStringWithNoCopy(data))
		if err != nil {
			log.Errorf(err, "Save mapping to etcd failed: %s", err)
		}
		newMaps = append(newMaps, val)
	}
	s.cleanExpired(s.GetMapByCluster(clusterName), newMaps)
}

// GetMapByCluster get map by clusterName of other cluster
func (s *storage) GetMapByCluster(clusterName string) (mapping pb.SyncMapping) {
	maps := make(pb.SyncMapping, 0, 10)
	s.getPrefixKey(mappingsKey+"/"+clusterName, func(key, val []byte) (next bool) {
		next = true
		item := &pb.MappingEntry{}
		if err := proto.Unmarshal(val, item); err != nil {
			log.Errorf(err, "Proto unmarshal '%s' failed: %s", val, err)
			return
		}

		maps = append(maps, item)
		return
	})
	return maps
}

// UpdateMaps update all maps to etcd
func (s *storage) UpdateMaps(maps pb.SyncMapping) {
	srcMaps := s.GetMaps()
	mappings := make(pb.SyncMapping, 0, len(maps))
	for _, val := range maps {
		key := mappingsKey + "/" + val.ClusterName + "/" + val.OrgInstanceID
		data, err := proto.Marshal(val)
		if err != nil {
			log.Errorf(err, "Proto marshal failed: %s", err)
			continue
		}
		_, err = s.client.Put(context.Background(), key, util.BytesToStringWithNoCopy(data))
		if err != nil {
			log.Errorf(err, "Save mapping to etcd failed: %s", err)
			continue
		}
		mappings = append(mappings, val)
	}
	s.cleanExpired(srcMaps, mappings)
}

// GetMaps Get maps from storage
func (s *storage) GetMaps() (mapping pb.SyncMapping) {
	maps := make(pb.SyncMapping, 0, 10)
	s.getPrefixKey(mappingsKey, func(key, val []byte) (next bool) {
		next = true
		item := &pb.MappingEntry{}
		if err := proto.Unmarshal(val, item); err != nil {
			log.Errorf(err, "Proto unmarshal '%s' failed: %s", val, err)
			return
		}
		maps = append(maps, item)
		return
	})
	return maps
}
