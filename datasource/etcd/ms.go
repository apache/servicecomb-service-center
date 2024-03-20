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
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/apache/servicecomb-service-center/datasource"
	"github.com/apache/servicecomb-service-center/datasource/etcd/cache"
	"github.com/apache/servicecomb-service-center/datasource/etcd/path"
	"github.com/apache/servicecomb-service-center/datasource/etcd/sd"
	"github.com/apache/servicecomb-service-center/datasource/etcd/state/kvstore"
	esync "github.com/apache/servicecomb-service-center/datasource/etcd/sync"
	eutil "github.com/apache/servicecomb-service-center/datasource/etcd/util"
	serviceUtil "github.com/apache/servicecomb-service-center/datasource/etcd/util"
	"github.com/apache/servicecomb-service-center/datasource/local"
	"github.com/apache/servicecomb-service-center/datasource/schema"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/util"
	"github.com/apache/servicecomb-service-center/server/core"
	"github.com/apache/servicecomb-service-center/server/plugin/uuid"
	quotasvc "github.com/apache/servicecomb-service-center/server/service/quota"
	"github.com/apache/servicecomb-service-center/syncer/service/event"
	pb "github.com/go-chassis/cari/discovery"
	"github.com/go-chassis/cari/pkg/errsvc"
	"github.com/go-chassis/cari/sync"
	"github.com/go-chassis/foundation/gopool"
	"github.com/little-cui/etcdadpt"
)

type MetadataManager struct {
	// InstanceTTL options
	InstanceTTL int64
}

const LOCAL = "local"

// RegisterService implement:
// 1. capsule request to etcd kv format
// 2. invoke etcd client to store data
// 3. check etcd-client response && construct createServiceResponse
func (ds *MetadataManager) RegisterService(ctx context.Context, request *pb.CreateServiceRequest) (
	response *pb.CreateServiceResponse, err error) {
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

	data, err := json.Marshal(service)
	if err != nil {
		log.Error(fmt.Sprintf("create micro-service[%s] failed, json marshal service failed, operator: %s",
			serviceFlag, remoteIP), err)
		return nil, pb.NewError(pb.ErrInternal, err.Error())
	}

	if schema.StorageType == LOCAL {
		contents := make([]*schema.ContentItem, len(service.Schemas))
		err = schema.Instance().PutManyContent(ctx, &schema.PutManyContentRequest{
			ServiceID: service.ServiceId,
			SchemaIDs: service.Schemas,
			Contents:  contents,
			Init:      true,
		})
		if err != nil {
			return nil, err
		}

		serviceMutex := local.GetOrCreateMutex(service.ServiceId)
		serviceMutex.Lock()
		defer serviceMutex.Unlock()
	}

	defer func() {
		if schema.StorageType == LOCAL && err != nil {
			cleanDirErr := local.CleanDir(filepath.Join(schema.RootFilePath, domainProject, service.ServiceId))
			if cleanDirErr != nil {
				log.Error("clean dir error when rollback in RegisterService", cleanDirErr)
			}
		}
	}()

	key := path.GenerateServiceKey(domainProject, service.ServiceId)
	alias := path.GenerateServiceAliasKey(serviceKey)

	opts := []etcdadpt.OpOptions{
		etcdadpt.OpPut(etcdadpt.WithStrKey(key), etcdadpt.WithValue(data)),
		etcdadpt.OpPut(etcdadpt.WithStrKey(index), etcdadpt.WithStrValue(service.ServiceId)),
	}
	uniqueCmpOpts := []etcdadpt.CmpOptions{
		etcdadpt.NotExistKey(key),
		etcdadpt.NotExistKey(index),
	}
	failOpts := []etcdadpt.OpOptions{
		etcdadpt.OpGet(etcdadpt.WithStrKey(index)),
	}

	if len(serviceKey.Alias) > 0 {
		opts = append(opts, etcdadpt.OpPut(etcdadpt.WithStrKey(alias), etcdadpt.WithStrValue(service.ServiceId)))
		uniqueCmpOpts = append(uniqueCmpOpts, etcdadpt.NotExistKey(alias))
		failOpts = append(failOpts, etcdadpt.OpGet(etcdadpt.WithStrKey(alias)))
	}

	syncOpts, err := esync.GenCreateOpts(ctx, datasource.ResourceService, request)
	if err != nil {
		log.Error("fail to create sync opts", err)
		return nil, pb.NewError(pb.ErrInternal, err.Error())
	}
	opts = append(opts, syncOpts...)

	resp, err := etcdadpt.TxnWithCmp(ctx, opts, uniqueCmpOpts, failOpts)
	if err != nil {
		log.Error(fmt.Sprintf("create micro-service[%s] failed, operator: %s",
			serviceFlag, remoteIP), err)
		return nil, pb.NewError(pb.ErrUnavailableBackend, err.Error())
	}

	if resp.Succeeded {
		log.Info(fmt.Sprintf("create micro-service[%s][%s] successfully, operator: %s",
			service.ServiceId, serviceFlag, remoteIP))

		return &pb.CreateServiceResponse{
			ServiceId: service.ServiceId,
		}, nil
	}

	if len(requestServiceID) != 0 {
		if len(resp.Kvs) == 0 || requestServiceID != util.BytesToStringWithNoCopy(resp.Kvs[0].Value) {
			log.Warn(fmt.Sprintf("create micro-service[%s] failed, service already exists, operator: %s",
				serviceFlag, remoteIP))
			return nil, pb.NewError(pb.ErrServiceAlreadyExists,
				"ServiceID conflict or found the same service with different id.")
		}
	}

	if len(resp.Kvs) == 0 {
		// internal error?
		log.Error(fmt.Sprintf("create micro-service[%s] failed, unexpected txn response, operator: %s",
			serviceFlag, remoteIP), nil)
		return nil, pb.NewError(pb.ErrInternal, "Unexpected txn response.")
	}

	existServiceID := util.BytesToStringWithNoCopy(resp.Kvs[0].Value)
	log.Warn(fmt.Sprintf("create micro-service[%s][%s] failed, service already exists, operator: %s",
		existServiceID, serviceFlag, remoteIP))
	return &pb.CreateServiceResponse{
		ServiceId: existServiceID,
	}, nil
}

func (ds *MetadataManager) ListService(ctx context.Context, _ *pb.GetServicesRequest) (
	*pb.GetServicesResponse, error) {
	services, err := eutil.GetAllServiceUtil(ctx)
	if err != nil {
		log.Error("get all services by domain failed", err)
		return nil, pb.NewError(pb.ErrInternal, err.Error())
	}

	return &pb.GetServicesResponse{
		Services: services,
	}, nil
}

