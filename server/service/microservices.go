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
	"fmt"
	"github.com/astaxie/beego"
	apt "github.com/ServiceComb/service-center/server/core"
	"github.com/ServiceComb/service-center/server/core/mux"
	pb "github.com/ServiceComb/service-center/server/core/proto"
	"github.com/ServiceComb/service-center/server/core/registry"
	"github.com/ServiceComb/service-center/server/infra/quota"
	"github.com/ServiceComb/service-center/server/plugins/dynamic"
	"github.com/ServiceComb/service-center/server/service/dependency"
	ms "github.com/ServiceComb/service-center/server/service/microservice"
	"github.com/ServiceComb/service-center/util"
	"golang.org/x/net/context"
	"strconv"
	"strings"
	"time"
)

type ServiceController struct {
}

// TODO 支持扩展
func GetServiceId(ctx context.Context, key *pb.MicroServiceKey) (string, error) {

	resp, err := registry.GetRegisterCenter().Do(ctx, &registry.PluginOp{
		Action: registry.GET,
		Key:    []byte(apt.GenerateServiceIndexKey(key)),
	})
	if err != nil {
		return "", err
	}
	if len(resp.Kvs) == 0 {
		if len(key.Alias) == 0 {
			return "", nil
		}
		// 别名查询
		util.LOGGER.Debugf("could not search microservice %s/%s/%s id by field 'serviceName', now try field 'alias'.",
			key.AppId, key.ServiceName, key.Version)
		resp, err = registry.GetRegisterCenter().Do(ctx, &registry.PluginOp{
			Action: registry.GET,
			Key:    []byte(apt.GenerateServiceAliasKey(key)),
		})
		if err != nil {
			return "", err
		}
		if len(resp.Kvs) == 0 {
			return "", nil
		}
	}
	return string(resp.Kvs[0].Value), nil
}

func FindServiceIds(ctx context.Context, versionRule string, key *pb.MicroServiceKey) ([]string, error) {
	// 版本规则
	ids := []string{}
	rangeIdx := strings.Index(versionRule, "-")

	alsoFindAlias := len(key.Alias) > 0
	keyGenerator := func(key *pb.MicroServiceKey) string { return apt.GenerateServiceIndexKey(key) }
	versionsFunc := func(key *pb.MicroServiceKey) (*registry.PluginResponse, error) {
		key.Version = ""
		prefix := keyGenerator(key)
		resp, err := registry.GetRegisterCenter().Do(context.TODO(), &registry.PluginOp{
			Action:     registry.GET,
			Key:        []byte(prefix),
			WithPrefix: true,
			SortOrder:  registry.SORT_DESCEND,
		})
		return resp, err
	}

FIND_RULE:
	switch {
	case versionRule == "latest":
		resp, err := versionsFunc(key)
		if err != nil {
			return nil, err
		}
		if len(resp.Kvs) == 0 {
			break
		}
		ids = ms.VersionRule(ms.Latest).GetServicesIds(resp.Kvs)
	case versionRule[len(versionRule)-1:] == "+":
		// 取最低版本及高版本集合
		start := versionRule[:len(versionRule)-1]
		resp, err := versionsFunc(key)
		if err != nil {
			return nil, err
		}
		if len(resp.Kvs) == 0 {
			break
		}
		ids = ms.VersionRule(ms.AtLess).GetServicesIds(resp.Kvs, start)
	case rangeIdx > 0:
		// 取版本范围集合
		start := versionRule[:rangeIdx]
		end := versionRule[rangeIdx+1:]
		resp, err := versionsFunc(key)
		if err != nil {
			return nil, err
		}
		if len(resp.Kvs) == 0 {
			break
		}
		ids = ms.VersionRule(ms.Range).GetServicesIds(resp.Kvs, start, end)
	default:
		// 精确匹配
		key.Version = versionRule
		serviceId, err := GetServiceId(ctx, key)
		if err != nil {
			return nil, err
		}
		if len(serviceId) <= 0 {
			break
		}
		ids = append(ids, serviceId)
	}
	if len(ids) == 0 && alsoFindAlias {
		alsoFindAlias = false
		keyGenerator = func(key *pb.MicroServiceKey) string { return apt.GenerateServiceAliasKey(key) }
		goto FIND_RULE
	}
	return ids, nil
}

func (s *ServiceController) ServiceExist(ctx context.Context, tenant string, serviceId string) bool {
	resp, err := registry.GetRegisterCenter().Do(ctx, &registry.PluginOp{
		Action:    registry.GET,
		Key:       []byte(apt.GenerateServiceKey(tenant, serviceId)),
		CountOnly: true,
	})
	if err != nil || resp.Count == 0 {
		return false
	}
	return true
}

func (s *ServiceController) Create(ctx context.Context, in *pb.CreateServiceRequest) (*pb.CreateServiceResponse, error) {
	remoteIP := util.GetIPFromContext(ctx)
	if in == nil || in.Service == nil {
		util.LOGGER.Errorf(nil, "create microservice failed: param empty.operator:%s", remoteIP)
		return &pb.CreateServiceResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "request format invalid"),
		}, nil
	}
	service := in.Service
	err := apt.Validate(service)

	serviceFlag := strings.Join([]string{service.AppId, service.ServiceName, service.Version}, "--")
	if err != nil {
		util.LOGGER.Errorf(err, "create microservice failed, %s: invalid parameters.operator:%s",
			serviceFlag, remoteIP)
		return &pb.CreateServiceResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, err.Error()),
		}, nil
	}

	tenant := util.ParaseTenantProject(ctx)

	consumer := &pb.MicroServiceKey{
		AppId:       service.AppId,
		ServiceName: service.ServiceName,
		Alias:       service.Alias,
		Version:     service.Version,
		Tenant:      tenant,
	}

	lockMutex := mux.MuxType(apt.GenerateServiceIndexKey(consumer))
	lock, err := mux.Lock(lockMutex)
	if err != nil {
		util.LOGGER.Errorf(err, "create microservice failed, %s:internel err,create lock failed.operator:%s",
			serviceFlag, remoteIP)
		return &pb.CreateServiceResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "begining transaction failed"),
		}, err
	}

	serviceId := in.Service.ServiceId
	serviceIdInner, err := GetServiceId(ctx, consumer)
	if err != nil {
		util.LOGGER.Errorf(err, "create microservice failed, %s:internel err,query service failed.operator:%s",
			serviceFlag, remoteIP)
		lock.Unlock()
		return &pb.CreateServiceResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "query service key failed"),
		}, err
	}
	if len(serviceIdInner) > 0 {
		util.LOGGER.Errorf(nil, "create microservice failed, %s:service already exist.operator:%s",
			serviceFlag, remoteIP)
		lock.Unlock()
		return &pb.CreateServiceResponse{
			Response:  pb.CreateResponse(pb.Response_FAIL, "register service already exists"),
			ServiceId: serviceId,
		}, nil
	}

	ok, err := quota.QuotaPlugins[beego.AppConfig.DefaultString("quota_plugin", "buildin")]().Apply4Quotas(ctx, quota.MicroServiceQuotaType, 0)
	if err != nil {
		util.LOGGER.Errorf(err, "create microservice failed, %s: check apply quota.operator:%s", serviceFlag, remoteIP)
		lock.Unlock()
		return &pb.CreateServiceResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, err.Error()),
		}, err
	}
	if !ok {
		util.LOGGER.Errorf(err, "create microservice failed, %s: no quota to apply.operator:%s", serviceFlag, remoteIP)
		lock.Unlock()
		return &pb.CreateServiceResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, fmt.Sprintf("No quota to create service,service name is %s", in.Service.ServiceName)),
		}, nil
	}

	// 产生全局service id
	if len(serviceId) == 0 {
		serviceId = dynamic.GetServiceId()
	}
	service.ServiceId = serviceId
	service.Timestamp = strconv.FormatInt(time.Now().Unix(), 10)

	data, err := json.Marshal(service)
	if err != nil {
		util.LOGGER.Errorf(err, "create microservice failed, %s: json marshal service failed.operator:%s",
			serviceFlag, remoteIP)
		lock.Unlock()
		return &pb.CreateServiceResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "body error "+err.Error()),
		}, nil
	}
	key := apt.GenerateServiceKey(tenant, serviceId)
	index := apt.GenerateServiceIndexKey(consumer)
	util.LOGGER.Debugf("start register service: %s %v", key, service)
	util.LOGGER.Debugf("start register service index: %s %v", index, serviceId)
	opts := []*registry.PluginOp{
		// Set key file
		{
			Action: registry.PUT,
			Key:    []byte(key),
			Value:  data,
		},
		{
			Action: registry.PUT,
			Key:    []byte(index),
			Value:  []byte(serviceId),
		},
	}

	if len(consumer.Alias) > 0 {
		opts = append(opts, &registry.PluginOp{
			Action: registry.PUT,
			Key:    []byte(apt.GenerateServiceAliasKey(consumer)),
			Value:  []byte(serviceId),
		})
	}

	_, err = registry.GetRegisterCenter().Txn(ctx, opts)
	if err != nil {
		util.LOGGER.Errorf(err, "create microservice failed, %s: commit data into etcd failed.operator:%s",
			serviceFlag, remoteIP)
		lock.Unlock()
		return &pb.CreateServiceResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "commit operations failed"),
		}, err
	}
	lock.Unlock()
	//创建服务间的依赖
	util.LOGGER.Infof("create microservice: add dependency for %s.operator:%s", serviceFlag, remoteIP)
	err = s.updateAsProviderDependency(ctx, serviceId, consumer)
	if err != nil {
		util.LOGGER.Errorf(err, "create microservice failed: dependency update,as provider failed.%s.operator:%s", serviceFlag, remoteIP)
		return &pb.CreateServiceResponse{
			Response:  pb.CreateResponse(pb.Response_FAIL, err.Error()),
			ServiceId: "",
		}, err
	}

	util.LOGGER.Infof("create microservice successful, %s, serviceId: %s. from remote %s",
		serviceFlag, service.ServiceId, util.GetIPFromContext(ctx))
	return &pb.CreateServiceResponse{
		Response:  pb.CreateResponse(pb.Response_SUCCESS, "register service successfully"),
		ServiceId: serviceId,
	}, nil
}

