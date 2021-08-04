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

package datasource_test

import (
	"context"
	"github.com/go-chassis/cari/pkg/errsvc"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/apache/servicecomb-service-center/datasource"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/util"
	"github.com/apache/servicecomb-service-center/server/plugin/quota"
	pb "github.com/go-chassis/cari/discovery"
	"github.com/stretchr/testify/assert"
)

func TestService_Register(t *testing.T) {
	t.Run("Register service after init & install, should pass", func(t *testing.T) {
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
				Framework: &pb.FrameWork{
					Name:    "service-ms-frameworkName",
					Version: "service-ms-frameworkVersion",
				},
				RegisterBy: "SDK",
				Timestamp:  strconv.FormatInt(time.Now().Unix(), 10),
			},
		}
		request.Service.ModTimestamp = request.Service.Timestamp
		resp, err := datasource.GetMetadataManager().RegisterService(getContext(), request)
		assert.NotNil(t, resp)
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, resp.Response.GetCode())
	})

	t.Run("register service with same key", func(t *testing.T) {
		// serviceName: some-relay-ms-service-name
		// alias: sr-ms-service-name
		resp, err := datasource.GetMetadataManager().RegisterService(getContext(), &pb.CreateServiceRequest{
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
		resp, err = datasource.GetMetadataManager().RegisterService(getContext(), &pb.CreateServiceRequest{
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
		resp, err = datasource.GetMetadataManager().RegisterService(getContext(), &pb.CreateServiceRequest{
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
		resp, err = datasource.GetMetadataManager().RegisterService(getContext(), &pb.CreateServiceRequest{
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
		resp, err = datasource.GetMetadataManager().RegisterService(getContext(), &pb.CreateServiceRequest{
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
		resp, err = datasource.GetMetadataManager().RegisterService(getContext(), &pb.CreateServiceRequest{
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
		resp, err = datasource.GetMetadataManager().RegisterService(getContext(), &pb.CreateServiceRequest{
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
			resp, err := datasource.GetMetadataManager().RegisterService(getContext(), &pb.CreateServiceRequest{
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
			resp, err = datasource.GetMetadataManager().RegisterService(getContext(), &pb.CreateServiceRequest{
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
	// get service test
	t.Run("query all services, should pass", func(t *testing.T) {
		resp, err := datasource.GetMetadataManager().GetServices(getContext(), &pb.GetServicesRequest{})
		assert.NoError(t, err)
		assert.Greater(t, len(resp.Services), 0)
	})

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

		resp, err := datasource.GetMetadataManager().RegisterService(getContext(), request)
		assert.NoError(t, err)
		assert.Equal(t, resp.Response.GetCode(), pb.ResponseSuccess)

		// search service by serviceID
		_, err = datasource.GetMetadataManager().GetService(getContext(), &pb.GetServiceRequest{
			ServiceId: "ms-service-query-id",
		})
		assert.NoError(t, err)
	})

	t.Run("query a service by a not existed serviceId, should not pass", func(t *testing.T) {
		// not exist service
		_, err := datasource.GetMetadataManager().GetService(getContext(), &pb.GetServiceRequest{
			ServiceId: "no-exist-service",
		})
		assert.NotNil(t, err)
		assert.Equal(t, pb.ErrServiceNotExists, err.(*errsvc.Error).Code)
	})
}

func TestService_Exist(t *testing.T) {
	var (
		serviceId1 string
		serviceId2 string
	)
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
		resp, err := datasource.GetMetadataManager().RegisterService(getContext(), &pb.CreateServiceRequest{
			Service: svc,
		})
		assert.NoError(t, err)
		assert.NotEqual(t, "", resp.ServiceId)
		serviceId1 = resp.ServiceId

		svc.ServiceId = ""
		svc.Environment = pb.ENV_PROD
		resp, err = datasource.GetMetadataManager().RegisterService(getContext(), &pb.CreateServiceRequest{
			Service: svc,
		})
		assert.NoError(t, err)
		assert.NotEqual(t, "", resp.ServiceId)
		serviceId2 = resp.ServiceId
	})

	t.Run("check exist when service does not exist", func(t *testing.T) {
		log.Info("check by querying a not exist serviceName")
		resp, err := datasource.GetMetadataManager().ExistService(getContext(), &pb.GetExistenceRequest{
			Type:        datasource.ExistTypeMicroservice,
			AppId:       "exist_appId",
			ServiceName: "notExistService_service_ms",
			Version:     "1.0.0",
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ErrServiceNotExists, resp.Response.GetCode())

		log.Info("check by querying a not exist env")
		resp, err = datasource.GetMetadataManager().ExistService(getContext(), &pb.GetExistenceRequest{
			Type:        datasource.ExistTypeMicroservice,
			Environment: pb.ENV_TEST,
			AppId:       "exist_appId_service_ms",
			ServiceName: "exist_service_service_ms",
			Version:     "1.0.0",
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ErrServiceNotExists, resp.Response.GetCode())

		log.Info("check by querying a not exist env with alias")
		resp, err = datasource.GetMetadataManager().ExistService(getContext(), &pb.GetExistenceRequest{
			Type:        datasource.ExistTypeMicroservice,
			Environment: pb.ENV_TEST,
			AppId:       "exist_appId_service_ms",
			ServiceName: "es_service_ms",
			Version:     "1.0.0",
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ErrServiceNotExists, resp.Response.GetCode())

		log.Info("check by querying with a mismatching version")
		resp, err = datasource.GetMetadataManager().ExistService(getContext(), &pb.GetExistenceRequest{
			Type:        datasource.ExistTypeMicroservice,
			AppId:       "exist_appId_service_ms",
			ServiceName: "exist_service_service_ms",
			Version:     "2.0.0",
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ErrServiceVersionNotExists, resp.Response.GetCode())
	})

	t.Run("check exist when service exists", func(t *testing.T) {
		log.Info("search with serviceName")
		resp, err := datasource.GetMetadataManager().ExistService(getContext(), &pb.GetExistenceRequest{
			Type:        datasource.ExistTypeMicroservice,
			AppId:       "exist_appId_service_ms",
			ServiceName: "exist_service_service_ms",
			Version:     "1.0.0",
		})
		assert.NoError(t, err)
		assert.Equal(t, serviceId1, resp.ServiceId)

		log.Info("check with serviceName and env")
		resp, err = datasource.GetMetadataManager().ExistService(getContext(), &pb.GetExistenceRequest{
			Type:        datasource.ExistTypeMicroservice,
			Environment: pb.ENV_PROD,
			AppId:       "exist_appId_service_ms",
			ServiceName: "exist_service_service_ms",
			Version:     "1.0.0",
		})
		assert.NoError(t, err)
		assert.Equal(t, serviceId2, resp.ServiceId)

		log.Info("check with alias")
		resp, err = datasource.GetMetadataManager().ExistService(getContext(), &pb.GetExistenceRequest{
			Type:        datasource.ExistTypeMicroservice,
			AppId:       "exist_appId_service_ms",
			ServiceName: "es_service_ms",
			Version:     "1.0.0",
		})
		assert.NoError(t, err)
		assert.Equal(t, serviceId1, resp.ServiceId)

		log.Info("check with alias and env")
		resp, err = datasource.GetMetadataManager().ExistService(getContext(), &pb.GetExistenceRequest{
			Type:        datasource.ExistTypeMicroservice,
			Environment: pb.ENV_PROD,
			AppId:       "exist_appId_service_ms",
			ServiceName: "es_service_ms",
			Version:     "1.0.0",
		})
		assert.NoError(t, err)
		assert.Equal(t, serviceId2, resp.ServiceId)
	})
}

func TestService_Update(t *testing.T) {
	var serviceId string

	t.Run("create service", func(t *testing.T) {
		resp, err := datasource.GetMetadataManager().RegisterService(getContext(), &pb.CreateServiceRequest{
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
		log.Info("shuold pass")
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
		resp, err := datasource.GetMetadataManager().UpdateService(getContext(), request)
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, resp.Response.GetCode())

		resp, err = datasource.GetMetadataManager().UpdateService(getContext(), request2)
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, resp.Response.GetCode())

		service, err := datasource.GetMetadataManager().GetService(getContext(), &pb.GetServiceRequest{
			ServiceId: serviceId,
		})
		assert.NoError(t, err)
		assert.Equal(t, serviceId, service.ServiceId)
		assert.Equal(t, "", service.Properties["test"])
		assert.Equal(t, "v", service.Properties["k"])
	})

	t.Run("update service that does not exist", func(t *testing.T) {
		log.Info("it should be failed")
		r := &pb.UpdateServicePropsRequest{
			ServiceId:  "not_exist_service_service_ms",
			Properties: make(map[string]string),
		}
		resp, err := datasource.GetMetadataManager().UpdateService(getContext(), r)
		assert.NoError(t, err)
		assert.NotEqual(t, pb.ResponseSuccess, resp.Response.GetCode())
	})

	t.Run("update service by removing the properties", func(t *testing.T) {
		log.Info("it should pass")
		r := &pb.UpdateServicePropsRequest{
			ServiceId:  serviceId,
			Properties: nil,
		}
		resp, err := datasource.GetMetadataManager().UpdateService(getContext(), r)
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, resp.Response.GetCode())

		log.Info("remove properties for service with empty serviceId")
		r = &pb.UpdateServicePropsRequest{
			ServiceId:  "",
			Properties: map[string]string{},
		}
		resp, err = datasource.GetMetadataManager().UpdateService(getContext(), r)
		assert.NoError(t, err)
		assert.NotEqual(t, pb.ResponseSuccess, resp.Response.GetCode())
	})
}

func TestService_Delete(t *testing.T) {
	var (
		serviceContainInstId string
		serviceNoInstId      string
	)

	t.Run("create service & instance", func(t *testing.T) {
		respCreate, err := datasource.GetMetadataManager().RegisterService(getContext(), &pb.CreateServiceRequest{
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
		respCreateIns, err := datasource.GetMetadataManager().RegisterInstance(getContext(), &pb.RegisterInstanceRequest{
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
		respCreate, err = datasource.GetMetadataManager().RegisterService(getContext(), &pb.CreateServiceRequest{
			Service: provider,
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respCreate.Response.GetCode())
		serviceNoInstId = respCreate.ServiceId

		respCreate, err = datasource.GetMetadataManager().RegisterService(getContext(), &pb.CreateServiceRequest{
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
		resp, err := datasource.GetMetadataManager().UnregisterService(getContext(), &pb.DeleteServiceRequest{
			ServiceId: serviceContainInstId,
			Force:     false,
		})
		assert.NoError(t, err)
		assert.NotEqual(t, pb.ResponseSuccess, resp.Response.GetCode())
	})

	t.Run("delete a service which contains instances with force flag", func(t *testing.T) {
		log.Info("should pass")
		resp, err := datasource.GetMetadataManager().UnregisterService(getContext(), &pb.DeleteServiceRequest{
			ServiceId: serviceContainInstId,
			Force:     true,
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, resp.Response.GetCode())
	})

	// todo: add delete service depended by consumer after finishing dependency management

	t.Run("delete a service which depended by consumer with force flag", func(t *testing.T) {
		log.Info("should pass")
		resp, err := datasource.GetMetadataManager().UnregisterService(getContext(), &pb.DeleteServiceRequest{
			ServiceId: serviceNoInstId,
			Force:     true,
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, resp.Response.GetCode())
	})

	t.Run("delete a service with no force flag", func(t *testing.T) {
		log.Info("should not pass")
		resp, err := datasource.GetMetadataManager().UnregisterService(getContext(), &pb.DeleteServiceRequest{
			ServiceId: serviceNoInstId,
			Force:     false,
		})
		assert.NoError(t, err)
		assert.NotEqual(t, pb.ResponseSuccess, resp.Response.GetCode())
	})
}

func TestService_Info(t *testing.T) {
	var (
		serviceID  string
		instanceID string
	)
	t.Run("register a tested service & instance, should be passed", func(t *testing.T) {
		resp, err := datasource.GetMetadataManager().RegisterService(getContext(), &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				ServiceName: "TestServic1",
				AppId:       "default",
				Version:     "1.0.0",
				Level:       "FRONT",
				Status:      "UP",
			},
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, resp.Response.GetCode())
		serviceID = resp.ServiceId
		instance := &pb.MicroServiceInstance{
			ServiceId: serviceID,
			Endpoints: []string{
				"rest://[::]:301000/",
			},
			HostName: "delete-host-ms",
			Status:   pb.MSI_UP,
		}
		respInstance, err := datasource.GetMetadataManager().RegisterInstance(getContext(), &pb.RegisterInstanceRequest{
			Instance: instance,
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respInstance.Response.GetCode())
		instanceID = respInstance.InstanceId
	})

	t.Run("get tested service info, should be passed", func(t *testing.T) {
		resp, err := datasource.GetMetadataManager().ListServiceDetail(getContext(), &pb.GetServicesInfoRequest{
			Options:     []string{"all"},
			AppId:       "default",
			ServiceName: "TestServic1",
		})
		assert.NoError(t, err)
		assert.NotEqual(t, 0, len(resp.AllServicesDetail))
		assert.NotEqual(t, 0, resp.Statistics.Services.Count)
		assert.NotEqual(t, 0, resp.Statistics.Instances.Count)
		assert.NotEqual(t, 0, resp.Statistics.Apps.Count)
	})

	t.Run("unregister tested service & instance, should be passed", func(t *testing.T) {
		resp, err := datasource.GetMetadataManager().UnregisterInstance(getContext(), &pb.UnregisterInstanceRequest{
			ServiceId:  serviceID,
			InstanceId: instanceID,
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, resp.Response.GetCode())
		respService, err := datasource.GetMetadataManager().UnregisterService(getContext(), &pb.DeleteServiceRequest{
			ServiceId: serviceID,
			Force:     true,
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respService.Response.GetCode())
	})

	t.Run("get all services", func(t *testing.T) {
		log.Info("should be passed")
		resp, err := datasource.GetMetadataManager().ListServiceDetail(getContext(), &pb.GetServicesInfoRequest{
			Options: []string{"all"},
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, resp.Response.GetCode())

		resp, err = datasource.GetMetadataManager().ListServiceDetail(getContext(), &pb.GetServicesInfoRequest{
			Options: []string{""},
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, resp.Response.GetCode())

		resp, err = datasource.GetMetadataManager().ListServiceDetail(getContext(), &pb.GetServicesInfoRequest{
			Options: []string{"tags", "rules", "instances", "schemas", "statistics"},
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, resp.Response.GetCode())

		resp, err = datasource.GetMetadataManager().ListServiceDetail(getContext(), &pb.GetServicesInfoRequest{
			Options: []string{"statistics"},
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, resp.Response.GetCode())

		resp, err = datasource.GetMetadataManager().ListServiceDetail(getContext(), &pb.GetServicesInfoRequest{
			Options:   []string{"instances"},
			CountOnly: true,
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, resp.Response.GetCode())
	})
}

func TestService_Detail(t *testing.T) {
	var (
		serviceId string
	)

	t.Run("execute 'get detail' operation", func(t *testing.T) {
		log.Info("should be passed")
		resp, err := datasource.GetMetadataManager().RegisterService(getContext(), &pb.CreateServiceRequest{
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

		datasource.GetMetadataManager().ModifySchema(getContext(), &pb.ModifySchemaRequest{
			ServiceId: serviceId,
			SchemaId:  "schemaId",
			Schema:    "detail",
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, resp.Response.GetCode())

		datasource.GetMetadataManager().RegisterInstance(getContext(), &pb.RegisterInstanceRequest{
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
		respD, err := datasource.GetMetadataManager().GetServiceDetail(getContext(), &pb.GetServiceRequest{
			ServiceId: "",
		})
		assert.Error(t, err)
		assert.Nil(t, respD)

		log.Info("when get a service detail, should be passed")
		respGetServiceDetail, err := datasource.GetMetadataManager().GetServiceDetail(getContext(), &pb.GetServiceRequest{
			ServiceId: serviceId,
		})
		assert.NoError(t, err)
		assert.Equal(t, serviceId, respGetServiceDetail.MicroService.ServiceId)

		respDelete, err := datasource.GetMetadataManager().UnregisterService(getContext(), &pb.DeleteServiceRequest{
			ServiceId: serviceId,
			Force:     true,
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respDelete.Response.GetCode())

		respGetServiceDetail, err = datasource.GetMetadataManager().GetServiceDetail(getContext(), &pb.GetServiceRequest{
			ServiceId: serviceId,
		})
		assert.Error(t, err)
		assert.Nil(t, respGetServiceDetail)
	})
}

func TestApplication_Get(t *testing.T) {
	t.Run("execute 'get apps' operation", func(t *testing.T) {
		log.Info("when request is valid, should be passed")
		resp, err := datasource.GetMetadataManager().ListApp(getContext(), &pb.GetAppsRequest{})
		assert.NoError(t, err)
		assert.NotNil(t, resp)

		resp, err = datasource.GetMetadataManager().ListApp(getContext(), &pb.GetAppsRequest{
			Environment: pb.ENV_ACCEPT,
		})
		assert.NoError(t, err)
		assert.NotNil(t, resp)
	})
}

func getContext() context.Context {
	return util.WithNoCache(util.SetDomainProject(context.Background(), "default", "default"))
}
