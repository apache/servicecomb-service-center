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
	"encoding/json"
	"io/ioutil"
	"sync"

	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/syncer/pkg/utils"
	pb "github.com/apache/servicecomb-service-center/syncer/proto"
)

var (
	defaultMapping = make(pb.SyncMapping, 0)
	snapshotPath   = "./data/syncer-snapshot"
)

func New() *Storage {
	return &Storage{
		data: &pb.SyncData{},
		maps: loadSnapshot(),
	}
}

// Storage of Syncer
type Storage struct {
	data *pb.SyncData
	// mapping table for other servicecenter instances
	maps map[string]pb.SyncMapping
	lock sync.RWMutex
}

// loadSnapshot Load snapshot of mapping table
func loadSnapshot() map[string]pb.SyncMapping {
	mapping := make(map[string]pb.SyncMapping)
	data, err := ioutil.ReadFile(snapshotPath)
	if err != nil {
		log.Warnf("get syncer snapshot from '%s' failed, error: %s", snapshotPath, err)
		return mapping
	}
	err = json.Unmarshal(data, &mapping)
	if err != nil {
		log.Warnf("unmarshal syncer snapshot failed, error: %s", err)
	}
	log.Infof("Loaded maps from disk to storage, maps = %s", data)
	return mapping
}

func (s *Storage) Stop() {
	s.flush()
}

// flush Refresh the mapping table to the hard disk
func (s *Storage) flush() {
	data, err := json.Marshal(&s.maps)
	if err != nil {
		log.Warnf("marshal syncer snapshot failed, error: %s", err)
		return
	}

	f, err := utils.OpenFile(snapshotPath)
	if err != nil {
		log.Warnf("open syncer snapshot file '%s' failed, error: %s", snapshotPath, err)
		return
	}
	defer f.Close()

	_, err = f.Write(data)
	if err != nil {
		log.Warnf("flush syncer snapshot to '%s' failed, error: %s", snapshotPath, err)
		return
	}
	log.Infof("Flushed maps to disk before exit, data = %s", data)
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
