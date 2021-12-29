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
package disco_test

import (
	"strconv"
	"testing"

	"github.com/go-chassis/cari/pkg/errsvc"
	"github.com/stretchr/testify/assert"

	"github.com/apache/servicecomb-service-center/datasource"
	"github.com/apache/servicecomb-service-center/server/service/disco"
	pb "github.com/go-chassis/cari/discovery"
)

func TestListConsumers(t *testing.T) {
	var (
		consumerId1 string
		providerId1 string
		providerId2 string
	)
	ctx := getContext()
	defer disco.UnregisterManyService(ctx, &pb.DelServicesRequest{ServiceIds: []string{
		consumerId1, providerId1, providerId2,
	}, Force: true})

	t.Run("prepare data, should be passed", func(t *testing.T) {
		respCreateService, err := disco.RegisterService(ctx, &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				AppId:       "get_dep_group",
				ServiceName: "get_dep_consumer",
				Version:     "1.0.0",
				Level:       "FRONT",
				Status:      pb.MS_UP,
			},
		})
		assert.NoError(t, err)
		consumerId1 = respCreateService.ServiceId

		respCreateService, err = disco.RegisterService(ctx, &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				AppId:       "get_dep_group",
				ServiceName: "get_dep_provider",
				Version:     "1.0.0",
				Level:       "FRONT",
				Status:      pb.MS_UP,
			},
		})
		assert.NoError(t, err)
		providerId1 = respCreateService.ServiceId

		respCreateService, err = disco.RegisterService(ctx, &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				AppId:       "get_dep_group",
				ServiceName: "get_dep_provider",
				Version:     "2.0.0",
				Level:       "FRONT",
				Status:      pb.MS_UP,
			},
		})
		assert.NoError(t, err)
		providerId2 = respCreateService.ServiceId
	})

	t.Run("when request is invalid, should be failed", func(t *testing.T) {
		_, err := disco.ListConsumers(ctx, &pb.GetDependenciesRequest{
			ServiceId: "",
		})
		testErr := err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrInvalidParams, testErr.Code)

		_, err = disco.ListConsumers(ctx, &pb.GetDependenciesRequest{
			ServiceId: "noneservice",
		})
		testErr = err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrServiceNotExists, testErr.Code)

		_, err = disco.ListProviders(ctx, &pb.GetDependenciesRequest{
			ServiceId: "",
		})
		testErr = err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrInvalidParams, testErr.Code)

		_, err = disco.ListProviders(ctx, &pb.GetDependenciesRequest{
			ServiceId: "noneservice",
		})
		testErr = err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrServiceNotExists, testErr.Code)
	})

	t.Run("when request is valid, should be passed", func(t *testing.T) {
		respPro, err := disco.ListConsumers(ctx, &pb.GetDependenciesRequest{
			ServiceId: providerId1,
		})
		assert.NoError(t, err)
		assert.Empty(t, respPro.Consumers)

		respCon, err := disco.ListProviders(ctx, &pb.GetDependenciesRequest{
			ServiceId: consumerId1,
		})
		assert.NoError(t, err)
		assert.Empty(t, respCon.Providers)
	})

	t.Run("when after finding instance, should be passed", func(t *testing.T) {
		_, err := disco.FindInstances(ctx, &pb.FindInstancesRequest{
			ConsumerServiceId: consumerId1,
			AppId:             "get_dep_group",
			ServiceName:       "get_dep_provider",
		})
		assert.NoError(t, err)

		DependencyHandle()

		respGetP, err := disco.ListConsumers(ctx, &pb.GetDependenciesRequest{
			ServiceId: providerId1,
		})
		assert.NoError(t, err)
		assert.Equal(t, 1, len(respGetP.Consumers))
		assert.Equal(t, consumerId1, respGetP.Consumers[0].ServiceId)

		respGetC, err := disco.ListProviders(ctx, &pb.GetDependenciesRequest{
			ServiceId: consumerId1,
		})
		assert.NoError(t, err)
		assert.Equal(t, 2, len(respGetC.Providers))

		_, err = disco.FindInstances(ctx, &pb.FindInstancesRequest{
			ConsumerServiceId: consumerId1,
			AppId:             "get_dep_group",
			ServiceName:       "get_dep_consumer",
		})
		assert.NoError(t, err)

		DependencyHandle()

		respGetC, err = disco.ListProviders(ctx, &pb.GetDependenciesRequest{
			ServiceId: consumerId1,
			NoSelf:    true,
		})
		assert.NoError(t, err)
		assert.Equal(t, 2, len(respGetC.Providers))

		respGetC, err = disco.ListProviders(ctx, &pb.GetDependenciesRequest{
			ServiceId: consumerId1,
			NoSelf:    false,
		})
		assert.NoError(t, err)
		assert.Equal(t, 3, len(respGetC.Providers))

		_, err = disco.FindInstances(ctx, &pb.FindInstancesRequest{
			ConsumerServiceId: providerId2,
			AppId:             "get_dep_group",
			ServiceName:       "get_dep_finder",
		})
		testErr := err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrServiceNotExists, testErr.Code)

		respCreateF, err := disco.RegisterService(ctx, &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				AppId:       "get_dep_group",
				ServiceName: "get_dep_finder",
				Version:     "1.0.0",
				Level:       "FRONT",
				Status:      pb.MS_UP,
			},
		})
		assert.NoError(t, err)
		finder1 := respCreateF.ServiceId
		defer disco.UnregisterService(ctx, &pb.DeleteServiceRequest{ServiceId: finder1, Force: true})

		_, err = disco.FindInstances(ctx, &pb.FindInstancesRequest{
			ConsumerServiceId: providerId2,
			AppId:             "get_dep_group",
			ServiceName:       "get_dep_finder",
		})
		assert.NoError(t, err)

		DependencyHandle()

		respGetC, err = disco.ListProviders(ctx, &pb.GetDependenciesRequest{
			ServiceId: providerId2,
		})
		assert.NoError(t, err)
		assert.Equal(t, 1, len(respGetC.Providers))
		assert.Equal(t, finder1, respGetC.Providers[0].ServiceId)

		err = disco.UnregisterService(ctx, &pb.DeleteServiceRequest{
			ServiceId: finder1, Force: true,
		})
		assert.NoError(t, err)

		DependencyHandle()

		respGetC, err = disco.ListProviders(ctx, &pb.GetDependenciesRequest{
			ServiceId: providerId2,
		})
		assert.NoError(t, err)
		assert.Equal(t, 0, len(respGetC.Providers))

		respCreateF, err = disco.RegisterService(ctx, &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				ServiceId:   finder1,
				AppId:       "get_dep_group",
				ServiceName: "get_dep_finder",
				Version:     "1.0.0",
				Level:       "FRONT",
				Status:      pb.MS_UP,
			},
		})
		assert.NoError(t, err)

		_, err = disco.FindInstances(ctx, &pb.FindInstancesRequest{
			ConsumerServiceId: providerId2,
			AppId:             "get_dep_group",
			ServiceName:       "get_dep_finder",
		})
		assert.NoError(t, err)

		DependencyHandle()

		respGetC, err = disco.ListProviders(ctx, &pb.GetDependenciesRequest{
			ServiceId: providerId2,
		})
		assert.NoError(t, err)
		assert.Equal(t, 1, len(respGetC.Providers))
		assert.Equal(t, finder1, respGetC.Providers[0].ServiceId)
	})
}

