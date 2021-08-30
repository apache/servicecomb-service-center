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
	"math"
	"strconv"
	"strings"

	discosvc "github.com/apache/servicecomb-service-center/server/service/disco"

	"github.com/apache/servicecomb-service-center/pkg/util"
	"github.com/apache/servicecomb-service-center/server/config"
	"github.com/apache/servicecomb-service-center/server/core"
	pb "github.com/go-chassis/cari/discovery"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var (
	TOO_LONG_HOSTNAME = strings.Repeat("x", 65)
	TOO_LONG_URL      = strings.Repeat("x", 513)
)

var _ = Describe("'Instance' service", func() {
	Describe("execute 'register' operartion", func() {
		var (
			serviceId1 string
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
			Expect(respCreate.Response.GetCode()).To(Equal(pb.ResponseSuccess))
			serviceId1 = respCreate.ServiceId
		})

		Context("when register a instance", func() {
			It("should be passed", func() {
				resp, err := discosvc.RegisterInstance(getContext(), &pb.RegisterInstanceRequest{
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
				Expect(resp.Response.GetCode()).To(Equal(pb.ResponseSuccess))
				Expect(resp.InstanceId).To(Not(Equal("")))

				resp, err = discosvc.RegisterInstance(getContext(), &pb.RegisterInstanceRequest{
					Instance: &pb.MicroServiceInstance{
						InstanceId: "customId",
						ServiceId:  serviceId1,
						Endpoints: []string{
							"createInstance:127.0.0.1:8080",
						},
						HostName: "UT-HOST",
						Status:   pb.MSI_UP,
					},
				})
				Expect(err).To(BeNil())
				Expect(resp.Response.GetCode()).To(Equal(pb.ResponseSuccess))
				Expect(resp.InstanceId).To(Equal("customId"))

				By("status is nil")
				resp, err = discosvc.RegisterInstance(getContext(), &pb.RegisterInstanceRequest{
					Instance: &pb.MicroServiceInstance{
						ServiceId: serviceId1,
						Endpoints: []string{
							"createInstance:127.0.0.1:8081",
						},
						HostName: "UT-HOST",
					},
				})
				Expect(err).To(BeNil())
				Expect(resp.Response.GetCode()).To(Equal(pb.ResponseSuccess))
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
				resp, err = discosvc.RegisterInstance(getContext(), &pb.RegisterInstanceRequest{
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
				Expect(resp.Response.GetCode()).To(Equal(pb.ResponseSuccess))
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
				resp, err := discosvc.RegisterInstance(getContext(), &pb.RegisterInstanceRequest{
					Instance: instance,
				})
				Expect(err).To(BeNil())
				Expect(resp.Response.GetCode()).To(Equal(pb.ResponseSuccess))

				resp, err = discosvc.RegisterInstance(getContext(), &pb.RegisterInstanceRequest{
					Instance: instance,
				})
				Expect(err).To(BeNil())
				Expect(resp.Response.GetCode()).To(Equal(pb.ResponseSuccess))
				Expect(resp.InstanceId).To(Equal(instance.InstanceId))
			})
		})

		Context("when register invalid instance", func() {
			It("should be failed", func() {
				By("endpoints are empty")
				resp, err := discosvc.RegisterInstance(getContext(), &pb.RegisterInstanceRequest{
					Instance: &pb.MicroServiceInstance{
						ServiceId: serviceId1,
						Endpoints: []string{""},
						HostName:  "UT-HOST",
						Status:    pb.MSI_UP,
					},
				})
				Expect(err).To(BeNil())
				Expect(resp.Response.GetCode()).To(Equal(pb.ErrInvalidParams))

				By("serviceId is empty")
				resp, err = discosvc.RegisterInstance(getContext(), &pb.RegisterInstanceRequest{
					Instance: &pb.MicroServiceInstance{
						HostName: "UT-HOST",
						Endpoints: []string{
							"check:127.0.0.1:8080",
						},
						Status: pb.MSI_UP,
					},
				})
				Expect(err).To(BeNil())
				Expect(resp.Response.GetCode()).ToNot(Equal(pb.ResponseSuccess))

				By("service does not exist")
				resp, err = discosvc.RegisterInstance(getContext(), &pb.RegisterInstanceRequest{
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
				Expect(resp.Response.GetCode()).ToNot(Equal(pb.ResponseSuccess))

				By("hostName is empty")
				resp, err = discosvc.RegisterInstance(getContext(), &pb.RegisterInstanceRequest{
					Instance: &pb.MicroServiceInstance{
						ServiceId: serviceId1,
						Endpoints: []string{
							"check:127.0.0.1:8080",
						},
						Status: pb.MSI_UP,
					},
				})
				Expect(err).To(BeNil())
				Expect(resp.Response.GetCode()).ToNot(Equal(pb.ResponseSuccess))
				resp, err = discosvc.RegisterInstance(getContext(), &pb.RegisterInstanceRequest{
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
				Expect(resp.Response.GetCode()).ToNot(Equal(pb.ResponseSuccess))
				resp, err = discosvc.RegisterInstance(getContext(), &pb.RegisterInstanceRequest{
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
				Expect(resp.Response.GetCode()).ToNot(Equal(pb.ResponseSuccess))

				By("instance is nil")
				resp, err = discosvc.RegisterInstance(getContext(), &pb.RegisterInstanceRequest{
					Instance: nil,
				})
				Expect(err).To(BeNil())
				Expect(resp.Response.GetCode()).ToNot(Equal(pb.ResponseSuccess))

				By("check normal push healthChceck")
				resp, err = discosvc.RegisterInstance(getContext(), &pb.RegisterInstanceRequest{
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
				Expect(resp.Response.GetCode()).To(Equal(pb.ResponseSuccess))

				resp, err = discosvc.RegisterInstance(getContext(), &pb.RegisterInstanceRequest{
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
				Expect(resp.Response.GetCode()).To(Equal(pb.ResponseSuccess))

				By("check normal pull healthChceck")
				resp, err = discosvc.RegisterInstance(getContext(), &pb.RegisterInstanceRequest{
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
				Expect(resp.Response.GetCode()).To(Equal(pb.ResponseSuccess))

				By("check invalid push healthChceck")
				resp, err = discosvc.RegisterInstance(getContext(), &pb.RegisterInstanceRequest{
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
				Expect(resp.Response.GetCode()).To(Equal(pb.ErrInvalidParams))

				By("check invalid pull healthChceck")
				resp, err = discosvc.RegisterInstance(getContext(), &pb.RegisterInstanceRequest{
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
				Expect(resp.Response.GetCode()).To(Equal(pb.ErrInvalidParams))
				resp, err = discosvc.RegisterInstance(getContext(), &pb.RegisterInstanceRequest{
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
				Expect(resp.Response.GetCode()).To(Equal(pb.ErrInvalidParams))
				resp, err = discosvc.RegisterInstance(getContext(), &pb.RegisterInstanceRequest{
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
				Expect(resp.Response.GetCode()).To(Equal(pb.ErrInvalidParams))

				By("invalid status")
				resp, err = discosvc.RegisterInstance(getContext(), &pb.RegisterInstanceRequest{
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
				Expect(resp.Response.GetCode()).ToNot(Equal(pb.ResponseSuccess))
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
			Expect(respCreate.Response.GetCode()).To(Equal(pb.ResponseSuccess))
			serviceId = respCreate.ServiceId

			resp, err := discosvc.RegisterInstance(getContext(), &pb.RegisterInstanceRequest{
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
			Expect(resp.Response.GetCode()).To(Equal(pb.ResponseSuccess))
			instanceId1 = resp.InstanceId

			resp, err = discosvc.RegisterInstance(getContext(), &pb.RegisterInstanceRequest{
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
			Expect(resp.Response.GetCode()).To(Equal(pb.ResponseSuccess))
			instanceId2 = resp.InstanceId
		})

		Context("when update a lease", func() {
			It("should be passed", func() {
				By("valid instance")
				resp, err := discosvc.Heartbeat(getContext(), &pb.HeartbeatRequest{
					ServiceId:  serviceId,
					InstanceId: instanceId1,
				})
				Expect(err).To(BeNil())
				Expect(resp.Response.GetCode()).To(Equal(pb.ResponseSuccess))

				By("serviceId/instanceId is invalid")
				resp, err = discosvc.Heartbeat(getContext(), &pb.HeartbeatRequest{
					ServiceId:  "",
					InstanceId: instanceId1,
				})
				Expect(err).To(BeNil())
				Expect(resp.Response.GetCode()).To(Equal(pb.ErrInvalidParams))
				resp, err = discosvc.Heartbeat(getContext(), &pb.HeartbeatRequest{
					ServiceId:  TOO_LONG_SERVICEID,
					InstanceId: instanceId1,
				})
				Expect(err).To(BeNil())
				Expect(resp.Response.GetCode()).To(Equal(pb.ErrInvalidParams))
				resp, err = discosvc.Heartbeat(getContext(), &pb.HeartbeatRequest{
					ServiceId:  serviceId,
					InstanceId: "",
				})
				Expect(err).To(BeNil())
				Expect(resp.Response.GetCode()).To(Equal(pb.ErrInvalidParams))
				resp, err = discosvc.Heartbeat(getContext(), &pb.HeartbeatRequest{
					ServiceId:  serviceId,
					InstanceId: "@",
				})
				Expect(err).To(BeNil())
				Expect(resp.Response.GetCode()).To(Equal(pb.ErrInvalidParams))
				resp, err = discosvc.Heartbeat(getContext(), &pb.HeartbeatRequest{
					ServiceId:  serviceId,
					InstanceId: TOO_LONG_SERVICEID,
				})
				Expect(err).To(BeNil())
				Expect(resp.Response.GetCode()).To(Equal(pb.ErrInvalidParams))

				By("serviceId does not exist")
				resp, err = discosvc.Heartbeat(getContext(), &pb.HeartbeatRequest{
					ServiceId:  "100000000000",
					InstanceId: instanceId1,
				})
				Expect(resp.Response.GetCode()).ToNot(Equal(pb.ResponseSuccess))

				By("instance does not exist")
				resp, err = discosvc.Heartbeat(getContext(), &pb.HeartbeatRequest{
					ServiceId:  serviceId,
					InstanceId: "not-exist-ins",
				})
				Expect(resp.Response.GetCode()).ToNot(Equal(pb.ResponseSuccess))
			})
		})

		Context("when batch update lease", func() {
			It("should be passed", func() {
				By("instances are empty")
				resp, err := discosvc.HeartbeatSet(getContext(), &pb.HeartbeatSetRequest{
					Instances: []*pb.HeartbeatSetElement{},
				})
				Expect(err).To(BeNil())
				Expect(resp.Response.GetCode()).ToNot(Equal(pb.ResponseSuccess))

				By("instances are nil")
				resp, err = discosvc.HeartbeatSet(getContext(), &pb.HeartbeatSetRequest{})
				Expect(err).To(BeNil())
				Expect(resp.Response.GetCode()).ToNot(Equal(pb.ResponseSuccess))

				By("request contains > 1 instances")
				resp, err = discosvc.HeartbeatSet(getContext(), &pb.HeartbeatSetRequest{
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
				Expect(resp.Response.GetCode()).To(Equal(pb.ResponseSuccess))

				By("request contains invalid instance")
				resp, err = discosvc.HeartbeatSet(getContext(), &pb.HeartbeatSetRequest{
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
				Expect(resp.Response.GetCode()).ToNot(Equal(pb.ResponseSuccess))
			})
		})
	})

	Describe("execute 'clusterHealth' operartion", func() {

		It("should be passed", func() {
			resp, err := serviceResource.Create(getContext(), core.CreateServiceRequest())
			Expect(err).To(BeNil())
			Expect(resp.Response.GetCode()).To(Equal(pb.ResponseSuccess))
		})

		Context("when SC does not exist", func() {
			It("should be failed", func() {
				old := core.Service.ServiceName
				core.Service.ServiceName = "x"
				respCluterhealth, err := discosvc.ClusterHealth(getContext())
				Expect(err).To(BeNil())
				Expect(respCluterhealth.Response.GetCode()).ToNot(Equal(pb.ResponseSuccess))
				core.Service.ServiceName = old
			})
		})

		Context("when SC registered", func() {
			It("should be passed", func() {
				respCluterhealth, err := discosvc.ClusterHealth(getContext())
				Expect(err).To(BeNil())
				Expect(respCluterhealth.Response.GetCode()).To(Equal(pb.ResponseSuccess))
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
			Expect(respCreate.Response.GetCode()).To(Equal(pb.ResponseSuccess))
			serviceId = respCreate.ServiceId

			resp, err := discosvc.RegisterInstance(getContext(), &pb.RegisterInstanceRequest{
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
			Expect(resp.Response.GetCode()).To(Equal(pb.ResponseSuccess))
			instanceId = resp.InstanceId
		})

		Context("when update instance status", func() {
			It("should be passed", func() {
				By("update instance status")
				respUpdateStatus, err := discosvc.UpdateInstanceStatus(getContext(), &pb.UpdateInstanceStatusRequest{
					ServiceId:  serviceId,
					InstanceId: instanceId,
					Status:     pb.MSI_DOWN,
				})
				Expect(err).To(BeNil())
				Expect(respUpdateStatus.Response.GetCode()).To(Equal(pb.ResponseSuccess))

				respUpdateStatus, err = discosvc.UpdateInstanceStatus(getContext(), &pb.UpdateInstanceStatusRequest{
					ServiceId:  serviceId,
					InstanceId: instanceId,
					Status:     pb.MSI_OUTOFSERVICE,
				})
				Expect(err).To(BeNil())
				Expect(respUpdateStatus.Response.GetCode()).To(Equal(pb.ResponseSuccess))

				respUpdateStatus, err = discosvc.UpdateInstanceStatus(getContext(), &pb.UpdateInstanceStatusRequest{
					ServiceId:  serviceId,
					InstanceId: instanceId,
					Status:     pb.MSI_STARTING,
				})
				Expect(err).To(BeNil())
				Expect(respUpdateStatus.Response.GetCode()).To(Equal(pb.ResponseSuccess))

				respUpdateStatus, err = discosvc.UpdateInstanceStatus(getContext(), &pb.UpdateInstanceStatusRequest{
					ServiceId:  serviceId,
					InstanceId: instanceId,
					Status:     pb.MSI_TESTING,
				})

				Expect(err).To(BeNil())
				Expect(respUpdateStatus.Response.GetCode()).To(Equal(pb.ResponseSuccess))

				respUpdateStatus, err = discosvc.UpdateInstanceStatus(getContext(), &pb.UpdateInstanceStatusRequest{
					ServiceId:  serviceId,
					InstanceId: instanceId,
					Status:     pb.MSI_UP,
				})
				Expect(err).To(BeNil())
				Expect(respUpdateStatus.Response.GetCode()).To(Equal(pb.ResponseSuccess))

				By("update status with a empty serviceId")
				respUpdateStatus, err = discosvc.UpdateInstanceStatus(getContext(), &pb.UpdateInstanceStatusRequest{
					ServiceId:  "",
					InstanceId: instanceId,
					Status:     pb.MSI_STARTING,
				})

				Expect(err).To(BeNil())
				Expect(respUpdateStatus.Response.GetCode()).ToNot(Equal(pb.ResponseSuccess))

				By("update status with a not exist service")
				respUpdateStatus, err = discosvc.UpdateInstanceStatus(getContext(), &pb.UpdateInstanceStatusRequest{
					ServiceId:  "notexistservice",
					InstanceId: instanceId,
					Status:     pb.MSI_STARTING,
				})

				Expect(err).To(BeNil())
				Expect(respUpdateStatus.Response.GetCode()).ToNot(Equal(pb.ResponseSuccess))

				By("update status with a not exist instance")
				respUpdateStatus, err = discosvc.UpdateInstanceStatus(getContext(), &pb.UpdateInstanceStatusRequest{
					ServiceId:  serviceId,
					InstanceId: "notexistins",
					Status:     pb.MSI_STARTING,
				})

				Expect(err).To(BeNil())
				Expect(respUpdateStatus.Response.GetCode()).ToNot(Equal(pb.ResponseSuccess))

				By("update invalid status")
				respUpdateStatus, err = discosvc.UpdateInstanceStatus(getContext(), &pb.UpdateInstanceStatusRequest{
					ServiceId:  serviceId,
					InstanceId: instanceId,
					Status:     "nonestatus",
				})

				Expect(err).To(BeNil())
				Expect(respUpdateStatus.Response.GetCode()).ToNot(Equal(pb.ResponseSuccess))
			})
		})

		Context("when update instance properties", func() {
			It("should be passed", func() {
				By("update instance properties")
				respUpdateProperties, err := discosvc.UpdateInstanceProperties(getContext(), &pb.UpdateInstancePropsRequest{
					ServiceId:  serviceId,
					InstanceId: instanceId,
					Properties: map[string]string{
						"test": "test",
					},
				})

				Expect(err).To(BeNil())
				Expect(respUpdateProperties.Response.GetCode()).To(Equal(pb.ResponseSuccess))

				By("all max")
				size := 1000
				properties := make(map[string]string, size)
				for i := 0; i < size; i++ {
					s := strconv.Itoa(i) + strings.Repeat("x", 253)
					properties[s] = s
				}
				respUpdateProperties, err = discosvc.UpdateInstanceProperties(getContext(), &pb.UpdateInstancePropsRequest{
					ServiceId:  serviceId,
					InstanceId: instanceId,
					Properties: properties,
				})

				Expect(err).To(BeNil())
				Expect(respUpdateProperties.Response.GetCode()).To(Equal(pb.ResponseSuccess))

				By("instance does not exist")
				respUpdateProperties, err = discosvc.UpdateInstanceProperties(getContext(), &pb.UpdateInstancePropsRequest{
					ServiceId:  serviceId,
					InstanceId: "notexistins",
					Properties: map[string]string{
						"test": "test",
					},
				})

				Expect(err).To(BeNil())
				Expect(respUpdateProperties.Response.GetCode()).ToNot(Equal(pb.ResponseSuccess))

				By("serviceId/instanceId/props is invalid")
				respUpdateProperties, err = discosvc.UpdateInstanceProperties(getContext(), &pb.UpdateInstancePropsRequest{
					ServiceId:  "",
					InstanceId: instanceId,
					Properties: map[string]string{
						"test": "test",
					},
				})

				Expect(err).To(BeNil())
				Expect(respUpdateProperties.Response.GetCode()).To(Equal(pb.ErrInvalidParams))
				respUpdateProperties, err = discosvc.UpdateInstanceProperties(getContext(), &pb.UpdateInstancePropsRequest{
					ServiceId:  TOO_LONG_SERVICEID,
					InstanceId: instanceId,
					Properties: map[string]string{
						"test": "test",
					},
				})

				Expect(err).To(BeNil())
				Expect(respUpdateProperties.Response.GetCode()).To(Equal(pb.ErrInvalidParams))
				respUpdateProperties, err = discosvc.UpdateInstanceProperties(getContext(), &pb.UpdateInstancePropsRequest{
					ServiceId:  serviceId,
					InstanceId: "",
					Properties: map[string]string{
						"test": "test",
					},
				})

				Expect(err).To(BeNil())
				Expect(respUpdateProperties.Response.GetCode()).To(Equal(pb.ErrInvalidParams))
				respUpdateProperties, err = discosvc.UpdateInstanceProperties(getContext(), &pb.UpdateInstancePropsRequest{
					ServiceId:  serviceId,
					InstanceId: "@",
					Properties: map[string]string{
						"test": "test",
					},
				})

				Expect(err).To(BeNil())
				Expect(respUpdateProperties.Response.GetCode()).To(Equal(pb.ErrInvalidParams))
				respUpdateProperties, err = discosvc.UpdateInstanceProperties(getContext(), &pb.UpdateInstancePropsRequest{
					ServiceId:  serviceId,
					InstanceId: TOO_LONG_SERVICEID,
					Properties: map[string]string{
						"test": "test",
					},
				})

				Expect(err).To(BeNil())
				Expect(respUpdateProperties.Response.GetCode()).To(Equal(pb.ErrInvalidParams))

				By("remove the properties")
				respUpdateProperties, err = discosvc.UpdateInstanceProperties(getContext(), &pb.UpdateInstancePropsRequest{
					ServiceId:  serviceId,
					InstanceId: instanceId,
				})
				Expect(err).To(BeNil())
				Expect(respUpdateProperties.Response.GetCode()).To(Equal(pb.ResponseSuccess))

				By("service does not exist")
				respUpdateProperties, err = discosvc.UpdateInstanceProperties(getContext(), &pb.UpdateInstancePropsRequest{
					ServiceId:  "notexistservice",
					InstanceId: instanceId,
					Properties: map[string]string{
						"test": "test",
					},
				})

				Expect(err).To(BeNil())
				Expect(respUpdateProperties.Response.GetCode()).ToNot(Equal(pb.ResponseSuccess))
			})
		})
	})

	Describe("execute 'query' operartion", func() {
		var (
			serviceId1   string
			serviceId2   string
			serviceId3   string
			serviceId4   string
			serviceId5   string
			serviceId6   string
			serviceId7   string
			serviceId8   string
			serviceId9   string
			serviceId10  string
			instanceId1  string
			instanceId2  string
			instanceId4  string
			instanceId5  string
			instanceId8  string
			instanceId9  string
			instanceId10 string
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
			Expect(respCreate.Response.GetCode()).To(Equal(pb.ResponseSuccess))
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
			Expect(respCreate.Response.GetCode()).To(Equal(pb.ResponseSuccess))
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
			Expect(respCreate.Response.GetCode()).To(Equal(pb.ResponseSuccess))
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
			Expect(respCreate.Response.GetCode()).To(Equal(pb.ResponseSuccess))
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
						pb.PropAllowCrossApp: "true",
					},
				},
			})
			Expect(err).To(BeNil())
			Expect(respCreate.Response.GetCode()).To(Equal(pb.ResponseSuccess))
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
			Expect(respCreate.Response.GetCode()).To(Equal(pb.ResponseSuccess))
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
			Expect(respCreate.Response.GetCode()).To(Equal(pb.ResponseSuccess))
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
			Expect(respCreate.Response.GetCode()).To(Equal(pb.ResponseSuccess))
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
			Expect(respCreate.Response.GetCode()).To(Equal(pb.ResponseSuccess))
			serviceId9 = respCreate.ServiceId

			respCreate, err = serviceResource.Create(getContext(), &pb.CreateServiceRequest{
				Service: &pb.MicroService{
					AppId:       "query_instance_alias",
					ServiceName: "query_instance_alias",
					Alias:       "query_instance_alias:query_instance_alias",
					Version:     "1.0.0",
					Level:       "FRONT",
					Status:      pb.MS_UP,
				},
			})
			Expect(err).To(BeNil())
			Expect(respCreate.Response.GetCode()).To(Equal(pb.ResponseSuccess))
			serviceId10 = respCreate.ServiceId

			resp, err := discosvc.RegisterInstance(getContext(), &pb.RegisterInstanceRequest{
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
			Expect(resp.Response.GetCode()).To(Equal(pb.ResponseSuccess))
			instanceId1 = resp.InstanceId
			Expect(instanceId1).To(Equal(instanceId1))

			resp, err = discosvc.RegisterInstance(getContext(), &pb.RegisterInstanceRequest{
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
			Expect(resp.Response.GetCode()).To(Equal(pb.ResponseSuccess))
			instanceId2 = resp.InstanceId

			resp, err = discosvc.RegisterInstance(getContext(), &pb.RegisterInstanceRequest{
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
			Expect(resp.Response.GetCode()).To(Equal(pb.ResponseSuccess))
			instanceId4 = resp.InstanceId

			resp, err = discosvc.RegisterInstance(getContext(), &pb.RegisterInstanceRequest{
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
			Expect(resp.Response.GetCode()).To(Equal(pb.ResponseSuccess))
			instanceId5 = resp.InstanceId

			resp, err = discosvc.RegisterInstance(getContext(), &pb.RegisterInstanceRequest{
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
			Expect(resp.Response.GetCode()).To(Equal(pb.ResponseSuccess))
			instanceId8 = resp.InstanceId

			resp, err = discosvc.RegisterInstance(getContext(), &pb.RegisterInstanceRequest{
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
			Expect(resp.Response.GetCode()).To(Equal(pb.ResponseSuccess))
			instanceId9 = resp.InstanceId

			resp, err = discosvc.RegisterInstance(getContext(), &pb.RegisterInstanceRequest{
				Instance: &pb.MicroServiceInstance{
					ServiceId: serviceId10,
					HostName:  "UT-HOST",
					Endpoints: []string{
						"find:127.0.0.9:8080",
					},
					Status: pb.MSI_UP,
				},
			})
			Expect(err).To(BeNil())
			Expect(resp.Response.GetCode()).To(Equal(pb.ResponseSuccess))
			instanceId10 = resp.InstanceId
		})

		Context("when query invalid parameters", func() {
			It("should be failed", func() {
				By("invalid appId")
				respFind, err := discosvc.FindInstances(getContext(), &pb.FindInstancesRequest{
					ConsumerServiceId: serviceId1,
					AppId:             TOO_LONG_APPID,
					ServiceName:       "query_instance_service",
				})
				Expect(err).To(BeNil())
				Expect(respFind.Response.GetCode()).To(Equal(pb.ErrInvalidParams))
				respFind, err = discosvc.FindInstances(getContext(), &pb.FindInstancesRequest{
					ConsumerServiceId: serviceId1,
					AppId:             "",
					ServiceName:       "query_instance_service",
				})
				Expect(err).To(BeNil())
				Expect(respFind.Response.GetCode()).To(Equal(pb.ErrInvalidParams))
				respFind, err = discosvc.FindInstances(getContext(), &pb.FindInstancesRequest{
					ConsumerServiceId: serviceId1,
					AppId:             " ",
					ServiceName:       "query_instance_service",
				})
				Expect(err).To(BeNil())
				Expect(respFind.Response.GetCode()).To(Equal(pb.ErrInvalidParams))

				By("invalid serviceName")
				respFind, err = discosvc.FindInstances(getContext(), &pb.FindInstancesRequest{
					ConsumerServiceId: serviceId1,
					AppId:             "query_instance",
					ServiceName:       TOO_LONG_EXISTENCE,
				})
				Expect(err).To(BeNil())
				Expect(respFind.Response.GetCode()).To(Equal(pb.ErrInvalidParams))
				respFind, err = discosvc.FindInstances(getContext(), &pb.FindInstancesRequest{
					ConsumerServiceId: serviceId1,
					AppId:             "query_instance",
					ServiceName:       "",
				})
				Expect(err).To(BeNil())
				Expect(respFind.Response.GetCode()).To(Equal(pb.ErrInvalidParams))
				respFind, err = discosvc.FindInstances(getContext(), &pb.FindInstancesRequest{
					ConsumerServiceId: serviceId1,
					AppId:             "query_instance",
					ServiceName:       " ",
				})
				Expect(err).To(BeNil())
				Expect(respFind.Response.GetCode()).To(Equal(pb.ErrInvalidParams))

				By("provider does not exist")
				respFind, err = discosvc.FindInstances(getContext(), &pb.FindInstancesRequest{
					ConsumerServiceId: serviceId1,
					AppId:             "query_instance",
					ServiceName:       "noneservice",
				})
				Expect(err).To(BeNil())
				Expect(respFind.Response.GetCode()).To(Equal(pb.ErrServiceNotExists))

				By("consumer does not exist")
				respFind, err = discosvc.FindInstances(getContext(), &pb.FindInstancesRequest{
					ConsumerServiceId: "notExistServiceId",
					AppId:             "query_instance",
					ServiceName:       "query_instance_service",
				})
				Expect(err).To(BeNil())
				Expect(respFind.Response.GetCode()).To(Equal(pb.ErrServiceNotExists))
			})
		})

		Context("when batch query invalid parameters", func() {
			It("should be failed", func() {
				By("invalid services")
				respFind, err := discosvc.BatchFindInstances(getContext(), &pb.BatchFindInstancesRequest{
					ConsumerServiceId: serviceId1,
					Services:          nil,
					Instances:         nil,
				})
				Expect(err).To(BeNil())
				Expect(respFind.Response.GetCode()).To(Equal(pb.ErrInvalidParams))
				respFind, err = discosvc.BatchFindInstances(getContext(), &pb.BatchFindInstancesRequest{
					ConsumerServiceId: serviceId1,
					Services:          []*pb.FindService{},
					Instances:         []*pb.FindInstance{},
				})
				Expect(err).To(BeNil())
				Expect(respFind.Response.GetCode()).To(Equal(pb.ErrInvalidParams))
				respFind, err = discosvc.BatchFindInstances(getContext(), &pb.BatchFindInstancesRequest{
					ConsumerServiceId: serviceId1,
					Services:          []*pb.FindService{{}},
				})
				Expect(err).To(BeNil())
				Expect(respFind.Response.GetCode()).To(Equal(pb.ErrInvalidParams))
				respFind, err = discosvc.BatchFindInstances(getContext(), &pb.BatchFindInstancesRequest{
					ConsumerServiceId: serviceId1,
					Instances:         []*pb.FindInstance{{}},
				})
				Expect(err).To(BeNil())
				Expect(respFind.Response.GetCode()).To(Equal(pb.ErrInvalidParams))

				By("invalid appId")
				respFind, err = discosvc.BatchFindInstances(getContext(), &pb.BatchFindInstancesRequest{
					ConsumerServiceId: serviceId1,
					Services: []*pb.FindService{
						{
							Service: &pb.MicroServiceKey{
								AppId:       TOO_LONG_APPID,
								ServiceName: "query_instance_service",
							},
						},
					},
				})
				Expect(err).To(BeNil())
				Expect(respFind.Response.GetCode()).To(Equal(pb.ErrInvalidParams))
				respFind, err = discosvc.BatchFindInstances(getContext(), &pb.BatchFindInstancesRequest{
					ConsumerServiceId: serviceId1,
					Services: []*pb.FindService{
						{
							Service: &pb.MicroServiceKey{
								AppId:       "",
								ServiceName: "query_instance_service",
							},
						},
					},
				})
				Expect(err).To(BeNil())
				Expect(respFind.Response.GetCode()).To(Equal(pb.ErrInvalidParams))
				respFind, err = discosvc.BatchFindInstances(getContext(), &pb.BatchFindInstancesRequest{
					ConsumerServiceId: serviceId1,
					Services: []*pb.FindService{
						{
							Service: &pb.MicroServiceKey{
								AppId:       " ",
								ServiceName: "query_instance_service",
							},
						},
					},
				})
				Expect(err).To(BeNil())
				Expect(respFind.Response.GetCode()).To(Equal(pb.ErrInvalidParams))

				By("invalid serviceName")
				respFind, err = discosvc.BatchFindInstances(getContext(), &pb.BatchFindInstancesRequest{
					ConsumerServiceId: serviceId1,
					Services: []*pb.FindService{
						{
							Service: &pb.MicroServiceKey{
								AppId:       "query_instance",
								ServiceName: TOO_LONG_EXISTENCE,
							},
						},
					},
				})
				Expect(err).To(BeNil())
				Expect(respFind.Response.GetCode()).To(Equal(pb.ErrInvalidParams))
				respFind, err = discosvc.BatchFindInstances(getContext(), &pb.BatchFindInstancesRequest{
					ConsumerServiceId: serviceId1,
					Services: []*pb.FindService{
						{
							Service: &pb.MicroServiceKey{
								AppId:       "query_instance",
								ServiceName: "",
							},
						},
					},
				})
				Expect(err).To(BeNil())
				Expect(respFind.Response.GetCode()).To(Equal(pb.ErrInvalidParams))
				respFind, err = discosvc.BatchFindInstances(getContext(), &pb.BatchFindInstancesRequest{
					ConsumerServiceId: serviceId1,
					Services: []*pb.FindService{
						{
							Service: &pb.MicroServiceKey{
								AppId:       "query_instance",
								ServiceName: " ",
							},
						},
					},
				})
				Expect(err).To(BeNil())
				Expect(respFind.Response.GetCode()).To(Equal(pb.ErrInvalidParams))

				By("invalid instance")
				respFind, err = discosvc.BatchFindInstances(getContext(), &pb.BatchFindInstancesRequest{
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
				Expect(respFind.Response.GetCode()).To(Equal(pb.ErrInvalidParams))
				respFind, err = discosvc.BatchFindInstances(getContext(), &pb.BatchFindInstancesRequest{
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
				Expect(respFind.Response.GetCode()).To(Equal(pb.ErrInvalidParams))

				By("consumerId is empty")
				respFind, err = discosvc.BatchFindInstances(getContext(), &pb.BatchFindInstancesRequest{
					Services: []*pb.FindService{
						{
							Service: &pb.MicroServiceKey{
								AppId:       "query_instance",
								ServiceName: "query_instance_service",
							},
						},
					},
				})
				Expect(err).To(BeNil())
				Expect(respFind.Response.GetCode()).To(Equal(pb.ResponseSuccess))

				By("provider does not exist")
				respFind, err = discosvc.BatchFindInstances(getContext(), &pb.BatchFindInstancesRequest{
					ConsumerServiceId: serviceId1,
					Services: []*pb.FindService{
						{
							Service: &pb.MicroServiceKey{
								AppId:       "query_instance",
								ServiceName: "noneservice",
							},
						},
					},
				})
				Expect(err).To(BeNil())
				Expect(respFind.Response.GetCode()).To(Equal(pb.ResponseSuccess))
				Expect(respFind.Services.Failed[0].Error.Code).To(Equal(pb.ErrServiceNotExists))
				Expect(respFind.Services.Failed[0].Indexes[0]).To(Equal(int64(0)))
				respFind, err = discosvc.BatchFindInstances(getContext(), &pb.BatchFindInstancesRequest{
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
				Expect(respFind.Response.GetCode()).To(Equal(pb.ResponseSuccess))
				Expect(respFind.Instances.Failed[0].Error.Code).To(Equal(pb.ErrInstanceNotExists))
				Expect(respFind.Instances.Failed[0].Indexes[0]).To(Equal(int64(0)))

				By("consumer does not exist")
				respFind, err = discosvc.BatchFindInstances(getContext(), &pb.BatchFindInstancesRequest{
					ConsumerServiceId: "notExistServiceId",
					Services: []*pb.FindService{
						{
							Service: &pb.MicroServiceKey{
								AppId:       "query_instance",
								ServiceName: "query_instance_service",
							},
						},
					},
				})
				Expect(err).To(BeNil())
				Expect(respFind.Response.GetCode()).To(Equal(pb.ResponseSuccess))
				Expect(respFind.Services.Failed[0].Indexes[0]).To(Equal(int64(0)))
				Expect(respFind.Services.Failed[0].Error.Code).To(Equal(pb.ErrServiceNotExists))
			})
		})

		Context("when query instances", func() {
			It("without consumerID should be passed", func() {
				By("consumerId is empty")
				respFind, err := discosvc.FindInstances(getContext(), &pb.FindInstancesRequest{
					ConsumerServiceId: "",
					AppId:             "query_instance",
					ServiceName:       "query_instance_service",
				})
				Expect(err).To(BeNil())
				Expect(respFind.Response.GetCode()).To(Equal(pb.ResponseSuccess))
			})

			It("with consumerID should be passed", func() {
				By("find")
				respFind, err := discosvc.FindInstances(getContext(), &pb.FindInstancesRequest{
					ConsumerServiceId: serviceId1,
					AppId:             "query_instance",
					ServiceName:       "query_instance_service",
				})
				Expect(err).To(BeNil())
				Expect(respFind.Response.GetCode()).To(Equal(pb.ResponseSuccess))
				assertInstanceContain(respFind.Instances, instanceId2)

				By("find with env")
				respFind, err = discosvc.FindInstances(getContext(), &pb.FindInstancesRequest{
					ConsumerServiceId: serviceId4,
					AppId:             "query_instance",
					ServiceName:       "query_instance_diff_env_service",
				})
				Expect(err).To(BeNil())
				Expect(respFind.Response.GetCode()).To(Equal(pb.ResponseSuccess))
				Expect(len(respFind.Instances)).To(Equal(1))
				Expect(respFind.Instances[0].InstanceId).To(Equal(instanceId4))

				respFind, err = discosvc.FindInstances(getContext(), &pb.FindInstancesRequest{
					Environment: pb.ENV_PROD,
					AppId:       "query_instance",
					ServiceName: "query_instance_diff_env_service",
				})
				Expect(err).To(BeNil())
				Expect(respFind.Response.GetCode()).To(Equal(pb.ResponseSuccess))
				Expect(len(respFind.Instances)).To(Equal(1))
				Expect(respFind.Instances[0].InstanceId).To(Equal(instanceId4))

				By("find with rev")
				ctx := util.SetContext(getContext(), util.CtxNocache, "")
				respFind, err = discosvc.FindInstances(ctx, &pb.FindInstancesRequest{
					ConsumerServiceId: serviceId8,
					AppId:             "query_instance",
					ServiceName:       "query_instance_with_rev",
				})
				Expect(err).To(BeNil())
				Expect(respFind.Response.GetCode()).To(Equal(pb.ResponseSuccess))
				rev, _ := ctx.Value(util.CtxResponseRevision).(string)
				Expect(respFind.Instances[0].InstanceId).To(Equal(instanceId8))
				Expect(len(rev)).NotTo(Equal(0))

				util.WithRequestRev(ctx, "x")
				respFind, err = discosvc.FindInstances(ctx, &pb.FindInstancesRequest{
					ConsumerServiceId: serviceId8,
					AppId:             "query_instance",
					ServiceName:       "query_instance_with_rev",
				})
				Expect(err).To(BeNil())
				Expect(respFind.Response.GetCode()).To(Equal(pb.ResponseSuccess))
				Expect(respFind.Instances[0].InstanceId).To(Equal(instanceId8))
				Expect(ctx.Value(util.CtxResponseRevision)).To(Equal(rev))

				util.WithRequestRev(ctx, rev)
				respFind, err = discosvc.FindInstances(ctx, &pb.FindInstancesRequest{
					ConsumerServiceId: serviceId8,
					AppId:             "query_instance",
					ServiceName:       "query_instance_with_rev",
				})
				Expect(err).To(BeNil())
				Expect(respFind.Response.GetCode()).To(Equal(pb.ResponseSuccess))
				Expect(len(respFind.Instances)).To(Equal(0))
				Expect(ctx.Value(util.CtxResponseRevision)).To(Equal(rev))

				By("find should return 200 even if consumer is diff apps")
				respFind, err = discosvc.FindInstances(getContext(), &pb.FindInstancesRequest{
					ConsumerServiceId: serviceId3,
					AppId:             "query_instance",
					ServiceName:       "query_instance_service",
				})
				Expect(err).To(BeNil())
				Expect(respFind.Response.GetCode()).To(Equal(pb.ResponseSuccess))
				Expect(len(respFind.Instances)).To(Equal(0))

				By("provider tag does not exist")
				respFind, err = discosvc.FindInstances(getContext(), &pb.FindInstancesRequest{
					ConsumerServiceId: serviceId1,
					AppId:             "query_instance",
					ServiceName:       "query_instance_service",
					Tags:              []string{"notexisttag"},
				})
				Expect(err).To(BeNil())
				Expect(respFind.Response.GetCode()).To(Equal(pb.ResponseSuccess))
				Expect(len(respFind.Instances)).To(Equal(0))

				By("shared service discovery")
				config.Server.Config.GlobalVisible = "query_instance_shared_provider"
				core.RegisterGlobalServices()
				core.Service.Environment = pb.ENV_PROD

				respFind, err = discosvc.FindInstances(
					util.SetTargetDomainProject(
						util.SetDomainProject(util.CloneContext(getContext()), "user", "user"),
						"default", "default"),
					&pb.FindInstancesRequest{
						ConsumerServiceId: serviceId6,
						AppId:             "default",
						ServiceName:       "query_instance_shared_provider",
					})
				Expect(err).To(BeNil())
				Expect(respFind.Response.GetCode()).To(Equal(pb.ResponseSuccess))
				Expect(len(respFind.Instances)).To(Equal(1))
				Expect(respFind.Instances[0].InstanceId).To(Equal(instanceId5))

				respFind, err = discosvc.FindInstances(getContext(), &pb.FindInstancesRequest{
					ConsumerServiceId: serviceId7,
					AppId:             "default",
					ServiceName:       "query_instance_shared_provider",
				})
				Expect(err).To(BeNil())
				Expect(respFind.Response.GetCode()).To(Equal(pb.ResponseSuccess))
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
				Expect(respGetC.Response.GetCode()).To(Equal(pb.ResponseSuccess))
				Expect(len(respGetC.Providers)).To(Equal(0))

				core.Service.Environment = pb.ENV_DEV

				By("find with alias")
				respFind, err = discosvc.FindInstances(getContext(), &pb.FindInstancesRequest{
					ConsumerServiceId: serviceId10,
					AppId:             "query_instance_alias",
					ServiceName:       "query_instance_alias:query_instance_alias",
					Alias:             "query_instance_alias:query_instance_alias",
				})
				Expect(err).To(BeNil())
				Expect(respFind.Response.GetCode()).To(Equal(pb.ResponseSuccess))
				Expect(len(respFind.Instances)).To(Equal(1))
				Expect(respFind.Instances[0].InstanceId).To(Equal(instanceId10))
			})
		})

		Context("when batch query instances", func() {
			It("should be passed", func() {
				By("find with version rule")
				respFind, err := discosvc.BatchFindInstances(getContext(), &pb.BatchFindInstancesRequest{
					ConsumerServiceId: serviceId1,
					Services: []*pb.FindService{
						{
							Service: &pb.MicroServiceKey{
								AppId:       "query_instance",
								ServiceName: "query_instance_service",
							},
						},
						{
							Service: &pb.MicroServiceKey{
								AppId:       "query_instance",
								ServiceName: "query_instance_diff_env_service",
							},
						},
						{
							Service: &pb.MicroServiceKey{
								AppId:       "query_instance_diff_app",
								ServiceName: "query_instance_service",
							},
						},
						{
							Service: &pb.MicroServiceKey{
								AppId:       "not-exists",
								ServiceName: "not-exists",
							},
						},
					},
				})
				Expect(err).To(BeNil())
				Expect(respFind.Response.GetCode()).To(Equal(pb.ResponseSuccess))
				Expect(respFind.Services.Updated[0].Index).To(Equal(int64(0)))
				assertInstanceContain(respFind.Services.Updated[0].Instances, instanceId1)
				assertInstanceContain(respFind.Services.Updated[0].Instances, instanceId2)
				Expect(respFind.Services.Updated[1].Index).To(Equal(int64(2)))
				Expect(respFind.Services.Updated[1].Instances).To(BeEmpty())
				Expect(len(respFind.Services.Failed[0].Indexes)).To(Equal(2))
				Expect(respFind.Services.Failed[0].Error.Code).To(Equal(pb.ErrServiceNotExists))

				By("find with env")
				respFind, err = discosvc.BatchFindInstances(getContext(), &pb.BatchFindInstancesRequest{
					ConsumerServiceId: serviceId4,
					Services: []*pb.FindService{
						{
							Service: &pb.MicroServiceKey{
								AppId:       "query_instance",
								ServiceName: "query_instance_diff_env_service",
							},
						},
					},
				})
				Expect(err).To(BeNil())
				Expect(respFind.Response.GetCode()).To(Equal(pb.ResponseSuccess))
				Expect(len(respFind.Services.Updated[0].Instances)).To(Equal(1))
				Expect(respFind.Services.Updated[0].Instances[0].InstanceId).To(Equal(instanceId4))

				respFind, err = discosvc.BatchFindInstances(getContext(), &pb.BatchFindInstancesRequest{
					Services: []*pb.FindService{
						{
							Service: &pb.MicroServiceKey{
								Environment: pb.ENV_PROD,
								AppId:       "query_instance",
								ServiceName: "query_instance_diff_env_service",
							},
						},
					},
				})
				Expect(err).To(BeNil())
				Expect(respFind.Response.GetCode()).To(Equal(pb.ResponseSuccess))
				Expect(len(respFind.Services.Updated[0].Instances)).To(Equal(1))
				Expect(respFind.Services.Updated[0].Instances[0].InstanceId).To(Equal(instanceId4))

				By("find with rev")
				ctx := util.SetContext(getContext(), util.CtxNocache, "")
				respFind, err = discosvc.BatchFindInstances(ctx, &pb.BatchFindInstancesRequest{
					ConsumerServiceId: serviceId8,
					Services: []*pb.FindService{
						{
							Service: &pb.MicroServiceKey{
								AppId:       "query_instance",
								ServiceName: "query_instance_with_rev",
							},
						},
						{
							Service: &pb.MicroServiceKey{
								AppId:       "query_instance",
								ServiceName: "batch_query_instance_with_rev",
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
				Expect(respFind.Response.GetCode()).To(Equal(pb.ResponseSuccess))
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

				respFind, err = discosvc.BatchFindInstances(ctx, &pb.BatchFindInstancesRequest{
					ConsumerServiceId: serviceId8,
					Services: []*pb.FindService{
						{
							Service: &pb.MicroServiceKey{
								AppId:       "query_instance",
								ServiceName: "query_instance_with_rev",
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
				Expect(respFind.Response.GetCode()).To(Equal(pb.ResponseSuccess))
				Expect(respFind.Services.Updated[0].Instances[0].InstanceId).To(Equal(instanceId8))
				Expect(respFind.Services.Updated[0].Rev).To(Equal(rev))
				Expect(respFind.Instances.Updated[0].Instances[0].InstanceId).To(Equal(instanceId9))
				Expect(respFind.Instances.Updated[0].Rev).To(Equal(instanceRev))

				respFind, err = discosvc.BatchFindInstances(ctx, &pb.BatchFindInstancesRequest{
					ConsumerServiceId: serviceId8,
					Services: []*pb.FindService{
						{
							Service: &pb.MicroServiceKey{
								AppId:       "query_instance",
								ServiceName: "query_instance_with_rev",
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
				Expect(respFind.Response.GetCode()).To(Equal(pb.ResponseSuccess))
				Expect(respFind.Services.NotModified[0]).To(Equal(int64(0)))
				Expect(respFind.Instances.NotModified[0]).To(Equal(int64(0)))

				By("find should return 200 even if consumer is diff apps")
				respFind, err = discosvc.BatchFindInstances(getContext(), &pb.BatchFindInstancesRequest{
					ConsumerServiceId: serviceId3,
					Services: []*pb.FindService{
						{
							Service: &pb.MicroServiceKey{
								AppId:       "query_instance",
								ServiceName: "query_instance_service",
							},
						},
					},
				})
				Expect(err).To(BeNil())
				Expect(respFind.Response.GetCode()).To(Equal(pb.ResponseSuccess))
				Expect(len(respFind.Services.Updated[0].Instances)).To(Equal(0))

				By("shared service discovery")
				config.Server.Config.GlobalVisible = "query_instance_shared_provider"
				core.RegisterGlobalServices()
				core.Service.Environment = pb.ENV_PROD

				respFind, err = discosvc.BatchFindInstances(
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
								},
							},
						},
					})
				Expect(err).To(BeNil())
				Expect(respFind.Response.GetCode()).To(Equal(pb.ResponseSuccess))
				Expect(len(respFind.Services.Updated[0].Instances)).To(Equal(1))
				Expect(respFind.Services.Updated[0].Instances[0].InstanceId).To(Equal(instanceId5))

				respFind, err = discosvc.BatchFindInstances(getContext(), &pb.BatchFindInstancesRequest{
					ConsumerServiceId: serviceId7,
					Services: []*pb.FindService{
						{
							Service: &pb.MicroServiceKey{
								AppId:       "default",
								ServiceName: "query_instance_shared_provider",
							},
						},
					},
				})
				Expect(err).To(BeNil())
				Expect(respFind.Response.GetCode()).To(Equal(pb.ResponseSuccess))
				Expect(len(respFind.Services.Updated[0].Instances)).To(Equal(1))
				Expect(respFind.Services.Updated[0].Instances[0].InstanceId).To(Equal(instanceId5))

				respFind, err = discosvc.BatchFindInstances(
					util.SetTargetDomainProject(
						util.SetDomainProject(util.CloneContext(getContext()), "user", "user"),
						"default", "default"),
					&pb.BatchFindInstancesRequest{
						ConsumerServiceId: serviceId6,
						Instances: []*pb.FindInstance{
							{
								Instance: &pb.HeartbeatSetElement{
									ServiceId:  serviceId5,
									InstanceId: instanceId5,
								},
							},
						},
					})
				Expect(err).To(BeNil())
				Expect(respFind.Response.GetCode()).To(Equal(pb.ResponseSuccess))
				Expect(respFind.Instances.Failed[0].Error.Code).To(Equal(pb.ErrServiceNotExists))

				respFind, err = discosvc.BatchFindInstances(getContext(), &pb.BatchFindInstancesRequest{
					ConsumerServiceId: serviceId7,
					Instances: []*pb.FindInstance{
						{
							Instance: &pb.HeartbeatSetElement{
								ServiceId:  serviceId5,
								InstanceId: instanceId5,
							},
						},
					},
				})
				Expect(err).To(BeNil())
				Expect(respFind.Response.GetCode()).To(Equal(pb.ResponseSuccess))
				Expect(len(respFind.Instances.Updated[0].Instances)).To(Equal(1))
				Expect(respFind.Instances.Updated[0].Instances[0].InstanceId).To(Equal(instanceId5))

				core.Service.Environment = pb.ENV_DEV
			})
		})

		Context("when query instances between diff dimensions", func() {
			It("should be failed", func() {
				By("diff appId")
				UTFunc := func(consumerId string, code int32) {
					respFind, err := discosvc.GetInstances(getContext(), &pb.GetInstancesRequest{
						ConsumerServiceId: consumerId,
						ProviderServiceId: serviceId2,
					})
					Expect(err).To(BeNil())
					Expect(respFind.Response.GetCode()).To(Equal(code))
				}

				UTFunc(serviceId3, pb.ErrServiceNotExists)

				UTFunc(serviceId1, pb.ResponseSuccess)

				By("diff env")
				respFind, err := discosvc.GetInstances(getContext(), &pb.GetInstancesRequest{
					ConsumerServiceId: serviceId4,
					ProviderServiceId: serviceId2,
				})
				Expect(err).To(BeNil())
				Expect(respFind.Response.GetCode()).ToNot(Equal(pb.ResponseSuccess))
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
			Expect(respCreate.Response.GetCode()).To(Equal(pb.ResponseSuccess))
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
			Expect(respCreate.Response.GetCode()).To(Equal(pb.ResponseSuccess))
			serviceId2 = respCreate.ServiceId

			resp, err := discosvc.RegisterInstance(getContext(), &pb.RegisterInstanceRequest{
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
			Expect(resp.Response.GetCode()).To(Equal(pb.ResponseSuccess))
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
				resp, err := discosvc.GetOneInstance(getContext(), &pb.GetOneInstanceRequest{
					ConsumerServiceId:  serviceId2,
					ProviderServiceId:  serviceId2,
					ProviderInstanceId: instanceId2,
				})
				Expect(err).To(BeNil())
				Expect(resp.Response.GetCode()).To(Equal(pb.ResponseSuccess))

				By("provider id is empty")
				resp, err = discosvc.GetOneInstance(getContext(), &pb.GetOneInstanceRequest{
					ConsumerServiceId:  serviceId1,
					ProviderServiceId:  "",
					ProviderInstanceId: instanceId2,
				})
				Expect(err).To(BeNil())
				Expect(resp.Response.GetCode()).ToNot(Equal(pb.ResponseSuccess))

				By("provider instance id is empty")
				resp, err = discosvc.GetOneInstance(getContext(), &pb.GetOneInstanceRequest{
					ConsumerServiceId:  serviceId1,
					ProviderServiceId:  serviceId2,
					ProviderInstanceId: "",
				})
				Expect(err).To(BeNil())
				Expect(resp.Response.GetCode()).ToNot(Equal(pb.ResponseSuccess))

				By("consumer id is empty")
				resp, err = discosvc.GetOneInstance(getContext(), &pb.GetOneInstanceRequest{
					ConsumerServiceId:  "",
					ProviderServiceId:  serviceId2,
					ProviderInstanceId: instanceId2,
				})
				Expect(err).To(BeNil())
				Expect(resp.Response.GetCode()).To(Equal(pb.ResponseSuccess))

				By("consumer does not exist")
				resp, err = discosvc.GetOneInstance(getContext(), &pb.GetOneInstanceRequest{
					ConsumerServiceId:  "not-exist-id",
					ProviderServiceId:  serviceId2,
					ProviderInstanceId: instanceId2,
				})
				Expect(err).To(BeNil())
				Expect(resp.Response.GetCode()).ToNot(Equal(pb.ResponseSuccess))

				By("provider tag does not exist")
				resp, err = discosvc.GetOneInstance(getContext(), &pb.GetOneInstanceRequest{
					ConsumerServiceId:  serviceId1,
					ProviderServiceId:  serviceId2,
					ProviderInstanceId: instanceId2,
					Tags:               []string{"not-exist-tag"},
				})
				Expect(err).To(BeNil())
				Expect(resp.Response.GetCode()).To(Equal(pb.ErrInstanceNotExists))

				By("provider tag exist")
				resp, err = discosvc.GetOneInstance(getContext(),
					&pb.GetOneInstanceRequest{
						ConsumerServiceId:  serviceId1,
						ProviderServiceId:  serviceId2,
						ProviderInstanceId: instanceId2,
						Tags:               []string{"test"},
					})
				Expect(err).To(BeNil())
				Expect(resp.Response.GetCode()).To(Equal(pb.ResponseSuccess))
			})
		})

		Context("when get between diff apps", func() {
			It("should be failed", func() {
				resp, err := discosvc.GetOneInstance(getContext(), &pb.GetOneInstanceRequest{
					ConsumerServiceId:  serviceId3,
					ProviderServiceId:  serviceId2,
					ProviderInstanceId: instanceId2,
				})
				Expect(err).To(BeNil())
				Expect(resp.Response.GetCode()).To(Equal(pb.ErrInstanceNotExists))

				respAll, err := discosvc.GetInstances(getContext(), &pb.GetInstancesRequest{
					ConsumerServiceId: serviceId3,
					ProviderServiceId: serviceId2,
				})
				Expect(err).To(BeNil())
				Expect(respAll.Response.GetCode()).ToNot(Equal(pb.ResponseSuccess))
			})
		})

		Context("when get instances request is invalid", func() {
			It("should be failed", func() {
				By("consumer id is empty")
				resp, err := discosvc.GetInstances(getContext(), &pb.GetInstancesRequest{
					ConsumerServiceId: "",
					ProviderServiceId: serviceId2,
				})
				Expect(err).To(BeNil())
				Expect(resp.Response.GetCode()).To(Equal(pb.ResponseSuccess))

				By("consumer does not exist")
				resp, err = discosvc.GetInstances(getContext(), &pb.GetInstancesRequest{
					ConsumerServiceId: "noneservice",
					ProviderServiceId: serviceId2,
				})
				Expect(err).To(BeNil())
				Expect(resp.Response.GetCode()).ToNot(Equal(pb.ResponseSuccess))

				By("valid request")
				resp, err = discosvc.GetInstances(getContext(), &pb.GetInstancesRequest{
					ConsumerServiceId: serviceId1,
					ProviderServiceId: serviceId2,
				})
				Expect(err).To(BeNil())
				Expect(resp.Response.GetCode()).To(Equal(pb.ResponseSuccess))
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
			Expect(respCreate.Response.GetCode()).To(Equal(pb.ResponseSuccess))
			serviceId = respCreate.ServiceId

			resp, err := discosvc.RegisterInstance(getContext(), &pb.RegisterInstanceRequest{
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
			Expect(resp.Response.GetCode()).To(Equal(pb.ResponseSuccess))
			instanceId = resp.InstanceId
		})

		Context("when request is valid", func() {
			It("should be passed", func() {
				resp, err := discosvc.UnregisterInstance(getContext(), &pb.UnregisterInstanceRequest{
					ServiceId:  serviceId,
					InstanceId: instanceId,
				})
				Expect(err).To(BeNil())
				Expect(resp.Response.GetCode()).To(Equal(pb.ResponseSuccess))
			})
		})

		Context("when request is invalid", func() {
			It("should be failed", func() {
				By("service id is invalid")
				resp, err := discosvc.UnregisterInstance(getContext(), &pb.UnregisterInstanceRequest{
					ServiceId:  "",
					InstanceId: instanceId,
				})
				Expect(err).To(BeNil())
				Expect(resp.Response.GetCode()).To(Equal(pb.ErrInvalidParams))
				resp, err = discosvc.UnregisterInstance(getContext(), &pb.UnregisterInstanceRequest{
					ServiceId:  TOO_LONG_SERVICEID,
					InstanceId: instanceId,
				})
				Expect(err).To(BeNil())
				Expect(resp.Response.GetCode()).To(Equal(pb.ErrInvalidParams))

				By("service does not exist")
				resp, err = discosvc.UnregisterInstance(getContext(), &pb.UnregisterInstanceRequest{
					ServiceId:  "not-exist-id",
					InstanceId: instanceId,
				})
				Expect(err).To(BeNil())
				Expect(resp.Response.GetCode()).ToNot(Equal(pb.ResponseSuccess))

				By("instance is invalid")
				resp, err = discosvc.UnregisterInstance(getContext(), &pb.UnregisterInstanceRequest{
					ServiceId:  serviceId,
					InstanceId: "",
				})
				Expect(err).To(BeNil())
				Expect(resp.Response.GetCode()).To(Equal(pb.ErrInvalidParams))
				resp, err = discosvc.UnregisterInstance(getContext(), &pb.UnregisterInstanceRequest{
					ServiceId:  serviceId,
					InstanceId: "@",
				})
				Expect(err).To(BeNil())
				Expect(resp.Response.GetCode()).To(Equal(pb.ErrInvalidParams))
				resp, err = discosvc.UnregisterInstance(getContext(), &pb.UnregisterInstanceRequest{
					ServiceId:  serviceId,
					InstanceId: TOO_LONG_SERVICEID,
				})
				Expect(err).To(BeNil())
				Expect(resp.Response.GetCode()).To(Equal(pb.ErrInvalidParams))

				By("instance does not exist")
				resp, err = discosvc.UnregisterInstance(getContext(), &pb.UnregisterInstanceRequest{
					ServiceId:  serviceId,
					InstanceId: "not-exist-id",
				})
				Expect(err).To(BeNil())
				Expect(resp.Response.GetCode()).ToNot(Equal(pb.ResponseSuccess))
			})
		})
	})
})

func assertInstanceContain(instances []*pb.MicroServiceInstance, instanceID string) {
	found := false
	for _, instance := range instances {
		if instance.InstanceId == instanceID {
			found = true
			break
		}
	}
	Expect(found).To(BeTrue())
}
