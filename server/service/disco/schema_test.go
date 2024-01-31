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
package disco_test

import (
	"context"
	"strconv"
	"strings"
	"testing"

	pb "github.com/go-chassis/cari/discovery"
	"github.com/go-chassis/cari/pkg/errsvc"
	csync "github.com/go-chassis/cari/sync"
	"github.com/stretchr/testify/assert"

	"github.com/apache/servicecomb-service-center/datasource"
	"github.com/apache/servicecomb-service-center/datasource/schema"
	"github.com/apache/servicecomb-service-center/eventbase/model"
	"github.com/apache/servicecomb-service-center/eventbase/service/task"
	"github.com/apache/servicecomb-service-center/eventbase/service/tombstone"
	"github.com/apache/servicecomb-service-center/pkg/util"
	"github.com/apache/servicecomb-service-center/server/service/disco"
	quotasvc "github.com/apache/servicecomb-service-center/server/service/quota"
	"github.com/apache/servicecomb-service-center/test"
)

const (
	invalidSchemaId = "@"
)

var (
	TooLongSummary = strings.Repeat("x", 129)
)

func TestPutSchema(t *testing.T) {
	var (
		serviceId    string
		serviceIdDev string
		serviceIdPro string
	)
	ctx := getContext()
	defer disco.UnregisterManyService(ctx, &pb.DelServicesRequest{
		ServiceIds: []string{serviceId, serviceIdDev, serviceIdPro}, Force: true,
	})

	t.Run("should be passed, create service", func(t *testing.T) {
		respCreateService, err := datasource.GetMetadataManager().RegisterService(ctx, &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				AppId:       "create_schema_group",
				ServiceName: "create_schema_service",
				Version:     "1.0.0",
				Level:       "FRONT",
				Status:      pb.MS_UP,
				Environment: pb.ENV_DEV,
			},
		})
		assert.NoError(t, err)
		serviceIdDev = respCreateService.ServiceId

		respCreateService, err = datasource.GetMetadataManager().RegisterService(ctx, &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				AppId:       "create_schema_group",
				ServiceName: "create_schema_service",
				Version:     "1.0.0",
				Level:       "FRONT",
				Status:      pb.MS_UP,
				Environment: pb.ENV_PROD,
			},
		})
		assert.NoError(t, err)
		serviceId = respCreateService.ServiceId

		respCreateService, err = datasource.GetMetadataManager().RegisterService(ctx, &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				AppId:       "create_schemas_prod",
				ServiceName: "create_schemas_service",
				Version:     "1.0.1",
				Level:       "FRONT",
				Schemas: []string{
					"first_schemaId",
					"second_schemaId",
				},
				Status:      pb.MS_UP,
				Environment: pb.ENV_PROD,
			},
		})
		assert.NoError(t, err)
		serviceIdPro = respCreateService.ServiceId
	})

	defer func() {
		datasource.GetMetadataManager().UnregisterService(ctx, &pb.DeleteServiceRequest{ServiceId: serviceIdDev, Force: true})
		datasource.GetMetadataManager().UnregisterService(ctx, &pb.DeleteServiceRequest{ServiceId: serviceId, Force: true})
		datasource.GetMetadataManager().UnregisterService(ctx, &pb.DeleteServiceRequest{ServiceId: serviceIdPro, Force: true})
	}()

	f := func() {
		err := disco.PutSchema(ctx, &pb.ModifySchemaRequest{
			ServiceId: "",
			SchemaId:  "com.huawei.test",
			Schema:    "create schema",
		})
		testErr := err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrInvalidParams, testErr.Code)

		err = disco.PutSchema(ctx, &pb.ModifySchemaRequest{
			ServiceId: "notExistService",
			SchemaId:  "com.huawei.test",
			Schema:    "create schema",
		})
		testErr = err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrServiceNotExists, testErr.Code)

		err = disco.PutSchema(ctx, &pb.ModifySchemaRequest{
			ServiceId: serviceIdDev,
			SchemaId:  invalidSchemaId,
			Schema:    "create schema",
		})
		testErr = err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrInvalidParams, testErr.Code)

		err = disco.PutSchema(ctx, &pb.ModifySchemaRequest{
			ServiceId: serviceIdDev,
			SchemaId:  "com.huawei.test",
			Schema:    "create schema",
			Summary:   TooLongSummary,
		})
		testErr = err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrInvalidParams, testErr.Code)

		err = disco.PutSchema(ctx, &pb.ModifySchemaRequest{
			ServiceId: serviceIdDev,
			SchemaId:  "com.huawei.test",
			Schema:    "create schema",
			Summary:   "_",
		})
		testErr = err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrInvalidParams, testErr.Code)
	}

	t.Run("should be failed in prod env, when create an invalid schema", func(t *testing.T) {
		old := serviceIdDev
		serviceIdDev = serviceId
		f()
		serviceIdDev = old
	})

	t.Run("should be failed in dev env, when create an invalid schema", func(t *testing.T) {
		f()
	})

	f = func() {
		err := disco.PutSchemas(ctx, &pb.ModifySchemasRequest{
			ServiceId: serviceIdDev,
		})
		testErr := err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrInvalidParams, testErr.Code)

		err = disco.PutSchemas(ctx, &pb.ModifySchemasRequest{
			ServiceId: "",
			Schemas: []*pb.Schema{
				{
					SchemaId: "com.huawei.test",
					Summary:  "create schema",
					Schema:   "create schema",
				},
			},
		})
		testErr = err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrInvalidParams, testErr.Code)

		err = disco.PutSchemas(ctx, &pb.ModifySchemasRequest{
			ServiceId: serviceIdDev,
			Schemas: []*pb.Schema{
				{
					SchemaId: invalidSchemaId,
					Summary:  "create schema",
					Schema:   "create schema",
				},
			},
		})
		testErr = err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrInvalidParams, testErr.Code)

		err = disco.PutSchemas(ctx, &pb.ModifySchemasRequest{
			ServiceId: serviceIdDev,
			Schemas: []*pb.Schema{
				{
					SchemaId: "com.huawei.test",
					Summary:  "create schema",
				},
			},
		})
		testErr = err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrInvalidParams, testErr.Code)

		err = disco.PutSchemas(ctx, &pb.ModifySchemasRequest{
			ServiceId: serviceIdDev,
			Schemas: []*pb.Schema{
				{
					SchemaId: "com.huawei.test",
					Schema:   "create schema",
				},
			},
		})
		testErr = err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrInvalidParams, testErr.Code)

		err = disco.PutSchemas(ctx, &pb.ModifySchemasRequest{
			ServiceId: serviceIdDev,
			Schemas: []*pb.Schema{
				{
					SchemaId: "com.huawei.test",
					Schema:   "create schema",
					Summary:  TooLongSummary,
				},
			},
		})
		testErr = err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrInvalidParams, testErr.Code)
	}

	t.Run("should be failed in dev env, when batch create invalid schemas", func(t *testing.T) {
		f()
	})

	t.Run("should be failed in prod env, when batch create invalid schemas", func(t *testing.T) {
		old := serviceIdDev
		serviceIdDev = serviceId
		f()
		serviceIdDev = old
	})

	max := int(quotasvc.SchemaQuota())
	size := max + 1
	schemaIds := make([]string, 0, size)
	schemas := make([]*pb.Schema, 0, size)
	for i := 0; i < size; i++ {
		s := "ServiceCombTestTheLimitOfSchemas" + strconv.Itoa(i)
		schemaIds = append(schemaIds, s)
		schemas = append(schemas, &pb.Schema{
			SchemaId: s,
			Schema:   s,
			Summary:  s,
		})
	}

	t.Run("should be failed, when put schemas out of gauge", func(t *testing.T) {
		err := disco.PutSchemas(ctx, &pb.ModifySchemasRequest{
			ServiceId: serviceIdDev,
			Schemas:   schemas,
		})
		testErr := err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrInvalidParams, testErr.Code)
	})

	err := disco.PutSchemas(ctx, &pb.ModifySchemasRequest{
		ServiceId: serviceIdDev,
		Schemas:   schemas[:max],
	})
	assert.NoError(t, err)

	t.Run("should be failed, when put schema out of gauge", func(t *testing.T) {
		schema := schemas[max]
		err = disco.PutSchema(ctx, &pb.ModifySchemaRequest{
			ServiceId: serviceIdDev,
			SchemaId:  schema.SchemaId,
			Schema:    schema.Schema,
		})
		testErr := err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrNotEnoughQuota, testErr.Code)
	})

	t.Run("should be ok, when put a exist schema", func(t *testing.T) {
		schema := schemas[0]
		err = disco.PutSchema(ctx, &pb.ModifySchemaRequest{
			ServiceId: serviceIdDev,
			SchemaId:  schema.SchemaId,
			Schema:    schema.Schema,
		})
		assert.NoError(t, err)
	})

	t.Run("should be failed when create service with exceed schemaIDs", func(t *testing.T) {
		_, err := disco.RegisterService(ctx, &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				AppId:       "check_schema_group",
				ServiceName: "check_schema_service",
				Version:     "1.0.0",
				Level:       "FRONT",
				Schemas:     schemaIds,
				Status:      pb.MS_UP,
			},
		})
		testErr := err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrInvalidParams, testErr.Code)
	})

	t.Run("should be failed, when modify schema and summary is empty", func(t *testing.T) {
		err := disco.PutSchema(ctx, &pb.ModifySchemaRequest{
			ServiceId: serviceIdPro,
			SchemaId:  "first_schemaId",
			Schema:    "first_schema",
		})
		assert.NoError(t, err)

		schemas := []*pb.Schema{
			{
				SchemaId: "first_schemaId",
				Schema:   "first_schema",
				Summary:  "first0summary",
			},
		}
		err = disco.PutSchemas(ctx, &pb.ModifySchemasRequest{
			ServiceId: serviceIdPro,
			Schemas:   schemas,
		})
		assert.NoError(t, err)

		respExist, err := disco.ExistSchema(ctx, &pb.GetSchemaRequest{
			ServiceId: serviceIdPro,
			SchemaId:  "first_schemaId",
		})
		assert.NoError(t, err)
		assert.Equal(t, "first0summary", respExist.Summary)

		schemas = []*pb.Schema{
			{
				SchemaId: "second_schemaId",
				Schema:   "second_schema",
				Summary:  "second0summary",
			},
		}
		err = disco.PutSchemas(ctx, &pb.ModifySchemasRequest{
			ServiceId: serviceIdPro,
			Schemas:   schemas,
		})
		assert.NoError(t, err)
	})
}

