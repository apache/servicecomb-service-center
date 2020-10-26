/*
 * Licensed to the Apache Software Foundation (ASF) under one or more
 * contributor license agreements.  See the NOTICE file distributed with
 * this work for additional information regarding copyright ownership.
 * The ASF licenses this file to You under the Apache License, Version 2.0
 * (the "License"); you may not use this file except request compliance with
 * the License.  You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to request writing, software
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
	"github.com/apache/servicecomb-service-center/datasource/etcd/client"
	registry "github.com/apache/servicecomb-service-center/datasource/etcd/client"
	"github.com/apache/servicecomb-service-center/datasource/etcd/kv"
	"github.com/apache/servicecomb-service-center/datasource/etcd/sd"
	serviceUtil "github.com/apache/servicecomb-service-center/datasource/etcd/util"
	"github.com/apache/servicecomb-service-center/pkg/gopool"
	"github.com/apache/servicecomb-service-center/pkg/log"
	pb "github.com/apache/servicecomb-service-center/pkg/registry"
	"github.com/apache/servicecomb-service-center/pkg/util"
	apt "github.com/apache/servicecomb-service-center/server/core"
	"github.com/apache/servicecomb-service-center/server/core/proto"
	"github.com/apache/servicecomb-service-center/server/plugin/quota"
	scerr "github.com/apache/servicecomb-service-center/server/scerror"
	"github.com/apache/servicecomb-service-center/server/service/cache"
	"sort"
	"strconv"
	"time"
)

var clustersIndex = make(map[string]int)

func init() {
	var clusters []string
	for name := range Configuration().Clusters {
		clusters = append(clusters, name)
	}
	sort.Strings(clusters)
	for i, name := range clusters {
		clustersIndex[name] = i
	}
}

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
	resp, err := client.Instance().TxnWithCmp(ctx, opts, uniqueCmpOpts, failOpts)

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

func (ds *DataSource) GetServiceDetail(ctx context.Context, request *pb.GetServiceRequest) (
	*pb.GetServiceDetailResponse, error) {
	ctx = util.SetContext(ctx, util.CtxCacheOnly, "1")

	domainProject := util.ParseDomainProject(ctx)
	options := []string{"tags", "rules", "instances", "schemas", "dependencies"}

	if len(request.ServiceId) == 0 {
		return &pb.GetServiceDetailResponse{
			Response: proto.CreateResponse(scerr.ErrInvalidParams, "Invalid request for getting service detail."),
		}, nil
	}

	service, err := serviceUtil.GetService(ctx, domainProject, request.ServiceId)
	if service == nil {
		return &pb.GetServiceDetailResponse{
			Response: proto.CreateResponse(scerr.ErrServiceNotExists, "Service does not exist."),
		}, nil
	}
	if err != nil {
		return &pb.GetServiceDetailResponse{
			Response: proto.CreateResponse(scerr.ErrInternal, err.Error()),
		}, err
	}

	key := &pb.MicroServiceKey{
		Tenant:      domainProject,
		Environment: service.Environment,
		AppId:       service.AppId,
		ServiceName: service.ServiceName,
		Version:     "",
	}
	versions, err := getServiceAllVersions(ctx, key)
	if err != nil {
		log.Errorf(err, "get service[%s/%s/%s] all versions failed",
			service.Environment, service.AppId, service.ServiceName)
		return &pb.GetServiceDetailResponse{
			Response: proto.CreateResponse(scerr.ErrInternal, err.Error()),
		}, err
	}

	serviceInfo, err := getServiceDetailUtil(ctx, ServiceDetailOpt{
		domainProject: domainProject,
		service:       service,
		options:       options,
	})
	if err != nil {
		return &pb.GetServiceDetailResponse{
			Response: proto.CreateResponse(scerr.ErrInternal, err.Error()),
		}, err
	}

	serviceInfo.MicroService = service
	serviceInfo.MicroServiceVersions = versions
	return &pb.GetServiceDetailResponse{
		Response: proto.CreateResponse(proto.Response_SUCCESS, "Get service successfully."),
		Service:  serviceInfo,
	}, nil
}

func (ds *DataSource) GetServicesInfo(ctx context.Context, request *pb.GetServicesInfoRequest) (
	*pb.GetServicesInfoResponse, error) {
	ctx = util.SetContext(ctx, util.CtxCacheOnly, "1")

	optionMap := make(map[string]struct{}, len(request.Options))
	for _, opt := range request.Options {
		optionMap[opt] = struct{}{}
	}

	options := make([]string, 0, len(optionMap))
	if _, ok := optionMap["all"]; ok {
		optionMap["statistics"] = struct{}{}
		options = []string{"tags", "rules", "instances", "schemas", "dependencies"}
	} else {
		for opt := range optionMap {
			options = append(options, opt)
		}
	}

	var st *pb.Statistics
	if _, ok := optionMap["statistics"]; ok {
		var err error
		st, err = statistics(ctx, request.WithShared)
		if err != nil {
			return &pb.GetServicesInfoResponse{
				Response: proto.CreateResponse(scerr.ErrInternal, err.Error()),
			}, err
		}
		if len(optionMap) == 1 {
			return &pb.GetServicesInfoResponse{
				Response:   proto.CreateResponse(proto.Response_SUCCESS, "Statistics successfully."),
				Statistics: st,
			}, nil
		}
	}

	//获取所有服务
	services, err := serviceUtil.GetAllServiceUtil(ctx)
	if err != nil {
		log.Errorf(err, "get all services by domain failed")
		return &pb.GetServicesInfoResponse{
			Response: proto.CreateResponse(scerr.ErrInternal, err.Error()),
		}, err
	}

	allServiceDetails := make([]*pb.ServiceDetail, 0, len(services))
	domainProject := util.ParseDomainProject(ctx)
	for _, service := range services {
		if !request.WithShared && apt.IsShared(proto.MicroServiceToKey(domainProject, service)) {
			continue
		}
		if len(request.AppId) > 0 {
			if request.AppId != service.AppId {
				continue
			}
			if len(request.ServiceName) > 0 && request.ServiceName != service.ServiceName {
				continue
			}
		}

		serviceDetail, err := getServiceDetailUtil(ctx, ServiceDetailOpt{
			domainProject: domainProject,
			service:       service,
			countOnly:     request.CountOnly,
			options:       options,
		})
		if err != nil {
			return &pb.GetServicesInfoResponse{
				Response: proto.CreateResponse(scerr.ErrInternal, err.Error()),
			}, err
		}
		serviceDetail.MicroService = service
		allServiceDetails = append(allServiceDetails, serviceDetail)
	}

	return &pb.GetServicesInfoResponse{
		Response:          proto.CreateResponse(proto.Response_SUCCESS, "Get services info successfully."),
		AllServicesDetail: allServiceDetails,
		Statistics:        st,
	}, nil
}

func (ds *DataSource) GetApplications(ctx context.Context, request *pb.GetAppsRequest) (*pb.GetAppsResponse, error) {
	domainProject := util.ParseDomainProject(ctx)
	key := GetServiceAppKey(domainProject, request.Environment, "")

	opts := append(serviceUtil.FromContext(ctx),
		registry.WithStrKey(key),
		registry.WithPrefix(),
		registry.WithKeyOnly())

	resp, err := kv.Store().ServiceIndex().Search(ctx, opts...)
	if err != nil {
		return nil, err
	}
	l := len(resp.Kvs)
	if l == 0 {
		return &pb.GetAppsResponse{
			Response: proto.CreateResponse(proto.Response_SUCCESS, "Get all applications successfully."),
		}, nil
	}

	apps := make([]string, 0, l)
	appMap := make(map[string]struct{}, l)
	for _, kv := range resp.Kvs {
		key := GetInfoFromSvcIndexKV(kv.Key)
		if !request.WithShared && apt.IsShared(key) {
			continue
		}
		if _, ok := appMap[key.AppId]; ok {
			continue
		}
		appMap[key.AppId] = struct{}{}
		apps = append(apps, key.AppId)
	}

	return &pb.GetAppsResponse{
		Response: proto.CreateResponse(proto.Response_SUCCESS, "Get all applications successfully."),
		AppIds:   apps,
	}, nil
}

func (ds *DataSource) ExistServiceByID(ctx context.Context, request *pb.GetExistenceByIDRequest) (*pb.GetExistenceByIDResponse, error) {
	domainProject := util.ParseDomainProject(ctx)
	return &pb.GetExistenceByIDResponse{
		Response: proto.CreateResponse(proto.Response_SUCCESS, "Get all applications successfully."),
		Exist:    serviceUtil.ServiceExist(ctx, domainProject, request.ServiceId),
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
	resp, err := client.Instance().TxnWithCmp(ctx,
		[]client.PluginOp{client.OpPut(client.WithStrKey(key), client.WithValue(data))},
		[]client.CompareOp{client.OpCmp(
			client.CmpVer(util.StringToBytesWithNoCopy(key)),
			client.CmpNotEqual, 0)},
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
		// 1. request the scenario the instance has been removed,
		//    the cast of registration operation can be reduced.
		// 2. request the self-protection scenario, the instance is unhealthy
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
		reporter = quota.Apply(ctx, res)
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

	leaseID, err := client.Instance().LeaseGrant(ctx, ttl)
	if err != nil {
		log.Errorf(err, "grant lease failed, %s, operator: %s", instanceFlag, remoteIP)
		return &pb.RegisterInstanceResponse{
			Response: proto.CreateResponse(scerr.ErrUnavailableBackend, err.Error()),
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

func (ds *DataSource) GetInstance(ctx context.Context, request *pb.GetOneInstanceRequest) (
	*pb.GetOneInstanceResponse, error) {
	domainProject := util.ParseDomainProject(ctx)

	service := &pb.MicroService{}
	var err error
	if len(request.ConsumerServiceId) > 0 {
		service, err = serviceUtil.GetService(ctx, domainProject, request.ConsumerServiceId)
		if err != nil {
			log.Errorf(err, "get consumer failed, consumer[%s] find provider instance[%s/%s]",
				request.ConsumerServiceId, request.ProviderServiceId, request.ProviderInstanceId)
			return &pb.GetOneInstanceResponse{
				Response: proto.CreateResponse(scerr.ErrInternal, err.Error()),
			}, err
		}
		if service == nil {
			log.Errorf(nil, "consumer does not exist, consumer[%s] find provider instance[%s/%s]",
				request.ConsumerServiceId, request.ProviderServiceId, request.ProviderInstanceId)
			return &pb.GetOneInstanceResponse{
				Response: proto.CreateResponse(scerr.ErrServiceNotExists,
					fmt.Sprintf("Consumer[%s] does not exist.", request.ConsumerServiceId)),
			}, nil
		}
	}

	provider, err := serviceUtil.GetService(ctx, domainProject, request.ProviderServiceId)
	if err != nil {
		log.Errorf(err, "get provider failed, consumer[%s] find provider instance[%s/%s]",
			request.ConsumerServiceId, request.ProviderServiceId, request.ProviderInstanceId)
		return &pb.GetOneInstanceResponse{
			Response: proto.CreateResponse(scerr.ErrInternal, err.Error()),
		}, err
	}
	if provider == nil {
		log.Errorf(nil, "provider does not exist, consumer[%s] find provider instance[%s/%s]",
			request.ConsumerServiceId, request.ProviderServiceId, request.ProviderInstanceId)
		return &pb.GetOneInstanceResponse{
			Response: proto.CreateResponse(scerr.ErrServiceNotExists,
				fmt.Sprintf("Provider[%s] does not exist.", request.ProviderServiceId)),
		}, nil
	}

	findFlag := func() string {
		return fmt.Sprintf("Consumer[%s][%s/%s/%s/%s] find provider[%s][%s/%s/%s/%s] instance[%s]",
			request.ConsumerServiceId, service.Environment, service.AppId, service.ServiceName, service.Version,
			provider.ServiceId, provider.Environment, provider.AppId, provider.ServiceName, provider.Version,
			request.ProviderInstanceId)
	}

	var item *cache.VersionRuleCacheItem
	rev, _ := ctx.Value(util.CtxRequestRevision).(string)
	item, err = cache.FindInstances.GetWithProviderID(ctx, service, proto.MicroServiceToKey(domainProject, provider),
		&pb.HeartbeatSetElement{
			ServiceId: request.ProviderServiceId, InstanceId: request.ProviderInstanceId,
		}, request.Tags, rev)
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

func (ds *DataSource) GetInstances(ctx context.Context, request *pb.GetInstancesRequest) (*pb.GetInstancesResponse,
	error) {
	domainProject := util.ParseDomainProject(ctx)

	service := &pb.MicroService{}
	var err error
	if len(request.ConsumerServiceId) > 0 {
		service, err = serviceUtil.GetService(ctx, domainProject, request.ConsumerServiceId)
		if err != nil {
			log.Errorf(err, "get consumer failed, consumer[%s] find provider instances",
				request.ConsumerServiceId, request.ProviderServiceId)
			return &pb.GetInstancesResponse{
				Response: proto.CreateResponse(scerr.ErrInternal, err.Error()),
			}, err
		}
		if service == nil {
			log.Errorf(nil, "consumer does not exist, consumer[%s] find provider instances",
				request.ConsumerServiceId, request.ProviderServiceId)
			return &pb.GetInstancesResponse{
				Response: proto.CreateResponse(scerr.ErrServiceNotExists,
					fmt.Sprintf("Consumer[%s] does not exist.", request.ConsumerServiceId)),
			}, nil
		}
	}

	provider, err := serviceUtil.GetService(ctx, domainProject, request.ProviderServiceId)
	if err != nil {
		log.Errorf(err, "get provider failed, consumer[%s] find provider instances",
			request.ConsumerServiceId, request.ProviderServiceId)
		return &pb.GetInstancesResponse{
			Response: proto.CreateResponse(scerr.ErrInternal, err.Error()),
		}, err
	}
	if provider == nil {
		log.Errorf(nil, "provider does not exist, consumer[%s] find provider instances",
			request.ConsumerServiceId, request.ProviderServiceId)
		return &pb.GetInstancesResponse{
			Response: proto.CreateResponse(scerr.ErrServiceNotExists,
				fmt.Sprintf("Provider[%s] does not exist.", request.ProviderServiceId)),
		}, nil
	}

	findFlag := func() string {
		return fmt.Sprintf("Consumer[%s][%s/%s/%s/%s] find provider[%s][%s/%s/%s/%s] instances",
			request.ConsumerServiceId, service.Environment, service.AppId, service.ServiceName, service.Version,
			provider.ServiceId, provider.Environment, provider.AppId, provider.ServiceName, provider.Version)
	}

	var item *cache.VersionRuleCacheItem
	rev, _ := ctx.Value(util.CtxRequestRevision).(string)
	item, err = cache.FindInstances.GetWithProviderID(ctx, service, proto.MicroServiceToKey(domainProject, provider),
		&pb.HeartbeatSetElement{
			ServiceId: request.ProviderServiceId,
		}, request.Tags, rev)
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

func (ds *DataSource) GetProviderInstances(ctx context.Context, request *pb.HeartbeatSetElement) (instances []*pb.MicroServiceInstance, rev string, err error) {
	var (
		maxRevs       = make([]int64, len(clustersIndex))
		counts        = make([]int64, len(clustersIndex))
		domainProject = util.ParseTargetDomainProject(ctx)
	)
	instances, err = ds.findInstances(ctx, domainProject, request.ServiceId, request.InstanceId, maxRevs, counts)
	if err != nil {
		return
	}
	return instances, serviceUtil.FormatRevision(maxRevs, counts), nil
}

func (ds *DataSource) BatchGetProviderInstances(ctx context.Context, request *pb.BatchGetInstancesRequest) (instances []*pb.MicroServiceInstance, rev string, err error) {
	var (
		maxRevs       = make([]int64, len(clustersIndex))
		counts        = make([]int64, len(clustersIndex))
		domainProject = util.ParseTargetDomainProject(ctx)
	)
	if request == nil || len(request.ServiceIds) == 0 {
		return nil, "", fmt.Errorf("invalid param BatchGetInstancesRequest")
	}

	for _, providerServiceID := range request.ServiceIds {
		insts, err := ds.findInstances(ctx, domainProject, providerServiceID, "", maxRevs, counts)
		if err != nil {
			return nil, "", err
		}
		instances = append(instances, insts...)
	}

	return instances, serviceUtil.FormatRevision(maxRevs, counts), nil
}

func (ds *DataSource) findInstances(ctx context.Context, domainProject, serviceID, instanceID string, maxRevs []int64, counts []int64) (instances []*pb.MicroServiceInstance, err error) {
	key := apt.GenerateInstanceKey(domainProject, serviceID, instanceID)
	opts := append(serviceUtil.FromContext(ctx), registry.WithStrKey(key), registry.WithPrefix())
	resp, err := kv.Store().Instance().Search(ctx, opts...)
	if err != nil {
		return nil, err
	}
	if len(resp.Kvs) == 0 {
		return
	}

	for _, kv := range resp.Kvs {
		if i, ok := clustersIndex[kv.ClusterName]; ok {
			if kv.ModRevision > maxRevs[i] {
				maxRevs[i] = kv.ModRevision
			}
			counts[i]++
		}
		instances = append(instances, kv.Value.(*pb.MicroServiceInstance))
	}
	return
}

func (ds *DataSource) FindInstances(ctx context.Context, request *pb.FindInstancesRequest) (*pb.FindInstancesResponse,
	error) {
	provider := &pb.MicroServiceKey{
		Tenant:      util.ParseTargetDomainProject(ctx),
		Environment: request.Environment,
		AppId:       request.AppId,
		ServiceName: request.ServiceName,
		Alias:       request.ServiceName,
		Version:     request.VersionRule,
	}

	rev, ok := ctx.Value(util.CtxRequestRevision).(string)
	if !ok {
		err := errors.New("rev request context is not type string")
		log.Error("", err)
		return &pb.FindInstancesResponse{
			Response: proto.CreateResponse(scerr.ErrInternal, err.Error()),
		}, err
	}

	if apt.IsShared(provider) {
		return ds.findSharedServiceInstance(ctx, request, provider, rev)
	}
	return ds.findInstance(ctx, request, provider, rev)
}

func (ds *DataSource) findInstance(ctx context.Context, request *pb.FindInstancesRequest,
	provider *pb.MicroServiceKey, rev string) (*pb.FindInstancesResponse, error) {
	var err error
	domainProject := util.ParseDomainProject(ctx)
	service := &pb.MicroService{Environment: request.Environment}
	if len(request.ConsumerServiceId) > 0 {
		service, err = serviceUtil.GetService(ctx, domainProject, request.ConsumerServiceId)
		if err != nil {
			log.Errorf(err, "get consumer failed, consumer[%s] find provider[%s/%s/%s/%s]",
				request.ConsumerServiceId, request.Environment, request.AppId, request.ServiceName, request.VersionRule)
			return &pb.FindInstancesResponse{
				Response: proto.CreateResponse(scerr.ErrInternal, err.Error()),
			}, err
		}
		if service == nil {
			log.Errorf(nil, "consumer does not exist, consumer[%s] find provider[%s/%s/%s/%s]",
				request.ConsumerServiceId, request.Environment, request.AppId, request.ServiceName, request.VersionRule)
			return &pb.FindInstancesResponse{
				Response: proto.CreateResponse(scerr.ErrServiceNotExists,
					fmt.Sprintf("Consumer[%s] does not exist.", request.ConsumerServiceId)),
			}, nil
		}
		provider.Environment = service.Environment
	}

	// provider is not a shared micro-service,
	// only allow shared micro-service instances found request different domains.
	ctx = util.SetTargetDomainProject(ctx, util.ParseDomain(ctx), util.ParseProject(ctx))
	provider.Tenant = util.ParseTargetDomainProject(ctx)

	findFlag := fmt.Sprintf("Consumer[%s][%s/%s/%s/%s] find provider[%s/%s/%s/%s]",
		request.ConsumerServiceId, service.Environment, service.AppId, service.ServiceName, service.Version,
		provider.Environment, provider.AppId, provider.ServiceName, provider.Version)

	// cache
	var item *cache.VersionRuleCacheItem
	item, err = cache.FindInstances.Get(ctx, service, provider, request.Tags, rev)
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
	if len(request.ConsumerServiceId) > 0 &&
		len(item.ServiceIds) > 0 &&
		!cache.DependencyRule.ExistVersionRule(ctx, request.ConsumerServiceId, provider) {
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

func (ds *DataSource) findSharedServiceInstance(ctx context.Context, request *pb.FindInstancesRequest,
	provider *pb.MicroServiceKey, rev string) (*pb.FindInstancesResponse, error) {
	var err error
	service := &pb.MicroService{Environment: request.Environment}
	// it means the shared micro-services must be the same env with SC.
	provider.Environment = apt.Service.Environment
	findFlag := fmt.Sprintf("find shared provider[%s/%s/%s/%s]", provider.Environment, provider.AppId, provider.ServiceName, provider.Version)

	// cache
	var item *cache.VersionRuleCacheItem
	item, err = cache.FindInstances.Get(ctx, service, provider, request.Tags, rev)
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

func (ds *DataSource) UpdateInstanceStatus(ctx context.Context, request *pb.UpdateInstanceStatusRequest) (*pb.UpdateInstanceStatusResponse, error) {
	domainProject := util.ParseDomainProject(ctx)
	updateStatusFlag := util.StringJoin([]string{request.ServiceId, request.InstanceId, request.Status}, "/")

	instance, err := serviceUtil.GetInstance(ctx, domainProject, request.ServiceId, request.InstanceId)
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
	copyInstanceRef.Status = request.Status

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

func (ds *DataSource) UpdateInstanceProperties(ctx context.Context, request *pb.UpdateInstancePropsRequest) (
	*pb.UpdateInstancePropsResponse, error) {
	domainProject := util.ParseDomainProject(ctx)
	instanceFlag := util.StringJoin([]string{request.ServiceId, request.InstanceId}, "/")

	instance, err := serviceUtil.GetInstance(ctx, domainProject, request.ServiceId, request.InstanceId)
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
	copyInstanceRef.Properties = request.Properties

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

func (ds *DataSource) HeartbeatSet(ctx context.Context, request *pb.HeartbeatSetRequest) (*pb.HeartbeatSetResponse, error) {
	domainProject := util.ParseDomainProject(ctx)

	heartBeatCount := len(request.Instances)
	existFlag := make(map[string]bool, heartBeatCount)
	instancesHbRst := make(chan *pb.InstanceHbRst, heartBeatCount)
	noMultiCounter := 0
	for _, heartbeatElement := range request.Instances {
		if _, ok := existFlag[heartbeatElement.ServiceId+heartbeatElement.InstanceId]; ok {
			log.Warnf("instance[%s/%s] is duplicate request heartbeat set",
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
	log.Errorf(nil, "batch update heartbeats failed, %v", request.Instances)
	return &pb.HeartbeatSetResponse{
		Response:  proto.CreateResponse(scerr.ErrInstanceNotExists, "Heartbeat set failed."),
		Instances: instanceHbRstArr,
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

func (ds *DataSource) batchFindServices(ctx context.Context, request *pb.BatchFindInstancesRequest) (
	*pb.BatchFindResult, error) {
	if len(request.Services) == 0 {
		return nil, nil
	}
	cloneCtx := util.CloneContext(ctx)

	services := &pb.BatchFindResult{}
	failedResult := make(map[int32]*pb.FindFailedResult)
	for index, key := range request.Services {
		findCtx := util.SetContext(cloneCtx, util.CtxRequestRevision, key.Rev)
		resp, err := ds.FindInstances(findCtx, &pb.FindInstancesRequest{
			ConsumerServiceId: request.ConsumerServiceId,
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

func (ds *DataSource) batchFindInstances(ctx context.Context, request *pb.BatchFindInstancesRequest) (*pb.BatchFindResult, error) {
	if len(request.Instances) == 0 {
		return nil, nil
	}
	cloneCtx := util.CloneContext(ctx)
	// can not find the shared provider instances
	cloneCtx = util.SetTargetDomainProject(cloneCtx, util.ParseDomain(ctx), util.ParseProject(ctx))

	instances := &pb.BatchFindResult{}
	failedResult := make(map[int32]*pb.FindFailedResult)
	for index, key := range request.Instances {
		getCtx := util.SetContext(cloneCtx, util.CtxRequestRevision, key.Rev)
		resp, err := ds.GetInstance(getCtx, &pb.GetOneInstanceRequest{
			ConsumerServiceId:  request.ConsumerServiceId,
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

func (ds *DataSource) GetSchema(ctx context.Context, request *pb.GetSchemaRequest) (*pb.GetSchemaResponse, error) {
	domainProject := util.ParseDomainProject(ctx)

	if !serviceUtil.ServiceExist(ctx, domainProject, request.ServiceId) {
		log.Errorf(nil, "get schema[%s/%s] failed, service does not exist",
			request.ServiceId, request.SchemaId)
		return &pb.GetSchemaResponse{
			Response: proto.CreateResponse(scerr.ErrServiceNotExists, "Service does not exist."),
		}, nil
	}

	key := apt.GenerateServiceSchemaKey(domainProject, request.ServiceId, request.SchemaId)
	opts := append(serviceUtil.FromContext(ctx), client.WithStrKey(key))
	resp, errDo := kv.Store().Schema().Search(ctx, opts...)
	if errDo != nil {
		log.Errorf(errDo, "get schema[%s/%s] failed", request.ServiceId, request.SchemaId)
		return &pb.GetSchemaResponse{
			Response: proto.CreateResponse(scerr.ErrUnavailableBackend, errDo.Error()),
		}, errDo
	}
	if resp.Count == 0 {
		log.Errorf(errDo, "get schema[%s/%s] failed, schema does not exists",
			request.ServiceId, request.SchemaId)
		return &pb.GetSchemaResponse{
			Response: proto.CreateResponse(scerr.ErrSchemaNotExists, "Do not have this schema info."),
		}, nil
	}

	schemaSummary, err := getSchemaSummary(ctx, domainProject, request.ServiceId, request.SchemaId)
	if err != nil {
		log.Errorf(err, "get schema[%s/%s] failed, get schema summary failed",
			request.ServiceId, request.SchemaId)
		return &pb.GetSchemaResponse{
			Response: proto.CreateResponse(scerr.ErrInternal, err.Error()),
		}, err
	}

	return &pb.GetSchemaResponse{
		Response:      proto.CreateResponse(proto.Response_SUCCESS, "Get schema info successfully."),
		Schema:        util.BytesToStringWithNoCopy(resp.Kvs[0].Value.([]byte)),
		SchemaSummary: schemaSummary,
	}, nil
}

func (ds *DataSource) GetAllSchemas(ctx context.Context, request *pb.GetAllSchemaRequest) (
	*pb.GetAllSchemaResponse, error) {
	domainProject := util.ParseDomainProject(ctx)

	service, err := serviceUtil.GetService(ctx, domainProject, request.ServiceId)
	if err != nil {
		log.Errorf(err, "get service[%s] all schemas failed, get service failed", request.ServiceId)
		return &pb.GetAllSchemaResponse{
			Response: proto.CreateResponse(scerr.ErrInternal, err.Error()),
		}, err
	}
	if service == nil {
		log.Errorf(nil, "get service[%s] all schemas failed, service does not exist", request.ServiceId)
		return &pb.GetAllSchemaResponse{
			Response: proto.CreateResponse(scerr.ErrServiceNotExists, "Service does not exist."),
		}, nil
	}

	schemasList := service.Schemas
	if len(schemasList) == 0 {
		return &pb.GetAllSchemaResponse{
			Response: proto.CreateResponse(proto.Response_SUCCESS, "Do not have this schema info."),
			Schemas:  []*pb.Schema{},
		}, nil
	}

	key := apt.GenerateServiceSchemaSummaryKey(domainProject, request.ServiceId, "")
	opts := append(serviceUtil.FromContext(ctx), client.WithStrKey(key), client.WithPrefix())
	resp, errDo := kv.Store().SchemaSummary().Search(ctx, opts...)
	if errDo != nil {
		log.Errorf(errDo, "get service[%s] all schema summaries failed", request.ServiceId)
		return &pb.GetAllSchemaResponse{
			Response: proto.CreateResponse(scerr.ErrUnavailableBackend, errDo.Error()),
		}, errDo
	}

	respWithSchema := &sd.Response{}
	if request.WithSchema {
		key := apt.GenerateServiceSchemaKey(domainProject, request.ServiceId, "")
		opts := append(serviceUtil.FromContext(ctx), client.WithStrKey(key), client.WithPrefix())
		respWithSchema, errDo = kv.Store().Schema().Search(ctx, opts...)
		if errDo != nil {
			log.Errorf(errDo, "get service[%s] all schemas failed", request.ServiceId)
			return &pb.GetAllSchemaResponse{
				Response: proto.CreateResponse(scerr.ErrUnavailableBackend, errDo.Error()),
			}, errDo
		}
	}

	schemas := make([]*pb.Schema, 0, len(schemasList))
	for _, schemaID := range schemasList {
		tempSchema := &pb.Schema{}
		tempSchema.SchemaId = schemaID
		for _, summarySchema := range resp.Kvs {
			_, _, schemaIDOfSummary := apt.GetInfoFromSchemaSummaryKV(summarySchema.Key)
			if schemaID == schemaIDOfSummary {
				tempSchema.Summary = summarySchema.Value.(string)
			}
		}

		for _, contentSchema := range respWithSchema.Kvs {
			_, _, schemaIDOfSchema := apt.GetInfoFromSchemaKV(contentSchema.Key)
			if schemaID == schemaIDOfSchema {
				tempSchema.Schema = util.BytesToStringWithNoCopy(contentSchema.Value.([]byte))
			}
		}
		schemas = append(schemas, tempSchema)
	}

	return &pb.GetAllSchemaResponse{
		Response: proto.CreateResponse(proto.Response_SUCCESS, "Get all schema info successfully."),
		Schemas:  schemas,
	}, nil
}

func (ds *DataSource) DeleteSchema(ctx context.Context, request *pb.DeleteSchemaRequest) (
	*pb.DeleteSchemaResponse, error) {
	remoteIP := util.GetIPFromContext(ctx)
	domainProject := util.ParseDomainProject(ctx)

	if !serviceUtil.ServiceExist(ctx, domainProject, request.ServiceId) {
		log.Errorf(nil, "delete schema[%s/%s] failed, service does not exist, operator: %s",
			request.ServiceId, request.SchemaId, remoteIP)
		return &pb.DeleteSchemaResponse{
			Response: proto.CreateResponse(scerr.ErrServiceNotExists, "Service does not exist."),
		}, nil
	}

	key := apt.GenerateServiceSchemaKey(domainProject, request.ServiceId, request.SchemaId)
	exist, err := serviceUtil.CheckSchemaInfoExist(ctx, key)
	if err != nil {
		log.Errorf(err, "delete schema[%s/%s] failed, operator: %s",
			request.ServiceId, request.SchemaId, remoteIP)
		return &pb.DeleteSchemaResponse{
			Response: proto.CreateResponse(scerr.ErrInternal, err.Error()),
		}, err
	}
	if !exist {
		log.Errorf(nil, "delete schema[%s/%s] failed, schema does not exist, operator: %s",
			request.ServiceId, request.SchemaId, remoteIP)
		return &pb.DeleteSchemaResponse{
			Response: proto.CreateResponse(scerr.ErrSchemaNotExists, "Schema info does not exist."),
		}, nil
	}
	epSummaryKey := apt.GenerateServiceSchemaSummaryKey(domainProject, request.ServiceId, request.SchemaId)
	opts := []client.PluginOp{
		client.OpDel(client.WithStrKey(epSummaryKey)),
		client.OpDel(client.WithStrKey(key)),
	}

	resp, errDo := client.Instance().TxnWithCmp(ctx, opts,
		[]client.CompareOp{client.OpCmp(
			client.CmpVer(util.StringToBytesWithNoCopy(apt.GenerateServiceKey(domainProject, request.ServiceId))),
			client.CmpNotEqual, 0)},
		nil)
	if errDo != nil {
		log.Errorf(errDo, "delete schema[%s/%s] failed, operator: %s",
			request.ServiceId, request.SchemaId, remoteIP)
		return &pb.DeleteSchemaResponse{
			Response: proto.CreateResponse(scerr.ErrUnavailableBackend, errDo.Error()),
		}, errDo
	}
	if !resp.Succeeded {
		log.Errorf(nil, "delete schema[%s/%s] failed, service does not exist, operator: %s",
			request.ServiceId, request.SchemaId, remoteIP)
		return &pb.DeleteSchemaResponse{
			Response: proto.CreateResponse(scerr.ErrServiceNotExists, "Service does not exist."),
		}, nil
	}

	log.Infof("delete schema[%s/%s] info successfully, operator: %s",
		request.ServiceId, request.SchemaId, remoteIP)
	return &pb.DeleteSchemaResponse{
		Response: proto.CreateResponse(proto.Response_SUCCESS, "Delete schema info successfully."),
	}, nil
}

func (ds *DataSource) AddTags(ctx context.Context, request *pb.AddServiceTagsRequest) (*pb.AddServiceTagsResponse, error) {
	remoteIP := util.GetIPFromContext(ctx)
	domainProject := util.ParseDomainProject(ctx)

	// service id存在性校验
	if !serviceUtil.ServiceExist(ctx, domainProject, request.ServiceId) {
		log.Errorf(nil, "add service[%s]'s tags %v failed, service does not exist, operator: %s",
			request.ServiceId, request.Tags, remoteIP)
		return &pb.AddServiceTagsResponse{
			Response: proto.CreateResponse(scerr.ErrServiceNotExists, "Service does not exist."),
		}, nil
	}

	addTags := request.Tags
	res := quota.NewApplyQuotaResource(quota.TagQuotaType, domainProject, request.ServiceId, int64(len(addTags)))
	rst := quota.Apply(ctx, res)
	errQuota := rst.Err
	if errQuota != nil {
		log.Errorf(errQuota, "add service[%s]'s tags %v failed, operator: %s", request.ServiceId, addTags, remoteIP)
		response := &pb.AddServiceTagsResponse{
			Response: proto.CreateResponseWithSCErr(errQuota),
		}
		if errQuota.InternalError() {
			return response, errQuota
		}
		return response, nil
	}

	dataTags, err := serviceUtil.GetTagsUtils(ctx, domainProject, request.ServiceId)
	if err != nil {
		log.Errorf(err, "add service[%s]'s tags %v failed, get existed tag failed, operator: %s",
			request.ServiceId, addTags, remoteIP)
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

	checkErr := serviceUtil.AddTagIntoETCD(ctx, domainProject, request.ServiceId, dataTags)
	if checkErr != nil {
		log.Errorf(checkErr, "add service[%s]'s tags %v failed, operator: %s", request.ServiceId, request.Tags, remoteIP)
		resp := &pb.AddServiceTagsResponse{
			Response: proto.CreateResponseWithSCErr(checkErr),
		}
		if checkErr.InternalError() {
			return resp, checkErr
		}
		return resp, nil
	}

	log.Infof("add service[%s]'s tags %v successfully, operator: %s", request.ServiceId, request.Tags, remoteIP)
	return &pb.AddServiceTagsResponse{
		Response: proto.CreateResponse(proto.Response_SUCCESS, "Add service tags successfully."),
	}, nil
}

func (ds *DataSource) GetTags(ctx context.Context, request *pb.GetServiceTagsRequest) (*pb.GetServiceTagsResponse, error) {
	var err error
	domainProject := util.ParseDomainProject(ctx)
	if !serviceUtil.ServiceExist(ctx, domainProject, request.ServiceId) {
		log.Errorf(err, "get service[%s]'s tags failed, service does not exist", request.ServiceId)
		return &pb.GetServiceTagsResponse{
			Response: proto.CreateResponse(scerr.ErrServiceNotExists, "Service does not exist."),
		}, nil
	}
	tags, err := serviceUtil.GetTagsUtils(ctx, domainProject, request.ServiceId)
	if err != nil {
		log.Errorf(err, "get service[%s]'s tags failed, get tags failed", request.ServiceId)
		return &pb.GetServiceTagsResponse{
			Response: proto.CreateResponse(scerr.ErrInternal, err.Error()),
		}, err
	}

	return &pb.GetServiceTagsResponse{
		Response: proto.CreateResponse(proto.Response_SUCCESS, "Get service tags successfully."),
		Tags:     tags,
	}, nil
}

func (ds *DataSource) UpdateTag(ctx context.Context, request *pb.UpdateServiceTagRequest) (*pb.UpdateServiceTagResponse, error) {
	var err error
	remoteIP := util.GetIPFromContext(ctx)
	tagFlag := util.StringJoin([]string{request.Key, request.Value}, "/")
	domainProject := util.ParseDomainProject(ctx)

	if !serviceUtil.ServiceExist(ctx, domainProject, request.ServiceId) {
		log.Errorf(err, "update service[%s]'s tag[%s] failed, service does not exist, operator: %s",
			request.ServiceId, tagFlag, remoteIP)
		return &pb.UpdateServiceTagResponse{
			Response: proto.CreateResponse(scerr.ErrServiceNotExists, "Service does not exist."),
		}, nil
	}

	tags, err := serviceUtil.GetTagsUtils(ctx, domainProject, request.ServiceId)
	if err != nil {
		log.Errorf(err, "update service[%s]'s tag[%s] failed, get tag failed, operator: %s",
			request.ServiceId, tagFlag, remoteIP)
		return &pb.UpdateServiceTagResponse{
			Response: proto.CreateResponse(scerr.ErrInternal, err.Error()),
		}, err
	}

	//check if the tag exists
	if _, ok := tags[request.Key]; !ok {
		log.Errorf(nil, "update service[%s]'s tag[%s] failed, tag does not exist, operator: %s",
			request.ServiceId, tagFlag, remoteIP)
		return &pb.UpdateServiceTagResponse{
			Response: proto.CreateResponse(scerr.ErrTagNotExists, "Tag does not exist, please add one first."),
		}, nil
	}

	copyTags := make(map[string]string, len(tags))
	for k, v := range tags {
		copyTags[k] = v
	}
	copyTags[request.Key] = request.Value

	checkErr := serviceUtil.AddTagIntoETCD(ctx, domainProject, request.ServiceId, copyTags)
	if checkErr != nil {
		log.Errorf(checkErr, "update service[%s]'s tag[%s] failed, operator: %s", request.ServiceId, tagFlag, remoteIP)
		resp := &pb.UpdateServiceTagResponse{
			Response: proto.CreateResponseWithSCErr(checkErr),
		}
		if checkErr.InternalError() {
			return resp, checkErr
		}
		return resp, nil
	}

	log.Infof("update service[%s]'s tag[%s] successfully, operator: %s", request.ServiceId, tagFlag, remoteIP)
	return &pb.UpdateServiceTagResponse{
		Response: proto.CreateResponse(proto.Response_SUCCESS, "Update service tag success."),
	}, nil
}

func (ds *DataSource) DeleteTags(ctx context.Context, request *pb.DeleteServiceTagsRequest) (*pb.DeleteServiceTagsResponse, error) {
	remoteIP := util.GetIPFromContext(ctx)
	domainProject := util.ParseDomainProject(ctx)

	if !serviceUtil.ServiceExist(ctx, domainProject, request.ServiceId) {
		log.Errorf(nil, "delete service[%s]'s tags %v failed, service does not exist, operator: %s",
			request.ServiceId, request.Keys, remoteIP)
		return &pb.DeleteServiceTagsResponse{
			Response: proto.CreateResponse(scerr.ErrServiceNotExists, "Service does not exist."),
		}, nil
	}

	tags, err := serviceUtil.GetTagsUtils(ctx, domainProject, request.ServiceId)
	if err != nil {
		log.Errorf(err, "delete service[%s]'s tags %v failed, get service tags failed, operator: %s",
			request.ServiceId, request.Keys, remoteIP)
		return &pb.DeleteServiceTagsResponse{
			Response: proto.CreateResponse(scerr.ErrInternal, err.Error()),
		}, err
	}

	copyTags := make(map[string]string, len(tags))
	for k, v := range tags {
		copyTags[k] = v
	}
	for _, key := range request.Keys {
		if _, ok := copyTags[key]; !ok {
			log.Errorf(nil, "delete service[%s]'s tags %v failed, tag[%s] does not exist, operator: %s",
				request.ServiceId, request.Keys, key, remoteIP)
			return &pb.DeleteServiceTagsResponse{
				Response: proto.CreateResponse(scerr.ErrTagNotExists, "Delete tags failed for this key "+key+" does not exist."),
			}, nil
		}
		delete(copyTags, key)
	}

	// the capacity of tags may be 0
	data, err := json.Marshal(copyTags)
	if err != nil {
		log.Errorf(err, "delete service[%s]'s tags %v failed, marshall service tags failed, operator: %s",
			request.ServiceId, request.Keys, remoteIP)
		return &pb.DeleteServiceTagsResponse{
			Response: proto.CreateResponse(scerr.ErrInternal, err.Error()),
		}, err
	}

	key := apt.GenerateServiceTagKey(domainProject, request.ServiceId)

	resp, err := client.Instance().TxnWithCmp(ctx,
		[]client.PluginOp{client.OpPut(client.WithStrKey(key), client.WithValue(data))},
		[]client.CompareOp{client.OpCmp(
			client.CmpVer(util.StringToBytesWithNoCopy(apt.GenerateServiceKey(domainProject, request.ServiceId))),
			client.CmpNotEqual, 0)},
		nil)
	if err != nil {
		log.Errorf(err, "delete service[%s]'s tags %v failed, operator: %s",
			request.ServiceId, request.Keys, remoteIP)
		return &pb.DeleteServiceTagsResponse{
			Response: proto.CreateResponse(scerr.ErrUnavailableBackend, err.Error()),
		}, err
	}
	if !resp.Succeeded {
		log.Errorf(err, "delete service[%s]'s tags %v failed, service does not exist, operator: %s",
			request.ServiceId, request.Keys, remoteIP)
		return &pb.DeleteServiceTagsResponse{
			Response: proto.CreateResponse(scerr.ErrServiceNotExists, "Service does not exist."),
		}, nil
	}

	log.Infof("delete service[%s]'s tags %v successfully, operator: %s", request.ServiceId, request.Keys, remoteIP)
	return &pb.DeleteServiceTagsResponse{
		Response: proto.CreateResponse(proto.Response_SUCCESS, "Delete service tags successfully."),
	}, nil
}

func (ds *DataSource) AddRule(ctx context.Context, request *pb.AddServiceRulesRequest) (
	*pb.AddServiceRulesResponse, error) {
	remoteIP := util.GetIPFromContext(ctx)
	domainProject := util.ParseDomainProject(ctx)

	// service id存在性校验
	if !serviceUtil.ServiceExist(ctx, domainProject, request.ServiceId) {
		log.Errorf(nil, "add service[%s] rule failed, service does not exist, operator: %s",
			request.ServiceId, remoteIP)
		return &pb.AddServiceRulesResponse{
			Response: proto.CreateResponse(scerr.ErrInvalidParams, "Service does not exist."),
		}, nil
	}
	res := quota.NewApplyQuotaResource(quota.RuleQuotaType, domainProject, request.ServiceId, int64(len(request.Rules)))
	rst := quota.Apply(ctx, res)
	errQuota := rst.Err
	if errQuota != nil {
		log.Errorf(errQuota, "add service[%s] rule failed, operator: %s", request.ServiceId, remoteIP)
		response := &pb.AddServiceRulesResponse{
			Response: proto.CreateResponseWithSCErr(errQuota),
		}
		if errQuota.InternalError() {
			return response, errQuota
		}
		return response, nil
	}

	ruleType, _, err := serviceUtil.GetServiceRuleType(ctx, domainProject, request.ServiceId)
	if err != nil {
		return &pb.AddServiceRulesResponse{
			Response: proto.CreateResponse(scerr.ErrInternal, err.Error()),
		}, err
	}
	ruleIDs := make([]string, 0, len(request.Rules))
	opts := make([]client.PluginOp, 0, 2*len(request.Rules))
	for _, rule := range request.Rules {
		//黑白名单只能存在一种，黑名单 or 白名单
		if len(ruleType) == 0 {
			ruleType = rule.RuleType
		} else if ruleType != rule.RuleType {
			log.Errorf(nil,
				"add service[%s] rule failed, can not add different RuleType at the same time, operator: %s",
				request.ServiceId, remoteIP)
			return &pb.AddServiceRulesResponse{
				Response: proto.CreateResponse(scerr.ErrBlackAndWhiteRule,
					"Service can only contain one rule type, BLACK or WHITE."),
			}, nil

		}

		//同一服务，attribute和pattern确定一个rule
		if serviceUtil.RuleExist(ctx, domainProject, request.ServiceId, rule.Attribute, rule.Pattern) {
			log.Infof("service[%s] rule[%s/%s] already exists, operator: %s",
				request.ServiceId, rule.Attribute, rule.Pattern, remoteIP)
			continue
		}

		// 产生全局rule id
		timestamp := strconv.FormatInt(time.Now().Unix(), 10)
		ruleAdd := &pb.ServiceRule{
			RuleId:       util.GenerateUUID(),
			RuleType:     rule.RuleType,
			Attribute:    rule.Attribute,
			Pattern:      rule.Pattern,
			Description:  rule.Description,
			Timestamp:    timestamp,
			ModTimestamp: timestamp,
		}

		key := apt.GenerateServiceRuleKey(domainProject, request.ServiceId, ruleAdd.RuleId)
		indexKey := apt.GenerateRuleIndexKey(domainProject, request.ServiceId, ruleAdd.Attribute, ruleAdd.Pattern)
		ruleIDs = append(ruleIDs, ruleAdd.RuleId)

		data, err := json.Marshal(ruleAdd)
		if err != nil {
			log.Errorf(err, "add service[%s] rule failed, marshal rule[%s/%s] failed, operator: %s",
				request.ServiceId, ruleAdd.Attribute, ruleAdd.Pattern, remoteIP)
			return &pb.AddServiceRulesResponse{
				Response: proto.CreateResponse(scerr.ErrInternal, err.Error()),
			}, err
		}

		opts = append(opts, client.OpPut(client.WithStrKey(key), client.WithValue(data)))
		opts = append(opts, client.OpPut(client.WithStrKey(indexKey), client.WithStrValue(ruleAdd.RuleId)))
	}
	if len(opts) <= 0 {
		log.Infof("add service[%s] rule successfully, no rules to add, operator: %s",
			request.ServiceId, remoteIP)
		return &pb.AddServiceRulesResponse{
			Response: proto.CreateResponse(proto.Response_SUCCESS, "Service rules has been added."),
		}, nil
	}

	resp, err := client.BatchCommitWithCmp(ctx, opts,
		[]client.CompareOp{client.OpCmp(
			client.CmpVer(util.StringToBytesWithNoCopy(apt.GenerateServiceKey(domainProject, request.ServiceId))),
			client.CmpNotEqual, 0)},
		nil)
	if err != nil {
		log.Errorf(err, "add service[%s] rule failed, operator: %s", request.ServiceId, remoteIP)
		return &pb.AddServiceRulesResponse{
			Response: proto.CreateResponse(scerr.ErrUnavailableBackend, err.Error()),
		}, err
	}
	if !resp.Succeeded {
		log.Errorf(nil, "add service[%s] rule failed, service does not exist, operator: %s",
			request.ServiceId, remoteIP)
		return &pb.AddServiceRulesResponse{
			Response: proto.CreateResponse(scerr.ErrServiceNotExists, "Service does not exist."),
		}, nil
	}

	log.Infof("add service[%s] rule %v successfully, operator: %s", request.ServiceId, ruleIDs, remoteIP)
	return &pb.AddServiceRulesResponse{
		Response: proto.CreateResponse(proto.Response_SUCCESS, "Add service rules successfully."),
		RuleIds:  ruleIDs,
	}, nil
}

func (ds *DataSource) GetRule(ctx context.Context, request *pb.GetServiceRulesRequest) (
	*pb.GetServiceRulesResponse, error) {
	domainProject := util.ParseDomainProject(ctx)

	// service id存在性校验
	if !serviceUtil.ServiceExist(ctx, domainProject, request.ServiceId) {
		log.Errorf(nil, "get service[%s] rule failed, service does not exist", request.ServiceId)
		return &pb.GetServiceRulesResponse{
			Response: proto.CreateResponse(scerr.ErrServiceNotExists, "Service does not exist."),
		}, nil
	}

	rules, err := serviceUtil.GetRulesUtil(ctx, domainProject, request.ServiceId)
	if err != nil {
		log.Errorf(err, "get service[%s] rule failed", request.ServiceId)
		return &pb.GetServiceRulesResponse{
			Response: proto.CreateResponse(scerr.ErrInternal, err.Error()),
		}, err
	}

	return &pb.GetServiceRulesResponse{
		Response: proto.CreateResponse(proto.Response_SUCCESS, "Get service rules successfully."),
		Rules:    rules,
	}, nil
}

func (ds *DataSource) UpdateRule(ctx context.Context, request *pb.UpdateServiceRuleRequest) (
	*pb.UpdateServiceRuleResponse, error) {
	remoteIP := util.GetIPFromContext(ctx)
	domainProject := util.ParseDomainProject(ctx)

	// service id存在性校验
	if !serviceUtil.ServiceExist(ctx, domainProject, request.ServiceId) {
		log.Errorf(nil, "update service rule[%s/%s] failed, service does not exist, operator: %s",
			request.ServiceId, request.RuleId, remoteIP)
		return &pb.UpdateServiceRuleResponse{
			Response: proto.CreateResponse(scerr.ErrServiceNotExists, "Service does not exist."),
		}, nil
	}

	//是否能改变ruleType
	ruleType, ruleNum, err := serviceUtil.GetServiceRuleType(ctx, domainProject, request.ServiceId)
	if err != nil {
		log.Errorf(err, "update service rule[%s/%s] failed, get rule type failed, operator: %s",
			request.ServiceId, request.RuleId, remoteIP)
		return &pb.UpdateServiceRuleResponse{
			Response: proto.CreateResponse(scerr.ErrInternal, err.Error()),
		}, err
	}
	if ruleNum >= 1 && ruleType != request.Rule.RuleType {
		log.Errorf(err, "update service rule[%s/%s] failed, can only exist one type, current type is %s, operator: %s",
			request.ServiceId, request.RuleId, ruleType, remoteIP)
		return &pb.UpdateServiceRuleResponse{
			Response: proto.CreateResponse(scerr.ErrModifyRuleNotAllow, "Exist multiple rules,can not change rule type. Rule type is "+ruleType),
		}, nil
	}

	rule, err := serviceUtil.GetOneRule(ctx, domainProject, request.ServiceId, request.RuleId)
	if err != nil {
		log.Errorf(err, "update service rule[%s/%s] failed, query service rule failed, operator: %s",
			request.ServiceId, request.RuleId, remoteIP)
		return &pb.UpdateServiceRuleResponse{
			Response: proto.CreateResponse(scerr.ErrInternal, err.Error()),
		}, err
	}
	if rule == nil {
		log.Errorf(err, "update service rule[%s/%s] failed, service rule does not exist, operator: %s",
			request.ServiceId, request.RuleId, remoteIP)
		return &pb.UpdateServiceRuleResponse{
			Response: proto.CreateResponse(scerr.ErrRuleNotExists, "This rule does not exist."),
		}, nil
	}

	copyRuleRef := *rule
	oldRulePatten := copyRuleRef.Pattern
	oldRuleAttr := copyRuleRef.Attribute
	isChangeIndex := false
	if copyRuleRef.Attribute != request.Rule.Attribute {
		isChangeIndex = true
		copyRuleRef.Attribute = request.Rule.Attribute
	}
	if copyRuleRef.Pattern != request.Rule.Pattern {
		isChangeIndex = true
		copyRuleRef.Pattern = request.Rule.Pattern
	}
	copyRuleRef.RuleType = request.Rule.RuleType
	copyRuleRef.Description = request.Rule.Description
	copyRuleRef.ModTimestamp = strconv.FormatInt(time.Now().Unix(), 10)

	key := apt.GenerateServiceRuleKey(domainProject, request.ServiceId, request.RuleId)
	data, err := json.Marshal(copyRuleRef)
	if err != nil {
		log.Errorf(err, "update service rule[%s/%s] failed, marshal service rule failed, operator: %s",
			request.ServiceId, request.RuleId, remoteIP)
		return &pb.UpdateServiceRuleResponse{
			Response: proto.CreateResponse(scerr.ErrInternal, err.Error()),
		}, err
	}
	var opts []client.PluginOp
	if isChangeIndex {
		//加入新的rule index
		indexKey := apt.GenerateRuleIndexKey(domainProject, request.ServiceId, copyRuleRef.Attribute, copyRuleRef.Pattern)
		opts = append(opts, client.OpPut(client.WithStrKey(indexKey), client.WithStrValue(copyRuleRef.RuleId)))

		//删除旧的rule index
		oldIndexKey := apt.GenerateRuleIndexKey(domainProject, request.ServiceId, oldRuleAttr, oldRulePatten)
		opts = append(opts, client.OpDel(client.WithStrKey(oldIndexKey)))
	}
	opts = append(opts, client.OpPut(client.WithStrKey(key), client.WithValue(data)))

	resp, err := client.Instance().TxnWithCmp(ctx, opts,
		[]client.CompareOp{client.OpCmp(
			client.CmpVer(util.StringToBytesWithNoCopy(apt.GenerateServiceKey(domainProject, request.ServiceId))),
			client.CmpNotEqual, 0)},
		nil)
	if err != nil {
		log.Errorf(err, "update service rule[%s/%s] failed, operator: %s", request.ServiceId, request.RuleId, remoteIP)
		return &pb.UpdateServiceRuleResponse{
			Response: proto.CreateResponse(scerr.ErrUnavailableBackend, err.Error()),
		}, err
	}
	if !resp.Succeeded {
		log.Errorf(err, "update service rule[%s/%s] failed, service does not exist, operator: %s",
			request.ServiceId, request.RuleId, remoteIP)
		return &pb.UpdateServiceRuleResponse{
			Response: proto.CreateResponse(scerr.ErrServiceNotExists, "Service does not exist."),
		}, nil
	}

	log.Infof("update service rule[%s/%s] successfully, operator: %s", request.ServiceId, request.RuleId, remoteIP)
	return &pb.UpdateServiceRuleResponse{
		Response: proto.CreateResponse(proto.Response_SUCCESS, "Get service rules successfully."),
	}, nil
}

func (ds *DataSource) DeleteRule(ctx context.Context, request *pb.DeleteServiceRulesRequest) (
	*pb.DeleteServiceRulesResponse, error) {
	remoteIP := util.GetIPFromContext(ctx)
	domainProject := util.ParseDomainProject(ctx)

	// service id存在性校验
	if !serviceUtil.ServiceExist(ctx, domainProject, request.ServiceId) {
		log.Errorf(nil, "delete service[%s] rules %v failed, service does not exist, operator: %s",
			request.ServiceId, request.RuleIds, remoteIP)
		return &pb.DeleteServiceRulesResponse{
			Response: proto.CreateResponse(scerr.ErrServiceNotExists, "Service does not exist."),
		}, nil
	}

	opts := []client.PluginOp{}
	key := ""
	indexKey := ""
	for _, ruleID := range request.RuleIds {
		key = apt.GenerateServiceRuleKey(domainProject, request.ServiceId, ruleID)
		log.Debugf("start delete service rule file: %s", key)
		data, err := serviceUtil.GetOneRule(ctx, domainProject, request.ServiceId, ruleID)
		if err != nil {
			log.Errorf(err, "delete service[%s] rules %v failed, get rule[%s] failed, operator: %s",
				request.ServiceId, request.RuleIds, ruleID, remoteIP)
			return &pb.DeleteServiceRulesResponse{
				Response: proto.CreateResponse(scerr.ErrInternal, err.Error()),
			}, err
		}
		if data == nil {
			log.Errorf(nil, "delete service[%s] rules %v failed, rule[%s] does not exist, operator: %s",
				request.ServiceId, request.RuleIds, ruleID, remoteIP)
			return &pb.DeleteServiceRulesResponse{
				Response: proto.CreateResponse(scerr.ErrRuleNotExists, "This rule does not exist."),
			}, nil
		}
		indexKey = apt.GenerateRuleIndexKey(domainProject, request.ServiceId, data.Attribute, data.Pattern)
		opts = append(opts,
			client.OpDel(client.WithStrKey(key)),
			client.OpDel(client.WithStrKey(indexKey)))
	}
	if len(opts) <= 0 {
		log.Errorf(nil, "delete service[%s] rules %v failed, no rule has been deleted, operator: %s",
			request.ServiceId, request.RuleIds, remoteIP)
		return &pb.DeleteServiceRulesResponse{
			Response: proto.CreateResponse(scerr.ErrRuleNotExists, "No service rule has been deleted."),
		}, nil
	}

	resp, err := client.BatchCommitWithCmp(ctx, opts,
		[]client.CompareOp{client.OpCmp(
			client.CmpVer(util.StringToBytesWithNoCopy(apt.GenerateServiceKey(domainProject, request.ServiceId))),
			client.CmpNotEqual, 0)},
		nil)
	if err != nil {
		log.Errorf(err, "delete service[%s] rules %v failed, operator: %s", request.ServiceId, request.RuleIds, remoteIP)
		return &pb.DeleteServiceRulesResponse{
			Response: proto.CreateResponse(scerr.ErrUnavailableBackend, err.Error()),
		}, err
	}
	if !resp.Succeeded {
		log.Errorf(err, "delete service[%s] rules %v failed, service does not exist, operator: %s",
			request.ServiceId, request.RuleIds, remoteIP)
		return &pb.DeleteServiceRulesResponse{
			Response: proto.CreateResponse(scerr.ErrServiceNotExists, "Service does not exist."),
		}, nil
	}

	log.Infof("delete service[%s] rules %v successfully, operator: %s", request.ServiceId, request.RuleIds, remoteIP)
	return &pb.DeleteServiceRulesResponse{
		Response: proto.CreateResponse(proto.Response_SUCCESS, "Delete service rules successfully."),
	}, nil
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

	pluginOps := make([]client.PluginOp, 0)
	if !ds.isSchemaEditable(service) {
		if len(service.Schemas) == 0 {
			res := quota.NewApplyQuotaResource(quota.SchemaQuotaType, domainProject, serviceID, int64(len(nonExistSchemaIds)))
			rst := quota.Apply(ctx, res)
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
					opts := schemaWithDatabaseOpera(client.OpPut, domainProject, serviceID, needUpdateSchema)
					pluginOps = append(pluginOps, opts...)
				} else {
					log.Warnf("schema[%s/%s] and it's summary already exist, skip to update, operator: %s",
						serviceID, needUpdateSchema.SchemaId, remoteIP)
				}
			}
		}

		for _, schema := range needAddSchemas {
			log.Infof("add new schema[%s/%s], operator: %s", serviceID, schema.SchemaId, remoteIP)
			opts := schemaWithDatabaseOpera(client.OpPut, domainProject, service.ServiceId, schema)
			pluginOps = append(pluginOps, opts...)
		}
	} else {
		quotaSize := len(needAddSchemas) - len(needDeleteSchemas)
		if quotaSize > 0 {
			res := quota.NewApplyQuotaResource(quota.SchemaQuotaType, domainProject, serviceID, int64(quotaSize))
			rst := quota.Apply(ctx, res)
			err := rst.Err
			if err != nil {
				log.Errorf(err, "modify service[%s] schemas failed, operator: %s", serviceID, remoteIP)
				return err
			}
		}

		var schemaIDs []string
		for _, schema := range needAddSchemas {
			log.Infof("add new schema[%s/%s], operator: %s", serviceID, schema.SchemaId, remoteIP)
			opts := schemaWithDatabaseOpera(client.OpPut, domainProject, service.ServiceId, schema)
			pluginOps = append(pluginOps, opts...)
			schemaIDs = append(schemaIDs, schema.SchemaId)
		}

		for _, schema := range needUpdateSchemas {
			log.Infof("update schema[%s/%s], operator: %s", serviceID, schema.SchemaId, remoteIP)
			opts := schemaWithDatabaseOpera(client.OpPut, domainProject, serviceID, schema)
			pluginOps = append(pluginOps, opts...)
			schemaIDs = append(schemaIDs, schema.SchemaId)
		}

		for _, schema := range needDeleteSchemas {
			log.Infof("delete non-existent schema[%s/%s], operator: %s", serviceID, schema.SchemaId, remoteIP)
			opts := schemaWithDatabaseOpera(client.OpDel, domainProject, serviceID, schema)
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
		resp, err := client.BatchCommitWithCmp(ctx, pluginOps,
			[]client.CompareOp{client.OpCmp(
				client.CmpVer(util.StringToBytesWithNoCopy(apt.GenerateServiceKey(domainProject, serviceID))),
				client.CmpNotEqual, 0)},
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

	var pluginOps []client.PluginOp
	isExist := isExistSchemaID(microService, []*pb.Schema{schema})

	if !ds.isSchemaEditable(microService) {
		if len(microService.Schemas) != 0 && !isExist {
			return scerr.NewError(scerr.ErrUndefinedSchemaID, "Non-existent schemaID can't be added request "+pb.ENV_PROD)
		}

		key := apt.GenerateServiceSchemaKey(domainProject, serviceID, schemaID)
		respSchema, err := kv.Store().Schema().Search(ctx, client.WithStrKey(key), client.WithCountOnly())
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
					"schema already exist, can not be changed request "+pb.ENV_PROD)
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
				return scerr.NewError(scerr.ErrModifySchemaNotAllow, "schema already exist, can not be changed request "+pb.ENV_PROD)
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

	resp, err := client.Instance().TxnWithCmp(ctx, pluginOps,
		[]client.CompareOp{client.OpCmp(
			client.CmpVer(util.StringToBytesWithNoCopy(apt.GenerateServiceKey(domainProject, serviceID))),
			client.CmpNotEqual, 0)},
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
		rsp, err := kv.Store().Instance().Search(ctx,
			client.WithStrKey(instancesKey),
			client.WithPrefix(),
			client.WithCountOnly())
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
	opts := []client.PluginOp{
		client.OpDel(client.WithStrKey(apt.GenerateServiceIndexKey(serviceKey))),
		client.OpDel(client.WithStrKey(apt.GenerateServiceAliasKey(serviceKey))),
		client.OpDel(client.WithStrKey(serviceIDKey)),
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
	opts = append(opts, client.OpDel(
		client.WithStrKey(apt.GenerateServiceRuleKey(domainProject, serviceID, "")),
		client.WithPrefix()))
	opts = append(opts, client.OpDel(client.WithStrKey(
		util.StringJoin([]string{apt.GetServiceRuleIndexRootKey(domainProject), serviceID, ""}, "/")),
		client.WithPrefix()))

	//删除schemas
	opts = append(opts, client.OpDel(
		client.WithStrKey(apt.GenerateServiceSchemaKey(domainProject, serviceID, "")),
		client.WithPrefix()))
	opts = append(opts, client.OpDel(
		client.WithStrKey(apt.GenerateServiceSchemaSummaryKey(domainProject, serviceID, "")),
		client.WithPrefix()))

	//删除tags
	opts = append(opts, client.OpDel(
		client.WithStrKey(apt.GenerateServiceTagKey(domainProject, serviceID))))

	//删除instances
	opts = append(opts, client.OpDel(
		client.WithStrKey(apt.GenerateInstanceKey(domainProject, serviceID, "")),
		client.WithPrefix()))
	opts = append(opts, client.OpDel(
		client.WithStrKey(apt.GenerateInstanceLeaseKey(domainProject, serviceID, "")),
		client.WithPrefix()))

	//删除实例
	err = serviceUtil.DeleteServiceAllInstances(ctx, serviceID)
	if err != nil {
		log.Errorf(err, "%s micro-service[%s] failed, revoke all instances failed, operator: %s",
			title, serviceID, remoteIP)
		return proto.CreateResponse(scerr.ErrUnavailableBackend, err.Error()), err
	}

	resp, err := client.Instance().TxnWithCmp(ctx, opts,
		[]client.CompareOp{client.OpCmp(
			client.CmpVer(util.StringToBytesWithNoCopy(serviceIDKey)),
			client.CmpNotEqual, 0)},
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
