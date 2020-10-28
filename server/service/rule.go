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
	"context"
	"encoding/json"
	"github.com/apache/servicecomb-service-center/datasource/etcd/client"
	serviceUtil "github.com/apache/servicecomb-service-center/datasource/etcd/util"
	"github.com/apache/servicecomb-service-center/pkg/log"
	pb "github.com/apache/servicecomb-service-center/pkg/registry"
	"github.com/apache/servicecomb-service-center/pkg/util"
	apt "github.com/apache/servicecomb-service-center/server/core"
	"github.com/apache/servicecomb-service-center/server/core/proto"
	"github.com/apache/servicecomb-service-center/server/plugin/quota"
	scerr "github.com/apache/servicecomb-service-center/server/scerror"
	"strconv"
	"time"
)

func (s *MicroServiceService) AddRule(ctx context.Context, in *pb.AddServiceRulesRequest) (*pb.AddServiceRulesResponse, error) {
	remoteIP := util.GetIPFromContext(ctx)
	err := Validate(in)
	if err != nil {
		log.Errorf(err, "add service[%s] rule failed, operator: %s", in.ServiceId, remoteIP)
		return &pb.AddServiceRulesResponse{
			Response: proto.CreateResponse(scerr.ErrInvalidParams, err.Error()),
		}, nil
	}

	domainProject := util.ParseDomainProject(ctx)

	// service id存在性校验
	if !serviceUtil.ServiceExist(ctx, domainProject, in.ServiceId) {
		log.Errorf(nil, "add service[%s] rule failed, service does not exist, operator: %s",
			in.ServiceId, remoteIP)
		return &pb.AddServiceRulesResponse{
			Response: proto.CreateResponse(scerr.ErrInvalidParams, "Service does not exist."),
		}, nil
	}
	res := quota.NewApplyQuotaResource(quota.RuleQuotaType, domainProject, in.ServiceId, int64(len(in.Rules)))
	rst := quota.Apply(ctx, res)
	errQuota := rst.Err
	if errQuota != nil {
		log.Errorf(errQuota, "add service[%s] rule failed, operator: %s", in.ServiceId, remoteIP)
		response := &pb.AddServiceRulesResponse{
			Response: proto.CreateResponseWithSCErr(errQuota),
		}
		if errQuota.InternalError() {
			return response, errQuota
		}
		return response, nil
	}

	ruleType, _, err := serviceUtil.GetServiceRuleType(ctx, domainProject, in.ServiceId)
	if err != nil {
		return &pb.AddServiceRulesResponse{
			Response: proto.CreateResponse(scerr.ErrInternal, err.Error()),
		}, err
	}
	ruleIDs := make([]string, 0, len(in.Rules))
	opts := make([]client.PluginOp, 0, 2*len(in.Rules))
	for _, rule := range in.Rules {
		//黑白名单只能存在一种，黑名单 or 白名单
		if len(ruleType) == 0 {
			ruleType = rule.RuleType
		} else if ruleType != rule.RuleType {
			log.Errorf(nil, "add service[%s] rule failed, can not add different RuleType at the same time, operator: %s",
				in.ServiceId, remoteIP)
			return &pb.AddServiceRulesResponse{
				Response: proto.CreateResponse(scerr.ErrBlackAndWhiteRule, "Service can only contain one rule type, BLACK or WHITE."),
			}, nil

		}

		//同一服务，attribute和pattern确定一个rule
		if serviceUtil.RuleExist(ctx, domainProject, in.ServiceId, rule.Attribute, rule.Pattern) {
			log.Infof("service[%s] rule[%s/%s] already exists, operator: %s",
				in.ServiceId, rule.Attribute, rule.Pattern, remoteIP)
			continue
		}

		// 产生全局rule id
		timestamp := strconv.FormatInt(time.Now().Unix(), 10)
		ruleAdd := &pb.ServiceRule{
			RuleId:       util.GenerateUUID(),
			RuleType:     rule.RuleType,
			Attribute:    rule.Attribute,
			Pattern:      rule.Pattern,
			Description:  rule.Description,
			Timestamp:    timestamp,
			ModTimestamp: timestamp,
		}

		key := apt.GenerateServiceRuleKey(domainProject, in.ServiceId, ruleAdd.RuleId)
		indexKey := apt.GenerateRuleIndexKey(domainProject, in.ServiceId, ruleAdd.Attribute, ruleAdd.Pattern)
		ruleIDs = append(ruleIDs, ruleAdd.RuleId)

		data, err := json.Marshal(ruleAdd)
		if err != nil {
			log.Errorf(err, "add service[%s] rule failed, marshal rule[%s/%s] failed, operator: %s",
				in.ServiceId, ruleAdd.Attribute, ruleAdd.Pattern, remoteIP)
			return &pb.AddServiceRulesResponse{
				Response: proto.CreateResponse(scerr.ErrInternal, err.Error()),
			}, err
		}

		opts = append(opts, client.OpPut(client.WithStrKey(key), client.WithValue(data)))
		opts = append(opts, client.OpPut(client.WithStrKey(indexKey), client.WithStrValue(ruleAdd.RuleId)))
	}
	if len(opts) <= 0 {
		log.Infof("add service[%s] rule successfully, no rules to add, operator: %s",
			in.ServiceId, remoteIP)
		return &pb.AddServiceRulesResponse{
			Response: proto.CreateResponse(proto.ResponseSuccess, "Service rules has been added."),
		}, nil
	}

	resp, err := client.BatchCommitWithCmp(ctx, opts,
		[]client.CompareOp{client.OpCmp(
			client.CmpVer(util.StringToBytesWithNoCopy(apt.GenerateServiceKey(domainProject, in.ServiceId))),
			client.CmpNotEqual, 0)},
		nil)
	if err != nil {
		log.Errorf(err, "add service[%s] rule failed, operator: %s", in.ServiceId, remoteIP)
		return &pb.AddServiceRulesResponse{
			Response: proto.CreateResponse(scerr.ErrUnavailableBackend, err.Error()),
		}, err
	}
	if !resp.Succeeded {
		log.Errorf(nil, "add service[%s] rule failed, service does not exist, operator: %s",
			in.ServiceId, remoteIP)
		return &pb.AddServiceRulesResponse{
			Response: proto.CreateResponse(scerr.ErrServiceNotExists, "Service does not exist."),
		}, nil
	}

	log.Infof("add service[%s] rule %v successfully, operator: %s", in.ServiceId, ruleIDs, remoteIP)
	return &pb.AddServiceRulesResponse{
		Response: proto.CreateResponse(proto.ResponseSuccess, "Add service rules successfully."),
		RuleIds:  ruleIDs,
	}, nil
}

