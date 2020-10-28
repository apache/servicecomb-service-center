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
	"encoding/json"
	"errors"
	"fmt"
	"github.com/apache/servicecomb-service-center/datasource/etcd/client"
	serviceUtil "github.com/apache/servicecomb-service-center/datasource/etcd/util"
	errorsEx "github.com/apache/servicecomb-service-center/pkg/errors"
	"github.com/apache/servicecomb-service-center/pkg/gopool"
	"github.com/apache/servicecomb-service-center/pkg/log"
	pb "github.com/apache/servicecomb-service-center/pkg/registry"
	"github.com/apache/servicecomb-service-center/pkg/util"
	apt "github.com/apache/servicecomb-service-center/server/core"
	"github.com/apache/servicecomb-service-center/server/health"
	"github.com/apache/servicecomb-service-center/server/plugin/quota"
	"github.com/apache/servicecomb-service-center/server/plugin/uuid"
	scerr "github.com/apache/servicecomb-service-center/server/scerror"
	"github.com/apache/servicecomb-service-center/server/service/cache"
	"os"
	"strconv"
	"time"
)

var (
	ttlFromEnv, _ = strconv.ParseInt(os.Getenv("INSTANCE_TTL"), 10, 0)
)

type InstanceService struct {
}

func (s *InstanceService) preProcessRegisterInstance(ctx context.Context, instance *pb.MicroServiceInstance) *scerr.Error {
	if len(instance.Status) == 0 {
		instance.Status = pb.MSI_UP
	}

	if len(instance.InstanceId) == 0 {
		instance.InstanceId = uuid.Generator().GetInstanceID(ctx)
	}

	instance.Timestamp = strconv.FormatInt(time.Now().Unix(), 10)
	instance.ModTimestamp = instance.Timestamp

	// 这里应该根据租约计时
	renewalInterval := apt.RegistryDefaultLeaseRenewalinterval
	retryTimes := apt.RegistryDefaultLeaseRetrytimes
	if instance.HealthCheck == nil {
		instance.HealthCheck = &pb.HealthCheck{
			Mode:     pb.CHECK_BY_HEARTBEAT,
			Interval: renewalInterval,
			Times:    retryTimes,
		}
	} else {
		// Health check对象仅用于呈现服务健康检查逻辑，如果CHECK_BY_PLATFORM类型，表明由sidecar代发心跳，实例120s超时
		switch instance.HealthCheck.Mode {
		case pb.CHECK_BY_HEARTBEAT:
			d := instance.HealthCheck.Interval * (instance.HealthCheck.Times + 1)
			if d <= 0 {
				return scerr.NewError(scerr.ErrInvalidParams, "Invalid 'healthCheck' settings in request body.")
			}
		case pb.CHECK_BY_PLATFORM:
			// 默认120s
			instance.HealthCheck.Interval = renewalInterval
			instance.HealthCheck.Times = retryTimes
		}
	}

	domainProject := util.ParseDomainProject(ctx)
	service, err := serviceUtil.GetService(ctx, domainProject, instance.ServiceId)
	if service == nil || err != nil {
		return scerr.NewError(scerr.ErrServiceNotExists, "Invalid 'serviceID' in request body.")
	}
	instance.Version = service.Version
	return nil
}

