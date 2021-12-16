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

	"github.com/apache/servicecomb-service-center/datasource/schema"
)

func init() {
	schema.Install("mongo", NewSchemaDAO)
}

func NewSchemaDAO(opts schema.Options) (schema.DAO, error) {
	return &SchemaDAO{}, nil
}

type SchemaDAO struct{}

func (SchemaDAO) GetRef(ctx context.Context, ref *schema.Ref) (*schema.Ref, error) {
	panic("implement me")
}

func (SchemaDAO) PutRef(ctx context.Context, ref *schema.Ref) error {
	panic("implement me")
}

func (SchemaDAO) DeleteRef(ctx context.Context, ref ...*schema.Ref) error {
	panic("implement me")
}

func (SchemaDAO) GetContent(ctx context.Context, hash *schema.ContentRequest) (string, error) {
	panic("implement me")
}

func (SchemaDAO) PutContent(ctx context.Context, content *schema.Content) error {
	panic("implement me")
}

func (SchemaDAO) DeleteContent(ctx context.Context, hash ...*schema.ContentRequest) error {
	panic("implement me")
}

func (SchemaDAO) ListHash(ctx context.Context) ([]*schema.Content, error) {
	panic("implement me")
}

func (SchemaDAO) ExistRef(ctx context.Context, hash *schema.ContentRequest) (*schema.Ref, error) {
	panic("implement me")
}
