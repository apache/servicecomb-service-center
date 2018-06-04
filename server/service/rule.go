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
package service

import (
	"encoding/json"
	"github.com/apache/incubator-servicecomb-service-center/pkg/util"
	apt "github.com/apache/incubator-servicecomb-service-center/server/core"
	"github.com/apache/incubator-servicecomb-service-center/server/core/backend"
	pb "github.com/apache/incubator-servicecomb-service-center/server/core/proto"
	scerr "github.com/apache/incubator-servicecomb-service-center/server/error"
	"github.com/apache/incubator-servicecomb-service-center/server/infra/quota"
	"github.com/apache/incubator-servicecomb-service-center/server/infra/registry"
	"github.com/apache/incubator-servicecomb-service-center/server/plugin"
	serviceUtil "github.com/apache/incubator-servicecomb-service-center/server/service/util"
	"golang.org/x/net/context"
	"strconv"
	"time"
)

func (s *MicroServiceService) AddRule(ctx context.Context, in *pb.AddServiceRulesRequest) (*pb.AddServiceRulesResponse, error) {
	err := Validate(in)
	if err != nil {
		util.Logger().Errorf(err, "add rule failed, serviceId is %s.", in.ServiceId)
		return &pb.AddServiceRulesResponse{
			Response: pb.CreateResponse(scerr.ErrInvalidParams, err.Error()),
		}, nil
	}

	domainProject := util.ParseDomainProject(ctx)

	// service id存在性校验
	if !serviceUtil.ServiceExist(ctx, domainProject, in.ServiceId) {
		util.Logger().Errorf(nil, "add rule failed, serviceId is %s: service not exist.", in.ServiceId)
		return &pb.AddServiceRulesResponse{
			Response: pb.CreateResponse(scerr.ErrInvalidParams, "Service does not exist."),
		}, nil
	}
	res := quota.NewApplyQuotaResource(quota.RuleQuotaType, domainProject, in.ServiceId, int64(len(in.Rules)))
	rst := plugin.Plugins().Quota().Apply4Quotas(ctx, res)
	errQuota := rst.Err
	if errQuota != nil {
		util.Logger().Errorf(errQuota, "")
		response := &pb.AddServiceRulesResponse{
			Response: pb.CreateResponseWithSCErr(errQuota),
		}
		if errQuota.InternalError() {
			return response, errQuota
		}
		return response, nil
	}

	ruleType, _, err := serviceUtil.GetServiceRuleType(ctx, domainProject, in.ServiceId)
	if err != nil {
		return &pb.AddServiceRulesResponse{
			Response: pb.CreateResponse(scerr.ErrInternal, err.Error()),
		}, err
	}
	ruleIds := make([]string, 0, len(in.Rules))
	opts := make([]registry.PluginOp, 0, 2*len(in.Rules))
	for _, rule := range in.Rules {
		//黑白名单只能存在一种，黑名单 or 白名单
		if len(ruleType) == 0 {
			ruleType = rule.RuleType
		} else {
			if ruleType != rule.RuleType {
				util.Logger().Errorf(nil, "add rule failed, serviceId is %s:can only exist one type, BLACK or WHITE.", in.ServiceId)
				return &pb.AddServiceRulesResponse{
					Response: pb.CreateResponse(scerr.ErrBlackAndWhiteRule, "Service can only contain one rule type, BLACK or WHITE."),
				}, nil
			}
		}

		//同一服务，attribute和pattern确定一个rule
		if serviceUtil.RuleExist(ctx, domainProject, in.ServiceId, rule.Attribute, rule.Pattern) {
			util.Logger().Infof("This rule more exists, %s ", in.ServiceId)
			continue
		}

		// 产生全局rule id
		timestamp := strconv.FormatInt(time.Now().Unix(), 10)
		ruleAdd := &pb.ServiceRule{
			RuleId:       util.GenerateUuid(),
			RuleType:     rule.RuleType,
			Attribute:    rule.Attribute,
			Pattern:      rule.Pattern,
			Description:  rule.Description,
			Timestamp:    timestamp,
			ModTimestamp: timestamp,
		}

		key := apt.GenerateServiceRuleKey(domainProject, in.ServiceId, ruleAdd.RuleId)
		indexKey := apt.GenerateRuleIndexKey(domainProject, in.ServiceId, ruleAdd.Attribute, ruleAdd.Pattern)
		ruleIds = append(ruleIds, ruleAdd.RuleId)

		util.Logger().Debugf("indexKey is : %s", indexKey)
		util.Logger().Debugf("start add service rule file: %s", key)
		data, err := json.Marshal(ruleAdd)
		if err != nil {
			util.Logger().Errorf(err, "add rule failed, serviceId is %s: marshal rule failed.", in.ServiceId)
			return &pb.AddServiceRulesResponse{
				Response: pb.CreateResponse(scerr.ErrInternal, "Service rule file marshal error."),
			}, err
		}

		opts = append(opts, registry.OpPut(registry.WithStrKey(key), registry.WithValue(data)))
		opts = append(opts, registry.OpPut(registry.WithStrKey(indexKey), registry.WithStrValue(ruleAdd.RuleId)))
	}
	if len(opts) <= 0 {
		util.Logger().Infof("add rule successful, serviceId is %s: rule already exists, no rules to add.", in.ServiceId)
		return &pb.AddServiceRulesResponse{
			Response: pb.CreateResponse(pb.Response_SUCCESS, "Service rules has been added."),
		}, nil
	}
	err = backend.BatchCommit(ctx, opts)
	if err != nil {
		util.Logger().Errorf(err, "add rule failed, serviceId is %s:commit date into etcd failed.", in.ServiceId)
		return &pb.AddServiceRulesResponse{
			Response: pb.CreateResponse(scerr.ErrUnavailableBackend, "Commit operations failed."),
		}, err
	}

	util.Logger().Infof("add rule successful, serviceId  %s.", in.ServiceId)
	return &pb.AddServiceRulesResponse{
		Response: pb.CreateResponse(pb.Response_SUCCESS, "Add service rules successfully."),
		RuleIds:  ruleIds,
	}, nil
}

