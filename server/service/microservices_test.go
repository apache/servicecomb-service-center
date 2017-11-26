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
	"github.com/ServiceComb/service-center/server/core"
	pb "github.com/ServiceComb/service-center/server/core/proto"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"strings"
)

var (
	TOO_LONG_SERVICEID   = strings.Repeat("x", 65)
	TOO_LONG_SERVICENAME = strings.Repeat("x", 129)
)

var _ = Describe("'Micro-service' service", func() {
	Describe("execute 'create' operartion", func() {
		Context("when service is nil", func() {
			It("should not be passed", func() {
				resp, err := serviceResource.Create(getContext(), &pb.CreateServiceRequest{
					Service: nil,
				})
				Expect(err).To(BeNil())
				Expect(resp.GetResponse().Code).ToNot(Equal(pb.Response_SUCCESS))
			})
		})

		Context("when service with rules/tags/instances", func() {
			It("should be passed", func() {
				By("prepare data for creating")
				tags := make(map[string]string, 10)
				tags["test"] = "test"
				rules := []*pb.AddOrUpdateServiceRule{
					{
						RuleType:    "BLACK",
						Attribute:   "ServiceName",
						Pattern:     "test",
						Description: "test",
					},
				}
				instances := []*pb.MicroServiceInstance{
					{
						Endpoints: []string{
							"createService:127.0.0.1:8080",
						},
						HostName: "UT-HOST",
						Status:   pb.MSI_UP,
					},
				}

				By("first: create")
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

				By("second: create")
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

				By("check the tags/rules/instances")
				respGetTags, err := serviceResource.GetTags(getContext(), &pb.GetServiceTagsRequest{
					ServiceId: resp.ServiceId,
				})
				Expect(err).To(BeNil())
				Expect(respGetTags.GetResponse().Code).To(Equal(pb.Response_SUCCESS))
				Expect(respGetTags.Tags["second"]).To(Equal("second"))

				respGetRules, err := serviceResource.GetRule(getContext(), &pb.GetServiceRulesRequest{
					ServiceId: resp.ServiceId,
				})
				Expect(err).To(BeNil())
				Expect(respGetRules.GetResponse().Code).To(Equal(pb.Response_SUCCESS))
				Expect(respGetRules.Rules[0].Attribute).To(Equal("ServiceName"))

				respGetInsts, err := instanceResource.GetInstances(getContext(), &pb.GetInstancesRequest{
					ConsumerServiceId: resp.ServiceId,
					ProviderServiceId: resp.ServiceId,
				})
				Expect(err).To(BeNil())
				Expect(respGetInsts.GetResponse().Code).To(Equal(pb.Response_SUCCESS))
				Expect(respGetInsts.Instances[0].HostName).To(Equal("UT-HOST"))

				By("delete service")
				respDelete, err := serviceResource.Delete(getContext(), &pb.DeleteServiceRequest{
					ServiceId: resp.ServiceId,
					Force:     true,
				})
				Expect(err).To(BeNil())
				Expect(respDelete.GetResponse().Code).To(Equal(pb.Response_SUCCESS))

			})
		})

		Context("when creating the same service", func() {
			It("should be failed", func() {
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
				Expect(resp.GetResponse().Code).ToNot(Equal(pb.Response_SUCCESS))
			})
		})

		Context("when service with properties", func() {
			It("should be passed", func() {
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
				Expect(resp.GetResponse().Code).To(Equal(pb.Response_SUCCESS))
			})
		})

		Context("when service body is invalid", func() {
			It("should be failed", func() {
				By("invalid appId")
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
				Expect(resp.GetResponse().Code).ToNot(Equal(pb.Response_SUCCESS))

				By("serviceName is nil")
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
				Expect(resp.GetResponse().Code).ToNot(Equal(pb.Response_SUCCESS))

				By("invalid status")
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
				Expect(resp.GetResponse().Code).ToNot(Equal(pb.Response_SUCCESS))

				By("invalid version")
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
				Expect(resp.GetResponse().Code).ToNot(Equal(pb.Response_SUCCESS))

				By("invalid level")
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
				Expect(resp.GetResponse().Code).ToNot(Equal(pb.Response_SUCCESS))

				By("alias contains illegal char")
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

				By("valid alias")
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
		})
	})

	Describe("execute 'exists' operartion", func() {
		var (
			serviceId string
		)

		It("should be passed", func() {
			respCreateService, err := serviceResource.Create(getContext(), &pb.CreateServiceRequest{
				Service: &pb.MicroService{
					Alias:       "es",
					ServiceName: "exist_service",
					AppId:       "exist_appId",
					Version:     "1.0.0",
					Level:       "FRONT",
					Schemas: []string{
						"first_schemaId",
					},
					Status: "UP",
				},
			})
			Expect(err).To(BeNil())
			Expect(respCreateService.ServiceId).ToNot(Equal(""))
			serviceId = respCreateService.ServiceId
		})

		Context("when type is invalid", func() {
			It("should be failed", func() {
				resp, err := serviceResource.Exist(getContext(), &pb.GetExistenceRequest{
					Type: "nonetype",
				})
				Expect(err).To(BeNil())
				Expect(resp.GetResponse().Code).ToNot(Equal(pb.Response_SUCCESS))
			})
		})

		Context("when param is invalid", func() {
			It("should be failed", func() {
				By("serviceName is too long")
				resp, err := serviceResource.Exist(getContext(), &pb.GetExistenceRequest{
					Type:        "microservice",
					ServiceName: TOO_LONG_SERVICENAME,
					Version:     "1.0.0",
					AppId:       "exist_appId",
				})
				Expect(err).To(BeNil())
				Expect(resp.GetResponse().Code).ToNot(Equal(pb.Response_SUCCESS))

				By("serviceName is empty")
				respExist, err := serviceResource.Exist(getContext(), &pb.GetExistenceRequest{
					Type:        "microservice",
					AppId:       "default",
					ServiceName: "",
					Version:     "3.0.0",
				})
				Expect(err).To(BeNil())
				Expect(respExist.GetResponse().Code).ToNot(Equal(pb.Response_SUCCESS))
			})
		})

		Context("when service does not exist", func() {
			It("should be failed", func() {
				resp, err := serviceResource.Exist(getContext(), &pb.GetExistenceRequest{
					Type:        "microservice",
					AppId:       "exist_appId",
					ServiceName: "notExistService",
					Version:     "1.0.0",
				})
				Expect(err).To(BeNil())
				Expect(resp.GetResponse().Code).ToNot(Equal(pb.Response_SUCCESS))
			})
		})

		Context("when service exists", func() {
			It("should be passed", func() {
				By("search with serviceName")
				resp, err := serviceResource.Exist(getContext(), &pb.GetExistenceRequest{
					Type:        "microservice",
					AppId:       "exist_appId",
					ServiceName: "exist_service",
					Version:     "1.0.0",
				})
				Expect(err).To(BeNil())
				Expect(resp.ServiceId).To(Equal(serviceId))

				By("search with alias")
				resp, err = serviceResource.Exist(getContext(), &pb.GetExistenceRequest{
					Type:        "microservice",
					AppId:       "exist_appId",
					ServiceName: "es",
					Version:     "1.0.0",
				})
				Expect(err).To(BeNil())
				Expect(resp.ServiceId).To(Equal(serviceId))

				By("search with latest versionRule")
				resp, err = serviceResource.Exist(getContext(), &pb.GetExistenceRequest{
					Type:        "microservice",
					AppId:       "exist_appId",
					ServiceName: "exist_service",
					Version:     "latest",
				})
				Expect(err).To(BeNil())
				Expect(resp.ServiceId).To(Equal(serviceId))

				By("search with 1.0.0+ versionRule")
				resp, err = serviceResource.Exist(getContext(), &pb.GetExistenceRequest{
					Type:        "microservice",
					AppId:       "exist_appId",
					ServiceName: "exist_service",
					Version:     "1.0.0+",
				})
				Expect(err).To(BeNil())
				Expect(resp.ServiceId).To(Equal(serviceId))

				By("search with range versionRule")
				resp, err = serviceResource.Exist(getContext(), &pb.GetExistenceRequest{
					Type:        "microservice",
					AppId:       "exist_appId",
					ServiceName: "exist_service",
					Version:     "0.9.1-1.0.1",
				})
				Expect(err).To(BeNil())
				Expect(resp.ServiceId).To(Equal(serviceId))
			})
		})
	})

	Describe("execute 'query' operartion", func() {
		Context("when request is nil", func() {
			It("should be failed", func() {
				resp, err := serviceResource.GetServices(getContext(), nil)
				Expect(err).To(BeNil())
				Expect(resp.Response.Code).ToNot(Equal(pb.Response_SUCCESS))
			})
		})

		Context("when query all services", func() {
			It("should be larger than 0", func() {

				resp, err := serviceResource.GetServices(getContext(), &pb.GetServicesRequest{})
				Expect(err).To(BeNil())
				Expect(len(resp.Services)).To(Not(Equal(0)))
			})
		})

		Context("when query a not exist service", func() {
			It("should be failed", func() {
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
		})
	})

	Describe("execute 'update' operartion", func() {
		var (
			serviceId string
		)

		It("should be passed", func() {
			respCreateService, err := serviceResource.Create(getContext(), &pb.CreateServiceRequest{
				Service: &pb.MicroService{
					Alias:       "es",
					ServiceName: "update_prop_service",
					AppId:       "update_prop_appId",
					Version:     "1.0.0",
					Level:       "FRONT",
					Status:      "UP",
				},
			})
			Expect(err).To(BeNil())
			Expect(respCreateService.ServiceId).ToNot(Equal(""))
			serviceId = respCreateService.ServiceId
		})

		Context("when property is not nil", func() {
			It("should be passed", func() {
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
				Expect(resp.Response.Code).To(Equal(pb.Response_SUCCESS))

				resp, err = serviceResource.UpdateProperties(getContext(), r2)
				Expect(err).To(BeNil())

				resp2, err := serviceResource.GetOne(getContext(), &pb.GetServiceRequest{
					ServiceId: serviceId,
				})
				Expect(err).To(BeNil())
				Expect(resp2.Service.ServiceId).To(Equal(serviceId))
				Expect(resp2.Service.Properties["test"]).To(Equal(""))
				Expect(resp2.Service.Properties["k"]).To(Equal("v"))
			})
		})

		Context("when service does not exist", func() {
			It("should be failed", func() {
				r := &pb.UpdateServicePropsRequest{
					ServiceId:  "notexistservice",
					Properties: make(map[string]string),
				}
				resp, err := serviceResource.UpdateProperties(getContext(), r)
				if err != nil {
				}
				Expect(err).To(BeNil())
				Expect(resp.Response.Code).ToNot(Equal(pb.Response_SUCCESS))
			})
		})

		Context("when property is nil or empty", func() {
			It("should be failed", func() {
				r := &pb.UpdateServicePropsRequest{
					ServiceId:  serviceId,
					Properties: nil,
				}
				resp, err := serviceResource.UpdateProperties(getContext(), r)
				Expect(err).To(BeNil())
				Expect(resp.Response.Code).ToNot(Equal(pb.Response_SUCCESS))

				r = &pb.UpdateServicePropsRequest{
					ServiceId:  "",
					Properties: map[string]string{},
				}
				resp, err = serviceResource.UpdateProperties(getContext(), r)
				Expect(err).To(BeNil())
				Expect(resp.Response.Code).ToNot(Equal(pb.Response_SUCCESS))
			})
		})

		Context("when request is invalid", func() {
			It("should be failed", func() {
				r := &pb.UpdateServicePropsRequest{
					ServiceId:  TOO_LONG_SERVICEID,
					Properties: map[string]string{},
				}
				resp, _ := serviceResource.UpdateProperties(getContext(), r)
				Expect(resp.Response.Code).ToNot(Equal(pb.Response_SUCCESS))
			})
		})
	})

	Describe("execute 'delete' operartion", func() {
		var (
			serviceContainInstId string
			serviceNoInstId      string
			serviceConsumerId    string
		)

		It("should be passed", func() {
			instances := []*pb.MicroServiceInstance{
				{
					Endpoints: []string{
						"deleteService:127.0.0.1:8080",
					},
					HostName: "delete-host",
					Status:   pb.MSI_UP,
				},
			}

			respCreate, err := serviceResource.Create(getContext(), &pb.CreateServiceRequest{
				Service: &pb.MicroService{
					ServiceName: "delete_service_with_inst",
					AppId:       "delete_service",
					Version:     "1.0.0",
					Level:       "FRONT",
					Status:      "UP",
				},
				Instances: instances,
			})
			Expect(err).To(BeNil())
			Expect(respCreate.GetResponse().Code).To(Equal(pb.Response_SUCCESS))
			serviceContainInstId = respCreate.ServiceId

			provider := &pb.MicroService{
				ServiceName: "delete_serivce_no_inst",
				AppId:       "delete_service",
				Version:     "1.0.0",
				Level:       "FRONT",
				Status:      "UP",
			}
			respCreate, err = serviceResource.Create(getContext(), &pb.CreateServiceRequest{
				Service: provider,
			})
			Expect(err).To(BeNil())
			Expect(respCreate.GetResponse().Code).To(Equal(pb.Response_SUCCESS))
			serviceNoInstId = respCreate.ServiceId

			respCreate, err = serviceResource.Create(getContext(), &pb.CreateServiceRequest{
				Service: &pb.MicroService{
					ServiceName: "delete_serivce_consuemr",
					AppId:       "delete_service",
					Version:     "1.0.0",
					Level:       "FRONT",
					Status:      "UP",
				},
			})
			Expect(err).To(BeNil())
			Expect(respCreate.GetResponse().Code).To(Equal(pb.Response_SUCCESS))
			serviceConsumerId = respCreate.ServiceId

			respFind, err := instanceResource.Find(getContext(), &pb.FindInstancesRequest{
				ConsumerServiceId: serviceConsumerId,
				AppId:             provider.AppId,
				ServiceName:       provider.ServiceName,
				VersionRule:       provider.Version,
			})
			Expect(err).To(BeNil())
			Expect(respFind.GetResponse().Code).To(Equal(pb.Response_SUCCESS))
		})

		Context("when request is invalid", func() {
			It("should be failed", func() {
				By("serviceId is empty")
				resp, err := serviceResource.Delete(getContext(), &pb.DeleteServiceRequest{
					ServiceId: "",
					Force:     true,
				})
				Expect(err).To(BeNil())
				Expect(resp.GetResponse().Code).ToNot(Equal(pb.Response_SUCCESS))

				By("serviceId is invalid")
				resp, err = serviceResource.Delete(getContext(), &pb.DeleteServiceRequest{
					ServiceId: TOO_LONG_SERVICEID,
					Force:     true,
				})
				Expect(err).To(BeNil())
				Expect(resp.GetResponse().Code).ToNot(Equal(pb.Response_SUCCESS))

				By("serviceId does not exist")
				resp, err = serviceResource.Delete(getContext(), &pb.DeleteServiceRequest{
					ServiceId: "notexistservice",
					Force:     true,
				})
				Expect(err).To(BeNil())
				Expect(resp.GetResponse().Code).ToNot(Equal(pb.Response_SUCCESS))
			})
		})

		Context("when delete sc service", func() {
			It("should be not allowed", func() {
				respExist, err := serviceResource.Exist(getContext(), &pb.GetExistenceRequest{
					Type:        "microservice",
					AppId:       core.Service.AppId,
					ServiceName: core.Service.ServiceName,
					Version:     core.Service.Version,
				})
				Expect(err).To(BeNil())
				core.Service.ServiceId = respExist.ServiceId
				resp, err := serviceResource.Delete(getContext(), &pb.DeleteServiceRequest{
					ServiceId: core.Service.ServiceId,
					Force:     true,
				})
				Expect(err).To(BeNil())
				Expect(resp.GetResponse().Code).ToNot(Equal(pb.Response_SUCCESS))
			})
		})

		Context("when delete a service contains instances with not force flag", func() {
			It("should be not allowed", func() {
				resp, err := serviceResource.Delete(getContext(), &pb.DeleteServiceRequest{
					ServiceId: serviceContainInstId,
					Force:     false,
				})
				Expect(err).To(BeNil())
				Expect(resp.GetResponse().Code).ToNot(Equal(pb.Response_SUCCESS))
			})
		})

		Context("when delete a service contains instances with force flag", func() {
			It("should be passed", func() {
				resp, err := serviceResource.Delete(getContext(), &pb.DeleteServiceRequest{
					ServiceId: serviceContainInstId,
					Force:     true,
				})
				Expect(err).To(BeNil())
				Expect(resp.GetResponse().Code).To(Equal(pb.Response_SUCCESS))
			})
		})

		Context("when delete a service depended on consumer with not force flag", func() {
			It("should be not allowed", func() {
				resp, err := serviceResource.Delete(getContext(), &pb.DeleteServiceRequest{
					ServiceId: serviceNoInstId,
					Force:     false,
				})
				Expect(err).To(BeNil())
				Expect(resp.GetResponse().Code).ToNot(Equal(pb.Response_SUCCESS))
			})
		})

		Context("when delete a service depended on consumer with force flag", func() {
			It("should be passed", func() {
				resp, err := serviceResource.Delete(getContext(), &pb.DeleteServiceRequest{
					ServiceId: serviceNoInstId,
					Force:     true,
				})
				Expect(err).To(BeNil())
				Expect(resp.GetResponse().Code).To(Equal(pb.Response_SUCCESS))
			})
		})

		Context("when delete a service with not force flag", func() {
			It("should be not allowed", func() {
				resp, err := serviceResource.Delete(getContext(), &pb.DeleteServiceRequest{
					ServiceId: serviceNoInstId,
					Force:     false,
				})
				Expect(err).To(BeNil())
				Expect(resp.GetResponse().Code).ToNot(Equal(pb.Response_SUCCESS))
			})
		})

		Context("when batch delete services", func() {
			It("should be OK", func() {
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

		Context("batch delete services", func() {
			var (
				serviceId1 string
				serviceId2 string
			)

			It("should be passed", func() {
				resp, err := serviceResource.Create(getContext(), &pb.CreateServiceRequest{
					Service: &pb.MicroService{
						ServiceName: "batch_delete_services_1",
						AppId:       "batch_delete",
						Version:     "1.0.0",
						Level:       "FRONT",
						Status:      "UP",
					},
				})
				Expect(err).To(BeNil())
				Expect(resp.GetResponse().Code).To(Equal(pb.Response_SUCCESS))
				serviceId1 = resp.ServiceId

				resp, err = serviceResource.Create(getContext(), &pb.CreateServiceRequest{
					Service: &pb.MicroService{
						ServiceName: "batch_delete_services_2",
						AppId:       "batch_delete",
						Version:     "1.0.0",
						Level:       "FRONT",
						Status:      "UP",
					},
				})
				Expect(err).To(BeNil())
				Expect(resp.GetResponse().Code).To(Equal(pb.Response_SUCCESS))
				serviceId2 = resp.ServiceId
			})

			It("should be passed", func() {
				resp, err := serviceResource.DeleteServices(getContext(), &pb.DelServicesRequest{
					ServiceIds: []string{serviceId1, serviceId2},
					Force:      false,
				},
				)
				Expect(err).To(BeNil())
				Expect(resp.GetResponse().Code).To(Equal(pb.Response_SUCCESS))
			})

		})

		Context("batch delete services contain instances", func() {
			var (
				serviceIdFailed1 string
				serviceIdFailed2 string
			)

			It("should be passed", func() {
				resp, err := serviceResource.Create(getContext(), &pb.CreateServiceRequest{
					Service: &pb.MicroService{
						ServiceName: "batch_delete_services_failed_1",
						AppId:       "batch_delete",
						Version:     "1.0.0",
						Level:       "FRONT",
						Status:      "UP",
					},
				})
				Expect(err).To(BeNil())
				serviceIdFailed1 = resp.ServiceId
				Expect(resp.GetResponse().Code).To(Equal(pb.Response_SUCCESS))

				instances := []*pb.MicroServiceInstance{
					{
						Endpoints: []string{
							"batchDeleteServices:127.0.0.2:8081",
						},
						HostName: "batch-delete-host",
						Status:   pb.MSI_UP,
					},
				}

				resp, err = serviceResource.Create(getContext(), &pb.CreateServiceRequest{
					Service: &pb.MicroService{
						ServiceName: "batch_delete_services_failed_2",
						AppId:       "batch_delete",
						Version:     "1.0.0",
						Level:       "FRONT",
						Status:      "UP",
					},
					Instances: instances,
				})
				Expect(err).To(BeNil())
				serviceIdFailed2 = resp.ServiceId
				Expect(resp.GetResponse().Code).To(Equal(pb.Response_SUCCESS))
			})

			It("should be failed", func() {
				resp, err := serviceResource.DeleteServices(getContext(), &pb.DelServicesRequest{
					ServiceIds: []string{serviceIdFailed1, serviceIdFailed2},
					Force:      false,
				},
				)
				Expect(err).To(BeNil())
				//期待结果失败
				Expect(resp.GetResponse().Code).ToNot(Equal(pb.Response_SUCCESS))
			})
		})
	})
})
