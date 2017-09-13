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
package util

import (
	"encoding/json"
	apt "github.com/ServiceComb/service-center/server/core"
	pb "github.com/ServiceComb/service-center/server/core/proto"
	"github.com/ServiceComb/service-center/server/core/registry"
	"github.com/ServiceComb/service-center/server/core/registry/store"
	ms "github.com/ServiceComb/service-center/server/service/microservice"
	"github.com/ServiceComb/service-center/util"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"golang.org/x/net/context"
	"strconv"
	"strings"
)

func GetLeaseId(ctx context.Context, tenant string, serviceId string, instanceId string) (int64, error) {
	resp, err := store.Store().Lease().Search(ctx,
		registry.WithStrKey(apt.GenerateInstanceLeaseKey(tenant, serviceId, instanceId)))
	if err != nil {
		return -1, err
	}
	if len(resp.Kvs) <= 0 {
		return -1, nil
	}
	leaseID, _ := strconv.ParseInt(util.BytesToStringWithNoCopy(resp.Kvs[0].Value), 10, 64)
	return leaseID, nil
}

func GetInstance(ctx context.Context, tenant string, serviceId string, instanceId string) (*pb.MicroServiceInstance, error) {
	key := apt.GenerateInstanceKey(tenant, serviceId, instanceId)
	resp, err := store.Store().Instance().Search(ctx,
		registry.WithStrKey(key))
	if err != nil {
		return nil, err
	}
	if len(resp.Kvs) == 0 {
		return nil, nil
	}

	var instance *pb.MicroServiceInstance
	err = json.Unmarshal(resp.Kvs[0].Value, &instance)
	if err != nil {
		return nil, err
	}
	return instance, nil
}

func GetAllInstancesOfOneService(ctx context.Context, tenant string, serviceId string, env string) ([]*pb.MicroServiceInstance, error) {
	key := apt.GenerateInstanceKey(tenant, serviceId, "")
	resp, err := store.Store().Instance().Search(ctx,
		registry.WithStrKey(key),
		registry.WithPrefix())
	if err != nil {
		util.Logger().Errorf(err, "Get instance of service %s from etcd failed.", serviceId)
		return nil, err
	}

	instances := []*pb.MicroServiceInstance{}
	for _, kvs := range resp.Kvs {
		util.Logger().Debugf("start unmarshal service instance file: %s", util.BytesToStringWithNoCopy(kvs.Key))
		instance := &pb.MicroServiceInstance{}
		err := json.Unmarshal(kvs.Value, instance)
		if err != nil {
			util.Logger().Errorf(err, "Unmarshal instance of service %s failed.", serviceId)
			return nil, err
		}
		if len(env) != 0 {
			if env == instance.Environment {
				instances = append(instances, instance)
			}
		} else {
			instances = append(instances, instance)
		}
	}
	return instances, nil
}

func InstanceExist(ctx context.Context, tenant string, serviceId string, instanceId string) (bool, error) {
	resp, err := store.Store().Instance().Search(ctx,
		registry.WithStrKey(apt.GenerateInstanceKey(tenant, serviceId, instanceId)),
		registry.WithCountOnly())
	if err != nil {
		return false, err
	}
	if resp.Count <= 0 {
		return false, nil
	}
	return true, nil
}

func CheckEndPoints(ctx context.Context, in *pb.RegisterInstanceRequest) (string, error) {
	tenant := util.ParseTenantProject(ctx)
	allInstancesKey := apt.GenerateInstanceKey(tenant, in.Instance.ServiceId, "")
	rsp, err := store.Store().Instance().Search(ctx,
		registry.WithStrKey(allInstancesKey),
		registry.WithPrefix())
	if err != nil {
		util.Logger().Errorf(nil, "Get all instance info failed.", err.Error())
		return "", err
	}
	if len(rsp.Kvs) == 0 {
		util.Logger().Debugf("There is no instance before this instance regists.")
		return "", nil
	}
	registerInstanceEndpoints := in.Instance.Endpoints
	nodeIpOfIn := ""
	if value, ok := in.GetInstance().Properties["nodeIP"]; ok {
		nodeIpOfIn = value
	}
	instance := &pb.MicroServiceInstance{}
	for _, kv := range rsp.Kvs {
		err = json.Unmarshal(kv.Value, instance)
		if err != nil {
			util.Logger().Errorf(nil, "Unmarshal instance info failed.", err.Error())
			return "", err
		}
		nodeIdFromETCD := ""
		if value, ok := instance.Properties["nodeIP"]; ok {
			nodeIdFromETCD = value
		}
		if nodeIdFromETCD != nodeIpOfIn {
			continue
		}
		tmpInstanceEndpoints := instance.Endpoints
		isEqual := true
		for _, endpoint := range registerInstanceEndpoints {
			if !isContain(tmpInstanceEndpoints, endpoint) {
				isEqual = false
				break
			}
		}
		if isEqual {
			arr := strings.Split(util.BytesToStringWithNoCopy(kv.Key), "/")
			return arr[len(arr)-1], nil
		}
	}
	return "", nil
}

