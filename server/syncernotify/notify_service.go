// Licensed to the Apache Software Foundation (ASF) under one or more
// contributor license agreements.  See the NOTICE file distributed with
// this work for additional information regarding copyright ownership.
// The ASF licenses this file to You under the Apache License, Version 2.0
// (the "License"); you may not use this file except in compliance with
// the License.  You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package syncernotify

import (
	pb "github.com/apache/servicecomb-service-center/pkg/dump"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"sync"
)

var syncerNotifyService *Service

func init() {
	syncerNotifyService = NewSyncerNotifyService()
}

func GetSyncerNotifyCenter() *Service {
	return syncerNotifyService
}

type Service struct {
	instEventCh chan *pb.WatchInstanceChangedEvent
	mux         sync.RWMutex
	isClose     bool
}

func NewSyncerNotifyService() *Service {
	return &Service{
		instEventCh: make(chan *pb.WatchInstanceChangedEvent, InstanceEventQueueSize),
		isClose:     true,
	}
}

func (s *Service) AddEvent(event *pb.WatchInstanceChangedEvent) {
	s.instEventCh <- event
	log.Debugf("add instance event to instance event channel, instEventCh len is:%s", len(s.instEventCh))
}

func (s *Service) Start() {
	if !s.Closed() {
		log.Warnf("syncer notify service is already running")
		return
	}

	s.mux.Lock()
	s.isClose = false
	s.mux.Unlock()

	log.Debugf("syncer notify service is started")
}

func (s *Service) Closed() (b bool) {
	s.mux.RLock()
	b = s.isClose
	s.mux.RUnlock()
	return b
}

func (s *Service) Stop() {
	if s.Closed() {
		return
	}
	s.mux.Lock()
	s.isClose = true
	s.mux.Unlock()

	log.Debug("syncer notify service stopped")
}
