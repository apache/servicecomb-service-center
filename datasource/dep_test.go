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
	"testing"

	"github.com/apache/servicecomb-service-center/datasource"
	"github.com/apache/servicecomb-service-center/pkg/util"
	pb "github.com/go-chassis/cari/discovery"
	"github.com/stretchr/testify/assert"
)

func Test_Get(t *testing.T) {
	var (
		consumerId1 string
		providerId1 string
		providerId2 string
	)

	t.Run("should be passed", func(t *testing.T) {
		resp, err := datasource.GetMetadataManager().RegisterService(depGetContext(), &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				AppId:       "dep_get_dep_group",
				ServiceName: "dep_get_dep_consumer",
				Version:     "1.0.0",
				Level:       "FRONT",
				Status:      pb.MS_UP,
			},
		})
		assert.NotNil(t, resp)
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, resp.Response.GetCode())
		consumerId1 = resp.ServiceId

		resp, err = datasource.GetMetadataManager().RegisterService(depGetContext(), &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				AppId:       "dep_get_dep_group",
				ServiceName: "dep_get_dep_provider",
				Version:     "1.0.0",
				Level:       "FRONT",
				Status:      pb.MS_UP,
			},
		})
		assert.NotNil(t, resp)
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, resp.Response.GetCode())
		providerId1 = resp.ServiceId

		resp, err = datasource.GetMetadataManager().RegisterService(depGetContext(), &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				AppId:       "dep_get_dep_group",
				ServiceName: "dep_get_dep_provider",
				Version:     "2.0.0",
				Level:       "FRONT",
				Status:      pb.MS_UP,
			},
		})
		assert.NotNil(t, resp)
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, resp.Response.GetCode())
		providerId2 = resp.ServiceId
	})

	t.Run("when request is valid, should be passed", func(t *testing.T) {
		//get provider
		resp, err := datasource.GetDependencyManager().SearchProviderDependency(depGetContext(), &pb.GetDependenciesRequest{
			ServiceId: providerId1,
		})
		assert.NotNil(t, resp)
		assert.NoError(t, err)

		//get consumer
		resp, err = datasource.GetDependencyManager().SearchProviderDependency(depGetContext(), &pb.GetDependenciesRequest{
			ServiceId: consumerId1,
		})
		assert.NotNil(t, resp)
		assert.NoError(t, err)
	})

	t.Run("when after finding instance, should created dependencies between C and P", func(t *testing.T) {
		// find provider
		resp, err := datasource.GetMetadataManager().FindInstances(depGetContext(), &pb.FindInstancesRequest{
			ConsumerServiceId: consumerId1,
			AppId:             "dep_get_dep_group",
			ServiceName:       "dep_get_dep_provider",
			VersionRule:       "1.0.0+",
		})
		assert.NotNil(t, resp)
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, resp.Response.GetCode())

		err = datasource.GetDependencyManager().DependencyHandle(getContext())
		assert.NoError(t, err)

		// get consumer's deps
		respGetP, err := datasource.GetDependencyManager().SearchProviderDependency(depGetContext(), &pb.GetDependenciesRequest{
			ServiceId: providerId1,
		})
		assert.NotNil(t, respGetP)
		assert.NoError(t, err)

		// get provider's deps
		respGetC, err := datasource.GetDependencyManager().SearchConsumerDependency(depGetContext(), &pb.GetDependenciesRequest{
			ServiceId: consumerId1,
		})
		assert.NotNil(t, respGetC)
		assert.NoError(t, err)

		// get self deps
		resp, err = datasource.GetMetadataManager().FindInstances(depGetContext(), &pb.FindInstancesRequest{
			ConsumerServiceId: consumerId1,
			AppId:             "dep_get_dep_group",
			ServiceName:       "dep_get_dep_consumer",
			VersionRule:       "1.0.0+",
		})
		assert.NotNil(t, resp)
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, resp.Response.GetCode())

		err = datasource.GetDependencyManager().DependencyHandle(getContext())
		assert.NoError(t, err)

		respGetC, err = datasource.GetDependencyManager().SearchConsumerDependency(depGetContext(), &pb.GetDependenciesRequest{
			ServiceId: consumerId1,
			NoSelf:    true,
		})
		assert.NotNil(t, respGetC)
		assert.NoError(t, err)

		// find before provider register
		resp, err = datasource.GetMetadataManager().FindInstances(depGetContext(), &pb.FindInstancesRequest{
			ConsumerServiceId: providerId2,
			AppId:             "dep_get_dep_group",
			ServiceName:       "dep_get_dep_finder",
			VersionRule:       "1.0.0+",
		})
		assert.NotNil(t, resp)
		assert.NoError(t, err)
		assert.Equal(t, pb.ErrServiceNotExists, resp.Response.GetCode())

		respCreateF, err := datasource.GetMetadataManager().RegisterService(depGetContext(), &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				AppId:       "dep_get_dep_group",
				ServiceName: "dep_get_dep_finder",
				Version:     "1.0.0",
				Level:       "FRONT",
				Status:      pb.MS_UP,
			},
		})
		assert.NotNil(t, respCreateF)
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respCreateF.Response.GetCode())
		finder1 := respCreateF.ServiceId

		resp, err = datasource.GetMetadataManager().FindInstances(depGetContext(), &pb.FindInstancesRequest{
			ConsumerServiceId: providerId2,
			AppId:             "dep_get_dep_group",
			ServiceName:       "dep_get_dep_finder",
			VersionRule:       "1.0.0+",
		})
		assert.NotNil(t, resp)
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, resp.Response.GetCode())

		err = datasource.GetDependencyManager().DependencyHandle(getContext())
		assert.NoError(t, err)

		respGetC, err = datasource.GetDependencyManager().SearchConsumerDependency(depGetContext(), &pb.GetDependenciesRequest{
			ServiceId: providerId2,
		})
		assert.NotNil(t, respGetC)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(respGetC.Providers))
		assert.Equal(t, finder1, respGetC.Providers[0].ServiceId)

		// find after delete micro service
		respDelP, err := datasource.GetMetadataManager().UnregisterService(depGetContext(), &pb.DeleteServiceRequest{
			ServiceId: finder1, Force: true,
		})
		assert.NotNil(t, respDelP)
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respDelP.Response.GetCode())

		err = datasource.GetDependencyManager().DependencyHandle(getContext())
		assert.NoError(t, err)

		respGetC, err = datasource.GetDependencyManager().SearchConsumerDependency(depGetContext(), &pb.GetDependenciesRequest{
			ServiceId: providerId2,
		})
		assert.NotNil(t, respGetC)
		assert.NoError(t, err)
		assert.Equal(t, 0, len(respGetC.Providers))

		respCreateF, err = datasource.GetMetadataManager().RegisterService(depGetContext(), &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				ServiceId:   finder1,
				AppId:       "dep_get_dep_group",
				ServiceName: "dep_get_dep_finder",
				Version:     "1.0.0",
				Level:       "FRONT",
				Status:      pb.MS_UP,
			},
		})
		assert.NotNil(t, respCreateF)
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respCreateF.Response.GetCode())

		resp, err = datasource.GetMetadataManager().FindInstances(depGetContext(), &pb.FindInstancesRequest{
			ConsumerServiceId: providerId2,
			AppId:             "dep_get_dep_group",
			ServiceName:       "dep_get_dep_finder",
			VersionRule:       "1.0.0+",
		})
		assert.NotNil(t, resp)
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, resp.Response.GetCode())

		err = datasource.GetDependencyManager().DependencyHandle(getContext())
		assert.NoError(t, err)

		respGetC, err = datasource.GetDependencyManager().SearchConsumerDependency(depGetContext(), &pb.GetDependenciesRequest{
			ServiceId: providerId2,
		})
		assert.NotNil(t, respGetC)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(respGetC.Providers))
		assert.Equal(t, finder1, respGetC.Providers[0].ServiceId)
	})
}

func depGetContext() context.Context {
	return util.WithNoCache(util.SetDomainProject(context.Background(), "new_default", "new_default"))
}
