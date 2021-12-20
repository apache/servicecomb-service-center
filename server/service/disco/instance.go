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

package disco

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/apache/servicecomb-service-center/datasource"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/util"
	"github.com/apache/servicecomb-service-center/server/config"
	apt "github.com/apache/servicecomb-service-center/server/core"
	"github.com/apache/servicecomb-service-center/server/health"
	quotasvc "github.com/apache/servicecomb-service-center/server/service/quota"
	"github.com/apache/servicecomb-service-center/server/service/validator"
	pb "github.com/go-chassis/cari/discovery"
)

const (
	defaultMinInterval = 5 * time.Second
	defaultMinTimes    = 3
)

func RegisterInstance(ctx context.Context, in *pb.RegisterInstanceRequest) (*pb.RegisterInstanceResponse, error) {
	if err := validator.Validate(in); err != nil {
		remoteIP := util.GetIPFromContext(ctx)
		log.Error(fmt.Sprintf("register instance failed, invalid parameters, operator %s", remoteIP), err)
		return &pb.RegisterInstanceResponse{
			Response: pb.CreateResponse(pb.ErrInvalidParams, err.Error()),
		}, nil
	}
	remoteIP := util.GetIPFromContext(ctx)
	if quotaErr := checkInstanceQuota(ctx); quotaErr != nil {
		log.Error(fmt.Sprintf("register instance failed, endpoints %v, host '%s', serviceID %s, operator %s",
			in.Instance.Endpoints, in.Instance.HostName, in.Instance.ServiceId, remoteIP), quotaErr)
		response, err := datasource.WrapErrResponse(quotaErr)
		return &pb.RegisterInstanceResponse{
			Response: response,
		}, err
	}
	if popErr := populateInstanceDefaultValue(ctx, in.Instance); popErr != nil {
		response, err := datasource.WrapErrResponse(popErr)
		return &pb.RegisterInstanceResponse{
			Response: response,
		}, err
	}
	return datasource.GetMetadataManager().RegisterInstance(ctx, in)
}

// instance util
func populateInstanceDefaultValue(ctx context.Context, instance *pb.MicroServiceInstance) error {
	if len(instance.Status) == 0 {
		instance.Status = pb.MSI_UP
	}

	instance.Timestamp = strconv.FormatInt(time.Now().Unix(), 10)
	instance.ModTimestamp = instance.Timestamp

	// 这里应该根据租约计时
	// Health check对象仅用于呈现服务健康检查逻辑，如果CHECK_BY_PLATFORM类型，表明由sidecar代发心跳，实例120s超时
	if instance.HealthCheck == nil {
		instance.HealthCheck = &pb.HealthCheck{
			Mode:     pb.CHECK_BY_HEARTBEAT,
			Interval: datasource.DefaultLeaseRenewalInterval,
			Times:    datasource.DefaultLeaseRetryTimes,
		}
	} else if instance.HealthCheck.Mode == pb.CHECK_BY_HEARTBEAT {
		renewalInterval := int32(config.GetDuration("registry.instance.minInterval", defaultMinInterval) / time.Second)
		if instance.HealthCheck.Interval < renewalInterval {
			instance.HealthCheck.Interval = renewalInterval
		}
		retryTimes := int32(config.GetInt("registry.instance.minTimes", defaultMinTimes))
		if instance.HealthCheck.Times < retryTimes {
			instance.HealthCheck.Times = retryTimes
		}
	} else if instance.HealthCheck.Mode == pb.CHECK_BY_PLATFORM {
		instance.HealthCheck.Interval = datasource.DefaultLeaseRenewalInterval
		instance.HealthCheck.Times = datasource.DefaultLeaseRetryTimes
	}

	microservice, err := GetService(ctx, &pb.GetServiceRequest{ServiceId: instance.ServiceId})
	if err != nil {
		return pb.NewError(pb.ErrServiceNotExists, "Invalid 'serviceID' in request body.")
	}
	instance.Version = microservice.Version
	return nil
}

