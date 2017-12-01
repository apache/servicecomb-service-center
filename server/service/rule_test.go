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
package service_test

import (
	pb "github.com/ServiceComb/service-center/server/core/proto"
	scerr "github.com/ServiceComb/service-center/server/error"
	"github.com/ServiceComb/service-center/server/plugin/infra/quota/buildin"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"strconv"
)

var _ = Describe("'Rule' service", func() {
	Describe("execute 'create' operartion", func() {
		var (
			serviceId string
		)

		It("should be passed", func() {
			respCreateService, err := serviceResource.Create(getContext(), &pb.CreateServiceRequest{
				Service: &pb.MicroService{
					AppId:       "create_rule_group",
					ServiceName: "create_rule_service",
					Version:     "1.0.0",
					Level:       "FRONT",
					Status:      pb.MS_UP,
				},
			})
			Expect(err).To(BeNil())
			Expect(respCreateService.Response.Code).To(Equal(pb.Response_SUCCESS))
			serviceId = respCreateService.ServiceId
		})

		Context("when request is invalid", func() {
			It("should be failed", func() {
				By("attribute is empty")
				respAddRule, err := serviceResource.AddRule(getContext(), &pb.AddServiceRulesRequest{
					ServiceId: serviceId,
					Rules: []*pb.AddOrUpdateServiceRule{
						{
							RuleType:    "BLACK",
							Attribute:   "",
							Pattern:     "Test*",
							Description: "test BLACK",
						},
					},
				})
				Expect(err).To(BeNil())
				Expect(respAddRule.Response.Code).ToNot(Equal(pb.Response_SUCCESS))

				By("attribute is tag but name is invalid")
				respAddRule, err = serviceResource.AddRule(getContext(), &pb.AddServiceRulesRequest{
					ServiceId: serviceId,
					Rules: []*pb.AddOrUpdateServiceRule{
						{
							RuleType:    "BLACK",
							Attribute:   "tag_@34",
							Pattern:     "Test*",
							Description: "test BLACK",
						},
					},
				})
				Expect(err).To(BeNil())
				Expect(respAddRule.Response.Code).ToNot(Equal(pb.Response_SUCCESS))

				By("attribute is a invalid field name")
				respAddRule, err = serviceResource.AddRule(getContext(), &pb.AddServiceRulesRequest{
					ServiceId: serviceId,
					Rules: []*pb.AddOrUpdateServiceRule{
						{
							RuleType:    "BLACK",
							Attribute:   "xxx",
							Pattern:     "Test*",
							Description: "test BLACK",
						},
					},
				})
				Expect(err).To(BeNil())
				Expect(respAddRule.Response.Code).ToNot(Equal(pb.Response_SUCCESS))

				By("service does not exist")
				respAddRule, err = serviceResource.AddRule(getContext(), &pb.AddServiceRulesRequest{
					ServiceId: "notexistservice",
					Rules: []*pb.AddOrUpdateServiceRule{
						{
							RuleType:    "BLACK",
							Attribute:   "ServiceName",
							Pattern:     "Test*",
							Description: "test white",
						},
					},
				})
				Expect(err).To(BeNil())
				Expect(respAddRule.Response.Code).ToNot(Equal(pb.Response_SUCCESS))

				By("service id is empty")
				respAddRule, err = serviceResource.AddRule(getContext(), &pb.AddServiceRulesRequest{
					ServiceId: "",
					Rules: []*pb.AddOrUpdateServiceRule{
						{
							RuleType:    "BLACK",
							Attribute:   "",
							Pattern:     "Test*",
							Description: "test white",
						},
					},
				})
				Expect(err).To(BeNil())
				Expect(respAddRule.Response.Code).ToNot(Equal(pb.Response_SUCCESS))

				By("rule is nil")
				respAddRule, err = serviceResource.AddRule(getContext(), &pb.AddServiceRulesRequest{
					ServiceId: serviceId,
				})
				Expect(err).To(BeNil())
				Expect(respAddRule.Response.Code).ToNot(Equal(pb.Response_SUCCESS))
			})
		})

		Context("when request is valid", func() {
			It("should be passed", func() {
				By("create a new black list")
				respAddRule, err := serviceResource.AddRule(getContext(), &pb.AddServiceRulesRequest{
					ServiceId: serviceId,
					Rules: []*pb.AddOrUpdateServiceRule{
						{
							RuleType:    "BLACK",
							Attribute:   "ServiceName",
							Pattern:     "Test*",
							Description: "test black",
						},
					},
				})
				Expect(err).To(BeNil())
				Expect(respAddRule.Response.Code).To(Equal(pb.Response_SUCCESS))
				ruleId := respAddRule.RuleIds[0]
				Expect(ruleId).ToNot(Equal(""))

				By("create the black list again")
				respAddRule, err = serviceResource.AddRule(getContext(), &pb.AddServiceRulesRequest{
					ServiceId: serviceId,
					Rules: []*pb.AddOrUpdateServiceRule{
						{
							RuleType:    "BLACK",
							Attribute:   "ServiceName",
							Pattern:     "Test*",
							Description: "test change black",
						},
					},
				})
				Expect(err).To(BeNil())
				Expect(respAddRule.Response.Code).To(Equal(pb.Response_SUCCESS))
				Expect(len(respAddRule.RuleIds)).To(Equal(0)) // no changed

				By("create a new white list when black list already exists")
				respAddRule, err = serviceResource.AddRule(getContext(), &pb.AddServiceRulesRequest{
					ServiceId: serviceId,
					Rules: []*pb.AddOrUpdateServiceRule{
						{
							RuleType:    "WHITE",
							Attribute:   "ServiceName",
							Pattern:     "Test*",
							Description: "test white",
						},
					},
				})
				Expect(err).To(BeNil())
				Expect(respAddRule.Response.Code).ToNot(Equal(pb.Response_SUCCESS))
			})
		})

		Context("when create rule out of gauge", func() {
			It("should be failed", func() {
				size := buildin.RULE_NUM_MAX_FOR_ONESERVICE + 1
				rules := make([]*pb.AddOrUpdateServiceRule, 0, size)
				for i := 0; i < size; i++ {
					rules = append(rules, &pb.AddOrUpdateServiceRule{
						RuleType:    "BLACK",
						Attribute:   "ServiceName",
						Pattern:     strconv.Itoa(i),
						Description: "test white",
					})
				}
				resp, err := serviceResource.AddRule(getContext(), &pb.AddServiceRulesRequest{
					ServiceId: serviceId,
					Rules:     rules,
				})
				Expect(err).To(BeNil())
				Expect(resp.Response.Code).ToNot(Equal(pb.Response_SUCCESS))
			})
		})
	})

	Describe("execute 'get' operartion", func() {
		var (
			serviceId string
			ruleId    string
		)

		It("should be passed", func() {
			respCreateService, err := serviceResource.Create(getContext(), &pb.CreateServiceRequest{
				Service: &pb.MicroService{
					AppId:       "get_rule_group",
					ServiceName: "get_rule_service",
					Version:     "1.0.0",
					Level:       "FRONT",
					Status:      pb.MS_UP,
				},
			})
			Expect(err).To(BeNil())
			Expect(respCreateService.Response.Code).To(Equal(pb.Response_SUCCESS))
			serviceId = respCreateService.ServiceId

			respAddRule, err := serviceResource.AddRule(getContext(), &pb.AddServiceRulesRequest{
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
			Expect(err).To(BeNil())
			Expect(respAddRule.Response.Code).To(Equal(pb.Response_SUCCESS))
			ruleId = respAddRule.RuleIds[0]
			Expect(ruleId).ToNot(Equal(""))
		})

		Context("when request is invalid", func() {
			It("should be failed", func() {
				By("service id is empty")
				respGetRule, err := serviceResource.GetRule(getContext(), &pb.GetServiceRulesRequest{
					ServiceId: "",
				})
				Expect(err).To(BeNil())
				Expect(respGetRule.Response.Code).ToNot(Equal(pb.Response_SUCCESS))

				By("service does not exist")
				respGetRule, err = serviceResource.GetRule(getContext(), &pb.GetServiceRulesRequest{
					ServiceId: "notexist",
				})
				Expect(err).To(BeNil())
				Expect(respGetRule.Response.Code).ToNot(Equal(pb.Response_SUCCESS))
			})
		})

		Context("when request is valid", func() {
			It("should be passed", func() {
				respGetRule, err := serviceResource.GetRule(getContext(), &pb.GetServiceRulesRequest{
					ServiceId: serviceId,
				})
				Expect(err).To(BeNil())
				Expect(respGetRule.Response.Code).To(Equal(pb.Response_SUCCESS))
				Expect(respGetRule.Rules[0].RuleId).To(Equal(ruleId))
			})
		})
	})

	Describe("execute 'update' operartion", func() {
		var (
			serviceId string
			ruleId    string
		)

		It("should be passed", func() {
			respCreateService, err := serviceResource.Create(getContext(), &pb.CreateServiceRequest{
				Service: &pb.MicroService{
					AppId:       "update_rule_group",
					ServiceName: "update_rule_service",
					Version:     "1.0.0",
					Level:       "FRONT",
					Status:      pb.MS_UP,
				},
			})
			Expect(err).To(BeNil())
			Expect(respCreateService.Response.Code).To(Equal(pb.Response_SUCCESS))
			serviceId = respCreateService.ServiceId

			respAddRule, err := serviceResource.AddRule(getContext(), &pb.AddServiceRulesRequest{
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
			Expect(err).To(BeNil())
			Expect(respAddRule.Response.Code).To(Equal(pb.Response_SUCCESS))
			ruleId = respAddRule.RuleIds[0]
			Expect(ruleId).ToNot(Equal(""))
		})

		Context("when request is invalid", func() {
			It("should be failed", func() {
				rule := &pb.AddOrUpdateServiceRule{
					RuleType:    "BLACK",
					Attribute:   "ServiceName",
					Pattern:     "Test*",
					Description: "test BLACK update",
				}
				By("service id is empty")
				respAddRule, err := serviceResource.UpdateRule(getContext(), &pb.UpdateServiceRuleRequest{
					ServiceId: "",
					RuleId:    ruleId,
					Rule:      rule,
				})
				Expect(err).To(BeNil())
				Expect(respAddRule.Response.Code).ToNot(Equal(pb.Response_SUCCESS))

				By("service does not exist")
				respAddRule, err = serviceResource.UpdateRule(getContext(), &pb.UpdateServiceRuleRequest{
					ServiceId: "notexistservice",
					RuleId:    ruleId,
					Rule:      rule,
				})
				Expect(err).To(BeNil())
				Expect(respAddRule.Response.Code).ToNot(Equal(pb.Response_SUCCESS))

				By("rule id is empty")
				respAddRule, err = serviceResource.UpdateRule(getContext(), &pb.UpdateServiceRuleRequest{
					ServiceId: serviceId,
					RuleId:    "",
					Rule:      rule,
				})
				Expect(err).To(BeNil())
				Expect(respAddRule.Response.Code).ToNot(Equal(pb.Response_SUCCESS))

				By("rule does not exist")
				respAddRule, err = serviceResource.UpdateRule(getContext(), &pb.UpdateServiceRuleRequest{
					ServiceId: serviceId,
					RuleId:    "notexistrule",
					Rule:      rule,
				})
				Expect(err).To(BeNil())
				Expect(respAddRule.Response.Code).ToNot(Equal(pb.Response_SUCCESS))

				By("rule type is invalid")
				respAddRule, err = serviceResource.UpdateRule(getContext(), &pb.UpdateServiceRuleRequest{
					ServiceId: serviceId,
					RuleId:    ruleId,
					Rule: &pb.AddOrUpdateServiceRule{
						RuleType:    "notType",
						Attribute:   "ServiceName",
						Pattern:     "Test*",
						Description: "test BLACK update",
					},
				})
				Expect(err).To(BeNil())
				Expect(respAddRule.Response.Code).ToNot(Equal(pb.Response_SUCCESS))

				By("attribute is a invalid field name")
				respAddRule, err = serviceResource.UpdateRule(getContext(), &pb.UpdateServiceRuleRequest{
					ServiceId: serviceId,
					RuleId:    ruleId,
					Rule: &pb.AddOrUpdateServiceRule{
						RuleType:    "BLACK",
						Attribute:   "noattribute",
						Pattern:     "Test*",
						Description: "test BLACK update",
					},
				})
				Expect(err).To(BeNil())
				Expect(respAddRule.Response.Code).ToNot(Equal(pb.Response_SUCCESS))

				By("attribute is tag but name is invalid")
				respAddRule, err = serviceResource.UpdateRule(getContext(), &pb.UpdateServiceRuleRequest{
					ServiceId: serviceId,
					RuleId:    ruleId,
					Rule: &pb.AddOrUpdateServiceRule{
						RuleType:    "BLACK",
						Attribute:   "tag_@34",
						Pattern:     "Test*",
						Description: "test BLACK update",
					},
				})
				Expect(err).To(BeNil())
				Expect(respAddRule.Response.Code).ToNot(Equal(pb.Response_SUCCESS))

				By("change rule type")
				respAddRule, err = serviceResource.UpdateRule(getContext(), &pb.UpdateServiceRuleRequest{
					ServiceId: serviceId,
					RuleId:    ruleId,
					Rule: &pb.AddOrUpdateServiceRule{
						RuleType:    "WHITE",
						Attribute:   "ServiceName",
						Pattern:     "Test*",
						Description: "test white update",
					},
				})
				Expect(err).To(BeNil())
				Expect(respAddRule.Response.Code).ToNot(Equal(pb.Response_SUCCESS))
			})
		})

		Context("when request is valid", func() {
			It("should be passed", func() {
				respAddRule, err := serviceResource.UpdateRule(getContext(), &pb.UpdateServiceRuleRequest{
					ServiceId: serviceId,
					RuleId:    ruleId,
					Rule: &pb.AddOrUpdateServiceRule{
						RuleType:    "BLACK",
						Attribute:   "AppId",
						Pattern:     "Test*",
						Description: "test white update",
					},
				})
				Expect(err).To(BeNil())
				Expect(respAddRule.Response.Code).To(Equal(pb.Response_SUCCESS))
			})
		})
	})

	Describe("execute 'delete' operartion", func() {
		var (
			serviceId string
			ruleId    string
		)

		It("should be passed", func() {
			respCreateService, err := serviceResource.Create(getContext(), &pb.CreateServiceRequest{
				Service: &pb.MicroService{
					AppId:       "delete_rule_group",
					ServiceName: "delete_rule_service",
					Version:     "1.0.0",
					Level:       "FRONT",
					Status:      pb.MS_UP,
				},
			})
			Expect(err).To(BeNil())
			Expect(respCreateService.Response.Code).To(Equal(pb.Response_SUCCESS))
			serviceId = respCreateService.ServiceId

			respAddRule, err := serviceResource.AddRule(getContext(), &pb.AddServiceRulesRequest{
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
			Expect(err).To(BeNil())
			Expect(respAddRule.Response.Code).To(Equal(pb.Response_SUCCESS))
			ruleId = respAddRule.RuleIds[0]
			Expect(ruleId).ToNot(Equal(""))
		})

		Context("when request is invalid", func() {
			It("should be failed", func() {
				By("service id is empty")
				respAddRule, err := serviceResource.DeleteRule(getContext(), &pb.DeleteServiceRulesRequest{
					ServiceId: "",
					RuleIds:   []string{"1000000"},
				})
				Expect(err).To(BeNil())
				Expect(respAddRule.Response.Code).ToNot(Equal(pb.Response_SUCCESS))

				By("service does not exist")
				respAddRule, err = serviceResource.DeleteRule(getContext(), &pb.DeleteServiceRulesRequest{
					ServiceId: "notexistservice",
					RuleIds:   []string{"1000000"},
				})
				Expect(err).To(BeNil())
				Expect(respAddRule.Response.Code).ToNot(Equal(pb.Response_SUCCESS))

				By("rule does not exist")
				respAddRule, err = serviceResource.DeleteRule(getContext(), &pb.DeleteServiceRulesRequest{
					ServiceId: serviceId,
					RuleIds:   []string{"notexistrule"},
				})
				Expect(err).To(BeNil())
				Expect(respAddRule.Response.Code).ToNot(Equal(pb.Response_SUCCESS))

				By("rules is empty")
				respAddRule, err = serviceResource.DeleteRule(getContext(), &pb.DeleteServiceRulesRequest{
					ServiceId: serviceId,
				})
				Expect(err).To(BeNil())
				Expect(respAddRule.Response.Code).ToNot(Equal(pb.Response_SUCCESS))

			})
		})

		Context("when request is valid", func() {
			It("should be passed", func() {
				respAddRule, err := serviceResource.DeleteRule(getContext(), &pb.DeleteServiceRulesRequest{
					ServiceId: serviceId,
					RuleIds:   []string{ruleId},
				})
				Expect(err).To(BeNil())
				Expect(respAddRule.Response.Code).To(Equal(pb.Response_SUCCESS))

				respGetRule, err := serviceResource.GetRule(getContext(), &pb.GetServiceRulesRequest{
					ServiceId: serviceId,
				})
				Expect(err).To(BeNil())
				Expect(respGetRule.Response.Code).To(Equal(pb.Response_SUCCESS))
				Expect(len(respGetRule.Rules)).To(Equal(0))
			})
		})
	})

	Describe("execute 'permission' operartion", func() {
		var (
			consumerVersion string
			consumerTag     string
			providerBlack   string
			providerWhite   string
		)

		It("should be passed", func() {
			respCreate, err := serviceResource.Create(getContext(), &pb.CreateServiceRequest{
				Service: &pb.MicroService{
					AppId:       "query_instance_tag",
					ServiceName: "query_instance_tag_service",
					Version:     "1.0.0",
					Level:       "FRONT",
					Status:      pb.MS_UP,
				},
			})
			Expect(err).To(BeNil())
			Expect(respCreate.Response.Code).To(Equal(pb.Response_SUCCESS))
			consumerVersion = respCreate.ServiceId

			respCreate, err = serviceResource.Create(getContext(), &pb.CreateServiceRequest{
				Service: &pb.MicroService{
					AppId:       "query_instance_tag",
					ServiceName: "query_instance_tag_service",
					Version:     "1.0.2",
					Level:       "FRONT",
					Status:      pb.MS_UP,
				},
			})
			Expect(err).To(BeNil())
			Expect(respCreate.Response.Code).To(Equal(pb.Response_SUCCESS))
			providerBlack = respCreate.ServiceId

			resp, err := serviceResource.AddRule(getContext(), &pb.AddServiceRulesRequest{
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
			Expect(err).To(BeNil())
			Expect(resp.Response.Code).To(Equal(pb.Response_SUCCESS))

			respCreate, err = serviceResource.Create(getContext(), &pb.CreateServiceRequest{
				Service: &pb.MicroService{
					AppId:       "query_instance_tag",
					ServiceName: "query_instance_tag_service",
					Version:     "1.0.3",
					Level:       "FRONT",
					Status:      pb.MS_UP,
				},
			})
			Expect(err).To(BeNil())
			Expect(respCreate.Response.Code).To(Equal(pb.Response_SUCCESS))
			providerWhite = respCreate.ServiceId

			resp, err = serviceResource.AddRule(getContext(), &pb.AddServiceRulesRequest{
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
			Expect(err).To(BeNil())
			Expect(resp.Response.Code).To(Equal(pb.Response_SUCCESS))

			respCreate, err = serviceResource.Create(getContext(), &pb.CreateServiceRequest{
				Service: &pb.MicroService{
					AppId:       "query_instance_tag",
					ServiceName: "query_instance_tag_service",
					Version:     "1.0.4",
					Level:       "FRONT",
					Status:      pb.MS_UP,
				},
			})
			Expect(err).To(BeNil())
			Expect(respCreate.Response.Code).To(Equal(pb.Response_SUCCESS))
			consumerTag = respCreate.ServiceId

			resp1, err := serviceResource.AddTags(getContext(), &pb.AddServiceTagsRequest{
				ServiceId: consumerTag,
				Tags:      map[string]string{"a": "b"},
			})
			Expect(err).To(BeNil())
			Expect(resp1.Response.Code).To(Equal(pb.Response_SUCCESS))
		})

		Context("when query instances", func() {
			It("should be failed", func() {
				By("consumer version in black list")
				resp, err := instanceResource.GetInstances(getContext(), &pb.GetInstancesRequest{
					ConsumerServiceId: consumerVersion,
					ProviderServiceId: providerBlack,
				})
				Expect(err).To(BeNil())
				Expect(resp.Response.Code).To(Equal(scerr.ErrPermissionDeny))

				By("consumer tag in black list")
				resp, err = instanceResource.GetInstances(getContext(), &pb.GetInstancesRequest{
					ConsumerServiceId: consumerTag,
					ProviderServiceId: providerBlack,
				})
				Expect(err).To(BeNil())
				Expect(resp.Response.Code).To(Equal(scerr.ErrPermissionDeny))

				By("consumer not in black list")
				resp, err = instanceResource.GetInstances(getContext(), &pb.GetInstancesRequest{
					ConsumerServiceId: providerWhite,
					ProviderServiceId: providerBlack,
				})
				Expect(err).To(BeNil())
				Expect(resp.Response.Code).To(Equal(pb.Response_SUCCESS))

				By("consumer not in white list")
				resp, err = instanceResource.GetInstances(getContext(), &pb.GetInstancesRequest{
					ConsumerServiceId: providerBlack,
					ProviderServiceId: providerWhite,
				})
				Expect(err).To(BeNil())
				Expect(resp.Response.Code).To(Equal(scerr.ErrPermissionDeny))

				By("consumer version in white list")
				resp, err = instanceResource.GetInstances(getContext(), &pb.GetInstancesRequest{
					ConsumerServiceId: consumerVersion,
					ProviderServiceId: providerWhite,
				})
				Expect(err).To(BeNil())
				Expect(resp.Response.Code).To(Equal(pb.Response_SUCCESS))

				By("consumer tag in white list")
				resp, err = instanceResource.GetInstances(getContext(), &pb.GetInstancesRequest{
					ConsumerServiceId: consumerTag,
					ProviderServiceId: providerWhite,
				})
				Expect(err).To(BeNil())
				Expect(resp.Response.Code).To(Equal(pb.Response_SUCCESS))
			})
		})
	})
})