func (s *MicroServiceService) UpdateRule(ctx context.Context, in *pb.UpdateServiceRuleRequest) (*pb.UpdateServiceRuleResponse, error) {
	remoteIP := util.GetIPFromContext(ctx)
	err := Validate(in)
	if err != nil {
		log.Errorf(err, "update service rule[%s/%s] failed, operator: %s", in.ServiceId, in.RuleId, remoteIP)
		return &pb.UpdateServiceRuleResponse{
			Response: proto.CreateResponse(scerr.ErrInvalidParams, err.Error()),
		}, nil
	}

	domainProject := util.ParseDomainProject(ctx)

	// service id存在性校验
	if !serviceUtil.ServiceExist(ctx, domainProject, in.ServiceId) {
		log.Errorf(nil, "update service rule[%s/%s] failed, service does not exist, operator: %s",
			in.ServiceId, in.RuleId, remoteIP)
		return &pb.UpdateServiceRuleResponse{
			Response: proto.CreateResponse(scerr.ErrServiceNotExists, "Service does not exist."),
		}, nil
	}

	//是否能改变ruleType
	ruleType, ruleNum, err := serviceUtil.GetServiceRuleType(ctx, domainProject, in.ServiceId)
	if err != nil {
		log.Errorf(err, "update service rule[%s/%s] failed, get rule type failed, operator: %s",
			in.ServiceId, in.RuleId, remoteIP)
		return &pb.UpdateServiceRuleResponse{
			Response: proto.CreateResponse(scerr.ErrInternal, err.Error()),
		}, err
	}
	if ruleNum >= 1 && ruleType != in.Rule.RuleType {
		log.Errorf(err, "update service rule[%s/%s] failed, can only exist one type, current type is %s, operator: %s",
			in.ServiceId, in.RuleId, ruleType, remoteIP)
		return &pb.UpdateServiceRuleResponse{
			Response: proto.CreateResponse(scerr.ErrModifyRuleNotAllow, "Exist multiple rules,can not change rule type. Rule type is "+ruleType),
		}, nil
	}

	rule, err := serviceUtil.GetOneRule(ctx, domainProject, in.ServiceId, in.RuleId)
	if err != nil {
		log.Errorf(err, "update service rule[%s/%s] failed, query service rule failed, operator: %s",
			in.ServiceId, in.RuleId, remoteIP)
		return &pb.UpdateServiceRuleResponse{
			Response: proto.CreateResponse(scerr.ErrInternal, err.Error()),
		}, err
	}
	if rule == nil {
		log.Errorf(err, "update service rule[%s/%s] failed, service rule does not exist, operator: %s",
			in.ServiceId, in.RuleId, remoteIP)
		return &pb.UpdateServiceRuleResponse{
			Response: proto.CreateResponse(scerr.ErrRuleNotExists, "This rule does not exist."),
		}, nil
	}

	copyRuleRef := *rule
	oldRulePatten := copyRuleRef.Pattern
	oldRuleAttr := copyRuleRef.Attribute
	isChangeIndex := false
	if copyRuleRef.Attribute != in.Rule.Attribute {
		isChangeIndex = true
		copyRuleRef.Attribute = in.Rule.Attribute
	}
	if copyRuleRef.Pattern != in.Rule.Pattern {
		isChangeIndex = true
		copyRuleRef.Pattern = in.Rule.Pattern
	}
	copyRuleRef.RuleType = in.Rule.RuleType
	copyRuleRef.Description = in.Rule.Description
	copyRuleRef.ModTimestamp = strconv.FormatInt(time.Now().Unix(), 10)

	key := apt.GenerateServiceRuleKey(domainProject, in.ServiceId, in.RuleId)
	data, err := json.Marshal(copyRuleRef)
	if err != nil {
		log.Errorf(err, "update service rule[%s/%s] failed, marshal service rule failed, operator: %s",
			in.ServiceId, in.RuleId, remoteIP)
		return &pb.UpdateServiceRuleResponse{
			Response: proto.CreateResponse(scerr.ErrInternal, err.Error()),
		}, err
	}
	var opts []client.PluginOp
	if isChangeIndex {
		//加入新的rule index
		indexKey := apt.GenerateRuleIndexKey(domainProject, in.ServiceId, copyRuleRef.Attribute, copyRuleRef.Pattern)
		opts = append(opts, client.OpPut(client.WithStrKey(indexKey), client.WithStrValue(copyRuleRef.RuleId)))

		//删除旧的rule index
		oldIndexKey := apt.GenerateRuleIndexKey(domainProject, in.ServiceId, oldRuleAttr, oldRulePatten)
		opts = append(opts, client.OpDel(client.WithStrKey(oldIndexKey)))
	}
	opts = append(opts, client.OpPut(client.WithStrKey(key), client.WithValue(data)))

	resp, err := client.Instance().TxnWithCmp(ctx, opts,
		[]client.CompareOp{client.OpCmp(
			client.CmpVer(util.StringToBytesWithNoCopy(apt.GenerateServiceKey(domainProject, in.ServiceId))),
			client.CmpNotEqual, 0)},
		nil)
	if err != nil {
		log.Errorf(err, "update service rule[%s/%s] failed, operator: %s", in.ServiceId, in.RuleId, remoteIP)
		return &pb.UpdateServiceRuleResponse{
			Response: proto.CreateResponse(scerr.ErrUnavailableBackend, err.Error()),
		}, err
	}
	if !resp.Succeeded {
		log.Errorf(err, "update service rule[%s/%s] failed, service does not exist, operator: %s",
			in.ServiceId, in.RuleId, remoteIP)
		return &pb.UpdateServiceRuleResponse{
			Response: proto.CreateResponse(scerr.ErrServiceNotExists, "Service does not exist."),
		}, nil
	}

	log.Infof("update service rule[%s/%s] successfully, operator: %s", in.ServiceId, in.RuleId, remoteIP)
	return &pb.UpdateServiceRuleResponse{
		Response: proto.CreateResponse(proto.ResponseSuccess, "Get service rules successfully."),
	}, nil
}

