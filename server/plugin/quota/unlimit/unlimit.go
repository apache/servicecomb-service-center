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

package unlimit

import (
	"context"

	"github.com/apache/servicecomb-service-center/pkg/log"
	mgr "github.com/apache/servicecomb-service-center/server/plugin"
	"github.com/apache/servicecomb-service-center/server/plugin/quota"
)

func init() {
	mgr.RegisterPlugin(mgr.Plugin{Kind: quota.QUOTA, Name: "unlimit", New: New})
}

type Unlimit struct {
}

func New() mgr.Instance {
	log.Warnf("quota init, all resources are unlimited")
	return &Unlimit{}
}

func (q *Unlimit) Apply4Quotas(ctx context.Context, res *quota.ApplyQuotaResource) *quota.ApplyQuotaResult {
	return quota.NewApplyQuotaResult(nil, nil)
}

func (q *Unlimit) RemandQuotas(ctx context.Context, quotaType quota.ResourceType) {
}
