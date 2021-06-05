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
	"crypto/sha1"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"time"

	"github.com/go-chassis/cari/discovery"
	"github.com/go-chassis/cari/pkg/errsvc"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/apache/servicecomb-service-center/datasource"
	"github.com/apache/servicecomb-service-center/datasource/cache"
	"github.com/apache/servicecomb-service-center/datasource/mongo/client"
	"github.com/apache/servicecomb-service-center/datasource/mongo/event"
	"github.com/apache/servicecomb-service-center/datasource/mongo/model"
	"github.com/apache/servicecomb-service-center/datasource/mongo/service/heartbeat"
	mutil "github.com/apache/servicecomb-service-center/datasource/mongo/util"
	"github.com/apache/servicecomb-service-center/pkg/gopool"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/util"
	apt "github.com/apache/servicecomb-service-center/server/core"
	"github.com/apache/servicecomb-service-center/server/plugin/uuid"
)

type InstanceSlice []*discovery.MicroServiceInstance

func (s InstanceSlice) Len() int {
	return len(s)
}

func (s InstanceSlice) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s InstanceSlice) Less(i, j int) bool {
	return s[i].InstanceId < s[j].InstanceId
}

// Instance management
func (ds *DataSource) RegisterInstance(ctx context.Context, request *discovery.RegisterInstanceRequest) (*discovery.RegisterInstanceResponse, error) {
	isCustomID := true
	if len(request.Instance.InstanceId) == 0 {
		isCustomID = false
		request.Instance.InstanceId = uuid.Generator().GetInstanceID(ctx)
	}

	// if queueSize is more than 0 and channel is not full, then do fast register instance
	if fastRegConfig.QueueSize > 0 && len(GetFastRegisterInstanceService().InstEventCh) < fastRegConfig.QueueSize {
		// fast register, just add instance to channel and batch register them later
		event := &event.InstanceRegisterEvent{Ctx: ctx, Request: request, IsCustomID: isCustomID, FailedTime: 0}
		GetFastRegisterInstanceService().AddEvent(event)

		return &discovery.RegisterInstanceResponse{
			Response:   discovery.CreateResponse(discovery.ResponseSuccess, "Register service instance successfully."),
			InstanceId: request.Instance.InstanceId,
		}, nil
	}

	return registerInstanceSingle(ctx, request, isCustomID)
}

func (ds *DataSource) ExistInstanceByID(ctx context.Context, request *discovery.MicroServiceInstanceKey) (*discovery.GetExistenceByIDResponse, error) {
	filter := mutil.NewBasicFilter(ctx, mutil.InstanceServiceID(request.ServiceId), mutil.InstanceInstanceID(request.InstanceId))
	exist, _ := existInstance(ctx, filter)
	if !exist {
		return &discovery.GetExistenceByIDResponse{
			Response: discovery.CreateResponse(discovery.ErrInstanceNotExists, "check instance exist failed."),
			Exist:    false,
		}, datasource.ErrInstanceNotExists
	}

	return &discovery.GetExistenceByIDResponse{
		Response: discovery.CreateResponse(discovery.ResponseSuccess, "check service exists successfully."),
		Exist:    exist,
	}, nil
}

