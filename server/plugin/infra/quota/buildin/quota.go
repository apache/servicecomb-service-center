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
	"github.com/ServiceComb/service-center/server/core/registry"
	"github.com/ServiceComb/service-center/server/core/registry/store"
	"github.com/ServiceComb/service-center/server/infra/quota"
	"golang.org/x/net/context"
)

type BuildInQuota struct {
}

func New() quota.QuotaManager {
	return &BuildInQuota{}
}
func init() {
	core.SchemaIdRule.Length = SCHEMA_NUM_MAX_FOR_ONESERVICE
	core.TagRule.Length = TAG_MAX_NUM_FOR_ONESERVICE
	quota.QuotaPlugins["buildin"] = New
}

const (
	SERVICE_MAX_NUMBER            = 12000
	INSTANCE_MAX_NUMBER           = 150000
	RULE_NUM_MAX_FOR_ONESERVICE   = 100
	SCHEMA_NUM_MAX_FOR_ONESERVICE = 1000
	TAG_MAX_NUM_FOR_ONESERVICE    = 100
)

//申请配额sourceType serviceinstance servicetype
func (q *BuildInQuota) Apply4Quotas(ctx context.Context, quotaType quota.ResourceType, domainProject string, serviceId string, quotaSize int16) (quota.QuotaReporter, bool, error) {
	var key string = ""
	var max int64 = 0
	var indexer *store.Indexer
	switch quotaType {
	case quota.MicroServiceInstanceQuotaType:
		key = core.GetInstanceRootKey(domainProject) + "/"
		max = INSTANCE_MAX_NUMBER
		indexer = store.Store().Instance()
	case quota.MicroServiceQuotaType:
		key = core.GetServiceRootKey(domainProject) + "/"
		max = SERVICE_MAX_NUMBER
		indexer = store.Store().Service()
	default:
		return ResourceLimitHandler(ctx, quotaType, domainProject, serviceId, quotaSize)
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
		util.Logger().Errorf(nil, "no quota to apply this source, %s", serviceId)
		return nil, false, nil
	}
	return nil, true, nil
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
		max = RULE_NUM_MAX_FOR_ONESERVICE
		indexer = store.Store().Rule()
	case quota.SCHEMAQuotaType:
		key = core.GenerateServiceSchemaKey(domainProject, serviceId, "")
		max = SCHEMA_NUM_MAX_FOR_ONESERVICE
		indexer = store.Store().Schema()
	case quota.TAGQuotaType:
		num := quotaSize
		if num > TAG_MAX_NUM_FOR_ONESERVICE {
			util.Logger().Errorf(nil,
				"fail to add tag for one service max tag num is %d, %s",
				TAG_MAX_NUM_FOR_ONESERVICE, serviceId)
			return nil, false, nil
		}
		return nil, true, nil
	default:
		return nil, false, fmt.Errorf("Unsurported Type %v", quotaType)
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
		util.Logger().Errorf(nil, "no quota to apply this source, %s", serviceId)
		return nil, false, nil
	}
	return nil, true, nil
}