func (s *InstanceService) Register(ctx context.Context, in *pb.RegisterInstanceRequest) (*pb.RegisterInstanceResponse, error) {
	remoteIP := util.GetIPFromContext(ctx)

	if err := Validate(in); err != nil {
		log.Errorf(err, "register instance failed, invalid parameters, operator %s", remoteIP)
		return &pb.RegisterInstanceResponse{
			Response: pb.CreateResponse(scerr.ErrInvalidParams, err.Error()),
		}, nil
	}

	instance := in.Instance

	//允许自定义id
	if len(instance.InstanceId) > 0 {
		// keep alive the lease ttl
		// there are two reasons for sending a heartbeat here:
		// 1. in the scenario the instance has been removed,
		//    the cast of registration operation can be reduced.
		// 2. in the self-protection scenario, the instance is unhealthy
		//    and needs to be re-registered.
		resp, err := s.Heartbeat(ctx, &pb.HeartbeatRequest{ServiceId: instance.ServiceId, InstanceId: instance.InstanceId})
		switch resp.Response.GetCode() {
		case pb.ResponseSuccess:
			log.Infof("register instance successful, reuse instance[%s/%s], operator %s",
				instance.ServiceId, instance.InstanceId, remoteIP)
			return &pb.RegisterInstanceResponse{
				Response:   resp.Response,
				InstanceId: instance.InstanceId,
			}, nil
		case scerr.ErrInstanceNotExists:
			// register a new one
		default:
			log.Errorf(err, "register instance failed, reuse instance[%s/%s], operator %s",
				instance.ServiceId, instance.InstanceId, remoteIP)
			return &pb.RegisterInstanceResponse{
				Response: resp.Response,
			}, err
		}
	}

	if err := s.preProcessRegisterInstance(ctx, instance); err != nil {
		log.Errorf(err, "register service[%s]'s instance failed, endpoints %v, host '%s', operator %s",
			instance.ServiceId, instance.Endpoints, instance.HostName, remoteIP)
		return &pb.RegisterInstanceResponse{
			Response: pb.CreateResponseWithSCErr(err),
		}, nil
	}

	ttl := int64(instance.HealthCheck.Interval * (instance.HealthCheck.Times + 1))
	if ttlFromEnv > 0 {
		ttl = ttlFromEnv
	}
	instanceFlag := fmt.Sprintf("ttl %ds, endpoints %v, host '%s', serviceID %s",
		ttl, instance.Endpoints, instance.HostName, instance.ServiceId)

	//先以domain/project的方式组装
	domainProject := util.ParseDomainProject(ctx)

	var reporter *quota.ApplyQuotaResult
	if !apt.IsSCInstance(ctx) {
		res := quota.NewApplyQuotaResource(quota.MicroServiceInstanceQuotaType,
			domainProject, in.Instance.ServiceId, 1)
		reporter = quota.Apply(ctx, res)
		defer reporter.Close(ctx)

		if reporter.Err != nil {
			log.Errorf(reporter.Err, "register instance failed, %s, operator %s",
				instanceFlag, remoteIP)
			response := &pb.RegisterInstanceResponse{
				Response: pb.CreateResponseWithSCErr(reporter.Err),
			}
			if reporter.Err.InternalError() {
				return response, reporter.Err
			}
			return response, nil
		}
	}

	instanceID := instance.InstanceId
	data, err := json.Marshal(instance)
	if err != nil {
		log.Errorf(err,
			"register instance failed, %s, instanceID %s, operator %s",
			instanceFlag, instanceID, remoteIP)
		return &pb.RegisterInstanceResponse{
			Response: pb.CreateResponse(scerr.ErrInternal, err.Error()),
		}, err
	}

	leaseID, err := client.Instance().LeaseGrant(ctx, ttl)
	if err != nil {
		log.Errorf(err, "grant lease failed, %s, operator: %s", instanceFlag, remoteIP)
		return &pb.RegisterInstanceResponse{
			Response: pb.CreateResponse(scerr.ErrUnavailableBackend, err.Error()),
		}, err
	}

	// build the request options
	key := apt.GenerateInstanceKey(domainProject, instance.ServiceId, instanceID)
	hbKey := apt.GenerateInstanceLeaseKey(domainProject, instance.ServiceId, instanceID)

	opts := []client.PluginOp{
		client.OpPut(client.WithStrKey(key), client.WithValue(data),
			client.WithLease(leaseID)),
		client.OpPut(client.WithStrKey(hbKey), client.WithStrValue(fmt.Sprintf("%d", leaseID)),
			client.WithLease(leaseID)),
	}

	resp, err := client.Instance().TxnWithCmp(ctx, opts,
		[]client.CompareOp{client.OpCmp(
			client.CmpVer(util.StringToBytesWithNoCopy(apt.GenerateServiceKey(domainProject, instance.ServiceId))),
			client.CmpNotEqual, 0)},
		nil)
	if err != nil {
		log.Errorf(err,
			"register instance failed, %s, instanceID %s, operator %s",
			instanceFlag, instanceID, remoteIP)
		return &pb.RegisterInstanceResponse{
			Response: pb.CreateResponse(scerr.ErrUnavailableBackend, err.Error()),
		}, err
	}
	if !resp.Succeeded {
		log.Errorf(nil,
			"register instance failed, %s, instanceID %s, operator %s: service does not exist",
			instanceFlag, instanceID, remoteIP)
		return &pb.RegisterInstanceResponse{
			Response: pb.CreateResponse(scerr.ErrServiceNotExists, "Service does not exist."),
		}, nil
	}

	if err := reporter.ReportUsedQuota(ctx); err != nil {
		log.Errorf(err,
			"register instance failed, %s, instanceID %s, operator %s",
			instanceFlag, instanceID, remoteIP)
	}

	log.Infof("register instance %s, instanceID %s, operator %s",
		instanceFlag, instanceID, remoteIP)
	return &pb.RegisterInstanceResponse{
		Response:   pb.CreateResponse(pb.ResponseSuccess, "Register service instance successfully."),
		InstanceId: instanceID,
	}, nil
}

