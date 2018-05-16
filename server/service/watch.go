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
	"errors"
	"github.com/apache/incubator-servicecomb-service-center/pkg/util"
	apt "github.com/apache/incubator-servicecomb-service-center/server/core"
	pb "github.com/apache/incubator-servicecomb-service-center/server/core/proto"
	nf "github.com/apache/incubator-servicecomb-service-center/server/service/notification"
	serviceUtil "github.com/apache/incubator-servicecomb-service-center/server/service/util"
	"github.com/gorilla/websocket"
	"golang.org/x/net/context"
)

func (s *InstanceService) WatchPreOpera(ctx context.Context, in *pb.WatchInstanceRequest) error {
	if in == nil || len(in.SelfServiceId) == 0 {
		return errors.New("Request format invalid.")
	}
	domainProject := util.ParseDomainProject(ctx)
	if !serviceUtil.ServiceExist(ctx, domainProject, in.SelfServiceId) {
		return errors.New("Service does not exist.")
	}
	return nil
}

func (s *InstanceService) Watch(in *pb.WatchInstanceRequest, stream pb.ServiceInstanceCtrl_WatchServer) error {
	var err error
	if err = s.WatchPreOpera(stream.Context(), in); err != nil {
		util.Logger().Errorf(err, "establish watch failed: invalid params.")
		return err
	}
	domainProject := util.ParseDomainProject(stream.Context())
	watcher := nf.NewInstanceListWatcher(in.SelfServiceId, apt.GetInstanceRootKey(domainProject)+"/", nil)
	err = nf.GetNotifyService().AddSubscriber(watcher)
	util.Logger().Infof("start watch instance status, watcher %s %s", watcher.Subject(), watcher.Id())
	return nf.HandleWatchJob(watcher, stream, nf.GetNotifyService().Config.NotifyTimeout)
}

func (s *InstanceService) WebSocketWatch(ctx context.Context, in *pb.WatchInstanceRequest, conn *websocket.Conn) {
	util.Logger().Infof("New a web socket watch with %s", in.SelfServiceId)
	if err := s.WatchPreOpera(ctx, in); err != nil {
		nf.EstablishWebSocketError(conn, err)
		return
	}
	nf.DoWebSocketListAndWatch(ctx, in.SelfServiceId, nil, conn)
}

func (s *InstanceService) WebSocketListAndWatch(ctx context.Context, in *pb.WatchInstanceRequest, conn *websocket.Conn) {
	util.Logger().Infof("New a web socket list and watch with %s", in.SelfServiceId)
	if err := s.WatchPreOpera(ctx, in); err != nil {
		nf.EstablishWebSocketError(conn, err)
		return
	}
	nf.DoWebSocketListAndWatch(ctx, in.SelfServiceId, func() ([]*pb.WatchInstanceResponse, int64) {
		return serviceUtil.QueryAllProvidersInstances(ctx, in.SelfServiceId)
	}, conn)
}
