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
	"github.com/go-chassis/cari/pkg/errsvc"

	"github.com/apache/servicecomb-service-center/pkg/util"
	"github.com/apache/servicecomb-service-center/server/plugin/quota/buildin"
	"github.com/apache/servicecomb-service-center/server/service/disco"
	pb "github.com/go-chassis/cari/discovery"
	"github.com/stretchr/testify/assert"
)

func TestSchemaUsage(t *testing.T) {
	ctx := util.WithNoCache(util.SetDomainProject(context.Background(), "default", "default"))

	resp, err := disco.RegisterService(ctx, &pb.CreateServiceRequest{
		Service: &pb.MicroService{
			ServiceName: "TestSchemaUsage",
		},
	})
	assert.NoError(t, err)
	serviceID := resp.ServiceId
	defer disco.UnregisterService(ctx, &pb.DeleteServiceRequest{ServiceId: serviceID, Force: true})

	t.Run("get not exist service schema usage, should failed", func(t *testing.T) {
		_, err := buildin.SchemaUsage(ctx, "note_exist_service_id")
		testErr := err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrServiceNotExists, testErr.Code)
	})

	t.Run("get usage without schemas, should return 0", func(t *testing.T) {
		usage, err := buildin.SchemaUsage(ctx, serviceID)
		assert.NoError(t, err)
		assert.Equal(t, int64(0), usage)
	})

	t.Run("get usage with schemas, should return 1", func(t *testing.T) {
		err := disco.PutSchema(ctx, &pb.ModifySchemaRequest{
			ServiceId: serviceID,
			SchemaId:  "schemaID_1",
			Schema:    "schema_1",
		})
		assert.NoError(t, err)

		usage, err := buildin.SchemaUsage(ctx, serviceID)
		assert.NoError(t, err)
		assert.Equal(t, int64(1), usage)
	})
}
