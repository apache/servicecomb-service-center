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

package mongo_test

import (
	"fmt"
	"strconv"
	"strings"
	"testing"
	"time"

	pb "github.com/go-chassis/cari/discovery"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"

	"github.com/apache/servicecomb-service-center/datasource"
	"github.com/apache/servicecomb-service-center/datasource/mongo"
	"github.com/apache/servicecomb-service-center/datasource/mongo/client"
	"github.com/apache/servicecomb-service-center/datasource/mongo/db"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/util"
	"github.com/apache/servicecomb-service-center/server/config"
	"github.com/apache/servicecomb-service-center/server/core"
	"github.com/apache/servicecomb-service-center/server/plugin/quota"
	"github.com/apache/servicecomb-service-center/server/service"
)

func TestService_Register(t *testing.T) {
	t.Run("Register service after init & install, should pass", func(t *testing.T) {
		size := quota.DefaultSchemaQuota + 1
		paths := make([]*pb.ServicePath, 0, size)
		properties := make(map[string]string, size)
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
		resp, err := datasource.Instance().RegisterService(getContext(), request)
		assert.NotNil(t, resp)
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, resp.Response.GetCode())
	})

	t.Run("register service with same key, should pass", func(t *testing.T) {
		resp, err := datasource.Instance().RegisterService(getContext(), &pb.CreateServiceRequest{
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
		assert.Equal(t, pb.ResponseSuccess, resp.Response.GetCode())
		sameId := resp.ServiceId

		// serviceName: some-relay-ms-service-name
		// alias: sr1-ms-service-name
		resp, err = datasource.Instance().RegisterService(getContext(), &pb.CreateServiceRequest{
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
		assert.Equal(t, pb.ResponseSuccess, resp.Response.GetCode())

		// serviceName: some-relay1-ms-service-name
		// alias: sr-ms-service-name
		resp, err = datasource.Instance().RegisterService(getContext(), &pb.CreateServiceRequest{
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
		assert.Equal(t, pb.ResponseSuccess, resp.Response.GetCode())

		// serviceName: some-relay1-ms-service-name
		// alias: sr-ms-service-name
		// add serviceId field: sameId
		resp, err = datasource.Instance().RegisterService(getContext(), &pb.CreateServiceRequest{
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
		assert.Equal(t, pb.ResponseSuccess, resp.Response.GetCode())

		// serviceName: some-relay-ms-service-name
		// alias: sr1-ms-service-name
		// serviceId: sameId
		resp, err = datasource.Instance().RegisterService(getContext(), &pb.CreateServiceRequest{
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
		assert.Equal(t, pb.ResponseSuccess, resp.Response.GetCode())

		// serviceName: some-relay-ms-service-name
		// alias: sr1-ms-service-name
		// serviceId: custom-id-ms-service-id -- different
		resp, err = datasource.Instance().RegisterService(getContext(), &pb.CreateServiceRequest{
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
		assert.Equal(t, pb.ErrServiceAlreadyExists, resp.Response.GetCode())

		// serviceName: some-relay1-ms-service-name
		// alias: sr-ms-service-name
		// serviceId: custom-id-ms-service-id -- different
		resp, err = datasource.Instance().RegisterService(getContext(), &pb.CreateServiceRequest{
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
		assert.Equal(t, pb.ErrServiceAlreadyExists, resp.Response.GetCode())
	})

	t.Run("same serviceId,different service, can not register again,error is same as the service register twice",
		func(t *testing.T) {
			resp, err := datasource.Instance().RegisterService(getContext(), &pb.CreateServiceRequest{
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
			assert.Equal(t, resp.Response.GetCode(), pb.ResponseSuccess)

			// same serviceId with different service name
			resp, err = datasource.Instance().RegisterService(getContext(), &pb.CreateServiceRequest{
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
			assert.Equal(t, pb.ErrServiceAlreadyExists, resp.Response.GetCode())
		})
}

func TestService_Get(t *testing.T) {
	t.Run("get a exist service, should pass", func(t *testing.T) {
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

		resp, err := datasource.Instance().RegisterService(getContext(), request)
		assert.NoError(t, err)
		assert.Equal(t, resp.Response.GetCode(), pb.ResponseSuccess)

		// search service by serviceID
		queryResp, err := datasource.Instance().GetService(getContext(), &pb.GetServiceRequest{
			ServiceId: "ms-service-query-id",
		})
		assert.NoError(t, err)
		assert.Equal(t, queryResp.Response.GetCode(), pb.ResponseSuccess)
	})
	t.Run("query all services, should pass", func(t *testing.T) {
		resp, err := datasource.Instance().GetServices(getContext(), &pb.GetServicesRequest{})
		assert.NoError(t, err)
		assert.Greater(t, len(resp.Services), 0)
	})

	t.Run("query a service by a not existed serviceId, should not pass", func(t *testing.T) {
		// not exist service
		resp, err := datasource.Instance().GetService(getContext(), &pb.GetServiceRequest{
			ServiceId: "no-exist-service",
		})
		assert.NoError(t, err)
		assert.Equal(t, resp.Response.GetCode(), pb.ErrServiceNotExists)
	})
}

func TestService_Update(t *testing.T) {
	var serviceId string

	t.Run("create service", func(t *testing.T) {
		resp, err := datasource.Instance().RegisterService(getContext(), &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				Alias:       "es_service_ms",
				ServiceName: "update_prop_service_service_ms",
				AppId:       "update_prop_appId_service_ms",
				Version:     "1.0.0",
				Level:       "FRONT",
				Status:      "UP",
			},
		})
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, pb.ResponseSuccess, resp.Response.GetCode())
		assert.NotEqual(t, "", resp.ServiceId)
		serviceId = resp.ServiceId
	})

	t.Run("update properties while properties not nil", func(t *testing.T) {
		request := &pb.UpdateServicePropsRequest{
			ServiceId:  serviceId,
			Properties: make(map[string]string),
		}
		request2 := &pb.UpdateServicePropsRequest{
			ServiceId:  serviceId,
			Properties: make(map[string]string),
		}
		request.Properties["test"] = "1"
		request2.Properties["k"] = "v"
		resp, err := datasource.Instance().UpdateService(getContext(), request)
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, resp.Response.GetCode())

		resp, err = datasource.Instance().UpdateService(getContext(), request2)
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, resp.Response.GetCode())

		respGetService, err := datasource.Instance().GetService(getContext(), &pb.GetServiceRequest{
			ServiceId: serviceId,
		})
		assert.NoError(t, err)
		assert.Equal(t, serviceId, respGetService.Service.ServiceId)
		assert.Equal(t, "", respGetService.Service.Properties["test"])
		assert.Equal(t, "v", respGetService.Service.Properties["k"])
	})

	t.Run("update service that does not exist", func(t *testing.T) {
		r := &pb.UpdateServicePropsRequest{
			ServiceId:  "not_exist_service_service_ms",
			Properties: make(map[string]string),
		}
		resp, err := datasource.Instance().UpdateService(getContext(), r)
		assert.NoError(t, err)
		assert.NotEqual(t, pb.ResponseSuccess, resp.Response.GetCode())
	})

	t.Run("update service by removing the properties", func(t *testing.T) {
		r := &pb.UpdateServicePropsRequest{
			ServiceId:  serviceId,
			Properties: nil,
		}
		resp, err := datasource.Instance().UpdateService(getContext(), r)
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, resp.Response.GetCode())

		r = &pb.UpdateServicePropsRequest{
			ServiceId:  "",
			Properties: map[string]string{},
		}
		resp, err = datasource.Instance().UpdateService(getContext(), r)
		assert.NoError(t, err)
		assert.NotEqual(t, pb.ResponseSuccess, resp.Response.GetCode())
	})
}

func TestService_Detail(t *testing.T) {
	var (
		serviceId string
	)

	t.Run("execute 'get detail' operation", func(t *testing.T) {
		log.Info("should be passed")
		resp, err := datasource.Instance().RegisterService(getContext(), &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				AppId:       "govern_service_group",
				ServiceName: "govern_service_name",
				Version:     "3.0.0",
				Level:       "FRONT",
				Status:      pb.MS_UP,
			},
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, resp.Response.GetCode())
		serviceId = resp.ServiceId

		datasource.Instance().ModifySchema(getContext(), &pb.ModifySchemaRequest{
			ServiceId: serviceId,
			SchemaId:  "schemaId",
			Schema:    "detail",
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, resp.Response.GetCode())

		datasource.Instance().RegisterInstance(getContext(), &pb.RegisterInstanceRequest{
			Instance: &pb.MicroServiceInstance{
				ServiceId: serviceId,
				Endpoints: []string{
					"govern:127.0.0.1:8080",
				},
				HostName: "UT-HOST",
				Status:   pb.MSI_UP,
			},
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, resp.Response.GetCode())

		log.Info("when get invalid service detail, should be failed")
		respD, err := datasource.Instance().GetServiceDetail(getContext(), &pb.GetServiceRequest{
			ServiceId: "",
		})
		assert.NoError(t, err)
		assert.NotEqual(t, pb.ResponseSuccess, respD.Response.GetCode())
		//
		log.Info("when get a service detail, should be passed")
		respGetServiceDetail, err := datasource.Instance().GetServiceDetail(getContext(), &pb.GetServiceRequest{
			ServiceId: serviceId,
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respGetServiceDetail.Response.GetCode())
		//
		respDelete, err := datasource.Instance().UnregisterService(getContext(), &pb.DeleteServiceRequest{
			ServiceId: serviceId,
			Force:     true,
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respDelete.Response.GetCode())
		//
		respGetServiceDetail, err = datasource.Instance().GetServiceDetail(getContext(), &pb.GetServiceRequest{
			ServiceId: serviceId,
		})
		assert.NoError(t, err)
		assert.NotEqual(t, pb.ResponseSuccess, respGetServiceDetail.Response.GetCode())
	})
}

func TestService_Info(t *testing.T) {
	t.Run("get all services", func(t *testing.T) {
		log.Info("should be passed")
		resp, err := datasource.Instance().GetServicesInfo(getContext(), &pb.GetServicesInfoRequest{
			Options: []string{"all"},
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, resp.Response.GetCode())

		resp, err = datasource.Instance().GetServicesInfo(getContext(), &pb.GetServicesInfoRequest{
			Options: []string{""},
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, resp.Response.GetCode())

		resp, err = datasource.Instance().GetServicesInfo(getContext(), &pb.GetServicesInfoRequest{
			Options: []string{"tags", "rules", "instances", "schemas", "statistics"},
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, resp.Response.GetCode())

		resp, err = datasource.Instance().GetServicesInfo(getContext(), &pb.GetServicesInfoRequest{
			Options: []string{"statistics"},
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, resp.Response.GetCode())

		resp, err = datasource.Instance().GetServicesInfo(getContext(), &pb.GetServicesInfoRequest{
			Options:   []string{"instances"},
			CountOnly: true,
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, resp.Response.GetCode())
	})
}

func TestService_Delete(t *testing.T) {
	var (
		serviceContainInstId string
		serviceNoInstId      string
	)

	t.Run("create service & instance", func(t *testing.T) {
		respCreate, err := datasource.Instance().RegisterService(getContext(), &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				ServiceName: "delete_service_with_inst_ms",
				AppId:       "delete_service_ms",
				Version:     "1.0.0",
				Level:       "FRONT",
				Status:      "UP",
			},
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respCreate.Response.GetCode())
		serviceContainInstId = respCreate.ServiceId

		log.Info("attach instance")
		instance := &pb.MicroServiceInstance{
			ServiceId: serviceContainInstId,
			Endpoints: []string{
				"deleteService:127.0.0.1:8080",
			},
			HostName: "delete-host-ms",
			Status:   pb.MSI_UP,
		}
		respCreateIns, err := datasource.Instance().RegisterInstance(getContext(), &pb.RegisterInstanceRequest{
			Instance: instance,
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respCreateIns.Response.GetCode())

		log.Info("create service without instance")
		provider := &pb.MicroService{
			ServiceName: "delete_service_no_inst_ms",
			AppId:       "delete_service_ms",
			Version:     "1.0.0",
			Level:       "FRONT",
			Status:      "UP",
		}
		respCreate, err = datasource.Instance().RegisterService(getContext(), &pb.CreateServiceRequest{
			Service: provider,
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respCreate.Response.GetCode())
		serviceNoInstId = respCreate.ServiceId

		respCreate, err = datasource.Instance().RegisterService(getContext(), &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				ServiceName: "delete_service_consumer_ms",
				AppId:       "delete_service_ms",
				Version:     "1.0.0",
				Level:       "FRONT",
				Status:      "UP",
			},
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respCreate.Response.GetCode())
	})

	t.Run("delete a service which contains instances with no force flag", func(t *testing.T) {
		log.Info("should not pass")
		resp, err := datasource.Instance().UnregisterService(getContext(), &pb.DeleteServiceRequest{
			ServiceId: serviceContainInstId,
			Force:     false,
		})
		assert.NoError(t, err)
		assert.NotEqual(t, pb.ResponseSuccess, resp.Response.GetCode())
	})

	t.Run("delete a service which contains instances with force flag", func(t *testing.T) {
		log.Info("should pass")
		resp, err := datasource.Instance().UnregisterService(getContext(), &pb.DeleteServiceRequest{
			ServiceId: serviceContainInstId,
			Force:     true,
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, resp.Response.GetCode())
	})

	// todo: add delete service depended by consumer after finishing dependency management

	t.Run("delete a service which depended by consumer with force flag", func(t *testing.T) {
		log.Info("should pass")
		resp, err := datasource.Instance().UnregisterService(getContext(), &pb.DeleteServiceRequest{
			ServiceId: serviceNoInstId,
			Force:     true,
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, resp.Response.GetCode())
	})

	t.Run("delete a service with no force flag", func(t *testing.T) {
		log.Info("should not pass")
		resp, err := datasource.Instance().UnregisterService(getContext(), &pb.DeleteServiceRequest{
			ServiceId: serviceNoInstId,
			Force:     false,
		})
		assert.NoError(t, err)
		assert.NotEqual(t, pb.ResponseSuccess, resp.Response.GetCode())
	})
}

func TestApplication_Get(t *testing.T) {
	t.Run("execute 'get apps' operation", func(t *testing.T) {
		log.Info("when request is valid, should be passed")
		resp, err := datasource.Instance().GetApplications(getContext(), &pb.GetAppsRequest{})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, resp.Response.GetCode())

		resp, err = datasource.Instance().GetApplications(getContext(), &pb.GetAppsRequest{
			Environment: pb.ENV_ACCEPT,
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, resp.Response.GetCode())
	})
}

func TestTags_Add(t *testing.T) {
	var (
		serviceId1 string
	)

	// create service
	t.Run("create service", func(t *testing.T) {
		svc1 := &pb.MicroService{
			AppId:       "create_tag_group_ms",
			ServiceName: "create_tag_service_ms",
			Version:     "1.0.0",
			Level:       "FRONT",
			Status:      pb.MS_UP,
		}
		resp, err := datasource.Instance().RegisterService(getContext(), &pb.CreateServiceRequest{
			Service: svc1,
		})
		assert.NoError(t, err)
		assert.NotEqual(t, "", resp.ServiceId)
		serviceId1 = resp.ServiceId
	})

	t.Run("the request is invalid", func(t *testing.T) {
		resp, err := datasource.Instance().AddTags(getContext(), &pb.AddServiceTagsRequest{
			ServiceId: "noServiceTest",
			Tags: map[string]string{
				"a": "test",
			},
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ErrServiceNotExists, resp.Response.GetCode())
	})

	t.Run("the request is valid", func(t *testing.T) {
		defaultQuota := quota.DefaultTagQuota
		tags := make(map[string]string, defaultQuota)
		for i := 0; i < defaultQuota; i++ {
			s := "tag" + strconv.Itoa(i)
			tags[s] = s
		}
		resp, err := datasource.Instance().AddTags(getContext(), &pb.AddServiceTagsRequest{
			ServiceId: serviceId1,
			Tags:      tags,
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, resp.Response.GetCode())
	})
}

func TestTags_Get(t *testing.T) {
	var serviceId string
	t.Run("create service and add tags", func(t *testing.T) {
		svc := &pb.MicroService{
			AppId:       "get_tag_group_ms",
			ServiceName: "get_tag_service_ms",
			Version:     "1.0.0",
			Level:       "FRONT",
			Status:      pb.MS_UP,
		}
		resp, err := datasource.Instance().RegisterService(getContext(), &pb.CreateServiceRequest{
			Service: svc,
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, resp.Response.GetCode())
		serviceId = resp.ServiceId

		respAddTags, err := datasource.Instance().AddTags(getContext(), &pb.AddServiceTagsRequest{
			ServiceId: serviceId,
			Tags: map[string]string{
				"a": "test",
				"b": "b",
			},
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respAddTags.Response.GetCode())
	})

	t.Run("the request is invalid", func(t *testing.T) {
		resp, err := datasource.Instance().GetTags(getContext(), &pb.GetServiceTagsRequest{
			ServiceId: "noThisService",
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ErrServiceNotExists, resp.Response.GetCode())

		resp, err = datasource.Instance().GetTags(getContext(), &pb.GetServiceTagsRequest{
			ServiceId: "",
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ErrServiceNotExists, resp.Response.GetCode())

		resp, err = datasource.Instance().GetTags(getContext(), &pb.GetServiceTagsRequest{
			ServiceId: strings.Repeat("x", 65),
		})
		assert.Equal(t, pb.ErrServiceNotExists, resp.Response.GetCode())
	})

	t.Run("the request is valid", func(t *testing.T) {
		resp, err := datasource.Instance().GetTags(getContext(), &pb.GetServiceTagsRequest{
			ServiceId: serviceId,
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, resp.Response.GetCode())
		assert.Equal(t, "test", resp.Tags["a"])
	})
}

func TestTag_Update(t *testing.T) {
	var serviceId string
	t.Run("add service and add tags", func(t *testing.T) {
		svc := &pb.MicroService{
			AppId:       "update_tag_group_ms",
			ServiceName: "update_tag_service_ms",
			Version:     "1.0.0",
			Level:       "FRONT",
			Status:      pb.MS_UP,
		}
		resp, err := datasource.Instance().RegisterService(getContext(), &pb.CreateServiceRequest{
			Service: svc,
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, resp.Response.GetCode())
		serviceId = resp.ServiceId

		respAddTags, err := datasource.Instance().AddTags(getContext(), &pb.AddServiceTagsRequest{
			ServiceId: serviceId,
			Tags: map[string]string{
				"a": "test",
				"b": "b",
			},
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respAddTags.Response.GetCode())
	})

	t.Run("the request is invalid", func(t *testing.T) {

		resp, err := datasource.Instance().UpdateTag(getContext(), &pb.UpdateServiceTagRequest{
			ServiceId: "noneservice",
			Key:       "a",
			Value:     "update",
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ErrServiceNotExists, resp.Response.GetCode())

		resp, err = datasource.Instance().UpdateTag(getContext(), &pb.UpdateServiceTagRequest{
			ServiceId: serviceId,
			Key:       "notexisttag",
			Value:     "update",
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ErrTagNotExists, resp.Response.GetCode())

		resp, err = datasource.Instance().UpdateTag(getContext(), &pb.UpdateServiceTagRequest{
			ServiceId: serviceId,
			Key:       strings.Repeat("x", 65),
			Value:     "v",
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ErrTagNotExists, resp.Response.GetCode())
	})

	t.Run("the request is valid", func(t *testing.T) {
		resp, err := datasource.Instance().UpdateTag(getContext(), &pb.UpdateServiceTagRequest{
			ServiceId: serviceId,
			Key:       "a",
			Value:     "update",
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, resp.Response.GetCode())
	})
}

func TestTags_Delete(t *testing.T) {
	var serviceId string
	t.Run("create service and add tags", func(t *testing.T) {
		resp, err := datasource.Instance().RegisterService(getContext(), &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				AppId:       "delete_tag_group_ms",
				ServiceName: "delete_tag_service_ms",
				Version:     "1.0.0",
				Level:       "FRONT",
				Status:      pb.MS_UP,
			},
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, resp.Response.GetCode())
		serviceId = resp.ServiceId

		respAddTages, err := datasource.Instance().AddTags(getContext(), &pb.AddServiceTagsRequest{
			ServiceId: serviceId,
			Tags: map[string]string{
				"a": "test",
				"b": "b",
			},
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respAddTages.Response.GetCode())
	})

	t.Run("the request is invalid", func(t *testing.T) {
		resp, err := datasource.Instance().DeleteTags(getContext(), &pb.DeleteServiceTagsRequest{
			ServiceId: "noneservice",
			Keys:      []string{"a", "b"},
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ErrServiceNotExists, resp.Response.GetCode())

		resp, err = datasource.Instance().DeleteTags(getContext(), &pb.DeleteServiceTagsRequest{
			ServiceId: serviceId,
			Keys:      []string{"c"},
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ErrTagNotExists, resp.Response.GetCode())
	})

	t.Run("the request is valid", func(t *testing.T) {
		resp, err := datasource.Instance().DeleteTags(getContext(), &pb.DeleteServiceTagsRequest{
			ServiceId: serviceId,
			Keys:      []string{"a", "b"},
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, resp.Response.GetCode())

		respGetTags, err := datasource.Instance().GetTags(getContext(), &pb.GetServiceTagsRequest{
			ServiceId: serviceId,
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, resp.Response.GetCode())
		assert.Equal(t, "", respGetTags.Tags["a"])
	})
}

func TestRule_Add(t *testing.T) {
	var (
		serviceId1 string
		serviceId2 string
	)

	t.Run("register service and instance", func(t *testing.T) {
		respCreateService, err := datasource.Instance().RegisterService(getContext(), &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				AppId:       "create_rule_group_ms",
				ServiceName: "create_rule_service_ms",
				Version:     "1.0.0",
				Level:       "FRONT",
				Status:      pb.MS_UP,
			},
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respCreateService.Response.GetCode())
		serviceId1 = respCreateService.ServiceId

		respCreateService, err = datasource.Instance().RegisterService(getContext(), &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				AppId:       "create_rule_group_ms",
				ServiceName: "create_rule_service_ms",
				Version:     "1.0.1",
				Level:       "FRONT",
				Status:      pb.MS_UP,
			},
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respCreateService.Response.GetCode())
		serviceId2 = respCreateService.ServiceId
	})

	t.Run("invalid request", func(t *testing.T) {
		respAddRule, err := datasource.Instance().AddRule(getContext(), &pb.AddServiceRulesRequest{
			ServiceId: "not_exist_service_ms",
			Rules: []*pb.AddOrUpdateServiceRule{
				{
					RuleType:    "BLACK",
					Attribute:   "ServiceName",
					Pattern:     "Test*",
					Description: "test white",
				},
			},
		})
		assert.NoError(t, err)
		assert.NotEqual(t, pb.ResponseSuccess, respAddRule)
	})

	t.Run("request is valid", func(t *testing.T) {
		respAddRule, err := datasource.Instance().AddRule(getContext(), &pb.AddServiceRulesRequest{
			ServiceId: serviceId1,
			Rules: []*pb.AddOrUpdateServiceRule{
				{
					RuleType:    "BLACK",
					Attribute:   "ServiceName",
					Pattern:     "Test*",
					Description: "test black",
				},
			},
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respAddRule.Response.GetCode())
		ruleId := respAddRule.RuleIds[0]
		assert.NotEqual(t, "", ruleId)

		respAddRule, err = datasource.Instance().AddRule(getContext(), &pb.AddServiceRulesRequest{
			ServiceId: serviceId1,
			Rules: []*pb.AddOrUpdateServiceRule{
				{
					RuleType:    "BLACK",
					Attribute:   "ServiceName",
					Pattern:     "Test*",
					Description: "test change black",
				},
			},
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respAddRule.Response.GetCode())
		assert.Equal(t, 0, len(respAddRule.RuleIds))

		respAddRule, err = datasource.Instance().AddRule(getContext(), &pb.AddServiceRulesRequest{
			ServiceId: serviceId1,
			Rules: []*pb.AddOrUpdateServiceRule{
				{
					RuleType:    "WHITE",
					Attribute:   "ServiceName",
					Pattern:     "Test*",
					Description: "test white",
				},
			},
		})
		assert.NoError(t, err)
		assert.NotEqual(t, pb.ResponseSuccess, respAddRule.Response.GetCode())
	})

	t.Run("create rule out of quota", func(t *testing.T) {
		size := quota.DefaultRuleQuota + 1
		rules := make([]*pb.AddOrUpdateServiceRule, 0, size)
		for i := 0; i < size; i++ {
			rules = append(rules, &pb.AddOrUpdateServiceRule{
				RuleType:    "BLACK",
				Attribute:   "ServiceName",
				Pattern:     strconv.Itoa(i),
				Description: "test white",
			})
		}

		resp, err := datasource.Instance().AddRule(getContext(), &pb.AddServiceRulesRequest{
			ServiceId: serviceId2,
			Rules:     rules[:size-1],
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, resp.Response.GetCode())

		resp, err = datasource.Instance().AddRule(getContext(), &pb.AddServiceRulesRequest{
			ServiceId: serviceId2,
			Rules:     rules[size-1:],
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ErrNotEnoughQuota, resp.Response.GetCode())
	})
}

func TestRule_Get(t *testing.T) {
	var (
		serviceId string
		ruleId    string
	)

	t.Run("register service and rules", func(t *testing.T) {
		respCreateService, err := datasource.Instance().RegisterService(getContext(), &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				AppId:       "get_rule_group_ms",
				ServiceName: "get_rule_service_ms",
				Version:     "1.0.0",
				Level:       "FRONT",
				Status:      pb.MS_UP,
			},
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respCreateService.Response.GetCode())
		serviceId = respCreateService.ServiceId

		respAddRule, err := datasource.Instance().AddRule(getContext(), &pb.AddServiceRulesRequest{
			ServiceId: serviceId,
			Rules: []*pb.AddOrUpdateServiceRule{
				{
					RuleType:    "BLACK",
					Attribute:   "ServiceName",
					Pattern:     "Test*",
					Description: "test BLACK",
				},
			},
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respAddRule.Response.GetCode())
		ruleId = respAddRule.RuleIds[0]
		assert.NotEqual(t, "", ruleId)
	})

	t.Run("get when request is invalid", func(t *testing.T) {
		respGetRule, err := datasource.Instance().GetRules(getContext(), &pb.GetServiceRulesRequest{
			ServiceId: "not_exist_service_ms",
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ErrServiceNotExists, respGetRule.Response.GetCode())
	})

	t.Run("get when request is valid", func(t *testing.T) {
		respGetRule, err := datasource.Instance().GetRules(getContext(), &pb.GetServiceRulesRequest{
			ServiceId: serviceId,
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respGetRule.Response.GetCode())
		assert.Equal(t, ruleId, respGetRule.Rules[0].RuleId)
	})
}

func TestRule_Delete(t *testing.T) {
	var (
		serviceId string
		ruleId    string
	)
	t.Run("register service and rules", func(t *testing.T) {
		respCreateService, err := datasource.Instance().RegisterService(getContext(), &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				AppId:       "delete_rule_group_ms",
				ServiceName: "delete_rule_service_ms",
				Version:     "1.0.0",
				Level:       "FRONT",
				Status:      pb.MS_UP,
			},
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respCreateService.Response.GetCode())
		serviceId = respCreateService.ServiceId

		respAddRule, err := datasource.Instance().AddRule(getContext(), &pb.AddServiceRulesRequest{
			ServiceId: serviceId,
			Rules: []*pb.AddOrUpdateServiceRule{
				{
					RuleType:    "BLACK",
					Attribute:   "ServiceName",
					Pattern:     "Test*",
					Description: "test BLACK",
				},
			},
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respAddRule.Response.GetCode())
		ruleId = respAddRule.RuleIds[0]
		assert.NotEqual(t, "", ruleId)
	})

	t.Run("delete when request is invalid", func(t *testing.T) {
		resp, err := datasource.Instance().DeleteRule(getContext(), &pb.DeleteServiceRulesRequest{
			ServiceId: "not_exist_service_ms",
			RuleIds:   []string{"1000000"},
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ErrServiceNotExists, resp.Response.GetCode())

		resp, err = datasource.Instance().DeleteRule(getContext(), &pb.DeleteServiceRulesRequest{
			ServiceId: serviceId,
			RuleIds:   []string{"not_exist_rule"},
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ErrRuleNotExists, resp.Response.GetCode())
	})

	t.Run("delete when request is valid", func(t *testing.T) {
		resp, err := datasource.Instance().DeleteRule(getContext(), &pb.DeleteServiceRulesRequest{
			ServiceId: serviceId,
			RuleIds:   []string{ruleId},
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, resp.Response.GetCode())

		respGetRule, err := datasource.Instance().GetRules(getContext(), &pb.GetServiceRulesRequest{
			ServiceId: serviceId,
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, resp.Response.GetCode())
		assert.Equal(t, 0, len(respGetRule.Rules))
	})
}

func TestRule_Update(t *testing.T) {
	var (
		serviceId string
		ruleId    string
	)

	t.Run("create service and rules", func(t *testing.T) {
		respCreateService, err := datasource.Instance().RegisterService(getContext(), &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				AppId:       "update_rule_group_ms",
				ServiceName: "update_rule_service_ms",
				Version:     "1.0.0",
				Level:       "FRONT",
				Status:      pb.MS_UP,
			},
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respCreateService.Response.GetCode())
		serviceId = respCreateService.ServiceId

		respAddRule, err := datasource.Instance().AddRule(getContext(), &pb.AddServiceRulesRequest{
			ServiceId: serviceId,
			Rules: []*pb.AddOrUpdateServiceRule{
				{
					RuleType:    "BLACK",
					Attribute:   "ServiceName",
					Pattern:     "Test*",
					Description: "test BLACK",
				},
			},
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respAddRule.Response.GetCode())
		ruleId = respAddRule.RuleIds[0]
		assert.NotEqual(t, "", ruleId)
	})

	t.Run("update when request is invalid", func(t *testing.T) {
		rule := &pb.AddOrUpdateServiceRule{
			RuleType:    "BLACK",
			Attribute:   "ServiceName",
			Pattern:     "Test*",
			Description: "test BLACK update",
		}
		resp, err := datasource.Instance().UpdateRule(getContext(), &pb.UpdateServiceRuleRequest{
			ServiceId: "not_exist_service_ms",
			RuleId:    ruleId,
			Rule:      rule,
		})
		assert.NoError(t, err)
		assert.NotEqual(t, pb.ResponseSuccess, resp.Response.GetCode())

		resp, err = datasource.Instance().UpdateRule(getContext(), &pb.UpdateServiceRuleRequest{
			ServiceId: serviceId,
			RuleId:    "not_exist_rule_ms",
			Rule:      rule,
		})
		assert.NoError(t, err)
		assert.NotEqual(t, pb.ResponseSuccess, resp.Response.GetCode())

		resp, err = datasource.Instance().UpdateRule(getContext(), &pb.UpdateServiceRuleRequest{
			ServiceId: serviceId,
			RuleId:    ruleId,
			Rule: &pb.AddOrUpdateServiceRule{
				RuleType:    "WHITE",
				Attribute:   "ServiceName",
				Pattern:     "Test*",
				Description: "test white update",
			},
		})
		assert.NoError(t, err)
		assert.NotEqual(t, pb.ResponseSuccess, resp.Response.GetCode())
	})

	t.Run("update when request is valid", func(t *testing.T) {
		resp, err := datasource.Instance().UpdateRule(getContext(), &pb.UpdateServiceRuleRequest{
			ServiceId: serviceId,
			RuleId:    ruleId,
			Rule: &pb.AddOrUpdateServiceRule{
				RuleType:    "BLACK",
				Attribute:   "AppId",
				Pattern:     "Test*",
				Description: "test white update",
			},
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, resp.Response.GetCode())
	})
}

func TestSchema_Create(t *testing.T) {
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
		resp, err := datasource.Instance().RegisterService(getContext(), &pb.CreateServiceRequest{
			Service: svc,
		})
		assert.NoError(t, err)
		assert.NotEqual(t, "", resp.ServiceId)
		assert.Equal(t, pb.ResponseSuccess, resp.Response.GetCode())
		serviceIdDev = resp.ServiceId

		resp, err = datasource.Instance().RegisterService(getContext(), &pb.CreateServiceRequest{
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
		assert.Equal(t, pb.ResponseSuccess, resp.Response.GetCode())
	})

	t.Run("create schemas out of gauge", func(t *testing.T) {
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

		resp, err := datasource.Instance().ModifySchemas(getContext(), &pb.ModifySchemasRequest{
			ServiceId: serviceIdDev,
			Schemas:   schemas,
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ErrNotEnoughQuota, resp.Response.GetCode())

		resp, err = datasource.Instance().ModifySchemas(getContext(), &pb.ModifySchemasRequest{
			ServiceId: serviceIdDev,
			Schemas:   schemas[:quota.DefaultSchemaQuota],
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, resp.Response.GetCode())

		//resp, err = datasource.Instance().ModifySchemas(getContext(), &pb.ModifySchemasRequest{
		//	ServiceId: serviceIdDev,
		//	Schemas:   schemas,
		//})
		//assert.NoError(t, err)
		//assert.Equal(t, pb.ErrNotEnoughQuota, resp.Response.GetCode())
	})

	//
	t.Run("batch create schemas in dev env", func(t *testing.T) {
		var (
			serviceIdDev1 string
			serviceIdDev2 string
		)

		resp, err := datasource.Instance().RegisterService(getContext(), &pb.CreateServiceRequest{
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
		assert.Equal(t, pb.ResponseSuccess, resp.Response.GetCode())
		serviceIdDev1 = resp.ServiceId

		resp, err = datasource.Instance().RegisterService(getContext(), &pb.CreateServiceRequest{
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
		assert.Equal(t, pb.ResponseSuccess, resp.Response.GetCode())
		serviceIdDev2 = resp.ServiceId

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
		respCreateSchema, err := datasource.Instance().ModifySchemas(getContext(), &pb.ModifySchemasRequest{
			ServiceId: serviceIdDev1,
			Schemas:   schemas,
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respCreateSchema.Response.GetCode())
		//
		schemas = []*pb.Schema{
			{
				SchemaId: "first_schemaId_service_ms",
				Schema:   "first_schema_change_service_ms",
				Summary:  "first0summary1change_service_ms",
			},
		}
		respCreateSchema, err = datasource.Instance().ModifySchemas(getContext(), &pb.ModifySchemasRequest{
			ServiceId: serviceIdDev1,
			Schemas:   schemas,
		})

		schemas = []*pb.Schema{
			{
				SchemaId: "second_schemaId_service_ms",
				Schema:   "second_schema_service_ms",
				Summary:  "second0summary_service_ms",
			},
		}
		respCreateSchema, err = datasource.Instance().ModifySchemas(getContext(), &pb.ModifySchemasRequest{
			ServiceId: serviceIdDev1,
			Schemas:   schemas,
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respCreateSchema.Response.GetCode())

		respGetService, err := datasource.Instance().GetService(getContext(), &pb.GetServiceRequest{
			ServiceId: serviceIdDev1,
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respGetService.Response.GetCode())
		assert.Equal(t, []string{"second_schemaId_service_ms"}, respGetService.Service.Schemas)

		schemas = []*pb.Schema{
			{
				SchemaId: "second_schemaId_service_ms",
				Schema:   "second_schema_service_ms",
				Summary:  "second0summary_service_ms",
			},
		}
		respCreateSchema, err = datasource.Instance().ModifySchemas(getContext(), &pb.ModifySchemasRequest{
			ServiceId: serviceIdDev2,
			Schemas:   schemas,
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respCreateSchema.Response.GetCode())

		respGetService, err = datasource.Instance().GetService(getContext(), &pb.GetServiceRequest{
			ServiceId: serviceIdDev2,
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respGetService.Response.GetCode())
		assert.Equal(t, []string{"second_schemaId_service_ms"}, respGetService.Service.Schemas)
	})
	//
	t.Run("batch create schemas in production env", func(t *testing.T) {
		var (
			serviceIdPro1 string
			serviceIdPro2 string
		)

		respCreateService, err := datasource.Instance().RegisterService(getContext(), &pb.CreateServiceRequest{
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
		assert.Equal(t, pb.ResponseSuccess, respCreateService.Response.GetCode())
		serviceIdPro1 = respCreateService.ServiceId

		respCreateService, err = datasource.Instance().RegisterService(getContext(), &pb.CreateServiceRequest{
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
		assert.Equal(t, pb.ResponseSuccess, respCreateService.Response.GetCode())
		serviceIdPro2 = respCreateService.ServiceId

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
		respModifySchemas, err := datasource.Instance().ModifySchemas(getContext(), &pb.ModifySchemasRequest{
			ServiceId: serviceIdPro1,
			Schemas:   schemas,
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respModifySchemas.Response.GetCode())

		respGetService, err := datasource.Instance().GetService(getContext(), &pb.GetServiceRequest{
			ServiceId: serviceIdPro1,
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respGetService.Response.GetCode())
		assert.Equal(t, []string{"first_schemaId_service_ms"}, respGetService.Service.Schemas)

		respModifySchemas, err = datasource.Instance().ModifySchemas(getContext(), &pb.ModifySchemasRequest{
			ServiceId: serviceIdPro1,
			Schemas:   schemas,
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respModifySchemas.Response.GetCode())

		schemas = []*pb.Schema{
			{
				SchemaId: "second_schemaId_service_ms",
				Schema:   "second_schema_service_ms",
				Summary:  "second0summary_service_ms",
			},
		}
		respModifySchemas, err = datasource.Instance().ModifySchemas(getContext(), &pb.ModifySchemasRequest{
			ServiceId: serviceIdPro1,
			Schemas:   schemas,
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ErrUndefinedSchemaID, respModifySchemas.Response.GetCode())
		fmt.Println(serviceIdPro2)

		respModifySchema, err := datasource.Instance().ModifySchema(getContext(), &pb.ModifySchemaRequest{
			ServiceId: serviceIdPro2,
			SchemaId:  "first_schemaId_service_ms",
			Schema:    "first_schema_service_ms",
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respModifySchema.Response.GetCode())

		schemas = []*pb.Schema{
			{
				SchemaId: "first_schemaId_service_ms",
				Schema:   "first_schema_service_ms",
				Summary:  "first0summary_service_ms",
			},
		}
		respModifySchemas, err = datasource.Instance().ModifySchemas(getContext(), &pb.ModifySchemasRequest{
			ServiceId: serviceIdPro2,
			Schemas:   schemas,
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respModifySchemas.Response.GetCode())
		//
		respExist, err := datasource.Instance().ExistSchema(getContext(), &pb.GetExistenceRequest{
			Type:      service.ExistTypeSchema,
			ServiceId: serviceIdPro2,
			SchemaId:  "first_schemaId_service_ms",
		})
		assert.NoError(t, err)
		assert.Equal(t, "first0summary_service_ms", respExist.Summary)
		//
		respModifySchemas, err = datasource.Instance().ModifySchemas(getContext(), &pb.ModifySchemasRequest{
			ServiceId: serviceIdPro2,
			Schemas:   schemas,
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respModifySchemas.Response.GetCode())
	})

	//
	t.Run("create a schema in dev env", func(t *testing.T) {
		var (
			serviceIdDev1 string
			serviceIdDev2 string
		)

		respCreateService, err := datasource.Instance().RegisterService(getContext(), &pb.CreateServiceRequest{
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
		assert.Equal(t, pb.ResponseSuccess, respCreateService.Response.GetCode())
		serviceIdDev1 = respCreateService.ServiceId

		respCreateService, err = datasource.Instance().RegisterService(getContext(), &pb.CreateServiceRequest{
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
		assert.Equal(t, pb.ResponseSuccess, respCreateService.Response.GetCode())
		serviceIdDev2 = respCreateService.ServiceId

		respModifySchema, err := datasource.Instance().ModifySchema(getContext(), &pb.ModifySchemaRequest{
			ServiceId: serviceIdDev1,
			SchemaId:  "first_schemaId_service_ms",
			Schema:    "first_schema_service_ms",
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respModifySchema.Response.GetCode())

		respModifySchema, err = datasource.Instance().ModifySchema(getContext(), &pb.ModifySchemaRequest{
			ServiceId: serviceIdDev2,
			SchemaId:  "first_schemaId_service_ms",
			Schema:    "first_schema_service_ms",
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respModifySchema.Response.GetCode())

		respModifySchema, err = datasource.Instance().ModifySchema(getContext(), &pb.ModifySchemaRequest{
			ServiceId: serviceIdDev1,
			SchemaId:  "first_schemaId_service_ms",
			Schema:    "first_schema_change_service_ms",
			Summary:   "first0summary1change_service_ms",
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respModifySchema.Response.GetCode())

		respModifySchema, err = datasource.Instance().ModifySchema(getContext(), &pb.ModifySchemaRequest{
			ServiceId: serviceIdDev1,
			SchemaId:  "first_schemaId_service_ms",
			Schema:    "first_schema_service_ms",
			Summary:   "first0summary_service_ms",
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respModifySchema.Response.GetCode())

		respModifySchema, err = datasource.Instance().ModifySchema(getContext(), &pb.ModifySchemaRequest{
			ServiceId: serviceIdDev1,
			SchemaId:  "second_schemaId_service_ms",
			Schema:    "second_schema_service_ms",
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respModifySchema.Response.GetCode())

	})
	//
	t.Run("create a schema in production env", func(t *testing.T) {
		var (
			serviceIdPro1 string
			serviceIdPro2 string
		)

		respCreateService, err := datasource.Instance().RegisterService(getContext(), &pb.CreateServiceRequest{
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
		assert.Equal(t, pb.ResponseSuccess, respCreateService.Response.GetCode())
		serviceIdPro1 = respCreateService.ServiceId

		respCreateService, err = datasource.Instance().RegisterService(getContext(), &pb.CreateServiceRequest{
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
		assert.Equal(t, pb.ResponseSuccess, respCreateService.Response.GetCode())
		serviceIdPro2 = respCreateService.ServiceId

		respModifySchema, err := datasource.Instance().ModifySchema(getContext(), &pb.ModifySchemaRequest{
			ServiceId: serviceIdPro1,
			SchemaId:  "first_schemaId_service_ms",
			Schema:    "first_schema_service_ms",
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respModifySchema.Response.GetCode())

		respModifySchema, err = datasource.Instance().ModifySchema(getContext(), &pb.ModifySchemaRequest{
			ServiceId: serviceIdPro1,
			SchemaId:  "first_schemaId_service_ms",
			Schema:    "first_schema_change_service_ms",
			Summary:   "first0summary1change_service_ms",
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respModifySchema.Response.GetCode())

		respModifySchema, err = datasource.Instance().ModifySchema(getContext(), &pb.ModifySchemaRequest{
			ServiceId: serviceIdPro1,
			SchemaId:  "first_schemaId_service_ms",
			Schema:    "first_schema_service_ms",
			Summary:   "first0summary_service_ms",
		})
		assert.NoError(t, err)
		assert.NotEqual(t, pb.ResponseSuccess, respModifySchema.Response.GetCode())

		respModifySchema, err = datasource.Instance().ModifySchema(getContext(), &pb.ModifySchemaRequest{
			ServiceId: serviceIdPro1,
			SchemaId:  "second_schemaId_service_ms",
			Schema:    "second_schema_service_ms",
		})
		assert.NoError(t, err)
		assert.NotEqual(t, pb.ResponseSuccess, respModifySchema.Response.GetCode())

		respModifySchema, err = datasource.Instance().ModifySchema(getContext(), &pb.ModifySchemaRequest{
			ServiceId: serviceIdPro2,
			SchemaId:  "first_schemaId_service_ms",
			Schema:    "first_schema_service_ms",
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respModifySchema.Response.GetCode())

	})
	//
	t.Run("create a schema in empty env", func(t *testing.T) {
		var (
			serviceIdPro1 string
			serviceIdPro2 string
		)

		log.Info("register service")
		respCreateService, err := datasource.Instance().RegisterService(getContext(), &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				AppId:       "create_schema_empty_service_ms",
				ServiceName: "create_schema_service_service_ms",
				Version:     "1.0.0",
				Level:       "FRONT",
				Status:      pb.MS_UP,
			},
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respCreateService.Response.GetCode())
		serviceIdPro1 = respCreateService.ServiceId

		respCreateService, err = datasource.Instance().RegisterService(getContext(), &pb.CreateServiceRequest{
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
		assert.Equal(t, pb.ResponseSuccess, respCreateService.Response.GetCode())
		serviceIdPro2 = respCreateService.ServiceId

		log.Info("create a schema for service whose schemaID is empty")
		respModifySchema, err := datasource.Instance().ModifySchema(getContext(), &pb.ModifySchemaRequest{
			ServiceId: serviceIdPro1,
			SchemaId:  "first_schemaId_service_ms",
			Schema:    "first_schema_service_ms",
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respModifySchema.Response.GetCode())

		log.Info("modify schema for the service whose schema summary is empty")
		respModifySchema, err = datasource.Instance().ModifySchema(getContext(), &pb.ModifySchemaRequest{
			ServiceId: serviceIdPro1,
			SchemaId:  "first_schemaId_service_ms",
			Schema:    "first_schema_change_service_ms",
			Summary:   "first0summary1change_service_ms",
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respModifySchema.Response.GetCode())

		log.Info("modify schema for the service whose schema summary already exist")
		respModifySchema, err = datasource.Instance().ModifySchema(getContext(), &pb.ModifySchemaRequest{
			ServiceId: serviceIdPro1,
			SchemaId:  "first_schemaId_service_ms",
			Schema:    "first_schema_service_ms",
			Summary:   "first0summary_service_ms",
		})
		assert.NoError(t, err)
		assert.NotEqual(t, pb.ResponseSuccess, respModifySchema.Response.GetCode())

		log.Info("add schema")
		respModifySchema, err = datasource.Instance().ModifySchema(getContext(), &pb.ModifySchemaRequest{
			ServiceId: serviceIdPro1,
			SchemaId:  "second_schemaId_service_ms",
			Schema:    "second_schema_service_ms",
		})
		assert.NoError(t, err)
		assert.NotEqual(t, pb.ResponseSuccess, respModifySchema.Response.GetCode())

		log.Info("modify schema for the service whose schemaId already exist")
		respModifySchema, err = datasource.Instance().ModifySchema(getContext(), &pb.ModifySchemaRequest{
			ServiceId: serviceIdPro2,
			SchemaId:  "first_schemaId_service_ms",
			Schema:    "first_schema_service_ms",
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respModifySchema.Response.GetCode())
	})
	//
	t.Run("add a schemaId in production env while schema editable is set", func(t *testing.T) {
		var (
			serviceIdPro1 string
		)
		log.Info("register service")
		respCreateService, err := datasource.Instance().RegisterService(getContext(), &pb.CreateServiceRequest{
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
		assert.Equal(t, pb.ResponseSuccess, respCreateService.Response.GetCode())
		serviceIdPro1 = respCreateService.ServiceId

		log.Info("add a schema with new schemaId, should pass")
		schemas := []*pb.Schema{
			{
				SchemaId: "first_schemaId_ms",
				Schema:   "first_schema_ms",
				Summary:  "first0summary_ms",
			},
		}
		respModifySchemas, err := datasource.Instance().ModifySchemas(getContext(), &pb.ModifySchemasRequest{
			ServiceId: serviceIdPro1,
			Schemas:   schemas,
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respModifySchemas.Response.GetCode())

		respService, err := datasource.Instance().GetService(getContext(), &pb.GetServiceRequest{
			ServiceId: serviceIdPro1,
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respService.Response.GetCode())
		assert.Equal(t, []string{"first_schemaId_ms"}, respService.Service.Schemas)

		schemas = []*pb.Schema{
			{
				SchemaId: "second_schemaId_ms",
				Schema:   "second_schema_ms",
				Summary:  "second0summary_ms",
			},
		}
		log.Info("schema edit not allowed, add a schema with new schemaId should fail")

		localMicroServiceDs := &mongo.DataSource{SchemaEditable: false}
		respModifySchemas, err = localMicroServiceDs.ModifySchemas(getContext(), &pb.ModifySchemasRequest{
			ServiceId: serviceIdPro1,
			Schemas:   schemas,
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ErrUndefinedSchemaID, respModifySchemas.Response.GetCode())

		log.Info("schema edit allowed, add a schema with new schemaId, should pass")
		localMicroServiceDs = &mongo.DataSource{SchemaEditable: true}
		respModifySchemas, err = localMicroServiceDs.ModifySchemas(getContext(), &pb.ModifySchemasRequest{
			ServiceId: serviceIdPro1,
			Schemas:   schemas,
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respModifySchemas.Response.GetCode())
	})
	//
	t.Run("modify a schema in production env while schema editable is set", func(t *testing.T) {
		var (
			serviceIdPro1 string
		)
		log.Info("register service")
		respCreateService, err := datasource.Instance().RegisterService(getContext(), &pb.CreateServiceRequest{
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
		assert.Equal(t, pb.ResponseSuccess, respCreateService.Response.GetCode())
		serviceIdPro1 = respCreateService.ServiceId

		log.Info("add schemas, should pass")
		schemas := []*pb.Schema{
			{
				SchemaId: "first_schemaId_ms",
				Schema:   "first_schema_ms",
				Summary:  "first0summary_ms",
			},
		}
		respModifySchemas, err := datasource.Instance().ModifySchemas(getContext(), &pb.ModifySchemasRequest{
			ServiceId: serviceIdPro1,
			Schemas:   schemas,
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respModifySchemas.Response.GetCode())

		respService, err := datasource.Instance().GetService(getContext(), &pb.GetServiceRequest{
			ServiceId: serviceIdPro1,
		})
		assert.NoError(t, err)
		assert.Equal(t, []string{"first_schemaId_ms"}, respService.Service.Schemas)

		log.Info("schema edit not allowed, modify schema should fail")
		localMicroServiceDs := &mongo.DataSource{SchemaEditable: false}
		respModifySchema, err := localMicroServiceDs.ModifySchema(getContext(), &pb.ModifySchemaRequest{
			ServiceId: serviceIdPro1,
			SchemaId:  schemas[0].SchemaId,
			Summary:   schemas[0].Summary,
			Schema:    schemas[0].SchemaId,
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ErrModifySchemaNotAllow, respModifySchema.Response.GetCode())

		log.Info("schema edit allowed, add a schema with new schemaId, should pass")
		localMicroServiceDs = &mongo.DataSource{SchemaEditable: true}
		respModifySchema, err = localMicroServiceDs.ModifySchema(getContext(), &pb.ModifySchemaRequest{
			ServiceId: serviceIdPro1,
			SchemaId:  schemas[0].SchemaId,
			Summary:   schemas[0].Summary,
			Schema:    schemas[0].SchemaId,
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respModifySchema.Response.GetCode())
	})
}

func TestSchema_Exist(t *testing.T) {
	var (
		serviceId string
	)

	t.Run("register service and add schema", func(t *testing.T) {
		log.Info("register service")
		respCreateService, err := datasource.Instance().RegisterService(getContext(), &pb.CreateServiceRequest{
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
		assert.Equal(t, pb.ResponseSuccess, respCreateService.Response.GetCode())
		serviceId = respCreateService.ServiceId

		log.Info("add schemas, should pass")
		resp, err := datasource.Instance().ModifySchema(getContext(), &pb.ModifySchemaRequest{
			ServiceId: serviceId,
			SchemaId:  "com.huawei.test.ms",
			Schema:    "query schema ms",
			Summary:   "summary_ms",
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, resp.Response.GetCode())

		resp, err = datasource.Instance().ModifySchema(getContext(), &pb.ModifySchemaRequest{
			ServiceId: serviceId,
			SchemaId:  "com.huawei.test.no.summary.ms",
			Schema:    "query schema ms",
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, resp.Response.GetCode())
	})

	t.Run("check exists", func(t *testing.T) {
		log.Info("check schema exist, should pass")
		resp, err := datasource.Instance().ExistSchema(getContext(), &pb.GetExistenceRequest{
			Type:      service.ExistTypeSchema,
			ServiceId: serviceId,
			SchemaId:  "com.huawei.test.ms",
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, resp.Response.GetCode())
		assert.Equal(t, "summary_ms", resp.Summary)

		resp, err = datasource.Instance().ExistSchema(getContext(), &pb.GetExistenceRequest{
			Type:        service.ExistTypeSchema,
			ServiceId:   serviceId,
			SchemaId:    "com.huawei.test.ms",
			AppId:       "()",
			ServiceName: "",
			Version:     "()",
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, resp.Response.GetCode())

		resp, err = datasource.Instance().ExistSchema(getContext(), &pb.GetExistenceRequest{
			Type:      service.ExistTypeSchema,
			ServiceId: serviceId,
			SchemaId:  "com.huawei.test.no.summary.ms",
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, resp.Response.GetCode())
		assert.Equal(t, "com.huawei.test.no.summary.ms", resp.SchemaId)
		assert.Equal(t, "", resp.Summary)
	})
}

func TestSchema_Delete(t *testing.T) {
	var (
		serviceId string
	)

	t.Run("register service and instance", func(t *testing.T) {
		respCreateService, err := datasource.Instance().RegisterService(getContext(), &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				AppId:       "delete_schema_group_ms",
				ServiceName: "delete_schema_service_ms",
				Version:     "1.0.0",
				Level:       "FRONT",
				Status:      pb.MS_UP,
			},
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respCreateService.Response.GetCode())
		serviceId = respCreateService.ServiceId

		resp, err := datasource.Instance().ModifySchema(getContext(), &pb.ModifySchemaRequest{
			ServiceId: serviceId,
			SchemaId:  "com.huawei.test.ms",
			Schema:    "delete schema ms",
			Summary:   "summary_ms",
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, resp.Response.GetCode())
	})

	t.Run("test delete when request is invalid", func(t *testing.T) {
		log.Info("schema id does not exist")
		resp, err := datasource.Instance().DeleteSchema(getContext(), &pb.DeleteSchemaRequest{
			ServiceId: serviceId,
			SchemaId:  "none_exist_schema",
		})
		assert.NoError(t, err)
		assert.NotEqual(t, pb.ResponseSuccess, resp.Response.GetCode())

		log.Info("service id does not exist")
		resp, err = datasource.Instance().DeleteSchema(getContext(), &pb.DeleteSchemaRequest{
			ServiceId: "not_exist_service",
			SchemaId:  "com.huawei.test.ms",
		})
		assert.NoError(t, err)
		assert.NotEqual(t, pb.ResponseSuccess, resp.Response.GetCode())
	})
	//
	t.Run("test delete when request is valid", func(t *testing.T) {
		resp, err := datasource.Instance().DeleteSchema(getContext(), &pb.DeleteSchemaRequest{
			ServiceId: serviceId,
			SchemaId:  "com.huawei.test.ms",
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, resp.Response.GetCode())

		respGet, err := datasource.Instance().GetSchema(getContext(), &pb.GetSchemaRequest{
			ServiceId: serviceId,
			SchemaId:  "com.huawei.test.ms",
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ErrSchemaNotExists, respGet.Response.GetCode())

		respExist, err := datasource.Instance().ExistSchema(getContext(), &pb.GetExistenceRequest{
			Type:      "schema",
			ServiceId: serviceId,
			SchemaId:  "com.huawei.test.ms",
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ErrSchemaNotExists, respExist.Response.GetCode())
	})
}

func TestSchema_Get(t *testing.T) {
	var (
		serviceId  string
		serviceId1 string
	)

	var (
		schemaId1     string = "all_schema1_ms"
		schemaId2     string = "all_schema2_ms"
		schemaId3     string = "all_schema3_ms"
		summary       string = "this0is1a2test3ms"
		schemaContent string = "the content is vary large"
	)

	t.Run("register service and instance", func(t *testing.T) {
		respCreateService, err := datasource.Instance().RegisterService(getContext(), &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				AppId:       "get_schema_group_ms",
				ServiceName: "get_schema_service_ms",
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

		respCreateSchema, err := datasource.Instance().ModifySchema(getContext(), &pb.ModifySchemaRequest{
			ServiceId: serviceId,
			SchemaId:  "com.huawei.test.ms",
			Schema:    "get schema ms",
			Summary:   "schema0summary1ms",
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respCreateSchema.Response.GetCode())

		respCreateService, err = datasource.Instance().RegisterService(getContext(), &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				AppId:       "get_all_schema_ms",
				ServiceName: "get_all_schema_ms",
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
		//
		respPutData, err := datasource.Instance().ModifySchema(getContext(), &pb.ModifySchemaRequest{
			ServiceId: serviceId1,
			SchemaId:  schemaId2,
			Schema:    schemaContent,
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respPutData.Response.GetCode())
		//
		respPutData, err = datasource.Instance().ModifySchema(getContext(), &pb.ModifySchemaRequest{
			ServiceId: serviceId1,
			SchemaId:  schemaId3,
			Schema:    schemaContent,
			Summary:   summary,
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respPutData.Response.GetCode())
		//
		respGetAllSchema, err := datasource.Instance().GetAllSchemas(getContext(), &pb.GetAllSchemaRequest{
			ServiceId:  serviceId1,
			WithSchema: false,
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respGetAllSchema.Response.GetCode())
		schemas := respGetAllSchema.Schemas
		for _, schema := range schemas {
			fmt.Println(schema.SchemaId)
			if schema.SchemaId == schemaId1 && schema.SchemaId == schemaId2 {
				assert.Empty(t, schema.Summary)
				assert.Empty(t, schema.Schema)
			}
			if schema.SchemaId == schemaId3 {
				assert.Equal(t, summary, schema.Summary)
				assert.Empty(t, schema.Schema)
			}
		}
		//
		respGetAllSchema, err = datasource.Instance().GetAllSchemas(getContext(), &pb.GetAllSchemaRequest{
			ServiceId:  serviceId1,
			WithSchema: true,
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respGetAllSchema.Response.GetCode())
		schemas = respGetAllSchema.Schemas
		for _, schema := range schemas {
			switch schema.SchemaId {
			case schemaId1:
				assert.Empty(t, schema.Summary)
				assert.Empty(t, schema.Schema)
			case schemaId2:
				assert.Empty(t, schema.Summary)
				assert.Equal(t, schemaContent, schema.Schema)
			case schemaId3:
				assert.Equal(t, summary, schema.Summary)
				assert.Equal(t, schemaContent, schema.Schema)
			}
		}
	})

	t.Run("test get when request is invalid", func(t *testing.T) {
		log.Info("service does not exist")
		respGetSchema, err := datasource.Instance().GetSchema(getContext(), &pb.GetSchemaRequest{
			ServiceId: "none_exist_service",
			SchemaId:  "com.huawei.test",
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ErrServiceNotExists, respGetSchema.Response.GetCode())

		respGetAllSchemas, err := datasource.Instance().GetAllSchemas(getContext(), &pb.GetAllSchemaRequest{
			ServiceId: "none_exist_service",
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ErrServiceNotExists, respGetAllSchemas.Response.GetCode())

		log.Info("schema id doest not exist")
		respGetSchema, err = datasource.Instance().GetSchema(getContext(), &pb.GetSchemaRequest{
			ServiceId: serviceId,
			SchemaId:  "none_exist_schema",
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ErrSchemaNotExists, respGetSchema.Response.GetCode())
	})

	t.Run("test get when request is valid", func(t *testing.T) {
		resp, err := datasource.Instance().GetSchema(getContext(), &pb.GetSchemaRequest{
			ServiceId: serviceId,
			SchemaId:  "com.huawei.test.ms",
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, resp.Response.GetCode())
		assert.Equal(t, "get schema ms", resp.Schema)
		assert.Equal(t, "schema0summary1ms", resp.SchemaSummary)

	})
}

func TestInstance_Create(t *testing.T) {
	var serviceID string

	t.Run("create service, when request is valid, should pass", func(t *testing.T) {
		respCreateService, err := datasource.Instance().RegisterService(getContext(), &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				ServiceName: "create_instance_service_ms",
				AppId:       "create_instance_ms",
				Version:     "1.0.0",
				Level:       "FRONT",
				Status:      pb.MS_UP,
			},
		})

		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respCreateService.Response.GetCode())
		serviceID = respCreateService.ServiceId
	})

	t.Run("register instance, when request is valid, should pass", func(t *testing.T) {
		respCreateInst, err := datasource.Instance().RegisterInstance(getContext(), &pb.RegisterInstanceRequest{
			Instance: &pb.MicroServiceInstance{
				ServiceId: serviceID,
				Endpoints: []string{
					"createInstance_ms:127.0.0.1:8080",
				},
				HostName: "UT-HOST",
				Status:   pb.MSI_UP,
			},
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respCreateInst.Response.GetCode())
		assert.NotEqual(t, "", respCreateInst.InstanceId)
		// delete instance
		_, err = datasource.Instance().UnregisterInstance(getContext(), &pb.UnregisterInstanceRequest{
			ServiceId:  serviceID,
			InstanceId: respCreateInst.InstanceId,
		})
		assert.NoError(t, err)

		respCreateInst, err = datasource.Instance().RegisterInstance(getContext(), &pb.RegisterInstanceRequest{
			Instance: &pb.MicroServiceInstance{
				InstanceId: "customId_ms",
				ServiceId:  serviceID,
				Endpoints: []string{
					"createInstance_ms:127.0.0.1:8080",
				},
				HostName: "UT-HOST",
				Status:   pb.MSI_UP,
			},
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respCreateInst.Response.GetCode())
		assert.Equal(t, "customId_ms", respCreateInst.InstanceId)
		// delete instance
		_, err = datasource.Instance().UnregisterInstance(getContext(), &pb.UnregisterInstanceRequest{
			ServiceId:  serviceID,
			InstanceId: "customId_ms",
		})
		assert.NoError(t, err)
	})

	t.Run("update the same instance, should pass", func(t *testing.T) {
		instance := &pb.MicroServiceInstance{
			ServiceId: serviceID,
			Endpoints: []string{
				"sameInstance:127.0.0.1:8080",
			},
			HostName: "UT-HOST",
			Status:   pb.MSI_UP,
		}
		resp, err := datasource.Instance().RegisterInstance(getContext(), &pb.RegisterInstanceRequest{
			Instance: instance,
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, resp.Response.GetCode())

		resp, err = datasource.Instance().RegisterInstance(getContext(), &pb.RegisterInstanceRequest{
			Instance: instance,
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, resp.Response.GetCode())
		assert.Equal(t, instance.InstanceId, resp.InstanceId)
		// delete instance
		_, err = datasource.Instance().UnregisterInstance(getContext(), &pb.UnregisterInstanceRequest{
			ServiceId:  serviceID,
			InstanceId: instance.InstanceId,
		})
		assert.NoError(t, err)
	})

	t.Run("delete service data", func(t *testing.T) {
		_, err := datasource.Instance().UnregisterService(getContext(), &pb.DeleteServiceRequest{
			ServiceId: serviceID,
		})
		assert.NoError(t, err)
	})
}

func TestInstance_HeartBeat(t *testing.T) {
	var (
		serviceID   string
		instanceID1 string
		instanceID2 string
	)

	t.Run("register service and instance, should pass", func(t *testing.T) {
		log.Info("register service")
		respCreateService, err := datasource.Instance().RegisterService(getContext(), &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				ServiceName: "heartbeat_service_ms",
				AppId:       "heartbeat_service_ms",
				Version:     "1.0.0",
				Level:       "FRONT",
				Status:      pb.MS_UP,
			},
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respCreateService.Response.GetCode())
		serviceID = respCreateService.ServiceId

		respCreateInstance, err := datasource.Instance().RegisterInstance(getContext(), &pb.RegisterInstanceRequest{
			Instance: &pb.MicroServiceInstance{
				ServiceId: serviceID,
				HostName:  "UT-HOST-MS",
				Endpoints: []string{
					"heartbeat:127.0.0.1:8080",
				},
				Status: pb.MSI_UP,
			},
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respCreateInstance.Response.GetCode())
		instanceID1 = respCreateInstance.InstanceId

		respCreateInstance, err = datasource.Instance().RegisterInstance(getContext(), &pb.RegisterInstanceRequest{
			Instance: &pb.MicroServiceInstance{
				ServiceId: serviceID,
				HostName:  "UT-HOST-MS",
				Endpoints: []string{
					"heartbeat:127.0.0.2:8080",
				},
				Status: pb.MSI_UP,
			},
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respCreateInstance.Response.GetCode())
		instanceID2 = respCreateInstance.InstanceId
	})

	t.Run("instance: update a lease", func(t *testing.T) {
		log.Info("valid instance")
		resp, err := datasource.Instance().Heartbeat(getContext(), &pb.HeartbeatRequest{
			ServiceId:  serviceID,
			InstanceId: instanceID1,
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, resp.Response.GetCode())

		log.Info("serviceId does not exist")
		resp, err = datasource.Instance().Heartbeat(getContext(), &pb.HeartbeatRequest{
			ServiceId:  "100000000000",
			InstanceId: instanceID1,
		})
		assert.NoError(t, err)
		assert.NotEqual(t, pb.ResponseSuccess, resp.Response.GetCode())

		log.Info("instance does not exist")
		resp, err = datasource.Instance().Heartbeat(getContext(), &pb.HeartbeatRequest{
			ServiceId:  serviceID,
			InstanceId: "not-exist-ins",
		})
		assert.NoError(t, err)
		assert.NotEqual(t, pb.ResponseSuccess, resp.Response.GetCode())
	})

	t.Run("batch update lease", func(t *testing.T) {
		log.Info("request contains at least 1 instances")
		resp, err := datasource.Instance().HeartbeatSet(getContext(), &pb.HeartbeatSetRequest{
			Instances: []*pb.HeartbeatSetElement{
				{
					ServiceId:  serviceID,
					InstanceId: instanceID1,
				},
				{
					ServiceId:  serviceID,
					InstanceId: instanceID2,
				},
			},
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, resp.Response.GetCode())
	})

	t.Run("delete related instances and services", func(t *testing.T) {
		// delete instances
		_, err := datasource.Instance().UnregisterInstance(getContext(), &pb.UnregisterInstanceRequest{
			ServiceId:  serviceID,
			InstanceId: instanceID1,
		})
		assert.NoError(t, err)
		_, err = datasource.Instance().UnregisterInstance(getContext(), &pb.UnregisterInstanceRequest{
			ServiceId:  serviceID,
			InstanceId: instanceID2,
		})
		assert.NoError(t, err)

		// delete service
		_, err = datasource.Instance().UnregisterService(getContext(), &pb.DeleteServiceRequest{
			ServiceId: serviceID,
		})
		assert.NoError(t, err)
	})
}

func TestInstance_Update(t *testing.T) {
	var (
		serviceID  string
		instanceID string
	)

	t.Run("register service and instance, should pass", func(t *testing.T) {
		log.Info("register service")
		respCreateService, err := datasource.Instance().RegisterService(getContext(), &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				ServiceName: "update_instance_service_ms",
				AppId:       "update_instance_service_ms",
				Version:     "1.0.0",
				Level:       "FRONT",
				Status:      pb.MS_UP,
			},
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respCreateService.Response.GetCode())
		serviceID = respCreateService.ServiceId

		log.Info("create instance")
		respCreateInstance, err := datasource.Instance().RegisterInstance(getContext(), &pb.RegisterInstanceRequest{
			Instance: &pb.MicroServiceInstance{
				ServiceId: serviceID,
				Endpoints: []string{
					"updateInstance:127.0.0.1:8080",
				},
				HostName:   "UT-HOST-MS",
				Status:     pb.MSI_UP,
				Properties: map[string]string{"nodeIP": "test"},
			},
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respCreateInstance.Response.GetCode())
		instanceID = respCreateInstance.InstanceId
	})

	t.Run("update instance status, should pass", func(t *testing.T) {
		log.Info("update instance status to DOWN")
		respUpdateStatus, err := datasource.Instance().UpdateInstanceStatus(getContext(), &pb.UpdateInstanceStatusRequest{
			ServiceId:  serviceID,
			InstanceId: instanceID,
			Status:     pb.MSI_DOWN,
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respUpdateStatus.Response.GetCode())

		log.Info("update instance status to OUTOFSERVICE")
		respUpdateStatus, err = datasource.Instance().UpdateInstanceStatus(getContext(), &pb.UpdateInstanceStatusRequest{
			ServiceId:  serviceID,
			InstanceId: instanceID,
			Status:     pb.MSI_OUTOFSERVICE,
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respUpdateStatus.Response.GetCode())

		log.Info("update instance status to STARTING")
		respUpdateStatus, err = datasource.Instance().UpdateInstanceStatus(getContext(), &pb.UpdateInstanceStatusRequest{
			ServiceId:  serviceID,
			InstanceId: instanceID,
			Status:     pb.MSI_STARTING,
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respUpdateStatus.Response.GetCode())

		log.Info("update instance status to TESTING")
		respUpdateStatus, err = datasource.Instance().UpdateInstanceStatus(getContext(), &pb.UpdateInstanceStatusRequest{
			ServiceId:  serviceID,
			InstanceId: instanceID,
			Status:     pb.MSI_TESTING,
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respUpdateStatus.Response.GetCode())

		log.Info("update instance status to UP")
		respUpdateStatus, err = datasource.Instance().UpdateInstanceStatus(getContext(), &pb.UpdateInstanceStatusRequest{
			ServiceId:  serviceID,
			InstanceId: instanceID,
			Status:     pb.MSI_UP,
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respUpdateStatus.Response.GetCode())

		log.Info("update instance status with a not exist instance")
		respUpdateStatus, err = datasource.Instance().UpdateInstanceStatus(getContext(), &pb.UpdateInstanceStatusRequest{
			ServiceId:  serviceID,
			InstanceId: "notexistins",
			Status:     pb.MSI_STARTING,
		})
		assert.NoError(t, err)
		assert.NotEqual(t, pb.ResponseSuccess, respUpdateStatus.Response.GetCode())
	})

	t.Run("update instance properties, should pass", func(t *testing.T) {
		log.Info("update one properties")
		respUpdateProperties, err := datasource.Instance().UpdateInstanceProperties(getContext(),
			&pb.UpdateInstancePropsRequest{
				ServiceId:  serviceID,
				InstanceId: instanceID,
				Properties: map[string]string{
					"test": "test",
				},
			})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respUpdateProperties.Response.GetCode())

		log.Info("all max properties updated")
		size := 1000
		properties := make(map[string]string, size)
		for i := 0; i < size; i++ {
			s := strconv.Itoa(i) + strings.Repeat("x", 253)
			properties[s] = s
		}
		respUpdateProperties, err = datasource.Instance().UpdateInstanceProperties(getContext(),
			&pb.UpdateInstancePropsRequest{
				ServiceId:  serviceID,
				InstanceId: instanceID,
				Properties: properties,
			})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respUpdateProperties.Response.GetCode())

		log.Info("update instance that does not exist")
		respUpdateProperties, err = datasource.Instance().UpdateInstanceProperties(getContext(),
			&pb.UpdateInstancePropsRequest{
				ServiceId:  serviceID,
				InstanceId: "not_exist_ins",
				Properties: map[string]string{
					"test": "test",
				},
			})
		assert.NoError(t, err)
		assert.NotEqual(t, pb.ResponseSuccess, respUpdateProperties.Response.GetCode())

		log.Info("remove properties")
		respUpdateProperties, err = datasource.Instance().UpdateInstanceProperties(getContext(),
			&pb.UpdateInstancePropsRequest{
				ServiceId:  serviceID,
				InstanceId: instanceID,
			})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respUpdateProperties.Response.GetCode())

		log.Info("update service that does not exist")
		respUpdateProperties, err = datasource.Instance().UpdateInstanceProperties(getContext(),
			&pb.UpdateInstancePropsRequest{
				ServiceId:  "not_exist_service",
				InstanceId: instanceID,
				Properties: map[string]string{
					"test": "test",
				},
			})
		assert.NoError(t, err)
		assert.NotEqual(t, pb.ResponseSuccess, respUpdateProperties.Response.GetCode())
	})

	t.Run("delete test data", func(t *testing.T) {
		_, err := client.GetMongoClient().Delete(getContext(), db.CollectionService, bson.M{"domain": "default", "project": "default"})
		assert.NoError(t, err)

		_, err = client.GetMongoClient().Delete(getContext(), db.CollectionInstance, bson.M{"domain": "default", "project": "default"})
		assert.NoError(t, err)
	})
}

func TestInstance_Query(t *testing.T) {
	var (
		serviceID1  string
		serviceID2  string
		serviceID3  string
		serviceID4  string
		serviceID5  string
		serviceID6  string
		serviceID7  string
		serviceID8  string
		serviceID9  string
		instanceID1 string
		instanceID2 string
		instanceID4 string
		instanceID5 string
	)

	t.Run("register services and instances for testInstance_query, when request is invalid, should pass", func(t *testing.T) {
		respCreateService, err := datasource.Instance().RegisterService(getContext(), &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				AppId:       "query_instance_ms",
				ServiceName: "query_instance_service_ms",
				Version:     "1.0.0",
				Level:       "FRONT",
				Status:      pb.MS_UP,
			},
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respCreateService.Response.GetCode())
		serviceID1 = respCreateService.ServiceId

		respCreateService, err = datasource.Instance().RegisterService(getContext(), &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				AppId:       "query_instance_ms",
				ServiceName: "query_instance_service_ms",
				Version:     "1.0.5",
				Level:       "FRONT",
				Status:      pb.MS_UP,
			},
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respCreateService.Response.GetCode())
		serviceID2 = respCreateService.ServiceId

		respCreateService, err = datasource.Instance().RegisterService(getContext(), &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				AppId:       "query_instance_diff_app_ms",
				ServiceName: "query_instance_service_ms",
				Version:     "1.0.0",
				Level:       "FRONT",
				Status:      pb.MS_UP,
			},
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respCreateService.Response.GetCode())
		serviceID3 = respCreateService.ServiceId

		respCreateService, err = datasource.Instance().RegisterService(getContext(), &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				Environment: pb.ENV_PROD,
				AppId:       "query_instance_ms",
				ServiceName: "query_instance_diff_env_service_ms",
				Version:     "1.0.0",
				Level:       "FRONT",
				Status:      pb.MS_UP,
			},
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respCreateService.Response.GetCode())
		serviceID4 = respCreateService.ServiceId

		respCreateService, err = datasource.Instance().RegisterService(getContext(), &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				Environment: pb.ENV_PROD,
				AppId:       "default",
				ServiceName: "query_instance_shared_provider_ms",
				Version:     "1.0.0",
				Level:       "FRONT",
				Status:      pb.MS_UP,
				Properties: map[string]string{
					pb.PropAllowCrossApp: "true",
				},
			},
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respCreateService.Response.GetCode())
		serviceID5 = respCreateService.ServiceId

		respCreateService, err = datasource.Instance().RegisterService(
			util.SetDomainProject(util.CloneContext(getContext()), "user", "user"),
			&pb.CreateServiceRequest{
				Service: &pb.MicroService{
					AppId:       "default",
					ServiceName: "query_instance_diff_domain_consumer_ms",
					Version:     "1.0.0",
					Level:       "FRONT",
					Status:      pb.MS_UP,
				},
			})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respCreateService.Response.GetCode())
		serviceID6 = respCreateService.ServiceId

		respCreateService, err = datasource.Instance().RegisterService(getContext(), &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				AppId:       "default",
				ServiceName: "query_instance_shared_consumer_ms",
				Version:     "1.0.0",
				Level:       "FRONT",
				Status:      pb.MS_UP,
			},
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respCreateService.Response.GetCode())
		serviceID7 = respCreateService.ServiceId

		respCreateService, err = datasource.Instance().RegisterService(getContext(), &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				AppId:       "query_instance_ms",
				ServiceName: "query_instance_with_rev_ms",
				Version:     "1.0.0",
				Level:       "FRONT",
				Status:      pb.MS_UP,
			},
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respCreateService.Response.GetCode())
		serviceID8 = respCreateService.ServiceId

		respCreateService, err = datasource.Instance().RegisterService(getContext(), &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				AppId:       "query_instance_ms",
				ServiceName: "batch_query_instance_with_rev_ms",
				Version:     "1.0.0",
				Level:       "FRONT",
				Status:      pb.MS_UP,
			},
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respCreateService.Response.GetCode())
		serviceID9 = respCreateService.ServiceId

		respCreateInstance, err := datasource.Instance().RegisterInstance(getContext(), &pb.RegisterInstanceRequest{
			Instance: &pb.MicroServiceInstance{
				ServiceId: serviceID1,
				HostName:  "UT-HOST-MS",
				Endpoints: []string{
					"find:127.0.0.1:8080",
				},
				Status: pb.MSI_UP,
			},
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respCreateInstance.Response.GetCode())
		instanceID1 = respCreateInstance.InstanceId

		respCreateInstance, err = datasource.Instance().RegisterInstance(getContext(), &pb.RegisterInstanceRequest{
			Instance: &pb.MicroServiceInstance{
				ServiceId: serviceID2,
				HostName:  "UT-HOST-MS",
				Endpoints: []string{
					"find:127.0.0.2:8080",
				},
				Status: pb.MSI_UP,
			},
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respCreateInstance.Response.GetCode())
		instanceID2 = respCreateInstance.InstanceId

		respCreateInstance, err = datasource.Instance().RegisterInstance(getContext(), &pb.RegisterInstanceRequest{
			Instance: &pb.MicroServiceInstance{
				ServiceId: serviceID4,
				HostName:  "UT-HOST-MS",
				Endpoints: []string{
					"find:127.0.0.4:8080",
				},
				Status: pb.MSI_UP,
			},
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respCreateInstance.Response.GetCode())
		instanceID4 = respCreateInstance.InstanceId

		respCreateInstance, err = datasource.Instance().RegisterInstance(getContext(), &pb.RegisterInstanceRequest{
			Instance: &pb.MicroServiceInstance{
				ServiceId: serviceID5,
				HostName:  "UT-HOST-MS",
				Endpoints: []string{
					"find:127.0.0.5:8080",
				},
				Status: pb.MSI_UP,
			},
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respCreateInstance.Response.GetCode())
		instanceID5 = respCreateInstance.InstanceId

		respCreateInstance, err = datasource.Instance().RegisterInstance(getContext(), &pb.RegisterInstanceRequest{
			Instance: &pb.MicroServiceInstance{
				ServiceId: serviceID8,
				HostName:  "UT-HOST-MS",
				Endpoints: []string{
					"find:127.0.0.8:8080",
				},
				Status: pb.MSI_UP,
			},
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respCreateInstance.Response.GetCode())

		respCreateInstance, err = datasource.Instance().RegisterInstance(getContext(), &pb.RegisterInstanceRequest{
			Instance: &pb.MicroServiceInstance{
				ServiceId: serviceID9,
				HostName:  "UT-HOST-MS",
				Endpoints: []string{
					"find:127.0.0.9:8080",
				},
				Status: pb.MSI_UP,
			},
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respCreateInstance.Response.GetCode())
	})

	t.Run("query instance, when request is invalid, should pass", func(t *testing.T) {
		log.Info("find with version rule")
		respFind, err := datasource.Instance().FindInstances(getContext(), &pb.FindInstancesRequest{
			ConsumerServiceId: serviceID1,
			AppId:             "query_instance_ms",
			ServiceName:       "query_instance_service_ms",
			VersionRule:       "latest",
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respFind.Response.GetCode())
		assert.Equal(t, instanceID2, respFind.Instances[0].InstanceId)

		respFind, err = datasource.Instance().FindInstances(getContext(), &pb.FindInstancesRequest{
			ConsumerServiceId: serviceID1,
			AppId:             "query_instance_ms",
			ServiceName:       "query_instance_service_ms",
			VersionRule:       "1.0.0+",
			Tags:              []string{},
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respFind.Response.GetCode())
		assert.Equal(t, instanceID2, respFind.Instances[0].InstanceId)

		respFind, err = datasource.Instance().FindInstances(getContext(), &pb.FindInstancesRequest{
			ConsumerServiceId: serviceID1,
			AppId:             "query_instance_ms",
			ServiceName:       "query_instance_service_ms",
			VersionRule:       "1.0.0",
			Tags:              []string{},
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respFind.Response.GetCode())
		assert.Equal(t, instanceID1, respFind.Instances[0].InstanceId)

		respFind, err = datasource.Instance().FindInstances(getContext(), &pb.FindInstancesRequest{
			ConsumerServiceId: serviceID1,
			AppId:             "query_instance",
			ServiceName:       "query_instance_service",
			VersionRule:       "0.0.0",
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ErrServiceNotExists, respFind.Response.GetCode())

		log.Info("find with env")
		respFind, err = datasource.Instance().FindInstances(getContext(), &pb.FindInstancesRequest{
			ConsumerServiceId: serviceID4,
			AppId:             "query_instance_ms",
			ServiceName:       "query_instance_diff_env_service_ms",
			VersionRule:       "1.0.0",
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respFind.Response.GetCode())
		assert.Equal(t, 1, len(respFind.Instances))
		assert.Equal(t, instanceID4, respFind.Instances[0].InstanceId)

		respFind, err = datasource.Instance().FindInstances(getContext(), &pb.FindInstancesRequest{
			Environment: pb.ENV_PROD,
			AppId:       "query_instance_ms",
			ServiceName: "query_instance_diff_env_service_ms",
			VersionRule: "1.0.0",
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respFind.Response.GetCode())
		assert.Equal(t, 1, len(respFind.Instances))
		assert.Equal(t, instanceID4, respFind.Instances[0].InstanceId)

		log.Info("find should return 200 if consumer is diff apps")
		respFind, err = datasource.Instance().FindInstances(getContext(), &pb.FindInstancesRequest{
			ConsumerServiceId: serviceID3,
			AppId:             "query_instance_ms",
			ServiceName:       "query_instance_service_ms",
			VersionRule:       "1.0.5",
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respFind.Response.GetCode())
		assert.Equal(t, 0, len(respFind.Instances))

		log.Info("provider tag does not exist")
		respFind, err = datasource.Instance().FindInstances(getContext(), &pb.FindInstancesRequest{
			ConsumerServiceId: serviceID1,
			AppId:             "query_instance_ms",
			ServiceName:       "query_instance_service_ms",
			VersionRule:       "latest",
			Tags:              []string{"not_exist_tag"},
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respFind.Response.GetCode())
		assert.Equal(t, 0, len(respFind.Instances))

		log.Info("shared service discovery")
		config.ServerInfo.Config.GlobalVisible = "query_instance_shared_provider_ms"
		core.RegisterGlobalServices()
		core.Service.Environment = pb.ENV_PROD
		respFind, err = datasource.Instance().FindInstances(
			util.SetTargetDomainProject(
				util.SetDomainProject(util.CloneContext(getContext()), "user", "user"),
				"default", "default"),
			&pb.FindInstancesRequest{
				ConsumerServiceId: serviceID6,
				AppId:             "default",
				ServiceName:       "query_instance_shared_provider_ms",
				VersionRule:       "1.0.0",
			})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respFind.Response.GetCode())
		assert.Equal(t, 1, len(respFind.Instances))
		assert.Equal(t, instanceID5, respFind.Instances[0].InstanceId)

		respFind, err = datasource.Instance().FindInstances(getContext(), &pb.FindInstancesRequest{
			ConsumerServiceId: serviceID7,
			AppId:             "default",
			ServiceName:       "query_instance_shared_provider_ms",
			VersionRule:       "1.0.0",
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respFind.Response.GetCode())
		assert.Equal(t, 1, len(respFind.Instances))
		assert.Equal(t, instanceID5, respFind.Instances[0].InstanceId)

		log.Info("query same domain deps")
		// todo finish ut after implementing GetConsumerDependencies interface

		core.Service.Environment = pb.ENV_DEV
	})

	t.Run("batch query instances, when request is invalid, should pass", func(t *testing.T) {
		log.Info("find with version rule")
		respFind, err := datasource.Instance().BatchFind(getContext(), &pb.BatchFindInstancesRequest{
			ConsumerServiceId: serviceID1,
			Services: []*pb.FindService{
				{
					Service: &pb.MicroServiceKey{
						AppId:       "query_instance_ms",
						ServiceName: "query_instance_service_ms",
						Version:     "latest",
					},
				},
				{
					Service: &pb.MicroServiceKey{
						AppId:       "query_instance_ms",
						ServiceName: "query_instance_service_ms",
						Version:     "1.0.0+",
					},
				},
				{
					Service: &pb.MicroServiceKey{
						AppId:       "query_instance_ms",
						ServiceName: "query_instance_service_ms",
						Version:     "0.0.0",
					},
				},
			},
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respFind.Response.GetCode())
		assert.Equal(t, int64(0), respFind.Services.Updated[0].Index)
		assert.Equal(t, instanceID2, respFind.Services.Updated[0].Instances[0].InstanceId)
		assert.Equal(t, int64(1), respFind.Services.Updated[1].Index)
		assert.Equal(t, instanceID2, respFind.Services.Updated[1].Instances[0].InstanceId)
		assert.Equal(t, int64(2), respFind.Services.Failed[0].Indexes[0])
		assert.Equal(t, pb.ErrServiceNotExists, respFind.Services.Failed[0].Error.Code)

		log.Info("find with env")
		respFind, err = datasource.Instance().BatchFind(getContext(), &pb.BatchFindInstancesRequest{
			ConsumerServiceId: serviceID4,
			Services: []*pb.FindService{
				{
					Service: &pb.MicroServiceKey{
						AppId:       "query_instance_ms",
						ServiceName: "query_instance_diff_env_service_ms",
						Version:     "1.0.0",
					},
				},
			},
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respFind.Response.GetCode())
		assert.Equal(t, 1, len(respFind.Services.Updated[0].Instances))
		assert.Equal(t, instanceID4, respFind.Services.Updated[0].Instances[0].InstanceId)

		log.Info("find should return 200 even if consumer is diff apps")
		respFind, err = datasource.Instance().BatchFind(getContext(), &pb.BatchFindInstancesRequest{
			ConsumerServiceId: serviceID3,
			Services: []*pb.FindService{
				{
					Service: &pb.MicroServiceKey{
						AppId:       "query_instance_ms",
						ServiceName: "query_instance_service_ms",
						Version:     "1.0.5",
					},
				},
			},
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respFind.Response.GetCode())
		assert.Equal(t, 0, len(respFind.Services.Updated[0].Instances))

		log.Info("shared service discovery")
		config.ServerInfo.Config.GlobalVisible = "query_instance_shared_provider_ms"
		core.RegisterGlobalServices()
		core.Service.Environment = pb.ENV_PROD
		respFind, err = datasource.Instance().BatchFind(
			util.SetTargetDomainProject(
				util.SetDomainProject(util.CloneContext(getContext()), "user", "user"),
				"default", "default"),
			&pb.BatchFindInstancesRequest{
				ConsumerServiceId: serviceID6,
				Services: []*pb.FindService{
					{
						Service: &pb.MicroServiceKey{
							AppId:       "default",
							ServiceName: "query_instance_shared_provider_ms",
							Version:     "1.0.0",
						},
					},
				},
			})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respFind.Response.GetCode())
		assert.Equal(t, 1, len(respFind.Services.Updated[0].Instances))
		assert.Equal(t, instanceID5, respFind.Services.Updated[0].Instances[0].InstanceId)

		respFind, err = datasource.Instance().BatchFind(getContext(), &pb.BatchFindInstancesRequest{
			ConsumerServiceId: serviceID7,
			Services: []*pb.FindService{
				{
					Service: &pb.MicroServiceKey{
						AppId:       "default",
						ServiceName: "query_instance_shared_provider_ms",
						Version:     "1.0.0",
					},
				},
			},
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respFind.Response.GetCode())
		assert.Equal(t, 1, len(respFind.Services.Updated[0].Instances))
		assert.Equal(t, instanceID5, respFind.Services.Updated[0].Instances[0].InstanceId)

		respFind, err = datasource.Instance().BatchFind(util.SetTargetDomainProject(
			util.SetDomainProject(util.CloneContext(getContext()), "user", "user"),
			"default", "default"),
			&pb.BatchFindInstancesRequest{
				ConsumerServiceId: serviceID6,
				Instances: []*pb.FindInstance{
					{
						Instance: &pb.HeartbeatSetElement{
							ServiceId:  serviceID5,
							InstanceId: instanceID5,
						},
					},
				},
			})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respFind.Response.GetCode())
		assert.Equal(t, pb.ErrServiceNotExists, respFind.Instances.Failed[0].Error.Code)

		respFind, err = datasource.Instance().BatchFind(getContext(), &pb.BatchFindInstancesRequest{
			ConsumerServiceId: serviceID7,
			Instances: []*pb.FindInstance{
				{
					Instance: &pb.HeartbeatSetElement{
						ServiceId:  serviceID5,
						InstanceId: instanceID5,
					},
				},
			},
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respFind.Response.GetCode())
		assert.Equal(t, 1, len(respFind.Instances.Updated[0].Instances))
		assert.Equal(t, instanceID5, respFind.Instances.Updated[0].Instances[0].InstanceId)

		core.Service.Environment = pb.ENV_DEV
	})

	t.Run("query instances between diff dimensions", func(t *testing.T) {
		log.Info("diff appId")
		UTFunc := func(consumerId string, code int32) {
			respFind, err := datasource.Instance().GetInstances(getContext(), &pb.GetInstancesRequest{
				ConsumerServiceId: consumerId,
				ProviderServiceId: serviceID2,
			})
			assert.NoError(t, err)
			assert.Equal(t, code, respFind.Response.GetCode())
		}

		UTFunc(serviceID3, pb.ErrServiceNotExists)

		UTFunc(serviceID1, pb.ResponseSuccess)

		log.Info("diff env")
		respFind, err := datasource.Instance().GetInstances(getContext(), &pb.GetInstancesRequest{
			ConsumerServiceId: serviceID4,
			ProviderServiceId: serviceID2,
		})
		assert.NoError(t, err)
		assert.NotEqual(t, pb.ResponseSuccess, respFind.Response.GetCode())
	})

	t.Run("delete test data", func(t *testing.T) {
		_, err := client.GetMongoClient().Delete(getContext(), db.CollectionService, bson.M{"domain": "default", "project": "default"})
		assert.NoError(t, err)

		_, err = client.GetMongoClient().Delete(getContext(), db.CollectionInstance, bson.M{"domain": "default", "project": "default"})
		assert.NoError(t, err)

		_, err = client.GetMongoClient().Delete(getContext(), db.CollectionService, bson.M{"domain": "user", "project": "user"})
		assert.NoError(t, err)

		_, err = client.GetMongoClient().Delete(getContext(), db.CollectionInstance, bson.M{"domain": "user", "project": "user"})
		assert.NoError(t, err)
	})
}

func TestInstance_GetOne(t *testing.T) {
	var (
		serviceID1  string
		serviceID2  string
		serviceID3  string
		instanceID2 string
	)

	t.Run("register service and instances, when request is invalid, should pass", func(t *testing.T) {
		respCreateService, err := datasource.Instance().RegisterService(getContext(), &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				AppId:       "get_instance_ms",
				ServiceName: "get_instance_service_ms",
				Version:     "1.0.0",
				Level:       "FRONT",
				Status:      pb.MS_UP,
			},
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respCreateService.Response.GetCode())
		serviceID1 = respCreateService.ServiceId

		respCreateService, err = datasource.Instance().RegisterService(getContext(), &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				AppId:       "get_instance_ms",
				ServiceName: "get_instance_service_ms",
				Version:     "1.0.5",
				Level:       "FRONT",
				Status:      pb.MS_UP,
			},
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respCreateService.Response.GetCode())
		serviceID2 = respCreateService.ServiceId

		respCreateInstance, err := datasource.Instance().RegisterInstance(getContext(), &pb.RegisterInstanceRequest{
			Instance: &pb.MicroServiceInstance{
				ServiceId: serviceID2,
				HostName:  "UT-HOST-MS",
				Endpoints: []string{
					"get:127.0.0.2:8080",
				},
				Status: pb.MSI_UP,
			},
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respCreateService.Response.GetCode())
		instanceID2 = respCreateInstance.InstanceId

		respCreateService, err = datasource.Instance().RegisterService(getContext(), &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				AppId:       "get_instance_cross_ms",
				ServiceName: "get_instance_service_ms",
				Version:     "1.0.0",
				Level:       "FRONT",
				Status:      pb.MS_UP,
			},
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respCreateService.Response.GetCode())
		serviceID3 = respCreateService.ServiceId
	})

	t.Run("get one instance when invalid request", func(t *testing.T) {
		log.Info("find service itself")
		resp, err := datasource.Instance().GetInstance(getContext(), &pb.GetOneInstanceRequest{
			ConsumerServiceId:  serviceID2,
			ProviderServiceId:  serviceID2,
			ProviderInstanceId: instanceID2,
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, resp.Response.GetCode())

		log.Info("consumer does not exist")
		resp, err = datasource.Instance().GetInstance(getContext(), &pb.GetOneInstanceRequest{
			ConsumerServiceId:  "not-exist-id-ms",
			ProviderServiceId:  serviceID2,
			ProviderInstanceId: instanceID2,
		})
		assert.NoError(t, err)
		assert.NotEqual(t, pb.ResponseSuccess, resp.Response.GetCode())
	})

	t.Run("get between diff apps", func(t *testing.T) {
		resp, err := datasource.Instance().GetInstance(getContext(), &pb.GetOneInstanceRequest{
			ConsumerServiceId:  serviceID3,
			ProviderServiceId:  serviceID2,
			ProviderInstanceId: instanceID2,
		})
		assert.NoError(t, err)
		assert.NotEqual(t, pb.ResponseSuccess, resp.Response.GetCode())

		respAll, err := datasource.Instance().GetInstances(getContext(), &pb.GetInstancesRequest{
			ConsumerServiceId: serviceID3,
			ProviderServiceId: serviceID2,
		})
		assert.NoError(t, err)
		assert.NotEqual(t, pb.ResponseSuccess, respAll.Response.GetCode())
	})

	t.Run("get instances when request is invalid", func(t *testing.T) {
		log.Info("consumer does not exist")
		resp, err := datasource.Instance().GetInstances(getContext(), &pb.GetInstancesRequest{
			ConsumerServiceId: "not-exist-service-ms",
			ProviderServiceId: serviceID2,
		})
		assert.NoError(t, err)
		assert.NotEqual(t, pb.ResponseSuccess, resp.Response.GetCode())

		log.Info("consumer does not exist")
		resp, err = datasource.Instance().GetInstances(getContext(), &pb.GetInstancesRequest{
			ConsumerServiceId: serviceID1,
			ProviderServiceId: serviceID2,
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, resp.Response.GetCode())
	})

	t.Run("delete test data", func(t *testing.T) {
		_, err := client.GetMongoClient().Delete(getContext(), db.CollectionService, bson.M{"domain": "default", "project": "default"})
		assert.NoError(t, err)

		_, err = client.GetMongoClient().Delete(getContext(), db.CollectionInstance, bson.M{"domain": "default", "project": "default"})
		assert.NoError(t, err)
	})
}

func TestInstance_GetAll(t *testing.T) {

	t.Run("register 2 instances, get all instances count should return 2", func(t *testing.T) {
		var (
			serviceId1 string
			serviceId2 string
		)
		ctx := util.WithNoCache(util.SetDomainProject(getContext(), "TestInstance_GetAll", "1"))
		respCreateService, err := datasource.Instance().RegisterService(ctx, &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				AppId:       "get_instance_ms",
				ServiceName: "get_instance_service_ms",
				Version:     "1.0.0",
				Level:       "FRONT",
				Status:      pb.MS_UP,
			},
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respCreateService.Response.GetCode())
		serviceId1 = respCreateService.ServiceId

		respCreateService, err = datasource.Instance().RegisterService(ctx, &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				AppId:       "get_instance_ms",
				ServiceName: "get_instance_service_ms",
				Version:     "1.0.5",
				Level:       "FRONT",
				Status:      pb.MS_UP,
			},
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respCreateService.Response.GetCode())
		serviceId2 = respCreateService.ServiceId

		respCreateInstance, err := datasource.Instance().RegisterInstance(ctx, &pb.RegisterInstanceRequest{
			Instance: &pb.MicroServiceInstance{
				ServiceId: serviceId1,
				HostName:  "UT-HOST-MS",
				Endpoints: []string{
					"get:127.0.0.2:8080",
				},
				Status: pb.MSI_UP,
			},
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respCreateInstance.Response.GetCode())

		respCreateInstance, err = datasource.Instance().RegisterInstance(ctx, &pb.RegisterInstanceRequest{
			Instance: &pb.MicroServiceInstance{
				ServiceId: serviceId2,
				HostName:  "UT-HOST-MS",
				Endpoints: []string{
					"get:127.0.0.3:8080",
				},
				Status: pb.MSI_UP,
			},
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respCreateInstance.Response.GetCode())

		respAll, err := datasource.Instance().GetAllInstances(ctx, &pb.GetAllInstancesRequest{})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respAll.Response.GetCode())
		assert.Equal(t, 2, len(respAll.Instances))
	})

	t.Run("domain contain no instances, get all instances should be pass, return 0 instance", func(t *testing.T) {
		ctx := util.WithNoCache(util.SetDomainProject(getContext(), "TestInstance_GetAll", "2"))
		respAll, err := datasource.Instance().GetAllInstances(ctx, &pb.GetAllInstancesRequest{})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respAll.Response.GetCode())
		assert.Equal(t, 0, len(respAll.Instances))
	})

	t.Run("delete test data", func(t *testing.T) {
		_, err := client.GetMongoClient().Delete(getContext(), db.CollectionService, bson.M{"domain": "TestInstance_GetAll", "project": "1"})
		assert.NoError(t, err)

		_, err = client.GetMongoClient().Delete(getContext(), db.CollectionInstance, bson.M{"domain": "TestInstance_GetAll", "project": "1"})
		assert.NoError(t, err)
	})
}

func TestInstance_Unregister(t *testing.T) {
	var (
		serviceId  string
		instanceId string
	)

	t.Run("register service and instances", func(t *testing.T) {
		respCreateService, err := datasource.Instance().RegisterService(getContext(), &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				AppId:       "unregister_instance_ms",
				ServiceName: "unregister_instance_service_ms",
				Version:     "1.0.5",
				Level:       "FRONT",
				Status:      pb.MS_UP,
			},
			Tags: map[string]string{
				"test": "test",
			},
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respCreateService.Response.GetCode())
		serviceId = respCreateService.ServiceId

		respCreateInstance, err := datasource.Instance().RegisterInstance(getContext(), &pb.RegisterInstanceRequest{
			Instance: &pb.MicroServiceInstance{
				ServiceId: serviceId,
				HostName:  "UT-HOST-MS",
				Endpoints: []string{
					"unregister:127.0.0.2:8080",
				},
				Status: pb.MSI_UP,
			},
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respCreateInstance.Response.GetCode())
		instanceId = respCreateInstance.InstanceId
	})

	t.Run("unregister instance", func(t *testing.T) {
		resp, err := datasource.Instance().UnregisterInstance(getContext(), &pb.UnregisterInstanceRequest{
			ServiceId:  serviceId,
			InstanceId: instanceId,
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, resp.Response.GetCode())
	})

	t.Run("unregister instance when request is invalid", func(t *testing.T) {
		log.Info("service id does not exist")
		resp, err := datasource.Instance().UnregisterInstance(getContext(), &pb.UnregisterInstanceRequest{
			ServiceId:  "not-exist-id-ms",
			InstanceId: instanceId,
		})
		assert.NoError(t, err)
		assert.NotEqual(t, pb.ResponseSuccess, resp.Response.GetCode())

		log.Info("instance id does not exist")
		resp, err = datasource.Instance().UnregisterInstance(getContext(), &pb.UnregisterInstanceRequest{
			ServiceId:  serviceId,
			InstanceId: "not-exist-id-ms",
		})
		assert.NoError(t, err)
		assert.NotEqual(t, pb.ResponseSuccess, resp.Response.GetCode())
	})
}
