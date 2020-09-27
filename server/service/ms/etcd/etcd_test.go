// Licensed to the Apache Software Foundation (ASF) under one or more
// contributor license agreements.  See the NOTICE file distributed with
// this work for additional information regarding copyright ownership.
// The ASF licenses this file to You under the Apache License, Version 2.0
// (the "License"); you may not use this file except in compliance with
// the License.  You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package etcd_test

import (
	"github.com/apache/servicecomb-service-center/pkg/log"
	pb "github.com/apache/servicecomb-service-center/pkg/registry"
	"github.com/apache/servicecomb-service-center/server/core/proto"
	"github.com/apache/servicecomb-service-center/server/plugin/quota"
	scerr "github.com/apache/servicecomb-service-center/server/scerror"
	"github.com/apache/servicecomb-service-center/server/service"
	"github.com/apache/servicecomb-service-center/server/service/ms"
	"github.com/apache/servicecomb-service-center/server/service/ms/etcd"
	"github.com/go-chassis/go-archaius"
	"github.com/stretchr/testify/assert"
	"strconv"
	"strings"
	"testing"
	"time"
)

var (
	TooLongServiceId   = strings.Repeat("x", 66)
	TooLongAppId       = strings.Repeat("x", 162)
	TooLongSchemaId    = strings.Repeat("x", 162)
	TooLongServiceName = strings.Repeat("x", 130)
	TooLongExistence   = strings.Repeat("x", 128+160+2)
	TooLongAlias       = strings.Repeat("x", 130)
	TooLongFramework   = strings.Repeat("x", 66)
)

func TestInit(t *testing.T) {
	_ = archaius.Init(archaius.WithMemorySource())
	_ = archaius.Set("servicecomb.ms.name", "etcd")
}

func TestService_Register(t *testing.T) {
	t.Run("Register service after init & install, should pass", func(t *testing.T) {
		ms.Install("etcd", func(opts ms.Options) (ms.DataSource, error) {
			return etcd.NewDataSource(), nil
		})

		err := ms.Init(ms.Options{
			Endpoint:       "",
			PluginImplName: ms.ImplName(archaius.GetString("servicecomb.ms.name", "etcd")),
		})
		assert.NoError(t, err)

		size := quota.DefaultSchemaQuota + 1
		paths := make([]*pb.ServicePath, 0, size)
		properties := make(map[string]string, size)
		for i := 0; i < size; i++ {
			s := strconv.Itoa(i) + strings.Repeat("x", 253)
			paths = append(paths, &pb.ServicePath{Path: s, Property: map[string]string{s: s}})
			properties[s] = s
		}
		request := &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				AppId:       "service-ms-appID",
				ServiceName: "service-ms-serviceName",
				Version:     "32767.32767.32767.32767",
				Alias:       "service-ms-alias",
				Level:       "BACK",
				Status:      "UP",
				Schemas:     []string{"service-ms-schema"},
				Paths:       paths,
				Properties:  properties,
				Framework: &pb.FrameWorkProperty{
					Name:    "service-ms-frameworkName",
					Version: "service-ms-frameworkVersion",
				},
				RegisterBy: "SDK",
				Timestamp:  strconv.FormatInt(time.Now().Unix(), 10),
			},
		}
		request.Service.ModTimestamp = request.Service.Timestamp
		resp, err := ms.MicroService().RegisterService(getContext(), request)
		assert.NotNil(t, resp)
		assert.NoError(t, err)
		assert.Equal(t, proto.Response_SUCCESS, resp.Response.GetCode())
	})

	t.Run("register service with same key", func(t *testing.T) {
		ms.Install("etcd", func(opts ms.Options) (ms.DataSource, error) {
			return etcd.NewDataSource(), nil
		})
		err := ms.Init(ms.Options{
			Endpoint:       "",
			PluginImplName: ms.ImplName(archaius.GetString("servicecomb.ms.name", "etcd")),
		})
		assert.NoError(t, err)

		// serviceName: some-relay-ms-service-name
		// alias: sr-ms-service-name
		resp, err := ms.MicroService().RegisterService(getContext(), &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				ServiceName: "some-relay-ms-service-name",
				Alias:       "sr-ms-service-name",
				AppId:       "default",
				Version:     "1.0.0",
				Level:       "FRONT",
				Schemas: []string{
					"xxxxxxxx",
				},
				Status: "UP",
			},
		})
		assert.NotNil(t, resp)
		assert.NoError(t, err)
		assert.Equal(t, proto.Response_SUCCESS, resp.Response.GetCode())
		sameId := resp.ServiceId

		// serviceName: some-relay-ms-service-name
		// alias: sr1-ms-service-name
		resp, err = ms.MicroService().RegisterService(getContext(), &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				ServiceName: "some-relay-ms-service-name",
				Alias:       "sr1-ms-service-name",
				AppId:       "default",
				Version:     "1.0.0",
				Level:       "FRONT",
				Schemas: []string{
					"xxxxxxxx",
				},
				Status: "UP",
			},
		})
		assert.NotNil(t, resp)
		assert.NoError(t, err)
		assert.Equal(t, scerr.ErrServiceAlreadyExists, resp.Response.GetCode())

		// serviceName: some-relay1-ms-service-name
		// alias: sr-ms-service-name
		resp, err = ms.MicroService().RegisterService(getContext(), &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				ServiceName: "some-relay1-ms-service-name",
				Alias:       "sr-ms-service-name",
				AppId:       "default",
				Version:     "1.0.0",
				Level:       "FRONT",
				Schemas: []string{
					"xxxxxxxx",
				},
				Status: "UP",
			},
		})
		assert.NotNil(t, resp)
		assert.NoError(t, err)
		assert.Equal(t, scerr.ErrServiceAlreadyExists, resp.Response.GetCode())

		// serviceName: some-relay1-ms-service-name
		// alias: sr-ms-service-name
		// add serviceId field: sameId
		resp, err = ms.MicroService().RegisterService(getContext(), &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				ServiceId:   sameId,
				ServiceName: "some-relay1-ms-service-name",
				Alias:       "sr-ms-service-name",
				AppId:       "default",
				Version:     "1.0.0",
				Level:       "FRONT",
				Schemas: []string{
					"xxxxxxxx",
				},
				Status: "UP",
			},
		})
		assert.NotNil(t, resp)
		assert.NoError(t, err)
		assert.Equal(t, proto.Response_SUCCESS, resp.Response.GetCode())

		// serviceName: some-relay-ms-service-name
		// alias: sr1-ms-service-name
		// serviceId: sameId
		resp, err = ms.MicroService().RegisterService(getContext(), &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				ServiceId:   sameId,
				ServiceName: "some-relay-ms-service-name",
				Alias:       "sr1-ms-service-name",
				AppId:       "default",
				Version:     "1.0.0",
				Level:       "FRONT",
				Schemas: []string{
					"xxxxxxxx",
				},
				Status: "UP",
			},
		})
		assert.NotNil(t, resp)
		assert.NoError(t, err)
		assert.Equal(t, proto.Response_SUCCESS, resp.Response.GetCode())

		// serviceName: some-relay-ms-service-name
		// alias: sr1-ms-service-name
		// serviceId: custom-id-ms-service-id -- different
		resp, err = ms.MicroService().RegisterService(getContext(), &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				ServiceId:   "custom-id-ms-service-id",
				ServiceName: "some-relay-ms-service-name",
				Alias:       "sr1-ms-service-name",
				AppId:       "default",
				Version:     "1.0.0",
				Level:       "FRONT",
				Schemas: []string{
					"xxxxxxxx",
				},
				Status: "UP",
			},
		})
		assert.NotNil(t, resp)
		assert.NoError(t, err)
		assert.Equal(t, scerr.ErrServiceAlreadyExists, resp.Response.GetCode())

		// serviceName: some-relay1-ms-service-name
		// alias: sr-ms-service-name
		// serviceId: custom-id-ms-service-id -- different
		resp, err = ms.MicroService().RegisterService(getContext(), &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				ServiceId:   "custom-id-ms-service-id",
				ServiceName: "some-relay1-ms-service-name",
				Alias:       "sr-ms-service-name",
				AppId:       "default",
				Version:     "1.0.0",
				Level:       "FRONT",
				Schemas: []string{
					"xxxxxxxx",
				},
				Status: "UP",
			},
		})
		assert.NotNil(t, resp)
		assert.NoError(t, err)
		assert.Equal(t, scerr.ErrServiceAlreadyExists, resp.Response.GetCode())
	})

	t.Run("same serviceId,different service, can not register again,error is same as the service register twice",
		func(t *testing.T) {
			ms.Install("etcd", func(opts ms.Options) (ms.DataSource, error) {
				return etcd.NewDataSource(), nil
			})
			err := ms.Init(ms.Options{
				Endpoint:       "",
				PluginImplName: ms.ImplName(archaius.GetString("servicecomb.ms.name", "etcd")),
			})
			assert.NoError(t, err)

			resp, err := ms.MicroService().RegisterService(getContext(), &pb.CreateServiceRequest{
				Service: &pb.MicroService{
					ServiceId:   "same-serviceId-service-ms",
					ServiceName: "serviceA-service-ms",
					AppId:       "default-service-ms",
					Version:     "1.0.0",
					Level:       "FRONT",
					Schemas: []string{
						"xxxxxxxx",
					},
					Status: "UP",
				},
			})

			assert.NotNil(t, resp)
			assert.NoError(t, err)
			assert.Equal(t, resp.Response.GetCode(), proto.Response_SUCCESS)

			// same serviceId with different service name
			resp, err = ms.MicroService().RegisterService(getContext(), &pb.CreateServiceRequest{
				Service: &pb.MicroService{
					ServiceId:   "same-serviceId-service-ms",
					ServiceName: "serviceB-service-ms",
					AppId:       "default-service-ms",
					Version:     "1.0.0",
					Level:       "FRONT",
					Schemas: []string{
						"xxxxxxxx",
					},
					Status: "UP",
				},
			})
			assert.NotNil(t, resp)
			assert.NoError(t, err)
			assert.Equal(t, scerr.ErrServiceAlreadyExists, resp.Response.GetCode())
		})
}

