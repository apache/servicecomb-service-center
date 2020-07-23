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
	"context"
	"fmt"
	"github.com/apache/servicecomb-service-center/pkg/log"
	pb "github.com/apache/servicecomb-service-center/pkg/registry"
	"github.com/apache/servicecomb-service-center/pkg/util"
	apt "github.com/apache/servicecomb-service-center/server/core"
	"github.com/apache/servicecomb-service-center/server/core/backend"
	"github.com/apache/servicecomb-service-center/server/core/proto"
	"github.com/apache/servicecomb-service-center/server/plugin/registry"
	scerr "github.com/apache/servicecomb-service-center/server/scerror"
	"reflect"
	"regexp"
	"strings"
)

type RuleFilter struct {
	DomainProject string
	ProviderRules []*pb.ServiceRule
}

func (rf *RuleFilter) Filter(ctx context.Context, consumerID string) (bool, error) {
	copyCtx := util.SetContext(util.CloneContext(ctx), util.CtxCacheOnly, "1")
	consumer, err := GetService(copyCtx, rf.DomainProject, consumerID)
	if consumer == nil {
		return false, err
	}

	if len(rf.ProviderRules) == 0 {
		return true, nil
	}

	tags, err := GetTagsUtils(copyCtx, rf.DomainProject, consumerID)
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

func (rf *RuleFilter) FilterAll(ctx context.Context, consumerIDs []string) (allow []string, deny []string, err error) {
	l := len(consumerIDs)
	if l == 0 || len(rf.ProviderRules) == 0 {
		return consumerIDs, nil, nil
	}

	allowIdx, denyIdx := 0, l
	consumers := make([]string, l)
	for _, consumerID := range consumerIDs {
		ok, err := rf.Filter(ctx, consumerID)
		if err != nil {
			return nil, nil, err
		}
		if ok {
			consumers[allowIdx] = consumerID
			allowIdx++
		} else {
			denyIdx--
			consumers[denyIdx] = consumerID
		}
	}
	return consumers[:allowIdx], consumers[denyIdx:], nil
}

func GetRulesUtil(ctx context.Context, domainProject string, serviceID string) ([]*pb.ServiceRule, error) {
	key := util.StringJoin([]string{
		apt.GetServiceRuleRootKey(domainProject),
		serviceID,
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

func RuleExist(ctx context.Context, domainProject string, serviceID string, attr string, pattern string) bool {
	opts := append(FromContext(ctx),
		registry.WithStrKey(apt.GenerateRuleIndexKey(domainProject, serviceID, attr, pattern)),
		registry.WithCountOnly())
	resp, err := backend.Store().RuleIndex().Search(ctx, opts...)
	if err != nil || resp.Count == 0 {
		return false
	}
	return true
}

func GetServiceRuleType(ctx context.Context, domainProject string, serviceID string) (string, int, error) {
	key := apt.GenerateServiceRuleKey(domainProject, serviceID, "")
	opts := append(FromContext(ctx),
		registry.WithStrKey(key),
		registry.WithPrefix())
	resp, err := backend.Store().Rule().Search(ctx, opts...)
	if err != nil {
		log.Errorf(err, "get service[%s] rule failed", serviceID)
		return "", 0, err
	}
	if len(resp.Kvs) == 0 {
		return "", 0, nil
	}
	return resp.Kvs[0].Value.(*pb.ServiceRule).RuleType, len(resp.Kvs), nil
}

func GetOneRule(ctx context.Context, domainProject, serviceID, ruleID string) (*pb.ServiceRule, error) {
	opts := append(FromContext(ctx),
		registry.WithStrKey(apt.GenerateServiceRuleKey(domainProject, serviceID, ruleID)))
	resp, err := backend.Store().Rule().Search(ctx, opts...)
	if err != nil {
		log.Errorf(err, "get service rule[%s/%s]", serviceID, ruleID)
		return nil, err
	}
	if len(resp.Kvs) == 0 {
		log.Errorf(nil, "get service rule[%s/%s] failed", serviceID, ruleID)
		return nil, nil
	}
	return resp.Kvs[0].Value.(*pb.ServiceRule), nil
}

func AllowAcrossDimension(ctx context.Context, providerService *pb.MicroService, consumerService *pb.MicroService) error {
	if providerService.AppId != consumerService.AppId {
		if len(providerService.Properties) == 0 {
			return fmt.Errorf("not allow across app access")
		}

		if allowCrossApp, ok := providerService.Properties[proto.PROP_ALLOW_CROSS_APP]; !ok || strings.ToLower(allowCrossApp) != "true" {
			return fmt.Errorf("not allow across app access")
		}
	}

	if !apt.IsShared(proto.MicroServiceToKey(util.ParseTargetDomainProject(ctx), providerService)) &&
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
	consumerID := consumer.ServiceId
	for _, rule := range rulesOfProvider {
		value, err := parsePattern(v, rule, tagsOfConsumer, consumerID)
		if err != nil {
			return err
		}
		if len(value) == 0 {
			continue
		}

		match, _ := regexp.MatchString(rule.Pattern, value)
		if match {
			log.Infof("consumer[%s][%s/%s/%s/%s] match white list, rule.Pattern is %s, value is %s",
				consumerID, consumer.Environment, consumer.AppId, consumer.ServiceName, consumer.Version,
				rule.Pattern, value)
			return nil
		}
	}
	return scerr.NewError(scerr.ErrPermissionDeny, "Not found in white list")
}

func parsePattern(v reflect.Value, rule *pb.ServiceRule, tagsOfConsumer map[string]string, consumerID string) (string, *scerr.Error) {
	if strings.HasPrefix(rule.Attribute, "tag_") {
		key := rule.Attribute[4:]
		value := tagsOfConsumer[key]
		if len(value) == 0 {
			log.Infof("can not find service[%s] tag[%s]", consumerID, key)
		}
		return value, nil
	}
	key := v.FieldByName(rule.Attribute)
	if !key.IsValid() {
		log.Errorf(nil, "can not find service[%] field[%s], ruleID is %s",
			consumerID, rule.Attribute, rule.RuleId)
		return "", scerr.NewErrorf(scerr.ErrInternal, "Can not find field '%s'", rule.Attribute)
	}
	return key.String(), nil

}

func patternBlackList(rulesOfProvider []*pb.ServiceRule, tagsOfConsumer map[string]string, consumer *pb.MicroService) *scerr.Error {
	v := reflect.Indirect(reflect.ValueOf(consumer))
	consumerID := consumer.ServiceId
	for _, rule := range rulesOfProvider {
		var value string
		value, err := parsePattern(v, rule, tagsOfConsumer, consumerID)
		if err != nil {
			return err
		}
		if len(value) == 0 {
			continue
		}

		match, _ := regexp.MatchString(rule.Pattern, value)
		if match {
			log.Warnf("no permission to access, consumer[%s][%s/%s/%s/%s] match black list, rule.Pattern is %s, value is %s",
				consumerID, consumer.Environment, consumer.AppId, consumer.ServiceName, consumer.Version,
				rule.Pattern, value)
			return scerr.NewError(scerr.ErrPermissionDeny, "Found in black list")
		}
	}
	return nil
}

func Accessible(ctx context.Context, consumerID string, providerID string) *scerr.Error {
	if len(consumerID) == 0 {
		return nil
	}

	domainProject := util.ParseDomainProject(ctx)
	targetDomainProject := util.ParseTargetDomainProject(ctx)

	consumerService, err := GetService(ctx, domainProject, consumerID)
	if err != nil {
		return scerr.NewErrorf(scerr.ErrInternal, "An error occurred in query consumer(%s)", err.Error())
	}
	if consumerService == nil {
		return scerr.NewError(scerr.ErrServiceNotExists, "consumer serviceID is invalid")
	}

	// 跨应用权限
	providerService, err := GetService(ctx, targetDomainProject, providerID)
	if err != nil {
		return scerr.NewErrorf(scerr.ErrInternal, "An error occurred in query provider(%s)", err.Error())
	}
	if providerService == nil {
		return scerr.NewError(scerr.ErrServiceNotExists, "provider serviceID is invalid")
	}

	err = AllowAcrossDimension(ctx, providerService, consumerService)
	if err != nil {
		return scerr.NewError(scerr.ErrPermissionDeny, err.Error())
	}

	ctx = util.SetContext(util.CloneContext(ctx), util.CtxCacheOnly, "1")

	// 黑白名单
	rules, err := GetRulesUtil(ctx, targetDomainProject, providerID)
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
