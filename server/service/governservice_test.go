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

var _ = Describe("GovernServiceController", func() {
	Describe("GetServicesInfo", func() {
		Context("normal", func() {
			It("服务治理获取所有服务信息", func() {
				fmt.Println("UT===========服务治理获取所有服务信息")

				resp, err := governService.GetServicesInfo(getContext(), &pb.GetServicesInfoRequest{
					Options: []string{"all"},
				})

				Expect(err).To(BeNil())
				Expect(resp.GetResponse().Code).To(Equal(pb.Response_SUCCESS))

				resp, err = governService.GetServicesInfo(getContext(), &pb.GetServicesInfoRequest{
					Options: []string{""},
				})

				Expect(err).To(BeNil())
				Expect(resp.GetResponse().Code).To(Equal(pb.Response_SUCCESS))

				resp, err = governService.GetServicesInfo(getContext(), &pb.GetServicesInfoRequest{
					Options: []string{"tags", "rules", "instances", "schemas"},
				})

				Expect(err).To(BeNil())
				Expect(resp.GetResponse().Code).To(Equal(pb.Response_SUCCESS))
			})
		})
	})
	Describe("GetServiceDetail", func() {
		Context("normal", func() {
			It("服务治理获取单个服务信息,参数校验", func() {
				fmt.Println("UT===========服务治理获取单个服务信息,参数校验")
				resp, err := governService.GetServiceDetail(getContext(), &pb.GetServiceRequest{
					ServiceId: "",
				})

				Expect(err).To(BeNil())
				Expect(resp.GetResponse().Code).To(Equal(pb.Response_FAIL))
			})

			It("服务治理获取单个服务信息", func() {
				resp, err := serviceResource.Create(getContext(), &pb.CreateServiceRequest{
					Service: &pb.MicroService{
						ServiceName: "service_name_govern",
						AppId:       "service_group_govern",
						Version:     "3.0.0",
						Level:       "FRONT",
						Schemas: []string{
							"schemaId",
						},
						Status: "UP",
					},
				})

				Expect(err).To(BeNil())
				serviceId := resp.ServiceId
				Expect(resp.GetResponse().Code).To(Equal(pb.Response_SUCCESS))
				fmt.Println("UT===========服务治理获取单个服务信息")

				serviceResource.ModifySchema(getContext(), &pb.ModifySchemaRequest{
					ServiceId: serviceId,
					SchemaId:  "schemaId",
					Schema:    "detail",
				})
				Expect(err).To(BeNil())
				Expect(resp.GetResponse().Code).To(Equal(pb.Response_SUCCESS))

				insResource.Register(getContext(), &pb.RegisterInstanceRequest{
					Instance: &pb.MicroServiceInstance{
						ServiceId: serviceId,
						Endpoints: []string{
							"rest:127.0.0.1:8080",
						},
						HostName:    "UT-HOST",
						Status:      pb.MSI_UP,
						Environment: "production",
					},
				})
				Expect(err).To(BeNil())
				Expect(resp.GetResponse().Code).To(Equal(pb.Response_SUCCESS))

				respGetServiceDetail, err := governService.GetServiceDetail(getContext(), &pb.GetServiceRequest{
					ServiceId: serviceId,
				})

				Expect(err).To(BeNil())
				Expect(respGetServiceDetail.GetResponse().Code).To(Equal(pb.Response_SUCCESS))

				respDelete, err := serviceResource.Delete(getContext(), &pb.DeleteServiceRequest{
					ServiceId: serviceId,
					Force:     true,
				})

				Expect(err).To(BeNil())
				Expect(respDelete.GetResponse().Code).To(Equal(pb.Response_SUCCESS))

				respGetServiceDetail, err = governService.GetServiceDetail(getContext(), &pb.GetServiceRequest{
					ServiceId: serviceId,
				})

				Expect(respGetServiceDetail.GetResponse().Code).To(Equal(pb.Response_FAIL))
			})
		})
	})
})