func TestPutSchemas(t *testing.T) {
	var (
		serviceIdDev1 string
		serviceIdDev2 string
	)
	respCreateService, err := datasource.GetMetadataManager().RegisterService(getContext(), &pb.CreateServiceRequest{
		Service: &pb.MicroService{
			AppId:       "create_schemas_dev",
			ServiceName: "create_schemas_service",
			Version:     "1.0.0",
			Level:       "FRONT",
			Status:      pb.MS_UP,
			Environment: pb.ENV_DEV,
		},
	})
	assert.NoError(t, err)
	serviceIdDev1 = respCreateService.ServiceId
	defer datasource.GetMetadataManager().UnregisterService(getContext(), &pb.DeleteServiceRequest{ServiceId: serviceIdDev1, Force: true})

	respCreateService, err = datasource.GetMetadataManager().RegisterService(getContext(), &pb.CreateServiceRequest{
		Service: &pb.MicroService{
			AppId:       "create_schemas_dev",
			ServiceName: "create_schemas_service",
			Version:     "1.0.1",
			Level:       "FRONT",
			Schemas: []string{
				"first_schemaId",
			},
			Status:      pb.MS_UP,
			Environment: pb.ENV_DEV,
		},
	})
	assert.NoError(t, err)
	serviceIdDev2 = respCreateService.ServiceId
	defer datasource.GetMetadataManager().UnregisterService(getContext(), &pb.DeleteServiceRequest{ServiceId: serviceIdDev2, Force: true})

	t.Run("should be passed, when create schemas when service schema id set is empty", func(t *testing.T) {
		schemas := []*pb.Schema{
			{
				SchemaId: "first_schemaId",
				Schema:   "first_schema",
				Summary:  "first0summary",
			},
		}
		err := disco.PutSchemas(getContext(), &pb.ModifySchemasRequest{
			ServiceId: serviceIdDev1,
			Schemas:   schemas,
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respCreateService.Response.GetCode())

		schemas, err = disco.ListSchema(getContext(), &pb.GetAllSchemaRequest{
			ServiceId: serviceIdDev1,
		})
		assert.NoError(t, err)
		assert.Equal(t, 1, len(schemas))

		schemas = []*pb.Schema{}
		err = disco.PutSchemas(getContext(), &pb.ModifySchemasRequest{
			ServiceId: serviceIdDev2,
			Schemas:   schemas,
		})
		testErr := err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrInvalidParams, testErr.Code)

		schemas = []*pb.Schema{
			{
				SchemaId: "first_schemaId",
				Schema:   "first_schema_change",
				Summary:  "first0summary1change",
			},
		}
		err = disco.PutSchemas(getContext(), &pb.ModifySchemasRequest{
			ServiceId: serviceIdDev1,
			Schemas:   schemas,
		})
		assert.NoError(t, err)

		schemas = []*pb.Schema{
			{
				SchemaId: "second_schemaId",
				Schema:   "second_schema",
				Summary:  "second0summary",
			},
		}
		err = disco.PutSchemas(getContext(), &pb.ModifySchemasRequest{
			ServiceId: serviceIdDev1,
			Schemas:   schemas,
		})
		assert.NoError(t, err)

		service, err := disco.GetService(getContext(), &pb.GetServiceRequest{
			ServiceId: serviceIdDev1,
		})
		assert.NoError(t, err)
		assert.Equal(t, []string{"second_schemaId"}, service.Schemas)

		schemas = []*pb.Schema{}
		err = disco.PutSchemas(getContext(), &pb.ModifySchemasRequest{
			ServiceId: serviceIdDev1,
			Schemas:   schemas,
		})
		testErr = err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrInvalidParams, testErr.Code)

		schemas = []*pb.Schema{
			{
				SchemaId: "second_schemaId",
				Schema:   "second_schema",
				Summary:  "second0summary",
			},
		}
		err = disco.PutSchemas(getContext(), &pb.ModifySchemasRequest{
			ServiceId: serviceIdDev2,
			Schemas:   schemas,
		})
		assert.NoError(t, err)

		service, err = disco.GetService(getContext(), &pb.GetServiceRequest{
			ServiceId: serviceIdDev2,
		})
		assert.NoError(t, err)
		assert.Equal(t, []string{"second_schemaId"}, service.Schemas)
	})
}