func (s *InstanceService) Unregister(ctx context.Context, in *pb.UnregisterInstanceRequest) (*pb.UnregisterInstanceResponse, error) {
	remoteIP := util.GetIPFromContext(ctx)

	if err := Validate(in); err != nil {
		log.Errorf(err, "unregister instance failed, invalid parameters, operator %s", remoteIP)
		return &pb.UnregisterInstanceResponse{
			Response: pb.CreateResponse(scerr.ErrInvalidParams, err.Error()),
		}, nil
	}

	domainProject := util.ParseDomainProject(ctx)
	serviceID := in.ServiceId
	instanceID := in.InstanceId

	instanceFlag := util.StringJoin([]string{serviceID, instanceID}, "/")

	err := revokeInstance(ctx, domainProject, serviceID, instanceID)
	if err != nil {
		log.Errorf(err, "unregister instance failed, instance[%s], operator %s: revoke instance failed", instanceFlag, remoteIP)
		resp := &pb.UnregisterInstanceResponse{
			Response: pb.CreateResponseWithSCErr(err),
		}
		if err.InternalError() {
			return resp, err
		}
		return resp, nil
	}

	log.Infof("unregister instance[%s], operator %s", instanceFlag, remoteIP)
	return &pb.UnregisterInstanceResponse{
		Response: pb.CreateResponse(pb.ResponseSuccess, "Unregister service instance successfully."),
	}, nil
}

func revokeInstance(ctx context.Context, domainProject string, serviceID string, instanceID string) *scerr.Error {
	leaseID, err := serviceUtil.GetLeaseID(ctx, domainProject, serviceID, instanceID)
	if err != nil {
		return scerr.NewError(scerr.ErrUnavailableBackend, err.Error())
	}
	if leaseID == -1 {
		return scerr.NewError(scerr.ErrInstanceNotExists, "Instance's leaseId not exist.")
	}

	err = client.Instance().LeaseRevoke(ctx, leaseID)
	if err != nil {
		if _, ok := err.(errorsEx.InternalError); !ok {
			return scerr.NewError(scerr.ErrInstanceNotExists, err.Error())
		}
		return scerr.NewError(scerr.ErrUnavailableBackend, err.Error())
	}
	return nil
}

func (s *InstanceService) Heartbeat(ctx context.Context, in *pb.HeartbeatRequest) (*pb.HeartbeatResponse, error) {
	remoteIP := util.GetIPFromContext(ctx)

	if err := Validate(in); err != nil {
		log.Errorf(err, "heartbeat failed, invalid parameters, operator %s", remoteIP)
		return &pb.HeartbeatResponse{
			Response: pb.CreateResponse(scerr.ErrInvalidParams, err.Error()),
		}, nil
	}

	domainProject := util.ParseDomainProject(ctx)
	instanceFlag := util.StringJoin([]string{in.ServiceId, in.InstanceId}, "/")

	_, ttl, err := serviceUtil.HeartbeatUtil(ctx, domainProject, in.ServiceId, in.InstanceId)
	if err != nil {
		log.Errorf(err, "heartbeat failed, instance[%s]. operator %s",
			instanceFlag, remoteIP)
		resp := &pb.HeartbeatResponse{
			Response: pb.CreateResponseWithSCErr(err),
		}
		if err.InternalError() {
			return resp, err
		}
		return resp, nil
	}

	if ttl == 0 {
		log.Errorf(errors.New("connect backend timed out"),
			"heartbeat successful, but renew instance[%s] failed. operator %s", instanceFlag, remoteIP)
	} else {
		log.Infof("heartbeat successful, renew instance[%s] ttl to %d. operator %s", instanceFlag, ttl, remoteIP)
	}
	return &pb.HeartbeatResponse{
		Response: pb.CreateResponse(pb.ResponseSuccess, "Update service instance heartbeat successfully."),
	}, nil
}

func (s *InstanceService) HeartbeatSet(ctx context.Context, in *pb.HeartbeatSetRequest) (*pb.HeartbeatSetResponse, error) {
	if len(in.Instances) == 0 {
		log.Errorf(nil, "heartbeats failed, invalid request. Body not contain Instances or is empty")
		return &pb.HeartbeatSetResponse{
			Response: pb.CreateResponse(scerr.ErrInvalidParams, "Request format invalid."),
		}, nil
	}
	domainProject := util.ParseDomainProject(ctx)

	heartBeatCount := len(in.Instances)
	existFlag := make(map[string]bool, heartBeatCount)
	instancesHbRst := make(chan *pb.InstanceHbRst, heartBeatCount)
	noMultiCounter := 0
	for _, heartbeatElement := range in.Instances {
		if _, ok := existFlag[heartbeatElement.ServiceId+heartbeatElement.InstanceId]; ok {
			log.Warnf("instance[%s/%s] is duplicate in heartbeat set", heartbeatElement.ServiceId, heartbeatElement.InstanceId)
			continue
		} else {
			existFlag[heartbeatElement.ServiceId+heartbeatElement.InstanceId] = true
			noMultiCounter++
		}
		gopool.Go(getHeartbeatFunc(ctx, domainProject, instancesHbRst, heartbeatElement))
	}
	count := 0
	successFlag := false
	failFlag := false
	instanceHbRstArr := make([]*pb.InstanceHbRst, 0, heartBeatCount)
	for heartbeat := range instancesHbRst {
		count++
		if len(heartbeat.ErrMessage) != 0 {
			failFlag = true
		} else {
			successFlag = true
		}
		instanceHbRstArr = append(instanceHbRstArr, heartbeat)
		if count == noMultiCounter {
			close(instancesHbRst)
		}
	}
	if !failFlag && successFlag {
		log.Infof("batch update heartbeats[%s] successfully", count)
		return &pb.HeartbeatSetResponse{
			Response:  pb.CreateResponse(pb.ResponseSuccess, "Heartbeat set successfully."),
			Instances: instanceHbRstArr,
		}, nil
	}
	log.Errorf(nil, "batch update heartbeats failed, %v", in.Instances)
	return &pb.HeartbeatSetResponse{
		Response:  pb.CreateResponse(scerr.ErrInstanceNotExists, "Heartbeat set failed."),
		Instances: instanceHbRstArr,
	}, nil
}

