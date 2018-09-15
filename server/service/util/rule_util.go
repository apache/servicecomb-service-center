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
package util

import (
	"fmt"
	"github.com/apache/incubator-servicecomb-service-center/pkg/log"
	"github.com/apache/incubator-servicecomb-service-center/pkg/util"
	apt "github.com/apache/incubator-servicecomb-service-center/server/core"
	"github.com/apache/incubator-servicecomb-service-center/server/core/backend"
	pb "github.com/apache/incubator-servicecomb-service-center/server/core/proto"
	scerr "github.com/apache/incubator-servicecomb-service-center/server/error"
	"github.com/apache/incubator-servicecomb-service-center/server/infra/registry"
	"golang.org/x/net/context"
	"reflect"
	"regexp"
	"strings"
)

type RuleFilter struct {
	DomainProject string
	ProviderRules []*pb.ServiceRule
}

func (rf *RuleFilter) Filter(ctx context.Context, consumerId string) (bool, error) {
	copyCtx := util.SetContext(util.CloneContext(ctx), CTX_CACHEONLY, "1")
	consumer, err := GetService(copyCtx, rf.DomainProject, consumerId)
	if consumer == nil {
		return false, err
	}

	if len(rf.ProviderRules) == 0 {
		return true, nil
	}

	tags, err := GetTagsUtils(copyCtx, rf.DomainProject, consumerId)
	if err != nil {
		return false, err
	}
	matchErr := MatchRules(rf.ProviderRules, consumer, tags)
	if matchErr != nil {
		if matchErr.Code == scerr.ErrPermissionDeny {
			return false, nil
		}
		return false, matchErr
	}
	return true, nil
}

