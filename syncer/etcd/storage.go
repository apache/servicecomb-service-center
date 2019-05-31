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

package etcd

import (
	"context"
	"sync"

	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/util"
	pb "github.com/apache/servicecomb-service-center/syncer/proto"
	"github.com/apache/servicecomb-service-center/syncer/servicecenter"
	"github.com/coreos/etcd/clientv3"
	"github.com/gogo/protobuf/proto"
)

var (
	// mappingsKey the key of instances mapping in etcd
	mappingsKey = "/syncer/v1/mappings"
)

type storage struct {
	client *clientv3.Client
	data   *pb.SyncData
	lock   sync.RWMutex
}

func NewStorage(client *clientv3.Client) servicecenter.Storage {
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
	s.lock.Lock()
	s.data = data
	s.lock.Unlock()
}

// GetData Get data from storage
func (s *storage) GetData() (data *pb.SyncData) {
	s.lock.RLock()
	data = s.data
	s.lock.RUnlock()
	return
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
		key := mappingsKey + "/" + entry.NodeName + "/" + entry.OrgInstanceID
		if _, err := s.client.Delete(context.Background(), key); err != nil {
			log.Errorf(err, "Delete instance nodeName=%s instanceID=%s failed", entry.NodeName, entry.OrgInstanceID)
		}
	}
}

// UpdateMapByNode update map to storage by nodeName of other node
func (s *storage) UpdateMapByNode(nodeName string, mapping pb.SyncMapping) {
	newMaps := make(pb.SyncMapping, 0, len(mapping))
	for _, val := range mapping {
		key := mappingsKey + "/" + nodeName + "/" + val.OrgInstanceID
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
	s.cleanExpired(s.GetMapByNode(nodeName), newMaps)
}

// GetMapByNode get map by nodeName of other node
func (s *storage) GetMapByNode(nodeName string) (mapping pb.SyncMapping) {
	maps := make(pb.SyncMapping, 0, 10)
	s.getPrefixKey(mappingsKey+"/"+nodeName, func(key, val []byte) (next bool) {
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
		key := mappingsKey + "/" + val.NodeName + "/" + val.OrgInstanceID
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