func TestExistSchema(t *testing.T) {
	var (
		serviceId string
	)
	respCreateService, err := datasource.GetMetadataManager().RegisterService(getContext(), &pb.CreateServiceRequest{
		Service: &pb.MicroService{
			AppId:       "query_schema_group",
			ServiceName: "query_schema_service",
			Version:     "1.0.0",
			Level:       "FRONT",
			Status:      pb.MS_UP,
			Environment: pb.ENV_DEV,
		},
	})
	assert.NoError(t, err)
	serviceId = respCreateService.ServiceId
	defer datasource.GetMetadataManager().UnregisterService(getContext(), &pb.DeleteServiceRequest{ServiceId: serviceId, Force: true})

	err = disco.PutSchema(getContext(), &pb.ModifySchemaRequest{
		ServiceId: serviceId,
		SchemaId:  "com.huawei.test",
		Schema:    "query schema",
		Summary:   "summary",
	})
	assert.NoError(t, err)

	err = disco.PutSchema(getContext(), &pb.ModifySchemaRequest{
		ServiceId: serviceId,
		SchemaId:  "com.huawei.test.no.summary",
		Schema:    "query schema",
	})
	assert.NoError(t, err)

	t.Run("should be failed, when request is invalid", func(t *testing.T) {
		_, err := disco.ExistSchema(getContext(), &pb.GetSchemaRequest{
			ServiceId: "",
			SchemaId:  "com.huawei.test",
		})
		testErr := err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrInvalidParams, testErr.Code)

		_, err = disco.ExistSchema(getContext(), &pb.GetSchemaRequest{
			ServiceId: serviceId,
			SchemaId:  "noneschema",
		})
		testErr = err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrSchemaNotExists, testErr.Code)

		_, err = disco.ExistSchema(getContext(), &pb.GetSchemaRequest{
			ServiceId: serviceId,
			SchemaId:  TooLongSchemaID,
		})
		testErr = err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrInvalidParams, testErr.Code)
	})

	t.Run("should be passed, when request is valid", func(t *testing.T) {
		resp, err := disco.ExistSchema(getContext(), &pb.GetSchemaRequest{
			ServiceId: serviceId,
			SchemaId:  "com.huawei.test",
		})
		assert.NoError(t, err)
		assert.Equal(t, "summary", resp.Summary)

		resp, err = disco.ExistSchema(getContext(), &pb.GetSchemaRequest{
			ServiceId: serviceId,
			SchemaId:  "com.huawei.test.no.summary",
		})
		assert.NoError(t, err)
		assert.Equal(t, "com.huawei.test.no.summary", resp.SchemaId)
		assert.Equal(t, "", resp.Summary)
	})
}

