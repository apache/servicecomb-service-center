//Copyright 2017 Huawei Technologies Co., Ltd
//
//Licensed under the Apache License, Version 2.0 (the "License");
//you may not use this file except in compliance with the License.
//You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
//Unless required by applicable law or agreed to in writing, software
//distributed under the License is distributed on an "AS IS" BASIS,
//WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//See the License for the specific language governing permissions and
//limitations under the License.
package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/astaxie/beego"
	apt "github.com/servicecomb/service-center/server/core"
	"github.com/servicecomb/service-center/server/core/mux"
	pb "github.com/servicecomb/service-center/server/core/proto"
	"github.com/servicecomb/service-center/server/core/registry"
	"github.com/servicecomb/service-center/server/infra/quota"
	nf "github.com/servicecomb/service-center/server/service/notification"
	serviceUtil "github.com/servicecomb/service-center/server/service/util"
	"github.com/servicecomb/service-center/util"
	"github.com/gorilla/websocket"
	"golang.org/x/net/context"
	"math"
	"strconv"
	"time"
)

type InstanceController struct {
	NotifyService *nf.NotifyService

	serviceCtrl ServiceController
}

func (s *InstanceController) Register(ctx context.Context, in *pb.RegisterInstanceRequest) (*pb.RegisterInstanceResponse, error) {
	if in == nil || in.Instance == nil {
		util.LOGGER.Errorf(nil, "register instance failed for instance empty.")
		return &pb.RegisterInstanceResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "request format invalid"),
		}, nil
	}
	instance := in.GetInstance()
	err := apt.Validate(instance)
	if err != nil {
		util.LOGGER.Errorf(err, "register instance failed for instance parameters valiedate failed.")
		return &pb.RegisterInstanceResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, err.Error()),
		}, nil
	}
	//先以tenant/project的方式组装
	tenant := util.ParaseTenant(ctx)

	// service id存在性校验
	if !s.serviceCtrl.ServiceExist(ctx, tenant, instance.ServiceId) {
		util.LOGGER.Errorf(nil, "register instance %s failed for service %s not exist.", instance.HostName, instance.ServiceId)
		return &pb.RegisterInstanceResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "service does not exist"),
		}, nil
	}

	instanceId := instance.InstanceId
	//允许自定义id
	//如果没填写 并且endpoints沒重復，則产生新的全局instance id
	oldInstanceId := ""

	if instanceId == "" {
		util.LOGGER.Infof("Start register a new instance for service %s", in.Instance.ServiceId)
		if len(instance.Endpoints) != 0 {
			oldInstanceId, err = serviceUtil.CheckEndPoints(ctx, in)
			if err != nil {
				return &pb.RegisterInstanceResponse{
					Response: pb.CreateResponse(pb.Response_FAIL, err.Error()),
				}, nil
			}
			if oldInstanceId != "" {
				instanceId = oldInstanceId
				instance.InstanceId = instanceId
			}
		}
	}

	if len(oldInstanceId) == 0 {
		ok, err := quota.QuotaPlugins[beego.AppConfig.DefaultString("quota_plugin", "buildin")]().Apply4Quotas(ctx, quota.MicroServiceInstanceQuotaType, 0)
		if err != nil {
			util.LOGGER.Errorf(err, "Check apply quota to create instance failed.")
			return &pb.RegisterInstanceResponse{
				Response: pb.CreateResponse(pb.Response_FAIL, err.Error()),
			}, err
		}
		if !ok {
			util.LOGGER.Errorf(err, "No quota to create instance.")
			return &pb.RegisterInstanceResponse{
				Response: pb.CreateResponse(pb.Response_FAIL, "No quota to create instance."),
			}, nil
		}
	}

	if len(instanceId) == 0 {
		respId, err := registry.GetRegisterCenter().Do(ctx, &registry.PluginOp{
			Action:     registry.PUT,
			Key:        []byte(apt.GetInstanceSequenceKey()),
			WithPrevKV: true,
		})
		if err != nil {
			util.LOGGER.Errorf(err, "register instance %s(%s) failed for generating instance id failed.", instance.ServiceId, instance.HostName)
			return &pb.RegisterInstanceResponse{
				Response: pb.CreateResponse(pb.Response_FAIL, "generate instance id error"),
			}, err
		}
		instanceId = "1"
		if respId.PrevKv != nil {
			instanceId = strconv.FormatInt(respId.PrevKv.Version+1, 10)
		}
		instance.InstanceId = instanceId
	}

	instance.Timestamp = strconv.FormatInt(time.Now().Unix(), 10)
	util.LOGGER.Debug(fmt.Sprintf("instance ID [%s]", instanceId))

	// 这里应该根据租约计时
	renewalInterval := apt.REGISTRY_DEFAULT_LEASE_RENEWALINTERVAL
	retryTimes := apt.REGISTRY_DEFAULT_LEASE_RETRYTIMES
	if instance.GetHealthCheck() == nil {
		instance.HealthCheck = &pb.HealthCheck{
			Mode:     pb.CHECK_BY_HEARTBEAT,
			Interval: renewalInterval,
			Times:    retryTimes,
		}
	} else {
		// Health check对象仅用于呈现服务健康检查逻辑，如果CHECK_BY_PLATFORM类型，表明由sidecar代发心跳，实例120s超时
		switch instance.HealthCheck.Mode {
		case pb.CHECK_BY_HEARTBEAT:
			if instance.HealthCheck.Interval <= 0 || instance.HealthCheck.Interval >= math.MaxInt32 ||
				instance.HealthCheck.Times <= 0 || instance.HealthCheck.Times >= math.MaxInt32 {
				util.LOGGER.Errorf(err, "register instance %s(%s) failed for invalid health check settings.", instance.ServiceId, instance.HostName)
				return &pb.RegisterInstanceResponse{
					Response: pb.CreateResponse(pb.Response_FAIL, "invalid health check settings"),
				}, nil
			}
			renewalInterval = instance.HealthCheck.Interval
			retryTimes = instance.HealthCheck.Times
		case pb.CHECK_BY_PLATFORM:
			// 默认120s
		}
	}

	ttl := int64(renewalInterval * (retryTimes + 1))
	var leaseID int64 = 0
	if ttl > 0 {
		leaseID, err = registry.GetRegisterCenter().LeaseGrant(ctx, ttl)
		if err != nil {
			util.LOGGER.Errorf(err, "register instance %s.%s(%s) failed for granting lease %ds failed.", instance.ServiceId, instance.InstanceId, instance.HostName, ttl)
			return &pb.RegisterInstanceResponse{
				Response: pb.CreateResponse(pb.Response_FAIL, "grant lease ID failed"),
			}, err
		}
	}

	data, err := json.Marshal(instance)
	if err != nil {
		util.LOGGER.Errorf(err, "register instance %s.%s(%s) failed for marshalling instance failed.", instance.ServiceId, instance.InstanceId, instance.HostName)
		return &pb.RegisterInstanceResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "instance file marshal error"),
		}, err
	}

	index := apt.GenerateInstanceIndexKey(tenant, instanceId)
	key := apt.GenerateInstanceKey(tenant, instance.ServiceId, instanceId)
	hbKey := apt.GenerateInstanceLeaseKey(tenant, instance.ServiceId, instanceId)

	util.LOGGER.Debugf("start register service instance: %s %v, lease: %s %ds", key, instance, hbKey, ttl)

	opts := []*registry.PluginOp{
		{
			Action: registry.PUT,
			Key:    []byte(key),
			Value:  data,
			Lease:  leaseID,
		},
		{
			Action: registry.PUT,
			Key:    []byte(index),
			Value:  []byte(instance.ServiceId),
			Lease:  leaseID,
		},
	}
	if leaseID != 0 {
		opts = append(opts, &registry.PluginOp{
			Action: registry.PUT,
			Key:    []byte(hbKey),
			Value:  []byte(fmt.Sprintf("%d", leaseID)),
			Lease:  leaseID,
		})
	}

	// Set key file
	_, err = registry.GetRegisterCenter().Txn(ctx, opts)
	if err != nil {
		util.LOGGER.Errorf(err, "register instance %s.%s(%s) failed for commiting operations failed.", instance.ServiceId, instance.HostName, instance.InstanceId)
		return &pb.RegisterInstanceResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "commit operations failed"),
		}, err
	}

	tenant = ctx.Value("tenant").(string)

	//新租户，则进行监听
	opt := &registry.PluginOp{
		Key:    []byte(apt.GenerateTenantKey(tenant)),
		Action: registry.PUT,
	}
	_, err = registry.GetRegisterCenter().PutNoOverride(ctx, opt)
	if err != nil {
		return &pb.RegisterInstanceResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "commit operations failed"),
		}, err
	}

	util.LOGGER.Warnf(nil, "register instance for service %s,  %s(%s) successfully from remote %s.", instance.ServiceId, instance.HostName, instance.InstanceId, util.GetIPFromContext(ctx))
	return &pb.RegisterInstanceResponse{
		Response:   pb.CreateResponse(pb.Response_SUCCESS, "register service instance successfully"),
		InstanceId: instanceId,
	}, nil
}

