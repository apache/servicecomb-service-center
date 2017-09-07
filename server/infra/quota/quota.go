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
	"github.com/astaxie/beego"
	"golang.org/x/net/context"
)

type QuotaManager interface {
	Apply4Quotas(quotaType ResourceType, tenant string, serviceId string, quotaSize int16) (bool, error)
	ReportCurrentQuotasUsage(ctx context.Context, quotaType int, usedQuotaSize int16) bool
}

var QuataType string

func init() {
	QuataType = beego.AppConfig.DefaultString("quota_plugin", "buildin")
	switch QuataType {
	case "buildin":
	case "fusionstage":
	case "unlimit":
	default:
		QuataType = "buildin"
	}
}

var QuotaPlugins map[string]func() QuotaManager

const (
	RULEQuotaType ResourceType = iota
	SCHEMAQuotaType
	TAGQuotaType
	MicroServiceQuotaType
	MicroServiceInstanceQuotaType
	endType
)

type ResourceType int

func init() {
	QuotaPlugins = make(map[string]func() QuotaManager)
}
