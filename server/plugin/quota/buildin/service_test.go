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

package buildin_test

import (
	"context"
	"testing"

	_ "github.com/apache/servicecomb-service-center/test"

	"github.com/apache/servicecomb-service-center/pkg/util"
	"github.com/apache/servicecomb-service-center/server/plugin/quota/buildin"
	"github.com/apache/servicecomb-service-center/server/service/disco"
	pb "github.com/go-chassis/cari/discovery"
	"github.com/stretchr/testify/assert"
)

func TestServiceUsage(t *testing.T) {
	t.Run("get domain/project without service usage, should return 0", func(t *testing.T) {
		usage, err := buildin.ServiceUsage(context.Background(), &pb.GetServiceCountRequest{
			Domain:  "domain_without_service",
			Project: "project_without_service",
		})
		assert.NoError(t, err)
		assert.Equal(t, int64(0), usage)
	})

	t.Run("get domain/project with 1 service usage, should return 1", func(t *testing.T) {
		ctx := util.SetDomainProject(context.Background(), "domain_with_service", "project_with_service")
		resp, err := disco.RegisterService(ctx, &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				ServiceName: "test",
			},
		})
		assert.NoError(t, err)
		defer disco.UnregisterService(ctx, &pb.DeleteServiceRequest{ServiceId: resp.ServiceId, Force: true})

		usage, err := buildin.ServiceUsage(context.Background(), &pb.GetServiceCountRequest{
			Domain:  "domain_with_service",
			Project: "project_with_service",
		})
		assert.NoError(t, err)
		assert.Equal(t, int64(1), usage)
	})
}
