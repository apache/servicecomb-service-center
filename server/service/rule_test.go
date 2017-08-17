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
	"fmt"
	pb "github.com/ServiceComb/service-center/server/core/proto"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ServiceController", func() {
	Describe("rule", func() {
		Context("normal", func() {
			It("创建rule,参数校验", func() {
				fmt.Println("UT===========创建rule，参数校验")
				respAddRule, err := serviceResource.AddRule(getContext(), &pb.AddServiceRulesRequest{
					ServiceId: serviceId,
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
				Expect(respAddRule.GetResponse().Code).To(Equal(pb.Response_FAIL))

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
				Expect(respAddRule.GetResponse().Code).To(Equal(pb.Response_FAIL))

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
				Expect(respAddRule.GetResponse().Code).To(Equal(pb.Response_FAIL))
			})
			ruleId := ""
			It("创建rule", func() {
				fmt.Println("UT===========创建rule")
				resp, err := serviceResource.Create(getContext(), &pb.CreateServiceRequest{
					Service: &pb.MicroService{
						ServiceName: "service_name_consumer",
						AppId:       "service_group_consumer",
						Version:     "3.0.0",
						Level:       "FRONT",
						Schemas: []string{
							"xxxxxxxx",
						},
						Status: "UP",
					},
				})
				Expect(err).To(BeNil())
				serviceId = resp.ServiceId

				respAddRule, err := serviceResource.AddRule(getContext(), &pb.AddServiceRulesRequest{
					ServiceId: serviceId,
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
				Expect(respAddRule.GetResponse().Code).To(Equal(pb.Response_SUCCESS))
				ruleId = respAddRule.RuleIds[0]

				//添加重复rule
				respAddRule, err = serviceResource.AddRule(getContext(), &pb.AddServiceRulesRequest{
					ServiceId: serviceId,
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
				Expect(respAddRule.GetResponse().Code).To(Equal(pb.Response_SUCCESS))
			})
			It("获取rule，参数校验", func() {
				fmt.Println("UT===========获取rule")
				respGetRule, err := serviceResource.GetRule(getContext(), &pb.GetServiceRulesRequest{
					ServiceId: "",
				})
				Expect(err).To(BeNil())
				Expect(respGetRule.GetResponse().Code).To(Equal(pb.Response_FAIL))

				respGetRule, err = serviceResource.GetRule(getContext(), &pb.GetServiceRulesRequest{
					ServiceId: "notexist",
				})
				Expect(err).To(BeNil())
				Expect(respGetRule.GetResponse().Code).To(Equal(pb.Response_FAIL))
			})
			It("获取rule", func() {
				respAddRule, err := serviceResource.AddRule(getContext(), &pb.AddServiceRulesRequest{
					ServiceId: serviceId,
					Rules: []*pb.AddOrUpdateServiceRule{
						{
							RuleType:    "notexistType",
							Attribute:   "ServiceName",
							Pattern:     "Test*",
							Description: "test white",
						},
					},
				})
				Expect(err).To(BeNil())
				Expect(respAddRule.GetResponse().Code).To(Equal(pb.Response_FAIL))

				//只能有一种ruleType
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
				Expect(respAddRule.GetResponse().Code).To(Equal(pb.Response_FAIL))

				fmt.Println("UT===========获取rule")
				respGetRule, err := serviceResource.GetRule(getContext(), &pb.GetServiceRulesRequest{
					ServiceId: serviceId,
				})
				Expect(err).To(BeNil())
				Expect(respGetRule.GetResponse().Code).To(Equal(pb.Response_SUCCESS))
			})
			It("修改rule,参数校验", func() {
				fmt.Println("UT===========修改rule,参数校验")
				respAddRule, err := serviceResource.UpdateRule(getContext(), &pb.UpdateServiceRuleRequest{
					ServiceId: "",
					RuleId:    ruleId,
					Rule: &pb.AddOrUpdateServiceRule{
						RuleType:    "BLACK",
						Attribute:   "ServiceName",
						Pattern:     "Test*",
						Description: "test white update",
					},
				})
				Expect(err).To(BeNil())
				Expect(respAddRule.GetResponse().Code).To(Equal(pb.Response_FAIL))

				respAddRule, err = serviceResource.UpdateRule(getContext(), &pb.UpdateServiceRuleRequest{
					ServiceId: "notexistservice",
					RuleId:    ruleId,
					Rule: &pb.AddOrUpdateServiceRule{
						RuleType:    "BLACK",
						Attribute:   "ServiceName",
						Pattern:     "Test*",
						Description: "test white update",
					},
				})
				Expect(err).To(BeNil())
				Expect(respAddRule.GetResponse().Code).To(Equal(pb.Response_FAIL))

				respAddRule, err = serviceResource.UpdateRule(getContext(), &pb.UpdateServiceRuleRequest{
					ServiceId: serviceId,
					RuleId:    ruleId,
					Rule: &pb.AddOrUpdateServiceRule{
						RuleType:    "notType",
						Attribute:   "ServiceName",
						Pattern:     "Test*",
						Description: "test white update",
					},
				})
				Expect(err).To(BeNil())
				Expect(respAddRule.GetResponse().Code).To(Equal(pb.Response_FAIL))

				respAddRule, err = serviceResource.UpdateRule(getContext(), &pb.UpdateServiceRuleRequest{
					ServiceId: serviceId,
					RuleId:    ruleId,
					Rule: &pb.AddOrUpdateServiceRule{
						RuleType:    "notType",
						Attribute:   "noattribute",
						Pattern:     "Test*",
						Description: "test white update",
					},
				})
				Expect(err).To(BeNil())
				Expect(respAddRule.GetResponse().Code).To(Equal(pb.Response_FAIL))
			})
			It("修改rule", func() {
				fmt.Println("UT===========修改rule")
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
				Expect(respAddRule.GetResponse().Code).To(Equal(pb.Response_SUCCESS))

				respAddRule, err = serviceResource.UpdateRule(getContext(), &pb.UpdateServiceRuleRequest{
					ServiceId: serviceId,
					RuleId:    ruleId,
					Rule: &pb.AddOrUpdateServiceRule{
						RuleType:    "BLACK",
						Attribute:   "AppId",
						Pattern:     "pattenChanged*",
						Description: "test white update",
					},
				})
				Expect(err).To(BeNil())
				Expect(respAddRule.GetResponse().Code).To(Equal(pb.Response_SUCCESS))

				respAddRule, err = serviceResource.UpdateRule(getContext(), &pb.UpdateServiceRuleRequest{
					ServiceId: serviceId,
					RuleId:    ruleId,
					Rule: &pb.AddOrUpdateServiceRule{
						RuleType:    "WHITE",
						Attribute:   "AppId",
						Pattern:     "Test*",
						Description: "test white update",
					},
				})
				Expect(respAddRule.GetResponse().Code).To(Equal(pb.Response_FAIL))

				respAddRule, err = serviceResource.UpdateRule(getContext(), &pb.UpdateServiceRuleRequest{
					ServiceId: serviceId,
					RuleId:    "notExistRuleId",
					Rule: &pb.AddOrUpdateServiceRule{
						RuleType:    "BLACK",
						Attribute:   "AppId",
						Pattern:     "Test*",
						Description: "test white update",
					},
				})
				Expect(respAddRule.GetResponse().Code).To(Equal(pb.Response_FAIL))
			})

			It("删除rule,参数校验", func() {
				fmt.Println("UT===========删除rule")
				respAddRule, err := serviceResource.DeleteRule(getContext(), &pb.DeleteServiceRulesRequest{
					ServiceId: serviceId,
					RuleIds:   []string{"1000000"},
				})
				Expect(err).To(BeNil())
				Expect(respAddRule.GetResponse().Code).To(Equal(pb.Response_FAIL))

				respAddRule, err = serviceResource.DeleteRule(getContext(), &pb.DeleteServiceRulesRequest{
					ServiceId: serviceId,
					RuleIds:   []string{},
				})
				Expect(err).To(BeNil())
				Expect(respAddRule.GetResponse().Code).To(Equal(pb.Response_FAIL))

			})

			It("删除rule", func() {
				fmt.Println("UT===========删除rule")
				respAddRule, err := serviceResource.DeleteRule(getContext(), &pb.DeleteServiceRulesRequest{
					ServiceId: serviceId,
					RuleIds:   []string{ruleId},
				})
				Expect(err).To(BeNil())
				Expect(respAddRule.GetResponse().Code).To(Equal(pb.Response_SUCCESS))

				respDel, err := serviceResource.Delete(getContext(), &pb.DeleteServiceRequest{
					ServiceId: serviceId,
					Force:     true,
				})
				Expect(err).To(BeNil())
				Expect(respDel.GetResponse().Code).To(Equal(pb.Response_SUCCESS))
			})
			It("删除serviceId不存在rule", func() {
				fmt.Println("UT===========删除serviceId不存在rule")
				respAddRule, _ := serviceResource.DeleteRule(getContext(), &pb.DeleteServiceRulesRequest{
					ServiceId: "",
					RuleIds:   []string{ruleId},
				})
				//Expect(err).To(BeNil())
				Expect(respAddRule.GetResponse().Code).To(Equal(pb.Response_FAIL))

				fmt.Println("UT===========删除serviceId不存在rule")
				respAddRule, _ = serviceResource.DeleteRule(getContext(), &pb.DeleteServiceRulesRequest{
					ServiceId: "unsure",
					RuleIds:   []string{ruleId},
				})
				//Expect(err).To(BeNil())
				Expect(respAddRule.GetResponse().Code).To(Equal(pb.Response_FAIL))
			})
		})
	})
})