func TestService_Get(t *testing.T) {
	// get service test
	t.Run("query all services, should pass", func(t *testing.T) {
		ms.Install("etcd", func(opts ms.Options) (ms.DataSource, error) {
			return etcd.NewDataSource(), nil
		})
		err := ms.Init(ms.Options{
			Endpoint:       "",
			PluginImplName: ms.ImplName(archaius.GetString("servicecomb.ms.name", "etcd")),
		})
		assert.NoError(t, err)
		resp, err := ms.MicroService().GetServices(getContext(), &pb.GetServicesRequest{})
		assert.NoError(t, err)
		assert.Greater(t, len(resp.Services), 0)
	})

	t.Run("get a exist service, should pass", func(t *testing.T) {
		ms.Install("etcd", func(opts ms.Options) (ms.DataSource, error) {
			return etcd.NewDataSource(), nil
		})
		err := ms.Init(ms.Options{
			Endpoint:       "",
			PluginImplName: ms.ImplName(archaius.GetString("servicecomb.ms.name", "etcd")),
		})
		assert.NoError(t, err)

		request := &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				ServiceId:   "ms-service-query-id",
				ServiceName: "ms-service-query",
				AppId:       "default",
				Version:     "1.0.4",
				Level:       "BACK",
				Properties:  make(map[string]string),
			},
		}

		resp, err := ms.MicroService().RegisterService(getContext(), request)
		assert.NoError(t, err)
		assert.Equal(t, resp.Response.GetCode(), proto.Response_SUCCESS)

		// search service by serviceID
		queryResp, err := ms.MicroService().GetService(getContext(), &pb.GetServiceRequest{
			ServiceId: "ms-service-query-id",
		})
		assert.NoError(t, err)
		assert.Equal(t, queryResp.Response.GetCode(), proto.Response_SUCCESS)
	})

	t.Run("query a service by a not existed serviceId, should not pass", func(t *testing.T) {
		ms.Install("etcd", func(opts ms.Options) (ms.DataSource, error) {
			return etcd.NewDataSource(), nil
		})
		err := ms.Init(ms.Options{
			Endpoint:       "",
			PluginImplName: ms.ImplName(archaius.GetString("servicecomb.ms.name", "etcd")),
		})
		assert.NoError(t, err)

		// not exist service
		resp, err := ms.MicroService().GetService(getContext(), &pb.GetServiceRequest{
			ServiceId: "no-exist-service",
		})
		assert.NoError(t, err)
		assert.Equal(t, resp.Response.GetCode(), scerr.ErrServiceNotExists)
	})
}

