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
	"github.com/servicecomb/service-center/util/errors"
)


func Accessible(ctx context.Context, tenant string, consumerID string, providerID string) (err error, isInnerErr bool) {
	consumerService, err := getServiceByServiceId(ctx, tenant, consumerID)
	if err != nil {
		return err, true
	}
	if consumerService == nil {
		return fmt.Errorf("consumer invalid"), false
	}

	// 跨应用权限
	providerService, err := getServiceByServiceId(ctx, tenant, providerID)
	if err != nil {
		return err, true
	}
	if providerService == nil {
		return fmt.Errorf("provider invalid"), false
	}

	providerFlag := strings.Join([]string{providerService.AppId, providerService.ServiceName, providerService.Version}, "--")
	consumerFlag := strings.Join([]string{consumerService.AppId, consumerService.ServiceName, consumerService.Version}, "--")
	if providerService.AppId != consumerService.AppId {
		if len(providerService.Properties) == 0 {
			util.LOGGER.Warnf(nil, "consumer %s can't access provider %s, different appid",
				consumerFlag, providerFlag)
			return errors.New("different appID can't access"), false
		}

		if allowCrossApp, ok := providerService.Properties[pb.PROP_ALLOW_CROSS_APP]; !ok || strings.ToLower(allowCrossApp) != "true" {
			util.LOGGER.Warnf(nil, "consumer %s can't access provider %s, different appid, no allowCrossApp defined in property",
				consumerFlag, providerFlag)
			return errors.New("different appID can't access"), false
		}
	}

	// 黑白名单
	validateTags, err := GetTagsUtils(ctx, tenant, consumerService.ServiceId)
	if err != nil {
		return err, true
	}

	ruleResp, err := registry.GetRegisterCenter().Do(ctx, &registry.PluginOp{
		Action:     registry.GET,
		Key:        []byte(apt.GenerateServiceRuleKey(tenant, providerID, "")),
		WithPrefix: true,
	})
	if err != nil {
		return fmt.Errorf("query service rules failed,%s", err), true
	}

	tagPattern := "tag_(.*)"
	tagRegEx, _ := regexp.Compile(tagPattern)
	hasWhite := false
	for _, kv := range ruleResp.Kvs {
		var rule pb.ServiceRule
		var value string
		err := json.Unmarshal(kv.Value, &rule)
		if err != nil {
			return fmt.Errorf("unmarshal service rules failed,%s", err), true
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
		if len(value) == 0 {
			util.LOGGER.Errorf(nil, "get attribute %s matched value is empty.", rule.Attribute)
			return errors.New("can't access, get attribute matched value is empty"), false
		}

		switch rule.RuleType {
		case "WHITE":
			hasWhite = true
			match, _ := regexp.MatchString(rule.Pattern, value)
			util.LOGGER.Debugf("match is %t, rule.Pattern is %s, value is %s", match, rule.Pattern, value)
			if match {
				return nil, false
			}

		case "BLACK":
			match, _ := regexp.MatchString(rule.Pattern, value)
			if match {
				util.LOGGER.Warnf(nil, "match black list %s, can't access.consumer %s, provider %s", rule.Pattern, consumerFlag, providerFlag)
				return errors.New("match black list,can't access."), false
			}
		}

	}
	if hasWhite {
		util.LOGGER.Warnf(nil, "not match white list , can't access.consumer %s, provider %s", consumerFlag, providerFlag)
		return errors.New("not match white list,can't access."), false
	}
	return nil, false
}
