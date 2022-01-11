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

package schema_test

import (
	"context"
	"testing"

	_ "github.com/apache/servicecomb-service-center/test"

	"github.com/apache/servicecomb-service-center/datasource"
	"github.com/apache/servicecomb-service-center/datasource/schema"
	"github.com/apache/servicecomb-service-center/pkg/util"
	pb "github.com/go-chassis/cari/discovery"
	"github.com/go-chassis/cari/pkg/errsvc"
	"github.com/stretchr/testify/assert"
)

func getContext() context.Context {
	return util.WithNoCache(util.SetDomainProject(context.Background(), "default", "default"))
}

func TestGetRef(t *testing.T) {
	var (
		serviceID string
		hash      = "TestGetRef_hash_1"
	)
	ctx := getContext()

	svc := &pb.MicroService{
		ServiceName: "TestGetRef",
	}
	resp, err := datasource.GetMetadataManager().RegisterService(ctx, &pb.CreateServiceRequest{
		Service: svc,
	})
	assert.NoError(t, err)
	assert.NotEqual(t, "", resp.ServiceId)
	serviceID = resp.ServiceId
	defer schema.Instance().DeleteContent(ctx, &schema.ContentRequest{
		Hash: hash,
	})
	defer datasource.GetMetadataManager().UnregisterService(ctx, &pb.DeleteServiceRequest{ServiceId: serviceID, Force: true})

	t.Run("get ref with invalid request, should failed", func(t *testing.T) {
		ref, err := schema.Instance().GetRef(ctx, &schema.RefRequest{
			ServiceID: "not_exist",
			SchemaID:  "not_exist",
		})
		testErr := err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrSchemaNotExists, testErr.Code)
		assert.Nil(t, ref)

		ref, err = schema.Instance().GetRef(ctx, &schema.RefRequest{
			ServiceID: serviceID,
			SchemaID:  "not_exist",
		})
		testErr = err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrSchemaNotExists, testErr.Code)
		assert.Nil(t, ref)
	})

	t.Run("get a exist ref, should ok", func(t *testing.T) {
		err := schema.Instance().PutContent(ctx, &schema.PutContentRequest{
			ServiceID: serviceID,
			SchemaID:  "schemaID_1",
			Content: &schema.ContentItem{
				Hash:    hash,
				Summary: "summary_1",
			},
		})
		assert.NoError(t, err)

		ref, err := schema.Instance().GetRef(ctx, &schema.RefRequest{
			ServiceID: serviceID,
			SchemaID:  "schemaID_1",
		})
		assert.NoError(t, err)
		assert.NotNil(t, ref)
		assert.Equal(t, "summary_1", ref.Summary)
		assert.Equal(t, hash, ref.Hash)
	})
}

func TestPutContent(t *testing.T) {
	var (
		serviceID string
		hash      = "TestPutContent_hash_1"
	)
	ctx := getContext()

	svc := &pb.MicroService{
		ServiceName: "TestPutContent",
		Schemas:     []string{"schemaID_1"},
	}
	resp, err := datasource.GetMetadataManager().RegisterService(ctx, &pb.CreateServiceRequest{
		Service: svc,
	})
	assert.NoError(t, err)
	assert.NotEqual(t, "", resp.ServiceId)
	serviceID = resp.ServiceId
	defer schema.Instance().DeleteContent(ctx, &schema.ContentRequest{Hash: hash})
	defer datasource.GetMetadataManager().UnregisterService(ctx, &pb.DeleteServiceRequest{ServiceId: serviceID, Force: true})

	t.Run("put content with invalid request, should failed", func(t *testing.T) {
		err := schema.Instance().PutContent(ctx, &schema.PutContentRequest{})
		testErr := err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrServiceNotExists, testErr.Code)
	})

	t.Run("put content with valid request, should ok", func(t *testing.T) {
		err := schema.Instance().PutContent(ctx, &schema.PutContentRequest{
			ServiceID: serviceID,
			SchemaID:  "schemaID_1",
			Content: &schema.ContentItem{
				Hash:    hash,
				Content: "content_1",
				Summary: "summary_1",
			},
		})
		assert.NoError(t, err)

		content, err := schema.Instance().GetContent(ctx, &schema.ContentRequest{
			Hash: hash,
		})
		assert.NoError(t, err)
		assert.NotNil(t, content)
		assert.Equal(t, hash, content.Hash)
		assert.Equal(t, "content_1", content.Content)

		ref, err := schema.Instance().GetRef(ctx, &schema.RefRequest{
			ServiceID: serviceID,
			SchemaID:  "schemaID_1",
		})
		assert.NoError(t, err)
		assert.NotNil(t, ref)
		assert.Equal(t, "summary_1", ref.Summary)
	})

	t.Run("put the same content, should ok", func(t *testing.T) {
		err := schema.Instance().PutContent(ctx, &schema.PutContentRequest{
			ServiceID: serviceID,
			SchemaID:  "schemaID_2",
			Content: &schema.ContentItem{
				Hash:    hash,
				Content: "content_2",
				Summary: "summary_2",
			},
		})
		assert.NoError(t, err)

		content, err := schema.Instance().GetContent(ctx, &schema.ContentRequest{
			Hash: hash,
		})
		assert.NoError(t, err)
		assert.NotNil(t, content)
		assert.Equal(t, hash, content.Hash)
		assert.Equal(t, "content_1", content.Content)

		ref, err := schema.Instance().GetRef(ctx, &schema.RefRequest{
			ServiceID: serviceID,
			SchemaID:  "schemaID_2",
		})
		assert.NoError(t, err)
		assert.NotNil(t, ref)
		assert.Equal(t, "summary_2", ref.Summary)
	})

	t.Run("check service schemaIDs after put content, should be updated", func(t *testing.T) {
		service, err := datasource.GetMetadataManager().GetService(ctx, &pb.GetServiceRequest{
			ServiceId: serviceID,
		})
		assert.NoError(t, err)
		assert.NotNil(t, service)
		assert.Equal(t, []string{"schemaID_1", "schemaID_2"}, service.Schemas)
	})
}

