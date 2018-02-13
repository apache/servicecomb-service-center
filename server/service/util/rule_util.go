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
	"encoding/json"
	"fmt"
	"github.com/apache/incubator-servicecomb-service-center/pkg/util"
	apt "github.com/apache/incubator-servicecomb-service-center/server/core"
	"github.com/apache/incubator-servicecomb-service-center/server/core/backend/store"
	pb "github.com/apache/incubator-servicecomb-service-center/server/core/proto"
	scerr "github.com/apache/incubator-servicecomb-service-center/server/error"
	"github.com/apache/incubator-servicecomb-service-center/server/infra/registry"
	"golang.org/x/net/context"
	"reflect"
	"regexp"
	"strings"
)

var tagRegEx *regexp.Regexp

func init() {
	tagRegEx, _ = regexp.Compile("tag_(.*)")
}

type RuleFilter struct {
	DomainProject string
	Provider      *pb.MicroService
	ProviderRules []*pb.ServiceRule
}

func (rf *RuleFilter) Filter(ctx context.Context, consumerId string) (bool, error) {
	copyCtx := util.SetContext(util.CloneContext(ctx), "cacheOnly", "1")
	consumer, err := GetService(copyCtx, rf.DomainProject, consumerId)
	if consumer == nil {
		return false, err
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

func GetRulesUtil(ctx context.Context, domainProject string, serviceId string) ([]*pb.ServiceRule, error) {
	key := util.StringJoin([]string{
		apt.GetServiceRuleRootKey(domainProject),
		serviceId,
		"",
	}, "/")

	opts := append(FromContext(ctx), registry.WithStrKey(key), registry.WithPrefix())
	resp, err := store.Store().Rule().Search(ctx, opts...)
	if err != nil {
		return nil, err
	}

	rules := []*pb.ServiceRule{}
	for _, kvs := range resp.Kvs {
		rule := &pb.ServiceRule{}
		err := json.Unmarshal(kvs.Value, rule)
		if err != nil {
			return nil, err
		}
		rules = append(rules, rule)
	}
	return rules, nil
}

func RuleExist(ctx context.Context, domainProject string, serviceId string, attr string, pattern string) bool {
	opts := append(FromContext(ctx),
		registry.WithStrKey(apt.GenerateRuleIndexKey(domainProject, serviceId, attr, pattern)),
		registry.WithCountOnly())
	resp, err := store.Store().RuleIndex().Search(ctx, opts...)
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
	resp, err := store.Store().Rule().Search(ctx, opts...)
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

func GetOneRule(ctx context.Context, domainProject, serviceId, ruleId string) (*pb.ServiceRule, error) {
	opts := append(FromContext(ctx),
		registry.WithStrKey(apt.GenerateServiceRuleKey(domainProject, serviceId, ruleId)))
	resp, err := store.Store().Rule().Search(ctx, opts...)
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

func MatchRules(rules []*pb.ServiceRule, service *pb.MicroService, serviceTags map[string]string) *scerr.Error {
	if service == nil {
		return scerr.NewError(scerr.ErrInvalidParams, "service is nil")
	}

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
				util.Logger().Errorf(nil, "can not find service %s field '%s', rule %s",
					service.ServiceId, rule.Attribute, rule.RuleId)
				return scerr.NewError(scerr.ErrInternal, fmt.Sprintf("Can not find field '%s'", rule.Attribute))
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
				return scerr.NewError(scerr.ErrPermissionDeny, "Found in black list")
			}
		}
	}
	if hasWhite {
		util.Logger().Infof("service %s do not match white list", service.ServiceId)
		return scerr.NewError(scerr.ErrPermissionDeny, "Not found in white list")
	}
	return nil
}

func Accessible(ctx context.Context, consumerId string, providerId string) *scerr.Error {
	domainProject := util.ParseDomainProject(ctx)
	targetDomainProject := util.ParseTargetDomainProject(ctx)

	consumerService, err := GetService(ctx, domainProject, consumerId)
	if err != nil {
		return scerr.NewError(scerr.ErrInternal, fmt.Sprintf("An error occurred in query consumer(%s)", err.Error()))
	}
	if consumerService == nil {
		return scerr.NewError(scerr.ErrServiceNotExists, "consumer serviceId is invalid")
	}

	// 跨应用权限
	providerService, err := GetService(ctx, targetDomainProject, providerId)
	if err != nil {
		return scerr.NewError(scerr.ErrInternal, fmt.Sprintf("An error occurred in query provider(%s)", err.Error()))
	}
	if providerService == nil {
		return scerr.NewError(scerr.ErrServiceNotExists, "provider serviceId is invalid")
	}

	err = AllowAcrossDimension(ctx, providerService, consumerService)
	if err != nil {
		return scerr.NewError(scerr.ErrPermissionDeny, err.Error())
	}

	ctx = util.SetContext(util.CloneContext(ctx), "cacheOnly", "1")

	// 黑白名单
	rules, err := GetRulesUtil(ctx, targetDomainProject, providerId)
	if err != nil {
		return scerr.NewError(scerr.ErrInternal, fmt.Sprintf("An error occurred in query provider rules(%s)", err.Error()))
	}

	if len(rules) == 0 {
		return nil
	}

	validateTags, err := GetTagsUtils(ctx, domainProject, consumerService.ServiceId)
	if err != nil {
		return scerr.NewError(scerr.ErrInternal, fmt.Sprintf("An error occurred in query consumer tags(%s)", err.Error()))
	}

	return MatchRules(rules, consumerService, validateTags)
}
