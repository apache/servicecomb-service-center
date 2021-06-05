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
	"fmt"
	"strconv"
	"time"

	"github.com/go-chassis/cari/discovery"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/apache/servicecomb-service-center/datasource/mongo/model"
	mutil "github.com/apache/servicecomb-service-center/datasource/mongo/util"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/util"
	"github.com/apache/servicecomb-service-center/server/plugin/quota"
)

func (ds *DataSource) AddRule(ctx context.Context, request *discovery.AddServiceRulesRequest) (*discovery.AddServiceRulesResponse, error) {
	remoteIP := util.GetIPFromContext(ctx)
	domain := util.ParseDomain(ctx)
	project := util.ParseProject(ctx)
	filter := mutil.NewDomainProjectFilter(domain, project, mutil.ServiceServiceID(request.ServiceId))
	exist, err := exitService(ctx, filter)
	if err != nil {
		log.Error(fmt.Sprintf("failed to add rules for service %s for get service failed", request.ServiceId), err)
		return &discovery.AddServiceRulesResponse{
			Response: discovery.CreateResponse(discovery.ErrInternal, "failed to check service exist"),
		}, nil
	}
	if !exist {
		return &discovery.AddServiceRulesResponse{Response: discovery.CreateResponse(discovery.ErrServiceNotExists, "service does not exist")}, nil
	}
	res := quota.NewApplyQuotaResource(quota.TypeRule, util.ParseDomainProject(ctx), request.ServiceId, int64(len(request.Rules)))
	errQuota := quota.Apply(ctx, res)
	if errQuota != nil {
		log.Error(fmt.Sprintf("add service[%s] rule failed, operator: %s", request.ServiceId, remoteIP), errQuota)
		response := &discovery.AddServiceRulesResponse{
			Response: discovery.CreateResponseWithSCErr(errQuota),
		}
		if errQuota.InternalError() {
			return response, errQuota
		}
		return response, nil
	}
	filter = mutil.NewDomainProjectFilter(domain, project, mutil.ServiceID(request.ServiceId))
	rules, err := getServiceRules(ctx, filter)
	if err != nil {
		return &discovery.AddServiceRulesResponse{
			Response: discovery.CreateResponse(discovery.ErrInternal, err.Error()),
		}, err
	}
	var ruleType string
	if len(rules) != 0 {
		ruleType = rules[0].RuleType
	}
	ruleIDs := make([]string, 0, len(request.Rules))
	for _, rule := range request.Rules {
		if len(ruleType) == 0 {
			ruleType = rule.RuleType
		} else if ruleType != rule.RuleType {
			return &discovery.AddServiceRulesResponse{
				Response: discovery.CreateResponse(discovery.ErrBlackAndWhiteRule, "service can only contain one rule type,Black or white."),
			}, nil
		}
		//the rule unique index is (serviceid,attribute,pattern)
		filter = mutil.NewFilter(
			mutil.ServiceID(request.ServiceId),
			mutil.RuleAttribute(rule.Attribute),
			mutil.RulePattern(rule.Pattern),
		)
		exist, err := existRule(ctx, filter)
		if err != nil {
			return &discovery.AddServiceRulesResponse{
				Response: discovery.CreateResponse(discovery.ErrUnavailableBackend, "can not check rule if exist."),
			}, nil
		}
		if exist {
			continue
		}
		timestamp := strconv.FormatInt(time.Now().Unix(), baseTen)
		ruleAdd := &model.Rule{
			Domain:    util.ParseDomain(ctx),
			Project:   util.ParseProject(ctx),
			ServiceID: request.ServiceId,
			Rule: &discovery.ServiceRule{
				RuleId:       util.GenerateUUID(),
				RuleType:     rule.RuleType,
				Attribute:    rule.Attribute,
				Pattern:      rule.Pattern,
				Description:  rule.Description,
				Timestamp:    timestamp,
				ModTimestamp: timestamp,
			},
		}
		ruleIDs = append(ruleIDs, ruleAdd.Rule.RuleId)
		err = insertRule(ctx, ruleAdd)
		if err != nil {
			return &discovery.AddServiceRulesResponse{
				Response: discovery.CreateResponse(discovery.ErrInternal, err.Error()),
			}, err
		}
	}
	return &discovery.AddServiceRulesResponse{
		Response: discovery.CreateResponse(discovery.ResponseSuccess, "Add service rules successfully."),
		RuleIds:  ruleIDs,
	}, nil
}