func (s *InstanceController) Unregister(ctx context.Context, in *pb.UnregisterInstanceRequest) (*pb.UnregisterInstanceResponse, error) {
	if in == nil || len(in.ServiceId) == 0 || len(in.InstanceId) == 0 {
		util.LOGGER.Errorf(nil, "unregister instance failed for instance empty.")
		return &pb.UnregisterInstanceResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "request format invalid"),
		}, nil
	}

	tenant := util.ParaseTenant(ctx)
	serviceId := in.ServiceId
	instanceId := in.InstanceId

	isExist, err := serviceUtil.InstanceExist(ctx, tenant, serviceId, instanceId)
	if err != nil {
		util.LOGGER.Errorf(err, "unregister instance %s.%s failed for querying instance failed.", in.ServiceId, in.InstanceId)
		return &pb.UnregisterInstanceResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "query service instance failed"),
		}, err
	}
	if !isExist {
		util.LOGGER.Errorf(nil, "unregister instance %s.%s failed for instance not exist.", in.ServiceId, in.InstanceId)
		return &pb.UnregisterInstanceResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "service instance does not exist"),
		}, nil
	}

	leaseID, err := serviceUtil.GetLeaseId(ctx, tenant, serviceId, instanceId)
	if leaseID == -1 && err == nil {
		util.LOGGER.Errorf(nil, "unregister instance %s.%s failed for instance not exist.", in.ServiceId, in.InstanceId)
		return &pb.UnregisterInstanceResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "service instance does not exist"),
		}, nil
	}
	if err != nil {
		util.LOGGER.Errorf(err, "unregister instance %s.%s failed for querying instance lease failed.", in.ServiceId, in.InstanceId)
		return &pb.UnregisterInstanceResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "query service instance lease failed"),
		}, err
	}

	err = registry.GetRegisterCenter().LeaseRevoke(ctx, leaseID)
	if err != nil {
		util.LOGGER.Errorf(err, "unregister instance %s.%s failed for committing operations failed.", in.ServiceId, in.InstanceId)
		return &pb.UnregisterInstanceResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "commit operations failed"),
		}, err
	}

	util.LOGGER.Warnf(err, "unregister instance %s.%s successfully from remote %s.", in.ServiceId, in.InstanceId, util.GetIPFromContext(ctx))
	return &pb.UnregisterInstanceResponse{
		Response: pb.CreateResponse(pb.Response_SUCCESS, "unregister service instance successfully"),
	}, nil
}

