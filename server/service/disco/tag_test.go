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
	"fmt"
	"strconv"
	"strings"

	"github.com/apache/servicecomb-service-center/server/service/disco"

	"github.com/apache/servicecomb-service-center/server/plugin/quota"
	pb "github.com/go-chassis/cari/discovery"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var (
	TooLongTag = strings.Repeat("x", 65)
)

var _ = Describe("'Tag' service", func() {
	Describe("execute 'create' operation", func() {
		var (
			serviceId1 string
			serviceId2 string
		)

		It("should be passed", func() {
			respCreateService, err := serviceResource.Create(getContext(), &pb.CreateServiceRequest{
				Service: &pb.MicroService{
					AppId:       "create_tag_group",
					ServiceName: "create_tag_service",
					Version:     "1.0.0",
					Level:       "FRONT",
					Status:      pb.MS_UP,
				},
			})
			Expect(err).To(BeNil())
			Expect(respCreateService.Response.GetCode()).To(Equal(pb.ResponseSuccess))
			serviceId1 = respCreateService.ServiceId

			respCreateService, err = serviceResource.Create(getContext(), &pb.CreateServiceRequest{
				Service: &pb.MicroService{
					AppId:       "create_tag_group",
					ServiceName: "create_tag_service",
					Version:     "1.0.1",
					Level:       "FRONT",
					Status:      pb.MS_UP,
				},
			})
			Expect(err).To(BeNil())
			Expect(respCreateService.Response.GetCode()).To(Equal(pb.ResponseSuccess))
			serviceId2 = respCreateService.ServiceId
		})

		Context("when request is invalid", func() {
			It("should be failed", func() {
				By("service id is empty")
				respAddTags, _ := serviceResource.AddTags(getContext(), &pb.AddServiceTagsRequest{
					ServiceId: "",
					Tags: map[string]string{
						"a": "test",
					},
				})
				Expect(respAddTags.Response.GetCode()).ToNot(Equal(pb.ResponseSuccess))

				By("service does not exist")
				respAddTags, _ = serviceResource.AddTags(getContext(), &pb.AddServiceTagsRequest{
					ServiceId: "noServiceTest",
					Tags: map[string]string{
						"a": "test",
					},
				})
				Expect(respAddTags.Response.GetCode()).ToNot(Equal(pb.ResponseSuccess))

				By("tag key is empty")
				respAddTags, _ = serviceResource.AddTags(getContext(), &pb.AddServiceTagsRequest{
					ServiceId: serviceId1,
					Tags: map[string]string{
						"": "value",
					},
				})
				Expect(respAddTags.Response.GetCode()).ToNot(Equal(pb.ResponseSuccess))
			})
		})

		Context("when request is valid", func() {
			It("should be passed", func() {
				By("all max")
				size := quota.DefaultTagQuota
				tags := make(map[string]string, size)
				for i := 0; i < size; i++ {
					s := "tag" + strconv.Itoa(i)
					tags[s] = s
				}
				respAddTags, err := serviceResource.AddTags(getContext(), &pb.AddServiceTagsRequest{
					ServiceId: serviceId1,
					Tags:      tags,
				})
				Expect(err).To(BeNil())
				Expect(respAddTags.Response.GetCode()).To(Equal(pb.ResponseSuccess))

				By("tag is empty")
				respAddTags, _ = serviceResource.AddTags(getContext(), &pb.AddServiceTagsRequest{
					ServiceId: serviceId1,
					Tags:      map[string]string{},
				})
				Expect(respAddTags.Response.GetCode()).To(Equal(pb.ResponseSuccess))
				getServiceTagsResponse, _ := serviceResource.GetTags(getContext(), &pb.GetServiceTagsRequest{ServiceId: serviceId1})
				Expect(getServiceTagsResponse.Tags).To(BeEmpty())
			})
		})

		Context("when create tag out of gauge", func() {
			It("should be failed", func() {
				size := quota.DefaultTagQuota + 1
				tags := make(map[string]string, size)
				for i := 0; i < size; i++ {
					s := "tag" + strconv.Itoa(i)
					tags[s] = s
				}
				respAddTags, err := serviceResource.AddTags(getContext(), &pb.AddServiceTagsRequest{
					ServiceId: serviceId2,
					Tags:      tags,
				})
				Expect(err).To(BeNil())
				Expect(respAddTags.Response.GetCode()).To(Equal(pb.ErrInvalidParams))

				size = quota.DefaultTagQuota
				tags = make(map[string]string, size)
				for i := 0; i < size; i++ {
					s := "tag" + strconv.Itoa(i)
					tags[s] = s
				}
				respAddTags, err = serviceResource.AddTags(getContext(), &pb.AddServiceTagsRequest{
					ServiceId: serviceId2,
					Tags:      tags,
				})
				Expect(err).To(BeNil())
				Expect(respAddTags.Response.GetCode()).To(Equal(pb.ResponseSuccess))
			})
		})
	})

	Describe("execute 'get' operartion", func() {
		var (
			serviceId string
		)

		It("should be passed", func() {
			respCreateService, err := serviceResource.Create(getContext(), &pb.CreateServiceRequest{
				Service: &pb.MicroService{
					AppId:       "get_tag_group",
					ServiceName: "get_tag_service",
					Version:     "1.0.0",
					Level:       "FRONT",
					Status:      pb.MS_UP,
				},
			})
			Expect(err).To(BeNil())
			Expect(respCreateService.Response.GetCode()).To(Equal(pb.ResponseSuccess))
			serviceId = respCreateService.ServiceId

			respAddTags, err := serviceResource.AddTags(getContext(), &pb.AddServiceTagsRequest{
				ServiceId: serviceId,
				Tags: map[string]string{
					"a": "test",
					"b": "b",
				},
			})
			Expect(err).To(BeNil())
			Expect(respAddTags.Response.GetCode()).To(Equal(pb.ResponseSuccess))
		})

		Context("when request is invalid", func() {
			It("should be failed", func() {
				By("service does not exits")
				resp, _ := serviceResource.GetTags(getContext(), &pb.GetServiceTagsRequest{
					ServiceId: "noThisService",
				})
				Expect(resp.Response.GetCode()).ToNot(Equal(pb.ResponseSuccess))

				By("service id is empty")
				resp, _ = serviceResource.GetTags(getContext(), &pb.GetServiceTagsRequest{
					ServiceId: "",
				})
				Expect(resp.Response.GetCode()).ToNot(Equal(pb.ResponseSuccess))

				By("service id is invalid")
				resp, _ = serviceResource.GetTags(getContext(), &pb.GetServiceTagsRequest{
					ServiceId: TOO_LONG_SERVICEID,
				})
				Expect(resp.Response.GetCode()).ToNot(Equal(pb.ResponseSuccess))
			})
		})

		Context("when request is valid", func() {
			It("should be passed", func() {
				resp, err := serviceResource.GetTags(getContext(), &pb.GetServiceTagsRequest{
					ServiceId: serviceId,
				})
				Expect(err).To(BeNil())
				Expect(resp.Response.GetCode()).To(Equal(pb.ResponseSuccess))
				Expect(resp.Tags["a"]).To(Equal("test"))
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
					AppId:       "update_tag_group",
					ServiceName: "update_tag_service",
					Version:     "1.0.0",
					Level:       "FRONT",
					Status:      pb.MS_UP,
				},
			})
			Expect(err).To(BeNil())
			Expect(respCreateService.Response.GetCode()).To(Equal(pb.ResponseSuccess))
			serviceId = respCreateService.ServiceId

			respAddTags, err := serviceResource.AddTags(getContext(), &pb.AddServiceTagsRequest{
				ServiceId: serviceId,
				Tags: map[string]string{
					"a": "test",
					"b": "b",
				},
			})
			Expect(err).To(BeNil())
			Expect(respAddTags.Response.GetCode()).To(Equal(pb.ResponseSuccess))
		})

		Context("when request is invalid", func() {
			It("should be failed", func() {
				By("service id is empty")
				resp, err := serviceResource.UpdateTag(getContext(), &pb.UpdateServiceTagRequest{
					ServiceId: "",
					Key:       "a",
					Value:     "update",
				})
				Expect(err).To(BeNil())
				Expect(resp.Response.GetCode()).ToNot(Equal(pb.ResponseSuccess))

				By("service does not exits")
				resp, err = serviceResource.UpdateTag(getContext(), &pb.UpdateServiceTagRequest{
					ServiceId: "noneservice",
					Key:       "a",
					Value:     "update",
				})
				Expect(err).To(BeNil())
				Expect(resp.Response.GetCode()).ToNot(Equal(pb.ResponseSuccess))

				By("tag key is empty")
				resp, err = serviceResource.UpdateTag(getContext(), &pb.UpdateServiceTagRequest{
					ServiceId: serviceId,
					Key:       "",
					Value:     "update",
				})
				Expect(err).To(BeNil())
				Expect(resp.Response.GetCode()).ToNot(Equal(pb.ResponseSuccess))

				By("tag key does not exist")
				resp, err = serviceResource.UpdateTag(getContext(), &pb.UpdateServiceTagRequest{
					ServiceId: serviceId,
					Key:       "notexisttag",
					Value:     "update",
				})
				Expect(err).To(BeNil())
				Expect(resp.Response.GetCode()).ToNot(Equal(pb.ResponseSuccess))

				By("tag key is invalid")
				resp, err = serviceResource.UpdateTag(getContext(), &pb.UpdateServiceTagRequest{
					ServiceId: serviceId,
					Key:       TooLongTag,
					Value:     "v",
				})
				Expect(err).To(BeNil())
				Expect(resp.Response.GetCode()).ToNot(Equal(pb.ResponseSuccess))
			})
		})

		Context("when request is valid", func() {
			It("should be passed", func() {
				resp, err := serviceResource.UpdateTag(getContext(), &pb.UpdateServiceTagRequest{
					ServiceId: serviceId,
					Key:       "a",
					Value:     "update",
				})
				Expect(err).To(BeNil())
				Expect(resp.Response.GetCode()).To(Equal(pb.ResponseSuccess))
			})

		})

		Context("find instance, contain tag", func() {
			It("should pass", func() {
				By("create consumer")
				resp, err := serviceResource.Create(getContext(), &pb.CreateServiceRequest{
					Service: &pb.MicroService{
						AppId:       "find_inst_tag_group",
						ServiceName: "find_inst_tag_consumer",
						Version:     "1.0.0",
						Level:       "FRONT",
						Status:      pb.MS_UP,
					},
				})
				Expect(err).To(BeNil())
				Expect(resp.Response.GetCode()).To(Equal(pb.ResponseSuccess))
				consumerId := resp.ServiceId

				By("create provider")
				resp, err = serviceResource.Create(getContext(), &pb.CreateServiceRequest{
					Service: &pb.MicroService{
						AppId:       "find_inst_tag_group",
						ServiceName: "find_inst_tag_provider",
						Version:     "1.0.1",
						Level:       "FRONT",
						Status:      pb.MS_UP,
					},
				})
				Expect(err).To(BeNil())
				Expect(resp.Response.GetCode()).To(Equal(pb.ResponseSuccess))
				providerId := resp.ServiceId

				addTagResp, err := serviceResource.AddTags(getContext(), &pb.AddServiceTagsRequest{
					ServiceId: providerId,
					Tags:      map[string]string{"filter_tag": "filter"},
				})
				Expect(err).To(BeNil())
				Expect(addTagResp.Response.GetCode()).To(Equal(pb.ResponseSuccess))

				instanceResp, err := disco.RegisterInstance(getContext(), &pb.RegisterInstanceRequest{
					Instance: &pb.MicroServiceInstance{
						ServiceId: providerId,
						Endpoints: []string{
							"findInstanceForTagFilter:127.0.0.1:8080",
						},
						HostName: "UT-HOST",
						Status:   pb.MSI_UP,
					},
				})
				Expect(err).To(BeNil())
				Expect(instanceResp.Response.GetCode()).To(Equal(pb.ResponseSuccess))

				findResp, err := disco.FindInstances(getContext(), &pb.FindInstancesRequest{
					ConsumerServiceId: consumerId,
					AppId:             "find_inst_tag_group",
					ServiceName:       "find_inst_tag_provider",
					Tags:              []string{"not-exist-tag"},
				})
				Expect(err).To(BeNil())
				Expect(findResp.Response.GetCode()).To(Equal(pb.ResponseSuccess))
				Expect(len(findResp.Instances)).To(Equal(0))

				findResp, err = disco.FindInstances(getContext(), &pb.FindInstancesRequest{
					ConsumerServiceId: consumerId,
					AppId:             "find_inst_tag_group",
					ServiceName:       "find_inst_tag_provider",
					Tags:              []string{"filter_tag"},
				})
				Expect(err).To(BeNil())
				Expect(findResp.Response.GetCode()).To(Equal(pb.ResponseSuccess))
				Expect(findResp.Instances[0].InstanceId).To(Equal(instanceResp.InstanceId))
			})
		})

	})

	Describe("execute 'delete' operation", func() {
		var (
			serviceId string
		)

		It("should be passed", func() {
			respCreateService, err := serviceResource.Create(getContext(), &pb.CreateServiceRequest{
				Service: &pb.MicroService{
					AppId:       "delete_tag_group",
					ServiceName: "delete_tag_service",
					Version:     "1.0.0",
					Level:       "FRONT",
					Status:      pb.MS_UP,
				},
			})
			Expect(err).To(BeNil())
			Expect(respCreateService.Response.GetCode()).To(Equal(pb.ResponseSuccess))
			serviceId = respCreateService.ServiceId

			respAddTags, err := serviceResource.AddTags(getContext(), &pb.AddServiceTagsRequest{
				ServiceId: serviceId,
				Tags: map[string]string{
					"a": "test",
					"b": "b",
				},
			})
			Expect(err).To(BeNil())
			Expect(respAddTags.Response.GetCode()).To(Equal(pb.ResponseSuccess))
		})

		Context("when request is invalid", func() {
			It("should be failed", func() {
				By("service id is empty")
				respAddTags, err := serviceResource.DeleteTags(getContext(), &pb.DeleteServiceTagsRequest{
					ServiceId: "",
					Keys:      []string{"a", "b"},
				})
				fmt.Println(err)
				fmt.Println(respAddTags.Response)
				Expect(err).To(BeNil())
				Expect(respAddTags.Response.GetCode()).To(Equal(pb.ErrInvalidParams))

				By("service does not exits")
				respAddTags, err = serviceResource.DeleteTags(getContext(), &pb.DeleteServiceTagsRequest{
					ServiceId: "noneservice",
					Keys:      []string{"a", "b"},
				})
				Expect(err).To(BeNil())
				Expect(respAddTags.Response.GetCode()).To(Equal(pb.ErrServiceNotExists))

				By("tag key does not exits")
				respAddTags, err = serviceResource.DeleteTags(getContext(), &pb.DeleteServiceTagsRequest{
					ServiceId: serviceId,
					Keys:      []string{"c"},
				})
				Expect(err).To(BeNil())
				Expect(respAddTags.Response.GetCode()).To(Equal(pb.ErrTagNotExists))

				By("tag key is empty")
				respAddTags, err = serviceResource.DeleteTags(getContext(), &pb.DeleteServiceTagsRequest{
					ServiceId: serviceId,
					Keys:      []string{""},
				})
				Expect(err).To(BeNil())
				Expect(respAddTags.Response.GetCode()).To(Equal(pb.ErrInvalidParams))
				respAddTags, err = serviceResource.DeleteTags(getContext(), &pb.DeleteServiceTagsRequest{
					ServiceId: serviceId,
				})
				Expect(err).To(BeNil())
				Expect(respAddTags.Response.GetCode()).To(Equal(pb.ErrInvalidParams))

				By("tag key is invalid")
				respAddTags, err = serviceResource.DeleteTags(getContext(), &pb.DeleteServiceTagsRequest{
					ServiceId: serviceId,
					Keys:      []string{TooLongTag},
				})
				Expect(err).To(BeNil())
				Expect(respAddTags.Response.GetCode()).To(Equal(pb.ErrInvalidParams))

				var arr []string
				for i := 0; i < quota.DefaultTagQuota+1; i++ {
					arr = append(arr, strconv.Itoa(i))
				}
				respAddTags, err = serviceResource.DeleteTags(getContext(), &pb.DeleteServiceTagsRequest{
					ServiceId: serviceId,
					Keys:      arr,
				})
				Expect(err).To(BeNil())
				Expect(respAddTags.Response.GetCode()).To(Equal(pb.ErrInvalidParams))
			})
		})

		Context("when request is valid", func() {
			It("should be passed", func() {
				respAddTags, err := serviceResource.DeleteTags(getContext(), &pb.DeleteServiceTagsRequest{
					ServiceId: serviceId,
					Keys:      []string{"a", "b"},
				})
				Expect(err).To(BeNil())
				Expect(respAddTags.Response.GetCode()).To(Equal(pb.ResponseSuccess))

				resp, err := serviceResource.GetTags(getContext(), &pb.GetServiceTagsRequest{
					ServiceId: serviceId,
				})
				Expect(err).To(BeNil())
				Expect(resp.Response.GetCode()).To(Equal(pb.ResponseSuccess))
				Expect(resp.Tags["a"]).To(Equal(""))
			})
		})
	})
})
