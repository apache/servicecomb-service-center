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
	"github.com/ServiceComb/service-center/server/core"
	pb "github.com/ServiceComb/service-center/server/core/proto"
	"github.com/ServiceComb/service-center/server/plugin/infra/quota/buildin"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"strconv"
)

var _ = Describe("ServiceController", func() {
	var serviceId string
	var serviceId2 string
	var serviceId3 string
	Describe("Create", func() {
		Context("normal", func() {
			By("param check", func() {
				It("service is nil", func() {
					resp, err := serviceResource.Create(getContext(), &pb.CreateServiceRequest{
						Service: nil,
					})
					Expect(err).To(BeNil())
					Expect(resp.GetResponse().Code).ToNot(Equal(pb.Response_SUCCESS))

				})
			})
			It("create service rule tag at once", func() {
				tags := make(map[string]string, 10)
				tags["test"] = "test"
				rules := []*pb.AddOrUpdateServiceRule{
					&pb.AddOrUpdateServiceRule{
						RuleType:    "BLACK",
						Attribute:   "ServiceName",
						Pattern:     "test",
						Description: "test",
					},
				}
				instances := []*pb.MicroServiceInstance{
					&pb.MicroServiceInstance{
						Endpoints: []string{
							"rest:127.0.0.1:8080",
						},
						HostName:    "UT-HOST",
						Status:      pb.MSI_UP,
						Environment: "production",
					},
				}
				fmt.Println("start is ------------>")
				resp, err := serviceResource.Create(getContext(), &pb.CreateServiceRequest{
					Service: &pb.MicroService{
						ServiceName: "create_serivce_rule_tag",
						AppId:       "default",
						Version:     "1.0.0",
						Level:       "FRONT",
						Schemas: []string{
							"xxxxxxxx",
						},
						Status: "UP",
					},
					Tags:      tags,
					Rules:     rules,
					Instances: instances,
				})
				Expect(err).To(BeNil())
				Expect(resp.GetResponse().Code).To(Equal(pb.Response_SUCCESS))

				delete(tags, "test")
				tags["second"] = "second"
				resp, err = serviceResource.Create(getContext(), &pb.CreateServiceRequest{
					Service: &pb.MicroService{
						ServiceName: "create_serivce_rule_tag",
						AppId:       "default",
						Version:     "1.0.0",
						Level:       "FRONT",
						Schemas: []string{
							"xxxxxxxx",
						},
						Status: "UP",
					},
					Tags: tags,
				})
				Expect(err).To(BeNil())
				Expect(resp.GetResponse().Code).To(Equal(pb.Response_SUCCESS))

				respDelete, err := serviceResource.Delete(getContext(), &pb.DeleteServiceRequest{
					ServiceId: resp.ServiceId,
					Force:     true,
				})
				Expect(err).To(BeNil())
				Expect(respDelete.GetResponse().Code).To(Equal(pb.Response_SUCCESS))

			})

			It("创建微服务1", func() {
				resp, err := serviceResource.Create(getContext(), &pb.CreateServiceRequest{
					Service: &pb.MicroService{
						ServiceName: "some-relay",
						Alias:       "sr",
						AppId:       "default",
						Version:     "1.0.0",
						Level:       "FRONT",
						Schemas: []string{
							"xxxxxxxx",
						},
						Status: "UP",
					},
				})
				Expect(err).To(BeNil())
				serviceId = resp.ServiceId
				fmt.Println("UT=========" + serviceId)
				Expect(resp.GetResponse().Code).To(Equal(pb.Response_SUCCESS))

				resp, err = serviceResource.Create(getContext(), &pb.CreateServiceRequest{
					Service: &pb.MicroService{
						ServiceName: "some-relay",
						Alias:       "sr1",
						AppId:       "default",
						Version:     "1.0.0",
						Level:       "FRONT",
						Schemas: []string{
							"xxxxxxxx",
						},
						Status: "UP",
					},
				})
				Expect(err).To(BeNil())
				fmt.Println("UT=============" + resp.String())
				Expect(resp.GetResponse().Code).ToNot(Equal(pb.Response_SUCCESS))

				resp, err = serviceResource.Create(getContext(), &pb.CreateServiceRequest{
					Service: &pb.MicroService{
						ServiceName: "some-relay1",
						Alias:       "sr",
						AppId:       "default",
						Version:     "1.0.0",
						Level:       "FRONT",
						Schemas: []string{
							"xxxxxxxx",
						},
						Status: "UP",
					},
				})
				Expect(err).To(BeNil())
				fmt.Println("UT=============" + resp.String())
				Expect(resp.GetResponse().Code).ToNot(Equal(pb.Response_SUCCESS))

				resp, err = serviceResource.Create(getContext(), &pb.CreateServiceRequest{
					Service: &pb.MicroService{
						ServiceId:   "custom_Id",
						ServiceName: "some-relay1",
						Alias:       "sr",
						AppId:       "default",
						Version:     "1.0.0",
						Level:       "FRONT",
						Schemas: []string{
							"xxxxxxxx",
						},
						Status: "UP",
					},
				})
				Expect(err).To(BeNil())
				fmt.Println("UT=============" + resp.String())
				Expect(resp.GetResponse().Code).ToNot(Equal(pb.Response_SUCCESS))
			})

			It("微服务是否存在,参数校验", func() {
				resp, err := serviceResource.Exist(getContext(), &pb.GetExistenceRequest{
					Type:      "schema",
					ServiceId: "noneservice",
					SchemaId:  "xxxxxxxx",
				})
				Expect(err).To(BeNil())
				fmt.Println("UT=============" + resp.String())
				Expect(resp.GetResponse().Code).ToNot(Equal(pb.Response_SUCCESS))

				resp, err = serviceResource.Exist(getContext(), &pb.GetExistenceRequest{
					Type:      "schema",
					ServiceId: serviceId,
					SchemaId:  "noneschema",
				})
				Expect(err).To(BeNil())
				fmt.Println("UT=============" + resp.String())
				Expect(resp.GetResponse().Code).ToNot(Equal(pb.Response_SUCCESS))

				resp, err = serviceResource.Exist(getContext(), &pb.GetExistenceRequest{
					Type:      "nonetype",
					ServiceId: serviceId,
					SchemaId:  "noneschema",
				})
				Expect(err).To(BeNil())
				fmt.Println("UT=============" + resp.String())
				Expect(resp.GetResponse().Code).ToNot(Equal(pb.Response_SUCCESS))

				resp, err = serviceResource.Exist(getContext(), &pb.GetExistenceRequest{
					Type:      "schema",
					ServiceId: serviceId,
					SchemaId:  invalidSchemaId,
				})
				//Expect(err).To(BeNil())
				fmt.Println("UT=============" + resp.String())
				Expect(resp.GetResponse().Code).ToNot(Equal(pb.Response_SUCCESS))

				resp, err = serviceResource.Exist(getContext(), &pb.GetExistenceRequest{
					Type:        "microservice",
					ServiceId:   serviceId,
					ServiceName: TOO_LONG_SERVICENAME,
					Version:     "2.0.0",
					AppId:       "default",
				})
				Expect(err).To(BeNil())
				fmt.Println("UT=============" + resp.String())
				Expect(resp.GetResponse().Code).ToNot(Equal(pb.Response_SUCCESS))
			})

			It("exist schema", func() {
				respCreateService, err := serviceResource.Create(getContext(), &pb.CreateServiceRequest{
					Service: &pb.MicroService{
						ServiceName: "exist_schema_service_name",
						AppId:       "exist_schema_appId",
						Version:     "1.0.0",
						Level:       "FRONT",
						Schemas: []string{
							"first_schemaId",
						},
						Status: "UP",
					},
				})
				Expect(err).To(BeNil())
				Expect(respCreateService.GetResponse().Code).To(Equal(pb.Response_SUCCESS))
				serviceId := respCreateService.ServiceId

				respCreateSchema, err := serviceResource.ModifySchema(getContext(), &pb.ModifySchemaRequest{
					ServiceId: serviceId,
					SchemaId:  "first_schemaId",
					Schema:    "first_schema",
				})
				Expect(err).To(BeNil())
				Expect(respCreateSchema.GetResponse().Code).To(Equal(pb.Response_SUCCESS))

				respSchemaExist, err := serviceResource.Exist(getContext(), &pb.GetExistenceRequest{
					Type:      "schema",
					ServiceId: serviceId,
					SchemaId:  "first_schemaId",
				})
				Expect(err).To(BeNil())
				Expect(respSchemaExist.GetResponse().Code).To(Equal(pb.Response_SUCCESS))
				Expect(respSchemaExist.Summary).To(Equal(""))

				respDeleteService, err := serviceResource.Delete(getContext(), &pb.DeleteServiceRequest{
					ServiceId: serviceId,
					Force:     true,
				})
				Expect(err).To(BeNil())
				Expect(respDeleteService.GetResponse().Code).To(Equal(pb.Response_SUCCESS))
			})

			It("微服务是否存在", func() {
				resp, err := serviceResource.Exist(getContext(), &pb.GetExistenceRequest{
					Type:        "microservice",
					AppId:       "default",
					ServiceName: "some-relay",
					Version:     "1.0.0",
				})
				Expect(err).To(BeNil())
				fmt.Println("UT=============" + resp.String())
				Expect(resp.ServiceId).To(Equal(serviceId))

				resp, err = serviceResource.Exist(getContext(), &pb.GetExistenceRequest{
					Type:        "microservice",
					AppId:       "default",
					ServiceName: "sr",
					Version:     "1.0.0",
				})
				Expect(err).To(BeNil())
				fmt.Println("UT=============" + resp.String())
				Expect(resp.ServiceId).To(Equal(serviceId))

				resp, err = serviceResource.Exist(getContext(), &pb.GetExistenceRequest{
					Type:        "microservice",
					AppId:       "default",
					ServiceName: "some-relay",
					Version:     "latest",
				})
				Expect(err).To(BeNil())
				fmt.Println("UT=============" + resp.String())
				Expect(resp.ServiceId).To(Equal(serviceId))

				resp, err = serviceResource.Exist(getContext(), &pb.GetExistenceRequest{
					Type:        "microservice",
					AppId:       "default",
					ServiceName: "some-relay",
					Version:     "1.0.0+",
				})
				Expect(err).To(BeNil())
				fmt.Println("UT=============" + resp.String())
				Expect(resp.ServiceId).To(Equal(serviceId))

				resp, err = serviceResource.Exist(getContext(), &pb.GetExistenceRequest{
					Type:        "microservice",
					AppId:       "default",
					ServiceName: "some-relay",
					Version:     "0.9.1-1.0.1",
				})
				Expect(err).To(BeNil())
				fmt.Println("UT=============" + resp.String())
				Expect(resp.ServiceId).To(Equal(serviceId))

				resp, err = serviceResource.Exist(getContext(), &pb.GetExistenceRequest{
					Type:        "microservice",
					AppId:       "default",
					ServiceName: "notExistService",
					Version:     "1.0.0",
				})
				Expect(err).To(BeNil())
				fmt.Println("UT=============" + resp.String())
				Expect(resp.GetResponse().Code).ToNot(Equal(pb.Response_SUCCESS))
			})

			It("创建微服务2", func() {
				r := &pb.CreateServiceRequest{
					Service: &pb.MicroService{
						ServiceName: "some-backend",
						AppId:       "default",
						Version:     "1.0.0",
						Level:       "BACK",
						Schemas: []string{
							"xxxxxxxx",
						},
						Properties: make(map[string]string),
						Status:     "UP",
					},
				}
				r.Service.Properties["project"] = "x"
				resp, err := serviceResource.Create(getContext(), r)
				Expect(err).To(BeNil())
				serviceId2 = resp.ServiceId
				fmt.Println("UT=========" + serviceId2)
				Expect(resp.GetResponse().Code).To(Equal(pb.Response_SUCCESS))
			})
			It("创建微服务3", func() {
				r := &pb.CreateServiceRequest{
					Service: &pb.MicroService{
						ServiceName: "some-project",
						AppId:       "default",
						Version:     "1.0.0",
						Level:       "BACK",
						Schemas: []string{
							"xxxxxxxx",
						},
						Properties: make(map[string]string),
						Status:     "UP",
					},
				}
				r.Service.Properties["project"] = "x"
				resp, err := serviceResource.Create(getContext(), r)
				Expect(err).To(BeNil())
				serviceId3 = resp.ServiceId
				fmt.Println("UT=========" + serviceId3)
				Expect(resp.GetResponse().Code).To(Equal(pb.Response_SUCCESS))
			})
			It("创建微服务4，字段验证", func() {
				r := &pb.CreateServiceRequest{
					Service: &pb.MicroService{
						ServiceName: "service-validate",
						Version:     "1.0.0",
						Level:       "BACK",
						Status:      "UP",
					},
				}
				resp, err := serviceResource.Create(getContext(), r)
				Expect(err).To(BeNil())
				fmt.Println("UT=========" + resp.ServiceId)
				Expect(resp.GetResponse().Code).ToNot(Equal(pb.Response_SUCCESS))

				r = &pb.CreateServiceRequest{
					Service: &pb.MicroService{
						AppId:   "default",
						Version: "1.0.0",
						Level:   "BACK",
						Status:  "UP",
					},
				}
				resp, err = serviceResource.Create(getContext(), r)
				Expect(err).To(BeNil())
				fmt.Println("UT=========" + resp.ServiceId)
				Expect(resp.GetResponse().Code).ToNot(Equal(pb.Response_SUCCESS))

				r = &pb.CreateServiceRequest{
					Service: &pb.MicroService{
						AppId:       "default",
						ServiceName: "service-validate",
						Level:       "BACK",
						Status:      "UP",
					},
				}
				resp, err = serviceResource.Create(getContext(), r)
				Expect(err).To(BeNil())
				fmt.Println("UT=========" + resp.ServiceId)
				Expect(resp.GetResponse().Code).ToNot(Equal(pb.Response_SUCCESS))

				r = &pb.CreateServiceRequest{
					Service: &pb.MicroService{
						AppId:       "default",
						ServiceName: "service-validate",
						Version:     "1.0.0",
						Level:       "INVALID",
						Status:      "UP",
					},
				}
				resp, err = serviceResource.Create(getContext(), r)
				Expect(err).To(BeNil())
				fmt.Println("UT=========" + resp.ServiceId)
				Expect(resp.GetResponse().Code).ToNot(Equal(pb.Response_SUCCESS))

				r = &pb.CreateServiceRequest{
					Service: &pb.MicroService{
						AppId:       "default",
						ServiceName: "service-validate",
						Version:     "1.0.0",
						Level:       "BACK",
					},
				}
				resp, err = serviceResource.Create(getContext(), r)
				Expect(err).To(BeNil())
				fmt.Println("UT=========" + resp.ServiceId)
				Expect(resp.GetResponse().Code).ToNot(Equal(pb.Response_SUCCESS))
			})

			It("create service, schema param check", func() {
				size := buildin.SCHEMA_NUM_MAX_FOR_ONESERVICE + 1
				schemas := make([]string, size)
				for i := 0; i < size; i++ {
					schemas = append(schemas, strconv.Itoa(i))
				}
				respServiceForSchema, _ := serviceResource.Create(getContext(), &pb.CreateServiceRequest{
					Service: &pb.MicroService{
						ServiceName: "service_name",
						AppId:       "service_group",
						Version:     "1.0.0",
						Level:       "FRONT",
						Schemas:     schemas,
						Status:      "UP",
					},
				})
				Expect(respServiceForSchema.GetResponse().Code).ToNot(Equal(pb.Response_SUCCESS))
			})

			It("创建微服务5,参数校验", func() {
				fmt.Println("UT=========" + "参数为空")
				r := &pb.CreateServiceRequest{
					Service: nil,
				}
				resp, err := serviceResource.Create(getContext(), r)
				Expect(err).To(BeNil())
				Expect(resp.GetResponse().Code).ToNot(Equal(pb.Response_SUCCESS))

				fmt.Println("UT=========" + "Alias 非法")
				r = &pb.CreateServiceRequest{
					Service: &pb.MicroService{
						AppId:       "default",
						ServiceName: "service-validate-alias",
						Alias:       "@$#%",
						Version:     "1.0.0",
						Level:       "BACK",
						Status:      "UP",
					},
				}
				resp, err = serviceResource.Create(getContext(), r)
				Expect(err).To(BeNil())
				Expect(resp.GetResponse().Code).ToNot(Equal(pb.Response_SUCCESS))

				fmt.Println("UT=========" + "Alias 合法")
				r = &pb.CreateServiceRequest{
					Service: &pb.MicroService{
						AppId:       "default",
						ServiceName: "service-validate-alias",
						Alias:       "aA:-_.1",
						Version:     "1.0.0",
						Level:       "BACK",
						Status:      "UP",
					},
				}
				resp, err = serviceResource.Create(getContext(), r)
				Expect(err).To(BeNil())
				Expect(resp.GetResponse().Code).To(Equal(pb.Response_SUCCESS))
			})

			It("查询所有服务,参数校验", func() {
				resp, err := serviceResource.GetServices(getContext(), nil)
				Expect(err).To(BeNil())
				fmt.Println(fmt.Sprintf("UT=============%s", resp.Services))
				Expect(resp.Response.Code).ToNot(Equal(pb.Response_SUCCESS))
			})

			It("查询所有服务", func() {

				resp, err := serviceResource.GetServices(getContext(), &pb.GetServicesRequest{})
				Expect(err).To(BeNil())
				fmt.Println(fmt.Sprintf("UT=============%s", resp.Services))
				Expect(len(resp.Services)).To(Not(Equal(0)))
			})
			It("查询单个服务，参数校验", func() {
				resp, err := serviceResource.GetOne(getContext(), &pb.GetServiceRequest{
					ServiceId: "",
				})
				Expect(err).To(BeNil())
				Expect(resp.Response.Code).ToNot(Equal(pb.Response_SUCCESS))

				resp, err = serviceResource.GetOne(getContext(), &pb.GetServiceRequest{
					ServiceId: "notexistservice",
				})
				Expect(err).To(BeNil())
				Expect(resp.Response.Code).ToNot(Equal(pb.Response_SUCCESS))

				resp, err = serviceResource.GetOne(getContext(), &pb.GetServiceRequest{
					ServiceId: TOO_LONG_SERVICEID,
				})
				Expect(err).To(BeNil())
				Expect(resp.Response.Code).ToNot(Equal(pb.Response_SUCCESS))
			})
			It("更新微服务1props", func() {
				r := &pb.UpdateServicePropsRequest{
					ServiceId:  serviceId,
					Properties: make(map[string]string),
				}
				r2 := &pb.UpdateServicePropsRequest{
					ServiceId:  serviceId,
					Properties: make(map[string]string),
				}
				r.Properties["test"] = "1"
				r2.Properties["k"] = "v"
				resp, err := serviceResource.UpdateProperties(getContext(), r)
				if err != nil {
					fmt.Println("UT===========" + err.Error())
				}
				Expect(resp.Response.Code).To(Equal(pb.Response_SUCCESS))

				resp, err = serviceResource.UpdateProperties(getContext(), r2)
				if err != nil {
					fmt.Println("UT===========" + err.Error())
				}
				Expect(err).To(BeNil())
				fmt.Println(fmt.Sprintf("UT=============%s", resp.Response.Code))
				resp2, err := serviceResource.GetOne(getContext(), &pb.GetServiceRequest{
					ServiceId: serviceId,
				})
				Expect(err).To(BeNil())
				fmt.Println(fmt.Sprintf("UT=============%s", resp2.Service.ServiceId))
				Expect(resp2.Service.ServiceId).To(Equal(serviceId))
				Expect(resp2.Service.Properties["k"]).To(Equal("v"))

				r.ServiceId = "notexistservice"
				resp, err = serviceResource.UpdateProperties(getContext(), r)
				if err != nil {
					fmt.Println("UT===========" + err.Error())
				}
				Expect(err).To(BeNil())
				fmt.Println(fmt.Sprintf("UT=============%s", resp.Response.Code))
				Expect(resp.Response.Code).ToNot(Equal(pb.Response_SUCCESS))
			})
			It("更新微服务3props，空结构更新", func() {
				r := &pb.UpdateServicePropsRequest{
					ServiceId:  serviceId,
					Properties: nil,
				}
				resp, err := serviceResource.UpdateProperties(getContext(), r)
				if err != nil {
					fmt.Println("UT===========" + err.Error())
				}
				Expect(err).To(BeNil())
				fmt.Println(fmt.Sprintf("UT=============%s", resp.Response.Code))
				Expect(resp.Response.Code).ToNot(Equal(pb.Response_SUCCESS))

				r = &pb.UpdateServicePropsRequest{
					ServiceId:  "",
					Properties: map[string]string{},
				}
				resp, err = serviceResource.UpdateProperties(getContext(), r)
				Expect(err).To(BeNil())
				fmt.Println(fmt.Sprintf("UT=============%s", resp.Response.Code))
				Expect(resp.Response.Code).ToNot(Equal(pb.Response_SUCCESS))
			})

			It("更新微服务3props, 参数校验", func() {
				r := &pb.UpdateServicePropsRequest{
					ServiceId:  TOO_LONG_SERVICEID,
					Properties: map[string]string{},
				}
				resp, _ := serviceResource.UpdateProperties(getContext(), r)
				Expect(resp.Response.Code).ToNot(Equal(pb.Response_SUCCESS))
			})

			It("创建微服务1实例", func() {
				resp, err := insResource.Register(getContext(), &pb.RegisterInstanceRequest{
					Instance: &pb.MicroServiceInstance{
						ServiceId: serviceId,
						Endpoints: []string{
							"rest:127.0.0.1:8080",
						},
						HostName:    "UT-HOST",
						Status:      pb.MSI_UP,
						Environment: "production",
					},
				})
				Expect(err).To(BeNil())
				fmt.Println("UT============" + resp.GetResponse().Message)
				fmt.Println("UT============" + resp.InstanceId)
				Expect(resp.GetResponse().Code).To(Equal(pb.Response_SUCCESS))
				Expect(resp.InstanceId).To(Not(Equal("")))
			})
			It("没添加黑名单 尝试访问", func() {
				resp, err := insResource.GetInstances(getContext(), &pb.GetInstancesRequest{
					ConsumerServiceId: serviceId3,
					ProviderServiceId: serviceId,
					Env:               "production",
				})
				Expect(err).To(BeNil())
				fmt.Println("UT============" + resp.GetResponse().Message)
				fmt.Println(fmt.Sprintf("UT============%s"), resp.Instances)
				Expect(resp.GetResponse().Code).To(Equal(pb.Response_SUCCESS))
				Expect(len(resp.Instances)).To(Not(Equal(0)))
			})
			It("添加黑名单 尝试访问", func() {
				req := &pb.AddServiceRulesRequest{
					ServiceId: serviceId,
				}
				rules := append(req.Rules, &pb.AddOrUpdateServiceRule{
					RuleType:  "BLACK",
					Pattern:   ".*project",
					Attribute: "ServiceName",
				})
				resp, err := serviceResource.AddRule(getContext(), &pb.AddServiceRulesRequest{
					ServiceId: serviceId,
					Rules:     rules,
				})
				Expect(err).To(BeNil())
				fmt.Println("UT============" + resp.GetResponse().Message)
				fmt.Println(fmt.Sprintf("UT============%s"), resp.String())
				resp2, err := insResource.GetInstances(getContext(), &pb.GetInstancesRequest{
					ConsumerServiceId: serviceId3,
					ProviderServiceId: serviceId,
					NoCache:           true,
				})
				Expect(err).To(BeNil())
				fmt.Println("UT============" + resp2.GetResponse().Message)
				fmt.Println(fmt.Sprintf("UT============%s"), resp2.Instances)
				Expect(resp2.GetResponse().Code).ToNot(Equal(pb.Response_SUCCESS))
				Expect(len(resp2.Instances)).To((Equal(0)))
			})
			It("添加白名单 尝试访问", func() {
				req := &pb.AddServiceRulesRequest{
					ServiceId: serviceId3,
				}
				rules := append(req.Rules, &pb.AddOrUpdateServiceRule{
					RuleType:  "WHITE",
					Pattern:   ".*relay",
					Attribute: "ServiceName",
				})
				resp, err := serviceResource.AddRule(getContext(), &pb.AddServiceRulesRequest{
					ServiceId: serviceId3,
					Rules:     rules,
				})
				Expect(err).To(BeNil())
				fmt.Println("UT============" + resp.GetResponse().Message)
				fmt.Println(fmt.Sprintf("UT============%s"), resp.String())
				Expect(resp.GetResponse().Code).To(Equal(pb.Response_SUCCESS))
				resp2, err := insResource.GetInstances(getContext(), &pb.GetInstancesRequest{
					ConsumerServiceId: serviceId,
					ProviderServiceId: serviceId3,
				})
				Expect(err).To(BeNil())
				fmt.Println("UT============" + resp2.GetResponse().Message)
				fmt.Println(fmt.Sprintf("UT============%s"), resp2.Instances)
				Expect(resp2.GetResponse().Code).To(Equal(pb.Response_SUCCESS))
				Expect(len(resp2.Instances)).To((Equal(0)))
			})

			It("添加tag 尝试访问", func() {
				resp1, err := serviceResource.AddTags(getContext(), &pb.AddServiceTagsRequest{
					ServiceId: serviceId3,
					Tags:      map[string]string{"a": "b"},
				})
				Expect(err).To(BeNil())
				fmt.Println("UT============" + resp1.GetResponse().Message)
				fmt.Println(fmt.Sprintf("UT============%s"), resp1.Response)
				Expect(resp1.GetResponse().Code).To(Equal(pb.Response_SUCCESS))

				resp2, err := insResource.GetInstances(getContext(), &pb.GetInstancesRequest{
					ConsumerServiceId: serviceId,
					ProviderServiceId: serviceId3,
					Tags:              []string{"a"},
					NoCache:           true,
				})
				Expect(err).To(BeNil())
				fmt.Println("UT============" + resp2.GetResponse().Message)
				fmt.Println(fmt.Sprintf("UT============%s"), resp2.Instances)
				Expect(resp2.GetResponse().Code).To(Equal(pb.Response_SUCCESS))
				Expect(len(resp2.Instances)).To((Equal(0)))
			})

			It("添加tag白名单 尝试访问", func() {
				req := &pb.AddServiceRulesRequest{
					ServiceId: serviceId3,
				}
				rules := append(req.Rules, &pb.AddOrUpdateServiceRule{
					RuleType:  "WHITE",
					Pattern:   ".*relay",
					Attribute: "tag_ServiceName",
				})
				resp, err := serviceResource.AddRule(getContext(), &pb.AddServiceRulesRequest{
					ServiceId: serviceId3,
					Rules:     rules,
				})
				Expect(err).To(BeNil())
				fmt.Println("UT============" + resp.GetResponse().Message)
				fmt.Println(fmt.Sprintf("UT============%s"), resp.String())
				Expect(resp.GetResponse().Code).To(Equal(pb.Response_SUCCESS))
				resp2, err := insResource.GetInstances(getContext(), &pb.GetInstancesRequest{
					ConsumerServiceId: serviceId2,
					ProviderServiceId: serviceId3,
					NoCache:           true,
				})
				Expect(err).To(BeNil())
				fmt.Println("UT============" + resp2.GetResponse().Message)
				fmt.Println(fmt.Sprintf("UT============%s"), resp2.Instances)
				Expect(resp2.GetResponse().Code).ToNot(Equal(pb.Response_SUCCESS))
				Expect(len(resp2.Instances)).To((Equal(0)))
			})
		})
	})

	Describe("Delete", func() {
		Context("normal", func() {
			It("删除微服务,参数校验", func() {
				resp, err := serviceResource.Delete(getContext(), &pb.DeleteServiceRequest{
					ServiceId: "",
					Force:     true,
				})
				Expect(err).To(BeNil())
				Expect(resp.GetResponse().Code).ToNot(Equal(pb.Response_SUCCESS))

				resp, err = serviceResource.Delete(getContext(), &pb.DeleteServiceRequest{
					ServiceId: TOO_LONG_SERVICEID,
					Force:     true,
				})
				Expect(err).To(BeNil())
				Expect(resp.GetResponse().Code).ToNot(Equal(pb.Response_SUCCESS))

				resp, err = serviceResource.Delete(getContext(), &pb.DeleteServiceRequest{
					ServiceId: "notexistservice",
					Force:     true,
				})
				Expect(err).To(BeNil())
				Expect(resp.GetResponse().Code).ToNot(Equal(pb.Response_SUCCESS))
			})

			It("删除微服务1", func() {
				resp, err := serviceResource.Delete(getContext(), &pb.DeleteServiceRequest{
					ServiceId: serviceId,
					Force:     true,
				})
				Expect(err).To(BeNil())
				Expect(resp.GetResponse().Code).To(Equal(pb.Response_SUCCESS))
			})
			It("删除微服务2", func() {
				resp, err := serviceResource.Delete(getContext(), &pb.DeleteServiceRequest{
					ServiceId: serviceId2,
					Force:     true,
				})
				Expect(err).To(BeNil())
				Expect(resp.GetResponse().Code).To(Equal(pb.Response_SUCCESS))
			})
			It("删除微服务3", func() {
				resp, err := serviceResource.Delete(getContext(), &pb.DeleteServiceRequest{
					ServiceId: serviceId3,
					Force:     false,
				})
				Expect(err).To(BeNil())
				Expect(resp.GetResponse().Code).To(Equal(pb.Response_SUCCESS))
			})

			It("删除微服务4，删除SC自身", func() {
				respExist, err := serviceResource.Exist(getContext(), &pb.GetExistenceRequest{
					Type:        "microservice",
					AppId:       core.Service.AppId,
					ServiceName: core.Service.ServiceName,
					Version:     core.Service.Version,
				})
				Expect(err).To(BeNil())
				core.Service.ServiceId = respExist.ServiceId
				fmt.Println("UT=================serviceId is ", core.Service.ServiceId)
				resp, err := serviceResource.Delete(getContext(), &pb.DeleteServiceRequest{
					ServiceId: core.Service.ServiceId,
					Force:     true,
				})
				fmt.Println("UT============" + resp.GetResponse().Message)
				Expect(err).To(BeNil())
				Expect(resp.GetResponse().Code).ToNot(Equal(pb.Response_SUCCESS))
			})
		})
	})

	Describe("DeleteServices", func() {
		Context("param check", func() {
			It("param check", func() {
				resp, err := serviceResource.DeleteServices(getContext(), &pb.DelServicesRequest{
					ServiceIds: []string{},
					Force:      false,
				})
				Expect(err).To(BeNil())
				Expect(resp.GetResponse().Code).ToNot(Equal(pb.Response_SUCCESS))

				resp, err = serviceResource.DeleteServices(getContext(), &pb.DelServicesRequest{
					ServiceIds: []string{"@#$44332_non-invalid_serviceId"},
					Force:      false,
				})
				Expect(err).To(BeNil())
				Expect(resp.GetResponse().Code).ToNot(Equal(pb.Response_SUCCESS))
			})

		})
		Context("normal", func() {
			var serviceId3 string
			var serviceId4 string
			var serviceId5 string
			var serviceId6 string
			var instanceId6 string
			// 创建服务4，服务5，服务6，其中服务6创建了实例关系
			It("批量删除服务，创建依赖的服务4", func() {
				resp, err := serviceResource.Create(getContext(), &pb.CreateServiceRequest{
					Service: &pb.MicroService{
						ServiceName: "test-services-03",
						Alias:       "ts_03",
						AppId:       "default_03",
						Version:     "1.0.0",
						Level:       "FRONT",
						Schemas: []string{
							"xxxxxxxxxx",
						},
						Status: "UP",
					},
				})
				Expect(err).To(BeNil())
				serviceId3 = resp.ServiceId
				fmt.Println("UT=========ServiceId" + serviceId3)
				Expect(resp.GetResponse().Code).To(Equal(pb.Response_SUCCESS))
			})

			It("批量删除服务，创建依赖的服务4", func() {
				resp, err := serviceResource.Create(getContext(), &pb.CreateServiceRequest{
					Service: &pb.MicroService{
						ServiceName: "test-services-04",
						Alias:       "ts_04",
						AppId:       "default_04",
						Version:     "1.0.0",
						Level:       "FRONT",
						Schemas: []string{
							"xxxxxxxxxx",
						},
						Status: "UP",
					},
				})
				Expect(err).To(BeNil())
				serviceId4 = resp.ServiceId
				fmt.Println("UT=========ServiceId" + serviceId4)
				Expect(resp.GetResponse().Code).To(Equal(pb.Response_SUCCESS))
			})
			It("批量删除服务，创建依赖的服务5", func() {
				resp, err := serviceResource.Create(getContext(), &pb.CreateServiceRequest{
					Service: &pb.MicroService{
						ServiceName: "test-services-05",
						Alias:       "ts_05",
						AppId:       "default_05",
						Version:     "1.0.0",
						Level:       "FRONT",
						Schemas: []string{
							"xxxxxxxx11",
						},
						Status: "UP",
					},
				})
				Expect(err).To(BeNil())
				serviceId5 = resp.ServiceId
				fmt.Println("UT=========ServiceId5 " + serviceId5)
				fmt.Printf("TEST CREATE service %d", resp.GetResponse().Code)
				Expect(resp.GetResponse().Code).To(Equal(pb.Response_SUCCESS))
			})
			It("批量删除服务，创建依赖的服务6", func() {
				resp, err := serviceResource.Create(getContext(), &pb.CreateServiceRequest{
					Service: &pb.MicroService{
						ServiceName: "test-services-06",
						Alias:       "ts_06",
						AppId:       "default_06",
						Version:     "1.0.0",
						Level:       "FRONT",
						Schemas: []string{
							"xxxxxxxxxxxxx",
						},
						Status: "UP",
					},
				})
				Expect(err).To(BeNil())
				serviceId6 = resp.ServiceId
				fmt.Println("UT=========ServiceId" + serviceId6)
				Expect(resp.GetResponse().Code).To(Equal(pb.Response_SUCCESS))
			})

			It("批量删除服务，创建依赖的服务6", func() {
				respReg, err := insResource.Register(getContext(), &pb.RegisterInstanceRequest{
					Instance: &pb.MicroServiceInstance{
						ServiceId: serviceId6,
						Endpoints: []string{
							"rest:127.0.0.2:8081",
						},
						HostName:    "UT-HOST",
						Status:      pb.MSI_UP,
						Environment: "production",
					},
				})
				instanceId6 = respReg.InstanceId
				Expect(err).To(BeNil())
				fmt.Println("UT============" + respReg.GetResponse().Message)
				Expect(respReg.GetResponse().Code).To(Equal(pb.Response_SUCCESS))
			})

			It("批量删除服务3和服务4", func() {
				resp, err := serviceResource.DeleteServices(getContext(), &pb.DelServicesRequest{
					ServiceIds: []string{serviceId3, serviceId4},
					Force:      false,
				},
				)
				Expect(err).To(BeNil())
				fmt.Println("UT============" + resp.GetResponse().Message)
				Expect(resp.GetResponse().Code).To(Equal(pb.Response_SUCCESS))
			})

			It("批量删除服务5和服务6", func() {
				resp, err := serviceResource.DeleteServices(getContext(), &pb.DelServicesRequest{
					ServiceIds: []string{serviceId5, serviceId6},
					Force:      false,
				},
				)
				Expect(err).To(BeNil())
				//期待结果失败
				fmt.Println("UT============" + resp.GetResponse().Message)
				Expect(resp.GetResponse().Code).ToNot(Equal(pb.Response_SUCCESS))
			})

			It("删除服务6的实例，删除服务5和服务6", func() {

				respReg, err := insResource.Unregister(getContext(), &pb.UnregisterInstanceRequest{
					ServiceId:  serviceId6,
					InstanceId: instanceId6,
				})

				Expect(err).To(BeNil())
				Expect(respReg.GetResponse().Code).To(Equal(pb.Response_SUCCESS))

				resp, err := serviceResource.DeleteServices(getContext(), &pb.DelServicesRequest{
					ServiceIds: []string{serviceId6},
					Force:      false,
				},
				)
				Expect(err).To(BeNil())
				//期待结果失败
				fmt.Println("UT============" + resp.GetResponse().Message)
				Expect(resp.GetResponse().Code).To(Equal(pb.Response_SUCCESS))
			})

		})
	})
})
