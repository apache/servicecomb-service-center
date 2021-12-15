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
	"encoding/hex"
)

type RefRequest struct {
	Domain    string
	Project   string
	ServiceID string `json:"serviceId" bson:"service_id"`
	SchemaID  string `json:"schemaId" bson:"schema_id"`
}

type Ref struct {
	Domain    string
	Project   string
	ServiceID string `json:"serviceId" bson:"service_id"`
	SchemaID  string `json:"schemaId" bson:"schema_id"`
	Hash      string
}

type ContentRequest struct {
	Domain  string
	Project string
	Hash    string
}

type Content struct {
	Domain  string
	Project string
	Hash    string
	Content string
}

type DAO interface {
	GetRef(ctx context.Context, ref *Ref) (*Ref, error)
	PutRef(ctx context.Context, ref *Ref) error
	DeleteRef(ctx context.Context, ref ...*Ref) error
	// GetContent get a schema content, hash is the result of MD5(schemaId+': '+content), see: Hash
	GetContent(ctx context.Context, hash *ContentRequest) (string, error)
	PutContent(ctx context.Context, content *Content) error
	DeleteContent(ctx context.Context, hash ...*ContentRequest) error
	// ListHash return Content list without content
	ListHash(ctx context.Context) ([]*Content, error)
	ExistRef(ctx context.Context, hash *ContentRequest) (*Ref, error)
}

func Hash(schemaID, content string) string {
	return hex.EncodeToString(md5.New().Sum([]byte(schemaID + ": " + content)))
}
