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
	"testing"

	"github.com/apache/servicecomb-service-center/datasource"
	pb "github.com/go-chassis/cari/discovery"
	"github.com/stretchr/testify/assert"
)

func TestDep_Creat(t *testing.T) {

	var (
		consumerId1 string
		consumerId3 string
	)

	t.Run("creat service, when request is valid, should not pass", func(t *testing.T) {
		respCreateService, err := datasource.Instance().RegisterService(getContext(), &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				ServiceId:   "dep1",
				AppId:       "create_dep_group",
				ServiceName: "create_dep_consumer",
				Version:     "1.0.0",
				Level:       "FRONT",
				Status:      pb.MS_UP,
			},
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respCreateService.Response.GetCode())
		consumerId1 = respCreateService.ServiceId

		respCreateService, err = datasource.Instance().RegisterService(getContext(), &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				ServiceId:   "dep2",
				AppId:       "create_dep_group",
				ServiceName: "create_dep_consumer_all",
				Version:     "1.0.0",
				Level:       "FRONT",
				Status:      pb.MS_UP,
			},
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respCreateService.Response.GetCode())
		consumerId3 = respCreateService.ServiceId

		respCreateService, err = datasource.Instance().RegisterService(getContext(), &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				ServiceId:   "dep3",
				Environment: pb.ENV_PROD,
				AppId:       "create_dep_group",
				ServiceName: "create_dep_consumer",
				Version:     "1.0.0",
				Level:       "FRONT",
				Status:      pb.MS_UP,
			},
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respCreateService.Response.GetCode())

		respCreateService, err = datasource.Instance().RegisterService(getContext(), &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				ServiceId:   "dep4",
				AppId:       "create_dep_group",
				ServiceName: "create_dep_provider",
				Version:     "1.0.0",
				Level:       "FRONT",
				Status:      pb.MS_UP,
			},
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respCreateService.Response.GetCode())

		respCreateService, err = datasource.Instance().RegisterService(getContext(), &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				ServiceId:   "dep5",
				AppId:       "create_dep_group",
				ServiceName: "create_dep_provider",
				Version:     "1.0.1",
				Level:       "FRONT",
				Status:      pb.MS_UP,
			},
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respCreateService.Response.GetCode())

		respCreateService, err = datasource.Instance().RegisterService(getContext(), &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				ServiceId:   "dep6",
				Environment: pb.ENV_PROD,
				AppId:       "create_dep_group",
				ServiceName: "create_dep_provider",
				Version:     "1.0.0",
				Level:       "FRONT",
				Status:      pb.MS_UP,
			},
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respCreateService.Response.GetCode())
	})

	t.Run("create dep, when request is valid, should be passed", func(t *testing.T) {

		respCreateDependency, err := datasource.Instance().AddOrUpdateDependencies(getContext(), []*pb.ConsumerDependency{
			{
				Consumer: &pb.MicroServiceKey{
					ServiceName: "create_dep_consumer",
					AppId:       "create_dep_group",
					Version:     "1.0.0",
				},
				Providers: []*pb.MicroServiceKey{
					{
						AppId:       "create_dep_group",
						ServiceName: "create_dep_provider",
						Version:     "latest",
					},
				},
			},
		}, false)
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respCreateDependency.GetCode())

		respPro, err := datasource.Instance().SearchConsumerDependency(getContext(), &pb.GetDependenciesRequest{
			ServiceId: consumerId1,
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respPro.Response.GetCode())
		assert.NotEqual(t, "1.0.1", respPro.Providers[0].Version)

		respCreateDependency, err = datasource.Instance().AddOrUpdateDependencies(getContext(), []*pb.ConsumerDependency{
			{
				Consumer: &pb.MicroServiceKey{
					ServiceName: "create_dep_consumer",
					AppId:       "create_dep_group",
					Version:     "1.0.0",
				},
				Providers: []*pb.MicroServiceKey{
					{
						AppId:       "create_dep_group",
						ServiceName: "create_dep_provider",
						Version:     "1.0.0+",
					},
				},
			},
		}, false)
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respCreateDependency.GetCode())

		respPro, err = datasource.Instance().SearchConsumerDependency(getContext(), &pb.GetDependenciesRequest{
			ServiceId: consumerId1,
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respPro.Response.GetCode())
		assert.NotEqual(t, "1.0.1", respPro.Providers[0].Version)

		respCreateDependency, err = datasource.Instance().AddOrUpdateDependencies(getContext(), []*pb.ConsumerDependency{
			{
				Consumer: &pb.MicroServiceKey{
					ServiceName: "create_dep_consumer",
					AppId:       "create_dep_group",
					Version:     "1.0.0",
				},
				Providers: []*pb.MicroServiceKey{
					{
						AppId:       "create_dep_group",
						ServiceName: "create_dep_provider",
						Version:     "1.0.0+",
					},
				},
			},
		}, false)
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respCreateDependency.GetCode())

		respPro, err = datasource.Instance().SearchConsumerDependency(getContext(), &pb.GetDependenciesRequest{
			ServiceId: consumerId1,
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respPro.Response.GetCode())
		assert.NotEqual(t, 2, len(respPro.Providers))

		respCreateDependency, err = datasource.Instance().AddOrUpdateDependencies(getContext(), []*pb.ConsumerDependency{
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
		}, false)
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respCreateDependency.GetCode())

		respPro, err = datasource.Instance().SearchConsumerDependency(getContext(), &pb.GetDependenciesRequest{
			ServiceId: consumerId3,
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respPro.Response.GetCode())
		assert.NotEqual(t, 0, len(respPro.Providers))

		respCreateDependency, err = datasource.Instance().AddOrUpdateDependencies(getContext(), []*pb.ConsumerDependency{
			{
				Consumer: &pb.MicroServiceKey{
					ServiceName: "create_dep_consumer_all",
					AppId:       "create_dep_group",
					Version:     "1.0.0",
				},
				Providers: nil,
			},
		}, false)
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respCreateDependency.GetCode())

		respCreateDependency, err = datasource.Instance().AddOrUpdateDependencies(getContext(), []*pb.ConsumerDependency{
			{
				Consumer: &pb.MicroServiceKey{
					ServiceName: "create_dep_consumer",
					AppId:       "create_dep_group",
					Version:     "1.0.0",
				},
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
		}, false)
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respCreateDependency.GetCode())

		respCreateDependency, err = datasource.Instance().AddOrUpdateDependencies(getContext(), []*pb.ConsumerDependency{
			{
				Consumer: &pb.MicroServiceKey{
					ServiceName: "create_dep_consumer",
					AppId:       "create_dep_group",
					Version:     "1.0.0",
				},
				Providers: []*pb.MicroServiceKey{
					{
						AppId:       "create_dep_group",
						ServiceName: "create_dep_provider",
						Version:     "1.0.0-1.0.1",
					},
				},
			},
		}, false)
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respCreateDependency.GetCode())

		respPro, err = datasource.Instance().SearchConsumerDependency(getContext(), &pb.GetDependenciesRequest{
			ServiceId: consumerId1,
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respPro.Response.GetCode())
		assert.Equal(t, "1.0.0", respPro.Providers[0].Version)
	})
}