func (s *ServiceController) Delete(ctx context.Context, in *pb.DeleteServiceRequest) (*pb.DeleteServiceResponse, error) {
	if in == nil || len(in.ServiceId) == 0 || in.ServiceId == apt.Service.ServiceId {
		util.LOGGER.Errorf(nil, "delete microservice failed: service empty.")
		return &pb.DeleteServiceResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "request format invalid"),
		}, nil
	}
	err := apt.Validate(in)
	if err != nil {
		util.LOGGER.Errorf(err, "delete microservice failed, serviceId is %s: invalid parameters.", in.ServiceId)
		return &pb.DeleteServiceResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, err.Error()),
		}, nil
	}

	force := in.Force
	tenant := util.ParaseTenantProject(ctx)

	service, err := getServiceByServiceId(ctx, tenant, in.ServiceId)
	if err != nil {
		util.LOGGER.Errorf(err, "delete microservice failed, serviceId is %s: get service failed.", in.ServiceId)
		return &pb.DeleteServiceResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, err.Error()),
		}, err
	}

	if service == nil {
		util.LOGGER.Errorf(err, "delete microservice failed, serviceId is %s: service not exist.", in.ServiceId)
		return &pb.DeleteServiceResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "service not exist"),
		}, nil
	}

	util.LOGGER.Infof("start delete service %s", in.ServiceId)

	// 强制删除，则与该服务相关的信息删除，非强制删除： 如果作为该被依赖（作为provider，提供服务,且不是只存在自依赖）或者存在实例，则不能删除
	if !force {
		keyConDependency := apt.GenerateProviderDependencyKey(tenant, in.ServiceId, "")
		services, err := GetDependencies(ctx, keyConDependency, tenant)
		if err != nil {
			util.LOGGER.Errorf(err, "delete microservice failed, serviceId is %s:(unforce) inner err, get service dependency failed.", in.ServiceId)
			return &pb.DeleteServiceResponse{
				Response: pb.CreateResponse(pb.Response_FAIL, "Get dependency info failed."),
			}, err
		}
		if len(services) > 1 || (len(services) == 1 && services[0].ServiceId != in.ServiceId) {
			util.LOGGER.Errorf(nil, "delete microservice failed, serviceId is %s:(unforce) can't delete, other services rely it.", in.ServiceId)
			return &pb.DeleteServiceResponse{
				Response: pb.CreateResponse(pb.Response_FAIL, "Can not delete this service, other service rely it."),
			}, err
		}

		instancesKey := apt.GenerateInstanceKey(tenant, in.ServiceId, "")
		rsp, err := registry.GetRegisterCenter().Do(ctx, &registry.PluginOp{
			Action:     registry.GET,
			Key:        []byte(instancesKey),
			WithPrefix: true,
			CountOnly:  true,
		})
		if err != nil {
			util.LOGGER.Errorf(err, "delete microservice failed, serviceId is %s:(unforce) inner err,get instances failed.", in.ServiceId)
			return &pb.DeleteServiceResponse{
				Response: pb.CreateResponse(pb.Response_FAIL, "Get instance failed."),
			}, err
		}

		if rsp.Count > 0 {
			util.LOGGER.Errorf(nil, "delete microservice failed, serviceId is %s:(unforce) can't delete, exist instance.", in.ServiceId)
			return &pb.DeleteServiceResponse{
				Response: pb.CreateResponse(pb.Response_FAIL, "Can not delete this service, exist instance."),
			}, err
		}
	}

	consumer := &pb.MicroServiceKey{
		AppId:       service.AppId,
		ServiceName: service.ServiceName,
		Version:     service.Version,
		Alias:       service.Alias,
		Tenant:      tenant,
	}

	//refresh msCache consumerCache, ensure that watch can notify consumers when no cache.
	err = refreshDependencyCache(tenant, in.ServiceId, service)
	if err != nil {
		util.LOGGER.Errorf(err, "delete microservice failed, serviceId is %s: inner err, refresh service dependency cache failed.", in.ServiceId)
		return &pb.DeleteServiceResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Refresh dependency cache failed."),
		}, err
	}

	opts := []*registry.PluginOp{
		{
			Action: registry.DELETE,
			Key:    []byte(apt.GenerateServiceIndexKey(consumer)),
		},
		{
			Action: registry.DELETE,
			Key:    []byte(apt.GenerateServiceAliasKey(consumer)),
		},
		{
			Action: registry.DELETE,
			Key:    []byte(apt.GenerateServiceKey(tenant, in.ServiceId)),
		},
		{
			Action: registry.DELETE,
			Key: []byte(strings.Join([]string{
				apt.GetServiceRuleRootKey(tenant),
				in.ServiceId,
				"",
			}, "/")),
		},
	}

	//删除依赖规则
	optsTmp, err := s.deteleDependencyForService(ctx, consumer, in.ServiceId)
	if err != nil {
		util.LOGGER.Errorf(err, "delete microservice failed, serviceId is %s: inner err, delete dependency failed.", in.ServiceId)
		return &pb.DeleteServiceResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, err.Error()),
		}, err
	}
	opts = append(opts, optsTmp...)

	//删除黑白名单
	rulekey := apt.GenerateServiceRuleKey(tenant, in.ServiceId, "")
	opt := &registry.PluginOp{
		Action:     registry.DELETE,
		Key:        []byte(rulekey),
		WithPrefix: true,
	}
	opts = append(opts, opt)
	indexKey := apt.GenerateRuleIndexKey(tenant, in.ServiceId, "", "")
	opts = append(opts, &registry.PluginOp{
		Action: registry.DELETE,
		Key:    []byte(indexKey),
	})
	opts = append(opts, opt)

	//删除shemas
	schemaKey := apt.GenerateServiceSchemaKey(tenant, in.ServiceId, "")
	opt = &registry.PluginOp{
		Action:     registry.DELETE,
		Key:        []byte(schemaKey),
		WithPrefix: true,
	}
	opts = append(opts, opt)

	//删除tags
	tagsKey := apt.GenerateServiceTagKey(tenant, in.ServiceId)
	opt = &registry.PluginOp{
		Action: registry.DELETE,
		Key:    []byte(tagsKey),
	}
	opts = append(opts, opt)

	//删除实例
	err = deleteServiceAllInstances(ctx, in)
	if err != nil {
		util.LOGGER.Errorf(err, "delete microservice failed, serviceId is %s: delete all instances failed.", in.ServiceId)
		return &pb.DeleteServiceResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Delete all instances failed for service."),
		}, err
	}

	_, err = registry.GetRegisterCenter().Txn(ctx, opts)
	if err != nil {
		util.LOGGER.Errorf(err, "delete microservice failed, serviceId is %s: commit data into etcd failed.", in.ServiceId)
		return &pb.DeleteServiceResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "commit operations failed"),
		}, nil
	}

	util.LOGGER.Infof("delete microservice successful: serviceid is %s,operator is %s.", in.ServiceId, util.GetIPFromContext(ctx))
	return &pb.DeleteServiceResponse{
		Response: pb.CreateResponse(pb.Response_SUCCESS, "unregister service successfully"),
	}, nil
}

func refreshDependencyCache(tenant string, providerId string, provider *pb.MicroService) error {
	key := apt.GenerateProviderDependencyKey(tenant, providerId, "")
	resp, err := registry.GetRegisterCenter().Do(context.TODO(), &registry.PluginOp{
		Action:     registry.GET,
		Key:        []byte(key),
		WithPrefix: true,
		KeyOnly:    true,
	})
	if err != nil {
		util.LOGGER.Errorf(err, "refresh dependency cache failed: query service consumers failed, provider id %s", providerId)
		return err
	}
	if resp == nil || len(resp.Kvs) == 0 {
		util.LOGGER.Infof("refresh dependency cache: this services %s has no consumer dependency.", providerId)
	} else {
		ms.MsCache().Set(providerId, provider, 5*time.Minute)
		dependency.ConsumerCache().Set(providerId, resp.Kvs, 5*time.Minute)
	}
	return nil
}

