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
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	pb "github.com/ServiceComb/service-center/server/core/proto"
)

const (
	invalidSchemaId = "@"
)

var _ = Describe("ServiceController", func() {
	Describe("Schema", func() {
		Context("normal", func() {
			It("创建schema", func() {
				respCreateService, err := serviceResource.Create(getContext(), &pb.CreateServiceRequest{
					Service: &pb.MicroService{
						ServiceName: "service_name",
						AppId:       "service_group",
						Version:     "1.0.0",
						Level:       "FRONT",
						Schemas: []string{
							"com.huawei.test",
						},
						Status: "UP",
					},
				})
				Expect(err).To(BeNil())
				serviceId = respCreateService.ServiceId
				Expect(respCreateService.GetResponse().Code).To(Equal(pb.Response_SUCCESS))

				fmt.Println("UT===========创建schema")
				resp, err := serviceResource.ModifySchema(getContext(), &pb.ModifySchemaRequest{
					ServiceId: serviceId,
					SchemaId:  "com.huawei.test",
					Schema:    "create schema",
				})
				Expect(err).To(BeNil())
				Expect(resp.GetResponse().Code).To(Equal(pb.Response_SUCCESS))
			})

			It("创建schema,参数校验", func() {
				fmt.Println("UT===========创建schema")
				resp, err := serviceResource.ModifySchema(getContext(), &pb.ModifySchemaRequest{
					ServiceId: "",
					SchemaId:  "com.huawei.test",
					Schema:    "create schema",
				})
				Expect(err).To(BeNil())
				Expect(resp.GetResponse().Code).To(Equal(pb.Response_FAIL))

				resp, err = serviceResource.ModifySchema(getContext(), &pb.ModifySchemaRequest{
					ServiceId: serviceId,
					SchemaId:  invalidSchemaId,
					Schema:    "create schema",
				})
				//Expect(err).To(BeNil())
				Expect(resp.GetResponse().Code).To(Equal(pb.Response_FAIL))

				resp, err = serviceResource.ModifySchema(getContext(), &pb.ModifySchemaRequest{
					ServiceId: "notExistService",
					SchemaId:  "com.huawei.test",
					Schema:    "create schema",
				})
				//Expect(err).To(BeNil())
				Expect(resp.GetResponse().Code).To(Equal(pb.Response_FAIL))
			})

			It("schema是否存在,参数校验", func() {
				resp, err := serviceResource.Exist(getContext(), &pb.GetExistenceRequest{
					Type:      "schema",
					ServiceId: "",
					SchemaId:  "com.huawei.test",
				})
				Expect(err).To(BeNil())
				Expect(resp.GetResponse().Code).To(Equal(pb.Response_FAIL))

				resp, err = serviceResource.Exist(getContext(), &pb.GetExistenceRequest{
					Type:      "schema",
					ServiceId: serviceId,
					SchemaId:  "noneschema",
				})
				Expect(err).To(BeNil())
				Expect(resp.GetResponse().Code).To(Equal(pb.Response_FAIL))

				resp, err = serviceResource.Exist(getContext(), nil)
				Expect(err).To(BeNil())
				Expect(resp.GetResponse().Code).To(Equal(pb.Response_FAIL))

				respExist, err := serviceResource.Exist(getContext(), &pb.GetExistenceRequest{
					Type:        "microservice",
					AppId:       "default",
					ServiceName: "",
					Version:     "3.0.0",
				})
				Expect(err).To(BeNil())
				Expect(respExist.GetResponse().Code).To(Equal(pb.Response_FAIL))
			})

			It("schema是否存在", func() {

				resp, err := serviceResource.Exist(getContext(), &pb.GetExistenceRequest{
					Type:      "schema",
					ServiceId: serviceId,
					SchemaId:  "com.huawei.test",
				})
				Expect(err).To(BeNil())
				Expect(resp.GetResponse().Code).To(Equal(pb.Response_SUCCESS))

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

			It("schema是否存在,参数校验", func() {

				resp, err := serviceResource.Exist(getContext(), &pb.GetExistenceRequest{
					Type:      "schema",
					ServiceId: "",
					SchemaId:  "com.huawei.test",
				})
				Expect(err).To(BeNil())
				Expect(resp.GetResponse().Code).To(Equal(pb.Response_FAIL))
			})

			It("查询schema", func() {
				fmt.Println("UT===========查询schema")
				resp, err := serviceResource.GetSchemaInfo(getContext(), &pb.GetSchemaRequest{
					ServiceId: serviceId,
					SchemaId:  "com.huawei.test",
				})
				Expect(err).To(BeNil())
				Expect(resp.GetResponse().Code).To(Equal(pb.Response_SUCCESS))
				Expect(resp.Schema).To(Equal("create schema"))
			})

			It("查询schema，参数校验", func() {
				fmt.Println("UT===========查询schema")
				resp, err := serviceResource.GetSchemaInfo(getContext(), &pb.GetSchemaRequest{
					ServiceId: "noneexistservice",
					SchemaId:  "com.huawei.test",
				})
				Expect(err).To(BeNil())
				Expect(resp.GetResponse().Code).To(Equal(pb.Response_FAIL))

				fmt.Println("UT===========查询schema")
				resp, err = serviceResource.GetSchemaInfo(getContext(), &pb.GetSchemaRequest{
					ServiceId: "",
					SchemaId:  "com.huawei.test",
				})
				Expect(err).To(BeNil())
				Expect(resp.GetResponse().Code).To(Equal(pb.Response_FAIL))

				fmt.Println("UT===========查询schema")
				resp, err = serviceResource.GetSchemaInfo(getContext(), &pb.GetSchemaRequest{
					ServiceId: serviceId,
					SchemaId:  "nonexistschema",
				})
				Expect(err).To(BeNil())
				Expect(resp.GetResponse().Code).To(Equal(pb.Response_FAIL))

				fmt.Println("UT===========查询schema")
				resp, err = serviceResource.GetSchemaInfo(getContext(), &pb.GetSchemaRequest{
					ServiceId: serviceId,
					SchemaId:  invalidSchemaId,
				})
				Expect(err).To(BeNil())
				Expect(resp.GetResponse().Code).To(Equal(pb.Response_FAIL))
			})

			It("修改schema,参数校验", func() {
				fmt.Println("UT===========修改schema,参数校验")
				resp, err := serviceResource.ModifySchema(getContext(), &pb.ModifySchemaRequest{
					ServiceId: serviceId,
					SchemaId:  "noneschma",
					Schema:    "change schema",
				})
				Expect(err).To(BeNil())
				Expect(resp.GetResponse().Code).To(Equal(pb.Response_FAIL))

				resp, err = serviceResource.ModifySchema(getContext(), &pb.ModifySchemaRequest{
					ServiceId: "noneservice",
					SchemaId:  "noneschma",
					Schema:    "change schema",
				})
				Expect(err).To(BeNil())
				Expect(resp.GetResponse().Code).To(Equal(pb.Response_FAIL))
			})

			It("修改schema", func() {
				fmt.Println("UT===========修改schema")
				resp, err := serviceResource.ModifySchema(getContext(), &pb.ModifySchemaRequest{
					ServiceId: serviceId,
					SchemaId:  "com.huawei.test",
					Schema:    "change schema",
				})
				Expect(err).To(BeNil())
				Expect(resp.GetResponse().Code).To(Equal(pb.Response_SUCCESS))
			})
			It("删除schema", func() {
				fmt.Println("UT===========删除schema，")
				resp, err := serviceResource.DeleteSchema(getContext(), &pb.DeleteSchemaRequest{
					ServiceId: serviceId,
					SchemaId:  "noneschema",
				})
				Expect(err).To(BeNil())
				Expect(resp.GetResponse().Code).To(Equal(pb.Response_FAIL))

				fmt.Println("UT===========删除schema")
				resp, err = serviceResource.DeleteSchema(getContext(), &pb.DeleteSchemaRequest{
					ServiceId: serviceId,
					SchemaId:  "com.huawei.test",
				})
				Expect(err).To(BeNil())
				Expect(resp.GetResponse().Code).To(Equal(pb.Response_SUCCESS))

				respDeleteService, err := serviceResource.Delete(getContext(), &pb.DeleteServiceRequest{
					ServiceId: serviceId,
					Force:     true,
				})
				Expect(err).To(BeNil())
				Expect(respDeleteService.GetResponse().Code).To(Equal(pb.Response_SUCCESS))
			})

			It("删除schema,参数校验", func() {
				fmt.Println("UT===========删除schema，无serviceId")
				resp, err := serviceResource.DeleteSchema(getContext(), &pb.DeleteSchemaRequest{
					ServiceId: "",
					SchemaId:  "com.huawei.test",
				})
				Expect(err).To(BeNil())
				Expect(resp.GetResponse().Code).To(Equal(pb.Response_FAIL))

				fmt.Println("UT===========删除schema，")
				resp, err = serviceResource.DeleteSchema(getContext(), &pb.DeleteSchemaRequest{
					ServiceId: "noexistservice",
					SchemaId:  "com.huawei.test",
				})
				Expect(err).To(BeNil())
				Expect(resp.GetResponse().Code).To(Equal(pb.Response_FAIL))

				fmt.Println("UT===========删除schema，")
				resp, err = serviceResource.DeleteSchema(getContext(), &pb.DeleteSchemaRequest{
					ServiceId: "noexistservice",
					SchemaId:  "com.huawei.test",
				})
				Expect(err).To(BeNil())
				Expect(resp.GetResponse().Code).To(Equal(pb.Response_FAIL))

				resp, err = serviceResource.DeleteSchema(getContext(), &pb.DeleteSchemaRequest{
					ServiceId: serviceId,
					SchemaId:  invalidSchemaId,
				})
				Expect(err).To(BeNil())
				Expect(resp.GetResponse().Code).To(Equal(pb.Response_FAIL))
			})
		})
	})
})
