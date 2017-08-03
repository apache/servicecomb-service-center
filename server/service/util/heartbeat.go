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
	"errors"
	apt "github.com/ServiceComb/service-center/server/core"
	"github.com/ServiceComb/service-center/server/core/registry"
	"github.com/ServiceComb/service-center/server/core/registry/store"
	"golang.org/x/net/context"
)

func HeartbeatUtil(ctx context.Context, tenant string, serviceId string, instanceId string) (leaseID int64, ttl int64, err error, isInnerErr bool) {
	leaseID, err = GetLeaseId(ctx, tenant, serviceId, instanceId)
	if err != nil {
		return leaseID, ttl, err, true
	}
	if leaseID == -1 {
		return leaseID, ttl, errors.New("leaseId not exist, instance not exist."), false
	}
	ttl, err = store.Store().Lease().KeepAlive(ctx, &registry.PluginOp{
		Action: registry.PUT,
		Key:    []byte(apt.GenerateInstanceLeaseKey(tenant, serviceId, instanceId)),
		Lease:  leaseID,
	})
	if err != nil {
		return leaseID, ttl, err, true
	}
	return leaseID, ttl, nil, false
}
