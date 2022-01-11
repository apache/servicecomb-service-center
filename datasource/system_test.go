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

package datasource_test

import (
	"testing"

	"github.com/apache/servicecomb-service-center/datasource"
	"github.com/apache/servicecomb-service-center/datasource/mongo/sd"
	pb "github.com/go-chassis/cari/discovery"
	"github.com/stretchr/testify/assert"
)

func TestDumpCache(t *testing.T) {
	var serviceID string
	var store = &sd.TypeStore{}
	store.Initialize()
	ctx := getContext()
	defer datasource.GetMetadataManager().UnregisterService(ctx, &pb.DeleteServiceRequest{ServiceId: serviceID, Force: true})

	t.Run("Register service && instance, check dump, should pass", func(t *testing.T) {
		service, err := datasource.GetMetadataManager().RegisterService(ctx, &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				ServiceName: "create_service_test",
				AppId:       "create_service_appId",
				Version:     "1.0.0",
				Level:       "BACK",
				Status:      pb.MS_UP,
			},
		})
		assert.NoError(t, err)
		serviceID = service.ServiceId

		_, err = datasource.GetMetadataManager().RegisterInstance(ctx, &pb.RegisterInstanceRequest{
			Instance: &pb.MicroServiceInstance{
				ServiceId: serviceID,
				Endpoints: []string{
					"localhost://127.0.0.1:8080",
				},
				HostName: "HOST_TEST",
				Status:   pb.MSI_UP,
			},
		})
		assert.NoError(t, err)

		cache := datasource.GetSystemManager().DumpCache(ctx)
		assert.NotNil(t, cache)
	})
}
