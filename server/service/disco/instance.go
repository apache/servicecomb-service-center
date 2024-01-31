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
	"sync"
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
	"github.com/go-chassis/cari/pkg/errsvc"
)

const (
	defaultMinInterval = 5 * time.Second
	defaultMinTimes    = 3
)

var (
	once          sync.Once
	propertiesMap map[string]string
)

func getInnerProperties() map[string]string {
	once.Do(func() {
		propertiesMap = config.GetStringMap("registry.instance.properties")
	})
	return propertiesMap
}

func RegisterInstance(ctx context.Context, in *pb.RegisterInstanceRequest) (*pb.RegisterInstanceResponse, error) {
	remoteIP := util.GetIPFromContext(ctx)

	if err := validator.ValidateRegisterInstanceRequest(in); err != nil {
		log.Error(fmt.Sprintf("register instance failed, invalid parameters, operator %s", remoteIP), err)
		return nil, pb.NewError(pb.ErrInvalidParams, err.Error())
	}

	if quotaErr := checkInstanceQuota(ctx); quotaErr != nil {
		log.Error(fmt.Sprintf("register instance failed, endpoints %v, host '%s', serviceID %s, operator %s",
			in.Instance.Endpoints, in.Instance.HostName, in.Instance.ServiceId, remoteIP), quotaErr)
		return nil, quotaErr
	}

	if popErr := populateInstanceDefaultValue(ctx, in.Instance); popErr != nil {
		return nil, popErr
	}

	if ok, err := datasource.EnableSync(ctx, in.Instance.ServiceId); err != nil {
		return nil, err
	} else if !ok {
		util.SetContext(ctx, util.CtxEnableSync, "0")
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
		log.Error("can not get service:"+instance.ServiceId, err)
		return pb.NewError(pb.ErrServiceNotExists, "Invalid 'serviceID' in request body.")
	}
	instance.Version = microservice.Version

	appendInnerPropertiesToInstance(instance)
	return nil
}

func appendInnerPropertiesToInstance(instance *pb.MicroServiceInstance) {
	if instance.Properties == nil {
		instance.Properties = make(map[string]string)
	}

	innerProps := getInnerProperties()
	if len(innerProps) <= 0 {
		return
	}

	for k, v := range innerProps {
		instance.Properties[k] = v
	}
}

func UnregisterInstance(ctx context.Context, in *pb.UnregisterInstanceRequest) error {
	remoteIP := util.GetIPFromContext(ctx)

	if err := validator.ValidateUnregisterInstanceRequest(in); err != nil {
		log.Error(fmt.Sprintf("unregister instance failed, invalid parameters, operator %s", remoteIP), err)
		return pb.NewError(pb.ErrInvalidParams, err.Error())
	}

	if ok, err := datasource.EnableSync(ctx, in.ServiceId); err != nil {
		return err
	} else if !ok {
		util.SetContext(ctx, util.CtxEnableSync, "0")
	}

	return datasource.GetMetadataManager().UnregisterInstance(ctx, in)
}

func SendHeartbeat(ctx context.Context, in *pb.HeartbeatRequest) error {
	remoteIP := util.GetIPFromContext(ctx)
	instanceID := in.InstanceId
	serviceID := in.ServiceId

	if err := validator.ValidateHeartbeatRequest(in); err != nil {
		log.Error(fmt.Sprintf("send heartbeat[%s/%s] failed, invalid parameters, operator %s",
			serviceID, instanceID, remoteIP), err)
		return pb.NewError(pb.ErrInvalidParams, err.Error())
	}

	if ok, err := datasource.EnableSync(ctx, in.ServiceId); err != nil {
		return err
	} else if !ok {
		util.SetContext(ctx, util.CtxEnableSync, "0")
	}

	err := datasource.GetMetadataManager().SendHeartbeat(ctx, in)
	if err != nil {
		log.Error(fmt.Sprintf("send heartbeat[%s/%s] failed, operator %s", serviceID, instanceID, remoteIP), err)
		return err
	}

	// append the inner properties
	err = appendInnerProperties(ctx, serviceID, instanceID)
	if err != nil {
		log.Error(fmt.Sprintf("append inner instance[%s/%s] properties failed, operator %s",
			serviceID, instanceID, remoteIP), err)
		return err
	}
	return nil
}