func (s *InstanceController) Heartbeat(ctx context.Context, in *pb.HeartbeatRequest) (*pb.HeartbeatResponse, error) {
	if in == nil || len(in.ServiceId) == 0 || len(in.InstanceId) == 0 {
		return &pb.HeartbeatResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "request format invalid"),
		}, nil
	}

	tenant := util.ParaseTenant(ctx)

	leaseID, err := serviceUtil.GetLeaseId(ctx, tenant, in.ServiceId, in.InstanceId)
	if leaseID == -1 && err == nil {
		util.LOGGER.Errorf(nil, "unregister instance %s.%s failed for instance not exist.", in.ServiceId, in.InstanceId)
		return &pb.HeartbeatResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "service instance does not exist"),
		}, nil
	}
	if err != nil {
		util.LOGGER.Errorf(err, "unregister instance %s.%s failed for querying instance lease failed.", in.ServiceId, in.InstanceId)
		return &pb.HeartbeatResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "query service instance lease failed"),
		}, err
	}

	ttl, err := registry.GetRegisterCenter().LeaseRenew(ctx, leaseID)
	if err != nil {
		return &pb.HeartbeatResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "update service instance lease failed"),
		}, err
	}
	util.LOGGER.Warnf(nil, "service %s instance %s renew ttl to %d", in.ServiceId, in.InstanceId, ttl)
	return &pb.HeartbeatResponse{
		Response: pb.CreateResponse(pb.Response_SUCCESS, "update service instance heartbeat successfully"),
	}, nil
}

