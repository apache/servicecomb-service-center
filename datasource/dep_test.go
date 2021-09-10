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

func Test_Create(t *testing.T) {
	var (
		consumerId1 string
		consumerId2 string
	)
	resp, err := datasource.GetMetadataManager().RegisterService(depGetContext(), &pb.CreateServiceRequest{
		Service: &pb.MicroService{
			AppId:       "dep_create_dep_group",
			ServiceName: "dep_create_dep_consumer",
			Version:     "1.0.0",
			Level:       "FRONT",
			Status:      pb.MS_UP,
		},
	})
	assert.NotNil(t, resp)
	assert.NoError(t, err)
	assert.Equal(t, pb.ResponseSuccess, resp.Response.GetCode())
	consumerId1 = resp.ServiceId
	defer datasource.GetMetadataManager().UnregisterService(depGetContext(), &pb.DeleteServiceRequest{ServiceId: resp.ServiceId, Force: true})

	resp, err = datasource.GetMetadataManager().RegisterService(depGetContext(), &pb.CreateServiceRequest{
		Service: &pb.MicroService{
			Environment: pb.ENV_PROD,
			AppId:       "dep_create_dep_group",
			ServiceName: "dep_create_dep_consumer",
			Version:     "1.0.0",
			Level:       "FRONT",
			Status:      pb.MS_UP,
		},
	})
	assert.NotNil(t, resp)
	assert.NoError(t, err)
	assert.Equal(t, pb.ResponseSuccess, resp.Response.GetCode())
	consumerId2 = resp.ServiceId
	defer datasource.GetMetadataManager().UnregisterService(depGetContext(), &pb.DeleteServiceRequest{ServiceId: resp.ServiceId, Force: true})

	resp, err = datasource.GetMetadataManager().RegisterService(depGetContext(), &pb.CreateServiceRequest{
		Service: &pb.MicroService{
			AppId:       "dep_create_dep_group",
			ServiceName: "dep_create_dep_provider",
			Version:     "1.0.0",
			Level:       "FRONT",
			Status:      pb.MS_UP,
		},
	})
	assert.NotNil(t, resp)
	assert.NoError(t, err)
	assert.Equal(t, pb.ResponseSuccess, resp.Response.GetCode())
	defer datasource.GetMetadataManager().UnregisterService(depGetContext(), &pb.DeleteServiceRequest{ServiceId: resp.ServiceId, Force: true})

	resp, err = datasource.GetMetadataManager().RegisterService(depGetContext(), &pb.CreateServiceRequest{
		Service: &pb.MicroService{
			AppId:       "dep_create_dep_group",
			ServiceName: "dep_create_dep_provider",
			Version:     "1.0.1",
			Level:       "FRONT",
			Status:      pb.MS_UP,
		},
	})
	assert.NotNil(t, resp)
	assert.NoError(t, err)
	assert.Equal(t, pb.ResponseSuccess, resp.Response.GetCode())
	defer datasource.GetMetadataManager().UnregisterService(depGetContext(), &pb.DeleteServiceRequest{ServiceId: resp.ServiceId, Force: true})

	resp, err = datasource.GetMetadataManager().RegisterService(depGetContext(), &pb.CreateServiceRequest{
		Service: &pb.MicroService{
			AppId:       "dep_create_dep_group",
			ServiceName: "dep_create_dep_provider_other",
			Version:     "1.0.0",
			Level:       "FRONT",
			Status:      pb.MS_UP,
		},
	})
	assert.NotNil(t, resp)
	assert.NoError(t, err)
	assert.Equal(t, pb.ResponseSuccess, resp.Response.GetCode())
	defer datasource.GetMetadataManager().UnregisterService(depGetContext(), &pb.DeleteServiceRequest{ServiceId: resp.ServiceId, Force: true})

	resp, err = datasource.GetMetadataManager().RegisterService(depGetContext(), &pb.CreateServiceRequest{
		Service: &pb.MicroService{
			Environment: pb.ENV_PROD,
			AppId:       "dep_create_dep_group",
			ServiceName: "dep_create_dep_provider",
			Version:     "1.0.0",
			Level:       "FRONT",
			Status:      pb.MS_UP,
		},
	})
	assert.NotNil(t, resp)
	assert.NoError(t, err)
	assert.Equal(t, pb.ResponseSuccess, resp.Response.GetCode())
	defer datasource.GetMetadataManager().UnregisterService(depGetContext(), &pb.DeleteServiceRequest{ServiceId: resp.ServiceId, Force: true})

	t.Run("add dep and search when request is invalid, should be failed", func(t *testing.T) {
		consumer := &pb.MicroServiceKey{
			AppId:       "dep_create_dep_group",
			ServiceName: "dep_create_dep_consumer",
			Version:     "1.0.0",
		}
		providers := []*pb.MicroServiceKey{
			{
				AppId:       "dep_create_dep_group",
				ServiceName: "dep_create_dep_provider",
			},
		}

		// consumer does not exist
		resp, err := datasource.GetDependencyManager().AddOrUpdateDependencies(depGetContext(), []*pb.ConsumerDependency{
			{
				Consumer: &pb.MicroServiceKey{
					AppId:       "noexistapp",
					ServiceName: "noexistservice",
					Version:     "1.0.0",
				},
				Providers: providers,
			},
		}, false)
		assert.NotNil(t, resp)
		assert.NoError(t, err)
		assert.Equal(t, pb.ErrServiceNotExists, resp.GetCode())

		err = datasource.GetDependencyManager().DependencyHandle(getContext())
		assert.NoError(t, err)

		// provider in diff env
		resp, err = datasource.GetDependencyManager().AddOrUpdateDependencies(depGetContext(), []*pb.ConsumerDependency{
			{
				Consumer: consumer,
				Providers: []*pb.MicroServiceKey{
					{
						Environment: pb.ENV_PROD,
						AppId:       "dep_service_group_provider",
						ServiceName: "dep_service_name_provider",
					},
				},
			},
		}, false)
		assert.NotNil(t, resp)
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, resp.GetCode())

		//consumer in diff env
		consumer.Environment = pb.ENV_PROD
		resp, err = datasource.GetDependencyManager().AddOrUpdateDependencies(depGetContext(), []*pb.ConsumerDependency{
			{
				Consumer: consumer,
				Providers: []*pb.MicroServiceKey{
					{
						AppId:       "dep_service_group_provider",
						ServiceName: "dep_service_name_provider",
					},
				},
			},
		}, false)
		assert.NotNil(t, resp)
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, resp.GetCode())

		err = datasource.GetDependencyManager().DependencyHandle(getContext())
		assert.NoError(t, err)

		respCon, err := datasource.GetDependencyManager().SearchConsumerDependency(depGetContext(), &pb.GetDependenciesRequest{
			ServiceId: consumerId1,
		})
		assert.NotNil(t, respCon)
		assert.NoError(t, err)
		assert.Equal(t, 0, len(respCon.Providers))

		respCon, err = datasource.GetDependencyManager().SearchConsumerDependency(depGetContext(), &pb.GetDependenciesRequest{
			ServiceId: consumerId2,
		})
		assert.NotNil(t, respCon)
		assert.NoError(t, err)
		assert.Equal(t, 0, len(respCon.Providers))
	})

	t.Run("add dep and search when request is valid, should be passed", func(t *testing.T) {
		consumer := &pb.MicroServiceKey{
			ServiceName: "dep_create_dep_consumer",
			AppId:       "dep_create_dep_group",
			Version:     "1.0.0",
		}

		resp, err := datasource.GetDependencyManager().AddOrUpdateDependencies(depGetContext(), []*pb.ConsumerDependency{
			{
				Consumer: consumer,
				Providers: []*pb.MicroServiceKey{
					{
						AppId:       "dep_create_dep_group",
						ServiceName: "dep_create_dep_provider",
					},
				},
			},
		}, false)
		assert.NotNil(t, resp)
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, resp.GetCode())

		err = datasource.GetDependencyManager().DependencyHandle(getContext())
		assert.NoError(t, err)

		respPro, err := datasource.GetDependencyManager().SearchConsumerDependency(depGetContext(), &pb.GetDependenciesRequest{
			ServiceId: consumerId1,
		})
		assert.NotNil(t, respPro)
		assert.NoError(t, err)
		assert.Equal(t, 2, len(respPro.Providers))

		// add multiple providers
		resp, err = datasource.GetDependencyManager().AddOrUpdateDependencies(depGetContext(), []*pb.ConsumerDependency{
			{
				Consumer: consumer,
				Providers: []*pb.MicroServiceKey{
					{
						AppId:       "dep_create_dep_group",
						ServiceName: "dep_create_dep_provider",
					},
					{
						AppId:       "dep_create_dep_group",
						ServiceName: "dep_create_dep_provider_other",
					},
				},
			},
		}, false)
		assert.NotNil(t, resp)
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, resp.GetCode())

		err = datasource.GetDependencyManager().DependencyHandle(getContext())
		assert.NoError(t, err)

		respPro, err = datasource.GetDependencyManager().SearchConsumerDependency(depGetContext(), &pb.GetDependenciesRequest{
			ServiceId: consumerId1,
		})
		assert.NotNil(t, respPro)
		assert.NoError(t, err)
		assert.Equal(t, 3, len(respPro.Providers))

		// override
		resp, err = datasource.GetDependencyManager().AddOrUpdateDependencies(depGetContext(), []*pb.ConsumerDependency{
			{
				Consumer: consumer,
				Providers: []*pb.MicroServiceKey{
					{
						AppId:       "dep_create_dep_group",
						ServiceName: "dep_create_dep_provider",
					},
				},
			},
		}, true)
		assert.NotNil(t, resp)
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, resp.GetCode())
		err = datasource.GetDependencyManager().DependencyHandle(getContext())
		assert.NoError(t, err)

		respPro, err = datasource.GetDependencyManager().SearchConsumerDependency(depGetContext(), &pb.GetDependenciesRequest{
			ServiceId: consumerId1,
		})
		assert.NotNil(t, respPro)
		assert.NoError(t, err)
		assert.Equal(t, 2, len(respPro.Providers))

		// clean all
		resp, err = datasource.GetDependencyManager().AddOrUpdateDependencies(depGetContext(), []*pb.ConsumerDependency{
			{
				Consumer:  consumer,
				Providers: nil,
			},
		}, true)
		assert.NotNil(t, resp)
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, resp.GetCode())

		err = datasource.GetDependencyManager().DependencyHandle(getContext())
		assert.NoError(t, err)

		respPro, err = datasource.GetDependencyManager().SearchConsumerDependency(depGetContext(), &pb.GetDependenciesRequest{
			ServiceId: consumerId1,
		})
		assert.NotNil(t, respPro)
		assert.NoError(t, err)
		assert.Equal(t, 0, len(respPro.Providers))
	})
}

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

