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
	"github.com/apache/servicecomb-service-center/pkg/util"
	"github.com/apache/servicecomb-service-center/server/core"
	pb "github.com/apache/servicecomb-service-center/server/core/proto"
	scerr "github.com/apache/servicecomb-service-center/server/error"
	serviceUtil "github.com/apache/servicecomb-service-center/server/service/util"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"math"
	"os"
	"strconv"
	"strings"
)

var (
	TOO_LONG_HOSTNAME = strings.Repeat("x", 65)
	TOO_LONG_URL      = strings.Repeat("x", 513)
)

var _ = Describe("'Instance' service", func() {
	Describe("execute 'register' operartion", func() {
		var (
			serviceId1 string
			serviceId2 string
		)

		It("should be passed", func() {
			respCreate, err := serviceResource.Create(getContext(), &pb.CreateServiceRequest{
				Service: &pb.MicroService{
					ServiceName: "create_instance_service",
					AppId:       "create_instance",
					Version:     "1.0.0",
					Level:       "FRONT",
					Status:      pb.MS_UP,
				},
			})
			Expect(err).To(BeNil())
			Expect(respCreate.Response.Code).To(Equal(pb.Response_SUCCESS))
			serviceId1 = respCreate.ServiceId

			respCreate, err = serviceResource.Create(getContext(), &pb.CreateServiceRequest{
				Service: &pb.MicroService{
					ServiceName: "create_instance_service",
					AppId:       "create_instance",
					Version:     "1.0.1",
					Level:       "FRONT",
					Status:      pb.MS_UP,
				},
			})
			Expect(err).To(BeNil())
			Expect(respCreate.Response.Code).To(Equal(pb.Response_SUCCESS))
			serviceId2 = respCreate.ServiceId
		})

		Context("when register a instance", func() {
			It("should be passed", func() {
				resp, err := instanceResource.Register(getContext(), &pb.RegisterInstanceRequest{
					Instance: &pb.MicroServiceInstance{
						ServiceId: serviceId1,
						Endpoints: []string{
							"createInstance:127.0.0.1:8080",
						},
						HostName: "UT-HOST",
						Status:   pb.MSI_UP,
					},
				})
				Expect(err).To(BeNil())
				Expect(resp.Response.Code).To(Equal(pb.Response_SUCCESS))
				Expect(resp.InstanceId).To(Not(Equal("")))

				By("status is nil")
				resp, err = instanceResource.Register(getContext(), &pb.RegisterInstanceRequest{
					Instance: &pb.MicroServiceInstance{
						ServiceId: serviceId1,
						Endpoints: []string{
							"createInstance:127.0.0.1:8081",
						},
						HostName: "UT-HOST",
					},
				})
				Expect(err).To(BeNil())
				Expect(resp.Response.Code).To(Equal(pb.Response_SUCCESS))
				Expect(resp.InstanceId).To(Not(Equal("")))

				By("all max")
				size := 1000
				var eps []string
				properties := make(map[string]string, size)
				for i := 0; i < size; i++ {
					s := strconv.Itoa(i) + strings.Repeat("x", 253)
					eps = append(eps, s)
					properties[s] = s
				}
				resp, err = instanceResource.Register(getContext(), &pb.RegisterInstanceRequest{
					Instance: &pb.MicroServiceInstance{
						ServiceId:  serviceId1,
						Endpoints:  eps,
						HostName:   TOO_LONG_HOSTNAME[:len(TOO_LONG_HOSTNAME)-1],
						Properties: properties,
						HealthCheck: &pb.HealthCheck{
							Mode:     "pull",
							Port:     math.MaxUint16,
							Interval: math.MaxInt32,
							Times:    math.MaxInt32,
							Url:      TOO_LONG_URL[:len(TOO_LONG_URL)-1],
						},
						DataCenterInfo: &pb.DataCenterInfo{
							Name:          TOO_LONG_SERVICENAME[:len(TOO_LONG_SERVICENAME)-1],
							Region:        TOO_LONG_SERVICENAME[:len(TOO_LONG_SERVICENAME)-1],
							AvailableZone: TOO_LONG_SERVICENAME[:len(TOO_LONG_SERVICENAME)-1],
						},
					},
				})
				Expect(err).To(BeNil())
				Expect(resp.Response.Code).To(Equal(pb.Response_SUCCESS))
				Expect(resp.InstanceId).To(Not(Equal("")))
			})
		})

		Context("when update the same instance", func() {
			It("should be passed", func() {
				instance := &pb.MicroServiceInstance{
					ServiceId: serviceId1,
					Endpoints: []string{
						"sameInstance:127.0.0.1:8080",
					},
					HostName: "UT-HOST",
					Status:   pb.MSI_UP,
				}
				resp, err := instanceResource.Register(getContext(), &pb.RegisterInstanceRequest{
					Instance: instance,
				})
				Expect(err).To(BeNil())
				Expect(resp.Response.Code).To(Equal(pb.Response_SUCCESS))

				resp, err = instanceResource.Register(getContext(), &pb.RegisterInstanceRequest{
					Instance: instance,
				})
				Expect(err).To(BeNil())
				Expect(resp.Response.Code).To(Equal(pb.Response_SUCCESS))
				Expect(resp.InstanceId).To(Equal(instance.InstanceId))
			})
		})

		Context("when register invalid instance", func() {
			It("should be failed", func() {
				By("endpoints are empty")
				resp, err := instanceResource.Register(getContext(), &pb.RegisterInstanceRequest{
					Instance: &pb.MicroServiceInstance{
						ServiceId: serviceId1,
						Endpoints: []string{""},
						HostName:  "UT-HOST",
						Status:    pb.MSI_UP,
					},
				})
				Expect(err).To(BeNil())
				Expect(resp.Response.Code).To(Equal(scerr.ErrInvalidParams))

				By("serviceId is empty")
				resp, err = instanceResource.Register(getContext(), &pb.RegisterInstanceRequest{
					Instance: &pb.MicroServiceInstance{
						HostName: "UT-HOST",
						Endpoints: []string{
							"check:127.0.0.1:8080",
						},
						Status: pb.MSI_UP,
					},
				})
				Expect(err).To(BeNil())
				Expect(resp.Response.Code).ToNot(Equal(pb.Response_SUCCESS))

				By("service does not exist")
				resp, err = instanceResource.Register(getContext(), &pb.RegisterInstanceRequest{
					Instance: &pb.MicroServiceInstance{
						ServiceId: "1000000000000000",
						Endpoints: []string{
							"check:127.0.0.1:8080",
						},
						HostName: "UT-HOST",
						Status:   pb.MSI_UP,
					},
				})
				Expect(err).To(BeNil())
				Expect(resp.Response.Code).ToNot(Equal(pb.Response_SUCCESS))

				By("hostName is empty")
				resp, err = instanceResource.Register(getContext(), &pb.RegisterInstanceRequest{
					Instance: &pb.MicroServiceInstance{
						ServiceId: serviceId1,
						Endpoints: []string{
							"check:127.0.0.1:8080",
						},
						Status: pb.MSI_UP,
					},
				})
				Expect(err).To(BeNil())
				Expect(resp.Response.Code).ToNot(Equal(pb.Response_SUCCESS))
				resp, err = instanceResource.Register(getContext(), &pb.RegisterInstanceRequest{
					Instance: &pb.MicroServiceInstance{
						ServiceId: serviceId1,
						HostName:  " ",
						Endpoints: []string{
							"check:127.0.0.1:8080",
						},
						Status: pb.MSI_UP,
					},
				})
				Expect(err).To(BeNil())
				Expect(resp.Response.Code).ToNot(Equal(pb.Response_SUCCESS))
				resp, err = instanceResource.Register(getContext(), &pb.RegisterInstanceRequest{
					Instance: &pb.MicroServiceInstance{
						ServiceId: serviceId1,
						HostName:  TOO_LONG_HOSTNAME,
						Endpoints: []string{
							"check:127.0.0.1:8080",
						},
						Status: pb.MSI_UP,
					},
				})
				Expect(err).To(BeNil())
				Expect(resp.Response.Code).ToNot(Equal(pb.Response_SUCCESS))

				By("instance is nil")
				resp, err = instanceResource.Register(getContext(), &pb.RegisterInstanceRequest{
					Instance: nil,
				})
				Expect(err).To(BeNil())
				Expect(resp.Response.Code).ToNot(Equal(pb.Response_SUCCESS))

				By("check normal push healthChceck")
				resp, err = instanceResource.Register(getContext(), &pb.RegisterInstanceRequest{
					Instance: &pb.MicroServiceInstance{
						ServiceId: serviceId1,
						Endpoints: []string{
							"checkpull:127.0.0.1:8080",
						},
						HostName: "UT-HOST",
						Status:   pb.MSI_UP,
						HealthCheck: &pb.HealthCheck{
							Mode:     "push",
							Interval: 30,
							Times:    1,
						},
					},
				})
				Expect(err).To(BeNil())
				Expect(resp.Response.Code).To(Equal(pb.Response_SUCCESS))

				resp, err = instanceResource.Register(getContext(), &pb.RegisterInstanceRequest{
					Instance: &pb.MicroServiceInstance{
						ServiceId: serviceId1,
						Endpoints: []string{
							"checkpush:127.0.0.1:8081",
						},
						HostName: "UT-HOST",
						Status:   pb.MSI_UP,
						HealthCheck: &pb.HealthCheck{
							Mode:     "push",
							Interval: 30,
							Times:    0,
						},
					},
				})
				Expect(err).To(BeNil())
				Expect(resp.Response.Code).To(Equal(pb.Response_SUCCESS))

				By("check normal pull healthChceck")
				resp, err = instanceResource.Register(getContext(), &pb.RegisterInstanceRequest{
					Instance: &pb.MicroServiceInstance{
						ServiceId: serviceId1,
						Endpoints: []string{
							"checkpush:127.0.0.1:8080",
						},
						HostName: "UT-HOST",
						Status:   pb.MSI_UP,
						HealthCheck: &pb.HealthCheck{
							Mode:     "pull",
							Interval: 30,
							Times:    1,
							Url:      "/abc/d",
						},
					},
				})
				Expect(err).To(BeNil())
				Expect(resp.Response.Code).To(Equal(pb.Response_SUCCESS))

				By("check invalid push healthChceck")
				resp, err = instanceResource.Register(getContext(), &pb.RegisterInstanceRequest{
					Instance: &pb.MicroServiceInstance{
						ServiceId: serviceId1,
						Endpoints: []string{
							"checkpush:127.0.0.1:8081",
						},
						HostName: "UT-HOST",
						Status:   pb.MSI_UP,
						HealthCheck: &pb.HealthCheck{
							Mode:     "push",
							Interval: 30,
							Times:    -1,
						},
					},
				})
				Expect(err).To(BeNil())
				Expect(resp.Response.Code).To(Equal(scerr.ErrInvalidParams))

				By("check invalid pull healthChceck")
				resp, err = instanceResource.Register(getContext(), &pb.RegisterInstanceRequest{
					Instance: &pb.MicroServiceInstance{
						ServiceId: serviceId1,
						Endpoints: []string{
							"checkpull:127.0.0.1:8081",
						},
						HostName: "UT-HOST",
						Status:   pb.MSI_UP,
						HealthCheck: &pb.HealthCheck{
							Mode:     "pull",
							Interval: 30,
							Times:    1,
							Url:      " ",
						},
					},
				})
				Expect(err).To(BeNil())
				Expect(resp.Response.Code).To(Equal(scerr.ErrInvalidParams))
				resp, err = instanceResource.Register(getContext(), &pb.RegisterInstanceRequest{
					Instance: &pb.MicroServiceInstance{
						ServiceId: serviceId1,
						Endpoints: []string{
							"checkpull:127.0.0.1:8081",
						},
						HostName: "UT-HOST",
						Status:   pb.MSI_UP,
						HealthCheck: &pb.HealthCheck{
							Mode:     "pull",
							Interval: 0,
							Times:    0,
						},
					},
				})
				Expect(err).To(BeNil())
				Expect(resp.Response.Code).To(Equal(scerr.ErrInvalidParams))
				resp, err = instanceResource.Register(getContext(), &pb.RegisterInstanceRequest{
					Instance: &pb.MicroServiceInstance{
						ServiceId: serviceId1,
						Endpoints: []string{
							"checkpull:127.0.0.1:8081",
						},
						HostName: "UT-HOST",
						Status:   pb.MSI_UP,
						HealthCheck: &pb.HealthCheck{
							Mode:     "pull",
							Interval: 30,
							Times:    1,
							Port:     math.MaxUint16 + 1,
						},
					},
				})
				Expect(err).To(BeNil())
				Expect(resp.Response.Code).To(Equal(scerr.ErrInvalidParams))

				By("invalid status")
				resp, err = instanceResource.Register(getContext(), &pb.RegisterInstanceRequest{
					Instance: &pb.MicroServiceInstance{
						ServiceId: serviceId1,
						Endpoints: []string{
							"createInstance:127.0.0.1:8083",
						},
						HostName: "UT-HOST",
						Status:   "Invalid",
					},
				})
				Expect(err).To(BeNil())
				Expect(resp.Response.Code).ToNot(Equal(pb.Response_SUCCESS))
			})
		})
	})

	Describe("execute 'heartbeat' operartion", func() {
		var (
			serviceId   string
			instanceId1 string
			instanceId2 string
		)

		It("should be passed", func() {
			respCreate, err := serviceResource.Create(getContext(), &pb.CreateServiceRequest{
				Service: &pb.MicroService{
					ServiceName: "heartbeat_service",
					AppId:       "heartbeat_service",
					Version:     "1.0.0",
					Level:       "FRONT",
					Status:      pb.MS_UP,
				},
			})
			Expect(err).To(BeNil())
			Expect(respCreate.Response.Code).To(Equal(pb.Response_SUCCESS))
			serviceId = respCreate.ServiceId

			resp, err := instanceResource.Register(getContext(), &pb.RegisterInstanceRequest{
				Instance: &pb.MicroServiceInstance{
					ServiceId: serviceId,
					HostName:  "UT-HOST",
					Endpoints: []string{
						"heartbeat:127.0.0.1:8080",
					},
					Status: pb.MSI_UP,
				},
			})
			Expect(err).To(BeNil())
			Expect(resp.Response.Code).To(Equal(pb.Response_SUCCESS))
			instanceId1 = resp.InstanceId

			resp, err = instanceResource.Register(getContext(), &pb.RegisterInstanceRequest{
				Instance: &pb.MicroServiceInstance{
					ServiceId: serviceId,
					HostName:  "UT-HOST",
					Endpoints: []string{
						"heartbeat:127.0.0.2:8080",
					},
					Status: pb.MSI_UP,
				},
			})
			Expect(err).To(BeNil())
			Expect(resp.Response.Code).To(Equal(pb.Response_SUCCESS))
			instanceId2 = resp.InstanceId
		})

		Context("when update a lease", func() {
			It("should be passed", func() {
				By("valid instance")
				resp, err := instanceResource.Heartbeat(getContext(), &pb.HeartbeatRequest{
					ServiceId:  serviceId,
					InstanceId: instanceId1,
				})
				Expect(err).To(BeNil())
				Expect(resp.Response.Code).To(Equal(pb.Response_SUCCESS))

				By("serviceId/instanceId is invalid")
				resp, err = instanceResource.Heartbeat(getContext(), &pb.HeartbeatRequest{
					ServiceId:  "",
					InstanceId: instanceId1,
				})
				Expect(err).To(BeNil())
				Expect(resp.Response.Code).To(Equal(scerr.ErrInvalidParams))
				resp, err = instanceResource.Heartbeat(getContext(), &pb.HeartbeatRequest{
					ServiceId:  TOO_LONG_SERVICEID,
					InstanceId: instanceId1,
				})
				Expect(err).To(BeNil())
				Expect(resp.Response.Code).To(Equal(scerr.ErrInvalidParams))
				resp, err = instanceResource.Heartbeat(getContext(), &pb.HeartbeatRequest{
					ServiceId:  serviceId,
					InstanceId: "",
				})
				Expect(err).To(BeNil())
				Expect(resp.Response.Code).To(Equal(scerr.ErrInvalidParams))
				resp, err = instanceResource.Heartbeat(getContext(), &pb.HeartbeatRequest{
					ServiceId:  serviceId,
					InstanceId: "@",
				})
				Expect(err).To(BeNil())
				Expect(resp.Response.Code).To(Equal(scerr.ErrInvalidParams))
				resp, err = instanceResource.Heartbeat(getContext(), &pb.HeartbeatRequest{
					ServiceId:  serviceId,
					InstanceId: TOO_LONG_SERVICEID,
				})
				Expect(err).To(BeNil())
				Expect(resp.Response.Code).To(Equal(scerr.ErrInvalidParams))

				By("serviceId does not exist")
				resp, err = instanceResource.Heartbeat(getContext(), &pb.HeartbeatRequest{
					ServiceId:  "100000000000",
					InstanceId: instanceId1,
				})
				Expect(resp.Response.Code).ToNot(Equal(pb.Response_SUCCESS))

				By("instance does not exist")
				resp, err = instanceResource.Heartbeat(getContext(), &pb.HeartbeatRequest{
					ServiceId:  serviceId,
					InstanceId: "not-exist-ins",
				})
				Expect(resp.Response.Code).ToNot(Equal(pb.Response_SUCCESS))
			})
		})

		Context("when batch update lease", func() {
			It("should be passed", func() {
				By("instances are empty")
				resp, err := instanceResource.HeartbeatSet(getContext(), &pb.HeartbeatSetRequest{
					Instances: []*pb.HeartbeatSetElement{},
				})
				Expect(err).To(BeNil())
				Expect(resp.Response.Code).ToNot(Equal(pb.Response_SUCCESS))

				By("instances are nil")
				resp, err = instanceResource.HeartbeatSet(getContext(), &pb.HeartbeatSetRequest{})
				Expect(err).To(BeNil())
				Expect(resp.Response.Code).ToNot(Equal(pb.Response_SUCCESS))

				By("request contains > 1 instances")
				resp, err = instanceResource.HeartbeatSet(getContext(), &pb.HeartbeatSetRequest{
					Instances: []*pb.HeartbeatSetElement{
						{
							ServiceId:  serviceId,
							InstanceId: instanceId1,
						},
						{
							ServiceId:  serviceId,
							InstanceId: instanceId2,
						},
					},
				})
				Expect(resp.Response.Code).To(Equal(pb.Response_SUCCESS))

				By("request contains invalid instance")
				resp, err = instanceResource.HeartbeatSet(getContext(), &pb.HeartbeatSetRequest{
					Instances: []*pb.HeartbeatSetElement{
						{
							ServiceId:  serviceId,
							InstanceId: instanceId1,
						},
						{
							ServiceId:  serviceId,
							InstanceId: "not-exist-instanceId",
						},
					},
				})
				Expect(resp.Response.Code).ToNot(Equal(pb.Response_SUCCESS))
			})
		})
	})

	Describe("execute 'clusterHealth' operartion", func() {
		var (
			scServiceId string
		)

		It("should be passed", func() {
			resp, err := serviceResource.Create(getContext(), core.CreateServiceRequest())
			Expect(err).To(BeNil())
			scServiceId = resp.ServiceId
			Expect(resp.Response.Code).To(Equal(pb.Response_SUCCESS))
		})

		Context("when SC does not exist", func() {
			It("should be failed", func() {
				old := core.Service.ServiceName
				core.Service.ServiceName = "x"
				respCluterhealth, err := instanceResource.ClusterHealth(getContext())
				Expect(err).To(BeNil())
				Expect(respCluterhealth.Response.Code).ToNot(Equal(pb.Response_SUCCESS))
				core.Service.ServiceName = old
			})
		})

		Context("when SC registered", func() {
			It("should be passed", func() {
				respCluterhealth, err := instanceResource.ClusterHealth(getContext())
				Expect(err).To(BeNil())
				Expect(respCluterhealth.Response.Code).To(Equal(pb.Response_SUCCESS))
			})
		})
	})

	Describe("execute 'udpate' operartion", func() {
		var (
			serviceId  string
			instanceId string
		)

		It("should be passed", func() {
			respCreate, err := serviceResource.Create(getContext(), &pb.CreateServiceRequest{
				Service: &pb.MicroService{
					ServiceName: "update_instance_service",
					AppId:       "update_instance_service",
					Version:     "1.0.0",
					Level:       "FRONT",
					Status:      pb.MS_UP,
				},
			})
			Expect(err).To(BeNil())
			Expect(respCreate.Response.Code).To(Equal(pb.Response_SUCCESS))
			serviceId = respCreate.ServiceId

			resp, err := instanceResource.Register(getContext(), &pb.RegisterInstanceRequest{
				Instance: &pb.MicroServiceInstance{
					ServiceId: serviceId,
					Endpoints: []string{
						"updateInstance:127.0.0.1:8080",
					},
					HostName:   "UT-HOST",
					Status:     pb.MSI_UP,
					Properties: map[string]string{"nodeIP": "test"},
				},
			})
			Expect(err).To(BeNil())
			Expect(resp.Response.Code).To(Equal(pb.Response_SUCCESS))
			instanceId = resp.InstanceId
		})

		Context("when update instance status", func() {
			It("should be passed", func() {
				By("update instance status")
				respUpdateStatus, err := instanceResource.UpdateStatus(getContext(), &pb.UpdateInstanceStatusRequest{
					ServiceId:  serviceId,
					InstanceId: instanceId,
					Status:     pb.MSI_DOWN,
				})
				Expect(err).To(BeNil())
				Expect(respUpdateStatus.Response.Code).To(Equal(pb.Response_SUCCESS))

				respUpdateStatus, err = instanceResource.UpdateStatus(getContext(), &pb.UpdateInstanceStatusRequest{
					ServiceId:  serviceId,
					InstanceId: instanceId,
					Status:     pb.MSI_OUTOFSERVICE,
				})
				Expect(err).To(BeNil())
				Expect(respUpdateStatus.Response.Code).To(Equal(pb.Response_SUCCESS))

				respUpdateStatus, err = instanceResource.UpdateStatus(getContext(), &pb.UpdateInstanceStatusRequest{
					ServiceId:  serviceId,
					InstanceId: instanceId,
					Status:     pb.MSI_STARTING,
				})
				Expect(err).To(BeNil())
				Expect(respUpdateStatus.Response.Code).To(Equal(pb.Response_SUCCESS))

				respUpdateStatus, err = instanceResource.UpdateStatus(getContext(), &pb.UpdateInstanceStatusRequest{
					ServiceId:  serviceId,
					InstanceId: instanceId,
					Status:     pb.MSI_TESTING,
				})

				Expect(err).To(BeNil())
				Expect(respUpdateStatus.Response.Code).To(Equal(pb.Response_SUCCESS))

				respUpdateStatus, err = instanceResource.UpdateStatus(getContext(), &pb.UpdateInstanceStatusRequest{
					ServiceId:  serviceId,
					InstanceId: instanceId,
					Status:     pb.MSI_UP,
				})
				Expect(err).To(BeNil())
				Expect(respUpdateStatus.Response.Code).To(Equal(pb.Response_SUCCESS))

				By("update status with a empty serviceId")
				respUpdateStatus, err = instanceResource.UpdateStatus(getContext(), &pb.UpdateInstanceStatusRequest{
					ServiceId:  "",
					InstanceId: instanceId,
					Status:     pb.MSI_STARTING,
				})

				Expect(err).To(BeNil())
				Expect(respUpdateStatus.Response.Code).ToNot(Equal(pb.Response_SUCCESS))

				By("update status with a not exist service")
				respUpdateStatus, err = instanceResource.UpdateStatus(getContext(), &pb.UpdateInstanceStatusRequest{
					ServiceId:  "notexistservice",
					InstanceId: instanceId,
					Status:     pb.MSI_STARTING,
				})

				Expect(err).To(BeNil())
				Expect(respUpdateStatus.Response.Code).ToNot(Equal(pb.Response_SUCCESS))

				By("update status with a not exist instance")
				respUpdateStatus, err = instanceResource.UpdateStatus(getContext(), &pb.UpdateInstanceStatusRequest{
					ServiceId:  serviceId,
					InstanceId: "notexistins",
					Status:     pb.MSI_STARTING,
				})

				Expect(err).To(BeNil())
				Expect(respUpdateStatus.Response.Code).ToNot(Equal(pb.Response_SUCCESS))

				By("update invalid status")
				respUpdateStatus, err = instanceResource.UpdateStatus(getContext(), &pb.UpdateInstanceStatusRequest{
					ServiceId:  serviceId,
					InstanceId: instanceId,
					Status:     "nonestatus",
				})

				Expect(err).To(BeNil())
				Expect(respUpdateStatus.Response.Code).ToNot(Equal(pb.Response_SUCCESS))
			})
		})

		Context("when update instance properties", func() {
			It("should be passed", func() {
				By("update instance properties")
				respUpdateProperties, err := instanceResource.UpdateInstanceProperties(getContext(), &pb.UpdateInstancePropsRequest{
					ServiceId:  serviceId,
					InstanceId: instanceId,
					Properties: map[string]string{
						"test": "test",
					},
				})

				Expect(err).To(BeNil())
				Expect(respUpdateProperties.Response.Code).To(Equal(pb.Response_SUCCESS))

				By("all max")
				size := 1000
				properties := make(map[string]string, size)
				for i := 0; i < size; i++ {
					s := strconv.Itoa(i) + strings.Repeat("x", 253)
					properties[s] = s
				}
				respUpdateProperties, err = instanceResource.UpdateInstanceProperties(getContext(), &pb.UpdateInstancePropsRequest{
					ServiceId:  serviceId,
					InstanceId: instanceId,
					Properties: properties,
				})

				Expect(err).To(BeNil())
				Expect(respUpdateProperties.Response.Code).To(Equal(pb.Response_SUCCESS))

				By("instance does not exist")
				respUpdateProperties, err = instanceResource.UpdateInstanceProperties(getContext(), &pb.UpdateInstancePropsRequest{
					ServiceId:  serviceId,
					InstanceId: "notexistins",
					Properties: map[string]string{
						"test": "test",
					},
				})

				Expect(err).To(BeNil())
				Expect(respUpdateProperties.Response.Code).ToNot(Equal(pb.Response_SUCCESS))

				By("serviceId/instanceId/props is invalid")
				respUpdateProperties, err = instanceResource.UpdateInstanceProperties(getContext(), &pb.UpdateInstancePropsRequest{
					ServiceId:  "",
					InstanceId: instanceId,
					Properties: map[string]string{
						"test": "test",
					},
				})

				Expect(err).To(BeNil())
				Expect(respUpdateProperties.Response.Code).To(Equal(scerr.ErrInvalidParams))
				respUpdateProperties, err = instanceResource.UpdateInstanceProperties(getContext(), &pb.UpdateInstancePropsRequest{
					ServiceId:  TOO_LONG_SERVICEID,
					InstanceId: instanceId,
					Properties: map[string]string{
						"test": "test",
					},
				})

				Expect(err).To(BeNil())
				Expect(respUpdateProperties.Response.Code).To(Equal(scerr.ErrInvalidParams))
				respUpdateProperties, err = instanceResource.UpdateInstanceProperties(getContext(), &pb.UpdateInstancePropsRequest{
					ServiceId:  serviceId,
					InstanceId: "",
					Properties: map[string]string{
						"test": "test",
					},
				})

				Expect(err).To(BeNil())
				Expect(respUpdateProperties.Response.Code).To(Equal(scerr.ErrInvalidParams))
				respUpdateProperties, err = instanceResource.UpdateInstanceProperties(getContext(), &pb.UpdateInstancePropsRequest{
					ServiceId:  serviceId,
					InstanceId: "@",
					Properties: map[string]string{
						"test": "test",
					},
				})

				Expect(err).To(BeNil())
				Expect(respUpdateProperties.Response.Code).To(Equal(scerr.ErrInvalidParams))
				respUpdateProperties, err = instanceResource.UpdateInstanceProperties(getContext(), &pb.UpdateInstancePropsRequest{
					ServiceId:  serviceId,
					InstanceId: TOO_LONG_SERVICEID,
					Properties: map[string]string{
						"test": "test",
					},
				})

				Expect(err).To(BeNil())
				Expect(respUpdateProperties.Response.Code).To(Equal(scerr.ErrInvalidParams))

				By("remove the properties")
				respUpdateProperties, err = instanceResource.UpdateInstanceProperties(getContext(), &pb.UpdateInstancePropsRequest{
					ServiceId:  serviceId,
					InstanceId: instanceId,
				})
				Expect(err).To(BeNil())
				Expect(respUpdateProperties.Response.Code).To(Equal(pb.Response_SUCCESS))

				By("service does not exist")
				respUpdateProperties, err = instanceResource.UpdateInstanceProperties(getContext(), &pb.UpdateInstancePropsRequest{
					ServiceId:  "notexistservice",
					InstanceId: instanceId,
					Properties: map[string]string{
						"test": "test",
					},
				})

				Expect(err).To(BeNil())
				Expect(respUpdateProperties.Response.Code).ToNot(Equal(pb.Response_SUCCESS))
			})
		})
	})

	Describe("execute 'query' operartion", func() {
		var (
			serviceId1  string
			serviceId2  string
			serviceId3  string
			serviceId4  string
			serviceId5  string
			serviceId6  string
			serviceId7  string
			serviceId8  string
			serviceId9  string
			instanceId1 string
			instanceId2 string
			instanceId4 string
			instanceId5 string
			instanceId8 string
			instanceId9 string
		)

		It("should be passed", func() {
			respCreate, err := serviceResource.Create(getContext(), &pb.CreateServiceRequest{
				Service: &pb.MicroService{
					AppId:       "query_instance",
					ServiceName: "query_instance_service",
					Version:     "1.0.0",
					Level:       "FRONT",
					Status:      pb.MS_UP,
				},
			})
			Expect(err).To(BeNil())
			Expect(respCreate.Response.Code).To(Equal(pb.Response_SUCCESS))
			serviceId1 = respCreate.ServiceId

			respCreate, err = serviceResource.Create(getContext(), &pb.CreateServiceRequest{
				Service: &pb.MicroService{
					AppId:       "query_instance",
					ServiceName: "query_instance_service",
					Version:     "1.0.5",
					Level:       "FRONT",
					Status:      pb.MS_UP,
				},
			})
			Expect(err).To(BeNil())
			Expect(respCreate.Response.Code).To(Equal(pb.Response_SUCCESS))
			serviceId2 = respCreate.ServiceId

			respCreate, err = serviceResource.Create(getContext(), &pb.CreateServiceRequest{
				Service: &pb.MicroService{
					AppId:       "query_instance_diff_app",
					ServiceName: "query_instance_service",
					Version:     "1.0.0",
					Level:       "FRONT",
					Status:      pb.MS_UP,
				},
			})
			Expect(err).To(BeNil())
			Expect(respCreate.Response.Code).To(Equal(pb.Response_SUCCESS))
			serviceId3 = respCreate.ServiceId

			respCreate, err = serviceResource.Create(getContext(), &pb.CreateServiceRequest{
				Service: &pb.MicroService{
					Environment: pb.ENV_PROD,
					AppId:       "query_instance",
					ServiceName: "query_instance_diff_env_service",
					Version:     "1.0.0",
					Level:       "FRONT",
					Status:      pb.MS_UP,
				},
			})
			Expect(err).To(BeNil())
			Expect(respCreate.Response.Code).To(Equal(pb.Response_SUCCESS))
			serviceId4 = respCreate.ServiceId

			respCreate, err = serviceResource.Create(getContext(), &pb.CreateServiceRequest{
				Service: &pb.MicroService{
					Environment: pb.ENV_PROD,
					AppId:       "default",
					ServiceName: "query_instance_shared_provider",
					Version:     "1.0.0",
					Level:       "FRONT",
					Status:      pb.MS_UP,
					Properties: map[string]string{
						pb.PROP_ALLOW_CROSS_APP: "true",
					},
				},
			})
			Expect(err).To(BeNil())
			Expect(respCreate.Response.Code).To(Equal(pb.Response_SUCCESS))
			serviceId5 = respCreate.ServiceId

			respCreate, err = serviceResource.Create(
				util.SetDomainProject(util.CloneContext(getContext()), "user", "user"),
				&pb.CreateServiceRequest{
					Service: &pb.MicroService{
						AppId:       "default",
						ServiceName: "query_instance_diff_domain_consumer",
						Version:     "1.0.0",
						Level:       "FRONT",
						Status:      pb.MS_UP,
					},
				})
			Expect(err).To(BeNil())
			Expect(respCreate.Response.Code).To(Equal(pb.Response_SUCCESS))
			serviceId6 = respCreate.ServiceId

			respCreate, err = serviceResource.Create(getContext(), &pb.CreateServiceRequest{
				Service: &pb.MicroService{
					AppId:       "default",
					ServiceName: "query_instance_shared_consumer",
					Version:     "1.0.0",
					Level:       "FRONT",
					Status:      pb.MS_UP,
				},
			})
			Expect(err).To(BeNil())
			Expect(respCreate.Response.Code).To(Equal(pb.Response_SUCCESS))
			serviceId7 = respCreate.ServiceId

			respCreate, err = serviceResource.Create(getContext(), &pb.CreateServiceRequest{
				Service: &pb.MicroService{
					AppId:       "query_instance",
					ServiceName: "query_instance_with_rev",
					Version:     "1.0.0",
					Level:       "FRONT",
					Status:      pb.MS_UP,
				},
			})
			Expect(err).To(BeNil())
			Expect(respCreate.Response.Code).To(Equal(pb.Response_SUCCESS))
			serviceId8 = respCreate.ServiceId

			respCreate, err = serviceResource.Create(getContext(), &pb.CreateServiceRequest{
				Service: &pb.MicroService{
					AppId:       "query_instance",
					ServiceName: "batch_query_instance_with_rev",
					Version:     "1.0.0",
					Level:       "FRONT",
					Status:      pb.MS_UP,
				},
			})
			Expect(err).To(BeNil())
			Expect(respCreate.Response.Code).To(Equal(pb.Response_SUCCESS))
			serviceId9 = respCreate.ServiceId

			resp, err := instanceResource.Register(getContext(), &pb.RegisterInstanceRequest{
				Instance: &pb.MicroServiceInstance{
					ServiceId: serviceId1,
					HostName:  "UT-HOST",
					Endpoints: []string{
						"find:127.0.0.1:8080",
					},
					Status: pb.MSI_UP,
				},
			})
			Expect(err).To(BeNil())
			Expect(resp.Response.Code).To(Equal(pb.Response_SUCCESS))
			instanceId1 = resp.InstanceId

			resp, err = instanceResource.Register(getContext(), &pb.RegisterInstanceRequest{
				Instance: &pb.MicroServiceInstance{
					ServiceId: serviceId2,
					HostName:  "UT-HOST",
					Endpoints: []string{
						"find:127.0.0.2:8080",
					},
					Status: pb.MSI_UP,
				},
			})
			Expect(err).To(BeNil())
			Expect(resp.Response.Code).To(Equal(pb.Response_SUCCESS))
			instanceId2 = resp.InstanceId

			resp, err = instanceResource.Register(getContext(), &pb.RegisterInstanceRequest{
				Instance: &pb.MicroServiceInstance{
					ServiceId: serviceId4,
					HostName:  "UT-HOST",
					Endpoints: []string{
						"find:127.0.0.4:8080",
					},
					Status: pb.MSI_UP,
				},
			})
			Expect(err).To(BeNil())
			Expect(resp.Response.Code).To(Equal(pb.Response_SUCCESS))
			instanceId4 = resp.InstanceId

			resp, err = instanceResource.Register(getContext(), &pb.RegisterInstanceRequest{
				Instance: &pb.MicroServiceInstance{
					ServiceId: serviceId5,
					HostName:  "UT-HOST",
					Endpoints: []string{
						"find:127.0.0.5:8080",
					},
					Status: pb.MSI_UP,
				},
			})
			Expect(err).To(BeNil())
			Expect(resp.Response.Code).To(Equal(pb.Response_SUCCESS))
			instanceId5 = resp.InstanceId

			resp, err = instanceResource.Register(getContext(), &pb.RegisterInstanceRequest{
				Instance: &pb.MicroServiceInstance{
					ServiceId: serviceId8,
					HostName:  "UT-HOST",
					Endpoints: []string{
						"find:127.0.0.8:8080",
					},
					Status: pb.MSI_UP,
				},
			})
			Expect(err).To(BeNil())
			Expect(resp.Response.Code).To(Equal(pb.Response_SUCCESS))
			instanceId8 = resp.InstanceId

			resp, err = instanceResource.Register(getContext(), &pb.RegisterInstanceRequest{
				Instance: &pb.MicroServiceInstance{
					ServiceId: serviceId9,
					HostName:  "UT-HOST",
					Endpoints: []string{
						"find:127.0.0.9:8080",
					},
					Status: pb.MSI_UP,
				},
			})
			Expect(err).To(BeNil())
			Expect(resp.Response.Code).To(Equal(pb.Response_SUCCESS))
			instanceId9 = resp.InstanceId
		})

		Context("when query invalid parameters", func() {
			It("should be failed", func() {
				By("invalid appId")
				respFind, err := instanceResource.Find(getContext(), &pb.FindInstancesRequest{
					ConsumerServiceId: serviceId1,
					AppId:             TOO_LONG_APPID,
					ServiceName:       "query_instance_service",
					VersionRule:       "1.0.0",
				})
				Expect(err).To(BeNil())
				Expect(respFind.Response.Code).To(Equal(scerr.ErrInvalidParams))
				respFind, err = instanceResource.Find(getContext(), &pb.FindInstancesRequest{
					ConsumerServiceId: serviceId1,
					AppId:             "",
					ServiceName:       "query_instance_service",
					VersionRule:       "1.0.0",
				})
				Expect(err).To(BeNil())
				Expect(respFind.Response.Code).To(Equal(scerr.ErrInvalidParams))
				respFind, err = instanceResource.Find(getContext(), &pb.FindInstancesRequest{
					ConsumerServiceId: serviceId1,
					AppId:             " ",
					ServiceName:       "query_instance_service",
					VersionRule:       "1.0.0",
				})
				Expect(err).To(BeNil())
				Expect(respFind.Response.Code).To(Equal(scerr.ErrInvalidParams))

				By("invalid serviceName")
				respFind, err = instanceResource.Find(getContext(), &pb.FindInstancesRequest{
					ConsumerServiceId: serviceId1,
					AppId:             "query_instance",
					ServiceName:       TOO_LONG_EXISTENCE,
					VersionRule:       "1.0.0",
				})
				Expect(err).To(BeNil())
				Expect(respFind.Response.Code).To(Equal(scerr.ErrInvalidParams))
				respFind, err = instanceResource.Find(getContext(), &pb.FindInstancesRequest{
					ConsumerServiceId: serviceId1,
					AppId:             "query_instance",
					ServiceName:       "",
					VersionRule:       "1.0.0",
				})
				Expect(err).To(BeNil())
				Expect(respFind.Response.Code).To(Equal(scerr.ErrInvalidParams))
				respFind, err = instanceResource.Find(getContext(), &pb.FindInstancesRequest{
					ConsumerServiceId: serviceId1,
					AppId:             "query_instance",
					ServiceName:       " ",
					VersionRule:       "1.0.0",
				})
				Expect(err).To(BeNil())
				Expect(respFind.Response.Code).To(Equal(scerr.ErrInvalidParams))

				By("invalid version")
				respFind, err = instanceResource.Find(getContext(), &pb.FindInstancesRequest{
					ConsumerServiceId: serviceId1,
					AppId:             "query_instance",
					ServiceName:       "query_instance_service",
					VersionRule:       "1.32768.0",
				})
				Expect(err).To(BeNil())
				Expect(respFind.Response.Code).To(Equal(scerr.ErrInvalidParams))
				respFind, err = instanceResource.Find(getContext(), &pb.FindInstancesRequest{
					ConsumerServiceId: serviceId1,
					AppId:             "query_instance",
					ServiceName:       "query_instance_service",
					VersionRule:       "1.0.0-1.32768.0",
				})
				Expect(err).To(BeNil())
				Expect(respFind.Response.Code).To(Equal(scerr.ErrInvalidParams))
				respFind, err = instanceResource.Find(getContext(), &pb.FindInstancesRequest{
					ConsumerServiceId: serviceId1,
					AppId:             "query_instance",
					ServiceName:       "query_instance_service",
					VersionRule:       " ",
				})
				Expect(err).To(BeNil())
				Expect(respFind.Response.Code).To(Equal(scerr.ErrInvalidParams))
				respFind, err = instanceResource.Find(getContext(), &pb.FindInstancesRequest{
					ConsumerServiceId: serviceId1,
					AppId:             "query_instance",
					ServiceName:       "query_instance_service",
					VersionRule:       "",
				})
				Expect(err).To(BeNil())
				Expect(respFind.Response.Code).To(Equal(scerr.ErrInvalidParams))

				By("consumerId is empty")
				respFind, err = instanceResource.Find(getContext(), &pb.FindInstancesRequest{
					ConsumerServiceId: "",
					AppId:             "query_instance",
					ServiceName:       "query_instance_service",
					VersionRule:       "1.0.0+",
				})
				Expect(err).To(BeNil())
				Expect(respFind.Response.Code).To(Equal(pb.Response_SUCCESS))

				By("provider does not exist")
				respFind, err = instanceResource.Find(getContext(), &pb.FindInstancesRequest{
					ConsumerServiceId: serviceId1,
					AppId:             "query_instance",
					ServiceName:       "noneservice",
					VersionRule:       "latest",
				})
				Expect(err).To(BeNil())
				Expect(respFind.Response.Code).ToNot(Equal(pb.Response_SUCCESS))

				By("provider does not contain 3.0.0+ versions")
				respFind, err = instanceResource.Find(getContext(), &pb.FindInstancesRequest{
					ConsumerServiceId: serviceId1,
					AppId:             "query_instance",
					ServiceName:       "query_instance_service",
					VersionRule:       "3.0.0+",
				})
				Expect(err).To(BeNil())
				Expect(respFind.Response.Code).To(Equal(pb.Response_SUCCESS))
				Expect(len(respFind.Instances)).To(Equal(0))

				By("provider does not contain 2.0.0-2.0.1 versions")
				respFind, err = instanceResource.Find(getContext(), &pb.FindInstancesRequest{
					ConsumerServiceId: serviceId1,
					AppId:             "query_instance",
					ServiceName:       "query_instance_service",
					VersionRule:       "2.0.0-2.0.1",
				})
				Expect(err).To(BeNil())
				Expect(respFind.Response.Code).To(Equal(pb.Response_SUCCESS))
				Expect(len(respFind.Instances)).To(Equal(0))

				By("provider does not contain 2.0.0 version")
				respFind, err = instanceResource.Find(getContext(), &pb.FindInstancesRequest{
					ConsumerServiceId: serviceId1,
					AppId:             "query_instance",
					ServiceName:       "query_instance_service",
					VersionRule:       "2.0.0",
				})
				Expect(err).To(BeNil())
				Expect(respFind.Response.Code).ToNot(Equal(pb.Response_SUCCESS))

				By("consumer does not exist")
				respFind, err = instanceResource.Find(getContext(), &pb.FindInstancesRequest{
					ConsumerServiceId: "notExistServiceId",
					AppId:             "query_instance",
					ServiceName:       "query_instance_service",
					VersionRule:       "2.0.0",
				})
				Expect(err).To(BeNil())
				Expect(respFind.Response.Code).To(Equal(scerr.ErrServiceNotExists))
			})
		})

		Context("when batch query invalid parameters", func() {
			It("should be failed", func() {
				By("invalid services")
				respFind, err := instanceResource.BatchFind(getContext(), &pb.BatchFindInstancesRequest{
					ConsumerServiceId: serviceId1,
					Services:          nil,
					Instances:         nil,
				})
				Expect(err).To(BeNil())
				Expect(respFind.Response.Code).To(Equal(scerr.ErrInvalidParams))
				respFind, err = instanceResource.BatchFind(getContext(), &pb.BatchFindInstancesRequest{
					ConsumerServiceId: serviceId1,
					Services:          []*pb.FindService{},
					Instances:         []*pb.FindInstance{},
				})
				Expect(err).To(BeNil())
				Expect(respFind.Response.Code).To(Equal(scerr.ErrInvalidParams))
				respFind, err = instanceResource.BatchFind(getContext(), &pb.BatchFindInstancesRequest{
					ConsumerServiceId: serviceId1,
					Services:          []*pb.FindService{{}},
				})
				Expect(err).To(BeNil())
				Expect(respFind.Response.Code).To(Equal(scerr.ErrInvalidParams))
				respFind, err = instanceResource.BatchFind(getContext(), &pb.BatchFindInstancesRequest{
					ConsumerServiceId: serviceId1,
					Instances:         []*pb.FindInstance{{}},
				})
				Expect(err).To(BeNil())
				Expect(respFind.Response.Code).To(Equal(scerr.ErrInvalidParams))

				By("invalid appId")
				respFind, err = instanceResource.BatchFind(getContext(), &pb.BatchFindInstancesRequest{
					ConsumerServiceId: serviceId1,
					Services: []*pb.FindService{
						{
							Service: &pb.MicroServiceKey{
								AppId:       TOO_LONG_APPID,
								ServiceName: "query_instance_service",
								Version:     "1.0.0",
							},
						},
					},
				})
				Expect(err).To(BeNil())
				Expect(respFind.Response.Code).To(Equal(scerr.ErrInvalidParams))
				respFind, err = instanceResource.BatchFind(getContext(), &pb.BatchFindInstancesRequest{
					ConsumerServiceId: serviceId1,
					Services: []*pb.FindService{
						{
							Service: &pb.MicroServiceKey{
								AppId:       "",
								ServiceName: "query_instance_service",
								Version:     "1.0.0",
							},
						},
					},
				})
				Expect(err).To(BeNil())
				Expect(respFind.Response.Code).To(Equal(scerr.ErrInvalidParams))
				respFind, err = instanceResource.BatchFind(getContext(), &pb.BatchFindInstancesRequest{
					ConsumerServiceId: serviceId1,
					Services: []*pb.FindService{
						{
							Service: &pb.MicroServiceKey{
								AppId:       " ",
								ServiceName: "query_instance_service",
								Version:     "1.0.0",
							},
						},
					},
				})
				Expect(err).To(BeNil())
				Expect(respFind.Response.Code).To(Equal(scerr.ErrInvalidParams))

				By("invalid serviceName")
				respFind, err = instanceResource.BatchFind(getContext(), &pb.BatchFindInstancesRequest{
					ConsumerServiceId: serviceId1,
					Services: []*pb.FindService{
						{
							Service: &pb.MicroServiceKey{
								AppId:       "query_instance",
								ServiceName: TOO_LONG_EXISTENCE,
								Version:     "1.0.0",
							},
						},
					},
				})
				Expect(err).To(BeNil())
				Expect(respFind.Response.Code).To(Equal(scerr.ErrInvalidParams))
				respFind, err = instanceResource.BatchFind(getContext(), &pb.BatchFindInstancesRequest{
					ConsumerServiceId: serviceId1,
					Services: []*pb.FindService{
						{
							Service: &pb.MicroServiceKey{
								AppId:       "query_instance",
								ServiceName: "",
								Version:     "1.0.0",
							},
						},
					},
				})
				Expect(err).To(BeNil())
				Expect(respFind.Response.Code).To(Equal(scerr.ErrInvalidParams))
				respFind, err = instanceResource.BatchFind(getContext(), &pb.BatchFindInstancesRequest{
					ConsumerServiceId: serviceId1,
					Services: []*pb.FindService{
						{
							Service: &pb.MicroServiceKey{
								AppId:       "query_instance",
								ServiceName: " ",
								Version:     "1.0.0",
							},
						},
					},
				})
				Expect(err).To(BeNil())
				Expect(respFind.Response.Code).To(Equal(scerr.ErrInvalidParams))

				By("invalid version")
				respFind, err = instanceResource.BatchFind(getContext(), &pb.BatchFindInstancesRequest{
					ConsumerServiceId: serviceId1,
					Services: []*pb.FindService{
						{
							Service: &pb.MicroServiceKey{
								AppId:       "query_instance",
								ServiceName: "query_instance_service",
								Version:     "1.32768.0",
							},
						},
					},
				})
				Expect(err).To(BeNil())
				Expect(respFind.Response.Code).To(Equal(scerr.ErrInvalidParams))
				respFind, err = instanceResource.BatchFind(getContext(), &pb.BatchFindInstancesRequest{
					ConsumerServiceId: serviceId1,
					Services: []*pb.FindService{
						{
							Service: &pb.MicroServiceKey{
								AppId:       "query_instance",
								ServiceName: "query_instance_service",
								Version:     "1.0.0-1.32768.0",
							},
						},
					},
				})
				Expect(err).To(BeNil())
				Expect(respFind.Response.Code).To(Equal(scerr.ErrInvalidParams))
				respFind, err = instanceResource.BatchFind(getContext(), &pb.BatchFindInstancesRequest{
					ConsumerServiceId: serviceId1,
					Services: []*pb.FindService{
						{
							Service: &pb.MicroServiceKey{
								AppId:       "query_instance",
								ServiceName: "query_instance_service",
								Version:     " ",
							},
						},
					},
				})
				Expect(err).To(BeNil())
				Expect(respFind.Response.Code).To(Equal(scerr.ErrInvalidParams))
				respFind, err = instanceResource.BatchFind(getContext(), &pb.BatchFindInstancesRequest{
					ConsumerServiceId: serviceId1,
					Services: []*pb.FindService{
						{
							Service: &pb.MicroServiceKey{
								AppId:       "query_instance",
								ServiceName: "query_instance_service",
								Version:     "",
							},
						},
					},
				})
				Expect(err).To(BeNil())
				Expect(respFind.Response.Code).To(Equal(scerr.ErrInvalidParams))

				By("invalid instance")
				respFind, err = instanceResource.BatchFind(getContext(), &pb.BatchFindInstancesRequest{
					ConsumerServiceId: serviceId1,
					Instances: []*pb.FindInstance{
						{
							Instance: &pb.HeartbeatSetElement{
								ServiceId: "query_instance",
							},
						},
					},
				})
				Expect(err).To(BeNil())
				Expect(respFind.Response.Code).To(Equal(scerr.ErrInvalidParams))
				respFind, err = instanceResource.BatchFind(getContext(), &pb.BatchFindInstancesRequest{
					ConsumerServiceId: serviceId1,
					Instances: []*pb.FindInstance{
						{
							Instance: &pb.HeartbeatSetElement{
								InstanceId: "query_instance",
							},
						},
					},
				})
				Expect(err).To(BeNil())
				Expect(respFind.Response.Code).To(Equal(scerr.ErrInvalidParams))

				By("consumerId is empty")
				respFind, err = instanceResource.BatchFind(getContext(), &pb.BatchFindInstancesRequest{
					ConsumerServiceId: serviceId1,
					Services: []*pb.FindService{
						{
							Service: &pb.MicroServiceKey{
								AppId:       "query_instance",
								ServiceName: "query_instance_service",
								Version:     "1.0.0+",
							},
						},
					},
				})
				Expect(err).To(BeNil())
				Expect(respFind.Response.Code).To(Equal(pb.Response_SUCCESS))

				By("provider does not exist")
				respFind, err = instanceResource.BatchFind(getContext(), &pb.BatchFindInstancesRequest{
					ConsumerServiceId: serviceId1,
					Services: []*pb.FindService{
						{
							Service: &pb.MicroServiceKey{
								AppId:       "query_instance",
								ServiceName: "noneservice",
								Version:     "latest",
							},
						},
					},
				})
				Expect(err).To(BeNil())
				Expect(respFind.Response.Code).To(Equal(pb.Response_SUCCESS))
				Expect(respFind.Services.Failed[0].Error.Code).To(Equal(scerr.ErrServiceNotExists))
				Expect(respFind.Services.Failed[0].Indexes[0]).To(Equal(int64(0)))
				respFind, err = instanceResource.BatchFind(getContext(), &pb.BatchFindInstancesRequest{
					ConsumerServiceId: serviceId1,
					Instances: []*pb.FindInstance{
						{
							Instance: &pb.HeartbeatSetElement{
								ServiceId:  serviceId1,
								InstanceId: "noninstance",
							},
						},
					},
				})
				Expect(err).To(BeNil())
				Expect(respFind.Response.Code).To(Equal(pb.Response_SUCCESS))
				Expect(respFind.Instances.Failed[0].Error.Code).To(Equal(scerr.ErrInstanceNotExists))
				Expect(respFind.Instances.Failed[0].Indexes[0]).To(Equal(int64(0)))

				By("provider does not contain 3.0.0+ versions")
				respFind, err = instanceResource.BatchFind(getContext(), &pb.BatchFindInstancesRequest{
					ConsumerServiceId: serviceId1,
					Services: []*pb.FindService{
						{
							Service: &pb.MicroServiceKey{
								AppId:       "query_instance",
								ServiceName: "query_instance_service",
								Version:     "3.0.0+",
							},
						},
					},
				})
				Expect(err).To(BeNil())
				Expect(respFind.Response.Code).To(Equal(pb.Response_SUCCESS))
				Expect(len(respFind.Services.Updated[0].Instances)).To(Equal(0))
				Expect(respFind.Services.Updated[0].Index).To(Equal(int64(0)))
				Expect(respFind.Services.Updated[0].Rev).ToNot(Equal(""))

				By("consumer does not exist")
				respFind, err = instanceResource.BatchFind(getContext(), &pb.BatchFindInstancesRequest{
					ConsumerServiceId: "notExistServiceId",
					Services: []*pb.FindService{
						{
							Service: &pb.MicroServiceKey{
								AppId:       "query_instance",
								ServiceName: "query_instance_service",
								Version:     "2.0.0",
							},
						},
					},
				})
				Expect(err).To(BeNil())
				Expect(respFind.Response.Code).To(Equal(pb.Response_SUCCESS))
				Expect(respFind.Services.Failed[0].Indexes[0]).To(Equal(int64(0)))
				Expect(respFind.Services.Failed[0].Error.Code).To(Equal(scerr.ErrServiceNotExists))
			})
		})

		Context("when query instances", func() {
			It("should be passed", func() {
				By("find with version rule")
				respFind, err := instanceResource.Find(getContext(), &pb.FindInstancesRequest{
					ConsumerServiceId: serviceId1,
					AppId:             "query_instance",
					ServiceName:       "query_instance_service",
					VersionRule:       "latest",
				})
				Expect(err).To(BeNil())
				Expect(respFind.Response.Code).To(Equal(pb.Response_SUCCESS))
				Expect(respFind.Instances[0].InstanceId).To(Equal(instanceId2))

				respFind, err = instanceResource.Find(getContext(), &pb.FindInstancesRequest{
					ConsumerServiceId: serviceId1,
					AppId:             "query_instance",
					ServiceName:       "query_instance_service",
					VersionRule:       "1.0.0+",
					Tags:              []string{},
				})
				Expect(err).To(BeNil())
				Expect(respFind.Response.Code).To(Equal(pb.Response_SUCCESS))
				Expect(respFind.Instances[0].InstanceId).To(Equal(instanceId2))

				respFind, err = instanceResource.Find(getContext(), &pb.FindInstancesRequest{
					ConsumerServiceId: serviceId1,
					AppId:             "query_instance",
					ServiceName:       "query_instance_service",
					VersionRule:       "1.0.0-1.0.1",
				})
				Expect(err).To(BeNil())
				Expect(respFind.Response.Code).To(Equal(pb.Response_SUCCESS))
				Expect(respFind.Instances[0].InstanceId).To(Equal(instanceId1))

				respFind, err = instanceResource.Find(getContext(), &pb.FindInstancesRequest{
					ConsumerServiceId: serviceId1,
					AppId:             "query_instance",
					ServiceName:       "query_instance_service",
					VersionRule:       "1.0.0",
				})
				Expect(err).To(BeNil())
				Expect(respFind.Response.Code).To(Equal(pb.Response_SUCCESS))
				Expect(respFind.Instances[0].InstanceId).To(Equal(instanceId1))

				respFind, err = instanceResource.Find(getContext(), &pb.FindInstancesRequest{
					ConsumerServiceId: serviceId1,
					AppId:             "query_instance",
					ServiceName:       "query_instance_service",
					VersionRule:       "0.0.0",
				})
				Expect(err).To(BeNil())
				Expect(respFind.Response.Code).To(Equal(scerr.ErrServiceNotExists))

				By("find with env")
				respFind, err = instanceResource.Find(getContext(), &pb.FindInstancesRequest{
					ConsumerServiceId: serviceId4,
					AppId:             "query_instance",
					ServiceName:       "query_instance_diff_env_service",
					VersionRule:       "1.0.0",
				})
				Expect(err).To(BeNil())
				Expect(respFind.Response.Code).To(Equal(pb.Response_SUCCESS))
				Expect(len(respFind.Instances)).To(Equal(1))
				Expect(respFind.Instances[0].InstanceId).To(Equal(instanceId4))

				respFind, err = instanceResource.Find(getContext(), &pb.FindInstancesRequest{
					Environment: pb.ENV_PROD,
					AppId:       "query_instance",
					ServiceName: "query_instance_diff_env_service",
					VersionRule: "1.0.0",
				})
				Expect(err).To(BeNil())
				Expect(respFind.Response.Code).To(Equal(pb.Response_SUCCESS))
				Expect(len(respFind.Instances)).To(Equal(1))
				Expect(respFind.Instances[0].InstanceId).To(Equal(instanceId4))

				By("find with rev")
				ctx := util.SetContext(getContext(), serviceUtil.CTX_NOCACHE, "")
				respFind, err = instanceResource.Find(ctx, &pb.FindInstancesRequest{
					ConsumerServiceId: serviceId8,
					AppId:             "query_instance",
					ServiceName:       "query_instance_with_rev",
					VersionRule:       "1.0.0",
				})
				Expect(err).To(BeNil())
				Expect(respFind.Response.Code).To(Equal(pb.Response_SUCCESS))
				rev, _ := ctx.Value(serviceUtil.CTX_RESPONSE_REVISION).(string)
				Expect(respFind.Instances[0].InstanceId).To(Equal(instanceId8))
				Expect(len(rev)).NotTo(Equal(0))

				util.SetContext(ctx, serviceUtil.CTX_REQUEST_REVISION, "x")
				respFind, err = instanceResource.Find(ctx, &pb.FindInstancesRequest{
					ConsumerServiceId: serviceId8,
					AppId:             "query_instance",
					ServiceName:       "query_instance_with_rev",
					VersionRule:       "1.0.0",
				})
				Expect(err).To(BeNil())
				Expect(respFind.Response.Code).To(Equal(pb.Response_SUCCESS))
				Expect(respFind.Instances[0].InstanceId).To(Equal(instanceId8))
				Expect(ctx.Value(serviceUtil.CTX_RESPONSE_REVISION)).To(Equal(rev))

				util.SetContext(ctx, serviceUtil.CTX_REQUEST_REVISION, rev)
				respFind, err = instanceResource.Find(ctx, &pb.FindInstancesRequest{
					ConsumerServiceId: serviceId8,
					AppId:             "query_instance",
					ServiceName:       "query_instance_with_rev",
					VersionRule:       "1.0.0",
				})
				Expect(err).To(BeNil())
				Expect(respFind.Response.Code).To(Equal(pb.Response_SUCCESS))
				Expect(len(respFind.Instances)).To(Equal(0))
				Expect(ctx.Value(serviceUtil.CTX_RESPONSE_REVISION)).To(Equal(rev))

				By("find should return 200 even if consumer is diff apps")
				respFind, err = instanceResource.Find(getContext(), &pb.FindInstancesRequest{
					ConsumerServiceId: serviceId3,
					AppId:             "query_instance",
					ServiceName:       "query_instance_service",
					VersionRule:       "1.0.5",
				})
				Expect(err).To(BeNil())
				Expect(respFind.Response.Code).To(Equal(pb.Response_SUCCESS))
				Expect(len(respFind.Instances)).To(Equal(0))

				By("provider tag does not exist")
				respFind, err = instanceResource.Find(getContext(), &pb.FindInstancesRequest{
					ConsumerServiceId: serviceId1,
					AppId:             "query_instance",
					ServiceName:       "query_instance_service",
					VersionRule:       "latest",
					Tags:              []string{"notexisttag"},
				})
				Expect(err).To(BeNil())
				Expect(respFind.Response.Code).To(Equal(pb.Response_SUCCESS))
				Expect(len(respFind.Instances)).To(Equal(0))

				By("shared service discovery")
				os.Setenv("CSE_SHARED_SERVICES", "query_instance_shared_provider")
				core.SetSharedMode()
				core.Service.Environment = pb.ENV_PROD

				respFind, err = instanceResource.Find(
					util.SetTargetDomainProject(
						util.SetDomainProject(util.CloneContext(getContext()), "user", "user"),
						"default", "default"),
					&pb.FindInstancesRequest{
						ConsumerServiceId: serviceId6,
						AppId:             "default",
						ServiceName:       "query_instance_shared_provider",
						VersionRule:       "1.0.0",
					})
				Expect(err).To(BeNil())
				Expect(respFind.Response.Code).To(Equal(pb.Response_SUCCESS))
				Expect(len(respFind.Instances)).To(Equal(1))
				Expect(respFind.Instances[0].InstanceId).To(Equal(instanceId5))

				respFind, err = instanceResource.Find(getContext(), &pb.FindInstancesRequest{
					ConsumerServiceId: serviceId7,
					AppId:             "default",
					ServiceName:       "query_instance_shared_provider",
					VersionRule:       "1.0.0",
				})
				Expect(err).To(BeNil())
				Expect(respFind.Response.Code).To(Equal(pb.Response_SUCCESS))
				Expect(len(respFind.Instances)).To(Equal(1))
				Expect(respFind.Instances[0].InstanceId).To(Equal(instanceId5))

				By("get same domain deps")
				respGetC, err := serviceResource.GetConsumerDependencies(
					util.SetDomainProject(util.CloneContext(getContext()), "user", "user"),
					&pb.GetDependenciesRequest{
						ServiceId:  serviceId6,
						SameDomain: true,
					})
				Expect(err).To(BeNil())
				Expect(respGetC.Response.Code).To(Equal(pb.Response_SUCCESS))
				Expect(len(respGetC.Providers)).To(Equal(0))

				core.Service.Environment = pb.ENV_DEV
			})
		})

		Context("when batch query instances", func() {
			It("should be passed", func() {
				By("find with version rule")
				respFind, err := instanceResource.BatchFind(getContext(), &pb.BatchFindInstancesRequest{
					ConsumerServiceId: serviceId1,
					Services: []*pb.FindService{
						{
							Service: &pb.MicroServiceKey{
								AppId:       "query_instance",
								ServiceName: "query_instance_service",
								Version:     "latest",
							},
						},
						{
							Service: &pb.MicroServiceKey{
								AppId:       "query_instance",
								ServiceName: "query_instance_service",
								Version:     "1.0.0+",
							},
						},
						{
							Service: &pb.MicroServiceKey{
								AppId:       "query_instance",
								ServiceName: "query_instance_service",
								Version:     "0.0.0",
							},
						},
					},
				})
				Expect(err).To(BeNil())
				Expect(respFind.Response.Code).To(Equal(pb.Response_SUCCESS))
				Expect(respFind.Services.Updated[0].Index).To(Equal(int64(0)))
				Expect(respFind.Services.Updated[0].Instances[0].InstanceId).To(Equal(instanceId2))
				Expect(respFind.Services.Updated[1].Index).To(Equal(int64(1)))
				Expect(respFind.Services.Updated[1].Instances[0].InstanceId).To(Equal(instanceId2))
				Expect(respFind.Services.Failed[0].Indexes[0]).To(Equal(int64(2)))
				Expect(respFind.Services.Failed[0].Error.Code).To(Equal(scerr.ErrServiceNotExists))

				By("find with env")
				respFind, err = instanceResource.BatchFind(getContext(), &pb.BatchFindInstancesRequest{
					ConsumerServiceId: serviceId4,
					Services: []*pb.FindService{
						{
							Service: &pb.MicroServiceKey{
								AppId:       "query_instance",
								ServiceName: "query_instance_diff_env_service",
								Version:     "1.0.0",
							},
						},
					},
				})
				Expect(err).To(BeNil())
				Expect(respFind.Response.Code).To(Equal(pb.Response_SUCCESS))
				Expect(len(respFind.Services.Updated[0].Instances)).To(Equal(1))
				Expect(respFind.Services.Updated[0].Instances[0].InstanceId).To(Equal(instanceId4))

				respFind, err = instanceResource.BatchFind(getContext(), &pb.BatchFindInstancesRequest{
					Services: []*pb.FindService{
						{
							Service: &pb.MicroServiceKey{
								Environment: pb.ENV_PROD,
								AppId:       "query_instance",
								ServiceName: "query_instance_diff_env_service",
								Version:     "1.0.0",
							},
						},
					},
				})
				Expect(err).To(BeNil())
				Expect(respFind.Response.Code).To(Equal(pb.Response_SUCCESS))
				Expect(len(respFind.Services.Updated[0].Instances)).To(Equal(1))
				Expect(respFind.Services.Updated[0].Instances[0].InstanceId).To(Equal(instanceId4))

				By("find with rev")
				ctx := util.SetContext(getContext(), serviceUtil.CTX_NOCACHE, "")
				respFind, err = instanceResource.BatchFind(ctx, &pb.BatchFindInstancesRequest{
					ConsumerServiceId: serviceId8,
					Services: []*pb.FindService{
						{
							Service: &pb.MicroServiceKey{
								AppId:       "query_instance",
								ServiceName: "query_instance_with_rev",
								Version:     "1.0.0",
							},
						},
						{
							Service: &pb.MicroServiceKey{
								AppId:       "query_instance",
								ServiceName: "batch_query_instance_with_rev",
								Version:     "1.0.0",
							},
						},
					},
					Instances: []*pb.FindInstance{
						{
							Instance: &pb.HeartbeatSetElement{
								ServiceId:  serviceId9,
								InstanceId: instanceId9,
							},
						},
						{
							Instance: &pb.HeartbeatSetElement{
								ServiceId:  serviceId8,
								InstanceId: instanceId8,
							},
						},
					},
				})
				Expect(err).To(BeNil())
				Expect(respFind.Response.Code).To(Equal(pb.Response_SUCCESS))
				rev := respFind.Services.Updated[0].Rev
				Expect(respFind.Services.Updated[0].Index).To(Equal(int64(0)))
				Expect(respFind.Services.Updated[1].Index).To(Equal(int64(1)))
				Expect(respFind.Services.Updated[0].Instances[0].InstanceId).To(Equal(instanceId8))
				Expect(respFind.Services.Updated[1].Instances[0].InstanceId).To(Equal(instanceId9))
				Expect(len(rev)).NotTo(Equal(0))
				instanceRev := respFind.Instances.Updated[0].Rev
				Expect(respFind.Instances.Updated[0].Index).To(Equal(int64(0)))
				Expect(respFind.Instances.Updated[1].Index).To(Equal(int64(1)))
				Expect(respFind.Instances.Updated[0].Instances[0].InstanceId).To(Equal(instanceId9))
				Expect(respFind.Instances.Updated[1].Instances[0].InstanceId).To(Equal(instanceId8))
				Expect(len(instanceRev)).NotTo(Equal(0))

				respFind, err = instanceResource.BatchFind(ctx, &pb.BatchFindInstancesRequest{
					ConsumerServiceId: serviceId8,
					Services: []*pb.FindService{
						{
							Service: &pb.MicroServiceKey{
								AppId:       "query_instance",
								ServiceName: "query_instance_with_rev",
								Version:     "1.0.0",
							},
							Rev: "x",
						},
					},
					Instances: []*pb.FindInstance{
						{
							Instance: &pb.HeartbeatSetElement{
								ServiceId:  serviceId9,
								InstanceId: instanceId9,
							},
							Rev: "x",
						},
					},
				})
				Expect(err).To(BeNil())
				Expect(respFind.Response.Code).To(Equal(pb.Response_SUCCESS))
				Expect(respFind.Services.Updated[0].Instances[0].InstanceId).To(Equal(instanceId8))
				Expect(respFind.Services.Updated[0].Rev).To(Equal(rev))
				Expect(respFind.Instances.Updated[0].Instances[0].InstanceId).To(Equal(instanceId9))
				Expect(respFind.Instances.Updated[0].Rev).To(Equal(instanceRev))

				respFind, err = instanceResource.BatchFind(ctx, &pb.BatchFindInstancesRequest{
					ConsumerServiceId: serviceId8,
					Services: []*pb.FindService{
						{
							Service: &pb.MicroServiceKey{
								AppId:       "query_instance",
								ServiceName: "query_instance_with_rev",
								Version:     "1.0.0",
							},
							Rev: rev,
						},
					},
					Instances: []*pb.FindInstance{
						{
							Instance: &pb.HeartbeatSetElement{
								ServiceId:  serviceId9,
								InstanceId: instanceId9,
							},
							Rev: instanceRev,
						},
					},
				})
				Expect(err).To(BeNil())
				Expect(respFind.Response.Code).To(Equal(pb.Response_SUCCESS))
				Expect(respFind.Services.NotModified[0]).To(Equal(int64(0)))
				Expect(respFind.Instances.NotModified[0]).To(Equal(int64(0)))

				By("find should return 200 even if consumer is diff apps")
				respFind, err = instanceResource.BatchFind(getContext(), &pb.BatchFindInstancesRequest{
					ConsumerServiceId: serviceId3,
					Services: []*pb.FindService{
						{
							Service: &pb.MicroServiceKey{
								AppId:       "query_instance",
								ServiceName: "query_instance_service",
								Version:     "1.0.5",
							},
						},
					},
				})
				Expect(err).To(BeNil())
				Expect(respFind.Response.Code).To(Equal(pb.Response_SUCCESS))
				Expect(len(respFind.Services.Updated[0].Instances)).To(Equal(0))

				By("shared service discovery")
				os.Setenv("CSE_SHARED_SERVICES", "query_instance_shared_provider")
				core.SetSharedMode()
				core.Service.Environment = pb.ENV_PROD

				respFind, err = instanceResource.BatchFind(
					util.SetTargetDomainProject(
						util.SetDomainProject(util.CloneContext(getContext()), "user", "user"),
						"default", "default"),
					&pb.BatchFindInstancesRequest{
						ConsumerServiceId: serviceId6,
						Services: []*pb.FindService{
							{
								Service: &pb.MicroServiceKey{
									AppId:       "default",
									ServiceName: "query_instance_shared_provider",
									Version:     "1.0.0",
								},
							},
						},
					})
				Expect(err).To(BeNil())
				Expect(respFind.Response.Code).To(Equal(pb.Response_SUCCESS))
				Expect(len(respFind.Services.Updated[0].Instances)).To(Equal(1))
				Expect(respFind.Services.Updated[0].Instances[0].InstanceId).To(Equal(instanceId5))

				respFind, err = instanceResource.BatchFind(getContext(), &pb.BatchFindInstancesRequest{
					ConsumerServiceId: serviceId7,
					Services: []*pb.FindService{
						{
							Service: &pb.MicroServiceKey{
								AppId:       "default",
								ServiceName: "query_instance_shared_provider",
								Version:     "1.0.0",
							},
						},
					},
				})
				Expect(err).To(BeNil())
				Expect(respFind.Response.Code).To(Equal(pb.Response_SUCCESS))
				Expect(len(respFind.Services.Updated[0].Instances)).To(Equal(1))
				Expect(respFind.Services.Updated[0].Instances[0].InstanceId).To(Equal(instanceId5))

				core.Service.Environment = pb.ENV_DEV
			})
		})

		Context("when query instances between diff dimensions", func() {
			It("should be failed", func() {
				By("diff appId")
				UTFunc := func(consumerId string, code int32) {
					respFind, err := instanceResource.GetInstances(getContext(), &pb.GetInstancesRequest{
						ConsumerServiceId: consumerId,
						ProviderServiceId: serviceId2,
					})
					Expect(err).To(BeNil())
					Expect(respFind.Response.Code).To(Equal(code))
				}

				UTFunc(serviceId3, scerr.ErrPermissionDeny)

				UTFunc(serviceId1, pb.Response_SUCCESS)

				By("diff env")
				respFind, err := instanceResource.GetInstances(getContext(), &pb.GetInstancesRequest{
					ConsumerServiceId: serviceId4,
					ProviderServiceId: serviceId2,
				})
				Expect(err).To(BeNil())
				Expect(respFind.Response.Code).ToNot(Equal(pb.Response_SUCCESS))
			})
		})
	})

	Describe("execute 'get one' operartion", func() {
		var (
			serviceId1  string
			serviceId2  string
			serviceId3  string
			instanceId2 string
		)

		It("should be passed", func() {
			respCreate, err := serviceResource.Create(getContext(), &pb.CreateServiceRequest{
				Service: &pb.MicroService{
					AppId:       "get_instance",
					ServiceName: "get_instance_service",
					Version:     "1.0.0",
					Level:       "FRONT",
					Status:      pb.MS_UP,
				},
			})
			Expect(err).To(BeNil())
			Expect(respCreate.Response.Code).To(Equal(pb.Response_SUCCESS))
			serviceId1 = respCreate.ServiceId

			respCreate, err = serviceResource.Create(getContext(), &pb.CreateServiceRequest{
				Service: &pb.MicroService{
					AppId:       "get_instance",
					ServiceName: "get_instance_service",
					Version:     "1.0.5",
					Level:       "FRONT",
					Status:      pb.MS_UP,
				},
				Tags: map[string]string{
					"test": "test",
				},
			})
			Expect(err).To(BeNil())
			Expect(respCreate.Response.Code).To(Equal(pb.Response_SUCCESS))
			serviceId2 = respCreate.ServiceId

			resp, err := instanceResource.Register(getContext(), &pb.RegisterInstanceRequest{
				Instance: &pb.MicroServiceInstance{
					ServiceId: serviceId2,
					HostName:  "UT-HOST",
					Endpoints: []string{
						"get:127.0.0.2:8080",
					},
					Status: pb.MSI_UP,
				},
			})
			Expect(err).To(BeNil())
			Expect(resp.Response.Code).To(Equal(pb.Response_SUCCESS))
			instanceId2 = resp.InstanceId

			respCreate, err = serviceResource.Create(getContext(), &pb.CreateServiceRequest{
				Service: &pb.MicroService{
					AppId:       "get_instance_cross",
					ServiceName: "get_instance_service",
					Version:     "1.0.0",
					Level:       "FRONT",
					Status:      pb.MS_UP,
				},
			})
			Expect(err).To(BeNil())
			serviceId3 = respCreate.ServiceId
		})

		Context("when get one instance request is invalid", func() {
			It("should be failed", func() {
				By("find service itself")
				resp, err := instanceResource.GetOneInstance(getContext(), &pb.GetOneInstanceRequest{
					ConsumerServiceId:  serviceId2,
					ProviderServiceId:  serviceId2,
					ProviderInstanceId: instanceId2,
				})
				Expect(err).To(BeNil())
				Expect(resp.Response.Code).To(Equal(pb.Response_SUCCESS))

				By("provider id is empty")
				resp, err = instanceResource.GetOneInstance(getContext(), &pb.GetOneInstanceRequest{
					ConsumerServiceId:  serviceId1,
					ProviderServiceId:  "",
					ProviderInstanceId: instanceId2,
				})
				Expect(err).To(BeNil())
				Expect(resp.Response.Code).ToNot(Equal(pb.Response_SUCCESS))

				By("provider instance id is empty")
				resp, err = instanceResource.GetOneInstance(getContext(), &pb.GetOneInstanceRequest{
					ConsumerServiceId:  serviceId1,
					ProviderServiceId:  serviceId2,
					ProviderInstanceId: "",
				})
				Expect(err).To(BeNil())
				Expect(resp.Response.Code).ToNot(Equal(pb.Response_SUCCESS))

				By("consumer id is empty")
				resp, err = instanceResource.GetOneInstance(getContext(), &pb.GetOneInstanceRequest{
					ConsumerServiceId:  "",
					ProviderServiceId:  serviceId2,
					ProviderInstanceId: instanceId2,
				})
				Expect(err).To(BeNil())
				Expect(resp.Response.Code).To(Equal(pb.Response_SUCCESS))

				By("consumer does not exist")
				resp, err = instanceResource.GetOneInstance(getContext(), &pb.GetOneInstanceRequest{
					ConsumerServiceId:  "not-exist-id",
					ProviderServiceId:  serviceId2,
					ProviderInstanceId: instanceId2,
				})
				Expect(err).To(BeNil())
				Expect(resp.Response.Code).ToNot(Equal(pb.Response_SUCCESS))

				By("provider tag does not exist")
				resp, err = instanceResource.GetOneInstance(getContext(), &pb.GetOneInstanceRequest{
					ConsumerServiceId:  serviceId1,
					ProviderServiceId:  serviceId2,
					ProviderInstanceId: instanceId2,
					Tags:               []string{"not-exist-tag"},
				})
				Expect(err).To(BeNil())
				Expect(resp.Response.Code).To(Equal(scerr.ErrInstanceNotExists))

				By("provider tag exist")
				resp, err = instanceResource.GetOneInstance(getContext(),
					&pb.GetOneInstanceRequest{
						ConsumerServiceId:  serviceId1,
						ProviderServiceId:  serviceId2,
						ProviderInstanceId: instanceId2,
						Tags:               []string{"test"},
					})
				Expect(err).To(BeNil())
				Expect(resp.Response.Code).To(Equal(pb.Response_SUCCESS))
			})
		})

		Context("when get between diff apps", func() {
			It("should be failed", func() {
				resp, err := instanceResource.GetOneInstance(getContext(), &pb.GetOneInstanceRequest{
					ConsumerServiceId:  serviceId3,
					ProviderServiceId:  serviceId2,
					ProviderInstanceId: instanceId2,
				})
				Expect(err).To(BeNil())
				Expect(resp.Response.Code).To(Equal(scerr.ErrInstanceNotExists))

				respAll, err := instanceResource.GetInstances(getContext(), &pb.GetInstancesRequest{
					ConsumerServiceId: serviceId3,
					ProviderServiceId: serviceId2,
				})
				Expect(err).To(BeNil())
				Expect(respAll.Response.Code).ToNot(Equal(pb.Response_SUCCESS))
			})
		})

		Context("when get instances request is invalid", func() {
			It("should be failed", func() {
				By("consumer id is empty")
				resp, err := instanceResource.GetInstances(getContext(), &pb.GetInstancesRequest{
					ConsumerServiceId: "",
					ProviderServiceId: serviceId2,
				})
				Expect(err).To(BeNil())
				Expect(resp.Response.Code).To(Equal(pb.Response_SUCCESS))

				By("consumer does not exist")
				resp, err = instanceResource.GetInstances(getContext(), &pb.GetInstancesRequest{
					ConsumerServiceId: "noneservice",
					ProviderServiceId: serviceId2,
				})
				Expect(err).To(BeNil())
				Expect(resp.Response.Code).ToNot(Equal(pb.Response_SUCCESS))

				By("valid request")
				resp, err = instanceResource.GetInstances(getContext(), &pb.GetInstancesRequest{
					ConsumerServiceId: serviceId1,
					ProviderServiceId: serviceId2,
				})
				Expect(err).To(BeNil())
				Expect(resp.Response.Code).To(Equal(pb.Response_SUCCESS))
			})
		})
	})

	Describe("execute 'unregister' operartion", func() {
		var (
			serviceId  string
			instanceId string
		)

		It("should be passed", func() {
			respCreate, err := serviceResource.Create(getContext(), &pb.CreateServiceRequest{
				Service: &pb.MicroService{
					AppId:       "unregister_instance",
					ServiceName: "unregister_instance_service",
					Version:     "1.0.5",
					Level:       "FRONT",
					Status:      pb.MS_UP,
				},
				Tags: map[string]string{
					"test": "test",
				},
			})
			Expect(err).To(BeNil())
			Expect(respCreate.Response.Code).To(Equal(pb.Response_SUCCESS))
			serviceId = respCreate.ServiceId

			resp, err := instanceResource.Register(getContext(), &pb.RegisterInstanceRequest{
				Instance: &pb.MicroServiceInstance{
					ServiceId: serviceId,
					HostName:  "UT-HOST",
					Endpoints: []string{
						"unregister:127.0.0.2:8080",
					},
					Status: pb.MSI_UP,
				},
			})
			Expect(err).To(BeNil())
			Expect(resp.Response.Code).To(Equal(pb.Response_SUCCESS))
			instanceId = resp.InstanceId
		})

		Context("when request is valid", func() {
			It("should be passed", func() {
				resp, err := instanceResource.Unregister(getContext(), &pb.UnregisterInstanceRequest{
					ServiceId:  serviceId,
					InstanceId: instanceId,
				})
				Expect(err).To(BeNil())
				Expect(resp.Response.Code).To(Equal(pb.Response_SUCCESS))
			})
		})

		Context("when request is invalid", func() {
			It("should be failed", func() {
				By("service id is invalid")
				resp, err := instanceResource.Unregister(getContext(), &pb.UnregisterInstanceRequest{
					ServiceId:  "",
					InstanceId: instanceId,
				})
				Expect(err).To(BeNil())
				Expect(resp.Response.Code).To(Equal(scerr.ErrInvalidParams))
				resp, err = instanceResource.Unregister(getContext(), &pb.UnregisterInstanceRequest{
					ServiceId:  TOO_LONG_SERVICEID,
					InstanceId: instanceId,
				})
				Expect(err).To(BeNil())
				Expect(resp.Response.Code).To(Equal(scerr.ErrInvalidParams))

				By("service does not exist")
				resp, err = instanceResource.Unregister(getContext(), &pb.UnregisterInstanceRequest{
					ServiceId:  "not-exist-id",
					InstanceId: instanceId,
				})
				Expect(err).To(BeNil())
				Expect(resp.Response.Code).ToNot(Equal(pb.Response_SUCCESS))

				By("instance is invalid")
				resp, err = instanceResource.Unregister(getContext(), &pb.UnregisterInstanceRequest{
					ServiceId:  serviceId,
					InstanceId: "",
				})
				Expect(err).To(BeNil())
				Expect(resp.Response.Code).To(Equal(scerr.ErrInvalidParams))
				resp, err = instanceResource.Unregister(getContext(), &pb.UnregisterInstanceRequest{
					ServiceId:  serviceId,
					InstanceId: "@",
				})
				Expect(err).To(BeNil())
				Expect(resp.Response.Code).To(Equal(scerr.ErrInvalidParams))
				resp, err = instanceResource.Unregister(getContext(), &pb.UnregisterInstanceRequest{
					ServiceId:  serviceId,
					InstanceId: TOO_LONG_SERVICEID,
				})
				Expect(err).To(BeNil())
				Expect(resp.Response.Code).To(Equal(scerr.ErrInvalidParams))

				By("instance does not exist")
				resp, err = instanceResource.Unregister(getContext(), &pb.UnregisterInstanceRequest{
					ServiceId:  serviceId,
					InstanceId: "not-exist-id",
				})
				Expect(err).To(BeNil())
				Expect(resp.Response.Code).ToNot(Equal(pb.Response_SUCCESS))
			})
		})
	})
})
