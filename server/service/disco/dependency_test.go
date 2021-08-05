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
	Describe("execute 'get' operation", func() {
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

				By("get self deps")
				resp, err = disco.FindInstances(getContext(), &pb.FindInstancesRequest{
					ConsumerServiceId: consumerId1,
					AppId:             "get_dep_group",
					ServiceName:       "get_dep_consumer",
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
				Expect(len(respGetC.Providers)).To(Equal(2))

				respGetC, err = serviceResource.GetConsumerDependencies(getContext(), &pb.GetDependenciesRequest{
					ServiceId: consumerId1,
					NoSelf:    false,
				})
				Expect(err).To(BeNil())
				Expect(respGetC.Response.GetCode()).To(Equal(pb.ResponseSuccess))
				Expect(len(respGetC.Providers)).To(Equal(3))

				By("find before provider register")
				resp, err = disco.FindInstances(getContext(), &pb.FindInstancesRequest{
					ConsumerServiceId: providerId2,
					AppId:             "get_dep_group",
					ServiceName:       "get_dep_finder",
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
