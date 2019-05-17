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
		syncData:    &pb.SyncData{},
		intsMapping: loadSnapshot(),
	}
}

// Storage of Syncer
type Storage struct {
	syncData *pb.SyncData

	// mapping table for other datacenter instances
	intsMapping map[string]pb.SyncMapping
	lock        sync.RWMutex
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
	return mapping
}

func (r *Storage) Stop() {
	r.flush()
}

// flush Refresh the mapping table to the hard disk
func (r *Storage) flush() {
	data, err := json.Marshal(&r.intsMapping)
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
}

// SaveSyncData Save self sync data
func (r *Storage) SaveSyncData(data *pb.SyncData) {
	r.lock.Lock()
	r.syncData = data
	r.lock.Unlock()
}

// GetSyncData Get self sync data
func (r *Storage) GetSyncData() (data *pb.SyncData) {
	r.lock.RLock()
	data = &pb.SyncData{Services: r.syncData.Services[:]}
	r.lock.RUnlock()
	return
}

// SaveSyncMapping Save mapping table for other datacenter instances
func (r *Storage) SaveSyncMapping(nodeName string, mapping pb.SyncMapping) {
	r.lock.Lock()
	r.intsMapping[nodeName] = mapping
	r.lock.Unlock()
}

// GetSyncMapping Get mapping table for other datacenter instances
func (r *Storage) GetSyncMapping(nodeName string) (mapping pb.SyncMapping) {
	r.lock.RLock()
	data, ok := r.intsMapping[nodeName]
	if !ok {
		data = defaultMapping
	}
	r.lock.RUnlock()
	return data
}

// GetAllMapping Get all mapping table for other datacenters instances
func (r *Storage) GetAllMapping() (mapping pb.SyncMapping) {
	r.lock.RLock()
	mapping = make(pb.SyncMapping, 0, 10)
	for _, data := range r.intsMapping {
		if data != nil {
			mapping = append(mapping, data...)
		}
	}
	r.lock.RUnlock()
	return
}