func isContain(endpoints []string, endpoint string) bool {
	for _, tmpEndpoint := range endpoints {
		if tmpEndpoint == endpoint {
			return true
		}
	}
	return false
}

func DeleteServiceAllInstances(ctx context.Context, ServiceId string) error {
	tenant := util.ParseTenantProject(ctx)

	instanceLeaseKey := apt.GenerateInstanceLeaseKey(tenant, ServiceId, "")
	resp, err := store.Store().Lease().Search(ctx,
		registry.WithStrKey(instanceLeaseKey),
		registry.WithPrefix(),
		registry.WithMode(registry.MODE_NO_CACHE))
	if err != nil {
		util.Logger().Errorf(err, "delete service all instance failed: get instance lease failed.")
		return err
	}
	if resp.Count <= 0 {
		util.Logger().Warnf(nil, "No instances to revoke.")
		return nil
	}
	for _, v := range resp.Kvs {
		leaseID, _ := strconv.ParseInt(util.BytesToStringWithNoCopy(v.Value), 10, 64)
		registry.GetRegisterCenter().LeaseRevoke(ctx, leaseID)
	}
	return nil
}

func QueryAllProvidersIntances(ctx context.Context, selfServiceId string) (results []*pb.WatchInstanceResponse, rev int64) {
	results = []*pb.WatchInstanceResponse{}

	tenant := util.ParseTenantProject(ctx)

	service, err := ms.GetService(ctx, tenant, selfServiceId)
	if err != nil {
		util.Logger().Errorf(err, "get service %s failed", selfServiceId)
		return
	}
	if service == nil {
		util.Logger().Errorf(nil, "service not exist, %s", selfServiceId)
		return
	}
	providerIds, _, err := GetProviderIdsByConsumerId(ctx, tenant, selfServiceId, service)
	if err != nil {
		util.Logger().Errorf(err, "get service %s providers id set failed.", selfServiceId)
		return
	}

	rev = store.Revision()

	for _, providerId := range providerIds {
		service, err := ms.GetServiceWithRev(ctx, tenant, providerId, rev)
		if err != nil {
			util.Logger().Errorf(err, "get service %s provider service %s file with revision %d failed.",
				selfServiceId, providerId, rev)
			return
		}
		if service == nil {
			continue
		}
		util.Logger().Debugf("query provider service %v with revision %d.", service, rev)

		kvs, err := queryServiceInstancesKvs(ctx, providerId, rev)
		if err != nil {
			util.Logger().Errorf(err, "get service %s provider %s instances with revision %d failed.",
				selfServiceId, providerId, rev)
			return
		}

		util.Logger().Debugf("query provider service %s instances[%d] with revision %d.", providerId, len(kvs), rev)
		for _, kv := range kvs {
			util.Logger().Debugf("start unmarshal service instance file with revision %d: %s",
				rev, util.BytesToStringWithNoCopy(kv.Key))
			instance := &pb.MicroServiceInstance{}
			err := json.Unmarshal(kv.Value, instance)
			if err != nil {
				util.Logger().Errorf(err, "unmarshal instance of service %s with revision %d failed.",
					providerId, rev)
				return
			}
			results = append(results, &pb.WatchInstanceResponse{
				Response: pb.CreateResponse(pb.Response_SUCCESS, "List instance successfully."),
				Action:   string(pb.EVT_CREATE),
				Key: &pb.MicroServiceKey{
					AppId:       service.AppId,
					ServiceName: service.ServiceName,
					Version:     service.Version,
				},
				Instance: instance,
			})
		}
	}
	return
}

func queryServiceInstancesKvs(ctx context.Context, serviceId string, rev int64) ([]*mvccpb.KeyValue, error) {
	tenant := util.ParseTenantProject(ctx)
	key := apt.GenerateInstanceKey(tenant, serviceId, "")
	resp, err := store.Store().Instance().Search(ctx,
		registry.WithStrKey(key),
		registry.WithPrefix(),
		registry.WithRev(rev))
	if err != nil {
		util.Logger().Errorf(err, "query instance of service %s with revision %d from etcd failed.",
			serviceId, rev)
		return nil, err
	}
	return resp.Kvs, nil
}
