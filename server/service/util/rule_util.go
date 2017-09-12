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
	"github.com/ServiceComb/service-center/server/service/dependency"
	ms "github.com/ServiceComb/service-center/server/service/microservice"
	"github.com/ServiceComb/service-center/util"
	errorsEx "github.com/ServiceComb/service-center/util/errors"
	"golang.org/x/net/context"
	"reflect"
	"regexp"
	"strings"
)

var tagRegEx *regexp.Regexp

func init() {
	tagRegEx, _ = regexp.Compile("tag_(.*)")
}

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

type RuleFilter struct {
	Tenant        string
	Provider      *pb.MicroService
	ProviderRules []*pb.ServiceRule
}

func (rf *RuleFilter) Filter(ctx context.Context, consumerId string) (bool, error) {
	consumer, err := ms.SearchService(ctx, rf.Tenant, consumerId, registry.MODE_CACHE)
	if consumer == nil {
		return false, err
	}

	tags, err := SearchTags(context.Background(), rf.Tenant, consumerId, registry.MODE_CACHE)
	if err != nil {
		return false, err
	}
	matchErr := MatchRules(rf.ProviderRules, consumer, tags)
	switch matchErr.(type) {
	case NotMatchWhiteListError, MatchBlackListError:
		return false, nil
	default:
	}
	return true, nil
}

