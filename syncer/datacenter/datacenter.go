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
	GetSyncData(allMapping pb.SyncMapping) (*pb.SyncData, error)
	SetSyncData(*pb.SyncData, pb.SyncMapping) (pb.SyncMapping, error)
}

type store struct {
	datacenter plugins.Datacenter
}

// NewStore new store with endpoints
func NewDataCenter(endpoints []string) (DataCenter, error) {
	datacenter, err := plugins.Plugins().Datacenter().New(endpoints)
	if err != nil {
		return nil, err
	}

	return &store{
		datacenter: datacenter,
	}, nil
}

// GetSyncData Get data from datacenter instance, excluded the data from other syncer
func (s *store) GetSyncData(allMapping pb.SyncMapping) (*pb.SyncData, error) {
	data, err := s.datacenter.GetAll(context.Background())
	if err != nil {
		log.Errorf(err, "Syncer discover instances failed")
		return nil, err
	}
	s.exclude(data, allMapping)
	return data, nil
}

// GetSyncData Get current datacenter information
func (s *store) SetSyncData(data *pb.SyncData, mapping pb.SyncMapping) (pb.SyncMapping, error) {
	return s.sync(data, mapping)
}
