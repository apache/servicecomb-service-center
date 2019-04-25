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
	"github.com/apache/servicecomb-service-center/pkg/log"

	"github.com/apache/servicecomb-service-center/syncer/notify"
	"github.com/apache/servicecomb-service-center/syncer/pkg/events"
	"github.com/apache/servicecomb-service-center/syncer/plugins"
	"github.com/apache/servicecomb-service-center/syncer/plugins/repository"
	"github.com/apache/servicecomb-service-center/syncer/plugins/storage"
	pb "github.com/apache/servicecomb-service-center/syncer/proto"
)

type Store interface {
	OnEvent(event events.ContextEvent)
	LocalInfo() *pb.SyncData
	Stop()
}

type store struct {
	repo  repository.Repository
	cache storage.Repository
}

func NewStore(endpoints []string) (Store, error) {
	repo, err := plugins.Plugins().Repository().New(endpoints)
	if err != nil {
		return nil, err
	}

	return &store{
		repo:  repo,
		cache: plugins.Plugins().Storage(),
	}, nil
}

func (s *store) Stop() {
	if s.cache == nil{
		return
	}
	s.cache.Stop()
}

func (s *store) LocalInfo() *pb.SyncData {
	return s.cache.GetSyncData()
}

func (s *store) OnEvent(event events.ContextEvent) {
	switch event.Type() {
	case notify.EventTicker:
		s.getLocalDataInfo(event)
	case notify.EventPullByPeer:
		s.syncPeerDataInfo(event)
	default:
	}
}

func (s *store) getLocalDataInfo(event events.ContextEvent) {
	ctx := event.Context()
	data, err := s.repo.GetAll(ctx)
	if err != nil {
		log.Errorf(err, "Syncer discover instances failed")
		return
	}
	s.exclude(data)
	s.cache.SaveSyncData(data)

	events.Dispatch(events.NewContextEvent(notify.EventDiscovery, nil))
}

func (s *store) syncPeerDataInfo(event events.ContextEvent) {
	ctx := event.Context()
	nodeData, ok := ctx.Value(event.Type()).(*pb.NodeDataInfo)
	if !ok {
		log.Error("save peer info failed", nil)
		return
	}
	s.sync(nodeData)
}
