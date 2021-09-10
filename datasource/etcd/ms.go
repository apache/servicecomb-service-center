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
	"strconv"
	"time"

	"github.com/apache/servicecomb-service-center/datasource"
	"github.com/apache/servicecomb-service-center/datasource/etcd/cache"
	"github.com/apache/servicecomb-service-center/datasource/etcd/path"
	"github.com/apache/servicecomb-service-center/datasource/etcd/sd"
	"github.com/apache/servicecomb-service-center/datasource/etcd/state/kvstore"
	serviceUtil "github.com/apache/servicecomb-service-center/datasource/etcd/util"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/util"
	"github.com/apache/servicecomb-service-center/server/core"
	"github.com/apache/servicecomb-service-center/server/plugin/quota"
	"github.com/apache/servicecomb-service-center/server/plugin/uuid"
	pb "github.com/go-chassis/cari/discovery"
	"github.com/go-chassis/cari/pkg/errsvc"
	"github.com/go-chassis/foundation/gopool"
	"github.com/jinzhu/copier"
	"github.com/little-cui/etcdadpt"
)

var (
	ErrUndefinedSchemaID    = pb.NewError(pb.ErrUndefinedSchemaID, datasource.ErrUndefinedSchemaID.Error())
	ErrModifySchemaNotAllow = pb.NewError(pb.ErrModifySchemaNotAllow, datasource.ErrModifySchemaNotAllow.Error())
)

type MetadataManager struct {
	// SchemaNotEditable determines whether schema modification is not allowed
	SchemaNotEditable bool
	// InstanceTTL options
	InstanceTTL int64
}

func newMetadataManager(schemaNotEditable bool, instanceTTL int64) datasource.MetadataManager {
	return &MetadataManager{
		SchemaNotEditable: schemaNotEditable,
		InstanceTTL:       instanceTTL,
	}
}

// RegisterService implement:
// 1. capsule request to etcd kv format
// 2. invoke etcd client to store data
// 3. check etcd-client response && construct createServiceResponse
func (ds *MetadataManager) RegisterService(ctx context.Context, request *pb.CreateServiceRequest) (
	*pb.CreateServiceResponse, error) {
	remoteIP := util.GetIPFromContext(ctx)
	service := request.Service
	serviceFlag := util.StringJoin([]string{
		service.Environment, service.AppId, service.ServiceName, service.Version}, path.SPLIT)
	domainProject := util.ParseDomainProject(ctx)

	serviceKey := &pb.MicroServiceKey{
		Tenant:      domainProject,
		Environment: service.Environment,
		AppId:       service.AppId,
		ServiceName: service.ServiceName,
		Alias:       service.Alias,
		Version:     service.Version,
	}

	index := path.GenerateServiceIndexKey(serviceKey)

	// 产生全局service id
	requestServiceID := service.ServiceId
	if len(requestServiceID) == 0 {
		ctx = util.SetContext(ctx, uuid.ContextKey, index)
		service.ServiceId = uuid.Generator().GetServiceID(ctx)
	}
	service.Timestamp = strconv.FormatInt(time.Now().Unix(), 10)
	service.ModTimestamp = service.Timestamp

	data, err := json.Marshal(service)
	if err != nil {
		log.Error(fmt.Sprintf("create micro-service[%s] failed, json marshal service failed, operator: %s",
			serviceFlag, remoteIP), err)
		return &pb.CreateServiceResponse{
			Response: pb.CreateResponse(pb.ErrInternal, err.Error()),
		}, err
	}

	key := path.GenerateServiceKey(domainProject, service.ServiceId)
	keyBytes := util.StringToBytesWithNoCopy(key)
	indexBytes := util.StringToBytesWithNoCopy(index)
	aliasBytes := util.StringToBytesWithNoCopy(path.GenerateServiceAliasKey(serviceKey))

	opts := []etcdadpt.OpOptions{
		etcdadpt.OpPut(etcdadpt.WithKey(keyBytes), etcdadpt.WithValue(data)),
		etcdadpt.OpPut(etcdadpt.WithKey(indexBytes), etcdadpt.WithStrValue(service.ServiceId)),
	}
	uniqueCmpOpts := []etcdadpt.CmpOptions{
		etcdadpt.OpCmp(etcdadpt.CmpVer(indexBytes), etcdadpt.CmpEqual, 0),
		etcdadpt.OpCmp(etcdadpt.CmpVer(keyBytes), etcdadpt.CmpEqual, 0),
	}
	failOpts := []etcdadpt.OpOptions{
		etcdadpt.OpGet(etcdadpt.WithKey(indexBytes)),
	}

	if len(serviceKey.Alias) > 0 {
		opts = append(opts, etcdadpt.OpPut(etcdadpt.WithKey(aliasBytes), etcdadpt.WithStrValue(service.ServiceId)))
		uniqueCmpOpts = append(uniqueCmpOpts,
			etcdadpt.OpCmp(etcdadpt.CmpVer(aliasBytes), etcdadpt.CmpEqual, 0))
		failOpts = append(failOpts, etcdadpt.OpGet(etcdadpt.WithKey(aliasBytes)))
	}

	resp, err := etcdadpt.TxnWithCmp(ctx, opts, uniqueCmpOpts, failOpts)
	if err != nil {
		log.Error(fmt.Sprintf("create micro-service[%s] failed, operator: %s",
			serviceFlag, remoteIP), err)
		return &pb.CreateServiceResponse{
			Response: pb.CreateResponse(pb.ErrUnavailableBackend, err.Error()),
		}, err
	}
	if !resp.Succeeded {
		if len(requestServiceID) != 0 {
			if len(resp.Kvs) == 0 ||
				requestServiceID != util.BytesToStringWithNoCopy(resp.Kvs[0].Value) {
				log.Warn(fmt.Sprintf("create micro-service[%s] failed, service already exists, operator: %s",
					serviceFlag, remoteIP))
				return &pb.CreateServiceResponse{
					Response: pb.CreateResponse(pb.ErrServiceAlreadyExists,
						"ServiceID conflict or found the same service with different id."),
				}, nil
			}
		}

		if len(resp.Kvs) == 0 {
			// internal error?
			log.Error(fmt.Sprintf("create micro-service[%s] failed, unexpected txn response, operator: %s",
				serviceFlag, remoteIP), nil)
			return &pb.CreateServiceResponse{
				Response: pb.CreateResponse(pb.ErrInternal, "Unexpected txn response."),
			}, nil
		}

		serviceIDInner := util.BytesToStringWithNoCopy(resp.Kvs[0].Value)
		log.Warn(fmt.Sprintf("create micro-service[%s][%s] failed, service already exists, operator: %s",
			serviceIDInner, serviceFlag, remoteIP))
		return &pb.CreateServiceResponse{
			Response:  pb.CreateResponse(pb.ResponseSuccess, "register service successfully"),
			ServiceId: serviceIDInner,
		}, nil
	}

	//TODO increase usage in quota system

	log.Info(fmt.Sprintf("create micro-service[%s][%s] successfully, operator: %s",
		service.ServiceId, serviceFlag, remoteIP))
	return &pb.CreateServiceResponse{
		Response:  pb.CreateResponse(pb.ResponseSuccess, "Register service successfully."),
		ServiceId: service.ServiceId,
	}, nil

}

func (ds *MetadataManager) GetServices(ctx context.Context, request *pb.GetServicesRequest) (
	*pb.GetServicesResponse, error) {
	services, err := serviceUtil.GetAllServiceUtil(ctx)
	if err != nil {
		log.Error("get all services by domain failed", err)
		return &pb.GetServicesResponse{
			Response: pb.CreateResponse(pb.ErrInternal, err.Error()),
		}, err
	}

	return &pb.GetServicesResponse{
		Response: pb.CreateResponse(pb.ResponseSuccess, "Get all services successfully."),
		Services: services,
	}, nil
}

func (ds *MetadataManager) GetService(ctx context.Context, request *pb.GetServiceRequest) (
	*pb.MicroService, error) {
	domainProject := util.ParseDomainProject(ctx)
	singleService, err := serviceUtil.GetService(ctx, domainProject, request.ServiceId)

	if err != nil {
		if errors.Is(err, datasource.ErrNoData) {
			log.Debug(fmt.Sprintf("get micro-service[%s] failed, service does not exist in db", request.ServiceId))
			return nil, pb.NewError(pb.ErrServiceNotExists, "Service does not exist.")
		}
		log.Error(fmt.Sprintf("get micro-service[%s] failed, get service file failed", request.ServiceId), err)
		return nil, pb.NewError(pb.ErrInternal, err.Error())
	}
	return singleService, nil
}

func (ds *MetadataManager) GetServiceDetail(ctx context.Context, request *pb.GetServiceRequest) (
	*pb.ServiceDetail, error) {
	domainProject := util.ParseDomainProject(ctx)

	service, err := serviceUtil.GetService(ctx, domainProject, request.ServiceId)
	if err != nil {
		if errors.Is(err, datasource.ErrNoData) {
			return nil, pb.NewError(pb.ErrServiceNotExists, "Service does not exist.")
		}
		return nil, pb.NewError(pb.ErrInternal, err.Error())
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
		log.Error(fmt.Sprintf("get service[%s/%s/%s] all versions failed",
			service.Environment, service.AppId, service.ServiceName), err)
		return nil, pb.NewError(pb.ErrInternal, err.Error())
	}

	options := []string{"tags", "instances", "schemas", "dependencies"}
	serviceInfo, err := getServiceDetailUtil(ctx, ServiceDetailOpt{
		domainProject: domainProject,
		service:       service,
		options:       options,
	})
	if err != nil {
		return nil, pb.NewError(pb.ErrInternal, err.Error())
	}

	serviceInfo.MicroService = service
	serviceInfo.MicroServiceVersions = versions
	return serviceInfo, nil
}

func (ds *MetadataManager) ListServiceDetail(ctx context.Context, request *pb.GetServicesInfoRequest) (
	*pb.GetServicesInfoResponse, error) {
	ctx = util.WithCacheOnly(ctx)

	optionMap := make(map[string]struct{}, len(request.Options))
	for _, opt := range request.Options {
		optionMap[opt] = struct{}{}
	}

	options := make([]string, 0, len(optionMap))
	if _, ok := optionMap["all"]; ok {
		optionMap["statistics"] = struct{}{}
		options = []string{"tags", "instances", "schemas", "dependencies"}
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
			return nil, pb.NewError(pb.ErrInternal, err.Error())
		}
		if len(optionMap) == 1 {
			return &pb.GetServicesInfoResponse{
				Statistics: st,
			}, nil
		}
	}

	//获取所有服务
	services, err := serviceUtil.GetAllServiceUtil(ctx)
	if err != nil {
		log.Error("get all services by domain failed", err)
		return nil, pb.NewError(pb.ErrInternal, err.Error())
	}

	allServiceDetails := make([]*pb.ServiceDetail, 0, len(services))
	domainProject := util.ParseDomainProject(ctx)
	for _, service := range services {
		if !ds.filterServices(domainProject, request, service) {
			continue
		}

		serviceDetail, err := getServiceDetailUtil(ctx, ServiceDetailOpt{
			domainProject: domainProject,
			service:       service,
			countOnly:     request.CountOnly,
			options:       options,
		})
		if err != nil {
			return nil, pb.NewError(pb.ErrInternal, err.Error())
		}
		serviceDetail.MicroService = service
		tmpServiceDetail := &pb.ServiceDetail{}
		err = copier.CopyWithOption(tmpServiceDetail, serviceDetail, copier.Option{DeepCopy: true})
		if err != nil {
			return nil, pb.NewError(pb.ErrInternal, err.Error())
		}
		tmpServiceDetail.MicroService.Properties = nil
		tmpServiceDetail.MicroService.Schemas = nil
		instances := tmpServiceDetail.Instances
		for _, instance := range instances {
			instance.Properties = nil
		}
		allServiceDetails = append(allServiceDetails, tmpServiceDetail)
	}

	return &pb.GetServicesInfoResponse{
		AllServicesDetail: allServiceDetails,
		Statistics:        st,
	}, nil
}

