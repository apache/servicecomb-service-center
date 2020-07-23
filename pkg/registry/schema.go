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

package registry

type ModifySchemasRequest struct {
	ServiceId string    `protobuf:"bytes,1,opt,name=serviceId" json:"serviceId,omitempty"`
	Schemas   []*Schema `protobuf:"bytes,2,rep,name=schemas" json:"schemas,omitempty"`
}

type Schema struct {
	SchemaId string `protobuf:"bytes,1,opt,name=schemaId" json:"schemaId,omitempty"`
	Summary  string `protobuf:"bytes,2,opt,name=summary" json:"summary,omitempty"`
	Schema   string `protobuf:"bytes,3,opt,name=schema" json:"schema,omitempty"`
}

type ModifySchemasResponse struct {
	Response *Response `protobuf:"bytes,1,opt,name=response" json:"response,omitempty"`
}
type ModifySchemaRequest struct {
	ServiceId string `protobuf:"bytes,1,opt,name=serviceId" json:"serviceId,omitempty"`
	SchemaId  string `protobuf:"bytes,2,opt,name=schemaId" json:"schemaId,omitempty"`
	Schema    string `protobuf:"bytes,3,opt,name=schema" json:"schema,omitempty"`
	Summary   string `protobuf:"bytes,4,opt,name=summary" json:"summary,omitempty"`
}
type ModifySchemaResponse struct {
	Response *Response `protobuf:"bytes,1,opt,name=response" json:"response,omitempty"`
}

type GetSchemaRequest struct {
	ServiceId string `protobuf:"bytes,1,opt,name=serviceId" json:"serviceId,omitempty"`
	SchemaId  string `protobuf:"bytes,2,opt,name=schemaId" json:"schemaId,omitempty"`
}

type GetAllSchemaRequest struct {
	ServiceId  string `protobuf:"bytes,1,opt,name=serviceId" json:"serviceId,omitempty"`
	WithSchema bool   `protobuf:"varint,2,opt,name=withSchema" json:"withSchema,omitempty"`
}

type GetSchemaResponse struct {
	Response      *Response `protobuf:"bytes,1,opt,name=response" json:"response,omitempty"`
	Schema        string    `protobuf:"bytes,2,opt,name=schema" json:"schema,omitempty"`
	SchemaSummary string    `protobuf:"bytes,3,opt,name=schemaSummary" json:"schemaSummary,omitempty"`
}

type GetAllSchemaResponse struct {
	Response *Response `protobuf:"bytes,1,opt,name=response" json:"response,omitempty"`
	Schemas  []*Schema `protobuf:"bytes,2,rep,name=schemas" json:"schemas,omitempty"`
}

type DeleteSchemaRequest struct {
	ServiceId string `protobuf:"bytes,1,opt,name=serviceId" json:"serviceId,omitempty"`
	SchemaId  string `protobuf:"bytes,2,opt,name=schemaId" json:"schemaId,omitempty"`
}

type DeleteSchemaResponse struct {
	Response *Response `protobuf:"bytes,1,opt,name=response" json:"response,omitempty"`
}
