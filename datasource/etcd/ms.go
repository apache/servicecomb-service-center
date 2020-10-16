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

package etcd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/apache/servicecomb-service-center/pkg/gopool"
	"github.com/apache/servicecomb-service-center/pkg/log"
	pb "github.com/apache/servicecomb-service-center/pkg/registry"
	"github.com/apache/servicecomb-service-center/pkg/util"
	apt "github.com/apache/servicecomb-service-center/server/core"
	"github.com/apache/servicecomb-service-center/server/core/backend"
	"github.com/apache/servicecomb-service-center/server/core/proto"
	"github.com/apache/servicecomb-service-center/server/plugin"
	"github.com/apache/servicecomb-service-center/server/plugin/quota"
	"github.com/apache/servicecomb-service-center/server/plugin/registry"
	scerr "github.com/apache/servicecomb-service-center/server/scerror"
	"github.com/apache/servicecomb-service-center/server/service/cache"
	serviceUtil "github.com/apache/servicecomb-service-center/server/service/util"
	"strconv"
	"time"
)

// RegisterService() implement:
// 1. capsule request to etcd kv format
// 2. invoke etcd client to store data
// 3. check etcd-client response && construct createServiceResponse
func (ds *DataSource) RegisterService(ctx context.Context, request *pb.CreateServiceRequest) (
	*pb.CreateServiceResponse, error) {
	// start to store service to etcd
	serviceBody := request.Service

	// construct data to invoke etcd client
	opts, uniqueCmpOpts, failOpts, err := capRegisterData(ctx, request)
	if err != nil {
		return &pb.CreateServiceResponse{
			Response: proto.CreateResponse(scerr.ErrInternal, err.Error()),
		}, err
	}
	resp, err := backend.Registry().TxnWithCmp(ctx, opts, uniqueCmpOpts, failOpts)

	return newRegisterServiceResp(ctx, serviceBody, resp, err)
}

func (ds *DataSource) GetServices(ctx context.Context, request *pb.GetServicesRequest) (
	*pb.GetServicesResponse, error) {
	services, err := serviceUtil.GetAllServiceUtil(ctx)
	if err != nil {
		log.Errorf(err, "get all services by domain failed")
		return &pb.GetServicesResponse{
			Response: proto.CreateResponse(scerr.ErrInternal, err.Error()),
		}, err
	}

	return &pb.GetServicesResponse{
		Response: proto.CreateResponse(proto.Response_SUCCESS, "Get all services successfully."),
		Services: services,
	}, nil
}

func (ds *DataSource) GetService(ctx context.Context, request *pb.GetServiceRequest) (
	*pb.GetServiceResponse, error) {
	domainProject := util.ParseDomainProject(ctx)
	singleService, err := serviceUtil.GetService(ctx, domainProject, request.ServiceId)

	if err != nil {
		log.Errorf(err, "get micro-service[%s] failed, get service file failed", request.ServiceId)
		return &pb.GetServiceResponse{
			Response: proto.CreateResponse(scerr.ErrInternal, err.Error()),
		}, err
	}
	if singleService == nil {
		log.Errorf(nil, "get micro-service[%s] failed, service does not exist", request.ServiceId)
		return &pb.GetServiceResponse{
			Response: proto.CreateResponse(scerr.ErrServiceNotExists, "Service does not exist."),
		}, nil
	}
	return &pb.GetServiceResponse{
		Response: proto.CreateResponse(proto.Response_SUCCESS, "Get service successfully."),
		Service:  singleService,
	}, nil
}

func (ds *DataSource) ExistService(ctx context.Context, request *pb.GetExistenceRequest) (*pb.GetExistenceResponse,
	error) {
	domainProject := util.ParseDomainProject(ctx)
	serviceFlag := util.StringJoin([]string{
		request.Environment, request.AppId, request.ServiceName, request.Version}, "/")

	ids, exist, err := serviceUtil.FindServiceIds(ctx, request.Version, &pb.MicroServiceKey{
		Environment: request.Environment,
		AppId:       request.AppId,
		ServiceName: request.ServiceName,
		Alias:       request.ServiceName,
		Version:     request.Version,
		Tenant:      domainProject,
	})
	if err != nil {
		log.Errorf(err, "micro-service[%s] exist failed, find serviceIDs failed", serviceFlag)
		return &pb.GetExistenceResponse{
			Response: proto.CreateResponse(scerr.ErrInternal, err.Error()),
		}, err
	}
	if !exist {
		log.Infof("micro-service[%s] exist failed, service does not exist", serviceFlag)
		return &pb.GetExistenceResponse{
			Response: proto.CreateResponse(scerr.ErrServiceNotExists, serviceFlag+" does not exist."),
		}, nil
	}
	if len(ids) == 0 {
		log.Infof("micro-service[%s] exist failed, version mismatch", serviceFlag)
		return &pb.GetExistenceResponse{
			Response: proto.CreateResponse(scerr.ErrServiceVersionNotExists, serviceFlag+" version mismatch."),
		}, nil
	}
	return &pb.GetExistenceResponse{
		Response:  proto.CreateResponse(proto.Response_SUCCESS, "get service id successfully."),
		ServiceId: ids[0], // 约定多个时，取较新版本
	}, nil
}

func (ds *DataSource) UpdateService(ctx context.Context, request *pb.UpdateServicePropsRequest) (
	*pb.UpdateServicePropsResponse, error) {
	remoteIP := util.GetIPFromContext(ctx)
	domainProject := util.ParseDomainProject(ctx)

	key := apt.GenerateServiceKey(domainProject, request.ServiceId)
	microservice, err := serviceUtil.GetService(ctx, domainProject, request.ServiceId)
	if err != nil {
		log.Errorf(err, "update service[%s] properties failed, get service file failed, operator: %s",
			request.ServiceId, remoteIP)
		return &pb.UpdateServicePropsResponse{
			Response: proto.CreateResponse(scerr.ErrInternal, err.Error()),
		}, err
	}
	if microservice == nil {
		log.Errorf(nil, "update service[%s] properties failed, service does not exist, operator: %s",
			request.ServiceId, remoteIP)
		return &pb.UpdateServicePropsResponse{
			Response: proto.CreateResponse(scerr.ErrServiceNotExists, "Service does not exist."),
		}, nil
	}

	copyServiceRef := *microservice
	copyServiceRef.Properties = request.Properties
	copyServiceRef.ModTimestamp = strconv.FormatInt(time.Now().Unix(), 10)

	data, err := json.Marshal(copyServiceRef)
	if err != nil {
		log.Errorf(err, "update service[%s] properties failed, json marshal service failed, operator: %s",
			request.ServiceId, remoteIP)
		return &pb.UpdateServicePropsResponse{
			Response: proto.CreateResponse(scerr.ErrInternal, err.Error()),
		}, err
	}

	// Set key file
	resp, err := backend.Registry().TxnWithCmp(ctx,
		[]registry.PluginOp{registry.OpPut(registry.WithStrKey(key), registry.WithValue(data))},
		[]registry.CompareOp{registry.OpCmp(
			registry.CmpVer(util.StringToBytesWithNoCopy(key)),
			registry.CmpNotEqual, 0)},
		nil)
	if err != nil {
		log.Errorf(err, "update service[%s] properties failed, operator: %s", request.ServiceId, remoteIP)
		return &pb.UpdateServicePropsResponse{
			Response: proto.CreateResponse(scerr.ErrUnavailableBackend, err.Error()),
		}, err
	}
	if !resp.Succeeded {
		log.Errorf(err, "update service[%s] properties failed, service does not exist, operator: %s",
			request.ServiceId, remoteIP)
		return &pb.UpdateServicePropsResponse{
			Response: proto.CreateResponse(scerr.ErrServiceNotExists, "Service does not exist."),
		}, nil
	}

	log.Infof("update service[%s] properties successfully, operator: %s", request.ServiceId, remoteIP)
	return &pb.UpdateServicePropsResponse{
		Response: proto.CreateResponse(proto.Response_SUCCESS, "update service successfully."),
	}, nil
}

