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

package etcd

import (
	"github.com/apache/servicecomb-service-center/datasource"
	pb "github.com/apache/servicecomb-service-center/pkg/registry"
	"github.com/apache/servicecomb-service-center/server/core/proto"
	scerr "github.com/apache/servicecomb-service-center/server/scerror"
	"github.com/apache/servicecomb-service-center/server/service/event"
	"github.com/go-chassis/go-archaius"
	"github.com/stretchr/testify/assert"
	"strconv"
	"testing"
)

var deh event.DependencyEventHandler

func Test_Creat(t *testing.T) {

	datasource.Install("etcd", func(opts datasource.Options) (datasource.DataSource, error) {
		return NewDataSource(), nil
	})

	err := datasource.Init(datasource.Options{
		Endpoint:       "",
		PluginImplName: datasource.ImplName(archaius.GetString("servicecomb.ms.name", "etcd")),
	})
	assert.NoError(t, err)

	var (
		consumerId1 string
		consumerId2 string
		consumerId3 string
	)
	t.Run("should be passed", func(t *testing.T) {
		resp, err := serviceResource.Create(getContext(), &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				AppId:       "create_dep_group",
				ServiceName: "create_dep_consumer",
				Version:     "1.0.0",
				Level:       "FRONT",
				Status:      pb.MS_UP,
			},
		})
		assert.NotNil(t, resp)
		assert.NoError(t, err)
		assert.Equal(t, proto.Response_SUCCESS, resp.Response.GetCode())
		consumerId1 = resp.ServiceId

		resp, err = serviceResource.Create(getContext(), &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				AppId:       "create_dep_group",
				ServiceName: "create_dep_consumer_all",
				Version:     "1.0.0",
				Level:       "FRONT",
				Status:      pb.MS_UP,
			},
		})
		assert.NotNil(t, resp)
		assert.NoError(t, err)
		assert.Equal(t, proto.Response_SUCCESS, resp.Response.GetCode())
		consumerId3 = resp.ServiceId

		resp, err = serviceResource.Create(getContext(), &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				Environment: pb.ENV_PROD,
				AppId:       "create_dep_group",
				ServiceName: "create_dep_consumer",
				Version:     "1.0.0",
				Level:       "FRONT",
				Status:      pb.MS_UP,
			},
		})
		assert.NotNil(t, resp)
		assert.NoError(t, err)
		assert.Equal(t, proto.Response_SUCCESS, resp.Response.GetCode())
		consumerId2 = resp.ServiceId

		resp, err = serviceResource.Create(getContext(), &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				AppId:       "create_dep_group",
				ServiceName: "create_dep_provider",
				Version:     "1.0.0",
				Level:       "FRONT",
				Status:      pb.MS_UP,
			},
		})
		assert.NotNil(t, resp)
		assert.NoError(t, err)
		assert.Equal(t, proto.Response_SUCCESS, resp.Response.GetCode())

		resp, err = serviceResource.Create(getContext(), &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				AppId:       "create_dep_group",
				ServiceName: "create_dep_provider",
				Version:     "1.0.1",
				Level:       "FRONT",
				Status:      pb.MS_UP,
			},
		})
		assert.NotNil(t, resp)
		assert.NoError(t, err)
		assert.Equal(t, proto.Response_SUCCESS, resp.Response.GetCode())

		resp, err = serviceResource.Create(getContext(), &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				Environment: pb.ENV_PROD,
				AppId:       "create_dep_group",
				ServiceName: "create_dep_provider",
				Version:     "1.0.0",
				Level:       "FRONT",
				Status:      pb.MS_UP,
			},
		})
		assert.NotNil(t, resp)
		assert.NoError(t, err)
		assert.Equal(t, proto.Response_SUCCESS, resp.Response.GetCode())
	})

	t.Run("when request is invalid, should be failed", func(t *testing.T) {
		consumer := &pb.MicroServiceKey{
			AppId:       "create_dep_group",
			ServiceName: "create_dep_consumer",
			Version:     "1.0.0",
		}
		providers := []*pb.MicroServiceKey{
			{
				AppId:       "create_dep_group",
				ServiceName: "create_dep_provider",
				Version:     "1.0.0",
			},
		}

		// consumer does not exist
		resp, err := datasource.DependencyMgr().CreateDependency(getContext(), &pb.CreateDependenciesRequest{
			Dependencies: []*pb.ConsumerDependency{
				{
					Consumer: &pb.MicroServiceKey{
						AppId:       "noexistapp",
						ServiceName: "noexistservice",
						Version:     "1.0.0",
					},
					Providers: providers,
				},
			},
		})
		assert.NotNil(t, resp)
		assert.NoError(t, err)
		assert.Equal(t, scerr.ErrServiceNotExists, resp.Response.GetCode())

		// provider version is invalid
		resp, err = datasource.DependencyMgr().CreateDependency(getContext(), &pb.CreateDependenciesRequest{
			Dependencies: []*pb.ConsumerDependency{
				{
					Consumer: consumer,
					Providers: []*pb.MicroServiceKey{
						{
							AppId:       "create_dep_group",
							ServiceName: "create_dep_provider",
							Version:     "1.0.32768",
						},
					},
				},
			},
		})
		assert.NotNil(t, resp)
		assert.NoError(t, err)
		assert.Equal(t, scerr.ErrInvalidParams, resp.Response.GetCode())

		// consumer version is invalid
		resp, err = datasource.DependencyMgr().CreateDependency(getContext(), &pb.CreateDependenciesRequest{
			Dependencies: []*pb.ConsumerDependency{
				{
					Consumer: &pb.MicroServiceKey{
						AppId:       "create_dep_group",
						ServiceName: "create_dep_consumer",
						Version:     "1.0.0+",
					},
					Providers: providers,
				},
			},
		})
		assert.NotNil(t, resp)
		assert.NoError(t, err)
		assert.Equal(t, scerr.ErrInvalidParams, resp.Response.GetCode())

		resp, err = datasource.DependencyMgr().CreateDependency(getContext(), &pb.CreateDependenciesRequest{
			Dependencies: []*pb.ConsumerDependency{
				{
					Consumer: &pb.MicroServiceKey{
						AppId:       "create_dep_group",
						ServiceName: "create_dep_consumer",
						Version:     "1.0.0-1.0.1",
					},
					Providers: providers,
				},
			},
		})
		assert.NotNil(t, resp)
		assert.NoError(t, err)
		assert.Equal(t, scerr.ErrInvalidParams, resp.Response.GetCode())

		resp, err = datasource.DependencyMgr().CreateDependency(getContext(), &pb.CreateDependenciesRequest{
			Dependencies: []*pb.ConsumerDependency{
				{
					Consumer: &pb.MicroServiceKey{
						AppId:       "create_dep_group",
						ServiceName: "create_dep_consumer",
						Version:     "latest",
					},
					Providers: providers,
				},
			},
		})
		assert.NotNil(t, resp)
		assert.NoError(t, err)
		assert.Equal(t, scerr.ErrInvalidParams, resp.Response.GetCode())

		resp, err = datasource.DependencyMgr().CreateDependency(getContext(), &pb.CreateDependenciesRequest{
			Dependencies: []*pb.ConsumerDependency{
				{
					Consumer: &pb.MicroServiceKey{
						AppId:       "create_dep_group",
						ServiceName: "create_dep_consumer",
						Version:     "",
					},
					Providers: providers,
				},
			},
		})
		assert.NotNil(t, resp)
		assert.NoError(t, err)
		assert.Equal(t, scerr.ErrInvalidParams, resp.Response.GetCode())

		resp, err = datasource.DependencyMgr().CreateDependency(getContext(), &pb.CreateDependenciesRequest{
			Dependencies: []*pb.ConsumerDependency{
				{
					Consumer: &pb.MicroServiceKey{
						AppId:       "create_dep_group",
						ServiceName: "*",
						Version:     "1.0.0",
					},
					Providers: providers,
				},
			},
		})
		assert.NotNil(t, resp)
		assert.NoError(t, err)
		assert.Equal(t, scerr.ErrInvalidParams, resp.Response.GetCode())

		// provider app is invalid
		resp, err = datasource.DependencyMgr().CreateDependency(getContext(), &pb.CreateDependenciesRequest{
			Dependencies: []*pb.ConsumerDependency{
				{
					Consumer: consumer,
					Providers: []*pb.MicroServiceKey{
						{
							AppId:       "*",
							ServiceName: "service_name_provider",
							Version:     "2.0.0",
						},
					},
				},
			},
		})
		assert.NotNil(t, resp)
		assert.NoError(t, err)
		assert.Equal(t, scerr.ErrInvalidParams, resp.Response.GetCode())

		// provider serviceName is invalid
		resp, err = datasource.DependencyMgr().CreateDependency(getContext(), &pb.CreateDependenciesRequest{
			Dependencies: []*pb.ConsumerDependency{
				{
					Consumer: consumer,
					Providers: []*pb.MicroServiceKey{
						{
							AppId:       "service_group_provider",
							ServiceName: "-",
							Version:     "2.0.0",
						},
					},
				},
			},
		})
		assert.NotNil(t, resp)
		assert.NoError(t, err)
		assert.Equal(t, scerr.ErrInvalidParams, resp.Response.GetCode())

		// provider version is invalid
		resp, err = datasource.DependencyMgr().CreateDependency(getContext(), &pb.CreateDependenciesRequest{
			Dependencies: []*pb.ConsumerDependency{
				{
					Consumer: consumer,
					Providers: []*pb.MicroServiceKey{
						{
							AppId:       "service_group_provider",
							ServiceName: "service_name_provider",
							Version:     "",
						},
					},
				},
			},
		})
		assert.NotNil(t, resp)
		assert.NoError(t, err)
		assert.Equal(t, scerr.ErrInvalidParams, resp.Response.GetCode())

		// provider in diff env
		resp, err = datasource.DependencyMgr().CreateDependency(getContext(), &pb.CreateDependenciesRequest{
			Dependencies: []*pb.ConsumerDependency{
				{
					Consumer: consumer,
					Providers: []*pb.MicroServiceKey{
						{
							Environment: pb.ENV_PROD,
							AppId:       "service_group_provider",
							ServiceName: "service_name_provider",
							Version:     "latest",
						},
					},
				},
			},
		})
		assert.NotNil(t, resp)
		assert.NoError(t, err)
		assert.Equal(t, proto.Response_SUCCESS, resp.Response.GetCode())

		// consumer in diff env
		consumer.Environment = pb.ENV_PROD
		resp, err = datasource.DependencyMgr().CreateDependency(getContext(), &pb.CreateDependenciesRequest{
			Dependencies: []*pb.ConsumerDependency{
				{
					Consumer: consumer,
					Providers: []*pb.MicroServiceKey{
						{
							AppId:       "service_group_provider",
							ServiceName: "service_name_provider",
							Version:     "latest",
						},
					},
				},
			},
		})
		assert.NotNil(t, resp)
		assert.NoError(t, err)
		assert.Equal(t, proto.Response_SUCCESS, resp.Response.GetCode())

		assert.Equal(t, nil, deh.Handle())

		respCon, err := serviceResource.GetConsumerDependencies(getContext(), &pb.GetDependenciesRequest{
			ServiceId: consumerId1,
		})
		assert.NotNil(t, respCon)
		assert.NoError(t, err)
		assert.Equal(t, proto.Response_SUCCESS, respCon.Response.GetCode())
		assert.Equal(t, 0, len(respCon.Providers))

		respCon, err = serviceResource.GetConsumerDependencies(getContext(), &pb.GetDependenciesRequest{
			ServiceId: consumerId2,
		})
		assert.NotNil(t, respCon)
		assert.NoError(t, err)
		assert.Equal(t, proto.Response_SUCCESS, respCon.Response.GetCode())
		assert.Equal(t, 0, len(respCon.Providers))

		// dependencies is invalid
		var deps []*pb.ConsumerDependency
		for i := 0; i < 101; i++ {
			deps = append(deps, &pb.ConsumerDependency{
				Consumer: &pb.MicroServiceKey{
					AppId:       "create_dep_group",
					ServiceName: "create_dep_consumer" + strconv.Itoa(i),
					Version:     "1.0.0",
				},
				Providers: []*pb.MicroServiceKey{
					{
						AppId:       "service_group_provider",
						ServiceName: "service_name_provider",
						Version:     "latest",
					},
				},
			})
		}
		resp, err = serviceResource.CreateDependenciesForMicroServices(getContext(), &pb.CreateDependenciesRequest{
			Dependencies: deps,
		})
		assert.NotNil(t, resp)
		assert.NoError(t, err)
		assert.Equal(t, scerr.ErrInvalidParams, resp.Response.GetCode())
	})

	t.Run("when request is valid, should be passed", func(t *testing.T) {
		consumer := &pb.MicroServiceKey{
			ServiceName: "create_dep_consumer",
			AppId:       "create_dep_group",
			Version:     "1.0.0",
		}

		// add latest
		resp, err := datasource.DependencyMgr().CreateDependency(getContext(), &pb.CreateDependenciesRequest{
			Dependencies: []*pb.ConsumerDependency{
				{
					Consumer: consumer,
					Providers: []*pb.MicroServiceKey{
						{
							AppId:       "create_dep_group",
							ServiceName: "create_dep_provider",
							Version:     "latest",
						},
					},
				},
			},
		})
		assert.NotNil(t, resp)
		assert.NoError(t, err)
		assert.Equal(t, proto.Response_SUCCESS, resp.Response.GetCode())

		assert.Equal(t, nil, deh.Handle())

		respPro, err := datasource.DependencyMgr().SearchConsumerDependency(getContext(), &pb.GetDependenciesRequest{
			ServiceId: consumerId1,
		})
		assert.NotNil(t, respPro)
		assert.NoError(t, err)
		assert.Equal(t, proto.Response_SUCCESS, respPro.Response.GetCode())

		// add 1.0.0+
		resp, err = datasource.DependencyMgr().CreateDependency(getContext(), &pb.CreateDependenciesRequest{
			Dependencies: []*pb.ConsumerDependency{
				{
					Consumer: consumer,
					Providers: []*pb.MicroServiceKey{
						{
							AppId:       "create_dep_group",
							ServiceName: "create_dep_provider",
							Version:     "1.0.0+",
						},
					},
				},
			},
		})
		assert.NotNil(t, resp)
		assert.NoError(t, err)
		assert.Equal(t, proto.Response_SUCCESS, resp.Response.GetCode())

		respPro, err = datasource.DependencyMgr().SearchConsumerDependency(getContext(), &pb.GetDependenciesRequest{
			ServiceId: consumerId1,
		})
		assert.NotNil(t, respPro)
		assert.NoError(t, err)
		assert.Equal(t, proto.Response_SUCCESS, respPro.Response.GetCode())

		// add *
		resp, err = datasource.DependencyMgr().CreateDependency(getContext(), &pb.CreateDependenciesRequest{
			Dependencies: []*pb.ConsumerDependency{
				{
					Consumer: &pb.MicroServiceKey{
						ServiceName: "create_dep_consumer_all",
						AppId:       "create_dep_group",
						Version:     "1.0.0",
					},
					Providers: []*pb.MicroServiceKey{
						{
							ServiceName: "*",
						},
					},
				},
			},
		})
		assert.NotNil(t, resp)
		assert.NoError(t, err)
		assert.Equal(t, proto.Response_SUCCESS, resp.Response.GetCode())

		respPro, err = datasource.DependencyMgr().SearchConsumerDependency(getContext(), &pb.GetDependenciesRequest{
			ServiceId: consumerId3,
		})
		assert.NotNil(t, respPro)
		assert.NoError(t, err)
		assert.Equal(t, proto.Response_SUCCESS, respPro.Response.GetCode())
		assert.Equal(t, 0, len(respPro.Providers))

		// clean all
		resp, err = datasource.DependencyMgr().CreateDependency(getContext(), &pb.CreateDependenciesRequest{
			Dependencies: []*pb.ConsumerDependency{
				{
					Consumer: &pb.MicroServiceKey{
						ServiceName: "create_dep_consumer_all",
						AppId:       "create_dep_group",
						Version:     "1.0.0",
					},
					Providers: nil,
				},
			},
		})
		assert.NotNil(t, resp)
		assert.NoError(t, err)
		assert.Equal(t, proto.Response_SUCCESS, resp.Response.GetCode())

		// add multiple providers
		resp, err = datasource.DependencyMgr().CreateDependency(getContext(), &pb.CreateDependenciesRequest{
			Dependencies: []*pb.ConsumerDependency{
				{
					Consumer: consumer,
					Providers: []*pb.MicroServiceKey{
						{
							AppId:       "create_dep_group",
							ServiceName: "create_dep_provider",
							Version:     "1.0.0",
						},
						{
							ServiceName: "*",
						},
					},
				},
			},
		})
		assert.NotNil(t, resp)
		assert.NoError(t, err)
		assert.Equal(t, proto.Response_SUCCESS, resp.Response.GetCode())

		// add 1.0.0-2.0.0 to override *
		resp, err = datasource.DependencyMgr().CreateDependency(getContext(), &pb.CreateDependenciesRequest{
			Dependencies: []*pb.ConsumerDependency{
				{
					Consumer: consumer,
					Providers: []*pb.MicroServiceKey{
						{
							AppId:       "create_dep_group",
							ServiceName: "create_dep_provider",
							Version:     "1.0.0-1.0.1",
						},
					},
				},
			},
		})
		assert.NotNil(t, resp)
		assert.NoError(t, err)
		assert.Equal(t, proto.Response_SUCCESS, resp.Response.GetCode())

		assert.Equal(t, nil, deh.Handle())

		respPro, err = datasource.DependencyMgr().SearchConsumerDependency(getContext(), &pb.GetDependenciesRequest{
			ServiceId: consumerId1,
		})
		assert.NotNil(t, respPro)
		assert.NoError(t, err)
		assert.Equal(t, proto.Response_SUCCESS, respPro.Response.GetCode())
		assert.Equal(t, "1.0.0", respPro.Providers[0].Version)

		// add not override
		respAdd, err := datasource.DependencyMgr().AddDependency(getContext(), &pb.AddDependenciesRequest{
			Dependencies: []*pb.ConsumerDependency{
				{
					Consumer: consumer,
					Providers: []*pb.MicroServiceKey{
						{
							AppId:       "create_dep_group",
							ServiceName: "create_dep_provider",
							Version:     "1.0.0-3.0.0",
						},
					},
				},
			},
		})
		assert.NotNil(t, respAdd)
		assert.NoError(t, err)
		assert.Equal(t, proto.Response_SUCCESS, respAdd.Response.GetCode())

		assert.Equal(t, nil, deh.Handle())

		respPro, err = datasource.DependencyMgr().SearchConsumerDependency(getContext(), &pb.GetDependenciesRequest{
			ServiceId: consumerId1,
		})
		assert.NotNil(t, respPro)
		assert.NoError(t, err)
		assert.Equal(t, proto.Response_SUCCESS, respPro.Response.GetCode())

		// add provider is empty
		resp, err = datasource.DependencyMgr().CreateDependency(getContext(), &pb.CreateDependenciesRequest{
			Dependencies: []*pb.ConsumerDependency{
				{
					Consumer:  consumer,
					Providers: []*pb.MicroServiceKey{},
				},
			},
		})
		assert.NotNil(t, resp)
		assert.NoError(t, err)
		assert.Equal(t, proto.Response_SUCCESS, resp.Response.GetCode())

		resp, err = datasource.DependencyMgr().CreateDependency(getContext(), &pb.CreateDependenciesRequest{
			Dependencies: []*pb.ConsumerDependency{
				{
					Consumer: consumer,
				},
			},
		})
		assert.NotNil(t, resp)
		assert.NoError(t, err)
		assert.Equal(t, proto.Response_SUCCESS, resp.Response.GetCode())

		assert.Equal(t, nil, deh.Handle())

		respPro, err = datasource.DependencyMgr().SearchConsumerDependency(getContext(), &pb.GetDependenciesRequest{
			ServiceId: consumerId1,
		})
		assert.NotNil(t, respPro)
		assert.NoError(t, err)
		assert.Equal(t, proto.Response_SUCCESS, respPro.Response.GetCode())
		//assert.Equal(t, 0, len(respPro.Providers))

	})
}

