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
package broker

import (
	"fmt"

	"github.com/apache/incubator-servicecomb-service-center/pkg/util"
	pb "github.com/apache/incubator-servicecomb-service-center/server/core/proto"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"golang.org/x/net/context"
)

const (
	TEST_BROKER_NO_SERVICE_ID      = "noServiceId"
	TEST_BROKER_NO_VERSION         = "noVersion"
	TEST_BROKER_TOO_LONG_SERVICEID = "addasdfasaddasdfasaddasdfasaddasdfasaddasdfasaddasdfasaddasdfasadafd"
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

var _ = Describe("BrokerController", func() {
	Describe("brokerDependency", func() {
		Context("normal", func() {
			It("PublishPact", func() {
				fmt.Println("UT===========PublishPact")

				//(1) create consumer service
				resp, err := serviceResource.Create(getContext(), &pb.CreateServiceRequest{
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
				Expect(resp.GetResponse().Code).To(Equal(pb.Response_SUCCESS))

				//(2) create provider service
				resp, err = serviceResource.Create(getContext(), &pb.CreateServiceRequest{
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
				Expect(resp.GetResponse().Code).To(Equal(pb.Response_SUCCESS))

				//(3) publish a pact between two services
				respPublishPact, err := brokerResource.PublishPact(getContext(),
					&PublishPactRequest{
						ProviderId: providerServiceId,
						ConsumerId: consumerServiceId,
						Version:    TEST_BROKER_CONSUMER_VERSION,
						Pact:       []byte("hello"),
					})

				Expect(err).To(BeNil())
				Expect(respPublishPact.GetResponse().Code).To(Equal(pb.Response_SUCCESS))
			})

			It("PublishPact-noProviderServiceId", func() {
				fmt.Println("UT===========PublishPact, no provider serviceID")

				//publish a pact between two services
				respPublishPact, _ := brokerResource.PublishPact(getContext(), &PublishPactRequest{
					ProviderId: TEST_BROKER_NO_SERVICE_ID,
					ConsumerId: consumerServiceId,
					Version:    TEST_BROKER_CONSUMER_VERSION,
					Pact:       []byte("hello"),
				})

				Expect(respPublishPact.GetResponse().Code).ToNot(Equal(pb.Response_SUCCESS))
			})

			It("PublishPact-noConumerServiceId", func() {
				fmt.Println("UT===========PublishPact, no consumer serviceID")

				//publish a pact between two services
				respPublishPact, _ := brokerResource.PublishPact(getContext(), &PublishPactRequest{
					ProviderId: providerServiceId,
					ConsumerId: TEST_BROKER_NO_SERVICE_ID,
					Version:    TEST_BROKER_CONSUMER_VERSION,
					Pact:       []byte("hello"),
				})

				Expect(respPublishPact.GetResponse().Code).ToNot(Equal(pb.Response_SUCCESS))
			})

			It("PublishPact-noConumerVersion", func() {
				fmt.Println("UT===========PublishPact, no consumer Version")

				//publish a pact between two services
				respPublishPact, _ := brokerResource.PublishPact(getContext(), &PublishPactRequest{
					ProviderId: providerServiceId,
					ConsumerId: consumerServiceId,
					Version:    TEST_BROKER_NO_VERSION,
					Pact:       []byte("hello"),
				})

				Expect(respPublishPact.GetResponse().Code).ToNot(Equal(pb.Response_SUCCESS))
			})

			It("GetBrokerHome", func() {
				fmt.Println("UT===========GetBrokerHome")

				respGetHome, _ := brokerResource.GetBrokerHome(getContext(), &BaseBrokerRequest{
					HostAddress: "localhost",
					Scheme:      "http",
				})

				Expect(respGetHome).NotTo(BeNil())

			})

			It("GetBrokerAllProviderPacts", func() {
				fmt.Println("UT===========GetBrokerAllProviderPacts")

				respGetAllProviderPacts, _ := brokerResource.GetAllProviderPacts(getContext(),
					&GetAllProviderPactsRequest{
						ProviderId: providerServiceId,
						BaseUrl: &BaseBrokerRequest{
							HostAddress: "localhost",
							Scheme:      "http",
						}})

				Expect(respGetAllProviderPacts).NotTo(BeNil())
				Expect(respGetAllProviderPacts.GetResponse().Code).To(Equal(pb.Response_SUCCESS))
			})

			It("GetBrokerPactsOfProvider", func() {
				fmt.Println("UT===========GetBrokerPactsOfProvider")

				respGetAllProviderPacts, _ := brokerResource.GetPactsOfProvider(getContext(),
					&GetProviderConsumerVersionPactRequest{
						ProviderId: providerServiceId,
						ConsumerId: consumerServiceId,
						Version:    TEST_BROKER_CONSUMER_VERSION,
						BaseUrl: &BaseBrokerRequest{
							HostAddress: "localhost",
							Scheme:      "http",
						}})

				Expect(respGetAllProviderPacts).NotTo(BeNil())
				Expect(respGetAllProviderPacts.GetResponse().Code).To(Equal(pb.Response_SUCCESS))
			})

		})
	})
})

func getContext() context.Context {
	ctx := context.TODO()
	ctx = util.SetContext(ctx, "domain", "default")
	ctx = util.SetContext(ctx, "project", "default")
	ctx = util.SetContext(ctx, "noCache", "1")
	return ctx
}