func (s *InstanceController) GetOneInstance(ctx context.Context, in *pb.GetOneInstanceRequest) (*pb.GetOneInstanceResponse, error) {
	ok, err, isInternalErr := s.getInstancePreCheck(ctx, in)
	if err != nil {
		if isInternalErr {
			util.LOGGER.Errorf(err, "Get instance pre check failed.")
			return &pb.GetOneInstanceResponse{
				Response: pb.CreateResponse(pb.Response_FAIL, err.Error()),
			}, err
		}
		util.LOGGER.Errorf(err, "Get instance pre check failed.")
		return &pb.GetOneInstanceResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, err.Error()),
		}, nil
	}
	if !ok {
		util.LOGGER.Errorf(err, "Can not access this service's instance for your consumer.")
		return &pb.GetOneInstanceResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Can not access this service's instance for your consumer."),
		}, nil
	}

	tenant := util.ParaseTenant(ctx)
	errAddDependence := s.addDependenceForService(ctx, tenant, in.ConsumerServiceId, in.ProviderServiceId)
	if errAddDependence != nil {
		util.LOGGER.Errorf(err, "Add dependency for consumer:%s , provider: %s", in.ConsumerServiceId, in.ProviderServiceId)
		return &pb.GetOneInstanceResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Add dependency failed."),
		}, err
	}

	serviceId := in.ProviderServiceId
	instanceId := in.ProviderInstanceId
	instance, err := serviceUtil.GetInstance(ctx, tenant, serviceId, instanceId)
	if instance == nil && err == nil {
		util.LOGGER.Errorf(nil, fmt.Sprintf("%s service 's %s instance not exist.", serviceId, instanceId))
		return &pb.GetOneInstanceResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "No this instance"),
		}, nil
	}
	if err != nil {
		util.LOGGER.Errorf(err, "update properties of instance %s.%s failed for get instance failed.", serviceId, instanceId)
		return &pb.GetOneInstanceResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Update service instance properties failed"),
		}, err
	}

	if len(in.Stage) != 0 && in.Stage != instance.Stage {
		util.LOGGER.Errorf(err, "This instance %s 's stage %s does not match %s.", instanceId, in.Stage, instance.Stage)
		return &pb.GetOneInstanceResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Stage does not match, can't access this instance."),
		}, nil
	}

	return &pb.GetOneInstanceResponse{
		Response: pb.CreateResponse(pb.Response_SUCCESS, "Get instance successful."),
		Instance: instance,
	}, nil
}

func (s *InstanceController) getInstancePreCheck(ctx context.Context, in interface{}) (bool, error, bool) {
	err := apt.Validate(in)
	if err != nil {
		return false, err, false
	}
	var providerServiceId, consumerServiceId string
	var tags []string
	tenant := util.ParaseTenant(ctx)
	if in == nil {
		return false, errors.New("Request format invalid."), false
	}
	switch in.(type) {
	case *pb.GetOneInstanceRequest:
		providerServiceId = in.(*pb.GetOneInstanceRequest).ProviderServiceId
		consumerServiceId = in.(*pb.GetOneInstanceRequest).ConsumerServiceId
		tags = in.(*pb.GetOneInstanceRequest).Tags
	case *pb.GetInstancesRequest:
		providerServiceId = in.(*pb.GetInstancesRequest).ProviderServiceId
		consumerServiceId = in.(*pb.GetInstancesRequest).ConsumerServiceId
		tags = in.(*pb.GetInstancesRequest).Tags
	}

	if !s.serviceCtrl.ServiceExist(ctx, tenant, providerServiceId) {
		return false, errors.New(fmt.Sprintf("Service does not exist.Service id is %s", providerServiceId)), false
	}

	// Tag过滤
	if len(tags) > 0 {
		tagsFromETCD, err := GetTagsUtils(ctx, tenant, providerServiceId)
		if err != nil {
			return false, err, true
		}
		for _, tag := range tags {
			if _, ok := tagsFromETCD[tag]; !ok {
				return false, errors.New(fmt.Sprintf("Instance's tag not contain %s", tag)), false
			}
		}
	}
	// TODO 服务分层管理
	// 黑白名单
	// 跨应用调用
	return Accessible(ctx, tenant, consumerServiceId, providerServiceId)
}

