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

package datasource_test

import (
	"github.com/apache/servicecomb-service-center/datasource"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/server/plugin/quota"
	pb "github.com/go-chassis/cari/discovery"
	"github.com/stretchr/testify/assert"
	"strconv"
	"testing"
)

func TestRule_Add(t *testing.T) {
	var (
		serviceId1 string
		serviceId2 string
	)

	t.Run("register service and instance", func(t *testing.T) {
		respCreateService, err := datasource.Instance().RegisterService(getContext(), &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				AppId:       "create_rule_group_ms",
				ServiceName: "create_rule_service_ms",
				Version:     "1.0.0",
				Level:       "FRONT",
				Status:      pb.MS_UP,
			},
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respCreateService.Response.GetCode())
		serviceId1 = respCreateService.ServiceId

		respCreateService, err = datasource.Instance().RegisterService(getContext(), &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				AppId:       "create_rule_group_ms",
				ServiceName: "create_rule_service_ms",
				Version:     "1.0.1",
				Level:       "FRONT",
				Status:      pb.MS_UP,
			},
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respCreateService.Response.GetCode())
		serviceId2 = respCreateService.ServiceId
	})

	t.Run("invalid request", func(t *testing.T) {
		log.Info("service does not exist")
		respAddRule, err := datasource.Instance().AddRule(getContext(), &pb.AddServiceRulesRequest{
			ServiceId: "not_exist_service_ms",
			Rules: []*pb.AddOrUpdateServiceRule{
				{
					RuleType:    "BLACK",
					Attribute:   "ServiceName",
					Pattern:     "Test*",
					Description: "test white",
				},
			},
		})
		assert.NoError(t, err)
		assert.NotEqual(t, pb.ResponseSuccess, respAddRule)
	})

	t.Run("request is valid", func(t *testing.T) {
		log.Info("create a new black list")
		respAddRule, err := datasource.Instance().AddRule(getContext(), &pb.AddServiceRulesRequest{
			ServiceId: serviceId1,
			Rules: []*pb.AddOrUpdateServiceRule{
				{
					RuleType:    "BLACK",
					Attribute:   "ServiceName",
					Pattern:     "Test*",
					Description: "test black",
				},
			},
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respAddRule.Response.GetCode())
		ruleId := respAddRule.RuleIds[0]
		assert.NotEqual(t, "", ruleId)

		log.Info("create the black list again")
		respAddRule, err = datasource.Instance().AddRule(getContext(), &pb.AddServiceRulesRequest{
			ServiceId: serviceId1,
			Rules: []*pb.AddOrUpdateServiceRule{
				{
					RuleType:    "BLACK",
					Attribute:   "ServiceName",
					Pattern:     "Test*",
					Description: "test change black",
				},
			},
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respAddRule.Response.GetCode())
		assert.Equal(t, 0, len(respAddRule.RuleIds))

		log.Info("create a new white list when black list already exists")
		respAddRule, err = datasource.Instance().AddRule(getContext(), &pb.AddServiceRulesRequest{
			ServiceId: serviceId1,
			Rules: []*pb.AddOrUpdateServiceRule{
				{
					RuleType:    "WHITE",
					Attribute:   "ServiceName",
					Pattern:     "Test*",
					Description: "test white",
				},
			},
		})
		assert.NoError(t, err)
		assert.NotEqual(t, pb.ResponseSuccess, respAddRule.Response.GetCode())
	})

	t.Run("create rule out of gaugue", func(t *testing.T) {
		size := quota.DefaultRuleQuota + 1
		rules := make([]*pb.AddOrUpdateServiceRule, 0, size)
		for i := 0; i < size; i++ {
			rules = append(rules, &pb.AddOrUpdateServiceRule{
				RuleType:    "BLACK",
				Attribute:   "ServiceName",
				Pattern:     strconv.Itoa(i),
				Description: "test white",
			})
		}

		resp, err := datasource.Instance().AddRule(getContext(), &pb.AddServiceRulesRequest{
			ServiceId: serviceId2,
			Rules:     rules[:size-1],
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, resp.Response.GetCode())

		resp, err = datasource.Instance().AddRule(getContext(), &pb.AddServiceRulesRequest{
			ServiceId: serviceId2,
			Rules:     rules[size-1:],
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ErrNotEnoughQuota, resp.Response.GetCode())
	})
}

func TestRule_Get(t *testing.T) {
	var (
		serviceId string
		ruleId    string
	)

	t.Run("register service and rules", func(t *testing.T) {
		respCreateService, err := datasource.Instance().RegisterService(getContext(), &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				AppId:       "get_rule_group_ms",
				ServiceName: "get_rule_service_ms",
				Version:     "1.0.0",
				Level:       "FRONT",
				Status:      pb.MS_UP,
			},
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respCreateService.Response.GetCode())
		serviceId = respCreateService.ServiceId

		respAddRule, err := datasource.Instance().AddRule(getContext(), &pb.AddServiceRulesRequest{
			ServiceId: serviceId,
			Rules: []*pb.AddOrUpdateServiceRule{
				{
					RuleType:    "BLACK",
					Attribute:   "ServiceName",
					Pattern:     "Test*",
					Description: "test BLACK",
				},
			},
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respAddRule.Response.GetCode())
		ruleId = respAddRule.RuleIds[0]
		assert.NotEqual(t, "", ruleId)
	})

	t.Run("get when request is invalid", func(t *testing.T) {
		log.Info("service not exists")
		respGetRule, err := datasource.Instance().GetRules(getContext(), &pb.GetServiceRulesRequest{
			ServiceId: "not_exist_service_ms",
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ErrServiceNotExists, respGetRule.Response.GetCode())
	})

	t.Run("get when request is valid", func(t *testing.T) {
		respGetRule, err := datasource.Instance().GetRules(getContext(), &pb.GetServiceRulesRequest{
			ServiceId: serviceId,
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respGetRule.Response.GetCode())
		assert.Equal(t, ruleId, respGetRule.Rules[0].RuleId)
	})
}

func TestRule_Update(t *testing.T) {
	var (
		serviceId string
		ruleId    string
	)

	t.Run("create service and rules", func(t *testing.T) {
		respCreateService, err := datasource.Instance().RegisterService(getContext(), &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				AppId:       "update_rule_group_ms",
				ServiceName: "update_rule_service_ms",
				Version:     "1.0.0",
				Level:       "FRONT",
				Status:      pb.MS_UP,
			},
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respCreateService.Response.GetCode())
		serviceId = respCreateService.ServiceId

		respAddRule, err := datasource.Instance().AddRule(getContext(), &pb.AddServiceRulesRequest{
			ServiceId: serviceId,
			Rules: []*pb.AddOrUpdateServiceRule{
				{
					RuleType:    "BLACK",
					Attribute:   "ServiceName",
					Pattern:     "Test*",
					Description: "test BLACK",
				},
			},
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respAddRule.Response.GetCode())
		ruleId = respAddRule.RuleIds[0]
		assert.NotEqual(t, "", ruleId)
	})

	t.Run("update when request is invalid", func(t *testing.T) {
		rule := &pb.AddOrUpdateServiceRule{
			RuleType:    "BLACK",
			Attribute:   "ServiceName",
			Pattern:     "Test*",
			Description: "test BLACK update",
		}
		log.Info("service does not exist")
		resp, err := datasource.Instance().UpdateRule(getContext(), &pb.UpdateServiceRuleRequest{
			ServiceId: "not_exist_service_ms",
			RuleId:    ruleId,
			Rule:      rule,
		})
		assert.NoError(t, err)
		assert.NotEqual(t, pb.ResponseSuccess, resp.Response.GetCode())

		log.Info("rule not exists")
		resp, err = datasource.Instance().UpdateRule(getContext(), &pb.UpdateServiceRuleRequest{
			ServiceId: serviceId,
			RuleId:    "not_exist_rule_ms",
			Rule:      rule,
		})
		assert.NoError(t, err)
		assert.NotEqual(t, pb.ResponseSuccess, resp.Response.GetCode())

		log.Info("change rule type")
		resp, err = datasource.Instance().UpdateRule(getContext(), &pb.UpdateServiceRuleRequest{
			ServiceId: serviceId,
			RuleId:    ruleId,
			Rule: &pb.AddOrUpdateServiceRule{
				RuleType:    "WHITE",
				Attribute:   "ServiceName",
				Pattern:     "Test*",
				Description: "test white update",
			},
		})
		assert.NoError(t, err)
		assert.NotEqual(t, pb.ResponseSuccess, resp.Response.GetCode())
	})

	t.Run("update when request is valid", func(t *testing.T) {
		resp, err := datasource.Instance().UpdateRule(getContext(), &pb.UpdateServiceRuleRequest{
			ServiceId: serviceId,
			RuleId:    ruleId,
			Rule: &pb.AddOrUpdateServiceRule{
				RuleType:    "BLACK",
				Attribute:   "AppId",
				Pattern:     "Test*",
				Description: "test white update",
			},
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, resp.Response.GetCode())
	})
}

func TestRule_Delete(t *testing.T) {
	var (
		serviceId string
		ruleId    string
	)
	t.Run("register service and rules", func(t *testing.T) {
		respCreateService, err := datasource.Instance().RegisterService(getContext(), &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				AppId:       "delete_rule_group_ms",
				ServiceName: "delete_rule_service_ms",
				Version:     "1.0.0",
				Level:       "FRONT",
				Status:      pb.MS_UP,
			},
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respCreateService.Response.GetCode())
		serviceId = respCreateService.ServiceId

		respAddRule, err := datasource.Instance().AddRule(getContext(), &pb.AddServiceRulesRequest{
			ServiceId: serviceId,
			Rules: []*pb.AddOrUpdateServiceRule{
				{
					RuleType:    "BLACK",
					Attribute:   "ServiceName",
					Pattern:     "Test*",
					Description: "test BLACK",
				},
			},
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respAddRule.Response.GetCode())
		ruleId = respAddRule.RuleIds[0]
		assert.NotEqual(t, "", ruleId)
	})

	t.Run("delete when request is invalid", func(t *testing.T) {
		log.Info("service not exist")
		resp, err := datasource.Instance().DeleteRule(getContext(), &pb.DeleteServiceRulesRequest{
			ServiceId: "not_exist_service_ms",
			RuleIds:   []string{"1000000"},
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ErrServiceNotExists, resp.Response.GetCode())

		log.Info("rule not exist")
		resp, err = datasource.Instance().DeleteRule(getContext(), &pb.DeleteServiceRulesRequest{
			ServiceId: serviceId,
			RuleIds:   []string{"not_exist_rule"},
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ErrRuleNotExists, resp.Response.GetCode())
	})

	t.Run("delete when request is valid", func(t *testing.T) {
		resp, err := datasource.Instance().DeleteRule(getContext(), &pb.DeleteServiceRulesRequest{
			ServiceId: serviceId,
			RuleIds:   []string{ruleId},
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, resp.Response.GetCode())

		respGetRule, err := datasource.Instance().GetRules(getContext(), &pb.GetServiceRulesRequest{
			ServiceId: serviceId,
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, resp.Response.GetCode())
		assert.Equal(t, 0, len(respGetRule.Rules))
	})
}

func TestRule_Permission(t *testing.T) {
	var (
		consumerVersion string
		consumerTag     string
		providerBlack   string
		providerWhite   string
	)
	t.Run("register service and rules", func(t *testing.T) {
		respCreateService, err := datasource.Instance().RegisterService(getContext(), &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				AppId:       "query_instance_tag_ms",
				ServiceName: "query_instance_version_consumer_ms",
				Version:     "1.0.0",
				Level:       "FRONT",
				Status:      pb.MS_UP,
			},
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respCreateService.Response.GetCode())
		consumerVersion = respCreateService.ServiceId

		respCreateService, err = datasource.Instance().RegisterService(getContext(), &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				AppId:       "query_instance_tag_ms",
				ServiceName: "query_instance_tag_service_ms",
				Version:     "1.0.2",
				Level:       "FRONT",
				Status:      pb.MS_UP,
			},
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respCreateService.Response.GetCode())
		providerBlack = respCreateService.ServiceId

		respAddRule, err := datasource.Instance().AddRule(getContext(), &pb.AddServiceRulesRequest{
			ServiceId: providerBlack,
			Rules: []*pb.AddOrUpdateServiceRule{
				{
					RuleType:  "BLACK",
					Attribute: "Version",
					Pattern:   "1.0.0",
				},
				{
					RuleType:  "BLACK",
					Attribute: "tag_a",
					Pattern:   "b",
				},
			},
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respAddRule.Response.GetCode())

		respCreateService, err = datasource.Instance().RegisterService(getContext(), &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				AppId:       "query_instance_tag_ms",
				ServiceName: "query_instance_tag_service_ms",
				Version:     "1.0.3",
				Level:       "FRONT",
				Status:      pb.MS_UP,
			},
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respCreateService.Response.GetCode())
		providerWhite = respCreateService.ServiceId

		respAddRule, err = datasource.Instance().AddRule(getContext(), &pb.AddServiceRulesRequest{
			ServiceId: providerWhite,
			Rules: []*pb.AddOrUpdateServiceRule{
				{
					RuleType:  "WHITE",
					Attribute: "Version",
					Pattern:   "1.0.0",
				},
				{
					RuleType:  "WHITE",
					Attribute: "tag_a",
					Pattern:   "b",
				},
			},
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respAddRule.Response.GetCode())

		respCreateService, err = datasource.Instance().RegisterService(getContext(), &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				AppId:       "query_instance_tag_ms",
				ServiceName: "query_instance_tag_consumer_ms",
				Version:     "1.0.4",
				Level:       "FRONT",
				Status:      pb.MS_UP,
			},
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respCreateService.Response.GetCode())
		consumerTag = respCreateService.ServiceId

		respAddTag, err := datasource.Instance().AddTags(getContext(), &pb.AddServiceTagsRequest{
			ServiceId: consumerTag,
			Tags:      map[string]string{"a": "b"},
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respAddTag.Response.GetCode())
	})

	t.Run("when query instances", func(t *testing.T) {
		log.Info("consumer version in black list")
		resp, err := datasource.Instance().GetInstances(getContext(), &pb.GetInstancesRequest{
			ConsumerServiceId: consumerVersion,
			ProviderServiceId: providerBlack,
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ErrServiceNotExists, resp.Response.GetCode())

		log.Info("consumer tag in black list")
		resp, err = datasource.Instance().GetInstances(getContext(), &pb.GetInstancesRequest{
			ConsumerServiceId: consumerTag,
			ProviderServiceId: providerBlack,
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ErrServiceNotExists, resp.Response.GetCode())

		log.Info("find should return 200 even if consumer permission deny")
		respFind, err := datasource.Instance().FindInstances(getContext(), &pb.FindInstancesRequest{
			ConsumerServiceId: consumerVersion,
			AppId:             "query_instance_tag_ms",
			ServiceName:       "query_instance_tag_service_ms",
			VersionRule:       "0+",
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respFind.Response.GetCode())
		assert.Equal(t, 0, len(respFind.Instances))

		respFind, err = datasource.Instance().FindInstances(getContext(), &pb.FindInstancesRequest{
			ConsumerServiceId: consumerTag,
			AppId:             "query_instance_tag_ms",
			ServiceName:       "query_instance_tag_service_ms",
			VersionRule:       "0+",
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respFind.Response.GetCode())
		assert.Equal(t, 0, len(respFind.Instances))

		log.Info("consumer not in black list")
		resp, err = datasource.Instance().GetInstances(getContext(), &pb.GetInstancesRequest{
			ConsumerServiceId: providerWhite,
			ProviderServiceId: providerBlack,
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, resp.Response.GetCode())

		log.Info("consumer not in white list")
		resp, err = datasource.Instance().GetInstances(getContext(), &pb.GetInstancesRequest{
			ConsumerServiceId: providerBlack,
			ProviderServiceId: providerWhite,
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ErrServiceNotExists, resp.Response.GetCode())

		log.Info("consumer version in white list")
		resp, err = datasource.Instance().GetInstances(getContext(), &pb.GetInstancesRequest{
			ConsumerServiceId: consumerVersion,
			ProviderServiceId: providerWhite,
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, resp.Response.GetCode())

		log.Info("consumer tag in white list")
		resp, err = datasource.Instance().GetInstances(getContext(), &pb.GetInstancesRequest{
			ConsumerServiceId: consumerTag,
			ProviderServiceId: providerWhite,
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, resp.Response.GetCode())
	})
}
