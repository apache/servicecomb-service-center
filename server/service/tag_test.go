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
	pb "github.com/servicecomb/service-center/server/core/proto"
)

var serviceId string
var serviceId2 string
var _ = Describe("ServiceController", func() {
	Describe("tag", func() {
		Context("normal", func() {
			It("创建tag", func() {
				fmt.Println("UT===========创建tag")
				resp, err := serviceResource.Create(getContext(), &pb.CreateServiceRequest{
					Service: &pb.MicroService{
						ServiceName: "service_name_consumer",
						AppId:       "service_group_consumer",
						Version:     "4.0.0",
						Level:       "FRONT",
						Schemas: []string{
							"xxxxxxxx",
						},
						Status: "UP",
					},
				})

				Expect(err).To(BeNil())
				serviceId = resp.ServiceId
				Expect(resp.GetResponse().Code).To(Equal(pb.Response_SUCCESS))

				resp, err = serviceResource.Create(getContext(), &pb.CreateServiceRequest{
					Service: &pb.MicroService{
						ServiceName: "service_name_consumer1",
						AppId:       "service_group_consumer2",
						Version:     "4.0.0",
						Level:       "FRONT",
						Schemas: []string{
							"xxxxxxxx",
						},
						Status: "UP",
					},
				})

				Expect(err).To(BeNil())
				serviceId2 = resp.ServiceId
				Expect(resp.GetResponse().Code).To(Equal(pb.Response_SUCCESS))

				respAddTags, err := serviceResource.AddTags(getContext(), &pb.AddServiceTagsRequest{
					ServiceId: serviceId,
					Tags: map[string]string{
						"a": "test",
					},
				})
				Expect(err).To(BeNil())
				Expect(respAddTags.GetResponse().Code).To(Equal(pb.Response_SUCCESS))

				respAddTags, err = serviceResource.AddTags(getContext(), &pb.AddServiceTagsRequest{
					ServiceId: serviceId,
					Tags: map[string]string{
						"a": "test",
						"b": "b",
					},
				})
				Expect(err).To(BeNil())
				Expect(respAddTags.GetResponse().Code).To(Equal(pb.Response_SUCCESS))
			})

			It("创建tag，参数校验", func() {
				fmt.Println("UT===========创建tag，参数校验")
				respAddTags, _ := serviceResource.AddTags(getContext(), &pb.AddServiceTagsRequest{
					ServiceId: "",
					Tags: map[string]string{
						"a": "test",
					},
				})
				Expect(respAddTags.GetResponse().Code).To(Equal(pb.Response_FAIL))

				respAddTags, _ = serviceResource.AddTags(getContext(), &pb.AddServiceTagsRequest{
					ServiceId: "noServiceTest",
					Tags: map[string]string{
						"a": "test",
					},
				})
				Expect(respAddTags.GetResponse().Code).To(Equal(pb.Response_FAIL))

				respAddTags, _ = serviceResource.AddTags(getContext(), &pb.AddServiceTagsRequest{
					ServiceId: serviceId,
					Tags:      map[string]string{},
				})
				Expect(respAddTags.GetResponse().Code).To(Equal(pb.Response_FAIL))

				respAddTags, _ = serviceResource.AddTags(getContext(), &pb.AddServiceTagsRequest{
					ServiceId: TOO_LONG_SERVICEID,
					Tags: map[string]string{
						"a": "b",
					},
				})
				Expect(respAddTags.GetResponse().Code).To(Equal(pb.Response_FAIL))

				respAddTags, _ = serviceResource.AddTags(getContext(), &pb.AddServiceTagsRequest{
					ServiceId: serviceId,
					Tags: map[string]string{
						"": "value",
					},
				})
				Expect(respAddTags.GetResponse().Code).To(Equal(pb.Response_FAIL))
			})
			It("获取tag，参数校验", func() {
				resp, _ := serviceResource.GetTags(getContext(), &pb.GetServiceTagsRequest{
					ServiceId: "noThisService",
				})
				Expect(resp.GetResponse().Code).To(Equal(pb.Response_FAIL))

				resp, _ = serviceResource.GetTags(getContext(), &pb.GetServiceTagsRequest{
					ServiceId: "",
				})
				Expect(resp.GetResponse().Code).To(Equal(pb.Response_FAIL))

				resp, _ = serviceResource.GetTags(getContext(), &pb.GetServiceTagsRequest{
					ServiceId: TOO_LONG_SERVICEID,
				})
				Expect(resp.GetResponse().Code).To(Equal(pb.Response_FAIL))
			})

			It("获取tag", func() {
				resp, err := serviceResource.GetTags(getContext(), &pb.GetServiceTagsRequest{
					ServiceId: serviceId,
				})
				Expect(err).To(BeNil())
				Expect(resp.GetResponse().Code).To(Equal(pb.Response_SUCCESS))
			})

			It("修改tag,参数校验", func() {
				fmt.Println("UT===========修改tag，参数校验")
				resp, err := serviceResource.UpdateTag(getContext(), &pb.UpdateServiceTagRequest{
					ServiceId: "",
					Key:       "a",
					Value:     "update",
				})
				Expect(err).To(BeNil())
				Expect(resp.GetResponse().Code).To(Equal(pb.Response_FAIL))

				resp, err = serviceResource.UpdateTag(getContext(), &pb.UpdateServiceTagRequest{
					ServiceId: "noneservice",
					Key:       "a",
					Value:     "update",
				})
				Expect(err).To(BeNil())
				Expect(resp.GetResponse().Code).To(Equal(pb.Response_FAIL))

				resp, err = serviceResource.UpdateTag(getContext(), &pb.UpdateServiceTagRequest{
					ServiceId: TOO_LONG_SERVICEID,
					Key:       "a",
					Value:     "update",
				})
				Expect(err).To(BeNil())
				Expect(resp.GetResponse().Code).To(Equal(pb.Response_FAIL))
			})

			It("修改tag", func() {
				fmt.Println("UT===========修改tag")
				resp, err := serviceResource.UpdateTag(getContext(), &pb.UpdateServiceTagRequest{
					ServiceId: serviceId,
					Key:       "a",
					Value:     "update",
				})
				Expect(err).To(BeNil())
				Expect(resp.GetResponse().Code).To(Equal(pb.Response_SUCCESS))

				resp, err = serviceResource.UpdateTag(getContext(), &pb.UpdateServiceTagRequest{
					ServiceId: serviceId,
					Key:       "notExistKey",
					Value:     "update",
				})
				Expect(err).To(BeNil())
				Expect(resp.GetResponse().Code).To(Equal(pb.Response_FAIL))

				resp, err = serviceResource.UpdateTag(getContext(), &pb.UpdateServiceTagRequest{
					ServiceId: TOO_LONG_SERVICEID,
					Key:       "",
					Value:     "update",
				})
				Expect(err).To(BeNil())
				Expect(resp.GetResponse().Code).To(Equal(pb.Response_FAIL))
			})

			It("删除tag,参数校验", func() {
				fmt.Println("UT===========删除tag,参数校验，serviceId为空")
				respAddTags, err := serviceResource.DeleteTags(getContext(), &pb.DeleteServiceTagsRequest{
					ServiceId: "",
					Keys:      []string{"a", "b"},
				})
				Expect(err).To(BeNil())
				Expect(respAddTags.GetResponse().Code).To(Equal(pb.Response_FAIL))

				fmt.Println("UT===========删除tag,参数校验，serviceId为空")
				respAddTags, err = serviceResource.DeleteTags(getContext(), &pb.DeleteServiceTagsRequest{
					ServiceId: "noneservice",
					Keys:      []string{"a", "b"},
				})
				Expect(err).To(BeNil())
				Expect(respAddTags.GetResponse().Code).To(Equal(pb.Response_FAIL))

				respAddTags, err = serviceResource.DeleteTags(getContext(), &pb.DeleteServiceTagsRequest{
					ServiceId: serviceId2,
					Keys:      []string{"a", "b"},
				})
				Expect(err).To(BeNil())
				Expect(respAddTags.GetResponse().Code).To(Equal(pb.Response_FAIL))

				respAddTags, err = serviceResource.DeleteTags(getContext(), &pb.DeleteServiceTagsRequest{
					ServiceId: serviceId,
					Keys:      []string{"c"},
				})
				Expect(err).To(BeNil())
				Expect(respAddTags.GetResponse().Code).To(Equal(pb.Response_FAIL))

				respAddTags, err = serviceResource.DeleteTags(getContext(), &pb.DeleteServiceTagsRequest{
					ServiceId: TOO_LONG_SERVICEID,
					Keys:      []string{""},
				})
				Expect(err).To(BeNil())
				Expect(respAddTags.GetResponse().Code).To(Equal(pb.Response_FAIL))
			})

			It("删除tag", func() {
				fmt.Println("UT===========删除tag")
				respAddTags, err := serviceResource.DeleteTags(getContext(), &pb.DeleteServiceTagsRequest{
					ServiceId: serviceId,
					Keys:      []string{"a", "b"},
				})
				Expect(err).To(BeNil())
				Expect(respAddTags.GetResponse().Code).To(Equal(pb.Response_SUCCESS))

				respDel, err := serviceResource.Delete(getContext(), &pb.DeleteServiceRequest{
					ServiceId: serviceId,
					Force:     true,
				})
				Expect(err).To(BeNil())
				Expect(respDel.GetResponse().Code).To(Equal(pb.Response_SUCCESS))

				respDel, err = serviceResource.Delete(getContext(), &pb.DeleteServiceRequest{
					ServiceId: serviceId2,
					Force:     true,
				})
				Expect(err).To(BeNil())
				Expect(respDel.GetResponse().Code).To(Equal(pb.Response_SUCCESS))
			})
		})
	})
})