func (s *InstanceController) GetInstances(ctx context.Context, in *pb.GetInstancesRequest) (*pb.GetInstancesResponse, error) {
	ok, err, isInternalErr := s.getInstancePreCheck(ctx, in)
	if err != nil {
		if isInternalErr {
			util.LOGGER.Errorf(err, "Get instances pre check failed.")
			return &pb.GetInstancesResponse{
				Response: pb.CreateResponse(pb.Response_FAIL, err.Error()),
			}, err
		}
		util.LOGGER.Errorf(err, "Get instances pre check failed.")
		return &pb.GetInstancesResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, err.Error()),
		}, nil
	}
	if !ok {
		util.LOGGER.Errorf(err, "Can not access this service's instances for your consumer.")
		return &pb.GetInstancesResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Can not access this service's instances for your consumer."),
		}, nil
	}

	tenant := util.ParaseTenant(ctx)
	errAddDependence := s.addDependenceForService(ctx, tenant, in.ConsumerServiceId, in.ProviderServiceId)
	if errAddDependence != nil {
		util.LOGGER.Errorf(errAddDependence, "Add dependency for consumer:%s , provider: %s", in.ConsumerServiceId, in.ProviderServiceId)
		return &pb.GetInstancesResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, errAddDependence.Error()),
		}, nil
	}

	instances := []*pb.MicroServiceInstance{}
	instances, err = getAllInstancesOfOneService(ctx, tenant, in.ProviderServiceId, in.Stage)
	if err != nil {
		util.LOGGER.Errorf(err, "Get instance of service %s failed.", in.ProviderServiceId)
		return &pb.GetInstancesResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, errAddDependence.Error()),
		}, nil
	}
	return &pb.GetInstancesResponse{
		Response:  pb.CreateResponse(pb.Response_SUCCESS, "query service instances successfully"),
		Instances: instances,
	}, nil
}

func getAllInstancesOfOneService(ctx context.Context, tenant string, serviceId string, stage string) ([]*pb.MicroServiceInstance, error) {
	key := apt.GenerateInstanceKey(tenant, serviceId, "")

	resp, err := registry.GetRegisterCenter().Do(ctx, &registry.PluginOp{
		Action:     registry.GET,
		Key:        []byte(key),
		WithPrefix: true,
	})
	if err != nil {
		util.LOGGER.Errorf(err, "Get instance of service %s from etcd failed.", serviceId)
		return nil, err
	}

	instances := []*pb.MicroServiceInstance{}
	for _, kvs := range resp.Kvs {
		util.LOGGER.Debugf("start unmarshal service instance file: %s", string(kvs.Key))
		instance := &pb.MicroServiceInstance{}
		err := json.Unmarshal(kvs.Value, instance)
		if err != nil {
			util.LOGGER.Errorf(err, "Unmarshal instance of service %s failed.", serviceId)
			return nil, err
		}
		if len(stage) != 0 {
			if stage == instance.Stage {
				instances = append(instances, instance)
			}
		} else {
			instances = append(instances, instance)
		}
	}
	return instances, nil
}

func (s *InstanceController) addDependenceForService(ctx context.Context, tenant string, consumerServiceId string, providerServiceId string) error {
	exist, err := s.existDependence(ctx, tenant, consumerServiceId, providerServiceId)
	if err != nil {
		return err
	}
	if exist {
		util.LOGGER.Warnf(nil, "consumerServiceId:%s , providerServiceId:%s dependency more exists", consumerServiceId, providerServiceId)
		return nil
	}
	dependenceConKey := apt.GenerateConsumerDependencyKey(tenant, consumerServiceId, providerServiceId)
	dependenceProKey := apt.GenerateProviderDependencyKey(tenant, providerServiceId, consumerServiceId)
	timeStamp := strconv.FormatInt(time.Now().Unix(), 10)
	util.LOGGER.Debugf("add service dependenceConKey, %s", dependenceConKey)
	util.LOGGER.Debugf("add service dependenceProKey, %s", dependenceProKey)
	optCon := &registry.PluginOp{
		Action: registry.PUT,
		Key:    []byte(dependenceConKey),
		Value:  []byte(timeStamp),
	}
	optPro := &registry.PluginOp{
		Action: registry.PUT,
		Key:    []byte(dependenceProKey),
		Value:  []byte(timeStamp),
	}
	opts := []*registry.PluginOp{}
	opts = append(opts, optCon)
	opts = append(opts, optPro)
	_, err = registry.GetRegisterCenter().Txn(ctx, opts)
	if err != nil {
		return err
	}
	util.LOGGER.Warnf(nil, "consumerServiceId:%s , providerServiceId:%s dependency add successful", consumerServiceId, providerServiceId)
	return nil
}