func (ds *DataSource) GetRules(ctx context.Context, request *discovery.GetServiceRulesRequest) (*discovery.GetServiceRulesResponse, error) {
	domain := util.ParseDomain(ctx)
	project := util.ParseProject(ctx)
	filter := mutil.NewDomainProjectFilter(domain, project, mutil.ServiceServiceID(request.ServiceId))
	exist, err := exitService(ctx, filter)
	if err != nil {
		return &discovery.GetServiceRulesResponse{
			Response: discovery.CreateResponse(discovery.ErrServiceNotExists, "get rules failed for get service failed."),
		}, nil
	}
	if !exist {
		return &discovery.GetServiceRulesResponse{
			Response: discovery.CreateResponse(discovery.ErrServiceNotExists, "get rules failed for service not exist."),
		}, nil
	}
	filter = mutil.NewDomainProjectFilter(domain, project, mutil.ServiceID(request.ServiceId))
	rules, err := getServiceRules(ctx, filter)
	if err != nil {
		return &discovery.GetServiceRulesResponse{
			Response: discovery.CreateResponse(discovery.ErrInternal, err.Error()),
		}, nil
	}
	return &discovery.GetServiceRulesResponse{
		Response: discovery.CreateResponse(discovery.ResponseSuccess, "get service rules successfully."),
		Rules:    rules,
	}, nil
}

func (ds *DataSource) DeleteRule(ctx context.Context, request *discovery.DeleteServiceRulesRequest) (*discovery.DeleteServiceRulesResponse, error) {
	domain := util.ParseDomain(ctx)
	project := util.ParseProject(ctx)
	filter := mutil.NewDomainProjectFilter(domain, project, mutil.ServiceServiceID(request.ServiceId))
	exist, err := exitService(ctx, filter)
	if err != nil {
		log.Error(fmt.Sprintf("failed to add tags for service %s for get service failed", request.ServiceId), err)
		return &discovery.DeleteServiceRulesResponse{
			Response: discovery.CreateResponse(discovery.ErrInternal, "failed to check service exist"),
		}, err
	}
	if !exist {
		return &discovery.DeleteServiceRulesResponse{Response: discovery.CreateResponse(discovery.ErrServiceNotExists, "service not exist")}, nil
	}
	var delRules []mongo.WriteModel
	for _, ruleID := range request.RuleIds {
		filter := mutil.NewDomainProjectFilter(domain, project, mutil.ServiceID(request.ServiceId), mutil.RuleRuleID(ruleID))
		exist, err := existRule(ctx, filter)
		if err != nil {
			return &discovery.DeleteServiceRulesResponse{
				Response: discovery.CreateResponse(discovery.ErrInternal, err.Error()),
			}, err
		}
		if !exist {
			return &discovery.DeleteServiceRulesResponse{
				Response: discovery.CreateResponse(discovery.ErrRuleNotExists, "this rule does not exist."),
			}, nil
		}
		filter = mutil.NewDomainProjectFilter(domain, project, mutil.ServiceID(request.ServiceId), mutil.RuleRuleID(ruleID))
		delRules = append(delRules, mongo.NewDeleteOneModel().SetFilter(filter))
	}
	if len(delRules) > 0 {
		err = batchDeleteRule(ctx, delRules)
		if err != nil {
			return &discovery.DeleteServiceRulesResponse{
				Response: discovery.CreateResponse(discovery.ErrInternal, err.Error()),
			}, err
		}
	}
	return &discovery.DeleteServiceRulesResponse{
		Response: discovery.CreateResponse(discovery.ResponseSuccess, "Delete service rules successfully."),
	}, nil
}