func (ds *MetadataManager) filterServices(domainProject string, request *pb.GetServicesInfoRequest, service *pb.MicroService) bool {
	if !request.WithShared && datasource.IsGlobal(pb.MicroServiceToKey(domainProject, service)) {
		return false
	}
	if len(request.Environment) > 0 && request.Environment != service.Environment {
		return false
	}
	if len(request.AppId) > 0 && request.AppId != service.AppId {
		return false
	}
	if len(request.ServiceName) > 0 && request.ServiceName != service.ServiceName {
		return false
	}
	return true
}

func (ds *MetadataManager) GetOverview(ctx context.Context, request *pb.GetServicesRequest) (
	*pb.Statistics, error) {
	ctx = util.WithCacheOnly(ctx)
	st, err := statistics(ctx, false)
	if err != nil {
		return nil, pb.NewError(pb.ErrInternal, err.Error())
	}
	return st, nil
}

func (ds *MetadataManager) ListApp(ctx context.Context, request *pb.GetAppsRequest) (*pb.GetAppsResponse, error) {
	domainProject := util.ParseDomainProject(ctx)
	key := path.GetServiceAppKey(domainProject, request.Environment, "")

	opts := append(serviceUtil.FromContext(ctx),
		etcdadpt.WithStrKey(key),
		etcdadpt.WithPrefix(),
		etcdadpt.WithKeyOnly())

	resp, err := sd.ServiceIndex().Search(ctx, opts...)
	if err != nil {
		return nil, err
	}
	l := len(resp.Kvs)
	if l == 0 {
		return &pb.GetAppsResponse{
			Response: pb.CreateResponse(pb.ResponseSuccess, "Get all applications successfully."),
		}, nil
	}

	apps := make([]string, 0, l)
	appMap := make(map[string]struct{}, l)
	for _, keyValue := range resp.Kvs {
		key := path.GetInfoFromSvcIndexKV(keyValue.Key)
		if !request.WithShared && datasource.IsGlobal(key) {
			continue
		}
		if _, ok := appMap[key.AppId]; ok {
			continue
		}
		appMap[key.AppId] = struct{}{}
		apps = append(apps, key.AppId)
	}

	return &pb.GetAppsResponse{
		Response: pb.CreateResponse(pb.ResponseSuccess, "Get all applications successfully."),
		AppIds:   apps,
	}, nil
}

func (ds *MetadataManager) ExistServiceByID(ctx context.Context, request *pb.GetExistenceByIDRequest) (*pb.GetExistenceByIDResponse, error) {
	domainProject := util.ParseDomainProject(ctx)
	return &pb.GetExistenceByIDResponse{
		Response: pb.CreateResponse(pb.ResponseSuccess, "Get all applications successfully."),
		Exist:    serviceUtil.ServiceExist(ctx, domainProject, request.ServiceId),
	}, nil
}

func (ds *MetadataManager) ExistService(ctx context.Context, request *pb.GetExistenceRequest) (*pb.GetExistenceResponse,
	error) {
	domainProject := util.ParseDomainProject(ctx)
	serviceFlag := util.StringJoin([]string{
		request.Environment, request.AppId, request.ServiceName, request.Version}, path.SPLIT)

	ids, exist, err := serviceUtil.FindServiceIds(ctx, &pb.MicroServiceKey{
		Environment: request.Environment,
		AppId:       request.AppId,
		ServiceName: request.ServiceName,
		Alias:       request.ServiceName,
		Version:     request.Version,
		Tenant:      domainProject,
	}, true)
	if err != nil {
		log.Error(fmt.Sprintf("micro-service[%s] exist failed, find serviceIDs failed", serviceFlag), err)
		return &pb.GetExistenceResponse{
			Response: pb.CreateResponse(pb.ErrInternal, err.Error()),
		}, err
	}
	if !exist {
		log.Info(fmt.Sprintf("micro-service[%s] exist failed, service does not exist", serviceFlag))
		return &pb.GetExistenceResponse{
			Response: pb.CreateResponse(pb.ErrServiceNotExists, serviceFlag+" does not exist."),
		}, nil
	}
	if len(ids) == 0 {
		log.Info(fmt.Sprintf("micro-service[%s] exist failed, version mismatch", serviceFlag))
		return &pb.GetExistenceResponse{
			Response: pb.CreateResponse(pb.ErrServiceVersionNotExists, serviceFlag+" version mismatch."),
		}, nil
	}
	return &pb.GetExistenceResponse{
		Response:  pb.CreateResponse(pb.ResponseSuccess, "get service id successfully."),
		ServiceId: ids[0], // 约定多个时，取较新版本
	}, nil
}

func (ds *MetadataManager) UpdateService(ctx context.Context, request *pb.UpdateServicePropsRequest) (
	*pb.UpdateServicePropsResponse, error) {
	remoteIP := util.GetIPFromContext(ctx)
	domainProject := util.ParseDomainProject(ctx)

	key := path.GenerateServiceKey(domainProject, request.ServiceId)
	microservice, err := serviceUtil.GetService(ctx, domainProject, request.ServiceId)
	if err != nil {
		if errors.Is(err, datasource.ErrNoData) {
			log.Debug(fmt.Sprintf("service does not exist, update service[%s] properties failed, operator: %s",
				request.ServiceId, remoteIP))
			return &pb.UpdateServicePropsResponse{
				Response: pb.CreateResponse(pb.ErrServiceNotExists, "Service does not exist."),
			}, nil
		}
		log.Error(fmt.Sprintf("update service[%s] properties failed, get service file failed, operator: %s",
			request.ServiceId, remoteIP), err)
		return &pb.UpdateServicePropsResponse{
			Response: pb.CreateResponse(pb.ErrInternal, err.Error()),
		}, err
	}

	copyServiceRef := *microservice
	copyServiceRef.Properties = request.Properties
	copyServiceRef.ModTimestamp = strconv.FormatInt(time.Now().Unix(), 10)

	data, err := json.Marshal(copyServiceRef)
	if err != nil {
		log.Error(fmt.Sprintf("update service[%s] properties failed, json marshal service failed, operator: %s",
			request.ServiceId, remoteIP), err)
		return &pb.UpdateServicePropsResponse{
			Response: pb.CreateResponse(pb.ErrInternal, err.Error()),
		}, err
	}

	// Set key file
	resp, err := etcdadpt.TxnWithCmp(ctx,
		etcdadpt.Ops(etcdadpt.OpPut(etcdadpt.WithStrKey(key), etcdadpt.WithValue(data))),
		etcdadpt.If(etcdadpt.NotEqualVer(key, 0)),
		nil)
	if err != nil {
		log.Error(fmt.Sprintf("update service[%s] properties failed, operator: %s", request.ServiceId, remoteIP), err)
		return &pb.UpdateServicePropsResponse{
			Response: pb.CreateResponse(pb.ErrUnavailableBackend, err.Error()),
		}, err
	}
	if !resp.Succeeded {
		log.Error(fmt.Sprintf("update service[%s] properties failed, service does not exist, operator: %s",
			request.ServiceId, remoteIP), err)
		return &pb.UpdateServicePropsResponse{
			Response: pb.CreateResponse(pb.ErrServiceNotExists, "Service does not exist."),
		}, nil
	}

	log.Info(fmt.Sprintf("update service[%s] properties successfully, operator: %s", request.ServiceId, remoteIP))
	return &pb.UpdateServicePropsResponse{
		Response: pb.CreateResponse(pb.ResponseSuccess, "update service successfully."),
	}, nil
}

func (ds *MetadataManager) UnregisterService(ctx context.Context, request *pb.DeleteServiceRequest) (
	*pb.DeleteServiceResponse, error) {
	resp, err := ds.DeleteServicePri(ctx, request.ServiceId, request.Force)
	return &pb.DeleteServiceResponse{
		Response: resp,
	}, err
}

// RegisterInstance TODO use ds.registerInstance() instead after refactor
func (ds *MetadataManager) RegisterInstance(ctx context.Context, request *pb.RegisterInstanceRequest) (
	*pb.RegisterInstanceResponse, error) {
	instanceID, respErr := ds.registerInstance(ctx, request)
	if respErr != nil {
		response, err := datasource.WrapErrResponse(respErr)
		return &pb.RegisterInstanceResponse{Response: response}, err
	}
	return &pb.RegisterInstanceResponse{
		Response:   pb.CreateResponse(pb.ResponseSuccess, "Register service instance successfully."),
		InstanceId: instanceID,
	}, nil
}

func (ds *MetadataManager) registerInstance(ctx context.Context, request *pb.RegisterInstanceRequest) (string, error) {
	remoteIP := util.GetIPFromContext(ctx)
	instance := request.Instance

	//允许自定义id
	if len(instance.InstanceId) > 0 {
		needRegister, err := ds.sendHeartbeatInstead(ctx, instance)
		if err != nil {
			return "", err
		}
		if !needRegister {
			return instance.InstanceId, nil
		}
	} else {
		instance.InstanceId = uuid.Generator().GetInstanceID(ctx)
	}

	ttl := ds.calcInstanceTTL(instance)
	instanceFlag := fmt.Sprintf("ttl %ds, endpoints %v, host '%s', serviceID %s",
		ttl, instance.Endpoints, instance.HostName, instance.ServiceId)

	//先以domain/project的方式组装
	domainProject := util.ParseDomainProject(ctx)

	instanceID := instance.InstanceId
	data, err := json.Marshal(instance)
	if err != nil {
		log.Error(fmt.Sprintf("register instance failed, %s, instanceID %s, operator %s",
			instanceFlag, instanceID, remoteIP), err)
		return "", pb.NewError(pb.ErrInternal, err.Error())
	}

	leaseID, err := etcdadpt.Instance().LeaseGrant(ctx, ttl)
	if err != nil {
		log.Error(fmt.Sprintf("grant lease failed, %s, operator: %s", instanceFlag, remoteIP), err)
		return "", pb.NewError(pb.ErrUnavailableBackend, err.Error())
	}

	// build the request options
	key := path.GenerateInstanceKey(domainProject, instance.ServiceId, instanceID)
	hbKey := path.GenerateInstanceLeaseKey(domainProject, instance.ServiceId, instanceID)

	leaseOp := etcdadpt.WithLease(leaseID)
	resp, err := etcdadpt.TxnWithCmp(ctx,
		etcdadpt.Ops(
			etcdadpt.OpPut(etcdadpt.WithStrKey(key), etcdadpt.WithValue(data), leaseOp),
			etcdadpt.OpPut(etcdadpt.WithStrKey(hbKey), etcdadpt.WithStrValue(fmt.Sprintf("%d", leaseID)), leaseOp),
		),
		etcdadpt.If(etcdadpt.NotEqualVer(path.GenerateServiceKey(domainProject, instance.ServiceId), 0)),
		nil)
	if err != nil {
		log.Error(fmt.Sprintf("register instance failed, %s, instanceID %s, operator %s",
			instanceFlag, instanceID, remoteIP), err)
		return "", pb.NewError(pb.ErrUnavailableBackend, err.Error())
	}
	if !resp.Succeeded {
		log.Error(fmt.Sprintf("register instance failed, %s, instanceID %s, operator %s: service does not exist",
			instanceFlag, instanceID, remoteIP), nil)
		return "", pb.NewError(pb.ErrServiceNotExists, "Service does not exist.")
	}

	log.Info(fmt.Sprintf("register instance %s, instanceID %s, operator %s",
		instanceFlag, instanceID, remoteIP))
	return instanceID, nil
}

