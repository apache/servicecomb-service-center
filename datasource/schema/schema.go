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

package schema

import (
	"context"
	"crypto/md5"
	"fmt"

	"github.com/go-chassis/cari/discovery"
)

var (
	ErrSchemaNotFound        = discovery.NewError(discovery.ErrSchemaNotExists, "schema ref not found.")
	ErrSchemaContentNotFound = discovery.NewError(discovery.ErrSchemaNotExists, "schema content not found.")
)

type RefRequest struct {
	ServiceID string `json:"serviceId" bson:"service_id"`
	SchemaID  string `json:"schemaId" bson:"schema_id"`
}

type Ref struct {
	Domain    string
	Project   string
	ServiceID string `json:"serviceId" bson:"service_id"`
	SchemaID  string `json:"schemaId" bson:"schema_id"`
	Hash      string
	Summary   string
	Content   string
}

type ContentRequest struct {
	Hash string
}

type Content struct {
	Domain  string
	Project string
	Hash    string
	Content string
}

type ContentItem struct {
	Hash    string
	Content string
	Summary string
}

type PutContentRequest struct {
	ServiceID string `json:"serviceId" bson:"service_id"`
	SchemaID  string `json:"schemaId" bson:"schema_id"`
	Content   *ContentItem
}

type PutManyContentRequest struct {
	ServiceID string `json:"serviceId" bson:"service_id"`
	SchemaIDs []string
	Contents  []*ContentItem
	Init      bool
}

type DAO interface {
	GetRef(ctx context.Context, refRequest *RefRequest) (*Ref, error)
	ListRef(ctx context.Context, refRequest *RefRequest) ([]*Ref, error)
	DeleteRef(ctx context.Context, refRequest *RefRequest) error
	// GetContent get a schema content, hash is the result of MD5(schemaId+': '+content), see: Hash
	GetContent(ctx context.Context, contentRequest *ContentRequest) (*Content, error)
	PutContent(ctx context.Context, contentRequest *PutContentRequest) error
	PutManyContent(ctx context.Context, contentRequest *PutManyContentRequest) error
	DeleteContent(ctx context.Context, contentRequest *ContentRequest) error
	DeleteNoRefContents(ctx context.Context) (int, error)
}

func Hash(schemaID, content string) string {
	return fmt.Sprintf("%x", md5.Sum([]byte(schemaID+": "+content)))
}