func (rf *RuleFilter) FilterAll(ctx context.Context, consumerIds []string) (allow []string, deny []string, err error) {
	l := len(consumerIds)
	if l == 0 || len(rf.ProviderRules) == 0 {
		return consumerIds, nil, nil
	}

	allowIdx, denyIdx := 0, l
	consumers := make([]string, l)
	for _, consumerId := range consumerIds {
		ok, err := rf.Filter(ctx, consumerId)
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

func GetRulesUtil(ctx context.Context, domainProject string, serviceId string) ([]*pb.ServiceRule, error) {
	key := util.StringJoin([]string{
		apt.GetServiceRuleRootKey(domainProject),
		serviceId,
		"",
	}, "/")

	opts := append(FromContext(ctx), registry.WithStrKey(key), registry.WithPrefix())
	resp, err := backend.Store().Rule().Search(ctx, opts...)
	if err != nil {
		return nil, err
	}

	rules := []*pb.ServiceRule{}
	for _, kv := range resp.Kvs {
		rules = append(rules, kv.Value.(*pb.ServiceRule))
	}
	return rules, nil
}

func RuleExist(ctx context.Context, domainProject string, serviceId string, attr string, pattern string) bool {
	opts := append(FromContext(ctx),
		registry.WithStrKey(apt.GenerateRuleIndexKey(domainProject, serviceId, attr, pattern)),
		registry.WithCountOnly())
	resp, err := backend.Store().RuleIndex().Search(ctx, opts...)
	if err != nil || resp.Count == 0 {
		return false
	}
	return true
}

func GetServiceRuleType(ctx context.Context, domainProject string, serviceId string) (string, int, error) {
	key := apt.GenerateServiceRuleKey(domainProject, serviceId, "")
	opts := append(FromContext(ctx),
		registry.WithStrKey(key),
		registry.WithPrefix())
	resp, err := backend.Store().Rule().Search(ctx, opts...)
	if err != nil {
		log.Errorf(err, "Get rule failed.%s", err.Error())
		return "", 0, err
	}
	if len(resp.Kvs) == 0 {
		return "", 0, nil
	}
	return resp.Kvs[0].Value.(*pb.ServiceRule).RuleType, len(resp.Kvs), nil
}

func GetOneRule(ctx context.Context, domainProject, serviceId, ruleId string) (*pb.ServiceRule, error) {
	opts := append(FromContext(ctx),
		registry.WithStrKey(apt.GenerateServiceRuleKey(domainProject, serviceId, ruleId)))
	resp, err := backend.Store().Rule().Search(ctx, opts...)
	if err != nil {
		log.Errorf(nil, "Get rule for service failed for %s.", err.Error())
		return nil, err
	}
	if len(resp.Kvs) == 0 {
		log.Errorf(nil, "Get rule failed, ruleId is %s.", ruleId)
		return nil, nil
	}
	return resp.Kvs[0].Value.(*pb.ServiceRule), nil
}

func AllowAcrossDimension(ctx context.Context, providerService *pb.MicroService, consumerService *pb.MicroService) error {
	if providerService.AppId != consumerService.AppId {
		if len(providerService.Properties) == 0 {
			return fmt.Errorf("not allow across app access")
		}

		if allowCrossApp, ok := providerService.Properties[pb.PROP_ALLOW_CROSS_APP]; !ok || strings.ToLower(allowCrossApp) != "true" {
			return fmt.Errorf("not allow across app access")
		}
	}

	if !apt.IsShared(pb.MicroServiceToKey(util.ParseTargetDomainProject(ctx), providerService)) &&
		providerService.Environment != consumerService.Environment {
		return fmt.Errorf("not allow across environment access")
	}

	return nil
}

func MatchRules(rulesOfProvider []*pb.ServiceRule, consumer *pb.MicroService, tagsOfConsumer map[string]string) *scerr.Error {
	if consumer == nil {
		return scerr.NewError(scerr.ErrInvalidParams, "consumer is nil")
	}

	if len(rulesOfProvider) <= 0 {
		return nil
	}
	if rulesOfProvider[0].RuleType == "WHITE" {
		return patternWhiteList(rulesOfProvider, tagsOfConsumer, consumer)
	}
	return patternBlackList(rulesOfProvider, tagsOfConsumer, consumer)
}

func patternWhiteList(rulesOfProvider []*pb.ServiceRule, tagsOfConsumer map[string]string, consumer *pb.MicroService) *scerr.Error {
	v := reflect.Indirect(reflect.ValueOf(consumer))
	consumerId := consumer.ServiceId
	for _, rule := range rulesOfProvider {
		value, err := parsePattern(v, rule, tagsOfConsumer, consumerId)
		if err != nil {
			return err
		}
		if len(value) == 0 {
			continue
		}

		match, _ := regexp.MatchString(rule.Pattern, value)
		if match {
			log.Infof("consumer %s match white list, rule.Pattern is %s, value is %s",
				consumerId, rule.Pattern, value)
			return nil
		}
	}
	return scerr.NewError(scerr.ErrPermissionDeny, "Not found in white list")
}

func parsePattern(v reflect.Value, rule *pb.ServiceRule, tagsOfConsumer map[string]string, consumerId string) (string, *scerr.Error) {
	if strings.HasPrefix(rule.Attribute, "tag_") {
		key := rule.Attribute[4:]
		value := tagsOfConsumer[key]
		if len(value) == 0 {
			log.Infof("can not find service %s tag '%s'", consumerId, key)
		}
		return value, nil
	}
	key := v.FieldByName(rule.Attribute)
	if !key.IsValid() {
		log.Errorf(nil, "can not find service %s field '%s', rule %s",
			consumerId, rule.Attribute, rule.RuleId)
		return "", scerr.NewErrorf(scerr.ErrInternal, "Can not find field '%s'", rule.Attribute)
	}
	return key.String(), nil

}

func patternBlackList(rulesOfProvider []*pb.ServiceRule, tagsOfConsumer map[string]string, consumer *pb.MicroService) *scerr.Error {
	v := reflect.Indirect(reflect.ValueOf(consumer))
	consumerId := consumer.ServiceId
	for _, rule := range rulesOfProvider {
		var value string
		value, err := parsePattern(v, rule, tagsOfConsumer, consumerId)
		if err != nil {
			return err
		}
		if len(value) == 0 {
			continue
		}

		match, _ := regexp.MatchString(rule.Pattern, value)
		if match {
			log.Warnf("no permission to access, consumer %s match black list, rule.Pattern is %s, value is %s",
				consumerId, rule.Pattern, value)
			return scerr.NewError(scerr.ErrPermissionDeny, "Found in black list")
		}
	}
	return nil
}

func Accessible(ctx context.Context, consumerId string, providerId string) *scerr.Error {
	if len(consumerId) == 0 {
		return nil
	}

	domainProject := util.ParseDomainProject(ctx)
	targetDomainProject := util.ParseTargetDomainProject(ctx)

	consumerService, err := GetService(ctx, domainProject, consumerId)
	if err != nil {
		return scerr.NewErrorf(scerr.ErrInternal, "An error occurred in query consumer(%s)", err.Error())
	}
	if consumerService == nil {
		return scerr.NewError(scerr.ErrServiceNotExists, "consumer serviceId is invalid")
	}

	// 跨应用权限
	providerService, err := GetService(ctx, targetDomainProject, providerId)
	if err != nil {
		return scerr.NewErrorf(scerr.ErrInternal, "An error occurred in query provider(%s)", err.Error())
	}
	if providerService == nil {
		return scerr.NewError(scerr.ErrServiceNotExists, "provider serviceId is invalid")
	}

	err = AllowAcrossDimension(ctx, providerService, consumerService)
	if err != nil {
		return scerr.NewError(scerr.ErrPermissionDeny, err.Error())
	}

	ctx = util.SetContext(util.CloneContext(ctx), CTX_CACHEONLY, "1")

	// 黑白名单
	rules, err := GetRulesUtil(ctx, targetDomainProject, providerId)
	if err != nil {
		return scerr.NewErrorf(scerr.ErrInternal, "An error occurred in query provider rules(%s)", err.Error())
	}

	if len(rules) == 0 {
		return nil
	}

	validateTags, err := GetTagsUtils(ctx, domainProject, consumerService.ServiceId)
	if err != nil {
		return scerr.NewErrorf(scerr.ErrInternal, "An error occurred in query consumer tags(%s)", err.Error())
	}

	return MatchRules(rules, consumerService, validateTags)
}