func (ds *DataSource) UnregisterService(ctx context.Context, request *pb.DeleteServiceRequest) (
	*pb.DeleteServiceResponse, error) {
	resp, err := ds.DeleteServicePri(ctx, request.ServiceId, request.Force)
	return &pb.DeleteServiceResponse{
		Response: resp,
	}, err
}

func (ds *DataSource) RegisterInstance(ctx context.Context, request *pb.RegisterInstanceRequest) (
	*pb.RegisterInstanceResponse, error) {
	remoteIP := util.GetIPFromContext(ctx)
	instance := request.Instance

	//允许自定义id
	if len(instance.InstanceId) > 0 {
		// keep alive the lease ttl
		// there are two reasons for sending a heartbeat here:
		// 1. in the scenario the instance has been removed,
		//    the cast of registration operation can be reduced.
		// 2. in the self-protection scenario, the instance is unhealthy
		//    and needs to be re-registered.
		resp, err := ds.Heartbeat(ctx, &pb.HeartbeatRequest{ServiceId: instance.ServiceId,
			InstanceId: instance.InstanceId})
		if resp == nil {
			log.Errorf(err, "register service[%s]'s instance failed, endpoints %v, host '%s', operator %s",
				instance.ServiceId, instance.Endpoints, instance.HostName, remoteIP)
			return &pb.RegisterInstanceResponse{
				Response: proto.CreateResponse(scerr.ErrInternal, err.Error()),
			}, nil
		}
		switch resp.Response.GetCode() {
		case proto.Response_SUCCESS:
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

	if err := preProcessRegisterInstance(ctx, instance); err != nil {
		log.Errorf(err, "register service[%s]'s instance failed, endpoints %v, host '%s', operator %s",
			instance.ServiceId, instance.Endpoints, instance.HostName, remoteIP)
		return &pb.RegisterInstanceResponse{
			Response: proto.CreateResponseWithSCErr(err),
		}, nil
	}

	ttl := int64(instance.HealthCheck.Interval * (instance.HealthCheck.Times + 1))
	if ds.ttlFromEnv > 0 {
		ttl = ds.ttlFromEnv
	}
	instanceFlag := fmt.Sprintf("ttl %ds, endpoints %v, host '%s', serviceID %s",
		ttl, instance.Endpoints, instance.HostName, instance.ServiceId)

	//先以domain/project的方式组装
	domainProject := util.ParseDomainProject(ctx)

	var reporter *quota.ApplyQuotaResult
	if !apt.IsSCInstance(ctx) {
		res := quota.NewApplyQuotaResource(quota.MicroServiceInstanceQuotaType,
			domainProject, request.Instance.ServiceId, 1)
		reporter = plugin.Plugins().Quota().Apply4Quotas(ctx, res)
		defer reporter.Close(ctx)

		if reporter.Err != nil {
			log.Errorf(reporter.Err, "register instance failed, %s, operator %s",
				instanceFlag, remoteIP)
			response := &pb.RegisterInstanceResponse{
				Response: proto.CreateResponseWithSCErr(reporter.Err),
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
			Response: proto.CreateResponse(scerr.ErrInternal, err.Error()),
		}, err
	}

	leaseID, err := backend.Registry().LeaseGrant(ctx, ttl)
	if err != nil {
		log.Errorf(err, "grant lease failed, %s, operator: %s", instanceFlag, remoteIP)
		return &pb.RegisterInstanceResponse{
			Response: proto.CreateResponse(scerr.ErrUnavailableBackend, err.Error()),
		}, err
	}

	// build the request options
	key := apt.GenerateInstanceKey(domainProject, instance.ServiceId, instanceID)
	hbKey := apt.GenerateInstanceLeaseKey(domainProject, instance.ServiceId, instanceID)

	opts := []registry.PluginOp{
		registry.OpPut(registry.WithStrKey(key), registry.WithValue(data),
			registry.WithLease(leaseID)),
		registry.OpPut(registry.WithStrKey(hbKey), registry.WithStrValue(fmt.Sprintf("%d", leaseID)),
			registry.WithLease(leaseID)),
	}

	resp, err := backend.Registry().TxnWithCmp(ctx, opts,
		[]registry.CompareOp{registry.OpCmp(
			registry.CmpVer(util.StringToBytesWithNoCopy(apt.GenerateServiceKey(domainProject, instance.ServiceId))),
			registry.CmpNotEqual, 0)},
		nil)
	if err != nil {
		log.Errorf(err,
			"register instance failed, %s, instanceID %s, operator %s",
			instanceFlag, instanceID, remoteIP)
		return &pb.RegisterInstanceResponse{
			Response: proto.CreateResponse(scerr.ErrUnavailableBackend, err.Error()),
		}, err
	}
	if !resp.Succeeded {
		log.Errorf(nil,
			"register instance failed, %s, instanceID %s, operator %s: service does not exist",
			instanceFlag, instanceID, remoteIP)
		return &pb.RegisterInstanceResponse{
			Response: proto.CreateResponse(scerr.ErrServiceNotExists, "Service does not exist."),
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
		Response:   proto.CreateResponse(proto.Response_SUCCESS, "Register service instance successfully."),
		InstanceId: instanceID,
	}, nil
}

func (ds *DataSource) GetInstance(ctx context.Context, in *pb.GetOneInstanceRequest) (
	*pb.GetOneInstanceResponse, error) {
	domainProject := util.ParseDomainProject(ctx)

	service := &pb.MicroService{}
	var err error
	if len(in.ConsumerServiceId) > 0 {
		service, err = serviceUtil.GetService(ctx, domainProject, in.ConsumerServiceId)
		if err != nil {
			log.Errorf(err, "get consumer failed, consumer[%s] find provider instance[%s/%s]",
				in.ConsumerServiceId, in.ProviderServiceId, in.ProviderInstanceId)
			return &pb.GetOneInstanceResponse{
				Response: proto.CreateResponse(scerr.ErrInternal, err.Error()),
			}, err
		}
		if service == nil {
			log.Errorf(nil, "consumer does not exist, consumer[%s] find provider instance[%s/%s]",
				in.ConsumerServiceId, in.ProviderServiceId, in.ProviderInstanceId)
			return &pb.GetOneInstanceResponse{
				Response: proto.CreateResponse(scerr.ErrServiceNotExists,
					fmt.Sprintf("Consumer[%s] does not exist.", in.ConsumerServiceId)),
			}, nil
		}
	}

	provider, err := serviceUtil.GetService(ctx, domainProject, in.ProviderServiceId)
	if err != nil {
		log.Errorf(err, "get provider failed, consumer[%s] find provider instance[%s/%s]",
			in.ConsumerServiceId, in.ProviderServiceId, in.ProviderInstanceId)
		return &pb.GetOneInstanceResponse{
			Response: proto.CreateResponse(scerr.ErrInternal, err.Error()),
		}, err
	}
	if provider == nil {
		log.Errorf(nil, "provider does not exist, consumer[%s] find provider instance[%s/%s]",
			in.ConsumerServiceId, in.ProviderServiceId, in.ProviderInstanceId)
		return &pb.GetOneInstanceResponse{
			Response: proto.CreateResponse(scerr.ErrServiceNotExists,
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
	item, err = cache.FindInstances.GetWithProviderID(ctx, service, proto.MicroServiceToKey(domainProject, provider),
		&pb.HeartbeatSetElement{
			ServiceId: in.ProviderServiceId, InstanceId: in.ProviderInstanceId,
		}, in.Tags, rev)
	if err != nil {
		log.Errorf(err, "FindInstances.GetWithProviderID failed, %s failed", findFlag())
		return &pb.GetOneInstanceResponse{
			Response: proto.CreateResponse(scerr.ErrInternal, err.Error()),
		}, err
	}
	if item == nil || len(item.Instances) == 0 {
		mes := fmt.Errorf("%s failed, provider instance does not exist", findFlag())
		log.Errorf(mes, "FindInstances.GetWithProviderID failed")
		return &pb.GetOneInstanceResponse{
			Response: proto.CreateResponse(scerr.ErrInstanceNotExists, mes.Error()),
		}, nil
	}

	instance := item.Instances[0]
	if rev == item.Rev {
		instance = nil // for gRPC
	}
	_ = util.SetContext(ctx, util.CtxResponseRevision, item.Rev)

	return &pb.GetOneInstanceResponse{
		Response: proto.CreateResponse(proto.Response_SUCCESS, "Get instance successfully."),
		Instance: instance,
	}, nil
}

func (ds *DataSource) FindInstances(ctx context.Context, in *pb.FindInstancesRequest) (*pb.FindInstancesResponse,
	error) {
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
		err := errors.New("rev in context is not type string")
		log.Error("", err)
		return &pb.FindInstancesResponse{
			Response: proto.CreateResponse(scerr.ErrInternal, err.Error()),
		}, err
	}

	if apt.IsShared(provider) {
		return ds.findSharedServiceInstance(ctx, in, provider, rev)
	}
	return ds.findInstance(ctx, in, provider, rev)
}

func (ds *DataSource) findInstance(ctx context.Context, in *pb.FindInstancesRequest,
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
				Response: proto.CreateResponse(scerr.ErrInternal, err.Error()),
			}, err
		}
		if service == nil {
			log.Errorf(nil, "consumer does not exist, consumer[%s] find provider[%s/%s/%s/%s]",
				in.ConsumerServiceId, in.Environment, in.AppId, in.ServiceName, in.VersionRule)
			return &pb.FindInstancesResponse{
				Response: proto.CreateResponse(scerr.ErrServiceNotExists,
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
			Response: proto.CreateResponse(scerr.ErrInternal, err.Error()),
		}, err
	}
	if item == nil {
		mes := fmt.Errorf("%s failed, provider does not exist", findFlag)
		log.Errorf(mes, "FindInstancesCache.Get failed")
		return &pb.FindInstancesResponse{
			Response: proto.CreateResponse(scerr.ErrServiceNotExists, mes.Error()),
		}, nil
	}

	// add dependency queue
	if len(in.ConsumerServiceId) > 0 &&
		len(item.ServiceIds) > 0 &&
		!cache.DependencyRule.ExistVersionRule(ctx, in.ConsumerServiceId, provider) {
		provider, err = ds.reshapeProviderKey(ctx, provider, item.ServiceIds[0])
		if err != nil {
			return nil, err
		}
		if provider != nil {
			err = serviceUtil.AddServiceVersionRule(ctx, domainProject, service, provider)
		} else {
			mes := fmt.Errorf("%s failed, provider does not exist", findFlag)
			log.Errorf(mes, "AddServiceVersionRule failed")
			return &pb.FindInstancesResponse{
				Response: proto.CreateResponse(scerr.ErrServiceNotExists, mes.Error()),
			}, nil
		}
		if err != nil {
			log.Errorf(err, "AddServiceVersionRule failed, %s failed", findFlag)
			return &pb.FindInstancesResponse{
				Response: proto.CreateResponse(scerr.ErrInternal, err.Error()),
			}, err
		}
	}

	return ds.genFindResult(ctx, rev, item)
}

func (ds *DataSource) findSharedServiceInstance(ctx context.Context, in *pb.FindInstancesRequest,
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
			Response: proto.CreateResponse(scerr.ErrInternal, err.Error()),
		}, err
	}
	if item == nil {
		mes := fmt.Errorf("%s failed, provider does not exist", findFlag)
		log.Errorf(mes, "FindInstancesCache.Get failed")
		return &pb.FindInstancesResponse{
			Response: proto.CreateResponse(scerr.ErrServiceNotExists, mes.Error()),
		}, nil
	}

	return ds.genFindResult(ctx, rev, item)
}

func (ds *DataSource) genFindResult(ctx context.Context, oldRev string, item *cache.VersionRuleCacheItem) (
	*pb.FindInstancesResponse, error) {
	instances := item.Instances
	if oldRev == item.Rev {
		instances = nil // for gRPC
	}
	// TODO support gRPC output context
	_ = util.SetContext(ctx, util.CtxResponseRevision, item.Rev)
	return &pb.FindInstancesResponse{
		Response:  proto.CreateResponse(proto.Response_SUCCESS, "Query service instances successfully."),
		Instances: instances,
	}, nil
}

func (ds *DataSource) reshapeProviderKey(ctx context.Context, provider *pb.MicroServiceKey, providerID string) (
	*pb.MicroServiceKey, error) {
	//维护version的规则,service name 可能是别名，所以重新获取
	providerService, err := serviceUtil.GetService(ctx, provider.Tenant, providerID)
	if providerService == nil {
		return nil, err
	}

	versionRule := provider.Version
	provider = proto.MicroServiceToKey(provider.Tenant, providerService)
	provider.Version = versionRule
	return provider, nil
}

func (ds *DataSource) UpdateInstanceStatus(ctx context.Context, in *pb.UpdateInstanceStatusRequest) (*pb.
	UpdateInstanceStatusResponse, error) {
	domainProject := util.ParseDomainProject(ctx)
	updateStatusFlag := util.StringJoin([]string{in.ServiceId, in.InstanceId, in.Status}, "/")

	instance, err := serviceUtil.GetInstance(ctx, domainProject, in.ServiceId, in.InstanceId)
	if err != nil {
		log.Errorf(err, "update instance[%s] status failed", updateStatusFlag)
		return &pb.UpdateInstanceStatusResponse{
			Response: proto.CreateResponse(scerr.ErrInternal, err.Error()),
		}, err
	}
	if instance == nil {
		log.Errorf(nil, "update instance[%s] status failed, instance does not exist", updateStatusFlag)
		return &pb.UpdateInstanceStatusResponse{
			Response: proto.CreateResponse(scerr.ErrInstanceNotExists, "Service instance does not exist."),
		}, nil
	}

	copyInstanceRef := *instance
	copyInstanceRef.Status = in.Status

	if err := serviceUtil.UpdateInstance(ctx, domainProject, &copyInstanceRef); err != nil {
		log.Errorf(err, "update instance[%s] status failed", updateStatusFlag)
		resp := &pb.UpdateInstanceStatusResponse{
			Response: proto.CreateResponseWithSCErr(err),
		}
		if err.InternalError() {
			return resp, err
		}
		return resp, nil
	}

	log.Infof("update instance[%s] status successfully", updateStatusFlag)
	return &pb.UpdateInstanceStatusResponse{
		Response: proto.CreateResponse(proto.Response_SUCCESS, "Update service instance status successfully."),
	}, nil
}

func (ds *DataSource) UpdateInstanceProperties(ctx context.Context, in *pb.UpdateInstancePropsRequest) (
	*pb.UpdateInstancePropsResponse, error) {
	domainProject := util.ParseDomainProject(ctx)
	instanceFlag := util.StringJoin([]string{in.ServiceId, in.InstanceId}, "/")

	instance, err := serviceUtil.GetInstance(ctx, domainProject, in.ServiceId, in.InstanceId)
	if err != nil {
		log.Errorf(err, "update instance[%s] properties failed", instanceFlag)
		return &pb.UpdateInstancePropsResponse{
			Response: proto.CreateResponse(scerr.ErrInternal, err.Error()),
		}, err
	}
	if instance == nil {
		log.Errorf(nil, "update instance[%s] properties failed, instance does not exist", instanceFlag)
		return &pb.UpdateInstancePropsResponse{
			Response: proto.CreateResponse(scerr.ErrInstanceNotExists, "Service instance does not exist."),
		}, nil
	}

	copyInstanceRef := *instance
	copyInstanceRef.Properties = in.Properties

	if err := serviceUtil.UpdateInstance(ctx, domainProject, &copyInstanceRef); err != nil {
		log.Errorf(err, "update instance[%s] properties failed", instanceFlag)
		resp := &pb.UpdateInstancePropsResponse{
			Response: proto.CreateResponseWithSCErr(err),
		}
		if err.InternalError() {
			return resp, err
		}
		return resp, nil
	}

	log.Infof("update instance[%s] properties successfully", instanceFlag)
	return &pb.UpdateInstancePropsResponse{
		Response: proto.CreateResponse(proto.Response_SUCCESS, "Update service instance properties successfully."),
	}, nil
}

func (ds *DataSource) HeartbeatSet(ctx context.Context, in *pb.HeartbeatSetRequest) (*pb.HeartbeatSetResponse, error) {
	domainProject := util.ParseDomainProject(ctx)

	heartBeatCount := len(in.Instances)
	existFlag := make(map[string]bool, heartBeatCount)
	instancesHbRst := make(chan *pb.InstanceHbRst, heartBeatCount)
	noMultiCounter := 0
	for _, heartbeatElement := range in.Instances {
		if _, ok := existFlag[heartbeatElement.ServiceId+heartbeatElement.InstanceId]; ok {
			log.Warnf("instance[%s/%s] is duplicate in heartbeat set",
				heartbeatElement.ServiceId, heartbeatElement.InstanceId)
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
			Response:  proto.CreateResponse(proto.Response_SUCCESS, "Heartbeat set successfully."),
			Instances: instanceHbRstArr,
		}, nil
	}
	log.Errorf(nil, "batch update heartbeats failed, %v", in.Instances)
	return &pb.HeartbeatSetResponse{
		Response:  proto.CreateResponse(scerr.ErrInstanceNotExists, "Heartbeat set failed."),
		Instances: instanceHbRstArr,
	}, nil
}

func (ds *DataSource) GetInstances(ctx context.Context, in *pb.GetInstancesRequest) (*pb.GetInstancesResponse,
	error) {
	domainProject := util.ParseDomainProject(ctx)

	service := &pb.MicroService{}
	var err error
	if len(in.ConsumerServiceId) > 0 {
		service, err = serviceUtil.GetService(ctx, domainProject, in.ConsumerServiceId)
		if err != nil {
			log.Errorf(err, "get consumer failed, consumer[%s] find provider instances",
				in.ConsumerServiceId, in.ProviderServiceId)
			return &pb.GetInstancesResponse{
				Response: proto.CreateResponse(scerr.ErrInternal, err.Error()),
			}, err
		}
		if service == nil {
			log.Errorf(nil, "consumer does not exist, consumer[%s] find provider instances",
				in.ConsumerServiceId, in.ProviderServiceId)
			return &pb.GetInstancesResponse{
				Response: proto.CreateResponse(scerr.ErrServiceNotExists,
					fmt.Sprintf("Consumer[%s] does not exist.", in.ConsumerServiceId)),
			}, nil
		}
	}

	provider, err := serviceUtil.GetService(ctx, domainProject, in.ProviderServiceId)
	if err != nil {
		log.Errorf(err, "get provider failed, consumer[%s] find provider instances",
			in.ConsumerServiceId, in.ProviderServiceId)
		return &pb.GetInstancesResponse{
			Response: proto.CreateResponse(scerr.ErrInternal, err.Error()),
		}, err
	}
	if provider == nil {
		log.Errorf(nil, "provider does not exist, consumer[%s] find provider instances",
			in.ConsumerServiceId, in.ProviderServiceId)
		return &pb.GetInstancesResponse{
			Response: proto.CreateResponse(scerr.ErrServiceNotExists,
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
	item, err = cache.FindInstances.GetWithProviderID(ctx, service, proto.MicroServiceToKey(domainProject, provider),
		&pb.HeartbeatSetElement{
			ServiceId: in.ProviderServiceId,
		}, in.Tags, rev)
	if err != nil {
		log.Errorf(err, "FindInstances.GetWithProviderID failed, %s failed", findFlag())
		return &pb.GetInstancesResponse{
			Response: proto.CreateResponse(scerr.ErrInternal, err.Error()),
		}, err
	}
	if item == nil || len(item.ServiceIds) == 0 {
		mes := fmt.Errorf("%s failed, provider instance does not exist", findFlag())
		log.Errorf(mes, "FindInstances.GetWithProviderID failed")
		return &pb.GetInstancesResponse{
			Response: proto.CreateResponse(scerr.ErrServiceNotExists, mes.Error()),
		}, nil
	}

	instances := item.Instances
	if rev == item.Rev {
		instances = nil // for gRPC
	}
	_ = util.SetContext(ctx, util.CtxResponseRevision, item.Rev)

	return &pb.GetInstancesResponse{
		Response:  proto.CreateResponse(proto.Response_SUCCESS, "Query service instances successfully."),
		Instances: instances,
	}, nil
}

func (ds *DataSource) BatchFind(ctx context.Context, request *pb.BatchFindInstancesRequest) (
	*pb.BatchFindInstancesResponse, error) {
	response := &pb.BatchFindInstancesResponse{
		Response: proto.CreateResponse(proto.Response_SUCCESS, "Batch query service instances successfully."),
	}

	var err error
	// find services
	response.Services, err = ds.batchFindServices(ctx, request)
	if err != nil {
		return &pb.BatchFindInstancesResponse{
			Response: proto.CreateResponse(scerr.ErrInternal, err.Error()),
		}, err
	}

	// find instance
	response.Instances, err = ds.batchFindInstances(ctx, request)
	if err != nil {
		return &pb.BatchFindInstancesResponse{
			Response: proto.CreateResponse(scerr.ErrInternal, err.Error()),
		}, err
	}

	return response, nil
}

func (ds *DataSource) batchFindServices(ctx context.Context, in *pb.BatchFindInstancesRequest) (
	*pb.BatchFindResult, error) {
	if len(in.Services) == 0 {
		return nil, nil
	}
	cloneCtx := util.CloneContext(ctx)

	services := &pb.BatchFindResult{}
	failedResult := make(map[int32]*pb.FindFailedResult)
	for index, key := range in.Services {
		findCtx := util.SetContext(cloneCtx, util.CtxRequestRevision, key.Rev)
		resp, err := ds.FindInstances(findCtx, &pb.FindInstancesRequest{
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

func (ds *DataSource) batchFindInstances(ctx context.Context, in *pb.BatchFindInstancesRequest) (*pb.BatchFindResult, error) {
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
		resp, err := ds.GetInstance(getCtx, &pb.GetOneInstanceRequest{
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

func (ds *DataSource) UnregisterInstance(ctx context.Context, request *pb.UnregisterInstanceRequest) (
	*pb.UnregisterInstanceResponse, error) {
	remoteIP := util.GetIPFromContext(ctx)
	domainProject := util.ParseDomainProject(ctx)
	serviceID := request.ServiceId
	instanceID := request.InstanceId

	instanceFlag := util.StringJoin([]string{serviceID, instanceID}, "/")

	err := revokeInstance(ctx, domainProject, serviceID, instanceID)
	if err != nil {
		log.Errorf(err, "unregister instance failed, instance[%s], operator %s: revoke instance failed",
			instanceFlag, remoteIP)
		resp := &pb.UnregisterInstanceResponse{
			Response: proto.CreateResponseWithSCErr(err),
		}
		if err.InternalError() {
			return resp, err
		}
		return resp, nil
	}

	log.Infof("unregister instance[%s], operator %s", instanceFlag, remoteIP)
	return &pb.UnregisterInstanceResponse{
		Response: proto.CreateResponse(proto.Response_SUCCESS, "Unregister service instance successfully."),
	}, nil
}

func (ds *DataSource) Heartbeat(ctx context.Context, request *pb.HeartbeatRequest) (*pb.HeartbeatResponse, error) {
	remoteIP := util.GetIPFromContext(ctx)
	domainProject := util.ParseDomainProject(ctx)
	instanceFlag := util.StringJoin([]string{request.ServiceId, request.InstanceId}, "/")

	_, ttl, err := serviceUtil.HeartbeatUtil(ctx, domainProject, request.ServiceId, request.InstanceId)
	if err != nil {
		log.Errorf(err, "heartbeat failed, instance[%s]. operator %s",
			instanceFlag, remoteIP)
		resp := &pb.HeartbeatResponse{
			Response: proto.CreateResponseWithSCErr(err),
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
		log.Infof("heartbeat successful, renew instance[%s] ttl to %d. operator %s",
			instanceFlag, ttl, remoteIP)
	}
	return &pb.HeartbeatResponse{
		Response: proto.CreateResponse(proto.Response_SUCCESS,
			"Update service instance heartbeat successfully."),
	}, nil
}

func (ds *DataSource) ModifySchemas(ctx context.Context, request *pb.ModifySchemasRequest) (
	*pb.ModifySchemasResponse, error) {
	remoteIP := util.GetIPFromContext(ctx)
	serviceID := request.ServiceId
	domainProject := util.ParseDomainProject(ctx)

	serviceInfo, err := serviceUtil.GetService(ctx, domainProject, serviceID)
	if err != nil {
		log.Errorf(err, "modify service[%s] schemas failed, get service failed, operator: %s",
			serviceID, remoteIP)
		return &pb.ModifySchemasResponse{
			Response: proto.CreateResponse(scerr.ErrInternal, err.Error()),
		}, err
	}
	if serviceInfo == nil {
		log.Errorf(nil, "modify service[%s] schemas failed, service does not exist, operator: %s",
			serviceID, remoteIP)
		return &pb.ModifySchemasResponse{
			Response: proto.CreateResponse(scerr.ErrServiceNotExists, "Service does not exist."),
		}, nil
	}

	respErr := ds.modifySchemas(ctx, domainProject, serviceInfo, request.Schemas)
	if respErr != nil {
		log.Errorf(nil, "modify service[%s] schemas failed, operator: %s", serviceID, remoteIP)
		resp := &pb.ModifySchemasResponse{
			Response: proto.CreateResponseWithSCErr(respErr),
		}
		if respErr.InternalError() {
			return resp, respErr
		}
		return resp, nil
	}

	return &pb.ModifySchemasResponse{
		Response: proto.CreateResponse(proto.Response_SUCCESS, "modify schemas info successfully."),
	}, nil
}

func (ds *DataSource) ModifySchema(ctx context.Context, request *pb.ModifySchemaRequest) (
	*pb.ModifySchemaResponse, error) {
	remoteIP := util.GetIPFromContext(ctx)
	serviceID := request.ServiceId
	schemaID := request.SchemaId

	schema := pb.Schema{
		SchemaId: schemaID,
		Summary:  request.Summary,
		Schema:   request.Schema,
	}
	err := ds.modifySchema(ctx, serviceID, &schema)
	if err != nil {
		log.Errorf(err, "modify schema[%s/%s] failed, operator: %s", serviceID, schemaID, remoteIP)
		resp := &pb.ModifySchemaResponse{
			Response: proto.CreateResponseWithSCErr(err),
		}
		if err.InternalError() {
			return resp, err
		}
		return resp, nil
	}

	log.Infof("modify schema[%s/%s] successfully, operator: %s", serviceID, schemaID, remoteIP)
	return &pb.ModifySchemaResponse{
		Response: proto.CreateResponse(proto.Response_SUCCESS, "modify schema info success"),
	}, nil
}

func (ds *DataSource) GetSchema() {
	panic("implement me")
}

func (ds *DataSource) DeleteSchema() {
	panic("implement me")
}

func (ds *DataSource) ExistSchema(ctx context.Context, request *pb.GetExistenceRequest) (
	*pb.GetExistenceResponse, error) {
	domainProject := util.ParseDomainProject(ctx)

	if !serviceUtil.ServiceExist(ctx, domainProject, request.ServiceId) {
		log.Warnf("schema[%s/%s] exist failed, service does not exist", request.ServiceId, request.SchemaId)
		return &pb.GetExistenceResponse{
			Response: proto.CreateResponse(scerr.ErrServiceNotExists, "service does not exist."),
		}, nil
	}

	key := apt.GenerateServiceSchemaKey(domainProject, request.ServiceId, request.SchemaId)
	exist, err := checkSchemaInfoExist(ctx, key)
	if err != nil {
		log.Errorf(err, "schema[%s/%s] exist failed, get schema failed", request.ServiceId, request.SchemaId)
		return &pb.GetExistenceResponse{
			Response: proto.CreateResponse(scerr.ErrInternal, err.Error()),
		}, err
	}
	if !exist {
		log.Infof("schema[%s/%s] exist failed, schema does not exist", request.ServiceId, request.SchemaId)
		return &pb.GetExistenceResponse{
			Response: proto.CreateResponse(scerr.ErrSchemaNotExists, "schema does not exist."),
		}, nil
	}
	schemaSummary, err := getSchemaSummary(ctx, domainProject, request.ServiceId, request.SchemaId)
	if err != nil {
		log.Errorf(err, "schema[%s/%s] exist failed, get schema summary failed",
			request.ServiceId, request.SchemaId)
		return &pb.GetExistenceResponse{
			Response: proto.CreateResponse(scerr.ErrInternal, err.Error()),
		}, err
	}
	return &pb.GetExistenceResponse{
		Response: proto.CreateResponse(proto.Response_SUCCESS, "Schema exist."),
		SchemaId: request.SchemaId,
		Summary:  schemaSummary,
	}, nil
}

func (ds *DataSource) AddTags(ctx context.Context, in *pb.AddServiceTagsRequest) (*pb.AddServiceTagsResponse, error) {
	remoteIP := util.GetIPFromContext(ctx)
	domainProject := util.ParseDomainProject(ctx)

	// service id存在性校验
	if !serviceUtil.ServiceExist(ctx, domainProject, in.ServiceId) {
		log.Errorf(nil, "add service[%s]'s tags %v failed, service does not exist, operator: %s",
			in.ServiceId, in.Tags, remoteIP)
		return &pb.AddServiceTagsResponse{
			Response: proto.CreateResponse(scerr.ErrServiceNotExists, "Service does not exist."),
		}, nil
	}

	addTags := in.Tags
	res := quota.NewApplyQuotaResource(quota.TagQuotaType, domainProject, in.ServiceId, int64(len(addTags)))
	rst := plugin.Plugins().Quota().Apply4Quotas(ctx, res)
	errQuota := rst.Err
	if errQuota != nil {
		log.Errorf(errQuota, "add service[%s]'s tags %v failed, operator: %s", in.ServiceId, addTags, remoteIP)
		response := &pb.AddServiceTagsResponse{
			Response: proto.CreateResponseWithSCErr(errQuota),
		}
		if errQuota.InternalError() {
			return response, errQuota
		}
		return response, nil
	}

	dataTags, err := serviceUtil.GetTagsUtils(ctx, domainProject, in.ServiceId)
	if err != nil {
		log.Errorf(err, "add service[%s]'s tags %v failed, get existed tag failed, operator: %s",
			in.ServiceId, addTags, remoteIP)
		return &pb.AddServiceTagsResponse{
			Response: proto.CreateResponse(scerr.ErrInternal, err.Error()),
		}, err
	}
	for key, value := range dataTags {
		if _, ok := addTags[key]; ok {
			continue
		}
		addTags[key] = value
	}
	dataTags = addTags

	checkErr := serviceUtil.AddTagIntoETCD(ctx, domainProject, in.ServiceId, dataTags)
	if checkErr != nil {
		log.Errorf(checkErr, "add service[%s]'s tags %v failed, operator: %s", in.ServiceId, in.Tags, remoteIP)
		resp := &pb.AddServiceTagsResponse{
			Response: proto.CreateResponseWithSCErr(checkErr),
		}
		if checkErr.InternalError() {
			return resp, checkErr
		}
		return resp, nil
	}

	log.Infof("add service[%s]'s tags %v successfully, operator: %s", in.ServiceId, in.Tags, remoteIP)
	return &pb.AddServiceTagsResponse{
		Response: proto.CreateResponse(proto.Response_SUCCESS, "Add service tags successfully."),
	}, nil
}

func (ds *DataSource) GetTag() {
	panic("implement me")
}

func (ds *DataSource) UpdateTag() {
	panic("implement me")
}

func (ds *DataSource) DeleteTag() {
	panic("implement me")
}

func (ds *DataSource) AddRule() {
	panic("implement me")
}

func (ds *DataSource) GetRule() {
	panic("implement me")
}

func (ds *DataSource) UpdateRule() {
	panic("implement me")
}

func (ds *DataSource) DeleteRule() {
	panic("implement me")
}

func (ds *DataSource) modifySchemas(ctx context.Context, domainProject string, service *pb.MicroService,
	schemas []*pb.Schema) *scerr.Error {
	remoteIP := util.GetIPFromContext(ctx)
	serviceID := service.ServiceId
	schemasFromDatabase, err := getSchemasFromDatabase(ctx, domainProject, serviceID)
	if err != nil {
		log.Errorf(nil, "modify service[%s] schemas failed, get schemas failed, operator: %s",
			serviceID, remoteIP)
		return scerr.NewError(scerr.ErrUnavailableBackend, err.Error())
	}

	needUpdateSchemas, needAddSchemas, needDeleteSchemas, nonExistSchemaIds :=
		schemasAnalysis(schemas, schemasFromDatabase, service.Schemas)

	pluginOps := make([]registry.PluginOp, 0)
	if !ds.isSchemaEditable(service) {
		if len(service.Schemas) == 0 {
			res := quota.NewApplyQuotaResource(quota.SchemaQuotaType, domainProject, serviceID, int64(len(nonExistSchemaIds)))
			rst := plugin.Plugins().Quota().Apply4Quotas(ctx, res)
			errQuota := rst.Err
			if errQuota != nil {
				log.Errorf(errQuota, "modify service[%s] schemas failed, operator: %s", serviceID, remoteIP)
				return errQuota
			}

			service.Schemas = nonExistSchemaIds
			opt, err := serviceUtil.UpdateService(domainProject, serviceID, service)
			if err != nil {
				log.Errorf(err, "modify service[%s] schemas failed, update service.Schemas failed, operator: %s",
					serviceID, remoteIP)
				return scerr.NewError(scerr.ErrInternal, err.Error())
			}
			pluginOps = append(pluginOps, opt)
		} else {
			if len(nonExistSchemaIds) != 0 {
				errInfo := fmt.Errorf("non-existent schemaIDs %v", nonExistSchemaIds)
				log.Errorf(errInfo, "modify service[%s] schemas failed, operator: %s", serviceID, remoteIP)
				return scerr.NewError(scerr.ErrUndefinedSchemaID, errInfo.Error())
			}
			for _, needUpdateSchema := range needUpdateSchemas {
				exist, err := isExistSchemaSummary(ctx, domainProject, serviceID, needUpdateSchema.SchemaId)
				if err != nil {
					return scerr.NewError(scerr.ErrInternal, err.Error())
				}
				if !exist {
					opts := schemaWithDatabaseOpera(registry.OpPut, domainProject, serviceID, needUpdateSchema)
					pluginOps = append(pluginOps, opts...)
				} else {
					log.Warnf("schema[%s/%s] and it's summary already exist, skip to update, operator: %s",
						serviceID, needUpdateSchema.SchemaId, remoteIP)
				}
			}
		}

		for _, schema := range needAddSchemas {
			log.Infof("add new schema[%s/%s], operator: %s", serviceID, schema.SchemaId, remoteIP)
			opts := schemaWithDatabaseOpera(registry.OpPut, domainProject, service.ServiceId, schema)
			pluginOps = append(pluginOps, opts...)
		}
	} else {
		quotaSize := len(needAddSchemas) - len(needDeleteSchemas)
		if quotaSize > 0 {
			res := quota.NewApplyQuotaResource(quota.SchemaQuotaType, domainProject, serviceID, int64(quotaSize))
			rst := plugin.Plugins().Quota().Apply4Quotas(ctx, res)
			err := rst.Err
			if err != nil {
				log.Errorf(err, "modify service[%s] schemas failed, operator: %s", serviceID, remoteIP)
				return err
			}
		}

		var schemaIDs []string
		for _, schema := range needAddSchemas {
			log.Infof("add new schema[%s/%s], operator: %s", serviceID, schema.SchemaId, remoteIP)
			opts := schemaWithDatabaseOpera(registry.OpPut, domainProject, service.ServiceId, schema)
			pluginOps = append(pluginOps, opts...)
			schemaIDs = append(schemaIDs, schema.SchemaId)
		}

		for _, schema := range needUpdateSchemas {
			log.Infof("update schema[%s/%s], operator: %s", serviceID, schema.SchemaId, remoteIP)
			opts := schemaWithDatabaseOpera(registry.OpPut, domainProject, serviceID, schema)
			pluginOps = append(pluginOps, opts...)
			schemaIDs = append(schemaIDs, schema.SchemaId)
		}

		for _, schema := range needDeleteSchemas {
			log.Infof("delete non-existent schema[%s/%s], operator: %s", serviceID, schema.SchemaId, remoteIP)
			opts := schemaWithDatabaseOpera(registry.OpDel, domainProject, serviceID, schema)
			pluginOps = append(pluginOps, opts...)
		}

		service.Schemas = schemaIDs
		opt, err := serviceUtil.UpdateService(domainProject, serviceID, service)
		if err != nil {
			log.Errorf(err, "modify service[%s] schemas failed, update service.Schemas failed, operator: %s",
				serviceID, remoteIP)
			return scerr.NewError(scerr.ErrInternal, err.Error())
		}
		pluginOps = append(pluginOps, opt)
	}

	if len(pluginOps) != 0 {
		resp, err := backend.BatchCommitWithCmp(ctx, pluginOps,
			[]registry.CompareOp{registry.OpCmp(
				registry.CmpVer(util.StringToBytesWithNoCopy(apt.GenerateServiceKey(domainProject, serviceID))),
				registry.CmpNotEqual, 0)},
			nil)
		if err != nil {
			return scerr.NewError(scerr.ErrUnavailableBackend, err.Error())
		}
		if !resp.Succeeded {
			return scerr.NewError(scerr.ErrServiceNotExists, "Service does not exist.")
		}
	}
	return nil
}

func (ds *DataSource) isSchemaEditable(service *pb.MicroService) bool {
	return (len(service.Environment) != 0 && service.Environment != pb.ENV_PROD) || ds.SchemaEditable
}

func (ds *DataSource) modifySchema(ctx context.Context, serviceID string, schema *pb.Schema) *scerr.Error {
	remoteIP := util.GetIPFromContext(ctx)
	domainProject := util.ParseDomainProject(ctx)
	schemaID := schema.SchemaId

	microService, err := serviceUtil.GetService(ctx, domainProject, serviceID)
	if err != nil {
		log.Errorf(err, "modify schema[%s/%s] failed, get `microService failed, operator: %s",
			serviceID, schemaID, remoteIP)
		return scerr.NewError(scerr.ErrInternal, err.Error())
	}
	if microService == nil {
		log.Errorf(nil, "modify schema[%s/%s] failed, microService does not exist, operator: %s",
			serviceID, schemaID, remoteIP)
		return scerr.NewError(scerr.ErrServiceNotExists, "Service does not exist")
	}

	var pluginOps []registry.PluginOp
	isExist := isExistSchemaID(microService, []*pb.Schema{schema})

	if !ds.isSchemaEditable(microService) {
		if len(microService.Schemas) != 0 && !isExist {
			return scerr.NewError(scerr.ErrUndefinedSchemaID, "Non-existent schemaID can't be added in "+pb.ENV_PROD)
		}

		key := apt.GenerateServiceSchemaKey(domainProject, serviceID, schemaID)
		respSchema, err := backend.Store().Schema().Search(ctx, registry.WithStrKey(key), registry.WithCountOnly())
		if err != nil {
			log.Errorf(err, "modify schema[%s/%s] failed, get schema summary failed, operator: %s",
				serviceID, schemaID, remoteIP)
			return scerr.NewError(scerr.ErrUnavailableBackend, err.Error())
		}

		if respSchema.Count != 0 {
			if len(schema.Summary) == 0 {
				log.Errorf(err, "%s mode, schema[%s/%s] already exists, can not be changed, operator: %s",
					pb.ENV_PROD, serviceID, schemaID, remoteIP)
				return scerr.NewError(scerr.ErrModifySchemaNotAllow,
					"schema already exist, can not be changed in "+pb.ENV_PROD)
			}

			exist, err := isExistSchemaSummary(ctx, domainProject, serviceID, schemaID)
			if err != nil {
				log.Errorf(err, "check schema[%s/%s] summary existence failed, operator: %s",
					serviceID, schemaID, remoteIP)
				return scerr.NewError(scerr.ErrInternal, err.Error())
			}
			if exist {
				log.Errorf(err, "%s mode, schema[%s/%s] already exist, can not be changed, operator: %s",
					pb.ENV_PROD, serviceID, schemaID, remoteIP)
				return scerr.NewError(scerr.ErrModifySchemaNotAllow, "schema already exist, can not be changed in "+pb.ENV_PROD)
			}
		}

		if len(microService.Schemas) == 0 {
			microService.Schemas = append(microService.Schemas, schemaID)
			opt, err := serviceUtil.UpdateService(domainProject, serviceID, microService)
			if err != nil {
				log.Errorf(err, "modify schema[%s/%s] failed, update microService.Schemas failed, operator: %s",
					serviceID, schemaID, remoteIP)
				return scerr.NewError(scerr.ErrInternal, err.Error())
			}
			pluginOps = append(pluginOps, opt)
		}
	} else {
		if !isExist {
			microService.Schemas = append(microService.Schemas, schemaID)
			opt, err := serviceUtil.UpdateService(domainProject, serviceID, microService)
			if err != nil {
				log.Errorf(err, "modify schema[%s/%s] failed, update microService.Schemas failed, operator: %s",
					serviceID, schemaID, remoteIP)
				return scerr.NewError(scerr.ErrInternal, err.Error())
			}
			pluginOps = append(pluginOps, opt)
		}
	}

	opts := commitSchemaInfo(domainProject, serviceID, schema)
	pluginOps = append(pluginOps, opts...)

	resp, err := backend.Registry().TxnWithCmp(ctx, pluginOps,
		[]registry.CompareOp{registry.OpCmp(
			registry.CmpVer(util.StringToBytesWithNoCopy(apt.GenerateServiceKey(domainProject, serviceID))),
			registry.CmpNotEqual, 0)},
		nil)
	if err != nil {
		return scerr.NewError(scerr.ErrUnavailableBackend, err.Error())
	}
	if !resp.Succeeded {
		return scerr.NewError(scerr.ErrServiceNotExists, "Service does not exist.")
	}
	return nil
}

func (ds *DataSource) DeleteServicePri(ctx context.Context, serviceID string, force bool) (*pb.Response, error) {
	remoteIP := util.GetIPFromContext(ctx)
	domainProject := util.ParseDomainProject(ctx)

	title := "delete"
	if force {
		title = "force delete"
	}

	if serviceID == apt.Service.ServiceId {
		err := errors.New("not allow to delete service center")
		log.Errorf(err, "%s micro-service[%s] failed, operator: %s", title, serviceID, remoteIP)
		return proto.CreateResponse(scerr.ErrInvalidParams, err.Error()), nil
	}

	microservice, err := serviceUtil.GetService(ctx, domainProject, serviceID)
	if err != nil {
		log.Errorf(err, "%s micro-service[%s] failed, get service file failed, operator: %s",
			title, serviceID, remoteIP)
		return proto.CreateResponse(scerr.ErrInternal, err.Error()), err
	}

	if microservice == nil {
		log.Errorf(err, "%s micro-service[%s] failed, service does not exist, operator: %s",
			title, serviceID, remoteIP)
		return proto.CreateResponse(scerr.ErrServiceNotExists, "Service does not exist."), nil
	}

	// 强制删除，则与该服务相关的信息删除，非强制删除： 如果作为该被依赖（作为provider，提供服务,且不是只存在自依赖）或者存在实例，则不能删除
	if !force {
		dr := serviceUtil.NewProviderDependencyRelation(ctx, domainProject, microservice)
		services, err := dr.GetDependencyConsumerIds()
		if err != nil {
			log.Errorf(err, "delete micro-service[%s] failed, get service dependency failed, operator: %s",
				serviceID, remoteIP)
			return proto.CreateResponse(scerr.ErrInternal, err.Error()), err
		}
		if l := len(services); l > 1 || (l == 1 && services[0] != serviceID) {
			log.Errorf(nil, "delete micro-service[%s] failed, other services[%d] depend on it, operator: %s",
				serviceID, l, remoteIP)
			return proto.CreateResponse(scerr.ErrDependedOnConsumer, "Can not delete this service, other service rely it."), err
		}

		instancesKey := apt.GenerateInstanceKey(domainProject, serviceID, "")
		rsp, err := backend.Store().Instance().Search(ctx,
			registry.WithStrKey(instancesKey),
			registry.WithPrefix(),
			registry.WithCountOnly())
		if err != nil {
			log.Errorf(err, "delete micro-service[%s] failed, get instances failed, operator: %s",
				serviceID, remoteIP)
			return proto.CreateResponse(scerr.ErrUnavailableBackend, err.Error()), err
		}

		if rsp.Count > 0 {
			log.Errorf(nil, "delete micro-service[%s] failed, service deployed instances[%s], operator: %s",
				serviceID, rsp.Count, remoteIP)
			return proto.CreateResponse(scerr.ErrDeployedInstance, "Can not delete the service deployed instance(s)."), err
		}
	}

	serviceIDKey := apt.GenerateServiceKey(domainProject, serviceID)
	serviceKey := &pb.MicroServiceKey{
		Tenant:      domainProject,
		Environment: microservice.Environment,
		AppId:       microservice.AppId,
		ServiceName: microservice.ServiceName,
		Version:     microservice.Version,
		Alias:       microservice.Alias,
	}
	opts := []registry.PluginOp{
		registry.OpDel(registry.WithStrKey(apt.GenerateServiceIndexKey(serviceKey))),
		registry.OpDel(registry.WithStrKey(apt.GenerateServiceAliasKey(serviceKey))),
		registry.OpDel(registry.WithStrKey(serviceIDKey)),
	}

	//删除依赖规则
	optDeleteDep, err := serviceUtil.DeleteDependencyForDeleteService(domainProject, serviceID, serviceKey)
	if err != nil {
		log.Errorf(err, "%s micro-service[%s] failed, delete dependency failed, operator: %s",
			title, serviceID, remoteIP)
		return proto.CreateResponse(scerr.ErrInternal, err.Error()), err
	}
	opts = append(opts, optDeleteDep)

	//删除黑白名单
	opts = append(opts, registry.OpDel(
		registry.WithStrKey(apt.GenerateServiceRuleKey(domainProject, serviceID, "")),
		registry.WithPrefix()))
	opts = append(opts, registry.OpDel(registry.WithStrKey(
		util.StringJoin([]string{apt.GetServiceRuleIndexRootKey(domainProject), serviceID, ""}, "/")),
		registry.WithPrefix()))

	//删除schemas
	opts = append(opts, registry.OpDel(
		registry.WithStrKey(apt.GenerateServiceSchemaKey(domainProject, serviceID, "")),
		registry.WithPrefix()))
	opts = append(opts, registry.OpDel(
		registry.WithStrKey(apt.GenerateServiceSchemaSummaryKey(domainProject, serviceID, "")),
		registry.WithPrefix()))

	//删除tags
	opts = append(opts, registry.OpDel(
		registry.WithStrKey(apt.GenerateServiceTagKey(domainProject, serviceID))))

	//删除instances
	opts = append(opts, registry.OpDel(
		registry.WithStrKey(apt.GenerateInstanceKey(domainProject, serviceID, "")),
		registry.WithPrefix()))
	opts = append(opts, registry.OpDel(
		registry.WithStrKey(apt.GenerateInstanceLeaseKey(domainProject, serviceID, "")),
		registry.WithPrefix()))

	//删除实例
	err = serviceUtil.DeleteServiceAllInstances(ctx, serviceID)
	if err != nil {
		log.Errorf(err, "%s micro-service[%s] failed, revoke all instances failed, operator: %s",
			title, serviceID, remoteIP)
		return proto.CreateResponse(scerr.ErrUnavailableBackend, err.Error()), err
	}

	resp, err := backend.Registry().TxnWithCmp(ctx, opts,
		[]registry.CompareOp{registry.OpCmp(
			registry.CmpVer(util.StringToBytesWithNoCopy(serviceIDKey)),
			registry.CmpNotEqual, 0)},
		nil)
	if err != nil {
		log.Errorf(err, "%s micro-service[%s] failed, operator: %s", title, serviceID, remoteIP)
		return proto.CreateResponse(scerr.ErrUnavailableBackend, err.Error()), err
	}
	if !resp.Succeeded {
		log.Errorf(err, "%s micro-service[%s] failed, service does not exist, operator: %s",
			title, serviceID, remoteIP)
		return proto.CreateResponse(scerr.ErrServiceNotExists, "Service does not exist."), nil
	}

	serviceUtil.RemandServiceQuota(ctx)

	log.Infof("%s micro-service[%s] successfully, operator: %s", title, serviceID, remoteIP)
	return proto.CreateResponse(proto.Response_SUCCESS, "Unregister service successfully."), nil
}

func (ds *DataSource) GetDeleteServiceFunc(ctx context.Context, serviceID string, force bool,
	serviceRespChan chan<- *pb.DelServicesRspInfo) func(context.Context) {
	return func(_ context.Context) {
		serviceRst := &pb.DelServicesRspInfo{
			ServiceId:  serviceID,
			ErrMessage: "",
		}
		resp, err := ds.DeleteServicePri(ctx, serviceID, force)
		if err != nil {
			serviceRst.ErrMessage = err.Error()
		} else if resp.GetCode() != proto.Response_SUCCESS {
			serviceRst.ErrMessage = resp.GetMessage()
		}

		serviceRespChan <- serviceRst
	}
}