func getHeartbeatFunc(ctx context.Context, domainProject string, instancesHbRst chan<- *pb.InstanceHbRst, element *pb.HeartbeatSetElement) func(context.Context) {
	return func(_ context.Context) {
		hbRst := &pb.InstanceHbRst{
			ServiceId:  element.ServiceId,
			InstanceId: element.InstanceId,
			ErrMessage: "",
		}
		_, _, err := serviceUtil.HeartbeatUtil(ctx, domainProject, element.ServiceId, element.InstanceId)
		if err != nil {
			hbRst.ErrMessage = err.Error()
			log.Errorf(err, "heartbeat set failed, %s/%s", element.ServiceId, element.InstanceId)
		}
		instancesHbRst <- hbRst
	}
}

func (s *InstanceService) GetOneInstance(ctx context.Context, in *pb.GetOneInstanceRequest) (*pb.GetOneInstanceResponse, error) {
	err := Validate(in)
	if err != nil {
		log.Errorf(err, "get instance failed: invalid parameters")
		return &pb.GetOneInstanceResponse{
			Response: pb.CreateResponse(scerr.ErrInvalidParams, err.Error()),
		}, nil
	}

	domainProject := util.ParseDomainProject(ctx)

	service := &pb.MicroService{}
	if len(in.ConsumerServiceId) > 0 {
		service, err = serviceUtil.GetService(ctx, domainProject, in.ConsumerServiceId)
		if err != nil {
			log.Errorf(err, "get consumer failed, consumer[%s] find provider instance[%s/%s]",
				in.ConsumerServiceId, in.ProviderServiceId, in.ProviderInstanceId)
			return &pb.GetOneInstanceResponse{
				Response: pb.CreateResponse(scerr.ErrInternal, err.Error()),
			}, err
		}
		if service == nil {
			log.Errorf(nil, "consumer does not exist, consumer[%s] find provider instance[%s/%s]",
				in.ConsumerServiceId, in.ProviderServiceId, in.ProviderInstanceId)
			return &pb.GetOneInstanceResponse{
				Response: pb.CreateResponse(scerr.ErrServiceNotExists,
					fmt.Sprintf("Consumer[%s] does not exist.", in.ConsumerServiceId)),
			}, nil
		}
	}

	provider, err := serviceUtil.GetService(ctx, domainProject, in.ProviderServiceId)
	if err != nil {
		log.Errorf(err, "get provider failed, consumer[%s] find provider instance[%s/%s]",
			in.ConsumerServiceId, in.ProviderServiceId, in.ProviderInstanceId)
		return &pb.GetOneInstanceResponse{
			Response: pb.CreateResponse(scerr.ErrInternal, err.Error()),
		}, err
	}
	if provider == nil {
		log.Errorf(nil, "provider does not exist, consumer[%s] find provider instance[%s/%s]",
			in.ConsumerServiceId, in.ProviderServiceId, in.ProviderInstanceId)
		return &pb.GetOneInstanceResponse{
			Response: pb.CreateResponse(scerr.ErrServiceNotExists,
				fmt.Sprintf("Provider[%s] does not exist.", in.ProviderServiceId)),
		}, nil
	}

	findFlag := func() string {
		return fmt.Sprintf("Consumer[%s][%s/%s/%s/%s] find provider[%s][%s/%s/%s/%s] instance[%s]",
			in.ConsumerServiceId, service.Environment, service.AppId, service.ServiceName, service.Version,
			provider.ServiceId, provider.Environment, provider.AppId, provider.ServiceName, provider.Version,
			in.ProviderInstanceId)
	}

	var item *cache.VersionRuleCacheItem
	rev, _ := ctx.Value(util.CtxRequestRevision).(string)
	item, err = cache.FindInstances.GetWithProviderID(ctx, service, pb.MicroServiceToKey(domainProject, provider),
		&pb.HeartbeatSetElement{
			ServiceId: in.ProviderServiceId, InstanceId: in.ProviderInstanceId,
		}, in.Tags, rev)
	if err != nil {
		log.Errorf(err, "FindInstances.GetWithProviderID failed, %s failed", findFlag())
		return &pb.GetOneInstanceResponse{
			Response: pb.CreateResponse(scerr.ErrInternal, err.Error()),
		}, err
	}
	if item == nil || len(item.Instances) == 0 {
		mes := fmt.Errorf("%s failed, provider instance does not exist", findFlag())
		log.Errorf(mes, "FindInstances.GetWithProviderID failed")
		return &pb.GetOneInstanceResponse{
			Response: pb.CreateResponse(scerr.ErrInstanceNotExists, mes.Error()),
		}, nil
	}

	instance := item.Instances[0]
	if rev == item.Rev {
		instance = nil // for gRPC
	}
	_ = util.SetContext(ctx, util.CtxResponseRevision, item.Rev)

	return &pb.GetOneInstanceResponse{
		Response: pb.CreateResponse(pb.ResponseSuccess, "Get instance successfully."),
		Instance: instance,
	}, nil
}

