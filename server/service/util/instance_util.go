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
	"github.com/ServiceComb/service-center/util"
	"golang.org/x/net/context"
	"strconv"
	"strings"
)

func GetLeaseId(ctx context.Context, tenant string, serviceId string, instanceId string) (int64, error) {
	resp, err := store.Store().Lease().Search(ctx, &registry.PluginOp{
		Action: registry.GET,
		Key:    util.StringToBytesWithNoCopy(apt.GenerateInstanceLeaseKey(tenant, serviceId, instanceId)),
	})
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
	resp, err := store.Store().Instance().Search(ctx, &registry.PluginOp{
		Action: registry.GET,
		Key:    util.StringToBytesWithNoCopy(key),
	})
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

func InstanceExist(ctx context.Context, tenant string, serviceId string, instanceId string) (bool, error) {
	resp, err := store.Store().Instance().Search(ctx, &registry.PluginOp{
		Action:    registry.GET,
		Key:       util.StringToBytesWithNoCopy(apt.GenerateInstanceKey(tenant, serviceId, instanceId)),
		CountOnly: true,
	})
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
	rsp, err := store.Store().Instance().Search(ctx, &registry.PluginOp{
		Action:     registry.GET,
		Key:        util.StringToBytesWithNoCopy(allInstancesKey),
		WithPrefix: true,
	})
	if err != nil {
		util.LOGGER.Errorf(nil, "Get all instance info failed.", err.Error())
		return "", err
	}
	if len(rsp.Kvs) == 0 {
		util.LOGGER.Debugf("There is no instance before this instance regists.")
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
			util.LOGGER.Errorf(nil, "Unmarshal instance info failed.", err.Error())
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
	resp, err := store.Store().Lease().Search(ctx, &registry.PluginOp{
		Action:      registry.GET,
		Key:         util.StringToBytesWithNoCopy(instanceLeaseKey),
		WithPrefix:  true,
		WithNoCache: true,
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
		leaseID, _ := strconv.ParseInt(util.BytesToStringWithNoCopy(v.Value), 10, 64)
		registry.GetRegisterCenter().LeaseRevoke(ctx, leaseID)
	}
	return nil
}
