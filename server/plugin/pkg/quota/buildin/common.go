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
	"errors"
	"fmt"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/server/core"
	"github.com/apache/servicecomb-service-center/server/core/backend"
	scerr "github.com/apache/servicecomb-service-center/server/error"
	mgr "github.com/apache/servicecomb-service-center/server/plugin"
	"github.com/apache/servicecomb-service-center/server/plugin/pkg/discovery"
	"github.com/apache/servicecomb-service-center/server/plugin/pkg/quota"
	"github.com/apache/servicecomb-service-center/server/plugin/pkg/registry"
	serviceUtil "github.com/apache/servicecomb-service-center/server/service/util"
	"golang.org/x/net/context"
)

type GetCurUsedNum func(context.Context, *quota.ApplyQuotaResource) (int64, error)
type GetLimitQuota func() int64

func CommonQuotaCheck(ctx context.Context, res *quota.ApplyQuotaResource,
	getLimitQuota GetLimitQuota, getCurUsedNum GetCurUsedNum) *quota.ApplyQuotaResult {
	if res == nil || getLimitQuota == nil || getCurUsedNum == nil {
		err := errors.New("invalid parameters")
		log.Errorf(err, "quota check failed")
		return quota.NewApplyQuotaResult(nil, scerr.NewError(scerr.ErrInternal, err.Error()))
	}

	limitQuota := getLimitQuota()
	curNum, err := getCurUsedNum(ctx, res)
	if err != nil {
		log.Errorf(err, "%s quota check failed", res.QuotaType)
		return quota.NewApplyQuotaResult(nil, scerr.NewError(scerr.ErrInternal, err.Error()))
	}
	if curNum+res.QuotaSize > limitQuota {
		mes := fmt.Sprintf("no quota to create %s, max num is %d, curNum is %d, apply num is %d",
			res.QuotaType, limitQuota, curNum, res.QuotaSize)
		log.Errorf(nil, mes)
		return quota.NewApplyQuotaResult(nil, scerr.NewError(scerr.ErrNotEnoughQuota, mes))
	}
	return quota.NewApplyQuotaResult(nil, nil)
}

func resourceQuota(t quota.ResourceType) GetLimitQuota {
	return func() int64 {
		switch t {
		case quota.MicroServiceInstanceQuotaType:
			return int64(quota.DefaultInstanceQuota)
		case quota.MicroServiceQuotaType:
			return int64(quota.DefaultServiceQuota)
		case quota.RuleQuotaType:
			return int64(quota.DefaultRuleQuota)
		case quota.SchemaQuotaType:
			return int64(quota.DefaultSchemaQuota)
		case quota.TagQuotaType:
			return int64(quota.DefaultTagQuota)
		default:
			return 0
		}
	}
}

func resourceLimitHandler(ctx context.Context, res *quota.ApplyQuotaResource) (int64, error) {
	var key string
	var indexer discovery.Indexer

	domainProject := res.DomainProject
	serviceId := res.ServiceId

	switch res.QuotaType {
	case quota.MicroServiceInstanceQuotaType:
		key = core.GetInstanceRootKey("")
		indexer = backend.Store().Instance()
	case quota.MicroServiceQuotaType:
		key = core.GetServiceRootKey("")
		indexer = backend.Store().Service()
	case quota.RuleQuotaType:
		key = core.GenerateServiceRuleKey(domainProject, serviceId, "")
		indexer = backend.Store().Rule()
	case quota.SchemaQuotaType:
		key = core.GenerateServiceSchemaKey(domainProject, serviceId, "")
		indexer = backend.Store().Schema()
	case quota.TagQuotaType:
		tags, err := serviceUtil.GetTagsUtils(ctx, domainProject, serviceId)
		if err != nil {
			return 0, err
		}
		return int64(len(tags)), nil
	default:
		return 0, fmt.Errorf("not define quota type '%s'", res.QuotaType)
	}

	resp, err := indexer.Search(ctx,
		registry.WithStrKey(key),
		registry.WithPrefix(),
		registry.WithCountOnly())
	if err != nil {
		return 0, err
	}
	return resp.Count, nil
}

func InitConfigs() {
	mgr.QUOTA.ActiveConfigs().
		Set("service", quota.DefaultServiceQuota).
		Set("instance", quota.DefaultInstanceQuota).
		Set("schema", quota.DefaultSchemaQuota).
		Set("tag", quota.DefaultTagQuota).
		Set("rule", quota.DefaultRuleQuota)
}
