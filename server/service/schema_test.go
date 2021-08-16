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
package service_test

import (
	"strconv"
	"strings"

	"github.com/apache/servicecomb-service-center/server/core/proto"

	pb "github.com/apache/servicecomb-service-center/pkg/registry"
	"github.com/apache/servicecomb-service-center/server/plugin/quota"
	scerr "github.com/apache/servicecomb-service-center/server/scerror"
	"github.com/apache/servicecomb-service-center/server/service"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

const (
	invalidSchemaId = "@"
)

var (
	TOO_LONG_SUMMARY = strings.Repeat("x", 129)
)

var _ = Describe("'Schema' service", func() {
	Describe("execute 'create' operation", func() {
		var (
			serviceIdDev string
			serviceId    string
		)

		It("should be passed, create service", func() {
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
			Expect(err).To(BeNil())
			Expect(respCreateService.Response.GetCode()).To(Equal(proto.Response_SUCCESS))
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
			Expect(err).To(BeNil())
			Expect(respCreateService.Response.GetCode()).To(Equal(proto.Response_SUCCESS))
			serviceId = respCreateService.ServiceId
		})

		Context("when create an invalid schema", func() {
			f := func() {
				By("service id is empty")
				resp, err := serviceResource.ModifySchema(getContext(), &pb.ModifySchemaRequest{
					ServiceId: "",
					SchemaId:  "com.huawei.test",
					Schema:    "create schema",
				})
				Expect(err).To(BeNil())
				Expect(resp.Response.GetCode()).To(Equal(scerr.ErrInvalidParams))

				By("service does not exist")
				resp, err = serviceResource.ModifySchema(getContext(), &pb.ModifySchemaRequest{
					ServiceId: "notExistService",
					SchemaId:  "com.huawei.test",
					Schema:    "create schema",
				})
				Expect(err).To(BeNil())
				Expect(resp.Response.GetCode()).To(Equal(scerr.ErrServiceNotExists))

				By("schema id is invalid")
				resp, err = serviceResource.ModifySchema(getContext(), &pb.ModifySchemaRequest{
					ServiceId: serviceIdDev,
					SchemaId:  invalidSchemaId,
					Schema:    "create schema",
				})
				Expect(err).To(BeNil())
				Expect(resp.Response.GetCode()).To(Equal(scerr.ErrInvalidParams))

				By("summary is invalid")
				resp, err = serviceResource.ModifySchema(getContext(), &pb.ModifySchemaRequest{
					ServiceId: serviceIdDev,
					SchemaId:  "com.huawei.test",
					Schema:    "create schema",
					Summary:   TOO_LONG_SUMMARY,
				})
				Expect(err).To(BeNil())
				Expect(resp.Response.GetCode()).To(Equal(scerr.ErrInvalidParams))

				resp, err = serviceResource.ModifySchema(getContext(), &pb.ModifySchemaRequest{
					ServiceId: serviceIdDev,
					SchemaId:  "com.huawei.test",
					Schema:    "create schema",
					Summary:   "_",
				})
				Expect(err).To(BeNil())
				Expect(resp.Response.GetCode()).To(Equal(scerr.ErrInvalidParams))
			}

			It("should be failed in prod env", func() {
				old := serviceIdDev
				serviceIdDev = serviceId
				f()
				serviceIdDev = old
			})
			It("should be failed in dev env", f)
		})

		Context("when batch create invalid schemas", func() {
			f := func() {
				By("service does not exist")
				respCreateService, err := serviceResource.ModifySchemas(getContext(), &pb.ModifySchemasRequest{
					ServiceId: "not_exist_serviceId",
				})
				Expect(err).To(BeNil())
				Expect(respCreateService.Response.GetCode()).ToNot(Equal(proto.Response_SUCCESS))

				By("service id is empty")
				respCreateService, err = serviceResource.ModifySchemas(getContext(), &pb.ModifySchemasRequest{
					ServiceId: "",
				})
				Expect(err).To(BeNil())
				Expect(respCreateService.Response.GetCode()).ToNot(Equal(proto.Response_SUCCESS))

				By("schema id is invalid")
				respCreateService, err = serviceResource.ModifySchemas(getContext(), &pb.ModifySchemasRequest{
					ServiceId: serviceIdDev,
					Schemas: []*pb.Schema{
						{
							SchemaId: invalidSchemaId,
						},
					},
				})
				Expect(err).To(BeNil())
				Expect(respCreateService.Response.GetCode()).ToNot(Equal(proto.Response_SUCCESS))

				By("schema is empty")
				respCreateService, err = serviceResource.ModifySchemas(getContext(), &pb.ModifySchemasRequest{
					ServiceId: serviceIdDev,
					Schemas: []*pb.Schema{
						{
							SchemaId: "com.huawei.test",
						},
					},
				})
				Expect(err).To(BeNil())
				Expect(respCreateService.Response.GetCode()).ToNot(Equal(proto.Response_SUCCESS))

				By("summary is empty")
				respCreateService, err = serviceResource.ModifySchemas(getContext(), &pb.ModifySchemasRequest{
					ServiceId: serviceIdDev,
					Schemas: []*pb.Schema{
						{
							SchemaId: "com.huawei.test",
							Schema:   "create schema",
						},
					},
				})
				Expect(err).To(BeNil())
				Expect(respCreateService.Response.GetCode()).ToNot(Equal(proto.Response_SUCCESS))

				By("summery is invalid")
				respCreateService, err = serviceResource.ModifySchemas(getContext(), &pb.ModifySchemasRequest{
					ServiceId: serviceIdDev,
					Schemas: []*pb.Schema{
						{
							SchemaId: "com.huawei.test",
							Schema:   "create schema",
							Summary:  TOO_LONG_SUMMARY,
						},
					},
				})
				Expect(err).To(BeNil())
				Expect(respCreateService.Response.GetCode()).ToNot(Equal(proto.Response_SUCCESS))
			}
			It("should be failed in dev env", f)

			It("should be failed in prod env", func() {
				old := serviceIdDev
				serviceIdDev = serviceId
				f()
				serviceIdDev = old
			})
		})

		Context("when create schemas out of gauge", func() {
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

			It("should be failed in dev env", func() {
				By("batch modify schemas 1")
				respCreateSchemas, err := serviceResource.ModifySchemas(getContext(), &pb.ModifySchemasRequest{
					ServiceId: serviceIdDev,
					Schemas:   schemas,
				})
				Expect(err).To(BeNil())
				Expect(respCreateSchemas.Response.GetCode()).To(Equal(scerr.ErrInvalidParams))

				By("batch modify schemas 2")
				respCreateSchemas, err = serviceResource.ModifySchemas(getContext(), &pb.ModifySchemasRequest{
					ServiceId: serviceIdDev,
					Schemas:   schemas[:quota.DefaultSchemaQuota],
				})
				Expect(err).To(BeNil())
				Expect(respCreateSchemas.Response.GetCode()).To(Equal(proto.Response_SUCCESS))

				By("modify one schema")
				respCreateService := &pb.ModifySchemaResponse{}
				schema := schemas[quota.DefaultSchemaQuota]
				respCreateService, err = serviceResource.ModifySchema(getContext(), &pb.ModifySchemaRequest{
					ServiceId: serviceIdDev,
					SchemaId:  schema.SchemaId,
					Schema:    schema.Schema,
				})
				Expect(err).To(BeNil())
				Expect(respCreateService.Response.GetCode()).To(Equal(scerr.ErrNotEnoughQuota))
			})

			It("should be failed in prod env", func() {
				respCreateService, err := serviceResource.ModifySchemas(getContext(), &pb.ModifySchemasRequest{
					ServiceId: serviceIdDev,
					Schemas:   schemas,
				})
				Expect(err).To(BeNil())
				Expect(respCreateService.Response.GetCode()).To(Equal(scerr.ErrInvalidParams))
			})

			It("should be failed when create service", func() {
				respServiceForSchema, err := serviceResource.Create(getContext(), &pb.CreateServiceRequest{
					Service: &pb.MicroService{
						AppId:       "check_schema_group",
						ServiceName: "check_schema_service",
						Version:     "1.0.0",
						Level:       "FRONT",
						Schemas:     schemaIds,
						Status:      pb.MS_UP,
					},
				})
				Expect(err).To(BeNil())
				Expect(respServiceForSchema.Response.GetCode()).ToNot(Equal(proto.Response_SUCCESS))
			})
		})

		Context("when batch create schemas in dev env", func() {
			var (
				serviceIdDev1 string
				serviceIdDev2 string
			)

			It("should be passed", func() {
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
				Expect(err).To(BeNil())
				Expect(respCreateService.Response.GetCode()).To(Equal(proto.Response_SUCCESS))
				serviceIdDev1 = respCreateService.ServiceId

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
				Expect(err).To(BeNil())
				Expect(respCreateService.Response.GetCode()).To(Equal(proto.Response_SUCCESS))
				serviceIdDev2 = respCreateService.ServiceId
			})

			It("should be passed", func() {
				By("create schemas when service schema id set is empty")
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
				respCreateService, err := serviceResource.ModifySchemas(getContext(), &pb.ModifySchemasRequest{
					ServiceId: serviceIdDev1,
					Schemas:   schemas,
				})
				Expect(err).To(BeNil())
				Expect(respCreateService.Response.GetCode()).To(Equal(proto.Response_SUCCESS))

				respGetAllSchema, err := serviceResource.GetAllSchemaInfo(getContext(), &pb.GetAllSchemaRequest{
					ServiceId: serviceIdDev1,
				})
				Expect(err).To(BeNil())
				Expect(respGetAllSchema.Response.GetCode()).To(Equal(proto.Response_SUCCESS))
				Expect(len(respGetAllSchema.Schemas)).To(Equal(1))

				By("modify schemas when service schema id already exists")
				schemas = []*pb.Schema{}
				respCreateService, err = serviceResource.ModifySchemas(getContext(), &pb.ModifySchemasRequest{
					ServiceId: serviceIdDev2,
					Schemas:   schemas,
				})
				Expect(err).To(BeNil())
				Expect(respCreateService.Response.GetCode()).To(Equal(scerr.ErrInvalidParams))

				By("modify schemas")
				schemas = []*pb.Schema{
					{
						SchemaId: "first_schemaId",
						Schema:   "first_schema_change",
						Summary:  "first0summary1change",
					},
				}
				respCreateService, err = serviceResource.ModifySchemas(getContext(), &pb.ModifySchemasRequest{
					ServiceId: serviceIdDev1,
					Schemas:   schemas,
				})
				Expect(err).To(BeNil())
				Expect(respCreateService.Response.GetCode()).To(Equal(proto.Response_SUCCESS))

				By("add schemas")
				schemas = []*pb.Schema{
					{
						SchemaId: "second_schemaId",
						Schema:   "second_schema",
						Summary:  "second0summary",
					},
				}
				respCreateService, err = serviceResource.ModifySchemas(getContext(), &pb.ModifySchemasRequest{
					ServiceId: serviceIdDev1,
					Schemas:   schemas,
				})
				Expect(err).To(BeNil())
				Expect(respCreateService.Response.GetCode()).To(Equal(proto.Response_SUCCESS))

				respGetOne, err := serviceResource.GetOne(getContext(), &pb.GetServiceRequest{
					ServiceId: serviceIdDev1,
				})
				Expect(err).To(BeNil())
				Expect(respGetOne.Response.GetCode()).To(Equal(proto.Response_SUCCESS))
				Expect(respGetOne.Service.Schemas).To(Equal([]string{"second_schemaId"}))

				By("create empty")
				schemas = []*pb.Schema{}
				respCreateService, err = serviceResource.ModifySchemas(getContext(), &pb.ModifySchemasRequest{
					ServiceId: serviceIdDev1,
					Schemas:   schemas,
				})
				Expect(err).To(BeNil())
				Expect(respCreateService.Response.GetCode()).To(Equal(scerr.ErrInvalidParams))

				By("add new schemaId not exist in service schemaId")
				schemas = []*pb.Schema{
					{
						SchemaId: "second_schemaId",
						Schema:   "second_schema",
						Summary:  "second0summary",
					},
				}
				respCreateService, err = serviceResource.ModifySchemas(getContext(), &pb.ModifySchemasRequest{
					ServiceId: serviceIdDev2,
					Schemas:   schemas,
				})
				Expect(err).To(BeNil())
				Expect(respCreateService.Response.GetCode()).To(Equal(proto.Response_SUCCESS))

				respGetOne, err = serviceResource.GetOne(getContext(), &pb.GetServiceRequest{
					ServiceId: serviceIdDev2,
				})
				Expect(err).To(BeNil())
				Expect(respGetOne.Response.GetCode()).To(Equal(proto.Response_SUCCESS))
				Expect(respGetOne.Service.Schemas).To(Equal([]string{"second_schemaId"}))
			})
		})

		Context("when batch create schemas in prod env", func() {
			var (
				serviceIdPro1 string
				serviceIdPro2 string
			)

			It("should be passed", func() {
				respCreateService, err := serviceResource.Create(getContext(), &pb.CreateServiceRequest{
					Service: &pb.MicroService{
						AppId:       "create_schemas_prod",
						ServiceName: "create_schemas_service",
						Version:     "1.0.0",
						Level:       "FRONT",
						Status:      pb.MS_UP,
						Environment: pb.ENV_PROD,
					},
				})
				Expect(err).To(BeNil())
				Expect(respCreateService.Response.GetCode()).To(Equal(proto.Response_SUCCESS))
				serviceIdPro1 = respCreateService.ServiceId

				respCreateService, err = serviceResource.Create(getContext(), &pb.CreateServiceRequest{
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
				Expect(err).To(BeNil())
				Expect(respCreateService.Response.GetCode()).To(Equal(proto.Response_SUCCESS))
				serviceIdPro2 = respCreateService.ServiceId
			})

			It("should be passed", func() {
				By("add schemas when service schema id set is empty")
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
				respModifySchemas, err := serviceResource.ModifySchemas(getContext(), &pb.ModifySchemasRequest{
					ServiceId: serviceIdPro1,
					Schemas:   schemas,
				})
				Expect(err).To(BeNil())
				Expect(respModifySchemas.Response.GetCode()).To(Equal(proto.Response_SUCCESS))

				respGetOne, err := serviceResource.GetOne(getContext(), &pb.GetServiceRequest{
					ServiceId: serviceIdPro1,
				})
				Expect(err).To(BeNil())
				Expect(respGetOne.Response.GetCode()).To(Equal(proto.Response_SUCCESS))
				Expect(respGetOne.Service.Schemas).To(Equal([]string{"first_schemaId"}))

				respGetAllSchema, err := serviceResource.GetAllSchemaInfo(getContext(), &pb.GetAllSchemaRequest{
					ServiceId: serviceIdPro1,
				})
				Expect(err).To(BeNil())
				Expect(respGetAllSchema.Response.GetCode()).To(Equal(proto.Response_SUCCESS))
				Expect(len(respGetAllSchema.Schemas)).To(Equal(1))

				By("modify schemas content already exists, will skip more exist schema")
				respModifySchemas, err = serviceResource.ModifySchemas(getContext(), &pb.ModifySchemasRequest{
					ServiceId: serviceIdPro1,
					Schemas:   schemas,
				})
				Expect(err).To(BeNil())
				Expect(respModifySchemas.Response.GetCode()).To(Equal(proto.Response_SUCCESS))

				By("add schemas, non-exist schemaId")
				schemas = []*pb.Schema{
					{
						SchemaId: "second_schemaId",
						Schema:   "second_schema",
						Summary:  "second0summary",
					},
				}
				respModifySchemas, err = serviceResource.ModifySchemas(getContext(), &pb.ModifySchemasRequest{
					ServiceId: serviceIdPro1,
					Schemas:   schemas,
				})
				Expect(err).To(BeNil())
				Expect(respModifySchemas.Response.GetCode()).To(Equal(scerr.ErrUndefinedSchemaID))

				respExist, err := serviceResource.Exist(getContext(), &pb.GetExistenceRequest{
					Type:      service.ExistTypeSchema,
					ServiceId: serviceIdPro1,
					SchemaId:  "first_schemaId",
				})

				Expect(err).To(BeNil())
				Expect(respExist.Summary).To(Equal("first0summary"))
			})

			It("summary is empty", func() {
				By("add schema when summary is empty")
				respModifySchema, err := serviceResource.ModifySchema(getContext(), &pb.ModifySchemaRequest{
					ServiceId: serviceIdPro2,
					SchemaId:  "first_schemaId",
					Schema:    "first_schema",
				})
				Expect(err).To(BeNil())
				Expect(respModifySchema.Response.GetCode()).To(Equal(proto.Response_SUCCESS))

				By("add schemas when summary in database is empty")
				schemas := []*pb.Schema{
					{
						SchemaId: "first_schemaId",
						Schema:   "first_schema",
						Summary:  "first0summary",
					},
				}

				respModifySchemas, err := serviceResource.ModifySchemas(getContext(), &pb.ModifySchemasRequest{
					ServiceId: serviceIdPro2,
					Schemas:   schemas,
				})
				Expect(err).To(BeNil())
				Expect(respModifySchemas.Response.GetCode()).To(Equal(proto.Response_SUCCESS))

				respExist, err := serviceResource.Exist(getContext(), &pb.GetExistenceRequest{
					Type:      service.ExistTypeSchema,
					ServiceId: serviceIdPro2,
					SchemaId:  "first_schemaId",
				})

				Expect(err).To(BeNil())
				Expect(respExist.Summary).To(Equal("first0summary"))

				schemas = []*pb.Schema{
					{
						SchemaId: "second_schemaId",
						Schema:   "second_schema",
						Summary:  "second0summary",
					},
				}
				respModifySchemas, err = serviceResource.ModifySchemas(getContext(), &pb.ModifySchemasRequest{
					ServiceId: serviceIdPro2,
					Schemas:   schemas,
				})
				Expect(err).To(BeNil())
				Expect(respModifySchemas.Response.GetCode()).To(Equal(proto.Response_SUCCESS))
			})
		})

		Context("when create a schema in dev env", func() {
			var (
				serviceIdDev1 string
				serviceIdDev2 string
			)

			It("should be passed", func() {
				respCreateService, err := serviceResource.Create(getContext(), &pb.CreateServiceRequest{
					Service: &pb.MicroService{
						AppId:       "create_schema_dev",
						ServiceName: "create_schema_service",
						Version:     "1.0.0",
						Level:       "FRONT",
						Status:      pb.MS_UP,
						Environment: pb.ENV_DEV,
					},
				})
				Expect(err).To(BeNil())
				Expect(respCreateService.Response.GetCode()).To(Equal(proto.Response_SUCCESS))
				serviceIdDev1 = respCreateService.ServiceId

				respCreateService, err = serviceResource.Create(getContext(), &pb.CreateServiceRequest{
					Service: &pb.MicroService{
						AppId:       "create_schema_dev",
						ServiceName: "create_schema_service",
						Version:     "1.0.1",
						Level:       "FRONT",
						Schemas: []string{
							"first_schemaId",
						},
						Status:      pb.MS_UP,
						Environment: pb.ENV_DEV,
					},
				})
				Expect(err).To(BeNil())
				Expect(respCreateService.Response.GetCode()).To(Equal(proto.Response_SUCCESS))
				serviceIdDev2 = respCreateService.ServiceId
			})

			It("should be passed", func() {
				By("create schema when service schema id set is empty")
				respModifySchema, err := serviceResource.ModifySchema(getContext(), &pb.ModifySchemaRequest{
					ServiceId: serviceIdDev1,
					SchemaId:  "first_schemaId",
					Schema:    "first_schema",
				})
				Expect(err).To(BeNil())
				Expect(respModifySchema.Response.GetCode()).To(Equal(proto.Response_SUCCESS))

				By("create schema when service schema id already exists")
				respModifySchema, err = serviceResource.ModifySchema(getContext(), &pb.ModifySchemaRequest{
					ServiceId: serviceIdDev2,
					SchemaId:  "first_schemaId",
					Schema:    "first_schema",
				})
				Expect(err).To(BeNil())
				Expect(respModifySchema.Response.GetCode()).To(Equal(proto.Response_SUCCESS))

				By("modify schema summary is empty")
				respModifySchema, err = serviceResource.ModifySchema(getContext(), &pb.ModifySchemaRequest{
					ServiceId: serviceIdDev1,
					SchemaId:  "first_schemaId",
					Schema:    "first_schema_change",
					Summary:   "first0summary1change",
				})
				Expect(err).To(BeNil())
				Expect(respModifySchema.Response.GetCode()).To(Equal(proto.Response_SUCCESS))

				By("modify schema summary already exists")
				respModifySchema, err = serviceResource.ModifySchema(getContext(), &pb.ModifySchemaRequest{
					ServiceId: serviceIdDev1,
					SchemaId:  "first_schemaId",
					Schema:    "first_schema",
					Summary:   "first0summary",
				})
				Expect(err).To(BeNil())
				Expect(respModifySchema.Response.GetCode()).To(Equal(proto.Response_SUCCESS))

				By("add schema")
				respModifySchema, err = serviceResource.ModifySchema(getContext(), &pb.ModifySchemaRequest{
					ServiceId: serviceIdDev1,
					SchemaId:  "second_schemaId",
					Schema:    "second_schema",
				})
				Expect(err).To(BeNil())
				Expect(respModifySchema.Response.GetCode()).To(Equal(proto.Response_SUCCESS))
			})
		})

		Context("when create a schema in prod env", func() {
			var (
				serviceIdPro1 string
				serviceIdPro2 string
			)

			It("should be passed", func() {
				respCreateService, err := serviceResource.Create(getContext(), &pb.CreateServiceRequest{
					Service: &pb.MicroService{
						AppId:       "create_schema_prod",
						ServiceName: "create_schema_service",
						Version:     "1.0.0",
						Level:       "FRONT",
						Status:      pb.MS_UP,
						Environment: pb.ENV_PROD,
					},
				})
				Expect(err).To(BeNil())
				Expect(respCreateService.Response.GetCode()).To(Equal(proto.Response_SUCCESS))
				serviceIdPro1 = respCreateService.ServiceId

				respCreateService, err = serviceResource.Create(getContext(), &pb.CreateServiceRequest{
					Service: &pb.MicroService{
						AppId:       "create_schema_prod",
						ServiceName: "create_schema_service",
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
				Expect(err).To(BeNil())
				Expect(respCreateService.Response.GetCode()).To(Equal(proto.Response_SUCCESS))
				serviceIdPro2 = respCreateService.ServiceId
			})

			It("should be passed", func() {
				By("create schema when service schema id set is empty")
				respModifySchema, err := serviceResource.ModifySchema(getContext(), &pb.ModifySchemaRequest{
					ServiceId: serviceIdPro1,
					SchemaId:  "first_schemaId",
					Schema:    "first_schema",
				})
				Expect(err).To(BeNil())
				Expect(respModifySchema.Response.GetCode()).To(Equal(proto.Response_SUCCESS))

				By("modify schema summary is empty")
				respModifySchema, err = serviceResource.ModifySchema(getContext(), &pb.ModifySchemaRequest{
					ServiceId: serviceIdPro1,
					SchemaId:  "first_schemaId",
					Schema:    "first_schema_change",
					Summary:   "first0summary1change",
				})
				Expect(err).To(BeNil())
				Expect(respModifySchema.Response.GetCode()).To(Equal(proto.Response_SUCCESS))

				By("modify schema summary already exists")
				respModifySchema, err = serviceResource.ModifySchema(getContext(), &pb.ModifySchemaRequest{
					ServiceId: serviceIdPro1,
					SchemaId:  "first_schemaId",
					Schema:    "first_schema",
					Summary:   "first0summary",
				})
				Expect(err).To(BeNil())
				Expect(respModifySchema.Response.GetCode()).ToNot(Equal(proto.Response_SUCCESS))

				By("add schema")
				respModifySchema, err = serviceResource.ModifySchema(getContext(), &pb.ModifySchemaRequest{
					ServiceId: serviceIdPro1,
					SchemaId:  "second_schemaId",
					Schema:    "second_schema",
				})
				Expect(err).To(BeNil())
				Expect(respModifySchema.Response.GetCode()).ToNot(Equal(proto.Response_SUCCESS))

				By("create schema when service schema id already exists")
				respModifySchema, err = serviceResource.ModifySchema(getContext(), &pb.ModifySchemaRequest{
					ServiceId: serviceIdPro2,
					SchemaId:  "first_schemaId",
					Schema:    "first_schema",
				})
				Expect(err).To(BeNil())
				Expect(respModifySchema.Response.GetCode()).To(Equal(proto.Response_SUCCESS))
			})
		})

		Context("when create a schema in empty env", func() {
			var (
				serviceIdPro1 string
				serviceIdPro2 string
			)

			It("should be passed", func() {
				respCreateService, err := serviceResource.Create(getContext(), &pb.CreateServiceRequest{
					Service: &pb.MicroService{
						AppId:       "create_schema_empty",
						ServiceName: "create_schema_service",
						Version:     "1.0.0",
						Level:       "FRONT",
						Status:      pb.MS_UP,
					},
				})
				Expect(err).To(BeNil())
				Expect(respCreateService.Response.GetCode()).To(Equal(proto.Response_SUCCESS))
				serviceIdPro1 = respCreateService.ServiceId

				respCreateService, err = serviceResource.Create(getContext(), &pb.CreateServiceRequest{
					Service: &pb.MicroService{
						AppId:       "create_schema_empty",
						ServiceName: "create_schema_service",
						Version:     "1.0.1",
						Level:       "FRONT",
						Schemas: []string{
							"first_schemaId",
							"second_schemaId",
						},
						Status: pb.MS_UP,
					},
				})
				Expect(err).To(BeNil())
				Expect(respCreateService.Response.GetCode()).To(Equal(proto.Response_SUCCESS))
				serviceIdPro2 = respCreateService.ServiceId
			})

			It("should be passed", func() {
				By("create schema when service schema id set is empty")
				respModifySchema, err := serviceResource.ModifySchema(getContext(), &pb.ModifySchemaRequest{
					ServiceId: serviceIdPro1,
					SchemaId:  "first_schemaId",
					Schema:    "first_schema",
				})
				Expect(err).To(BeNil())
				Expect(respModifySchema.Response.GetCode()).To(Equal(proto.Response_SUCCESS))

				By("modify schema summary is empty")
				respModifySchema, err = serviceResource.ModifySchema(getContext(), &pb.ModifySchemaRequest{
					ServiceId: serviceIdPro1,
					SchemaId:  "first_schemaId",
					Schema:    "first_schema_change",
					Summary:   "first0summary1change",
				})
				Expect(err).To(BeNil())
				Expect(respModifySchema.Response.GetCode()).To(Equal(proto.Response_SUCCESS))

				By("modify schema summary already exists")
				respModifySchema, err = serviceResource.ModifySchema(getContext(), &pb.ModifySchemaRequest{
					ServiceId: serviceIdPro1,
					SchemaId:  "first_schemaId",
					Schema:    "first_schema",
					Summary:   "first0summary",
				})
				Expect(err).To(BeNil())
				Expect(respModifySchema.Response.GetCode()).ToNot(Equal(proto.Response_SUCCESS))

				By("add schema")
				respModifySchema, err = serviceResource.ModifySchema(getContext(), &pb.ModifySchemaRequest{
					ServiceId: serviceIdPro1,
					SchemaId:  "second_schemaId",
					Schema:    "second_schema",
				})
				Expect(err).To(BeNil())
				Expect(respModifySchema.Response.GetCode()).ToNot(Equal(proto.Response_SUCCESS))

				By("create schema when service schema id already exists")
				respModifySchema, err = serviceResource.ModifySchema(getContext(), &pb.ModifySchemaRequest{
					ServiceId: serviceIdPro2,
					SchemaId:  "first_schemaId",
					Schema:    "first_schema",
				})
				Expect(err).To(BeNil())
				Expect(respModifySchema.Response.GetCode()).To(Equal(proto.Response_SUCCESS))
			})
		})

		Context("when add a schemaId in prod env, schema editable is set in the meanwhile", func() {
			var (
				serviceIdPro1 string
			)

			It("create service, should pass", func() {
				respCreateService, err := serviceResource.Create(getContext(), &pb.CreateServiceRequest{
					Service: &pb.MicroService{
						AppId:       "add_a_schemaId_prod_schema_lock",
						ServiceName: "add_a_schemaId_prod_schema_lock_service",
						Version:     "1.0.0",
						Level:       "FRONT",
						Status:      pb.MS_UP,
						Environment: pb.ENV_PROD,
					},
				})
				Expect(err).To(BeNil())
				Expect(respCreateService.Response.GetCode()).To(Equal(proto.Response_SUCCESS))
				serviceIdPro1 = respCreateService.ServiceId
			})

			It("add a schema with new schemaId, should pass", func() {
				By("add schemas")
				schemas := []*pb.Schema{
					{
						SchemaId: "first_schemaId",
						Schema:   "first_schema",
						Summary:  "first0summary",
					},
				}
				respModifySchemas, err := serviceResource.ModifySchemas(getContext(), &pb.ModifySchemasRequest{
					ServiceId: serviceIdPro1,
					Schemas:   schemas,
				})
				Expect(err).To(BeNil())
				Expect(respModifySchemas.Response.GetCode()).To(Equal(proto.Response_SUCCESS))

				respGetOne, err := serviceResource.GetOne(getContext(), &pb.GetServiceRequest{
					ServiceId: serviceIdPro1,
				})
				Expect(err).To(BeNil())
				Expect(respGetOne.Response.GetCode()).To(Equal(proto.Response_SUCCESS))
				Expect(respGetOne.Service.Schemas).To(Equal([]string{"first_schemaId"}))

				schemas = []*pb.Schema{
					{
						SchemaId: "second_schemaId",
						Schema:   "second_schema",
						Summary:  "second0summary",
					},
				}
				By("schema edit not allowed, add a schema with new schemaId should fail")
				localServiceResource := service.NewMicroServiceService(false, instanceResource)
				respModifySchemas, err = localServiceResource.ModifySchemas(getContext(), &pb.ModifySchemasRequest{
					ServiceId: serviceIdPro1,
					Schemas:   schemas,
				})
				Expect(err).To(BeNil())
				Expect(respModifySchemas.Response.GetCode()).To(Equal(scerr.ErrUndefinedSchemaID))

				By("schema edit allowed, add a schema with new schemaId should success")

				localServiceResource = service.NewMicroServiceService(true, instanceResource)
				respModifySchemas, err = localServiceResource.ModifySchemas(getContext(), &pb.ModifySchemasRequest{
					ServiceId: serviceIdPro1,
					Schemas:   schemas,
				})
				Expect(err).To(BeNil())
				Expect(respModifySchemas.Response.GetCode()).To(Equal(proto.Response_SUCCESS))
			})
		})

		Context("when modify a schema in prod env, schema editable is set in the meanwhile", func() {
			var (
				serviceIdPro1 string
			)

			It("create service, should pass", func() {
				respCreateService, err := serviceResource.Create(getContext(), &pb.CreateServiceRequest{
					Service: &pb.MicroService{
						AppId:       "modify_a_schema_prod_schema_lock",
						ServiceName: "modify_a_schema_prod_schema_lock_service",
						Version:     "1.0.0",
						Level:       "FRONT",
						Status:      pb.MS_UP,
						Environment: pb.ENV_PROD,
					},
				})
				Expect(err).To(BeNil())
				Expect(respCreateService.Response.GetCode()).To(Equal(proto.Response_SUCCESS))
				serviceIdPro1 = respCreateService.ServiceId
			})

			It("add schemas, should pass", func() {
				By("add schemas")
				schemas := []*pb.Schema{
					{
						SchemaId: "first_schemaId",
						Schema:   "first_schema",
						Summary:  "first0summary",
					},
				}
				respModifySchemas, err := serviceResource.ModifySchemas(getContext(), &pb.ModifySchemasRequest{
					ServiceId: serviceIdPro1,
					Schemas:   schemas,
				})
				Expect(err).To(BeNil())
				Expect(respModifySchemas.Response.GetCode()).To(Equal(proto.Response_SUCCESS))

				respGetOne, err := serviceResource.GetOne(getContext(), &pb.GetServiceRequest{
					ServiceId: serviceIdPro1,
				})
				Expect(err).To(BeNil())
				Expect(respGetOne.Response.GetCode()).To(Equal(proto.Response_SUCCESS))
				Expect(respGetOne.Service.Schemas).To(Equal([]string{"first_schemaId"}))

				By("schema edit not allowed, modify a schema should fail")
				localServiceResource := service.NewMicroServiceService(false, instanceResource)
				respModifySchema, err := localServiceResource.ModifySchema(getContext(), &pb.ModifySchemaRequest{
					ServiceId: serviceIdPro1,
					SchemaId:  schemas[0].SchemaId,
					Summary:   schemas[0].Summary,
					Schema:    schemas[0].SchemaId,
				})
				Expect(err).To(BeNil())
				Expect(respModifySchema.Response.GetCode()).To(Equal(scerr.ErrModifySchemaNotAllow))

				By("schema edit allowed, modify a schema should success")
				localServiceResource = service.NewMicroServiceService(true, instanceResource)
				respModifySchema, err = localServiceResource.ModifySchema(getContext(), &pb.ModifySchemaRequest{
					ServiceId: serviceIdPro1,
					SchemaId:  schemas[0].SchemaId,
					Summary:   schemas[0].Summary,
					Schema:    schemas[0].SchemaId,
				})
				Expect(err).To(BeNil())
				Expect(respModifySchema.Response.GetCode()).To(Equal(proto.Response_SUCCESS))
			})
		})
	})

	Describe("execute 'exist' operation", func() {
		var (
			serviceId string
		)

		It("should be passed", func() {
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
			Expect(err).To(BeNil())
			Expect(respCreateService.Response.GetCode()).To(Equal(proto.Response_SUCCESS))
			serviceId = respCreateService.ServiceId

			resp, err := serviceResource.ModifySchema(getContext(), &pb.ModifySchemaRequest{
				ServiceId: serviceId,
				SchemaId:  "com.huawei.test",
				Schema:    "query schema",
				Summary:   "summary",
			})
			Expect(err).To(BeNil())
			Expect(resp.Response.GetCode()).To(Equal(proto.Response_SUCCESS))

			resp, err = serviceResource.ModifySchema(getContext(), &pb.ModifySchemaRequest{
				ServiceId: serviceId,
				SchemaId:  "com.huawei.test.no.summary",
				Schema:    "query schema",
			})
			Expect(err).To(BeNil())
			Expect(resp.Response.GetCode()).To(Equal(proto.Response_SUCCESS))
		})

		Context("when request is invalid", func() {
			It("should be failed", func() {
				By("service id is empty")
				resp, err := serviceResource.Exist(getContext(), &pb.GetExistenceRequest{
					Type:      "schema",
					ServiceId: "",
					SchemaId:  "com.huawei.test",
				})
				Expect(err).To(BeNil())
				Expect(resp.Response.GetCode()).ToNot(Equal(proto.Response_SUCCESS))

				By("schema id does not exist")
				resp, err = serviceResource.Exist(getContext(), &pb.GetExistenceRequest{
					Type:      "schema",
					ServiceId: serviceId,
					SchemaId:  "noneschema",
				})
				Expect(err).To(BeNil())
				Expect(resp.Response.GetCode()).ToNot(Equal(proto.Response_SUCCESS))

				By("schema id is invalid")
				resp, err = serviceResource.Exist(getContext(), &pb.GetExistenceRequest{
					Type:      "schema",
					ServiceId: serviceId,
					SchemaId:  TOO_LONG_SCHEMAID,
				})
				Expect(err).To(BeNil())
				Expect(resp.Response.GetCode()).ToNot(Equal(proto.Response_SUCCESS))
			})
		})

		Context("when request is valid", func() {
			It("should be passed", func() {
				resp, err := serviceResource.Exist(getContext(), &pb.GetExistenceRequest{
					Type:      "schema",
					ServiceId: serviceId,
					SchemaId:  "com.huawei.test",
				})
				Expect(err).To(BeNil())
				Expect(resp.Response.GetCode()).To(Equal(proto.Response_SUCCESS))
				Expect(resp.Summary).To(Equal("summary"))

				resp, err = serviceResource.Exist(getContext(), &pb.GetExistenceRequest{
					Type:        "schema",
					ServiceId:   serviceId,
					SchemaId:    "com.huawei.test",
					AppId:       "()",
					ServiceName: "",
					Version:     "()",
				})
				Expect(err).To(BeNil())
				Expect(resp.Response.GetCode()).To(Equal(proto.Response_SUCCESS))

				resp, err = serviceResource.Exist(getContext(), &pb.GetExistenceRequest{
					Type:      "schema",
					ServiceId: serviceId,
					SchemaId:  "com.huawei.test.no.summary",
				})
				Expect(err).To(BeNil())
				Expect(resp.Response.GetCode()).To(Equal(proto.Response_SUCCESS))
				Expect(resp.SchemaId).To(Equal("com.huawei.test.no.summary"))
				Expect(resp.Summary).To(Equal(""))
			})
		})
	})

	Describe("execute 'get' operation", func() {
		var (
			serviceId  string
			serviceId1 string
		)

		var (
			schemaId1     string = "all_schema1"
			schemaId2     string = "all_schema2"
			schemaId3     string = "all_schema3"
			summary       string = "this0is1a2test"
			schemaContent string = "the content is vary large"
		)

		It("should be passed", func() {
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
			Expect(err).To(BeNil())
			Expect(respCreateService.Response.GetCode()).To(Equal(proto.Response_SUCCESS))
			serviceId = respCreateService.ServiceId

			resp, err := serviceResource.ModifySchema(getContext(), &pb.ModifySchemaRequest{
				ServiceId: serviceId,
				SchemaId:  "com.huawei.test",
				Schema:    "get schema",
				Summary:   "schema0summary",
			})
			Expect(err).To(BeNil())
			Expect(resp.Response.GetCode()).To(Equal(proto.Response_SUCCESS))

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
			Expect(err).To(BeNil())
			Expect(respCreateService.Response.GetCode()).To(Equal(proto.Response_SUCCESS))
			serviceId1 = respCreateService.ServiceId

			respPutData, err := serviceResource.ModifySchema(getContext(), &pb.ModifySchemaRequest{
				ServiceId: serviceId1,
				SchemaId:  schemaId2,
				Schema:    schemaContent,
			})
			Expect(err).To(BeNil())
			Expect(respPutData.Response.GetCode()).To(Equal(proto.Response_SUCCESS))

			respPutData, err = serviceResource.ModifySchema(getContext(), &pb.ModifySchemaRequest{
				ServiceId: serviceId1,
				SchemaId:  schemaId3,
				Schema:    schemaContent,
				Summary:   summary,
			})
			Expect(err).To(BeNil())
			Expect(respPutData.Response.GetCode()).To(Equal(proto.Response_SUCCESS))

			respGetAllSchema, err := serviceResource.GetAllSchemaInfo(getContext(), &pb.GetAllSchemaRequest{
				ServiceId:  serviceId1,
				WithSchema: false,
			})
			Expect(err).To(BeNil())
			Expect(respGetAllSchema.Response.GetCode()).To(Equal(proto.Response_SUCCESS))
			schemas := respGetAllSchema.Schemas
			for _, schema := range schemas {
				if schema.SchemaId == schemaId1 {
					Expect(schema.Summary).To(BeEmpty())
					Expect(schema.Schema).To(BeEmpty())
				}
				if schema.SchemaId == schemaId2 {
					Expect(schema.Summary).To(BeEmpty())
					Expect(schema.Schema).To(BeEmpty())
				}
				if schema.SchemaId == schemaId3 {
					Expect(schema.Summary).To(Equal(summary))
					Expect(schema.Schema).To(BeEmpty())
				}
			}

			respGetAllSchema, err = serviceResource.GetAllSchemaInfo(getContext(), &pb.GetAllSchemaRequest{
				ServiceId:  serviceId1,
				WithSchema: true,
			})
			Expect(err).To(BeNil())
			Expect(respGetAllSchema.Response.GetCode()).To(Equal(proto.Response_SUCCESS))
			schemas = respGetAllSchema.Schemas
			for _, schema := range schemas {
				if schema.SchemaId == schemaId1 {
					Expect(schema.Schema).To(BeEmpty())
					Expect(schema.Schema).To(BeEmpty())
				}
				if schema.SchemaId == schemaId2 {
					Expect(schema.Summary).To(BeEmpty())
					Expect(schema.Schema).To(Equal(schemaContent))
				}
				if schema.SchemaId == schemaId3 {
					Expect(schema.Summary).To(Equal(summary))
					Expect(schema.Schema).To(Equal(schemaContent))
				}
			}

		})

		Context("when request is invalid", func() {
			It("should be failed", func() {
				By("service does not exist")
				resp, err := serviceResource.GetSchemaInfo(getContext(), &pb.GetSchemaRequest{
					ServiceId: "noneexistservice",
					SchemaId:  "com.huawei.test",
				})
				Expect(err).To(BeNil())
				Expect(resp.Response.GetCode()).To(Equal(scerr.ErrServiceNotExists))

				respA, err := serviceResource.GetAllSchemaInfo(getContext(), &pb.GetAllSchemaRequest{
					ServiceId: "noneexistservice",
				})
				Expect(err).To(BeNil())
				Expect(respA.Response.GetCode()).To(Equal(scerr.ErrServiceNotExists))

				By("service id is empty")
				resp, err = serviceResource.GetSchemaInfo(getContext(), &pb.GetSchemaRequest{
					ServiceId: "",
					SchemaId:  "com.huawei.test",
				})
				Expect(err).To(BeNil())
				Expect(resp.Response.GetCode()).To(Equal(scerr.ErrInvalidParams))

				respA, err = serviceResource.GetAllSchemaInfo(getContext(), &pb.GetAllSchemaRequest{
					ServiceId: "",
				})
				Expect(err).To(BeNil())
				Expect(respA.Response.GetCode()).To(Equal(scerr.ErrInvalidParams))

				By("service id is invalid")
				resp, err = serviceResource.GetSchemaInfo(getContext(), &pb.GetSchemaRequest{
					ServiceId: TOO_LONG_SERVICEID,
					SchemaId:  "com.huawei.test",
				})
				Expect(err).To(BeNil())
				Expect(resp.Response.GetCode()).To(Equal(scerr.ErrInvalidParams))

				respA, err = serviceResource.GetAllSchemaInfo(getContext(), &pb.GetAllSchemaRequest{
					ServiceId: TOO_LONG_SERVICEID,
				})
				Expect(err).To(BeNil())
				Expect(respA.Response.GetCode()).To(Equal(scerr.ErrInvalidParams))

				By("schema id does not exist")
				resp, err = serviceResource.GetSchemaInfo(getContext(), &pb.GetSchemaRequest{
					ServiceId: serviceId,
					SchemaId:  "nonexistschema",
				})
				Expect(err).To(BeNil())
				Expect(resp.Response.GetCode()).To(Equal(scerr.ErrSchemaNotExists))

				By("schema id is invalid")
				resp, err = serviceResource.GetSchemaInfo(getContext(), &pb.GetSchemaRequest{
					ServiceId: serviceId,
					SchemaId:  "",
				})
				Expect(err).To(BeNil())
				Expect(resp.Response.GetCode()).To(Equal(scerr.ErrInvalidParams))
				resp, err = serviceResource.GetSchemaInfo(getContext(), &pb.GetSchemaRequest{
					ServiceId: serviceId,
					SchemaId:  TOO_LONG_SCHEMAID,
				})
				Expect(err).To(BeNil())
				Expect(resp.Response.GetCode()).To(Equal(scerr.ErrInvalidParams))
				resp, err = serviceResource.GetSchemaInfo(getContext(), &pb.GetSchemaRequest{
					ServiceId: serviceId,
					SchemaId:  invalidSchemaId,
				})
				Expect(err).To(BeNil())
				Expect(resp.Response.GetCode()).To(Equal(scerr.ErrInvalidParams))

				By("schema content does not exist")
				resp, err = serviceResource.GetSchemaInfo(getContext(), &pb.GetSchemaRequest{
					ServiceId: serviceId,
					SchemaId:  "non-schema-content",
				})
				Expect(err).To(BeNil())
				Expect(resp.Response.GetCode()).To(Equal(scerr.ErrSchemaNotExists))
			})
		})

		Context("when request is valid", func() {
			It("should be passed", func() {
				resp, err := serviceResource.GetSchemaInfo(getContext(), &pb.GetSchemaRequest{
					ServiceId: serviceId,
					SchemaId:  "com.huawei.test",
				})
				Expect(err).To(BeNil())
				Expect(resp.Response.GetCode()).To(Equal(proto.Response_SUCCESS))
				Expect(resp.Schema).To(Equal("get schema"))
				Expect(resp.SchemaSummary).To(Equal("schema0summary"))
			})
		})
	})

	Describe("execute 'delete' operation", func() {
		var (
			serviceId string
		)

		It("should be passed", func() {
			respCreateService, err := serviceResource.Create(getContext(), &pb.CreateServiceRequest{
				Service: &pb.MicroService{
					AppId:       "delete_schema_group",
					ServiceName: "delete_schema_service",
					Version:     "1.0.0",
					Level:       "FRONT",
					Status:      pb.MS_UP,
				},
			})
			Expect(err).To(BeNil())
			Expect(respCreateService.Response.GetCode()).To(Equal(proto.Response_SUCCESS))
			serviceId = respCreateService.ServiceId

			resp, err := serviceResource.ModifySchema(getContext(), &pb.ModifySchemaRequest{
				ServiceId: serviceId,
				SchemaId:  "com.huawei.test",
				Schema:    "delete schema",
				Summary:   "summary",
			})
			Expect(err).To(BeNil())
			Expect(resp.Response.GetCode()).To(Equal(proto.Response_SUCCESS))
		})

		Context("when request is invalid", func() {
			It("should be failed", func() {
				By("schema id does not exist")
				resp, err := serviceResource.DeleteSchema(getContext(), &pb.DeleteSchemaRequest{
					ServiceId: serviceId,
					SchemaId:  "noneschema",
				})
				Expect(err).To(BeNil())
				Expect(resp.Response.GetCode()).ToNot(Equal(proto.Response_SUCCESS))

				By("service id is empty")
				resp, err = serviceResource.DeleteSchema(getContext(), &pb.DeleteSchemaRequest{
					ServiceId: "",
					SchemaId:  "com.huawei.test",
				})
				Expect(err).To(BeNil())
				Expect(resp.Response.GetCode()).ToNot(Equal(proto.Response_SUCCESS))

				By("service id does not exist")
				resp, err = serviceResource.DeleteSchema(getContext(), &pb.DeleteSchemaRequest{
					ServiceId: "noexistservice",
					SchemaId:  "com.huawei.test",
				})
				Expect(err).To(BeNil())
				Expect(resp.Response.GetCode()).ToNot(Equal(proto.Response_SUCCESS))

				By("schema id is invalid")
				resp, err = serviceResource.DeleteSchema(getContext(), &pb.DeleteSchemaRequest{
					ServiceId: serviceId,
					SchemaId:  invalidSchemaId,
				})
				Expect(err).To(BeNil())
				Expect(resp.Response.GetCode()).ToNot(Equal(proto.Response_SUCCESS))
			})
		})

		Context("when request is valid", func() {
			It("should be passed", func() {
				resp, err := serviceResource.DeleteSchema(getContext(), &pb.DeleteSchemaRequest{
					ServiceId: serviceId,
					SchemaId:  "com.huawei.test",
				})
				Expect(err).To(BeNil())
				Expect(resp.Response.GetCode()).To(Equal(proto.Response_SUCCESS))

				respGet, err := serviceResource.GetSchemaInfo(getContext(), &pb.GetSchemaRequest{
					ServiceId: serviceId,
					SchemaId:  "com.huawei.test",
				})
				Expect(err).To(BeNil())
				Expect(respGet.Response.GetCode()).To(Equal(scerr.ErrSchemaNotExists))

				respExist, err := serviceResource.Exist(getContext(), &pb.GetExistenceRequest{
					Type:      "schema",
					ServiceId: serviceId,
					SchemaId:  "com.huawei.test",
				})
				Expect(err).To(BeNil())
				Expect(respExist.Response.GetCode()).To(Equal(scerr.ErrSchemaNotExists))
			})
		})
	})
})
