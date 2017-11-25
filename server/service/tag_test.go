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
	pb "github.com/ServiceComb/service-center/server/core/proto"
	"github.com/ServiceComb/service-center/server/plugin/infra/quota/buildin"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"strconv"
)

var _ = Describe("'Tag' service", func() {
	Describe("execute 'create' operartion", func() {
		var (
			serviceId string
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
			Expect(respCreateService.GetResponse().Code).To(Equal(pb.Response_SUCCESS))
			serviceId = respCreateService.ServiceId
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
				Expect(respAddTags.GetResponse().Code).ToNot(Equal(pb.Response_SUCCESS))

				By("service does not exist")
				respAddTags, _ = serviceResource.AddTags(getContext(), &pb.AddServiceTagsRequest{
					ServiceId: "noServiceTest",
					Tags: map[string]string{
						"a": "test",
					},
				})
				Expect(respAddTags.GetResponse().Code).ToNot(Equal(pb.Response_SUCCESS))

				By("tag is empty")
				respAddTags, _ = serviceResource.AddTags(getContext(), &pb.AddServiceTagsRequest{
					ServiceId: serviceId,
					Tags:      map[string]string{},
				})
				Expect(respAddTags.GetResponse().Code).ToNot(Equal(pb.Response_SUCCESS))

				By("tag key is empty")
				respAddTags, _ = serviceResource.AddTags(getContext(), &pb.AddServiceTagsRequest{
					ServiceId: serviceId,
					Tags: map[string]string{
						"": "value",
					},
				})
				Expect(respAddTags.GetResponse().Code).ToNot(Equal(pb.Response_SUCCESS))
			})
		})

		Context("when request is valid", func() {
			It("should be passed", func() {
				respAddTags, err := serviceResource.AddTags(getContext(), &pb.AddServiceTagsRequest{
					ServiceId: serviceId,
					Tags: map[string]string{
						"a": "test",
						"b": "b",
					},
				})
				Expect(err).To(BeNil())
				Expect(respAddTags.GetResponse().Code).To(Equal(pb.Response_SUCCESS))
			})
		})

		Context("when create tag out of gauge", func() {
			It("should be failed", func() {
				size := buildin.TAG_MAX_NUM_FOR_ONESERVICE + 1
				tags := make(map[string]string, size)
				for i := 0; i < size; i++ {
					s := strconv.Itoa(i)
					tags[s] = s
				}
				respAddTags, _ := serviceResource.AddTags(getContext(), &pb.AddServiceTagsRequest{
					ServiceId: serviceId,
					Tags:      tags,
				})
				Expect(respAddTags.GetResponse().Code).ToNot(Equal(pb.Response_SUCCESS))
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
			Expect(respCreateService.GetResponse().Code).To(Equal(pb.Response_SUCCESS))
			serviceId = respCreateService.ServiceId

			respAddTags, err := serviceResource.AddTags(getContext(), &pb.AddServiceTagsRequest{
				ServiceId: serviceId,
				Tags: map[string]string{
					"a": "test",
					"b": "b",
				},
			})
			Expect(err).To(BeNil())
			Expect(respAddTags.GetResponse().Code).To(Equal(pb.Response_SUCCESS))
		})

		Context("when request is invalid", func() {
			It("should be failed", func() {
				By("service does not exits")
				resp, _ := serviceResource.GetTags(getContext(), &pb.GetServiceTagsRequest{
					ServiceId: "noThisService",
				})
				Expect(resp.GetResponse().Code).ToNot(Equal(pb.Response_SUCCESS))

				By("service id is empty")
				resp, _ = serviceResource.GetTags(getContext(), &pb.GetServiceTagsRequest{
					ServiceId: "",
				})
				Expect(resp.GetResponse().Code).ToNot(Equal(pb.Response_SUCCESS))
			})
		})

		Context("when request is valid", func() {
			It("should be passed", func() {
				resp, err := serviceResource.GetTags(getContext(), &pb.GetServiceTagsRequest{
					ServiceId: serviceId,
				})
				Expect(err).To(BeNil())
				Expect(resp.GetResponse().Code).To(Equal(pb.Response_SUCCESS))
				Expect(resp.Tags["a"]).To(Equal("test"))
			})
		})

	})

	Describe("execute 'update' operartion", func() {
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
			Expect(respCreateService.GetResponse().Code).To(Equal(pb.Response_SUCCESS))
			serviceId = respCreateService.ServiceId

			respAddTags, err := serviceResource.AddTags(getContext(), &pb.AddServiceTagsRequest{
				ServiceId: serviceId,
				Tags: map[string]string{
					"a": "test",
					"b": "b",
				},
			})
			Expect(err).To(BeNil())
			Expect(respAddTags.GetResponse().Code).To(Equal(pb.Response_SUCCESS))
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
				Expect(resp.GetResponse().Code).ToNot(Equal(pb.Response_SUCCESS))

				By("service does not exits")
				resp, err = serviceResource.UpdateTag(getContext(), &pb.UpdateServiceTagRequest{
					ServiceId: "noneservice",
					Key:       "a",
					Value:     "update",
				})
				Expect(err).To(BeNil())
				Expect(resp.GetResponse().Code).ToNot(Equal(pb.Response_SUCCESS))

				By("tag key is empty")
				resp, err = serviceResource.UpdateTag(getContext(), &pb.UpdateServiceTagRequest{
					ServiceId: serviceId,
					Key:       "",
					Value:     "update",
				})
				Expect(err).To(BeNil())
				Expect(resp.GetResponse().Code).ToNot(Equal(pb.Response_SUCCESS))
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
				Expect(resp.GetResponse().Code).To(Equal(pb.Response_SUCCESS))
			})

		})

	})

	Describe("execute 'delete' operartion", func() {
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
			Expect(respCreateService.GetResponse().Code).To(Equal(pb.Response_SUCCESS))
			serviceId = respCreateService.ServiceId

			respAddTags, err := serviceResource.AddTags(getContext(), &pb.AddServiceTagsRequest{
				ServiceId: serviceId,
				Tags: map[string]string{
					"a": "test",
					"b": "b",
				},
			})
			Expect(err).To(BeNil())
			Expect(respAddTags.GetResponse().Code).To(Equal(pb.Response_SUCCESS))
		})

		Context("when request is invalid", func() {
			It("should be failed", func() {
				By("service id is empty")
				respAddTags, err := serviceResource.DeleteTags(getContext(), &pb.DeleteServiceTagsRequest{
					ServiceId: "",
					Keys:      []string{"a", "b"},
				})
				Expect(err).To(BeNil())
				Expect(respAddTags.GetResponse().Code).ToNot(Equal(pb.Response_SUCCESS))

				By("service does not exits")
				respAddTags, err = serviceResource.DeleteTags(getContext(), &pb.DeleteServiceTagsRequest{
					ServiceId: "noneservice",
					Keys:      []string{"a", "b"},
				})
				Expect(err).To(BeNil())
				Expect(respAddTags.GetResponse().Code).ToNot(Equal(pb.Response_SUCCESS))

				By("tag key does not exits")
				respAddTags, err = serviceResource.DeleteTags(getContext(), &pb.DeleteServiceTagsRequest{
					ServiceId: serviceId,
					Keys:      []string{"c"},
				})
				Expect(err).To(BeNil())
				Expect(respAddTags.GetResponse().Code).ToNot(Equal(pb.Response_SUCCESS))

				By("tag key is empty")
				respAddTags, err = serviceResource.DeleteTags(getContext(), &pb.DeleteServiceTagsRequest{
					ServiceId: serviceId,
					Keys:      []string{""},
				})
				Expect(err).To(BeNil())
				Expect(respAddTags.GetResponse().Code).ToNot(Equal(pb.Response_SUCCESS))
			})
		})

		Context("when request is valid", func() {
			It("should be passed", func() {
				respAddTags, err := serviceResource.DeleteTags(getContext(), &pb.DeleteServiceTagsRequest{
					ServiceId: serviceId,
					Keys:      []string{"a", "b"},
				})
				Expect(err).To(BeNil())
				Expect(respAddTags.GetResponse().Code).To(Equal(pb.Response_SUCCESS))

				resp, err := serviceResource.GetTags(getContext(), &pb.GetServiceTagsRequest{
					ServiceId: serviceId,
				})
				Expect(err).To(BeNil())
				Expect(resp.GetResponse().Code).To(Equal(pb.Response_SUCCESS))
				Expect(resp.Tags["a"]).To(Equal(""))
			})
		})
	})
})