func TestGetSchema(t *testing.T) {
	var (
		serviceId     string
		serviceId1    string
		schemaId1     = "all_schema1"
		schemaId2     = "all_schema2"
		schemaId3     = "all_schema3"
		summary       = "this0is1a2test"
		schemaContent = "the content is vary large"
	)

	respCreateService, err := datasource.GetMetadataManager().RegisterService(getContext(), &pb.CreateServiceRequest{
		Service: &pb.MicroService{
			AppId:       "get_schema_group",
			ServiceName: "get_schema_service",
			Version:     "1.0.0",
			Level:       "FRONT",
			Schemas: []string{
				"non-schema-content",
			},
			Status:      pb.MS_UP,
			Environment: pb.ENV_DEV,
		},
	})
	assert.NoError(t, err)
	serviceId = respCreateService.ServiceId
	defer datasource.GetMetadataManager().UnregisterService(getContext(), &pb.DeleteServiceRequest{ServiceId: serviceId, Force: true})

	respCreateService, err = datasource.GetMetadataManager().RegisterService(getContext(), &pb.CreateServiceRequest{
		Service: &pb.MicroService{
			AppId:       "get_all_schema",
			ServiceName: "get_all_schema",
			Version:     "1.0.0",
			Level:       "FRONT",
			Schemas: []string{
				schemaId1,
				schemaId2,
				schemaId3,
			},
			Status: pb.MS_UP,
		},
	})
	assert.NoError(t, err)
	serviceId1 = respCreateService.ServiceId
	defer datasource.GetMetadataManager().UnregisterService(getContext(), &pb.DeleteServiceRequest{ServiceId: serviceId1, Force: true})

	err = disco.PutSchema(getContext(), &pb.ModifySchemaRequest{
		ServiceId: serviceId,
		SchemaId:  "com.huawei.test",
		Schema:    "get schema",
		Summary:   "schema0summary",
	})
	assert.NoError(t, err)

	err = disco.PutSchema(getContext(), &pb.ModifySchemaRequest{
		ServiceId: serviceId1,
		SchemaId:  schemaId2,
		Schema:    schemaContent,
	})
	assert.NoError(t, err)

	err = disco.PutSchema(getContext(), &pb.ModifySchemaRequest{
		ServiceId: serviceId1,
		SchemaId:  schemaId3,
		Schema:    schemaContent,
		Summary:   summary,
	})
	assert.NoError(t, err)

	t.Run("should be pass, when list schema", func(t *testing.T) {
		schemas, err := disco.ListSchema(getContext(), &pb.GetAllSchemaRequest{
			ServiceId:  serviceId1,
			WithSchema: false,
		})
		assert.NoError(t, err)
		for _, schema := range schemas {
			if schema.SchemaId == schemaId1 {
				assert.Empty(t, schema.Summary)
				assert.Empty(t, schema.Schema)
			}
			if schema.SchemaId == schemaId2 {
				assert.Empty(t, schema.Summary)
				assert.Empty(t, schema.Schema)
			}
			if schema.SchemaId == schemaId3 {
				assert.Equal(t, summary, schema.Summary)
				assert.Empty(t, schema.Schema)
			}
		}

		schemas, err = disco.ListSchema(getContext(), &pb.GetAllSchemaRequest{
			ServiceId:  serviceId1,
			WithSchema: true,
		})
		assert.NoError(t, err)
		for _, schema := range schemas {
			if schema.SchemaId == schemaId1 {
				assert.Empty(t, schema.Summary)
				assert.Empty(t, schema.Schema)
			}
			if schema.SchemaId == schemaId2 {
				assert.Empty(t, schema.Summary)
				assert.Equal(t, schemaContent, schema.Schema)
			}
			if schema.SchemaId == schemaId3 {
				assert.Equal(t, summary, schema.Summary)
				assert.Equal(t, schemaContent, schema.Schema)
			}
		}

	})

	t.Run("should be failed, when request is invalid", func(t *testing.T) {
		_, err := disco.GetSchema(getContext(), &pb.GetSchemaRequest{
			ServiceId: "noneexistservice",
			SchemaId:  "com.huawei.test",
		})
		testErr := err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrServiceNotExists, testErr.Code)

		_, err = disco.ListSchema(getContext(), &pb.GetAllSchemaRequest{
			ServiceId: "noneexistservice",
		})
		testErr = err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrServiceNotExists, testErr.Code)

		_, err = disco.GetSchema(getContext(), &pb.GetSchemaRequest{
			ServiceId: "",
			SchemaId:  "com.huawei.test",
		})
		testErr = err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrInvalidParams, testErr.Code)

		_, err = disco.ListSchema(getContext(), &pb.GetAllSchemaRequest{
			ServiceId: "",
		})
		testErr = err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrInvalidParams, testErr.Code)

		_, err = disco.GetSchema(getContext(), &pb.GetSchemaRequest{
			ServiceId: TooLongServiceID,
			SchemaId:  "com.huawei.test",
		})
		testErr = err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrInvalidParams, testErr.Code)

		_, err = disco.ListSchema(getContext(), &pb.GetAllSchemaRequest{
			ServiceId: TooLongServiceID,
		})
		testErr = err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrInvalidParams, testErr.Code)

		_, err = disco.GetSchema(getContext(), &pb.GetSchemaRequest{
			ServiceId: serviceId,
			SchemaId:  "nonexistschema",
		})
		testErr = err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrSchemaNotExists, testErr.Code)

		_, err = disco.GetSchema(getContext(), &pb.GetSchemaRequest{
			ServiceId: serviceId,
			SchemaId:  "",
		})
		testErr = err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrInvalidParams, testErr.Code)

		_, err = disco.GetSchema(getContext(), &pb.GetSchemaRequest{
			ServiceId: serviceId,
			SchemaId:  TooLongSchemaID,
		})
		testErr = err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrInvalidParams, testErr.Code)

		_, err = disco.GetSchema(getContext(), &pb.GetSchemaRequest{
			ServiceId: serviceId,
			SchemaId:  invalidSchemaId,
		})
		testErr = err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrInvalidParams, testErr.Code)

		_, err = disco.GetSchema(getContext(), &pb.GetSchemaRequest{
			ServiceId: serviceId,
			SchemaId:  "non-schema-content",
		})
		testErr = err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrSchemaNotExists, testErr.Code)
	})

	t.Run("should be passed, when request is valid", func(t *testing.T) {
		schema, err := disco.GetSchema(getContext(), &pb.GetSchemaRequest{
			ServiceId: serviceId,
			SchemaId:  "com.huawei.test",
		})
		assert.NoError(t, err)
		assert.NotNil(t, schema)
		assert.Equal(t, "com.huawei.test", schema.SchemaId)
		assert.Equal(t, "get schema", schema.Schema)
		assert.Equal(t, "schema0summary", schema.Summary)
	})
}

func TestDeleteSchema(t *testing.T) {
	var (
		serviceId string
	)
	respCreateService, err := datasource.GetMetadataManager().RegisterService(getContext(), &pb.CreateServiceRequest{
		Service: &pb.MicroService{
			AppId:       "delete_schema_group",
			ServiceName: "delete_schema_service",
			Version:     "1.0.0",
			Level:       "FRONT",
			Status:      pb.MS_UP,
		},
	})
	assert.NoError(t, err)
	serviceId = respCreateService.ServiceId
	defer datasource.GetMetadataManager().UnregisterService(getContext(), &pb.DeleteServiceRequest{ServiceId: serviceId, Force: true})

	err = disco.PutSchema(getContext(), &pb.ModifySchemaRequest{
		ServiceId: serviceId,
		SchemaId:  "com.huawei.test",
		Schema:    "delete schema",
		Summary:   "summary",
	})
	assert.NoError(t, err)

	t.Run("should be failed, when request is invalid", func(t *testing.T) {
		err := disco.DeleteSchema(getContext(), &pb.DeleteSchemaRequest{
			ServiceId: serviceId,
			SchemaId:  "noneschema",
		})
		testErr := err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrSchemaNotExists, testErr.Code)

		err = disco.DeleteSchema(getContext(), &pb.DeleteSchemaRequest{
			ServiceId: "",
			SchemaId:  "com.huawei.test",
		})
		testErr = err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrInvalidParams, testErr.Code)

		err = disco.DeleteSchema(getContext(), &pb.DeleteSchemaRequest{
			ServiceId: "noexistservice",
			SchemaId:  "com.huawei.test",
		})
		testErr = err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrServiceNotExists, testErr.Code)

		err = disco.DeleteSchema(getContext(), &pb.DeleteSchemaRequest{
			ServiceId: serviceId,
			SchemaId:  invalidSchemaId,
		})
		testErr = err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrInvalidParams, testErr.Code)
	})

	t.Run("should be passed, when request is valid", func(t *testing.T) {
		err := disco.DeleteSchema(getContext(), &pb.DeleteSchemaRequest{
			ServiceId: serviceId,
			SchemaId:  "com.huawei.test",
		})
		assert.NoError(t, err)

		_, err = disco.GetSchema(getContext(), &pb.GetSchemaRequest{
			ServiceId: serviceId,
			SchemaId:  "com.huawei.test",
		})
		testErr := err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrSchemaNotExists, testErr.Code)

		_, err = disco.ExistSchema(getContext(), &pb.GetSchemaRequest{
			ServiceId: serviceId,
			SchemaId:  "com.huawei.test",
		})
		testErr = err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrSchemaNotExists, testErr.Code)
	})
}

