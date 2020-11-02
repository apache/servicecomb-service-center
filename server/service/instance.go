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
	"github.com/apache/servicecomb-service-center/datasource"
	serviceUtil "github.com/apache/servicecomb-service-center/datasource/etcd/util"
	"github.com/apache/servicecomb-service-center/pkg/log"
	pb "github.com/apache/servicecomb-service-center/pkg/registry"
	"github.com/apache/servicecomb-service-center/pkg/util"
	apt "github.com/apache/servicecomb-service-center/server/core"
	"github.com/apache/servicecomb-service-center/server/health"
	scerr "github.com/apache/servicecomb-service-center/server/scerror"
)

type InstanceService struct {
}

func (s *InstanceService) Register(ctx context.Context, in *pb.RegisterInstanceRequest) (*pb.RegisterInstanceResponse, error) {
	if err := Validate(in); err != nil {
		remoteIP := util.GetIPFromContext(ctx)
		log.Errorf(err, "register instance failed, invalid parameters, operator %s", remoteIP)
		return &pb.RegisterInstanceResponse{
			Response: pb.CreateResponse(scerr.ErrInvalidParams, err.Error()),
		}, nil
	}

	return datasource.Instance().RegisterInstance(ctx, in)
}

func (s *InstanceService) Unregister(ctx context.Context, in *pb.UnregisterInstanceRequest) (*pb.UnregisterInstanceResponse, error) {
	if err := Validate(in); err != nil {
		remoteIP := util.GetIPFromContext(ctx)
		log.Errorf(err, "unregister instance failed, invalid parameters, operator %s", remoteIP)
		return &pb.UnregisterInstanceResponse{
			Response: pb.CreateResponse(scerr.ErrInvalidParams, err.Error()),
		}, nil
	}

	return datasource.Instance().UnregisterInstance(ctx, in)
}

func (s *InstanceService) Heartbeat(ctx context.Context, in *pb.HeartbeatRequest) (*pb.HeartbeatResponse, error) {
	if err := Validate(in); err != nil {
		remoteIP := util.GetIPFromContext(ctx)
		log.Errorf(err, "heartbeat failed, invalid parameters, operator %s", remoteIP)
		return &pb.HeartbeatResponse{
			Response: pb.CreateResponse(scerr.ErrInvalidParams, err.Error()),
		}, nil
	}

	return datasource.Instance().Heartbeat(ctx, in)
}

func (s *InstanceService) HeartbeatSet(ctx context.Context, in *pb.HeartbeatSetRequest) (*pb.HeartbeatSetResponse, error) {
	if len(in.Instances) == 0 {
		log.Errorf(nil, "heartbeats failed, invalid request. Body not contain Instances or is empty")
		return &pb.HeartbeatSetResponse{
			Response: pb.CreateResponse(scerr.ErrInvalidParams, "Request format invalid."),
		}, nil
	}
	return datasource.Instance().HeartbeatSet(ctx, in)
}

func (s *InstanceService) GetOneInstance(ctx context.Context, in *pb.GetOneInstanceRequest) (*pb.GetOneInstanceResponse, error) {
	err := Validate(in)
	if err != nil {
		log.Errorf(err, "get instance failed: invalid parameters")
		return &pb.GetOneInstanceResponse{
			Response: pb.CreateResponse(scerr.ErrInvalidParams, err.Error()),
		}, nil
	}

	return datasource.Instance().GetInstance(ctx, in)
}

func (s *InstanceService) GetInstances(ctx context.Context, in *pb.GetInstancesRequest) (*pb.GetInstancesResponse, error) {
	err := Validate(in)
	if err != nil {
		log.Errorf(err, "get instances failed: invalid parameters")
		return &pb.GetInstancesResponse{
			Response: pb.CreateResponse(scerr.ErrInvalidParams, err.Error()),
		}, nil
	}

	return datasource.Instance().GetInstances(ctx, in)
}

func (s *InstanceService) Find(ctx context.Context, in *pb.FindInstancesRequest) (*pb.FindInstancesResponse, error) {
	err := Validate(in)
	if err != nil {
		log.Errorf(err, "find instance failed: invalid parameters")
		return &pb.FindInstancesResponse{
			Response: pb.CreateResponse(scerr.ErrInvalidParams, err.Error()),
		}, nil
	}

	return datasource.Instance().FindInstances(ctx, in)
}

