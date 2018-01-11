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
package buildin

import (
	"fmt"
	"github.com/apache/incubator-servicecomb-service-center/pkg/util"
	"github.com/apache/incubator-servicecomb-service-center/server/core"
	"github.com/apache/incubator-servicecomb-service-center/server/core/backend/store"
	"github.com/apache/incubator-servicecomb-service-center/server/infra/quota"
	"github.com/apache/incubator-servicecomb-service-center/server/infra/registry"
	mgr "github.com/apache/incubator-servicecomb-service-center/server/plugin"
	serviceUtil "github.com/apache/incubator-servicecomb-service-center/server/service/util"
	"golang.org/x/net/context"
	"strings"
	scerr "github.com/apache/incubator-servicecomb-service-center/server/error"
)

const (
	SERVICE_NUM_MAX_LIMIT            = 12000
	INSTANCE_NUM_MAX_LIMIT           = 150000
	RULE_NUM_MAX_LIMIT_PER_SERVICE   = 100
	SCHEMA_NUM_MAX_LIMIT_PER_SERVICE = 1000
	TAG_NUM_MAX_LIMIT_PER_SERVICE    = 100
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
func (q *BuildInQuota) Apply4Quotas(ctx context.Context, res *quota.ApplyQuotaRes) *quota.ApplyQuotaResult {
	data := &QuotaApplyData{
		domain:    strings.Split(res.DomainProject, "/")[0],
		quotaSize: res.QuotaSize,
	}
	switch res.QuotaType {
	case quota.MicroServiceInstanceQuotaType:
		return instanceQuotaCheck(ctx, data)
	case quota.MicroServiceQuotaType:
		return serviceQuotaCheck(ctx, data)
	default:
		return ResourceLimitHandler(ctx, res)
	}
}

//向配额中心上报配额使用量
func (q *BuildInQuota) RemandQuotas(ctx context.Context, quotaType quota.ResourceType) {
}

func ResourceLimitHandler(ctx context.Context, res *quota.ApplyQuotaRes) *quota.ApplyQuotaResult {
	var key string
	var max int64 = 0
	var indexer *store.Indexer

	domainProject := res.DomainProject
	serviceId := res.ServiceId
	switch res.QuotaType {
	case quota.RuleQuotaType:
		key = core.GenerateServiceRuleKey(domainProject, serviceId, "")
		max = RULE_NUM_MAX_LIMIT_PER_SERVICE
		indexer = store.Store().Rule()
	case quota.SchemaQuotaType:
		key = core.GenerateServiceSchemaKey(domainProject, serviceId, "")
		max = SCHEMA_NUM_MAX_LIMIT_PER_SERVICE
		indexer = store.Store().Schema()
	case quota.TagQuotaType:
		applyNum := res.QuotaSize
		max = TAG_NUM_MAX_LIMIT_PER_SERVICE
		tags, err := serviceUtil.GetTagsUtils(ctx, domainProject, serviceId)
		if err != nil {
			return quota.NewApplyQuotaResult(nil, scerr.NewError(scerr.ErrInternal, err.Error()))
		}
		curNum := int64(len(tags))
		if curNum+applyNum > max {
			mes := fmt.Sprintf("no quota to apply %s , max quota is %d, current used quota is %d, apply quota num is %d",
				res.QuotaType, max, curNum, applyNum)
			util.Logger().Errorf(nil, "%s, serviceId is %s", mes, serviceId)
			return quota.NewApplyQuotaResult(nil, scerr.NewError(scerr.ErrNotEnoughQuota, mes))
		}
		return quota.NewApplyQuotaResult(nil, nil)
	default:
		return quota.NewApplyQuotaResult(nil, scerr.NewError(scerr.ErrNotDefineQuotaType, ""))
	}

	resp, err := indexer.Search(ctx,
		registry.WithStrKey(key),
		registry.WithPrefix(),
		registry.WithCountOnly())
	if err != nil {
		return quota.NewApplyQuotaResult(nil, scerr.NewError(scerr.ErrInternal, err.Error()))
	}
	num := resp.Count + int64(res.QuotaSize)
	util.Logger().Debugf("resource num is %d", num)
	if num > max {
		mes := fmt.Sprintf("no quota to apply %s, max quota is %d, current used quota is %d, apply quota num is %d",
			res.QuotaType, max, resp.Count, res.QuotaSize)
		util.Logger().Errorf(nil, "%s,  serviceId %s", mes, serviceId)
		return quota.NewApplyQuotaResult(nil, scerr.NewError(scerr.ErrNotEnoughQuota, mes))
	}
	return quota.NewApplyQuotaResult(nil, nil)
}

type QuotaApplyData struct {
	domain        string
	project       string
	domainProject string
	quotaSize     int64
}

type GetCurUsedNum func(context.Context, *QuotaApplyData) (int64, error)
type GetLimitQuota func() int64

type QuotaCheckResult struct {
	IsOk   bool
	CurNum int64
	Err    error
}

func NewQuotaCheckResult(isOk bool, curNum int64, err error) QuotaCheckResult {
	return QuotaCheckResult{
		isOk,
		curNum,
		err,
	}
}

func quotaCheck(ctx context.Context, data *QuotaApplyData, getLimitQuota GetLimitQuota, getCurUsedNum GetCurUsedNum) QuotaCheckResult {
	limitQuota := getLimitQuota()
	curNum, err := getCurUsedNum(ctx, data)
	if err != nil {
		return NewQuotaCheckResult(false, 0, err)
	}
	if curNum+data.quotaSize > limitQuota {
		return NewQuotaCheckResult(false, curNum, nil)
	}
	return NewQuotaCheckResult(true, curNum, nil)
}

func instanceQuotaCheck(ctx context.Context, data *QuotaApplyData) *quota.ApplyQuotaResult {
	rst := quotaCheck(ctx, data, getInstanceMaxLimit, getAllInstancesNum)
	err := rst.Err
	if rst.Err != nil {
		util.Logger().Errorf(rst.Err, "instance quota check failed")
		return quota.NewApplyQuotaResult(nil, scerr.NewError(scerr.ErrInternal, err.Error()))
	}
	if !rst.IsOk {
		mes := fmt.Sprintf("no quota to create instance, max num is %d, curNum is %d, apply num is %d",
			getInstanceMaxLimit(), rst.CurNum, data.quotaSize)
		util.Logger().Errorf(nil, mes)
		return quota.NewApplyQuotaResult(nil, scerr.NewError(scerr.ErrNotEnoughQuota, mes))
	}
	return quota.NewApplyQuotaResult(nil, nil)
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

func serviceQuotaCheck(ctx context.Context, data *QuotaApplyData) *quota.ApplyQuotaResult {
	rst := quotaCheck(ctx, data, getServiceMaxLimit, getAllServicesNum)
	err := rst.Err
	if err != nil {
		util.Logger().Errorf(err, "service quota check failed")
		return quota.NewApplyQuotaResult(nil, scerr.NewError(scerr.ErrInternal, err.Error()))
	}
	if !rst.IsOk {
		mes := fmt.Sprintf("no quota to create service, max quota is %d, current used quota is %d, apply quota num is %d",
			getServiceMaxLimit(), rst.CurNum, data.quotaSize)
		util.Logger().Errorf(err, mes)
		return quota.NewApplyQuotaResult(nil, scerr.NewError(scerr.ErrNotEnoughQuota, mes))
	}
	return quota.NewApplyQuotaResult(nil,nil)
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
