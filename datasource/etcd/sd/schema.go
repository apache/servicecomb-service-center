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

package sd

import (
	"github.com/apache/servicecomb-service-center/datasource/etcd/path"
	"github.com/apache/servicecomb-service-center/datasource/etcd/state"
	"github.com/apache/servicecomb-service-center/datasource/etcd/state/kvstore"
	"github.com/apache/servicecomb-service-center/datasource/etcd/state/parser"
)

var (
	TypeSchemaRef     kvstore.Type
	TypeSchema        kvstore.Type
	TypeSchemaSummary kvstore.Type
)

func init() {
	TypeSchemaRef = state.MustRegister("SCHEMA_REF", path.GetServiceSchemaRefRootKey(""),
		state.WithInitSize(100),
		state.WithParser(parser.StringParser))
	TypeSchema = state.MustRegister("SCHEMA", path.GetServiceSchemaRootKey(""),
		state.WithInitSize(0))
	TypeSchemaSummary = state.MustRegister("SCHEMA_SUMMARY", path.GetServiceSchemaSummaryRootKey(""),
		state.WithInitSize(100),
		state.WithParser(parser.StringParser))
}

func SchemaRef() state.State     { return state.Get(TypeSchemaRef) }
func Schema() state.State        { return state.Get(TypeSchema) }
func SchemaSummary() state.State { return state.Get(TypeSchemaSummary) }
