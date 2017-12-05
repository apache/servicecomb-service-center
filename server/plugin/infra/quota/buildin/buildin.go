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
package buildin

import (
	"fmt"
	"github.com/ServiceComb/service-center/pkg/util"
	"github.com/ServiceComb/service-center/server/core"
	"github.com/ServiceComb/service-center/server/core/backend/store"
	"github.com/ServiceComb/service-center/server/infra/quota"
	"github.com/ServiceComb/service-center/server/infra/registry"
	mgr "github.com/ServiceComb/service-center/server/plugin"
	serviceUtil "github.com/ServiceComb/service-center/server/service/util"
	"golang.org/x/net/context"
	"strings"
)

const (
	SERVICE_NUM_MAX_LIMIT             = 12000
	SERVICE_NUM_MAX_LIMIT_PER_TENANT  = 100
	INSTANCE_NUM_MAX_LIMIT            = 150000
	INSTANCE_NUM_MAX_LIMIT_PER_TENANT = 100
	RULE_NUM_MAX_LIMIT_PER_SERVICE    = 100
	SCHEMA_NUM_MAX_LIMIT_PER_SERVICE  = 1000
	TAG_NUM_MAX_LIMIT_PER_SERVICE     = 100
)

func init() {
	core.SchemaIdRule.Length = SCHEMA_NUM_MAX_LIMIT_PER_SERVICE
	core.TagRule.Length = TAG_NUM_MAX_LIMIT_PER_SERVICE

	mgr.RegisterPlugin(mgr.Plugin{mgr.STATIC, mgr.QUOTA, "buildin", New})
}

func New() mgr.PluginInstance {
	return &BuildInQuota{}
}

type BuildInQuota struct {
}

//申请配额sourceType serviceinstance servicetype
func (q *BuildInQuota) Apply4Quotas(ctx context.Context, quotaType quota.ResourceType, domainProject string, serviceId string, quotaSize int16) (quota.QuotaReporter, bool, error) {
	data := &QuotaApplyData {
		domain: strings.Split(domainProject, "/")[0],
		quotaSize: int64(quotaSize),
	}
	switch quotaType {
	case quota.MicroServiceInstanceQuotaType:
		isOk, err := instanceQuotaCheck(ctx, data)
		return nil, isOk, err
	case quota.MicroServiceQuotaType:
		isOk, err := serviceQuotaCheck(ctx, data)
		return nil, isOk, err
	default:
		return ResourceLimitHandler(ctx, quotaType, domainProject, serviceId, quotaSize)
	}
}

//向配额中心上报配额使用量
func (q *BuildInQuota) RemandQuotas(ctx context.Context, quotaType quota.ResourceType) {
}

func ResourceLimitHandler(ctx context.Context, quotaType quota.ResourceType, domainProject string, serviceId string, quotaSize int16) (quota.QuotaReporter, bool, error) {
	var key string
	var max int64 = 0
	var indexer *store.Indexer

	switch quotaType {
	case quota.RULEQuotaType:
		key = core.GenerateServiceRuleKey(domainProject, serviceId, "")
		max = RULE_NUM_MAX_LIMIT_PER_SERVICE
		indexer = store.Store().Rule()
	case quota.SCHEMAQuotaType:
		key = core.GenerateServiceSchemaKey(domainProject, serviceId, "")
		max = SCHEMA_NUM_MAX_LIMIT_PER_SERVICE
		indexer = store.Store().Schema()
	case quota.TAGQuotaType:
		num := quotaSize
		max = TAG_NUM_MAX_LIMIT_PER_SERVICE
		tags, err := serviceUtil.GetTagsUtils(ctx, domainProject, serviceId)
		if err != nil {
			return nil, false, err
		}
		if int64(len(tags))+int64(num) > max {
			util.Logger().Errorf(nil, "no quota(%d) to apply resource '%s', %s", max, quotaType, serviceId)
			return nil, false, nil
		}
		return nil, true, nil
	default:
		return nil, false, fmt.Errorf("Unsurported resource '%s'", quotaType)
	}

	resp, err := indexer.Search(ctx,
		registry.WithStrKey(key),
		registry.WithPrefix(),
		registry.WithCountOnly())
	if err != nil {
		return nil, false, err
	}
	num := resp.Count + int64(quotaSize)
	util.Logger().Debugf("resource num is %d", num)
	if num > max {
		util.Logger().Errorf(nil, "no quota(%d) to apply resource '%s', %s", max, quotaType, serviceId)
		return nil, false, nil
	}
	return nil, true, nil
}

