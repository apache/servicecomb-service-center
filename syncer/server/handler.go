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

	"github.com/apache/servicecomb-service-center/pkg/dump"

	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/tlsutil"
	"github.com/apache/servicecomb-service-center/pkg/util"
	"github.com/apache/servicecomb-service-center/server/alarm"
	"github.com/apache/servicecomb-service-center/syncer/client"
	pb "github.com/apache/servicecomb-service-center/syncer/proto"
)

const (
	EventDiscovered       = "discovered"
	EventIncrementPulled  = "incrementPulled"
	EventNotifyFullPulled = "notifyFullPulled"
	BufferSize            = 1000
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
	s.mux.Lock()
	defer s.mux.Unlock()
	if s.triggered {
		err := s.serf.UserEvent(EventNotifyFullPulled, util.StringToBytesWithNoCopy(""))
		s.triggered = false
		if err != nil {
			log.Error("Syncer send notifyFullPulled user event failed", err)
		}
		err = alarm.Clear(alarm.IDIncrementPullError)
		if err != nil {
			log.Error("", err)
		}
	} else {
		err := s.serf.UserEvent(EventIncrementPulled, util.StringToBytesWithNoCopy(s.conf.Cluster))
		if err != nil {
			log.Error("Syncer send incrementPulled user event failed", err)
		}
	}
}

// Pull returns sync data of servicecenter
func (s *Server) Pull(ctx context.Context, req *pb.PullRequest) (*pb.SyncData, error) {
	if _, ok := s.channelMap[req.GetAddr()]; !ok {
		s.channelMap[req.GetAddr()] = make(chan *dump.WatchInstanceChangedEvent, BufferSize)
	}
	return s.servicecenter.Discovery(), nil
}

func (s *Server) IncrementPull(ctx context.Context, req *pb.IncrementPullRequest) (*pb.SyncData, error) {
	incrementQueue := s.GetIncrementQueue(req.GetAddr())
	return s.GetIncrementData(ctx, incrementQueue), nil
}

func (s *Server) DeclareDataLength(ctx context.Context, req *pb.DeclareRequest) (*pb.DeclareResponse, error) {
	return s.getSyncDataLength(req.GetAddr()), nil
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
	syncData, err := cli.Pull(context.Background(), s.conf.Listener.RPCAddr)
	if err != nil {
		log.Errorf(err, "Pull other serf instances failed, node name is '%s'", members[0].Name)
		return
	}
	// Registry instances to servicecenter and update storage of it
	s.servicecenter.Registry(clusterName, syncData)
	return true
}

func (s *Server) incrementUserEvent(data ...[]byte) (success bool) {
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
		log.Warn(fmt.Sprintf("serf member = %s is not found", clusterName))
		return
	}

	// Get dta from remote member
	endpoint := fmt.Sprintf("%s:%s", members[0].Addr, members[0].Tags[tagKeyRPCPort])
	log.Debug(fmt.Sprintf("Going to pull data from %s %s", members[0].Name, endpoint))

	enabled, err := strconv.ParseBool(members[0].Tags[tagKeyTLSEnabled])
	if err != nil {
		log.Warn(fmt.Sprintf("get tls enabled failed, err = %s", err))
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
	declareResponse, err := cli.DeclareDataLength(context.Background(), s.conf.Listener.RPCAddr)
	if err != nil {
		log.Error(fmt.Sprintf("Get syncData length from other node failed, node name is '%s'", members[0].Name), err)
		return
	}
	syncDataLength := declareResponse.SyncDataLength

	if syncDataLength != 0 {
		syncData, err := cli.IncrementPull(context.Background(), s.conf.Listener.RPCAddr)
		if err != nil {
			log.Error(fmt.Sprintf("IncrementPull other serf instances failed, node name is '%s'", members[0].Name), err)
			return
		}

		if syncDataLength != int64(len(syncData.Instances)) {
			err = alarm.Raise(alarm.IDIncrementPullError, alarm.AdditionalContext("%v", err))
			if err != nil {
				log.Error("", err)
			}
		}
		s.servicecenter.IncrementRegistry(clusterName, syncData)
	}
	return true
}

func (s *Server) notifyUserEvent(data ...[]byte) (success bool) {
	err := s.serf.UserEvent(EventDiscovered, util.StringToBytesWithNoCopy(s.conf.Cluster))
	if err != nil {
		log.Error("Syncer send discovered user event failed", err)
	}
	return true
}
