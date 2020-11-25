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

package mongo

import (
	"context"
	"fmt"
	"github.com/apache/servicecomb-service-center/datasource/mongo/client"
	"github.com/apache/servicecomb-service-center/datasource/mongo/heartbeat"
	"github.com/apache/servicecomb-service-center/pkg/gopool"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/util"
	"github.com/apache/servicecomb-service-center/server/plugin/uuid"
	pb "github.com/go-chassis/cari/discovery"
	"go.mongodb.org/mongo-driver/bson"
	"strconv"
	"time"
)

func (ds *DataSource) RegisterService(ctx context.Context, request *pb.CreateServiceRequest) (
	*pb.CreateServiceResponse, error) {
	service := request.Service
	SetServiceDefaultValue(service)

	domain := util.ParseDomain(ctx)
	project := util.ParseProject(ctx)
	//todo add quota check
	requestServiceID := service.ServiceId

	if len(requestServiceID) == 0 {
		ctx = util.SetContext(ctx, uuid.ContextKey, util.StringJoin([]string{domain, project, service.Environment, service.AppId, service.ServiceName, service.Alias, service.Version}, "/"))
		service.ServiceId = uuid.Generator().GetServiceID(ctx)
	}

	//filter := GeneratorServiceFilter(ctx, service.ServiceId)
	svc, err := GetService(ctx, service.ServiceId)
	if err != nil {
		log.Errorf(err, "check service %s exist failed", service.ServiceId)
		return &pb.CreateServiceResponse{
			Response: pb.CreateResponse(pb.ErrInternal, "check service exist failed"),
		}, nil
	}
	if svc != nil {
		return &pb.CreateServiceResponse{
			Response: pb.CreateResponse(pb.ErrServiceAlreadyExists, "ServiceID conflict"),
		}, nil
	}
	insertRes, err := client.GetMongoClient().Insert(ctx, TBL_SERVICE, &MongoService{Domain: domain, Project: project, Service: service})
	if err != nil {
		log.Errorf(err, "failed to insert service to mongodb")
		return &pb.CreateServiceResponse{
			Response: pb.CreateResponse(pb.ErrInternal, "insert service failed"),
		}, nil
	}

	remoteIP := util.GetIPFromContext(ctx)
	log.Infof("create micro-service[%s][%s] successfully,operator: %s", service.ServiceId, insertRes.InsertedID, remoteIP)

	return &pb.CreateServiceResponse{
		Response:  pb.CreateResponse(pb.ResponseSuccess, "Register service successfully"),
		ServiceId: service.ServiceId,
	}, nil
}

func (ds *DataSource) GetServices(ctx context.Context, request *pb.GetServicesRequest) (*pb.GetServicesResponse, error) {
	return &pb.GetServicesResponse{}, nil
}

func (ds *DataSource) GetService(ctx context.Context, request *pb.GetServiceRequest) (*pb.GetServiceResponse, error) {
	return &pb.GetServiceResponse{}, nil
}

func (ds *DataSource) GetServiceDetail(ctx context.Context, request *pb.GetServiceRequest) (*pb.GetServiceDetailResponse, error) {
	return &pb.GetServiceDetailResponse{}, nil
}

func (ds *DataSource) GetServicesInfo(ctx context.Context, request *pb.GetServicesInfoRequest) (*pb.GetServicesInfoResponse, error) {
	return &pb.GetServicesInfoResponse{}, nil
}

func (ds *DataSource) GetApplications(ctx context.Context, request *pb.GetAppsRequest) (*pb.GetAppsResponse, error) {
	return &pb.GetAppsResponse{}, nil
}

func (ds *DataSource) ExistServiceByID(ctx context.Context, request *pb.GetExistenceByIDRequest) (*pb.GetExistenceByIDResponse, error) {
	return &pb.GetExistenceByIDResponse{}, nil
}

func (ds *DataSource) ExistService(ctx context.Context, request *pb.GetExistenceRequest) (*pb.GetExistenceResponse, error) {
	return &pb.GetExistenceResponse{}, nil
}

func (ds *DataSource) UpdateService(ctx context.Context, request *pb.UpdateServicePropsRequest) (*pb.UpdateServicePropsResponse, error) {
	return &pb.UpdateServicePropsResponse{}, nil
}

func (ds *DataSource) UnregisterService(ctx context.Context, request *pb.DeleteServiceRequest) (*pb.DeleteServiceResponse, error) {
	return &pb.DeleteServiceResponse{}, nil
}