func (ds *MetadataManager) GetService(ctx context.Context, request *pb.GetServiceRequest) (
	*pb.MicroService, error) {
	domainProject := util.ParseDomainProject(ctx)
	singleService, err := eutil.GetService(ctx, domainProject, request.ServiceId)

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

func (ds *MetadataManager) GetOverview(ctx context.Context, _ *pb.GetServicesRequest) (
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

	opts := append(eutil.FromContext(ctx),
		etcdadpt.WithStrKey(key),
		etcdadpt.WithPrefix(),
		etcdadpt.WithKeyOnly())

	resp, err := sd.ServiceIndex().Search(ctx, opts...)
	if err != nil {
		return nil, err
	}
	l := len(resp.Kvs)
	if l == 0 {
		return &pb.GetAppsResponse{}, nil
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
		AppIds: apps,
	}, nil
}

func (ds *MetadataManager) ExistServiceByID(ctx context.Context, request *pb.GetExistenceByIDRequest) (*pb.GetExistenceByIDResponse, error) {
	domainProject := util.ParseDomainProject(ctx)
	return &pb.GetExistenceByIDResponse{
		Exist: eutil.ServiceExist(ctx, domainProject, request.ServiceId),
	}, nil
}

func (ds *MetadataManager) ExistService(ctx context.Context, request *pb.GetExistenceRequest) (string, error) {
	domainProject := util.ParseDomainProject(ctx)
	serviceFlag := util.StringJoin([]string{
		request.Environment, request.AppId, request.ServiceName, request.Version}, path.SPLIT)

	ids, exist, err := eutil.FindServiceIds(ctx, &pb.MicroServiceKey{
		Environment: request.Environment,
		AppId:       request.AppId,
		ServiceName: request.ServiceName,
		Alias:       request.ServiceName,
		Version:     request.Version,
		Tenant:      domainProject,
	}, true)
	if err != nil {
		log.Error(fmt.Sprintf("micro-service[%s] exist failed, find serviceIDs failed", serviceFlag), err)
		return "", pb.NewError(pb.ErrInternal, err.Error())
	}
	if !exist {
		log.Info(fmt.Sprintf("micro-service[%s] exist failed, service does not exist", serviceFlag))
		return "", pb.NewError(pb.ErrServiceNotExists, serviceFlag+" does not exist.")
	}
	if len(ids) == 0 {
		log.Info(fmt.Sprintf("micro-service[%s] exist failed, version mismatch", serviceFlag))
		return "", pb.NewError(pb.ErrServiceVersionNotExists, serviceFlag+" version mismatch.")
	}
	// 约定多个时，取较新版本
	return ids[0], nil
}

func (ds *MetadataManager) FindService(ctx context.Context, request *pb.MicroServiceKey) (*pb.GetServicesResponse, error) {
	copyKey := *request
	copyKey.Tenant = util.ParseDomainProject(ctx)
	copyKey.Version = ""
	key := path.GenerateServiceIndexKey(&copyKey)
	opts := append(eutil.FromContext(ctx), etcdadpt.WithStrKey(key), etcdadpt.WithPrefix())

	resp, err := sd.ServiceIndex().Search(ctx, opts...)
	if err != nil {
		return nil, err
	}
	result := &pb.GetServicesResponse{}
	for _, kv := range resp.Kvs {
		svc, err := ds.GetService(ctx, &pb.GetServiceRequest{ServiceId: kv.Value.(string)})
		if err != nil {
			return nil, err
		}
		result.Services = append(result.Services, svc)
	}
	return result, nil
}

func (ds *MetadataManager) PutServiceProperties(ctx context.Context, request *pb.UpdateServicePropsRequest) error {
	remoteIP := util.GetIPFromContext(ctx)
	domainProject := util.ParseDomainProject(ctx)

	key := path.GenerateServiceKey(domainProject, request.ServiceId)
	microservice, err := eutil.GetService(ctx, domainProject, request.ServiceId)
	if err != nil {
		if errors.Is(err, datasource.ErrNoData) {
			log.Debug(fmt.Sprintf("service does not exist, update service[%s] properties failed, operator: %s",
				request.ServiceId, remoteIP))
			return pb.NewError(pb.ErrServiceNotExists, "Service does not exist.")
		}
		log.Error(fmt.Sprintf("update service[%s] properties failed, get service file failed, operator: %s",
			request.ServiceId, remoteIP), err)
		return pb.NewError(pb.ErrInternal, err.Error())
	}

	copyServiceRef := *microservice
	copyServiceRef.Properties = request.Properties
	copyServiceRef.ModTimestamp = strconv.FormatInt(time.Now().Unix(), 10)

	data, err := json.Marshal(copyServiceRef)
	if err != nil {
		log.Error(fmt.Sprintf("update service[%s] properties failed, json marshal service failed, operator: %s",
			request.ServiceId, remoteIP), err)
		return pb.NewError(pb.ErrInternal, err.Error())
	}

	opts := []etcdadpt.OpOptions{
		etcdadpt.OpPut(etcdadpt.WithStrKey(key), etcdadpt.WithValue(data)),
	}
	syncOpts, err := esync.GenUpdateOpts(ctx, datasource.ResourceService, request)
	if err != nil {
		log.Error("fail to create task", err)
		return pb.NewError(pb.ErrInternal, err.Error())
	}
	opts = append(opts, syncOpts...)

	// Set key file
	resp, err := etcdadpt.TxnWithCmp(ctx, opts, etcdadpt.If(etcdadpt.NotEqualVer(key, 0)), nil)
	if err != nil {
		log.Error(fmt.Sprintf("update service[%s] properties failed, operator: %s", request.ServiceId, remoteIP), err)
		return pb.NewError(pb.ErrUnavailableBackend, err.Error())
	}
	if !resp.Succeeded {
		log.Error(fmt.Sprintf("update service[%s] properties failed, service does not exist, operator: %s",
			request.ServiceId, remoteIP), err)
		return pb.NewError(pb.ErrServiceNotExists, "Service does not exist.")
	}

	log.Info(fmt.Sprintf("update service[%s] properties successfully, operator: %s", request.ServiceId, remoteIP))
	return nil
}

func (ds *MetadataManager) RegisterInstance(ctx context.Context, request *pb.RegisterInstanceRequest) (
	*pb.RegisterInstanceResponse, error) {
	instanceID, err := ds.registerInstance(ctx, request)
	if err != nil {
		return nil, err
	}
	return &pb.RegisterInstanceResponse{
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
	sendEvent(ctx, sync.CreateAction, datasource.ResourceInstance, request)
	log.Info(fmt.Sprintf("register instance %s, instanceID %s, operator %s",
		instanceFlag, instanceID, remoteIP))
	return instanceID, nil
}

func sendEvent(ctx context.Context, action string, resourceType string, resource interface{}) {
	if !util.EnableSync(ctx) {
		return
	}
	event.Publish(ctx, action, resourceType, resource)
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
	err := ds.SendHeartbeat(ctx, &pb.HeartbeatRequest{ServiceId: instance.ServiceId,
		InstanceId: instance.InstanceId})
	if err == nil {
		log.Info(fmt.Sprintf("register instance successful, reuse instance[%s/%s], operator %s",
			instance.ServiceId, instance.InstanceId, remoteIP))
		return false, nil
	}

	if errsvc.IsErrEqualCode(err, pb.ErrInstanceNotExists) {
		// register a new one
		return true, nil
	}

	log.Error(fmt.Sprintf("register service[%s]'s instance failed, endpoints %v, host '%s', operator %s",
		instance.ServiceId, instance.Endpoints, instance.HostName, remoteIP), err)
	return false, err
}

func (ds *MetadataManager) ExistInstance(ctx context.Context, request *pb.MicroServiceInstanceKey) (*pb.GetExistenceByIDResponse, error) {
	domainProject := util.ParseDomainProject(ctx)
	exist, err := eutil.ExistInstance(ctx, domainProject, request.ServiceId, request.InstanceId)
	if err != nil {
		return nil, pb.NewError(pb.ErrInternal, err.Error())
	}
	return &pb.GetExistenceByIDResponse{
		Exist: exist,
	}, nil
}

func (ds *MetadataManager) GetInstance(ctx context.Context, request *pb.GetOneInstanceRequest) (
	*pb.GetOneInstanceResponse, error) {
	domainProject := util.ParseDomainProject(ctx)

	service := &pb.MicroService{}
	var err error
	if len(request.ConsumerServiceId) > 0 {
		service, err = eutil.GetService(ctx, domainProject, request.ConsumerServiceId)
		if err != nil {
			if errors.Is(err, datasource.ErrNoData) {
				log.Debug(fmt.Sprintf("consumer does not exist in db, consumer[%s] find provider instance[%s/%s]",
					request.ConsumerServiceId, request.ProviderServiceId, request.ProviderInstanceId))
				return nil, pb.NewError(pb.ErrServiceNotExists,
					fmt.Sprintf("Consumer[%s] does not exist.", request.ConsumerServiceId))
			}
			log.Error(fmt.Sprintf("get consumer failed, consumer[%s] find provider instance[%s/%s]",
				request.ConsumerServiceId, request.ProviderServiceId, request.ProviderInstanceId), err)
			return nil, pb.NewError(pb.ErrInternal, err.Error())
		}
	}

	provider, err := eutil.GetService(ctx, domainProject, request.ProviderServiceId)
	if err != nil {
		if errors.Is(err, datasource.ErrNoData) {
			log.Debug(fmt.Sprintf("provider does not exist in db, consumer[%s] find provider instance[%s/%s]",
				request.ConsumerServiceId, request.ProviderServiceId, request.ProviderInstanceId))
			return nil, pb.NewError(pb.ErrServiceNotExists,
				fmt.Sprintf("Provider[%s] does not exist.", request.ProviderServiceId))
		}
		log.Error(fmt.Sprintf("get provider failed, consumer[%s] find provider instance[%s/%s]",
			request.ConsumerServiceId, request.ProviderServiceId, request.ProviderInstanceId), err)
		return nil, pb.NewError(pb.ErrInternal, err.Error())
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
		return nil, pb.NewError(pb.ErrInternal, err.Error())
	}
	if item == nil || len(item.Instances) == 0 {
		mes := fmt.Errorf("%s failed, provider instance does not exist", findFlag())
		log.Error("find Instances by ProviderID failed", mes)
		return nil, pb.NewError(pb.ErrInstanceNotExists, mes.Error())
	}

	instance := item.Instances[0]
	if rev == item.Rev {
		instance = nil // for gRPC
	}
	_ = util.WithResponseRev(ctx, item.Rev)

	return &pb.GetOneInstanceResponse{
		Instance: instance,
	}, nil
}

func (ds *MetadataManager) ListInstance(ctx context.Context, request *pb.GetInstancesRequest) (*pb.GetInstancesResponse,
	error) {
	domainProject := util.ParseDomainProject(ctx)

	service := &pb.MicroService{}
	var err error
	if len(request.ConsumerServiceId) > 0 {
		service, err = eutil.GetService(ctx, domainProject, request.ConsumerServiceId)
		if err != nil {
			if errors.Is(err, datasource.ErrNoData) {
				log.Debug(fmt.Sprintf("consumer does not exist in db, consumer[%s] find provider[%s] instances",
					request.ConsumerServiceId, request.ProviderServiceId))
				return nil, pb.NewError(pb.ErrServiceNotExists,
					fmt.Sprintf("Consumer[%s] does not exist.", request.ConsumerServiceId))
			}
			log.Error(fmt.Sprintf("get consumer failed, consumer[%s] find provider[%s] instances",
				request.ConsumerServiceId, request.ProviderServiceId), err)
			return nil, pb.NewError(pb.ErrInternal, err.Error())
		}
	}

	provider, err := eutil.GetService(ctx, domainProject, request.ProviderServiceId)
	if err != nil {
		if errors.Is(err, datasource.ErrNoData) {
			log.Debug(fmt.Sprintf("provider does not exist, consumer[%s] find provider[%s] instances",
				request.ConsumerServiceId, request.ProviderServiceId))
			return nil, pb.NewError(pb.ErrServiceNotExists,
				fmt.Sprintf("Provider[%s] does not exist.", request.ProviderServiceId))
		}
		log.Error(fmt.Sprintf("get provider failed, consumer[%s] find provider[%s] instances",
			request.ConsumerServiceId, request.ProviderServiceId), err)
		return nil, pb.NewError(pb.ErrInternal, err.Error())
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
		return nil, pb.NewError(pb.ErrInternal, err.Error())
	}
	if item == nil || len(item.ServiceIds) == 0 {
		err := fmt.Errorf("%s failed, provider instance does not exist", findFlag())
		log.Error("FindInstances.GetWithProviderID failed", err)
		return nil, pb.NewError(pb.ErrServiceNotExists, err.Error())
	}

	instances := item.Instances
	if rev == item.Rev {
		instances = nil // for gRPC
	}
	_ = util.WithResponseRev(ctx, item.Rev)

	return &pb.GetInstancesResponse{
		Instances: instances,
	}, nil
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
		return nil, pb.NewError(pb.ErrInternal, err.Error())
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
		service, err = eutil.GetService(ctx, domainProject, request.ConsumerServiceId)
		if err != nil {
			if errors.Is(err, datasource.ErrNoData) {
				log.Debug(fmt.Sprintf("consumer does not exist, consumer[%s] find provider[%s/%s/%s]",
					request.ConsumerServiceId, request.Environment, request.AppId, request.ServiceName))
				return nil, pb.NewError(pb.ErrServiceNotExists,
					fmt.Sprintf("Consumer[%s] does not exist.", request.ConsumerServiceId))
			}
			log.Error(fmt.Sprintf("get consumer failed, consumer[%s] find provider[%s/%s/%s]",
				request.ConsumerServiceId, request.Environment, request.AppId, request.ServiceName), err)
			return nil, pb.NewError(pb.ErrInternal, err.Error())
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
		return nil, pb.NewError(pb.ErrInternal, err.Error())
	}
	if item == nil {
		err := fmt.Errorf("%s failed, provider does not exist", findFlag)
		log.Error("FindInstancesCache.Get failed", err)
		return nil, pb.NewError(pb.ErrServiceNotExists, err.Error())
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
			err = eutil.AddServiceVersionRule(ctx, domainProject, service, provider)
		} else {
			err := fmt.Errorf("%s failed, provider does not exist", findFlag)
			log.Error("AddServiceVersionRule failed", err)
			return nil, pb.NewError(pb.ErrServiceNotExists, err.Error())
		}
		if err != nil {
			log.Error(fmt.Sprintf("AddServiceVersionRule failed, %s failed", findFlag), err)
			return nil, pb.NewError(pb.ErrInternal, err.Error())
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
		return nil, pb.NewError(pb.ErrInternal, err.Error())
	}
	if item == nil {
		err := fmt.Errorf("%s failed, provider does not exist", findFlag)
		log.Error("FindInstancesCache.Get failed", err)
		return nil, pb.NewError(pb.ErrServiceNotExists, err.Error())
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
		Instances: instances,
	}, nil
}

func (ds *MetadataManager) reshapeProviderKey(ctx context.Context, provider *pb.MicroServiceKey, providerID string) (
	*pb.MicroServiceKey, error) {
	// service name 可能是别名，所以重新获取
	providerService, err := eutil.GetService(ctx, provider.Tenant, providerID)
	if err != nil {
		return nil, err
	}

	provider = pb.MicroServiceToKey(provider.Tenant, providerService)
	provider.Version = datasource.AllVersions // just compatible to old version
	return provider, nil
}

func (ds *MetadataManager) PutInstance(ctx context.Context, request *pb.RegisterInstanceRequest) error {
	domainProject := util.ParseDomainProject(ctx)
	instance := request.Instance
	serviceID := instance.ServiceId
	instanceID := instance.InstanceId
	exist, err := eutil.ExistInstance(ctx, domainProject, serviceID, instanceID)
	if err != nil {
		log.Error(fmt.Sprintf("update instance[%s/%s] failed", serviceID, instanceID), err)
		return pb.NewError(pb.ErrInternal, err.Error())
	}
	if !exist {
		log.Error(fmt.Sprintf("update instance[%s/%s] failed, instance does not exist",
			serviceID, instanceID), nil)
		return pb.NewError(pb.ErrInstanceNotExists, "Service instance does not exist.")
	}

	if err := eutil.UpdateInstance(ctx, domainProject, instance); err != nil {
		log.Error(fmt.Sprintf("update instance[%s/%s] failed", serviceID, instanceID), err)
		return err
	}
	sendEvent(ctx, sync.UpdateAction, datasource.ResourceInstance, instance)
	log.Info(fmt.Sprintf("update instance[%s/%s] successfully", serviceID, instanceID))
	return nil
}

func (ds *MetadataManager) PutInstanceStatus(ctx context.Context, request *pb.UpdateInstanceStatusRequest) error {
	domainProject := util.ParseDomainProject(ctx)
	updateStatusFlag := util.StringJoin([]string{request.ServiceId, request.InstanceId, request.Status}, path.SPLIT)

	instance, err := eutil.GetInstance(ctx, domainProject, request.ServiceId, request.InstanceId)
	if err != nil {
		log.Error(fmt.Sprintf("update instance[%s] status failed", updateStatusFlag), err)
		return pb.NewError(pb.ErrInternal, err.Error())
	}
	if instance == nil {
		log.Error(fmt.Sprintf("update instance[%s] status failed, instance does not exist", updateStatusFlag), nil)
		return pb.NewError(pb.ErrInstanceNotExists, "Service instance does not exist.")
	}

	copyInstanceRef := *instance
	copyInstanceRef.Status = request.Status
	copyInstanceRef.ModTimestamp = strconv.FormatInt(time.Now().Unix(), 10)

	if err := eutil.UpdateInstance(ctx, domainProject, &copyInstanceRef); err != nil {
		log.Error(fmt.Sprintf("update instance[%s] status failed", updateStatusFlag), err)
		return err
	}
	sendEvent(ctx, sync.UpdateAction, datasource.ResourceInstance, copyInstanceRef)
	log.Info(fmt.Sprintf("update instance[%s] status successfully", updateStatusFlag))
	return nil
}

func (ds *MetadataManager) PutInstanceProperties(ctx context.Context, request *pb.UpdateInstancePropsRequest) error {
	domainProject := util.ParseDomainProject(ctx)
	instanceFlag := util.StringJoin([]string{request.ServiceId, request.InstanceId}, path.SPLIT)

	instance, err := eutil.GetInstance(ctx, domainProject, request.ServiceId, request.InstanceId)
	if err != nil {
		log.Error(fmt.Sprintf("update instance[%s] properties failed", instanceFlag), err)
		return pb.NewError(pb.ErrInternal, err.Error())
	}
	if instance == nil {
		log.Error(fmt.Sprintf("update instance[%s] properties failed, instance does not exist", instanceFlag), nil)
		return pb.NewError(pb.ErrInstanceNotExists, "Service instance does not exist.")
	}

	copyInstanceRef := *instance
	copyInstanceRef.Properties = request.Properties
	copyInstanceRef.ModTimestamp = strconv.FormatInt(time.Now().Unix(), 10)

	if err := eutil.UpdateInstance(ctx, domainProject, &copyInstanceRef); err != nil {
		log.Error(fmt.Sprintf("update instance[%s] properties failed", instanceFlag), err)
		return err
	}

	sendEvent(ctx, sync.UpdateAction, datasource.ResourceInstance, copyInstanceRef)
	log.Info(fmt.Sprintf("update instance[%s] properties successfully", instanceFlag))
	return nil
}

func (ds *MetadataManager) SendManyHeartbeat(ctx context.Context, request *pb.HeartbeatSetRequest) (*pb.HeartbeatSetResponse, error) {
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
		}
		existFlag[heartbeatElement.ServiceId+heartbeatElement.InstanceId] = true
		noMultiCounter++
		gopool.Go(getHeartbeatFunc(ctx, domainProject, instancesHbRst, heartbeatElement))
	}
	count := 0
	instanceHbRstArr := make([]*pb.InstanceHbRst, 0, heartBeatCount)
	for heartbeat := range instancesHbRst {
		count++
		instanceHbRstArr = append(instanceHbRstArr, heartbeat)
		sendEvent(ctx, sync.UpdateAction, datasource.ResourceHeartbeat,
			&pb.HeartbeatRequest{ServiceId: heartbeat.ServiceId, InstanceId: heartbeat.InstanceId})
		if count == noMultiCounter {
			close(instancesHbRst)
		}
	}
	log.Info(fmt.Sprintf("batch update heartbeats, %v", instanceHbRstArr))
	return &pb.HeartbeatSetResponse{
		Instances: instanceHbRstArr,
	}, nil
}

