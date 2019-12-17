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
package server

import (
	"context"
	"crypto/tls"
	"fmt"
	"strconv"

	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/util"
	"github.com/apache/servicecomb-service-center/syncer/grpc"
	pb "github.com/apache/servicecomb-service-center/syncer/proto"
	myserf "github.com/apache/servicecomb-service-center/syncer/serf"
	"github.com/hashicorp/serf/serf"
)

const (
	EventDiscovered = "discovered"
)

// tickHandler Timed task handler
func (s *Server) tickHandler(ctx context.Context) {
	log.Debugf("is leader: %v", s.etcd.IsLeader())
	if !s.etcd.IsLeader() {
		return
	}
	log.Debugf("Handle Tick")
	// Flush data to the storage of servicecenter
	s.servicecenter.FlushData()

	// sends a UserEvent on Serf, the event will be broadcast between members
	err := s.agent.UserEvent(EventDiscovered, util.StringToBytesWithNoCopy(s.conf.ClusterName), true)
	if err != nil {
		log.Errorf(err, "Syncer send user event failed")
	}
}

// GetData Sync Data to GRPC
func (s *Server) Discovery() *pb.SyncData {
	return s.servicecenter.Discovery()
}

// HandleEvent Handles serf.EventUser/serf.EventQuery,
// used for message passing and processing between serf nodes
func (s *Server) HandleEvent(event serf.Event) {
	log.Debugf("is leader: %v", s.etcd.IsLeader())
	if !s.etcd.IsLeader() {
		return
	}

	switch event.EventType() {
	case serf.EventUser:
		s.userEvent(event.(serf.UserEvent))
	case serf.EventQuery:
		s.queryEvent(event.(*serf.Query))
	default:
		log.Infof("serf event = %s", event)
	}
}

// userEvent Handles "EventUser" notification events, no response required
func (s *Server) userEvent(event serf.UserEvent) {
	log.Debug("Receive serf user event")
	clusterName := util.BytesToStringWithNoCopy(event.Payload)

	// Excludes notifications from self, as the gossip protocol inevitably has redundant notifications
	if s.conf.ClusterName == clusterName {
		return
	}

	// Get member information and get synchronized data from it
	members := s.agent.GroupMembers(clusterName)
	if members == nil || len(members) == 0 {
		log.Warnf("serf member = %s is not found", clusterName)
		return
	}

	// todo: grpc supports multi-address polling
	// Get dta from remote member
	endpoint := fmt.Sprintf("%s:%s", members[0].Addr, members[0].Tags[myserf.TagKeyRPCPort])
	log.Debugf("Going to pull data from %s %s", members[0].Name, endpoint)

	enabled, err := strconv.ParseBool(members[0].Tags[myserf.TagKeyTLSEnabled])
	if err != nil {
		log.Warnf("get tls enabled failed, err = %s", err)
	}
	var tlsConfig *tls.Config
	if enabled {
		tlsConfig, err = s.conf.TLSConfig.ClientTlsConfig()
		if err != nil {
			log.Error("get grpc client tls config failed", err)
			return
		}
	}

	data, err := grpc.Pull(context.Background(), endpoint, tlsConfig)
	if err != nil {
		log.Errorf(err, "Pull other serf instances failed, node name is '%s'", members[0].Name)
		return
	}
	// Registry instances to servicecenter and update storage of it
	s.servicecenter.Registry(clusterName, data)
}

// queryEvent Handles "EventQuery" query events and respond if conditions are met
func (s *Server) queryEvent(query *serf.Query) {
	// todo: Get instances requested
}