func TestDep_Get(t *testing.T) {

	var (
		consumerId1 string
		providerId1 string
	)

	t.Run("create service, when request is valid, should be passed", func(t *testing.T) {
		respCreateService, err := datasource.Instance().RegisterService(getContext(), &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				ServiceId:   "dep7",
				AppId:       "get_dep_group",
				ServiceName: "get_dep_consumer",
				Version:     "1.0.0",
				Level:       "FRONT",
				Status:      pb.MS_UP,
			},
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respCreateService.Response.GetCode())
		consumerId1 = respCreateService.ServiceId

		respCreateService, err = datasource.Instance().RegisterService(getContext(), &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				ServiceId:   "dep8",
				AppId:       "get_dep_group",
				ServiceName: "get_dep_provider",
				Version:     "1.0.0",
				Level:       "FRONT",
				Status:      pb.MS_UP,
			},
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respCreateService.Response.GetCode())
		providerId1 = respCreateService.ServiceId

		respCreateService, err = datasource.Instance().RegisterService(getContext(), &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				ServiceId:   "dep9",
				AppId:       "get_dep_group",
				ServiceName: "get_dep_provider",
				Version:     "2.0.0",
				Level:       "FRONT",
				Status:      pb.MS_UP,
			},
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respCreateService.Response.GetCode())
	})

	t.Run("execute 'search dep' operation, when request is valid,should be passed", func(t *testing.T) {
		respPro, err := datasource.Instance().SearchProviderDependency(getContext(), &pb.GetDependenciesRequest{
			ServiceId: providerId1,
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respPro.Response.GetCode())

		respCon, err := datasource.Instance().SearchConsumerDependency(getContext(), &pb.GetDependenciesRequest{
			ServiceId: consumerId1,
		})

		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respCon.Response.GetCode())

	})
}