func deleteServiceAllInstances(ctx context.Context, in *pb.DeleteServiceRequest) error {
	tenant := util.ParaseTenantProject(ctx)

	instanceLeaseKey := apt.GenerateInstanceLeaseKey(tenant, in.ServiceId, "")

	resp, err := registry.GetRegisterCenter().Do(ctx, &registry.PluginOp{
		Action:     registry.GET,
		Key:        []byte(instanceLeaseKey),
		WithPrefix: true,
	})
	if err != nil {
		util.LOGGER.Errorf(err, "delete service all instance failed: get instance lease failed.")
		return err
	}
	if resp.Count <= 0 {
		util.LOGGER.Warnf(nil, "No instances to revoke.")
		return nil
	}
	for _, v := range resp.Kvs {
		leaseID, _ := strconv.ParseInt(string(v.Value), 10, 64)
		registry.GetRegisterCenter().LeaseRevoke(ctx, leaseID)
	}
	return nil
}

func (s *ServiceController) deteleDependencyForService(ctx context.Context, consumer *pb.MicroServiceKey, serviceId string) ([]*registry.PluginOp, error) {
	opts := []*registry.PluginOp{}
	optsTmps := []*registry.PluginOp{}
	tenant := consumer.Tenant
	flag := map[string]bool{}
	//删除依赖规则
	conKey := apt.GenerateConsumerDependencyRuleKey(tenant, consumer)
	err, providerValue := transferToMircroServiceDependency(ctx, conKey)
	if err != nil {
		return nil, err
	}
	if providerValue != nil && len(providerValue.Dependency) != 0 {
		proProkey := ""
		for _, providerRule := range providerValue.Dependency {
			proProkey = apt.GenerateProviderDependencyRuleKey(tenant, providerRule)
			err, consumers := transferToMircroServiceDependency(ctx, proProkey)
			if err != nil {
				return nil, err
			}
			err = deleteDependencyRuleUtil(ctx, consumers, consumer, proProkey)
			if err != nil {
				return nil, err
			}
		}

		opt := &registry.PluginOp{
			Action: registry.DELETE,
			Key:    []byte(conKey),
		}
		util.LOGGER.Debugf("conKey is %s.", conKey)
		opts = append(opts, opt)
	}
	//作为provider的依赖规则
	providerKey := apt.GenerateProviderDependencyRuleKey(tenant, consumer)
	opt := &registry.PluginOp{
		Action: registry.DELETE,
		Key:    []byte(providerKey),
	}
	util.LOGGER.Debugf("providerKey is %s", providerKey)
	opts = append(opts, opt)

	//删除依赖关系
	optsTmps, err = s.deleteDependencyUtil(ctx, "c", tenant, serviceId, flag)
	if err != nil {
		return nil, err
	}
	opts = append(opts, optsTmps...)
	util.LOGGER.Debugf("flag is %s", flag)
	optsTmps, err = s.deleteDependencyUtil(ctx, "p", tenant, serviceId, flag)
	if err != nil {
		return nil, err
	}
	util.LOGGER.Debugf("flag is %s", flag)
	opts = append(opts, optsTmps...)
	return opts, nil
}

func deleteDependencyRuleUtil(ctx context.Context, mircroServiceDependency *MircroServiceDependency, service *pb.MicroServiceKey, serviceKey string) error {
	for key, serviceTmp := range mircroServiceDependency.Dependency {
		if ok := equalServiceDependency(serviceTmp, service); ok {
			mircroServiceDependency.Dependency = append(mircroServiceDependency.Dependency[:key], mircroServiceDependency.Dependency[key+1:]...)
			util.LOGGER.Debugf("delete versionRule from %s", serviceTmp.ServiceName)
			break
		}
	}
	opt := &registry.PluginOp{}
	if len(mircroServiceDependency.Dependency) == 0 {
		opt = &registry.PluginOp{
			Action: registry.DELETE,
			Key:    []byte(serviceKey),
		}
		util.LOGGER.Debugf("serviceKey is .", serviceKey)
		util.LOGGER.Debugf("After deleting versionRule from %s,provider's consumer is empty.", serviceKey)

	} else {
		data, err := json.Marshal(mircroServiceDependency)
		if err != nil {
			util.LOGGER.Errorf(nil, "Marshal tmpValue failed.")
			return err
		}
		opt = &registry.PluginOp{
			Action: registry.PUT,
			Key:    []byte(serviceKey),
			Value:  []byte(data),
		}
		util.LOGGER.Debugf("serviceKey is %s.", serviceKey)
	}
	_, err := registry.GetRegisterCenter().Do(ctx, opt)
	if err != nil {
		util.LOGGER.Errorf(err, "Submit update dependency failed.")
		return err
	}
	return nil
}

func (s *ServiceController) deleteDependencyUtil(ctx context.Context, serviceType string, tenant string, serviceId string, flag map[string]bool) ([]*registry.PluginOp, error) {
	serviceKey := apt.GenerateServiceDependencyKey(serviceType, tenant, serviceId, "")
	opt := &registry.PluginOp{
		Action:     registry.GET,
		Key:        []byte(serviceKey),
		WithPrefix: true,
	}
	rsp, err := registry.GetRegisterCenter().Do(ctx, opt)
	if err != nil {
		return nil, err
	}
	opts := []*registry.PluginOp{}
	if rsp != nil {
		serviceTmpId := ""
		serviceTmpKey := ""
		deleteKey := ""
		for _, kv := range rsp.Kvs {
			tmpKeyArr := strings.Split(string(kv.Key), "/")
			serviceTmpId = tmpKeyArr[len(tmpKeyArr)-1]
			if serviceType == "p" {
				serviceTmpKey = apt.GenerateConsumerDependencyKey(tenant, serviceTmpId, serviceId)
				deleteKey = strings.Join([]string{"c", serviceTmpId, serviceId}, "/")
			} else {
				serviceTmpKey = apt.GenerateProviderDependencyKey(tenant, serviceTmpId, serviceId)
				deleteKey = strings.Join([]string{"p", serviceTmpId, serviceId}, "/")
			}
			if _, ok := flag[serviceTmpKey]; ok {
				util.LOGGER.Debugf("serviceTmpKey is more exist.%s", serviceTmpKey)
				continue
			}
			flag[serviceTmpKey] = true
			util.LOGGER.Infof("delete dependency %s", deleteKey)
			opt = &registry.PluginOp{
				Action: registry.DELETE,
				Key:    []byte(serviceTmpKey),
			}
			opts = append(opts, opt)
		}
		opt = &registry.PluginOp{
			Action:     registry.DELETE,
			Key:        []byte(serviceKey),
			WithPrefix: true,
		}
		util.LOGGER.Infof("delete dependency serviceKey is %s", serviceType+"/"+serviceId)
		opts = append(opts, opt)
	}
	return opts, nil
}

func (s *ServiceController) GetOne(ctx context.Context, in *pb.GetServiceRequest) (*pb.GetServiceResponse, error) {
	if in == nil || len(in.ServiceId) == 0 {
		return &pb.GetServiceResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "request format invalid"),
		}, nil
	}
	err := apt.Validate(in)
	if err != nil {
		util.LOGGER.Errorf(err, "get microservice failed, serviceId is %s: invalid parameters.",
			in.ServiceId)
		return &pb.GetServiceResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, err.Error()),
		}, nil
	}
	tenant := util.ParaseTenantProject(ctx)
	service, err := getServiceByServiceId(ctx, tenant, in.ServiceId)

	if err != nil {
		util.LOGGER.Errorf(err, "get microservice failed, serviceId is %s: inner err,get service failed.", in.ServiceId)
		return &pb.GetServiceResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "get service file failed"),
		}, err
	}
	if service == nil {
		util.LOGGER.Errorf(nil, "get microservice failed, serviceId is %s: service not exist.", in.ServiceId)
		return &pb.GetServiceResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "There is no this service."),
		}, nil
	}
	return &pb.GetServiceResponse{
		Response: pb.CreateResponse(pb.Response_SUCCESS, "get service successfully"),
		Service:  service,
	}, nil
}

func (s *ServiceController) GetServices(ctx context.Context, in *pb.GetServicesRequest) (*pb.GetServicesResponse, error) {
	if in == nil {
		util.LOGGER.Errorf(nil, "get services failed: invalid params.")
		return &pb.GetServicesResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Request format invalid"),
		}, nil
	}
	services, err := GetAllServiceUtil(ctx)
	if err != nil {
		util.LOGGER.Errorf(err, "get services failed: inner err.")
		return &pb.GetServicesResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Get all service failed."),
		}, err
	}

	return &pb.GetServicesResponse{
		Response: pb.CreateResponse(pb.Response_SUCCESS, "Get services successfully"),
		Services: services,
	}, nil
}

