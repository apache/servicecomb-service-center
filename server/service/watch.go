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
	"fmt"

	"github.com/apache/servicecomb-service-center/datasource"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/proto"
	"github.com/apache/servicecomb-service-center/server/connection/grpc"
	"github.com/apache/servicecomb-service-center/server/connection/ws"
	pb "github.com/go-chassis/cari/discovery"
	"github.com/gorilla/websocket"
)

func (s *InstanceService) WatchPreOpera(ctx context.Context, in *pb.WatchInstanceRequest) error {
	if in == nil || len(in.SelfServiceId) == 0 {
		return errors.New("request format invalid")
	}
	resp, err := datasource.Instance().ExistServiceByID(ctx, &pb.GetExistenceByIDRequest{
		ServiceId: in.SelfServiceId,
	})
	if err != nil {
		return err
	}
	if !resp.Exist {
		return datasource.ErrServiceNotExists
	}
	return nil
}

func (s *InstanceService) Watch(in *pb.WatchInstanceRequest, stream proto.ServiceInstanceCtrlWatchServer) error {
	log.Infof("new a stream list and watch with service[%s]", in.SelfServiceId)
	if err := s.WatchPreOpera(stream.Context(), in); err != nil {
		log.Errorf(err, "service[%s] establish watch failed: invalid params", in.SelfServiceId)
		return err
	}

	return grpc.ListAndWatch(stream.Context(), in.SelfServiceId, nil, stream)
}

func (s *InstanceService) WebSocketWatch(ctx context.Context, in *pb.WatchInstanceRequest, conn *websocket.Conn) {
	log.Infof("new a web socket watch with service[%s]", in.SelfServiceId)
	if err := s.WatchPreOpera(ctx, in); err != nil {
		ws.SendEstablishError(conn, err)
		return
	}
	ws.ListAndWatch(ctx, in.SelfServiceId, nil, conn)
}

func (s *InstanceService) WebSocketListAndWatch(ctx context.Context, in *pb.WatchInstanceRequest, conn *websocket.Conn) {
	log.Infof("new a web socket list and watch with service[%s]", in.SelfServiceId)
	if err := s.WatchPreOpera(ctx, in); err != nil {
		ws.SendEstablishError(conn, err)
		return
	}
	ws.ListAndWatch(ctx, in.SelfServiceId, func() ([]*pb.WatchInstanceResponse, int64) {
		return s.QueryAllProvidersInstances(ctx, in)
	}, conn)
}

func (s *InstanceService) QueryAllProvidersInstances(ctx context.Context, in *pb.WatchInstanceRequest) ([]*pb.WatchInstanceResponse, int64) {
	depResp, err := datasource.Instance().SearchConsumerDependency(ctx, &pb.GetDependenciesRequest{
		ServiceId: in.SelfServiceId,
	})
	if err != nil {
		log.Error(fmt.Sprintf("search service[%s] dependencies failed", in.SelfServiceId), err)
		return nil, 0
	}
	if depResp.Response.GetCode() != pb.ResponseSuccess {
		log.Error(fmt.Sprintf("search service[%s] dependencies failed. %s",
			in.SelfServiceId, depResp.Response.GetMessage()), nil)
		return nil, 0
	}
	var results []*pb.WatchInstanceResponse
	for _, provider := range depResp.Providers {
		instResp, err := datasource.Instance().GetInstances(ctx, &pb.GetInstancesRequest{
			ProviderServiceId: provider.ServiceId,
		})
		if err != nil {
			log.Error(fmt.Sprintf("get service[%s] instances failed", in.SelfServiceId), err)
			return nil, 0
		}
		if instResp.Response.GetCode() != pb.ResponseSuccess {
			log.Error(fmt.Sprintf("get service[%s] instances failed. %s",
				in.SelfServiceId, instResp.Response.GetMessage()), nil)
			return nil, 0
		}
		for _, instance := range instResp.Instances {
			results = append(results, &pb.WatchInstanceResponse{
				Response: pb.CreateResponse(pb.ResponseSuccess, "List instance successfully."),
				Action:   string(pb.EVT_INIT),
				Key: &pb.MicroServiceKey{
					Environment: provider.Environment,
					AppId:       provider.AppId,
					ServiceName: provider.ServiceName,
					Version:     provider.Version,
				},
				Instance: instance,
			})
		}
	}
	return results, 0
}
