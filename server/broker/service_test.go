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
package broker_test

import (
	"context"
	"fmt"
	"github.com/apache/servicecomb-service-center/pkg/util"
	"github.com/apache/servicecomb-service-center/server/broker"
	"github.com/apache/servicecomb-service-center/server/broker/brokerpb"
	"github.com/apache/servicecomb-service-center/server/core"
	pb "github.com/go-chassis/cari/discovery"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

const (
	TEST_BROKER_NO_SERVICE_ID = "noServiceId"
	TEST_BROKER_NO_VERSION    = "noVersion"
	//Consumer
	TEST_BROKER_CONSUMER_VERSION = "4.0.0"
	TEST_BROKER_CONSUMER_NAME    = "broker_name_consumer"
	TEST_BROKER_CONSUMER_APP     = "broker_group_consumer"
	//Provider
	TEST_BROKER_PROVIDER_VERSION = "3.0.0"
	TEST_BROKER_PROVIDER_NAME    = "broker_name_provider"
	TEST_BROKER_PROVIDER_APP     = "broker_group_provider"
)

var consumerServiceId string
var providerServiceId string

var _ = Describe("Controller", func() {
	Describe("brokerDependency", func() {
		Context("normal", func() {
			It("PublishPact", func() {
				fmt.Println("UT===========PublishPact")

				//(1) create consumer service
				resp, err := core.ServiceAPI.Create(getContext(), &pb.CreateServiceRequest{
					Service: &pb.MicroService{
						ServiceName: TEST_BROKER_CONSUMER_NAME,
						AppId:       TEST_BROKER_CONSUMER_APP,
						Version:     TEST_BROKER_CONSUMER_VERSION,
						Level:       "FRONT",
						Schemas: []string{
							"xxxxxxxx",
						},
						Status: "UP",
					},
				})

				Expect(err).To(BeNil())
				consumerServiceId = resp.ServiceId
				Expect(resp.Response.GetCode()).To(Equal(pb.ResponseSuccess))

				//(2) create provider service
				resp, err = core.ServiceAPI.Create(getContext(), &pb.CreateServiceRequest{
					Service: &pb.MicroService{
						ServiceName: TEST_BROKER_PROVIDER_NAME,
						AppId:       TEST_BROKER_PROVIDER_APP,
						Version:     TEST_BROKER_PROVIDER_VERSION,
						Level:       "FRONT",
						Schemas: []string{
							"xxxxxxxx",
						},
						Status:     "UP",
						Properties: map[string]string{"allowCrossApp": "true"},
					},
				})
				Expect(err).To(BeNil())
				providerServiceId = resp.ServiceId
				Expect(resp.Response.GetCode()).To(Equal(pb.ResponseSuccess))

				//(3) publish a pact between two services
				respPublishPact, err := brokerResource.PublishPact(getContext(),
					&brokerpb.PublishPactRequest{
						ProviderId: providerServiceId,
						ConsumerId: consumerServiceId,
						Version:    TEST_BROKER_CONSUMER_VERSION,
						Pact:       []byte("hello"),
					})

				Expect(err).To(BeNil())
				Expect(respPublishPact.Response.GetCode()).To(Equal(pb.ResponseSuccess))
			})

			It("PublishPact-noProviderServiceId", func() {
				fmt.Println("UT===========PublishPact, no provider serviceID")

				//publish a pact between two services
				respPublishPact, _ := brokerResource.PublishPact(getContext(), &brokerpb.PublishPactRequest{
					ProviderId: TEST_BROKER_NO_SERVICE_ID,
					ConsumerId: consumerServiceId,
					Version:    TEST_BROKER_CONSUMER_VERSION,
					Pact:       []byte("hello"),
				})

				Expect(respPublishPact.Response.GetCode()).ToNot(Equal(pb.ResponseSuccess))
			})

			It("PublishPact-noConumerServiceId", func() {
				fmt.Println("UT===========PublishPact, no consumer serviceID")

				//publish a pact between two services
				respPublishPact, _ := brokerResource.PublishPact(getContext(), &brokerpb.PublishPactRequest{
					ProviderId: providerServiceId,
					ConsumerId: TEST_BROKER_NO_SERVICE_ID,
					Version:    TEST_BROKER_CONSUMER_VERSION,
					Pact:       []byte("hello"),
				})

				Expect(respPublishPact.Response.GetCode()).ToNot(Equal(pb.ResponseSuccess))
			})

			It("PublishPact-noConumerVersion", func() {
				fmt.Println("UT===========PublishPact, no consumer Version")

				//publish a pact between two services
				respPublishPact, _ := brokerResource.PublishPact(getContext(), &brokerpb.PublishPactRequest{
					ProviderId: providerServiceId,
					ConsumerId: consumerServiceId,
					Version:    TEST_BROKER_NO_VERSION,
					Pact:       []byte("hello"),
				})

				Expect(respPublishPact.Response.GetCode()).ToNot(Equal(pb.ResponseSuccess))
			})

			It("GetBrokerHome", func() {
				fmt.Println("UT===========GetBrokerHome")

				respGetHome, _ := brokerResource.GetBrokerHome(getContext(), &brokerpb.BaseBrokerRequest{
					HostAddress: "localhost",
					Scheme:      "http",
				})

				Expect(respGetHome).NotTo(BeNil())

			})

			It("GetBrokerAllProviderPacts", func() {
				fmt.Println("UT===========GetBrokerAllProviderPacts")

				respGetAllProviderPacts, _ := brokerResource.GetAllProviderPacts(getContext(),
					&brokerpb.GetAllProviderPactsRequest{
						ProviderId: providerServiceId,
						BaseUrl: &brokerpb.BaseBrokerRequest{
							HostAddress: "localhost",
							Scheme:      "http",
						}})

				Expect(respGetAllProviderPacts).NotTo(BeNil())
				Expect(respGetAllProviderPacts.Response.GetCode()).To(Equal(pb.ResponseSuccess))
			})

			It("GetBrokerPactsOfProvider", func() {
				fmt.Println("UT===========GetBrokerPactsOfProvider")

				respGetAllProviderPacts, _ := brokerResource.GetPactsOfProvider(getContext(),
					&brokerpb.GetProviderConsumerVersionPactRequest{
						ProviderId: providerServiceId,
						ConsumerId: consumerServiceId,
						Version:    TEST_BROKER_CONSUMER_VERSION,
						BaseUrl: &brokerpb.BaseBrokerRequest{
							HostAddress: "localhost",
							Scheme:      "http",
						}})

				Expect(respGetAllProviderPacts).NotTo(BeNil())
				Expect(respGetAllProviderPacts.Response.GetCode()).To(Equal(pb.ResponseSuccess))
			})

			It("PublishVerificationResults", func() {
				fmt.Println("UT===========PublishVerificationResults")

				id, err := broker.GetData(context.Background(), broker.GetBrokerLatestPactIDKey())
				Expect(err).To(BeNil())
				respResults, err := brokerResource.PublishVerificationResults(getContext(),
					&brokerpb.PublishVerificationRequest{
						ProviderId:                 providerServiceId,
						ConsumerId:                 consumerServiceId,
						PactId:                     int32(id),
						ProviderApplicationVersion: TEST_BROKER_PROVIDER_VERSION,
					})

				Expect(respResults).NotTo(BeNil())
				Expect(respResults.Response.GetCode()).To(Equal(pb.ResponseSuccess))
			})

			It("RetrieveVerificationResults", func() {
				fmt.Println("UT===========RetrieveVerificationResults")

				respVerification, _ := brokerResource.RetrieveVerificationResults(getContext(),
					&brokerpb.RetrieveVerificationRequest{
						ConsumerId:      consumerServiceId,
						ConsumerVersion: TEST_BROKER_CONSUMER_VERSION,
					})

				Expect(respVerification).NotTo(BeNil())
				Expect(respVerification.Response.GetCode()).To(Equal(pb.ResponseSuccess))
			})

			It("RetrieveProviderPacts", func() {
				fmt.Println("UT===========RetrieveProviderPacts")

				respProviderPact, _ := brokerResource.RetrieveProviderPacts(getContext(),
					&brokerpb.GetAllProviderPactsRequest{
						ProviderId: providerServiceId,
						BaseUrl: &brokerpb.BaseBrokerRequest{
							HostAddress: "localhost",
							Scheme:      "http",
						},
					})

				Expect(respProviderPact).NotTo(BeNil())
				Expect(respProviderPact.Response.GetCode()).To(Equal(pb.ResponseSuccess))
			})
		})
	})
})

func getContext() context.Context {
	ctx := context.TODO()
	ctx = util.SetContext(ctx, "domain", "default")
	ctx = util.SetContext(ctx, "project", "default")
	ctx = util.SetContext(ctx, util.CtxNocache, "1")
	return ctx
}