func (ds *MetadataManager) calcInstanceTTL(instance *pb.MicroServiceInstance) int64 {
	if instance.HealthCheck == nil {
		instance.HealthCheck = &pb.HealthCheck{
			Mode:     pb.CHECK_BY_HEARTBEAT,
			Interval: datasource.DefaultLeaseRenewalInterval,
			Times:    datasource.DefaultLeaseRetryTimes,
		}
	}
	ttl := int64(instance.HealthCheck.Interval * (instance.HealthCheck.Times + 1))
	if ds.InstanceTTL > 0 {
		ttl = ds.InstanceTTL
	}
	return ttl
}

func (ds *MetadataManager) sendHeartbeatInstead(ctx context.Context, instance *pb.MicroServiceInstance) (bool, error) {
	remoteIP := util.GetIPFromContext(ctx)
	// keep alive the lease ttl
	// there are two reasons for sending a heartbeat here:
	// 1. request the scenario the instance has been removed,
	//    the cast of registration operation can be reduced.
	// 2. request the self-protection scenario, the instance is unhealthy
	//    and needs to be re-registered.
	resp, err := ds.Heartbeat(ctx, &pb.HeartbeatRequest{ServiceId: instance.ServiceId,
		InstanceId: instance.InstanceId})
	if resp == nil {
		log.Error(fmt.Sprintf("register service[%s]'s instance failed, endpoints %v, host '%s', operator %s",
			instance.ServiceId, instance.Endpoints, instance.HostName, remoteIP), err)
		return false, pb.NewError(pb.ErrInternal, err.Error())
	}
	switch resp.Response.GetCode() {
	case pb.ResponseSuccess:
		log.Info(fmt.Sprintf("register instance successful, reuse instance[%s/%s], operator %s",
			instance.ServiceId, instance.InstanceId, remoteIP))
		return false, nil
	case pb.ErrInstanceNotExists:
		// register a new one
	default:
		log.Error(fmt.Sprintf("register instance failed, reuse instance[%s/%s], operator %s",
			instance.ServiceId, instance.InstanceId, remoteIP), err)
		return false, err
	}
	return true, nil
}

func (ds *MetadataManager) ExistInstanceByID(ctx context.Context, request *pb.MicroServiceInstanceKey) (*pb.GetExistenceByIDResponse, error) {
	domainProject := util.ParseDomainProject(ctx)
	exist, _ := serviceUtil.InstanceExist(ctx, domainProject, request.ServiceId, request.InstanceId)
	if !exist {
		return &pb.GetExistenceByIDResponse{
			Response: pb.CreateResponse(pb.ErrInstanceNotExists, "Check instance exist failed."),
			Exist:    false,
		}, datasource.ErrInstanceNotExists
	}
	return &pb.GetExistenceByIDResponse{
		Response: pb.CreateResponse(pb.ResponseSuccess, "Check service exists successfully."),
		Exist:    exist,
	}, nil
}

func (ds *MetadataManager) GetInstance(ctx context.Context, request *pb.GetOneInstanceRequest) (
	*pb.GetOneInstanceResponse, error) {
	domainProject := util.ParseDomainProject(ctx)

	service := &pb.MicroService{}
	var err error
	if len(request.ConsumerServiceId) > 0 {
		service, err = serviceUtil.GetService(ctx, domainProject, request.ConsumerServiceId)
		if err != nil {
			if errors.Is(err, datasource.ErrNoData) {
				log.Debug(fmt.Sprintf("consumer does not exist in db, consumer[%s] find provider instance[%s/%s]",
					request.ConsumerServiceId, request.ProviderServiceId, request.ProviderInstanceId))
				return &pb.GetOneInstanceResponse{
					Response: pb.CreateResponse(pb.ErrServiceNotExists,
						fmt.Sprintf("Consumer[%s] does not exist.", request.ConsumerServiceId)),
				}, nil
			}
			log.Error(fmt.Sprintf("get consumer failed, consumer[%s] find provider instance[%s/%s]",
				request.ConsumerServiceId, request.ProviderServiceId, request.ProviderInstanceId), err)
			return &pb.GetOneInstanceResponse{
				Response: pb.CreateResponse(pb.ErrInternal, err.Error()),
			}, err
		}
	}

	provider, err := serviceUtil.GetService(ctx, domainProject, request.ProviderServiceId)
	if err != nil {
		if errors.Is(err, datasource.ErrNoData) {
			log.Debug(fmt.Sprintf("provider does not exist in db, consumer[%s] find provider instance[%s/%s]",
				request.ConsumerServiceId, request.ProviderServiceId, request.ProviderInstanceId))
			return &pb.GetOneInstanceResponse{
				Response: pb.CreateResponse(pb.ErrServiceNotExists,
					fmt.Sprintf("Provider[%s] does not exist.", request.ProviderServiceId)),
			}, nil
		}
		log.Error(fmt.Sprintf("get provider failed, consumer[%s] find provider instance[%s/%s]",
			request.ConsumerServiceId, request.ProviderServiceId, request.ProviderInstanceId), err)
		return &pb.GetOneInstanceResponse{
			Response: pb.CreateResponse(pb.ErrInternal, err.Error()),
		}, err
	}

	findFlag := func() string {
		return fmt.Sprintf("Consumer[%s][%s/%s/%s/%s] find provider[%s][%s/%s/%s/%s] instance[%s]",
			request.ConsumerServiceId, service.Environment, service.AppId, service.ServiceName, service.Version,
			provider.ServiceId, provider.Environment, provider.AppId, provider.ServiceName, provider.Version,
			request.ProviderInstanceId)
	}

	var item *cache.VersionRuleCacheItem
	rev, _ := ctx.Value(util.CtxRequestRevision).(string)
	item, err = cache.FindInstances.GetWithProviderID(ctx, service, pb.MicroServiceToKey(domainProject, provider),
		&pb.HeartbeatSetElement{
			ServiceId: request.ProviderServiceId, InstanceId: request.ProviderInstanceId,
		}, request.Tags, rev)
	if err != nil {
		log.Error(fmt.Sprintf("find Instances by providerID failed, %s failed", findFlag()), err)
		return &pb.GetOneInstanceResponse{
			Response: pb.CreateResponse(pb.ErrInternal, err.Error()),
		}, err
	}
	if item == nil || len(item.Instances) == 0 {
		mes := fmt.Errorf("%s failed, provider instance does not exist", findFlag())
		log.Error("find Instances by ProviderID failed", mes)
		return &pb.GetOneInstanceResponse{
			Response: pb.CreateResponse(pb.ErrInstanceNotExists, mes.Error()),
		}, nil
	}

	instance := item.Instances[0]
	if rev == item.Rev {
		instance = nil // for gRPC
	}
	_ = util.WithResponseRev(ctx, item.Rev)

	return &pb.GetOneInstanceResponse{
		Response: pb.CreateResponse(pb.ResponseSuccess, "Get instance successfully."),
		Instance: instance,
	}, nil
}

func (ds *MetadataManager) GetInstances(ctx context.Context, request *pb.GetInstancesRequest) (*pb.GetInstancesResponse,
	error) {
	domainProject := util.ParseDomainProject(ctx)

	service := &pb.MicroService{}
	var err error
	if len(request.ConsumerServiceId) > 0 {
		service, err = serviceUtil.GetService(ctx, domainProject, request.ConsumerServiceId)
		if err != nil {
			if errors.Is(err, datasource.ErrNoData) {
				log.Debug(fmt.Sprintf("consumer does not exist in db, consumer[%s] find provider[%s] instances",
					request.ConsumerServiceId, request.ProviderServiceId))
				return &pb.GetInstancesResponse{
					Response: pb.CreateResponse(pb.ErrServiceNotExists,
						fmt.Sprintf("Consumer[%s] does not exist.", request.ConsumerServiceId)),
				}, nil
			}
			log.Error(fmt.Sprintf("get consumer failed, consumer[%s] find provider[%s] instances",
				request.ConsumerServiceId, request.ProviderServiceId), err)
			return &pb.GetInstancesResponse{
				Response: pb.CreateResponse(pb.ErrInternal, err.Error()),
			}, err
		}
	}

	provider, err := serviceUtil.GetService(ctx, domainProject, request.ProviderServiceId)
	if err != nil {
		if errors.Is(err, datasource.ErrNoData) {
			log.Debug(fmt.Sprintf("provider does not exist, consumer[%s] find provider[%s] instances",
				request.ConsumerServiceId, request.ProviderServiceId))
			return &pb.GetInstancesResponse{
				Response: pb.CreateResponse(pb.ErrServiceNotExists,
					fmt.Sprintf("Provider[%s] does not exist.", request.ProviderServiceId)),
			}, nil
		}
		log.Error(fmt.Sprintf("get provider failed, consumer[%s] find provider[%s] instances",
			request.ConsumerServiceId, request.ProviderServiceId), err)
		return &pb.GetInstancesResponse{
			Response: pb.CreateResponse(pb.ErrInternal, err.Error()),
		}, err
	}

	findFlag := func() string {
		return fmt.Sprintf("Consumer[%s][%s/%s/%s/%s] find provider[%s][%s/%s/%s/%s] instances",
			request.ConsumerServiceId, service.Environment, service.AppId, service.ServiceName, service.Version,
			provider.ServiceId, provider.Environment, provider.AppId, provider.ServiceName, provider.Version)
	}

	var item *cache.VersionRuleCacheItem
	rev, _ := ctx.Value(util.CtxRequestRevision).(string)
	item, err = cache.FindInstances.GetWithProviderID(ctx, service, pb.MicroServiceToKey(domainProject, provider),
		&pb.HeartbeatSetElement{
			ServiceId: request.ProviderServiceId,
		}, request.Tags, rev)
	if err != nil {
		log.Error(fmt.Sprintf("FindInstances.GetWithProviderID failed, %s failed", findFlag()), err)
		return &pb.GetInstancesResponse{
			Response: pb.CreateResponse(pb.ErrInternal, err.Error()),
		}, err
	}
	if item == nil || len(item.ServiceIds) == 0 {
		err := fmt.Errorf("%s failed, provider instance does not exist", findFlag())
		log.Error("FindInstances.GetWithProviderID failed", err)
		return &pb.GetInstancesResponse{
			Response: pb.CreateResponse(pb.ErrServiceNotExists, err.Error()),
		}, nil
	}

	instances := item.Instances
	if rev == item.Rev {
		instances = nil // for gRPC
	}
	_ = util.WithResponseRev(ctx, item.Rev)

	return &pb.GetInstancesResponse{
		Response:  pb.CreateResponse(pb.ResponseSuccess, "Query service instances successfully."),
		Instances: instances,
	}, nil
}

func (ds *MetadataManager) GetProviderInstances(ctx context.Context, request *pb.GetProviderInstancesRequest) (instances []*pb.MicroServiceInstance, rev string, err error) {
	var (
		maxRevs       = make([]int64, len(clustersIndex))
		counts        = make([]int64, len(clustersIndex))
		domainProject = util.ParseTargetDomainProject(ctx)
	)
	instances, err = ds.findInstances(ctx, domainProject, request.ProviderServiceId, maxRevs, counts)
	if err != nil {
		return
	}
	return instances, serviceUtil.FormatRevision(maxRevs, counts), nil
}