func (s *InstanceController) Find(ctx context.Context, in *pb.FindInstancesRequest) (*pb.FindInstancesResponse, error) {
	err := apt.Validate(in)
	if err != nil {
		util.LOGGER.Errorf(err, "Find instance failed for validating parameters failed.")
		return &pb.FindInstancesResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, err.Error()),
		}, nil
	}

	tenant := util.ParaseTenant(ctx)

	service, err := getServiceByServiceId(ctx, tenant, in.ConsumerServiceId)
	if err != nil {
		util.LOGGER.Errorf(err, "Find instance fialed,consumer %s not exist.", in.ConsumerServiceId)
		return &pb.FindInstancesResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, err.Error()),
		}, nil
	}
	if service == nil {
		util.LOGGER.Errorf(err, "Find instance fialed,consumer %s not exist.", in.ConsumerServiceId)
		return &pb.FindInstancesResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Consumer not exist."),
		}, nil
	}

	// 版本规则
	ids, err := FindServiceIds(ctx, in.VersionRule, &pb.MicroServiceKey{
		Tenant:      tenant,
		AppId:       in.AppId,
		ServiceName: in.ServiceName,
		Alias:       in.ServiceName,
	})
	if err != nil {
		message := fmt.Sprintf("Get service %s/%s Id version: %s failed.", in.AppId, in.ServiceName, in.VersionRule)
		util.LOGGER.Errorf(err, message)
		return &pb.FindInstancesResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Get serviceId failed"),
		}, err
	}
	if len(ids) == 0 {
		return &pb.FindInstancesResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Service does not exist"),
		}, nil
	}

	instances := []*pb.MicroServiceInstance{}
	for _, serviceId := range ids {
		resp, err := s.GetInstances(ctx, &pb.GetInstancesRequest{
			ConsumerServiceId: in.ConsumerServiceId,
			ProviderServiceId: serviceId,
			Tags:              in.Tags,
			Stage:             in.Stage,
		})
		if err != nil || resp.GetResponse().Code != pb.Response_SUCCESS {
			return &pb.FindInstancesResponse{
				Response: resp.GetResponse(),
			}, err
		}
		if len(resp.GetInstances()) > 0 {
			instances = append(instances, resp.GetInstances()...)
		}
	}
	consumer := toMicroServiceKey(tenant, service)
	//维护version的规则
	provider := &pb.MicroServiceKey{
		Tenant:      tenant,
		AppId:       in.AppId,
		ServiceName: in.ServiceName,
		Version:     in.VersionRule,
	}
	lock, err := mux.Lock(mux.SERVICE_LOCK)
	err, _ = AddServiceVersionRule(ctx, provider, tenant, consumer)
	if err != nil {
		lock.Unlock()
		util.LOGGER.Errorf(err, "Find instance failed for check version rule existence.")
		return &pb.FindInstancesResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, err.Error()),
		}, nil
	}
	lock.Unlock()

	return &pb.FindInstancesResponse{
		Response:  pb.CreateResponse(pb.Response_SUCCESS, "query service instances successfully"),
		Instances: instances,
	}, nil
}

func (s *InstanceController) existDependence(ctx context.Context, tenant string, consumerServiceId string, providerServiceId string) (bool, error) {
	dependenceKey := apt.GenerateConsumerDependencyKey(tenant, consumerServiceId, providerServiceId)
	util.LOGGER.Debugf("add service dependence, %s", dependenceKey)
	rsp, err := registry.GetRegisterCenter().Do(ctx, &registry.PluginOp{
		Action:    registry.GET,
		Key:       []byte(dependenceKey),
		CountOnly: true,
	})
	if err != nil {
		util.LOGGER.Errorf(nil, "Get %s and %s dependency info failed.", consumerServiceId, providerServiceId)
		return false, err
	}
	if rsp.Count > 0 {
		util.LOGGER.Infof("%s and %s dependency more existed.", consumerServiceId, providerServiceId)
		return true, nil
	}
	return false, nil
}