func (s *InstanceService) GetInstances(ctx context.Context, in *pb.GetInstancesRequest) (*pb.GetInstancesResponse, error) {
	err := Validate(in)
	if err != nil {
		log.Errorf(err, "get instances failed: invalid parameters")
		return &pb.GetInstancesResponse{
			Response: pb.CreateResponse(scerr.ErrInvalidParams, err.Error()),
		}, nil
	}

	domainProject := util.ParseDomainProject(ctx)

	service := &pb.MicroService{}
	if len(in.ConsumerServiceId) > 0 {
		service, err = serviceUtil.GetService(ctx, domainProject, in.ConsumerServiceId)
		if err != nil {
			log.Errorf(err, "get consumer failed, consumer[%s] find provider instances",
				in.ConsumerServiceId, in.ProviderServiceId)
			return &pb.GetInstancesResponse{
				Response: pb.CreateResponse(scerr.ErrInternal, err.Error()),
			}, err
		}
		if service == nil {
			log.Errorf(nil, "consumer does not exist, consumer[%s] find provider instances",
				in.ConsumerServiceId, in.ProviderServiceId)
			return &pb.GetInstancesResponse{
				Response: pb.CreateResponse(scerr.ErrServiceNotExists,
					fmt.Sprintf("Consumer[%s] does not exist.", in.ConsumerServiceId)),
			}, nil
		}
	}

	provider, err := serviceUtil.GetService(ctx, domainProject, in.ProviderServiceId)
	if err != nil {
		log.Errorf(err, "get provider failed, consumer[%s] find provider instances",
			in.ConsumerServiceId, in.ProviderServiceId)
		return &pb.GetInstancesResponse{
			Response: pb.CreateResponse(scerr.ErrInternal, err.Error()),
		}, err
	}
	if provider == nil {
		log.Errorf(nil, "provider does not exist, consumer[%s] find provider instances",
			in.ConsumerServiceId, in.ProviderServiceId)
		return &pb.GetInstancesResponse{
			Response: pb.CreateResponse(scerr.ErrServiceNotExists,
				fmt.Sprintf("Provider[%s] does not exist.", in.ProviderServiceId)),
		}, nil
	}

	findFlag := func() string {
		return fmt.Sprintf("Consumer[%s][%s/%s/%s/%s] find provider[%s][%s/%s/%s/%s] instances",
			in.ConsumerServiceId, service.Environment, service.AppId, service.ServiceName, service.Version,
			provider.ServiceId, provider.Environment, provider.AppId, provider.ServiceName, provider.Version)
	}

	var item *cache.VersionRuleCacheItem
	rev, _ := ctx.Value(util.CtxRequestRevision).(string)
	item, err = cache.FindInstances.GetWithProviderID(ctx, service, pb.MicroServiceToKey(domainProject, provider),
		&pb.HeartbeatSetElement{
			ServiceId: in.ProviderServiceId,
		}, in.Tags, rev)
	if err != nil {
		log.Errorf(err, "FindInstances.GetWithProviderID failed, %s failed", findFlag())
		return &pb.GetInstancesResponse{
			Response: pb.CreateResponse(scerr.ErrInternal, err.Error()),
		}, err
	}
	if item == nil || len(item.ServiceIds) == 0 {
		mes := fmt.Errorf("%s failed, provider instance does not exist", findFlag())
		log.Errorf(mes, "FindInstances.GetWithProviderID failed")
		return &pb.GetInstancesResponse{
			Response: pb.CreateResponse(scerr.ErrServiceNotExists, mes.Error()),
		}, nil
	}

	instances := item.Instances
	if rev == item.Rev {
		instances = nil // for gRPC
	}
	_ = util.SetContext(ctx, util.CtxResponseRevision, item.Rev)

	return &pb.GetInstancesResponse{
		Response:  pb.CreateResponse(pb.ResponseSuccess, "Query service instances successfully."),
		Instances: instances,
	}, nil
}

