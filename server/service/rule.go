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
	"encoding/json"
	"fmt"
	apt "github.com/ServiceComb/service-center/server/core"
	pb "github.com/ServiceComb/service-center/server/core/proto"
	"github.com/ServiceComb/service-center/server/core/registry"
	"github.com/ServiceComb/service-center/server/infra/quota"
	"github.com/ServiceComb/service-center/server/plugins/dynamic"
	serviceUtil "github.com/ServiceComb/service-center/server/service/util"
	"github.com/ServiceComb/service-center/util"
	errorsEx "github.com/ServiceComb/service-center/util/errors"
	"golang.org/x/net/context"
	"strconv"
	"time"
)

func Accessible(ctx context.Context, tenant string, consumerId string, providerId string) error {
	consumerService, err := serviceUtil.GetService(ctx, tenant, consumerId)
	if err != nil {
		util.Logger().Errorf(err,
			"consumer %s can't access provider %s for internal error", consumerId, providerId)
		return errorsEx.InternalError(err.Error())
	}
	if consumerService == nil {
		util.Logger().Warnf(nil,
			"consumer %s can't access provider %s for invalid consumer", consumerId, providerId)
		return fmt.Errorf("consumer invalid")
	}

	consumerFlag := fmt.Sprintf("%s/%s/%s", consumerService.AppId, consumerService.ServiceName, consumerService.Version)

	// 跨应用权限
	providerService, err := serviceUtil.GetService(ctx, tenant, providerId)
	if err != nil {
		util.Logger().Errorf(err, "consumer %s can't access provider %s for internal error",
			consumerFlag, providerId)
		return errorsEx.InternalError(err.Error())
	}
	if providerService == nil {
		util.Logger().Warnf(nil, "consumer %s can't access provider %s for invalid provider",
			consumerFlag, providerId)
		return fmt.Errorf("provider invalid")
	}

	providerFlag := fmt.Sprintf("%s/%s/%s", providerService.AppId, providerService.ServiceName, providerService.Version)

	err = serviceUtil.AllowAcrossApp(providerService, consumerService)
	if err != nil {
		util.Logger().Warnf(nil,
			"consumer %s can't access provider %s which property 'allowCrossApp' is not true or does not exist",
			consumerFlag, providerFlag)
		return err
	}

	// 黑白名单
	rules, err := serviceUtil.GetRulesUtil(ctx, tenant, providerId, registry.WithCacheOnly())
	if err != nil {
		util.Logger().Errorf(err, "consumer %s can't access provider %s for internal error",
			consumerFlag, providerFlag)
		return errorsEx.InternalError(err.Error())
	}

	if len(rules) == 0 {
		return nil
	}

	validateTags, err := serviceUtil.GetTagsUtils(ctx, tenant, consumerService.ServiceId, registry.WithCacheOnly())
	if err != nil {
		util.Logger().Errorf(err, "consumer %s can't access provider %s for internal error",
			consumerFlag, providerFlag)
		return errorsEx.InternalError(err.Error())
	}

	err = serviceUtil.MatchRules(rules, consumerService, validateTags)
	if err != nil {
		switch err.(type) {
		case errorsEx.InternalError:
			util.Logger().Errorf(err, "consumer %s can't access provider %s for internal error",
				consumerFlag, providerFlag)
		default:
			util.Logger().Warnf(err, "consumer %s can't access provider %s", consumerFlag, providerFlag)
		}
		return err
	}

	return nil
}

