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
	"github.com/apache/incubator-servicecomb-service-center/pkg/util"
	pb "github.com/apache/incubator-servicecomb-service-center/server/core/proto"
	"github.com/apache/incubator-servicecomb-service-center/server/service/event"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"golang.org/x/net/context"
)

var deh event.DependencyEventHandler

var _ = Describe("'Dependency' service", func() {
	Describe("execute 'create' operation", func() {
		var (
			consumerId1 string
			consumerId2 string
		)

		It("should be passed", func() {
			respCreateService, err := serviceResource.Create(getContext(), &pb.CreateServiceRequest{
				Service: &pb.MicroService{
					AppId:       "create_dep_group",
					ServiceName: "create_dep_consumer",
					Version:     "1.0.0",
					Level:       "FRONT",
					Status:      pb.MS_UP,
				},
			})
			Expect(err).To(BeNil())
			Expect(respCreateService.Response.Code).To(Equal(pb.Response_SUCCESS))
			consumerId1 = respCreateService.ServiceId

			respCreateService, err = serviceResource.Create(getContext(), &pb.CreateServiceRequest{
				Service: &pb.MicroService{
					Environment: pb.ENV_PROD,
					AppId:       "create_dep_group",
					ServiceName: "create_dep_consumer",
					Version:     "1.0.0",
					Level:       "FRONT",
					Status:      pb.MS_UP,
				},
			})
			Expect(err).To(BeNil())
			Expect(respCreateService.Response.Code).To(Equal(pb.Response_SUCCESS))
			consumerId2 = respCreateService.ServiceId

			respCreateService, err = serviceResource.Create(getContext(), &pb.CreateServiceRequest{
				Service: &pb.MicroService{
					AppId:       "create_dep_group",
					ServiceName: "create_dep_provider",
					Version:     "1.0.0",
					Level:       "FRONT",
					Status:      pb.MS_UP,
				},
			})
			Expect(err).To(BeNil())
			Expect(respCreateService.Response.Code).To(Equal(pb.Response_SUCCESS))

			respCreateService, err = serviceResource.Create(getContext(), &pb.CreateServiceRequest{
				Service: &pb.MicroService{
					Environment: pb.ENV_PROD,
					AppId:       "create_dep_group",
					ServiceName: "create_dep_provider",
					Version:     "1.0.0",
					Level:       "FRONT",
					Status:      pb.MS_UP,
				},
			})
			Expect(err).To(BeNil())
			Expect(respCreateService.Response.Code).To(Equal(pb.Response_SUCCESS))
		})

		Context("when request is invalid", func() {
			It("should be failed", func() {
				By("dependency is nil")
				respCreateDependency, err := serviceResource.CreateDependenciesForMicroServices(getContext(), &pb.CreateDependenciesRequest{})
				Expect(err).To(BeNil())
				Expect(respCreateDependency.Response.Code).ToNot(Equal(pb.Response_SUCCESS))

				consumer := &pb.MicroServiceKey{
					AppId:       "create_dep_group",
					ServiceName: "create_dep_consumer",
					Version:     "1.0.0",
				}
				providers := []*pb.MicroServiceKey{
					{
						AppId:       "create_dep_group",
						ServiceName: "create_dep_provider",
						Version:     "1.0.0",
					},
				}

				By("consumer does not exist")
				respCreateDependency, err = serviceResource.CreateDependenciesForMicroServices(getContext(), &pb.CreateDependenciesRequest{
					Dependencies: []*pb.ConsumerDependency{
						{
							Consumer: &pb.MicroServiceKey{
								AppId:       "noexistapp",
								ServiceName: "noexistservice",
								Version:     "1.0.0",
							},
							Providers: providers,
						},
					},
				})
				Expect(err).To(BeNil())
				Expect(respCreateDependency.Response.Code).ToNot(Equal(pb.Response_SUCCESS))

				By("consumer version is invalid")
				respCreateDependency, err = serviceResource.CreateDependenciesForMicroServices(getContext(), &pb.CreateDependenciesRequest{
					Dependencies: []*pb.ConsumerDependency{
						{
							Consumer: &pb.MicroServiceKey{
								AppId:       "create_dep_group",
								ServiceName: "create_dep_consumer",
								Version:     "1.0.0+",
							},
							Providers: providers,
						},
					},
				})
				Expect(err).To(BeNil())
				Expect(respCreateDependency.Response.Code).ToNot(Equal(pb.Response_SUCCESS))

				By("consumer serviceName is invalid")
				respCreateDependency, err = serviceResource.CreateDependenciesForMicroServices(getContext(), &pb.CreateDependenciesRequest{
					Dependencies: []*pb.ConsumerDependency{
						{
							Consumer: &pb.MicroServiceKey{
								AppId:       "create_dep_group",
								ServiceName: "*",
								Version:     "1.0.0",
							},
							Providers: providers,
						},
					},
				})
				Expect(err).To(BeNil())
				Expect(respCreateDependency.Response.Code).ToNot(Equal(pb.Response_SUCCESS))

				By("provider is nil")
				respCreateDependency, err = serviceResource.CreateDependenciesForMicroServices(getContext(), &pb.CreateDependenciesRequest{
					Dependencies: []*pb.ConsumerDependency{
						{
							Consumer: consumer,
						},
					},
				})
				Expect(err).To(BeNil())
				Expect(respCreateDependency.Response.Code).ToNot(Equal(pb.Response_SUCCESS))

				By("provider app is invalid")
				respCreateDependency, err = serviceResource.CreateDependenciesForMicroServices(getContext(), &pb.CreateDependenciesRequest{
					Dependencies: []*pb.ConsumerDependency{
						{
							Consumer: consumer,
							Providers: []*pb.MicroServiceKey{
								{
									AppId:       "*",
									ServiceName: "service_name_provider",
									Version:     "2.0.0",
								},
							},
						},
					},
				})
				Expect(err).To(BeNil())
				Expect(respCreateDependency.Response.Code).ToNot(Equal(pb.Response_SUCCESS))

				By("provider serviceName is invalid")
				respCreateDependency, err = serviceResource.CreateDependenciesForMicroServices(getContext(), &pb.CreateDependenciesRequest{
					Dependencies: []*pb.ConsumerDependency{
						{
							Consumer: consumer,
							Providers: []*pb.MicroServiceKey{
								{
									AppId:       "service_group_provider",
									ServiceName: "-",
									Version:     "2.0.0",
								},
							},
						},
					},
				})
				Expect(err).To(BeNil())
				Expect(respCreateDependency.Response.Code).ToNot(Equal(pb.Response_SUCCESS))

				By("provider version is invalid")
				respCreateDependency, err = serviceResource.CreateDependenciesForMicroServices(getContext(), &pb.CreateDependenciesRequest{
					Dependencies: []*pb.ConsumerDependency{
						{
							Consumer: consumer,
							Providers: []*pb.MicroServiceKey{
								{
									AppId:       "service_group_provider",
									ServiceName: "service_name_provider",
									Version:     "",
								},
							},
						},
					},
				})
				Expect(err).To(BeNil())
				Expect(respCreateDependency.Response.Code).ToNot(Equal(pb.Response_SUCCESS))

				By("provider in diff env")
				respCreateDependency, err = serviceResource.CreateDependenciesForMicroServices(getContext(), &pb.CreateDependenciesRequest{
					Dependencies: []*pb.ConsumerDependency{
						{
							Consumer: consumer,
							Providers: []*pb.MicroServiceKey{
								{
									Environment: pb.ENV_PROD,
									AppId:       "service_group_provider",
									ServiceName: "service_name_provider",
									Version:     "latest",
								},
							},
						},
					},
				})
				Expect(err).To(BeNil())
				Expect(respCreateDependency.Response.Code).To(Equal(pb.Response_SUCCESS))

				By("consumer in diff env")
				consumer.Environment = pb.ENV_PROD
				respCreateDependency, err = serviceResource.CreateDependenciesForMicroServices(getContext(), &pb.CreateDependenciesRequest{
					Dependencies: []*pb.ConsumerDependency{
						{
							Consumer: consumer,
							Providers: []*pb.MicroServiceKey{
								{
									AppId:       "service_group_provider",
									ServiceName: "service_name_provider",
									Version:     "latest",
								},
							},
						},
					},
				})
				Expect(err).To(BeNil())
				Expect(respCreateDependency.Response.Code).To(Equal(pb.Response_SUCCESS))

				Expect(deh.Handle()).To(BeNil())

				respCon, err := serviceResource.GetConsumerDependencies(getContext(), &pb.GetDependenciesRequest{
					ServiceId: consumerId1,
				})
				Expect(err).To(BeNil())
				Expect(respCon.Response.Code).To(Equal(pb.Response_SUCCESS))
				Expect(len(respCon.Providers)).To(Equal(0))

				respCon, err = serviceResource.GetConsumerDependencies(getContext(), &pb.GetDependenciesRequest{
					ServiceId: consumerId2,
				})
				Expect(err).To(BeNil())
				Expect(respCon.Response.Code).To(Equal(pb.Response_SUCCESS))
				Expect(len(respCon.Providers)).To(Equal(0))
			})
		})

		Context("when request is valid", func() {
			It("should be passed", func() {
				consumer := &pb.MicroServiceKey{
					ServiceName: "create_dep_consumer",
					AppId:       "create_dep_group",
					Version:     "1.0.0",
				}

				By("add provider is empty")
				respCreateDependency, err := serviceResource.CreateDependenciesForMicroServices(getContext(), &pb.CreateDependenciesRequest{
					Dependencies: []*pb.ConsumerDependency{
						{
							Consumer:  consumer,
							Providers: []*pb.MicroServiceKey{},
						},
					},
				})
				Expect(err).To(BeNil())
				Expect(respCreateDependency.Response.Code).To(Equal(pb.Response_SUCCESS))

				By("add latest")
				respCreateDependency, err = serviceResource.CreateDependenciesForMicroServices(getContext(), &pb.CreateDependenciesRequest{
					Dependencies: []*pb.ConsumerDependency{
						{
							Consumer: consumer,
							Providers: []*pb.MicroServiceKey{
								{
									AppId:       "create_dep_group",
									ServiceName: "create_dep_provider",
									Version:     "latest",
								},
							},
						},
					},
				})
				Expect(err).To(BeNil())
				Expect(respCreateDependency.Response.Code).To(Equal(pb.Response_SUCCESS))

				By("add 1.0.0+")
				respCreateDependency, err = serviceResource.CreateDependenciesForMicroServices(getContext(), &pb.CreateDependenciesRequest{
					Dependencies: []*pb.ConsumerDependency{
						{
							Consumer: consumer,
							Providers: []*pb.MicroServiceKey{
								{
									AppId:       "create_dep_group",
									ServiceName: "create_dep_provider",
									Version:     "1.0.0+",
								},
							},
						},
					},
				})
				Expect(err).To(BeNil())
				Expect(respCreateDependency.Response.Code).To(Equal(pb.Response_SUCCESS))

				By("add *")
				respCreateDependency, err = serviceResource.CreateDependenciesForMicroServices(getContext(), &pb.CreateDependenciesRequest{
					Dependencies: []*pb.ConsumerDependency{
						{
							Consumer: consumer,
							Providers: []*pb.MicroServiceKey{
								{
									ServiceName: "*",
								},
							},
						},
					},
				})
				Expect(err).To(BeNil())
				Expect(respCreateDependency.Response.Code).To(Equal(pb.Response_SUCCESS))

				By("add multiple providers")
				respCreateDependency, err = serviceResource.CreateDependenciesForMicroServices(getContext(), &pb.CreateDependenciesRequest{
					Dependencies: []*pb.ConsumerDependency{
						{
							Consumer: consumer,
							Providers: []*pb.MicroServiceKey{
								{
									AppId:       "create_dep_group",
									ServiceName: "create_dep_provider",
									Version:     "1.0.0",
								},
								{
									ServiceName: "*",
								},
							},
						},
					},
				})
				Expect(err).To(BeNil())
				Expect(respCreateDependency.Response.Code).To(Equal(pb.Response_SUCCESS))

				By("add 1.0.0-2.0.0")
				respCreateDependency, err = serviceResource.CreateDependenciesForMicroServices(getContext(), &pb.CreateDependenciesRequest{
					Dependencies: []*pb.ConsumerDependency{
						{
							Consumer: consumer,
							Providers: []*pb.MicroServiceKey{
								{
									AppId:       "create_dep_group",
									ServiceName: "create_dep_provider",
									Version:     "1.0.0-2.0.0",
								},
							},
						},
					},
				})
				Expect(err).To(BeNil())
				Expect(respCreateDependency.Response.Code).To(Equal(pb.Response_SUCCESS))

				By("add not override")
				respAddDependency, err := serviceResource.AddDependenciesForMicroServices(getContext(), &pb.AddDependenciesRequest{
					Dependencies: []*pb.ConsumerDependency{
						{
							Consumer: consumer,
							Providers: []*pb.MicroServiceKey{
								{
									AppId:       "create_dep_group",
									ServiceName: "create_dep_provider",
									Version:     "1.0.0-2.0.0",
								},
							},
						},
					},
				})
				Expect(err).To(BeNil())
				Expect(respAddDependency.Response.Code).To(Equal(pb.Response_SUCCESS))
			})
		})
	})

	Describe("execute 'get' operartion", func() {
		var (
			consumerId1 string
			consumerId2 string
			providerId1 string
			providerId2 string
		)

		It("should be passed", func() {
			respCreateService, err := serviceResource.Create(getContext(), &pb.CreateServiceRequest{
				Service: &pb.MicroService{
					AppId:       "get_dep_group",
					ServiceName: "get_dep_consumer",
					Version:     "1.0.0",
					Level:       "FRONT",
					Status:      pb.MS_UP,
				},
			})
			Expect(err).To(BeNil())
			Expect(respCreateService.Response.Code).To(Equal(pb.Response_SUCCESS))
			consumerId1 = respCreateService.ServiceId

			respCreateService, err = serviceResource.Create(util.SetDomainProject(context.Background(), "user", "user"),
				&pb.CreateServiceRequest{
					Service: &pb.MicroService{
						Environment: pb.ENV_DEV,
						AppId:       "get_dep_group",
						ServiceName: "get_same_domain_dep_consumer",
						Version:     "1.0.0",
						Level:       "FRONT",
						Status:      pb.MS_UP,
					},
				})
			Expect(err).To(BeNil())
			Expect(respCreateService.Response.Code).To(Equal(pb.Response_SUCCESS))
			consumerId2 = respCreateService.ServiceId

			respCreateService, err = serviceResource.Create(getContext(), &pb.CreateServiceRequest{
				Service: &pb.MicroService{
					AppId:       "get_dep_group",
					ServiceName: "get_dep_provider",
					Version:     "1.0.0",
					Level:       "FRONT",
					Status:      pb.MS_UP,
				},
			})
			Expect(err).To(BeNil())
			Expect(respCreateService.Response.Code).To(Equal(pb.Response_SUCCESS))
			providerId1 = respCreateService.ServiceId

			respCreateService, err = serviceResource.Create(getContext(), &pb.CreateServiceRequest{
				Service: &pb.MicroService{
					AppId:       "get_dep_group",
					ServiceName: "get_dep_provider",
					Version:     "2.0.0",
					Level:       "FRONT",
					Status:      pb.MS_UP,
				},
			})
			Expect(err).To(BeNil())
			Expect(respCreateService.Response.Code).To(Equal(pb.Response_SUCCESS))
			providerId2 = respCreateService.ServiceId
		})

		Context("when request is invalid", func() {
			It("should be failed", func() {
				By("service id is empty when get provider")
				respPro, err := serviceResource.GetProviderDependencies(getContext(), &pb.GetDependenciesRequest{
					ServiceId: "",
				})
				Expect(err).To(BeNil())
				Expect(respPro.Response.Code).ToNot(Equal(pb.Response_SUCCESS))

				By("service does not exist when get provider")
				respPro, err = serviceResource.GetProviderDependencies(getContext(), &pb.GetDependenciesRequest{
					ServiceId: "noneservice",
				})
				Expect(err).To(BeNil())
				Expect(respPro.Response.Code).ToNot(Equal(pb.Response_SUCCESS))

				By("service id is empty when get consumer")
				respCon, err := serviceResource.GetConsumerDependencies(getContext(), &pb.GetDependenciesRequest{
					ServiceId: "",
				})
				Expect(err).To(BeNil())
				Expect(respCon.Response.Code).ToNot(Equal(pb.Response_SUCCESS))

				By("service does not exist when get consumer")
				respCon, err = serviceResource.GetConsumerDependencies(getContext(), &pb.GetDependenciesRequest{
					ServiceId: "noneservice",
				})
				Expect(err).To(BeNil())
				Expect(respCon.Response.Code).ToNot(Equal(pb.Response_SUCCESS))
			})
		})

		Context("when request is valid", func() {
			It("should be passed", func() {
				By("get provider")
				respPro, err := serviceResource.GetProviderDependencies(getContext(), &pb.GetDependenciesRequest{
					ServiceId: providerId1,
				})
				Expect(err).To(BeNil())
				Expect(respPro.Response.Code).To(Equal(pb.Response_SUCCESS))

				By("get consumer")
				respCon, err := serviceResource.GetConsumerDependencies(getContext(), &pb.GetDependenciesRequest{
					ServiceId: consumerId1,
				})
				Expect(err).To(BeNil())
				Expect(respCon.Response.Code).To(Equal(pb.Response_SUCCESS))
			})
		})

		Context("when after finding instance", func() {
			It("should created dependencies between C and P", func() {
				By("find provider")
				resp, err := instanceResource.Find(getContext(), &pb.FindInstancesRequest{
					ConsumerServiceId: consumerId1,
					AppId:             "get_dep_group",
					ServiceName:       "get_dep_provider",
					VersionRule:       "1.0.0+",
				})
				Expect(err).To(BeNil())
				Expect(resp.Response.Code).To(Equal(pb.Response_SUCCESS))

				Expect(deh.Handle()).To(BeNil())

				By("get consumer's deps")
				respGetP, err := serviceResource.GetProviderDependencies(getContext(), &pb.GetDependenciesRequest{
					ServiceId: providerId1,
				})
				Expect(err).To(BeNil())
				Expect(respGetP.Response.Code).To(Equal(pb.Response_SUCCESS))
				Expect(respGetP.Consumers[0].ServiceId).To(Equal(consumerId1))

				By("get provider's deps")
				respGetC, err := serviceResource.GetConsumerDependencies(getContext(), &pb.GetDependenciesRequest{
					ServiceId: consumerId1,
				})
				Expect(err).To(BeNil())
				Expect(respGetC.Response.Code).To(Equal(pb.Response_SUCCESS))
				Expect(len(respGetC.Providers)).To(Equal(2))

				//重复find
				resp, err = instanceResource.Find(getContext(), &pb.FindInstancesRequest{
					ConsumerServiceId: consumerId1,
					AppId:             "get_dep_group",
					ServiceName:       "get_dep_provider",
					VersionRule:       "2.0.0+",
				})
				Expect(err).To(BeNil())
				Expect(resp.Response.Code).To(Equal(pb.Response_SUCCESS))

				Expect(deh.Handle()).To(BeNil())

				By("get consumer again")
				respGetP, err = serviceResource.GetProviderDependencies(getContext(), &pb.GetDependenciesRequest{
					ServiceId: providerId1,
				})
				Expect(err).To(BeNil())
				Expect(respGetP.Response.Code).To(Equal(pb.Response_SUCCESS))
				Expect(len(respGetP.Consumers)).To(Equal(0))

				By("get provider again")
				respGetC, err = serviceResource.GetConsumerDependencies(getContext(), &pb.GetDependenciesRequest{
					ServiceId: consumerId1,
				})
				Expect(err).To(BeNil())
				Expect(respGetC.Response.Code).To(Equal(pb.Response_SUCCESS))
				Expect(len(respGetC.Providers)).To(Equal(1))
				Expect(respGetC.Providers[0].ServiceId).To(Equal(providerId2))

				By("get self deps")
				resp, err = instanceResource.Find(getContext(), &pb.FindInstancesRequest{
					ConsumerServiceId: consumerId1,
					AppId:             "get_dep_group",
					ServiceName:       "get_dep_consumer",
					VersionRule:       "1.0.0+",
				})
				Expect(err).To(BeNil())
				Expect(resp.Response.Code).To(Equal(pb.Response_SUCCESS))

				respGetC, err = serviceResource.GetConsumerDependencies(getContext(), &pb.GetDependenciesRequest{
					ServiceId: consumerId1,
					NoSelf:    true,
				})
				Expect(err).To(BeNil())
				Expect(respGetC.Response.Code).To(Equal(pb.Response_SUCCESS))
				Expect(len(respGetC.Providers)).To(Equal(1))

				By("get same domain deps")
				respFind, err := instanceResource.Find(
					util.SetTargetDomainProject(
						util.SetDomainProject(util.CloneContext(getContext()), "user", "user"),
						"default", "default"),
					&pb.FindInstancesRequest{
						ConsumerServiceId: consumerId2,
						AppId:             "default",
						ServiceName:       "SERVICECENTER",
						VersionRule:       "latest",
					})
				Expect(err).To(BeNil())
				Expect(respFind.Response.Code).To(Equal(pb.Response_SUCCESS))

				respGetC, err = serviceResource.GetConsumerDependencies(
					util.SetDomainProject(context.Background(), "user", "user"),
					&pb.GetDependenciesRequest{
						ServiceId:  consumerId2,
						SameDomain: true,
					})
				Expect(err).To(BeNil())
				Expect(respGetC.Response.Code).To(Equal(pb.Response_SUCCESS))
				Expect(len(respGetC.Providers)).To(Equal(0))
			})
		})
	})
})