func (s *InstanceService) Find(ctx context.Context, in *pb.FindInstancesRequest) (*pb.FindInstancesResponse, error) {
	err := Validate(in)
	if err != nil {
		log.Errorf(err, "find instance failed: invalid parameters")
		return &pb.FindInstancesResponse{
			Response: pb.CreateResponse(scerr.ErrInvalidParams, err.Error()),
		}, nil
	}

	provider := &pb.MicroServiceKey{
		Tenant:      util.ParseTargetDomainProject(ctx),
		Environment: in.Environment,
		AppId:       in.AppId,
		ServiceName: in.ServiceName,
		Alias:       in.ServiceName,
		Version:     in.VersionRule,
	}

	rev, ok := ctx.Value(util.CtxRequestRevision).(string)
	if !ok {
		err = errors.New("rev in context is not type string")
		log.Error("", err)
		return &pb.FindInstancesResponse{
			Response: pb.CreateResponse(scerr.ErrInternal, err.Error()),
		}, err
	}

	if apt.IsShared(provider) {
		return s.findSharedServiceInstance(ctx, in, provider, rev)
	}
	return s.findInstance(ctx, in, provider, rev)
}

func (s *InstanceService) findInstance(ctx context.Context, in *pb.FindInstancesRequest,
	provider *pb.MicroServiceKey, rev string) (*pb.FindInstancesResponse, error) {
	var err error
	domainProject := util.ParseDomainProject(ctx)
	service := &pb.MicroService{Environment: in.Environment}
	if len(in.ConsumerServiceId) > 0 {
		service, err = serviceUtil.GetService(ctx, domainProject, in.ConsumerServiceId)
		if err != nil {
			log.Errorf(err, "get consumer failed, consumer[%s] find provider[%s/%s/%s/%s]",
				in.ConsumerServiceId, in.Environment, in.AppId, in.ServiceName, in.VersionRule)
			return &pb.FindInstancesResponse{
				Response: pb.CreateResponse(scerr.ErrInternal, err.Error()),
			}, err
		}
		if service == nil {
			log.Errorf(nil, "consumer does not exist, consumer[%s] find provider[%s/%s/%s/%s]",
				in.ConsumerServiceId, in.Environment, in.AppId, in.ServiceName, in.VersionRule)
			return &pb.FindInstancesResponse{
				Response: pb.CreateResponse(scerr.ErrServiceNotExists,
					fmt.Sprintf("Consumer[%s] does not exist.", in.ConsumerServiceId)),
			}, nil
		}
		provider.Environment = service.Environment
	}

	// provider is not a shared micro-service,
	// only allow shared micro-service instances found in different domains.
	ctx = util.SetTargetDomainProject(ctx, util.ParseDomain(ctx), util.ParseProject(ctx))
	provider.Tenant = util.ParseTargetDomainProject(ctx)

	findFlag := fmt.Sprintf("Consumer[%s][%s/%s/%s/%s] find provider[%s/%s/%s/%s]",
		in.ConsumerServiceId, service.Environment, service.AppId, service.ServiceName, service.Version,
		provider.Environment, provider.AppId, provider.ServiceName, provider.Version)

	// cache
	var item *cache.VersionRuleCacheItem
	item, err = cache.FindInstances.Get(ctx, service, provider, in.Tags, rev)
	if err != nil {
		log.Errorf(err, "FindInstancesCache.Get failed, %s failed", findFlag)
		return &pb.FindInstancesResponse{
			Response: pb.CreateResponse(scerr.ErrInternal, err.Error()),
		}, err
	}
	if item == nil {
		mes := fmt.Errorf("%s failed, provider does not exist", findFlag)
		log.Errorf(mes, "FindInstancesCache.Get failed")
		return &pb.FindInstancesResponse{
			Response: pb.CreateResponse(scerr.ErrServiceNotExists, mes.Error()),
		}, nil
	}

	// add dependency queue
	if len(in.ConsumerServiceId) > 0 &&
		len(item.ServiceIds) > 0 &&
		!cache.DependencyRule.ExistVersionRule(ctx, in.ConsumerServiceId, provider) {
		provider, err = s.reshapeProviderKey(ctx, provider, item.ServiceIds[0])
		if err != nil {
			return nil, err
		}
		if provider != nil {
			err = serviceUtil.AddServiceVersionRule(ctx, domainProject, service, provider)
		} else {
			mes := fmt.Errorf("%s failed, provider does not exist", findFlag)
			log.Errorf(mes, "AddServiceVersionRule failed")
			return &pb.FindInstancesResponse{
				Response: pb.CreateResponse(scerr.ErrServiceNotExists, mes.Error()),
			}, nil
		}
		if err != nil {
			log.Errorf(err, "AddServiceVersionRule failed, %s failed", findFlag)
			return &pb.FindInstancesResponse{
				Response: pb.CreateResponse(scerr.ErrInternal, err.Error()),
			}, err
		}
	}

	return s.genFindResult(ctx, rev, item)
}

