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
	"errors"
	"fmt"

	"github.com/apache/servicecomb-service-center/datasource"
	"github.com/apache/servicecomb-service-center/datasource/schema"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/util"
	quotasvc "github.com/apache/servicecomb-service-center/server/service/quota"
	"github.com/apache/servicecomb-service-center/server/service/validator"
	mapset "github.com/deckarep/golang-set"
	pb "github.com/go-chassis/cari/discovery"
)

const LOCAL = "local"

// ExistSchema only return the summary without content if schema exist
func ExistSchema(ctx context.Context, request *pb.GetSchemaRequest) (*pb.Schema, error) {
	remoteIP := util.GetIPFromContext(ctx)
	serviceID := request.ServiceId
	schemaID := request.SchemaId

	if checkErr := validator.ValidateGetSchema(request); checkErr != nil {
		log.Error(fmt.Sprintf("invalid get service[%s] schema[%s] request, operator: %s",
			serviceID, schemaID, remoteIP), nil)
		return nil, pb.NewError(pb.ErrInvalidParams, checkErr.Error())
	}

	ref, err := schema.Instance().GetRef(ctx, &schema.RefRequest{
		ServiceID: serviceID,
		SchemaID:  schemaID,
	})

	if err != nil {
		if errors.Is(err, schema.ErrSchemaNotFound) {
			return existOldSchema(ctx, request)
		}
		log.Error(fmt.Sprintf("get service[%s] schema-ref[%s] failed, operator: %s",
			serviceID, schemaID, remoteIP), nil)
		return nil, err
	}

	// return directly when using local fs
	if schema.StorageType == LOCAL {
		return &pb.Schema{
			SchemaId: schemaID,
			Schema:   ref.Content,
			Summary:  ref.Summary,
		}, nil
	}

	return &pb.Schema{
		SchemaId: schemaID,
		Summary:  ref.Summary,
	}, nil
}

func existOldSchema(ctx context.Context, request *pb.GetSchemaRequest) (*pb.Schema, error) {
	resp, err := datasource.GetMetadataManager().ExistSchema(ctx, &pb.GetExistenceRequest{
		Type:      datasource.ExistTypeSchema,
		ServiceId: request.ServiceId,
		SchemaId:  request.SchemaId,
	})
	if err != nil {
		return nil, err
	}
	return &pb.Schema{
		SchemaId: request.SchemaId,
		Summary:  resp.Summary,
	}, nil
}

func GetSchema(ctx context.Context, request *pb.GetSchemaRequest) (*pb.Schema, error) {
	remoteIP := util.GetIPFromContext(ctx)
	serviceID := request.ServiceId
	schemaID := request.SchemaId

	if checkErr := validator.ValidateGetSchema(request); checkErr != nil {
		log.Error(fmt.Sprintf("invalid get service[%s] schema[%s] request, operator: %s",
			serviceID, schemaID, remoteIP), nil)
		return nil, pb.NewError(pb.ErrInvalidParams, checkErr.Error())
	}

	ref, err := schema.Instance().GetRef(ctx, &schema.RefRequest{
		ServiceID: serviceID,
		SchemaID:  schemaID,
	})

	if err != nil {
		if errors.Is(err, schema.ErrSchemaNotFound) {
			return getOldSchema(ctx, request)
		}
		log.Error(fmt.Sprintf("get service[%s] schema-ref[%s] failed, operator: %s",
			serviceID, schemaID, remoteIP), nil)
		return nil, err
	}

	// return directly when using local fs
	if schema.StorageType == LOCAL {
		return &pb.Schema{
			SchemaId: schemaID,
			Schema:   ref.Content,
			Summary:  ref.Summary,
		}, nil
	}

	content, err := schema.Instance().GetContent(ctx, &schema.ContentRequest{
		Hash: ref.Hash,
	})
	if err != nil {
		log.Error(fmt.Sprintf("get service[%s] schema[%s] failed, operator: %s",
			serviceID, schemaID, remoteIP), nil)
		return nil, err
	}

	return &pb.Schema{
		SchemaId: schemaID,
		Schema:   content.Content,
		Summary:  ref.Summary,
	}, nil
}

func getOldSchema(ctx context.Context, request *pb.GetSchemaRequest) (*pb.Schema, error) {
	resp, err := datasource.GetMetadataManager().GetSchema(ctx, request)
	if err != nil {
		return nil, err
	}
	return &pb.Schema{
		SchemaId: request.SchemaId,
		Schema:   resp.Schema,
		Summary:  resp.SchemaSummary,
	}, nil
}