func TestParamsChecker(t *testing.T) {
	p := datasource.ParamsChecker(nil, nil)
	assert.Nil(t, p)

	p = datasource.ParamsChecker(&pb.MicroServiceKey{
		AppId:       "a",
		ServiceName: "b",
		Version:     "1.0.0",
	}, nil)
	assert.Nil(t, p)

	producer := &pb.MicroServiceKey{ServiceName: "*"}
	p = datasource.ParamsChecker(&pb.MicroServiceKey{
		AppId:       "a",
		ServiceName: "b",
		Version:     "1.0.0",
	}, []*pb.MicroServiceKey{
		producer,
	})
	assert.Nil(t, p)
	assert.Equal(t, "a", producer.AppId)
	assert.Equal(t, "0.0.0+", producer.Version)

	producer.ServiceName = "a"
	producer.Version = "1"
	p = datasource.ParamsChecker(&pb.MicroServiceKey{
		AppId:       "a",
		ServiceName: "b",
		Version:     "1.0.0",
	}, []*pb.MicroServiceKey{
		producer,
	})
	assert.Nil(t, p)
	assert.Equal(t, "0.0.0+", producer.Version)

	p = datasource.ParamsChecker(&pb.MicroServiceKey{
		AppId:       "a",
		ServiceName: "b",
		Version:     "1.0.0",
	}, []*pb.MicroServiceKey{
		{},
	})
	assert.NotNil(t, p)

	p = datasource.ParamsChecker(&pb.MicroServiceKey{
		AppId:       "a",
		ServiceName: "b",
		Version:     "1.0.0",
	}, []*pb.MicroServiceKey{
		{ServiceName: "a", Version: "1"},
		{ServiceName: "a", Version: "1"},
	})
	assert.NotNil(t, p)
}