func (ds *MetadataManager) UnregisterInstance(ctx context.Context, request *pb.UnregisterInstanceRequest) error {
	remoteIP := util.GetIPFromContext(ctx)
	domainProject := util.ParseDomainProject(ctx)
	serviceID := request.ServiceId
	instanceID := request.InstanceId

	instanceFlag := util.StringJoin([]string{serviceID, instanceID}, path.SPLIT)

	err := revokeInstance(ctx, domainProject, serviceID, instanceID)
	if err != nil {
		log.Error(fmt.Sprintf("unregister instance failed, instance[%s], operator %s: revoke instance failed",
			instanceFlag, remoteIP), err)
		return err
	}
	sendEvent(ctx, sync.DeleteAction, datasource.ResourceInstance, request)
	log.Info(fmt.Sprintf("unregister instance[%s], operator %s", instanceFlag, remoteIP))
	return nil
}

func (ds *MetadataManager) SendHeartbeat(ctx context.Context, request *pb.HeartbeatRequest) error {
	remoteIP := util.GetIPFromContext(ctx)
	domainProject := util.ParseDomainProject(ctx)
	instanceFlag := util.StringJoin([]string{request.ServiceId, request.InstanceId}, path.SPLIT)

	_, ttl, err := eutil.HeartbeatUtil(ctx, domainProject, request.ServiceId, request.InstanceId)
	if err != nil {
		log.Error(fmt.Sprintf("heartbeat failed, instance[%s]. operator %s",
			instanceFlag, remoteIP), err)
		return err
	}

	if ttl == 0 {
		log.Error(fmt.Sprintf("heartbeat successful, but renew instance[%s] failed. operator %s", instanceFlag, remoteIP),
			errors.New("connect backend timed out"))
	} else {
		log.Info(fmt.Sprintf("heartbeat successful, renew instance[%s] ttl to %d. operator %s",
			instanceFlag, ttl, remoteIP))
	}
	sendEvent(ctx, sync.UpdateAction, datasource.ResourceHeartbeat, request)
	return nil
}