func (s *ServiceController) UpdateProperties(ctx context.Context, in *pb.UpdateServicePropsRequest) (*pb.UpdateServicePropsResponse, error) {
	if in == nil || len(in.ServiceId) == 0 || in.Properties == nil {
		util.LOGGER.Errorf(nil, "update service properties failed: invalid params.")
		return &pb.UpdateServicePropsResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "request format invalid"),
		}, nil
	}
	err := apt.Validate(in)
	if err != nil {
		util.LOGGER.Errorf(err, "update service properties failed, serviceId is %s: invalid parameters.", in.ServiceId)
		return &pb.UpdateServicePropsResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, err.Error()),
		}, nil
	}

	tenant := util.ParaseTenantProject(ctx)

	key := apt.GenerateServiceKey(tenant, in.ServiceId)
	resp, err := registry.GetRegisterCenter().Do(ctx, &registry.PluginOp{
		Action: registry.GET,
		Key:    []byte(key),
	})
	if err != nil {
		util.LOGGER.Errorf(err, "update service properties failed, serviceId is %s: query service failed.", in.ServiceId)
		return &pb.UpdateServicePropsResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "query service file failed"),
		}, err
	}
	if len(resp.Kvs) == 0 {
		util.LOGGER.Errorf(nil, "update service properties failed, serviceId is %s: service not exist.", in.ServiceId)
		return &pb.UpdateServicePropsResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "service does not exist"),
		}, nil
	}

	service := pb.MicroService{}

	err = json.Unmarshal(resp.Kvs[0].Value, &service)
	if err != nil {
		util.LOGGER.Errorf(err, "update service properties failed, serviceId is %s: json unmarshal service failed.", in.ServiceId)
		return &pb.UpdateServicePropsResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "service file unmarshal error"),
		}, err
	}
	service.Properties = make(map[string]string)
	for propertyKey := range in.Properties {
		service.Properties[propertyKey] = in.Properties[propertyKey]
	}

	data, err := json.Marshal(service)
	if err != nil {
		util.LOGGER.Errorf(err, "update service properties failed, serviceId is %s: json marshal service failed.", in.ServiceId)
		return &pb.UpdateServicePropsResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "service file marshal error"),
		}, err
	}

	// Set key file
	_, err = registry.GetRegisterCenter().Do(ctx, &registry.PluginOp{
		Action: registry.PUT,
		Key:    []byte(key),
		Value:  data,
	})
	if err != nil {
		util.LOGGER.Errorf(err, "update service properties failed, serviceId is %s: commit data into etcd failed.", in.ServiceId)
		return &pb.UpdateServicePropsResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "commit operations failed"),
		}, err
	}

	util.LOGGER.Infof("update service properties successful: serviceId is %s.", in.ServiceId)
	return &pb.UpdateServicePropsResponse{
		Response: pb.CreateResponse(pb.Response_SUCCESS, "update service successfully"),
	}, nil
}

func (s *ServiceController) RuleExist(ctx context.Context, tenant string, serviceId string, attr string, pattern string) bool {
	resp, err := registry.GetRegisterCenter().Do(ctx, &registry.PluginOp{
		Action:    registry.GET,
		Key:       []byte(apt.GenerateRuleIndexKey(tenant, serviceId, attr, pattern)),
		CountOnly: true,
	})
	if err != nil || resp.Count == 0 {
		return false
	}
	return true
}

func (s *ServiceController) AddRule(ctx context.Context, in *pb.AddServiceRulesRequest) (*pb.AddServiceRulesResponse, error) {
	if in == nil || len(in.ServiceId) == 0 || len(in.GetRules()) == 0 {
		util.LOGGER.Errorf(nil, "add rule failed: invalid parameters.")
		return &pb.AddServiceRulesResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "request format invalid"),
		}, nil
	}

	tenant := util.ParaseTenantProject(ctx)

	// service id存在性校验
	if !s.ServiceExist(ctx, tenant, in.ServiceId) {
		util.LOGGER.Errorf(nil, "add rule failed, serviceId is %s: service not exist.", in.ServiceId)
		return &pb.AddServiceRulesResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "service does not exist"),
		}, nil
	}

	opts := []*registry.PluginOp{}
	ruleType, _, err := getServiceRuleType(ctx, tenant, in.ServiceId)
	util.LOGGER.Debugf("ruleType is %s", ruleType)
	if err != nil {
		return &pb.AddServiceRulesResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, err.Error()),
		}, nil
	}
	ruleIds := []string{}
	for _, rule := range in.Rules {
		err := apt.Validate(rule)
		if err != nil {
			util.LOGGER.Errorf(err, "add rule failed, serviceId is %s: invalid rule.", in.ServiceId)
			return &pb.AddServiceRulesResponse{
				Response: pb.CreateResponse(pb.Response_FAIL, err.Error()),
			}, nil
		}
		//黑白名单只能存在一种，黑名单 or 白名单
		if len(ruleType) == 0 {
			ruleType = rule.RuleType
		} else {
			if ruleType != rule.RuleType {
				util.LOGGER.Errorf(nil, "add rule failed, serviceId is %s:can only exist one type, BLACK or WHITE.", in.ServiceId)
				return &pb.AddServiceRulesResponse{
					Response: pb.CreateResponse(pb.Response_FAIL, "Service can only contain one rule type, BLACK or WHITE."),
				}, nil
			}
		}

		//同一服务，attribute和pattern确定一个rule
		if s.RuleExist(ctx, tenant, in.ServiceId, rule.Attribute, rule.Pattern) {
			util.LOGGER.Infof("This rule more exists, %s ", in.ServiceId)
			continue
		}

		// 产生全局rule id
		resp, err := registry.GetRegisterCenter().Do(ctx, &registry.PluginOp{
			Action:     registry.PUT,
			Key:        []byte(apt.GetRuleSequenceKey()),
			WithPrevKV: true,
		})
		if err != nil {
			util.LOGGER.Errorf(err, "add rule failed, serviceId is %s:generate ruleId failed.", in.ServiceId)
			return &pb.AddServiceRulesResponse{
				Response: pb.CreateResponse(pb.Response_FAIL, "generate service id error"),
			}, err
		}
		timestamp := strconv.FormatInt(time.Now().Unix(), 10)
		ruleAdd := &pb.ServiceRule{
			RuleId:      "1",
			RuleType:    rule.RuleType,
			Attribute:   rule.Attribute,
			Pattern:     rule.Pattern,
			Description: rule.Description,
			Timestamp:   timestamp,
		}
		if resp.PrevKv != nil {
			ruleAdd.RuleId = strconv.FormatInt(resp.PrevKv.Version+1, 10)
		}

		key := apt.GenerateServiceRuleKey(tenant, in.ServiceId, ruleAdd.RuleId)
		indexKey := apt.GenerateRuleIndexKey(tenant, in.ServiceId, ruleAdd.Attribute, ruleAdd.Pattern)
		ruleIds = append(ruleIds, ruleAdd.RuleId)

		util.LOGGER.Debugf("indexKey is : %s", indexKey)
		util.LOGGER.Debugf("start add service rule file: %s", key)
		data, err := json.Marshal(ruleAdd)
		if err != nil {
			util.LOGGER.Errorf(err, "add rule failed, serviceId is %s: marshal rule failed.", in.ServiceId)
			return &pb.AddServiceRulesResponse{
				Response: pb.CreateResponse(pb.Response_FAIL, "service rule file marshal error"),
			}, err
		}

		opts = append(opts, &registry.PluginOp{
			Action: registry.PUT,
			Key:    []byte(key),
			Value:  data,
		})
		opts = append(opts, &registry.PluginOp{
			Action: registry.PUT,
			Key:    []byte(indexKey),
			Value:  []byte(ruleAdd.RuleId),
		})
	}
	if len(opts) <= 0 {
		util.LOGGER.Infof("add rule successful, serviceId is %s: rule more exists,no rules to add.", in.ServiceId)
		return &pb.AddServiceRulesResponse{
			Response: pb.CreateResponse(pb.Response_SUCCESS, "service rules has been added"),
		}, nil
	}
	_, err = registry.GetRegisterCenter().Txn(ctx, opts)
	if err != nil {
		util.LOGGER.Errorf(err, "add rule failed, serviceId is %s:commit date into etcd failed.", in.ServiceId)
		return &pb.AddServiceRulesResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "commit operations failed"),
		}, err
	}

	util.LOGGER.Infof("add rule successful, serviceId  %s.", in.ServiceId)
	return &pb.AddServiceRulesResponse{
		Response: pb.CreateResponse(pb.Response_SUCCESS, "add service rules successfully"),
		RuleIds:  ruleIds,
	}, nil
}

func getServiceRuleType(ctx context.Context, tenant string, serviceId string) (string, int, error) {
	key := apt.GenerateServiceRuleKey(tenant, serviceId, "")
	resp, err := registry.GetRegisterCenter().Do(ctx, &registry.PluginOp{
		Action:     registry.GET,
		Key:        []byte(key),
		WithPrefix: true,
	})
	if err != nil {
		util.LOGGER.Errorf(err, "Get rule failed.%s", err.Error())
		return "", 0, err
	}
	if len(resp.Kvs) == 0 {
		return "", 0, nil
	}
	rule := &pb.ServiceRule{}
	err = json.Unmarshal(resp.Kvs[0].Value, rule)
	if err != nil {
		util.LOGGER.Errorf(err, "Unmarshal rule data failed.%s", err.Error())
	}
	return rule.RuleType, len(resp.Kvs), nil
}