// GetInstance returns instance under the current domain
func (ds *DataSource) GetInstance(ctx context.Context, request *discovery.GetOneInstanceRequest) (*discovery.GetOneInstanceResponse, error) {
	var service *model.Service
	var err error
	var serviceIDs []string
	domain := util.ParseDomain(ctx)
	project := util.ParseProject(ctx)
	if len(request.ConsumerServiceId) > 0 {
		filter := mutil.NewDomainProjectFilter(domain, project, mutil.ServiceServiceID(request.ConsumerServiceId))
		service, err = findService(ctx, filter)
		if err != nil {
			if errors.Is(err, datasource.ErrNoData) {
				log.Debug(fmt.Sprintf("consumer does not exist, consumer %s find provider instance %s %s",
					request.ConsumerServiceId, request.ProviderServiceId, request.ProviderInstanceId))
				return &discovery.GetOneInstanceResponse{
					Response: discovery.CreateResponse(discovery.ErrServiceNotExists,
						fmt.Sprintf("Consumer[%s] does not exist.", request.ConsumerServiceId)),
				}, nil
			}
			log.Error(fmt.Sprintf(" get consumer failed, consumer %s find provider instance %s",
				request.ConsumerServiceId, request.ProviderInstanceId), err)
			return &discovery.GetOneInstanceResponse{
				Response: discovery.CreateResponse(discovery.ErrInternal, err.Error()),
			}, err
		}
	}
	filter := mutil.NewDomainProjectFilter(domain, project, mutil.ServiceServiceID(request.ProviderServiceId))
	provider, err := findService(ctx, filter)
	if err != nil {
		if errors.Is(err, datasource.ErrNoData) {
			log.Debug(fmt.Sprintf("provider does not exist, consumer %s find provider instance %s %s",
				request.ConsumerServiceId, request.ProviderServiceId, request.ProviderInstanceId))
			return &discovery.GetOneInstanceResponse{
				Response: discovery.CreateResponse(discovery.ErrServiceNotExists,
					fmt.Sprintf("Provider[%s] does not exist.", request.ProviderServiceId)),
			}, nil
		}
		log.Error(fmt.Sprintf("get provider failed, consumer %s find provider instance %s %s",
			request.ConsumerServiceId, request.ProviderServiceId, request.ProviderInstanceId), err)
		return &discovery.GetOneInstanceResponse{
			Response: discovery.CreateResponse(discovery.ErrInternal, err.Error()),
		}, err
	}
	findFlag := func() string {
		return fmt.Sprintf("consumer[%s][%s/%s/%s/%s] find provider[%s][%s/%s/%s/%s] instance[%s]",
			request.ConsumerServiceId, service.Service.Environment, service.Service.AppId, service.Service.ServiceName, service.Service.Version,
			provider.Service.ServiceId, provider.Service.Environment, provider.Service.AppId, provider.Service.ServiceName, provider.Service.Version,
			request.ProviderInstanceId)
	}

	domainProject := util.ParseDomainProject(ctx)
	services, err := filterServices(ctx, discovery.MicroServiceToKey(domainProject, provider.Service))
	if err != nil {
		log.Error(fmt.Sprintf("get instance failed %s", findFlag()), err)
		return &discovery.GetOneInstanceResponse{
			Response: discovery.CreateResponse(discovery.ErrInternal, err.Error()),
		}, err
	}
	rev, _ := ctx.Value(util.CtxRequestRevision).(string)
	serviceIDs = filterServiceIDs(ctx, request.ConsumerServiceId, request.Tags, services)
	if len(serviceIDs) == 0 {
		mes := fmt.Errorf("%s failed, provider instance does not exist", findFlag())
		log.Error("query service failed", mes)
		return &discovery.GetOneInstanceResponse{
			Response: discovery.CreateResponse(discovery.ErrInstanceNotExists, mes.Error()),
		}, nil
	}
	instances, ok := cache.GetMicroServiceInstancesByID(request.ProviderServiceId)
	if !ok {
		instances, err = getMicroServiceInstancesByServiceID(ctx, request.ProviderServiceId)
		if err != nil {
			log.Error(fmt.Sprintf("get instance failed %s", findFlag()), err)
			return &discovery.GetOneInstanceResponse{
				Response: discovery.CreateResponse(discovery.ErrInternal, err.Error()),
			}, err
		}
	}

	instance := instances[0]
	// use explicit instanceId to query
	if len(request.ProviderInstanceId) != 0 {
		isExist := false
		for _, ins := range instances {
			if ins.InstanceId == request.ProviderInstanceId {
				instance = ins
				isExist = true
				break
			}
		}
		if !isExist {
			mes := fmt.Errorf("%s failed, provider instance does not exist", findFlag())
			log.Error("get instance failed", mes)
			return &discovery.GetOneInstanceResponse{
				Response: discovery.CreateResponse(discovery.ErrInstanceNotExists, mes.Error()),
			}, nil
		}
	}

	newRev, _ := formatRevision(request.ConsumerServiceId, instances)
	if rev == newRev {
		instance = nil // for gRPC
	}
	// TODO support gRPC output context
	_ = util.WithResponseRev(ctx, newRev)

	return &discovery.GetOneInstanceResponse{
		Response: discovery.CreateResponse(discovery.ResponseSuccess, "Get instance successfully."),
		Instance: instance,
	}, nil
}

func (ds *DataSource) GetInstances(ctx context.Context, request *discovery.GetInstancesRequest) (*discovery.GetInstancesResponse, error) {
	service := &model.Service{}
	var err error

	if len(request.ConsumerServiceId) > 0 {
		var exist bool
		service, exist = cache.GetServiceByID(request.ConsumerServiceId)
		if !exist {
			service, err = getServiceByServiceID(ctx, request.ConsumerServiceId)
			if err != nil {
				if errors.Is(err, datasource.ErrNoData) {
					log.Debug(fmt.Sprintf("consumer does not exist, consumer %s find provider %s instances",
						request.ConsumerServiceId, request.ProviderServiceId))
					return &discovery.GetInstancesResponse{
						Response: discovery.CreateResponse(discovery.ErrServiceNotExists,
							fmt.Sprintf("Consumer[%s] does not exist.", request.ConsumerServiceId)),
					}, nil
				}
				log.Error(fmt.Sprintf("get consumer failed, consumer %s find provider %s instances",
					request.ConsumerServiceId, request.ProviderServiceId), err)
				return &discovery.GetInstancesResponse{
					Response: discovery.CreateResponse(discovery.ErrInternal, err.Error()),
				}, err
			}
		}
	}

	provider, ok := cache.GetServiceByID(request.ProviderServiceId)
	if !ok {
		provider, err = getServiceByServiceID(ctx, request.ProviderServiceId)
		if err != nil {
			if errors.Is(err, datasource.ErrNoData) {
				log.Debug(fmt.Sprintf("provider does not exist, consumer %s find provider %s  instances",
					request.ConsumerServiceId, request.ProviderServiceId))
				return &discovery.GetInstancesResponse{
					Response: discovery.CreateResponse(discovery.ErrServiceNotExists,
						fmt.Sprintf("provider[%s] does not exist.", request.ProviderServiceId)),
				}, nil
			}
			log.Error(fmt.Sprintf("get provider failed, consumer %s find provider instances %s",
				request.ConsumerServiceId, request.ProviderServiceId), err)
			return &discovery.GetInstancesResponse{
				Response: discovery.CreateResponse(discovery.ErrInternal, err.Error()),
			}, err
		}
	}

	findFlag := func() string {
		return fmt.Sprintf("consumer[%s][%s/%s/%s/%s] find provider[%s][%s/%s/%s/%s] instances",
			request.ConsumerServiceId, service.Service.Environment, service.Service.AppId, service.Service.ServiceName, service.Service.Version,
			provider.Service.ServiceId, provider.Service.Environment, provider.Service.AppId, provider.Service.ServiceName, provider.Service.Version)
	}

	rev, _ := ctx.Value(util.CtxRequestRevision).(string)
	serviceIDs := filterServiceIDs(ctx, request.ConsumerServiceId, request.Tags, []*model.Service{provider})
	if len(serviceIDs) == 0 {
		mes := fmt.Errorf("%s failed, provider does not exist", findFlag())
		log.Error("query service failed", mes)
		return &discovery.GetInstancesResponse{
			Response: discovery.CreateResponse(discovery.ErrServiceNotExists, mes.Error()),
		}, nil
	}
	instances, ok := cache.GetMicroServiceInstancesByID(request.ProviderServiceId)
	if !ok {
		instances, err = getMicroServiceInstancesByServiceID(ctx, request.ProviderServiceId)
		if err != nil {
			log.Error(fmt.Sprintf("get instances failed %s", findFlag()), err)
			return &discovery.GetInstancesResponse{
				Response: discovery.CreateResponse(discovery.ErrInternal, err.Error()),
			}, err
		}
	}
	newRev, _ := formatRevision(request.ConsumerServiceId, instances)
	if rev == newRev {
		instances = nil // for gRPC
	}
	_ = util.WithResponseRev(ctx, newRev)
	return &discovery.GetInstancesResponse{
		Response:  discovery.CreateResponse(discovery.ResponseSuccess, "query service instances successfully."),
		Instances: instances,
	}, nil
}