func (ds *DataSource) GetDeleteServiceFunc(ctx context.Context, serviceID string, force bool, serviceRespChan chan<- *pb.DelServicesRspInfo) func(context.Context) {
	return func(_ context.Context) {

	}
}

// Instance management
func (ds *DataSource) RegisterInstance(ctx context.Context, request *pb.RegisterInstanceRequest) (*pb.RegisterInstanceResponse, error) {
	remoteIP := util.GetIPFromContext(ctx)
	instance := request.Instance

	// 允许自定义 id
	if len(instance.InstanceId) > 0 {
		resp, err := ds.Heartbeat(ctx, &pb.HeartbeatRequest{
			InstanceId: instance.InstanceId,
			ServiceId:  instance.ServiceId,
		})
		if resp == nil {
			log.Errorf(err, "register service[%s]'s instance failed, endpoints %v, host '%s', operator %s",
				instance.ServiceId, instance.Endpoints, instance.HostName, remoteIP)
			return &pb.RegisterInstanceResponse{
				Response: pb.CreateResponse(pb.ErrInternal, err.Error()),
			}, nil
		}
		switch resp.Response.GetCode() {
		case pb.ResponseSuccess:
			log.Infof("register instance successful, reuse instance[%s/%s], operator %s",
				instance.ServiceId, instance.InstanceId, remoteIP)
			return &pb.RegisterInstanceResponse{
				Response:   resp.Response,
				InstanceId: instance.InstanceId,
			}, nil
		case pb.ErrInstanceNotExists:
			// register a new one
			return registryInstance(ctx, request)
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
			Response: pb.CreateResponseWithSCErr(err),
		}, nil
	}

	return registryInstance(ctx, request)
}

func registryInstance(ctx context.Context, request *pb.RegisterInstanceRequest) (*pb.RegisterInstanceResponse, error) {
	domain := util.ParseDomain(ctx)
	project := util.ParseProject(ctx)
	remoteIP := util.GetIPFromContext(ctx)
	instance := request.Instance
	instanceID := instance.InstanceId
	data := &MongoInstance{
		Domain:       domain,
		Project:      project,
		RefreshTime:  time.Now(),
		InstanceInfo: instance,
	}

	instanceFlag := fmt.Sprintf("endpoints %v, host '%s', serviceID %s",
		instance.Endpoints, instance.HostName, instance.ServiceId)

	insertRes, err := client.GetMongoClient().Insert(ctx, TBL_INSTANCE, data)
	if err != nil {
		log.Errorf(err,
			"register instance failed, %s, instanceID %s, operator %s",
			instanceFlag, instanceID, remoteIP)
		return &pb.RegisterInstanceResponse{
			Response: pb.CreateResponse(pb.ErrUnavailableBackend, err.Error()),
		}, err
	}

	log.Infof("register instance %s, instanceID %s, operator %s",
		instanceFlag, insertRes.InsertedID, remoteIP)
	return &pb.RegisterInstanceResponse{
		Response:   pb.CreateResponse(pb.ResponseSuccess, "Register service instance successfully."),
		InstanceId: instanceID,
	}, nil
}