func (s *ServiceController) UpdateRule(ctx context.Context, in *pb.UpdateServiceRuleRequest) (*pb.UpdateServiceRuleResponse, error) {
	if in == nil || in.GetRule() == nil || len(in.ServiceId) == 0 || len(in.RuleId) == 0 {
		util.LOGGER.Errorf(nil, "update rule failed: invalid parameters.")
		return &pb.UpdateServiceRuleResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "request format invalid"),
		}, nil
	}

	tenant := util.ParaseTenantProject(ctx)

	// service id存在性校验
	if !s.ServiceExist(ctx, tenant, in.ServiceId) {
		util.LOGGER.Errorf(nil, "update rule failed, serviceId is %s, ruleId is %s: service not exist.", in.ServiceId, in.RuleId)
		return &pb.UpdateServiceRuleResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "service does not exist"),
		}, nil
	}
	err := apt.Validate(in.Rule)
	if err != nil {
		util.LOGGER.Errorf(err, "update rule failed, serviceId is %s, ruleId is %s: invalid service rule.", in.ServiceId, in.RuleId)
		return &pb.UpdateServiceRuleResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, err.Error()),
		}, nil
	}

	//是否能改变ruleType
	ruleType, ruleNum, err := getServiceRuleType(ctx, tenant, in.ServiceId)
	if err != nil {
		util.LOGGER.Errorf(err, "update rule failed, serviceId is %s, ruleId is %s: get rule type failed.", in.ServiceId, in.RuleId)
		return &pb.UpdateServiceRuleResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, err.Error()),
		}, err
	}
	if ruleNum >= 1 && ruleType != in.Rule.RuleType {
		util.LOGGER.Errorf(err, "update rule failed, serviceId is %s, ruleId is %s: rule type can exist one type, BLACK or WHITE.rule type is %s", in.ServiceId, in.RuleId, in.Rule.RuleType)
		return &pb.UpdateServiceRuleResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Exist multiple rules,can not change rule type.Rule type is "+ruleType),
		}, nil
	}

	key := apt.GenerateServiceRuleKey(tenant, in.ServiceId, in.RuleId)
	resp, err := registry.GetRegisterCenter().Do(ctx, &registry.PluginOp{
		Action: registry.GET,
		Key:    []byte(key),
	})
	if err != nil {
		util.LOGGER.Errorf(err, "update rule failed, serviceId is %s, ruleId is %s: query service rule failed.", in.ServiceId, in.RuleId)
		return &pb.UpdateServiceRuleResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "get service rule file failed"),
		}, err
	}

	if len(resp.Kvs) == 0 {
		util.LOGGER.Errorf(err, "update rule failed, serviceId is %s, ruleId is %s:this rule does not exist,can't update.", in.ServiceId, in.RuleId)
		return &pb.UpdateServiceRuleResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "This rule does not exist."),
		}, err
	}
	rule := &pb.ServiceRule{}
	err = json.Unmarshal(resp.Kvs[0].Value, rule)
	if err != nil {
		util.LOGGER.Errorf(err, "update rule failed, serviceId is %s, ruleId is %s: unmarshal service rule failed.", in.ServiceId, in.RuleId)
		return &pb.UpdateServiceRuleResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "unmarshal service rule file failed"),
		}, err
	}

	oldRulePatten := rule.Pattern
	oldRuleAttr := rule.Attribute
	isChangeIndex := false
	if rule.Attribute != in.GetRule().Attribute {
		isChangeIndex = true
		rule.Attribute = in.GetRule().Attribute
	}
	if rule.Pattern != in.GetRule().Pattern {
		isChangeIndex = true
		rule.Pattern = in.GetRule().Pattern
	}
	rule.RuleType = in.GetRule().RuleType
	rule.Description = in.GetRule().Description
	rule.Timestamp = strconv.FormatInt(time.Now().Unix(), 10)

	util.LOGGER.Debugf("start update service rule file: %s", key)
	data, err := json.Marshal(rule)
	if err != nil {
		util.LOGGER.Errorf(err, "update rule failed, serviceId is %s, ruleId is %s: marshal service rule failed.", in.ServiceId, in.RuleId)
		return &pb.UpdateServiceRuleResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "service rule file marshal error"),
		}, err
	}
	opts := []*registry.PluginOp{}
	if isChangeIndex {
		//加入新的rule index
		indexKey := apt.GenerateRuleIndexKey(tenant, in.ServiceId, rule.Attribute, rule.Pattern)
		opt := &registry.PluginOp{
			Action: registry.PUT,
			Key:    []byte(indexKey),
			Value:  []byte(rule.RuleId),
		}
		opts = append(opts, opt)

		//删除旧的rule index
		oldIndexKey := apt.GenerateRuleIndexKey(tenant, in.ServiceId, oldRuleAttr, oldRulePatten)
		opt = &registry.PluginOp{
			Action: registry.DELETE,
			Key:    []byte(oldIndexKey),
		}

		opts = append(opts, opt)
	}
	opt := &registry.PluginOp{
		Action: registry.PUT,
		Key:    []byte(key),
		Value:  data,
	}
	opts = append(opts, opt)
	_, err = registry.GetRegisterCenter().Txn(ctx, opts)
	if err != nil {
		util.LOGGER.Errorf(err, "update rule failed, serviceId is %s, ruleId is %s: commit date into etcd failed.", in.ServiceId, in.RuleId)
		return &pb.UpdateServiceRuleResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "commit operations failed"),
		}, err
	}

	util.LOGGER.Infof("update rule successful: servieId is , ruleId is %s.", in.ServiceId, in.RuleId)
	return &pb.UpdateServiceRuleResponse{
		Response: pb.CreateResponse(pb.Response_SUCCESS, "get service rules successfully"),
	}, nil
}

func (s *ServiceController) GetRule(ctx context.Context, in *pb.GetServiceRulesRequest) (*pb.GetServiceRulesResponse, error) {
	if in == nil || len(in.ServiceId) == 0 {
		util.LOGGER.Errorf(nil, "get service rule failed, serviceId is %s: invalid params.", in.ServiceId)
		return &pb.GetServiceRulesResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "request format invalid"),
		}, nil
	}

	tenant := util.ParaseTenantProject(ctx)
	// service id存在性校验
	if !s.ServiceExist(ctx, tenant, in.ServiceId) {
		util.LOGGER.Errorf(nil, "get service rule failed, serviceId is %s: service not exist.", in.ServiceId)
		return &pb.GetServiceRulesResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "service does not exist"),
		}, nil
	}

	rules, err := GetRulesUtil(ctx, tenant, in.ServiceId)
	if err != nil {
		util.LOGGER.Errorf(nil, "get service rule failed, serviceId is %s: get rule failed.", in.ServiceId)
		return &pb.GetServiceRulesResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "get service rules faild."),
		}, nil
	}

	return &pb.GetServiceRulesResponse{
		Response: pb.CreateResponse(pb.Response_SUCCESS, "get service rules successfully"),
		Rules:    rules,
	}, nil
}

func (s *ServiceController) DeleteRule(ctx context.Context, in *pb.DeleteServiceRulesRequest) (*pb.DeleteServiceRulesResponse, error) {
	if in == nil || len(in.ServiceId) == 0 {
		util.LOGGER.Errorf(nil, "delete service rule failed: invalid parameters.")
		return &pb.DeleteServiceRulesResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "request format invalid"),
		}, nil
	}

	tenant := util.ParaseTenantProject(ctx)
	// service id存在性校验
	if !s.ServiceExist(ctx, tenant, in.ServiceId) {
		util.LOGGER.Errorf(nil, "delete service rule failed, serviceId is %s, rule is %v: service not exist.", in.ServiceId, in.RuleIds)
		return &pb.DeleteServiceRulesResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "service does not exist"),
		}, nil
	}

	opts := []*registry.PluginOp{}
	key := ""
	indexKey := ""
	for _, ruleId := range in.RuleIds {
		key = apt.GenerateServiceRuleKey(tenant, in.ServiceId, ruleId)
		util.LOGGER.Debugf("start delete service rule file: %s", key)
		data, err := s.getOneRule(ctx, key, ruleId)
		if err != nil {
			util.LOGGER.Errorf(err, "delete service rule failed, serviceId is %s, rule is %v: get rule of ruleId %s failed.", in.ServiceId, in.RuleIds, ruleId)
			return &pb.DeleteServiceRulesResponse{
				Response: pb.CreateResponse(pb.Response_FAIL, err.Error()),
			}, err
		}
		if data == nil {
			util.LOGGER.Errorf(nil, "delete service rule failed, serviceId is %s, rule is %v: ruleId %s not exist.", in.ServiceId, in.RuleIds, ruleId)
			return &pb.DeleteServiceRulesResponse{
				Response: pb.CreateResponse(pb.Response_FAIL, "This rule does not exist"),
			}, nil
		}
		indexKey = apt.GenerateRuleIndexKey(tenant, in.ServiceId, data.Attribute, data.Pattern)
		opts = append(opts, &registry.PluginOp{
			Action: registry.DELETE,
			Key:    []byte(key),
		})
		opts = append(opts, &registry.PluginOp{
			Action: registry.DELETE,
			Key:    []byte(indexKey),
		})
	}
	if len(opts) <= 0 {
		util.LOGGER.Errorf(nil, "delete service rule failed, serviceId is %s, rule is %v: rule has been deleted.", in.ServiceId, in.RuleIds)
		return &pb.DeleteServiceRulesResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "no service rule has been deleted"),
		}, nil
	}
	_, err := registry.GetRegisterCenter().Txn(ctx, opts)
	if err != nil {
		util.LOGGER.Errorf(err, "delete service rule failed, serviceId is %s, rule is %v: commit data into etcd failed.", in.ServiceId, in.RuleIds)
		return &pb.DeleteServiceRulesResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "commit operations failed"),
		}, err
	}

	util.LOGGER.Infof("delete rule successful: serviceId %s, ruleIds %v", in.ServiceId, in.RuleIds)
	return &pb.DeleteServiceRulesResponse{
		Response: pb.CreateResponse(pb.Response_SUCCESS, "delete service rules successfully"),
	}, nil
}

