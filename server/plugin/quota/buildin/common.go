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
	"context"
	"errors"
	"fmt"

	"github.com/apache/servicecomb-service-center/datasource"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/metrics"
	"github.com/apache/servicecomb-service-center/server/plugin/quota"
	pb "github.com/go-chassis/cari/discovery"
)

const (
	TotalService  = "db_service_total"
	TotalInstance = "db_instance_total"
)

type GetCurUsedNum func(context.Context, *quota.ApplyQuotaResource) (int64, error)
type GetLimitQuota func() int64

func CommonQuotaCheck(ctx context.Context, res *quota.ApplyQuotaResource,
	getLimitQuota GetLimitQuota, getCurUsedNum GetCurUsedNum) *quota.ApplyQuotaResult {
	if res == nil || getLimitQuota == nil || getCurUsedNum == nil {
		err := errors.New("invalid parameters")
		log.Errorf(err, "quota check failed")
		return quota.NewApplyQuotaResult(nil, pb.NewError(pb.ErrInternal, err.Error()))
	}

	limitQuota := getLimitQuota()
	curNum, err := getCurUsedNum(ctx, res)
	if err != nil {
		log.Errorf(err, "%s quota check failed", res.QuotaType)
		return quota.NewApplyQuotaResult(nil, pb.NewError(pb.ErrInternal, err.Error()))
	}
	if curNum+res.QuotaSize > limitQuota {
		mes := fmt.Sprintf("no quota to create %s, max num is %d, curNum is %d, apply num is %d",
			res.QuotaType, limitQuota, curNum, res.QuotaSize)
		log.Errorf(nil, mes)
		return quota.NewApplyQuotaResult(nil, pb.NewError(pb.ErrNotEnoughQuota, mes))
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
	serviceID := res.ServiceID

	switch res.QuotaType {
	case quota.MicroServiceQuotaType:
		return metrics.GaugeValue(TotalService, nil), nil
	case quota.MicroServiceInstanceQuotaType:
		return metrics.GaugeValue(TotalInstance, nil), nil
	case quota.RuleQuotaType:
		{
			resp, err := datasource.Instance().GetRules(ctx, &pb.GetServiceRulesRequest{
				ServiceId: serviceID,
			})
			if err != nil {
				return 0, err
			}
			return int64(len(resp.Rules)), nil
		}
	case quota.SchemaQuotaType:
		{
			resp, err := datasource.Instance().GetAllSchemas(ctx, &pb.GetAllSchemaRequest{
				ServiceId:  serviceID,
				WithSchema: false,
			})
			if err != nil {
				return 0, err
			}
			return int64(len(resp.Schemas)), nil
		}
	case quota.TagQuotaType:
		// always re-create the service old tags
		return 0, nil
	default:
		return 0, fmt.Errorf("not define quota type '%s'", res.QuotaType)
	}
}