func (ds *MetadataManager) ListManyInstances(ctx context.Context, _ *pb.GetAllInstancesRequest) (*pb.GetAllInstancesResponse, error) {
	domainProject := util.ParseDomainProject(ctx)
	key := path.GetInstanceRootKey(domainProject) + path.SPLIT
	opts := append(eutil.FromContext(ctx), etcdadpt.WithStrKey(key), etcdadpt.WithPrefix())
	kvs, err := sd.Instance().Search(ctx, opts...)
	if err != nil {
		return nil, err
	}
	resp := &pb.GetAllInstancesResponse{}
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

	serviceInfo, err := eutil.GetService(ctx, domainProject, serviceID)
	if err != nil {
		if errors.Is(err, datasource.ErrNoData) {
			log.Debug(fmt.Sprintf("modify service[%s] schemas failed, service does not exist in db, operator: %s",
				serviceID, remoteIP))
			return nil, pb.NewError(pb.ErrServiceNotExists, "Service does not exist.")
		}
		log.Error(fmt.Sprintf("modify service[%s] schemas failed, get service failed, operator: %s",
			serviceID, remoteIP), err)
		return nil, pb.NewError(pb.ErrInternal, err.Error())
	}

	if respErr := ds.modifySchemas(ctx, domainProject, serviceInfo, request.Schemas); respErr != nil {
		log.Error(fmt.Sprintf("modify service[%s] schemas failed, operator: %s", serviceID, remoteIP), respErr)
		return nil, respErr
	}

	return &pb.ModifySchemasResponse{}, nil
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
		return nil, err
	}

	log.Info(fmt.Sprintf("modify schema[%s/%s] successfully, operator: %s", serviceID, schemaID, remoteIP))
	return &pb.ModifySchemaResponse{}, nil
}

