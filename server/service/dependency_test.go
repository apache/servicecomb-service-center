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
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	pb "github.com/ServiceComb/service-center/server/core/proto"
)

var consumerId string
var providerId string
var consumerId2 string
var _ = Describe("ServiceController", func() {
	Describe("serviceDependency", func() {
		Context("normal", func() {
			It("创建Dependency,参数校验", func() {
				fmt.Println("UT===========创建Dependency，参数校验")
				resp, err := serviceResource.Create(getContext(), &pb.CreateServiceRequest{
					Service: &pb.MicroService{
						ServiceName: "service_name_consumer",
						AppId:       "service_group_consumer",
						Version:     "6.0.0",
						Level:       "FRONT",
						Schemas: []string{
							"com.huawei.test",
						},
						Status: "UP",
					},
				})
				Expect(err).To(BeNil())
				consumerId = resp.ServiceId
				Expect(resp.GetResponse().Code).To(Equal(pb.Response_SUCCESS))

				respCreateDependency, err := serviceResource.CreateDependenciesForMircServices(getContext(), &pb.CreateDependenciesRequest{
					Dependencies: nil,
				})
				Expect(err).To(BeNil())
				Expect(respCreateDependency.GetResponse().Code).To(Equal(pb.Response_FAIL))

				respCreateDependency, err = serviceResource.CreateDependenciesForMircServices(getContext(), &pb.CreateDependenciesRequest{
					Dependencies: []*pb.MircroServiceDependency{
						{
							Consumer: &pb.DependencyMircroService{
								ServiceName: "service_name_consumer2",
								AppId:       "",
								Version:     "3.0.0",
							},
							Providers: nil,
						},
					},
				})
				Expect(respCreateDependency.GetResponse().Code).To(Equal(pb.Response_FAIL))

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
				Expect(respCreateDependency.GetResponse().Code).To(Equal(pb.Response_FAIL))

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
				Expect(respCreateDependency.GetResponse().Code).To(Equal(pb.Response_FAIL))

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
				Expect(respCreateDependency.GetResponse().Code).To(Equal(pb.Response_FAIL))

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
				Expect(respCreateDependency.GetResponse().Code).To(Equal(pb.Response_FAIL))

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
				Expect(respCreateDependency.GetResponse().Code).To(Equal(pb.Response_FAIL))

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
				Expect(respCreateDependency.GetResponse().Code).To(Equal(pb.Response_FAIL))
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
									ServiceName: "service_name_consumer",
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
				Expect(resp.GetResponse().Code).To(Equal(pb.Response_FAIL))
			})

			It("查询provider对应的所有consumer,参数校验", func() {
				fmt.Println("UT===========查询privider的所有consumer,参数校验")
				resp, err := serviceResource.GetProviderDependencies(getContext(), &pb.GetDependenciesRequest{
					ServiceId: "",
				})
				Expect(err).To(BeNil())
				Expect(resp.GetResponse().Code).To(Equal(pb.Response_FAIL))

				resp, err = serviceResource.GetProviderDependencies(getContext(), &pb.GetDependenciesRequest{
					ServiceId: "noneservice",
				})
				Expect(err).To(BeNil())
				Expect(resp.GetResponse().Code).To(Equal(pb.Response_FAIL))
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
				Expect(resp.GetResponse().Code).To(Equal(pb.Response_FAIL))

				resp, err = serviceResource.GetConsumerDependencies(getContext(), &pb.GetDependenciesRequest{
					ServiceId: "noneservice",
				})
				Expect(err).To(BeNil())
				Expect(resp.GetResponse().Code).To(Equal(pb.Response_FAIL))
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
				resp, err := insResource.Find(getContext(), &pb.FindInstancesRequest{
					ConsumerServiceId: consumerId,
					AppId:             "service_group_provider",
					ServiceName:       "service_name_provider",
					VersionRule:       "latest",
				})
				Expect(err).To(BeNil())
				Expect(resp.GetResponse().Code).To(Equal(pb.Response_SUCCESS))

				//重复find
				resp, err = insResource.Find(getContext(), &pb.FindInstancesRequest{
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
	})
})