func appendInnerProperties(ctx context.Context, serviceID string, instanceID string) error {
	resp, err := datasource.GetMetadataManager().GetInstance(ctx, &pb.GetOneInstanceRequest{ProviderServiceId: serviceID, ProviderInstanceId: instanceID})
	if err != nil {
		log.Error(fmt.Sprintf("get instance[%s/%s] failed", serviceID, instanceID), err)
		return err
	}
	instance := resp.Instance
	if !shouldAppendInnerProperties(instance) {
		return nil
	}
	props := make(map[string]string, len(resp.Instance.Properties))
	for k, v := range resp.Instance.Properties {
		props[k] = v
	}
	return PutInstanceProperties(ctx, &pb.UpdateInstancePropsRequest{
		ServiceId:  serviceID,
		InstanceId: instanceID,
		Properties: props,
	})
}

func shouldAppendInnerProperties(instance *pb.MicroServiceInstance) bool {
	instProps := instance.Properties
	innerProps := getInnerProperties()
	if len(innerProps) == 0 {
		return false
	}
	for k, v := range innerProps {
		if prop, ok := instProps[k]; !ok || prop != v {
			return true
		}
	}
	return false
}

func SendManyHeartbeat(ctx context.Context, in *pb.HeartbeatSetRequest) (*pb.HeartbeatSetResponse, error) {
	remoteIP := util.GetIPFromContext(ctx)

	if len(in.Instances) == 0 {
		log.Error(fmt.Sprintf("invalid heartbeat set request. instances is empty, operator: %s", remoteIP), nil)
		return nil, pb.NewError(pb.ErrInvalidParams, "Request format invalid.")
	}
	return datasource.GetMetadataManager().SendManyHeartbeat(ctx, in)
}

func GetInstance(ctx context.Context, in *pb.GetOneInstanceRequest) (*pb.GetOneInstanceResponse, error) {
	remoteIP := util.GetIPFromContext(ctx)

	if err := validator.ValidateGetOneInstanceRequest(in); err != nil {
		log.Error(fmt.Sprintf("get instance failed: invalid parameters, operator: %s", remoteIP), err)
		return nil, pb.NewError(pb.ErrInvalidParams, err.Error())
	}

	return datasource.GetMetadataManager().GetInstance(ctx, in)
}

func ListInstance(ctx context.Context, in *pb.GetInstancesRequest) (*pb.GetInstancesResponse, error) {
	remoteIP := util.GetIPFromContext(ctx)

	if err := validator.ValidateGetInstancesRequest(in); err != nil {
		log.Error(fmt.Sprintf("get instances failed: invalid parameters, operator: %s", remoteIP), err)
		return nil, pb.NewError(pb.ErrInvalidParams, err.Error())
	}

	return datasource.GetMetadataManager().ListInstance(ctx, in)
}

func FindInstances(ctx context.Context, in *pb.FindInstancesRequest) (*pb.FindInstancesResponse, error) {
	remoteIP := util.GetIPFromContext(ctx)

	if err := validator.ValidateFindInstancesRequest(in); err != nil {
		log.Error(fmt.Sprintf("find instance failed: invalid parameters, operator: %s", remoteIP), err)
		return nil, pb.NewError(pb.ErrInvalidParams, err.Error())
	}

	if ok, err := datasource.EnableSync(ctx, in.ConsumerServiceId); err != nil {
		return nil, err
	} else if !ok {
		util.SetContext(ctx, util.CtxEnableSync, "0")
	}

	return datasource.GetMetadataManager().FindInstances(ctx, in)
}

func FindManyInstances(ctx context.Context, request *pb.BatchFindInstancesRequest) (*pb.BatchFindInstancesResponse, error) {
	remoteIP := util.GetIPFromContext(ctx)

	if len(request.Services) == 0 && len(request.Instances) == 0 {
		err := errors.New("Required services or instances")
		log.Error(fmt.Sprintf("batch find instance failed: invalid parameters, operator: %s", remoteIP), err)
		return nil, pb.NewError(pb.ErrInvalidParams, err.Error())
	}

	if err := validator.ValidateFindManyInstancesRequest(request); err != nil {
		log.Error(fmt.Sprintf("batch find instance failed: invalid parameters, operator: %s", remoteIP), err)
		return nil, pb.NewError(pb.ErrInvalidParams, err.Error())
	}

	response := &pb.BatchFindInstancesResponse{}

	var err error
	// find services
	response.Services, err = batchFindServices(ctx, request)
	if err != nil {
		log.Error(fmt.Sprintf("batch find instance failed, operator: %s", remoteIP), err)
		return nil, pb.NewError(pb.ErrInternal, err.Error())
	}

	// find instance
	response.Instances, err = batchFindInstances(ctx, request)
	if err != nil {
		log.Error(fmt.Sprintf("batch find instance failed, operator: %s", remoteIP), err)
		return nil, pb.NewError(pb.ErrInternal, err.Error())
	}

	return response, nil
}