func (ds *MetadataManager) ExistSchema(ctx context.Context, request *pb.GetExistenceRequest) (
	*pb.GetExistenceResponse, error) {
	domainProject := util.ParseDomainProject(ctx)

	if !eutil.ServiceExist(ctx, domainProject, request.ServiceId) {
		log.Warn(fmt.Sprintf("schema[%s/%s] exist failed, service does not exist", request.ServiceId, request.SchemaId))
		return nil, pb.NewError(pb.ErrServiceNotExists, "service does not exist.")
	}

	key := path.GenerateServiceSchemaKey(domainProject, request.ServiceId, request.SchemaId)
	exist, err := checkSchemaInfoExist(ctx, key)
	if err != nil {
		log.Error(fmt.Sprintf("schema[%s/%s] exist failed, get schema failed", request.ServiceId, request.SchemaId), err)
		return nil, pb.NewError(pb.ErrInternal, err.Error())
	}
	if !exist {
		log.Info(fmt.Sprintf("schema[%s/%s] exist failed, schema does not exist", request.ServiceId, request.SchemaId))
		return nil, schema.ErrSchemaNotFound
	}
	schemaSummary, err := getSchemaSummary(ctx, domainProject, request.ServiceId, request.SchemaId)
	if err != nil {
		log.Error(fmt.Sprintf("schema[%s/%s] exist failed, get schema summary failed",
			request.ServiceId, request.SchemaId), err)
		return nil, pb.NewError(pb.ErrInternal, err.Error())
	}
	return &pb.GetExistenceResponse{
		SchemaId: request.SchemaId,
		Summary:  schemaSummary,
	}, nil
}

func (ds *MetadataManager) GetSchema(ctx context.Context, request *pb.GetSchemaRequest) (*pb.GetSchemaResponse, error) {
	domainProject := util.ParseDomainProject(ctx)

	if !eutil.ServiceExist(ctx, domainProject, request.ServiceId) {
		log.Error(fmt.Sprintf("get schema[%s/%s] failed, service does not exist",
			request.ServiceId, request.SchemaId), nil)
		return nil, pb.NewError(pb.ErrServiceNotExists, "Service does not exist.")
	}

	key := path.GenerateServiceSchemaKey(domainProject, request.ServiceId, request.SchemaId)
	opts := append(eutil.FromContext(ctx), etcdadpt.WithStrKey(key))
	resp, errDo := sd.Schema().Search(ctx, opts...)
	if errDo != nil {
		log.Error(fmt.Sprintf("get schema[%s/%s] failed", request.ServiceId, request.SchemaId), errDo)
		return nil, pb.NewError(pb.ErrUnavailableBackend, errDo.Error())
	}
	if resp.Count == 0 {
		log.Error(fmt.Sprintf("get schema[%s/%s] failed, schema does not exists",
			request.ServiceId, request.SchemaId), errDo)
		return nil, schema.ErrSchemaNotFound
	}

	schemaSummary, err := getSchemaSummary(ctx, domainProject, request.ServiceId, request.SchemaId)
	if err != nil {
		log.Error(fmt.Sprintf("get schema[%s/%s] failed, get schema summary failed",
			request.ServiceId, request.SchemaId), err)
		return nil, pb.NewError(pb.ErrInternal, err.Error())
	}

	return &pb.GetSchemaResponse{
		Schema:        util.BytesToStringWithNoCopy(resp.Kvs[0].Value.([]byte)),
		SchemaSummary: schemaSummary,
	}, nil
}

