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
	"fmt"

	pb "github.com/go-chassis/cari/discovery"
	ev "github.com/go-chassis/cari/env"

	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/plugin"
	"github.com/apache/servicecomb-service-center/pkg/util"
	"github.com/apache/servicecomb-service-center/server/config"
	"github.com/apache/servicecomb-service-center/server/plugin/quota"
	"github.com/apache/servicecomb-service-center/server/service/disco"
	quotasvc "github.com/apache/servicecomb-service-center/server/service/quota"
	"github.com/apache/servicecomb-service-center/server/service/rbac"
)

const (
	defaultServiceLimit     int64 = 50000
	defaultInstanceLimit    int64 = 150000
	defaultSchemaLimit      int64 = 100
	defaultTagLimit         int64 = 100
	defaultAccountLimit     int64 = 1000
	defaultRoleLimit        int64 = 100
	defaultEnvironmentLimit int64 = 50
)

func init() {
	plugin.RegisterPlugin(plugin.Plugin{Kind: quota.QUOTA, Name: "buildin", New: New})
}

func New() plugin.Instance {
	q := &Quota{
		ServiceQuota:     config.GetInt64("quota.cap.service.limit", defaultServiceLimit, config.WithENV("QUOTA_SERVICE")),
		InstanceQuota:    config.GetInt64("quota.cap.instance.limit", defaultInstanceLimit, config.WithENV("QUOTA_INSTANCE")),
		SchemaQuota:      config.GetInt64("quota.cap.schema.limit", defaultSchemaLimit, config.WithENV("QUOTA_SCHEMA")),
		TagQuota:         config.GetInt64("quota.cap.tag.limit", defaultTagLimit, config.WithENV("QUOTA_TAG")),
		AccountQuota:     config.GetInt64("quota.cap.account.limit", defaultAccountLimit, config.WithENV("QUOTA_ACCOUNT")),
		RoleQuota:        config.GetInt64("quota.cap.role.limit", defaultRoleLimit, config.WithENV("QUOTA_ROLE")),
		EnvironmentQuota: config.GetInt64("quota.cap.environment.limit", defaultEnvironmentLimit, config.WithENV("QUOTA_ENVIRONMENT")),
	}
	log.Info(fmt.Sprintf("quota init, service: %d, instance: %d, schema: %d/service, tag: %d/service"+
		", account: %d, role: %d",
		q.ServiceQuota, q.InstanceQuota, q.SchemaQuota, q.TagQuota,
		q.AccountQuota, q.RoleQuota))
	return q
}

type Quota struct {
	ServiceQuota     int64
	InstanceQuota    int64
	SchemaQuota      int64
	TagQuota         int64
	AccountQuota     int64
	RoleQuota        int64
	EnvironmentQuota int64
}

func (q *Quota) GetQuota(_ context.Context, t quota.ResourceType) int64 {
	switch t {
	case quotasvc.TypeInstance:
		return q.InstanceQuota
	case quotasvc.TypeService:
		return q.ServiceQuota
	case quotasvc.TypeSchema:
		return q.SchemaQuota
	case quotasvc.TypeTag:
		return q.TagQuota
	case quotasvc.TypeAccount:
		return q.AccountQuota
	case quotasvc.TypeRole:
		return q.RoleQuota
	case quotasvc.TypeEnvironment:
		return q.EnvironmentQuota
	default:
		return 0
	}
}

// 向配额中心上报配额使用量
func (q *Quota) RemandQuotas(ctx context.Context, resourceType quota.ResourceType) {
	df, ok := plugin.DynamicPluginFunc(quota.QUOTA, "RemandQuotas").(func(context.Context, quota.ResourceType))
	if ok {
		df(ctx, resourceType)
		return
	}
}

func (q *Quota) Usage(ctx context.Context, req *quota.Request) (int64, error) {
	switch req.QuotaType {
	case quotasvc.TypeInstance:
		return disco.InstanceUsage(ctx, &pb.GetServiceCountRequest{
			Domain:  util.ParseDomain(ctx),
			Project: util.ParseProject(ctx),
		})
	case quotasvc.TypeService:
		return disco.ServiceUsage(ctx, &pb.GetServiceCountRequest{
			Domain:  util.ParseDomain(ctx),
			Project: util.ParseProject(ctx),
		})
	case quotasvc.TypeSchema:
		return disco.Usage(ctx, req.ServiceID)
	case quotasvc.TypeTag:
		// always re-create the service old tags
		return 0, nil
	case quotasvc.TypeAccount:
		return rbac.AccountUsage(ctx)
	case quotasvc.TypeRole:
		return rbac.RoleUsage(ctx)
	case quotasvc.TypeEnvironment:
		return disco.EnvironmentUsage(ctx, &ev.GetEnvironmentCountRequest{
			Domain:  util.ParseDomain(ctx),
			Project: util.ParseProject(ctx),
		})
	default:
		return 0, nil
	}
}