func (s *InstanceService) findSharedServiceInstance(ctx context.Context, in *pb.FindInstancesRequest,
	provider *pb.MicroServiceKey, rev string) (*pb.FindInstancesResponse, error) {
	var err error
	service := &pb.MicroService{Environment: in.Environment}
	// it means the shared micro-services must be the same env with SC.
	provider.Environment = apt.Service.Environment
	findFlag := fmt.Sprintf("find shared provider[%s/%s/%s/%s]", provider.Environment, provider.AppId, provider.ServiceName, provider.Version)

	// cache
	var item *cache.VersionRuleCacheItem
	item, err = cache.FindInstances.Get(ctx, service, provider, in.Tags, rev)
	if err != nil {
		log.Errorf(err, "FindInstancesCache.Get failed, %s failed", findFlag)
		return &pb.FindInstancesResponse{
			Response: pb.CreateResponse(scerr.ErrInternal, err.Error()),
		}, err
	}
	if item == nil {
		mes := fmt.Errorf("%s failed, provider does not exist", findFlag)
		log.Errorf(mes, "FindInstancesCache.Get failed")
		return &pb.FindInstancesResponse{
			Response: pb.CreateResponse(scerr.ErrServiceNotExists, mes.Error()),
		}, nil
	}

	return s.genFindResult(ctx, rev, item)
}

func (s *InstanceService) genFindResult(ctx context.Context, oldRev string, item *cache.VersionRuleCacheItem) (*pb.FindInstancesResponse, error) {
	instances := item.Instances
	if oldRev == item.Rev {
		instances = nil // for gRPC
	}
	// TODO support gRPC output context
	_ = util.SetContext(ctx, util.CtxResponseRevision, item.Rev)
	return &pb.FindInstancesResponse{
		Response:  pb.CreateResponse(pb.ResponseSuccess, "Query service instances successfully."),
		Instances: instances,
	}, nil
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

	response := &pb.BatchFindInstancesResponse{
		Response: pb.CreateResponse(pb.ResponseSuccess, "Batch query service instances successfully."),
	}

	// find services
	response.Services, err = s.batchFindServices(ctx, in)
	if err != nil {
		return &pb.BatchFindInstancesResponse{
			Response: pb.CreateResponse(scerr.ErrInternal, err.Error()),
		}, err
	}

	// find instance
	response.Instances, err = s.batchFindInstances(ctx, in)
	if err != nil {
		return &pb.BatchFindInstancesResponse{
			Response: pb.CreateResponse(scerr.ErrInternal, err.Error()),
		}, err
	}

	return response, nil
}

func (s *InstanceService) batchFindServices(ctx context.Context, in *pb.BatchFindInstancesRequest) (*pb.BatchFindResult, error) {
	if len(in.Services) == 0 {
		return nil, nil
	}
	cloneCtx := util.CloneContext(ctx)

	services := &pb.BatchFindResult{}
	failedResult := make(map[int32]*pb.FindFailedResult)
	for index, key := range in.Services {
		findCtx := util.SetContext(cloneCtx, util.CtxRequestRevision, key.Rev)
		resp, err := s.Find(findCtx, &pb.FindInstancesRequest{
			ConsumerServiceId: in.ConsumerServiceId,
			AppId:             key.Service.AppId,
			ServiceName:       key.Service.ServiceName,
			VersionRule:       key.Service.Version,
			Environment:       key.Service.Environment,
		})
		if err != nil {
			return nil, err
		}
		failed, ok := failedResult[resp.Response.GetCode()]
		serviceUtil.AppendFindResponse(findCtx, int64(index), resp.Response, resp.Instances,
			&services.Updated, &services.NotModified, &failed)
		if !ok && failed != nil {
			failedResult[resp.Response.GetCode()] = failed
		}
	}
	for _, result := range failedResult {
		services.Failed = append(services.Failed, result)
	}
	return services, nil
}

func (s *InstanceService) batchFindInstances(ctx context.Context, in *pb.BatchFindInstancesRequest) (*pb.BatchFindResult, error) {
	if len(in.Instances) == 0 {
		return nil, nil
	}
	cloneCtx := util.CloneContext(ctx)
	// can not find the shared provider instances
	cloneCtx = util.SetTargetDomainProject(cloneCtx, util.ParseDomain(ctx), util.ParseProject(ctx))

	instances := &pb.BatchFindResult{}
	failedResult := make(map[int32]*pb.FindFailedResult)
	for index, key := range in.Instances {
		getCtx := util.SetContext(cloneCtx, util.CtxRequestRevision, key.Rev)
		resp, err := s.GetOneInstance(getCtx, &pb.GetOneInstanceRequest{
			ConsumerServiceId:  in.ConsumerServiceId,
			ProviderServiceId:  key.Instance.ServiceId,
			ProviderInstanceId: key.Instance.InstanceId,
		})
		if err != nil {
			return nil, err
		}
		failed, ok := failedResult[resp.Response.GetCode()]
		serviceUtil.AppendFindResponse(getCtx, int64(index), resp.Response, []*pb.MicroServiceInstance{resp.Instance},
			&instances.Updated, &instances.NotModified, &failed)
		if !ok && failed != nil {
			failedResult[resp.Response.GetCode()] = failed
		}
	}
	for _, result := range failedResult {
		instances.Failed = append(instances.Failed, result)
	}
	return instances, nil
}

