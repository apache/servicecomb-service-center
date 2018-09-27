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
package service_test

import (
	"github.com/apache/incubator-servicecomb-service-center/server/core"
	pb "github.com/apache/incubator-servicecomb-service-center/server/core/proto"
	scerr "github.com/apache/incubator-servicecomb-service-center/server/error"
	"github.com/apache/incubator-servicecomb-service-center/server/plugin/pkg/quota"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"strconv"
	"strings"
)

var (
	TOO_LONG_APPID       = strings.Repeat("x", 161)
	TOO_LONG_SCHEMAID    = strings.Repeat("x", 161)
	TOO_LONG_SERVICEID   = strings.Repeat("x", 65)
	TOO_LONG_SERVICENAME = strings.Repeat("x", 129)
	TOO_LONG_EXISTENCE   = strings.Repeat("x", 128+160+2)
	TOO_LONG_ALIAS       = strings.Repeat("x", 129)
	TOO_LONG_FRAMEWORK   = strings.Repeat("x", 65)
	TOO_LONG_DESCRIPTION = strings.Repeat("x", 257)
)

var _ = Describe("'Micro-service' service", func() {
	Describe("execute 'create' operartion", func() {
		Context("when service is nil", func() {
			It("should not be passed", func() {
				resp, err := serviceResource.Create(getContext(), &pb.CreateServiceRequest{
					Service: nil,
				})
				Expect(err).To(BeNil())
				Expect(resp.Response.Code).To(Equal(scerr.ErrInvalidParams))
			})
		})

		Context("all max", func() {
			It("should be passed", func() {
				size := quota.DefaultSchemaQuota + 1
				paths := make([]*pb.ServicePath, 0, size)
				properties := make(map[string]string, size)
				for i := 0; i < size; i++ {
					s := strconv.Itoa(i) + strings.Repeat("x", 253)
					paths = append(paths, &pb.ServicePath{Path: s, Property: map[string]string{s: s}})
					properties[s] = s
				}
				r := &pb.CreateServiceRequest{
					Service: &pb.MicroService{
						AppId:       TOO_LONG_APPID[:len(TOO_LONG_APPID)-1],
						ServiceName: TOO_LONG_SERVICENAME[:len(TOO_LONG_SERVICENAME)-1],
						Version:     "32767.32767.32767",
						Alias:       TOO_LONG_ALIAS[:len(TOO_LONG_ALIAS)-1],
						Level:       "BACK",
						Status:      "UP",
						Schemas:     []string{TOO_LONG_SCHEMAID[:len(TOO_LONG_SCHEMAID)-1]},
						Paths:       paths,
						Properties:  properties,
						Framework: &pb.FrameWorkProperty{
							Name:    TOO_LONG_FRAMEWORK[:len(TOO_LONG_FRAMEWORK)-1],
							Version: TOO_LONG_FRAMEWORK[:len(TOO_LONG_FRAMEWORK)-1],
						},
						RegisterBy: "SDK",
					},
				}
				resp, err := serviceResource.Create(getContext(), r)
				Expect(err).To(BeNil())
				Expect(resp.Response.Code).To(Equal(pb.Response_SUCCESS))
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
						ServiceName: "create_service_rule_tag",
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
				Expect(resp.Response.Code).To(Equal(pb.Response_SUCCESS))

				By("second: create")
				delete(tags, "test")
				tags["second"] = "second"
				resp, err = serviceResource.Create(getContext(), &pb.CreateServiceRequest{
					Service: &pb.MicroService{
						ServiceName: "create_service_rule_tag",
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
				Expect(resp.Response.Code).To(Equal(pb.Response_SUCCESS))

				By("check the tags/rules/instances")
				respGetTags, err := serviceResource.GetTags(getContext(), &pb.GetServiceTagsRequest{
					ServiceId: resp.ServiceId,
				})
				Expect(err).To(BeNil())
				Expect(respGetTags.Response.Code).To(Equal(pb.Response_SUCCESS))
				Expect(respGetTags.Tags["second"]).To(Equal("second"))

				respGetRules, err := serviceResource.GetRule(getContext(), &pb.GetServiceRulesRequest{
					ServiceId: resp.ServiceId,
				})
				Expect(err).To(BeNil())
				Expect(respGetRules.Response.Code).To(Equal(pb.Response_SUCCESS))
				Expect(respGetRules.Rules[0].Attribute).To(Equal("ServiceName"))

				respGetInsts, err := instanceResource.GetInstances(getContext(), &pb.GetInstancesRequest{
					ConsumerServiceId: resp.ServiceId,
					ProviderServiceId: resp.ServiceId,
				})
				Expect(err).To(BeNil())
				Expect(respGetInsts.Response.Code).To(Equal(pb.Response_SUCCESS))
				Expect(respGetInsts.Instances[0].HostName).To(Equal("UT-HOST"))

				By("delete service")
				respDelete, err := serviceResource.Delete(getContext(), &pb.DeleteServiceRequest{
					ServiceId: resp.ServiceId,
					Force:     true,
				})
				Expect(err).To(BeNil())
				Expect(respDelete.Response.Code).To(Equal(pb.Response_SUCCESS))

			})
		})

		Context("when creating the same service", func() {
			It("should be failed", func() {
				By("the same serviceName")
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
				Expect(resp.Response.Code).To(Equal(pb.Response_SUCCESS))
				sameId := resp.ServiceId

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
				Expect(resp.Response.Code).To(Equal(pb.Response_SUCCESS))
				Expect(resp.ServiceId).To(Equal(sameId))

				By("the same alias")
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
				Expect(resp.Response.Code).To(Equal(pb.Response_SUCCESS))
				Expect(resp.ServiceId).To(Equal(sameId))

				By("the same serviceId and the same serviceName")
				resp, err = serviceResource.Create(getContext(), &pb.CreateServiceRequest{
					Service: &pb.MicroService{
						ServiceId:   sameId,
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
				Expect(resp.Response.Code).To(Equal(pb.Response_SUCCESS))
				Expect(resp.ServiceId).To(Equal(sameId))

				By("the same serviceId and the same alias")
				resp, err = serviceResource.Create(getContext(), &pb.CreateServiceRequest{
					Service: &pb.MicroService{
						ServiceId:   sameId,
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
				Expect(resp.Response.Code).To(Equal(pb.Response_SUCCESS))
				Expect(resp.ServiceId).To(Equal(sameId))

				By("the same service key but with diff serviceId")
				resp, err = serviceResource.Create(getContext(), &pb.CreateServiceRequest{
					Service: &pb.MicroService{
						ServiceId:   "customId",
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
				Expect(resp.Response.Code).To(Equal(scerr.ErrServiceAlreadyExists))
				resp, err = serviceResource.Create(getContext(), &pb.CreateServiceRequest{
					Service: &pb.MicroService{
						ServiceId:   "customId",
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
				Expect(resp.Response.Code).To(Equal(scerr.ErrServiceAlreadyExists))

				By("the same service id but with diff key")
				resp, err = serviceResource.Create(getContext(), &pb.CreateServiceRequest{
					Service: &pb.MicroService{
						ServiceId:   sameId,
						ServiceName: "some-relay2",
						Alias:       "sr2",
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
				Expect(resp.Response.Code).To(Equal(scerr.ErrServiceAlreadyExists))
			})
			It("same serviceId,different service, can not register again,error is same as the service register twice", func() {
				resp, err := serviceResource.Create(getContext(), &pb.CreateServiceRequest{
					Service: &pb.MicroService{
						ServiceId:   "same_serviceId",
						ServiceName: "serviceA",
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
				Expect(resp.Response.Code).To(Equal(pb.Response_SUCCESS))

				resp, err = serviceResource.Create(getContext(), &pb.CreateServiceRequest{
					Service: &pb.MicroService{
						ServiceId:   "same_serviceId",
						ServiceName: "serviceB",
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
				Expect(resp.Response.Code).ToNot(Equal(pb.Response_SUCCESS))
			})
		})

		Context("when creating a diff env service", func() {
			It("should be passed", func() {
				service := &pb.MicroService{
					ServiceName: "diff_env_service",
					AppId:       "default",
					Version:     "1.0.0",
					Level:       "FRONT",
					Schemas: []string{
						"xxxxxxxx",
					},
					Status: "UP",
				}
				resp, err := serviceResource.Create(getContext(), &pb.CreateServiceRequest{
					Service: service,
				})
				Expect(err).To(BeNil())
				Expect(resp.Response.Code).To(Equal(pb.Response_SUCCESS))

				service.ServiceId = ""
				service.Environment = pb.ENV_PROD
				resp, err = serviceResource.Create(getContext(), &pb.CreateServiceRequest{
					Service: service,
				})
				Expect(err).To(BeNil())
				Expect(resp.Response.Code).To(Equal(pb.Response_SUCCESS))
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
				Expect(resp.Response.Code).To(Equal(pb.Response_SUCCESS))
			})
		})

		Context("when service body is invalid", func() {
			It("should be failed", func() {
				By("invalid serviceId")
				r := &pb.CreateServiceRequest{
					Service: &pb.MicroService{
						ServiceId:   " ",
						AppId:       "default",
						ServiceName: "service-validate",
						Version:     "1.0.0",
						Level:       "BACK",
						Status:      "UP",
					},
				}
				resp, err := serviceResource.Create(getContext(), r)
				Expect(err).To(BeNil())
				Expect(resp.Response.Code).To(Equal(scerr.ErrInvalidParams))

				By("invalid appId")
				r = &pb.CreateServiceRequest{
					Service: &pb.MicroService{
						AppId:       TOO_LONG_APPID,
						ServiceName: "service-validate",
						Version:     "1.0.0",
						Level:       "BACK",
						Status:      "UP",
					},
				}
				resp, err = serviceResource.Create(getContext(), r)
				Expect(err).To(BeNil())
				Expect(resp.Response.Code).To(Equal(scerr.ErrInvalidParams))

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
				Expect(resp.Response.Code).To(Equal(scerr.ErrInvalidParams))

				By("invalid version 1")
				r = &pb.CreateServiceRequest{
					Service: &pb.MicroService{
						AppId:       "default",
						ServiceName: "service-validate",
						Version:     "1.",
						Level:       "BACK",
						Status:      "UP",
					},
				}
				resp, err = serviceResource.Create(getContext(), r)
				Expect(err).To(BeNil())
				Expect(resp.Response.Code).To(Equal(scerr.ErrInvalidParams))

				By("invalid version 2")
				r = &pb.CreateServiceRequest{
					Service: &pb.MicroService{
						AppId:       "default",
						ServiceName: "service-validate",
						Version:     "1.a.0",
						Level:       "BACK",
						Status:      "UP",
					},
				}
				resp, err = serviceResource.Create(getContext(), r)
				Expect(err).To(BeNil())
				Expect(resp.Response.Code).To(Equal(scerr.ErrInvalidParams))

				By("invalid version 3")
				r = &pb.CreateServiceRequest{
					Service: &pb.MicroService{
						AppId:       "default",
						ServiceName: "service-validate",
						Version:     "1.32768.0",
						Level:       "BACK",
						Status:      "UP",
					},
				}
				resp, err = serviceResource.Create(getContext(), r)
				Expect(err).To(BeNil())
				Expect(resp.Response.Code).To(Equal(scerr.ErrInvalidParams))

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
				Expect(resp.Response.Code).To(Equal(scerr.ErrInvalidParams))

				By("invalid env")
				r = &pb.CreateServiceRequest{
					Service: &pb.MicroService{
						Environment: "notexistenv",
						AppId:       "default",
						ServiceName: "service-invalidate-env",
						Version:     "1.0.0",
						Level:       "BACK",
						Status:      "UP",
					},
				}
				resp, err = serviceResource.Create(getContext(), r)
				Expect(err).To(BeNil())
				Expect(resp.Response.Code).To(Equal(scerr.ErrInvalidParams))

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
				Expect(resp.Response.Code).To(Equal(scerr.ErrInvalidParams))

				By("invalid alias")
				r = &pb.CreateServiceRequest{
					Service: &pb.MicroService{
						AppId:       "default",
						ServiceName: "service-validate-alias",
						Alias:       TOO_LONG_ALIAS,
						Version:     "1.0.0",
						Level:       "BACK",
						Status:      "UP",
					},
				}
				resp, err = serviceResource.Create(getContext(), r)
				Expect(err).To(BeNil())
				Expect(resp.Response.Code).To(Equal(scerr.ErrInvalidParams))

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
				Expect(resp.Response.Code).To(Equal(pb.Response_SUCCESS))

				By("invalid framework version")
				r = &pb.CreateServiceRequest{
					Service: &pb.MicroService{
						ServiceName: "framework-test",
						AppId:       "default",
						Version:     "1.0.4",
						Level:       "BACK",
						Framework: &pb.FrameWorkProperty{
							Version: TOO_LONG_FRAMEWORK,
						},
						Properties: make(map[string]string),
						Status:     "UP",
					},
				}
				resp, err = serviceResource.Create(getContext(), r)
				Expect(err).To(BeNil())
				Expect(resp.Response.Code).To(Equal(scerr.ErrInvalidParams))

				By("invalid framework name 1")
				r = &pb.CreateServiceRequest{
					Service: &pb.MicroService{
						ServiceName: "framework-test",
						AppId:       "default",
						Version:     "1.0.5",
						Level:       "BACK",
						Framework: &pb.FrameWorkProperty{
							Name: TOO_LONG_FRAMEWORK,
						},
						Properties: make(map[string]string),
						Status:     "UP",
					},
				}
				resp, err = serviceResource.Create(getContext(), r)
				Expect(err).To(BeNil())
				Expect(resp.Response.Code).To(Equal(scerr.ErrInvalidParams))

				By("invalid framework name 2")
				r = &pb.CreateServiceRequest{
					Service: &pb.MicroService{
						ServiceName: "framework-test",
						AppId:       "default",
						Version:     "1.0.5",
						Level:       "BACK",
						Framework: &pb.FrameWorkProperty{
							Name: "test@$",
						},
						Properties: make(map[string]string),
						Status:     "UP",
					},
				}
				resp, err = serviceResource.Create(getContext(), r)
				Expect(err).To(BeNil())
				Expect(resp.Response.Code).To(Equal(scerr.ErrInvalidParams))

				By("invalid registerBy")
				r = &pb.CreateServiceRequest{
					Service: &pb.MicroService{
						AppId:       "default",
						ServiceName: "framework-test",
						Version:     "1.0.6",
						Level:       "BACK",
						Status:      "UP",
						RegisterBy:  "InValid",
					},
				}
				resp, err = serviceResource.Create(getContext(), r)
				Expect(err).To(BeNil())
				Expect(resp.Response.Code).To(Equal(scerr.ErrInvalidParams))

				By("valid registerBy")
				r = &pb.CreateServiceRequest{
					Service: &pb.MicroService{
						AppId:       "default",
						ServiceName: "registerBy-test",
						Version:     "1.0.10",
						Level:       "BACK",
						Status:      "UP",
						RegisterBy:  "PLATFORM",
					},
				}
				resp, err = serviceResource.Create(getContext(), r)
				Expect(err).To(BeNil())
				Expect(resp.Response.Code).To(Equal(pb.Response_SUCCESS))

				By("valid registerBy")
				r = &pb.CreateServiceRequest{
					Service: &pb.MicroService{
						AppId:       "default",
						ServiceName: "registerBy-test",
						Version:     "1.0.11",
						Level:       "BACK",
						Status:      "UP",
						RegisterBy:  "SIDECAR",
					},
				}
				resp, err = serviceResource.Create(getContext(), r)
				Expect(err).To(BeNil())
				Expect(resp.Response.Code).To(Equal(pb.Response_SUCCESS))

				By("invalid description")
				r = &pb.CreateServiceRequest{
					Service: &pb.MicroService{
						AppId:       "default",
						ServiceName: "description-test",
						Version:     "1.0.6",
						Level:       "BACK",
						Status:      "UP",
						Description: TOO_LONG_DESCRIPTION,
					},
				}
				resp, err = serviceResource.Create(getContext(), r)
				Expect(err).To(BeNil())
				Expect(resp.Response.Code).To(Equal(scerr.ErrInvalidParams))

				By("schemaId out of range")
				r = &pb.CreateServiceRequest{
					Service: &pb.MicroService{
						AppId:       "default",
						ServiceName: "schema-test",
						Level:       "BACK",
						Status:      "UP",
						Schemas:     []string{TOO_LONG_SCHEMAID},
					},
				}
				resp, err = serviceResource.Create(getContext(), r)
				Expect(err).To(BeNil())
				Expect(resp.Response.Code).To(Equal(scerr.ErrInvalidParams))
			})
		})

		Context("when service with framework/registerBy/status", func() {
			It("should be passed", func() {
				By("framework name is nil")
				r := &pb.CreateServiceRequest{
					Service: &pb.MicroService{
						ServiceName: "framework-test",
						AppId:       "default",
						Version:     "1.0.1",
						Level:       "BACK",
						Framework: &pb.FrameWorkProperty{
							Version: "1.0.0",
						},
						Properties: make(map[string]string),
						Status:     "UP",
					},
				}
				resp, err := serviceResource.Create(getContext(), r)
				Expect(err).To(BeNil())
				Expect(resp.Response.Code).To(Equal(pb.Response_SUCCESS))

				By("framework version is nil")
				r = &pb.CreateServiceRequest{
					Service: &pb.MicroService{
						ServiceName: "framework-test",
						AppId:       "default",
						Version:     "1.0.2",
						Level:       "BACK",
						Framework: &pb.FrameWorkProperty{
							Name: "framework",
						},
						Properties: make(map[string]string),
						Status:     "UP",
					},
				}
				resp, err = serviceResource.Create(getContext(), r)
				Expect(err).To(BeNil())
				Expect(resp.Response.Code).To(Equal(pb.Response_SUCCESS))

				By("status is nil")
				r = &pb.CreateServiceRequest{
					Service: &pb.MicroService{
						ServiceName: "status-test",
						AppId:       "default",
						Version:     "1.0.3",
						Level:       "BACK",
						Properties:  make(map[string]string),
					},
				}
				resp, err = serviceResource.Create(getContext(), r)
				Expect(err).To(BeNil())
				Expect(resp.Response.Code).To(Equal(pb.Response_SUCCESS))
			})
		})
	})

	Describe("execute 'exists' operation", func() {
		var (
			serviceId1 string
			serviceId2 string
		)

		It("should be passed", func() {
			service := &pb.MicroService{
				Alias:       "es",
				ServiceName: "exist_service",
				AppId:       "exist_appId",
				Version:     "1.0.0",
				Level:       "FRONT",
				Schemas: []string{
					"first_schemaId",
				},
				Status: "UP",
			}
			respCreateService, err := serviceResource.Create(getContext(), &pb.CreateServiceRequest{
				Service: service,
			})
			Expect(err).To(BeNil())
			Expect(respCreateService.ServiceId).ToNot(Equal(""))
			serviceId1 = respCreateService.ServiceId

			service.ServiceId = ""
			service.Environment = pb.ENV_PROD
			respCreateService, err = serviceResource.Create(getContext(), &pb.CreateServiceRequest{
				Service: service,
			})
			Expect(err).To(BeNil())
			Expect(respCreateService.ServiceId).ToNot(Equal(""))
			serviceId2 = respCreateService.ServiceId
		})

		Context("when type is invalid", func() {
			It("should be failed", func() {
				resp, err := serviceResource.Exist(getContext(), &pb.GetExistenceRequest{
					Type: "nonetype",
				})
				Expect(err).To(BeNil())
				Expect(resp.Response.Code).To(Equal(scerr.ErrInvalidParams))
			})
		})

		Context("when param is invalid", func() {
			It("should be failed", func() {
				By("serviceName is too long")
				resp, err := serviceResource.Exist(getContext(), &pb.GetExistenceRequest{
					Type:        "microservice",
					ServiceName: TOO_LONG_EXISTENCE,
					Version:     "1.0.0",
					AppId:       "exist_appId",
				})
				Expect(err).To(BeNil())
				Expect(resp.Response.Code).To(Equal(scerr.ErrInvalidParams))

				By("serviceName is empty")
				resp, err = serviceResource.Exist(getContext(), &pb.GetExistenceRequest{
					Type:        "microservice",
					AppId:       "default",
					ServiceName: "",
					Version:     "3.0.0",
				})
				Expect(err).To(BeNil())
				Expect(resp.Response.Code).To(Equal(scerr.ErrInvalidParams))

				By("invalid appId")
				resp, err = serviceResource.Exist(getContext(), &pb.GetExistenceRequest{
					Type:        "microservice",
					AppId:       TOO_LONG_APPID,
					ServiceName: "exist-invalid-appid",
					Version:     "3.0.0",
				})
				Expect(err).To(BeNil())
				Expect(resp.Response.Code).To(Equal(scerr.ErrInvalidParams))

				By("invalid version")
				resp, err = serviceResource.Exist(getContext(), &pb.GetExistenceRequest{
					Type:        "microservice",
					AppId:       "default",
					ServiceName: "exist-invalid-version",
					Version:     "3.32768.0",
				})
				Expect(err).To(BeNil())
				Expect(resp.Response.Code).To(Equal(scerr.ErrInvalidParams))

				By("version is empty")
				resp, err = serviceResource.Exist(getContext(), &pb.GetExistenceRequest{
					Type:        "microservice",
					AppId:       "default",
					ServiceName: "exist-empty-version",
					Version:     "",
				})
				Expect(err).To(BeNil())
				Expect(resp.Response.Code).To(Equal(scerr.ErrInvalidParams))
			})
		})

		Context("when service does not exist", func() {
			It("should be failed", func() {
				By("query a not exist serviceName")
				resp, err := serviceResource.Exist(getContext(), &pb.GetExistenceRequest{
					Type:        "microservice",
					AppId:       "exist_appId",
					ServiceName: "notExistService",
					Version:     "1.0.0",
				})
				Expect(err).To(BeNil())
				Expect(resp.Response.Code).To(Equal(scerr.ErrServiceNotExists))

				By("query a not exist env")
				resp, err = serviceResource.Exist(getContext(), &pb.GetExistenceRequest{
					Type:        "microservice",
					Environment: pb.ENV_TEST,
					AppId:       "exist_appId",
					ServiceName: "exist_service",
					Version:     "1.0.0",
				})
				Expect(err).To(BeNil())
				Expect(resp.Response.Code).To(Equal(scerr.ErrServiceNotExists))

				By("query a not exist env with alias")
				resp, err = serviceResource.Exist(getContext(), &pb.GetExistenceRequest{
					Type:        "microservice",
					Environment: pb.ENV_TEST,
					AppId:       "exist_appId",
					ServiceName: "es",
					Version:     "1.0.0",
				})
				Expect(err).To(BeNil())
				Expect(resp.Response.Code).To(Equal(scerr.ErrServiceNotExists))

				By("version mismatch")
				resp, err = serviceResource.Exist(getContext(), &pb.GetExistenceRequest{
					Type:        "microservice",
					AppId:       "exist_appId",
					ServiceName: "exist_service",
					Version:     "2.0.0",
				})
				Expect(err).To(BeNil())
				Expect(resp.Response.Code).To(Equal(scerr.ErrServiceNotExists))
				resp, err = serviceResource.Exist(getContext(), &pb.GetExistenceRequest{
					Type:        "microservice",
					AppId:       "exist_appId",
					ServiceName: "exist_service",
					Version:     "0.0.0-1.0.0",
				})
				Expect(err).To(BeNil())
				Expect(resp.Response.Code).To(Equal(scerr.ErrServiceVersionNotExists))
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
				Expect(resp.ServiceId).To(Equal(serviceId1))

				By("search with serviceName and env")
				resp, err = serviceResource.Exist(getContext(), &pb.GetExistenceRequest{
					Type:        "microservice",
					Environment: pb.ENV_PROD,
					AppId:       "exist_appId",
					ServiceName: "exist_service",
					Version:     "1.0.0",
				})
				Expect(err).To(BeNil())
				Expect(resp.ServiceId).To(Equal(serviceId2))

				By("search with alias")
				resp, err = serviceResource.Exist(getContext(), &pb.GetExistenceRequest{
					Type:        "microservice",
					AppId:       "exist_appId",
					ServiceName: "es",
					Version:     "1.0.0",
				})
				Expect(err).To(BeNil())
				Expect(resp.ServiceId).To(Equal(serviceId1))

				By("search with alias and env")
				resp, err = serviceResource.Exist(getContext(), &pb.GetExistenceRequest{
					Type:        "microservice",
					Environment: pb.ENV_PROD,
					AppId:       "exist_appId",
					ServiceName: "es",
					Version:     "1.0.0",
				})
				Expect(err).To(BeNil())
				Expect(resp.ServiceId).To(Equal(serviceId2))

				By("search with latest versionRule")
				resp, err = serviceResource.Exist(getContext(), &pb.GetExistenceRequest{
					Type:        "microservice",
					AppId:       "exist_appId",
					ServiceName: "exist_service",
					Version:     "latest",
				})
				Expect(err).To(BeNil())
				Expect(resp.ServiceId).To(Equal(serviceId1))

				By("search with 1.0.0+ versionRule")
				resp, err = serviceResource.Exist(getContext(), &pb.GetExistenceRequest{
					Type:        "microservice",
					AppId:       "exist_appId",
					ServiceName: "exist_service",
					Version:     "1.0.0+",
				})
				Expect(err).To(BeNil())
				Expect(resp.ServiceId).To(Equal(serviceId1))

				By("search with range versionRule")
				resp, err = serviceResource.Exist(getContext(), &pb.GetExistenceRequest{
					Type:        "microservice",
					AppId:       "exist_appId",
					ServiceName: "exist_service",
					Version:     "0.9.1-1.0.1",
				})
				Expect(err).To(BeNil())
				Expect(resp.ServiceId).To(Equal(serviceId1))
			})
		})
	})

	Describe("execute 'query' operation", func() {
		Context("when query all services", func() {
			It("should be larger than 0", func() {
				resp, err := serviceResource.GetServices(getContext(), &pb.GetServicesRequest{})
				Expect(err).To(BeNil())
				Expect(len(resp.Services)).To(Not(Equal(0)))
			})
		})

		Context("when query a not exist service by serviceId", func() {
			It("should be failed", func() {
				resp, err := serviceResource.GetOne(getContext(), &pb.GetServiceRequest{
					ServiceId: "",
				})
				Expect(err).To(BeNil())
				Expect(resp.Response.Code).To(Equal(scerr.ErrInvalidParams))

				resp, err = serviceResource.GetOne(getContext(), &pb.GetServiceRequest{
					ServiceId: "notexistservice",
				})
				Expect(err).To(BeNil())
				Expect(resp.Response.Code).To(Equal(scerr.ErrServiceNotExists))

				resp, err = serviceResource.GetOne(getContext(), &pb.GetServiceRequest{
					ServiceId: TOO_LONG_SERVICEID,
				})
				Expect(err).To(BeNil())
				Expect(resp.Response.Code).To(Equal(scerr.ErrInvalidParams))
			})
		})
	})

	Describe("execute 'update' operation", func() {
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
			Expect(respCreate.Response.Code).To(Equal(pb.Response_SUCCESS))
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
			Expect(respCreate.Response.Code).To(Equal(pb.Response_SUCCESS))
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
			Expect(respCreate.Response.Code).To(Equal(pb.Response_SUCCESS))
			serviceConsumerId = respCreate.ServiceId

			respFind, err := instanceResource.Find(getContext(), &pb.FindInstancesRequest{
				ConsumerServiceId: serviceConsumerId,
				AppId:             provider.AppId,
				ServiceName:       provider.ServiceName,
				VersionRule:       provider.Version,
			})
			Expect(err).To(BeNil())
			Expect(respFind.Response.Code).To(Equal(pb.Response_SUCCESS))

			Expect(deh.Handle()).To(BeNil())
		})

		Context("when request is invalid", func() {
			It("should be failed", func() {
				By("serviceId is empty")
				resp, err := serviceResource.Delete(getContext(), &pb.DeleteServiceRequest{
					ServiceId: "",
					Force:     true,
				})
				Expect(err).To(BeNil())
				Expect(resp.Response.Code).ToNot(Equal(pb.Response_SUCCESS))

				By("serviceId is invalid")
				resp, err = serviceResource.Delete(getContext(), &pb.DeleteServiceRequest{
					ServiceId: TOO_LONG_SERVICEID,
					Force:     true,
				})
				Expect(err).To(BeNil())
				Expect(resp.Response.Code).ToNot(Equal(pb.Response_SUCCESS))

				By("serviceId does not exist")
				resp, err = serviceResource.Delete(getContext(), &pb.DeleteServiceRequest{
					ServiceId: "notexistservice",
					Force:     true,
				})
				Expect(err).To(BeNil())
				Expect(resp.Response.Code).ToNot(Equal(pb.Response_SUCCESS))
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
				Expect(resp.Response.Code).ToNot(Equal(pb.Response_SUCCESS))
			})
		})

		Context("when delete a service contains instances with not force flag", func() {
			It("should be not allowed", func() {
				resp, err := serviceResource.Delete(getContext(), &pb.DeleteServiceRequest{
					ServiceId: serviceContainInstId,
					Force:     false,
				})
				Expect(err).To(BeNil())
				Expect(resp.Response.Code).ToNot(Equal(pb.Response_SUCCESS))
			})
		})

		Context("when delete a service contains instances with force flag", func() {
			It("should be passed", func() {
				resp, err := serviceResource.Delete(getContext(), &pb.DeleteServiceRequest{
					ServiceId: serviceContainInstId,
					Force:     true,
				})
				Expect(err).To(BeNil())
				Expect(resp.Response.Code).To(Equal(pb.Response_SUCCESS))
			})
		})

		Context("when delete a service depended on consumer with not force flag", func() {
			It("should be not allowed", func() {
				resp, err := serviceResource.Delete(getContext(), &pb.DeleteServiceRequest{
					ServiceId: serviceNoInstId,
					Force:     false,
				})
				Expect(err).To(BeNil())
				Expect(resp.Response.Code).ToNot(Equal(pb.Response_SUCCESS))
			})
		})

		Context("when delete a service depended on consumer with force flag", func() {
			It("should be passed", func() {
				resp, err := serviceResource.Delete(getContext(), &pb.DeleteServiceRequest{
					ServiceId: serviceNoInstId,
					Force:     true,
				})
				Expect(err).To(BeNil())
				Expect(resp.Response.Code).To(Equal(pb.Response_SUCCESS))
			})
		})

		Context("when delete a service with not force flag", func() {
			It("should be not allowed", func() {
				resp, err := serviceResource.Delete(getContext(), &pb.DeleteServiceRequest{
					ServiceId: serviceNoInstId,
					Force:     false,
				})
				Expect(err).To(BeNil())
				Expect(resp.Response.Code).ToNot(Equal(pb.Response_SUCCESS))
			})
		})

		Context("when batch delete services", func() {
			It("should be OK", func() {
				resp, err := serviceResource.DeleteServices(getContext(), &pb.DelServicesRequest{
					ServiceIds: []string{},
					Force:      false,
				})
				Expect(err).To(BeNil())
				Expect(resp.Response.Code).ToNot(Equal(pb.Response_SUCCESS))

				resp, err = serviceResource.DeleteServices(getContext(), &pb.DelServicesRequest{
					ServiceIds: []string{"@#$44332_non-invalid_serviceId"},
					Force:      false,
				})
				Expect(err).To(BeNil())
				Expect(resp.Response.Code).ToNot(Equal(pb.Response_SUCCESS))
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
				Expect(resp.Response.Code).To(Equal(pb.Response_SUCCESS))
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
				Expect(resp.Response.Code).To(Equal(pb.Response_SUCCESS))
				serviceId2 = resp.ServiceId
			})

			It("should be passed", func() {
				resp, err := serviceResource.DeleteServices(getContext(), &pb.DelServicesRequest{
					ServiceIds: []string{serviceId1, serviceId2},
					Force:      false,
				},
				)
				Expect(err).To(BeNil())
				Expect(resp.Response.Code).To(Equal(pb.Response_SUCCESS))
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
				Expect(resp.Response.Code).To(Equal(pb.Response_SUCCESS))

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
				Expect(resp.Response.Code).To(Equal(pb.Response_SUCCESS))
			})

			It("should be failed", func() {
				resp, err := serviceResource.DeleteServices(getContext(), &pb.DelServicesRequest{
					ServiceIds: []string{serviceIdFailed1, serviceIdFailed2},
					Force:      false,
				},
				)
				Expect(err).To(BeNil())
				//期待结果失败
				Expect(resp.Response.Code).ToNot(Equal(pb.Response_SUCCESS))
			})
		})
	})
})