// FindInstances returns instances under the specified domain
func (ds *DataSource) FindInstances(ctx context.Context, request *discovery.FindInstancesRequest) (*discovery.FindInstancesResponse, error) {
	provider := &discovery.MicroServiceKey{
		Tenant:      util.ParseTargetDomainProject(ctx),
		Environment: request.Environment,
		AppId:       request.AppId,
		ServiceName: request.ServiceName,
		Alias:       request.Alias,
		Version:     request.VersionRule,
	}
	rev, ok := ctx.Value(util.CtxRequestRevision).(string)
	if !ok {
		err := errors.New("rev request context is not type string")
		log.Error("", err)
		return &discovery.FindInstancesResponse{
			Response: discovery.CreateResponse(discovery.ErrInternal, err.Error()),
		}, err
	}

	if apt.IsGlobal(provider) {
		return findSharedServiceInstance(ctx, request, provider, rev)
	}

	return getInstance(ctx, request, provider, rev)
}

// GetProviderInstances returns instances under the specified domain
func (ds *DataSource) GetProviderInstances(ctx context.Context, request *discovery.GetProviderInstancesRequest) (instances []*discovery.MicroServiceInstance, rev string, err error) {
	filter := mutil.NewBasicFilter(ctx, mutil.InstanceServiceID(request.ProviderServiceId))
	instances, err = findMicroServiceInstances(ctx, filter)
	if err != nil {
		return
	}
	return instances, "", nil
}

func (ds *DataSource) GetAllInstances(ctx context.Context, request *discovery.GetAllInstancesRequest) (*discovery.GetAllInstancesResponse, error) {
	filter := mutil.NewBasicFilter(ctx)
	instances, err := findMicroServiceInstances(ctx, filter)
	if err != nil {
		return nil, err
	}
	resp := &discovery.GetAllInstancesResponse{
		Response: discovery.CreateResponse(discovery.ResponseSuccess, "get all instances successfully"),
	}
	resp.Instances = instances
	return resp, nil
}

func (ds *DataSource) BatchGetProviderInstances(ctx context.Context, request *discovery.BatchGetInstancesRequest) (instances []*discovery.MicroServiceInstance, rev string, err error) {
	if request == nil || len(request.ServiceIds) == 0 {
		return nil, "", mutil.ErrInvalidParam
	}
	domain := util.ParseDomain(ctx)
	project := util.ParseProject(ctx)
	for _, providerServiceID := range request.ServiceIds {
		filter := mutil.NewDomainProjectFilter(domain, project, mutil.InstanceServiceID(providerServiceID))
		microServiceInstances, err := findMicroServiceInstances(ctx, filter)
		if err != nil {
			return instances, "", nil
		}
		instances = append(instances, microServiceInstances...)
	}
	return instances, "", nil
}

func (ds *DataSource) UpdateInstanceStatus(ctx context.Context, request *discovery.UpdateInstanceStatusRequest) (*discovery.UpdateInstanceStatusResponse, error) {
	updateStatusFlag := util.StringJoin([]string{request.ServiceId, request.InstanceId, request.Status}, "/")

	// todo finish get instance
	filter := mutil.NewBasicFilter(ctx, mutil.InstanceServiceID(request.ServiceId), mutil.InstanceInstanceID(request.InstanceId))
	instance, err := findInstance(ctx, filter)

	if err != nil && err != datasource.ErrNoData {
		log.Error(fmt.Sprintf("update instance %s status failed", updateStatusFlag), err)
		return &discovery.UpdateInstanceStatusResponse{
			Response: discovery.CreateResponse(discovery.ErrInternal, err.Error()),
		}, err
	}
	if instance == nil {
		log.Error(fmt.Sprintf("update instance %s status failed, instance does not exist", updateStatusFlag), err)
		return &discovery.UpdateInstanceStatusResponse{
			Response: discovery.CreateResponse(discovery.ErrInstanceNotExists, "service instance does not exist."),
		}, nil
	}

	copyInstanceRef := *instance
	copyInstanceRef.Instance.Status = request.Status
	setValue := mutil.NewFilter(
		mutil.InstanceModTime(strconv.FormatInt(time.Now().Unix(), baseTen)),
		mutil.InstanceStatus(copyInstanceRef.Instance.Status),
	)
	updateFilter := mutil.NewFilter(mutil.Set(setValue))
	if err := updateInstance(ctx, filter, updateFilter); err != nil {
		log.Error(fmt.Sprintf("update instance %s status failed", updateStatusFlag), err)
		resp := &discovery.UpdateInstanceStatusResponse{
			Response: discovery.CreateResponseWithSCErr(err),
		}
		if err.InternalError() {
			return resp, err
		}
		return resp, nil
	}

	log.Info(fmt.Sprintf("update instance[%s] status successfully", updateStatusFlag))
	return &discovery.UpdateInstanceStatusResponse{
		Response: discovery.CreateResponse(discovery.ResponseSuccess, "update service instance status successfully."),
	}, nil
}

