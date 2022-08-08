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
package govern_test

import (
	"context"
	"testing"

	_ "github.com/apache/servicecomb-service-center/test"

	"github.com/apache/servicecomb-service-center/pkg/util"
	"github.com/apache/servicecomb-service-center/server/service/disco"
	"github.com/apache/servicecomb-service-center/server/service/govern"
	pb "github.com/go-chassis/cari/discovery"
	"github.com/go-chassis/cari/pkg/errsvc"
	"github.com/stretchr/testify/assert"
)

func getContext() context.Context {
	return util.WithNoCache(util.SetDomainProject(context.Background(), "default", "default"))
}

func TestListServiceDetail(t *testing.T) {
	ctx := getContext()
	respC, err := disco.RegisterService(ctx, &pb.CreateServiceRequest{
		Service: &pb.MicroService{
			AppId:       "govern_service_group",
			ServiceName: "govern_service_graph",
			Version:     "1.0.0",
			Level:       "FRONT",
			Status:      pb.MS_UP,
		},
	})
	assert.NoError(t, err)
	assert.NotNil(t, respC)
	defer disco.UnregisterService(ctx, &pb.DeleteServiceRequest{ServiceId: respC.ServiceId, Force: true})

	t.Run("when get all services, should be passed", func(t *testing.T) {
		resp, err := govern.ListServiceDetail(ctx, &pb.GetServicesInfoRequest{
			Options: []string{"all"},
		})
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.NotNil(t, resp.Statistics)
		assert.NotNil(t, resp.AllServicesDetail)

		resp, err = govern.ListServiceDetail(ctx, &pb.GetServicesInfoRequest{
			Options: []string{""},
		})
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Nil(t, resp.Statistics)
		assert.NotNil(t, resp.AllServicesDetail)

		resp, err = govern.ListServiceDetail(ctx, &pb.GetServicesInfoRequest{
			Options: []string{"tags", "rules", "instances", "schemas", "statistics"},
		})
		assert.NoError(t, err)
		assert.NotNil(t, resp)

		resp, err = govern.ListServiceDetail(ctx, &pb.GetServicesInfoRequest{
			Options: []string{"statistics"},
		})
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.NotNil(t, resp.Statistics)
		assert.Nil(t, resp.AllServicesDetail)

		resp, err = govern.ListServiceDetail(ctx, &pb.GetServicesInfoRequest{
			Options:   []string{"instances"},
			CountOnly: true,
		})
		assert.NoError(t, err)
		assert.NotNil(t, resp)
	})

	t.Run("when get top graph, should be passed", func(t *testing.T) {

		graph, err := govern.Draw(ctx, false)
		assert.NoError(t, err)
		assert.NotEqual(t, 0, len(graph.Nodes))
	})
}

func TestListApp(t *testing.T) {
	t.Run("when request is invalid, should be failed", func(t *testing.T) {
		resp, err := govern.ListApp(getContext(), &pb.GetAppsRequest{
			Environment: "non-exist-env",
		})
		assert.True(t, errsvc.IsErrEqualCode(err, pb.ErrInvalidParams), err)
		assert.Nil(t, resp)
	})

	t.Run("when request is valid, should be passed", func(t *testing.T) {
		resp, err := govern.ListApp(getContext(), &pb.GetAppsRequest{})
		assert.NoError(t, err)
		assert.NotNil(t, resp)

		resp, err = govern.ListApp(getContext(), &pb.GetAppsRequest{
			Environment: pb.ENV_ACCEPT,
		})
		assert.NoError(t, err)
		assert.NotNil(t, resp)
	})
}

