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

package disco

import (
	"context"
	"fmt"

	"github.com/apache/servicecomb-service-center/datasource"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/util"
	"github.com/apache/servicecomb-service-center/server/plugin/quota"
	"github.com/apache/servicecomb-service-center/server/service/validator"
	pb "github.com/go-chassis/cari/discovery"
)

func GetSchema(ctx context.Context, in *pb.GetSchemaRequest) (*pb.GetSchemaResponse, error) {
	err := validator.Validate(in)
	if err != nil {
		log.Error(fmt.Sprintf("get schema[%s/%s] failed", in.ServiceId, in.SchemaId), nil)
		return nil, pb.NewError(pb.ErrInvalidParams, err.Error())
	}

	return datasource.GetMetadataManager().GetSchema(ctx, in)
}

func ListSchema(ctx context.Context, in *pb.GetAllSchemaRequest) (*pb.GetAllSchemaResponse, error) {
	err := validator.Validate(in)
	if err != nil {
		log.Error(fmt.Sprintf("get service[%s] all schemas failed", in.ServiceId), nil)
		return nil, pb.NewError(pb.ErrInvalidParams, err.Error())
	}

	return datasource.GetMetadataManager().GetAllSchemas(ctx, in)
}

func DeleteSchema(ctx context.Context, in *pb.DeleteSchemaRequest) (*pb.DeleteSchemaResponse, error) {
	err := validator.Validate(in)
	if err != nil {
		remoteIP := util.GetIPFromContext(ctx)
		log.Error(fmt.Sprintf("delete schema[%s/%s] failed, operator: %s", in.ServiceId, in.SchemaId, remoteIP), err)
		return nil, pb.NewError(pb.ErrInvalidParams, err.Error())
	}

	return datasource.GetMetadataManager().DeleteSchema(ctx, in)
}

// PutSchemas covers all the schemas of a service.
// To cover the old schemas, ModifySchemas adds new schemas into, delete and
// modify the old schemas.
// 1. When the service is in production environment and schema is not editable:
// If the request contains a new schemaID (the number of schemaIDs of
// the service is also required to be 0, or the request will be rejected),
// the new schemaID will be automatically added to the service information.
// Schema is only allowed to add.
// 2. Other cases:
// If the request contains a new schemaID,
// the new schemaID will be automatically added to the service information.
// Schema is allowed to add/delete/modify.
func PutSchemas(ctx context.Context, in *pb.ModifySchemasRequest) (*pb.ModifySchemasResponse, error) {
	err := validator.Validate(in)
	if err != nil {
		remoteIP := util.GetIPFromContext(ctx)
		log.Error(fmt.Sprintf("modify service[%s] schemas failed, operator: %s", in.ServiceId, remoteIP), err)
		return nil, pb.NewError(pb.ErrInvalidParams, "Invalid request.")
	}
	return datasource.GetMetadataManager().ModifySchemas(ctx, in)
}

// PutSchema modifies a specific schema
// 1. When the service is in production environment and schema is not editable:
// If the request contains a new schemaID (the number of schemaIDs of
// the service is also required to be 0, or the request will be rejected),
// the new schemaID will be automatically added to the service information.
// Schema is only allowed to add.
// 2. Other cases:
// If the request contains a new schemaID,
// the new schemaID will be automatically added to the service information.
// Schema is allowed to add/modify.
func PutSchema(ctx context.Context, request *pb.ModifySchemaRequest) (*pb.ModifySchemaResponse, error) {
	domainProject := util.ParseDomainProject(ctx)
	err := canModifySchema(ctx, domainProject, request)
	if err != nil {
		return nil, err
	}

	return datasource.GetMetadataManager().ModifySchema(ctx, request)
}

func canModifySchema(ctx context.Context, domainProject string, in *pb.ModifySchemaRequest) error {
	remoteIP := util.GetIPFromContext(ctx)
	serviceID := in.ServiceId
	schemaID := in.SchemaId
	if len(schemaID) == 0 || len(serviceID) == 0 {
		log.Error(fmt.Sprintf("update schema[%s/%s] failed, invalid params, operator: %s",
			serviceID, schemaID, remoteIP), nil)
		return pb.NewError(pb.ErrInvalidParams, "serviceID or schemaID is nil")
	}
	err := validator.Validate(in)
	if err != nil {
		log.Error(fmt.Sprintf("update schema[%s/%s] failed, operator: %s", serviceID, schemaID, remoteIP), err)
		return pb.NewError(pb.ErrInvalidParams, err.Error())
	}

	res := quota.NewApplyQuotaResource(quota.TypeSchema, domainProject, serviceID, 1)
	if errQuota := quota.Apply(ctx, res); errQuota != nil {
		log.Error(fmt.Sprintf("update schema[%s/%s] failed, operator: %s", serviceID, schemaID, remoteIP), errQuota)
		return errQuota
	}
	if len(in.Summary) == 0 {
		log.Warn(fmt.Sprintf("schema[%s/%s]'s summary is empty, operator: %s", serviceID, schemaID, remoteIP))
	}
	return nil
}