// GetInstances returns instances under the current domain
func (ds *DataSource) GetInstance(ctx context.Context, request *pb.GetOneInstanceRequest) (*pb.GetOneInstanceResponse, error) {

	service := &MongoService{}
	var err error
	if len(request.ConsumerServiceId) > 0 {
		service, err = GetService(ctx, request.ConsumerServiceId)
		if err != nil {
			log.Errorf(err, "get consumer failed, consumer[%s] find provider instance[%s/%s]",
				request.ConsumerServiceId, request.ProviderServiceId, request.ProviderInstanceId)
			return &pb.GetOneInstanceResponse{
				Response: pb.CreateResponse(pb.ErrInternal, err.Error()),
			}, err
		}
		if service == nil {
			log.Errorf(nil, "consumer does not exist, consumer[%s] find provider instance[%s/%s]",
				request.ConsumerServiceId, request.ProviderServiceId, request.ProviderInstanceId)
			return &pb.GetOneInstanceResponse{
				Response: pb.CreateResponse(pb.ErrServiceNotExists,
					fmt.Sprintf("Consumer[%s] does not exist.", request.ConsumerServiceId)),
			}, nil
		}
	}

	provider, err := GetService(ctx, request.ProviderServiceId)
	if err != nil {
		log.Errorf(err, "get provider failed, consumer[%s] find provider instance[%s/%s]",
			request.ConsumerServiceId, request.ProviderServiceId, request.ProviderInstanceId)
		return &pb.GetOneInstanceResponse{
			Response: pb.CreateResponse(pb.ErrInternal, err.Error()),
		}, err
	}
	if provider == nil {
		log.Errorf(nil, "provider does not exist, consumer[%s] find provider instance[%s/%s]",
			request.ConsumerServiceId, request.ProviderServiceId, request.ProviderInstanceId)
		return &pb.GetOneInstanceResponse{
			Response: pb.CreateResponse(pb.ErrServiceNotExists,
				fmt.Sprintf("Provider[%s] does not exist.", request.ProviderServiceId)),
		}, nil
	}

	findFlag := func() string {
		return fmt.Sprintf("Consumer[%s][%s/%s/%s/%s] find provider[%s][%s/%s/%s/%s] instance[%s]",
			request.ConsumerServiceId, service.Service.Environment, service.Service.AppId, service.Service.ServiceName, service.Service.Version,
			provider.Service.ServiceId, provider.Service.Environment, provider.Service.AppId, provider.Service.ServiceName, provider.Service.Version,
			request.ProviderInstanceId)
	}

	domain := util.ParseDomain(ctx)
	project := util.ParseProject(ctx)
	filter := bson.M{"domain": domain, "project": project, "instanceinfo.serviceid": request.ProviderServiceId}
	findOneRes, err := client.GetMongoClient().FindOne(ctx, TBL_INSTANCE, filter)
	if err != nil {
		mes := fmt.Errorf("%s failed, provider instance does not exist", findFlag())
		log.Errorf(mes, "FindInstances.GetWithProviderID failed")
		return &pb.GetOneInstanceResponse{
			Response: pb.CreateResponse(pb.ErrInstanceNotExists, mes.Error()),
		}, nil
	}
	var instance MongoInstance
	err = findOneRes.Decode(&instance)
	if err != nil {
		log.Errorf(err, "FindInstances.GetWithProviderID failed, %s failed", findFlag())
		return &pb.GetOneInstanceResponse{
			Response: pb.CreateResponse(pb.ErrInternal, err.Error()),
		}, err
	}

	return &pb.GetOneInstanceResponse{
		Response: pb.CreateResponse(pb.ResponseSuccess, "Get instance successfully."),
		Instance: instance.InstanceInfo,
	}, nil
}

func (ds *DataSource) GetInstances(ctx context.Context, request *pb.GetInstancesRequest) (*pb.GetInstancesResponse, error) {

	service := &MongoService{}
	var err error

	if len(request.ConsumerServiceId) > 0 {
		service, err = GetService(ctx, request.ConsumerServiceId)
		if err != nil {
			log.Errorf(err, "get consumer failed, consumer[%s] find provider instances",
				request.ConsumerServiceId, request.ProviderServiceId)
			return &pb.GetInstancesResponse{
				Response: pb.CreateResponse(pb.ErrInternal, err.Error()),
			}, err
		}
		if service == nil {
			log.Errorf(nil, "consumer does not exist, consumer[%s] find provider instances",
				request.ConsumerServiceId, request.ProviderServiceId)
			return &pb.GetInstancesResponse{
				Response: pb.CreateResponse(pb.ErrServiceNotExists,
					fmt.Sprintf("Consumer[%s] does not exist.", request.ConsumerServiceId)),
			}, nil
		}
	}

	provider, err := GetService(ctx, request.ProviderServiceId)
	if err != nil {
		log.Errorf(err, "get provider failed, consumer[%s] find provider instances",
			request.ConsumerServiceId, request.ProviderServiceId)
		return &pb.GetInstancesResponse{
			Response: pb.CreateResponse(pb.ErrInternal, err.Error()),
		}, err
	}
	if provider == nil {
		log.Errorf(nil, "provider does not exist, consumer[%s] find provider instances",
			request.ConsumerServiceId, request.ProviderServiceId)
		return &pb.GetInstancesResponse{
			Response: pb.CreateResponse(pb.ErrServiceNotExists,
				fmt.Sprintf("Provider[%s] does not exist.", request.ProviderServiceId)),
		}, nil
	}

	findFlag := fmt.Sprintf("Consumer[%s][%s/%s/%s/%s] find provider[%s][%s/%s/%s/%s] instances",
		request.ConsumerServiceId, service.Service.Environment, service.Service.AppId, service.Service.ServiceName, service.Service.Version,
		provider.Service.ServiceId, provider.Service.Environment, provider.Service.AppId, provider.Service.ServiceName, provider.Service.Version)

	domain := util.ParseDomain(ctx)
	project := util.ParseProject(ctx)
	resp, err := client.GetMongoClient().Find(ctx, TBL_INSTANCE, bson.M{"domain": domain, "project": project, "instanceinfo.serviceid": request.ProviderServiceId})
	if err != nil {
		log.Errorf(err, "FindInstancesCache.Get failed, %s failed", findFlag)
		return &pb.GetInstancesResponse{
			Response: pb.CreateResponse(pb.ErrInternal, err.Error()),
		}, err
	}
	if resp == nil {
		mes := fmt.Errorf("%s failed, provider does not exist", findFlag)
		log.Errorf(mes, "FindInstancesCache.Get failed")
		return &pb.GetInstancesResponse{
			Response: pb.CreateResponse(pb.ErrServiceNotExists, mes.Error()),
		}, nil
	}

	var instances []*pb.MicroServiceInstance
	for resp.Next(ctx) {
		var instance MongoInstance
		err := resp.Decode(&instance)
		if err != nil {
			log.Errorf(err, "FindInstances.GetWithProviderID failed, %s failed", findFlag)
			return &pb.GetInstancesResponse{
				Response: pb.CreateResponse(pb.ErrInternal, err.Error()),
			}, err
		}
		instances = append(instances, instance.InstanceInfo)
	}

	return &pb.GetInstancesResponse{
		Response:  pb.CreateResponse(pb.ResponseSuccess, "Query service instances successfully."),
		Instances: instances,
	}, nil
}

