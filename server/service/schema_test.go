//Copyright 2017 Huawei Technologies Co., Ltd
//
//Licensed under the Apache License, Version 2.0 (the "License");
//you may not use this file except in compliance with the License.
//You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
//Unless required by applicable law or agreed to in writing, software
//distributed under the License is distributed on an "AS IS" BASIS,
//WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//See the License for the specific language governing permissions and
//limitations under the License.
package service_test

import (
	pb "github.com/ServiceComb/service-center/server/core/proto"
	scerr "github.com/ServiceComb/service-center/server/error"
	"github.com/ServiceComb/service-center/server/plugin/infra/quota/buildin"
	"github.com/ServiceComb/service-center/version"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"strconv"
	"strings"
)

const (
	invalidSchemaId = "@"
)

var tooLongSummary = strings.Repeat("x", 513)

var _ = Describe("'Schema' service", func() {
	Describe("execute 'create' operartion", func() {
		var (
			serviceId string
		)

		It("should be passed", func() {
			respCreateService, err := serviceResource.Create(getContext(), &pb.CreateServiceRequest{
				Service: &pb.MicroService{
					AppId:       "create_schema_group",
					ServiceName: "create_schema_service",
					Version:     "1.0.0",
					Level:       "FRONT",
					Status:      pb.MS_UP,
				},
			})
			Expect(err).To(BeNil())
			Expect(respCreateService.GetResponse().Code).To(Equal(pb.Response_SUCCESS))
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
				Expect(resp.GetResponse().Code).ToNot(Equal(pb.Response_SUCCESS))

				By("service does not exist")
				resp, err = serviceResource.ModifySchema(getContext(), &pb.ModifySchemaRequest{
					ServiceId: "notExistService",
					SchemaId:  "com.huawei.test",
					Schema:    "create schema",
				})
				Expect(err).To(BeNil())
				Expect(resp.GetResponse().Code).ToNot(Equal(pb.Response_SUCCESS))

				By("schema id does not exist")
				resp, err = serviceResource.ModifySchema(getContext(), &pb.ModifySchemaRequest{
					ServiceId: serviceId,
					SchemaId:  invalidSchemaId,
					Schema:    "create schema",
				})
				Expect(err).To(BeNil())
				Expect(resp.GetResponse().Code).ToNot(Equal(pb.Response_SUCCESS))

				By("summary is invalid")
				resp, err = serviceResource.ModifySchema(getContext(), &pb.ModifySchemaRequest{
					ServiceId: serviceId,
					SchemaId:  "com.huawei.test",
					Schema:    "create schema",
					Summary:   tooLongSummary,
				})
				Expect(err).To(BeNil())
				Expect(resp.GetResponse().Code).ToNot(Equal(pb.Response_SUCCESS))
			}
			It("should be failed in dev env", f)

			It("should be failed in prod env", func() {
				version.Ver().RunMode = "prod"
				f()
				version.Ver().RunMode = "dev"
			})
		})

		Context("when batch create invalid schemas", func() {
			f := func() {
				By("service does not exist")
				respCreateService, err := serviceResource.ModifySchemas(getContext(), &pb.ModifySchemasRequest{
					ServiceId: "not_exist_serviceId",
				})
				Expect(err).To(BeNil())
				Expect(respCreateService.GetResponse().Code).ToNot(Equal(pb.Response_SUCCESS))

				By("service id is empty")
				respCreateService, err = serviceResource.ModifySchemas(getContext(), &pb.ModifySchemasRequest{
					ServiceId: "",
				})
				Expect(err).To(BeNil())
				Expect(respCreateService.GetResponse().Code).ToNot(Equal(pb.Response_SUCCESS))

				By("schema id is invalid")
				respCreateService, err = serviceResource.ModifySchemas(getContext(), &pb.ModifySchemasRequest{
					ServiceId: serviceId,
					Schemas: []*pb.Schema{
						{
							SchemaId: invalidSchemaId,
						},
					},
				})
				Expect(err).To(BeNil())
				Expect(respCreateService.GetResponse().Code).ToNot(Equal(pb.Response_SUCCESS))

				By("schema is empty")
				respCreateService, err = serviceResource.ModifySchemas(getContext(), &pb.ModifySchemasRequest{
					ServiceId: serviceId,
					Schemas: []*pb.Schema{
						{
							SchemaId: "com.huawei.test",
						},
					},
				})
				Expect(err).To(BeNil())
				Expect(respCreateService.GetResponse().Code).ToNot(Equal(pb.Response_SUCCESS))

				By("summary is empty")
				respCreateService, err = serviceResource.ModifySchemas(getContext(), &pb.ModifySchemasRequest{
					ServiceId: serviceId,
					Schemas: []*pb.Schema{
						{
							SchemaId: "com.huawei.test",
							Schema:   "create schema",
						},
					},
				})
				Expect(err).To(BeNil())
				Expect(respCreateService.GetResponse().Code).ToNot(Equal(pb.Response_SUCCESS))

				By("summery is invalid")
				respCreateService, err = serviceResource.ModifySchemas(getContext(), &pb.ModifySchemasRequest{
					ServiceId: serviceId,
					Schemas: []*pb.Schema{
						{
							SchemaId: "com.huawei.test",
							Schema:   "create schema",
							Summary:  tooLongSummary,
						},
					},
				})
				Expect(err).To(BeNil())
				Expect(respCreateService.GetResponse().Code).ToNot(Equal(pb.Response_SUCCESS))
			}
			It("should be failed in dev env", f)

			It("should be failed in prod env", func() {
				version.Ver().RunMode = "prod"
				f()
				version.Ver().RunMode = "dev"
			})
		})

		Context("when create schema out of gauge", func() {
			size := buildin.SCHEMA_NUM_MAX_FOR_ONESERVICE + 1
			schemaIds := make([]string, 0, size)
			schemas := make([]*pb.Schema, 0, size)
			for i := 0; i < size; i++ {
				s := strconv.Itoa(i)

				schemaIds = append(schemaIds, s)
				schemas = append(schemas, &pb.Schema{
					SchemaId: s,
					Schema:   s,
					Summary:  s,
				})
			}
			f := func() {
				By("create service")
				size := buildin.SCHEMA_NUM_MAX_FOR_ONESERVICE + 1
				schemaIds := make([]string, size)
				for i := 0; i < size; i++ {
					schemaIds = append(schemaIds, strconv.Itoa(i))
				}
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
				Expect(respServiceForSchema.GetResponse().Code).ToNot(Equal(pb.Response_SUCCESS))

				By("batch modify schemas")
				respCreateService, err := serviceResource.ModifySchemas(getContext(), &pb.ModifySchemasRequest{
					ServiceId: serviceId,
					Schemas:   schemas,
				})
				Expect(err).To(BeNil())
				Expect(respCreateService.GetResponse().Code).ToNot(Equal(pb.Response_SUCCESS))

			}
			It("should be failed in dev env", func() {
				f()

				By("modify one schema")
				var err error
				respCreateService := &pb.ModifySchemaResponse{}
				for _, schema := range schemas {
					respCreateService, err = serviceResource.ModifySchema(getContext(), &pb.ModifySchemaRequest{
						ServiceId: serviceId,
						Schema:    schema.Schema,
					})
					Expect(err).To(BeNil())
					if respCreateService.GetResponse().Code != pb.Response_SUCCESS {
						break
					}
				}
				Expect(respCreateService.GetResponse().Code).ToNot(Equal(pb.Response_SUCCESS))
			})

			It("should be failed in prod env", func() {
				version.Ver().RunMode = "prod"
				f()
				version.Ver().RunMode = "dev"
			})
		})

		Context("when batch create schemas in dev env", func() {
			var (
				serviceId1 string
				serviceId2 string
			)

			It("should be passed", func() {
				respCreateService, err := serviceResource.Create(getContext(), &pb.CreateServiceRequest{
					Service: &pb.MicroService{
						AppId:       "create_schemas_dev",
						ServiceName: "create_schemas_service",
						Version:     "1.0.0",
						Level:       "FRONT",
						Status:      pb.MS_UP,
					},
				})
				Expect(err).To(BeNil())
				Expect(respCreateService.GetResponse().Code).To(Equal(pb.Response_SUCCESS))
				serviceId1 = respCreateService.ServiceId

				respCreateService, err = serviceResource.Create(getContext(), &pb.CreateServiceRequest{
					Service: &pb.MicroService{
						AppId:       "create_schemas_dev",
						ServiceName: "create_schemas_service",
						Version:     "1.0.1",
						Level:       "FRONT",
						Schemas: []string{
							"first_schemaId",
						},
						Status: pb.MS_UP,
					},
				})
				Expect(err).To(BeNil())
				Expect(respCreateService.GetResponse().Code).To(Equal(pb.Response_SUCCESS))
				serviceId2 = respCreateService.ServiceId
			})

			It("should be passed", func() {
				By("create schemas when service schema id set is empty")
				schemas := []*pb.Schema{
					{
						SchemaId: "first_schemaId",
						Schema:   "first_schema",
						Summary:  "first_summary",
					},
				}
				respCreateService, err := serviceResource.ModifySchemas(getContext(), &pb.ModifySchemasRequest{
					ServiceId: serviceId1,
					Schemas:   schemas,
				})
				Expect(err).To(BeNil())
				Expect(respCreateService.GetResponse().Code).To(Equal(pb.Response_SUCCESS))

				By("modify schemas when service schema id already exists")
				schemas = []*pb.Schema{}
				respCreateService, err = serviceResource.ModifySchemas(getContext(), &pb.ModifySchemasRequest{
					ServiceId: serviceId2,
					Schemas:   schemas,
				})
				Expect(err).To(BeNil())
				Expect(respCreateService.GetResponse().Code).To(Equal(pb.Response_SUCCESS))

				By("modify schemas")
				schemas = []*pb.Schema{
					{
						SchemaId: "first_schemaId",
						Schema:   "first_schema_change",
						Summary:  "first_summary_change",
					},
				}
				respCreateService, err = serviceResource.ModifySchemas(getContext(), &pb.ModifySchemasRequest{
					ServiceId: serviceId1,
					Schemas:   schemas,
				})
				Expect(err).To(BeNil())
				Expect(respCreateService.GetResponse().Code).To(Equal(pb.Response_SUCCESS))

				By("add schemas")
				schemas = []*pb.Schema{
					{
						SchemaId: "second_schemaId",
						Schema:   "second_schema",
						Summary:  "second_summary",
					},
				}
				respCreateService, err = serviceResource.ModifySchemas(getContext(), &pb.ModifySchemasRequest{
					ServiceId: serviceId1,
					Schemas:   schemas,
				})
				Expect(err).To(BeNil())
				Expect(respCreateService.GetResponse().Code).To(Equal(pb.Response_SUCCESS))

				By("create empty")
				schemas = []*pb.Schema{}
				respCreateService, err = serviceResource.ModifySchemas(getContext(), &pb.ModifySchemasRequest{
					ServiceId: serviceId1,
					Schemas:   schemas,
				})
				Expect(err).To(BeNil())
				Expect(respCreateService.GetResponse().Code).To(Equal(pb.Response_SUCCESS))

			})
		})

		Context("when batch create schemas in prod env", func() {
			var (
				serviceId1 string
				serviceId2 string
			)

			It("should be passed", func() {
				respCreateService, err := serviceResource.Create(getContext(), &pb.CreateServiceRequest{
					Service: &pb.MicroService{
						AppId:       "create_schemas_prod",
						ServiceName: "create_schemas_service",
						Version:     "1.0.0",
						Level:       "FRONT",
						Status:      pb.MS_UP,
					},
				})
				Expect(err).To(BeNil())
				Expect(respCreateService.GetResponse().Code).To(Equal(pb.Response_SUCCESS))
				serviceId1 = respCreateService.ServiceId

				respCreateService, err = serviceResource.Create(getContext(), &pb.CreateServiceRequest{
					Service: &pb.MicroService{
						AppId:       "create_schemas_prod",
						ServiceName: "create_schemas_service",
						Version:     "1.0.1",
						Level:       "FRONT",
						Schemas: []string{
							"first_schemaId",
						},
						Status: pb.MS_UP,
					},
				})
				Expect(err).To(BeNil())
				Expect(respCreateService.GetResponse().Code).To(Equal(pb.Response_SUCCESS))
				serviceId2 = respCreateService.ServiceId
			})

			It("should be passed", func() {
				version.Ver().RunMode = "prod"

				By("add schemas when service schema id set is empty")
				schemas := []*pb.Schema{
					{
						SchemaId: "first_schemaId",
						Schema:   "first_schema",
						Summary:  "first_summary",
					},
				}
				respModifySchemas, err := serviceResource.ModifySchemas(getContext(), &pb.ModifySchemasRequest{
					ServiceId: serviceId1,
					Schemas:   schemas,
				})
				Expect(err).To(BeNil())
				Expect(respModifySchemas.GetResponse().Code).To(Equal(pb.Response_SUCCESS))

				By("modify schemas content already exists")
				respModifySchemas, err = serviceResource.ModifySchemas(getContext(), &pb.ModifySchemasRequest{
					ServiceId: serviceId1,
					Schemas:   schemas,
				})
				Expect(err).To(BeNil())
				Expect(respModifySchemas.GetResponse().Code).ToNot(Equal(pb.Response_SUCCESS))

				By("modify schemas content is empty")
				respModifySchemas, err = serviceResource.ModifySchemas(getContext(), &pb.ModifySchemasRequest{
					ServiceId: serviceId2,
					Schemas:   schemas,
				})
				Expect(err).To(BeNil())
				Expect(respModifySchemas.GetResponse().Code).To(Equal(pb.Response_SUCCESS))

				By("add schemas")
				schemas = []*pb.Schema{
					{
						SchemaId: "second_schemaId",
						Schema:   "second_schema",
						Summary:  "second_summary",
					},
				}
				respModifySchemas, err = serviceResource.ModifySchemas(getContext(), &pb.ModifySchemasRequest{
					ServiceId: serviceId1,
					Schemas:   schemas,
				})
				Expect(err).To(BeNil())
				Expect(respModifySchemas.GetResponse().Code).ToNot(Equal(pb.Response_SUCCESS))

				respModifySchemas, err = serviceResource.ModifySchemas(getContext(), &pb.ModifySchemasRequest{
					ServiceId: serviceId2,
					Schemas:   schemas,
				})
				Expect(err).To(BeNil())
				Expect(respModifySchemas.GetResponse().Code).ToNot(Equal(pb.Response_SUCCESS))

				version.Ver().RunMode = "dev"
			})
		})

		Context("when create a schema in dev env", func() {
			var (
				serviceId1 string
				serviceId2 string
			)

			It("should be passed", func() {
				respCreateService, err := serviceResource.Create(getContext(), &pb.CreateServiceRequest{
					Service: &pb.MicroService{
						AppId:       "create_schema_dev",
						ServiceName: "create_schema_service",
						Version:     "1.0.0",
						Level:       "FRONT",
						Status:      pb.MS_UP,
					},
				})
				Expect(err).To(BeNil())
				Expect(respCreateService.GetResponse().Code).To(Equal(pb.Response_SUCCESS))
				serviceId1 = respCreateService.ServiceId

				respCreateService, err = serviceResource.Create(getContext(), &pb.CreateServiceRequest{
					Service: &pb.MicroService{
						AppId:       "create_schema_dev",
						ServiceName: "create_schema_service",
						Version:     "1.0.1",
						Level:       "FRONT",
						Schemas: []string{
							"first_schemaId",
						},
						Status: pb.MS_UP,
					},
				})
				Expect(err).To(BeNil())
				Expect(respCreateService.GetResponse().Code).To(Equal(pb.Response_SUCCESS))
				serviceId2 = respCreateService.ServiceId
			})

			It("should be passed", func() {
				By("create schema when service schema id set is empty")
				respModifySchema, err := serviceResource.ModifySchema(getContext(), &pb.ModifySchemaRequest{
					ServiceId: serviceId1,
					SchemaId:  "first_schemaId",
					Schema:    "first_schema",
				})
				Expect(err).To(BeNil())
				Expect(respModifySchema.GetResponse().Code).To(Equal(pb.Response_SUCCESS))

				By("create schema when service schema id already exists")
				respModifySchema, err = serviceResource.ModifySchema(getContext(), &pb.ModifySchemaRequest{
					ServiceId: serviceId2,
					SchemaId:  "first_schemaId",
					Schema:    "first_schema",
				})
				Expect(err).To(BeNil())
				Expect(respModifySchema.GetResponse().Code).To(Equal(pb.Response_SUCCESS))

				By("modify schema summary is empty")
				respModifySchema, err = serviceResource.ModifySchema(getContext(), &pb.ModifySchemaRequest{
					ServiceId: serviceId1,
					SchemaId:  "first_schemaId",
					Schema:    "first_schema_change",
					Summary:   "first_summary_change",
				})
				Expect(err).To(BeNil())
				Expect(respModifySchema.GetResponse().Code).To(Equal(pb.Response_SUCCESS))

				By("modify schema summary already exists")
				respModifySchema, err = serviceResource.ModifySchema(getContext(), &pb.ModifySchemaRequest{
					ServiceId: serviceId1,
					SchemaId:  "first_schemaId",
					Schema:    "first_schema",
					Summary:   "first_summary",
				})
				Expect(err).To(BeNil())
				Expect(respModifySchema.GetResponse().Code).To(Equal(pb.Response_SUCCESS))

				By("add schema")
				respModifySchema, err = serviceResource.ModifySchema(getContext(), &pb.ModifySchemaRequest{
					ServiceId: serviceId1,
					SchemaId:  "second_schemaId",
					Schema:    "second_schema",
				})
				Expect(err).To(BeNil())
				Expect(respModifySchema.GetResponse().Code).To(Equal(pb.Response_SUCCESS))
			})
		})

		Context("when create a schemas in prod env", func() {
			var (
				serviceId1 string
				serviceId2 string
			)

			It("should be passed", func() {
				respCreateService, err := serviceResource.Create(getContext(), &pb.CreateServiceRequest{
					Service: &pb.MicroService{
						AppId:       "create_schema_prod",
						ServiceName: "create_schema_service",
						Version:     "1.0.0",
						Level:       "FRONT",
						Status:      pb.MS_UP,
					},
				})
				Expect(err).To(BeNil())
				Expect(respCreateService.GetResponse().Code).To(Equal(pb.Response_SUCCESS))
				serviceId1 = respCreateService.ServiceId

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
						Status: pb.MS_UP,
					},
				})
				Expect(err).To(BeNil())
				Expect(respCreateService.GetResponse().Code).To(Equal(pb.Response_SUCCESS))
				serviceId2 = respCreateService.ServiceId

				respModifySchema, err := serviceResource.ModifySchema(getContext(), &pb.ModifySchemaRequest{
					ServiceId: serviceId2,
					SchemaId:  "second_schemaId",
					Schema:    "second_schema",
				})
				Expect(err).To(BeNil())
				Expect(respModifySchema.GetResponse().Code).To(Equal(pb.Response_SUCCESS))
			})

			It("should be passed", func() {
				By("create schema when service schema id set is empty")
				respModifySchema, err := serviceResource.ModifySchema(getContext(), &pb.ModifySchemaRequest{
					ServiceId: serviceId1,
					SchemaId:  "first_schemaId",
					Schema:    "first_schema",
				})
				Expect(err).To(BeNil())
				Expect(respModifySchema.GetResponse().Code).To(Equal(pb.Response_SUCCESS))

				By("create schema when service schema id already exists")
				respModifySchema, err = serviceResource.ModifySchema(getContext(), &pb.ModifySchemaRequest{
					ServiceId: serviceId2,
					SchemaId:  "first_schemaId",
					Schema:    "first_schema",
				})
				Expect(err).To(BeNil())
				Expect(respModifySchema.GetResponse().Code).To(Equal(pb.Response_SUCCESS))

				By("modify schema summary is empty")
				respModifySchema, err = serviceResource.ModifySchema(getContext(), &pb.ModifySchemaRequest{
					ServiceId: serviceId1,
					SchemaId:  "first_schemaId",
					Schema:    "first_schema_change",
					Summary:   "first_summary_change",
				})
				Expect(err).To(BeNil())
				Expect(respModifySchema.GetResponse().Code).To(Equal(pb.Response_SUCCESS))

				By("modify schema summary already exists")
				respModifySchema, err = serviceResource.ModifySchema(getContext(), &pb.ModifySchemaRequest{
					ServiceId: serviceId1,
					SchemaId:  "first_schemaId",
					Schema:    "first_schema",
					Summary:   "first_summary",
				})
				Expect(err).To(BeNil())
				Expect(respModifySchema.GetResponse().Code).ToNot(Equal(pb.Response_SUCCESS))

				By("add schema")
				respModifySchema, err = serviceResource.ModifySchema(getContext(), &pb.ModifySchemaRequest{
					ServiceId: serviceId1,
					SchemaId:  "second_schemaId",
					Schema:    "second_schema",
				})
				Expect(err).To(BeNil())
				Expect(respModifySchema.GetResponse().Code).ToNot(Equal(pb.Response_SUCCESS))
			})
		})
	})

	Describe("execute 'exist' operartion", func() {
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
				},
			})
			Expect(err).To(BeNil())
			Expect(respCreateService.GetResponse().Code).To(Equal(pb.Response_SUCCESS))
			serviceId = respCreateService.ServiceId

			resp, err := serviceResource.ModifySchema(getContext(), &pb.ModifySchemaRequest{
				ServiceId: serviceId,
				SchemaId:  "com.huawei.test",
				Schema:    "query schema",
				Summary:   "xxx",
			})
			Expect(err).To(BeNil())
			Expect(resp.GetResponse().Code).To(Equal(pb.Response_SUCCESS))
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
				Expect(resp.GetResponse().Code).ToNot(Equal(pb.Response_SUCCESS))

				By("schema id does not exist")
				resp, err = serviceResource.Exist(getContext(), &pb.GetExistenceRequest{
					Type:      "schema",
					ServiceId: serviceId,
					SchemaId:  "noneschema",
				})
				Expect(err).To(BeNil())
				Expect(resp.GetResponse().Code).ToNot(Equal(pb.Response_SUCCESS))

				By("request is nil")
				resp, err = serviceResource.Exist(getContext(), nil)
				Expect(err).To(BeNil())
				Expect(resp.GetResponse().Code).ToNot(Equal(pb.Response_SUCCESS))
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
				Expect(resp.GetResponse().Code).To(Equal(pb.Response_SUCCESS))
				Expect(resp.Summary).To(Equal("xxx"))

				resp, err = serviceResource.Exist(getContext(), &pb.GetExistenceRequest{
					Type:        "schema",
					ServiceId:   serviceId,
					SchemaId:    "com.huawei.test",
					AppId:       "()",
					ServiceName: "",
					Version:     "()",
				})
				Expect(err).To(BeNil())
				Expect(resp.GetResponse().Code).To(Equal(pb.Response_SUCCESS))
			})
		})
	})

	Describe("execute 'get' operartion", func() {
		var (
			serviceId string
		)

		It("should be passed", func() {
			respCreateService, err := serviceResource.Create(getContext(), &pb.CreateServiceRequest{
				Service: &pb.MicroService{
					AppId:       "get_schema_group",
					ServiceName: "get_schema_service",
					Version:     "1.0.0",
					Level:       "FRONT",
					Status:      pb.MS_UP,
				},
			})
			Expect(err).To(BeNil())
			Expect(respCreateService.GetResponse().Code).To(Equal(pb.Response_SUCCESS))
			serviceId = respCreateService.ServiceId

			resp, err := serviceResource.ModifySchema(getContext(), &pb.ModifySchemaRequest{
				ServiceId: serviceId,
				SchemaId:  "com.huawei.test",
				Schema:    "get schema",
				Summary:   "xxx",
			})
			Expect(err).To(BeNil())
			Expect(resp.GetResponse().Code).To(Equal(pb.Response_SUCCESS))
		})

		Context("when request is invalid", func() {
			It("should be failed", func() {
				By("service does not exist")
				resp, err := serviceResource.GetSchemaInfo(getContext(), &pb.GetSchemaRequest{
					ServiceId: "noneexistservice",
					SchemaId:  "com.huawei.test",
				})
				Expect(err).To(BeNil())
				Expect(resp.GetResponse().Code).ToNot(Equal(pb.Response_SUCCESS))

				By("service id is empty")
				resp, err = serviceResource.GetSchemaInfo(getContext(), &pb.GetSchemaRequest{
					ServiceId: "",
					SchemaId:  "com.huawei.test",
				})
				Expect(err).To(BeNil())
				Expect(resp.GetResponse().Code).ToNot(Equal(pb.Response_SUCCESS))

				By("schema id is empty")
				resp, err = serviceResource.GetSchemaInfo(getContext(), &pb.GetSchemaRequest{
					ServiceId: serviceId,
					SchemaId:  "nonexistschema",
				})
				Expect(err).To(BeNil())
				Expect(resp.GetResponse().Code).ToNot(Equal(pb.Response_SUCCESS))

				By("schema id is invalid")
				resp, err = serviceResource.GetSchemaInfo(getContext(), &pb.GetSchemaRequest{
					ServiceId: serviceId,
					SchemaId:  invalidSchemaId,
				})
				Expect(err).To(BeNil())
				Expect(resp.GetResponse().Code).ToNot(Equal(pb.Response_SUCCESS))

				By("schema content does not exist")
				resp, err = serviceResource.GetSchemaInfo(getContext(), &pb.GetSchemaRequest{
					ServiceId: serviceId,
					SchemaId:  "com.huawei.test",
				})
				Expect(err).To(BeNil())
				Expect(resp.GetResponse().Code).ToNot(Equal(pb.Response_SUCCESS))
			})
		})

		Context("when request is valid", func() {
			It("should be passed", func() {
				resp, err := serviceResource.GetSchemaInfo(getContext(), &pb.GetSchemaRequest{
					ServiceId: serviceId,
					SchemaId:  "com.huawei.test",
				})
				Expect(err).To(BeNil())
				Expect(resp.GetResponse().Code).To(Equal(pb.Response_SUCCESS))
				Expect(resp.Schema).To(Equal("create schema"))
			})
		})
	})

	Describe("execute 'delete' operartion", func() {
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
			Expect(respCreateService.GetResponse().Code).To(Equal(pb.Response_SUCCESS))
			serviceId = respCreateService.ServiceId

			resp, err := serviceResource.ModifySchema(getContext(), &pb.ModifySchemaRequest{
				ServiceId: serviceId,
				SchemaId:  "com.huawei.test",
				Schema:    "delete schema",
				Summary:   "xxx",
			})
			Expect(err).To(BeNil())
			Expect(resp.GetResponse().Code).To(Equal(pb.Response_SUCCESS))
		})

		Context("when request is invalid", func() {
			It("should be failed", func() {
				By("schema id does not exist")
				resp, err := serviceResource.DeleteSchema(getContext(), &pb.DeleteSchemaRequest{
					ServiceId: serviceId,
					SchemaId:  "noneschema",
				})
				Expect(err).To(BeNil())
				Expect(resp.GetResponse().Code).ToNot(Equal(pb.Response_SUCCESS))

				By("service id is empty")
				resp, err = serviceResource.DeleteSchema(getContext(), &pb.DeleteSchemaRequest{
					ServiceId: "",
					SchemaId:  "com.huawei.test",
				})
				Expect(err).To(BeNil())
				Expect(resp.GetResponse().Code).ToNot(Equal(pb.Response_SUCCESS))

				By("service id does not exist")
				resp, err = serviceResource.DeleteSchema(getContext(), &pb.DeleteSchemaRequest{
					ServiceId: "noexistservice",
					SchemaId:  "com.huawei.test",
				})
				Expect(err).To(BeNil())
				Expect(resp.GetResponse().Code).ToNot(Equal(pb.Response_SUCCESS))

				By("schema id is invalid")
				resp, err = serviceResource.DeleteSchema(getContext(), &pb.DeleteSchemaRequest{
					ServiceId: serviceId,
					SchemaId:  invalidSchemaId,
				})
				Expect(err).To(BeNil())
				Expect(resp.GetResponse().Code).ToNot(Equal(pb.Response_SUCCESS))
			})
		})

		Context("when request is valid", func() {
			It("should be passed", func() {
				resp, err := serviceResource.DeleteSchema(getContext(), &pb.DeleteSchemaRequest{
					ServiceId: serviceId,
					SchemaId:  "com.huawei.test",
				})
				Expect(err).To(BeNil())
				Expect(resp.GetResponse().Code).To(Equal(pb.Response_SUCCESS))

				respGet, err := serviceResource.GetSchemaInfo(getContext(), &pb.GetSchemaRequest{
					ServiceId: serviceId,
					SchemaId:  "com.huawei.test",
				})
				Expect(err).To(BeNil())
				Expect(respGet.GetResponse().Code).To(Equal(scerr.ErrSchemaNotExists))

				respExist, err := serviceResource.Exist(getContext(), &pb.GetExistenceRequest{
					Type:      "schema",
					ServiceId: serviceId,
					SchemaId:  "com.huawei.test",
				})
				Expect(err).To(BeNil())
				Expect(respExist.GetResponse().Code).To(Equal(scerr.ErrSchemaNotExists))
			})
		})
	})
})