func (s *InstanceController) UpdateStatus(ctx context.Context, in *pb.UpdateInstanceStatusRequest) (*pb.UpdateInstanceStatusResponse, error) {
	if in == nil || len(in.ServiceId) == 0 || len(in.InstanceId) == 0 {
		util.LOGGER.Errorf(nil, "update instance status failed for instance empty.")
		return &pb.UpdateInstanceStatusResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "request format invalid"),
		}, nil
	}
	//project := in.Project
	tenant := util.ParaseTenant(ctx)
	if !apt.InstanseStatusRule.Match(in.Status) {
		util.LOGGER.Errorf(nil, "update status of instance %s.%s failed for status must be UP|DOWN|STARTING|OUTOFSERVICE.", in.ServiceId, in.InstanceId)
		return &pb.UpdateInstanceStatusResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "status must be UP|DOWN|STARTING|OUTOFSERVICE"),
		}, nil
	}

	// query lease
	var err error
	var leaseID int64
	//project := in.Project
	serviceId := in.ServiceId
	instanceId := in.InstanceId

	var instance *pb.MicroServiceInstance
	instance, err = serviceUtil.GetInstance(ctx, tenant, in.ServiceId, in.InstanceId)
	if instance == nil && err == nil {
		util.LOGGER.Errorf(nil, fmt.Sprintf("%s service 's %s instance not exist.", serviceId, instanceId))
		return &pb.UpdateInstanceStatusResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "No this instance"),
		}, nil
	}
	if err != nil {
		util.LOGGER.Errorf(err, "update properties of instance %s.%s failed for get instance failed.", in.ServiceId, in.InstanceId)
		return &pb.UpdateInstanceStatusResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Update service instance properties failed"),
		}, err
	}

	// query lease
	leaseID, err = serviceUtil.GetLeaseId(ctx, tenant, serviceId, instanceId)
	if err == nil && leaseID == -1 {
		util.LOGGER.Errorf(err, "update properties of instance %s.%s failed for instance not existing.", in.ServiceId, in.InstanceId)
		return &pb.UpdateInstanceStatusResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Update service instance properties failed"),
		}, nil
	}
	if err != nil {
		util.LOGGER.Errorf(err, "update properties of instance %s.%s failed for get instance leaseId failed.", in.ServiceId, in.InstanceId)
		return &pb.UpdateInstanceStatusResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Update service instance properties failed"),
		}, err
	}

	key := apt.GenerateInstanceKey(tenant, in.ServiceId, in.InstanceId)
	instance.Status = in.Status

	data, err := json.Marshal(instance)
	if err != nil {
		util.LOGGER.Errorf(err, "update status of instance %s.%s failed for marshalling instance failed.", in.ServiceId, in.InstanceId)
		return &pb.UpdateInstanceStatusResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "service instance file marshal failed"),
		}, err
	}
	_, err = registry.GetRegisterCenter().Do(ctx, &registry.PluginOp{
		Action: registry.PUT,
		Key:    []byte(key),
		Value:  data,
		Lease:  leaseID,
	})
	if err != nil {
		util.LOGGER.Errorf(err, "update status of instance %s.%s failed for update instance information failed.", in.ServiceId, in.InstanceId)
		return &pb.UpdateInstanceStatusResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "update service instance information failed"),
		}, err
	}

	_, err = registry.GetRegisterCenter().LeaseRenew(ctx, leaseID)
	if err != nil {
		return &pb.UpdateInstanceStatusResponse{
			Response: pb.CreateResponse(pb.Response_FAIL,
				fmt.Sprintf("service instance does not exist(%s)", err.Error())),
		}, nil
	}

	util.LOGGER.Warnf(nil, "update status of instance %s.%s to %s successfully.", in.ServiceId, in.InstanceId, in.Status)
	return &pb.UpdateInstanceStatusResponse{
		Response: pb.CreateResponse(pb.Response_SUCCESS, "update service instance information successfully"),
	}, nil
}