func (s *ServiceController) AddRule(ctx context.Context, in *pb.AddServiceRulesRequest) (*pb.AddServiceRulesResponse, error) {
	if in == nil || len(in.ServiceId) == 0 || len(in.GetRules()) == 0 {
		util.Logger().Errorf(nil, "add rule failed: invalid parameters.")
		return &pb.AddServiceRulesResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Request format invalid."),
		}, nil
	}

	tenant := util.ParseTenantProject(ctx)

	// service id存在性校验
	if !serviceUtil.ServiceExist(ctx, tenant, in.ServiceId) {
		util.Logger().Errorf(nil, "add rule failed, serviceId is %s: service not exist.", in.ServiceId)
		return &pb.AddServiceRulesResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Service does not exist."),
		}, nil
	}
	ok, err := quota.QuotaPlugins[quota.QuataType]().Apply4Quotas(ctx, quota.RULEQuotaType, tenant, in.ServiceId, 1)
	if err != nil {
		util.Logger().Errorf(err, "check can apply resource failed.%s", in.ServiceId)
		return &pb.AddServiceRulesResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, err.Error()),
		}, err
	}
	if !ok {
		util.Logger().Errorf(err, "no size to add tag, max size is 100 for one servivce.%s", in.ServiceId)
		return &pb.AddServiceRulesResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "no size to add tag, max size is 100 for one servivce"),
		}, nil
	}

	opts := []registry.PluginOp{}
	ruleType, _, err := serviceUtil.GetServiceRuleType(ctx, tenant, in.ServiceId)
	util.Logger().Debugf("ruleType is %s", ruleType)
	if err != nil {
		return &pb.AddServiceRulesResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, err.Error()),
		}, nil
	}
	ruleIds := []string{}
	for _, rule := range in.Rules {
		err := apt.Validate(rule)
		if err != nil {
			util.Logger().Errorf(err, "add rule failed, serviceId is %s: invalid rule.", in.ServiceId)
			return &pb.AddServiceRulesResponse{
				Response: pb.CreateResponse(pb.Response_FAIL, err.Error()),
			}, nil
		}
		//黑白名单只能存在一种，黑名单 or 白名单
		if len(ruleType) == 0 {
			ruleType = rule.RuleType
		} else {
			if ruleType != rule.RuleType {
				util.Logger().Errorf(nil, "add rule failed, serviceId is %s:can only exist one type, BLACK or WHITE.", in.ServiceId)
				return &pb.AddServiceRulesResponse{
					Response: pb.CreateResponse(pb.Response_FAIL, "Service can only contain one rule type, BLACK or WHITE."),
				}, nil
			}
		}

		//同一服务，attribute和pattern确定一个rule
		if serviceUtil.RuleExist(ctx, tenant, in.ServiceId, rule.Attribute, rule.Pattern) {
			util.Logger().Infof("This rule more exists, %s ", in.ServiceId)
			continue
		}

		// 产生全局rule id
		timestamp := strconv.FormatInt(time.Now().Unix(), 10)
		ruleAdd := &pb.ServiceRule{
			RuleId:       dynamic.GenerateUuid(),
			RuleType:     rule.RuleType,
			Attribute:    rule.Attribute,
			Pattern:      rule.Pattern,
			Description:  rule.Description,
			Timestamp:    timestamp,
			ModTimestamp: timestamp,
		}

		key := apt.GenerateServiceRuleKey(tenant, in.ServiceId, ruleAdd.RuleId)
		indexKey := apt.GenerateRuleIndexKey(tenant, in.ServiceId, ruleAdd.Attribute, ruleAdd.Pattern)
		ruleIds = append(ruleIds, ruleAdd.RuleId)

		util.Logger().Debugf("indexKey is : %s", indexKey)
		util.Logger().Debugf("start add service rule file: %s", key)
		data, err := json.Marshal(ruleAdd)
		if err != nil {
			util.Logger().Errorf(err, "add rule failed, serviceId is %s: marshal rule failed.", in.ServiceId)
			return &pb.AddServiceRulesResponse{
				Response: pb.CreateResponse(pb.Response_FAIL, "Service rule file marshal error."),
			}, err
		}

		opts = append(opts, registry.OpPut(registry.WithStrKey(key), registry.WithValue(data)))
		opts = append(opts, registry.OpPut(registry.WithStrKey(indexKey), registry.WithStrValue(ruleAdd.RuleId)))
	}
	if len(opts) <= 0 {
		util.Logger().Infof("add rule successful, serviceId is %s: rule more exists,no rules to add.", in.ServiceId)
		return &pb.AddServiceRulesResponse{
			Response: pb.CreateResponse(pb.Response_SUCCESS, "Service rules has been added."),
		}, nil
	}
	_, err = registry.GetRegisterCenter().Txn(ctx, opts)
	if err != nil {
		util.Logger().Errorf(err, "add rule failed, serviceId is %s:commit date into etcd failed.", in.ServiceId)
		return &pb.AddServiceRulesResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Commit operations failed."),
		}, err
	}

	util.Logger().Infof("add rule successful, serviceId  %s.", in.ServiceId)
	return &pb.AddServiceRulesResponse{
		Response: pb.CreateResponse(pb.Response_SUCCESS, "Add service rules successfully."),
		RuleIds:  ruleIds,
	}, nil
}

