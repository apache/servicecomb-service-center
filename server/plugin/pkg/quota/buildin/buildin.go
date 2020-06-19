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
	"github.com/apache/servicecomb-service-center/pkg/log"
	mgr "github.com/apache/servicecomb-service-center/server/plugin"
	"github.com/apache/servicecomb-service-center/server/plugin/pkg/quota"
	"github.com/apache/servicecomb-service-center/server/plugin/pkg/quota/counter"
)

func init() {
	mgr.RegisterPlugin(mgr.Plugin{mgr.QUOTA, "buildin", New})
	counter.RegisterCounterListener("buildin")
}

func New() mgr.PluginInstance {
	InitConfigs()
	log.Infof("quota init, service: %d, instance: %d, schema: %d/service, tag: %d/service, rule: %d/service",
		quota.DefaultServiceQuota, quota.DefaultInstanceQuota,
		quota.DefaultSchemaQuota, quota.DefaultTagQuota, quota.DefaultRuleQuota)
	return &BuildInQuota{}
}

type BuildInQuota struct {
}

//申请配额sourceType serviceinstance servicetype
func (q *BuildInQuota) Apply4Quotas(ctx context.Context, res *quota.ApplyQuotaResource) *quota.ApplyQuotaResult {
	df, ok := mgr.DynamicPluginFunc(mgr.QUOTA, "Apply4Quotas").(func(context.Context, *quota.ApplyQuotaResource) *quota.ApplyQuotaResult)
	if ok {
		return df(ctx, res)
	}

	return CommonQuotaCheck(ctx, res, resourceQuota(res.QuotaType), resourceLimitHandler)
}

//向配额中心上报配额使用量
func (q *BuildInQuota) RemandQuotas(ctx context.Context, quotaType quota.ResourceType) {
	df, ok := mgr.DynamicPluginFunc(mgr.QUOTA, "RemandQuotas").(func(context.Context, quota.ResourceType))
	if ok {
		df(ctx, quotaType)
		return
	}
}