func batchFindServices(ctx context.Context, request *pb.BatchFindInstancesRequest) (*pb.BatchFindResult, error) {
	if len(request.Services) == 0 {
		return nil, nil
	}
	cloneCtx := util.CloneContext(ctx)

	result := &pb.BatchFindResult{}
	failedResult := make(map[int32]*pb.FindFailedResult)
	for index, key := range request.Services {
		findCtx := util.WithRequestRev(cloneCtx, key.Rev)
		resp, err := FindInstances(findCtx, &pb.FindInstancesRequest{
			ConsumerServiceId: request.ConsumerServiceId,
			AppId:             key.Service.AppId,
			ServiceName:       key.Service.ServiceName,
			Environment:       key.Service.Environment,
		})
		errCode, errMsg, internalErr := parseError(err)
		if internalErr != nil {
			return nil, internalErr
		}

		var instances []*pb.MicroServiceInstance
		if resp != nil {
			instances = resp.Instances
		}

		failed, ok := failedResult[errCode]
		AppendFindResponse(findCtx, int64(index),
			pb.CreateResponse(errCode, errMsg),
			instances,
			&result.Updated, &result.NotModified, &failed)
		if !ok && failed != nil {
			failedResult[errCode] = failed
		}
	}
	for _, rs := range failedResult {
		result.Failed = append(result.Failed, rs)
	}
	return result, nil
}

func batchFindInstances(ctx context.Context, request *pb.BatchFindInstancesRequest) (*pb.BatchFindResult, error) {
	if len(request.Instances) == 0 {
		return nil, nil
	}
	cloneCtx := util.CloneContext(ctx)
	// can not find the shared provider instances
	cloneCtx = util.SetTargetDomainProject(cloneCtx, util.ParseDomain(ctx), util.ParseProject(ctx))

	result := &pb.BatchFindResult{}
	failedResult := make(map[int32]*pb.FindFailedResult)
	for index, key := range request.Instances {
		getCtx := util.WithRequestRev(cloneCtx, key.Rev)
		resp, err := GetInstance(getCtx, &pb.GetOneInstanceRequest{
			ConsumerServiceId:  request.ConsumerServiceId,
			ProviderServiceId:  key.Instance.ServiceId,
			ProviderInstanceId: key.Instance.InstanceId,
		})
		errCode, errMsg, internalErr := parseError(err)
		if internalErr != nil {
			return nil, internalErr
		}

		var instances []*pb.MicroServiceInstance
		if resp != nil {
			instances = append(instances, resp.Instance)
		}

		failed, ok := failedResult[errCode]
		AppendFindResponse(getCtx, int64(index),
			pb.CreateResponse(errCode, errMsg),
			instances,
			&result.Updated, &result.NotModified, &failed)
		if !ok && failed != nil {
			failedResult[errCode] = failed
		}
	}
	for _, rs := range failedResult {
		result.Failed = append(result.Failed, rs)
	}
	return result, nil
}

func parseError(err error) (int32, string, error) {
	testErr, ok := err.(*errsvc.Error)
	if !ok || testErr.InternalError() {
		return 0, "", err
	}
	return testErr.Code, testErr.Error(), nil
}

func AppendFindResponse(ctx context.Context, index int64, resp *pb.Response, instances []*pb.MicroServiceInstance,
	updatedResult *[]*pb.FindResult, notModifiedResult *[]int64, failedResult **pb.FindFailedResult) {
	if code := resp.GetCode(); code != pb.ResponseSuccess {
		if *failedResult == nil {
			*failedResult = &pb.FindFailedResult{
				Error: pb.NewError(code, resp.GetMessage()),
			}
		}
		(*failedResult).Indexes = append((*failedResult).Indexes, index)
		return
	}
	iv, _ := ctx.Value(util.CtxRequestRevision).(string)
	ov, _ := ctx.Value(util.CtxResponseRevision).(string)
	if len(iv) > 0 && iv == ov {
		*notModifiedResult = append(*notModifiedResult, index)
		return
	}
	*updatedResult = append(*updatedResult, &pb.FindResult{
		Index:     index,
		Instances: instances,
		Rev:       ov,
	})
}