func TestListSchema(t *testing.T) {
	var serviceID string
	ctx := getContext()

	respCreateService, err := datasource.GetMetadataManager().RegisterService(ctx, &pb.CreateServiceRequest{
		Service: &pb.MicroService{
			ServiceName: "TestListSchema",
			Schemas: []string{
				"schemaID_1",
			},
		},
	})
	assert.NoError(t, err)
	serviceID = respCreateService.ServiceId
	defer datasource.GetMetadataManager().UnregisterService(ctx, &pb.DeleteServiceRequest{ServiceId: serviceID, Force: true})

	err = disco.PutSchema(ctx, &pb.ModifySchemaRequest{
		ServiceId: serviceID,
		SchemaId:  "schemaID_1",
		Schema:    "schema_1",
		Summary:   "summary1",
	})
	assert.NoError(t, err)

	t.Run("list schema with invalid request, should failed", func(t *testing.T) {
		schemas, err := disco.ListSchema(ctx, &pb.GetAllSchemaRequest{})
		testErr := err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrInvalidParams, testErr.Code)
		assert.Nil(t, schemas)

		schemas, err = disco.ListSchema(ctx, &pb.GetAllSchemaRequest{
			ServiceId: "not_exist_id",
		})
		testErr = err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrServiceNotExists, testErr.Code)
		assert.Nil(t, schemas)
	})

	t.Run("list schema, should ok", func(t *testing.T) {
		schemas, err := disco.ListSchema(ctx, &pb.GetAllSchemaRequest{
			ServiceId: serviceID,
		})
		assert.NoError(t, err)
		assert.Equal(t, 1, len(schemas))

		schema := schemas[0]
		assert.Equal(t, "schemaID_1", schema.SchemaId)
		assert.Equal(t, "summary1", schema.Summary)
		assert.Empty(t, schema.Schema)

		schemas, err = disco.ListSchema(ctx, &pb.GetAllSchemaRequest{
			ServiceId:  serviceID,
			WithSchema: true,
		})
		assert.NoError(t, err)
		assert.Equal(t, 1, len(schemas))

		schema = schemas[0]
		assert.Equal(t, "schemaID_1", schema.SchemaId)
		assert.Equal(t, "summary1", schema.Summary)
		assert.Equal(t, "schema_1", schema.Schema)
	})
}

func TestCompatibleOperateSchema(t *testing.T) {
	var serviceID string
	ctx := getContext()

	respCreateService, err := datasource.GetMetadataManager().RegisterService(ctx, &pb.CreateServiceRequest{
		Service: &pb.MicroService{
			ServiceName: "TestListSchema",
			Schemas: []string{
				"schemaID_1",
			},
		},
	})
	assert.NoError(t, err)
	serviceID = respCreateService.ServiceId
	defer schema.Instance().DeleteContent(ctx, &schema.ContentRequest{
		Hash: schema.Hash("schemaID_1", "schema_1"),
	})
	defer schema.Instance().DeleteContent(ctx, &schema.ContentRequest{
		Hash: schema.Hash("schemaID_2", "schema_2"),
	})
	defer datasource.GetMetadataManager().UnregisterService(ctx, &pb.DeleteServiceRequest{ServiceId: serviceID, Force: true})

	t.Run("get/list schema with old content, should ok", func(t *testing.T) {
		_, err := datasource.GetMetadataManager().ModifySchema(ctx, &pb.ModifySchemaRequest{
			ServiceId: serviceID,
			SchemaId:  "schemaID_1",
			Schema:    "schema_1",
			Summary:   "summary1",
		})
		assert.NoError(t, err)

		schema, err := disco.GetSchema(ctx, &pb.GetSchemaRequest{
			ServiceId: serviceID,
			SchemaId:  "schemaID_1",
		})
		assert.NoError(t, err)
		assert.Equal(t, "schema_1", schema.Schema)
		assert.Equal(t, "summary1", schema.Summary)

		schemas, err := disco.ListSchema(ctx, &pb.GetAllSchemaRequest{
			ServiceId:  serviceID,
			WithSchema: true,
		})
		assert.NoError(t, err)
		assert.Equal(t, 1, len(schemas))
		schema = schemas[0]
		assert.Equal(t, "schema_1", schema.Schema)
		assert.Equal(t, "summary1", schema.Summary)
	})

	t.Run("put schema overwrite old content, should ok", func(t *testing.T) {
		_, err := datasource.GetMetadataManager().ModifySchema(ctx, &pb.ModifySchemaRequest{
			ServiceId: serviceID,
			SchemaId:  "schemaID_2",
			Schema:    "schema_2",
			Summary:   "summary2",
		})
		assert.NoError(t, err)

		err = disco.PutSchema(ctx, &pb.ModifySchemaRequest{
			ServiceId: serviceID,
			SchemaId:  "schemaID_2",
			Schema:    "schema_2",
			Summary:   "summary2",
		})
		assert.NoError(t, err)

		schema, err := disco.GetSchema(ctx, &pb.GetSchemaRequest{
			ServiceId: serviceID,
			SchemaId:  "schemaID_2",
		})
		assert.NoError(t, err)
		assert.Equal(t, "schema_2", schema.Schema)
		assert.Equal(t, "summary2", schema.Summary)

		schemas, err := disco.ListSchema(ctx, &pb.GetAllSchemaRequest{
			ServiceId:  serviceID,
			WithSchema: true,
		})
		assert.NoError(t, err)
		assert.Equal(t, 2, len(schemas))
		schema = findSchemaBySchemaID(schemas, "schemaID_2")
		assert.NotNil(t, schema)
		assert.Equal(t, "schema_2", schema.Schema)
		assert.Equal(t, "summary2", schema.Summary)
	})

	t.Run("delete schema with old content, should ok", func(t *testing.T) {
		err := disco.DeleteSchema(ctx, &pb.DeleteSchemaRequest{
			ServiceId: serviceID,
			SchemaId:  "schemaID_1",
		})
		assert.NoError(t, err)

		schema, err := disco.GetSchema(ctx, &pb.GetSchemaRequest{
			ServiceId: serviceID,
			SchemaId:  "schemaID_1",
		})
		testErr := err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrSchemaNotExists, testErr.Code)

		schemas, err := disco.ListSchema(ctx, &pb.GetAllSchemaRequest{
			ServiceId:  serviceID,
			WithSchema: true,
		})
		assert.NoError(t, err)
		assert.Equal(t, 2, len(schemas))

		schema = findSchemaBySchemaID(schemas, "schemaID_1")
		assert.NotNil(t, schema)
		assert.Empty(t, schema.Schema)
		assert.Empty(t, schema.Summary)

		schema = findSchemaBySchemaID(schemas, "schemaID_2")
		assert.NotNil(t, schema)
		assert.Equal(t, "schema_2", schema.Schema)
		assert.Equal(t, "summary2", schema.Summary)
	})

	t.Run("delete schema with new/old content, should ok", func(t *testing.T) {
		err := disco.DeleteSchema(ctx, &pb.DeleteSchemaRequest{
			ServiceId: serviceID,
			SchemaId:  "schemaID_2",
		})
		assert.NoError(t, err)

		schema, err := disco.GetSchema(ctx, &pb.GetSchemaRequest{
			ServiceId: serviceID,
			SchemaId:  "schemaID_2",
		})
		testErr := err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrSchemaNotExists, testErr.Code)

		schemas, err := disco.ListSchema(ctx, &pb.GetAllSchemaRequest{
			ServiceId:  serviceID,
			WithSchema: true,
		})
		assert.NoError(t, err)
		assert.Equal(t, 2, len(schemas))

		schema = findSchemaBySchemaID(schemas, "schemaID_1")
		assert.NotNil(t, schema)
		assert.Empty(t, schema.Schema)
		assert.Empty(t, schema.Summary)

		schema = findSchemaBySchemaID(schemas, "schemaID_2")
		assert.NotNil(t, schema)
		assert.Empty(t, schema.Schema)
		assert.Empty(t, schema.Summary)
	})
}

