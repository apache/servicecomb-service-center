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
	"context"
	"testing"

	"github.com/onsi/ginkgo/reporters"

	_ "github.com/apache/servicecomb-service-center/test"

	"github.com/apache/servicecomb-service-center/pkg/util"
	"github.com/apache/servicecomb-service-center/server/core"
	"github.com/apache/servicecomb-service-center/server/service/disco"
	"github.com/apache/servicecomb-service-center/server/service/govern"
	pb "github.com/go-chassis/cari/discovery"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestGovern(t *testing.T) {
	RegisterFailHandler(Fail)
	junitReporter := reporters.NewJUnitReporter("model.junit.xml")
	RunSpecsWithDefaultAndCustomReporters(t, "model Suite", []Reporter{junitReporter})
}

func getContext() context.Context {
	return util.WithNoCache(util.SetDomainProject(context.Background(), "default", "default"))
}

var _ = Describe("'Govern' service", func() {
	Describe("execute 'get all' operation", func() {
		Context("when get all services", func() {
			It("should be passed", func() {
				By("all options")
				resp, err := govern.ListServiceDetail(getContext(), &pb.GetServicesInfoRequest{
					Options: []string{"all"},
				})
				Expect(err).To(BeNil())
				Expect(resp).ToNot(BeNil())

				By("only service metadata")
				resp, err = govern.ListServiceDetail(getContext(), &pb.GetServicesInfoRequest{
					Options: []string{""},
				})
				Expect(err).To(BeNil())
				Expect(resp).ToNot(BeNil())

				By("custom options")
				resp, err = govern.ListServiceDetail(getContext(), &pb.GetServicesInfoRequest{
					Options: []string{"tags", "rules", "instances", "schemas", "statistics"},
				})
				Expect(err).To(BeNil())
				Expect(resp).ToNot(BeNil())

				By("'statistics' option")
				resp, err = govern.ListServiceDetail(getContext(), &pb.GetServicesInfoRequest{
					Options: []string{"statistics"},
				})
				Expect(err).To(BeNil())
				Expect(resp).ToNot(BeNil())

				By("get instance count")
				resp, err = govern.ListServiceDetail(getContext(), &pb.GetServicesInfoRequest{
					Options:   []string{"instances"},
					CountOnly: true,
				})
				Expect(err).To(BeNil())
				Expect(resp).ToNot(BeNil())
			})
		})

		Context("when get top graph", func() {
			It("should be passed", func() {
				respC, err := core.ServiceAPI.Create(getContext(), &pb.CreateServiceRequest{
					Service: &pb.MicroService{
						AppId:       "govern_service_group",
						ServiceName: "govern_service_graph",
						Version:     "1.0.0",
						Level:       "FRONT",
						Status:      pb.MS_UP,
					},
				})
				Expect(err).To(BeNil())
				Expect(respC.Response.GetCode()).To(Equal(pb.ResponseSuccess))

				graph, err := govern.Draw(getContext(), false)
				Expect(err).To(BeNil())
				Expect(len(graph.Nodes)).ToNot(Equal(0))
			})
		})
	})

	Describe("execute 'get detail' operation", func() {
		var (
			serviceId string
		)

		It("should be passed", func() {
			resp, err := core.ServiceAPI.Create(getContext(), &pb.CreateServiceRequest{
				Service: &pb.MicroService{
					AppId:       "govern_service_group",
					ServiceName: "govern_service_name",
					Version:     "3.0.0",
					Level:       "FRONT",
					Status:      pb.MS_UP,
				},
			})
			Expect(err).To(BeNil())
			Expect(resp.Response.GetCode()).To(Equal(pb.ResponseSuccess))
			serviceId = resp.ServiceId

			core.ServiceAPI.ModifySchema(getContext(), &pb.ModifySchemaRequest{
				ServiceId: serviceId,
				SchemaId:  "schemaId",
				Schema:    "detail",
			})
			Expect(err).To(BeNil())
			Expect(resp.Response.GetCode()).To(Equal(pb.ResponseSuccess))

			disco.RegisterInstance(getContext(), &pb.RegisterInstanceRequest{
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
			Expect(resp.Response.GetCode()).To(Equal(pb.ResponseSuccess))
		})

		Context("when get invalid service detail", func() {
			It("should be failed", func() {
				resp, err := govern.GetServiceDetail(getContext(), &pb.GetServiceRequest{
					ServiceId: "",
				})
				Expect(err).ToNot(BeNil())
				Expect(resp).To(BeNil())
			})
		})

		Context("when get a service detail", func() {
			It("should be passed", func() {
				respGetServiceDetail, err := govern.GetServiceDetail(getContext(), &pb.GetServiceRequest{
					ServiceId: serviceId,
				})
				Expect(err).To(BeNil())
				Expect(respGetServiceDetail).ToNot(BeNil())

				respDelete, err := core.ServiceAPI.Delete(getContext(), &pb.DeleteServiceRequest{
					ServiceId: serviceId,
					Force:     true,
				})
				Expect(err).To(BeNil())
				Expect(respDelete.Response.GetCode()).To(Equal(pb.ResponseSuccess))

				respGetServiceDetail, err = govern.GetServiceDetail(getContext(), &pb.GetServiceRequest{
					ServiceId: serviceId,
				})
				Expect(err).ToNot(BeNil())
				Expect(respGetServiceDetail).To(BeNil())
			})
		})
	})

	Describe("execute 'get apps' operation", func() {
		Context("when request is invalid", func() {
			It("should be failed", func() {
				resp, err := govern.ListApp(getContext(), &pb.GetAppsRequest{
					Environment: "non-exist-env",
				})
				Expect(err).ToNot(BeNil())
				Expect(resp).To(BeNil())
			})
		})

		Context("when request is valid", func() {
			It("should be passed", func() {
				resp, err := govern.ListApp(getContext(), &pb.GetAppsRequest{})
				Expect(err).To(BeNil())
				Expect(resp.Response.GetCode()).To(Equal(pb.ResponseSuccess))

				resp, err = govern.ListApp(getContext(), &pb.GetAppsRequest{
					Environment: pb.ENV_ACCEPT,
				})
				Expect(err).To(BeNil())
				Expect(resp.Response.GetCode()).To(Equal(pb.ResponseSuccess))
			})
		})
	})
})