func (s *InstanceService) reshapeProviderKey(ctx context.Context, provider *pb.MicroServiceKey, providerID string) (*pb.MicroServiceKey, error) {
	//维护version的规则,service name 可能是别名，所以重新获取
	providerService, err := serviceUtil.GetService(ctx, provider.Tenant, providerID)
	if providerService == nil {
		return nil, err
	}

	versionRule := provider.Version
	provider = pb.MicroServiceToKey(provider.Tenant, providerService)
	provider.Version = versionRule
	return provider, nil
}

func (s *InstanceService) UpdateStatus(ctx context.Context, in *pb.UpdateInstanceStatusRequest) (*pb.UpdateInstanceStatusResponse, error) {
	domainProject := util.ParseDomainProject(ctx)
	updateStatusFlag := util.StringJoin([]string{in.ServiceId, in.InstanceId, in.Status}, "/")
	if err := Validate(in); err != nil {
		log.Errorf(nil, "update instance[%s] status failed", updateStatusFlag)
		return &pb.UpdateInstanceStatusResponse{
			Response: pb.CreateResponse(scerr.ErrInvalidParams, err.Error()),
		}, nil
	}

	instance, err := serviceUtil.GetInstance(ctx, domainProject, in.ServiceId, in.InstanceId)
	if err != nil {
		log.Errorf(err, "update instance[%s] status failed", updateStatusFlag)
		return &pb.UpdateInstanceStatusResponse{
			Response: pb.CreateResponse(scerr.ErrInternal, err.Error()),
		}, err
	}
	if instance == nil {
		log.Errorf(nil, "update instance[%s] status failed, instance does not exist", updateStatusFlag)
		return &pb.UpdateInstanceStatusResponse{
			Response: pb.CreateResponse(scerr.ErrInstanceNotExists, "Service instance does not exist."),
		}, nil
	}

	copyInstanceRef := *instance
	copyInstanceRef.Status = in.Status

	if err := serviceUtil.UpdateInstance(ctx, domainProject, &copyInstanceRef); err != nil {
		log.Errorf(err, "update instance[%s] status failed", updateStatusFlag)
		resp := &pb.UpdateInstanceStatusResponse{
			Response: pb.CreateResponseWithSCErr(err),
		}
		if err.InternalError() {
			return resp, err
		}
		return resp, nil
	}

	log.Infof("update instance[%s] status successfully", updateStatusFlag)
	return &pb.UpdateInstanceStatusResponse{
		Response: pb.CreateResponse(pb.ResponseSuccess, "Update service instance status successfully."),
	}, nil
}

func (s *InstanceService) UpdateInstanceProperties(ctx context.Context, in *pb.UpdateInstancePropsRequest) (*pb.UpdateInstancePropsResponse, error) {
	domainProject := util.ParseDomainProject(ctx)
	instanceFlag := util.StringJoin([]string{in.ServiceId, in.InstanceId}, "/")
	if err := Validate(in); err != nil {
		log.Errorf(nil, "update instance[%s] properties failed", instanceFlag)
		return &pb.UpdateInstancePropsResponse{
			Response: pb.CreateResponse(scerr.ErrInvalidParams, err.Error()),
		}, nil
	}

	instance, err := serviceUtil.GetInstance(ctx, domainProject, in.ServiceId, in.InstanceId)
	if err != nil {
		log.Errorf(err, "update instance[%s] properties failed", instanceFlag)
		return &pb.UpdateInstancePropsResponse{
			Response: pb.CreateResponse(scerr.ErrInternal, err.Error()),
		}, err
	}
	if instance == nil {
		log.Errorf(nil, "update instance[%s] properties failed, instance does not exist", instanceFlag)
		return &pb.UpdateInstancePropsResponse{
			Response: pb.CreateResponse(scerr.ErrInstanceNotExists, "Service instance does not exist."),
		}, nil
	}

	copyInstanceRef := *instance
	copyInstanceRef.Properties = in.Properties

	if err := serviceUtil.UpdateInstance(ctx, domainProject, &copyInstanceRef); err != nil {
		log.Errorf(err, "update instance[%s] properties failed", instanceFlag)
		resp := &pb.UpdateInstancePropsResponse{
			Response: pb.CreateResponseWithSCErr(err),
		}
		if err.InternalError() {
			return resp, err
		}
		return resp, nil
	}

	log.Infof("update instance[%s] properties successfully", instanceFlag)
	return &pb.UpdateInstancePropsResponse{
		Response: pb.CreateResponse(pb.ResponseSuccess, "Update service instance properties successfully."),
	}, nil
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
