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
package quota

import (
	"fmt"
	"golang.org/x/net/context"
)

type QuotaManager interface {
	Apply4Quotas(ctx context.Context, quotaType ResourceType, domainProject string, serviceId string, quotaSize int16) (QuotaReporter, bool, error)
	RemandQuotas(ctx context.Context, quotaType ResourceType)
}

type QuotaReporter interface {
	ReportUsedQuota(ctx context.Context) error
	Close()
}

const (
	RuleQuotaType                 ResourceType = iota
	SchemaQuotaType
	TagQuotaType
	MicroServiceQuotaType
	MicroServiceInstanceQuotaType
	typeEnd
)

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
		return "RESOURCE" + fmt.Sprint(r)
	}
}