func (s *ServiceController) getOneRule(ctx context.Context, key string, ruleId string) (*pb.ServiceRule, error) {
	opt := &registry.PluginOp{
		Action: registry.GET,
		Key:    []byte(key),
	}
	resp, err := registry.GetRegisterCenter().Do(ctx, opt)
	if err != nil {
		util.LOGGER.Errorf(nil, "Get rule for service failed for %s.", err.Error())
		return nil, err
	}
	rule := &pb.ServiceRule{}
	if len(resp.Kvs) == 0 {
		util.LOGGER.Errorf(nil, "Get rule failed, ruleId is %s.", ruleId)
		return nil, nil
	}
	err = json.Unmarshal(resp.Kvs[0].Value, rule)
	if err != nil {
		util.LOGGER.Errorf(nil, "unmarshal resp failed for %s.", err.Error())
		return nil, err
	}
	return rule, nil
}

func (s *ServiceController) Exist(ctx context.Context, in *pb.GetExistenceRequest) (*pb.GetExistenceResponse, error) {
	if in == nil {
		util.LOGGER.Errorf(nil, "exist failed: invalid params.")
		return &pb.GetExistenceResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "request format invalid"),
		}, nil
	}

	tenant := util.ParaseTenantProject(ctx)
	switch in.Type {
	case "microservice":
		if len(in.AppId) == 0 || len(in.ServiceName) == 0 || len(in.Version) == 0 {
			util.LOGGER.Errorf(nil, "microservice exist failed: invalid params.")
			return &pb.GetExistenceResponse{
				Response: pb.CreateResponse(pb.Response_FAIL, "Invalid request."),
			}, nil
		}
		err := apt.GetMSExistsReqValidator.Validate(in)
		serviceFlag := strings.Join([]string{in.AppId, in.ServiceName, in.Version}, "--")
		if err != nil {
			util.LOGGER.Errorf(err, "microservice exist failed, service %s: invalid params.", serviceFlag)
			return &pb.GetExistenceResponse{
				Response: pb.CreateResponse(pb.Response_FAIL, err.Error()),
			}, nil
		}

		ids, err := FindServiceIds(ctx, in.Version, &pb.MicroServiceKey{
			AppId:       in.AppId,
			ServiceName: in.ServiceName,
			Alias:       in.ServiceName,
			Version:     in.Version,
			Tenant:      tenant,
		})
		if err != nil {
			util.LOGGER.Errorf(err, "microservice exist failed, service %s: find serviceIds failed.", serviceFlag)
			return &pb.GetExistenceResponse{
				Response: pb.CreateResponse(pb.Response_FAIL, "get service file failed"),
			}, err
		}
		if len(ids) <= 0 {
			util.LOGGER.Errorf(nil, "microservice exist failed, service %s: service not exist.", serviceFlag)
			return &pb.GetExistenceResponse{
				Response: pb.CreateResponse(pb.Response_FAIL, "service does not exist"),
			}, nil
		}
		return &pb.GetExistenceResponse{
			Response:  pb.CreateResponse(pb.Response_SUCCESS, "get service id successfully"),
			ServiceId: ids[0], // 约定多个时，取较新版本
		}, nil
	case "schema":
		if len(in.SchemaId) == 0 || len(in.ServiceId) == 0 {
			util.LOGGER.Errorf(nil, "schema exist failed, serviceId %s, schemaId : invalid params.", in.ServiceId, in.SchemaId)
			return &pb.GetExistenceResponse{
				Response: pb.CreateResponse(pb.Response_FAIL, "Invalid request."),
			}, nil
		}

		err := apt.GetSchemaExistsReqValidator.Validate(in)
		if err != nil {
			util.LOGGER.Errorf(err, "schema exist failed, serviceId %s, schemaId : invalid params.", in.ServiceId, in.SchemaId)
			return &pb.GetExistenceResponse{
				Response: pb.CreateResponse(pb.Response_FAIL, err.Error()),
			}, nil
		}
		if !s.ServiceExist(ctx, tenant, in.ServiceId) {
			util.LOGGER.Errorf(nil, "schema exist failed, serviceId %s, schemaId : service not exist.", in.ServiceId, in.SchemaId)
			return &pb.GetExistenceResponse{
				Response: pb.CreateResponse(pb.Response_FAIL, "Service does not exist"),
			}, nil
		}

		key := apt.GenerateServiceSchemaKey(tenant, in.ServiceId, in.SchemaId)
		err, exist := s.checkSchemaInfoExist(ctx, key)
		if err != nil {
			util.LOGGER.Errorf(err, "schema exist failed, serviceId %s, schemaId : get schema failed.", in.ServiceId, in.SchemaId)
			return &pb.GetExistenceResponse{
				Response: pb.CreateResponse(pb.Response_FAIL, err.Error()),
			}, err
		}
		if !exist {
			util.LOGGER.Errorf(nil, "schema exist failed, serviceId %s, schemaId %s: schema not exist.", in.ServiceId, in.SchemaId)
			return &pb.GetExistenceResponse{
				Response: pb.CreateResponse(pb.Response_FAIL, "Schema does not exist"),
			}, nil
		}
		return &pb.GetExistenceResponse{
			Response: pb.CreateResponse(pb.Response_SUCCESS, "Schema exist."),
			SchemaId: in.SchemaId,
		}, nil
	default:
		return &pb.GetExistenceResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Only microservice and schema can be used as type."),
		}, nil
	}
}

func (s *ServiceController) AddTags(ctx context.Context, in *pb.AddServiceTagsRequest) (*pb.AddServiceTagsResponse, error) {
	if in == nil || len(in.ServiceId) == 0 || len(in.GetTags()) == 0 {
		util.LOGGER.Errorf(nil, "add service tags failed: invalid parameters.")
		return &pb.AddServiceTagsResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "request format invalid"),
		}, nil
	}

	err := apt.Validate(in)
	if err != nil {
		util.LOGGER.Errorf(err, "add service tags failed, serviceId %s, tags %v: invalid parameters.", in.ServiceId, in.Tags)
		return &pb.AddServiceTagsResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, err.Error()),
		}, nil
	}

	tenant := util.ParaseTenantProject(ctx)
	// service id存在性校验
	if !s.ServiceExist(ctx, tenant, in.ServiceId) {
		util.LOGGER.Errorf(nil, "add service tags failed, serviceId %s, tags %v: service not exist.", in.ServiceId, in.Tags)
		return &pb.AddServiceTagsResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "service does not exist"),
		}, nil
	}

	addTags := in.GetTags()

	key := apt.GenerateServiceTagKey(tenant, in.ServiceId)

	resp, err := registry.GetRegisterCenter().Do(ctx, &registry.PluginOp{
		Action: registry.GET,
		Key:    []byte(key),
	})
	if err != nil {
		util.LOGGER.Errorf(err, "add service tags failed, serviceId %s, tags %v: get existed tag failed.", in.ServiceId, in.Tags)
		return &pb.AddServiceTagsResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Get tags failed."),
		}, err
	}

	dataTags := map[string]string{}
	if len(resp.Kvs) > 0 {
		err = json.Unmarshal(resp.Kvs[0].Value, &dataTags)
		if err != nil {
			util.LOGGER.Errorf(err, "add service tags failed, serviceId %s, tags %v: json unmarshal tag data failed.", in.ServiceId, in.Tags)
			return &pb.AddServiceTagsResponse{
				Response: pb.CreateResponse(pb.Response_FAIL, "Unmarshal tags failed."),
			}, nil
		}
		if len(dataTags) > 0 {
			for key, value := range addTags {
				dataTags[key] = value
			}
		} else {
			dataTags = addTags
		}
	} else {
		dataTags = addTags
	}

	err = addTagIntoETCD(ctx, tenant, in.ServiceId, dataTags)
	if err != nil {
		util.LOGGER.Errorf(err, "add service tags failed, serviceId %s, tags %v: commit tag data into etcd failed.", in.ServiceId, in.Tags)
		return &pb.AddServiceTagsResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "commit operations failed"),
		}, err
	}

	util.LOGGER.Infof("add service tags successful, serviceId %s, tags %v.", in.ServiceId, in.Tags)
	return &pb.AddServiceTagsResponse{
		Response: pb.CreateResponse(pb.Response_SUCCESS, "add service tags successfully"),
	}, nil
}

