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
package syncer

import (
	"context"
	"fmt"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"time"

	"github.com/apache/servicecomb-service-center/syncer/notify"
	"github.com/apache/servicecomb-service-center/syncer/pkg/events"
	pb "github.com/apache/servicecomb-service-center/syncer/proto"
	"github.com/gogo/protobuf/proto"
	"github.com/hashicorp/serf/serf"
)

func (s *Server) OnEvent(event events.ContextEvent) {
	data, _ := proto.Marshal(&pb.Member{
		NodeName: s.conf.NodeName,
		RPCPort:  int32(s.conf.RPCPort),
		Time:     fmt.Sprintf("%d", time.Now().UTC().Second()),
	})

	err := s.agent.UserEvent(event.Type(), data, true)
	if err != nil {
		log.Errorf(err, "Syncer send user event failed")
	}
}

func (s *Server) HandleEvent(event serf.Event) {
	switch event.EventType() {
	case serf.EventUser:
		s.userEvent(event.(serf.UserEvent))
	case serf.EventQuery:
		s.queryEvent(event.(*serf.Query))
	}
}

func (s *Server) userEvent(event serf.UserEvent) {
	m := &pb.Member{}
	err := proto.Unmarshal(event.Payload, m)
	if err != nil {
		log.Errorf(err, "trigger user event '%s' handler failed", event.EventType())
		return
	}

	if s.agent.LocalMember().Name == m.NodeName {
		return
	}

	member := s.agent.Member(m.NodeName)
	data, err := s.broker.Pull(context.Background(), fmt.Sprintf("%s:%d", member.Addr, m.RPCPort))
	if err != nil {
		log.Errorf(err, "pull other peer instances failed, node name is '%s'", m.NodeName)
		return
	}
	ctx := context.WithValue(context.Background(), notify.EventPullByPeer, &pb.NodeDataInfo{NodeName: m.NodeName, DataInfo: data})
	events.Dispatch(events.NewContextEvent(notify.EventPullByPeer, ctx))
}

func (s *Server) queryEvent(query *serf.Query) {
	// todo: 按需获取实例信息
}