func (ds *DataSource) UpdateRule(ctx context.Context, request *discovery.UpdateServiceRuleRequest) (*discovery.UpdateServiceRuleResponse, error) {
	domain := util.ParseDomain(ctx)
	project := util.ParseProject(ctx)
	filter := mutil.NewDomainProjectFilter(domain, project, mutil.ServiceServiceID(request.ServiceId))
	exist, err := exitService(ctx, filter)
	if err != nil {
		return &discovery.UpdateServiceRuleResponse{
			Response: discovery.CreateResponse(discovery.ErrServiceNotExists, "updateRule failed for get service failed."),
		}, nil
	}
	if !exist {
		return &discovery.UpdateServiceRuleResponse{
			Response: discovery.CreateResponse(discovery.ErrServiceNotExists, "updateRule failed for service not exist."),
		}, nil
	}
	filter = mutil.NewDomainProjectFilter(domain, project, mutil.ServiceID(request.ServiceId))
	rules, err := getServiceRules(ctx, filter)
	if err != nil {
		return &discovery.UpdateServiceRuleResponse{
			Response: discovery.CreateResponse(discovery.ErrUnavailableBackend, "updateRule failed for get rule."),
		}, nil
	}
	if len(rules) >= 1 && rules[0].RuleType != request.Rule.RuleType {
		return &discovery.UpdateServiceRuleResponse{
			Response: discovery.CreateResponse(discovery.ErrModifyRuleNotAllow, "exist multiple rules, can not change rule type. Rule type is ."+rules[0].RuleType),
		}, nil
	}
	filter = mutil.NewDomainProjectFilter(domain, project, mutil.ServiceID(request.ServiceId), mutil.RuleRuleID(request.RuleId))
	exist, err = existRule(ctx, filter)
	if err != nil {
		return &discovery.UpdateServiceRuleResponse{
			Response: discovery.CreateResponse(discovery.ErrInternal, err.Error()),
		}, nil
	}
	if !exist {
		return &discovery.UpdateServiceRuleResponse{
			Response: discovery.CreateResponse(discovery.ErrRuleNotExists, "this rule does not exist."),
		}, nil
	}
	filter = mutil.NewDomainProjectFilter(domain, project, mutil.ServiceID(request.ServiceId), mutil.RuleRuleID(request.RuleId))
	setValue := mutil.NewFilter(
		mutil.RuleRuleType(request.Rule.RuleType),
		mutil.RulePattern(request.Rule.Pattern),
		mutil.RuleAttribute(request.Rule.Attribute),
		mutil.RuleDescription(request.Rule.Description),
		mutil.RuleModTime(strconv.FormatInt(time.Now().Unix(), baseTen)),
	)
	updateFilter := mutil.NewFilter(mutil.Set(setValue))
	err = updateRule(ctx, filter, updateFilter)
	if err != nil {
		return &discovery.UpdateServiceRuleResponse{
			Response: discovery.CreateResponse(discovery.ErrInternal, err.Error()),
		}, err
	}
	return &discovery.UpdateServiceRuleResponse{
		Response: discovery.CreateResponse(discovery.ResponseSuccess, "update service rules succesfully."),
	}, nil
}

func addServiceVersionRule(ctx context.Context, domainProject string, consumer *discovery.MicroService, provider *discovery.MicroServiceKey) error {
	consumerKey := discovery.MicroServiceToKey(domainProject, consumer)
	exist, err := existDependencyRuleByProviderConsumer(ctx, provider, consumerKey)
	if exist || err != nil {
		return err
	}
	r := &discovery.ConsumerDependency{
		Consumer:  consumerKey,
		Providers: []*discovery.MicroServiceKey{provider},
		Override:  false,
	}
	err = syncDependencyRule(ctx, domainProject, r)
	if err != nil {
		return err
	}
	return nil
}