func (ds *MetadataManager) GetAllSchemas(ctx context.Context, request *pb.GetAllSchemaRequest) (
	*pb.GetAllSchemaResponse, error) {
	domainProject := util.ParseDomainProject(ctx)

	service, err := eutil.GetService(ctx, domainProject, request.ServiceId)
	if err != nil {
		if errors.Is(err, datasource.ErrNoData) {
			log.Debug(fmt.Sprintf("get service[%s] all schemas failed, service does not exist in db", request.ServiceId))
			return nil, pb.NewError(pb.ErrServiceNotExists, "Service does not exist.")
		}
		log.Error(fmt.Sprintf("get service[%s] all schemas failed, get service failed", request.ServiceId), err)
		return nil, pb.NewError(pb.ErrInternal, err.Error())
	}

	schemasList := service.Schemas
	if len(schemasList) == 0 {
		return &pb.GetAllSchemaResponse{
			Schemas: []*pb.Schema{},
		}, nil
	}

	key := path.GenerateServiceSchemaSummaryKey(domainProject, request.ServiceId, "")
	opts := append(eutil.FromContext(ctx), etcdadpt.WithStrKey(key), etcdadpt.WithPrefix())
	resp, errDo := sd.SchemaSummary().Search(ctx, opts...)
	if errDo != nil {
		log.Error(fmt.Sprintf("get service[%s] all schema summaries failed", request.ServiceId), errDo)
		return nil, pb.NewError(pb.ErrUnavailableBackend, errDo.Error())
	}

	respWithSchema := &kvstore.Response{}
	if request.WithSchema {
		key := path.GenerateServiceSchemaKey(domainProject, request.ServiceId, "")
		opts := append(eutil.FromContext(ctx), etcdadpt.WithStrKey(key), etcdadpt.WithPrefix())
		respWithSchema, errDo = sd.Schema().Search(ctx, opts...)
		if errDo != nil {
			log.Error(fmt.Sprintf("get service[%s] all schemas failed", request.ServiceId), errDo)
			return nil, pb.NewError(pb.ErrUnavailableBackend, errDo.Error())
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
		Schemas: schemas,
	}, nil
}

func (ds *MetadataManager) DeleteSchema(ctx context.Context, request *pb.DeleteSchemaRequest) error {
	remoteIP := util.GetIPFromContext(ctx)
	domainProject := util.ParseDomainProject(ctx)

	key := path.GenerateServiceSchemaKey(domainProject, request.ServiceId, request.SchemaId)
	exist, err := eutil.CheckSchemaInfoExist(ctx, key)
	if err != nil {
		log.Error(fmt.Sprintf("delete schema[%s/%s] failed, operator: %s",
			request.ServiceId, request.SchemaId, remoteIP), err)
		return pb.NewError(pb.ErrInternal, err.Error())
	}
	if !exist {
		log.Error(fmt.Sprintf("delete schema[%s/%s] failed, schema does not exist, operator: %s",
			request.ServiceId, request.SchemaId, remoteIP), nil)
		return schema.ErrSchemaNotFound
	}
	epSummaryKey := path.GenerateServiceSchemaSummaryKey(domainProject, request.ServiceId, request.SchemaId)
	opts := []etcdadpt.OpOptions{etcdadpt.OpDel(etcdadpt.WithStrKey(epSummaryKey)), etcdadpt.OpDel(etcdadpt.WithStrKey(key))}
	schemaKeyOpt, err := esync.GenDeleteOpts(ctx, datasource.ResourceKV, key, key)
	if err != nil {
		log.Error("fail to create delete opts", err)
		return err
	}
	opts = append(opts, schemaKeyOpt...)
	schemaSummaryKeyOpt, err := esync.GenDeleteOpts(ctx, datasource.ResourceKV, epSummaryKey, epSummaryKey)
	if err != nil {
		log.Error("fail to create delete opts", err)
		return err
	}
	opts = append(opts, schemaSummaryKeyOpt...)

	resp, errDo := etcdadpt.TxnWithCmp(ctx, opts,
		etcdadpt.If(etcdadpt.NotEqualVer(path.GenerateServiceKey(domainProject, request.ServiceId), 0)),
		nil)

	if errDo != nil {
		log.Error(fmt.Sprintf("delete schema[%s/%s] failed, operator: %s",
			request.ServiceId, request.SchemaId, remoteIP), errDo)
		return pb.NewError(pb.ErrUnavailableBackend, errDo.Error())
	}
	if !resp.Succeeded {
		log.Error(fmt.Sprintf("delete schema[%s/%s] failed, service does not exist, operator: %s",
			request.ServiceId, request.SchemaId, remoteIP), nil)
		return pb.NewError(pb.ErrServiceNotExists, "Service does not exist.")
	}

	log.Info(fmt.Sprintf("delete schema[%s/%s] info successfully, operator: %s",
		request.ServiceId, request.SchemaId, remoteIP))
	return nil
}

func (ds *MetadataManager) PutManyTags(ctx context.Context, request *pb.AddServiceTagsRequest) error {
	remoteIP := util.GetIPFromContext(ctx)
	domainProject := util.ParseDomainProject(ctx)

	// service id存在性校验
	if !eutil.ServiceExist(ctx, domainProject, request.ServiceId) {
		log.Error(fmt.Sprintf("add service[%s]'s tags %v failed, service does not exist, operator: %s",
			request.ServiceId, request.Tags, remoteIP), nil)
		return pb.NewError(pb.ErrServiceNotExists, "Service does not exist.")
	}

	checkErr := eutil.AddTagIntoETCD(ctx, domainProject, request.ServiceId, request.Tags)
	if checkErr != nil {
		log.Error(fmt.Sprintf("add service[%s]'s tags %v failed, operator: %s",
			request.ServiceId, request.Tags, remoteIP), checkErr)
		return checkErr
	}

	log.Info(fmt.Sprintf("add service[%s]'s tags %v successfully, operator: %s", request.ServiceId, request.Tags, remoteIP))
	return nil
}

func (ds *MetadataManager) ListTag(ctx context.Context, request *pb.GetServiceTagsRequest) (*pb.GetServiceTagsResponse, error) {
	var err error
	domainProject := util.ParseDomainProject(ctx)
	if !eutil.ServiceExist(ctx, domainProject, request.ServiceId) {
		log.Error(fmt.Sprintf("get service[%s]'s tags failed, service does not exist", request.ServiceId), err)
		return nil, pb.NewError(pb.ErrServiceNotExists, "Service does not exist.")
	}
	tags, err := eutil.GetTagsUtils(ctx, domainProject, request.ServiceId)
	if err != nil {
		log.Error(fmt.Sprintf("get service[%s]'s tags failed, get tags failed", request.ServiceId), err)
		return nil, pb.NewError(pb.ErrInternal, err.Error())
	}

	return &pb.GetServiceTagsResponse{
		Tags: tags,
	}, nil
}

func (ds *MetadataManager) PutTag(ctx context.Context, request *pb.UpdateServiceTagRequest) error {
	var err error
	remoteIP := util.GetIPFromContext(ctx)
	tagFlag := util.StringJoin([]string{request.Key, request.Value}, path.SPLIT)
	domainProject := util.ParseDomainProject(ctx)

	if !eutil.ServiceExist(ctx, domainProject, request.ServiceId) {
		log.Error(fmt.Sprintf("update service[%s]'s tag[%s] failed, service does not exist, operator: %s",
			request.ServiceId, tagFlag, remoteIP), err)
		return pb.NewError(pb.ErrServiceNotExists, "Service does not exist.")
	}

	tags, err := eutil.GetTagsUtils(ctx, domainProject, request.ServiceId)
	if err != nil {
		log.Error(fmt.Sprintf("update service[%s]'s tag[%s] failed, get tag failed, operator: %s",
			request.ServiceId, tagFlag, remoteIP), err)
		return pb.NewError(pb.ErrInternal, err.Error())
	}

	//check if the tag exists
	if _, ok := tags[request.Key]; !ok {
		log.Error(fmt.Sprintf("update service[%s]'s tag[%s] failed, tag does not exist, operator: %s",
			request.ServiceId, tagFlag, remoteIP), nil)
		return pb.NewError(pb.ErrTagNotExists, "Tag does not exist, please add one first.")
	}

	copyTags := make(map[string]string, len(tags))
	for k, v := range tags {
		copyTags[k] = v
	}
	copyTags[request.Key] = request.Value

	checkErr := eutil.AddTagIntoETCD(ctx, domainProject, request.ServiceId, copyTags)
	if checkErr != nil {
		log.Error(fmt.Sprintf("update service[%s]'s tag[%s] failed, operator: %s",
			request.ServiceId, tagFlag, remoteIP), checkErr)
		return checkErr
	}

	log.Info(fmt.Sprintf("update service[%s]'s tag[%s] successfully, operator: %s", request.ServiceId, tagFlag, remoteIP))
	return nil
}

func (ds *MetadataManager) DeleteManyTags(ctx context.Context, request *pb.DeleteServiceTagsRequest) error {
	remoteIP := util.GetIPFromContext(ctx)
	domainProject := util.ParseDomainProject(ctx)

	if !eutil.ServiceExist(ctx, domainProject, request.ServiceId) {
		log.Error(fmt.Sprintf("delete service[%s]'s tags %v failed, service does not exist, operator: %s",
			request.ServiceId, request.Keys, remoteIP), nil)
		return pb.NewError(pb.ErrServiceNotExists, "Service does not exist.")
	}

	tags, err := eutil.GetTagsUtils(ctx, domainProject, request.ServiceId)
	if err != nil {
		log.Error(fmt.Sprintf("delete service[%s]'s tags %v failed, get service tags failed, operator: %s",
			request.ServiceId, request.Keys, remoteIP), err)
		return pb.NewError(pb.ErrInternal, err.Error())
	}

	copyTags := make(map[string]string, len(tags))
	for k, v := range tags {
		copyTags[k] = v
	}
	for _, key := range request.Keys {
		if _, ok := copyTags[key]; !ok {
			log.Error(fmt.Sprintf("delete service[%s]'s tags %v failed, tag[%s] does not exist, operator: %s",
				request.ServiceId, request.Keys, key, remoteIP), nil)
			return pb.NewError(pb.ErrTagNotExists, "Delete tags failed for this key "+key+" does not exist.")
		}
		delete(copyTags, key)
	}

	// the capacity of tags may be 0
	data, err := json.Marshal(copyTags)
	if err != nil {
		log.Error(fmt.Sprintf("delete service[%s]'s tags %v failed, marshall service tags failed, operator: %s",
			request.ServiceId, request.Keys, remoteIP), err)
		return pb.NewError(pb.ErrInternal, err.Error())
	}

	key := path.GenerateServiceTagKey(domainProject, request.ServiceId)

	opts := etcdadpt.Ops(etcdadpt.OpPut(etcdadpt.WithStrKey(key), etcdadpt.WithValue(data)))
	syncOpts, err := esync.GenDeleteOpts(ctx, datasource.ResourceKV, key, data,
		esync.WithOpts(map[string]string{"key": key}))
	if err != nil {
		return pb.NewError(pb.ErrInternal, err.Error())
	}
	opts = append(opts, syncOpts...)

	resp, err := etcdadpt.TxnWithCmp(ctx, opts,
		etcdadpt.If(etcdadpt.NotEqualVer(path.GenerateServiceKey(domainProject, request.ServiceId), 0)), nil)
	if err != nil {
		log.Error(fmt.Sprintf("delete service[%s]'s tags %v failed, operator: %s",
			request.ServiceId, request.Keys, remoteIP), err)
		return pb.NewError(pb.ErrUnavailableBackend, err.Error())
	}
	if !resp.Succeeded {
		log.Error(fmt.Sprintf("delete service[%s]'s tags %v failed, service does not exist, operator: %s",
			request.ServiceId, request.Keys, remoteIP), err)
		return pb.NewError(pb.ErrServiceNotExists, "Service does not exist.")
	}

	log.Info(fmt.Sprintf("delete service[%s]'s tags %v successfully, operator: %s", request.ServiceId, request.Keys, remoteIP))
	return nil
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

	needUpdateSchemas, needAddSchemas, needDeleteSchemas, _ :=
		datasource.SchemasAnalysis(schemas, schemasFromDatabase, service.Schemas)

	pluginOps := make([]etcdadpt.OpOptions, 0)
	quotaSize := len(needAddSchemas) - len(needDeleteSchemas)
	if quotaSize > 0 {
		errQuota := quotasvc.ApplySchema(ctx, serviceID, int64(quotaSize))
		if errQuota != nil {
			log.Error(fmt.Sprintf("modify service[%s] schemas failed, operator: %s", serviceID, remoteIP), errQuota)
			return errQuota
		}
	}

	var schemaIDs []string
	for _, schema := range needAddSchemas {
		log.Info(fmt.Sprintf("add new schema[%s/%s], operator: %s", serviceID, schema.SchemaId, remoteIP))
		opts, _ := putSchema(ctx, domainProject, service.ServiceId, schema)
		pluginOps = append(pluginOps, opts...)
		schemaIDs = append(schemaIDs, schema.SchemaId)
	}

	for _, schema := range needUpdateSchemas {
		log.Info(fmt.Sprintf("update schema[%s/%s], operator: %s", serviceID, schema.SchemaId, remoteIP))
		opts, _ := putSchema(ctx, domainProject, serviceID, schema)
		pluginOps = append(pluginOps, opts...)
		schemaIDs = append(schemaIDs, schema.SchemaId)
	}

	for _, schema := range needDeleteSchemas {
		log.Info(fmt.Sprintf("delete non-existent schema[%s/%s], operator: %s", serviceID, schema.SchemaId, remoteIP))
		opts, _ := deleteSchema(ctx, domainProject, serviceID, schema)
		pluginOps = append(pluginOps, opts...)
	}

	service.Schemas = schemaIDs
	opts, err := eutil.UpdateService(ctx, domainProject, serviceID, service)
	if err != nil {
		log.Error(fmt.Sprintf("modify service[%s] schemas failed, update service.Schemas failed, operator: %s",
			serviceID, remoteIP), err)
		return pb.NewError(pb.ErrInternal, err.Error())
	}
	pluginOps = append(pluginOps, opts...)

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

func (ds *MetadataManager) modifySchema(ctx context.Context, serviceID string, schema *pb.Schema) *errsvc.Error {
	remoteIP := util.GetIPFromContext(ctx)
	domainProject := util.ParseDomainProject(ctx)
	schemaID := schema.SchemaId

	microService, err := eutil.GetService(ctx, domainProject, serviceID)
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
	if !isExist {
		microService.Schemas = append(microService.Schemas, schemaID)
		opts, err := eutil.UpdateService(ctx, domainProject, serviceID, microService)
		if err != nil {
			log.Error(fmt.Sprintf("modify schema[%s/%s] failed, update microService.Schemas failed, operator: %s",
				serviceID, schemaID, remoteIP), err)
			return pb.NewError(pb.ErrInternal, err.Error())
		}
		pluginOps = append(pluginOps, opts...)
	}

	opts, err := commitSchemaInfo(ctx, domainProject, serviceID, schema)
	if err != nil {
		return pb.NewError(pb.ErrInternal, err.Error())
	}
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

func (ds *MetadataManager) UnregisterService(ctx context.Context, request *pb.DeleteServiceRequest) (err error) {
	serviceID := request.ServiceId
	force := request.Force
	remoteIP := util.GetIPFromContext(ctx)
	domainProject := util.ParseDomainProject(ctx)

	title := "delete"
	if force {
		title = "force delete"
	}

	if serviceID == core.Service.ServiceId {
		err := errors.New("not allow to delete service center")
		log.Error(fmt.Sprintf("%s micro-service[%s] failed, operator: %s", title, serviceID, remoteIP), err)
		return pb.NewError(pb.ErrInvalidParams, err.Error())
	}

	// try to delete schema files
	if schema.StorageType == LOCAL {
		tmpPath := filepath.Join(schema.RootFilePath, "tmp", domainProject, serviceID)
		originPath := filepath.Join(schema.RootFilePath, domainProject, serviceID)

		err = local.MoveDir(originPath, tmpPath)
		if err != nil {
			log.Error(fmt.Sprintf("%s micro-service[%s] failed, clean local schmea dir failed, operator: %s",
				title, serviceID, remoteIP), err)
			return err
		}

		serviceMutex := local.GetOrCreateMutex(serviceID)
		serviceMutex.Lock()
		defer serviceMutex.Unlock()
	}

	defer func() {
		if schema.StorageType == LOCAL {
			tmpPath := filepath.Join(schema.RootFilePath, "tmp", domainProject, serviceID)
			originPath := filepath.Join(schema.RootFilePath, domainProject, serviceID)
			var rollbackErr error
			if err != nil {
				rollbackErr = local.MoveDir(tmpPath, originPath)
				if rollbackErr != nil {
					log.Error("clean dir error when rollback in UnregisterService", err)
				}
			} else {
				rollbackErr = local.CleanDir(tmpPath)
				if rollbackErr != nil {
					log.Error("clean tmp dir error when rollback in UnregisterService", err)
				}
				rollbackErr = os.Remove(originPath)
				if rollbackErr != nil {
					log.Error("clean origin dir error when rollback in UnregisterService", err)
				}
			}
		}
	}()

	microservice, err := eutil.GetService(ctx, domainProject, serviceID)
	if err != nil {
		if errors.Is(err, datasource.ErrNoData) {
			log.Debug(fmt.Sprintf("service does not exist, %s micro-service[%s] failed, operator: %s",
				title, serviceID, remoteIP))
			return pb.NewError(pb.ErrServiceNotExists, "Service does not exist.")
		}
		log.Error(fmt.Sprintf("%s micro-service[%s] failed, get service file failed, operator: %s",
			title, serviceID, remoteIP), err)
		return pb.NewError(pb.ErrInternal, err.Error())
	}

	// 强制删除，则与该服务相关的信息删除，非强制删除： 如果作为该被依赖（作为provider，提供服务,且不是只存在自依赖）或者存在实例，则不能删除
	if !force {
		services, err := eutil.GetConsumerIds(ctx, domainProject, microservice)
		if err != nil {
			log.Error(fmt.Sprintf("delete micro-service[%s] failed, get service dependency failed, operator: %s",
				serviceID, remoteIP), err)
			return pb.NewError(pb.ErrInternal, err.Error())
		}
		if l := len(services); l > 1 || (l == 1 && services[0] != serviceID) {
			log.Error(fmt.Sprintf("delete micro-service[%s] failed, other services[%d] depend on it, operator: %s",
				serviceID, l, remoteIP), nil)
			return pb.NewError(pb.ErrDependedOnConsumer, "Can not delete this service, other service rely it.")
		}

		instancesKey := path.GenerateInstanceKey(domainProject, serviceID, "")
		rsp, err := sd.Instance().Search(ctx,
			etcdadpt.WithStrKey(instancesKey),
			etcdadpt.WithPrefix(),
			etcdadpt.WithCountOnly())
		if err != nil {
			log.Error(fmt.Sprintf("delete micro-service[%s] failed, get instances failed, operator: %s",
				serviceID, remoteIP), err)
			return pb.NewError(pb.ErrUnavailableBackend, err.Error())
		}

		if rsp.Count > 0 {
			log.Error(fmt.Sprintf("delete micro-service[%s] failed, service deployed instances[%d], operator: %s",
				serviceID, rsp.Count, remoteIP), nil)
			return pb.NewError(pb.ErrDeployedInstance, "Can not delete the service deployed instance(s).")
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

	syncOpts, err := esync.GenDeleteOpts(ctx, datasource.ResourceService, serviceID,
		&pb.DeleteServiceRequest{ServiceId: serviceID, Force: force})
	if err != nil {
		log.Error("fail to sync opt", err)
		return pb.NewError(pb.ErrInternal, err.Error())
	}
	opts = append(opts, syncOpts...)

	//删除依赖规则
	optDeleteDep, err := eutil.DeleteDependencyForDeleteService(domainProject, serviceID, serviceKey)
	if err != nil {
		log.Error(fmt.Sprintf("%s micro-service[%s] failed, delete dependency failed, operator: %s",
			title, serviceID, remoteIP), err)
		return pb.NewError(pb.ErrInternal, err.Error())
	}
	opts = append(opts, optDeleteDep)

	// 删除schemas
	opts = append(opts, etcdadpt.OpDel(
		etcdadpt.WithStrKey(path.GenerateServiceSchemaKey(domainProject, serviceID, "")),
		etcdadpt.WithPrefix()))
	opts = append(opts, etcdadpt.OpDel(
		etcdadpt.WithStrKey(path.GenerateServiceSchemaSummaryKey(domainProject, serviceID, "")),
		etcdadpt.WithPrefix()))
	opts = append(opts, etcdadpt.OpDel(
		etcdadpt.WithStrKey(path.GenerateServiceSchemaRefKey(domainProject, serviceID, "")),
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
	err = eutil.DeleteServiceAllInstances(ctx, serviceID)
	if err != nil {
		log.Error(fmt.Sprintf("%s micro-service[%s] failed, revoke all instances failed, operator: %s",
			title, serviceID, remoteIP), err)
		return pb.NewError(pb.ErrUnavailableBackend, err.Error())
	}

	resp, err := etcdadpt.TxnWithCmp(ctx, opts, etcdadpt.If(etcdadpt.NotEqualVer(serviceIDKey, 0)), nil)
	if err != nil {
		log.Error(fmt.Sprintf("%s micro-service[%s] failed, operator: %s", title, serviceID, remoteIP), err)
		return pb.NewError(pb.ErrUnavailableBackend, err.Error())
	}
	if !resp.Succeeded {
		log.Error(fmt.Sprintf("%s micro-service[%s] failed, service does not exist, operator: %s",
			title, serviceID, remoteIP), err)
		return pb.NewError(pb.ErrServiceNotExists, "Service does not exist.")
	}

	quotasvc.RemandService(ctx)

	log.Info(fmt.Sprintf("%s micro-service[%s] successfully, operator: %s", title, serviceID, remoteIP))
	return nil
}

func (ds *MetadataManager) Statistics(ctx context.Context, withShared bool) (*pb.Statistics, error) {
	return statistics(ctx, withShared)
}

func (ds *MetadataManager) UpdateManyInstanceStatus(ctx context.Context, match *datasource.MatchPolicy, status string) error {
	resp, _ := ds.ListManyInstances(ctx, &pb.GetAllInstancesRequest{})
	instances := resp.Instances
	if len(instances) == 0 {
		return nil
	}
	options := make([]etcdadpt.OpOptions, 0)
	cmps := make([]etcdadpt.CmpOptions, 0)

	domainProject := util.ParseDomainProject(ctx)

	for _, instance := range instances {
		var t = true
		for k, v := range match.Properties {
			value, ok := instance.Properties[k]
			if ok {
				if value != v {
					t = false
					break
				}
			} else {
				t = false
				break
			}
		}
		if t {
			key := path.GenerateInstanceKey(domainProject, instance.ServiceId, instance.InstanceId)
			//更新状态
			instance.Status = status
			data, _ := json.Marshal(instance)
			leaseID, err := serviceUtil.GetLeaseID(ctx, domainProject, instance.ServiceId, instance.InstanceId)
			if err != nil {
				log.Error(fmt.Sprintf("get leaseId %s error", instance.InstanceId), err)
				continue
			}
			options = append(options, etcdadpt.Ops(etcdadpt.OpPut(etcdadpt.WithStrKey(key), etcdadpt.WithValue(data), etcdadpt.WithLease(leaseID)))...)
			cmps = append(cmps, etcdadpt.If(etcdadpt.NotEqualVer(path.GenerateServiceKey(domainProject, instance.ServiceId), 0))...)
		}
	}
	_, err := etcdadpt.TxnWithCmp(ctx,
		options,
		cmps,
		nil)

	if err != nil {
		log.Error("UpdateManyInstanceStatus error", err)

		return pb.NewError(pb.ErrUnavailableBackend, err.Error())
	}
	return nil
}
