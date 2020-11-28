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
	"strconv"

	"github.com/apache/servicecomb-service-center/server/config"
	"github.com/apache/servicecomb-service-center/server/plugin"
	"github.com/go-chassis/cari/discovery"
)

const QUOTA plugin.Kind = "quota"

const (
	defaultServiceLimit  = 50000
	defaultInstanceLimit = 150000
	defaultSchemaLimit   = 100
	defaultRuleLimit     = 100
	defaultTagLimit      = 100
)

const (
	RuleQuotaType ResourceType = iota
	SchemaQuotaType
	TagQuotaType
	MicroServiceQuotaType
	MicroServiceInstanceQuotaType
)

var (
	DefaultServiceQuota  = defaultServiceLimit
	DefaultInstanceQuota = defaultInstanceLimit
	DefaultSchemaQuota   = defaultSchemaLimit
	DefaultTagQuota      = defaultTagLimit
	DefaultRuleQuota     = defaultRuleLimit
)

func Init() {
	DefaultServiceQuota = config.GetInt("quota.cap.service", defaultServiceLimit, config.WithStandby("QUOTA_SERVICE"))
	DefaultInstanceQuota = config.GetInt("quota.cap.instance", defaultInstanceLimit, config.WithStandby("QUOTA_INSTANCE"))
	DefaultSchemaQuota = config.GetInt("quota.cap.schema", defaultSchemaLimit, config.WithStandby("QUOTA_SCHEMA"))
	DefaultTagQuota = config.GetInt("quota.cap.tag", defaultTagLimit, config.WithStandby("QUOTA_TAG"))
	DefaultRuleQuota = config.GetInt("quota.cap.rule", defaultRuleLimit, config.WithStandby("QUOTA_RULE"))
}

type ApplyQuotaResult struct {
	Err *discovery.Error

	reporter Reporter
}

func (r *ApplyQuotaResult) ReportUsedQuota(ctx context.Context) error {
	if r == nil || r.reporter == nil {
		return nil
	}
	return r.reporter.ReportUsedQuota(ctx)
}

func (r *ApplyQuotaResult) Close(ctx context.Context) {
	if r == nil || r.reporter == nil {
		return
	}
	r.reporter.Close(ctx)
}

func NewApplyQuotaResult(reporter Reporter, err *discovery.Error) *ApplyQuotaResult {
	return &ApplyQuotaResult{
		reporter: reporter,
		Err:      err,
	}
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
	Apply4Quotas(ctx context.Context, res *ApplyQuotaResource) *ApplyQuotaResult
	RemandQuotas(ctx context.Context, quotaType ResourceType)
}

type Reporter interface {
	ReportUsedQuota(ctx context.Context) error
	Close(ctx context.Context)
}

type ResourceType int

func (r ResourceType) String() string {
	switch r {
	case RuleQuotaType:
		return "RULE"
	case SchemaQuotaType:
		return "SCHEMA"
	case TagQuotaType:
		return "TAG"
	case MicroServiceQuotaType:
		return "SERVICE"
	case MicroServiceInstanceQuotaType:
		return "INSTANCE"
	default:
		return "RESOURCE" + strconv.Itoa(int(r))
	}
}

func Apply(ctx context.Context, res *ApplyQuotaResource) *ApplyQuotaResult {
	return plugin.Plugins().Instance(QUOTA).(Manager).Apply4Quotas(ctx, res)
}

func Remand(ctx context.Context, quotaType ResourceType) {
	plugin.Plugins().Instance(QUOTA).(Manager).RemandQuotas(ctx, quotaType)
}
