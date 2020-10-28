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

package service

import (
	"context"
	"errors"
	serviceUtil "github.com/apache/servicecomb-service-center/datasource/etcd/util"
	"github.com/apache/servicecomb-service-center/pkg/log"
	pb "github.com/apache/servicecomb-service-center/pkg/registry"
	"github.com/apache/servicecomb-service-center/pkg/util"
	"github.com/apache/servicecomb-service-center/server/core/proto"
	"github.com/apache/servicecomb-service-center/server/notify"
	"github.com/gorilla/websocket"
)

func (s *InstanceService) WatchPreOpera(ctx context.Context, in *pb.WatchInstanceRequest) error {
	if in == nil || len(in.SelfServiceId) == 0 {
		return errors.New("request format invalid")
	}
	domainProject := util.ParseDomainProject(ctx)
	if !serviceUtil.ServiceExist(ctx, domainProject, in.SelfServiceId) {
		return errors.New("service does not exist")
	}
	return nil
}

func (s *InstanceService) Watch(in *pb.WatchInstanceRequest, stream proto.ServiceInstanceCtrlWatchServer) error {
	log.Infof("new a stream list and watch with service[%s]", in.SelfServiceId)
	if err := s.WatchPreOpera(stream.Context(), in); err != nil {
		log.Errorf(err, "service[%s] establish watch failed: invalid params", in.SelfServiceId)
		return err
	}

	return notify.DoStreamListAndWatch(stream.Context(), in.SelfServiceId, nil, stream)
}

func (s *InstanceService) WebSocketWatch(ctx context.Context, in *pb.WatchInstanceRequest, conn *websocket.Conn) {
	log.Infof("new a web socket watch with service[%s]", in.SelfServiceId)
	if err := s.WatchPreOpera(ctx, in); err != nil {
		notify.EstablishWebSocketError(conn, err)
		return
	}
	notify.DoWebSocketListAndWatch(ctx, in.SelfServiceId, nil, conn)
}

func (s *InstanceService) WebSocketListAndWatch(ctx context.Context, in *pb.WatchInstanceRequest, conn *websocket.Conn) {
	log.Infof("new a web socket list and watch with service[%s]", in.SelfServiceId)
	if err := s.WatchPreOpera(ctx, in); err != nil {
		notify.EstablishWebSocketError(conn, err)
		return
	}
	notify.DoWebSocketListAndWatch(ctx, in.SelfServiceId, func() ([]*pb.WatchInstanceResponse, int64) {
		return serviceUtil.QueryAllProvidersInstances(ctx, in.SelfServiceId)
	}, conn)
}
