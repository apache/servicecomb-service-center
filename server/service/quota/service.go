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

package quota

import (
	"context"

	"github.com/apache/servicecomb-service-center/pkg/util"
	"github.com/apache/servicecomb-service-center/server/plugin/quota"
)

const TypeService quota.ResourceType = "SERVICE"

func ServiceQuota() int64 {
	return quota.GetQuota(context.Background(), TypeService)
}

func ApplyService(ctx context.Context, size int64) error {
	return quota.Apply(ctx, &quota.Request{
		QuotaType: TypeService,
		Domain:    util.ParseDomain(ctx),
		Project:   util.ParseProject(ctx),
		QuotaSize: size,
	})
}

func RemandService(ctx context.Context) {
	quota.Remand(ctx, TypeService)
}
