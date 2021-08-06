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
	"errors"
	"fmt"
	"strconv"

	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/plugin"
	"github.com/apache/servicecomb-service-center/pkg/util"
	"github.com/apache/servicecomb-service-center/server/config"
	"github.com/apache/servicecomb-service-center/server/service/quota"
	pb "github.com/go-chassis/cari/discovery"
)

const QUOTA plugin.Kind = "quota"

const (
	defaultServiceLimit  = 50000
	defaultInstanceLimit = 150000
	defaultSchemaLimit   = 100
	defaultTagLimit      = 100
	defaultAccountLimit  = 1000
	defaultRoleLimit     = 100
)

const (
	TypeSchema ResourceType = iota
	TypeTag
	TypeService
	TypeInstance
	TypeAccount
	TypeRole
)

var (
	DefaultServiceQuota  = defaultServiceLimit
	DefaultInstanceQuota = defaultInstanceLimit
	DefaultSchemaQuota   = defaultSchemaLimit
	DefaultTagQuota      = defaultTagLimit
	DefaultAccountQuota  = defaultAccountLimit
	DefaultRoleQuota     = defaultRoleLimit
)

func Init() {
	DefaultServiceQuota = config.GetInt("quota.cap.service.limit", defaultServiceLimit, config.WithENV("QUOTA_SERVICE"))
	DefaultInstanceQuota = config.GetInt("quota.cap.instance.limit", defaultInstanceLimit, config.WithENV("QUOTA_INSTANCE"))
	DefaultSchemaQuota = config.GetInt("quota.cap.schema.limit", defaultSchemaLimit, config.WithENV("QUOTA_SCHEMA"))
	DefaultTagQuota = config.GetInt("quota.cap.tag.limit", defaultTagLimit, config.WithENV("QUOTA_TAG"))
	DefaultAccountQuota = config.GetInt("quota.cap.account.limit", defaultAccountLimit, config.WithENV("QUOTA_ACCOUNT"))
	DefaultRoleQuota = config.GetInt("quota.cap.role.limit", defaultRoleLimit, config.WithENV("QUOTA_ROLE"))
}

type ApplyQuotaResource struct {
	QuotaType     ResourceType
	DomainProject string
	ServiceID     string
	QuotaSize     int64
}

func NewApplyQuotaResource(quotaType ResourceType, domainProject, serviceID string, quotaSize int64) *ApplyQuotaResource {
	return &ApplyQuotaResource{
		quotaType,
		domainProject,
		serviceID,
		quotaSize,
	}
}

type Manager interface {
	RemandQuotas(ctx context.Context, quotaType ResourceType)
	GetQuota(ctx context.Context, t ResourceType) int64
}

type ResourceType int

func (r ResourceType) String() string {
	switch r {
	case TypeSchema:
		return "SCHEMA"
	case TypeTag:
		return "TAG"
	case TypeService:
		return "SERVICE"
	case TypeInstance:
		return "INSTANCE"
	case TypeAccount:
		return "ACCOUNT"
	case TypeRole:
		return "ROLE"
	default:
		return "RESOURCE" + strconv.Itoa(int(r))
	}
}

// Apply 申请配额sourceType serviceinstance servicetype
func Apply(ctx context.Context, res *ApplyQuotaResource) error {
	if res == nil {
		err := errors.New("invalid parameters")
		log.Error("quota check failed", err)
		return pb.NewError(pb.ErrInternal, err.Error())
	}

	limitQuota := plugin.Plugins().Instance(QUOTA).(Manager).GetQuota(ctx, res.QuotaType)
	curNum, err := GetResourceUsage(ctx, res)
	if err != nil {
		log.Error(fmt.Sprintf("%s quota check failed", res.QuotaType), err)
		return pb.NewError(pb.ErrInternal, err.Error())
	}
	if curNum+res.QuotaSize > limitQuota {
		mes := fmt.Sprintf("no quota to create %s, max num is %d, curNum is %d, apply num is %d",
			res.QuotaType, limitQuota, curNum, res.QuotaSize)
		log.Error(mes, nil)
		return pb.NewError(pb.ErrNotEnoughQuota, mes)
	}
	return nil
}

func Remand(ctx context.Context, quotaType ResourceType) {
	plugin.Plugins().Instance(QUOTA).(Manager).RemandQuotas(ctx, quotaType)
}
func GetResourceUsage(ctx context.Context, res *ApplyQuotaResource) (int64, error) {
	serviceID := res.ServiceID
	switch res.QuotaType {
	case TypeService:
		return quota.ServiceUsage(ctx, &pb.GetServiceCountRequest{
			Domain:  util.ParseDomain(ctx),
			Project: util.ParseProject(ctx),
		})
	case TypeInstance:
		return quota.InstanceUsage(ctx, &pb.GetServiceCountRequest{
			Domain:  util.ParseDomain(ctx),
			Project: util.ParseProject(ctx),
		})
	case TypeSchema:
		return quota.SchemaUsage(ctx, serviceID)
	case TypeTag:
		// always re-create the service old tags
		return 0, nil
	case TypeRole:
		return quota.RoleUsage(ctx)
	case TypeAccount:
		return quota.AccountUsage(ctx)
	default:
		return 0, fmt.Errorf("not define quota type '%s'", res.QuotaType)
	}
}
