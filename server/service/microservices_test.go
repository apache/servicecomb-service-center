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
	var serviceId string
	var serviceId2 string
	var serviceId3 string
	Describe("Create", func() {
		Context("normal", func() {
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
				Expect(resp.GetResponse().Code).To(Equal(pb.Response_FAIL))

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
				Expect(resp.GetResponse().Code).To(Equal(pb.Response_FAIL))

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
				Expect(resp.GetResponse().Code).To(Equal(pb.Response_FAIL))
			})

			It("微服务是否存在,参数校验", func() {
				resp, err := serviceResource.Exist(getContext(), &pb.GetExistenceRequest{
					Type:      "schema",
					ServiceId: "noneservice",
					SchemaId:  "xxxxxxxx",
				})
				Expect(err).To(BeNil())
				fmt.Println("UT=============" + resp.String())
				Expect(resp.GetResponse().Code).To(Equal(pb.Response_FAIL))

				resp, err = serviceResource.Exist(getContext(), &pb.GetExistenceRequest{
					Type:      "schema",
					ServiceId: serviceId,
					SchemaId:  "noneschema",
				})
				Expect(err).To(BeNil())
				fmt.Println("UT=============" + resp.String())
				Expect(resp.GetResponse().Code).To(Equal(pb.Response_FAIL))

				resp, err = serviceResource.Exist(getContext(), &pb.GetExistenceRequest{
					Type:      "nonetype",
					ServiceId: serviceId,
					SchemaId:  "noneschema",
				})
				Expect(err).To(BeNil())
				fmt.Println("UT=============" + resp.String())
				Expect(resp.GetResponse().Code).To(Equal(pb.Response_FAIL))

				resp, err = serviceResource.Exist(getContext(), &pb.GetExistenceRequest{
					Type:      "schema",
					ServiceId: serviceId,
					SchemaId:  invalidSchemaId,
				})
				//Expect(err).To(BeNil())
				fmt.Println("UT=============" + resp.String())
				Expect(resp.GetResponse().Code).To(Equal(pb.Response_FAIL))

				resp, err = serviceResource.Exist(getContext(), &pb.GetExistenceRequest{
					Type:        "microservice",
					ServiceId:   serviceId,
					ServiceName: TOO_LONG_SERVICENAME,
					Version:     "2.0.0",
					AppId:       "default",
				})
				Expect(err).To(BeNil())
				fmt.Println("UT=============" + resp.String())
				Expect(resp.GetResponse().Code).To(Equal(pb.Response_FAIL))
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
					Version:     "0.9.1-1.0.0",
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
				Expect(resp.GetResponse().Code).To(Equal(pb.Response_FAIL))
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
				Expect(resp.GetResponse().Code).To(Equal(pb.Response_FAIL))

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
				Expect(resp.GetResponse().Code).To(Equal(pb.Response_FAIL))

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
				Expect(resp.GetResponse().Code).To(Equal(pb.Response_FAIL))

				r = &pb.CreateServiceRequest{
					Service: &pb.MicroService{
						AppId:       "default",
						ServiceName: "service-validate",
						Version:     "1.0.0",
						Status:      "UP",
					},
				}
				resp, err = serviceResource.Create(getContext(), r)
				Expect(err).To(BeNil())
				fmt.Println("UT=========" + resp.ServiceId)
				Expect(resp.GetResponse().Code).To(Equal(pb.Response_FAIL))

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
				Expect(resp.GetResponse().Code).To(Equal(pb.Response_FAIL))
			})

			It("创建微服务5,参数校验", func() {
				fmt.Println("UT=========" + "参数为空")
				r := &pb.CreateServiceRequest{
					Service: nil,
				}
				resp, err := serviceResource.Create(getContext(), r)
				Expect(err).To(BeNil())
				Expect(resp.GetResponse().Code).To(Equal(pb.Response_FAIL))

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
				Expect(resp.GetResponse().Code).To(Equal(pb.Response_FAIL))

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
				Expect(resp.Response.Code).To(Equal(pb.Response_FAIL))
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
				Expect(resp.Response.Code).To(Equal(pb.Response_FAIL))

				resp, err = serviceResource.GetOne(getContext(), &pb.GetServiceRequest{
					ServiceId: "notexistservice",
				})
				Expect(err).To(BeNil())
				Expect(resp.Response.Code).To(Equal(pb.Response_FAIL))

				resp, err = serviceResource.GetOne(getContext(), &pb.GetServiceRequest{
					ServiceId: TOO_LONG_SERVICEID,
				})
				Expect(err).To(BeNil())
				Expect(resp.Response.Code).To(Equal(pb.Response_FAIL))
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
				Expect(resp.Response.Code).To(Equal(pb.Response_FAIL))
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
				Expect(resp.Response.Code).To(Equal(pb.Response_FAIL))

				r = &pb.UpdateServicePropsRequest{
					ServiceId:  "",
					Properties: map[string]string{},
				}
				resp, err = serviceResource.UpdateProperties(getContext(), r)
				Expect(err).To(BeNil())
				fmt.Println(fmt.Sprintf("UT=============%s", resp.Response.Code))
				Expect(resp.Response.Code).To(Equal(pb.Response_FAIL))
			})

			It("更新微服务3props, 参数校验", func() {
				r := &pb.UpdateServicePropsRequest{
					ServiceId:  TOO_LONG_SERVICEID,
					Properties: map[string]string{},
				}
				resp, _ := serviceResource.UpdateProperties(getContext(), r)
				Expect(resp.Response.Code).To(Equal(pb.Response_FAIL))
			})

			It("创建微服务1实例", func() {
				resp, err := insResource.Register(getContext(), &pb.RegisterInstanceRequest{
					Instance: &pb.MicroServiceInstance{
						ServiceId: serviceId,
						Endpoints: []string{
							"rest:127.0.0.1:8080",
						},
						HostName: "UT-HOST",
						Status:   pb.MSI_UP,
						Stage:    "prod",
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
					Stage:             "prod",
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
				})
				Expect(err).To(BeNil())
				fmt.Println("UT============" + resp2.GetResponse().Message)
				fmt.Println(fmt.Sprintf("UT============%s"), resp2.Instances)
				Expect(resp2.GetResponse().Code).To(Equal(pb.Response_FAIL))
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
				})
				Expect(err).To(BeNil())
				fmt.Println("UT============" + resp2.GetResponse().Message)
				fmt.Println(fmt.Sprintf("UT============%s"), resp2.Instances)
				Expect(resp2.GetResponse().Code).To(Equal(pb.Response_FAIL))
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
				Expect(resp.GetResponse().Code).To(Equal(pb.Response_FAIL))

				resp, err = serviceResource.Delete(getContext(), &pb.DeleteServiceRequest{
					ServiceId: TOO_LONG_SERVICEID,
					Force:     true,
				})
				Expect(err).To(BeNil())
				Expect(resp.GetResponse().Code).To(Equal(pb.Response_FAIL))

				resp, err = serviceResource.Delete(getContext(), &pb.DeleteServiceRequest{
					ServiceId: "notexistservice",
					Force:     true,
				})
				Expect(err).To(BeNil())
				Expect(resp.GetResponse().Code).To(Equal(pb.Response_FAIL))
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
					AppId:       "default",
					ServiceName: "SERVICECENTER",
					Version:     "3.0.0",
				})
				Expect(err).To(BeNil())
				serviceId := respExist.ServiceId
				fmt.Println("UT=================serviceId is ", serviceId)
				resp, err := serviceResource.Delete(getContext(), &pb.DeleteServiceRequest{
					ServiceId: serviceId,
					Force:     true,
				})
				fmt.Println("UT============" + resp.GetResponse().Message)
				Expect(err).To(BeNil())
				Expect(resp.GetResponse().Code).To(Equal(pb.Response_FAIL))
			})
		})
	})
})