type QuotaApplyData struct {
	domain        string
	project       string
	domainProject string
	quotaSize     int64
}

type GetCurUsedNum func(context.Context, *QuotaApplyData) (int64, error)
type GetLimitQuota func() int64

func quotaCheck(ctx context.Context, data *QuotaApplyData, getLimitQuota GetLimitQuota, getCurUsedNum GetCurUsedNum) (bool, error) {
	limitQuota := getLimitQuota()
	curNum, err := getCurUsedNum(ctx, data)
	if err != nil {
		return false, err
	}
	if curNum+data.quotaSize > limitQuota {
		return false, nil
	}
	return true, nil
}

func instanceQuotaCheck(ctx context.Context, data *QuotaApplyData) (isOk bool, err error) {
	isOk, err = quotaCheck(ctx, data, getInstanceMaxLimitUnderOneTenant, getAllInstancesNumUnderOneTenant)
	if err != nil {
		util.Logger().Errorf(err, "instance quota check failed under one tenant")
		return
	}
	if !isOk {
		util.Logger().Errorf(err, "no quota to create instance under one tenant")
		return
	}

	isOk, err = quotaCheck(ctx, data, getInstanceMaxLimit, getAllInstancesNum)
	if err != nil {
		util.Logger().Errorf(err, "instance quota check failed")
		return
	}
	if !isOk {
		util.Logger().Errorf(err, "no quota to create instance")
		return
	}
	return
}

func getInstanceMaxLimitUnderOneTenant() int64 {
	return INSTANCE_NUM_MAX_LIMIT_PER_TENANT
}

func getInstanceMaxLimit() int64 {
	return INSTANCE_NUM_MAX_LIMIT
}

func getInstancesNum(ctx context.Context, key string) (int64, error) {
	resp, err := store.Store().Instance().Search(ctx,
		registry.WithStrKey(key),
		registry.WithPrefix(),
		registry.WithCountOnly())
	if err != nil {
		return 0, err
	}
	return resp.Count, nil
}

func getAllInstancesNum(ctx context.Context, data *QuotaApplyData) (int64, error) {
	key := core.GetInstanceRootKey("")
	return getInstancesNum(ctx, key)
}

func getAllInstancesNumUnderOneTenant(ctx context.Context, data *QuotaApplyData) (int64, error) {
	key := core.GetInstanceRootKey(data.domain)
	return getInstancesNum(ctx, key)
}

func serviceQuotaCheck(ctx context.Context, data *QuotaApplyData) (isOk bool, err error) {
	isOk, err = quotaCheck(ctx, data, getServiceMaxLimitUnderOneTenant, getAllServicesNumUnderOneTenant)
	if err != nil {
		util.Logger().Errorf(err, "service quota check failed under one tenant ")
		return
	}
	if !isOk {
		util.Logger().Errorf(err, "no quota to create service under one tenant")
		return
	}

	isOk, err = quotaCheck(ctx, data, getServiceMaxLimit, getAllServicesNum)
	if err != nil {
		util.Logger().Errorf(err, "service quota check failed")
		return
	}
	if !isOk {
		util.Logger().Errorf(err, "no quota to create service")
		return
	}
	return
}

func getServiceMaxLimitUnderOneTenant() int64 {
	return SERVICE_NUM_MAX_LIMIT_PER_TENANT
}

func getServiceMaxLimit() int64 {
	return SERVICE_NUM_MAX_LIMIT
}

func getServicesNum(ctx context.Context, key string) (int64, error) {
	resp, err := store.Store().Service().Search(ctx,
		registry.WithStrKey(key),
		registry.WithPrefix(),
		registry.WithCountOnly())
	if err != nil {
		return 0, err
	}
	return resp.Count, nil
}

func getAllServicesNum(ctx context.Context, data *QuotaApplyData) (int64, error) {
	key := core.GetServiceRootKey("")
	return getServicesNum(ctx, key)
}

func getAllServicesNumUnderOneTenant(ctx context.Context, data *QuotaApplyData) (int64, error) {
	key := core.GetServiceRootKey(data.domain)
	return getServicesNum(ctx, key)
}