func (s *ServiceController) UpdateRule(ctx context.Context, in *pb.UpdateServiceRuleRequest) (*pb.UpdateServiceRuleResponse, error) {
	if in == nil || in.GetRule() == nil || len(in.ServiceId) == 0 || len(in.RuleId) == 0 {
		util.Logger().Errorf(nil, "update rule failed: invalid parameters.")
		return &pb.UpdateServiceRuleResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Request format invalid."),
		}, nil
	}

	tenant := util.ParseTenantProject(ctx)

	// service id存在性校验
	if !serviceUtil.ServiceExist(ctx, tenant, in.ServiceId) {
		util.Logger().Errorf(nil, "update rule failed, serviceId is %s, ruleId is %s: service not exist.", in.ServiceId, in.RuleId)
		return &pb.UpdateServiceRuleResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Service does not exist."),
		}, nil
	}
	err := apt.Validate(in.Rule)
	if err != nil {
		util.Logger().Errorf(err, "update rule failed, serviceId is %s, ruleId is %s: invalid service rule.", in.ServiceId, in.RuleId)
		return &pb.UpdateServiceRuleResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, err.Error()),
		}, nil
	}

	//是否能改变ruleType
	ruleType, ruleNum, err := serviceUtil.GetServiceRuleType(ctx, tenant, in.ServiceId)
	if err != nil {
		util.Logger().Errorf(err, "update rule failed, serviceId is %s, ruleId is %s: get rule type failed.", in.ServiceId, in.RuleId)
		return &pb.UpdateServiceRuleResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, err.Error()),
		}, err
	}
	if ruleNum >= 1 && ruleType != in.Rule.RuleType {
		util.Logger().Errorf(err, "update rule failed, serviceId is %s, ruleId is %s: rule type can exist one type, BLACK or WHITE.rule type is %s", in.ServiceId, in.RuleId, in.Rule.RuleType)
		return &pb.UpdateServiceRuleResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Exist multiple rules,can not change rule type. Rule type is "+ruleType),
		}, nil
	}

	rule, err := serviceUtil.GetOneRule(ctx, tenant, in.ServiceId, in.RuleId)
	if err != nil {
		util.Logger().Errorf(err, "update rule failed, serviceId is %s, ruleId is %s: query service rule failed.", in.ServiceId, in.RuleId)
		return &pb.UpdateServiceRuleResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Get service rule file failed."),
		}, err
	}
	if rule == nil {
		util.Logger().Errorf(err, "update rule failed, serviceId is %s, ruleId is %s:this rule does not exist,can't update.", in.ServiceId, in.RuleId)
		return &pb.UpdateServiceRuleResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "This rule does not exist."),
		}, nil
	}

	oldRulePatten := rule.Pattern
	oldRuleAttr := rule.Attribute
	isChangeIndex := false
	if rule.Attribute != in.GetRule().Attribute {
		isChangeIndex = true
		rule.Attribute = in.GetRule().Attribute
	}
	if rule.Pattern != in.GetRule().Pattern {
		isChangeIndex = true
		rule.Pattern = in.GetRule().Pattern
	}
	rule.RuleType = in.GetRule().RuleType
	rule.Description = in.GetRule().Description
	rule.ModTimestamp = strconv.FormatInt(time.Now().Unix(), 10)

	key := apt.GenerateServiceRuleKey(tenant, in.ServiceId, in.RuleId)
	util.Logger().Debugf("start update service rule file: %s", key)
	data, err := json.Marshal(rule)
	if err != nil {
		util.Logger().Errorf(err, "update rule failed, serviceId is %s, ruleId is %s: marshal service rule failed.", in.ServiceId, in.RuleId)
		return &pb.UpdateServiceRuleResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Service rule file marshal error."),
		}, err
	}
	opts := []registry.PluginOp{}
	if isChangeIndex {
		//加入新的rule index
		indexKey := apt.GenerateRuleIndexKey(tenant, in.ServiceId, rule.Attribute, rule.Pattern)
		opts = append(opts, registry.OpPut(registry.WithStrKey(indexKey), registry.WithStrValue(rule.RuleId)))

		//删除旧的rule index
		oldIndexKey := apt.GenerateRuleIndexKey(tenant, in.ServiceId, oldRuleAttr, oldRulePatten)
		opts = append(opts, registry.OpDel(registry.WithStrKey(oldIndexKey)))
	}
	opts = append(opts, registry.OpPut(registry.WithStrKey(key), registry.WithValue(data)))
	_, err = registry.GetRegisterCenter().Txn(ctx, opts)
	if err != nil {
		util.Logger().Errorf(err, "update rule failed, serviceId is %s, ruleId is %s: commit date into etcd failed.", in.ServiceId, in.RuleId)
		return &pb.UpdateServiceRuleResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Commit operations failed."),
		}, err
	}

	util.Logger().Infof("update rule successful: servieId is %s, ruleId is %s.", in.ServiceId, in.RuleId)
	return &pb.UpdateServiceRuleResponse{
		Response: pb.CreateResponse(pb.Response_SUCCESS, "Get service rules successfully."),
	}, nil
}