func TestService_Exist(t *testing.T) {
	var (
		serviceId1 string
		serviceId2 string
	)

	ms.Install("etcd", func(opts ms.Options) (ms.DataSource, error) {
		return etcd.NewDataSource(), nil
	})

	err := ms.Init(ms.Options{
		Endpoint:       "",
		PluginImplName: ms.ImplName(archaius.GetString("servicecomb.ms.name", "etcd")),
	})
	assert.NoError(t, err)

	t.Run("create service", func(t *testing.T) {
		svc := &pb.MicroService{
			Alias:       "es_service_ms",
			ServiceName: "exist_service_service_ms",
			AppId:       "exist_appId_service_ms",
			Version:     "1.0.0",
			Level:       "FRONT",
			Schemas: []string{
				"first_schemaId_service_ms",
			},
			Status: "UP",
		}
		resp, err := ms.MicroService().RegisterService(getContext(), &pb.CreateServiceRequest{
			Service: svc,
		})
		assert.NoError(t, err)
		assert.NotEqual(t, "", resp.ServiceId)
		serviceId1 = resp.ServiceId

		svc.ServiceId = ""
		svc.Environment = pb.ENV_PROD
		resp, err = ms.MicroService().RegisterService(getContext(), &pb.CreateServiceRequest{
			Service: svc,
		})
		assert.NoError(t, err)
		assert.NotEqual(t, "", resp.ServiceId)
		serviceId2 = resp.ServiceId
	})

	t.Run("check exist when service does not exist", func(t *testing.T) {
		log.Info("check by querying a not exist serviceName")
		resp, err := ms.MicroService().ExistService(getContext(), &pb.GetExistenceRequest{
			Type:        service.ExistTypeMicroservice,
			AppId:       "exist_appId",
			ServiceName: "notExistService_service_ms",
			Version:     "1.0.0",
		})
		assert.NoError(t, err)
		assert.Equal(t, scerr.ErrServiceNotExists, resp.Response.GetCode())

		log.Info("check by querying a not exist env")
		resp, err = ms.MicroService().ExistService(getContext(), &pb.GetExistenceRequest{
			Type:        service.ExistTypeMicroservice,
			Environment: pb.ENV_TEST,
			AppId:       "exist_appId_service_ms",
			ServiceName: "exist_service_service_ms",
			Version:     "1.0.0",
		})
		assert.NoError(t, err)
		assert.Equal(t, scerr.ErrServiceNotExists, resp.Response.GetCode())

		log.Info("check by querying a not exist env with alias")
		resp, err = ms.MicroService().ExistService(getContext(), &pb.GetExistenceRequest{
			Type:        service.ExistTypeMicroservice,
			Environment: pb.ENV_TEST,
			AppId:       "exist_appId_service_ms",
			ServiceName: "es_service_ms",
			Version:     "1.0.0",
		})
		assert.NoError(t, err)
		assert.Equal(t, scerr.ErrServiceNotExists, resp.Response.GetCode())

		log.Info("check by querying with a mismatching version")
		resp, err = ms.MicroService().ExistService(getContext(), &pb.GetExistenceRequest{
			Type:        service.ExistTypeMicroservice,
			AppId:       "exist_appId_service_ms",
			ServiceName: "exist_service_service_ms",
			Version:     "2.0.0",
		})
		assert.NoError(t, err)
		assert.Equal(t, scerr.ErrServiceNotExists, resp.Response.GetCode())
		resp, err = ms.MicroService().ExistService(getContext(), &pb.GetExistenceRequest{
			Type:        service.ExistTypeMicroservice,
			AppId:       "exist_appId_service_ms",
			ServiceName: "exist_service_service_ms",
			Version:     "0.0.0-1.0.0",
		})
		assert.NoError(t, err)
		assert.Equal(t, scerr.ErrServiceVersionNotExists, resp.Response.GetCode())
	})

	t.Run("check exist when service exists", func(t *testing.T) {
		log.Info("search with serviceName")
		resp, err := ms.MicroService().ExistService(getContext(), &pb.GetExistenceRequest{
			Type:        service.ExistTypeMicroservice,
			AppId:       "exist_appId_service_ms",
			ServiceName: "exist_service_service_ms",
			Version:     "1.0.0",
		})
		assert.NoError(t, err)
		assert.Equal(t, serviceId1, resp.ServiceId)

		log.Info("check with serviceName and env")
		resp, err = ms.MicroService().ExistService(getContext(), &pb.GetExistenceRequest{
			Type:        service.ExistTypeMicroservice,
			Environment: pb.ENV_PROD,
			AppId:       "exist_appId_service_ms",
			ServiceName: "exist_service_service_ms",
			Version:     "1.0.0",
		})
		assert.NoError(t, err)
		assert.Equal(t, serviceId2, resp.ServiceId)

		log.Info("check with alias")
		resp, err = ms.MicroService().ExistService(getContext(), &pb.GetExistenceRequest{
			Type:        service.ExistTypeMicroservice,
			AppId:       "exist_appId_service_ms",
			ServiceName: "es_service_ms",
			Version:     "1.0.0",
		})
		assert.NoError(t, err)
		assert.Equal(t, serviceId1, resp.ServiceId)

		log.Info("check with alias and env")
		resp, err = ms.MicroService().ExistService(getContext(), &pb.GetExistenceRequest{
			Type:        service.ExistTypeMicroservice,
			Environment: pb.ENV_PROD,
			AppId:       "exist_appId_service_ms",
			ServiceName: "es_service_ms",
			Version:     "1.0.0",
		})
		assert.NoError(t, err)
		assert.Equal(t, serviceId2, resp.ServiceId)

		log.Info("check with latest versionRule")
		resp, err = ms.MicroService().ExistService(getContext(), &pb.GetExistenceRequest{
			Type:        service.ExistTypeMicroservice,
			AppId:       "exist_appId_service_ms",
			ServiceName: "es_service_ms",
			Version:     "latest",
		})
		assert.NoError(t, err)
		assert.Equal(t, serviceId1, resp.ServiceId)

		log.Info("check with 1.0.0+ versionRule")
		resp, err = ms.MicroService().ExistService(getContext(), &pb.GetExistenceRequest{
			Type:        service.ExistTypeMicroservice,
			AppId:       "exist_appId_service_ms",
			ServiceName: "es_service_ms",
			Version:     "1.0.0+",
		})
		assert.NoError(t, err)
		assert.Equal(t, serviceId1, resp.ServiceId)

		log.Info("check with range versionRule")
		resp, err = ms.MicroService().ExistService(getContext(), &pb.GetExistenceRequest{
			Type:        service.ExistTypeMicroservice,
			AppId:       "exist_appId_service_ms",
			ServiceName: "es_service_ms",
			Version:     "0.9.1-1.0.1",
		})
		assert.NoError(t, err)
		assert.Equal(t, serviceId1, resp.ServiceId)
	})
}