// GetProviderInstances returns instances under the specified domain
func (ds *DataSource) GetProviderInstances(ctx context.Context, request *pb.GetProviderInstancesRequest) (instances []*pb.MicroServiceInstance, rev string, err error) {

	domain := util.ParseDomain(ctx)
	project := util.ParseProject(ctx)
	filter := bson.M{"domain": domain, "project": project, "instanceinfo.serviceid": request.ProviderServiceId}

	findRes, err := client.GetMongoClient().Find(ctx, TBL_INSTANCE, filter)
	if err != nil {
		return
	}

	for findRes.Next(ctx) {
		var mongoInstance MongoInstance
		err := findRes.Decode(&mongoInstance)
		if err == nil {
			instances = append(instances, mongoInstance.InstanceInfo)
		}
	}

	return instances, "", nil
}

func (ds *DataSource) reshapeProviderKey(ctx context.Context, provider *pb.MicroServiceKey, providerID string) (
	*pb.MicroServiceKey, error) {
	//维护version的规则,service name 可能是别名，所以重新获取
	providerService, err := GetService(ctx, providerID)
	if providerService == nil {
		return nil, err
	}

	versionRule := provider.Version
	provider = pb.MicroServiceToKey(provider.Tenant, providerService.Service)
	provider.Version = versionRule
	return provider, nil
}