func ListSchema(ctx context.Context, request *pb.GetAllSchemaRequest) ([]*pb.Schema, error) {
	remoteIP := util.GetIPFromContext(ctx)
	serviceID := request.ServiceId

	if checkErr := validator.ValidateListSchema(request); checkErr != nil {
		log.Error(fmt.Sprintf("invalid list service[%s] schemas request, operator: %s", serviceID, remoteIP), nil)
		return nil, pb.NewError(pb.ErrInvalidParams, checkErr.Error())
	}

	schemaRefs, err := schema.Instance().ListRef(ctx, &schema.RefRequest{
		ServiceID: serviceID,
	})
	if err != nil {
		log.Error(fmt.Sprintf("list service[%s] schemaIDs failed, operator: %s", serviceID, remoteIP), nil)
		return nil, err
	}

	// return directly when using local fs
	if schema.StorageType == LOCAL {
		schemas := make([]*pb.Schema, 0, len(schemaRefs))
		for _, ref := range schemaRefs {
			item := &pb.Schema{
				SchemaId: ref.SchemaID,
				Summary:  ref.Summary,
				Schema:   ref.Content,
			}
			schemas = append(schemas, item)
		}
		return schemas, nil
	}

	oldSchemaIDs, err := getOldSchemaIDs(ctx, serviceID)
	if err != nil {
		log.Error(fmt.Sprintf("list service[%s] schemaIDs failed, operator: %s", serviceID, remoteIP), err)
		return nil, err
	}

	requests, err := mergeRequests(ctx, serviceID, schemaRefs, oldSchemaIDs)
	if err != nil {
		log.Error(fmt.Sprintf("list service[%s] schema-refs failed, operator: %s", serviceID, remoteIP), err)
		return nil, err
	}

	schemas := make([]*pb.Schema, 0, len(requests))
	for _, req := range requests {
		tmp, err := getSchema(ctx, req, request.WithSchema)
		if err != nil && !errors.Is(err, schema.ErrSchemaNotFound) {
			return nil, err
		}

		item := &pb.Schema{
			SchemaId: req.SchemaId,
		}
		if tmp != nil {
			item = tmp
		}
		schemas = append(schemas, item)
	}
	return schemas, nil
}

func getSchema(ctx context.Context, req *pb.GetSchemaRequest, withSchema bool) (*pb.Schema, error) {
	if withSchema {
		return GetSchema(ctx, req)
	}
	return ExistSchema(ctx, req)
}

func getOldSchemaIDs(ctx context.Context, serviceID string) ([]string, error) {
	resp, err := datasource.GetMetadataManager().GetAllSchemas(ctx, &pb.GetAllSchemaRequest{
		ServiceId:  serviceID,
		WithSchema: false,
	})
	if err != nil {
		return nil, err
	}
	schemaIDs := make([]string, 0, len(resp.Schemas))
	for _, item := range resp.Schemas {
		schemaIDs = append(schemaIDs, item.SchemaId)
	}
	return schemaIDs, nil
}

func mergeRequests(_ context.Context, serviceID string, refs []*schema.Ref, oldSchemaIDs []string) ([]*pb.GetSchemaRequest, error) {
	set := mapset.NewSet()
	for _, schemaID := range oldSchemaIDs {
		set.Add(schemaID)
	}
	for _, ref := range refs {
		set.Add(ref.SchemaID)
	}

	var requests []*pb.GetSchemaRequest
	for item := range set.Iter() {
		requests = append(requests, &pb.GetSchemaRequest{
			ServiceId: serviceID,
			SchemaId:  item.(string),
		})
	}
	return requests, nil
}

func DeleteSchema(ctx context.Context, request *pb.DeleteSchemaRequest) error {
	remoteIP := util.GetIPFromContext(ctx)

	if checkErr := validator.ValidateDeleteSchema(request); checkErr != nil {
		log.Error(fmt.Sprintf("invalid delete service[%s] schema[%s] request, operator: %s",
			request.ServiceId, request.SchemaId, remoteIP), checkErr)
		return pb.NewError(pb.ErrInvalidParams, checkErr.Error())
	}

	_, svcErr := datasource.GetMetadataManager().GetService(ctx, &pb.GetServiceRequest{
		ServiceId: request.ServiceId,
	})
	if svcErr != nil {
		log.Error(fmt.Sprintf("get service[%s] failed, operator: %s", request.ServiceId, remoteIP), svcErr)
		return svcErr
	}

	err := schema.Instance().DeleteRef(ctx, &schema.RefRequest{
		ServiceID: request.ServiceId,
		SchemaID:  request.SchemaId,
	})

	if err != nil {
		if errors.Is(err, schema.ErrSchemaNotFound) {
			return deleteOldSchema(ctx, request)
		}
		log.Error(fmt.Sprintf("delete service[%s] schema[%s] failed, operator: %s",
			request.ServiceId, request.SchemaId, remoteIP), err)
		return err
	}
	log.Info(fmt.Sprintf("delete service[%s] schema[%s], operator: %s", request.ServiceId, request.SchemaId, remoteIP))

	// return directly when using local fs
	if schema.StorageType == LOCAL {
		return err
	}

	err = deleteOldSchema(ctx, request)
	if err != nil && !errors.Is(err, schema.ErrSchemaNotFound) {
		log.Error(fmt.Sprintf("delete old service[%s] schema[%s] failed, operator: %s",
			request.ServiceId, request.SchemaId, remoteIP), svcErr)
		return err
	}
	return nil
}