func (s *MicroServiceService) GetRule(ctx context.Context, in *pb.GetServiceRulesRequest) (*pb.GetServiceRulesResponse, error) {
	err := Validate(in)
	if err != nil {
		log.Errorf(err, "get service[%s] rule failed", in.ServiceId)
		return &pb.GetServiceRulesResponse{
			Response: proto.CreateResponse(scerr.ErrInvalidParams, err.Error()),
		}, nil
	}

	domainProject := util.ParseDomainProject(ctx)

	// service id存在性校验
	if !serviceUtil.ServiceExist(ctx, domainProject, in.ServiceId) {
		log.Errorf(nil, "get service[%s] rule failed, service does not exist", in.ServiceId)
		return &pb.GetServiceRulesResponse{
			Response: proto.CreateResponse(scerr.ErrServiceNotExists, "Service does not exist."),
		}, nil
	}

	rules, err := serviceUtil.GetRulesUtil(ctx, domainProject, in.ServiceId)
	if err != nil {
		log.Errorf(err, "get service[%s] rule failed", in.ServiceId)
		return &pb.GetServiceRulesResponse{
			Response: proto.CreateResponse(scerr.ErrInternal, err.Error()),
		}, err
	}

	return &pb.GetServiceRulesResponse{
		Response: proto.CreateResponse(proto.ResponseSuccess, "Get service rules successfully."),
		Rules:    rules,
	}, nil
}