func (ds *DataSource) findInstance(ctx context.Context, request *pb.FindInstancesRequest, provider *pb.MicroServiceKey) (*pb.FindInstancesResponse, error) {
	var err error
	domainProject := util.ParseDomainProject(ctx)
	service := &MongoService{Service: &pb.MicroService{Environment: request.Environment}}
	if len(request.ConsumerServiceId) > 0 {
		service, err = GetService(ctx, request.ConsumerServiceId)
		if err != nil {
			log.Errorf(err, "get consumer failed, consumer[%s] find provider[%s/%s/%s/%s]",
				request.ConsumerServiceId, request.Environment, request.AppId, request.ServiceName, request.VersionRule)
			return &pb.FindInstancesResponse{
				Response: pb.CreateResponse(pb.ErrInternal, err.Error()),
			}, err
		}
		if service == nil {
			log.Errorf(nil, "consumer does not exist, consumer[%s] find provider[%s/%s/%s/%s]",
				request.ConsumerServiceId, request.Environment, request.AppId, request.ServiceName, request.VersionRule)
			return &pb.FindInstancesResponse{
				Response: pb.CreateResponse(pb.ErrServiceNotExists,
					fmt.Sprintf("Consumer[%s] does not exist.", request.ConsumerServiceId)),
			}, nil
		}
		provider.Environment = service.Service.Environment
	}

	// provider is not a shared micro-service,
	// only allow shared micro-service instances found request different domains.
	ctx = util.SetTargetDomainProject(ctx, util.ParseDomain(ctx), util.ParseProject(ctx))
	provider.Tenant = util.ParseTargetDomainProject(ctx)

	findFlag := fmt.Sprintf("Consumer[%s][%s/%s/%s/%s] find provider[%s/%s/%s/%s]",
		request.ConsumerServiceId, service.Service.Environment, service.Service.AppId, service.Service.ServiceName, service.Service.Version,
		provider.Environment, provider.AppId, provider.ServiceName, provider.Version)

	domain := util.ParseDomain(ctx)
	project := util.ParseProject(ctx)
	resp, err := client.GetMongoClient().Find(ctx, TBL_INSTANCE, bson.M{"domain": domain, "project": project})
	if err != nil {
		log.Errorf(err, "FindInstancesCache.Get failed, %s failed", findFlag)
		return &pb.FindInstancesResponse{
			Response: pb.CreateResponse(pb.ErrInternal, err.Error()),
		}, err
	}
	if resp == nil {
		mes := fmt.Errorf("%s failed, provider does not exist", findFlag)
		log.Errorf(mes, "FindInstancesCache.Get failed")
		return &pb.FindInstancesResponse{
			Response: pb.CreateResponse(pb.ErrServiceNotExists, mes.Error()),
		}, nil
	}

	var instances []*pb.MicroServiceInstance
	for resp.Next(ctx) {
		var instance MongoInstance
		err := resp.Decode(&instance)
		if err != nil {
			log.Errorf(err, "FindInstances.GetWithProviderID failed, %s failed", findFlag)
			return &pb.FindInstancesResponse{
				Response: pb.CreateResponse(pb.ErrInternal, err.Error()),
			}, err
		}
		instances = append(instances, instance.InstanceInfo)
	}

	// add dependency queue
	if len(request.ConsumerServiceId) > 0 &&
		len(instances) > 0 {
		provider, err = ds.reshapeProviderKey(ctx, provider, instances[0].ServiceId)
		if err != nil {
			return nil, err
		}
		if provider != nil {
			err = AddServiceVersionRule(ctx, domainProject, service.Service, provider)
		} else {
			mes := fmt.Errorf("%s failed, provider does not exist", findFlag)
			log.Errorf(mes, "AddServiceVersionRule failed")
			return &pb.FindInstancesResponse{
				Response: pb.CreateResponse(pb.ErrServiceNotExists, mes.Error()),
			}, nil
		}
		if err != nil {
			log.Errorf(err, "AddServiceVersionRule failed, %s failed", findFlag)
			return &pb.FindInstancesResponse{
				Response: pb.CreateResponse(pb.ErrInternal, err.Error()),
			}, err
		}
	}

	return &pb.FindInstancesResponse{
		Response:  pb.CreateResponse(pb.ResponseSuccess, "Query service instances successfully."),
		Instances: instances,
	}, nil
}

func (ds *DataSource) BatchGetProviderInstances(ctx context.Context, request *pb.BatchGetInstancesRequest) (instances []*pb.MicroServiceInstance, rev string, err error) {
	if request == nil || len(request.ServiceIds) == 0 {
		return nil, "", fmt.Errorf("invalid param BatchGetInstancesRequest")
	}

	domain := util.ParseDomain(ctx)
	project := util.ParseProject(ctx)

	for _, providerServiceID := range request.ServiceIds {
		// todo  finish find instances

		filter := bson.M{"domain": domain, "project": project, "instanceinfo.serviceid": providerServiceID}
		findRes, err := client.GetMongoClient().Find(ctx, TBL_INSTANCE, filter)
		if err != nil {
			return instances, "", nil
		}

		for findRes.Next(ctx) {
			var mongoInstance MongoInstance
			err := findRes.Decode(&mongoInstance)
			if err == nil {
				instances = append(instances, mongoInstance.InstanceInfo)
			}
		}
	}

	return instances, "", nil
}

// FindInstances returns instances under the specified domain
func (ds *DataSource) FindInstances(ctx context.Context, request *pb.FindInstancesRequest) (*pb.FindInstancesResponse, error) {

	provider := &pb.MicroServiceKey{
		Tenant:      util.ParseTargetDomainProject(ctx),
		Environment: request.Environment,
		AppId:       request.AppId,
		ServiceName: request.ServiceName,
		Alias:       request.ServiceName,
		Version:     request.VersionRule,
	}

	return ds.findInstance(ctx, request, provider)
}