func findSchemaBySchemaID(schemas []*pb.Schema, schemaID string) *pb.Schema {
	for _, schema := range schemas {
		if schema.SchemaId == schemaID {
			return schema
		}
	}
	return nil
}

func TestSchemaUsage(t *testing.T) {
	ctx := util.WithNoCache(util.SetDomainProject(context.Background(), "default", "default"))

	resp, err := disco.RegisterService(ctx, &pb.CreateServiceRequest{
		Service: &pb.MicroService{
			ServiceName: "TestSchemaUsage",
		},
	})
	assert.NoError(t, err)
	serviceID := resp.ServiceId
	defer disco.UnregisterService(ctx, &pb.DeleteServiceRequest{ServiceId: serviceID, Force: true})

	t.Run("get not exist service schema usage, should failed", func(t *testing.T) {
		_, err := disco.Usage(ctx, "note_exist_service_id")
		testErr := err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrServiceNotExists, testErr.Code)
	})

	t.Run("get usage without schemas, should return 0", func(t *testing.T) {
		usage, err := disco.Usage(ctx, serviceID)
		assert.NoError(t, err)
		assert.Equal(t, int64(0), usage)
	})

	t.Run("get usage with schemas, should return 1", func(t *testing.T) {
		err := disco.PutSchema(ctx, &pb.ModifySchemaRequest{
			ServiceId: serviceID,
			SchemaId:  "schemaID_1",
			Schema:    "schema_1",
		})
		assert.NoError(t, err)

		usage, err := disco.Usage(ctx, serviceID)
		assert.NoError(t, err)
		assert.Equal(t, int64(1), usage)
	})
}

