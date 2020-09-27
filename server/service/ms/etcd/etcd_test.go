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
	pb "github.com/apache/servicecomb-service-center/pkg/registry"
	"github.com/apache/servicecomb-service-center/server/core/proto"
	"github.com/apache/servicecomb-service-center/server/plugin/quota"
	scerr "github.com/apache/servicecomb-service-center/server/scerror"
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
	TooLongAlias       = strings.Repeat("x", 130)
	TooLongFramework   = strings.Repeat("x", 66)
)

func TestInit(t *testing.T) {
	_ = archaius.Init(archaius.WithMemorySource())
	_ = archaius.Set("servicecomb.ms.name", "etcd")
}

func TestEtcd_RegisterService(t *testing.T) {
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

func TestEtcd_GetService(t *testing.T) {
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