func (ds *DataSource) UpdateInstanceStatus(ctx context.Context, request *pb.UpdateInstanceStatusRequest) (*pb.UpdateInstanceStatusResponse, error) {

	updateStatusFlag := util.StringJoin([]string{request.ServiceId, request.InstanceId, request.Status}, "/")

	// todo finish get instance
	instance, err := GetInstance(ctx, request.ServiceId, request.InstanceId)
	if err != nil {
		log.Errorf(err, "update instance[%s] status failed", updateStatusFlag)
		return &pb.UpdateInstanceStatusResponse{
			Response: pb.CreateResponse(pb.ErrInternal, err.Error()),
		}, err
	}
	if instance == nil {
		log.Errorf(nil, "update instance[%s] status failed, instance does not exist", updateStatusFlag)
		return &pb.UpdateInstanceStatusResponse{
			Response: pb.CreateResponse(pb.ErrInstanceNotExists, "Service instance does not exist."),
		}, nil
	}

	copyInstanceRef := *instance
	copyInstanceRef.InstanceInfo.Status = request.Status

	if err := UpdateInstanceS(ctx, copyInstanceRef.InstanceInfo); err != nil {
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

func (ds *DataSource) UpdateInstanceProperties(ctx context.Context, request *pb.UpdateInstancePropsRequest) (*pb.UpdateInstancePropsResponse, error) {

	instanceFlag := util.StringJoin([]string{request.ServiceId, request.InstanceId}, "/")

	instance, err := GetInstance(ctx, request.ServiceId, request.InstanceId)
	if err != nil {
		log.Errorf(err, "update instance[%s] properties failed", instanceFlag)
		return &pb.UpdateInstancePropsResponse{
			Response: pb.CreateResponse(pb.ErrInternal, err.Error()),
		}, err
	}
	if instance == nil {
		log.Errorf(nil, "update instance[%s] properties failed, instance does not exist", instanceFlag)
		return &pb.UpdateInstancePropsResponse{
			Response: pb.CreateResponse(pb.ErrInstanceNotExists, "Service instance does not exist."),
		}, nil
	}

	copyInstanceRef := *instance
	copyInstanceRef.InstanceInfo.Properties = request.Properties

	// todo finish update instance
	if err := UpdateInstanceP(ctx, copyInstanceRef.InstanceInfo); err != nil {
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

func (ds *DataSource) UnregisterInstance(ctx context.Context, request *pb.UnregisterInstanceRequest) (*pb.UnregisterInstanceResponse, error) {

	remoteIP := util.GetIPFromContext(ctx)
	serviceID := request.ServiceId
	instanceID := request.InstanceId

	instanceFlag := util.StringJoin([]string{serviceID, instanceID}, "/")

	// todo finish revoke instance
	//err := revokeInstance(ctx, "", serviceID, instanceID)
	//if err != nil {
	//	log.Errorf(err, "unregister instance failed, instance[%s], operator %s: revoke instance failed",
	//		instanceFlag, remoteIP)
	//	resp := &pb.UnregisterInstanceResponse{
	//		Response: pb.CreateResponseWithSCErr(err),
	//	}
	//	if err.InternalError() {
	//		return resp, err
	//	}
	//	return resp, nil
	//}

	domain := util.ParseDomain(ctx)
	project := util.ParseProject(ctx)
	filter := bson.M{"domain": domain, "project": project, "instance.serviceid": serviceID, "instance.instanceid": instanceID}
	_, err := client.GetMongoClient().Delete(ctx, TBL_INSTANCE, filter)
	if err != nil {
		log.Errorf(err, "unregister instance failed, instance[%s], operator %s: revoke instance failed",
			instanceFlag, remoteIP)
		return &pb.UnregisterInstanceResponse{
			Response: pb.CreateResponse(pb.ErrInternal, "delete instance failed"),
		}, err
	}

	log.Infof("unregister instance[%s], operator %s", instanceFlag, remoteIP)
	return &pb.UnregisterInstanceResponse{
		Response: pb.CreateResponse(pb.ResponseSuccess, "Unregister service instance successfully."),
	}, nil
}

func (ds *DataSource) Heartbeat(ctx context.Context, request *pb.HeartbeatRequest) (*pb.HeartbeatResponse, error) {
	remoteIP := util.GetIPFromContext(ctx)
	instanceFlag := util.StringJoin([]string{request.ServiceId, request.InstanceId}, "/")
	err := KeepAliveLease(ctx, request)
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
	return &pb.HeartbeatResponse{
		Response: pb.CreateResponse(pb.ResponseSuccess,
			"Update service instance heartbeat successfully."),
	}, nil
}

func KeepAliveLease(ctx context.Context, request *pb.HeartbeatRequest) *pb.Error {
	_, err := heartbeat.Instance().Heartbeat(ctx, request)
	if err != nil {
		return pb.NewError(pb.ErrInstanceNotExists, err.Error())
	}
	return nil
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
		log.Infof("batch update heartbeats[%d] successfully", count)
		return &pb.HeartbeatSetResponse{
			Response:  pb.CreateResponse(pb.ResponseSuccess, "Heartbeat set successfully."),
			Instances: instanceHbRstArr,
		}, nil
	}

	log.Errorf(nil, "batch update heartbeats failed, %v", request.Instances)
	return &pb.HeartbeatSetResponse{
		Response:  pb.CreateResponse(pb.ErrInstanceNotExists, "Heartbeat set failed."),
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

		req := &pb.HeartbeatRequest{
			InstanceId: element.InstanceId,
			ServiceId:  element.ServiceId,
		}

		err := KeepAliveLease(ctx, req)
		if err != nil {
			hbRst.ErrMessage = err.Error()
			log.Errorf(err, "heartbeat set failed, %s/%s", element.ServiceId, element.InstanceId)
		}
		instancesHbRst <- hbRst
	}
}

func (ds *DataSource) BatchFind(ctx context.Context, request *pb.BatchFindInstancesRequest) (*pb.BatchFindInstancesResponse, error) {

	response := &pb.BatchFindInstancesResponse{
		Response: pb.CreateResponse(pb.ResponseSuccess, "Batch query service instances successfully."),
	}

	var err error

	response.Services, err = ds.batchFindServices(ctx, request)
	if err != nil {
		return &pb.BatchFindInstancesResponse{
			Response: pb.CreateResponse(pb.ErrInternal, err.Error()),
		}, err
	}

	response.Instances, err = ds.batchFindInstances(ctx, request)
	if err != nil {
		return &pb.BatchFindInstancesResponse{
			Response: pb.CreateResponse(pb.ErrInternal, err.Error()),
		}, err
	}

	return response, nil
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
		AppendFindResponse(getCtx, int64(index), resp.Response, []*pb.MicroServiceInstance{resp.Instance},
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
		AppendFindResponse(findCtx, int64(index), resp.Response, resp.Instances,
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

func (ds *DataSource) modifySchemas(ctx context.Context, domainProject string, service *pb.MicroService,
	schemas []*pb.Schema) *pb.Error {
	return nil
}

func getSchemasFromDatabase(ctx context.Context, domainProject string, serviceID string) ([]*pb.Schema, error) {
	// todo  table unknown
	return nil, nil
}

func (ds *DataSource) GetAllInstances(ctx context.Context, request *pb.GetAllInstancesRequest) (*pb.GetAllInstancesResponse, error) {
	return &pb.GetAllInstancesResponse{}, nil
}

// Schema management
func (ds *DataSource) ModifySchemas(ctx context.Context, request *pb.ModifySchemasRequest) (*pb.ModifySchemasResponse, error) {
	return &pb.ModifySchemasResponse{}, nil
}

func (ds *DataSource) ModifySchema(ctx context.Context, request *pb.ModifySchemaRequest) (*pb.ModifySchemaResponse, error) {
	return &pb.ModifySchemaResponse{}, nil
}

func (ds *DataSource) ExistSchema(ctx context.Context, request *pb.GetExistenceRequest) (*pb.GetExistenceResponse, error) {
	return &pb.GetExistenceResponse{}, nil
}

func (ds *DataSource) GetSchema(ctx context.Context, request *pb.GetSchemaRequest) (*pb.GetSchemaResponse, error) {
	return &pb.GetSchemaResponse{}, nil
}

func (ds *DataSource) GetAllSchemas(ctx context.Context, request *pb.GetAllSchemaRequest) (*pb.GetAllSchemaResponse, error) {
	return &pb.GetAllSchemaResponse{}, nil
}

func (ds *DataSource) DeleteSchema(ctx context.Context, request *pb.DeleteSchemaRequest) (*pb.DeleteSchemaResponse, error) {
	return &pb.DeleteSchemaResponse{}, nil
}

// Tag management
func (ds *DataSource) AddTags(ctx context.Context, request *pb.AddServiceTagsRequest) (*pb.AddServiceTagsResponse, error) {
	return &pb.AddServiceTagsResponse{}, nil
}

func (ds *DataSource) GetTags(ctx context.Context, request *pb.GetServiceTagsRequest) (*pb.GetServiceTagsResponse, error) {
	return &pb.GetServiceTagsResponse{}, nil
}

func (ds *DataSource) UpdateTag(ctx context.Context, request *pb.UpdateServiceTagRequest) (*pb.UpdateServiceTagResponse, error) {
	return &pb.UpdateServiceTagResponse{}, nil
}

func (ds *DataSource) DeleteTags(ctx context.Context, request *pb.DeleteServiceTagsRequest) (*pb.DeleteServiceTagsResponse, error) {
	return &pb.DeleteServiceTagsResponse{}, nil
}

// White/black list management
func (ds *DataSource) AddRule(ctx context.Context, request *pb.AddServiceRulesRequest) (*pb.AddServiceRulesResponse, error) {
	return &pb.AddServiceRulesResponse{}, nil
}

func (ds *DataSource) GetRules(ctx context.Context, request *pb.GetServiceRulesRequest) (*pb.GetServiceRulesResponse, error) {
	return &pb.GetServiceRulesResponse{}, nil
}

func (ds *DataSource) UpdateRule(ctx context.Context, request *pb.UpdateServiceRuleRequest) (*pb.UpdateServiceRuleResponse, error) {
	return &pb.UpdateServiceRuleResponse{}, nil
}

func (ds *DataSource) DeleteRule(ctx context.Context, request *pb.DeleteServiceRulesRequest) (*pb.DeleteServiceRulesResponse, error) {
	return &pb.DeleteServiceRulesResponse{}, nil
}

func GetService(ctx context.Context, serviceID string) (*MongoService, error) {
	filter := GeneratorServiceFilter(ctx, serviceID)

	findRes, err := client.GetMongoClient().Find(ctx, TBL_SERVICE, filter)
	if err != nil {
		return nil, err
	}
	var service *MongoService
	for findRes.Next(ctx) {
		err := findRes.Decode(&service)
		if err != nil {
			return nil, err
		}
		return service, nil
	}
	return nil, nil
}

func GeneratorServiceFilter(ctx context.Context, serviceID string) bson.M {
	domain := util.ParseDomain(ctx)
	project := util.ParseProject(ctx)

	return bson.M{FLD_SERVICE_DOMAIN: domain, FLD_SERVICE_PROJECT: project, FLD_SERVICE_ID: serviceID}
}

func GetInstance(ctx context.Context, serviceID string, instanceID string) (*MongoInstance, error) {
	domain := util.ParseDomain(ctx)
	project := util.ParseProject(ctx)
	filter := bson.M{"domain": domain, "project": project, "instanceinfo.serviceid": serviceID, "instanceinfo.instanceid": instanceID}
	findRes, err := client.GetMongoClient().Find(ctx, TBL_INSTANCE, filter)
	if err != nil {
		return nil, err
	}
	var service *MongoInstance
	for findRes.Next(ctx) {
		err := findRes.Decode(&service)
		if err != nil {
			return nil, err
		}
		return service, nil
	}
	return nil, nil
}

func UpdateInstanceS(ctx context.Context, instance *pb.MicroServiceInstance) *pb.Error {
	domain := util.ParseDomain(ctx)
	project := util.ParseProject(ctx)
	filter := bson.M{"domain": domain, "project": project, "instance.instanceid": instance.InstanceId, "instance.serviceid": instance.ServiceId}
	_, err := client.GetMongoClient().Update(ctx, TBL_INSTANCE, filter, bson.M{"$set": bson.M{"instance.motTimestamp": strconv.FormatInt(time.Now().Unix(), 10), "instance.status": instance.Status}})
	if err != nil {
		return pb.NewError(pb.ErrUnavailableBackend, err.Error())
	}
	return nil
}

func UpdateInstanceP(ctx context.Context, instance *pb.MicroServiceInstance) *pb.Error {
	domain := util.ParseDomain(ctx)
	project := util.ParseProject(ctx)
	filter := bson.M{"domain": domain, "project": project, "instance.instanceid": instance.InstanceId, "instance.serviceid": instance.ServiceId}
	_, err := client.GetMongoClient().Update(ctx, TBL_INSTANCE, filter, bson.M{"$set": bson.M{"instance.motTimestamp": strconv.FormatInt(time.Now().Unix(), 10), "instance.properties": instance.Properties}})
	if err != nil {
		return pb.NewError(pb.ErrUnavailableBackend, err.Error())
	}
	return nil
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

func SetServiceDefaultValue(service *pb.MicroService) {
	if len(service.AppId) == 0 {
		service.AppId = pb.AppID
	}
	if len(service.Version) == 0 {
		service.Version = pb.VERSION
	}
	if len(service.Level) == 0 {
		service.Level = "BACK"
	}
	if len(service.Status) == 0 {
		service.Status = pb.MS_UP
	}
}

func AddServiceVersionRule(ctx context.Context, domainProject string, consumer *pb.MicroService, provider *pb.MicroServiceKey) error {
	return nil
}