func UnregisterInstance(ctx context.Context,
	in *pb.UnregisterInstanceRequest) (*pb.UnregisterInstanceResponse, error) {
	if err := validator.Validate(in); err != nil {
		remoteIP := util.GetIPFromContext(ctx)
		log.Error(fmt.Sprintf("unregister instance failed, invalid parameters, operator %s", remoteIP), err)
		return &pb.UnregisterInstanceResponse{
			Response: pb.CreateResponse(pb.ErrInvalidParams, err.Error()),
		}, nil
	}

	return datasource.GetMetadataManager().UnregisterInstance(ctx, in)
}

func Heartbeat(ctx context.Context, in *pb.HeartbeatRequest) (*pb.HeartbeatResponse, error) {
	if err := validator.Validate(in); err != nil {
		remoteIP := util.GetIPFromContext(ctx)
		log.Error(fmt.Sprintf("heartbeat failed, invalid parameters, operator %s", remoteIP), err)
		return &pb.HeartbeatResponse{
			Response: pb.CreateResponse(pb.ErrInvalidParams, err.Error()),
		}, nil
	}

	return datasource.GetMetadataManager().Heartbeat(ctx, in)
}

func HeartbeatSet(ctx context.Context,
	in *pb.HeartbeatSetRequest) (*pb.HeartbeatSetResponse, error) {
	if len(in.Instances) == 0 {
		log.Error("heartbeats failed, invalid request. Body not contain Instances or is empty", nil)
		return &pb.HeartbeatSetResponse{
			Response: pb.CreateResponse(pb.ErrInvalidParams, "Request format invalid."),
		}, nil
	}
	return datasource.GetMetadataManager().HeartbeatSet(ctx, in)
}

func GetOneInstance(ctx context.Context,
	in *pb.GetOneInstanceRequest) (*pb.GetOneInstanceResponse, error) {
	err := validator.Validate(in)
	if err != nil {
		log.Error("get instance failed: invalid parameters", err)
		return &pb.GetOneInstanceResponse{
			Response: pb.CreateResponse(pb.ErrInvalidParams, err.Error()),
		}, nil
	}

	return datasource.GetMetadataManager().GetInstance(ctx, in)
}

func GetInstances(ctx context.Context, in *pb.GetInstancesRequest) (*pb.GetInstancesResponse, error) {
	err := validator.Validate(in)
	if err != nil {
		log.Error("get instances failed: invalid parameters", err)
		return &pb.GetInstancesResponse{
			Response: pb.CreateResponse(pb.ErrInvalidParams, err.Error()),
		}, nil
	}

	return datasource.GetMetadataManager().GetInstances(ctx, in)
}

func FindInstances(ctx context.Context, in *pb.FindInstancesRequest) (*pb.FindInstancesResponse, error) {
	err := validator.Validate(in)
	if err != nil {
		log.Error("find instance failed: invalid parameters", err)
		return &pb.FindInstancesResponse{
			Response: pb.CreateResponse(pb.ErrInvalidParams, err.Error()),
		}, nil
	}

	return datasource.GetMetadataManager().FindInstances(ctx, in)
}

func BatchFindInstances(ctx context.Context, in *pb.BatchFindInstancesRequest) (*pb.BatchFindInstancesResponse, error) {
	if len(in.Services) == 0 && len(in.Instances) == 0 {
		err := errors.New("Required services or instances")
		log.Error("batch find instance failed: invalid parameters", err)
		return &pb.BatchFindInstancesResponse{
			Response: pb.CreateResponse(pb.ErrInvalidParams, err.Error()),
		}, nil
	}

	err := validator.Validate(in)
	if err != nil {
		log.Error("batch find instance failed: invalid parameters", err)
		return &pb.BatchFindInstancesResponse{
			Response: pb.CreateResponse(pb.ErrInvalidParams, err.Error()),
		}, nil
	}

	return datasource.GetMetadataManager().BatchFind(ctx, in)
}