func (ds *MetadataManager) BatchGetProviderInstances(ctx context.Context, request *pb.BatchGetInstancesRequest) (instances []*pb.MicroServiceInstance, rev string, err error) {
	var (
		maxRevs       = make([]int64, len(clustersIndex))
		counts        = make([]int64, len(clustersIndex))
		domainProject = util.ParseTargetDomainProject(ctx)
	)
	if request == nil || len(request.ServiceIds) == 0 {
		return nil, "", fmt.Errorf("invalid param BatchGetInstancesRequest")
	}

	for _, providerServiceID := range request.ServiceIds {
		insts, err := ds.findInstances(ctx, domainProject, providerServiceID, maxRevs, counts)
		if err != nil {
			return nil, "", err
		}
		instances = append(instances, insts...)
	}

	return instances, serviceUtil.FormatRevision(maxRevs, counts), nil
}

func (ds *MetadataManager) findInstances(ctx context.Context, domainProject, serviceID string, maxRevs []int64, counts []int64) (instances []*pb.MicroServiceInstance, err error) {
	key := path.GenerateInstanceKey(domainProject, serviceID, "")
	opts := append(serviceUtil.FromContext(ctx), etcdadpt.WithStrKey(key), etcdadpt.WithPrefix())
	resp, err := sd.Instance().Search(ctx, opts...)
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

func (ds *MetadataManager) FindInstances(ctx context.Context, request *pb.FindInstancesRequest) (*pb.FindInstancesResponse,
	error) {
	provider := &pb.MicroServiceKey{
		Tenant:      util.ParseTargetDomainProject(ctx),
		Environment: request.Environment,
		AppId:       request.AppId,
		ServiceName: request.ServiceName,
		Alias:       request.Alias,
	}

	rev, ok := ctx.Value(util.CtxRequestRevision).(string)
	if !ok {
		err := errors.New("rev request context is not type string")
		log.Error("", err)
		return &pb.FindInstancesResponse{
			Response: pb.CreateResponse(pb.ErrInternal, err.Error()),
		}, err
	}

	if datasource.IsGlobal(provider) {
		return ds.findSharedServiceInstance(ctx, request, provider, rev)
	}
	return ds.findInstance(ctx, request, provider, rev)
}

func (ds *MetadataManager) findInstance(ctx context.Context, request *pb.FindInstancesRequest,
	provider *pb.MicroServiceKey, rev string) (*pb.FindInstancesResponse, error) {
	var err error
	domainProject := util.ParseDomainProject(ctx)
	service := &pb.MicroService{Environment: request.Environment}
	if len(request.ConsumerServiceId) > 0 {
		service, err = serviceUtil.GetService(ctx, domainProject, request.ConsumerServiceId)
		if err != nil {
			if errors.Is(err, datasource.ErrNoData) {
				log.Debug(fmt.Sprintf("consumer does not exist, consumer[%s] find provider[%s/%s/%s]",
					request.ConsumerServiceId, request.Environment, request.AppId, request.ServiceName))
				return &pb.FindInstancesResponse{
					Response: pb.CreateResponse(pb.ErrServiceNotExists,
						fmt.Sprintf("Consumer[%s] does not exist.", request.ConsumerServiceId)),
				}, nil
			}
			log.Error(fmt.Sprintf("get consumer failed, consumer[%s] find provider[%s/%s/%s]",
				request.ConsumerServiceId, request.Environment, request.AppId, request.ServiceName), err)
			return &pb.FindInstancesResponse{
				Response: pb.CreateResponse(pb.ErrInternal, err.Error()),
			}, err
		}
		provider.Environment = service.Environment
	}

	// provider is not a shared micro-service,
	// only allow shared micro-service instances found request different domains.
	ctx = util.SetTargetDomainProject(ctx, util.ParseDomain(ctx), util.ParseProject(ctx))
	provider.Tenant = util.ParseTargetDomainProject(ctx)

	findFlag := fmt.Sprintf("Consumer[%s][%s/%s/%s/%s] find provider[%s/%s/%s]",
		request.ConsumerServiceId, service.Environment, service.AppId, service.ServiceName, service.Version,
		provider.Environment, provider.AppId, provider.ServiceName)

	// cache
	var item *cache.VersionRuleCacheItem
	item, err = cache.FindInstances.Get(ctx, service, provider, request.Tags, rev)
	if err != nil {
		log.Error(fmt.Sprintf("FindInstancesCache.Get failed, %s failed", findFlag), err)
		return &pb.FindInstancesResponse{
			Response: pb.CreateResponse(pb.ErrInternal, err.Error()),
		}, err
	}
	if item == nil {
		err := fmt.Errorf("%s failed, provider does not exist", findFlag)
		log.Error("FindInstancesCache.Get failed", err)
		return &pb.FindInstancesResponse{
			Response: pb.CreateResponse(pb.ErrServiceNotExists, err.Error()),
		}, nil
	}

	// add dependency queue
	if len(request.ConsumerServiceId) > 0 &&
		len(item.ServiceIds) > 0 &&
		!cache.DependencyRule.ExistRule(ctx, request.ConsumerServiceId, provider) {
		provider, err = ds.reshapeProviderKey(ctx, provider, item.ServiceIds[0])
		if err != nil {
			return nil, err
		}
		if provider != nil {
			err = serviceUtil.AddServiceVersionRule(ctx, domainProject, service, provider)
		} else {
			err := fmt.Errorf("%s failed, provider does not exist", findFlag)
			log.Error("AddServiceVersionRule failed", err)
			return &pb.FindInstancesResponse{
				Response: pb.CreateResponse(pb.ErrServiceNotExists, err.Error()),
			}, nil
		}
		if err != nil {
			log.Error(fmt.Sprintf("AddServiceVersionRule failed, %s failed", findFlag), err)
			return &pb.FindInstancesResponse{
				Response: pb.CreateResponse(pb.ErrInternal, err.Error()),
			}, err
		}
	}

	return ds.genFindResult(ctx, rev, item)
}

func (ds *MetadataManager) findSharedServiceInstance(ctx context.Context, request *pb.FindInstancesRequest,
	provider *pb.MicroServiceKey, rev string) (*pb.FindInstancesResponse, error) {
	var err error
	service := &pb.MicroService{Environment: request.Environment}
	// it means the shared micro-services must be the same env with SC.
	provider.Environment = core.Service.Environment
	findFlag := fmt.Sprintf("find shared provider[%s/%s/%s/%s]", provider.Environment, provider.AppId, provider.ServiceName, provider.Version)

	// cache
	var item *cache.VersionRuleCacheItem
	item, err = cache.FindInstances.Get(ctx, service, provider, request.Tags, rev)
	if err != nil {
		log.Error(fmt.Sprintf("FindInstancesCache.Get failed, %s failed", findFlag), err)
		return &pb.FindInstancesResponse{
			Response: pb.CreateResponse(pb.ErrInternal, err.Error()),
		}, err
	}
	if item == nil {
		err := fmt.Errorf("%s failed, provider does not exist", findFlag)
		log.Error("FindInstancesCache.Get failed", err)
		return &pb.FindInstancesResponse{
			Response: pb.CreateResponse(pb.ErrServiceNotExists, err.Error()),
		}, nil
	}

	return ds.genFindResult(ctx, rev, item)
}

func (ds *MetadataManager) genFindResult(ctx context.Context, oldRev string, item *cache.VersionRuleCacheItem) (
	*pb.FindInstancesResponse, error) {
	instances := item.Instances
	if oldRev == item.Rev {
		instances = nil // for gRPC
	}
	// TODO support gRPC output context
	_ = util.WithResponseRev(ctx, item.Rev)
	return &pb.FindInstancesResponse{
		Response:  pb.CreateResponse(pb.ResponseSuccess, "Query service instances successfully."),
		Instances: instances,
	}, nil
}

func (ds *MetadataManager) reshapeProviderKey(ctx context.Context, provider *pb.MicroServiceKey, providerID string) (
	*pb.MicroServiceKey, error) {
	// service name 可能是别名，所以重新获取
	providerService, err := serviceUtil.GetService(ctx, provider.Tenant, providerID)
	if err != nil {
		return nil, err
	}

	provider = pb.MicroServiceToKey(provider.Tenant, providerService)
	provider.Version = datasource.AllVersions // just compatible to old version
	return provider, nil
}

func (ds *MetadataManager) UpdateInstanceStatus(ctx context.Context, request *pb.UpdateInstanceStatusRequest) (*pb.UpdateInstanceStatusResponse, error) {
	domainProject := util.ParseDomainProject(ctx)
	updateStatusFlag := util.StringJoin([]string{request.ServiceId, request.InstanceId, request.Status}, path.SPLIT)

	instance, err := serviceUtil.GetInstance(ctx, domainProject, request.ServiceId, request.InstanceId)
	if err != nil {
		log.Error(fmt.Sprintf("update instance[%s] status failed", updateStatusFlag), err)
		return &pb.UpdateInstanceStatusResponse{
			Response: pb.CreateResponse(pb.ErrInternal, err.Error()),
		}, err
	}
	if instance == nil {
		log.Error(fmt.Sprintf("update instance[%s] status failed, instance does not exist", updateStatusFlag), nil)
		return &pb.UpdateInstanceStatusResponse{
			Response: pb.CreateResponse(pb.ErrInstanceNotExists, "Service instance does not exist."),
		}, nil
	}

	copyInstanceRef := *instance
	copyInstanceRef.Status = request.Status

	if err := serviceUtil.UpdateInstance(ctx, domainProject, &copyInstanceRef); err != nil {
		log.Error(fmt.Sprintf("update instance[%s] status failed", updateStatusFlag), err)
		resp := &pb.UpdateInstanceStatusResponse{
			Response: pb.CreateResponseWithSCErr(err),
		}
		if err.InternalError() {
			return resp, err
		}
		return resp, nil
	}

	log.Info(fmt.Sprintf("update instance[%s] status successfully", updateStatusFlag))
	return &pb.UpdateInstanceStatusResponse{
		Response: pb.CreateResponse(pb.ResponseSuccess, "Update service instance status successfully."),
	}, nil
}

func (ds *MetadataManager) UpdateInstanceProperties(ctx context.Context, request *pb.UpdateInstancePropsRequest) (
	*pb.UpdateInstancePropsResponse, error) {
	domainProject := util.ParseDomainProject(ctx)
	instanceFlag := util.StringJoin([]string{request.ServiceId, request.InstanceId}, path.SPLIT)

	instance, err := serviceUtil.GetInstance(ctx, domainProject, request.ServiceId, request.InstanceId)
	if err != nil {
		log.Error(fmt.Sprintf("update instance[%s] properties failed", instanceFlag), err)
		return &pb.UpdateInstancePropsResponse{
			Response: pb.CreateResponse(pb.ErrInternal, err.Error()),
		}, err
	}
	if instance == nil {
		log.Error(fmt.Sprintf("update instance[%s] properties failed, instance does not exist", instanceFlag), nil)
		return &pb.UpdateInstancePropsResponse{
			Response: pb.CreateResponse(pb.ErrInstanceNotExists, "Service instance does not exist."),
		}, nil
	}

	copyInstanceRef := *instance
	copyInstanceRef.Properties = request.Properties

	if err := serviceUtil.UpdateInstance(ctx, domainProject, &copyInstanceRef); err != nil {
		log.Error(fmt.Sprintf("update instance[%s] properties failed", instanceFlag), err)
		resp := &pb.UpdateInstancePropsResponse{
			Response: pb.CreateResponseWithSCErr(err),
		}
		if err.InternalError() {
			return resp, err
		}
		return resp, nil
	}

	log.Info(fmt.Sprintf("update instance[%s] properties successfully", instanceFlag))
	return &pb.UpdateInstancePropsResponse{
		Response: pb.CreateResponse(pb.ResponseSuccess, "Update service instance properties successfully."),
	}, nil
}

func (ds *MetadataManager) HeartbeatSet(ctx context.Context, request *pb.HeartbeatSetRequest) (*pb.HeartbeatSetResponse, error) {
	domainProject := util.ParseDomainProject(ctx)

	heartBeatCount := len(request.Instances)
	existFlag := make(map[string]bool, heartBeatCount)
	instancesHbRst := make(chan *pb.InstanceHbRst, heartBeatCount)
	noMultiCounter := 0
	for _, heartbeatElement := range request.Instances {
		if _, ok := existFlag[heartbeatElement.ServiceId+heartbeatElement.InstanceId]; ok {
			log.Warn(fmt.Sprintf("instance[%s/%s] is duplicate request heartbeat set",
				heartbeatElement.ServiceId, heartbeatElement.InstanceId))
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
		log.Info(fmt.Sprintf("batch update heartbeats[%d] successfully", count))
		return &pb.HeartbeatSetResponse{
			Response:  pb.CreateResponse(pb.ResponseSuccess, "Heartbeat set successfully."),
			Instances: instanceHbRstArr,
		}, nil
	}
	log.Error(fmt.Sprintf("batch update heartbeats failed, %v", request.Instances), nil)
	return &pb.HeartbeatSetResponse{
		Response:  pb.CreateResponse(pb.ErrInstanceNotExists, "Heartbeat set failed."),
		Instances: instanceHbRstArr,
	}, nil
}

func (ds *MetadataManager) BatchFind(ctx context.Context, request *pb.BatchFindInstancesRequest) (
	*pb.BatchFindInstancesResponse, error) {
	response := &pb.BatchFindInstancesResponse{
		Response: pb.CreateResponse(pb.ResponseSuccess, "Batch query service instances successfully."),
	}

	var err error
	// find services
	response.Services, err = ds.batchFindServices(ctx, request)
	if err != nil {
		return &pb.BatchFindInstancesResponse{
			Response: pb.CreateResponse(pb.ErrInternal, err.Error()),
		}, err
	}

	// find instance
	response.Instances, err = ds.batchFindInstances(ctx, request)
	if err != nil {
		return &pb.BatchFindInstancesResponse{
			Response: pb.CreateResponse(pb.ErrInternal, err.Error()),
		}, err
	}

	return response, nil
}

func (ds *MetadataManager) batchFindServices(ctx context.Context, request *pb.BatchFindInstancesRequest) (
	*pb.BatchFindResult, error) {
	if len(request.Services) == 0 {
		return nil, nil
	}
	cloneCtx := util.CloneContext(ctx)

	services := &pb.BatchFindResult{}
	failedResult := make(map[int32]*pb.FindFailedResult)
	for index, key := range request.Services {
		findCtx := util.WithRequestRev(cloneCtx, key.Rev)
		resp, err := ds.FindInstances(findCtx, &pb.FindInstancesRequest{
			ConsumerServiceId: request.ConsumerServiceId,
			AppId:             key.Service.AppId,
			ServiceName:       key.Service.ServiceName,
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

func (ds *MetadataManager) batchFindInstances(ctx context.Context, request *pb.BatchFindInstancesRequest) (*pb.BatchFindResult, error) {
	if len(request.Instances) == 0 {
		return nil, nil
	}
	cloneCtx := util.CloneContext(ctx)
	// can not find the shared provider instances
	cloneCtx = util.SetTargetDomainProject(cloneCtx, util.ParseDomain(ctx), util.ParseProject(ctx))

	instances := &pb.BatchFindResult{}
	failedResult := make(map[int32]*pb.FindFailedResult)
	for index, key := range request.Instances {
		getCtx := util.WithRequestRev(cloneCtx, key.Rev)
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

func (ds *MetadataManager) UnregisterInstance(ctx context.Context, request *pb.UnregisterInstanceRequest) (
	*pb.UnregisterInstanceResponse, error) {
	remoteIP := util.GetIPFromContext(ctx)
	domainProject := util.ParseDomainProject(ctx)
	serviceID := request.ServiceId
	instanceID := request.InstanceId

	instanceFlag := util.StringJoin([]string{serviceID, instanceID}, path.SPLIT)

	err := revokeInstance(ctx, domainProject, serviceID, instanceID)
	if err != nil {
		log.Error(fmt.Sprintf("unregister instance failed, instance[%s], operator %s: revoke instance failed",
			instanceFlag, remoteIP), err)
		resp := &pb.UnregisterInstanceResponse{
			Response: pb.CreateResponseWithSCErr(err),
		}
		if err.InternalError() {
			return resp, err
		}
		return resp, nil
	}

	log.Info(fmt.Sprintf("unregister instance[%s], operator %s", instanceFlag, remoteIP))
	return &pb.UnregisterInstanceResponse{
		Response: pb.CreateResponse(pb.ResponseSuccess, "Unregister service instance successfully."),
	}, nil
}

func (ds *MetadataManager) Heartbeat(ctx context.Context, request *pb.HeartbeatRequest) (*pb.HeartbeatResponse, error) {
	remoteIP := util.GetIPFromContext(ctx)
	domainProject := util.ParseDomainProject(ctx)
	instanceFlag := util.StringJoin([]string{request.ServiceId, request.InstanceId}, path.SPLIT)

	_, ttl, err := serviceUtil.HeartbeatUtil(ctx, domainProject, request.ServiceId, request.InstanceId)
	if err != nil {
		log.Error(fmt.Sprintf("heartbeat failed, instance[%s]. operator %s",
			instanceFlag, remoteIP), err)
		resp := &pb.HeartbeatResponse{
			Response: pb.CreateResponseWithSCErr(err),
		}
		if err.InternalError() {
			return resp, err
		}
		return resp, nil
	}

	if ttl == 0 {
		log.Error(fmt.Sprintf("heartbeat successful, but renew instance[%s] failed. operator %s", instanceFlag, remoteIP),
			errors.New("connect backend timed out"))
	} else {
		log.Info(fmt.Sprintf("heartbeat successful, renew instance[%s] ttl to %d. operator %s",
			instanceFlag, ttl, remoteIP))
	}
	return &pb.HeartbeatResponse{
		Response: pb.CreateResponse(pb.ResponseSuccess,
			"Update service instance heartbeat successfully."),
	}, nil
}

func (ds *MetadataManager) GetAllInstances(ctx context.Context, request *pb.GetAllInstancesRequest) (*pb.GetAllInstancesResponse, error) {
	domainProject := util.ParseDomainProject(ctx)
	key := path.GetInstanceRootKey(domainProject) + path.SPLIT
	opts := append(serviceUtil.FromContext(ctx), etcdadpt.WithStrKey(key), etcdadpt.WithPrefix())
	kvs, err := sd.Instance().Search(ctx, opts...)
	if err != nil {
		return nil, err
	}
	resp := &pb.GetAllInstancesResponse{
		Response: pb.CreateResponse(pb.ResponseSuccess, "Get all instances successfully"),
	}
	for _, keyValue := range kvs.Kvs {
		instance, ok := keyValue.Value.(*pb.MicroServiceInstance)
		if !ok {
			log.Warn(fmt.Sprintf("Unexpected value format! %s", util.BytesToStringWithNoCopy(keyValue.Key)))
			continue
		}
		resp.Instances = append(resp.Instances, instance)
	}
	return resp, nil
}

func (ds *MetadataManager) ModifySchemas(ctx context.Context, request *pb.ModifySchemasRequest) (
	*pb.ModifySchemasResponse, error) {
	remoteIP := util.GetIPFromContext(ctx)
	serviceID := request.ServiceId
	domainProject := util.ParseDomainProject(ctx)

	serviceInfo, err := serviceUtil.GetService(ctx, domainProject, serviceID)
	if err != nil {
		if errors.Is(err, datasource.ErrNoData) {
			log.Debug(fmt.Sprintf("modify service[%s] schemas failed, service does not exist in db, operator: %s",
				serviceID, remoteIP))
			return &pb.ModifySchemasResponse{
				Response: pb.CreateResponse(pb.ErrServiceNotExists, "Service does not exist."),
			}, nil
		}
		log.Error(fmt.Sprintf("modify service[%s] schemas failed, get service failed, operator: %s",
			serviceID, remoteIP), err)
		return &pb.ModifySchemasResponse{
			Response: pb.CreateResponse(pb.ErrInternal, err.Error()),
		}, err
	}

	if respErr := ds.modifySchemas(ctx, domainProject, serviceInfo, request.Schemas); respErr != nil {
		log.Error(fmt.Sprintf("modify service[%s] schemas failed, operator: %s", serviceID, remoteIP), respErr)
		response, err := datasource.WrapErrResponse(respErr)
		return &pb.ModifySchemasResponse{
			Response: response,
		}, err
	}

	return &pb.ModifySchemasResponse{
		Response: pb.CreateResponse(pb.ResponseSuccess, "modify schemas info successfully."),
	}, nil
}

func (ds *MetadataManager) ModifySchema(ctx context.Context, request *pb.ModifySchemaRequest) (
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
		log.Error(fmt.Sprintf("modify schema[%s/%s] failed, operator: %s", serviceID, schemaID, remoteIP), err)
		resp := &pb.ModifySchemaResponse{
			Response: pb.CreateResponseWithSCErr(err),
		}
		if err.InternalError() {
			return resp, err
		}
		return resp, nil
	}

	log.Info(fmt.Sprintf("modify schema[%s/%s] successfully, operator: %s", serviceID, schemaID, remoteIP))
	return &pb.ModifySchemaResponse{
		Response: pb.CreateResponse(pb.ResponseSuccess, "modify schema info success"),
	}, nil
}

func (ds *MetadataManager) ExistSchema(ctx context.Context, request *pb.GetExistenceRequest) (
	*pb.GetExistenceResponse, error) {
	domainProject := util.ParseDomainProject(ctx)

	if !serviceUtil.ServiceExist(ctx, domainProject, request.ServiceId) {
		log.Warn(fmt.Sprintf("schema[%s/%s] exist failed, service does not exist", request.ServiceId, request.SchemaId))
		return &pb.GetExistenceResponse{
			Response: pb.CreateResponse(pb.ErrServiceNotExists, "service does not exist."),
		}, nil
	}

	key := path.GenerateServiceSchemaKey(domainProject, request.ServiceId, request.SchemaId)
	exist, err := checkSchemaInfoExist(ctx, key)
	if err != nil {
		log.Error(fmt.Sprintf("schema[%s/%s] exist failed, get schema failed", request.ServiceId, request.SchemaId), err)
		return &pb.GetExistenceResponse{
			Response: pb.CreateResponse(pb.ErrInternal, err.Error()),
		}, err
	}
	if !exist {
		log.Info(fmt.Sprintf("schema[%s/%s] exist failed, schema does not exist", request.ServiceId, request.SchemaId))
		return &pb.GetExistenceResponse{
			Response: pb.CreateResponse(pb.ErrSchemaNotExists, "schema does not exist."),
		}, nil
	}
	schemaSummary, err := getSchemaSummary(ctx, domainProject, request.ServiceId, request.SchemaId)
	if err != nil {
		log.Error(fmt.Sprintf("schema[%s/%s] exist failed, get schema summary failed",
			request.ServiceId, request.SchemaId), err)
		return &pb.GetExistenceResponse{
			Response: pb.CreateResponse(pb.ErrInternal, err.Error()),
		}, err
	}
	return &pb.GetExistenceResponse{
		Response: pb.CreateResponse(pb.ResponseSuccess, "Schema exist."),
		SchemaId: request.SchemaId,
		Summary:  schemaSummary,
	}, nil
}

func (ds *MetadataManager) GetSchema(ctx context.Context, request *pb.GetSchemaRequest) (*pb.GetSchemaResponse, error) {
	domainProject := util.ParseDomainProject(ctx)

	if !serviceUtil.ServiceExist(ctx, domainProject, request.ServiceId) {
		log.Error(fmt.Sprintf("get schema[%s/%s] failed, service does not exist",
			request.ServiceId, request.SchemaId), nil)
		return &pb.GetSchemaResponse{
			Response: pb.CreateResponse(pb.ErrServiceNotExists, "Service does not exist."),
		}, nil
	}

	key := path.GenerateServiceSchemaKey(domainProject, request.ServiceId, request.SchemaId)
	opts := append(serviceUtil.FromContext(ctx), etcdadpt.WithStrKey(key))
	resp, errDo := sd.Schema().Search(ctx, opts...)
	if errDo != nil {
		log.Error(fmt.Sprintf("get schema[%s/%s] failed", request.ServiceId, request.SchemaId), errDo)
		return &pb.GetSchemaResponse{
			Response: pb.CreateResponse(pb.ErrUnavailableBackend, errDo.Error()),
		}, errDo
	}
	if resp.Count == 0 {
		log.Error(fmt.Sprintf("get schema[%s/%s] failed, schema does not exists",
			request.ServiceId, request.SchemaId), errDo)
		return &pb.GetSchemaResponse{
			Response: pb.CreateResponse(pb.ErrSchemaNotExists, "Do not have this schema info."),
		}, nil
	}

	schemaSummary, err := getSchemaSummary(ctx, domainProject, request.ServiceId, request.SchemaId)
	if err != nil {
		log.Error(fmt.Sprintf("get schema[%s/%s] failed, get schema summary failed",
			request.ServiceId, request.SchemaId), err)
		return &pb.GetSchemaResponse{
			Response: pb.CreateResponse(pb.ErrInternal, err.Error()),
		}, err
	}

	return &pb.GetSchemaResponse{
		Response:      pb.CreateResponse(pb.ResponseSuccess, "Get schema info successfully."),
		Schema:        util.BytesToStringWithNoCopy(resp.Kvs[0].Value.([]byte)),
		SchemaSummary: schemaSummary,
	}, nil
}

func (ds *MetadataManager) GetAllSchemas(ctx context.Context, request *pb.GetAllSchemaRequest) (
	*pb.GetAllSchemaResponse, error) {
	domainProject := util.ParseDomainProject(ctx)

	service, err := serviceUtil.GetService(ctx, domainProject, request.ServiceId)
	if err != nil {
		if errors.Is(err, datasource.ErrNoData) {
			log.Debug(fmt.Sprintf("get service[%s] all schemas failed, service does not exist in db", request.ServiceId))
			return &pb.GetAllSchemaResponse{
				Response: pb.CreateResponse(pb.ErrServiceNotExists, "Service does not exist."),
			}, nil
		}
		log.Error(fmt.Sprintf("get service[%s] all schemas failed, get service failed", request.ServiceId), err)
		return &pb.GetAllSchemaResponse{
			Response: pb.CreateResponse(pb.ErrInternal, err.Error()),
		}, err
	}

	schemasList := service.Schemas
	if len(schemasList) == 0 {
		return &pb.GetAllSchemaResponse{
			Response: pb.CreateResponse(pb.ResponseSuccess, "Do not have this schema info."),
			Schemas:  []*pb.Schema{},
		}, nil
	}

	key := path.GenerateServiceSchemaSummaryKey(domainProject, request.ServiceId, "")
	opts := append(serviceUtil.FromContext(ctx), etcdadpt.WithStrKey(key), etcdadpt.WithPrefix())
	resp, errDo := sd.SchemaSummary().Search(ctx, opts...)
	if errDo != nil {
		log.Error(fmt.Sprintf("get service[%s] all schema summaries failed", request.ServiceId), errDo)
		return &pb.GetAllSchemaResponse{
			Response: pb.CreateResponse(pb.ErrUnavailableBackend, errDo.Error()),
		}, errDo
	}

	respWithSchema := &kvstore.Response{}
	if request.WithSchema {
		key := path.GenerateServiceSchemaKey(domainProject, request.ServiceId, "")
		opts := append(serviceUtil.FromContext(ctx), etcdadpt.WithStrKey(key), etcdadpt.WithPrefix())
		respWithSchema, errDo = sd.Schema().Search(ctx, opts...)
		if errDo != nil {
			log.Error(fmt.Sprintf("get service[%s] all schemas failed", request.ServiceId), errDo)
			return &pb.GetAllSchemaResponse{
				Response: pb.CreateResponse(pb.ErrUnavailableBackend, errDo.Error()),
			}, errDo
		}
	}

	schemas := make([]*pb.Schema, 0, len(schemasList))
	for _, schemaID := range schemasList {
		tempSchema := &pb.Schema{}
		tempSchema.SchemaId = schemaID
		for _, summarySchema := range resp.Kvs {
			_, _, schemaIDOfSummary := path.GetInfoFromSchemaSummaryKV(summarySchema.Key)
			if schemaID == schemaIDOfSummary {
				tempSchema.Summary = summarySchema.Value.(string)
			}
		}

		for _, contentSchema := range respWithSchema.Kvs {
			_, _, schemaIDOfSchema := path.GetInfoFromSchemaKV(contentSchema.Key)
			if schemaID == schemaIDOfSchema {
				tempSchema.Schema = util.BytesToStringWithNoCopy(contentSchema.Value.([]byte))
			}
		}
		schemas = append(schemas, tempSchema)
	}

	return &pb.GetAllSchemaResponse{
		Response: pb.CreateResponse(pb.ResponseSuccess, "Get all schema info successfully."),
		Schemas:  schemas,
	}, nil
}

func (ds *MetadataManager) DeleteSchema(ctx context.Context, request *pb.DeleteSchemaRequest) (
	*pb.DeleteSchemaResponse, error) {
	remoteIP := util.GetIPFromContext(ctx)
	domainProject := util.ParseDomainProject(ctx)

	if !serviceUtil.ServiceExist(ctx, domainProject, request.ServiceId) {
		log.Error(fmt.Sprintf("delete schema[%s/%s] failed, service does not exist, operator: %s",
			request.ServiceId, request.SchemaId, remoteIP), nil)
		return &pb.DeleteSchemaResponse{
			Response: pb.CreateResponse(pb.ErrServiceNotExists, "Service does not exist."),
		}, nil
	}

	key := path.GenerateServiceSchemaKey(domainProject, request.ServiceId, request.SchemaId)
	exist, err := serviceUtil.CheckSchemaInfoExist(ctx, key)
	if err != nil {
		log.Error(fmt.Sprintf("delete schema[%s/%s] failed, operator: %s",
			request.ServiceId, request.SchemaId, remoteIP), err)
		return &pb.DeleteSchemaResponse{
			Response: pb.CreateResponse(pb.ErrInternal, err.Error()),
		}, err
	}
	if !exist {
		log.Error(fmt.Sprintf("delete schema[%s/%s] failed, schema does not exist, operator: %s",
			request.ServiceId, request.SchemaId, remoteIP), nil)
		return &pb.DeleteSchemaResponse{
			Response: pb.CreateResponse(pb.ErrSchemaNotExists, "Schema info does not exist."),
		}, nil
	}
	epSummaryKey := path.GenerateServiceSchemaSummaryKey(domainProject, request.ServiceId, request.SchemaId)
	resp, errDo := etcdadpt.TxnWithCmp(ctx,
		etcdadpt.Ops(
			etcdadpt.OpDel(etcdadpt.WithStrKey(epSummaryKey)),
			etcdadpt.OpDel(etcdadpt.WithStrKey(key)),
		),
		etcdadpt.If(etcdadpt.NotEqualVer(path.GenerateServiceKey(domainProject, request.ServiceId), 0)),
		nil)
	if errDo != nil {
		log.Error(fmt.Sprintf("delete schema[%s/%s] failed, operator: %s",
			request.ServiceId, request.SchemaId, remoteIP), errDo)
		return &pb.DeleteSchemaResponse{
			Response: pb.CreateResponse(pb.ErrUnavailableBackend, errDo.Error()),
		}, errDo
	}
	if !resp.Succeeded {
		log.Error(fmt.Sprintf("delete schema[%s/%s] failed, service does not exist, operator: %s",
			request.ServiceId, request.SchemaId, remoteIP), nil)
		return &pb.DeleteSchemaResponse{
			Response: pb.CreateResponse(pb.ErrServiceNotExists, "Service does not exist."),
		}, nil
	}

	log.Info(fmt.Sprintf("delete schema[%s/%s] info successfully, operator: %s",
		request.ServiceId, request.SchemaId, remoteIP))
	return &pb.DeleteSchemaResponse{
		Response: pb.CreateResponse(pb.ResponseSuccess, "Delete schema info successfully."),
	}, nil
}

func (ds *MetadataManager) AddTags(ctx context.Context, request *pb.AddServiceTagsRequest) (*pb.AddServiceTagsResponse, error) {
	remoteIP := util.GetIPFromContext(ctx)
	domainProject := util.ParseDomainProject(ctx)

	// service id存在性校验
	if !serviceUtil.ServiceExist(ctx, domainProject, request.ServiceId) {
		log.Error(fmt.Sprintf("add service[%s]'s tags %v failed, service does not exist, operator: %s",
			request.ServiceId, request.Tags, remoteIP), nil)
		return &pb.AddServiceTagsResponse{
			Response: pb.CreateResponse(pb.ErrServiceNotExists, "Service does not exist."),
		}, nil
	}

	checkErr := serviceUtil.AddTagIntoETCD(ctx, domainProject, request.ServiceId, request.Tags)
	if checkErr != nil {
		log.Error(fmt.Sprintf("add service[%s]'s tags %v failed, operator: %s",
			request.ServiceId, request.Tags, remoteIP), checkErr)
		resp := &pb.AddServiceTagsResponse{
			Response: pb.CreateResponseWithSCErr(checkErr),
		}
		if checkErr.InternalError() {
			return resp, checkErr
		}
		return resp, nil
	}

	log.Info(fmt.Sprintf("add service[%s]'s tags %v successfully, operator: %s", request.ServiceId, request.Tags, remoteIP))
	return &pb.AddServiceTagsResponse{
		Response: pb.CreateResponse(pb.ResponseSuccess, "Add service tags successfully."),
	}, nil
}

func (ds *MetadataManager) GetTags(ctx context.Context, request *pb.GetServiceTagsRequest) (*pb.GetServiceTagsResponse, error) {
	var err error
	domainProject := util.ParseDomainProject(ctx)
	if !serviceUtil.ServiceExist(ctx, domainProject, request.ServiceId) {
		log.Error(fmt.Sprintf("get service[%s]'s tags failed, service does not exist", request.ServiceId), err)
		return &pb.GetServiceTagsResponse{
			Response: pb.CreateResponse(pb.ErrServiceNotExists, "Service does not exist."),
		}, nil
	}
	tags, err := serviceUtil.GetTagsUtils(ctx, domainProject, request.ServiceId)
	if err != nil {
		log.Error(fmt.Sprintf("get service[%s]'s tags failed, get tags failed", request.ServiceId), err)
		return &pb.GetServiceTagsResponse{
			Response: pb.CreateResponse(pb.ErrInternal, err.Error()),
		}, err
	}

	return &pb.GetServiceTagsResponse{
		Response: pb.CreateResponse(pb.ResponseSuccess, "Get service tags successfully."),
		Tags:     tags,
	}, nil
}

func (ds *MetadataManager) UpdateTag(ctx context.Context, request *pb.UpdateServiceTagRequest) (*pb.UpdateServiceTagResponse, error) {
	var err error
	remoteIP := util.GetIPFromContext(ctx)
	tagFlag := util.StringJoin([]string{request.Key, request.Value}, path.SPLIT)
	domainProject := util.ParseDomainProject(ctx)

	if !serviceUtil.ServiceExist(ctx, domainProject, request.ServiceId) {
		log.Error(fmt.Sprintf("update service[%s]'s tag[%s] failed, service does not exist, operator: %s",
			request.ServiceId, tagFlag, remoteIP), err)
		return &pb.UpdateServiceTagResponse{
			Response: pb.CreateResponse(pb.ErrServiceNotExists, "Service does not exist."),
		}, nil
	}

	tags, err := serviceUtil.GetTagsUtils(ctx, domainProject, request.ServiceId)
	if err != nil {
		log.Error(fmt.Sprintf("update service[%s]'s tag[%s] failed, get tag failed, operator: %s",
			request.ServiceId, tagFlag, remoteIP), err)
		return &pb.UpdateServiceTagResponse{
			Response: pb.CreateResponse(pb.ErrInternal, err.Error()),
		}, err
	}

	//check if the tag exists
	if _, ok := tags[request.Key]; !ok {
		log.Error(fmt.Sprintf("update service[%s]'s tag[%s] failed, tag does not exist, operator: %s",
			request.ServiceId, tagFlag, remoteIP), nil)
		return &pb.UpdateServiceTagResponse{
			Response: pb.CreateResponse(pb.ErrTagNotExists, "Tag does not exist, please add one first."),
		}, nil
	}

	copyTags := make(map[string]string, len(tags))
	for k, v := range tags {
		copyTags[k] = v
	}
	copyTags[request.Key] = request.Value

	checkErr := serviceUtil.AddTagIntoETCD(ctx, domainProject, request.ServiceId, copyTags)
	if checkErr != nil {
		log.Error(fmt.Sprintf("update service[%s]'s tag[%s] failed, operator: %s",
			request.ServiceId, tagFlag, remoteIP), checkErr)
		resp := &pb.UpdateServiceTagResponse{
			Response: pb.CreateResponseWithSCErr(checkErr),
		}
		if checkErr.InternalError() {
			return resp, checkErr
		}
		return resp, nil
	}

	log.Info(fmt.Sprintf("update service[%s]'s tag[%s] successfully, operator: %s", request.ServiceId, tagFlag, remoteIP))
	return &pb.UpdateServiceTagResponse{
		Response: pb.CreateResponse(pb.ResponseSuccess, "Update service tag success."),
	}, nil
}

func (ds *MetadataManager) DeleteTags(ctx context.Context, request *pb.DeleteServiceTagsRequest) (*pb.DeleteServiceTagsResponse, error) {
	remoteIP := util.GetIPFromContext(ctx)
	domainProject := util.ParseDomainProject(ctx)

	if !serviceUtil.ServiceExist(ctx, domainProject, request.ServiceId) {
		log.Error(fmt.Sprintf("delete service[%s]'s tags %v failed, service does not exist, operator: %s",
			request.ServiceId, request.Keys, remoteIP), nil)
		return &pb.DeleteServiceTagsResponse{
			Response: pb.CreateResponse(pb.ErrServiceNotExists, "Service does not exist."),
		}, nil
	}

	tags, err := serviceUtil.GetTagsUtils(ctx, domainProject, request.ServiceId)
	if err != nil {
		log.Error(fmt.Sprintf("delete service[%s]'s tags %v failed, get service tags failed, operator: %s",
			request.ServiceId, request.Keys, remoteIP), err)
		return &pb.DeleteServiceTagsResponse{
			Response: pb.CreateResponse(pb.ErrInternal, err.Error()),
		}, err
	}

	copyTags := make(map[string]string, len(tags))
	for k, v := range tags {
		copyTags[k] = v
	}
	for _, key := range request.Keys {
		if _, ok := copyTags[key]; !ok {
			log.Error(fmt.Sprintf("delete service[%s]'s tags %v failed, tag[%s] does not exist, operator: %s",
				request.ServiceId, request.Keys, key, remoteIP), nil)
			return &pb.DeleteServiceTagsResponse{
				Response: pb.CreateResponse(pb.ErrTagNotExists, "Delete tags failed for this key "+key+" does not exist."),
			}, nil
		}
		delete(copyTags, key)
	}

	// the capacity of tags may be 0
	data, err := json.Marshal(copyTags)
	if err != nil {
		log.Error(fmt.Sprintf("delete service[%s]'s tags %v failed, marshall service tags failed, operator: %s",
			request.ServiceId, request.Keys, remoteIP), err)
		return &pb.DeleteServiceTagsResponse{
			Response: pb.CreateResponse(pb.ErrInternal, err.Error()),
		}, err
	}

	key := path.GenerateServiceTagKey(domainProject, request.ServiceId)

	resp, err := etcdadpt.TxnWithCmp(ctx,
		etcdadpt.Ops(etcdadpt.OpPut(etcdadpt.WithStrKey(key), etcdadpt.WithValue(data))),
		etcdadpt.If(etcdadpt.NotEqualVer(path.GenerateServiceKey(domainProject, request.ServiceId), 0)),
		nil)
	if err != nil {
		log.Error(fmt.Sprintf("delete service[%s]'s tags %v failed, operator: %s",
			request.ServiceId, request.Keys, remoteIP), err)
		return &pb.DeleteServiceTagsResponse{
			Response: pb.CreateResponse(pb.ErrUnavailableBackend, err.Error()),
		}, err
	}
	if !resp.Succeeded {
		log.Error(fmt.Sprintf("delete service[%s]'s tags %v failed, service does not exist, operator: %s",
			request.ServiceId, request.Keys, remoteIP), err)
		return &pb.DeleteServiceTagsResponse{
			Response: pb.CreateResponse(pb.ErrServiceNotExists, "Service does not exist."),
		}, nil
	}

	log.Info(fmt.Sprintf("delete service[%s]'s tags %v successfully, operator: %s", request.ServiceId, request.Keys, remoteIP))
	return &pb.DeleteServiceTagsResponse{
		Response: pb.CreateResponse(pb.ResponseSuccess, "Delete service tags successfully."),
	}, nil
}

func (ds *MetadataManager) modifySchemas(ctx context.Context, domainProject string, service *pb.MicroService,
	schemas []*pb.Schema) error {
	remoteIP := util.GetIPFromContext(ctx)
	serviceID := service.ServiceId
	schemasFromDatabase, err := getSchemasFromDatabase(ctx, domainProject, serviceID)
	if err != nil {
		log.Error(fmt.Sprintf("modify service[%s] schemas failed, get schemas failed, operator: %s",
			serviceID, remoteIP), nil)
		return pb.NewError(pb.ErrUnavailableBackend, err.Error())
	}

	needUpdateSchemas, needAddSchemas, needDeleteSchemas, nonExistSchemaIds :=
		datasource.SchemasAnalysis(schemas, schemasFromDatabase, service.Schemas)

	pluginOps := make([]etcdadpt.OpOptions, 0)
	if !ds.isSchemaEditable() {
		if len(service.Schemas) == 0 {
			res := quota.NewApplyQuotaResource(quota.TypeSchema, domainProject, serviceID, int64(len(nonExistSchemaIds)))
			errQuota := quota.Apply(ctx, res)
			if errQuota != nil {
				log.Error(fmt.Sprintf("modify service[%s] schemas failed, operator: %s", serviceID, remoteIP), errQuota)
				return errQuota
			}

			service.Schemas = nonExistSchemaIds
			opt, err := serviceUtil.UpdateService(domainProject, serviceID, service)
			if err != nil {
				log.Error(fmt.Sprintf("modify service[%s] schemas failed, update service.Schemas failed, operator: %s",
					serviceID, remoteIP), err)
				return pb.NewError(pb.ErrInternal, err.Error())
			}
			pluginOps = append(pluginOps, opt)
		} else {
			if len(nonExistSchemaIds) != 0 {
				errInfo := fmt.Errorf("non-existent schemaIDs %v", nonExistSchemaIds)
				log.Error(fmt.Sprintf("modify service[%s] schemas failed, operator: %s", serviceID, remoteIP), errInfo)
				return pb.NewError(pb.ErrUndefinedSchemaID, errInfo.Error())
			}
			for _, needUpdateSchema := range needUpdateSchemas {
				exist, err := isExistSchemaSummary(ctx, domainProject, serviceID, needUpdateSchema.SchemaId)
				if err != nil {
					return pb.NewError(pb.ErrInternal, err.Error())
				}
				if !exist {
					opts := schemaWithDatabaseOpera(etcdadpt.OpPut, domainProject, serviceID, needUpdateSchema)
					pluginOps = append(pluginOps, opts...)
				} else {
					log.Warn(fmt.Sprintf("schema[%s/%s] and it's summary already exist, skip to update, operator: %s",
						serviceID, needUpdateSchema.SchemaId, remoteIP))
				}
			}
		}

		for _, schema := range needAddSchemas {
			log.Info(fmt.Sprintf("add new schema[%s/%s], operator: %s", serviceID, schema.SchemaId, remoteIP))
			opts := schemaWithDatabaseOpera(etcdadpt.OpPut, domainProject, service.ServiceId, schema)
			pluginOps = append(pluginOps, opts...)
		}
	} else {
		quotaSize := len(needAddSchemas) - len(needDeleteSchemas)
		if quotaSize > 0 {
			res := quota.NewApplyQuotaResource(quota.TypeSchema, domainProject, serviceID, int64(quotaSize))
			errQuota := quota.Apply(ctx, res)
			if errQuota != nil {
				log.Error(fmt.Sprintf("modify service[%s] schemas failed, operator: %s", serviceID, remoteIP), errQuota)
				return errQuota
			}
		}

		var schemaIDs []string
		for _, schema := range needAddSchemas {
			log.Info(fmt.Sprintf("add new schema[%s/%s], operator: %s", serviceID, schema.SchemaId, remoteIP))
			opts := schemaWithDatabaseOpera(etcdadpt.OpPut, domainProject, service.ServiceId, schema)
			pluginOps = append(pluginOps, opts...)
			schemaIDs = append(schemaIDs, schema.SchemaId)
		}

		for _, schema := range needUpdateSchemas {
			log.Info(fmt.Sprintf("update schema[%s/%s], operator: %s", serviceID, schema.SchemaId, remoteIP))
			opts := schemaWithDatabaseOpera(etcdadpt.OpPut, domainProject, serviceID, schema)
			pluginOps = append(pluginOps, opts...)
			schemaIDs = append(schemaIDs, schema.SchemaId)
		}

		for _, schema := range needDeleteSchemas {
			log.Info(fmt.Sprintf("delete non-existent schema[%s/%s], operator: %s", serviceID, schema.SchemaId, remoteIP))
			opts := schemaWithDatabaseOpera(etcdadpt.OpDel, domainProject, serviceID, schema)
			pluginOps = append(pluginOps, opts...)
		}

		service.Schemas = schemaIDs
		opt, err := serviceUtil.UpdateService(domainProject, serviceID, service)
		if err != nil {
			log.Error(fmt.Sprintf("modify service[%s] schemas failed, update service.Schemas failed, operator: %s",
				serviceID, remoteIP), err)
			return pb.NewError(pb.ErrInternal, err.Error())
		}
		pluginOps = append(pluginOps, opt)
	}

	if len(pluginOps) != 0 {
		resp, err := etcdadpt.TxnWithCmp(ctx, pluginOps,
			etcdadpt.If(etcdadpt.NotEqualVer(path.GenerateServiceKey(domainProject, serviceID), 0)),
			nil)
		if err != nil {
			return pb.NewError(pb.ErrUnavailableBackend, err.Error())
		}
		if !resp.Succeeded {
			return pb.NewError(pb.ErrServiceNotExists, "Service does not exist.")
		}
	}
	return nil
}

func (ds *MetadataManager) isSchemaEditable() bool {
	return !ds.SchemaNotEditable
}

func (ds *MetadataManager) modifySchema(ctx context.Context, serviceID string, schema *pb.Schema) *errsvc.Error {
	remoteIP := util.GetIPFromContext(ctx)
	domainProject := util.ParseDomainProject(ctx)
	schemaID := schema.SchemaId

	microService, err := serviceUtil.GetService(ctx, domainProject, serviceID)
	if err != nil {
		if errors.Is(err, datasource.ErrNoData) {
			log.Debug(fmt.Sprintf("microService does not exist, modify schema[%s/%s] failed, operator: %s",
				serviceID, schemaID, remoteIP))
			return pb.NewError(pb.ErrServiceNotExists, "Service does not exist")
		}
		log.Error(fmt.Sprintf("modify schema[%s/%s] failed, get `microService failed, operator: %s",
			serviceID, schemaID, remoteIP), err)
		return pb.NewError(pb.ErrInternal, err.Error())
	}

	var pluginOps []etcdadpt.OpOptions
	isExist := isExistSchemaID(microService, []*pb.Schema{schema})

	if !ds.isSchemaEditable() {
		if len(microService.Schemas) != 0 && !isExist {
			return ErrUndefinedSchemaID
		}

		key := path.GenerateServiceSchemaKey(domainProject, serviceID, schemaID)
		respSchema, err := sd.Schema().Search(ctx, etcdadpt.WithStrKey(key), etcdadpt.WithCountOnly())
		if err != nil {
			log.Error(fmt.Sprintf("modify schema[%s/%s] failed, get schema summary failed, operator: %s",
				serviceID, schemaID, remoteIP), err)
			return pb.NewError(pb.ErrUnavailableBackend, err.Error())
		}

		if respSchema.Count != 0 {
			if len(schema.Summary) == 0 {
				log.Error(fmt.Sprintf("schema readonly mode, schema[%s/%s] already exists, can not be changed, operator: %s",
					serviceID, schemaID, remoteIP), err)
				return ErrModifySchemaNotAllow
			}

			exist, err := isExistSchemaSummary(ctx, domainProject, serviceID, schemaID)
			if err != nil {
				log.Error(fmt.Sprintf("check schema[%s/%s] summary existence failed, operator: %s",
					serviceID, schemaID, remoteIP), err)
				return pb.NewError(pb.ErrInternal, err.Error())
			}
			if exist {
				log.Error(fmt.Sprintf("schema readonly mode, schema[%s/%s] already exist, can not be changed, operator: %s",
					serviceID, schemaID, remoteIP), err)
				return ErrModifySchemaNotAllow
			}
		}

		if len(microService.Schemas) == 0 {
			microService.Schemas = append(microService.Schemas, schemaID)
			opt, err := serviceUtil.UpdateService(domainProject, serviceID, microService)
			if err != nil {
				log.Error(fmt.Sprintf("modify schema[%s/%s] failed, update microService.Schemas failed, operator: %s",
					serviceID, schemaID, remoteIP), err)
				return pb.NewError(pb.ErrInternal, err.Error())
			}
			pluginOps = append(pluginOps, opt)
		}
	} else {
		if !isExist {
			microService.Schemas = append(microService.Schemas, schemaID)
			opt, err := serviceUtil.UpdateService(domainProject, serviceID, microService)
			if err != nil {
				log.Error(fmt.Sprintf("modify schema[%s/%s] failed, update microService.Schemas failed, operator: %s",
					serviceID, schemaID, remoteIP), err)
				return pb.NewError(pb.ErrInternal, err.Error())
			}
			pluginOps = append(pluginOps, opt)
		}
	}

	opts := commitSchemaInfo(domainProject, serviceID, schema)
	pluginOps = append(pluginOps, opts...)

	resp, err := etcdadpt.TxnWithCmp(ctx, pluginOps,
		etcdadpt.If(etcdadpt.NotEqualVer(path.GenerateServiceKey(domainProject, serviceID), 0)),
		nil)
	if err != nil {
		return pb.NewError(pb.ErrUnavailableBackend, err.Error())
	}
	if !resp.Succeeded {
		return pb.NewError(pb.ErrServiceNotExists, "Service does not exist.")
	}
	return nil
}

func (ds *MetadataManager) DeleteServicePri(ctx context.Context, serviceID string, force bool) (*pb.Response, error) {
	remoteIP := util.GetIPFromContext(ctx)
	domainProject := util.ParseDomainProject(ctx)

	title := "delete"
	if force {
		title = "force delete"
	}

	if serviceID == core.Service.ServiceId {
		err := errors.New("not allow to delete service center")
		log.Error(fmt.Sprintf("%s micro-service[%s] failed, operator: %s", title, serviceID, remoteIP), err)
		return pb.CreateResponse(pb.ErrInvalidParams, err.Error()), nil
	}

	microservice, err := serviceUtil.GetService(ctx, domainProject, serviceID)
	if err != nil {
		if errors.Is(err, datasource.ErrNoData) {
			log.Debug(fmt.Sprintf("service does not exist, %s micro-service[%s] failed, operator: %s",
				title, serviceID, remoteIP))
			return pb.CreateResponse(pb.ErrServiceNotExists, "Service does not exist."), nil
		}
		log.Error(fmt.Sprintf("%s micro-service[%s] failed, get service file failed, operator: %s",
			title, serviceID, remoteIP), err)
		return pb.CreateResponse(pb.ErrInternal, err.Error()), err
	}

	// 强制删除，则与该服务相关的信息删除，非强制删除： 如果作为该被依赖（作为provider，提供服务,且不是只存在自依赖）或者存在实例，则不能删除
	if !force {
		services, err := serviceUtil.GetConsumerIds(ctx, domainProject, microservice)
		if err != nil {
			log.Error(fmt.Sprintf("delete micro-service[%s] failed, get service dependency failed, operator: %s",
				serviceID, remoteIP), err)
			return pb.CreateResponse(pb.ErrInternal, err.Error()), err
		}
		if l := len(services); l > 1 || (l == 1 && services[0] != serviceID) {
			log.Error(fmt.Sprintf("delete micro-service[%s] failed, other services[%d] depend on it, operator: %s",
				serviceID, l, remoteIP), nil)
			return pb.CreateResponse(pb.ErrDependedOnConsumer, "Can not delete this service, other service rely it."), err
		}

		instancesKey := path.GenerateInstanceKey(domainProject, serviceID, "")
		rsp, err := sd.Instance().Search(ctx,
			etcdadpt.WithStrKey(instancesKey),
			etcdadpt.WithPrefix(),
			etcdadpt.WithCountOnly())
		if err != nil {
			log.Error(fmt.Sprintf("delete micro-service[%s] failed, get instances failed, operator: %s",
				serviceID, remoteIP), err)
			return pb.CreateResponse(pb.ErrUnavailableBackend, err.Error()), err
		}

		if rsp.Count > 0 {
			log.Error(fmt.Sprintf("delete micro-service[%s] failed, service deployed instances[%d], operator: %s",
				serviceID, rsp.Count, remoteIP), nil)
			return pb.CreateResponse(pb.ErrDeployedInstance, "Can not delete the service deployed instance(s)."), err
		}
	}

	serviceIDKey := path.GenerateServiceKey(domainProject, serviceID)
	serviceKey := &pb.MicroServiceKey{
		Tenant:      domainProject,
		Environment: microservice.Environment,
		AppId:       microservice.AppId,
		ServiceName: microservice.ServiceName,
		Version:     microservice.Version,
		Alias:       microservice.Alias,
	}
	opts := []etcdadpt.OpOptions{
		etcdadpt.OpDel(etcdadpt.WithStrKey(path.GenerateServiceIndexKey(serviceKey))),
		etcdadpt.OpDel(etcdadpt.WithStrKey(path.GenerateServiceAliasKey(serviceKey))),
		etcdadpt.OpDel(etcdadpt.WithStrKey(serviceIDKey)),
	}

	//删除依赖规则
	optDeleteDep, err := serviceUtil.DeleteDependencyForDeleteService(domainProject, serviceID, serviceKey)
	if err != nil {
		log.Error(fmt.Sprintf("%s micro-service[%s] failed, delete dependency failed, operator: %s",
			title, serviceID, remoteIP), err)
		return pb.CreateResponse(pb.ErrInternal, err.Error()), err
	}
	opts = append(opts, optDeleteDep)

	//删除schemas
	opts = append(opts, etcdadpt.OpDel(
		etcdadpt.WithStrKey(path.GenerateServiceSchemaKey(domainProject, serviceID, "")),
		etcdadpt.WithPrefix()))
	opts = append(opts, etcdadpt.OpDel(
		etcdadpt.WithStrKey(path.GenerateServiceSchemaSummaryKey(domainProject, serviceID, "")),
		etcdadpt.WithPrefix()))

	//删除tags
	opts = append(opts, etcdadpt.OpDel(
		etcdadpt.WithStrKey(path.GenerateServiceTagKey(domainProject, serviceID))))

	//删除instances
	opts = append(opts, etcdadpt.OpDel(
		etcdadpt.WithStrKey(path.GenerateInstanceKey(domainProject, serviceID, "")),
		etcdadpt.WithPrefix()))
	opts = append(opts, etcdadpt.OpDel(
		etcdadpt.WithStrKey(path.GenerateInstanceLeaseKey(domainProject, serviceID, "")),
		etcdadpt.WithPrefix()))

	//删除实例
	err = serviceUtil.DeleteServiceAllInstances(ctx, serviceID)
	if err != nil {
		log.Error(fmt.Sprintf("%s micro-service[%s] failed, revoke all instances failed, operator: %s",
			title, serviceID, remoteIP), err)
		return pb.CreateResponse(pb.ErrUnavailableBackend, err.Error()), err
	}

	resp, err := etcdadpt.TxnWithCmp(ctx, opts, etcdadpt.If(etcdadpt.NotEqualVer(serviceIDKey, 0)), nil)
	if err != nil {
		log.Error(fmt.Sprintf("%s micro-service[%s] failed, operator: %s", title, serviceID, remoteIP), err)
		return pb.CreateResponse(pb.ErrUnavailableBackend, err.Error()), err
	}
	if !resp.Succeeded {
		log.Error(fmt.Sprintf("%s micro-service[%s] failed, service does not exist, operator: %s",
			title, serviceID, remoteIP), err)
		return pb.CreateResponse(pb.ErrServiceNotExists, "Service does not exist."), nil
	}

	serviceUtil.RemandServiceQuota(ctx)

	log.Info(fmt.Sprintf("%s micro-service[%s] successfully, operator: %s", title, serviceID, remoteIP))
	return pb.CreateResponse(pb.ResponseSuccess, "Unregister service successfully."), nil
}

func (ds *MetadataManager) GetDeleteServiceFunc(ctx context.Context, serviceID string, force bool,
	serviceRespChan chan<- *pb.DelServicesRspInfo) func(context.Context) {
	return func(_ context.Context) {
		serviceRst := &pb.DelServicesRspInfo{
			ServiceId:  serviceID,
			ErrMessage: "",
		}
		resp, err := ds.DeleteServicePri(ctx, serviceID, force)
		if err != nil {
			serviceRst.ErrMessage = err.Error()
		} else if resp.GetCode() != pb.ResponseSuccess {
			serviceRst.ErrMessage = resp.GetMessage()
		}

		serviceRespChan <- serviceRst
	}
}
