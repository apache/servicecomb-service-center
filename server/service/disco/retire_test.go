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

package disco_test

import (
	"fmt"
	"testing"

	"github.com/apache/servicecomb-service-center/datasource"
	"github.com/apache/servicecomb-service-center/datasource/schema"
	discosvc "github.com/apache/servicecomb-service-center/server/service/disco"
	"github.com/apache/servicecomb-service-center/test"
	pb "github.com/go-chassis/cari/discovery"
	"github.com/stretchr/testify/assert"
)

func TestRetireService(t *testing.T) {
	if !test.IsETCD() {
		return
	}

	const serviceIDPrefix = "TestRetireMicroservice"
	ctx := getContext()

	t.Run("normal case should return ok", func(t *testing.T) {
		const count = 5
		for i := 0; i < count; i++ {
			idx := fmt.Sprintf("%d", i)
			_, err := discosvc.RegisterService(ctx, &pb.CreateServiceRequest{
				Service: &pb.MicroService{
					ServiceId:   serviceIDPrefix + idx,
					ServiceName: serviceIDPrefix,
					Version:     "1.0." + idx,
				},
			})
			assert.NoError(t, err)
		}
		defer func() {
			for i := 0; i < count; i++ {
				idx := fmt.Sprintf("%d", i)
				discosvc.UnregisterService(ctx, &pb.DeleteServiceRequest{
					ServiceId: serviceIDPrefix + idx,
					Force:     true,
				})
			}
		}()

		err := discosvc.RetireService(ctx, &datasource.RetirePlan{Reserve: 1})
		assert.NoError(t, err)

		resp, err := datasource.GetMetadataManager().ListServiceDetail(ctx, &pb.GetServicesInfoRequest{
			ServiceName: serviceIDPrefix,
		})
		assert.NoError(t, err)
		assert.Equal(t, 1, len(resp.AllServicesDetail))
		assert.Equal(t, serviceIDPrefix+"4", resp.AllServicesDetail[0].MicroService.ServiceId)
	})
}

func TestRetireSchema(t *testing.T) {
	if !test.IsETCD() {
		return
	}

	var serviceID string

	ctx := getContext()
	service, err := discosvc.RegisterService(ctx, &pb.CreateServiceRequest{
		Service: &pb.MicroService{
			ServiceName: "TestRetireSchema",
		},
	})
	assert.NoError(t, err)
	serviceID = service.ServiceId
	schemaID := "schemaID_1"
	content := "content_1"
	hash := schema.Hash(schemaID, content)
	defer schema.Instance().DeleteContent(ctx, &schema.ContentRequest{Hash: hash})
	defer discosvc.UnregisterService(ctx, &pb.DeleteServiceRequest{ServiceId: serviceID, Force: true})

	t.Run("retire schema with ref, should not delete it", func(t *testing.T) {
		err = discosvc.PutSchema(ctx, &pb.ModifySchemaRequest{
			ServiceId: serviceID,
			SchemaId:  schemaID,
			Schema:    content,
		})
		assert.NoError(t, err)

		err := discosvc.RetireSchema(ctx)
		assert.NoError(t, err)

		_, err = discosvc.GetSchema(ctx, &pb.GetSchemaRequest{ServiceId: serviceID, SchemaId: schemaID})
		assert.NoError(t, err)
	})

	t.Run("retire schema without ref, should delete it", func(t *testing.T) {
		err := discosvc.DeleteSchema(ctx, &pb.DeleteSchemaRequest{
			ServiceId: serviceID,
			SchemaId:  schemaID,
		})
		assert.NoError(t, err)

		err = discosvc.RetireSchema(ctx)
		assert.NoError(t, err)

		_, err = schema.Instance().GetContent(ctx, &schema.ContentRequest{Hash: hash})
		assert.ErrorIs(t, schema.ErrSchemaContentNotFound, err)
	})
}
