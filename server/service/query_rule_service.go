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
package service

import (
	apt "github.com/servicecomb/service-center/server/core"
	pb "github.com/servicecomb/service-center/server/core/proto"
	"github.com/servicecomb/service-center/server/core/registry"
	"github.com/servicecomb/service-center/util"
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"strings"
)


func Accessible(ctx context.Context, tenant string, consumerID string, providerID string) (bool, error, bool) {
	consumerService, err := getServiceByServiceId(ctx, tenant, consumerID)
	if err != nil {
		return false, err, true
	}
	if consumerService == nil {
		return false, fmt.Errorf("consumer invalid"), false
	}

	// 跨应用权限
	providerService, err := getServiceByServiceId(ctx, tenant, providerID)
	if err != nil {
		return false, err, true
	}
	if providerService == nil {
		return false, fmt.Errorf("provider invalid"), false
	}

	if providerService.AppId != consumerService.AppId {
		if len(providerService.Properties) == 0 {
                        util.LOGGER.Warnf(nil, "provider service %s/%s/%s is not defined whether allowed to invoked by other AppId or not, consumer service is %s/%s/%s",
                                providerService.AppId, providerService.ServiceName, providerService.Version,
                                consumerService.AppId, consumerService.ServiceName, consumerService.Version)
			return false, nil, false
		}

		if allowCrossApp, ok := providerService.Properties[pb.PROP_ALLOW_CROSS_APP]; !ok || strings.ToLower(allowCrossApp) != "true" {
                        util.LOGGER.Warnf(nil, "provider service %s/%s/%s is not allowed to be invoked by other AppId, consumer service is %s/%s/%s, %s=%s",
                                providerService.AppId, providerService.ServiceName, providerService.Version,
                                consumerService.AppId, consumerService.ServiceName, consumerService.Version,
				pb.PROP_ALLOW_CROSS_APP, allowCrossApp)
			return false, nil, false
		}
	}

	// 黑白名单
	validateTags, err := GetTagsUtils(ctx, tenant, consumerService.ServiceId)
	if err != nil {
		return false, err, true
	}

	ruleResp, err := registry.GetRegisterCenter().Do(ctx, &registry.PluginOp{
		Action:     registry.GET,
		Key:        []byte(apt.GenerateServiceRuleKey(tenant, providerID, "")),
		WithPrefix: true,
	})
	if err != nil {
		return false, fmt.Errorf("query service rules failed,%s", err), true
	}

	tagPattern := "tag_(.*)"
	tagRegEx, _ := regexp.Compile(tagPattern)
	hasWhite := false
	for _, kv := range ruleResp.Kvs {
		var rule pb.ServiceRule
		var value string
		err := json.Unmarshal(kv.Value, &rule)
		if err != nil {
			return false, fmt.Errorf("unmarshal service rules failed,%s", err), true
		}
		tagRule, _ := regexp.MatchString(tagPattern, rule.Attribute)
		if tagRule {
			key := tagRegEx.FindStringSubmatch(rule.Attribute)[1]
			value = validateTags[key]
		} else {
			v := reflect.ValueOf(consumerService)
			key := reflect.Indirect(v).FieldByName(rule.Attribute)
			value = key.String()
		}

		switch rule.RuleType {
		case "WHITE":
			hasWhite = true
			match, _ := regexp.MatchString(rule.Pattern, value)
			util.LOGGER.Debugf("match is %s, rule.Pattern is %s, value is %s", match, rule.Pattern, value)
			if match {
				return true, nil, false
			}

		case "BLACK":
			match, _ := regexp.MatchString(rule.Pattern, value)
			if match {
				return false, nil, false
			}
		}

	}
	if hasWhite {
		return false, nil, false
	}
	return true, nil, false

}
