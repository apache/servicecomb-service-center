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
	"strconv"
	"strings"
	"testing"

	"github.com/apache/servicecomb-service-center/datasource"
	"github.com/apache/servicecomb-service-center/server/plugin/quota"
	"github.com/apache/servicecomb-service-center/server/service/disco"
	pb "github.com/go-chassis/cari/discovery"
	"github.com/go-chassis/cari/pkg/errsvc"
	"github.com/stretchr/testify/assert"
)

const (
	invalidSchemaId = "@"
)

var (
	TOO_LONG_SUMMARY = strings.Repeat("x", 129)
)

func TestPutSchema(t *testing.T) {
	var (
		serviceIdDev string
		serviceId    string
	)

	t.Run("should be passed, create service", func(t *testing.T) {
		respCreateService, err := serviceResource.Create(getContext(), &pb.CreateServiceRequest{
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
		assert.Equal(t, pb.ResponseSuccess, respCreateService.Response.GetCode())
		serviceIdDev = respCreateService.ServiceId

		respCreateService, err = serviceResource.Create(getContext(), &pb.CreateServiceRequest{
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
		assert.Equal(t, pb.ResponseSuccess, respCreateService.Response.GetCode())
		serviceId = respCreateService.ServiceId
	})

	defer func() {
		serviceResource.Delete(getContext(), &pb.DeleteServiceRequest{ServiceId: serviceIdDev, Force: true})
		serviceResource.Delete(getContext(), &pb.DeleteServiceRequest{ServiceId: serviceId, Force: true})
	}()

	f := func() {
		_, err := disco.PutSchema(getContext(), &pb.ModifySchemaRequest{
			ServiceId: "",
			SchemaId:  "com.huawei.test",
			Schema:    "create schema",
		})
		testErr := err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrInvalidParams, testErr.Code)

		_, err = disco.PutSchema(getContext(), &pb.ModifySchemaRequest{
			ServiceId: "notExistService",
			SchemaId:  "com.huawei.test",
			Schema:    "create schema",
		})
		testErr = err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrServiceNotExists, testErr.Code)

		_, err = disco.PutSchema(getContext(), &pb.ModifySchemaRequest{
			ServiceId: serviceIdDev,
			SchemaId:  invalidSchemaId,
			Schema:    "create schema",
		})
		testErr = err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrInvalidParams, testErr.Code)

		_, err = disco.PutSchema(getContext(), &pb.ModifySchemaRequest{
			ServiceId: serviceIdDev,
			SchemaId:  "com.huawei.test",
			Schema:    "create schema",
			Summary:   TOO_LONG_SUMMARY,
		})
		testErr = err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrInvalidParams, testErr.Code)

		_, err = disco.PutSchema(getContext(), &pb.ModifySchemaRequest{
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
		_, err := disco.PutSchemas(getContext(), &pb.ModifySchemasRequest{
			ServiceId: serviceIdDev,
		})
		testErr := err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrInvalidParams, testErr.Code)

		_, err = disco.PutSchemas(getContext(), &pb.ModifySchemasRequest{
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

		_, err = disco.PutSchemas(getContext(), &pb.ModifySchemasRequest{
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

		_, err = disco.PutSchemas(getContext(), &pb.ModifySchemasRequest{
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

		_, err = disco.PutSchemas(getContext(), &pb.ModifySchemasRequest{
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

		_, err = disco.PutSchemas(getContext(), &pb.ModifySchemasRequest{
			ServiceId: serviceIdDev,
			Schemas: []*pb.Schema{
				{
					SchemaId: "com.huawei.test",
					Schema:   "create schema",
					Summary:  TOO_LONG_SUMMARY,
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

	size := quota.DefaultSchemaQuota + 1
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

	t.Run("should be failed in dev env, when create schemas out of gauge", func(t *testing.T) {
		_, err := disco.PutSchemas(getContext(), &pb.ModifySchemasRequest{
			ServiceId: serviceIdDev,
			Schemas:   schemas,
		})
		testErr := err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrInvalidParams, testErr.Code)

		_, err = disco.PutSchemas(getContext(), &pb.ModifySchemasRequest{
			ServiceId: serviceIdDev,
			Schemas:   schemas[:quota.DefaultSchemaQuota],
		})
		assert.NoError(t, err)

		schema := schemas[quota.DefaultSchemaQuota]
		_, err = disco.PutSchema(getContext(), &pb.ModifySchemaRequest{
			ServiceId: serviceIdDev,
			SchemaId:  schema.SchemaId,
			Schema:    schema.Schema,
		})
		testErr = err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrNotEnoughQuota, testErr.Code)
	})

	t.Run("should be failed in prod env, when create schemas out of gauge", func(t *testing.T) {
		_, err := disco.PutSchemas(getContext(), &pb.ModifySchemasRequest{
			ServiceId: serviceIdDev,
			Schemas:   schemas,
		})
		testErr := err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrInvalidParams, testErr.Code)
	})

	t.Run("should be failed when create service", func(t *testing.T) {
		createServiceResponse, err := serviceResource.Create(getContext(), &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				AppId:       "check_schema_group",
				ServiceName: "check_schema_service",
				Version:     "1.0.0",
				Level:       "FRONT",
				Schemas:     schemaIds,
				Status:      pb.MS_UP,
			},
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ErrInvalidParams, createServiceResponse.Response.GetCode())
	})

	var (
		serviceIdPro string
	)
	respCreateService, err := serviceResource.Create(getContext(), &pb.CreateServiceRequest{
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
	assert.Equal(t, pb.ResponseSuccess, respCreateService.Response.GetCode())
	serviceIdPro = respCreateService.ServiceId
	defer serviceResource.Delete(getContext(), &pb.DeleteServiceRequest{ServiceId: serviceIdPro, Force: true})

	t.Run("should be failed, when modify schema and summary is empty", func(t *testing.T) {
		respModifySchema, err := disco.PutSchema(getContext(), &pb.ModifySchemaRequest{
			ServiceId: serviceIdPro,
			SchemaId:  "first_schemaId",
			Schema:    "first_schema",
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respModifySchema.Response.GetCode())

		schemas := []*pb.Schema{
			{
				SchemaId: "first_schemaId",
				Schema:   "first_schema",
				Summary:  "first0summary",
			},
		}
		respModifySchemas, err := disco.PutSchemas(getContext(), &pb.ModifySchemasRequest{
			ServiceId: serviceIdPro,
			Schemas:   schemas,
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respModifySchemas.Response.GetCode())

		respExist, err := serviceResource.Exist(getContext(), &pb.GetExistenceRequest{
			Type:      datasource.ExistTypeSchema,
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
		respModifySchemas, err = disco.PutSchemas(getContext(), &pb.ModifySchemasRequest{
			ServiceId: serviceIdPro,
			Schemas:   schemas,
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respModifySchemas.Response.GetCode())
	})
}

func TestPutSchemas(t *testing.T) {
	var (
		serviceIdDev1 string
		serviceIdDev2 string
	)
	respCreateService, err := serviceResource.Create(getContext(), &pb.CreateServiceRequest{
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
	assert.Equal(t, pb.ResponseSuccess, respCreateService.Response.GetCode())
	serviceIdDev1 = respCreateService.ServiceId
	defer serviceResource.Delete(getContext(), &pb.DeleteServiceRequest{ServiceId: serviceIdDev1, Force: true})

	respCreateService, err = serviceResource.Create(getContext(), &pb.CreateServiceRequest{
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
	assert.Equal(t, pb.ResponseSuccess, respCreateService.Response.GetCode())
	serviceIdDev2 = respCreateService.ServiceId
	defer serviceResource.Delete(getContext(), &pb.DeleteServiceRequest{ServiceId: serviceIdDev2, Force: true})

	t.Run("should be passed, when create schemas when service schema id set is empty", func(t *testing.T) {
		schemas := []*pb.Schema{
			{
				SchemaId: "first_schemaId",
				Schema:   "first_schema",
				Summary:  "first0summary",
			},
			{
				SchemaId: "first_schemaId",
				Schema:   "first_schema",
				Summary:  "first0summary",
			},
		}
		respCreateService, err := disco.PutSchemas(getContext(), &pb.ModifySchemasRequest{
			ServiceId: serviceIdDev1,
			Schemas:   schemas,
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respCreateService.Response.GetCode())

		respGetAllSchema, err := disco.ListSchema(getContext(), &pb.GetAllSchemaRequest{
			ServiceId: serviceIdDev1,
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respGetAllSchema.Response.GetCode())
		assert.Equal(t, 1, len(respGetAllSchema.Schemas))

		schemas = []*pb.Schema{}
		respCreateService, err = disco.PutSchemas(getContext(), &pb.ModifySchemasRequest{
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
		respCreateService, err = disco.PutSchemas(getContext(), &pb.ModifySchemasRequest{
			ServiceId: serviceIdDev1,
			Schemas:   schemas,
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respCreateService.Response.GetCode())

		schemas = []*pb.Schema{
			{
				SchemaId: "second_schemaId",
				Schema:   "second_schema",
				Summary:  "second0summary",
			},
		}
		respCreateService, err = disco.PutSchemas(getContext(), &pb.ModifySchemasRequest{
			ServiceId: serviceIdDev1,
			Schemas:   schemas,
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respCreateService.Response.GetCode())

		service, err := disco.GetService(getContext(), &pb.GetServiceRequest{
			ServiceId: serviceIdDev1,
		})
		assert.NoError(t, err)
		assert.Equal(t, []string{"second_schemaId"}, service.Schemas)

		schemas = []*pb.Schema{}
		respCreateService, err = disco.PutSchemas(getContext(), &pb.ModifySchemasRequest{
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
		respCreateService, err = disco.PutSchemas(getContext(), &pb.ModifySchemasRequest{
			ServiceId: serviceIdDev2,
			Schemas:   schemas,
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respCreateService.Response.GetCode())

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
	respCreateService, err := serviceResource.Create(getContext(), &pb.CreateServiceRequest{
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
	assert.Equal(t, pb.ResponseSuccess, respCreateService.Response.GetCode())
	serviceId = respCreateService.ServiceId
	defer serviceResource.Delete(getContext(), &pb.DeleteServiceRequest{ServiceId: serviceId, Force: true})

	resp, err := disco.PutSchema(getContext(), &pb.ModifySchemaRequest{
		ServiceId: serviceId,
		SchemaId:  "com.huawei.test",
		Schema:    "query schema",
		Summary:   "summary",
	})
	assert.NoError(t, err)
	assert.Equal(t, pb.ResponseSuccess, resp.Response.GetCode())

	resp, err = disco.PutSchema(getContext(), &pb.ModifySchemaRequest{
		ServiceId: serviceId,
		SchemaId:  "com.huawei.test.no.summary",
		Schema:    "query schema",
	})
	assert.NoError(t, err)
	assert.Equal(t, pb.ResponseSuccess, resp.Response.GetCode())

	t.Run("should be failed, when request is invalid", func(t *testing.T) {
		resp, err := serviceResource.Exist(getContext(), &pb.GetExistenceRequest{
			Type:      "schema",
			ServiceId: "",
			SchemaId:  "com.huawei.test",
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ErrInvalidParams, resp.Response.GetCode())

		resp, err = serviceResource.Exist(getContext(), &pb.GetExistenceRequest{
			Type:      "schema",
			ServiceId: serviceId,
			SchemaId:  "noneschema",
		})
		testErr := err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrSchemaNotExists, testErr.Code)

		resp, err = serviceResource.Exist(getContext(), &pb.GetExistenceRequest{
			Type:      "schema",
			ServiceId: serviceId,
			SchemaId:  TOO_LONG_SCHEMAID,
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ErrInvalidParams, resp.Response.GetCode())
	})

	t.Run("should be passed, when request is valid", func(t *testing.T) {
		resp, err := serviceResource.Exist(getContext(), &pb.GetExistenceRequest{
			Type:      "schema",
			ServiceId: serviceId,
			SchemaId:  "com.huawei.test",
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, resp.Response.GetCode())
		assert.Equal(t, "summary", resp.Summary)

		resp, err = serviceResource.Exist(getContext(), &pb.GetExistenceRequest{
			Type:        "schema",
			ServiceId:   serviceId,
			SchemaId:    "com.huawei.test",
			AppId:       "()",
			ServiceName: "",
			Version:     "()",
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, resp.Response.GetCode())

		resp, err = serviceResource.Exist(getContext(), &pb.GetExistenceRequest{
			Type:      "schema",
			ServiceId: serviceId,
			SchemaId:  "com.huawei.test.no.summary",
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, resp.Response.GetCode())
		assert.Equal(t, "com.huawei.test.no.summary", resp.SchemaId)
		assert.Equal(t, "", resp.Summary)
	})
}

func TestGetSchema(t *testing.T) {
	var (
		serviceId     string
		serviceId1    string
		schemaId1     string = "all_schema1"
		schemaId2     string = "all_schema2"
		schemaId3     string = "all_schema3"
		summary       string = "this0is1a2test"
		schemaContent string = "the content is vary large"
	)

	respCreateService, err := serviceResource.Create(getContext(), &pb.CreateServiceRequest{
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
	assert.Equal(t, pb.ResponseSuccess, respCreateService.Response.GetCode())
	serviceId = respCreateService.ServiceId
	defer serviceResource.Delete(getContext(), &pb.DeleteServiceRequest{ServiceId: serviceId, Force: true})

	respCreateService, err = serviceResource.Create(getContext(), &pb.CreateServiceRequest{
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
	assert.Equal(t, pb.ResponseSuccess, respCreateService.Response.GetCode())
	serviceId1 = respCreateService.ServiceId
	defer serviceResource.Delete(getContext(), &pb.DeleteServiceRequest{ServiceId: serviceId1, Force: true})

	resp, err := disco.PutSchema(getContext(), &pb.ModifySchemaRequest{
		ServiceId: serviceId,
		SchemaId:  "com.huawei.test",
		Schema:    "get schema",
		Summary:   "schema0summary",
	})
	assert.NoError(t, err)
	assert.Equal(t, pb.ResponseSuccess, resp.Response.GetCode())

	respPutData, err := disco.PutSchema(getContext(), &pb.ModifySchemaRequest{
		ServiceId: serviceId1,
		SchemaId:  schemaId2,
		Schema:    schemaContent,
	})
	assert.NoError(t, err)
	assert.Equal(t, pb.ResponseSuccess, respPutData.Response.GetCode())

	respPutData, err = disco.PutSchema(getContext(), &pb.ModifySchemaRequest{
		ServiceId: serviceId1,
		SchemaId:  schemaId3,
		Schema:    schemaContent,
		Summary:   summary,
	})
	assert.NoError(t, err)
	assert.Equal(t, pb.ResponseSuccess, respPutData.Response.GetCode())

	t.Run("should be pass, when list schema", func(t *testing.T) {
		respGetAllSchema, err := disco.ListSchema(getContext(), &pb.GetAllSchemaRequest{
			ServiceId:  serviceId1,
			WithSchema: false,
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respGetAllSchema.Response.GetCode())
		schemas := respGetAllSchema.Schemas
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

		respGetAllSchema, err = disco.ListSchema(getContext(), &pb.GetAllSchemaRequest{
			ServiceId:  serviceId1,
			WithSchema: true,
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respGetAllSchema.Response.GetCode())
		schemas = respGetAllSchema.Schemas
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
			ServiceId: TOO_LONG_SERVICEID,
			SchemaId:  "com.huawei.test",
		})
		testErr = err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrInvalidParams, testErr.Code)

		_, err = disco.ListSchema(getContext(), &pb.GetAllSchemaRequest{
			ServiceId: TOO_LONG_SERVICEID,
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
			SchemaId:  TOO_LONG_SCHEMAID,
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
		resp, err := disco.GetSchema(getContext(), &pb.GetSchemaRequest{
			ServiceId: serviceId,
			SchemaId:  "com.huawei.test",
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respPutData.Response.GetCode())
		assert.Equal(t, "get schema", resp.Schema)
		assert.Equal(t, "schema0summary", resp.SchemaSummary)
	})
}

func TestDeleteSchema(t *testing.T) {
	var (
		serviceId string
	)
	respCreateService, err := serviceResource.Create(getContext(), &pb.CreateServiceRequest{
		Service: &pb.MicroService{
			AppId:       "delete_schema_group",
			ServiceName: "delete_schema_service",
			Version:     "1.0.0",
			Level:       "FRONT",
			Status:      pb.MS_UP,
		},
	})
	assert.NoError(t, err)
	assert.Equal(t, pb.ResponseSuccess, respCreateService.Response.GetCode())
	serviceId = respCreateService.ServiceId
	defer serviceResource.Delete(getContext(), &pb.DeleteServiceRequest{ServiceId: serviceId, Force: true})

	resp, err := disco.PutSchema(getContext(), &pb.ModifySchemaRequest{
		ServiceId: serviceId,
		SchemaId:  "com.huawei.test",
		Schema:    "delete schema",
		Summary:   "summary",
	})
	assert.NoError(t, err)
	assert.Equal(t, pb.ResponseSuccess, resp.Response.GetCode())

	t.Run("should be failed, when request is invalid", func(t *testing.T) {
		_, err := disco.DeleteSchema(getContext(), &pb.DeleteSchemaRequest{
			ServiceId: serviceId,
			SchemaId:  "noneschema",
		})
		testErr := err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrSchemaNotExists, testErr.Code)

		_, err = disco.DeleteSchema(getContext(), &pb.DeleteSchemaRequest{
			ServiceId: "",
			SchemaId:  "com.huawei.test",
		})
		testErr = err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrInvalidParams, testErr.Code)

		_, err = disco.DeleteSchema(getContext(), &pb.DeleteSchemaRequest{
			ServiceId: "noexistservice",
			SchemaId:  "com.huawei.test",
		})
		testErr = err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrServiceNotExists, testErr.Code)

		_, err = disco.DeleteSchema(getContext(), &pb.DeleteSchemaRequest{
			ServiceId: serviceId,
			SchemaId:  invalidSchemaId,
		})
		testErr = err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrInvalidParams, testErr.Code)
	})

	t.Run("should be passed, when request is valid", func(t *testing.T) {
		resp, err := disco.DeleteSchema(getContext(), &pb.DeleteSchemaRequest{
			ServiceId: serviceId,
			SchemaId:  "com.huawei.test",
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, resp.Response.GetCode())

		_, err = disco.GetSchema(getContext(), &pb.GetSchemaRequest{
			ServiceId: serviceId,
			SchemaId:  "com.huawei.test",
		})
		testErr := err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrSchemaNotExists, testErr.Code)

		_, err = serviceResource.Exist(getContext(), &pb.GetExistenceRequest{
			Type:      "schema",
			ServiceId: serviceId,
			SchemaId:  "com.huawei.test",
		})
		testErr = err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrSchemaNotExists, testErr.Code)
	})
}
