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

var _ = Describe("'Dependency' service", func() {
	Describe("execute 'create' operartion", func() {
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
		Expect(respCreateService.GetResponse().Code).To(Equal(pb.Response_SUCCESS))
		consumerId := respCreateService.ServiceId

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
		Expect(respCreateService.GetResponse().Code).To(Equal(pb.Response_SUCCESS))
		providerId := respCreateService.ServiceId

		Context("when request is invalid", func() {
			It("should be failed", func() {
				By("dependency is nil")
				respCreateDependency, err := serviceResource.CreateDependenciesForMircServices(getContext(), &pb.CreateDependenciesRequest{})
				Expect(err).To(BeNil())
				Expect(respCreateDependency.GetResponse().Code).ToNot(Equal(pb.Response_SUCCESS))

				By("consumer appId is empty")
				respCreateDependency, err = serviceResource.CreateDependenciesForMircServices(getContext(), &pb.CreateDependenciesRequest{
					Dependencies: []*pb.MircroServiceDependency{
						{
							Consumer: &pb.DependencyMircroService{
								ServiceName: "create_dep_consumer",
								AppId:       "",
								Version:     "3.0.0",
							},
							Providers: nil,
						},
					},
				})
				Expect(respCreateDependency.GetResponse().Code).ToNot(Equal(pb.Response_SUCCESS))

				respCreateDependency, err = serviceResource.CreateDependenciesForMircServices(getContext(), &pb.CreateDependenciesRequest{
					Dependencies: []*pb.MircroServiceDependency{
						{
							Consumer: &pb.DependencyMircroService{
								ServiceName: "service_name_not_exist",
								AppId:       "service_group_not_exist",
								Version:     "3.0.0",
							},
							Providers: []*pb.DependencyMircroService{
								{
									AppId:       "service_group_provider",
									ServiceName: "service_name_provider",
									Version:     "2.0.0",
								},
								{
									AppId:       "service_group_consumer",
									ServiceName: "service_name_consumer",
									Version:     "2.0.0",
								},
							},
						},
					},
				})
				Expect(err).To(BeNil())
				Expect(respCreateDependency.GetResponse().Code).ToNot(Equal(pb.Response_SUCCESS))

				respCreateDependency, err = serviceResource.CreateDependenciesForMircServices(getContext(), &pb.CreateDependenciesRequest{
					Dependencies: []*pb.MircroServiceDependency{
						{
							Consumer: &pb.DependencyMircroService{
								ServiceName: "service_name_consumer",
								AppId:       "service_group_consumer",
								Version:     "6.0.0",
							},
							Providers: []*pb.DependencyMircroService{
								{
									AppId:       "service_group_provider",
									ServiceName: "service_name_provider",
									Version:     "2.0.0",
								},
							},
						},
					},
				})
				Expect(err).To(BeNil())
				Expect(respCreateDependency.GetResponse().Code).To(Equal(pb.Response_SUCCESS))

				respCreateDependency, err = serviceResource.CreateDependenciesForMircServices(getContext(), &pb.CreateDependenciesRequest{
					Dependencies: []*pb.MircroServiceDependency{
						{
							Consumer: &pb.DependencyMircroService{
								ServiceName: "*",
								AppId:       "service_group_consumer",
								Version:     "6.0.0",
							},
							Providers: []*pb.DependencyMircroService{
								{
									AppId:       "service_group_provider",
									ServiceName: "service_name_provider",
									Version:     "2.0.0",
								},
							},
						},
					},
				})
				Expect(err).To(BeNil())
				Expect(respCreateDependency.GetResponse().Code).ToNot(Equal(pb.Response_SUCCESS))

				respCreateDependency, err = serviceResource.CreateDependenciesForMircServices(getContext(), &pb.CreateDependenciesRequest{
					Dependencies: []*pb.MircroServiceDependency{
						{
							Consumer: &pb.DependencyMircroService{
								ServiceName: "service_name_consumer",
								AppId:       "service_group_consumer",
								Version:     "6.0.0+",
							},
							Providers: []*pb.DependencyMircroService{
								{
									AppId:       "service_group_provider",
									ServiceName: "service_name_provider",
									Version:     "2.0.0",
								},
							},
						},
					},
				})
				Expect(err).To(BeNil())
				Expect(respCreateDependency.GetResponse().Code).ToNot(Equal(pb.Response_SUCCESS))

				respCreateDependency, err = serviceResource.CreateDependenciesForMircServices(getContext(), &pb.CreateDependenciesRequest{
					Dependencies: []*pb.MircroServiceDependency{
						{
							Consumer: &pb.DependencyMircroService{
								ServiceName: "service_name_consumer",
								AppId:       "service_group_consumer",
								Version:     "6.0.0",
							},
							Providers: []*pb.DependencyMircroService{
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
				Expect(respCreateDependency.GetResponse().Code).ToNot(Equal(pb.Response_SUCCESS))

				respCreateDependency, err = serviceResource.CreateDependenciesForMircServices(getContext(), &pb.CreateDependenciesRequest{
					Dependencies: []*pb.MircroServiceDependency{
						{
							Consumer: &pb.DependencyMircroService{
								ServiceName: "service_name_consumer",
								AppId:       "service_group_consumer",
								Version:     "6.0.0",
							},
							Providers: []*pb.DependencyMircroService{
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
				Expect(respCreateDependency.GetResponse().Code).ToNot(Equal(pb.Response_SUCCESS))

				respCreateDependency, err = serviceResource.CreateDependenciesForMircServices(getContext(), &pb.CreateDependenciesRequest{
					Dependencies: []*pb.MircroServiceDependency{
						{
							Consumer: &pb.DependencyMircroService{
								ServiceName: "service_name_consumer",
								AppId:       "service_group_consumer",
								Version:     "6.0.0",
							},
							Providers: []*pb.DependencyMircroService{
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
				Expect(respCreateDependency.GetResponse().Code).ToNot(Equal(pb.Response_SUCCESS))
			})
			It("创建Dependency", func() {
				fmt.Println("UT===========创建Dependency")
				resp, err := serviceResource.Create(getContext(), &pb.CreateServiceRequest{
					Service: &pb.MicroService{
						ServiceName: "service_name_consumer",
						AppId:       "service_group_consumer",
						Version:     "2.0.0",
						Level:       "FRONT",
						Schemas: []string{
							"xxxxxxxx",
						},
						Status: "UP",
					},
				})
				Expect(err).To(BeNil())
				consumerId = resp.ServiceId
				Expect(resp.GetResponse().Code).To(Equal(pb.Response_SUCCESS))

				respCreateDependency, err := serviceResource.CreateDependenciesForMircServices(getContext(), &pb.CreateDependenciesRequest{
					Dependencies: []*pb.MircroServiceDependency{
						{
							Consumer: &pb.DependencyMircroService{
								AppId:       "service_group_consumer",
								ServiceName: "service_name_consumer",
								Version:     "2.0.0",
							},
							Providers: []*pb.DependencyMircroService{
								{
									AppId:       "service_group_provider",
									ServiceName: "service_name_provider",
									Version:     "latest",
								},
							},
						},
					},
				})

				resp, err = serviceResource.Create(getContext(), &pb.CreateServiceRequest{
					Service: &pb.MicroService{
						ServiceName: "service_name_provider",
						AppId:       "service_group_provider",
						Version:     "3.0.0",
						Level:       "FRONT",
						Schemas: []string{
							"xxxxxxxx",
						},
						Status:     "UP",
						Properties: map[string]string{"allowCrossApp": "true"},
					},
				})
				Expect(err).To(BeNil())
				providerId = resp.ServiceId
				Expect(resp.GetResponse().Code).To(Equal(pb.Response_SUCCESS))

				respCreateDependency, err = serviceResource.CreateDependenciesForMircServices(getContext(), &pb.CreateDependenciesRequest{
					Dependencies: []*pb.MircroServiceDependency{
						{
							Consumer: &pb.DependencyMircroService{
								AppId:       "service_group_consumer",
								ServiceName: "service_name_consumer",
								Version:     "2.0.0",
							},
							Providers: []*pb.DependencyMircroService{
								{
									AppId:       "service_group_consumer2",
									ServiceName: "service_name_consumer2",
									Version:     "2.0.0+",
								},
							},
						},
					},
				})
				Expect(err).To(BeNil())
				Expect(resp.GetResponse().Code).To(Equal(pb.Response_SUCCESS))

				resp, err = serviceResource.Create(getContext(), &pb.CreateServiceRequest{
					Service: &pb.MicroService{
						ServiceName: "service_name_consumer2",
						AppId:       "service_group_consumer2",
						Version:     "3.0.0",
						Level:       "FRONT",
						Schemas: []string{
							"xxxxxxxx",
						},
						Status:     "UP",
						Properties: map[string]string{"allowCrossApp": "true"},
					},
				})
				Expect(err).To(BeNil())
				consumerId2 = resp.ServiceId
				Expect(resp.GetResponse().Code).To(Equal(pb.Response_SUCCESS))

				fmt.Println("UT===========创建Dependency，serviceName 为*")

				respCreateDependency, err = serviceResource.CreateDependenciesForMircServices(getContext(), &pb.CreateDependenciesRequest{
					Dependencies: []*pb.MircroServiceDependency{
						{
							Consumer: &pb.DependencyMircroService{
								AppId:       "service_group_consumer",
								ServiceName: "service_name_consumer",
								Version:     "2.0.0",
							},
							Providers: []*pb.DependencyMircroService{
								{
									AppId:       "service_group_provider",
									ServiceName: "*",
									Version:     "2.0.0",
								},
							},
						},
					},
				})
				Expect(err).To(BeNil())
				Expect(respCreateDependency.GetResponse().Code).To(Equal(pb.Response_SUCCESS))

				respCreateDependency, err = serviceResource.CreateDependenciesForMircServices(getContext(), &pb.CreateDependenciesRequest{
					Dependencies: []*pb.MircroServiceDependency{
						{
							Consumer: &pb.DependencyMircroService{
								AppId:       "service_group_consumer",
								ServiceName: "service_name_consumer",
								Version:     "2.0.0",
							},
							Providers: []*pb.DependencyMircroService{
								{
									AppId:       "service_group_provider",
									ServiceName: "service_name_provider",
									Version:     "2.0.0",
								},
								{
									AppId:       "service_group_consumer",
									ServiceName: "*",
									Version:     "2.0.0",
								},
							},
						},
					},
				})
				Expect(err).To(BeNil())
				Expect(respCreateDependency.GetResponse().Code).To(Equal(pb.Response_SUCCESS))

				respCreateDependency, err = serviceResource.CreateDependenciesForMircServices(getContext(), &pb.CreateDependenciesRequest{
					Dependencies: []*pb.MircroServiceDependency{
						{
							Consumer: &pb.DependencyMircroService{
								ServiceName: "service_name_consumer2",
								AppId:       "service_group_consumer2",
								Version:     "3.0.0",
							},
							Providers: []*pb.DependencyMircroService{
								{
									AppId:       "service_group_provider",
									ServiceName: "service_name_provider",
									Version:     "2.0.0",
								},
								{
									AppId:       "service_group_consumer",
									ServiceName: "service_name_consumer",
									Version:     "2.0.0",
								},
							},
						},
					},
				})
				Expect(err).To(BeNil())
				Expect(respCreateDependency.GetResponse().Code).To(Equal(pb.Response_SUCCESS))
			})

			It("删除微服务,实例存在，存在依赖不能删除", func() {
				resp, err := serviceResource.Delete(getContext(), &pb.DeleteServiceRequest{
					ServiceId: consumerId,
					Force:     false,
				})
				Expect(err).To(BeNil())
				Expect(resp.GetResponse().Code).ToNot(Equal(pb.Response_SUCCESS))
			})

			It("查询provider对应的所有consumer,参数校验", func() {
				fmt.Println("UT===========查询privider的所有consumer,参数校验")
				resp, err := serviceResource.GetProviderDependencies(getContext(), &pb.GetDependenciesRequest{
					ServiceId: "",
				})
				Expect(err).To(BeNil())
				Expect(resp.GetResponse().Code).ToNot(Equal(pb.Response_SUCCESS))

				resp, err = serviceResource.GetProviderDependencies(getContext(), &pb.GetDependenciesRequest{
					ServiceId: "noneservice",
				})
				Expect(err).To(BeNil())
				Expect(resp.GetResponse().Code).ToNot(Equal(pb.Response_SUCCESS))
			})

			It("查询provider对应的所有consumer", func() {
				fmt.Println("UT===========查询privider的所有consumer")
				resp, err := serviceResource.GetProviderDependencies(getContext(), &pb.GetDependenciesRequest{
					ServiceId: providerId,
				})
				Expect(err).To(BeNil())
				Expect(resp.GetResponse().Code).To(Equal(pb.Response_SUCCESS))
			})

			It("查询consumer对应的所有provider,参数校验", func() {
				fmt.Println("UT===========查询consumer的所有privider")
				resp, err := serviceResource.GetConsumerDependencies(getContext(), &pb.GetDependenciesRequest{
					ServiceId: "",
				})
				Expect(err).To(BeNil())
				Expect(resp.GetResponse().Code).ToNot(Equal(pb.Response_SUCCESS))

				resp, err = serviceResource.GetConsumerDependencies(getContext(), &pb.GetDependenciesRequest{
					ServiceId: "noneservice",
				})
				Expect(err).To(BeNil())
				Expect(resp.GetResponse().Code).ToNot(Equal(pb.Response_SUCCESS))
			})

			It("查询consumer对应的所有provider", func() {
				fmt.Println("UT===========查询consumer的所有privider")
				resp, err := serviceResource.GetConsumerDependencies(getContext(), &pb.GetDependenciesRequest{
					ServiceId: consumerId,
				})
				Expect(err).To(BeNil())
				Expect(resp.GetResponse().Code).To(Equal(pb.Response_SUCCESS))
			})
			It("修改Dependency", func() {
				fmt.Println("UT===========修改Dependency")
				resp, err := serviceResource.CreateDependenciesForMircServices(getContext(), &pb.CreateDependenciesRequest{
					Dependencies: []*pb.MircroServiceDependency{
						{
							Consumer: &pb.DependencyMircroService{
								AppId:       "service_group_consumer",
								ServiceName: "service_name_consumer",
								Version:     "2.0.0",
							},
							Providers: []*pb.DependencyMircroService{
								{
									AppId:       "service_group_provider",
									ServiceName: "service_name_provider",
									Version:     "2.0.0",
								},
							},
						},
					},
				})
				Expect(err).To(BeNil())
				Expect(resp.GetResponse().Code).To(Equal(pb.Response_SUCCESS))

				resp, err = serviceResource.CreateDependenciesForMircServices(getContext(), &pb.CreateDependenciesRequest{
					Dependencies: []*pb.MircroServiceDependency{
						{
							Consumer: &pb.DependencyMircroService{
								AppId:       "service_group_consumer",
								ServiceName: "service_name_consumer",
								Version:     "2.0.0",
							},
							Providers: []*pb.DependencyMircroService{
								{
									AppId:       "service_group_provider",
									ServiceName: "service_name_provider",
									Version:     "2.0.0+",
								},
							},
						},
					},
				})
				Expect(err).To(BeNil())
				Expect(resp.GetResponse().Code).To(Equal(pb.Response_SUCCESS))

				resp, err = serviceResource.CreateDependenciesForMircServices(getContext(), &pb.CreateDependenciesRequest{
					Dependencies: []*pb.MircroServiceDependency{
						{
							Consumer: &pb.DependencyMircroService{
								AppId:       "service_group_consumer",
								ServiceName: "service_name_consumer",
								Version:     "2.0.0",
							},
							Providers: []*pb.DependencyMircroService{
								{
									AppId:       "service_group_provider",
									ServiceName: "service_name_provider_bk",
									Version:     "2.0.0",
								},
							},
						},
					},
				})
				Expect(err).To(BeNil())
				Expect(resp.GetResponse().Code).To(Equal(pb.Response_SUCCESS))

				resp, err = serviceResource.CreateDependenciesForMircServices(getContext(), &pb.CreateDependenciesRequest{
					Dependencies: []*pb.MircroServiceDependency{
						{
							Consumer: &pb.DependencyMircroService{
								AppId:       "service_group_consumer",
								ServiceName: "service_name_consumer",
								Version:     "2.0.0",
							},
							Providers: []*pb.DependencyMircroService{
								{
									AppId:       "service_group_provider",
									ServiceName: "service_name_provider",
									Version:     "3.0.0",
								},
							},
						},
					},
				})
				Expect(err).To(BeNil())
				Expect(resp.GetResponse().Code).To(Equal(pb.Response_SUCCESS))

				resp, err = serviceResource.CreateDependenciesForMircServices(getContext(), &pb.CreateDependenciesRequest{
					Dependencies: []*pb.MircroServiceDependency{
						{
							Consumer: &pb.DependencyMircroService{
								AppId:       "service_group_consumer",
								ServiceName: "service_name_consumer",
								Version:     "2.0.0",
							},
							Providers: []*pb.DependencyMircroService{
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
				Expect(resp.GetResponse().Code).To(Equal(pb.Response_SUCCESS))

				resp, err = serviceResource.CreateDependenciesForMircServices(getContext(), &pb.CreateDependenciesRequest{
					Dependencies: []*pb.MircroServiceDependency{
						{
							Consumer: &pb.DependencyMircroService{
								AppId:       "service_group_consumer",
								ServiceName: "service_name_consumer",
								Version:     "2.0.0",
							},
							Providers: []*pb.DependencyMircroService{
								{
									AppId:       "service_group_provider",
									ServiceName: "service_name_provider",
									Version:     "3.0.0",
								},
							},
						},
					},
				})
				Expect(err).To(BeNil())
				Expect(resp.GetResponse().Code).To(Equal(pb.Response_SUCCESS))

				resp, err = serviceResource.CreateDependenciesForMircServices(getContext(), &pb.CreateDependenciesRequest{
					Dependencies: []*pb.MircroServiceDependency{
						{
							Consumer: &pb.DependencyMircroService{
								AppId:       "service_group_consumer",
								ServiceName: "service_name_consumer",
								Version:     "2.0.0",
							},
							Providers: []*pb.DependencyMircroService{
								{
									AppId:       "service_group_provider",
									ServiceName: "service_name_provider",
									Version:     "2.0.0-3.0.0",
								},
							},
						},
					},
				})
				Expect(err).To(BeNil())
				Expect(resp.GetResponse().Code).To(Equal(pb.Response_SUCCESS))

				resp, err = serviceResource.CreateDependenciesForMircServices(getContext(), &pb.CreateDependenciesRequest{
					Dependencies: []*pb.MircroServiceDependency{
						{
							Consumer: &pb.DependencyMircroService{
								AppId:       "service_group_consumer",
								ServiceName: "service_name_consumer",
								Version:     "2.0.0",
							},
							Providers: []*pb.DependencyMircroService{
								{
									AppId:       "",
									ServiceName: "*",
									Version:     "",
								},
							},
						},
					},
				})
				Expect(err).To(BeNil())
				Expect(resp.GetResponse().Code).To(Equal(pb.Response_SUCCESS))
			})
			It("Find 添加依赖", func() {
				fmt.Println("UT===========find 添加依赖")
				resp, err := instanceResource.Find(getContext(), &pb.FindInstancesRequest{
					ConsumerServiceId: consumerId,
					AppId:             "service_group_provider",
					ServiceName:       "service_name_provider",
					VersionRule:       "latest",
				})
				Expect(err).To(BeNil())
				Expect(resp.GetResponse().Code).To(Equal(pb.Response_SUCCESS))

				//重复find
				resp, err = instanceResource.Find(getContext(), &pb.FindInstancesRequest{
					ConsumerServiceId: consumerId,
					AppId:             "service_group_provider",
					ServiceName:       "service_name_provider",
					VersionRule:       "latest",
				})
				Expect(err).To(BeNil())
				Expect(resp.GetResponse().Code).To(Equal(pb.Response_SUCCESS))
			})

			It("删除Dependency", func() {
				fmt.Println("UT===========删除Dependency")
				resp, err := serviceResource.Delete(getContext(), &pb.DeleteServiceRequest{
					ServiceId: consumerId,
					Force:     true,
				})
				Expect(err).To(BeNil())
				Expect(resp.GetResponse().Code).To(Equal(pb.Response_SUCCESS))

				resp, err = serviceResource.Delete(getContext(), &pb.DeleteServiceRequest{
					ServiceId: providerId,
					Force:     true,
				})
				Expect(err).To(BeNil())
				Expect(resp.GetResponse().Code).To(Equal(pb.Response_SUCCESS))

				resp, err = serviceResource.Delete(getContext(), &pb.DeleteServiceRequest{
					ServiceId: consumerId2,
					Force:     true,
				})
				Expect(err).To(BeNil())
				Expect(resp.GetResponse().Code).To(Equal(pb.Response_SUCCESS))
			})
		})

		Context("dep find", func() {
			var providerIdDep1, providerIdDep2, providerIdDep3, consumerIdDep string
			It("find接口， 建立依赖关系", func() {
				resp, err := serviceResource.Create(getContext(), &pb.CreateServiceRequest{
					Service: &pb.MicroService{
						ServiceName: "serviceName_consumer1",
						AppId:       "appId_consumer1",
						Version:     "2.0.0",
						Level:       "FRONT",
						Schemas: []string{
							"com.huawei.test",
						},
						Status: "UP",
					},
				})
				Expect(err).To(BeNil())
				Expect(resp.GetResponse().Code).To(Equal(pb.Response_SUCCESS))
				consumerIdDep = resp.ServiceId

				resp, err = serviceResource.Create(getContext(), &pb.CreateServiceRequest{
					Service: &pb.MicroService{
						ServiceName: "serviceName_provider1",
						AppId:       "appId_provider1",
						Version:     "2.0.0",
						Level:       "FRONT",
						Schemas: []string{
							"com.huawei.test",
						},
						Status: "UP",
					},
				})
				Expect(err).To(BeNil())
				Expect(resp.GetResponse().Code).To(Equal(pb.Response_SUCCESS))
				providerIdDep1 = resp.ServiceId

				resp, err = serviceResource.Create(getContext(), &pb.CreateServiceRequest{
					Service: &pb.MicroService{
						ServiceName: "serviceName_provider2",
						AppId:       "appId_provider2",
						Version:     "2.0.0",
						Level:       "FRONT",
						Schemas: []string{
							"com.huawei.test",
						},
						Status: "UP",
					},
				})
				Expect(err).To(BeNil())
				Expect(resp.GetResponse().Code).To(Equal(pb.Response_SUCCESS))
				providerIdDep2 = resp.ServiceId

				respFind, err := instanceResource.Find(getContext(), &pb.FindInstancesRequest{
					ConsumerServiceId: consumerIdDep,
					AppId:             "appId_provider1",
					ServiceName:       "serviceName_provider1",
					VersionRule:       "latest",
				})
				Expect(err).To(BeNil())
				Expect(respFind.GetResponse().Code).To(Equal(pb.Response_SUCCESS))

				respFind, err = instanceResource.Find(getContext(), &pb.FindInstancesRequest{
					ConsumerServiceId: consumerIdDep,
					AppId:             "appId_provider2",
					ServiceName:       "serviceName_provider2",
					VersionRule:       "latest",
				})
				Expect(err).To(BeNil())
				Expect(respFind.GetResponse().Code).To(Equal(pb.Response_SUCCESS))

				respPro, err := serviceResource.GetConsumerDependencies(getContext(), &pb.GetDependenciesRequest{
					ServiceId: consumerIdDep,
				})
				Expect(err).To(BeNil())
				Expect(respPro.GetResponse().Code).To(Equal(pb.Response_SUCCESS))

				var flag bool
				if len(respPro.Providers) == 2 {
					if (respPro.Providers[0].ServiceId == providerIdDep1 || respPro.Providers[0].ServiceId == providerIdDep2) &&
						(respPro.Providers[1].ServiceId == providerIdDep1 || respPro.Providers[1].ServiceId == providerIdDep2) {
						flag = true
					}
				}
				Expect(flag).To(Equal(true))

			})
			It("find ,find 不同version，相同servicename和appId", func() {
				respCreate, err := serviceResource.Create(getContext(), &pb.CreateServiceRequest{
					Service: &pb.MicroService{
						ServiceName: "serviceName_provider2",
						AppId:       "appId_provider2",
						Version:     "1.0",
						Level:       "FRONT",
						Schemas: []string{
							"com.huawei.test",
						},
						Status: "UP",
					},
				})
				Expect(err).To(BeNil())
				Expect(respCreate.GetResponse().Code).To(Equal(pb.Response_SUCCESS))
				providerIdDep3 = respCreate.ServiceId

				respFind, err := instanceResource.Find(getContext(), &pb.FindInstancesRequest{
					ConsumerServiceId: consumerIdDep,
					AppId:             "appId_provider2",
					ServiceName:       "serviceName_provider2",
					VersionRule:       "1.0",
				})
				Expect(err).To(BeNil())
				Expect(respFind.GetResponse().Code).To(Equal(pb.Response_SUCCESS))

				respPro, err := serviceResource.GetConsumerDependencies(getContext(), &pb.GetDependenciesRequest{
					ServiceId: consumerIdDep,
				})
				Expect(err).To(BeNil())
				Expect(respPro.GetResponse().Code).To(Equal(pb.Response_SUCCESS))

				var flag bool
				if len(respPro.Providers) == 2 {
					if (respPro.Providers[0].ServiceId == providerIdDep1 || respPro.Providers[0].ServiceId == providerIdDep3) &&
						(respPro.Providers[1].ServiceId == providerIdDep1 || respPro.Providers[1].ServiceId == providerIdDep3) {
						flag = true
					}
				}
				Expect(flag).To(Equal(true))
			})

			It("clean", func() {
				respDelete, err := serviceResource.Delete(getContext(), &pb.DeleteServiceRequest{
					ServiceId: consumerIdDep,
					Force:     true,
				})
				Expect(err).To(BeNil())
				Expect(respDelete.GetResponse().Code).To(Equal(pb.Response_SUCCESS))

				respDelete, err = serviceResource.Delete(getContext(), &pb.DeleteServiceRequest{
					ServiceId: providerIdDep1,
					Force:     true,
				})
				Expect(err).To(BeNil())
				Expect(respDelete.GetResponse().Code).To(Equal(pb.Response_SUCCESS))

				respDelete, err = serviceResource.Delete(getContext(), &pb.DeleteServiceRequest{
					ServiceId: providerIdDep2,
					Force:     true,
				})
				Expect(err).To(BeNil())
				Expect(respDelete.GetResponse().Code).To(Equal(pb.Response_SUCCESS))

				respDelete, err = serviceResource.Delete(getContext(), &pb.DeleteServiceRequest{
					ServiceId: providerIdDep3,
					Force:     true,
				})
				Expect(err).To(BeNil())
				Expect(respDelete.GetResponse().Code).To(Equal(pb.Response_SUCCESS))
			})

			It("find instance, create dependency", func() {
				ctx := getCunstomContext("find_dep", "find_dep")
				resp, err := serviceResource.Create(ctx, &pb.CreateServiceRequest{
					Service: &pb.MicroService{
						ServiceName: "find_create_dep_consumer",
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
				consumerId := resp.ServiceId

				resp, err = serviceResource.Create(ctx, &pb.CreateServiceRequest{
					Service: &pb.MicroService{
						ServiceName: "find_create_dep_provider",
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
				providerId := resp.ServiceId

				respFind, err := instanceResource.Find(ctx, &pb.FindInstancesRequest{
					ConsumerServiceId: consumerId,
					AppId:             "default",
					ServiceName:       "find_create_dep_provider",
					VersionRule:       "latest",
					Tags:              []string{},
					Env:               "development",
				})
				Expect(err).To(BeNil())
				Expect(respFind.GetResponse().Code).To(Equal(pb.Response_SUCCESS))

				respPro, err := serviceResource.GetConsumerDependencies(ctx, &pb.GetDependenciesRequest{
					ServiceId: consumerId,
				})
				Expect(err).To(BeNil())
				Expect(providerId).To(Equal(respPro.Providers[0].ServiceId))

				respCon, err := serviceResource.GetProviderDependencies(ctx, &pb.GetDependenciesRequest{
					ServiceId: providerId,
				})
				Expect(err).To(BeNil())
				Expect(consumerId).To(Equal(respCon.Consumers[0].ServiceId))

				resDel, err := serviceResource.Delete(ctx, &pb.DeleteServiceRequest{
					ServiceId: consumerId,
					Force:     true,
				})
				Expect(err).To(BeNil())
				Expect(resDel.GetResponse().Code).To(Equal(pb.Response_SUCCESS))

				resDel, err = serviceResource.Delete(ctx, &pb.DeleteServiceRequest{
					ServiceId: providerId,
					Force:     true,
				})
				Expect(err).To(BeNil())
				Expect(resDel.GetResponse().Code).To(Equal(pb.Response_SUCCESS))

			})
		})

		It("删除微服务,作为provider，有consumer", func() {
			var consumerId, providerId string
			resp, err := serviceResource.Create(getContext(), &pb.CreateServiceRequest{
				Service: &pb.MicroService{
					ServiceName: "serviceName_consumer",
					AppId:       "appId_consumer",
					Version:     "1.0.0",
					Level:       "FRONT",
					Schemas: []string{
						"com.huawei.test",
					},
					Status: "UP",
				},
			})
			Expect(err).To(BeNil())
			Expect(resp.GetResponse().Code).To(Equal(pb.Response_SUCCESS))
			consumerId = resp.ServiceId

			resp, err = serviceResource.Create(getContext(), &pb.CreateServiceRequest{
				Service: &pb.MicroService{
					ServiceName: "serviceName_provider",
					AppId:       "appId_provider",
					Version:     "1.0.0",
					Level:       "FRONT",
					Schemas: []string{
						"com.huawei.test",
					},
					Status: "UP",
				},
			})
			Expect(err).To(BeNil())
			providerId = resp.ServiceId
			Expect(resp.GetResponse().Code).To(Equal(pb.Response_SUCCESS))

			respCreateDependency, err := serviceResource.CreateDependenciesForMircServices(getContext(), &pb.CreateDependenciesRequest{
				Dependencies: []*pb.MircroServiceDependency{
					{
						Consumer: &pb.DependencyMircroService{
							AppId:       "appId_consumer",
							ServiceName: "serviceName_consumer",
							Version:     "1.0.0",
						},
						Providers: []*pb.DependencyMircroService{
							{
								AppId:       "",
								ServiceName: "*",
								Version:     "",
							},
						},
					},
				},
			})
			Expect(err).To(BeNil())
			Expect(respCreateDependency.GetResponse().Code).To(Equal(pb.Response_SUCCESS))

			respDelete, err := serviceResource.Delete(getContext(), &pb.DeleteServiceRequest{
				ServiceId: providerId,
				Force:     false,
			})
			Expect(err).To(BeNil())
			Expect(respDelete.GetResponse().Code).ToNot(Equal(pb.Response_SUCCESS))

			respDelete, err = serviceResource.Delete(getContext(), &pb.DeleteServiceRequest{
				ServiceId: consumerId,
				Force:     true,
			})
			Expect(err).To(BeNil())
			Expect(respDelete.GetResponse().Code).To(Equal(pb.Response_SUCCESS))

			respDelete, err = serviceResource.Delete(getContext(), &pb.DeleteServiceRequest{
				ServiceId: providerId,
				Force:     true,
			})
			Expect(err).To(BeNil())
			Expect(respDelete.GetResponse().Code).To(Equal(pb.Response_SUCCESS))
		})
	})
})