func TestDeleteContent(t *testing.T) {
	var (
		serviceID string
		hash      = "TestDeleteContent_hash_1"
	)
	ctx := getContext()

	svc := &pb.MicroService{
		ServiceName: "TestDeleteContent",
	}
	resp, err := datasource.GetMetadataManager().RegisterService(ctx, &pb.CreateServiceRequest{
		Service: svc,
	})
	assert.NoError(t, err)
	assert.NotEqual(t, "", resp.ServiceId)
	serviceID = resp.ServiceId

	t.Run("delete content with invalid request, should failed", func(t *testing.T) {
		err := schema.Instance().DeleteContent(ctx, &schema.ContentRequest{})
		testErr := err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrSchemaNotExists, testErr.Code)
	})

	t.Run("delete content referenced by service, should failed", func(t *testing.T) {
		err := schema.Instance().PutContent(ctx, &schema.PutContentRequest{
			ServiceID: serviceID,
			SchemaID:  "schemaID_1",
			Content: &schema.ContentItem{
				Hash:    hash,
				Content: "content_1",
				Summary: "summary_1",
			},
		})
		assert.NoError(t, err)

		err = schema.Instance().DeleteContent(ctx, &schema.ContentRequest{
			Hash: hash,
		})
		testErr := err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrInvalidParams, testErr.Code)
	})

	t.Run("delete content without ref, should ok", func(t *testing.T) {
		err := datasource.GetMetadataManager().UnregisterService(ctx, &pb.DeleteServiceRequest{ServiceId: serviceID, Force: true})
		assert.NoError(t, err)

		err = schema.Instance().DeleteContent(ctx, &schema.ContentRequest{
			Hash: hash,
		})
		assert.NoError(t, err)

		_, err = schema.Instance().GetRef(ctx, &schema.RefRequest{
			ServiceID: serviceID,
			SchemaID:  "schemaID_1",
		})
		testErr := err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrSchemaNotExists, testErr.Code)
	})
}