func TestGetServiceDetail(t *testing.T) {
	var (
		serviceId string
	)

	t.Run("prepare data, should be passed", func(t *testing.T) {
		resp, err := disco.RegisterService(getContext(), &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				AppId:       "govern_service_group",
				ServiceName: "govern_service_name",
				Version:     "3.0.0",
				Level:       "FRONT",
				Status:      pb.MS_UP,
			},
		})
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		serviceId = resp.ServiceId

		err = disco.PutSchema(getContext(), &pb.ModifySchemaRequest{
			ServiceId: serviceId,
			SchemaId:  "schemaId",
			Schema:    "detail",
		})
		assert.NoError(t, err)

		_, err = disco.RegisterInstance(getContext(), &pb.RegisterInstanceRequest{
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
	})

	t.Run("when get invalid service detail, should be failed", func(t *testing.T) {
		resp, err := govern.GetServiceDetail(getContext(), &pb.GetServiceRequest{
			ServiceId: "",
		})
		assert.True(t, errsvc.IsErrEqualCode(err, pb.ErrInvalidParams), err)
		assert.Nil(t, resp)
	})

	t.Run("when get a service detail, should be passed", func(t *testing.T) {
		respGetServiceDetail, err := govern.GetServiceDetail(getContext(), &pb.GetServiceRequest{
			ServiceId: serviceId,
		})
		assert.NoError(t, err)
		assert.NotNil(t, respGetServiceDetail)

		err = disco.UnregisterService(getContext(), &pb.DeleteServiceRequest{
			ServiceId: serviceId,
			Force:     true,
		})
		assert.NoError(t, err)

		respGetServiceDetail, err = govern.GetServiceDetail(getContext(), &pb.GetServiceRequest{
			ServiceId: serviceId,
		})
		assert.True(t, errsvc.IsErrEqualCode(err, pb.ErrServiceNotExists), err)
		assert.Nil(t, respGetServiceDetail)
	})
}

func TestNewServiceOverview(t *testing.T) {
	t.Run("no instances, should be ok", func(t *testing.T) {
		_, err := govern.NewServiceOverview(&pb.ServiceDetail{
			MicroService: &pb.MicroService{},
		}, nil)
		assert.NoError(t, err)

		_, err = govern.NewServiceOverview(&pb.ServiceDetail{
			MicroService: &pb.MicroService{},
			Instances:    []*pb.MicroServiceInstance{},
		}, nil)
		assert.NoError(t, err)
	})

	t.Run("has schema or service properties, should be ok", func(t *testing.T) {
		overview, err := govern.NewServiceOverview(&pb.ServiceDetail{
			MicroService: &pb.MicroService{
				Schemas:    []string{"test"},
				Properties: map[string]string{"test": "A"},
			},
		}, nil)
		assert.NoError(t, err)
		assert.Empty(t, overview.MicroService.Schemas)
		assert.NotEmpty(t, overview.MicroService.Properties)
	})

	t.Run("has instance properties, should be ok", func(t *testing.T) {
		overview, err := govern.NewServiceOverview(&pb.ServiceDetail{
			MicroService: &pb.MicroService{},
			Instances: []*pb.MicroServiceInstance{
				{
					Properties: map[string]string{"test": "A"},
				},
			},
		}, nil)
		assert.NoError(t, err)
		assert.Empty(t, overview.Instances[0].Properties)

		overview, err = govern.NewServiceOverview(&pb.ServiceDetail{
			MicroService: &pb.MicroService{},
			Instances: []*pb.MicroServiceInstance{
				{
					Properties: map[string]string{"test": "A"},
				},
			},
		}, map[string]string{"inner": "B"})
		assert.NoError(t, err)
		assert.Empty(t, overview.Instances[0].Properties)

		overview, err = govern.NewServiceOverview(&pb.ServiceDetail{
			MicroService: &pb.MicroService{},
			Instances: []*pb.MicroServiceInstance{
				{
					Properties: map[string]string{"test": "A", "inner": "C"},
				},
			},
		}, map[string]string{"inner": "B"})
		assert.NoError(t, err)
		assert.Equal(t, "C", overview.Instances[0].Properties["inner"])
	})
}
