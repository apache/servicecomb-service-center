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
	"sync"

	pb "github.com/apache/servicecomb-service-center/syncer/proto"
)

var (
	defaultMapping = make(pb.SyncMapping, 0)
)

func New() *Storage {
	return &Storage{
		data: &pb.SyncData{},
		maps: make(map[string]pb.SyncMapping),
	}
}

// Storage of Syncer
type Storage struct {
	data *pb.SyncData
	// mapping table for other servicecenter instances
	maps map[string]pb.SyncMapping
	lock sync.RWMutex
}

// UpdateData Update data to storage
func (s *Storage) UpdateData(data *pb.SyncData) {
	s.lock.Lock()
	s.data = data
	s.lock.Unlock()
}

// GetData Get data from storage
func (s *Storage) GetData() (data *pb.SyncData) {
	s.lock.RLock()
	data = s.data
	s.lock.RUnlock()
	return
}

// UpdateMapByNode update map to storage by nodeName of other node
func (s *Storage) UpdateMapByNode(nodeName string, mapping pb.SyncMapping) {
	s.lock.Lock()
	s.maps[nodeName] = mapping
	s.lock.Unlock()
}

// GetMapByNode get map by nodeName of other node
func (s *Storage) GetMapByNode(nodeName string) (mapping pb.SyncMapping) {
	s.lock.RLock()
	data, ok := s.maps[nodeName]
	if !ok {
		data = defaultMapping
	}
	s.lock.RUnlock()
	return data
}

func (s *Storage) UpdateMaps(maps pb.SyncMapping) {
	s.lock.Lock()
	mappings := make(map[string]pb.SyncMapping)
	for _, item := range maps {
		mapping, ok := mappings[item.NodeName]
		if !ok {
			mapping = make(pb.SyncMapping, 0, 10)
		}
		mapping = append(mapping, item)
		mappings[item.NodeName] = mapping
	}
	s.maps = mappings
	s.lock.Unlock()
}

// GetMaps Get maps from storage
func (s *Storage) GetMaps() (mapping pb.SyncMapping) {
	s.lock.RLock()
	mapping = make(pb.SyncMapping, 0, 10)
	for _, data := range s.maps {
		if data != nil {
			mapping = append(mapping, data...)
		}
	}
	s.lock.RUnlock()
	return
}