func GetRulesUtil(ctx context.Context, tenant string, serviceId string) ([]*pb.ServiceRule, error) {
	key := util.StringJoin([]string{
		apt.GetServiceRuleRootKey(tenant),
		serviceId,
		"",
	}, "/")

	resp, err := store.Store().Rule().Search(ctx,
		registry.WithStrKey(key),
		registry.WithPrefix())
	if err != nil {
		return nil, err
	}

	rules := []*pb.ServiceRule{}
	for _, kvs := range resp.Kvs {
		util.Logger().Debugf("start unmarshal service rule file: %s", util.BytesToStringWithNoCopy(kvs.Key))
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
	resp, err := store.Store().RuleIndex().Search(ctx,
		registry.WithStrKey(apt.GenerateRuleIndexKey(tenant, serviceId, attr, pattern)),
		registry.WithCountOnly())
	if err != nil || resp.Count == 0 {
		return false
	}
	return true
}

func GetServiceRuleType(ctx context.Context, tenant string, serviceId string) (string, int, error) {
	key := apt.GenerateServiceRuleKey(tenant, serviceId, "")
	resp, err := store.Store().Rule().Search(ctx,
		registry.WithStrKey(key),
		registry.WithPrefix())
	if err != nil {
		util.Logger().Errorf(err, "Get rule failed.%s", err.Error())
		return "", 0, err
	}
	if len(resp.Kvs) == 0 {
		return "", 0, nil
	}
	rule := &pb.ServiceRule{}
	err = json.Unmarshal(resp.Kvs[0].Value, rule)
	if err != nil {
		util.Logger().Errorf(err, "Unmarshal rule data failed.%s", err.Error())
	}
	return rule.RuleType, len(resp.Kvs), nil
}

func GetOneRule(ctx context.Context, tenant, serviceId, ruleId string) (*pb.ServiceRule, error) {
	resp, err := store.Store().Rule().Search(ctx,
		registry.WithStrKey(apt.GenerateServiceRuleKey(tenant, serviceId, ruleId)))
	if err != nil {
		util.Logger().Errorf(nil, "Get rule for service failed for %s.", err.Error())
		return nil, err
	}
	rule := &pb.ServiceRule{}
	if len(resp.Kvs) == 0 {
		util.Logger().Errorf(nil, "Get rule failed, ruleId is %s.", ruleId)
		return nil, nil
	}
	err = json.Unmarshal(resp.Kvs[0].Value, rule)
	if err != nil {
		util.Logger().Errorf(nil, "unmarshal resp failed for %s.", err.Error())
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

	hasWhite := false
	for _, rule := range rules {
		var value string
		if tagRegEx.MatchString(rule.Attribute) {
			key := tagRegEx.FindStringSubmatch(rule.Attribute)[1]
			value = serviceTags[key]
			if len(value) == 0 {
				util.Logger().Infof("can not find service %s tag '%s'", service.ServiceId, key)
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
				util.Logger().Infof("service %s match white list, rule.Pattern is %s, value is %s",
					service.ServiceId, rule.Pattern, value)
				return nil
			}
		case "BLACK":
			match, _ := regexp.MatchString(rule.Pattern, value)
			if match {
				util.Logger().Infof("service %s match black list, rule.Pattern is %s, value is %s",
					service.ServiceId, rule.Pattern, value)
				return MatchBlackListError("Found in black list")
			}
		}

	}
	if hasWhite {
		util.Logger().Infof("service %s do not match white list", service.ServiceId)
		return NotMatchWhiteListError("Not found in white list")
	}
	return nil
}

func getConsumerIdsWithFilter(ctx context.Context, tenant, providerId string, provider *pb.MicroService,
	filter func(ctx context.Context, consumerId string) (bool, error)) (allow []string, deny []string, err error) {
	consumerIds, err := dependency.GetConsumersInCache(tenant, providerId, provider)
	if err != nil {
		return nil, nil, err
	}
	return filterConsumerIds(ctx, consumerIds, filter)
}

func filterConsumerIds(ctx context.Context, consumerIds []string,
	filter func(ctx context.Context, consumerId string) (bool, error)) (allow []string, deny []string, err error) {
	l := len(consumerIds)
	if l == 0 {
		return nil, nil, nil
	}
	allowIdx, denyIdx := 0, l
	consumers := make([]string, l)
	for _, consumerId := range consumerIds {
		ok, err := filter(ctx, consumerId)
		if err != nil {
			return nil, nil, err
		}
		if ok {
			consumers[allowIdx] = consumerId
			allowIdx++
		} else {
			denyIdx--
			consumers[denyIdx] = consumerId
		}
	}
	return consumers[:allowIdx], consumers[denyIdx:], nil
}

func noFilter(_ context.Context, _ string) (bool, error) {
	return true, nil
}

func GetConsumerIds(ctx context.Context, tenant string, provider *pb.MicroService) (allow []string, deny []string, _ error) {
	if provider == nil || len(provider.ServiceId) == 0 {
		return nil, nil, fmt.Errorf("invalid provider")
	}

	//todo 删除服务，最后实例推送有误差
	providerRules, err := GetRulesUtil(ctx, tenant, provider.ServiceId)
	if err != nil {
		return nil, nil, err
	}
	if len(providerRules) == 0 {
		return getConsumerIdsWithFilter(ctx, tenant, provider.ServiceId, provider, noFilter)
	}

	rf := RuleFilter{
		Tenant:        tenant,
		Provider:      provider,
		ProviderRules: providerRules,
	}

	allow, deny, err = getConsumerIdsWithFilter(ctx, tenant, provider.ServiceId, provider, rf.Filter)
	if err != nil {
		return nil, nil, err
	}
	return allow, deny, nil
}

func GetProvidersInCache(tenant string, consumerId string, provider *pb.MicroService) ([]string, error) {
	return dependency.GetProvidersInCache(tenant, consumerId, provider)
}

func GetProviderIdsByConsumerId(ctx context.Context, tenant, consumerId string, server *pb.MicroService) (allow []string, deny []string, _ error) {
	providerIdsInCache, err := GetProvidersInCache(tenant, consumerId, server)
	if err != nil {
		return nil, nil, err
	}
	l := len(providerIdsInCache)
	rf := RuleFilter{
		Tenant: tenant,
	}
	allowIdx, denyIdx := 0, l
	providerIds := make([]string, l)
	for _, providerId := range providerIdsInCache {
		provider, err := ms.GetService(ctx, tenant, providerId)
		if provider == nil {
			continue
		}
		providerRules, err := GetRulesUtil(ctx, tenant, provider.ServiceId)
		if err != nil {
			return nil, nil, err
		}
		if len(providerRules) == 0 {
			providerIds[allowIdx] = providerId
			allowIdx++
			continue
		}
		rf.Provider = provider
		rf.ProviderRules = providerRules
		ok, err := rf.Filter(ctx, consumerId)
		if err != nil {
			return nil, nil, err
		}
		if ok {
			providerIds[allowIdx] = providerId
			allowIdx++
		} else {
			denyIdx--
			providerIds[denyIdx] = providerId
		}
	}
	return providerIds[:allowIdx], providerIds[denyIdx:], nil
}
