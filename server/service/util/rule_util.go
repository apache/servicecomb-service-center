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
package util

import (
	"encoding/json"
	"fmt"
	apt "github.com/ServiceComb/service-center/server/core"
	pb "github.com/ServiceComb/service-center/server/core/proto"
	"github.com/ServiceComb/service-center/server/core/registry"
	"github.com/ServiceComb/service-center/server/core/registry/store"
	"github.com/ServiceComb/service-center/util"
	errorsEx "github.com/ServiceComb/service-center/util/errors"
	"golang.org/x/net/context"
	"reflect"
	"regexp"
	"strings"
)

type NotAllowAcrossAppError string

func (e NotAllowAcrossAppError) Error() string {
	return string(e)
}

type NotMatchWhiteListError string

func (e NotMatchWhiteListError) Error() string {
	return string(e)
}

type MatchBlackListError string

func (e MatchBlackListError) Error() string {
	return string(e)
}

func GetRulesUtil(ctx context.Context, tenant string, serviceId string) ([]*pb.ServiceRule, error) {
	key := util.StringJoin([]string{
		apt.GetServiceRuleRootKey(tenant),
		serviceId,
		"",
	}, "/")

	resp, err := store.Store().Rule().Search(ctx, &registry.PluginOp{
		Action:     registry.GET,
		Key:        util.StringToBytesWithNoCopy(key),
		WithPrefix: true,
	})
	if err != nil {
		return nil, err
	}

	rules := []*pb.ServiceRule{}
	for _, kvs := range resp.Kvs {
		util.LOGGER.Debugf("start unmarshal service rule file: %s", string(kvs.Key))
		rule := &pb.ServiceRule{}
		err := json.Unmarshal(kvs.Value, rule)
		if err != nil {
			return nil, err
		}
		rules = append(rules, rule)
	}
	return rules, nil
}

func RuleExist(ctx context.Context, tenant string, serviceId string, attr string, pattern string) bool {
	resp, err := store.Store().RuleIndex().Search(ctx, &registry.PluginOp{
		Action:    registry.GET,
		Key:       util.StringToBytesWithNoCopy(apt.GenerateRuleIndexKey(tenant, serviceId, attr, pattern)),
		CountOnly: true,
	})
	if err != nil || resp.Count == 0 {
		return false
	}
	return true
}

func GetServiceRuleType(ctx context.Context, tenant string, serviceId string) (string, int, error) {
	key := apt.GenerateServiceRuleKey(tenant, serviceId, "")
	resp, err := store.Store().Rule().Search(ctx, &registry.PluginOp{
		Action:     registry.GET,
		Key:        util.StringToBytesWithNoCopy(key),
		WithPrefix: true,
	})
	if err != nil {
		util.LOGGER.Errorf(err, "Get rule failed.%s", err.Error())
		return "", 0, err
	}
	if len(resp.Kvs) == 0 {
		return "", 0, nil
	}
	rule := &pb.ServiceRule{}
	err = json.Unmarshal(resp.Kvs[0].Value, rule)
	if err != nil {
		util.LOGGER.Errorf(err, "Unmarshal rule data failed.%s", err.Error())
	}
	return rule.RuleType, len(resp.Kvs), nil
}

func GetOneRule(ctx context.Context, tenant, serviceId, ruleId string) (*pb.ServiceRule, error) {
	opt := &registry.PluginOp{
		Action: registry.GET,
		Key:    util.StringToBytesWithNoCopy(apt.GenerateServiceRuleKey(tenant, serviceId, ruleId)),
	}
	resp, err := store.Store().Rule().Search(ctx, opt)
	if err != nil {
		util.LOGGER.Errorf(nil, "Get rule for service failed for %s.", err.Error())
		return nil, err
	}
	rule := &pb.ServiceRule{}
	if len(resp.Kvs) == 0 {
		util.LOGGER.Errorf(nil, "Get rule failed, ruleId is %s.", ruleId)
		return nil, nil
	}
	err = json.Unmarshal(resp.Kvs[0].Value, rule)
	if err != nil {
		util.LOGGER.Errorf(nil, "unmarshal resp failed for %s.", err.Error())
		return nil, err
	}
	return rule, nil
}

func AllowAcrossApp(providerService *pb.MicroService, consumerService *pb.MicroService) error {
	if providerService.AppId != consumerService.AppId {
		if len(providerService.Properties) == 0 {
			return NotAllowAcrossAppError("not allow across app access")
		}

		if allowCrossApp, ok := providerService.Properties[pb.PROP_ALLOW_CROSS_APP]; !ok || strings.ToLower(allowCrossApp) != "true" {
			return NotAllowAcrossAppError("not allow across app access")
		}
	}
	return nil
}

func MatchRules(rules []*pb.ServiceRule, service *pb.MicroService, serviceTags map[string]string) error {
	v := reflect.Indirect(reflect.ValueOf(service))

	tagPattern := "tag_(.*)"
	tagRegEx, _ := regexp.Compile(tagPattern)
	hasWhite := false
	for _, rule := range rules {
		var value string
		if tagRegEx.MatchString(rule.Attribute) {
			key := tagRegEx.FindStringSubmatch(rule.Attribute)[1]
			value = serviceTags[key]
			if len(value) == 0 {
				util.LOGGER.Infof("can not find service %s tag '%s'", service.ServiceId, key)
				continue
			}
		} else {
			key := v.FieldByName(rule.Attribute)
			if !key.IsValid() {
				return errorsEx.InternalError(fmt.Sprintf("can not find field '%s'", rule.Attribute))
			}
			value = key.String()
		}

		switch rule.RuleType {
		case "WHITE":
			hasWhite = true
			match, _ := regexp.MatchString(rule.Pattern, value)
			if match {
				util.LOGGER.Infof("service %s match white list, rule.Pattern is %s, value is %s",
					service.ServiceId, rule.Pattern, value)
				return nil
			}
		case "BLACK":
			match, _ := regexp.MatchString(rule.Pattern, value)
			if match {
				util.LOGGER.Infof("service %s match black list, rule.Pattern is %s, value is %s",
					service.ServiceId, rule.Pattern, value)
				return MatchBlackListError("Found in black list")
			}
		}

	}
	if hasWhite {
		return NotMatchWhiteListError("Not found in white list")
	}
	return nil
}
