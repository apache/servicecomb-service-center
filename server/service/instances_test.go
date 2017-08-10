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
	"github.com/ServiceComb/service-center/server/service"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

var instanceId string
var _ = Describe("InstanceController", func() {
	var consumerId string
	var providerId string
	Describe("Register", func() {
		Context("normal", func() {
			It("创建实例", func() {
				fmt.Println("UT===========创建实例")
				respCreate, err := serviceResource.Create(getContext(), &pb.CreateServiceRequest{
					Service: &pb.MicroService{
						ServiceName: "service_name_consumer_register",
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
				consumerId = respCreate.ServiceId
				Expect(respCreate.GetResponse().Code).To(Equal(pb.Response_SUCCESS))

				resp, err := insResource.Register(getContext(), &pb.RegisterInstanceRequest{
					Instance: &pb.MicroServiceInstance{
						ServiceId: consumerId,
						Endpoints: []string{
							"rest:127.0.0.1:8080",
						},
						HostName: "UT-HOST",
						Status:   pb.MSI_UP,
						Stage:    "prod",
					},
				})
				resp, err = insResource.Register(getContext(), &pb.RegisterInstanceRequest{
					Instance: &pb.MicroServiceInstance{
						ServiceId: consumerId,
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
				Expect(resp.GetResponse().Code).To(Equal(pb.Response_SUCCESS))

				resp, err = insResource.Register(getContext(), &pb.RegisterInstanceRequest{
					Instance: &pb.MicroServiceInstance{
						ServiceId: consumerId,
						Endpoints: []string{
							"rest:127.0.0.1:8080",
						},
						HostName:   "UT-HOST",
						Status:     pb.MSI_UP,
						Stage:      "prod",
						Properties: map[string]string{"nodeIP": "test"},
					},
				})
				Expect(err).To(BeNil())
				fmt.Println("UT============" + resp.GetResponse().Message)
				Expect(resp.GetResponse().Code).To(Equal(pb.Response_SUCCESS))

				instanceId = resp.InstanceId

				respUpdateProperties, err := insResource.UpdateInstanceProperties(getContext(), &pb.UpdateInstancePropsRequest{
					ServiceId:  consumerId,
					InstanceId: instanceId,
					Properties: map[string]string{
						"test": "test",
					},
				})

				Expect(err).To(BeNil())
				fmt.Println("UT============" + respUpdateProperties.GetResponse().Message)
				Expect(respUpdateProperties.GetResponse().Code).To(Equal(pb.Response_SUCCESS))

				respUpdateStatus, err := insResource.UpdateStatus(getContext(), &pb.UpdateInstanceStatusRequest{
					ServiceId:  consumerId,
					InstanceId: instanceId,
					Status:     pb.MSI_STARTING,
				})

				Expect(err).To(BeNil())
				fmt.Println("UT============" + respUpdateStatus.GetResponse().Message)
				Expect(respUpdateStatus.GetResponse().Code).To(Equal(pb.Response_SUCCESS))
			})

			It("批量心跳接口", func() {
				fmt.Println("UT===========实例心跳上报批量接口， 参数校验")
				resp, err := insResource.HeartbeatSet(getContext(), &pb.HeartbeatSetRequest{
					Instances: []*pb.HeartbeatSetElement{},
				})
				Expect(err).To(BeNil())
				Expect(resp.GetResponse().Code).To(Equal(pb.Response_FAIL))

				fmt.Println("UT===========实例心跳上报批量接口")
				resp, err = insResource.HeartbeatSet(getContext(), &pb.HeartbeatSetRequest{
					Instances: []*pb.HeartbeatSetElement{
						&pb.HeartbeatSetElement{
							ServiceId:  consumerId,
							InstanceId: instanceId,
						},
					},
				})
				Expect(resp.GetResponse().Code).To(Equal(pb.Response_SUCCESS))

				resp, err = insResource.HeartbeatSet(getContext(), &pb.HeartbeatSetRequest{
					Instances: []*pb.HeartbeatSetElement{
						&pb.HeartbeatSetElement{
							ServiceId:  consumerId,
							InstanceId: instanceId,
						},
						&pb.HeartbeatSetElement{
							ServiceId:  consumerId,
							InstanceId: instanceId,
						},
					},
				})
				Expect(resp.GetResponse().Code).To(Equal(pb.Response_SUCCESS))

				resp, err = insResource.HeartbeatSet(getContext(), &pb.HeartbeatSetRequest{
					Instances: []*pb.HeartbeatSetElement{
						&pb.HeartbeatSetElement{
							ServiceId:  consumerId,
							InstanceId: "not-exist-instanceId",
						},
						&pb.HeartbeatSetElement{
							ServiceId:  consumerId,
							InstanceId: instanceId,
						},
					},
				})
				Expect(resp.GetResponse().Code).To(Equal(pb.Response_FAIL))
			})

			It("创建实例,参数校验", func() {
				fmt.Println("UT===========创建实例， 参数校验")
				resp, err := insResource.Register(getContext(), &pb.RegisterInstanceRequest{
					Instance: &pb.MicroServiceInstance{
						ServiceId: "1000000000000000",
						Endpoints: []string{
							"rest:127.0.0.1:8080",
						},
						HostName: "UT-HOST",
						Status:   pb.MSI_UP,
						Stage:    "prod",
					},
				})
				Expect(err).To(BeNil())
				Expect(resp.GetResponse().Code).To(Equal(pb.Response_FAIL))
			})

			It("health", func() {
				//注册sc
				fmt.Println("UT===========health")

				respCluterhealth, err := insResource.CluterHealth(getContext())
				Expect(err).To(BeNil())
				Expect(respCluterhealth.GetResponse().Code).To(Equal(pb.Response_FAIL))

				resp, err := serviceResource.Create(getContext(), core.CreateServiceRequest())
				Expect(err).To(BeNil())
				scServiceId := resp.ServiceId
				Expect(resp.GetResponse().Code).To(Equal(pb.Response_SUCCESS))

				//注册实例
				respIns, err := insResource.Register(getContext(), &pb.RegisterInstanceRequest{
					Instance: &pb.MicroServiceInstance{
						ServiceId: scServiceId,
						Endpoints: []string{
							"rest:127.0.0.1:8080",
						},
						HostName: "UT-HOST",
						Status:   pb.MSI_UP,
						Stage:    "prod",
					},
				})
				Expect(respIns.GetResponse().Code).To(Equal(pb.Response_SUCCESS))

				respCluterhealth, err = insResource.CluterHealth(getContext())
				Expect(err).To(BeNil())
				Expect(respCluterhealth.GetResponse().Code).To(Equal(pb.Response_SUCCESS))
			})

			It("修改实例的status，参数校验", func() {
				respUpdateStatus, err := insResource.UpdateStatus(getContext(), &pb.UpdateInstanceStatusRequest{
					ServiceId:  "",
					InstanceId: instanceId,
					Status:     pb.MSI_STARTING,
				})

				Expect(err).To(BeNil())
				fmt.Println("UT============" + respUpdateStatus.GetResponse().Message)
				Expect(respUpdateStatus.GetResponse().Code).To(Equal(pb.Response_FAIL))

				respUpdateStatus, err = insResource.UpdateStatus(getContext(), &pb.UpdateInstanceStatusRequest{
					ServiceId:  "notexistservice",
					InstanceId: instanceId,
					Status:     pb.MSI_STARTING,
				})

				Expect(err).To(BeNil())
				fmt.Println("UT============" + respUpdateStatus.GetResponse().Message)
				Expect(respUpdateStatus.GetResponse().Code).To(Equal(pb.Response_FAIL))

				respUpdateStatus, err = insResource.UpdateStatus(getContext(), &pb.UpdateInstanceStatusRequest{
					ServiceId:  consumerId,
					InstanceId: "notexistins",
					Status:     pb.MSI_STARTING,
				})

				Expect(err).To(BeNil())
				fmt.Println("UT============" + respUpdateStatus.GetResponse().Message)
				Expect(respUpdateStatus.GetResponse().Code).To(Equal(pb.Response_FAIL))

				respUpdateStatus, err = insResource.UpdateStatus(getContext(), &pb.UpdateInstanceStatusRequest{
					ServiceId:  consumerId,
					InstanceId: instanceId,
					Status:     "nonestatus",
				})

				Expect(err).To(BeNil())
				fmt.Println("UT============" + respUpdateStatus.GetResponse().Message)
				Expect(respUpdateStatus.GetResponse().Code).To(Equal(pb.Response_FAIL))
			})

			It("修改实例的properties", func() {
				respUpdateProperties, err := insResource.UpdateInstanceProperties(getContext(), &pb.UpdateInstancePropsRequest{
					ServiceId:  consumerId,
					InstanceId: "notexistins",
					Properties: map[string]string{
						"test": "test",
					},
				})

				Expect(err).To(BeNil())
				fmt.Println("UT============" + respUpdateProperties.GetResponse().Message)
				Expect(respUpdateProperties.GetResponse().Code).To(Equal(pb.Response_FAIL))

				respUpdateProperties, err = insResource.UpdateInstanceProperties(getContext(), &pb.UpdateInstancePropsRequest{
					ServiceId:  "",
					InstanceId: instanceId,
					Properties: map[string]string{
						"test": "test",
					},
				})

				Expect(err).To(BeNil())
				fmt.Println("UT============" + respUpdateProperties.GetResponse().Message)
				Expect(respUpdateProperties.GetResponse().Code).To(Equal(pb.Response_FAIL))

				respUpdateProperties, err = insResource.UpdateInstanceProperties(getContext(), &pb.UpdateInstancePropsRequest{
					ServiceId:  "notexistservice",
					InstanceId: instanceId,
					Properties: map[string]string{
						"test": "test",
					},
				})

				Expect(err).To(BeNil())
				fmt.Println("UT============" + respUpdateProperties.GetResponse().Message)
				Expect(respUpdateProperties.GetResponse().Code).To(Equal(pb.Response_FAIL))
			})

			It("创建实例，字段校验", func() {
				fmt.Println("UT===========创建实例")
				respCreate, err := serviceResource.Create(getContext(), &pb.CreateServiceRequest{
					Service: &pb.MicroService{
						ServiceName: "instance_validate",
						AppId:       "instance_validate",
						Version:     "1.0.0",
						Level:       "FRONT",
						Status:      "UP",
					},
				})
				Expect(err).To(BeNil())
				// Expect(respCreate.GetResponse().Code).To(Equal(pb.Response_SUCCESS))

				resp, err := insResource.Register(getContext(), &pb.RegisterInstanceRequest{
					Instance: &pb.MicroServiceInstance{
						ServiceId: respCreate.ServiceId,
						HostName:  "UT-HOST",
						Status:    pb.MSI_UP,
						Stage:     "prod",
					},
				})
				Expect(err).To(BeNil())
				fmt.Println("UT============" + resp.GetResponse().Message)
				Expect(resp.GetResponse().Code).To(Equal(pb.Response_SUCCESS))

				resp, err = insResource.Register(getContext(), &pb.RegisterInstanceRequest{
					Instance: &pb.MicroServiceInstance{
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
				Expect(resp.GetResponse().Code).To(Equal(pb.Response_FAIL))

				resp, err = insResource.Register(getContext(), &pb.RegisterInstanceRequest{
					Instance: &pb.MicroServiceInstance{
						ServiceId: consumerId,
						Endpoints: []string{
							"rest:127.0.0.1:8080",
						},
						Status: pb.MSI_UP,
						Stage:  "prod",
					},
				})
				Expect(err).To(BeNil())
				fmt.Println("UT============" + resp.GetResponse().Message)
				Expect(resp.GetResponse().Code).To(Equal(pb.Response_FAIL))

				resp, err = insResource.Register(getContext(), &pb.RegisterInstanceRequest{
					Instance: &pb.MicroServiceInstance{
						ServiceId: consumerId,
						Endpoints: []string{
							"rest:127.0.0.1:8080",
						},
						HostName: "UT-HOST",
						Stage:    "prod",
					},
				})
				Expect(err).To(BeNil())
				fmt.Println("UT============" + resp.GetResponse().Message)
				Expect(resp.GetResponse().Code).To(Equal(pb.Response_FAIL))

				resp, err = insResource.Register(getContext(), &pb.RegisterInstanceRequest{
					Instance: nil,
				})
				Expect(err).To(BeNil())
				fmt.Println("UT============" + resp.GetResponse().Message)
				Expect(resp.GetResponse().Code).To(Equal(pb.Response_FAIL))
			})

			It("创建实例，字段校验,服务不存在", func() {
				resp, err := insResource.Register(getContext(), &pb.RegisterInstanceRequest{
					Instance: &pb.MicroServiceInstance{
						ServiceId: consumerId,
						Endpoints: []string{
							"rest:127.0.0.1:8080",
						},
						HostName: "UT-HOST",
						Stage:    "prod",
					},
				})
				Expect(err).To(BeNil())
				fmt.Println("UT============" + resp.GetResponse().Message)
				Expect(resp.GetResponse().Code).To(Equal(pb.Response_FAIL))
			})
			It("创建实例,含healthcheck", func() {
				fmt.Println("UT============创建实例，含checkhealth")
				resp, err := insResource.Register(getContext(), &pb.RegisterInstanceRequest{
					Instance: &pb.MicroServiceInstance{
						ServiceId: consumerId,
						HostName:  "UT-HOST",
						Status:    pb.MSI_UP,
						Stage:     "prod",
						HealthCheck: &pb.HealthCheck{
							Mode:     "push",
							Interval: 30,
							Times:    1,
						},
					},
				})
				Expect(err).To(BeNil())
				fmt.Println("UT============" + resp.GetResponse().Message)
				Expect(resp.GetResponse().Code).To(Equal(pb.Response_SUCCESS))

				resp, err = insResource.Register(getContext(), &pb.RegisterInstanceRequest{
					Instance: &pb.MicroServiceInstance{
						ServiceId: consumerId,
						HostName:  "UT-HOST",
						Status:    pb.MSI_UP,
						Stage:     "prod",
						HealthCheck: &pb.HealthCheck{
							Mode:     "pull",
							Interval: 30,
							Times:    1,
							Url:      "/abc/d",
						},
					},
				})
				Expect(err).To(BeNil())
				fmt.Println("UT============" + resp.GetResponse().Message)
				Expect(resp.GetResponse().Code).To(Equal(pb.Response_SUCCESS))

				resp, err = insResource.Register(getContext(), &pb.RegisterInstanceRequest{
					Instance: &pb.MicroServiceInstance{
						ServiceId: consumerId,
						HostName:  "UT-HOST",
						Status:    pb.MSI_UP,
						Stage:     "prod",
						HealthCheck: &pb.HealthCheck{
							Mode:     "push",
							Interval: 30,
							Times:    0,
						},
					},
				})
				Expect(err).To(BeNil())
				fmt.Println("UT============" + resp.GetResponse().Message)
				Expect(resp.GetResponse().Code).To(Equal(pb.Response_FAIL))

				resp, err = insResource.Register(getContext(), &pb.RegisterInstanceRequest{
					Instance: &pb.MicroServiceInstance{
						ServiceId: consumerId,
						HostName:  "UT-HOST",
						Status:    pb.MSI_UP,
						Stage:     "prod",
						HealthCheck: &pb.HealthCheck{
							Mode:     "pull",
							Interval: 30,
							Times:    1,
							Url:      "*",
						},
					},
				})
				Expect(err).To(BeNil())
				fmt.Println("UT============" + resp.GetResponse().Message)
				Expect(resp.GetResponse().Code).To(Equal(pb.Response_FAIL))
			})
		})
	})
	Describe("findInstance", func() {
		Context("normal", func() {
			var consumerIdFind string
			It("发现实例,参数校验", func() {
				fmt.Println("UT===========发现实例")
				respCreate, err := serviceResource.Create(getContext(), &pb.CreateServiceRequest{
					Service: &pb.MicroService{
						ServiceName: "service_name_consumer_find",
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
				consumerIdFind = respCreate.ServiceId
				Expect(respCreate.GetResponse().Code).To(Equal(pb.Response_SUCCESS))

				respFind, err := insResource.Find(getContext(), &pb.FindInstancesRequest{
					ConsumerServiceId: consumerIdFind,
					AppId:             "service_group_provider",
					ServiceName:       "service_name_provider",
					VersionRule:       "1.0.0+",
					Tags:              []string{},
					Stage:             "nonestage",
				})
				Expect(err).To(BeNil())
				fmt.Println("UT============" + respFind.GetResponse().Message)
				Expect(respFind.GetResponse().Code).To(Equal(pb.Response_FAIL))

				respFind, err = insResource.Find(getContext(), &pb.FindInstancesRequest{
					ConsumerServiceId: "",
					AppId:             "service_group_provider",
					ServiceName:       "service_name_provider",
					VersionRule:       "1.0.0+",
					Tags:              []string{},
					Stage:             "dev",
				})
				Expect(err).To(BeNil())
				fmt.Println("UT============" + respFind.GetResponse().Message)
				Expect(respFind.GetResponse().Code).To(Equal(pb.Response_FAIL))

				respFind, err = insResource.Find(getContext(), &pb.FindInstancesRequest{
					ConsumerServiceId: consumerIdFind,
					AppId:             "service_group_provider",
					ServiceName:       "noneservice",
					VersionRule:       "latest",
					Tags:              []string{},
					Stage:             "dev",
				})
				Expect(err).To(BeNil())
				fmt.Println("UT============" + respFind.GetResponse().Message)
				Expect(respFind.GetResponse().Code).To(Equal(pb.Response_FAIL))

				respFind, err = insResource.Find(getContext(), &pb.FindInstancesRequest{
					ConsumerServiceId: consumerIdFind,
					AppId:             "service_group_provider",
					ServiceName:       "service_name_provider",
					VersionRule:       "3.0.0+",
					Tags:              []string{},
					Stage:             "dev",
				})
				Expect(err).To(BeNil())
				fmt.Println("UT============" + respFind.GetResponse().Message)
				Expect(respFind.GetResponse().Code).To(Equal(pb.Response_FAIL))

				respFind, err = insResource.Find(getContext(), &pb.FindInstancesRequest{
					ConsumerServiceId: consumerIdFind,
					AppId:             "service_group_provider",
					ServiceName:       "service_name_provider",
					VersionRule:       "2.0.0-2.0.1",
					Tags:              []string{},
					Stage:             "dev",
				})
				Expect(err).To(BeNil())
				fmt.Println("UT============" + respFind.GetResponse().Message)
				Expect(respFind.GetResponse().Code).To(Equal(pb.Response_FAIL))

				respFind, err = insResource.Find(getContext(), &pb.FindInstancesRequest{
					ConsumerServiceId: consumerIdFind,
					AppId:             "service_group_provider",
					ServiceName:       "service_name_provider",
					VersionRule:       "2.0.0",
					Tags:              []string{},
					Stage:             "dev",
				})
				Expect(err).To(BeNil())

				fmt.Println("UT============" + respFind.GetResponse().Message)
				Expect(respFind.GetResponse().Code).To(Equal(pb.Response_FAIL))

				respFind, err = insResource.Find(getContext(), &pb.FindInstancesRequest{
					ConsumerServiceId: "notExistServiceId",
					AppId:             "service_group_provider",
					ServiceName:       "service_name_provider",
					VersionRule:       "2.0.0",
					Tags:              []string{},
					Stage:             "dev",
				})
				Expect(err).To(BeNil())

				fmt.Println("UT============" + respFind.GetResponse().Message)
				Expect(respFind.GetResponse().Code).To(Equal(pb.Response_FAIL))
			})

			It("发现实例", func() {
				fmt.Println("UT===========发现实例")
				resp, err := serviceResource.Create(getContext(), &pb.CreateServiceRequest{
					Service: &pb.MicroService{
						ServiceName: "service_name_provider",
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
				providerId = resp.ServiceId

				respFind, err := insResource.Find(getContext(), &pb.FindInstancesRequest{
					ConsumerServiceId: consumerId,
					AppId:             "default",
					ServiceName:       "service_name_provider",
					VersionRule:       "latest",
					Tags:              []string{},
					Stage:             "dev",
				})
				fmt.Println("test resp.GetResponse() ", respFind.GetResponse())
				Expect(err).To(BeNil())
				fmt.Println("UT============" + respFind.GetResponse().Message)
				Expect(respFind.GetResponse().Code).To(Equal(pb.Response_SUCCESS))

				respFind, err = insResource.Find(getContext(), &pb.FindInstancesRequest{
					ConsumerServiceId: consumerId,
					AppId:             "default",
					ServiceName:       "service_name_provider",
					VersionRule:       "1.0.0+",
					Tags:              []string{},
					Stage:             "dev",
				})
				Expect(err).To(BeNil())
				fmt.Println("UT============" + respFind.GetResponse().Message)
				Expect(respFind.GetResponse().Code).To(Equal(pb.Response_SUCCESS))

				respFind, err = insResource.Find(getContext(), &pb.FindInstancesRequest{
					ConsumerServiceId: consumerId,
					AppId:             "default",
					ServiceName:       "service_name_provider",
					VersionRule:       "1.0.0-1.0.1",
					Tags:              []string{},
					Stage:             "dev",
				})
				Expect(err).To(BeNil())
				fmt.Println("UT============" + resp.GetResponse().Message)
				Expect(respFind.GetResponse().Code).To(Equal(pb.Response_SUCCESS))

				respFind, err = insResource.Find(getContext(), &pb.FindInstancesRequest{
					ConsumerServiceId: consumerId,
					AppId:             "default",
					ServiceName:       "service_name_provider",
					VersionRule:       "1.0.0",
					Tags:              []string{},
					Stage:             "dev",
				})
				Expect(err).To(BeNil())

				fmt.Println("UT============" + resp.GetResponse().Message)
				Expect(respFind.GetResponse().Code).To(Equal(pb.Response_SUCCESS))
			})

			It("跨App发现实例", func() {
				fmt.Println("UT===========跨App发现实例")
				resp, err := serviceResource.Create(getContext(), &pb.CreateServiceRequest{
					Service: &pb.MicroService{
						ServiceName: "service_name_app_consumer",
						AppId:       "service_group_app_consumer",
						Version:     "1.0.0",
						Level:       "FRONT",
						Schemas: []string{
							"xxxxxxxx",
						},
						Status: "UP",
					},
				})
				Expect(err).To(BeNil())
				consumerId := resp.ServiceId

				resp, err = serviceResource.Create(getContext(), &pb.CreateServiceRequest{
					Service: &pb.MicroService{
						ServiceName: "service_name_app_provider_fail",
						AppId:       "service_group_app_provider",
						Version:     "1.0.0",
						Level:       "FRONT",
						Schemas: []string{
							"xxxxxxxx",
						},
						Status: "UP",
					},
				})
				Expect(err).To(BeNil())
				providerFailId := resp.ServiceId

				resp, err = serviceResource.Create(getContext(), &pb.CreateServiceRequest{
					Service: &pb.MicroService{
						ServiceName: "service_name_app_provider_ok",
						AppId:       "service_group_app_provider",
						Version:     "1.0.0",
						Level:       "FRONT",
						Schemas: []string{
							"xxxxxxxx",
						},
						Status: "UP",
						Properties: map[string]string{
							pb.PROP_ALLOW_CROSS_APP: "true",
						},
					},
				})
				Expect(err).To(BeNil())
				providerOkId := resp.ServiceId

				UTFunc := func(providerId string, code pb.Response_Code) {
					respFind, err := insResource.GetInstances(getContext(), &pb.GetInstancesRequest{
						ConsumerServiceId: consumerId,
						ProviderServiceId: providerId,
					})
					fmt.Println("test resp.GetResponse() ", respFind.GetResponse())
					Expect(err).To(BeNil())
					fmt.Println("UT============" + respFind.GetResponse().Message)
					Expect(respFind.GetResponse().Code).To(Equal(code))
				}

				UTFunc(providerFailId, pb.Response_FAIL)

				UTFunc(providerOkId, pb.Response_SUCCESS)
			})

			It("实例心跳", func() {
				fmt.Println("UT===========实例心跳")
				resp, err := insResource.Heartbeat(getContext(), &pb.HeartbeatRequest{
					ServiceId:  consumerId,
					InstanceId: instanceId,
				})
				Expect(err).To(BeNil())
				Expect(resp.GetResponse().Code).To(Equal(pb.Response_SUCCESS))
			})

			It("实例心跳,参数校验", func() {
				fmt.Println("UT===========实例心跳")
				resp, err := insResource.Heartbeat(getContext(), &pb.HeartbeatRequest{
					ServiceId:  "",
					InstanceId: instanceId,
				})
				Expect(err).To(BeNil())
				Expect(resp.GetResponse().Code).To(Equal(pb.Response_FAIL))

				resp, err = insResource.Heartbeat(getContext(), &pb.HeartbeatRequest{
					ServiceId:  consumerId,
					InstanceId: "100000000000",
				})
				Expect(resp.GetResponse().Code).To(Equal(pb.Response_FAIL))

				resp, err = insResource.Heartbeat(getContext(), &pb.HeartbeatRequest{
					ServiceId:  consumerId,
					InstanceId: "not-exist-ins",
				})
				Expect(resp.GetResponse().Code).To(Equal(pb.Response_FAIL))
			})

			It("删除微服务,实例存在，不能删除", func() {
				resp, err := serviceResource.Delete(getContext(), &pb.DeleteServiceRequest{
					ServiceId: consumerId,
					Force:     false,
				})
				Expect(err).To(BeNil())
				Expect(resp.GetResponse().Code).To(Equal(pb.Response_FAIL))
			})
		})
	})

	Describe("getInstances", func() {
		Context("normal", func() {
			It("查找单个实例,参数校验", func() {
				fmt.Println("UT===========查找实例，参数校验")
				resp, err := insResource.GetOneInstance(getContext(), &pb.GetOneInstanceRequest{
					ConsumerServiceId:  consumerId,
					ProviderServiceId:  "",
					ProviderInstanceId: instanceId,
				})
				Expect(err).To(BeNil())
				fmt.Println("UT============" + resp.GetResponse().Message)
				Expect(resp.GetResponse().Code).To(Equal(pb.Response_FAIL))

				fmt.Println("UT===========查找实例，参数校验,")
				resp, err = insResource.GetOneInstance(getContext(), &pb.GetOneInstanceRequest{
					ConsumerServiceId:  consumerId,
					ProviderServiceId:  consumerId,
					ProviderInstanceId: instanceId,
					Stage:              "dev",
				})
				Expect(err).To(BeNil())
				fmt.Println("UT============" + resp.GetResponse().Message)
				Expect(resp.GetResponse().Code).To(Equal(pb.Response_FAIL))
			})

			It("查找单个实例,夸app调用", func() {
				respCreate, err := serviceResource.Create(getContext(), &pb.CreateServiceRequest{
					Service: &pb.MicroService{
						ServiceName: "service_name_cross_app",
						AppId:       "crossApp",
						Version:     "1.0.0",
						Level:       "FRONT",
						Schemas: []string{
							"xxxxxxxx",
						},
						Status: "UP",
					},
				})
				Expect(err).To(BeNil())
				crossAppId := respCreate.ServiceId

				fmt.Println("UT===========查找实例，参数校验")
				resp, err := insResource.GetOneInstance(getContext(), &pb.GetOneInstanceRequest{
					ConsumerServiceId:  crossAppId,
					ProviderServiceId:  consumerId,
					ProviderInstanceId: instanceId,
				})
				Expect(err).To(BeNil())
				fmt.Println("UT============" + resp.GetResponse().Message)
				Expect(resp.GetResponse().Code).To(Equal(pb.Response_FAIL))
			})

			It("查找单个实例", func() {
				fmt.Println("UT===========查找实例，参数校验")
				resp, err := insResource.GetOneInstance(getContext(), &pb.GetOneInstanceRequest{
					ConsumerServiceId:  consumerId,
					ProviderServiceId:  consumerId,
					ProviderInstanceId: instanceId,
				})
				Expect(err).To(BeNil())
				fmt.Println("UT============" + resp.GetResponse().Message)
				Expect(resp.GetResponse().Code).To(Equal(pb.Response_SUCCESS))

				fmt.Println("UT===========查找实例，含有不存在的tag")
				resp, err = insResource.GetOneInstance(getContext(), &pb.GetOneInstanceRequest{
					ConsumerServiceId:  consumerId,
					ProviderServiceId:  consumerId,
					ProviderInstanceId: instanceId,
					Tags:               []string{"not-exist-tag"},
				})
				Expect(err).To(BeNil())
				fmt.Println("UT============" + resp.GetResponse().Message)
				Expect(resp.GetResponse().Code).To(Equal(pb.Response_FAIL))

				fmt.Println("UT===========查找实例，含有存在的tag")
				respAddTags, err := serviceResource.AddTags(getContext(), &pb.AddServiceTagsRequest{
					ServiceId: consumerId,
					Tags: map[string]string{
						"test": "test",
					},
				})
				Expect(err).To(BeNil())
				Expect(respAddTags.GetResponse().Code).To(Equal(pb.Response_SUCCESS))

				resp, err = insResource.GetOneInstance(getContext(), &pb.GetOneInstanceRequest{
					ConsumerServiceId:  consumerId,
					ProviderServiceId:  consumerId,
					ProviderInstanceId: instanceId,
					Tags:               []string{"test"},
				})
				Expect(err).To(BeNil())
				fmt.Println("UT============" + resp.GetResponse().Message)
				Expect(resp.GetResponse().Code).To(Equal(pb.Response_SUCCESS))

			})

			It("查找实例，参数校验", func() {
				fmt.Println("UT===========查找实例，参数校验")
				resp, err := insResource.GetInstances(getContext(), &pb.GetInstancesRequest{
					ConsumerServiceId: "",
					ProviderServiceId: instanceId,
				})
				Expect(err).To(BeNil())
				fmt.Println("UT============" + resp.GetResponse().Message)
				Expect(resp.GetResponse().Code).To(Equal(pb.Response_FAIL))

				resp, err = insResource.GetInstances(getContext(), &pb.GetInstancesRequest{
					ConsumerServiceId: "noneservice",
					ProviderServiceId: "noneservice",
					Tags:              []string{},
					Stage:             "prod",
				})
				Expect(err).To(BeNil())
				fmt.Println("UT============" + resp.GetResponse().Message)
				Expect(resp.GetResponse().Code).To(Equal(pb.Response_FAIL))

				resp, err = insResource.GetInstances(getContext(), &pb.GetInstancesRequest{
					ConsumerServiceId: consumerId,
					ProviderServiceId: providerId,
					Tags:              []string{},
				})
				Expect(err).To(BeNil())
				fmt.Println("UT============" + resp.GetResponse().Message)
				Expect(resp.GetResponse().Code).To(Equal(pb.Response_SUCCESS))
			})
		})
	})

	Describe("Unregister", func() {
		Context("normal", func() {
			It("去注册实例", func() {
				fmt.Println("UT===========去注册实例")
				resp, err := insResource.Unregister(getContext(), &pb.UnregisterInstanceRequest{
					ServiceId:  consumerId,
					InstanceId: instanceId,
				})
				Expect(err).To(BeNil())
				fmt.Println("UT============" + resp.GetResponse().Message)
				//Expect(resp.GetResponse().Code).To(Equal(pb.Response_SUCCESS))

				respDeleteService, err := serviceResource.Delete(getContext(), &pb.DeleteServiceRequest{
					ServiceId: consumerId,
					Force:     true,
				})
				Expect(err).To(BeNil())
				fmt.Println("UT============" + resp.GetResponse().Message)
				Expect(respDeleteService.GetResponse().Code).To(Equal(pb.Response_SUCCESS))
			})

			It("去注册实例, 参数校验", func() {
				fmt.Println("UT===========去注册实例")
				resp, err := insResource.Unregister(getContext(), &pb.UnregisterInstanceRequest{
					ServiceId:  "",
					InstanceId: instanceId,
				})
				Expect(err).To(BeNil())
				fmt.Println("UT============" + resp.GetResponse().Message)
				Expect(resp.GetResponse().Code).To(Equal(pb.Response_FAIL))

				resp, err = insResource.Unregister(getContext(), &pb.UnregisterInstanceRequest{
					ServiceId:  consumerId,
					InstanceId: instanceId,
				})
				Expect(err).To(BeNil())
				fmt.Println("UT============" + resp.GetResponse().Message)
				Expect(resp.GetResponse().Code).To(Equal(pb.Response_FAIL))

				resp, err = insResource.Unregister(getContext(), &pb.UnregisterInstanceRequest{
					ServiceId:  consumerId,
					InstanceId: "100000000000",
				})
				Expect(err).To(BeNil())
				fmt.Println("UT============" + resp.GetResponse().Message)
				Expect(resp.GetResponse().Code).To(Equal(pb.Response_FAIL))
			})
		})
	})

	Describe("Watcher", func() {
		Context("normal", func() {
			It("参数校验，不存在", func() {
				fmt.Println("UT===========参数校验，不存在")
				IC := insResource.(*service.InstanceController)
				err := IC.WatchPreOpera(getContext(), &pb.WatchInstanceRequest{
					SelfServiceId: "-1",
				})
				fmt.Println("UT============" + err.Error())
				Expect(err).NotTo(BeNil())

				err = IC.Watch(&pb.WatchInstanceRequest{
					SelfServiceId: "-1",
				}, &grpcWatchServer{})
				fmt.Println("UT============" + err.Error())
				Expect(err).NotTo(BeNil())
			})
			It("参数校验，非法", func() {
				fmt.Println("UT===========参数校验，非法")
				err := insResource.(*service.InstanceController).WatchPreOpera(getContext(), &pb.WatchInstanceRequest{
					SelfServiceId: "",
				})
				fmt.Println("UT============" + err.Error())
				Expect(err).NotTo(BeNil())
			})
			It("参数校验，OK", func() {
				fmt.Println("UT===========参数校验，OK")
				respCreate, err := serviceResource.Create(getContext(), &pb.CreateServiceRequest{
					Service: &pb.MicroService{
						ServiceName: "service_name_watch",
						AppId:       "watch",
						Version:     "1.0.0",
						Level:       "BACK",
						Status:      "UP",
					},
				})
				Expect(err).To(BeNil())

				err = insResource.(*service.InstanceController).WatchPreOpera(getContext(),
					&pb.WatchInstanceRequest{
						SelfServiceId: respCreate.ServiceId,
					})
				Expect(err).To(BeNil())
			})
		})
	})
})

type grpcWatchServer struct {
	grpc.ServerStream
}

func (x *grpcWatchServer) Send(m *pb.WatchInstanceResponse) error {
	return nil
}

func (x *grpcWatchServer) Context() context.Context {
	return getContext()
}
