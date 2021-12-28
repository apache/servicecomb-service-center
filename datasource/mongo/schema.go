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

package mongo

import (
	"context"

	"github.com/apache/servicecomb-service-center/datasource"
	"github.com/apache/servicecomb-service-center/datasource/schema"
	"github.com/go-chassis/cari/discovery"
)

func init() {
	schema.Install("mongo", NewSchemaDAO)
}

func NewSchemaDAO(opts schema.Options) (schema.DAO, error) {
	return &SchemaDAO{}, nil
}

type SchemaDAO struct{}

func (s *SchemaDAO) GetRef(ctx context.Context, refRequest *schema.RefRequest) (*schema.Ref, error) {
	return nil, schema.ErrSchemaNotFound
}

func (s *SchemaDAO) ListRef(ctx context.Context, refRequest *schema.RefRequest) ([]*schema.Ref, error) {
	return nil, nil
}

func (s *SchemaDAO) DeleteRef(ctx context.Context, refRequest *schema.RefRequest) error {
	return schema.ErrSchemaNotFound
}

func (s *SchemaDAO) GetContent(ctx context.Context, contentRequest *schema.ContentRequest) (*schema.Content, error) {
	return nil, schema.ErrSchemaNotFound
}

func (s *SchemaDAO) PutContent(ctx context.Context, contentRequest *schema.PutContentRequest) error {
	_, err := datasource.GetMetadataManager().ModifySchema(ctx, &discovery.ModifySchemaRequest{
		ServiceId: contentRequest.ServiceID,
		SchemaId:  contentRequest.SchemaID,
		Schema:    contentRequest.Content.Content,
		Summary:   contentRequest.Content.Summary,
	})
	if err != nil {
		return err
	}
	return nil
}

func (s *SchemaDAO) PutManyContent(ctx context.Context, contentRequest *schema.PutManyContentRequest) error {
	var schemas []*discovery.Schema
	for i, item := range contentRequest.Contents {
		schemaID := contentRequest.SchemaIDs[i]
		schemas = append(schemas, &discovery.Schema{
			SchemaId: schemaID,
			Summary:  item.Summary,
			Schema:   item.Content,
		})
	}
	_, err := datasource.GetMetadataManager().ModifySchemas(ctx, &discovery.ModifySchemasRequest{
		ServiceId: contentRequest.ServiceID,
		Schemas:   schemas,
	})
	if err != nil {
		return err
	}
	return nil
}

func (s *SchemaDAO) DeleteContent(ctx context.Context, contentRequest *schema.ContentRequest) error {
	return schema.ErrSchemaContentNotFound
}

func (s *SchemaDAO) DeleteNoRefContents(ctx context.Context) (int, error) {
	return 0, nil
}