func PutInstance(ctx context.Context, in *pb.RegisterInstanceRequest) error {
	remoteIP := util.GetIPFromContext(ctx)

	if err := validator.ValidateRegisterInstanceRequest(in); err != nil {
		log.Error(fmt.Sprintf("update instance failed, invalid parameters, operator %s", remoteIP), err)
		return pb.NewError(pb.ErrInvalidParams, err.Error())
	}

	return datasource.GetMetadataManager().PutInstance(ctx, in)
}

func PutInstanceStatus(ctx context.Context, in *pb.UpdateInstanceStatusRequest) error {
	remoteIP := util.GetIPFromContext(ctx)

	if err := validator.ValidateUpdateInstanceStatusRequest(in); err != nil {
		updateStatusFlag := util.StringJoin([]string{in.ServiceId, in.InstanceId, in.Status}, "/")
		log.Error(fmt.Sprintf("update instance[%s] status failed, operator: %s", updateStatusFlag, remoteIP), nil)
		return pb.NewError(pb.ErrInvalidParams, err.Error())
	}

	if ok, err := datasource.EnableSync(ctx, in.ServiceId); err != nil {
		return err
	} else if !ok {
		util.SetContext(ctx, util.CtxEnableSync, "0")
	}

	return datasource.GetMetadataManager().PutInstanceStatus(ctx, in)
}

func PutInstanceProperties(ctx context.Context, in *pb.UpdateInstancePropsRequest) error {
	remoteIP := util.GetIPFromContext(ctx)

	if err := validator.ValidateUpdateInstancePropsRequest(in); err != nil {
		instanceFlag := util.StringJoin([]string{in.ServiceId, in.InstanceId}, "/")
		log.Error(fmt.Sprintf("update instance[%s] properties failed, operator: %s", instanceFlag, remoteIP), nil)
		return pb.NewError(pb.ErrInvalidParams, err.Error())
	}

	properties := getInnerProperties()
	if in.Properties == nil {
		in.Properties = make(map[string]string, len(properties))
	}
	for k, v := range properties {
		in.Properties[k] = v
	}

	if ok, err := datasource.EnableSync(ctx, in.ServiceId); err != nil {
		return err
	} else if !ok {
		util.SetContext(ctx, util.CtxEnableSync, "0")
	}

	return datasource.GetMetadataManager().PutInstanceProperties(ctx, in)
}

func ClusterHealth(ctx context.Context) (*pb.GetInstancesResponse, error) {
	remoteIP := util.GetIPFromContext(ctx)

	if err := health.GlobalHealthChecker().Healthy(); err != nil {
		return nil, pb.NewError(pb.ErrUnhealthy, err.Error())
	}
	cloneContext := util.SetDomainProject(util.CloneContext(ctx), datasource.RegistryDomain, datasource.RegistryProject)
	serviceID, err := datasource.GetMetadataManager().ExistService(cloneContext, &pb.GetExistenceRequest{
		Type:        pb.ExistenceMicroservice,
		AppId:       apt.Service.AppId,
		Environment: apt.Service.Environment,
		ServiceName: apt.Service.ServiceName,
		Version:     apt.Service.Version,
	})
	if err != nil {
		log.Error(fmt.Sprintf("health check failed: get service center[%s/%s/%s/%s]'s serviceID failed, operator: %s",
			apt.Service.Environment, apt.Service.AppId, apt.Service.ServiceName, apt.Service.Version, remoteIP), err)
		return nil, err
	}

	instResp, err := datasource.GetMetadataManager().ListInstance(cloneContext, &pb.GetInstancesRequest{
		ProviderServiceId: serviceID,
	})
	if err != nil {
		log.Error(fmt.Sprintf("health check failed: get service center[%s][%s/%s/%s/%s]'s instances failed, operator: %s",
			serviceID, apt.Service.Environment, apt.Service.AppId, apt.Service.ServiceName, apt.Service.Version, remoteIP), err)
		return nil, err
	}
	return &pb.GetInstancesResponse{
		Instances: instResp.Instances,
	}, nil
}

func checkInstanceQuota(ctx context.Context) error {
	if !apt.IsSCInstance(ctx) {
		return quotasvc.ApplyInstance(ctx, 1)
	}
	return nil
}

func InstanceUsage(ctx context.Context, request *pb.GetServiceCountRequest) (int64, error) {
	resp, err := datasource.GetMetadataManager().CountInstance(ctx, request)
	if err != nil {
		return 0, err
	}
	return resp.Count, nil
}

func UpdateManyInstanceStatus(ctx context.Context, match *datasource.MatchPolicy, status string) error {
	err := datasource.GetMetadataManager().UpdateManyInstanceStatus(ctx, match, status)
	return err
}