func TestSyncSchema(t *testing.T) {
	if !test.IsETCD() {
		return
	}
	initWhiteList()

	var serviceID string
	var serviceIDNotInWhiteList string

	t.Run("register a microservice", func(t *testing.T) {
		t.Run("register a microservice named sync_schemas_service will create a service task should pass", func(t *testing.T) {
			resp, err := disco.RegisterService(schemaContext(), &pb.CreateServiceRequest{
				Service: &pb.MicroService{
					AppId:       "sync_schemas_prod",
					ServiceName: "sync_schemas_service",
					Version:     "1.0.1",
					Level:       "FRONT",
					Status:      pb.MS_UP,
					Environment: pb.ENV_PROD,
				},
			})
			assert.NoError(t, err)
			assert.Equal(t, pb.ResponseSuccess, resp.Response.GetCode())
			serviceID = resp.ServiceId
			listTaskReq := model.ListTaskRequest{
				Domain:       schemaDomain,
				Project:      schemaProject,
				Action:       csync.CreateAction,
				ResourceType: datasource.ResourceService,
				Status:       csync.PendingStatus,
			}
			tasks, err := task.List(context.Background(), &listTaskReq)
			assert.NoError(t, err)
			assert.Equal(t, 1, len(tasks))
			err = task.Delete(context.Background(), tasks...)
			assert.NoError(t, err)
		})

		t.Run("register a microservice named CCC will not create a service task should pass", func(t *testing.T) {
			resp, err := disco.RegisterService(schemaContext(), &pb.CreateServiceRequest{
				Service: &pb.MicroService{
					AppId:       "sync_schemas_prod",
					ServiceName: "CCC",
					Version:     "1.0.1",
					Level:       "FRONT",
					Status:      pb.MS_UP,
					Environment: pb.ENV_PROD,
				},
			})
			assert.NoError(t, err)
			assert.Equal(t, pb.ResponseSuccess, resp.Response.GetCode())
			serviceIDNotInWhiteList = resp.ServiceId
			listTaskReq := model.ListTaskRequest{
				Domain:       schemaDomain,
				Project:      schemaProject,
				Action:       csync.CreateAction,
				ResourceType: datasource.ResourceService,
				Status:       csync.PendingStatus,
			}
			tasks, err := task.List(context.Background(), &listTaskReq)
			assert.NoError(t, err)
			assert.Equal(t, 0, len(tasks))
		})
	})

	t.Run("put schema will execute the PutContent func", func(t *testing.T) {
		t.Run("update a microservice named sync_schemas_service's schema with valid request, will create 3 kv tasks(hash summary content) should pass", func(t *testing.T) {
			err := disco.PutSchema(schemaContext(), &pb.ModifySchemaRequest{
				ServiceId: serviceID,
				SchemaId:  "12345678",
				Schema:    "schema-one",
				Summary:   "12345678",
			})
			assert.NoError(t, err)

			ref, err := disco.GetSchema(schemaContext(), &pb.GetSchemaRequest{
				ServiceId: serviceID,
				SchemaId:  "12345678",
			})
			assert.NoError(t, err)
			assert.NotNil(t, ref)
			assert.Equal(t, "12345678", ref.Summary)
			assert.Equal(t, "schema-one", ref.Schema)
			listTaskReq := model.ListTaskRequest{
				Domain:       schemaDomain,
				Project:      schemaProject,
				Action:       csync.UpdateAction,
				ResourceType: datasource.ResourceKV,
				Status:       csync.PendingStatus,
			}
			tasks, err := task.List(schemaContext(), &listTaskReq)
			assert.NoError(t, err)
			// append the schemaID into service.Schemas if schemaID is new will create a kv task
			assert.Equal(t, 4, len(tasks))
			err = task.Delete(schemaContext(), tasks...)
			assert.NoError(t, err)
		})

		t.Run("update a microservice named CCC's schema with valid request, will not create 3 kv tasks(hash summary content) should pass", func(t *testing.T) {
			err := disco.PutSchema(schemaContext(), &pb.ModifySchemaRequest{
				ServiceId: serviceIDNotInWhiteList,
				SchemaId:  "87654321",
				Schema:    "schema-two",
				Summary:   "87654321",
			})
			assert.NoError(t, err)

			ref, err := disco.GetSchema(schemaContext(), &pb.GetSchemaRequest{
				ServiceId: serviceIDNotInWhiteList,
				SchemaId:  "87654321",
			})
			assert.NoError(t, err)
			assert.NotNil(t, ref)
			assert.Equal(t, "87654321", ref.Summary)
			assert.Equal(t, "schema-two", ref.Schema)
			listTaskReq := model.ListTaskRequest{
				Domain:       schemaDomain,
				Project:      schemaProject,
				Action:       csync.UpdateAction,
				ResourceType: datasource.ResourceKV,
				Status:       csync.PendingStatus,
			}
			tasks, err := task.List(schemaContext(), &listTaskReq)
			assert.NoError(t, err)
			assert.Equal(t, 0, len(tasks))
		})
	})

	t.Run("put schemas will execute the PutManyContent func", func(t *testing.T) {
		t.Run("update a microservice named sync_schemas_service's schemas with valid request, will create 7 kv update task (2 ref tasks, 2 content tasks, 2 summary tasks"+
			" 1 service task), two delete kv task, two tombstones(ref and summary) should pass", func(t *testing.T) {
			err := disco.PutSchemas(schemaContext(), &pb.ModifySchemasRequest{
				ServiceId: serviceID,
				Schemas: []*pb.Schema{
					{
						SchemaId: "11111111",
						Summary:  "11111111",
						Schema:   "11111111",
					},
					{
						SchemaId: "22222222",
						Summary:  "22222222",
						Schema:   "22222222",
					},
				},
			})
			assert.NoError(t, err)
			ref, err := disco.GetSchema(schemaContext(), &pb.GetSchemaRequest{
				ServiceId: serviceID,
				SchemaId:  "11111111",
			})
			assert.NoError(t, err)
			assert.NotNil(t, ref)
			assert.Equal(t, "11111111", ref.Summary)
			assert.Equal(t, "11111111", ref.Schema)
			ref, err = disco.GetSchema(schemaContext(), &pb.GetSchemaRequest{
				ServiceId: serviceID,
				SchemaId:  "22222222",
			})
			assert.NoError(t, err)
			assert.NotNil(t, ref)
			assert.Equal(t, "22222222", ref.Summary)
			assert.Equal(t, "22222222", ref.Schema)
			listTaskReq := model.ListTaskRequest{
				Domain:       schemaDomain,
				Project:      schemaProject,
				Action:       csync.UpdateAction,
				ResourceType: datasource.ResourceKV,
				Status:       csync.PendingStatus,
			}
			tasks, err := task.List(schemaContext(), &listTaskReq)
			assert.NoError(t, err)
			assert.Equal(t, 7, len(tasks))
			err = task.Delete(schemaContext(), tasks...)
			assert.NoError(t, err)
			listTaskReq = model.ListTaskRequest{
				Domain:       schemaDomain,
				Project:      schemaProject,
				Action:       csync.DeleteAction,
				ResourceType: datasource.ResourceKV,
				Status:       csync.PendingStatus,
			}
			tasks, err = task.List(schemaContext(), &listTaskReq)
			assert.NoError(t, err)
			assert.Equal(t, 2, len(tasks))
			err = task.Delete(schemaContext(), tasks...)
			assert.NoError(t, err)
			tombstoneListReq := model.ListTombstoneRequest{
				Domain:       schemaDomain,
				Project:      schemaProject,
				ResourceType: datasource.ResourceKV,
			}
			tombstones, err := tombstone.List(schemaContext(), &tombstoneListReq)
			assert.NoError(t, err)
			assert.Equal(t, 2, len(tombstones))
			err = tombstone.Delete(schemaContext(), tombstones...)
			assert.NoError(t, err)
		})

		t.Run("update a microservice named CCC's schemas with valid request, will not create 7 kv update task (2 ref tasks, 2 content tasks, 2 summary tasks"+
			" 1 service task), two delete kv task, two tombstones(ref and summary) should pass", func(t *testing.T) {
			err := disco.PutSchemas(schemaContext(), &pb.ModifySchemasRequest{
				ServiceId: serviceIDNotInWhiteList,
				Schemas: []*pb.Schema{
					{
						SchemaId: "33333333",
						Summary:  "33333333",
						Schema:   "33333333",
					},
					{
						SchemaId: "44444444",
						Summary:  "44444444",
						Schema:   "44444444",
					},
				},
			})
			assert.NoError(t, err)
			ref, err := disco.GetSchema(schemaContext(), &pb.GetSchemaRequest{
				ServiceId: serviceIDNotInWhiteList,
				SchemaId:  "33333333",
			})
			assert.NoError(t, err)
			assert.NotNil(t, ref)
			assert.Equal(t, "33333333", ref.Summary)
			assert.Equal(t, "33333333", ref.Schema)
			ref, err = disco.GetSchema(schemaContext(), &pb.GetSchemaRequest{
				ServiceId: serviceIDNotInWhiteList,
				SchemaId:  "44444444",
			})
			assert.NoError(t, err)
			assert.NotNil(t, ref)
			assert.Equal(t, "44444444", ref.Summary)
			assert.Equal(t, "44444444", ref.Schema)
			listTaskReq := model.ListTaskRequest{
				Domain:       schemaDomain,
				Project:      schemaProject,
				Action:       csync.UpdateAction,
				ResourceType: datasource.ResourceKV,
				Status:       csync.PendingStatus,
			}
			tasks, err := task.List(schemaContext(), &listTaskReq)
			assert.NoError(t, err)
			assert.Equal(t, 0, len(tasks))
			listTaskReq = model.ListTaskRequest{
				Domain:       schemaDomain,
				Project:      schemaProject,
				Action:       csync.DeleteAction,
				ResourceType: datasource.ResourceKV,
				Status:       csync.PendingStatus,
			}
			tasks, err = task.List(schemaContext(), &listTaskReq)
			assert.NoError(t, err)
			assert.Equal(t, 0, len(tasks))
			tombstoneListReq := model.ListTombstoneRequest{
				Domain:       schemaDomain,
				Project:      schemaProject,
				ResourceType: datasource.ResourceKV,
			}
			tombstones, err := tombstone.List(schemaContext(), &tombstoneListReq)
			assert.NoError(t, err)
			assert.Equal(t, 0, len(tombstones))
		})

	})

	t.Run("delete schemas ", func(t *testing.T) {
		t.Run("delete a microservice named sync_schemas_service's schemas 11111111 and 22222222 will create 4 tasks(2 from DeleteRef, 2 from DeleteSchema ) "+
			"and 4 tombstones (2 from DeleteRef, 2 from DeleteSchema) should pass", func(t *testing.T) {
			err := disco.DeleteSchema(schemaContext(), &pb.DeleteSchemaRequest{
				ServiceId: serviceID,
				SchemaId:  "11111111",
			})
			assert.NoError(t, err)

			err = disco.DeleteSchema(schemaContext(), &pb.DeleteSchemaRequest{
				ServiceId: serviceID,
				SchemaId:  "22222222",
			})
			assert.NoError(t, err)
			listTaskReq := model.ListTaskRequest{
				Domain:       schemaDomain,
				Project:      schemaProject,
				Action:       csync.DeleteAction,
				ResourceType: datasource.ResourceKV,
				Status:       csync.PendingStatus,
			}
			tasks, err := task.List(schemaContext(), &listTaskReq)
			assert.NoError(t, err)
			assert.Equal(t, 4, len(tasks))
			err = task.Delete(schemaContext(), tasks...)
			assert.NoError(t, err)
			tombstoneListReq := model.ListTombstoneRequest{
				Domain:       schemaDomain,
				Project:      schemaProject,
				ResourceType: datasource.ResourceKV,
			}
			tombstones, err := tombstone.List(schemaContext(), &tombstoneListReq)
			assert.NoError(t, err)
			assert.Equal(t, 4, len(tombstones))
			err = tombstone.Delete(context.Background(), tombstones...)
			assert.NoError(t, err)
		})

		t.Run("delete a microservice named CCC's schemas 33333333 and 44444444 will not create 4 tasks(2 from DeleteRef, 2 from DeleteSchema ) "+
			"and 4 tombstones (2 from DeleteRef, 2 from DeleteSchema) should pass", func(t *testing.T) {
			err := disco.DeleteSchema(schemaContext(), &pb.DeleteSchemaRequest{
				ServiceId: serviceIDNotInWhiteList,
				SchemaId:  "33333333",
			})
			assert.NoError(t, err)

			err = disco.DeleteSchema(schemaContext(), &pb.DeleteSchemaRequest{
				ServiceId: serviceIDNotInWhiteList,
				SchemaId:  "44444444",
			})
			assert.NoError(t, err)
			listTaskReq := model.ListTaskRequest{
				Domain:       schemaDomain,
				Project:      schemaProject,
				Action:       csync.DeleteAction,
				ResourceType: datasource.ResourceKV,
				Status:       csync.PendingStatus,
			}
			tasks, err := task.List(schemaContext(), &listTaskReq)
			assert.NoError(t, err)
			assert.Equal(t, 0, len(tasks))
			tombstoneListReq := model.ListTombstoneRequest{
				Domain:       schemaDomain,
				Project:      schemaProject,
				ResourceType: datasource.ResourceKV,
			}
			tombstones, err := tombstone.List(schemaContext(), &tombstoneListReq)
			assert.NoError(t, err)
			assert.Equal(t, 0, len(tombstones))
		})
	})

	t.Run("unregister microservice", func(t *testing.T) {
		t.Run("unregister a microservice named sync_schemas_service will create a task and a tombstone should pass", func(t *testing.T) {
			err := disco.UnregisterService(schemaContext(), &pb.DeleteServiceRequest{
				ServiceId: serviceID,
				Force:     true,
			})
			assert.NoError(t, err)

			listTaskReq := model.ListTaskRequest{
				Domain:       schemaDomain,
				Project:      schemaProject,
				ResourceType: datasource.ResourceService,
				Action:       csync.DeleteAction,
				Status:       csync.PendingStatus,
			}
			tasks, err := task.List(schemaContext(), &listTaskReq)
			assert.NoError(t, err)
			assert.Equal(t, 1, len(tasks))
			err = task.Delete(context.Background(), tasks...)
			assert.NoError(t, err)
			tasks, err = task.List(schemaContext(), &listTaskReq)
			assert.NoError(t, err)
			assert.Equal(t, 0, len(tasks))
			tombstoneListReq := model.ListTombstoneRequest{
				Domain:       schemaDomain,
				Project:      schemaProject,
				ResourceType: datasource.ResourceService,
			}
			tombstones, err := tombstone.List(schemaContext(), &tombstoneListReq)
			assert.NoError(t, err)
			assert.Equal(t, 1, len(tombstones))
			err = tombstone.Delete(schemaContext(), tombstones...)
			assert.NoError(t, err)
		})

		t.Run("unregister a microservice named CCC will not create a task and a tombstone should pass", func(t *testing.T) {
			err := disco.UnregisterService(schemaContext(), &pb.DeleteServiceRequest{
				ServiceId: serviceIDNotInWhiteList,
				Force:     true,
			})
			assert.NoError(t, err)

			listTaskReq := model.ListTaskRequest{
				Domain:       schemaDomain,
				Project:      schemaProject,
				ResourceType: datasource.ResourceService,
				Action:       csync.DeleteAction,
				Status:       csync.PendingStatus,
			}
			tasks, err := task.List(schemaContext(), &listTaskReq)
			assert.NoError(t, err)
			assert.Equal(t, 0, len(tasks))
			tombstoneListReq := model.ListTombstoneRequest{
				Domain:       schemaDomain,
				Project:      schemaProject,
				ResourceType: datasource.ResourceService,
			}
			tombstones, err := tombstone.List(schemaContext(), &tombstoneListReq)
			assert.NoError(t, err)
			assert.Equal(t, 0, len(tombstones))
		})
	})
}
