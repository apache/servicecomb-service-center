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

package service

import (
	"github.com/apache/servicecomb-service-center/datasource"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/util"
	"github.com/apache/servicecomb-service-center/server/plugin/quota"
	pb "github.com/go-chassis/cari/discovery"

	"context"
)

func (s *MicroServiceService) GetSchemaInfo(ctx context.Context, in *pb.GetSchemaRequest) (*pb.GetSchemaResponse, error) {
	err := Validate(in)
	if err != nil {
		log.Errorf(nil, "get schema[%s/%s] failed", in.ServiceId, in.SchemaId)
		return &pb.GetSchemaResponse{
			Response: pb.CreateResponse(pb.ErrInvalidParams, err.Error()),
		}, nil
	}

	return datasource.Instance().GetSchema(ctx, in)
}

func (s *MicroServiceService) GetAllSchemaInfo(ctx context.Context, in *pb.GetAllSchemaRequest) (*pb.GetAllSchemaResponse, error) {
	err := Validate(in)
	if err != nil {
		log.Errorf(nil, "get service[%s] all schemas failed", in.ServiceId)
		return &pb.GetAllSchemaResponse{
			Response: pb.CreateResponse(pb.ErrInvalidParams, err.Error()),
		}, nil
	}

	return datasource.Instance().GetAllSchemas(ctx, in)
}

func (s *MicroServiceService) DeleteSchema(ctx context.Context, in *pb.DeleteSchemaRequest) (*pb.DeleteSchemaResponse, error) {
	err := Validate(in)
	if err != nil {
		remoteIP := util.GetIPFromContext(ctx)
		log.Errorf(err, "delete schema[%s/%s] failed, operator: %s", in.ServiceId, in.SchemaId, remoteIP)
		return &pb.DeleteSchemaResponse{
			Response: pb.CreateResponse(pb.ErrInvalidParams, err.Error()),
		}, nil
	}

	return datasource.Instance().DeleteSchema(ctx, in)
}

// ModifySchemas covers all the schemas of a service.
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
func (s *MicroServiceService) ModifySchemas(ctx context.Context, in *pb.ModifySchemasRequest) (*pb.ModifySchemasResponse, error) {
	err := Validate(in)
	if err != nil {
		remoteIP := util.GetIPFromContext(ctx)
		log.Errorf(err, "modify service[%s] schemas failed, operator: %s", in.ServiceId, remoteIP)
		return &pb.ModifySchemasResponse{
			Response: pb.CreateResponse(pb.ErrInvalidParams, "Invalid request."),
		}, nil
	}
	return datasource.Instance().ModifySchemas(ctx, in)
}

// ModifySchema modifies a specific schema
// 1. When the service is in production environment and schema is not editable:
// If the request contains a new schemaID (the number of schemaIDs of
// the service is also required to be 0, or the request will be rejected),
// the new schemaID will be automatically added to the service information.
// Schema is only allowed to add.
// 2. Other cases:
// If the request contains a new schemaID,
// the new schemaID will be automatically added to the service information.
// Schema is allowed to add/modify.
func (s *MicroServiceService) ModifySchema(ctx context.Context, request *pb.ModifySchemaRequest) (*pb.ModifySchemaResponse, error) {
	domainProject := util.ParseDomainProject(ctx)
	respErr := s.canModifySchema(ctx, domainProject, request)
	if respErr != nil {
		resp := &pb.ModifySchemaResponse{
			Response: pb.CreateResponseWithSCErr(respErr),
		}
		if respErr.InternalError() {
			return resp, respErr
		}
		return resp, nil
	}

	return datasource.Instance().ModifySchema(ctx, request)
}

func (s *MicroServiceService) canModifySchema(ctx context.Context, domainProject string, in *pb.ModifySchemaRequest) *pb.Error {
	remoteIP := util.GetIPFromContext(ctx)
	serviceID := in.ServiceId
	schemaID := in.SchemaId
	if len(schemaID) == 0 || len(serviceID) == 0 {
		log.Errorf(nil, "update schema[%s/%s] failed, invalid params, operator: %s",
			serviceID, schemaID, remoteIP)
		return pb.NewError(pb.ErrInvalidParams, "serviceID or schemaID is nil")
	}
	err := Validate(in)
	if err != nil {
		log.Errorf(err, "update schema[%s/%s] failed, operator: %s", serviceID, schemaID, remoteIP)
		return pb.NewError(pb.ErrInvalidParams, err.Error())
	}

	res := quota.NewApplyQuotaResource(quota.SchemaQuotaType, domainProject, serviceID, 1)
	rst := quota.Apply(ctx, res)
	errQuota := rst.Err
	if errQuota != nil {
		log.Errorf(errQuota, "update schema[%s/%s] failed, operator: %s", serviceID, schemaID, remoteIP)
		return errQuota
	}
	if len(in.Summary) == 0 {
		log.Warnf("schema[%s/%s]'s summary is empty, operator: %s", serviceID, schemaID, remoteIP)
	}
	return nil
}