func (ds *DataSource) UpdateInstanceProperties(ctx context.Context, request *discovery.UpdateInstancePropsRequest) (*discovery.UpdateInstancePropsResponse, error) {
	instanceFlag := util.StringJoin([]string{request.ServiceId, request.InstanceId}, "/")
	domain := util.ParseDomain(ctx)
	project := util.ParseProject(ctx)
	filter := mutil.NewDomainProjectFilter(domain, project, mutil.InstanceServiceID(request.ServiceId), mutil.InstanceInstanceID(request.InstanceId))
	instance, err := findInstance(ctx, filter)
	if err != nil && err != datasource.ErrNoData {
		log.Error(fmt.Sprintf("update instance %s properties failed", instanceFlag), err)
		return &discovery.UpdateInstancePropsResponse{
			Response: discovery.CreateResponse(discovery.ErrInternal, err.Error()),
		}, err
	}
	if instance == nil {
		log.Error(fmt.Sprintf("update instance %s properties failed, instance does not exist", instanceFlag), err)
		return &discovery.UpdateInstancePropsResponse{
			Response: discovery.CreateResponse(discovery.ErrInstanceNotExists, "Service instance does not exist."),
		}, nil
	}

	copyInstanceRef := *instance
	copyInstanceRef.Instance.Properties = request.Properties

	// todo finish update instance
	filter = mutil.NewDomainProjectFilter(domain, project, mutil.InstanceServiceID(request.ServiceId), mutil.InstanceInstanceID(request.InstanceId))
	setValue := mutil.NewFilter(
		mutil.InstanceModTime(strconv.FormatInt(time.Now().Unix(), baseTen)),
		mutil.InstanceProperties(copyInstanceRef.Instance.Properties),
	)
	updateFilter := mutil.NewFilter(mutil.Set(setValue))
	if err := updateInstance(ctx, filter, updateFilter); err != nil {
		log.Error(fmt.Sprintf("update instance %s properties failed", instanceFlag), err)
		resp := &discovery.UpdateInstancePropsResponse{
			Response: discovery.CreateResponseWithSCErr(err),
		}
		if err.InternalError() {
			return resp, err
		}
		return resp, nil
	}

	log.Info(fmt.Sprintf("update instance[%s] properties successfully", instanceFlag))
	return &discovery.UpdateInstancePropsResponse{
		Response: discovery.CreateResponse(discovery.ResponseSuccess, "update service instance properties successfully."),
	}, nil
}

func (ds *DataSource) UnregisterInstance(ctx context.Context, request *discovery.UnregisterInstanceRequest) (*discovery.UnregisterInstanceResponse, error) {
	remoteIP := util.GetIPFromContext(ctx)
	serviceID := request.ServiceId
	instanceID := request.InstanceId

	instanceFlag := util.StringJoin([]string{serviceID, instanceID}, "/")

	filter := mutil.NewBasicFilter(ctx, mutil.InstanceServiceID(serviceID), mutil.InstanceInstanceID(instanceID))
	isDeleted, err := deleteInstance(ctx, filter)
	if err != nil || !isDeleted {
		log.Error(fmt.Sprintf("unregister instance failed, instance %s, operator %s revoke instance failed", instanceFlag, remoteIP), err)
		return &discovery.UnregisterInstanceResponse{
			Response: discovery.CreateResponse(discovery.ErrInternal, "delete instance failed"),
		}, err
	}

	log.Info(fmt.Sprintf("unregister instance[%s], operator %s", instanceFlag, remoteIP))
	return &discovery.UnregisterInstanceResponse{
		Response: discovery.CreateResponse(discovery.ResponseSuccess, "unregister service instance successfully."),
	}, nil
}

func (ds *DataSource) Heartbeat(ctx context.Context, request *discovery.HeartbeatRequest) (*discovery.HeartbeatResponse, error) {
	remoteIP := util.GetIPFromContext(ctx)
	instanceFlag := util.StringJoin([]string{request.ServiceId, request.InstanceId}, "/")
	err := keepAliveLease(ctx, request)
	if err != nil {
		log.Error(fmt.Sprintf("heartbeat failed, instance %s operator %s", instanceFlag, remoteIP), err)
		resp := &discovery.HeartbeatResponse{
			Response: discovery.CreateResponseWithSCErr(err),
		}
		if err.InternalError() {
			return resp, err
		}
		return resp, nil
	}
	return &discovery.HeartbeatResponse{
		Response: discovery.CreateResponse(discovery.ResponseSuccess,
			"Update service instance heartbeat successfully."),
	}, nil
}

func (ds *DataSource) HeartbeatSet(ctx context.Context, request *discovery.HeartbeatSetRequest) (*discovery.HeartbeatSetResponse, error) {
	heartBeatCount := len(request.Instances)
	existFlag := make(map[string]bool, heartBeatCount)
	instancesHbRst := make(chan *discovery.InstanceHbRst, heartBeatCount)
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
		gopool.Go(getHeartbeatFunc(ctx, instancesHbRst, heartbeatElement))
	}

	count := 0
	successFlag := false
	failFlag := false
	instanceHbRstArr := make([]*discovery.InstanceHbRst, 0, heartBeatCount)

	for hbRst := range instancesHbRst {
		count++
		if len(hbRst.ErrMessage) != 0 {
			failFlag = true
		} else {
			successFlag = true
		}
		instanceHbRstArr = append(instanceHbRstArr, hbRst)
		if count == noMultiCounter {
			close(instancesHbRst)
		}
	}

	if !failFlag && successFlag {
		log.Info(fmt.Sprintf("batch update heartbeats[%d] successfully", count))
		return &discovery.HeartbeatSetResponse{
			Response:  discovery.CreateResponse(discovery.ResponseSuccess, "heartbeat set successfully."),
			Instances: instanceHbRstArr,
		}, nil
	}

	log.Info(fmt.Sprintf("batch update heartbeats failed %v", request.Instances))
	return &discovery.HeartbeatSetResponse{
		Response:  discovery.CreateResponse(discovery.ErrInstanceNotExists, "heartbeat set failed."),
		Instances: instanceHbRstArr,
	}, nil
}