func deleteOldSchema(ctx context.Context, request *pb.DeleteSchemaRequest) error {
	return datasource.GetMetadataManager().DeleteSchema(ctx, request)
}

// PutSchemas covers all the schemas of a service.
// To cover the old schemas, ModifySchemas adds new schemas into, delete and
// modify the old schemas.
func PutSchemas(ctx context.Context, request *pb.ModifySchemasRequest) error {
	remoteIP := util.GetIPFromContext(ctx)
	serviceID := request.ServiceId

	if checkErr := validator.ValidatePutSchemas(request); checkErr != nil {
		log.Error(fmt.Sprintf("invalid modify service[%s] schemas request, operator: %s", serviceID, remoteIP), checkErr)
		return pb.NewError(pb.ErrInvalidParams, "Invalid request.")
	}

	// no need to check quota usage because overwrite existing.
	apply := len(request.Schemas)
	schemaIDs := make([]string, 0, apply)
	contentItems := make([]*schema.ContentItem, 0, apply)
	for _, item := range request.Schemas {
		schemaIDs = append(schemaIDs, item.SchemaId)
		contentItems = append(contentItems, &schema.ContentItem{
			Hash:    schema.Hash(item.SchemaId, item.Schema),
			Content: item.Schema,
			Summary: item.Summary,
		})
	}

	err := schema.Instance().PutManyContent(ctx, &schema.PutManyContentRequest{
		ServiceID: serviceID,
		SchemaIDs: schemaIDs,
		Contents:  contentItems,
	})
	if err != nil {
		log.Error(fmt.Sprintf("put modify service[%s] schemas[len: %d] failed, operator: %s",
			serviceID, apply, remoteIP), err)
		return err
	}
	log.Info(fmt.Sprintf("put service[%s] schemas[len: %d], operator: %s", serviceID, apply, remoteIP))
	return nil
}

// PutSchema modifies a specific schema.
func PutSchema(ctx context.Context, request *pb.ModifySchemaRequest) error {
	remoteIP := util.GetIPFromContext(ctx)
	serviceID := request.ServiceId
	schemaID := request.SchemaId
	chars := len(request.Schema)

	if checkErr := validator.ValidatePutSchema(request); checkErr != nil {
		log.Error(fmt.Sprintf("invalid put service[%s] schemas[%s] request, operator: %s",
			serviceID, schemaID, remoteIP), checkErr)
		return pb.NewError(pb.ErrInvalidParams, checkErr.Error())
	}

	if quotaErr := checkSchemaQuota(ctx, serviceID, schemaID); quotaErr != nil {
		log.Error(fmt.Sprintf("check service[%s] schema quota failed, operator: %s", serviceID, remoteIP), quotaErr)
		return quotaErr
	}

	if len(request.Summary) == 0 {
		log.Warn(fmt.Sprintf("service[%s] schema[%s]'s summary is empty, operator: %s",
			serviceID, schemaID, remoteIP))
	}

	err := schema.Instance().PutContent(ctx,
		&schema.PutContentRequest{
			ServiceID: request.ServiceId,
			SchemaID:  request.SchemaId,
			Content: &schema.ContentItem{
				Hash:    schema.Hash(request.SchemaId, request.Schema),
				Content: request.Schema,
				Summary: request.Summary,
			},
		})
	if err != nil {
		log.Error(fmt.Sprintf("put service[%s] schema[%s chars: %d] failed, operator: %s",
			serviceID, schemaID, chars, remoteIP), err)
		return err
	}
	log.Info(fmt.Sprintf("put service[%s] schema[%s, chars: %d], operator: %s", serviceID, schemaID, chars, remoteIP))
	return nil
}

func checkSchemaQuota(ctx context.Context, serviceID string, schemaID string) error {
	service, err := datasource.GetMetadataManager().GetService(ctx, &pb.GetServiceRequest{
		ServiceId: serviceID,
	})
	if err != nil {
		return err
	}

	if util.SliceHave(service.Schemas, schemaID) {
		return nil
	}

	if errQuota := quotasvc.ApplySchema(ctx, serviceID, 1); errQuota != nil {
		return errQuota
	}
	return nil
}

func Usage(ctx context.Context, serviceID string) (int64, error) {
	schemas, err := ListSchema(ctx, &pb.GetAllSchemaRequest{
		ServiceId:  serviceID,
		WithSchema: false,
	})
	if err != nil {
		return 0, err
	}
	return int64(len(schemas)), nil
}