func (s *ServiceController) UpdateTag(ctx context.Context, in *pb.UpdateServiceTagRequest) (*pb.UpdateServiceTagResponse, error) {
	if in == nil || len(in.ServiceId) == 0 || len(in.Key) == 0 || len(in.Value) == 0 {
		util.LOGGER.Errorf(nil, "update service tag failed: invalid parameters.")
		return &pb.UpdateServiceTagResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "request format invalid"),
		}, nil
	}
	tagFlag := strings.Join([]string{in.Key, in.Value}, "--")
	err := apt.Validate(in)
	if err != nil {
		util.LOGGER.Errorf(err, "update service tag failed, serviceId %s, tag %s: invalid params.", in.ServiceId, tagFlag)
		return &pb.UpdateServiceTagResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, err.Error()),
		}, nil
	}

	tenant := util.ParaseTenantProject(ctx)

	if !s.ServiceExist(ctx, tenant, in.ServiceId) {
		util.LOGGER.Errorf(err, "update service tag failed, serviceId %s, tag %s: service not exist.", in.ServiceId, tagFlag)
		return &pb.UpdateServiceTagResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Service does not exist."),
		}, nil
	}

	tags, err := GetTagsUtils(ctx, tenant, in.ServiceId)
	if err != nil {
		util.LOGGER.Errorf(err, "update service tag failed, serviceId %s, tag %s: get tag failed.", in.ServiceId, tagFlag)
		return &pb.UpdateServiceTagResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Get tags for serivce failed."),
		}, err
	}
	//check tag 是否存在
	if _, ok := tags[in.Key]; !ok {
		util.LOGGER.Errorf(nil, "update service tag failed, serviceId %s, tag %s: tag not exist,please add first.", in.ServiceId, tagFlag)
		return &pb.UpdateServiceTagResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Update tag for service failed for update tags not exist,please add first."),
		}, err
	}
	tags[in.Key] = in.Value

	err = addTagIntoETCD(ctx, tenant, in.ServiceId, tags)

	if err != nil {
		util.LOGGER.Errorf(err, "update service tag failed, serviceId %s, tag %s: adding service tags failed.", in.ServiceId, tagFlag)
		return &pb.UpdateServiceTagResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Commit into etcd failed."),
		}, err
	}

	util.LOGGER.Infof("update tag successful, serviceId %s, tag %s.", in.ServiceId, tagFlag)
	return &pb.UpdateServiceTagResponse{
		Response: pb.CreateResponse(pb.Response_SUCCESS, "update service tag success."),
	}, nil
}

func addTagIntoETCD(ctx context.Context, tenant string, serviceId string, dataTags map[string]string) error {
	key := apt.GenerateServiceTagKey(tenant, serviceId)
	data, err := json.Marshal(dataTags)
	if err != nil {
		util.LOGGER.Errorf(err, "add tag into etcd,serviceId %s:json marshal tag data failed.", serviceId)
		return err
	}

	_, err = registry.GetRegisterCenter().Do(ctx, &registry.PluginOp{
		Action: registry.PUT,
		Key:    []byte(key),
		Value:  data,
	})
	if err != nil {
		util.LOGGER.Errorf(err, "add tag into etcd,serviceId %s: commit tag data into etcd failed.", serviceId)
		return err
	}
	return nil
}

func (s *ServiceController) DeleteTags(ctx context.Context, in *pb.DeleteServiceTagsRequest) (*pb.DeleteServiceTagsResponse, error) {
	if in == nil || len(in.ServiceId) == 0 || len(in.Keys) == 0 {
		util.LOGGER.Errorf(nil, "delete service tags failed: invalid parameters.")
		return &pb.DeleteServiceTagsResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "request format invalid"),
		}, nil
	}
	err := apt.Validate(in)
	if err != nil {
		util.LOGGER.Errorf(err, "delete service tags failed, serviceId %s, tags %v: invalid params.", in.ServiceId, in.Keys)
		return &pb.DeleteServiceTagsResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, err.Error()),
		}, nil
	}

	tenant := util.ParaseTenantProject(ctx)

	if !s.ServiceExist(ctx, tenant, in.ServiceId) {
		util.LOGGER.Errorf(nil, "delete service tags failed, serviceId %s, tags %v: service not exist.", in.ServiceId, in.Keys)
		return &pb.DeleteServiceTagsResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Service does not exist."),
		}, nil
	}

	resp, err := registry.GetRegisterCenter().Do(ctx, &registry.PluginOp{
		Action: registry.GET,
		Key:    []byte(apt.GenerateServiceTagKey(tenant, in.ServiceId)),
	})
	if err != nil {
		util.LOGGER.Errorf(err, "delete service tags failed, serviceId %s, tags %v: query service failed.", in.ServiceId, in.Keys)
		return &pb.DeleteServiceTagsResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "get service tags file failed"),
		}, err
	}
	if len(resp.Kvs) == 0 {
		util.LOGGER.Errorf(err, "delete service tags failed, serviceId %s, tags %v: tag not exist,can't delete.", in.ServiceId, in.Keys)
		return &pb.DeleteServiceTagsResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "tags do not exist"),
		}, nil
	}

	var tags map[string]string
	err = json.Unmarshal(resp.Kvs[0].Value, &tags)
	if err != nil {
		util.LOGGER.Errorf(err, "delete service tags failed, serviceId %s, tags %v: unmarshal tag data failed.", in.ServiceId, in.Keys)
		return &pb.DeleteServiceTagsResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "unmarshal tags file failed"),
		}, err
	}
	for _, key := range in.Keys {
		if _, ok := tags[key]; !ok {
			util.LOGGER.Errorf(nil, "delete service tags failed, serviceId %s, tags %v: tag %s not exist.", in.ServiceId, in.Keys, key)
			return &pb.DeleteServiceTagsResponse{
				Response: pb.CreateResponse(pb.Response_FAIL, "Delete tags for service failed.for this key "+key+" does not exist"),
			}, nil
		}
		delete(tags, key)
	}

	// tags 可能size == 0
	data, err := json.Marshal(tags)
	if err != nil {
		util.LOGGER.Errorf(err, "delete service tags failed, serviceId %s, tags %v: marshall service tag failed.", in.ServiceId, in.Keys)
		return &pb.DeleteServiceTagsResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "marshal service tags file failed"),
		}, err
	}

	key := apt.GenerateServiceTagKey(tenant, in.ServiceId)

	util.LOGGER.Debugf("start delete service tags file: %s %v", key, in.Keys)
	_, err = registry.GetRegisterCenter().Do(ctx, &registry.PluginOp{
		Action: registry.PUT,
		Key:    []byte(key),
		Value:  data,
	})
	if err != nil {
		util.LOGGER.Errorf(err, "delete service tags failed, serviceId %s, tags %v: commit tag data into etcd failed.", in.ServiceId, in.Keys)
		return &pb.DeleteServiceTagsResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "commit operations failed"),
		}, err
	}

	util.LOGGER.Infof("delete service tags successful: serviceId %s, tag %v.", in.ServiceId, in.Keys)
	return &pb.DeleteServiceTagsResponse{
		Response: pb.CreateResponse(pb.Response_SUCCESS, "delete service tags successfully"),
	}, nil
}

func (s *ServiceController) GetTags(ctx context.Context, in *pb.GetServiceTagsRequest) (*pb.GetServiceTagsResponse, error) {
	if in == nil || len(in.ServiceId) == 0 {
		util.LOGGER.Errorf(nil, "get service tags failed: invalid params.")
		return &pb.GetServiceTagsResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "request format invalid"),
		}, nil
	}
	err := apt.Validate(in)
	if err != nil {
		util.LOGGER.Errorf(err, "get service tags failed, serviceId %s: invalid parameters.", in.ServiceId)
		return &pb.GetServiceTagsResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, err.Error()),
		}, nil
	}

	tenant := util.ParaseTenantProject(ctx)

	if !s.ServiceExist(ctx, tenant, in.ServiceId) {
		util.LOGGER.Errorf(err, "get service tags failed, serviceId %s: service not exist.", in.ServiceId)
		return &pb.GetServiceTagsResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Service does not exist"),
		}, nil
	}

	tags, err := GetTagsUtils(ctx, tenant, in.ServiceId)
	if err != nil {
		util.LOGGER.Errorf(err, "get service tags failed, serviceId %s: get tag failed.", in.ServiceId)
		return &pb.GetServiceTagsResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Get tags for serivce failed."),
		}, err
	}

	return &pb.GetServiceTagsResponse{
		Response: pb.CreateResponse(pb.Response_SUCCESS, "get service tags successfully"),
		Tags:     tags,
	}, nil
}

