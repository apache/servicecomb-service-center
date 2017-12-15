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
package govern_test

import (
	pb "github.com/ServiceComb/service-center/server/core/proto"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("'Govern' service", func() {
	Describe("execute 'get all' operation", func() {
		Context("when get all services", func() {
			It("should be passed", func() {
				By("all options")
				resp, err := governService.GetServicesInfo(getContext(), &pb.GetServicesInfoRequest{
					Options: []string{"all"},
				})
				Expect(err).To(BeNil())
				Expect(resp.Response.Code).To(Equal(pb.Response_SUCCESS))

				By("only service metadata")
				resp, err = governService.GetServicesInfo(getContext(), &pb.GetServicesInfoRequest{
					Options: []string{""},
				})
				Expect(err).To(BeNil())
				Expect(resp.Response.Code).To(Equal(pb.Response_SUCCESS))

				By("custom options")
				resp, err = governService.GetServicesInfo(getContext(), &pb.GetServicesInfoRequest{
					Options: []string{"tags", "rules", "instances", "schemas", "statistics"},
				})
				Expect(err).To(BeNil())
				Expect(resp.Response.Code).To(Equal(pb.Response_SUCCESS))

				By("'statistics' option")
				resp, err = governService.GetServicesInfo(getContext(), &pb.GetServicesInfoRequest{
					Options: []string{"statistics"},
				})
				Expect(err).To(BeNil())
				Expect(resp.Response.Code).To(Equal(pb.Response_SUCCESS))

				By("get instance count")
				resp, err = governService.GetServicesInfo(getContext(), &pb.GetServicesInfoRequest{
					Options: []string{"instances"},
					CountOnly: true,
				})
				Expect(err).To(BeNil())
				Expect(resp.Response.Code).To(Equal(pb.Response_SUCCESS))
			})
		})
	})

	Describe("execute 'get detail' operation", func() {
		var (
			serviceId string
		)

		It("should be passed", func() {
			resp, err := serviceResource.Create(getContext(), &pb.CreateServiceRequest{
				Service: &pb.MicroService{
					AppId:       "govern_service_group",
					ServiceName: "govern_service_name",
					Version:     "3.0.0",
					Level:       "FRONT",
					Status:      pb.MS_UP,
				},
			})
			Expect(err).To(BeNil())
			Expect(resp.Response.Code).To(Equal(pb.Response_SUCCESS))
			serviceId = resp.ServiceId

			serviceResource.ModifySchema(getContext(), &pb.ModifySchemaRequest{
				ServiceId: serviceId,
				SchemaId:  "schemaId",
				Schema:    "detail",
			})
			Expect(err).To(BeNil())
			Expect(resp.Response.Code).To(Equal(pb.Response_SUCCESS))

			instanceResource.Register(getContext(), &pb.RegisterInstanceRequest{
				Instance: &pb.MicroServiceInstance{
					ServiceId: serviceId,
					Endpoints: []string{
						"govern:127.0.0.1:8080",
					},
					HostName: "UT-HOST",
					Status:   pb.MSI_UP,
				},
			})
			Expect(err).To(BeNil())
			Expect(resp.Response.Code).To(Equal(pb.Response_SUCCESS))
		})

		Context("when get invalid service detail", func() {
			It("should be failed", func() {
				resp, err := governService.GetServiceDetail(getContext(), &pb.GetServiceRequest{
					ServiceId: "",
				})
				Expect(err).To(BeNil())
				Expect(resp.Response.Code).ToNot(Equal(pb.Response_SUCCESS))
			})
		})

		Context("when get a service detail", func() {
			It("should be passed", func() {
				respGetServiceDetail, err := governService.GetServiceDetail(getContext(), &pb.GetServiceRequest{
					ServiceId: serviceId,
				})
				Expect(err).To(BeNil())
				Expect(respGetServiceDetail.Response.Code).To(Equal(pb.Response_SUCCESS))

				respDelete, err := serviceResource.Delete(getContext(), &pb.DeleteServiceRequest{
					ServiceId: serviceId,
					Force:     true,
				})
				Expect(err).To(BeNil())
				Expect(respDelete.Response.Code).To(Equal(pb.Response_SUCCESS))

				respGetServiceDetail, err = governService.GetServiceDetail(getContext(), &pb.GetServiceRequest{
					ServiceId: serviceId,
				})
				Expect(err).To(BeNil())
				Expect(respGetServiceDetail.Response.Code).ToNot(Equal(pb.Response_SUCCESS))
			})
		})
	})

	Describe("execute 'get apps' operation", func() {
		Context("when request is invalid", func() {
			It("should be failed", func() {
				resp, err := governService.GetApplications(getContext(), &pb.GetAppsRequest{
					Environment: "non-exist-env",
				})
				Expect(err).To(BeNil())
				Expect(resp.Response.Code).ToNot(Equal(pb.Response_SUCCESS))
			})
		})

		Context("when request is valid", func() {
			It("should be passed", func() {
				resp, err := governService.GetApplications(getContext(), &pb.GetAppsRequest{})
				Expect(err).To(BeNil())
				Expect(resp.Response.Code).To(Equal(pb.Response_SUCCESS))

				resp, err = governService.GetApplications(getContext(), &pb.GetAppsRequest{
					Environment: pb.ENV_ACCEPT,
				})
				Expect(err).To(BeNil())
				Expect(resp.Response.Code).To(Equal(pb.Response_SUCCESS))
			})
		})
	})
})
