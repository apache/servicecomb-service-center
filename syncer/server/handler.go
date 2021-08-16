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
	"github.com/apache/servicecomb-service-center/syncer/client"
	pb "github.com/apache/servicecomb-service-center/syncer/proto"
	"github.com/go-chassis/foundation/tlsutil"
)

const (
	EventDiscovered = "discovered"
)

// tickHandler Timed task handler
func (s *Server) tickHandler() {
	log.Debugf("is leader: %v", s.etcd.IsLeader())
	if !s.etcd.IsLeader() {
		return
	}
	log.Debugf("Handle Tick")
	// Flush data to the storage of servicecenter
	s.servicecenter.FlushData()

	// sends a UserEvent on Serf, the event will be broadcast between members
	err := s.serf.UserEvent(EventDiscovered, util.StringToBytesWithNoCopy(s.conf.Cluster))
	if err != nil {
		log.Errorf(err, "Syncer send user event failed")
	}
}

// Pull returns sync data of servicecenter
func (s *Server) Pull(context.Context, *pb.PullRequest) (*pb.SyncData, error) {
	return s.servicecenter.Discovery(), nil
}

// userEvent Handles "EventUser" notification events, no response required
func (s *Server) userEvent(data ...[]byte) (success bool) {
	log.Debug("Receive serf user event")
	clusterName := util.BytesToStringWithNoCopy(data[0])

	// Excludes notifications from self, as the gossip protocol inevitably has redundant notifications
	if s.conf.Cluster == clusterName {
		return
	}

	tags := map[string]string{tagKeyClusterName: clusterName}
	// Get member information and get synchronized data from it
	members := s.serf.MembersByTags(tags)
	if len(members) == 0 {
		log.Warnf("serf member = %s is not found", clusterName)
		return
	}

	// todo: grpc supports multi-address polling
	// Get dta from remote member
	endpoint := fmt.Sprintf("%s:%s", members[0].Addr, members[0].Tags[tagKeyRPCPort])
	log.Debugf("Going to pull data from %s %s", members[0].Name, endpoint)

	enabled, err := strconv.ParseBool(members[0].Tags[tagKeyTLSEnabled])
	if err != nil {
		log.Warnf("get tls enabled failed, err = %s", err)
	}
	var tlsConfig *tls.Config
	if enabled {
		conf := s.conf.GetTLSConfig(s.conf.Listener.TLSMount.Name)
		sslOps := append(tlsutil.DefaultClientTLSOptions(), tlsConfigToOptions(conf)...)
		tlsConfig, err = tlsutil.GetClientTLSConfig(sslOps...)
		if err != nil {
			log.Error("get grpc client tls config failed", err)
			return
		}
	}

	cli := client.NewSyncClient(endpoint, tlsConfig)
	syncData, err := cli.Pull(context.Background())
	if err != nil {
		log.Errorf(err, "Pull other serf instances failed, node name is '%s'", members[0].Name)
		return
	}
	// Registry instances to servicecenter and update storage of it
	s.servicecenter.Registry(clusterName, syncData)
	return true
}