func TestSchema_Create(t *testing.T) {
	ms.Install("etcd", func(opts ms.Options) (ms.DataSource, error) {
		return etcd.NewDataSource(), nil
	})
	err := ms.Init(ms.Options{
		Endpoint:       "",
		PluginImplName: ms.ImplName(archaius.GetString("servicecomb.ms.name", "etcd")),
	})
	assert.NoError(t, err)

	var (
		serviceIdDev string
	)

	t.Run("create service, should pass", func(t *testing.T) {
		svc := &pb.MicroService{
			Alias:       "create_schema_group_service_ms",
			ServiceName: "create_schema_service_service_ms",
			Version:     "1.0.0",
			Level:       "FRONT",
			Status:      pb.MS_UP,
			Environment: pb.ENV_DEV,
		}
		resp, err := ms.MicroService().RegisterService(getContext(), &pb.CreateServiceRequest{
			Service: svc,
		})
		assert.NoError(t, err)
		assert.NotEqual(t, "", resp.ServiceId)
		assert.Equal(t, proto.Response_SUCCESS, resp.Response.GetCode())
		serviceIdDev = resp.ServiceId

		resp, err = ms.MicroService().RegisterService(getContext(), &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				AppId:       "create_schema_group_service_ms",
				ServiceName: "create_schema_service_service_ms",
				Version:     "1.0.0",
				Level:       "FRONT",
				Status:      pb.MS_UP,
				Environment: pb.ENV_PROD,
			},
		})
		assert.NoError(t, err)
		assert.NotEqual(t, "", resp.ServiceId)
		assert.Equal(t, proto.Response_SUCCESS, resp.Response.GetCode())
	})

	t.Run("create schemas out of gauge", func(t *testing.T) {
		log.Info("create schemas out of gauge")
		size := quota.DefaultSchemaQuota + 1
		schemaIds := make([]string, 0, size)
		schemas := make([]*pb.Schema, 0, size)
		for i := 0; i < size; i++ {
			s := "ServiceCombTestTheLimitOfSchemasServiceMS" + strconv.Itoa(i)

			schemaIds = append(schemaIds, s)
			schemas = append(schemas, &pb.Schema{
				SchemaId: s,
				Schema:   s,
				Summary:  s,
			})
		}

		log.Info("batch modify schemas 1, should failed")
		resp, err := ms.MicroService().ModifySchemas(getContext(), &pb.ModifySchemasRequest{
			ServiceId: serviceIdDev,
			Schemas:   schemas,
		})
		assert.NoError(t, err)
		assert.Equal(t, scerr.ErrNotEnoughQuota, resp.Response.GetCode())

		log.Info("batch modify schemas 2")
		resp, err = ms.MicroService().ModifySchemas(getContext(), &pb.ModifySchemasRequest{
			ServiceId: serviceIdDev,
			Schemas:   schemas[:quota.DefaultSchemaQuota],
		})
		assert.NoError(t, err)
		assert.Equal(t, proto.Response_SUCCESS, resp.Response.GetCode())

		log.Info("should be failed in production env")
		resp, err = ms.MicroService().ModifySchemas(getContext(), &pb.ModifySchemasRequest{
			ServiceId: serviceIdDev,
			Schemas:   schemas,
		})
		assert.NoError(t, err)
		assert.Equal(t, scerr.ErrNotEnoughQuota, resp.Response.GetCode())
	})

	t.Run("batch create schemas in dev env", func(t *testing.T) {
		var (
			serviceIdDev1 string
			serviceIdDev2 string
		)

		log.Info("register service, should pass")
		resp, err := ms.MicroService().RegisterService(getContext(), &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				AppId:       "create_schemas_dev_service_ms",
				ServiceName: "create_schemas_service_service_ms",
				Version:     "1.0.0",
				Level:       "FRONT",
				Status:      pb.MS_UP,
				Environment: pb.ENV_DEV,
			},
		})
		assert.NoError(t, err)
		assert.Equal(t, proto.Response_SUCCESS, resp.Response.GetCode())
		serviceIdDev1 = resp.ServiceId

		resp, err = ms.MicroService().RegisterService(getContext(), &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				AppId:       "create_schemas_dev_service_ms",
				ServiceName: "create_schemas_service_service_ms",
				Version:     "1.0.1",
				Level:       "FRONT",
				Schemas: []string{
					"first_schemaId_service_ms",
				},
				Status:      pb.MS_UP,
				Environment: pb.ENV_DEV,
			},
		})
		assert.NoError(t, err)
		assert.Equal(t, proto.Response_SUCCESS, resp.Response.GetCode())
		serviceIdDev2 = resp.ServiceId

		log.Info("create schemas with service schemaId set is empty")
		schemas := []*pb.Schema{
			{
				SchemaId: "first_schemaId_service_ms",
				Schema:   "first_schema_service_ms",
				Summary:  "first0summary_service_ms",
			},
			{
				SchemaId: "first_schemaId_service_ms",
				Schema:   "first_schema_service_ms",
				Summary:  "first0summary_service_ms",
			},
		}
		respCreateSchema, err := ms.MicroService().ModifySchemas(getContext(), &pb.ModifySchemasRequest{
			ServiceId: serviceIdDev1,
			Schemas:   schemas,
		})
		assert.NoError(t, err)
		assert.Equal(t, proto.Response_SUCCESS, respCreateSchema.Response.GetCode())

		// todo: test GetAllSchemaInfo interface refers to schema_test line 342

		log.Info("modify schemas")
		schemas = []*pb.Schema{
			{
				SchemaId: "first_schemaId_service_ms",
				Schema:   "first_schema_change_service_ms",
				Summary:  "first0summary1change_service_ms",
			},
		}
		respCreateSchema, err = ms.MicroService().ModifySchemas(getContext(), &pb.ModifySchemasRequest{
			ServiceId: serviceIdDev1,
			Schemas:   schemas,
		})

		log.Info("add schemas")
		schemas = []*pb.Schema{
			{
				SchemaId: "second_schemaId_service_ms",
				Schema:   "second_schema_service_ms",
				Summary:  "second0summary_service_ms",
			},
		}
		respCreateSchema, err = ms.MicroService().ModifySchemas(getContext(), &pb.ModifySchemasRequest{
			ServiceId: serviceIdDev1,
			Schemas:   schemas,
		})
		assert.NoError(t, err)
		assert.Equal(t, proto.Response_SUCCESS, respCreateSchema.Response.GetCode())

		log.Info("query service by serviceID to obtain schema info")
		respGetService, err := ms.MicroService().GetService(getContext(), &pb.GetServiceRequest{
			ServiceId: serviceIdDev1,
		})
		assert.NoError(t, err)
		assert.Equal(t, proto.Response_SUCCESS, respGetService.Response.GetCode())
		assert.Equal(t, []string{"second_schemaId_service_ms"}, respGetService.Service.Schemas)

		log.Info("add new schemaId not exist in service's schemaId list")
		schemas = []*pb.Schema{
			{
				SchemaId: "second_schemaId_service_ms",
				Schema:   "second_schema_service_ms",
				Summary:  "second0summary_service_ms",
			},
		}
		respCreateSchema, err = ms.MicroService().ModifySchemas(getContext(), &pb.ModifySchemasRequest{
			ServiceId: serviceIdDev2,
			Schemas:   schemas,
		})
		assert.NoError(t, err)
		assert.Equal(t, proto.Response_SUCCESS, respCreateSchema.Response.GetCode())

		respGetService, err = ms.MicroService().GetService(getContext(), &pb.GetServiceRequest{
			ServiceId: serviceIdDev2,
		})
		assert.NoError(t, err)
		assert.Equal(t, proto.Response_SUCCESS, respGetService.Response.GetCode())
		assert.Equal(t, []string{"second_schemaId_service_ms"}, respGetService.Service.Schemas)
	})

	t.Run("batch create schemas in production env", func(t *testing.T) {
		var (
			serviceIdPro1 string
			serviceIdPro2 string
		)

		log.Info("register service")
		respCreateService, err := ms.MicroService().RegisterService(getContext(), &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				AppId:       "create_schemas_prod_service_ms",
				ServiceName: "create_schemas_service_service_ms",
				Version:     "1.0.0",
				Level:       "FRONT",
				Status:      pb.MS_UP,
				Environment: pb.ENV_PROD,
			},
		})
		assert.NoError(t, err)
		assert.Equal(t, proto.Response_SUCCESS, respCreateService.Response.GetCode())
		serviceIdPro1 = respCreateService.ServiceId

		respCreateService, err = ms.MicroService().RegisterService(getContext(), &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				AppId:       "create_schemas_prod_service_ms",
				ServiceName: "create_schemas_service_service_ms",
				Version:     "1.0.1",
				Level:       "FRONT",
				Schemas: []string{
					"first_schemaId_service_ms",
					"second_schemaId_service_ms",
				},
				Status:      pb.MS_UP,
				Environment: pb.ENV_PROD,
			},
		})
		assert.NoError(t, err)
		assert.Equal(t, proto.Response_SUCCESS, respCreateService.Response.GetCode())
		serviceIdPro2 = respCreateService.ServiceId

		log.Info("add schemas to service whose schemaId set is empty")
		schemas := []*pb.Schema{
			{
				SchemaId: "first_schemaId_service_ms",
				Schema:   "first_schema_service_ms",
				Summary:  "first0summary_service_ms",
			},
			{
				SchemaId: "first_schemaId_service_ms",
				Schema:   "first_schema_service_ms",
				Summary:  "first0summary_service_ms",
			},
		}
		respModifySchemas, err := ms.MicroService().ModifySchemas(getContext(), &pb.ModifySchemasRequest{
			ServiceId: serviceIdPro1,
			Schemas:   schemas,
		})
		assert.NoError(t, err)
		assert.Equal(t, proto.Response_SUCCESS, respModifySchemas.Response.GetCode())
		respGetService, err := ms.MicroService().GetService(getContext(), &pb.GetServiceRequest{
			ServiceId: serviceIdPro1,
		})
		assert.NoError(t, err)
		assert.Equal(t, proto.Response_SUCCESS, respGetService.Response.GetCode())
		assert.Equal(t, []string{"first_schemaId_service_ms"}, respGetService.Service.Schemas)

		// todo: finish ut after implementing GetAllSchemaInfo, refer to schema_test.go line. 496

		log.Info("modify schemas content already exists, will skip more exist schema")
		respModifySchemas, err = ms.MicroService().ModifySchemas(getContext(), &pb.ModifySchemasRequest{
			ServiceId: serviceIdPro1,
			Schemas:   schemas,
		})
		assert.NoError(t, err)
		assert.Equal(t, proto.Response_SUCCESS, respModifySchemas.Response.GetCode())

		log.Info("add schemas, non-exist schemaId")
		schemas = []*pb.Schema{
			{
				SchemaId: "second_schemaId_service_ms",
				Schema:   "second_schema_service_ms",
				Summary:  "second0summary_service_ms",
			},
		}
		respModifySchemas, err = ms.MicroService().ModifySchemas(getContext(), &pb.ModifySchemasRequest{
			ServiceId: serviceIdPro1,
			Schemas:   schemas,
		})
		assert.NoError(t, err)
		assert.Equal(t, scerr.ErrUndefinedSchemaID, respModifySchemas.Response.GetCode())

		log.Info("add schema when summary is empty")
		respModifySchema, err := ms.MicroService().ModifySchema(getContext(), &pb.ModifySchemaRequest{
			ServiceId: serviceIdPro2,
			SchemaId:  "first_schemaId_service_ms",
			Schema:    "first_schema_service_ms",
		})
		assert.NoError(t, err)
		assert.Equal(t, proto.Response_SUCCESS, respModifySchema.Response.GetCode())

		log.Info("add schemas when summary in database is empty")
		schemas = []*pb.Schema{
			{
				SchemaId: "first_schemaId_service_ms",
				Schema:   "first_schema_service_ms",
				Summary:  "first0summary_service_ms",
			},
		}
		respModifySchemas, err = ms.MicroService().ModifySchemas(getContext(), &pb.ModifySchemasRequest{
			ServiceId: serviceIdPro2,
			Schemas:   schemas,
		})
		assert.NoError(t, err)
		assert.Equal(t, proto.Response_SUCCESS, respModifySchemas.Response.GetCode())
		respExist, err := ms.MicroService().ExistSchema(getContext(), &pb.GetExistenceRequest{
			Type:      service.ExistTypeSchema,
			ServiceId: serviceIdPro2,
			SchemaId:  "first_schemaId_service_ms",
		})
		assert.NoError(t, err)
		assert.Equal(t, "first0summary_service_ms", respExist.Summary)

		respModifySchemas, err = ms.MicroService().ModifySchemas(getContext(), &pb.ModifySchemasRequest{
			ServiceId: serviceIdPro2,
			Schemas:   schemas,
		})
		assert.NoError(t, err)
		assert.Equal(t, proto.Response_SUCCESS, respModifySchemas.Response.GetCode())
	})

	t.Run("create a schema in dev env", func(t *testing.T) {
		var (
			serviceIdDev1 string
			serviceIdDev2 string
		)

		log.Info("register service")
		respCreateService, err := ms.MicroService().RegisterService(getContext(), &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				AppId:       "create_schema_dev_service_ms",
				ServiceName: "create_schema_service_service_ms",
				Version:     "1.0.0",
				Level:       "FRONT",
				Status:      pb.MS_UP,
				Environment: pb.ENV_DEV,
			},
		})
		assert.NoError(t, err)
		assert.Equal(t, proto.Response_SUCCESS, respCreateService.Response.GetCode())
		serviceIdDev1 = respCreateService.ServiceId

		respCreateService, err = ms.MicroService().RegisterService(getContext(), &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				AppId:       "create_schema_dev_service_ms",
				ServiceName: "create_schema_service_service_ms",
				Version:     "1.0.1",
				Level:       "FRONT",
				Schemas: []string{
					"first_schemaId_service_ms",
				},
				Status:      pb.MS_UP,
				Environment: pb.ENV_DEV,
			},
		})
		assert.NoError(t, err)
		assert.Equal(t, proto.Response_SUCCESS, respCreateService.Response.GetCode())
		serviceIdDev2 = respCreateService.ServiceId

		log.Info("create a schema for service whose schemaID is empty")
		respModifySchema, err := ms.MicroService().ModifySchema(getContext(), &pb.ModifySchemaRequest{
			ServiceId: serviceIdDev1,
			SchemaId:  "first_schemaId_service_ms",
			Schema:    "first_schema_service_ms",
		})
		assert.NoError(t, err)
		assert.Equal(t, proto.Response_SUCCESS, respModifySchema.Response.GetCode())

		log.Info("create schema for the service whose schemaId already exist")
		respModifySchema, err = ms.MicroService().ModifySchema(getContext(), &pb.ModifySchemaRequest{
			ServiceId: serviceIdDev2,
			SchemaId:  "first_schemaId_service_ms",
			Schema:    "first_schema_service_ms",
		})
		assert.NoError(t, err)
		assert.Equal(t, proto.Response_SUCCESS, respModifySchema.Response.GetCode())

		log.Info("create schema for the service whose schema summary is empty")
		respModifySchema, err = ms.MicroService().ModifySchema(getContext(), &pb.ModifySchemaRequest{
			ServiceId: serviceIdDev1,
			SchemaId:  "first_schemaId_service_ms",
			Schema:    "first_schema_change_service_ms",
			Summary:   "first0summary1change_service_ms",
		})
		assert.NoError(t, err)
		assert.Equal(t, proto.Response_SUCCESS, respModifySchema.Response.GetCode())

		log.Info("create schema for the service whose schema summary already exist")
		respModifySchema, err = ms.MicroService().ModifySchema(getContext(), &pb.ModifySchemaRequest{
			ServiceId: serviceIdDev1,
			SchemaId:  "first_schemaId_service_ms",
			Schema:    "first_schema_service_ms",
			Summary:   "first0summary_service_ms",
		})
		assert.NoError(t, err)
		assert.Equal(t, proto.Response_SUCCESS, respModifySchema.Response.GetCode())

		log.Info("add schema")
		respModifySchema, err = ms.MicroService().ModifySchema(getContext(), &pb.ModifySchemaRequest{
			ServiceId: serviceIdDev1,
			SchemaId:  "second_schemaId_service_ms",
			Schema:    "second_schema_service_ms",
		})
		assert.NoError(t, err)
		assert.Equal(t, proto.Response_SUCCESS, respModifySchema.Response.GetCode())

	})

	t.Run("create a schema in production env", func(t *testing.T) {
		var (
			serviceIdPro1 string
			serviceIdPro2 string
		)

		log.Info("register service")
		respCreateService, err := ms.MicroService().RegisterService(getContext(), &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				AppId:       "create_schema_prod_service_ms",
				ServiceName: "create_schema_service_service_ms",
				Version:     "1.0.0",
				Level:       "FRONT",
				Status:      pb.MS_UP,
				Environment: pb.ENV_PROD,
			},
		})
		assert.NoError(t, err)
		assert.Equal(t, proto.Response_SUCCESS, respCreateService.Response.GetCode())
		serviceIdPro1 = respCreateService.ServiceId

		respCreateService, err = ms.MicroService().RegisterService(getContext(), &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				AppId:       "create_schema_prod_service_ms",
				ServiceName: "create_schema_service_service_ms",
				Version:     "1.0.1",
				Level:       "FRONT",
				Schemas: []string{
					"first_schemaId_service_ms",
					"second_schemaId_service_ms",
				},
				Status:      pb.MS_UP,
				Environment: pb.ENV_PROD,
			},
		})
		assert.NoError(t, err)
		assert.Equal(t, proto.Response_SUCCESS, respCreateService.Response.GetCode())
		serviceIdPro2 = respCreateService.ServiceId

		log.Info("create a schema for service whose schemaID is empty")
		respModifySchema, err := ms.MicroService().ModifySchema(getContext(), &pb.ModifySchemaRequest{
			ServiceId: serviceIdPro1,
			SchemaId:  "first_schemaId_service_ms",
			Schema:    "first_schema_service_ms",
		})
		assert.NoError(t, err)
		assert.Equal(t, proto.Response_SUCCESS, respModifySchema.Response.GetCode())

		log.Info("modify schema for the service whose schema summary is empty")
		respModifySchema, err = ms.MicroService().ModifySchema(getContext(), &pb.ModifySchemaRequest{
			ServiceId: serviceIdPro1,
			SchemaId:  "first_schemaId_service_ms",
			Schema:    "first_schema_change_service_ms",
			Summary:   "first0summary1change_service_ms",
		})
		assert.NoError(t, err)
		assert.Equal(t, proto.Response_SUCCESS, respModifySchema.Response.GetCode())

		log.Info("modify schema for the service whose schema summary already exist")
		respModifySchema, err = ms.MicroService().ModifySchema(getContext(), &pb.ModifySchemaRequest{
			ServiceId: serviceIdPro1,
			SchemaId:  "first_schemaId_service_ms",
			Schema:    "first_schema_service_ms",
			Summary:   "first0summary_service_ms",
		})
		assert.NoError(t, err)
		assert.NotEqual(t, proto.Response_SUCCESS, respModifySchema.Response.GetCode())

		log.Info("add schema")
		respModifySchema, err = ms.MicroService().ModifySchema(getContext(), &pb.ModifySchemaRequest{
			ServiceId: serviceIdPro1,
			SchemaId:  "second_schemaId_service_ms",
			Schema:    "second_schema_service_ms",
		})
		assert.NoError(t, err)
		assert.NotEqual(t, proto.Response_SUCCESS, respModifySchema.Response.GetCode())

		log.Info("modify schema for the service whose schemaId already exist")
		respModifySchema, err = ms.MicroService().ModifySchema(getContext(), &pb.ModifySchemaRequest{
			ServiceId: serviceIdPro2,
			SchemaId:  "first_schemaId_service_ms",
			Schema:    "first_schema_service_ms",
		})
		assert.NoError(t, err)
		assert.Equal(t, proto.Response_SUCCESS, respModifySchema.Response.GetCode())

	})

	t.Run("create a schema in empty env", func(t *testing.T) {
		var (
			serviceIdPro1 string
			serviceIdPro2 string
		)

		log.Info("register service")
		respCreateService, err := ms.MicroService().RegisterService(getContext(), &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				AppId:       "create_schema_empty_service_ms",
				ServiceName: "create_schema_service_service_ms",
				Version:     "1.0.0",
				Level:       "FRONT",
				Status:      pb.MS_UP,
			},
		})
		assert.NoError(t, err)
		assert.Equal(t, proto.Response_SUCCESS, respCreateService.Response.GetCode())
		serviceIdPro1 = respCreateService.ServiceId

		respCreateService, err = ms.MicroService().RegisterService(getContext(), &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				AppId:       "create_schema_empty_service_ms",
				ServiceName: "create_schema_service_service_ms",
				Version:     "1.0.1",
				Level:       "FRONT",
				Schemas: []string{
					"first_schemaId_service_ms",
					"second_schemaId_service_ms",
				},
				Status: pb.MS_UP,
			},
		})
		assert.NoError(t, err)
		assert.Equal(t, proto.Response_SUCCESS, respCreateService.Response.GetCode())
		serviceIdPro2 = respCreateService.ServiceId

		log.Info("create a schema for service whose schemaID is empty")
		respModifySchema, err := ms.MicroService().ModifySchema(getContext(), &pb.ModifySchemaRequest{
			ServiceId: serviceIdPro1,
			SchemaId:  "first_schemaId_service_ms",
			Schema:    "first_schema_service_ms",
		})
		assert.NoError(t, err)
		assert.Equal(t, proto.Response_SUCCESS, respModifySchema.Response.GetCode())

		log.Info("modify schema for the service whose schema summary is empty")
		respModifySchema, err = ms.MicroService().ModifySchema(getContext(), &pb.ModifySchemaRequest{
			ServiceId: serviceIdPro1,
			SchemaId:  "first_schemaId_service_ms",
			Schema:    "first_schema_change_service_ms",
			Summary:   "first0summary1change_service_ms",
		})
		assert.NoError(t, err)
		assert.Equal(t, proto.Response_SUCCESS, respModifySchema.Response.GetCode())

		log.Info("modify schema for the service whose schema summary already exist")
		respModifySchema, err = ms.MicroService().ModifySchema(getContext(), &pb.ModifySchemaRequest{
			ServiceId: serviceIdPro1,
			SchemaId:  "first_schemaId_service_ms",
			Schema:    "first_schema_service_ms",
			Summary:   "first0summary_service_ms",
		})
		assert.NoError(t, err)
		assert.NotEqual(t, proto.Response_SUCCESS, respModifySchema.Response.GetCode())

		log.Info("add schema")
		respModifySchema, err = ms.MicroService().ModifySchema(getContext(), &pb.ModifySchemaRequest{
			ServiceId: serviceIdPro1,
			SchemaId:  "second_schemaId_service_ms",
			Schema:    "second_schema_service_ms",
		})
		assert.NoError(t, err)
		assert.NotEqual(t, proto.Response_SUCCESS, respModifySchema.Response.GetCode())

		log.Info("modify schema for the service whose schemaId already exist")
		respModifySchema, err = ms.MicroService().ModifySchema(getContext(), &pb.ModifySchemaRequest{
			ServiceId: serviceIdPro2,
			SchemaId:  "first_schemaId_service_ms",
			Schema:    "first_schema_service_ms",
		})
		assert.NoError(t, err)
		assert.Equal(t, proto.Response_SUCCESS, respModifySchema.Response.GetCode())
	})

	t.Run("add a schemaId in production env while schema editable is set", func(t *testing.T) {
		var (
			serviceIdPro1 string
		)
		log.Info("register service")
		respCreateService, err := ms.MicroService().RegisterService(getContext(), &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				AppId:       "add_a_schemaId_prod_schema_lock_ms",
				ServiceName: "add_a_schemaId_prod_schema_lock_service_ms",
				Version:     "1.0.0",
				Level:       "FRONT",
				Status:      pb.MS_UP,
				Environment: pb.ENV_PROD,
			},
		})
		assert.NoError(t, err)
		assert.Equal(t, proto.Response_SUCCESS, respCreateService.Response.GetCode())
		serviceIdPro1 = respCreateService.ServiceId

		log.Info("add a schema with new schemaId, should pass")
		schemas := []*pb.Schema{
			{
				SchemaId: "first_schemaId_ms",
				Schema:   "first_schema_ms",
				Summary:  "first0summary_ms",
			},
		}
		respModifySchemas, err := ms.MicroService().ModifySchemas(getContext(), &pb.ModifySchemasRequest{
			ServiceId: serviceIdPro1,
			Schemas:   schemas,
		})
		assert.NoError(t, err)
		assert.Equal(t, proto.Response_SUCCESS, respModifySchemas.Response.GetCode())

		respService, err := ms.MicroService().GetService(getContext(), &pb.GetServiceRequest{
			ServiceId: serviceIdPro1,
		})
		assert.NoError(t, err)
		assert.Equal(t, proto.Response_SUCCESS, respService.Response.GetCode())
		assert.Equal(t, []string{"first_schemaId_ms"}, respService.Service.Schemas)

		schemas = []*pb.Schema{
			{
				SchemaId: "second_schemaId_ms",
				Schema:   "second_schema_ms",
				Summary:  "second0summary_ms",
			},
		}
		log.Info("schema edit not allowed, add a schema with new schemaId should fail")

		localMicroServiceDs := &etcd.DataSource{SchemaEditable: false}
		respModifySchemas, err = localMicroServiceDs.ModifySchemas(getContext(), &pb.ModifySchemasRequest{
			ServiceId: serviceIdPro1,
			Schemas:   schemas,
		})
		assert.NoError(t, err)
		assert.Equal(t, scerr.ErrUndefinedSchemaID, respModifySchemas.Response.GetCode())

		log.Info("schema edit allowed, add a schema with new schemaId, should pass")
		localMicroServiceDs = &etcd.DataSource{SchemaEditable: true}
		respModifySchemas, err = localMicroServiceDs.ModifySchemas(getContext(), &pb.ModifySchemasRequest{
			ServiceId: serviceIdPro1,
			Schemas:   schemas,
		})
		assert.NoError(t, err)
		assert.Equal(t, proto.Response_SUCCESS, respModifySchemas.Response.GetCode())
	})

	t.Run("modify a schema in production env while schema editable is set", func(t *testing.T) {
		var (
			serviceIdPro1 string
		)
		log.Info("register service")
		respCreateService, err := ms.MicroService().RegisterService(getContext(), &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				AppId:       "modify_a_schema_prod_schema_lock_ms",
				ServiceName: "modify_a_schema_prod_schema_lock_service_ms",
				Version:     "1.0.0",
				Level:       "FRONT",
				Status:      pb.MS_UP,
				Environment: pb.ENV_PROD,
			},
		})
		assert.NoError(t, err)
		assert.Equal(t, proto.Response_SUCCESS, respCreateService.Response.GetCode())
		serviceIdPro1 = respCreateService.ServiceId

		log.Info("add schemas, should pass")
		schemas := []*pb.Schema{
			{
				SchemaId: "first_schemaId_ms",
				Schema:   "first_schema_ms",
				Summary:  "first0summary_ms",
			},
		}
		respModifySchemas, err := ms.MicroService().ModifySchemas(getContext(), &pb.ModifySchemasRequest{
			ServiceId: serviceIdPro1,
			Schemas:   schemas,
		})
		assert.NoError(t, err)
		assert.Equal(t, proto.Response_SUCCESS, respModifySchemas.Response.GetCode())

		respService, err := ms.MicroService().GetService(getContext(), &pb.GetServiceRequest{
			ServiceId: serviceIdPro1,
		})
		assert.NoError(t, err)
		assert.Equal(t, []string{"first_schemaId_ms"}, respService.Service.Schemas)

		log.Info("schema edit not allowed, modify schema should fail")
		localMicroServiceDs := &etcd.DataSource{SchemaEditable: false}
		respModifySchema, err := localMicroServiceDs.ModifySchema(getContext(), &pb.ModifySchemaRequest{
			ServiceId: serviceIdPro1,
			SchemaId:  schemas[0].SchemaId,
			Summary:   schemas[0].Summary,
			Schema:    schemas[0].SchemaId,
		})
		assert.NoError(t, err)
		assert.Equal(t, scerr.ErrModifySchemaNotAllow, respModifySchema.Response.GetCode())

		log.Info("schema edit allowed, add a schema with new schemaId, should pass")
		localMicroServiceDs = &etcd.DataSource{SchemaEditable: true}
		respModifySchema, err = localMicroServiceDs.ModifySchema(getContext(), &pb.ModifySchemaRequest{
			ServiceId: serviceIdPro1,
			SchemaId:  schemas[0].SchemaId,
			Summary:   schemas[0].Summary,
			Schema:    schemas[0].SchemaId,
		})
		assert.NoError(t, err)
		assert.Equal(t, proto.Response_SUCCESS, respModifySchema.Response.GetCode())
	})
}