func (ds *DataSource) BatchFind(ctx context.Context, request *discovery.BatchFindInstancesRequest) (*discovery.BatchFindInstancesResponse, error) {
	response := &discovery.BatchFindInstancesResponse{
		Response: discovery.CreateResponse(discovery.ResponseSuccess, "batch query service instances successfully."),
	}

	var err error

	response.Services, err = ds.batchFindServices(ctx, request)
	if err != nil {
		return &discovery.BatchFindInstancesResponse{
			Response: discovery.CreateResponse(discovery.ErrInternal, err.Error()),
		}, err
	}

	response.Instances, err = ds.batchFindInstances(ctx, request)
	if err != nil {
		return &discovery.BatchFindInstancesResponse{
			Response: discovery.CreateResponse(discovery.ErrInternal, err.Error()),
		}, err
	}

	return response, nil
}

func (ds *DataSource) batchFindServices(ctx context.Context, request *discovery.BatchFindInstancesRequest) (*discovery.BatchFindResult, error) {
	if len(request.Services) == 0 {
		return nil, nil
	}
	cloneCtx := util.CloneContext(ctx)

	services := &discovery.BatchFindResult{}
	failedResult := make(map[int32]*discovery.FindFailedResult)
	for index, key := range request.Services {
		findCtx := util.SetContext(cloneCtx, util.CtxRequestRevision, key.Rev)
		resp, err := ds.FindInstances(findCtx, &discovery.FindInstancesRequest{
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
		appendFindResponse(findCtx, int64(index), resp.Response, resp.Instances,
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

func (ds *DataSource) batchFindInstances(ctx context.Context, request *discovery.BatchFindInstancesRequest) (*discovery.BatchFindResult, error) {
	if len(request.Instances) == 0 {
		return nil, nil
	}
	cloneCtx := util.CloneContext(ctx)
	// can not find the shared provider instances
	cloneCtx = util.SetTargetDomainProject(cloneCtx, util.ParseDomain(ctx), util.ParseProject(ctx))

	instances := &discovery.BatchFindResult{}
	failedResult := make(map[int32]*discovery.FindFailedResult)
	for index, key := range request.Instances {
		getCtx := util.SetContext(cloneCtx, util.CtxRequestRevision, key.Rev)
		resp, err := ds.GetInstance(getCtx, &discovery.GetOneInstanceRequest{
			ConsumerServiceId:  request.ConsumerServiceId,
			ProviderServiceId:  key.Instance.ServiceId,
			ProviderInstanceId: key.Instance.InstanceId,
		})
		if err != nil {
			return nil, err
		}
		failed, ok := failedResult[resp.Response.GetCode()]
		appendFindResponse(getCtx, int64(index), resp.Response, []*discovery.MicroServiceInstance{resp.Instance},
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

func appendFindResponse(ctx context.Context, index int64, resp *discovery.Response, instances []*discovery.MicroServiceInstance,
	updatedResult *[]*discovery.FindResult, notModifiedResult *[]int64, failedResult **discovery.FindFailedResult) {
	if code := resp.GetCode(); code != discovery.ResponseSuccess {
		if *failedResult == nil {
			*failedResult = &discovery.FindFailedResult{
				Error: discovery.NewError(code, resp.GetMessage()),
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
	*updatedResult = append(*updatedResult, &discovery.FindResult{
		Index:     index,
		Instances: instances,
		Rev:       ov,
	})
}

func registerInstanceSingle(ctx context.Context, request *discovery.RegisterInstanceRequest, isUserDefinedID bool) (*discovery.RegisterInstanceResponse, error) {
	resp, needRegister, err := preProcessRegister(ctx, request.Instance, isUserDefinedID)
	if err != nil || !needRegister {
		log.Error("pre process instance err, or instance already existed", err)
		return resp, err
	}

	return registryInstance(ctx, request)
}

func batchRegisterInstance(ctx context.Context, events []*event.InstanceRegisterEvent) (*discovery.RegisterInstanceResponse, error) {
	instances := make([]interface{}, len(events))

	for i, event := range events {
		eventCtx := event.Ctx
		instance := event.Request.Instance

		resp, needRegister, err := preProcessRegister(eventCtx, instance, event.IsCustomID)

		if err != nil || !needRegister {
			log.Error("pre process instance err, or instance existed", err)
			return resp, err
		}

		domain := util.ParseDomain(eventCtx)
		project := util.ParseProject(eventCtx)

		data := model.Instance{
			Domain:      domain,
			Project:     project,
			RefreshTime: time.Now(),
			Instance:    instance,
		}
		instances[i] = data
	}

	return registryInstances(ctx, instances)
}

func preProcessRegister(ctx context.Context, instance *discovery.MicroServiceInstance, isUserDefinedID bool) (*discovery.RegisterInstanceResponse, bool, error) {
	remoteIP := util.GetIPFromContext(ctx)
	// 允许自定义 id
	if isUserDefinedID {
		resp, err := datasource.Instance().Heartbeat(ctx, &discovery.HeartbeatRequest{
			InstanceId: instance.InstanceId,
			ServiceId:  instance.ServiceId,
		})
		if err != nil || resp == nil {
			log.Error(fmt.Sprintf("register service %s's instance failed, endpoints %s, host '%s', operator %s",
				instance.ServiceId, instance.Endpoints, instance.HostName, remoteIP), err)
			return &discovery.RegisterInstanceResponse{
				Response: discovery.CreateResponse(discovery.ErrInternal, err.Error()),
			}, false, nil
		}
		switch resp.Response.GetCode() {
		case discovery.ResponseSuccess:
			log.Info(fmt.Sprintf("register instance successful, reuse instance[%s/%s], operator %s",
				instance.ServiceId, instance.InstanceId, remoteIP))
			return &discovery.RegisterInstanceResponse{
				Response:   resp.Response,
				InstanceId: instance.InstanceId,
			}, false, nil
		case discovery.ErrInstanceNotExists:
			//register a new one
		default:
			log.Error(fmt.Sprintf("register instance failed, reuse instance %s %s, operator %s",
				instance.ServiceId, instance.InstanceId, remoteIP), err)
			return &discovery.RegisterInstanceResponse{
				Response: resp.Response,
			}, false, err
		}

	}

	if err := preProcessRegisterInstance(ctx, instance); err != nil {
		log.Error(fmt.Sprintf("register service %s instance failed, endpoints %s, host %s operator %s",
			instance.ServiceId, instance.Endpoints, instance.HostName, remoteIP), err)
		return &discovery.RegisterInstanceResponse{
			Response: discovery.CreateResponseWithSCErr(err),
		}, false, nil
	}

	return &discovery.RegisterInstanceResponse{
		Response:   discovery.CreateResponse(discovery.ResponseSuccess, "process success"),
		InstanceId: instance.InstanceId,
	}, true, nil
}

func preProcessRegisterInstance(ctx context.Context, instance *discovery.MicroServiceInstance) *errsvc.Error {
	if len(instance.Status) == 0 {
		instance.Status = discovery.MSI_UP
	}

	if len(instance.InstanceId) == 0 {
		instance.InstanceId = uuid.Generator().GetInstanceID(ctx)
	}

	instance.Timestamp = strconv.FormatInt(time.Now().Unix(), baseTen)
	instance.ModTimestamp = instance.Timestamp

	// 这里应该根据租约计时
	renewalInterval := apt.RegistryDefaultLeaseRenewalinterval
	retryTimes := apt.RegistryDefaultLeaseRetrytimes
	if instance.HealthCheck == nil {
		instance.HealthCheck = &discovery.HealthCheck{
			Mode:     discovery.CHECK_BY_HEARTBEAT,
			Interval: renewalInterval,
			Times:    retryTimes,
		}
	} else {
		// Health check对象仅用于呈现服务健康检查逻辑，如果CHECK_BY_PLATFORM类型，表明由sidecar代发心跳，实例120s超时
		switch instance.HealthCheck.Mode {
		case discovery.CHECK_BY_HEARTBEAT:
			d := instance.HealthCheck.Interval * (instance.HealthCheck.Times + 1)
			if d <= 0 {
				return discovery.NewError(discovery.ErrInvalidParams, "invalid 'healthCheck' settings in request body.")
			}
		case discovery.CHECK_BY_PLATFORM:
			// 默认120s
			instance.HealthCheck.Interval = renewalInterval
			instance.HealthCheck.Times = retryTimes
		}
	}

	cacheService, ok := cache.GetServiceByID(instance.ServiceId)

	var microService *discovery.MicroService
	if ok {
		microService = cacheService.Service
		instance.Version = microService.Version
	} else {
		filter := mutil.NewBasicFilter(ctx, mutil.ServiceServiceID(instance.ServiceId))
		microservice, err := findService(ctx, filter)
		if err != nil {
			log.Error("get service failed", err)
			return discovery.NewError(discovery.ErrServiceNotExists, "invalid 'serviceID' in request body.")
		}
		instance.Version = microservice.Service.Version
	}
	return nil
}

func registryInstance(ctx context.Context, request *discovery.RegisterInstanceRequest) (*discovery.RegisterInstanceResponse, error) {
	domain := util.ParseDomain(ctx)
	project := util.ParseProject(ctx)
	remoteIP := util.GetIPFromContext(ctx)
	instance := request.Instance
	ttl := instance.HealthCheck.Interval * (instance.HealthCheck.Times + 1)

	instanceFlag := fmt.Sprintf("ttl %ds, endpoints %v, host '%s', serviceID %s",
		ttl, instance.Endpoints, instance.HostName, instance.ServiceId)

	instanceID := instance.InstanceId
	data := &model.Instance{
		Domain:      domain,
		Project:     project,
		RefreshTime: time.Now(),
		Instance:    instance,
	}

	insertRes, err := client.GetMongoClient().Insert(ctx, model.CollectionInstance, data)
	if err != nil {
		if mutil.IsDuplicateKey(err) {
			return &discovery.RegisterInstanceResponse{
				Response:   discovery.CreateResponse(discovery.ResponseSuccess, "Register service instance successfully."),
				InstanceId: instanceID,
			}, nil
		}
		log.Error(fmt.Sprintf("register instance failed %s instanceID %s operator %s", instanceFlag, instanceID, remoteIP), err)
		return &discovery.RegisterInstanceResponse{
			Response: discovery.CreateResponse(discovery.ErrUnavailableBackend, err.Error()),
		}, err
	}

	// need to complete the instance offline function in time, so you need to check the heartbeat after registering the instance
	err = heartbeat.Instance().CheckInstance(ctx, instance)
	if err != nil {
		log.Error(fmt.Sprintf("fail to check instance, instance[%s]. operator %s", instance.InstanceId, remoteIP), err)
	}

	log.Info(fmt.Sprintf("register instance %s, instanceID %s, operator %s",
		instanceFlag, insertRes.InsertedID, remoteIP))
	return &discovery.RegisterInstanceResponse{
		Response:   discovery.CreateResponse(discovery.ResponseSuccess, "Register service instance successfully."),
		InstanceId: instanceID,
	}, nil
}

func registryInstances(ctx context.Context, instances []interface{}) (*discovery.RegisterInstanceResponse, error) {
	opts := options.InsertManyOptions{}
	opts.SetOrdered(false)
	opts.SetBypassDocumentValidation(true)
	err := batchInsertMicroServiceInstances(ctx, instances, &opts)
	if err != nil {
		log.Error("batch register instance failed", err)
		return &discovery.RegisterInstanceResponse{
			Response: discovery.CreateResponse(discovery.ErrUnavailableBackend, err.Error()),
		}, err
	}
	return &discovery.RegisterInstanceResponse{
		Response: discovery.CreateResponse(discovery.ResponseSuccess, "register service instance successfully."),
	}, nil
}

func formatRevision(consumerServiceID string, instances []*discovery.MicroServiceInstance) (string, error) {
	if instances == nil {
		return fmt.Sprintf("%x", sha1.Sum(util.StringToBytesWithNoCopy(consumerServiceID))), nil
	}
	copyInstance := make([]*discovery.MicroServiceInstance, len(instances))
	copy(copyInstance, instances)
	sort.Sort(InstanceSlice(copyInstance))
	data, err := json.Marshal(copyInstance)
	if err != nil {
		log.Error("fail to marshal instance json", err)
		return "", err
	}
	s := fmt.Sprintf("%s.%x", consumerServiceID, sha1.Sum(data))
	return fmt.Sprintf("%x", sha1.Sum(util.StringToBytesWithNoCopy(s))), nil
}

func getMicroServiceInstancesByServiceID(ctx context.Context, serviceID string) ([]*discovery.MicroServiceInstance, error) {
	filter := mutil.NewFilter(mutil.InstanceServiceID(serviceID))
	option := &options.FindOptions{Sort: bson.D{{Key: mutil.ConnectWithDot([]string{model.ColumnInstance, model.ColumnVersion}), Value: -1}}}
	return findMicroServiceInstances(ctx, filter, option)
}

func findSharedServiceInstance(ctx context.Context, request *discovery.FindInstancesRequest, provider *discovery.MicroServiceKey, rev string) (*discovery.FindInstancesResponse, error) {
	var err error
	// it means the shared micro-services must be the same env with SC.
	provider.Environment = apt.Service.Environment
	findFlag := func() string {
		return fmt.Sprintf("find shared provider[%s/%s/%s/%s]", provider.Environment, provider.AppId, provider.ServiceName, provider.Version)
	}
	basicConditionsFilterServices, err := filterServicesByBasicConditions(ctx, provider)
	if err != nil {
		log.Error(fmt.Sprintf("find shared service instance failed %s", findFlag()), err)
		return &discovery.FindInstancesResponse{
			Response: discovery.CreateResponse(discovery.ErrInternal, err.Error()),
		}, err
	}
	services, err := filterServices(ctx, provider)
	if err != nil {
		log.Error(fmt.Sprintf("find shared service instance failed %s", findFlag()), err)
		return &discovery.FindInstancesResponse{
			Response: discovery.CreateResponse(discovery.ErrInternal, err.Error()),
		}, err
	}
	if services == nil && len(basicConditionsFilterServices) == 0 {
		mes := fmt.Errorf("%s failed, provider does not exist", findFlag())
		log.Error("find shared service instance failed", mes)
		return &discovery.FindInstancesResponse{
			Response: discovery.CreateResponse(discovery.ErrServiceNotExists, mes.Error()),
		}, nil
	}
	serviceIDs := filterServiceIDs(ctx, request.ConsumerServiceId, request.Tags, services)
	inFilter := mutil.NewFilter(mutil.In(serviceIDs))
	filter := mutil.NewFilter(mutil.InstanceServiceID(inFilter))
	option := &options.FindOptions{Sort: bson.D{{Key: mutil.ConnectWithDot([]string{model.ColumnInstance, model.ColumnVersion}), Value: -1}}}
	instances, err := findMicroServiceInstances(ctx, filter, option)
	if err != nil {
		log.Error(fmt.Sprintf("find shared service instance failed %s", findFlag()), err)
		return &discovery.FindInstancesResponse{
			Response: discovery.CreateResponse(discovery.ErrInternal, err.Error()),
		}, err
	}
	newRev, _ := formatRevision(request.ConsumerServiceId, instances)
	if rev == newRev {
		instances = nil // for gRPC
	}
	// TODO support gRPC output context
	_ = util.WithResponseRev(ctx, newRev)
	return &discovery.FindInstancesResponse{
		Response:  discovery.CreateResponse(discovery.ResponseSuccess, "query service instances successfully."),
		Instances: instances,
	}, nil
}

func getInstance(ctx context.Context, request *discovery.FindInstancesRequest, provider *discovery.MicroServiceKey, rev string) (*discovery.FindInstancesResponse, error) {
	var err error
	domainProject := util.ParseDomainProject(ctx)
	service := &model.Service{Service: &discovery.MicroService{Environment: request.Environment}}
	if len(request.ConsumerServiceId) > 0 {
		filter := mutil.NewBasicFilter(ctx, mutil.ServiceServiceID(request.ConsumerServiceId))
		service, err = findService(ctx, filter)
		if err != nil {
			if errors.Is(err, datasource.ErrNoData) {
				log.Debug(fmt.Sprintf("consumer does not exist, consumer %s find provider %s/%s/%s/%s",
					request.ConsumerServiceId, request.Environment, request.AppId, request.ServiceName, request.VersionRule))
				return &discovery.FindInstancesResponse{
					Response: discovery.CreateResponse(discovery.ErrServiceNotExists,
						fmt.Sprintf("consumer[%s] does not exist.", request.ConsumerServiceId)),
				}, nil
			}
			log.Error(fmt.Sprintf("get consumer failed, consumer %s find provider %s/%s/%s/%s",
				request.ConsumerServiceId, request.Environment, request.AppId, request.ServiceName, request.VersionRule), err)
			return &discovery.FindInstancesResponse{
				Response: discovery.CreateResponse(discovery.ErrInternal, err.Error()),
			}, err
		}
		provider.Environment = service.Service.Environment
	}

	// provider is not a shared micro-service,
	// only allow shared micro-service instances found request different domains.
	ctx = util.SetTargetDomainProject(ctx, util.ParseDomain(ctx), util.ParseProject(ctx))
	provider.Tenant = util.ParseTargetDomainProject(ctx)

	findFlag := func() string {
		return fmt.Sprintf("consumer[%s][%s/%s/%s/%s] find provider[%s/%s/%s/%s]",
			request.ConsumerServiceId, service.Service.Environment, service.Service.AppId, service.Service.ServiceName, service.Service.Version,
			provider.Environment, provider.AppId, provider.ServiceName, provider.Version)
	}
	basicConditionsFilterServices, err := filterServicesByBasicConditions(ctx, provider)
	if err != nil {
		log.Error(fmt.Sprintf("find instance failed %s", findFlag()), err)
		return &discovery.FindInstancesResponse{
			Response: discovery.CreateResponse(discovery.ErrInternal, err.Error()),
		}, err
	}
	services, err := filterServices(ctx, provider)
	if err != nil {
		log.Error(fmt.Sprintf("find instance failed %s", findFlag()), err)
		return &discovery.FindInstancesResponse{
			Response: discovery.CreateResponse(discovery.ErrInternal, err.Error()),
		}, err
	}
	if services == nil && len(basicConditionsFilterServices) == 0 {
		mes := fmt.Errorf("%s failed, provider does not exist", findFlag())
		log.Error("find instance failed", mes)
		return &discovery.FindInstancesResponse{
			Response: discovery.CreateResponse(discovery.ErrServiceNotExists, mes.Error()),
		}, nil
	}
	serviceIDs := filterServiceIDs(ctx, request.ConsumerServiceId, request.Tags, services)
	inFilter := mutil.NewFilter(mutil.In(serviceIDs))
	filter := mutil.NewFilter(mutil.InstanceServiceID(inFilter))
	option := &options.FindOptions{Sort: bson.D{{Key: mutil.ConnectWithDot([]string{model.ColumnInstance, model.ColumnVersion}), Value: -1}}}
	instances, err := findMicroServiceInstances(ctx, filter, option)
	if err != nil {
		log.Error(fmt.Sprintf("find instance failed %s", findFlag()), err)
		return &discovery.FindInstancesResponse{
			Response: discovery.CreateResponse(discovery.ErrInternal, err.Error()),
		}, err
	}
	// add dependency queue
	if len(request.ConsumerServiceId) > 0 &&
		len(serviceIDs) > 0 {
		provider, err = reshapeProviderKey(ctx, provider, serviceIDs[0])
		if err != nil {
			return nil, err
		}
		if provider != nil {
			err = addServiceVersionRule(ctx, domainProject, service.Service, provider)
		} else {
			mes := fmt.Errorf("%s failed, provider does not exist", findFlag())
			log.Error("add service version rule failed", mes)
			return &discovery.FindInstancesResponse{
				Response: discovery.CreateResponse(discovery.ErrServiceNotExists, mes.Error()),
			}, nil
		}
		if err != nil {
			log.Error(fmt.Sprintf("add service version rule failed %s", findFlag()), err)
			return &discovery.FindInstancesResponse{
				Response: discovery.CreateResponse(discovery.ErrInternal, err.Error()),
			}, err
		}
	}
	newRev, _ := formatRevision(request.ConsumerServiceId, instances)
	if rev == newRev {
		instances = nil // for gRPC
	}
	// TODO support gRPC output context
	_ = util.WithResponseRev(ctx, newRev)
	return &discovery.FindInstancesResponse{
		Response:  discovery.CreateResponse(discovery.ResponseSuccess, "query service instances successfully."),
		Instances: instances,
	}, nil
}

func reshapeProviderKey(ctx context.Context, provider *discovery.MicroServiceKey, providerID string) (*discovery.MicroServiceKey, error) {
	//维护version的规则,service name 可能是别名，所以重新获取
	filter := mutil.NewBasicFilter(ctx, mutil.ServiceServiceID(providerID))
	providerService, err := findService(ctx, filter)
	if err != nil {
		return nil, err
	}

	versionRule := provider.Version
	provider = discovery.MicroServiceToKey(provider.Tenant, providerService.Service)
	provider.Version = versionRule
	return provider, nil
}

func keepAliveLease(ctx context.Context, request *discovery.HeartbeatRequest) *errsvc.Error {
	_, err := heartbeat.Instance().Heartbeat(ctx, request)
	if err != nil {
		return discovery.NewError(discovery.ErrInstanceNotExists, err.Error())
	}
	return nil
}

func getHeartbeatFunc(ctx context.Context, instancesHbRst chan<- *discovery.InstanceHbRst, element *discovery.HeartbeatSetElement) func(context.Context) {
	return func(_ context.Context) {
		hbRst := &discovery.InstanceHbRst{
			ServiceId:  element.ServiceId,
			InstanceId: element.InstanceId,
			ErrMessage: "",
		}

		req := &discovery.HeartbeatRequest{
			InstanceId: element.InstanceId,
			ServiceId:  element.ServiceId,
		}

		err := keepAliveLease(ctx, req)
		if err != nil {
			hbRst.ErrMessage = err.Error()
			log.Error(fmt.Sprintf("heartbeat set failed %s %s", element.ServiceId, element.InstanceId), err)
		}
		instancesHbRst <- hbRst
	}
}
