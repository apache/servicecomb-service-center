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
package pubsub_test

import (
	"context"
	"testing"

	"github.com/apache/servicecomb-service-center/pkg/util"
	"github.com/apache/servicecomb-service-center/server/pubsub"
	"github.com/apache/servicecomb-service-center/server/service/disco"

	pb "github.com/go-chassis/cari/discovery"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func getContext() context.Context {
	return util.WithNoCache(util.SetDomainProject(context.Background(), "default", "default"))
}

func TestInstanceService_WebSocketWatch(t *testing.T) {
	defer func() {
		recover()
	}()
	pubsub.Watch(context.Background(), &pb.WatchInstanceRequest{}, nil)
}

var _ = Describe("'Instance' service", func() {
	Describe("execute 'watch' operartion", func() {
		var (
			serviceId string
		)

		It("should be passed", func() {
			respCreate, err := disco.RegisterService(getContext(), &pb.CreateServiceRequest{
				Service: &pb.MicroService{
					ServiceName: "service_name_watch",
					AppId:       "service_name_watch",
					Version:     "1.0.0",
					Level:       "BACK",
					Status:      pb.MS_UP,
				},
			})
			Expect(err).To(BeNil())
			serviceId = respCreate.ServiceId
		})

		Context("when request is invalid", func() {
			It("should be failed", func() {
				By("service does not exist")
				err := pubsub.ExistService(getContext(), "-1")
				Expect(err).NotTo(BeNil())

				By("service id is empty")
				err = pubsub.ExistService(getContext(), "")
				Expect(err).NotTo(BeNil())

				By("request is valid")
				err = pubsub.ExistService(getContext(), serviceId)
				Expect(err).To(BeNil())
			})
		})
	})
})