func UpdateInstanceStatus(ctx context.Context, in *pb.UpdateInstanceStatusRequest) (*pb.UpdateInstanceStatusResponse, error) {
	if err := validator.Validate(in); err != nil {
		updateStatusFlag := util.StringJoin([]string{in.ServiceId, in.InstanceId, in.Status}, "/")
		log.Error(fmt.Sprintf("update instance[%s] status failed", updateStatusFlag), nil)
		return &pb.UpdateInstanceStatusResponse{
			Response: pb.CreateResponse(pb.ErrInvalidParams, err.Error()),
		}, nil
	}

	return datasource.GetMetadataManager().UpdateInstanceStatus(ctx, in)
}

func UpdateInstanceProperties(ctx context.Context, in *pb.UpdateInstancePropsRequest) (*pb.UpdateInstancePropsResponse, error) {
	if err := validator.Validate(in); err != nil {
		instanceFlag := util.StringJoin([]string{in.ServiceId, in.InstanceId}, "/")
		log.Error(fmt.Sprintf("update instance[%s] properties failed", instanceFlag), nil)
		return &pb.UpdateInstancePropsResponse{
			Response: pb.CreateResponse(pb.ErrInvalidParams, err.Error()),
		}, nil
	}

	return datasource.GetMetadataManager().UpdateInstanceProperties(ctx, in)
}

func ClusterHealth(ctx context.Context) (*pb.GetInstancesResponse, error) {
	if err := health.GlobalHealthChecker().Healthy(); err != nil {
		return &pb.GetInstancesResponse{
			Response: pb.CreateResponse(pb.ErrUnhealthy, err.Error()),
		}, nil
	}
	cloneContext := util.SetDomainProject(util.CloneContext(ctx), datasource.RegistryDomain, datasource.RegistryProject)
	svcResp, err := datasource.GetMetadataManager().ExistService(cloneContext, &pb.GetExistenceRequest{
		Type:        pb.ExistenceMicroservice,
		AppId:       apt.Service.AppId,
		Environment: apt.Service.Environment,
		ServiceName: apt.Service.ServiceName,
		Version:     apt.Service.Version,
	})

	if err != nil {
		log.Error(fmt.Sprintf("health check failed: get service center[%s/%s/%s/%s]'s serviceID failed",
			apt.Service.Environment, apt.Service.AppId, apt.Service.ServiceName, apt.Service.Version), err)
		return &pb.GetInstancesResponse{
			Response: pb.CreateResponse(pb.ErrInternal, err.Error()),
		}, err
	}
	if len(svcResp.ServiceId) == 0 {
		log.Error(fmt.Sprintf("health check failed: service center[%s/%s/%s/%s]'s serviceID does not exist",
			apt.Service.Environment, apt.Service.AppId, apt.Service.ServiceName, apt.Service.Version), nil)
		return &pb.GetInstancesResponse{
			Response: pb.CreateResponse(pb.ErrServiceNotExists, "ServiceCenter's serviceID not exist."),
		}, nil
	}

	instResp, err := datasource.GetMetadataManager().GetInstances(cloneContext, &pb.GetInstancesRequest{
		ProviderServiceId: svcResp.ServiceId,
	})
	if err != nil {
		log.Error(fmt.Sprintf("health check failed: get service center[%s][%s/%s/%s/%s]'s instances failed",
			svcResp.ServiceId, apt.Service.Environment, apt.Service.AppId, apt.Service.ServiceName, apt.Service.Version), err)
		return &pb.GetInstancesResponse{
			Response: pb.CreateResponse(pb.ErrInternal, err.Error()),
		}, err
	}
	return &pb.GetInstancesResponse{
		Response:  pb.CreateResponse(pb.ResponseSuccess, "Health check successfully."),
		Instances: instResp.Instances,
	}, nil
}

func checkInstanceQuota(ctx context.Context) error {
	if !apt.IsSCInstance(ctx) {
		return quotasvc.ApplyInstance(ctx, 1)
	}
	return nil
}