func TestPutManyContent(t *testing.T) {
	var serviceID string
	ctx := getContext()

	svc := &pb.MicroService{
		ServiceName: "TestPutManyContent",
		Schemas:     []string{"schemaID_1"},
	}
	resp, err := datasource.GetMetadataManager().RegisterService(ctx, &pb.CreateServiceRequest{
		Service: svc,
	})
	assert.NoError(t, err)
	assert.NotEqual(t, "", resp.ServiceId)
	serviceID = resp.ServiceId
	defer schema.Instance().DeleteContent(ctx, &schema.ContentRequest{Hash: "hash_3"})
	defer datasource.GetMetadataManager().UnregisterService(ctx, &pb.DeleteServiceRequest{ServiceId: serviceID, Force: true})

	t.Run("put many content with invalid request, should failed", func(t *testing.T) {
		err := schema.Instance().PutManyContent(ctx, &schema.PutManyContentRequest{})
		testErr := err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrServiceNotExists, testErr.Code)

		err = schema.Instance().PutManyContent(ctx, &schema.PutManyContentRequest{
			ServiceID: serviceID,
			SchemaIDs: []string{"schemaID_1"},
			Contents: []*schema.ContentItem{
				{
					Hash: "hash_1",
				},
				{
					Hash: "hash_2",
				},
			},
		})
		testErr = err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrInvalidParams, testErr.Code)
	})

	t.Run("put content with valid request, should ok", func(t *testing.T) {
		err = schema.Instance().PutManyContent(ctx, &schema.PutManyContentRequest{
			ServiceID: serviceID,
			SchemaIDs: []string{"schemaID_1", "schemaID_2"},
			Contents: []*schema.ContentItem{
				{
					Hash:    "hash_1",
					Content: "content_1",
					Summary: "summary_1",
				},
				{
					Hash:    "hash_2",
					Content: "content_2",
					Summary: "summary_2",
				},
			},
		})
		assert.NoError(t, err)

		content, err := schema.Instance().GetContent(ctx, &schema.ContentRequest{
			Hash: "hash_1",
		})
		assert.NoError(t, err)
		assert.NotNil(t, content)
		assert.Equal(t, "hash_1", content.Hash)
		assert.Equal(t, "content_1", content.Content)

		content, err = schema.Instance().GetContent(ctx, &schema.ContentRequest{
			Hash: "hash_2",
		})
		assert.NoError(t, err)
		assert.NotNil(t, content)
		assert.Equal(t, "hash_2", content.Hash)
		assert.Equal(t, "content_2", content.Content)

		ref, err := schema.Instance().GetRef(ctx, &schema.RefRequest{
			ServiceID: serviceID,
			SchemaID:  "schemaID_1",
		})
		assert.NoError(t, err)
		assert.NotNil(t, ref)
		assert.Equal(t, "summary_1", ref.Summary)

		ref, err = schema.Instance().GetRef(ctx, &schema.RefRequest{
			ServiceID: serviceID,
			SchemaID:  "schemaID_2",
		})
		assert.NoError(t, err)
		assert.NotNil(t, ref)
		assert.Equal(t, "summary_2", ref.Summary)

		service, err := datasource.GetMetadataManager().GetService(ctx, &pb.GetServiceRequest{
			ServiceId: serviceID,
		})
		assert.NoError(t, err)
		assert.NotNil(t, service)
		assert.Equal(t, []string{"schemaID_1", "schemaID_2"}, service.Schemas)
	})

	t.Run("put many content again, should overwrite", func(t *testing.T) {
		err = schema.Instance().PutManyContent(ctx, &schema.PutManyContentRequest{
			ServiceID: serviceID,
			SchemaIDs: []string{"schemaID_3"},
			Contents: []*schema.ContentItem{
				{
					Hash:    "hash_3",
					Content: "content_3",
					Summary: "summary_3",
				},
			},
		})
		assert.NoError(t, err)

		content, err := schema.Instance().GetContent(ctx, &schema.ContentRequest{
			Hash: "hash_1",
		})
		assert.NoError(t, err)
		assert.NotNil(t, content)
		assert.Equal(t, "hash_1", content.Hash)
		assert.Equal(t, "content_1", content.Content)

		content, err = schema.Instance().GetContent(ctx, &schema.ContentRequest{
			Hash: "hash_3",
		})
		assert.NoError(t, err)
		assert.NotNil(t, content)
		assert.Equal(t, "hash_3", content.Hash)
		assert.Equal(t, "content_3", content.Content)

		ref, err := schema.Instance().GetRef(ctx, &schema.RefRequest{
			ServiceID: serviceID,
			SchemaID:  "schemaID_1",
		})
		testErr := err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrSchemaNotExists, testErr.Code)

		ref, err = schema.Instance().GetRef(ctx, &schema.RefRequest{
			ServiceID: serviceID,
			SchemaID:  "schemaID_3",
		})
		assert.NoError(t, err)
		assert.NotNil(t, ref)
		assert.Equal(t, "summary_3", ref.Summary)

		service, err := datasource.GetMetadataManager().GetService(ctx, &pb.GetServiceRequest{
			ServiceId: serviceID,
		})
		assert.NoError(t, err)
		assert.NotNil(t, service)
		assert.Equal(t, []string{"schemaID_3"}, service.Schemas)

		err = schema.Instance().DeleteContent(ctx, &schema.ContentRequest{Hash: "hash_1"})
		assert.NoError(t, err)
		err = schema.Instance().DeleteContent(ctx, &schema.ContentRequest{Hash: "hash_2"})
		assert.NoError(t, err)
	})
}