func (s *InstanceService) BatchFind(ctx context.Context, in *pb.BatchFindInstancesRequest) (*pb.BatchFindInstancesResponse, error) {
	if len(in.Services) == 0 && len(in.Instances) == 0 {
		err := errors.New("Required services or instances")
		log.Errorf(err, "batch find instance failed: invalid parameters")
		return &pb.BatchFindInstancesResponse{
			Response: pb.CreateResponse(scerr.ErrInvalidParams, err.Error()),
		}, nil
	}

	err := Validate(in)
	if err != nil {
		log.Errorf(err, "batch find instance failed: invalid parameters")
		return &pb.BatchFindInstancesResponse{
			Response: pb.CreateResponse(scerr.ErrInvalidParams, err.Error()),
		}, nil
	}

	return datasource.Instance().BatchFind(ctx, in)
}

func (s *InstanceService) UpdateStatus(ctx context.Context, in *pb.UpdateInstanceStatusRequest) (*pb.UpdateInstanceStatusResponse, error) {
	if err := Validate(in); err != nil {
		updateStatusFlag := util.StringJoin([]string{in.ServiceId, in.InstanceId, in.Status}, "/")
		log.Errorf(nil, "update instance[%s] status failed", updateStatusFlag)
		return &pb.UpdateInstanceStatusResponse{
			Response: pb.CreateResponse(scerr.ErrInvalidParams, err.Error()),
		}, nil
	}

	return datasource.Instance().UpdateInstanceStatus(ctx, in)
}

func (s *InstanceService) UpdateInstanceProperties(ctx context.Context, in *pb.UpdateInstancePropsRequest) (*pb.UpdateInstancePropsResponse, error) {
	if err := Validate(in); err != nil {
		instanceFlag := util.StringJoin([]string{in.ServiceId, in.InstanceId}, "/")
		log.Errorf(nil, "update instance[%s] properties failed", instanceFlag)
		return &pb.UpdateInstancePropsResponse{
			Response: pb.CreateResponse(scerr.ErrInvalidParams, err.Error()),
		}, nil
	}

	return datasource.Instance().UpdateInstanceProperties(ctx, in)
}

func (s *InstanceService) ClusterHealth(ctx context.Context) (*pb.GetInstancesResponse, error) {
	if err := health.GlobalHealthChecker().Healthy(); err != nil {
		return &pb.GetInstancesResponse{
			Response: pb.CreateResponse(scerr.ErrUnhealthy, err.Error()),
		}, nil
	}

	domainProject := apt.RegistryDomainProject
	serviceID, err := serviceUtil.GetServiceID(ctx, &pb.MicroServiceKey{
		AppId:       apt.Service.AppId,
		Environment: apt.Service.Environment,
		ServiceName: apt.Service.ServiceName,
		Version:     apt.Service.Version,
		Tenant:      domainProject,
	})

	if err != nil {
		log.Errorf(err, "health check failed: get service center[%s/%s/%s/%s]'s serviceID failed",
			apt.Service.Environment, apt.Service.AppId, apt.Service.ServiceName, apt.Service.Version)
		return &pb.GetInstancesResponse{
			Response: pb.CreateResponse(scerr.ErrInternal, err.Error()),
		}, err
	}
	if len(serviceID) == 0 {
		log.Errorf(nil, "health check failed: service center[%s/%s/%s/%s]'s serviceID does not exist",
			apt.Service.Environment, apt.Service.AppId, apt.Service.ServiceName, apt.Service.Version)
		return &pb.GetInstancesResponse{
			Response: pb.CreateResponse(scerr.ErrServiceNotExists, "ServiceCenter's serviceID not exist."),
		}, nil
	}

	instances, err := serviceUtil.GetAllInstancesOfOneService(ctx, domainProject, serviceID)
	if err != nil {
		log.Errorf(err, "health check failed: get service center[%s][%s/%s/%s/%s]'s instances failed",
			serviceID, apt.Service.Environment, apt.Service.AppId, apt.Service.ServiceName, apt.Service.Version)
		return &pb.GetInstancesResponse{
			Response: pb.CreateResponse(scerr.ErrInternal, err.Error()),
		}, err
	}
	return &pb.GetInstancesResponse{
		Response:  pb.CreateResponse(pb.ResponseSuccess, "Health check successfully."),
		Instances: instances,
	}, nil
}