func TestPutDependencies(t *testing.T) {
	var (
		consumerId1 string
		consumerId2 string
	)
	ctx := getContext()

	respCreateService, err := disco.RegisterService(ctx, &pb.CreateServiceRequest{
		Service: &pb.MicroService{
			AppId:       "create_dep_group",
			ServiceName: "create_dep_consumer",
			Version:     "1.0.0",
			Level:       "FRONT",
			Status:      pb.MS_UP,
		},
	})
	assert.NoError(t, err)
	consumerId1 = respCreateService.ServiceId

	respCreateService, err = disco.RegisterService(ctx, &pb.CreateServiceRequest{
		Service: &pb.MicroService{
			Environment: pb.ENV_PROD,
			AppId:       "create_dep_group",
			ServiceName: "create_dep_consumer",
			Version:     "1.0.0",
			Level:       "FRONT",
			Status:      pb.MS_UP,
		},
	})
	assert.NoError(t, err)
	consumerId2 = respCreateService.ServiceId

	respCreateService, err = disco.RegisterService(ctx, &pb.CreateServiceRequest{
		Service: &pb.MicroService{
			AppId:       "create_dep_group",
			ServiceName: "create_dep_provider",
			Version:     "1.0.0",
			Level:       "FRONT",
			Status:      pb.MS_UP,
		},
	})
	assert.NoError(t, err)

	respCreateService, err = disco.RegisterService(ctx, &pb.CreateServiceRequest{
		Service: &pb.MicroService{
			AppId:       "create_dep_group",
			ServiceName: "create_dep_provider",
			Version:     "1.0.1",
			Level:       "FRONT",
			Status:      pb.MS_UP,
		},
	})
	assert.NoError(t, err)
	providerID1 := respCreateService.ServiceId

	respCreateService, err = disco.RegisterService(ctx, &pb.CreateServiceRequest{
		Service: &pb.MicroService{
			Environment: pb.ENV_PROD,
			AppId:       "create_dep_group",
			ServiceName: "create_dep_provider",
			Version:     "1.0.0",
			Level:       "FRONT",
			Status:      pb.MS_UP,
		},
	})
	assert.NoError(t, err)
	providerID2 := respCreateService.ServiceId

	respCreateService, err = disco.RegisterService(ctx, &pb.CreateServiceRequest{
		Service: &pb.MicroService{
			AppId:       "create_dep_group",
			ServiceName: "create_dep_provider_other",
			Version:     "1.0.0",
			Level:       "FRONT",
			Status:      pb.MS_UP,
		},
	})
	assert.NoError(t, err)
	providerID3 := respCreateService.ServiceId
	defer disco.UnregisterManyService(ctx, &pb.DelServicesRequest{ServiceIds: []string{
		consumerId1, consumerId2,
		providerID1, providerID2, providerID3,
	}, Force: true})

	t.Run("when request is invalid, should be failed", func(t *testing.T) {
		err := disco.PutDependencies(ctx, &pb.CreateDependenciesRequest{})
		testErr := err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrInvalidParams, testErr.Code)

		consumer := &pb.MicroServiceKey{
			AppId:       "create_dep_group",
			ServiceName: "create_dep_consumer",
			Version:     "1.0.0",
		}
		providers := []*pb.MicroServiceKey{
			{
				AppId:       "create_dep_group",
				ServiceName: "create_dep_provider",
			},
		}

		err = disco.PutDependencies(ctx, &pb.CreateDependenciesRequest{
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
		testErr = err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrServiceNotExists, testErr.Code)

		err = disco.PutDependencies(ctx, &pb.CreateDependenciesRequest{
			Dependencies: []*pb.ConsumerDependency{
				{
					Consumer: consumer,
					Providers: []*pb.MicroServiceKey{
						{
							AppId:       "create_dep_group",
							ServiceName: "",
						},
					},
				},
			},
		})
		testErr = err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrInvalidParams, testErr.Code)

		err = disco.PutDependencies(ctx, &pb.CreateDependenciesRequest{
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
		testErr = err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrInvalidParams, testErr.Code)

		err = disco.PutDependencies(ctx, &pb.CreateDependenciesRequest{
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
		testErr = err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrInvalidParams, testErr.Code)

		err = disco.PutDependencies(ctx, &pb.CreateDependenciesRequest{
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
		testErr = err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrInvalidParams, testErr.Code)

		err = disco.PutDependencies(ctx, &pb.CreateDependenciesRequest{
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
		testErr = err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrInvalidParams, testErr.Code)

		err = disco.PutDependencies(ctx, &pb.CreateDependenciesRequest{
			Dependencies: []*pb.ConsumerDependency{
				{
					Consumer: &pb.MicroServiceKey{
						AppId:       "create_dep_group",
						ServiceName: "create_dep_consumer",
						Version:     "1.0.32768",
					},
					Providers: providers,
				},
			},
		})
		testErr = err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrInvalidParams, testErr.Code)

		err = disco.PutDependencies(ctx, &pb.CreateDependenciesRequest{
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
		testErr = err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrInvalidParams, testErr.Code)

		err = disco.PutDependencies(ctx, &pb.CreateDependenciesRequest{
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
		testErr = err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrInvalidParams, testErr.Code)

		err = disco.PutDependencies(ctx, &pb.CreateDependenciesRequest{
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
		testErr = err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrInvalidParams, testErr.Code)

		err = disco.PutDependencies(ctx, &pb.CreateDependenciesRequest{
			Dependencies: []*pb.ConsumerDependency{
				{
					Consumer: consumer,
					Providers: []*pb.MicroServiceKey{
						{
							Environment: pb.ENV_PROD,
							AppId:       "service_group_provider",
							ServiceName: "service_name_provider",
						},
					},
				},
			},
		})
		assert.NoError(t, err)

		consumer.Environment = pb.ENV_PROD
		err = disco.PutDependencies(ctx, &pb.CreateDependenciesRequest{
			Dependencies: []*pb.ConsumerDependency{
				{
					Consumer: consumer,
					Providers: []*pb.MicroServiceKey{
						{
							AppId:       "service_group_provider",
							ServiceName: "service_name_provider",
						},
					},
				},
			},
		})
		assert.NoError(t, err)

		DependencyHandle()

		respCon, err := disco.ListProviders(ctx, &pb.GetDependenciesRequest{
			ServiceId: consumerId1,
		})
		assert.NoError(t, err)
		assert.Equal(t, 0, len(respCon.Providers))

		respCon, err = disco.ListProviders(ctx, &pb.GetDependenciesRequest{
			ServiceId: consumerId2,
		})
		assert.NoError(t, err)
		assert.Equal(t, 0, len(respCon.Providers))

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
					},
				},
			})
		}
		err = disco.PutDependencies(ctx, &pb.CreateDependenciesRequest{
			Dependencies: deps,
		})
		testErr = err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrInvalidParams, testErr.Code)
	})

	t.Run("when request is valid, should be passed", func(t *testing.T) {
		consumer := &pb.MicroServiceKey{
			ServiceName: "create_dep_consumer",
			AppId:       "create_dep_group",
			Version:     "1.0.0",
		}

		err := disco.PutDependencies(ctx, &pb.CreateDependenciesRequest{
			Dependencies: []*pb.ConsumerDependency{
				{
					Consumer: consumer,
					Providers: []*pb.MicroServiceKey{
						{
							AppId:       "create_dep_group",
							ServiceName: "create_dep_provider",
						},
					},
				},
			},
		})
		assert.NoError(t, err)

		DependencyHandle()

		respPro, err := disco.ListProviders(ctx, &pb.GetDependenciesRequest{
			ServiceId: consumerId1,
		})
		assert.NoError(t, err)
		assert.Equal(t, 2, len(respPro.Providers))

		err = disco.AddDependencies(ctx, &pb.AddDependenciesRequest{
			Dependencies: []*pb.ConsumerDependency{
				{
					Consumer: consumer,
					Providers: []*pb.MicroServiceKey{
						{
							AppId:       "create_dep_group",
							ServiceName: "create_dep_provider_other",
						},
					},
				},
			},
		})
		assert.NoError(t, err)

		DependencyHandle()

		respPro, err = disco.ListProviders(ctx, &pb.GetDependenciesRequest{
			ServiceId: consumerId1,
		})
		assert.NoError(t, err)
		assert.Equal(t, 3, len(respPro.Providers))

		err = disco.PutDependencies(ctx, &pb.CreateDependenciesRequest{
			Dependencies: []*pb.ConsumerDependency{
				{
					Consumer:  consumer,
					Providers: nil,
				},
			},
		})
		assert.NoError(t, err)

		DependencyHandle()

		respPro, err = disco.ListProviders(ctx, &pb.GetDependenciesRequest{
			ServiceId: consumerId1,
		})
		assert.NoError(t, err)
		assert.Equal(t, 0, len(respPro.Providers))
	})
}

func DependencyHandle() {
	datasource.GetDependencyManager().DependencyHandle(getContext())
}