func Test_Get(t *testing.T) {
	datasource.Install("etcd", func(opts datasource.Options) (datasource.DataSource, error) {
		return NewDataSource(), nil
	})

	err := datasource.Init(datasource.Options{
		Endpoint:       "",
		PluginImplName: datasource.ImplName(archaius.GetString("servicecomb.ms.name", "etcd")),
	})
	assert.NoError(t, err)

	var (
		consumerId1 string
		providerId1 string
		providerId2 string
	)

	t.Run("should be passed", func(t *testing.T) {
		resp, err := serviceResource.Create(getContext(), &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				AppId:       "get_dep_group",
				ServiceName: "get_dep_consumer",
				Version:     "1.0.0",
				Level:       "FRONT",
				Status:      pb.MS_UP,
			},
		})
		assert.NotNil(t, resp)
		assert.NoError(t, err)
		assert.Equal(t, proto.Response_SUCCESS, resp.Response.GetCode())
		consumerId1 = resp.ServiceId

		resp, err = serviceResource.Create(getContext(), &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				AppId:       "get_dep_group",
				ServiceName: "get_dep_provider",
				Version:     "1.0.0",
				Level:       "FRONT",
				Status:      pb.MS_UP,
			},
		})
		assert.NotNil(t, resp)
		assert.NoError(t, err)
		assert.Equal(t, proto.Response_SUCCESS, resp.Response.GetCode())
		providerId1 = resp.ServiceId

		resp, err = serviceResource.Create(getContext(), &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				AppId:       "get_dep_group",
				ServiceName: "get_dep_provider",
				Version:     "2.0.0",
				Level:       "FRONT",
				Status:      pb.MS_UP,
			},
		})
		assert.NotNil(t, resp)
		assert.NoError(t, err)
		assert.Equal(t, proto.Response_SUCCESS, resp.Response.GetCode())
		providerId2 = resp.ServiceId
	})

	t.Run("when request is invalid, should be failed", func(t *testing.T) {
		//service id is empty when get provider
		resp, err := datasource.DependencyMgr().SearchProviderDependency(getContext(), &pb.GetDependenciesRequest{
			ServiceId: "",
		})
		assert.NotNil(t, resp)
		assert.NoError(t, err)
		assert.Equal(t, scerr.ErrInvalidParams, resp.Response.GetCode())

		//service does not exist when get provider
		resp, err = datasource.DependencyMgr().SearchProviderDependency(getContext(), &pb.GetDependenciesRequest{
			ServiceId: "noneservice",
		})
		assert.NotNil(t, resp)
		assert.NoError(t, err)
		assert.Equal(t, scerr.ErrServiceNotExists, resp.Response.GetCode())

		//service id is empty when get consumer
		resp, err = datasource.DependencyMgr().SearchProviderDependency(getContext(), &pb.GetDependenciesRequest{
			ServiceId: "",
		})
		assert.NotNil(t, resp)
		assert.NoError(t, err)
		assert.Equal(t, scerr.ErrInvalidParams, resp.Response.GetCode())

		//service does not exist when get consumer
		resp, err = datasource.DependencyMgr().SearchProviderDependency(getContext(), &pb.GetDependenciesRequest{
			ServiceId: "noneservice",
		})
		assert.NotNil(t, resp)
		assert.NoError(t, err)
		assert.Equal(t, scerr.ErrServiceNotExists, resp.Response.GetCode())
	})

	t.Run("when request is valid, should be passed", func(t *testing.T) {
		//get provider
		resp, err := datasource.DependencyMgr().SearchProviderDependency(getContext(), &pb.GetDependenciesRequest{
			ServiceId: providerId1,
		})
		assert.NotNil(t, resp)
		assert.NoError(t, err)
		assert.Equal(t, proto.Response_SUCCESS, resp.Response.GetCode())

		//get consumer
		resp, err = datasource.DependencyMgr().SearchProviderDependency(getContext(), &pb.GetDependenciesRequest{
			ServiceId: consumerId1,
		})
		assert.NotNil(t, resp)
		assert.NoError(t, err)
		assert.Equal(t, proto.Response_SUCCESS, resp.Response.GetCode())
	})

	t.Run("when after finding instance, should created dependencies between C and P", func(t *testing.T) {
		// find provider
		resp, err := instanceResource.Find(getContext(), &pb.FindInstancesRequest{
			ConsumerServiceId: consumerId1,
			AppId:             "get_dep_group",
			ServiceName:       "get_dep_provider",
			VersionRule:       "1.0.0+",
		})
		assert.NotNil(t, resp)
		assert.NoError(t, err)
		assert.Equal(t, proto.Response_SUCCESS, resp.Response.GetCode())

		assert.Equal(t, nil, deh.Handle())

		// get consumer's deps
		respGetP, err := datasource.DependencyMgr().SearchProviderDependency(getContext(), &pb.GetDependenciesRequest{
			ServiceId: providerId1,
		})
		assert.NotNil(t, respGetP)
		assert.NoError(t, err)
		assert.Equal(t, proto.Response_SUCCESS, respGetP.Response.GetCode())

		// get provider's deps
		respGetC, err := serviceResource.GetConsumerDependencies(getContext(), &pb.GetDependenciesRequest{
			ServiceId: consumerId1,
		})
		assert.NotNil(t, respGetC)
		assert.NoError(t, err)
		assert.Equal(t, proto.Response_SUCCESS, respGetC.Response.GetCode())

		// get self deps
		resp, err = instanceResource.Find(getContext(), &pb.FindInstancesRequest{
			ConsumerServiceId: consumerId1,
			AppId:             "get_dep_group",
			ServiceName:       "get_dep_consumer",
			VersionRule:       "1.0.0+",
		})
		assert.NotNil(t, resp)
		assert.NoError(t, err)
		assert.Equal(t, proto.Response_SUCCESS, resp.Response.GetCode())

		assert.Equal(t, nil, deh.Handle())

		respGetC, err = datasource.DependencyMgr().SearchConsumerDependency(getContext(), &pb.GetDependenciesRequest{
			ServiceId: consumerId1,
			NoSelf:    true,
		})
		assert.NotNil(t, respGetC)
		assert.NoError(t, err)
		assert.Equal(t, proto.Response_SUCCESS, respGetC.Response.GetCode())

		// find before provider register
		resp, err = instanceResource.Find(getContext(), &pb.FindInstancesRequest{
			ConsumerServiceId: providerId2,
			AppId:             "get_dep_group",
			ServiceName:       "get_dep_finder",
			VersionRule:       "1.0.0+",
		})
		assert.NotNil(t, resp)
		assert.NoError(t, err)
		assert.Equal(t, proto.Response_SUCCESS, resp.Response.GetCode())

		respCreateF, err := serviceResource.Create(getContext(), &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				AppId:       "get_dep_group",
				ServiceName: "get_dep_finder",
				Version:     "1.0.0",
				Level:       "FRONT",
				Status:      pb.MS_UP,
			},
		})
		assert.NotNil(t, respCreateF)
		assert.NoError(t, err)
		assert.Equal(t, proto.Response_SUCCESS, respCreateF.Response.GetCode())
		finder1 := respCreateF.ServiceId

		resp, err = instanceResource.Find(getContext(), &pb.FindInstancesRequest{
			ConsumerServiceId: providerId2,
			AppId:             "get_dep_group",
			ServiceName:       "get_dep_finder",
			VersionRule:       "1.0.0+",
		})
		assert.NotNil(t, resp)
		assert.NoError(t, err)
		assert.Equal(t, proto.Response_SUCCESS, resp.Response.GetCode())

		assert.Equal(t, nil, deh.Handle())

		respGetC, err = datasource.DependencyMgr().SearchConsumerDependency(getContext(), &pb.GetDependenciesRequest{
			ServiceId: providerId2,
		})
		assert.NotNil(t, respGetC)
		assert.NoError(t, err)
		assert.Equal(t, proto.Response_SUCCESS, respGetC.Response.GetCode())
		assert.Equal(t, 1, len(respGetC.Providers))
		assert.Equal(t, finder1, respGetC.Providers[0].ServiceId)

		// find after delete micro service
		respDelP, err := serviceResource.Delete(getContext(), &pb.DeleteServiceRequest{
			ServiceId: finder1, Force: true,
		})
		assert.NotNil(t, respDelP)
		assert.NoError(t, err)
		assert.Equal(t, proto.Response_SUCCESS, respDelP.Response.GetCode())

		assert.Equal(t, nil, deh.Handle())

		respGetC, err = datasource.DependencyMgr().SearchConsumerDependency(getContext(), &pb.GetDependenciesRequest{
			ServiceId: providerId2,
		})
		assert.NotNil(t, respGetC)
		assert.NoError(t, err)
		assert.Equal(t, proto.Response_SUCCESS, respGetC.Response.GetCode())
		assert.Equal(t, 0, len(respGetC.Providers))

		respCreateF, err = serviceResource.Create(getContext(), &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				ServiceId:   finder1,
				AppId:       "get_dep_group",
				ServiceName: "get_dep_finder",
				Version:     "1.0.0",
				Level:       "FRONT",
				Status:      pb.MS_UP,
			},
		})
		assert.NotNil(t, respCreateF)
		assert.NoError(t, err)
		assert.Equal(t, proto.Response_SUCCESS, respCreateF.Response.GetCode())

		resp, err = instanceResource.Find(getContext(), &pb.FindInstancesRequest{
			ConsumerServiceId: providerId2,
			AppId:             "get_dep_group",
			ServiceName:       "get_dep_finder",
			VersionRule:       "1.0.0+",
		})
		assert.NotNil(t, resp)
		assert.NoError(t, err)
		assert.Equal(t, proto.Response_SUCCESS, resp.Response.GetCode())

		assert.Equal(t, nil, deh.Handle())

		respGetC, err = datasource.DependencyMgr().SearchConsumerDependency(getContext(), &pb.GetDependenciesRequest{
			ServiceId: providerId2,
		})
		assert.NotNil(t, respGetC)
		assert.NoError(t, err)
		assert.Equal(t, proto.Response_SUCCESS, respGetC.Response.GetCode())
		assert.Equal(t, 1, len(respGetC.Providers))
		assert.Equal(t, finder1, respGetC.Providers[0].ServiceId)
	})
}