func (s *MicroServiceService) UpdateRule(ctx context.Context, in *pb.UpdateServiceRuleRequest) (*pb.UpdateServiceRuleResponse, error) {
	err := Validate(in)
	if err != nil {
		util.Logger().Errorf(err, "update rule failed, serviceId is %s, ruleId is %s.", in.ServiceId, in.RuleId)
		return &pb.UpdateServiceRuleResponse{
			Response: pb.CreateResponse(scerr.ErrInvalidParams, err.Error()),
		}, nil
	}

	domainProject := util.ParseDomainProject(ctx)

	// service id存在性校验
	if !serviceUtil.ServiceExist(ctx, domainProject, in.ServiceId) {
		util.Logger().Errorf(nil, "update rule failed, serviceId is %s, ruleId is %s: service not exist.", in.ServiceId, in.RuleId)
		return &pb.UpdateServiceRuleResponse{
			Response: pb.CreateResponse(scerr.ErrServiceNotExists, "Service does not exist."),
		}, nil
	}

	//是否能改变ruleType
	ruleType, ruleNum, err := serviceUtil.GetServiceRuleType(ctx, domainProject, in.ServiceId)
	if err != nil {
		util.Logger().Errorf(err, "update rule failed, serviceId is %s, ruleId is %s: get rule type failed.", in.ServiceId, in.RuleId)
		return &pb.UpdateServiceRuleResponse{
			Response: pb.CreateResponse(scerr.ErrInternal, err.Error()),
		}, err
	}
	if ruleNum >= 1 && ruleType != in.Rule.RuleType {
		util.Logger().Errorf(err, "update rule failed, serviceId is %s, ruleId is %s: rule type can exist one type, BLACK or WHITE.rule type is %s", in.ServiceId, in.RuleId, in.Rule.RuleType)
		return &pb.UpdateServiceRuleResponse{
			Response: pb.CreateResponse(scerr.ErrModifyRuleNotAllow, "Exist multiple rules,can not change rule type. Rule type is "+ruleType),
		}, nil
	}

	rule, err := serviceUtil.GetOneRule(ctx, domainProject, in.ServiceId, in.RuleId)
	if err != nil {
		util.Logger().Errorf(err, "update rule failed, serviceId is %s, ruleId is %s: query service rule failed.", in.ServiceId, in.RuleId)
		return &pb.UpdateServiceRuleResponse{
			Response: pb.CreateResponse(scerr.ErrInternal, "Get service rule file failed."),
		}, err
	}
	if rule == nil {
		util.Logger().Errorf(err, "update rule failed, serviceId is %s, ruleId is %s:this rule does not exist,can't update.", in.ServiceId, in.RuleId)
		return &pb.UpdateServiceRuleResponse{
			Response: pb.CreateResponse(scerr.ErrRuleNotExists, "This rule does not exist."),
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

	key := apt.GenerateServiceRuleKey(domainProject, in.ServiceId, in.RuleId)
	util.Logger().Debugf("start update service rule file: %s", key)
	data, err := json.Marshal(rule)
	if err != nil {
		util.Logger().Errorf(err, "update rule failed, serviceId is %s, ruleId is %s: marshal service rule failed.", in.ServiceId, in.RuleId)
		return &pb.UpdateServiceRuleResponse{
			Response: pb.CreateResponse(scerr.ErrInternal, "Service rule file marshal error."),
		}, err
	}
	opts := []registry.PluginOp{}
	if isChangeIndex {
		//加入新的rule index
		indexKey := apt.GenerateRuleIndexKey(domainProject, in.ServiceId, rule.Attribute, rule.Pattern)
		opts = append(opts, registry.OpPut(registry.WithStrKey(indexKey), registry.WithStrValue(rule.RuleId)))

		//删除旧的rule index
		oldIndexKey := apt.GenerateRuleIndexKey(domainProject, in.ServiceId, oldRuleAttr, oldRulePatten)
		opts = append(opts, registry.OpDel(registry.WithStrKey(oldIndexKey)))
	}
	opts = append(opts, registry.OpPut(registry.WithStrKey(key), registry.WithValue(data)))
	_, err = backend.Registry().Txn(ctx, opts)
	if err != nil {
		util.Logger().Errorf(err, "update rule failed, serviceId is %s, ruleId is %s: commit date into etcd failed.", in.ServiceId, in.RuleId)
		return &pb.UpdateServiceRuleResponse{
			Response: pb.CreateResponse(scerr.ErrUnavailableBackend, "Commit operations failed."),
		}, err
	}

	util.Logger().Infof("update rule successful: servieId is %s, ruleId is %s.", in.ServiceId, in.RuleId)
	return &pb.UpdateServiceRuleResponse{
		Response: pb.CreateResponse(pb.Response_SUCCESS, "Get service rules successfully."),
	}, nil
}

func (s *MicroServiceService) GetRule(ctx context.Context, in *pb.GetServiceRulesRequest) (*pb.GetServiceRulesResponse, error) {
	err := Validate(in)
	if err != nil {
		util.Logger().Errorf(err, "get service rule failed, serviceId %s.", in.ServiceId)
		return &pb.GetServiceRulesResponse{
			Response: pb.CreateResponse(scerr.ErrInvalidParams, err.Error()),
		}, nil
	}

	domainProject := util.ParseDomainProject(ctx)

	// service id存在性校验
	if !serviceUtil.ServiceExist(ctx, domainProject, in.ServiceId) {
		util.Logger().Errorf(nil, "get service rule failed, serviceId is %s: service not exist.", in.ServiceId)
		return &pb.GetServiceRulesResponse{
			Response: pb.CreateResponse(scerr.ErrServiceNotExists, "Service does not exist."),
		}, nil
	}

	rules, err := serviceUtil.GetRulesUtil(ctx, domainProject, in.ServiceId)
	if err != nil {
		util.Logger().Errorf(err, "get service rule failed, serviceId is %s: get rule failed.", in.ServiceId)
		return &pb.GetServiceRulesResponse{
			Response: pb.CreateResponse(scerr.ErrInternal, "Get service rules failed."),
		}, err
	}

	return &pb.GetServiceRulesResponse{
		Response: pb.CreateResponse(pb.Response_SUCCESS, "Get service rules successfully."),
		Rules:    rules,
	}, nil
}

func (s *MicroServiceService) DeleteRule(ctx context.Context, in *pb.DeleteServiceRulesRequest) (*pb.DeleteServiceRulesResponse, error) {
	err := Validate(in)
	if err != nil {
		util.Logger().Errorf(err, "delete rule failed, serviceId is %s, ruleIds are %s.", in.ServiceId, in.RuleIds)
		return &pb.DeleteServiceRulesResponse{
			Response: pb.CreateResponse(scerr.ErrInvalidParams, err.Error()),
		}, nil
	}

	domainProject := util.ParseDomainProject(ctx)

	// service id存在性校验
	if !serviceUtil.ServiceExist(ctx, domainProject, in.ServiceId) {
		util.Logger().Errorf(nil, "delete service rule failed, serviceId is %s, rule is %v: service not exist.", in.ServiceId, in.RuleIds)
		return &pb.DeleteServiceRulesResponse{
			Response: pb.CreateResponse(scerr.ErrServiceNotExists, "Service does not exist."),
		}, nil
	}

	opts := []registry.PluginOp{}
	key := ""
	indexKey := ""
	for _, ruleId := range in.RuleIds {
		key = apt.GenerateServiceRuleKey(domainProject, in.ServiceId, ruleId)
		util.Logger().Debugf("start delete service rule file: %s", key)
		data, err := serviceUtil.GetOneRule(ctx, domainProject, in.ServiceId, ruleId)
		if err != nil {
			util.Logger().Errorf(err, "delete service rule failed, serviceId is %s, rule is %v: get rule of ruleId %s failed.", in.ServiceId, in.RuleIds, ruleId)
			return &pb.DeleteServiceRulesResponse{
				Response: pb.CreateResponse(scerr.ErrInternal, err.Error()),
			}, err
		}
		if data == nil {
			util.Logger().Errorf(nil, "delete service rule failed, serviceId is %s, rule is %v: ruleId %s not exist.", in.ServiceId, in.RuleIds, ruleId)
			return &pb.DeleteServiceRulesResponse{
				Response: pb.CreateResponse(scerr.ErrRuleNotExists, "This rule does not exist."),
			}, nil
		}
		indexKey = apt.GenerateRuleIndexKey(domainProject, in.ServiceId, data.Attribute, data.Pattern)
		opts = append(opts,
			registry.OpDel(registry.WithStrKey(key)),
			registry.OpDel(registry.WithStrKey(indexKey)))
	}
	if len(opts) <= 0 {
		util.Logger().Errorf(nil, "delete service rule failed, serviceId is %s, rule is %v: rule has been deleted.", in.ServiceId, in.RuleIds)
		return &pb.DeleteServiceRulesResponse{
			Response: pb.CreateResponse(scerr.ErrRuleNotExists, "No service rule has been deleted."),
		}, nil
	}

	err = backend.BatchCommit(ctx, opts)
	if err != nil {
		util.Logger().Errorf(err, "delete service rule failed, serviceId is %s, rule is %v: commit data into etcd failed.", in.ServiceId, in.RuleIds)
		return &pb.DeleteServiceRulesResponse{
			Response: pb.CreateResponse(scerr.ErrUnavailableBackend, "Commit operations failed."),
		}, err
	}

	util.Logger().Infof("delete rule successful: serviceId %s, ruleIds %v", in.ServiceId, in.RuleIds)
	return &pb.DeleteServiceRulesResponse{
		Response: pb.CreateResponse(pb.Response_SUCCESS, "Delete service rules successfully."),
	}, nil
}
