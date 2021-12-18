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

const TypeTag quota.ResourceType = "TAG"

func TagQuota() int64 {
	return quota.GetQuota(context.Background(), TypeTag)
}

func ApplyTag(ctx context.Context, serviceID string, size int64) error {
	return quota.Apply(ctx, &quota.Request{
		QuotaType:     TypeTag,
		DomainProject: util.ParseDomainProject(ctx),
		ServiceID:     serviceID,
		QuotaSize:     size,
	})
}