func (s *MicroServiceService) DeleteRule(ctx context.Context, in *pb.DeleteServiceRulesRequest) (*pb.DeleteServiceRulesResponse, error) {
	remoteIP := util.GetIPFromContext(ctx)
	err := Validate(in)
	if err != nil {
		log.Errorf(err, "delete service[%s] rules %v failed, operator: %s", in.ServiceId, in.RuleIds, remoteIP)
		return &pb.DeleteServiceRulesResponse{
			Response: proto.CreateResponse(scerr.ErrInvalidParams, err.Error()),
		}, nil
	}

	domainProject := util.ParseDomainProject(ctx)

	// service id存在性校验
	if !serviceUtil.ServiceExist(ctx, domainProject, in.ServiceId) {
		log.Errorf(nil, "delete service[%s] rules %v failed, service does not exist, operator: %s",
			in.ServiceId, in.RuleIds, remoteIP)
		return &pb.DeleteServiceRulesResponse{
			Response: proto.CreateResponse(scerr.ErrServiceNotExists, "Service does not exist."),
		}, nil
	}

	opts := []client.PluginOp{}
	key := ""
	indexKey := ""
	for _, ruleID := range in.RuleIds {
		key = apt.GenerateServiceRuleKey(domainProject, in.ServiceId, ruleID)
		log.Debugf("start delete service rule file: %s", key)
		data, err := serviceUtil.GetOneRule(ctx, domainProject, in.ServiceId, ruleID)
		if err != nil {
			log.Errorf(err, "delete service[%s] rules %v failed, get rule[%s] failed, operator: %s",
				in.ServiceId, in.RuleIds, ruleID, remoteIP)
			return &pb.DeleteServiceRulesResponse{
				Response: proto.CreateResponse(scerr.ErrInternal, err.Error()),
			}, err
		}
		if data == nil {
			log.Errorf(nil, "delete service[%s] rules %v failed, rule[%s] does not exist, operator: %s",
				in.ServiceId, in.RuleIds, ruleID, remoteIP)
			return &pb.DeleteServiceRulesResponse{
				Response: proto.CreateResponse(scerr.ErrRuleNotExists, "This rule does not exist."),
			}, nil
		}
		indexKey = apt.GenerateRuleIndexKey(domainProject, in.ServiceId, data.Attribute, data.Pattern)
		opts = append(opts,
			client.OpDel(client.WithStrKey(key)),
			client.OpDel(client.WithStrKey(indexKey)))
	}
	if len(opts) <= 0 {
		log.Errorf(nil, "delete service[%s] rules %v failed, no rule has been deleted, operator: %s",
			in.ServiceId, in.RuleIds, remoteIP)
		return &pb.DeleteServiceRulesResponse{
			Response: proto.CreateResponse(scerr.ErrRuleNotExists, "No service rule has been deleted."),
		}, nil
	}

	resp, err := client.BatchCommitWithCmp(ctx, opts,
		[]client.CompareOp{client.OpCmp(
			client.CmpVer(util.StringToBytesWithNoCopy(apt.GenerateServiceKey(domainProject, in.ServiceId))),
			client.CmpNotEqual, 0)},
		nil)
	if err != nil {
		log.Errorf(err, "delete service[%s] rules %v failed, operator: %s", in.ServiceId, in.RuleIds, remoteIP)
		return &pb.DeleteServiceRulesResponse{
			Response: proto.CreateResponse(scerr.ErrUnavailableBackend, err.Error()),
		}, err
	}
	if !resp.Succeeded {
		log.Errorf(err, "delete service[%s] rules %v failed, service does not exist, operator: %s",
			in.ServiceId, in.RuleIds, remoteIP)
		return &pb.DeleteServiceRulesResponse{
			Response: proto.CreateResponse(scerr.ErrServiceNotExists, "Service does not exist."),
		}, nil
	}

	log.Infof("delete service[%s] rules %v successfully, operator: %s", in.ServiceId, in.RuleIds, remoteIP)
	return &pb.DeleteServiceRulesResponse{
		Response: proto.CreateResponse(proto.ResponseSuccess, "Delete service rules successfully."),
	}, nil
}
