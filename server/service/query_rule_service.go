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
	"fmt"
	pb "github.com/ServiceComb/service-center/server/core/proto"
	ms "github.com/ServiceComb/service-center/server/service/microservice"
	serviceUtil "github.com/ServiceComb/service-center/server/service/util"
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

type NotMatchTagError string

func (e NotMatchTagError) Error() string {
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
				return NotMatchTagError(
					fmt.Sprintf("Can not find service tag '%s'", key))
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
				util.LOGGER.Infof("match white list, rule.Pattern is %s, value is %s", rule.Pattern, value)
				return nil
			}
		case "BLACK":
			match, _ := regexp.MatchString(rule.Pattern, value)
			if match {
				util.LOGGER.Infof("match black list, rule.Pattern is %s, value is %s", rule.Pattern, value)
				return MatchBlackListError("Found in black list")
			}
		}

	}
	if hasWhite {
		return NotMatchWhiteListError("Not found in white list")
	}
	return nil
}

func Accessible(ctx context.Context, tenant string, consumerId string, providerId string) error {
	consumerService, err := ms.GetServiceByServiceId(ctx, tenant, consumerId)
	if err != nil {
		util.LOGGER.Errorf(err,
			"consumer %s can't access provider %s for internal error", consumerId, providerId)
		return errorsEx.InternalError(err.Error())
	}
	if consumerService == nil {
		util.LOGGER.Warnf(nil,
			"consumer %s can't access provider %s for invalid consumer", consumerId, providerId)
		return fmt.Errorf("consumer invalid")
	}

	consumerFlag := fmt.Sprintf("%s/%s/%s", consumerService.AppId, consumerService.ServiceName, consumerService.Version)

	// 跨应用权限
	providerService, err := ms.GetServiceByServiceId(ctx, tenant, providerId)
	if err != nil {
		util.LOGGER.Errorf(err, "consumer %s can't access provider %s for internal error",
			consumerFlag, providerId)
		return errorsEx.InternalError(err.Error())
	}
	if providerService == nil {
		util.LOGGER.Warnf(nil, "consumer %s can't access provider %s for invalid provider",
			consumerFlag, providerId)
		return fmt.Errorf("provider invalid")
	}

	providerFlag := fmt.Sprintf("%s/%s/%s", providerService.AppId, providerService.ServiceName, providerService.Version)

	err = AllowAcrossApp(providerService, consumerService)
	if err != nil {
		util.LOGGER.Warnf(nil,
			"consumer %s can't access provider %s which property 'allowCrossApp' is not true or does not exist",
			consumerFlag, providerFlag)
		return err
	}

	// 黑白名单
	rules, err := serviceUtil.GetRulesUtil(ctx, tenant, providerId)
	if err != nil {
		util.LOGGER.Errorf(err, "consumer %s can't access provider %s for internal error",
			consumerFlag, providerFlag)
		return errorsEx.InternalError(err.Error())
	}

	if len(rules) == 0 {
		return nil
	}

	validateTags, err := serviceUtil.GetTagsUtils(ctx, tenant, consumerService.ServiceId)
	if err != nil {
		util.LOGGER.Errorf(err, "consumer %s can't access provider %s for internal error",
			consumerFlag, providerFlag)
		return errorsEx.InternalError(err.Error())
	}

	err = MatchRules(rules, consumerService, validateTags)
	if err != nil {
		switch err.(type) {
		case errorsEx.InternalError:
			util.LOGGER.Errorf(err, "consumer %s can't access provider %s for internal error",
				consumerFlag, providerFlag)
		default:
			util.LOGGER.Warnf(err, "consumer %s can't access provider %s", consumerFlag, providerFlag)
		}
		return err
	}

	return nil
}
