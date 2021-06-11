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
package disco_test

import (
	"strconv"

	"github.com/apache/servicecomb-service-center/server/service/disco"

	pb "github.com/go-chassis/cari/discovery"
	"github.com/go-chassis/go-archaius"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/apache/servicecomb-service-center/datasource/etcd/client"
	"github.com/apache/servicecomb-service-center/datasource/etcd/event"
	"github.com/apache/servicecomb-service-center/datasource/etcd/kv"
	"github.com/apache/servicecomb-service-center/datasource/etcd/path"
)

var deh event.DependencyEventHandler

var _ = Describe("'Dependency' service", func() {
	Describe("execute 'create' operation", func() {
		var (
			consumerId1 string
			consumerId2 string
			consumerId3 string
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
			Expect(respCreateService.Response.GetCode()).To(Equal(pb.ResponseSuccess))
			consumerId1 = respCreateService.ServiceId

			respCreateService, err = serviceResource.Create(getContext(), &pb.CreateServiceRequest{
				Service: &pb.MicroService{
					AppId:       "create_dep_group",
					ServiceName: "create_dep_consumer_all",
					Version:     "1.0.0",
					Level:       "FRONT",
					Status:      pb.MS_UP,
				},
			})
			Expect(err).To(BeNil())
			Expect(respCreateService.Response.GetCode()).To(Equal(pb.ResponseSuccess))
			consumerId3 = respCreateService.ServiceId

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
			Expect(respCreateService.Response.GetCode()).To(Equal(pb.ResponseSuccess))
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
			Expect(respCreateService.Response.GetCode()).To(Equal(pb.ResponseSuccess))

			respCreateService, err = serviceResource.Create(getContext(), &pb.CreateServiceRequest{
				Service: &pb.MicroService{
					AppId:       "create_dep_group",
					ServiceName: "create_dep_provider",
					Version:     "1.0.1",
					Level:       "FRONT",
					Status:      pb.MS_UP,
				},
			})
			Expect(err).To(BeNil())
			Expect(respCreateService.Response.GetCode()).To(Equal(pb.ResponseSuccess))

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
			Expect(respCreateService.Response.GetCode()).To(Equal(pb.ResponseSuccess))
		})

		Context("when request is invalid", func() {
			It("should be failed", func() {
				By("dependency is nil")
				respCreateDependency, err := serviceResource.CreateDependenciesForMicroServices(getContext(), &pb.CreateDependenciesRequest{})
				Expect(err).To(BeNil())
				Expect(respCreateDependency.Response.GetCode()).ToNot(Equal(pb.ResponseSuccess))

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
				Expect(respCreateDependency.Response.GetCode()).ToNot(Equal(pb.ResponseSuccess))

				By("provider version is invalid")
				respCreateDependency, err = serviceResource.CreateDependenciesForMicroServices(getContext(), &pb.CreateDependenciesRequest{
					Dependencies: []*pb.ConsumerDependency{
						{
							Consumer: consumer,
							Providers: []*pb.MicroServiceKey{
								{
									AppId:       "create_dep_group",
									ServiceName: "create_dep_provider",
									Version:     "1.0.32768",
								},
							},
						},
					},
				})
				Expect(err).To(BeNil())
				Expect(respCreateDependency.Response.GetCode()).To(Equal(pb.ErrInvalidParams))

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
				Expect(respCreateDependency.Response.GetCode()).To(Equal(pb.ErrInvalidParams))
				respCreateDependency, err = serviceResource.CreateDependenciesForMicroServices(getContext(), &pb.CreateDependenciesRequest{
					Dependencies: []*pb.ConsumerDependency{
						{
							Consumer: &pb.MicroServiceKey{
								AppId:       "create_dep_group",
								ServiceName: "create_dep_consumer",
								Version:     "1.0.0-1.0.1",
							},
							Providers: providers,
						},
					},
				})
				Expect(err).To(BeNil())
				Expect(respCreateDependency.Response.GetCode()).To(Equal(pb.ErrInvalidParams))
				respCreateDependency, err = serviceResource.CreateDependenciesForMicroServices(getContext(), &pb.CreateDependenciesRequest{
					Dependencies: []*pb.ConsumerDependency{
						{
							Consumer: &pb.MicroServiceKey{
								AppId:       "create_dep_group",
								ServiceName: "create_dep_consumer",
								Version:     "latest",
							},
							Providers: providers,
						},
					},
				})
				Expect(err).To(BeNil())
				Expect(respCreateDependency.Response.GetCode()).To(Equal(pb.ErrInvalidParams))
				respCreateDependency, err = serviceResource.CreateDependenciesForMicroServices(getContext(), &pb.CreateDependenciesRequest{
					Dependencies: []*pb.ConsumerDependency{
						{
							Consumer: &pb.MicroServiceKey{
								AppId:       "create_dep_group",
								ServiceName: "create_dep_consumer",
								Version:     "",
							},
							Providers: providers,
						},
					},
				})
				Expect(err).To(BeNil())
				Expect(respCreateDependency.Response.GetCode()).To(Equal(pb.ErrInvalidParams))
				respCreateDependency, err = serviceResource.CreateDependenciesForMicroServices(getContext(), &pb.CreateDependenciesRequest{
					Dependencies: []*pb.ConsumerDependency{
						{
							Consumer: &pb.MicroServiceKey{
								AppId:       "create_dep_group",
								ServiceName: "create_dep_consumer",
								Version:     "1.0.32768",
							},
							Providers: providers,
						},
					},
				})
				Expect(err).To(BeNil())
				Expect(respCreateDependency.Response.GetCode()).To(Equal(pb.ErrInvalidParams))

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
				Expect(respCreateDependency.Response.GetCode()).ToNot(Equal(pb.ResponseSuccess))

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
				Expect(respCreateDependency.Response.GetCode()).ToNot(Equal(pb.ResponseSuccess))

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
				Expect(respCreateDependency.Response.GetCode()).ToNot(Equal(pb.ResponseSuccess))

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
				Expect(respCreateDependency.Response.GetCode()).ToNot(Equal(pb.ResponseSuccess))

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
				Expect(respCreateDependency.Response.GetCode()).To(Equal(pb.ResponseSuccess))

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
				Expect(respCreateDependency.Response.GetCode()).To(Equal(pb.ResponseSuccess))

				DependencyHandle()

				respCon, err := serviceResource.GetConsumerDependencies(getContext(), &pb.GetDependenciesRequest{
					ServiceId: consumerId1,
				})
				Expect(err).To(BeNil())
				Expect(respCon.Response.GetCode()).To(Equal(pb.ResponseSuccess))
				Expect(len(respCon.Providers)).To(Equal(0))

				respCon, err = serviceResource.GetConsumerDependencies(getContext(), &pb.GetDependenciesRequest{
					ServiceId: consumerId2,
				})
				Expect(err).To(BeNil())
				Expect(respCon.Response.GetCode()).To(Equal(pb.ResponseSuccess))
				Expect(len(respCon.Providers)).To(Equal(0))

				By("dependencies is invalid")
				var deps []*pb.ConsumerDependency
				for i := 0; i < 101; i++ {
					deps = append(deps, &pb.ConsumerDependency{
						Consumer: &pb.MicroServiceKey{
							AppId:       "create_dep_group",
							ServiceName: "create_dep_consumer" + strconv.Itoa(i),
							Version:     "1.0.0",
						},
						Providers: []*pb.MicroServiceKey{
							{
								AppId:       "service_group_provider",
								ServiceName: "service_name_provider",
								Version:     "latest",
							},
						},
					})
				}
				respCreateDependency, err = serviceResource.CreateDependenciesForMicroServices(getContext(), &pb.CreateDependenciesRequest{
					Dependencies: deps,
				})
				Expect(err).To(BeNil())
				Expect(respCreateDependency.Response.GetCode()).To(Equal(pb.ErrInvalidParams))
			})
		})

		Context("when request is valid", func() {
			It("should be passed", func() {
				consumer := &pb.MicroServiceKey{
					ServiceName: "create_dep_consumer",
					AppId:       "create_dep_group",
					Version:     "1.0.0",
				}

				By("add latest")
				respCreateDependency, err := serviceResource.CreateDependenciesForMicroServices(getContext(), &pb.CreateDependenciesRequest{
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
				Expect(respCreateDependency.Response.GetCode()).To(Equal(pb.ResponseSuccess))

				DependencyHandle()

				respPro, err := serviceResource.GetConsumerDependencies(getContext(), &pb.GetDependenciesRequest{
					ServiceId: consumerId1,
				})
				Expect(err).To(BeNil())
				Expect(respPro.Response.GetCode()).To(Equal(pb.ResponseSuccess))
				Expect(len(respPro.Providers)).To(Equal(1))
				Expect(respPro.Providers[0].Version).To(Equal("1.0.1"))

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
				Expect(respCreateDependency.Response.GetCode()).To(Equal(pb.ResponseSuccess))

				DependencyHandle()

				respPro, err = serviceResource.GetConsumerDependencies(getContext(), &pb.GetDependenciesRequest{
					ServiceId: consumerId1,
				})
				Expect(err).To(BeNil())
				Expect(respPro.Response.GetCode()).To(Equal(pb.ResponseSuccess))
				Expect(len(respPro.Providers)).To(Equal(2))

				By("add *")
				respCreateDependency, err = serviceResource.CreateDependenciesForMicroServices(getContext(), &pb.CreateDependenciesRequest{
					Dependencies: []*pb.ConsumerDependency{
						{
							Consumer: &pb.MicroServiceKey{
								ServiceName: "create_dep_consumer_all",
								AppId:       "create_dep_group",
								Version:     "1.0.0",
							},
							Providers: []*pb.MicroServiceKey{
								{
									ServiceName: "*",
								},
							},
						},
					},
				})
				Expect(err).To(BeNil())
				Expect(respCreateDependency.Response.GetCode()).To(Equal(pb.ResponseSuccess))

				DependencyHandle()

				respPro, err = serviceResource.GetConsumerDependencies(getContext(), &pb.GetDependenciesRequest{
					ServiceId: consumerId3,
				})
				Expect(err).To(BeNil())
				Expect(respPro.Response.GetCode()).To(Equal(pb.ResponseSuccess))
				Expect(len(respPro.Providers)).ToNot(Equal(0))

				By("clean all")
				respCreateDependency, err = serviceResource.CreateDependenciesForMicroServices(getContext(), &pb.CreateDependenciesRequest{
					Dependencies: []*pb.ConsumerDependency{
						{
							Consumer: &pb.MicroServiceKey{
								ServiceName: "create_dep_consumer_all",
								AppId:       "create_dep_group",
								Version:     "1.0.0",
							},
							Providers: nil,
						},
					},
				})
				Expect(err).To(BeNil())
				Expect(respCreateDependency.Response.GetCode()).To(Equal(pb.ResponseSuccess))

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
				Expect(respCreateDependency.Response.GetCode()).To(Equal(pb.ResponseSuccess))

				By("add 1.0.0-2.0.0 to override *")
				respCreateDependency, err = serviceResource.CreateDependenciesForMicroServices(getContext(), &pb.CreateDependenciesRequest{
					Dependencies: []*pb.ConsumerDependency{
						{
							Consumer: consumer,
							Providers: []*pb.MicroServiceKey{
								{
									AppId:       "create_dep_group",
									ServiceName: "create_dep_provider",
									Version:     "1.0.0-1.0.1",
								},
							},
						},
					},
				})
				Expect(err).To(BeNil())
				Expect(respCreateDependency.Response.GetCode()).To(Equal(pb.ResponseSuccess))

				DependencyHandle()

				respPro, err = serviceResource.GetConsumerDependencies(getContext(), &pb.GetDependenciesRequest{
					ServiceId: consumerId1,
				})
				Expect(err).To(BeNil())
				Expect(respPro.Response.GetCode()).To(Equal(pb.ResponseSuccess))
				Expect(len(respPro.Providers)).To(Equal(1))
				Expect(respPro.Providers[0].Version).To(Equal("1.0.0"))

				By("add not override")
				respAddDependency, err := serviceResource.AddDependenciesForMicroServices(getContext(), &pb.AddDependenciesRequest{
					Dependencies: []*pb.ConsumerDependency{
						{
							Consumer: consumer,
							Providers: []*pb.MicroServiceKey{
								{
									AppId:       "create_dep_group",
									ServiceName: "create_dep_provider",
									Version:     "1.0.0-3.0.0",
								},
							},
						},
					},
				})
				Expect(err).To(BeNil())
				Expect(respAddDependency.Response.GetCode()).To(Equal(pb.ResponseSuccess))

				DependencyHandle()

				respPro, err = serviceResource.GetConsumerDependencies(getContext(), &pb.GetDependenciesRequest{
					ServiceId: consumerId1,
				})
				Expect(err).To(BeNil())
				Expect(respPro.Response.GetCode()).To(Equal(pb.ResponseSuccess))
				Expect(len(respPro.Providers)).To(Equal(2))

				By("add provider is empty")
				respCreateDependency, err = serviceResource.CreateDependenciesForMicroServices(getContext(), &pb.CreateDependenciesRequest{
					Dependencies: []*pb.ConsumerDependency{
						{
							Consumer:  consumer,
							Providers: []*pb.MicroServiceKey{},
						},
					},
				})
				Expect(err).To(BeNil())
				Expect(respCreateDependency.Response.GetCode()).To(Equal(pb.ResponseSuccess))
				respCreateDependency, err = serviceResource.CreateDependenciesForMicroServices(getContext(), &pb.CreateDependenciesRequest{
					Dependencies: []*pb.ConsumerDependency{
						{
							Consumer: consumer,
						},
					},
				})
				Expect(err).To(BeNil())
				Expect(respCreateDependency.Response.GetCode()).To(Equal(pb.ResponseSuccess))

				DependencyHandle()

				respPro, err = serviceResource.GetConsumerDependencies(getContext(), &pb.GetDependenciesRequest{
					ServiceId: consumerId1,
				})
				Expect(err).To(BeNil())
				Expect(respPro.Response.GetCode()).To(Equal(pb.ResponseSuccess))
				Expect(len(respPro.Providers)).To(Equal(0))
			})
		})
	})

	Describe("execute 'get' operartion", func() {
		var (
			consumerId1 string
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
			Expect(respCreateService.Response.GetCode()).To(Equal(pb.ResponseSuccess))
			consumerId1 = respCreateService.ServiceId

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
			Expect(respCreateService.Response.GetCode()).To(Equal(pb.ResponseSuccess))
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
			Expect(respCreateService.Response.GetCode()).To(Equal(pb.ResponseSuccess))
			providerId2 = respCreateService.ServiceId
		})

		Context("when request is invalid", func() {
			It("should be failed", func() {
				By("service id is empty when get provider")
				respPro, err := serviceResource.GetProviderDependencies(getContext(), &pb.GetDependenciesRequest{
					ServiceId: "",
				})
				Expect(err).To(BeNil())
				Expect(respPro.Response.GetCode()).ToNot(Equal(pb.ResponseSuccess))

				By("service does not exist when get provider")
				respPro, err = serviceResource.GetProviderDependencies(getContext(), &pb.GetDependenciesRequest{
					ServiceId: "noneservice",
				})
				Expect(err).To(BeNil())
				Expect(respPro.Response.GetCode()).ToNot(Equal(pb.ResponseSuccess))

				By("service id is empty when get consumer")
				respCon, err := serviceResource.GetConsumerDependencies(getContext(), &pb.GetDependenciesRequest{
					ServiceId: "",
				})
				Expect(err).To(BeNil())
				Expect(respCon.Response.GetCode()).ToNot(Equal(pb.ResponseSuccess))

				By("service does not exist when get consumer")
				respCon, err = serviceResource.GetConsumerDependencies(getContext(), &pb.GetDependenciesRequest{
					ServiceId: "noneservice",
				})
				Expect(err).To(BeNil())
				Expect(respCon.Response.GetCode()).ToNot(Equal(pb.ResponseSuccess))
			})
		})

		Context("when request is valid", func() {
			It("should be passed", func() {
				By("get provider")
				respPro, err := serviceResource.GetProviderDependencies(getContext(), &pb.GetDependenciesRequest{
					ServiceId: providerId1,
				})
				Expect(err).To(BeNil())
				Expect(respPro.Response.GetCode()).To(Equal(pb.ResponseSuccess))

				By("get consumer")
				respCon, err := serviceResource.GetConsumerDependencies(getContext(), &pb.GetDependenciesRequest{
					ServiceId: consumerId1,
				})
				Expect(err).To(BeNil())
				Expect(respCon.Response.GetCode()).To(Equal(pb.ResponseSuccess))
			})
		})

		Context("when after finding instance", func() {
			It("should created dependencies between C and P", func() {
				By("find provider")
				resp, err := disco.FindInstances(getContext(), &pb.FindInstancesRequest{
					ConsumerServiceId: consumerId1,
					AppId:             "get_dep_group",
					ServiceName:       "get_dep_provider",
					VersionRule:       "1.0.0+",
				})
				Expect(err).To(BeNil())
				Expect(resp.Response.GetCode()).To(Equal(pb.ResponseSuccess))

				DependencyHandle()

				By("get consumer's deps")
				respGetP, err := serviceResource.GetProviderDependencies(getContext(), &pb.GetDependenciesRequest{
					ServiceId: providerId1,
				})
				Expect(err).To(BeNil())
				Expect(respGetP.Response.GetCode()).To(Equal(pb.ResponseSuccess))
				Expect(len(respGetP.Consumers)).NotTo(Equal(0))
				Expect(respGetP.Consumers[0].ServiceId).To(Equal(consumerId1))

				By("get provider's deps")
				respGetC, err := serviceResource.GetConsumerDependencies(getContext(), &pb.GetDependenciesRequest{
					ServiceId: consumerId1,
				})
				Expect(err).To(BeNil())
				Expect(respGetC.Response.GetCode()).To(Equal(pb.ResponseSuccess))
				Expect(len(respGetC.Providers)).To(Equal(2))

				//重复find
				resp, err = disco.FindInstances(getContext(), &pb.FindInstancesRequest{
					ConsumerServiceId: consumerId1,
					AppId:             "get_dep_group",
					ServiceName:       "get_dep_provider",
					VersionRule:       "2.0.0+",
				})
				Expect(err).To(BeNil())
				Expect(resp.Response.GetCode()).To(Equal(pb.ResponseSuccess))

				DependencyHandle()

				By("get consumer again")
				respGetP, err = serviceResource.GetProviderDependencies(getContext(), &pb.GetDependenciesRequest{
					ServiceId: providerId1,
				})
				Expect(err).To(BeNil())
				Expect(respGetP.Response.GetCode()).To(Equal(pb.ResponseSuccess))
				Expect(len(respGetP.Consumers)).To(Equal(0))

				By("get provider again")
				respGetC, err = serviceResource.GetConsumerDependencies(getContext(), &pb.GetDependenciesRequest{
					ServiceId: consumerId1,
				})
				Expect(err).To(BeNil())
				Expect(respGetC.Response.GetCode()).To(Equal(pb.ResponseSuccess))
				Expect(len(respGetC.Providers)).To(Equal(1))
				Expect(respGetC.Providers[0].ServiceId).To(Equal(providerId2))

				By("get self deps")
				resp, err = disco.FindInstances(getContext(), &pb.FindInstancesRequest{
					ConsumerServiceId: consumerId1,
					AppId:             "get_dep_group",
					ServiceName:       "get_dep_consumer",
					VersionRule:       "1.0.0+",
				})
				Expect(err).To(BeNil())
				Expect(resp.Response.GetCode()).To(Equal(pb.ResponseSuccess))

				DependencyHandle()

				respGetC, err = serviceResource.GetConsumerDependencies(getContext(), &pb.GetDependenciesRequest{
					ServiceId: consumerId1,
					NoSelf:    true,
				})
				Expect(err).To(BeNil())
				Expect(respGetC.Response.GetCode()).To(Equal(pb.ResponseSuccess))
				Expect(len(respGetC.Providers)).To(Equal(1))

				By("find before provider register")
				resp, err = disco.FindInstances(getContext(), &pb.FindInstancesRequest{
					ConsumerServiceId: providerId2,
					AppId:             "get_dep_group",
					ServiceName:       "get_dep_finder",
					VersionRule:       "1.0.0+",
				})
				Expect(err).To(BeNil())
				Expect(resp.Response.GetCode()).To(Equal(pb.ErrServiceNotExists))

				respCreateF, err := serviceResource.Create(getContext(), &pb.CreateServiceRequest{
					Service: &pb.MicroService{
						AppId:       "get_dep_group",
						ServiceName: "get_dep_finder",
						Version:     "1.0.0",
						Level:       "FRONT",
						Status:      pb.MS_UP,
					},
				})
				Expect(err).To(BeNil())
				Expect(respCreateF.Response.GetCode()).To(Equal(pb.ResponseSuccess))
				finder1 := respCreateF.ServiceId

				resp, err = disco.FindInstances(getContext(), &pb.FindInstancesRequest{
					ConsumerServiceId: providerId2,
					AppId:             "get_dep_group",
					ServiceName:       "get_dep_finder",
					VersionRule:       "1.0.0+",
				})
				Expect(err).To(BeNil())
				Expect(resp.Response.GetCode()).To(Equal(pb.ResponseSuccess))

				DependencyHandle()

				respGetC, err = serviceResource.GetConsumerDependencies(getContext(), &pb.GetDependenciesRequest{
					ServiceId: providerId2,
				})
				Expect(err).To(BeNil())
				Expect(respGetC.Response.GetCode()).To(Equal(pb.ResponseSuccess))
				Expect(len(respGetC.Providers)).To(Equal(1))
				Expect(respGetC.Providers[0].ServiceId).To(Equal(finder1))

				By("find after delete micro service")
				respDelP, err := serviceResource.Delete(getContext(), &pb.DeleteServiceRequest{
					ServiceId: finder1, Force: true,
				})
				Expect(err).To(BeNil())
				Expect(respDelP.Response.GetCode()).To(Equal(pb.ResponseSuccess))

				DependencyHandle()

				respGetC, err = serviceResource.GetConsumerDependencies(getContext(), &pb.GetDependenciesRequest{
					ServiceId: providerId2,
				})
				Expect(err).To(BeNil())
				Expect(respGetC.Response.GetCode()).To(Equal(pb.ResponseSuccess))
				Expect(len(respGetC.Providers)).To(Equal(0))

				respCreateF, err = serviceResource.Create(getContext(), &pb.CreateServiceRequest{
					Service: &pb.MicroService{
						ServiceId:   finder1,
						AppId:       "get_dep_group",
						ServiceName: "get_dep_finder",
						Version:     "1.0.0",
						Level:       "FRONT",
						Status:      pb.MS_UP,
					},
				})
				Expect(err).To(BeNil())
				Expect(respCreateF.Response.GetCode()).To(Equal(pb.ResponseSuccess))

				resp, err = disco.FindInstances(getContext(), &pb.FindInstancesRequest{
					ConsumerServiceId: providerId2,
					AppId:             "get_dep_group",
					ServiceName:       "get_dep_finder",
					VersionRule:       "1.0.0+",
				})
				Expect(err).To(BeNil())
				Expect(resp.Response.GetCode()).To(Equal(pb.ResponseSuccess))

				DependencyHandle()

				respGetC, err = serviceResource.GetConsumerDependencies(getContext(), &pb.GetDependenciesRequest{
					ServiceId: providerId2,
				})
				Expect(err).To(BeNil())
				Expect(respGetC.Response.GetCode()).To(Equal(pb.ResponseSuccess))
				Expect(len(respGetC.Providers)).To(Equal(1))
				Expect(respGetC.Providers[0].ServiceId).To(Equal(finder1))
			})
		})
	})
})

func DependencyHandle() {
	t := archaius.Get("TEST_MODE")
	if t == nil {
		t = "etcd"
	}
	if t == "etcd" {
		for {
			Expect(deh.Handle()).To(BeNil())

			key := path.GetServiceDependencyQueueRootKey("")
			resp, err := kv.Store().DependencyQueue().Search(getContext(),
				client.WithStrKey(key), client.WithPrefix(), client.WithCountOnly())

			Expect(err).To(BeNil())

			// maintain dependency rules.
			if resp.Count == 0 {
				break
			}
		}
	}
}