func TestSchema_Exist(t *testing.T) {
	ms.Install("etcd", func(opts ms.Options) (ms.DataSource, error) {
		return etcd.NewDataSource(), nil
	})
	err := ms.Init(ms.Options{
		Endpoint:       "",
		PluginImplName: ms.ImplName(archaius.GetString("servicecomb.ms.name", "etcd")),
	})
	assert.NoError(t, err)

	var (
		serviceId string
	)

	t.Run("register service and add schema", func(t *testing.T) {
		log.Info("register service")
		respCreateService, err := ms.MicroService().RegisterService(getContext(), &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				AppId:       "query_schema_group_ms",
				ServiceName: "query_schema_service_ms",
				Version:     "1.0.0",
				Level:       "FRONT",
				Status:      pb.MS_UP,
				Environment: pb.ENV_DEV,
			},
		})
		assert.NoError(t, err)
		assert.Equal(t, proto.Response_SUCCESS, respCreateService.Response.GetCode())
		serviceId = respCreateService.ServiceId

		log.Info("add schemas, should pass")
		resp, err := ms.MicroService().ModifySchema(getContext(), &pb.ModifySchemaRequest{
			ServiceId: serviceId,
			SchemaId:  "com.huawei.test.ms",
			Schema:    "query schema ms",
			Summary:   "summary_ms",
		})
		assert.NoError(t, err)
		assert.Equal(t, proto.Response_SUCCESS, resp.Response.GetCode())

		resp, err = ms.MicroService().ModifySchema(getContext(), &pb.ModifySchemaRequest{
			ServiceId: serviceId,
			SchemaId:  "com.huawei.test.no.summary.ms",
			Schema:    "query schema ms",
		})
		assert.NoError(t, err)
		assert.Equal(t, proto.Response_SUCCESS, resp.Response.GetCode())
	})

	t.Run("check exists", func(t *testing.T) {
		log.Info("check schema exist, should pass")
		resp, err := ms.MicroService().ExistSchema(getContext(), &pb.GetExistenceRequest{
			Type:      service.ExistTypeSchema,
			ServiceId: serviceId,
			SchemaId:  "com.huawei.test.ms",
		})
		assert.NoError(t, err)
		assert.Equal(t, proto.Response_SUCCESS, resp.Response.GetCode())
		assert.Equal(t, "summary_ms", resp.Summary)

		resp, err = ms.MicroService().ExistSchema(getContext(), &pb.GetExistenceRequest{
			Type:        service.ExistTypeSchema,
			ServiceId:   serviceId,
			SchemaId:    "com.huawei.test.ms",
			AppId:       "()",
			ServiceName: "",
			Version:     "()",
		})
		assert.NoError(t, err)
		assert.Equal(t, proto.Response_SUCCESS, resp.Response.GetCode())

		resp, err = ms.MicroService().ExistSchema(getContext(), &pb.GetExistenceRequest{
			Type:      service.ExistTypeSchema,
			ServiceId: serviceId,
			SchemaId:  "com.huawei.test.no.summary.ms",
		})
		assert.NoError(t, err)
		assert.Equal(t, proto.Response_SUCCESS, resp.Response.GetCode())
		assert.Equal(t, "com.huawei.test.no.summary.ms", resp.SchemaId)
		assert.Equal(t, "", resp.Summary)
	})
}