func (s *InstanceController) UpdateInstanceProperties(ctx context.Context, in *pb.UpdateInstancePropsRequest) (*pb.UpdateInstancePropsResponse, error) {
	if in == nil || len(in.ServiceId) == 0 || len(in.InstanceId) == 0 || in.Properties == nil {
		util.LOGGER.Errorf(nil, "update instance properties failed for instance empty.")
		return &pb.UpdateInstancePropsResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "request format invalid"),
		}, nil
	}

	var err error
	var leaseID int64
	//project := in.Project
	tenant := util.ParaseTenant(ctx)
	serviceId := in.ServiceId
	instanceId := in.InstanceId

	var instance *pb.MicroServiceInstance

	instance, err = serviceUtil.GetInstance(ctx, tenant, in.ServiceId, in.InstanceId)
	if instance == nil && err == nil {
		util.LOGGER.Errorf(nil, fmt.Sprintf("%s service 's %s instance not exist.Update properties failed.", serviceId, instanceId))
		return &pb.UpdateInstancePropsResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "No this instance"),
		}, nil
	}
	if err != nil {
		util.LOGGER.Errorf(err, "update properties of instance %s.%s failed for get instance failed.", in.ServiceId, in.InstanceId)
		return &pb.UpdateInstancePropsResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Update service instance properties failed"),
		}, err
	}

	// query lease
	leaseID, err = serviceUtil.GetLeaseId(ctx, tenant, serviceId, instanceId)
	if err != nil {
		util.LOGGER.Errorf(err, "update properties of instance %s.%s failed for get instance leaseId failed.", in.ServiceId, in.InstanceId)
		return &pb.UpdateInstancePropsResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Update service instance properties failed"),
		}, err
	}

	key := apt.GenerateInstanceKey(tenant, in.ServiceId, in.InstanceId)

	instance.Properties = map[string]string{}
	for property := range in.Properties {
		instance.Properties[property] = in.Properties[property]
	}

	data, err := json.Marshal(instance)
	if err != nil {
		util.LOGGER.Errorf(err, "update properties of instance %s.%s failed for marshalling instance failed.", in.ServiceId, in.InstanceId)
		return &pb.UpdateInstancePropsResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "service instance file marshal failed"),
		}, err
	}
	_, err = registry.GetRegisterCenter().Do(ctx, &registry.PluginOp{
		Action: registry.PUT,
		Key:    []byte(key),
		Value:  data,
		Lease:  leaseID,
	})
	if err != nil {
		util.LOGGER.Errorf(err, "update properties of instance %s.%s failed for updating instance information failed.", in.ServiceId, in.InstanceId)
		return &pb.UpdateInstancePropsResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "update service instance information failed"),
		}, err
	}

	_, err = registry.GetRegisterCenter().LeaseRenew(ctx, leaseID)
	if err != nil {
		return &pb.UpdateInstancePropsResponse{
			Response: pb.CreateResponse(pb.Response_FAIL,
				fmt.Sprintf("service instance does not exist(%s)", err.Error())),
		}, nil
	}

	util.LOGGER.Warnf(nil, "update properties of instance %s.%s successfully.", in.ServiceId, in.InstanceId)
	return &pb.UpdateInstancePropsResponse{
		Response: pb.CreateResponse(pb.Response_SUCCESS, "update service instance information successfully"),
	}, nil
}

func (s *InstanceController) Watch(in *pb.WatchInstanceRequest, stream pb.ServiceInstanceCtrl_WatchServer) error {
	var watcher *nf.Watcher
	var err error
	if err, watcher = nf.WatchPreOpera(in, stream.Context(), s.NotifyService); err != nil {
		util.LOGGER.Errorf(err, "Establish watch pre operate failed.")
		return err
	}
	err = s.NotifyService.AddNotifier(watcher)
	util.LOGGER.Infof("start watch instance status, watcher %s %s", watcher.GetSubject(), watcher.GetId())
	return nf.WatchJobHandler(watcher, stream, s.NotifyService.Config.NotifyTimeout)
}

func (s *InstanceController) WebSocketWatch(ctx context.Context, in *pb.WatchInstanceRequest, conn *websocket.Conn) {
	var watcher *nf.Watcher
	var err error
	if err, watcher = nf.WatchPreOpera(in, ctx, s.NotifyService); err != nil {
		util.LOGGER.Errorf(err, "Establish web socket watch pre operate failed.")
		err := conn.WriteMessage(websocket.TextMessage, []byte(err.Error()))
		if err != nil {
			util.LOGGER.Error("write message error", err)
		}
		return
	}
	err = s.NotifyService.AddNotifier(watcher)
	util.LOGGER.Warnf(nil, "start watching instance status, watcher %s %s", watcher.GetSubject(), watcher.GetId())
	nf.WatchWebSocketJobHandler(conn, watcher, s.NotifyService.Config.NotifyTimeout)
}