func (s *ServiceController) GetSchemaInfo(ctx context.Context, request *pb.GetSchemaRequest) (*pb.GetSchemaResponse, error) {
	if request == nil || len(request.ServiceId) == 0 || len(request.SchemaId) == 0 {
		util.LOGGER.Errorf(nil, "get schema failed: invalid params.")
		return &pb.GetSchemaResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Invalid request path."),
		}, nil
	}

	err := apt.Validate(request)
	if err != nil {
		util.LOGGER.Errorf(nil, "get schema failed, serviceId %s, schemaId %s: invalid params.", request.ServiceId, request.SchemaId)
		return &pb.GetSchemaResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, err.Error()),
		}, nil
	}

	tenant := util.ParaseTenantProject(ctx)

	if !s.ServiceExist(ctx, tenant, request.ServiceId) {
		util.LOGGER.Errorf(nil, "get schema failed, serviceId %s, schemaId %s: service not exist.", request.ServiceId, request.SchemaId)
		return &pb.GetSchemaResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Service does not exist"),
		}, nil
	}

	key := apt.GenerateServiceSchemaKey(tenant, request.ServiceId, request.SchemaId)
	resp, errDo := registry.GetRegisterCenter().Do(ctx, &registry.PluginOp{
		Action: registry.GET,
		Key:    []byte(key),
	})
	if errDo != nil {
		util.LOGGER.Errorf(errDo, "get schema failed, serviceId %s, schemaId %s: get schema info failed.", request.ServiceId, request.SchemaId)
		return &pb.GetSchemaResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Get schema info failded."),
		}, errDo
	}
	if resp.Count == 0 {
		util.LOGGER.Errorf(errDo, "get schema failed, serviceId %s, schemaId %s: schema not exists.", request.ServiceId, request.SchemaId)
		return &pb.GetSchemaResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Do not have this schema info."),
		}, nil
	}
	return &pb.GetSchemaResponse{
		Response: pb.CreateResponse(pb.Response_SUCCESS, "Get schema info successfully"),
		Schema:   string(resp.Kvs[0].Value),
	}, nil
}

func (s *ServiceController) DeleteSchema(ctx context.Context, request *pb.DeleteSchemaRequest) (*pb.DeleteSchemaResponse, error) {
	if request == nil || len(request.ServiceId) == 0 || len(request.SchemaId) == 0 {
		util.LOGGER.Errorf(nil, "delete schema failded: invalid params.")
		return &pb.DeleteSchemaResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Invalid request path."),
		}, nil
	}
	err := apt.Validate(request)
	if err != nil {
		util.LOGGER.Errorf(err, "delete schema failded, serviceId %s, schemaId %s: invalid params.", request.ServiceId, request.SchemaId)
		return &pb.DeleteSchemaResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, err.Error()),
		}, nil
	}
	tenant := util.ParaseTenantProject(ctx)

	if !s.ServiceExist(ctx, tenant, request.ServiceId) {
		util.LOGGER.Errorf(nil, "delete schema failded, serviceId %s, schemaId %s: service not exist.", request.ServiceId, request.SchemaId)
		return &pb.DeleteSchemaResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Service does not exist"),
		}, nil
	}

	key := apt.GenerateServiceSchemaKey(tenant, request.ServiceId, request.SchemaId)
	err, exist := s.checkSchemaInfoExist(ctx, key)
	if err != nil {
		util.LOGGER.Errorf(err, "delete schema failded, serviceId %s, schemaId %s: get schema failed.", request.ServiceId, request.SchemaId)
		return &pb.DeleteSchemaResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Schema info does not exist."),
		}, err
	}
	if !exist {
		util.LOGGER.Errorf(nil, "delete schema failded, serviceId %s, schemaId %s: schema not exist.", request.ServiceId, request.SchemaId)
		return &pb.DeleteSchemaResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Schema info does not exist."),
		}, nil
	}
	_, errDo := registry.GetRegisterCenter().Do(ctx, &registry.PluginOp{
		Action: registry.DELETE,
		Key:    []byte(key),
	})
	if errDo != nil {
		util.LOGGER.Errorf(errDo, "delete schema failded, serviceId %s, schemaId %s: delete schema from etcd faild.", request.ServiceId, request.SchemaId)
		return &pb.DeleteSchemaResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Delete schema info failded."),
		}, errDo
	}
	util.LOGGER.Infof("delete schema info successfully.%s", request.SchemaId)
	return &pb.DeleteSchemaResponse{
		Response: pb.CreateResponse(pb.Response_SUCCESS, "Delete schema info successfully."),
	}, nil
}

func (s *ServiceController) ModifySchema(ctx context.Context, request *pb.ModifySchemaRequest) (*pb.ModifySchemaResponse, error) {
	err, rst := s.canModifySchema(ctx, request)
	if err != nil {
		return &pb.ModifySchemaResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Modify schema info failded."),
		}, err
	}
	if !rst {
		return &pb.ModifySchemaResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Modify schema info failded,service or schemaId not exist."),
		}, nil
	}

	tenant := util.ParaseTenantProject(ctx)
	key := apt.GenerateServiceSchemaKey(tenant, request.ServiceId, request.SchemaId)
	_, errDo := registry.GetRegisterCenter().Do(ctx, &registry.PluginOp{
		Action: registry.PUT,
		Key:    []byte(key),
		Value:  []byte(request.Schema),
	})
	if errDo != nil {
		util.LOGGER.Errorf(errDo, "update schema failded, serviceId %s, schemaId %s: commit schema into etcd failed.", request.ServiceId, request.SchemaId)
		return &pb.ModifySchemaResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Modify schema info failded."),
		}, errDo
	}
	util.LOGGER.Infof("update schema success: serviceId %s, schemaId %s.", request.ServiceId, request.SchemaId)
	return &pb.ModifySchemaResponse{
		Response: pb.CreateResponse(pb.Response_SUCCESS, "Modify schema info success."),
	}, nil

}

func (s *ServiceController) canModifySchema(ctx context.Context, request *pb.ModifySchemaRequest) (error, bool) {
	if request == nil || len(request.Schema) == 0 || len(request.SchemaId) == 0 || len(request.ServiceId) == 0 {
		util.LOGGER.Errorf(nil, "update schema failded: invalid params.")
		return nil, false
	}
	err := apt.Validate(request)
	if err != nil {
		util.LOGGER.Errorf(err, "update schema failded, serviceId %s, schemaId %s: invalid params.", request.ServiceId, request.SchemaId)
		return err, false
	}
	tenant := util.ParaseTenantProject(ctx)
	service, err := getServiceByServiceId(ctx, tenant, request.ServiceId)
	if err != nil {
		util.LOGGER.Errorf(err, "update schema failded, serviceId %s, schemaId %s: get service failed.", request.ServiceId, request.SchemaId)
		return err, false
	}
	if service == nil {
		util.LOGGER.Errorf(nil, "update schema failded, serviceId %s, schemaId %s: service not exist,%s", request.ServiceId, request.SchemaId)
		return nil, false
	}
	schemas := service.Schemas
	if !containsValueInSlice(schemas, request.SchemaId) {
		message := "Do not contain " + request.SchemaId + "Schema in service file."
		util.LOGGER.Errorf(nil, "update schema failded, serviceId %s, schemaId %s:%s", request.ServiceId, request.SchemaId, message)
		return nil, false
	}
	return nil, true
}

func containsValueInSlice(in []string, value string) bool {
	if in == nil || len(value) == 0 {
		return false
	}
	for _, i := range in {
		if i == value {
			return true
		}
	}
	return false
}

func (s *ServiceController) checkSchemaInfoExist(ctx context.Context, key string) (error, bool) {
	resp, errDo := registry.GetRegisterCenter().Do(ctx, &registry.PluginOp{
		Action:    registry.GET,
		Key:       []byte(key),
		CountOnly: true,
	})
	if errDo != nil {
		return errDo, false
	}
	if resp.Count == 0 {
		return nil, false
	}
	return nil, true
}

func GetAllServiceUtil(ctx context.Context) ([]*pb.MicroService, error) {
	tenant := util.ParaseTenantProject(ctx)
	services, err := ms.GetServicesByTenant(ctx, tenant)
	if err != nil {
		return nil, err
	}
	return services, nil
}

func GetTagsUtils(ctx context.Context, tenant string, serviceId string) (map[string]string, error) {
	tags := map[string]string{}

	key := apt.GenerateServiceTagKey(tenant, serviceId)
	resp, err := registry.GetRegisterCenter().Do(ctx, &registry.PluginOp{
		Action: registry.GET,
		Key:    []byte(key),
	})
	if err != nil {
		return tags, err
	}
	if len(resp.Kvs) != 0 {
		util.LOGGER.Debugf("start unmarshal service tags file: %s", key)
		err = json.Unmarshal(resp.Kvs[0].Value, &tags)
		if err != nil {
			return tags, err
		}
	}
	return tags, nil
}

func GetRulesUtil(ctx context.Context, tenant string, serviceId string) ([]*pb.ServiceRule, error) {
	key := strings.Join([]string{
		apt.GetServiceRuleRootKey(tenant),
		serviceId,
		"",
	}, "/")

	resp, err := registry.GetRegisterCenter().Do(ctx, &registry.PluginOp{
		Action:     registry.GET,
		Key:        []byte(key),
		WithPrefix: true,
	})
	if err != nil {
		return nil, err
	}

	rules := []*pb.ServiceRule{}
	for _, kvs := range resp.Kvs {
		util.LOGGER.Debugf("start unmarshal service rule file: %s", string(kvs.Key))
		rule := &pb.ServiceRule{}
		err := json.Unmarshal(kvs.Value, rule)
		if err != nil {
			return nil, err
		}
		rules = append(rules, rule)
	}
	return rules, nil
}