func (s *ServiceController) GetRule(ctx context.Context, in *pb.GetServiceRulesRequest) (*pb.GetServiceRulesResponse, error) {
	if in == nil || len(in.ServiceId) == 0 {
		util.Logger().Errorf(nil, "get service rule failed, serviceId is %s: invalid params.", in.ServiceId)
		return &pb.GetServiceRulesResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Request format invalid."),
		}, nil
	}

	tenant := util.ParseTenantProject(ctx)
	// service id存在性校验
	if !serviceUtil.ServiceExist(ctx, tenant, in.ServiceId) {
		util.Logger().Errorf(nil, "get service rule failed, serviceId is %s: service not exist.", in.ServiceId)
		return &pb.GetServiceRulesResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Service does not exist."),
		}, nil
	}

	rules, err := serviceUtil.GetRulesUtil(ctx, tenant, in.ServiceId)
	if err != nil {
		util.Logger().Errorf(nil, "get service rule failed, serviceId is %s: get rule failed.", in.ServiceId)
		return &pb.GetServiceRulesResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Get service rules failed."),
		}, nil
	}

	return &pb.GetServiceRulesResponse{
		Response: pb.CreateResponse(pb.Response_SUCCESS, "Get service rules successfully."),
		Rules:    rules,
	}, nil
}

func (s *ServiceController) DeleteRule(ctx context.Context, in *pb.DeleteServiceRulesRequest) (*pb.DeleteServiceRulesResponse, error) {
	if in == nil || len(in.ServiceId) == 0 {
		util.Logger().Errorf(nil, "delete service rule failed: invalid parameters.")
		return &pb.DeleteServiceRulesResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Request format invalid."),
		}, nil
	}

	tenant := util.ParseTenantProject(ctx)
	// service id存在性校验
	if !serviceUtil.ServiceExist(ctx, tenant, in.ServiceId) {
		util.Logger().Errorf(nil, "delete service rule failed, serviceId is %s, rule is %v: service not exist.", in.ServiceId, in.RuleIds)
		return &pb.DeleteServiceRulesResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Service does not exist."),
		}, nil
	}

	opts := []registry.PluginOp{}
	key := ""
	indexKey := ""
	for _, ruleId := range in.RuleIds {
		key = apt.GenerateServiceRuleKey(tenant, in.ServiceId, ruleId)
		util.Logger().Debugf("start delete service rule file: %s", key)
		data, err := serviceUtil.GetOneRule(ctx, tenant, in.ServiceId, ruleId)
		if err != nil {
			util.Logger().Errorf(err, "delete service rule failed, serviceId is %s, rule is %v: get rule of ruleId %s failed.", in.ServiceId, in.RuleIds, ruleId)
			return &pb.DeleteServiceRulesResponse{
				Response: pb.CreateResponse(pb.Response_FAIL, err.Error()),
			}, err
		}
		if data == nil {
			util.Logger().Errorf(nil, "delete service rule failed, serviceId is %s, rule is %v: ruleId %s not exist.", in.ServiceId, in.RuleIds, ruleId)
			return &pb.DeleteServiceRulesResponse{
				Response: pb.CreateResponse(pb.Response_FAIL, "This rule does not exist."),
			}, nil
		}
		indexKey = apt.GenerateRuleIndexKey(tenant, in.ServiceId, data.Attribute, data.Pattern)
		opts = append(opts,
			registry.OpDel(registry.WithStrKey(key)),
			registry.OpDel(registry.WithStrKey(indexKey)))
	}
	if len(opts) <= 0 {
		util.Logger().Errorf(nil, "delete service rule failed, serviceId is %s, rule is %v: rule has been deleted.", in.ServiceId, in.RuleIds)
		return &pb.DeleteServiceRulesResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "No service rule has been deleted."),
		}, nil
	}
	_, err := registry.GetRegisterCenter().Txn(ctx, opts)
	if err != nil {
		util.Logger().Errorf(err, "delete service rule failed, serviceId is %s, rule is %v: commit data into etcd failed.", in.ServiceId, in.RuleIds)
		return &pb.DeleteServiceRulesResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Commit operations failed."),
		}, err
	}

	util.Logger().Infof("delete rule successful: serviceId %s, ruleIds %v", in.ServiceId, in.RuleIds)
	return &pb.DeleteServiceRulesResponse{
		Response: pb.CreateResponse(pb.Response_SUCCESS, "Delete service rules successfully."),
	}, nil
}
