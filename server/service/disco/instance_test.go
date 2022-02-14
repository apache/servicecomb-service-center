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
	"context"
	"fmt"
	"math"
	"strconv"
	"strings"
	"testing"

	_ "github.com/apache/servicecomb-service-center/test"

	"github.com/apache/servicecomb-service-center/pkg/util"
	"github.com/apache/servicecomb-service-center/server/config"
	"github.com/apache/servicecomb-service-center/server/core"
	discosvc "github.com/apache/servicecomb-service-center/server/service/disco"
	pb "github.com/go-chassis/cari/discovery"
	"github.com/go-chassis/cari/pkg/errsvc"
	"github.com/stretchr/testify/assert"
)

var (
	TOO_LONG_HOSTNAME = strings.Repeat("x", 65)
	TOO_LONG_URL      = strings.Repeat("x", 513)
)

func TestRegisterInstance(t *testing.T) {
	var (
		serviceId1 string
	)
	ctx := getContext()
	defer discosvc.UnregisterService(ctx, &pb.DeleteServiceRequest{ServiceId: serviceId1, Force: true})

	t.Run("prepare data, should be passed", func(t *testing.T) {
		respCreate, err := discosvc.RegisterService(ctx, &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				ServiceName: "create_instance_service",
				AppId:       "create_instance",
				Version:     "1.0.0",
				Level:       "FRONT",
				Status:      pb.MS_UP,
			},
		})
		assert.NoError(t, err)
		serviceId1 = respCreate.ServiceId
	})

	t.Run("when register a instance, should be passed", func(t *testing.T) {
		resp, err := discosvc.RegisterInstance(ctx, &pb.RegisterInstanceRequest{
			Instance: &pb.MicroServiceInstance{
				ServiceId: serviceId1,
				Endpoints: []string{
					"createInstance:127.0.0.1:8080",
				},
				HostName: "UT-HOST",
				Status:   pb.MSI_UP,
			},
		})
		assert.NoError(t, err)
		assert.NotEmpty(t, resp.InstanceId)

		resp, err = discosvc.RegisterInstance(ctx, &pb.RegisterInstanceRequest{
			Instance: &pb.MicroServiceInstance{
				InstanceId: "customId",
				ServiceId:  serviceId1,
				Endpoints: []string{
					"createInstance:127.0.0.1:8080",
				},
				HostName: "UT-HOST",
				Status:   pb.MSI_UP,
			},
		})
		assert.NoError(t, err)
		assert.Equal(t, "customId", resp.InstanceId)

		resp, err = discosvc.RegisterInstance(ctx, &pb.RegisterInstanceRequest{
			Instance: &pb.MicroServiceInstance{
				ServiceId: serviceId1,
				Endpoints: []string{
					"createInstance:127.0.0.1:8081",
				},
				HostName: "UT-HOST",
			},
		})
		assert.NoError(t, err)
		assert.NotEmpty(t, resp.InstanceId)

		size := 1000
		var eps []string
		properties := make(map[string]string, size)
		for i := 0; i < size; i++ {
			s := strconv.Itoa(i) + strings.Repeat("x", 253)
			eps = append(eps, s)
			properties[s] = s
		}
		resp, err = discosvc.RegisterInstance(ctx, &pb.RegisterInstanceRequest{
			Instance: &pb.MicroServiceInstance{
				ServiceId:  serviceId1,
				Endpoints:  eps,
				HostName:   TOO_LONG_HOSTNAME[:len(TOO_LONG_HOSTNAME)-1],
				Properties: properties,
				HealthCheck: &pb.HealthCheck{
					Mode:     "pull",
					Port:     math.MaxUint16,
					Interval: math.MaxInt32,
					Times:    math.MaxInt32,
					Url:      TOO_LONG_URL[:len(TOO_LONG_URL)-1],
				},
				DataCenterInfo: &pb.DataCenterInfo{
					Name:          TooLongServiceName[:len(TooLongServiceName)-1],
					Region:        TooLongServiceName[:len(TooLongServiceName)-1],
					AvailableZone: TooLongServiceName[:len(TooLongServiceName)-1],
				},
			},
		})
		assert.NoError(t, err)
		assert.NotEmpty(t, resp.InstanceId)
	})

	t.Run("when update the same instance, should be passed", func(t *testing.T) {
		instance := &pb.MicroServiceInstance{
			ServiceId: serviceId1,
			Endpoints: []string{
				"sameInstance:127.0.0.1:8080",
			},
			HostName: "UT-HOST",
			Status:   pb.MSI_UP,
		}
		resp, err := discosvc.RegisterInstance(ctx, &pb.RegisterInstanceRequest{
			Instance: instance,
		})
		assert.NoError(t, err)

		resp, err = discosvc.RegisterInstance(ctx, &pb.RegisterInstanceRequest{
			Instance: instance,
		})
		assert.NoError(t, err)
		assert.Equal(t, instance.InstanceId, resp.InstanceId)
	})

	t.Run("when register invalid instance, should be failed", func(t *testing.T) {
		_, err := discosvc.RegisterInstance(ctx, &pb.RegisterInstanceRequest{
			Instance: &pb.MicroServiceInstance{
				ServiceId: serviceId1,
				Endpoints: []string{""},
				HostName:  "UT-HOST",
				Status:    pb.MSI_UP,
			},
		})
		testErr := err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrInvalidParams, testErr.Code)

		_, err = discosvc.RegisterInstance(ctx, &pb.RegisterInstanceRequest{
			Instance: &pb.MicroServiceInstance{
				HostName: "UT-HOST",
				Endpoints: []string{
					"check:127.0.0.1:8080",
				},
				Status: pb.MSI_UP,
			},
		})
		testErr = err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrInvalidParams, testErr.Code)

		_, err = discosvc.RegisterInstance(ctx, &pb.RegisterInstanceRequest{
			Instance: &pb.MicroServiceInstance{
				ServiceId: "1000000000000000",
				Endpoints: []string{
					"check:127.0.0.1:8080",
				},
				HostName: "UT-HOST",
				Status:   pb.MSI_UP,
			},
		})
		testErr = err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrServiceNotExists, testErr.Code)

		_, err = discosvc.RegisterInstance(ctx, &pb.RegisterInstanceRequest{
			Instance: &pb.MicroServiceInstance{
				ServiceId: serviceId1,
				Endpoints: []string{
					"check:127.0.0.1:8080",
				},
				Status: pb.MSI_UP,
			},
		})
		testErr = err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrInvalidParams, testErr.Code)

		_, err = discosvc.RegisterInstance(ctx, &pb.RegisterInstanceRequest{
			Instance: &pb.MicroServiceInstance{
				ServiceId: serviceId1,
				HostName:  " ",
				Endpoints: []string{
					"check:127.0.0.1:8080",
				},
				Status: pb.MSI_UP,
			},
		})
		testErr = err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrInvalidParams, testErr.Code)

		_, err = discosvc.RegisterInstance(ctx, &pb.RegisterInstanceRequest{
			Instance: &pb.MicroServiceInstance{
				ServiceId: serviceId1,
				HostName:  TOO_LONG_HOSTNAME,
				Endpoints: []string{
					"check:127.0.0.1:8080",
				},
				Status: pb.MSI_UP,
			},
		})
		testErr = err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrInvalidParams, testErr.Code)

		_, err = discosvc.RegisterInstance(ctx, &pb.RegisterInstanceRequest{
			Instance: nil,
		})
		testErr = err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrInvalidParams, testErr.Code)

		_, err = discosvc.RegisterInstance(ctx, &pb.RegisterInstanceRequest{
			Instance: &pb.MicroServiceInstance{
				ServiceId: serviceId1,
				Endpoints: []string{
					"checkpull:127.0.0.1:8080",
				},
				HostName: "UT-HOST",
				Status:   pb.MSI_UP,
				HealthCheck: &pb.HealthCheck{
					Mode:     "push",
					Interval: 30,
					Times:    1,
				},
			},
		})
		assert.NoError(t, err)

		_, err = discosvc.RegisterInstance(ctx, &pb.RegisterInstanceRequest{
			Instance: &pb.MicroServiceInstance{
				ServiceId: serviceId1,
				Endpoints: []string{
					"checkpush:127.0.0.1:8081",
				},
				HostName: "UT-HOST",
				Status:   pb.MSI_UP,
				HealthCheck: &pb.HealthCheck{
					Mode:     "push",
					Interval: 30,
					Times:    0,
				},
			},
		})
		assert.NoError(t, err)

		_, err = discosvc.RegisterInstance(ctx, &pb.RegisterInstanceRequest{
			Instance: &pb.MicroServiceInstance{
				ServiceId: serviceId1,
				Endpoints: []string{
					"checkpush:127.0.0.1:8080",
				},
				HostName: "UT-HOST",
				Status:   pb.MSI_UP,
				HealthCheck: &pb.HealthCheck{
					Mode:     "pull",
					Interval: 30,
					Times:    1,
					Url:      "/abc/d",
				},
			},
		})
		assert.NoError(t, err)

		_, err = discosvc.RegisterInstance(ctx, &pb.RegisterInstanceRequest{
			Instance: &pb.MicroServiceInstance{
				ServiceId: serviceId1,
				Endpoints: []string{
					"checkpush:127.0.0.1:8081",
				},
				HostName: "UT-HOST",
				Status:   pb.MSI_UP,
				HealthCheck: &pb.HealthCheck{
					Mode:     "push",
					Interval: 30,
					Times:    -1,
				},
			},
		})
		testErr = err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrInvalidParams, testErr.Code)

		_, err = discosvc.RegisterInstance(ctx, &pb.RegisterInstanceRequest{
			Instance: &pb.MicroServiceInstance{
				ServiceId: serviceId1,
				Endpoints: []string{
					"checkpull:127.0.0.1:8081",
				},
				HostName: "UT-HOST",
				Status:   pb.MSI_UP,
				HealthCheck: &pb.HealthCheck{
					Mode:     "pull",
					Interval: 30,
					Times:    1,
					Url:      " ",
				},
			},
		})
		testErr = err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrInvalidParams, testErr.Code)

		_, err = discosvc.RegisterInstance(ctx, &pb.RegisterInstanceRequest{
			Instance: &pb.MicroServiceInstance{
				ServiceId: serviceId1,
				Endpoints: []string{
					"checkpull:127.0.0.1:8081",
				},
				HostName: "UT-HOST",
				Status:   pb.MSI_UP,
				HealthCheck: &pb.HealthCheck{
					Mode:     "pull",
					Interval: 0,
					Times:    0,
				},
			},
		})
		testErr = err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrInvalidParams, testErr.Code)

		_, err = discosvc.RegisterInstance(ctx, &pb.RegisterInstanceRequest{
			Instance: &pb.MicroServiceInstance{
				ServiceId: serviceId1,
				Endpoints: []string{
					"checkpull:127.0.0.1:8081",
				},
				HostName: "UT-HOST",
				Status:   pb.MSI_UP,
				HealthCheck: &pb.HealthCheck{
					Mode:     "pull",
					Interval: 30,
					Times:    1,
					Port:     math.MaxUint16 + 1,
				},
			},
		})
		testErr = err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrInvalidParams, testErr.Code)

		_, err = discosvc.RegisterInstance(ctx, &pb.RegisterInstanceRequest{
			Instance: &pb.MicroServiceInstance{
				ServiceId: serviceId1,
				Endpoints: []string{
					"createInstance:127.0.0.1:8083",
				},
				HostName: "UT-HOST",
				Status:   "Invalid",
			},
		})
		testErr = err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrInvalidParams, testErr.Code)
	})

	t.Run(" register properties info to instance, should be passed", func(t *testing.T) {
		instance := &pb.MicroServiceInstance{
			ServiceId: serviceId1,
			Endpoints: []string{
				"sameInstance:127.0.0.1:8080",
			},
			HostName: "UT-HOST",
			Status:   pb.MSI_UP,
		}
		_, err := discosvc.RegisterInstance(ctx, &pb.RegisterInstanceRequest{
			Instance: instance,
		})
		assert.NoError(t, err)
		assert.Equal(t, 2, len(instance.Properties))
		engineID, ok1 := instance.Properties["engineID"]
		engineName, ok2 := instance.Properties["engineName"]
		assert.Equal(t, true, ok1 && ok2)
		assert.Equal(t, "test_engineID", engineID)
		assert.Equal(t, "test_engineName", engineName)
	})
}

func TestFindManyInstances(t *testing.T) {
	var (
		serviceId1  string
		serviceId2  string
		serviceId3  string
		serviceId4  string
		serviceId5  string
		serviceId6  string
		serviceId7  string
		serviceId8  string
		serviceId9  string
		serviceId10 string
		instanceId1 string
		instanceId2 string
		instanceId4 string
		instanceId5 string
		instanceId8 string
		instanceId9 string
	)
	ctx := getContext()
	defer discosvc.UnregisterManyService(ctx, &pb.DelServicesRequest{ServiceIds: []string{
		serviceId1, serviceId2, serviceId3, serviceId4, serviceId5,
		serviceId7, serviceId8, serviceId9, serviceId10,
	}, Force: true})
	defer discosvc.UnregisterService(util.SetDomainProject(util.CloneContext(ctx), "user", "user"), &pb.DeleteServiceRequest{
		ServiceId: serviceId6, Force: true,
	})

	t.Run("prepare data, should be passed", func(t *testing.T) {
		respCreate, err := discosvc.RegisterService(ctx, &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				AppId:       "query_instance",
				ServiceName: "query_instance_service",
				Version:     "1.0.0",
				Level:       "FRONT",
				Status:      pb.MS_UP,
			},
		})
		assert.NoError(t, err)
		serviceId1 = respCreate.ServiceId

		respCreate, err = discosvc.RegisterService(ctx, &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				AppId:       "query_instance",
				ServiceName: "query_instance_service",
				Version:     "1.0.5",
				Level:       "FRONT",
				Status:      pb.MS_UP,
			},
		})
		assert.NoError(t, err)
		serviceId2 = respCreate.ServiceId

		respCreate, err = discosvc.RegisterService(ctx, &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				AppId:       "query_instance_diff_app",
				ServiceName: "query_instance_service",
				Version:     "1.0.0",
				Level:       "FRONT",
				Status:      pb.MS_UP,
			},
		})
		assert.NoError(t, err)
		serviceId3 = respCreate.ServiceId

		respCreate, err = discosvc.RegisterService(ctx, &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				Environment: pb.ENV_PROD,
				AppId:       "query_instance",
				ServiceName: "query_instance_diff_env_service",
				Version:     "1.0.0",
				Level:       "FRONT",
				Status:      pb.MS_UP,
			},
		})
		assert.NoError(t, err)
		serviceId4 = respCreate.ServiceId

		respCreate, err = discosvc.RegisterService(ctx, &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				Environment: pb.ENV_PROD,
				AppId:       "default",
				ServiceName: "query_instance_shared_provider",
				Version:     "1.0.0",
				Level:       "FRONT",
				Status:      pb.MS_UP,
				Properties: map[string]string{
					pb.PropAllowCrossApp: "true",
				},
			},
		})
		assert.NoError(t, err)
		serviceId5 = respCreate.ServiceId

		respCreate, err = discosvc.RegisterService(
			util.SetDomainProject(util.CloneContext(ctx), "user", "user"),
			&pb.CreateServiceRequest{
				Service: &pb.MicroService{
					AppId:       "default",
					ServiceName: "query_instance_diff_domain_consumer",
					Version:     "1.0.0",
					Level:       "FRONT",
					Status:      pb.MS_UP,
				},
			})
		assert.NoError(t, err)
		serviceId6 = respCreate.ServiceId

		respCreate, err = discosvc.RegisterService(ctx, &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				AppId:       "default",
				ServiceName: "query_instance_shared_consumer",
				Version:     "1.0.0",
				Level:       "FRONT",
				Status:      pb.MS_UP,
			},
		})
		assert.NoError(t, err)
		serviceId7 = respCreate.ServiceId

		respCreate, err = discosvc.RegisterService(ctx, &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				AppId:       "query_instance",
				ServiceName: "query_instance_with_rev",
				Version:     "1.0.0",
				Level:       "FRONT",
				Status:      pb.MS_UP,
			},
		})
		assert.NoError(t, err)
		serviceId8 = respCreate.ServiceId

		respCreate, err = discosvc.RegisterService(ctx, &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				AppId:       "query_instance",
				ServiceName: "batch_query_instance_with_rev",
				Version:     "1.0.0",
				Level:       "FRONT",
				Status:      pb.MS_UP,
			},
		})
		assert.NoError(t, err)
		serviceId9 = respCreate.ServiceId

		respCreate, err = discosvc.RegisterService(ctx, &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				AppId:       "query_instance_alias",
				ServiceName: "query_instance_alias",
				Alias:       "query_instance_alias:query_instance_alias",
				Version:     "1.0.0",
				Level:       "FRONT",
				Status:      pb.MS_UP,
			},
		})
		assert.NoError(t, err)
		serviceId10 = respCreate.ServiceId

		resp, err := discosvc.RegisterInstance(ctx, &pb.RegisterInstanceRequest{
			Instance: &pb.MicroServiceInstance{
				ServiceId: serviceId1,
				HostName:  "UT-HOST",
				Endpoints: []string{
					"find:127.0.0.1:8080",
				},
				Status: pb.MSI_UP,
			},
		})
		assert.NoError(t, err)
		instanceId1 = resp.InstanceId

		resp, err = discosvc.RegisterInstance(ctx, &pb.RegisterInstanceRequest{
			Instance: &pb.MicroServiceInstance{
				ServiceId: serviceId2,
				HostName:  "UT-HOST",
				Endpoints: []string{
					"find:127.0.0.2:8080",
				},
				Status: pb.MSI_UP,
			},
		})
		assert.NoError(t, err)
		instanceId2 = resp.InstanceId

		resp, err = discosvc.RegisterInstance(ctx, &pb.RegisterInstanceRequest{
			Instance: &pb.MicroServiceInstance{
				ServiceId: serviceId4,
				HostName:  "UT-HOST",
				Endpoints: []string{
					"find:127.0.0.4:8080",
				},
				Status: pb.MSI_UP,
			},
		})
		assert.NoError(t, err)
		instanceId4 = resp.InstanceId

		resp, err = discosvc.RegisterInstance(ctx, &pb.RegisterInstanceRequest{
			Instance: &pb.MicroServiceInstance{
				ServiceId: serviceId5,
				HostName:  "UT-HOST",
				Endpoints: []string{
					"find:127.0.0.5:8080",
				},
				Status: pb.MSI_UP,
			},
		})
		assert.NoError(t, err)
		instanceId5 = resp.InstanceId

		resp, err = discosvc.RegisterInstance(ctx, &pb.RegisterInstanceRequest{
			Instance: &pb.MicroServiceInstance{
				ServiceId: serviceId8,
				HostName:  "UT-HOST",
				Endpoints: []string{
					"find:127.0.0.8:8080",
				},
				Status: pb.MSI_UP,
			},
		})
		assert.NoError(t, err)
		instanceId8 = resp.InstanceId

		resp, err = discosvc.RegisterInstance(ctx, &pb.RegisterInstanceRequest{
			Instance: &pb.MicroServiceInstance{
				ServiceId: serviceId9,
				HostName:  "UT-HOST",
				Endpoints: []string{
					"find:127.0.0.9:8080",
				},
				Status: pb.MSI_UP,
			},
		})
		assert.NoError(t, err)
		instanceId9 = resp.InstanceId

		resp, err = discosvc.RegisterInstance(ctx, &pb.RegisterInstanceRequest{
			Instance: &pb.MicroServiceInstance{
				ServiceId: serviceId10,
				HostName:  "UT-HOST",
				Endpoints: []string{
					"find:127.0.0.9:8080",
				},
				Status: pb.MSI_UP,
			},
		})
		assert.NoError(t, err)
	})

	t.Run("when batch query invalid parameters, should be failed", func(t *testing.T) {
		respFind, err := discosvc.FindManyInstances(ctx, &pb.BatchFindInstancesRequest{
			ConsumerServiceId: serviceId1,
			Services:          nil,
			Instances:         nil,
		})
		testErr := err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrInvalidParams, testErr.Code)

		respFind, err = discosvc.FindManyInstances(ctx, &pb.BatchFindInstancesRequest{
			ConsumerServiceId: serviceId1,
			Services:          []*pb.FindService{},
			Instances:         []*pb.FindInstance{},
		})
		testErr = err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrInvalidParams, testErr.Code)

		respFind, err = discosvc.FindManyInstances(ctx, &pb.BatchFindInstancesRequest{
			ConsumerServiceId: serviceId1,
			Services:          []*pb.FindService{{}},
		})
		testErr = err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrInvalidParams, testErr.Code)

		respFind, err = discosvc.FindManyInstances(ctx, &pb.BatchFindInstancesRequest{
			ConsumerServiceId: serviceId1,
			Instances:         []*pb.FindInstance{{}},
		})
		testErr = err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrInvalidParams, testErr.Code)

		respFind, err = discosvc.FindManyInstances(ctx, &pb.BatchFindInstancesRequest{
			ConsumerServiceId: serviceId1,
			Services: []*pb.FindService{
				{
					Service: &pb.MicroServiceKey{
						AppId:       TooLongAppID,
						ServiceName: "query_instance_service",
					},
				},
			},
		})
		testErr = err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrInvalidParams, testErr.Code)

		respFind, err = discosvc.FindManyInstances(ctx, &pb.BatchFindInstancesRequest{
			ConsumerServiceId: serviceId1,
			Services: []*pb.FindService{
				{
					Service: &pb.MicroServiceKey{
						AppId:       "",
						ServiceName: "query_instance_service",
					},
				},
			},
		})
		testErr = err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrInvalidParams, testErr.Code)

		respFind, err = discosvc.FindManyInstances(ctx, &pb.BatchFindInstancesRequest{
			ConsumerServiceId: serviceId1,
			Services: []*pb.FindService{
				{
					Service: &pb.MicroServiceKey{
						AppId:       " ",
						ServiceName: "query_instance_service",
					},
				},
			},
		})
		testErr = err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrInvalidParams, testErr.Code)

		respFind, err = discosvc.FindManyInstances(ctx, &pb.BatchFindInstancesRequest{
			ConsumerServiceId: serviceId1,
			Services: []*pb.FindService{
				{
					Service: &pb.MicroServiceKey{
						AppId:       "query_instance",
						ServiceName: TooLongExistence,
					},
				},
			},
		})
		testErr = err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrInvalidParams, testErr.Code)

		respFind, err = discosvc.FindManyInstances(ctx, &pb.BatchFindInstancesRequest{
			ConsumerServiceId: serviceId1,
			Services: []*pb.FindService{
				{
					Service: &pb.MicroServiceKey{
						AppId:       "query_instance",
						ServiceName: "",
					},
				},
			},
		})
		testErr = err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrInvalidParams, testErr.Code)

		respFind, err = discosvc.FindManyInstances(ctx, &pb.BatchFindInstancesRequest{
			ConsumerServiceId: serviceId1,
			Services: []*pb.FindService{
				{
					Service: &pb.MicroServiceKey{
						AppId:       "query_instance",
						ServiceName: " ",
					},
				},
			},
		})
		testErr = err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrInvalidParams, testErr.Code)

		respFind, err = discosvc.FindManyInstances(ctx, &pb.BatchFindInstancesRequest{
			ConsumerServiceId: serviceId1,
			Instances: []*pb.FindInstance{
				{
					Instance: &pb.HeartbeatSetElement{
						ServiceId: "query_instance",
					},
				},
			},
		})
		testErr = err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrInvalidParams, testErr.Code)

		respFind, err = discosvc.FindManyInstances(ctx, &pb.BatchFindInstancesRequest{
			ConsumerServiceId: serviceId1,
			Instances: []*pb.FindInstance{
				{
					Instance: &pb.HeartbeatSetElement{
						InstanceId: "query_instance",
					},
				},
			},
		})
		testErr = err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrInvalidParams, testErr.Code)

		respFind, err = discosvc.FindManyInstances(ctx, &pb.BatchFindInstancesRequest{
			Services: []*pb.FindService{
				{
					Service: &pb.MicroServiceKey{
						AppId:       "query_instance",
						ServiceName: "query_instance_service",
					},
				},
			},
		})
		assert.NoError(t, err)
		assert.NotEqual(t, 0, len(respFind.Services.Updated))

		respFind, err = discosvc.FindManyInstances(ctx, &pb.BatchFindInstancesRequest{
			ConsumerServiceId: serviceId1,
			Services: []*pb.FindService{
				{
					Service: &pb.MicroServiceKey{
						AppId:       "query_instance",
						ServiceName: "noneservice",
					},
				},
			},
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ErrServiceNotExists, respFind.Services.Failed[0].Error.Code)
		assert.Equal(t, int64(0), respFind.Services.Failed[0].Indexes[0])

		respFind, err = discosvc.FindManyInstances(ctx, &pb.BatchFindInstancesRequest{
			ConsumerServiceId: serviceId1,
			Instances: []*pb.FindInstance{
				{
					Instance: &pb.HeartbeatSetElement{
						ServiceId:  serviceId1,
						InstanceId: "noninstance",
					},
				},
			},
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ErrInstanceNotExists, respFind.Instances.Failed[0].Error.Code)
		assert.Equal(t, int64(0), respFind.Instances.Failed[0].Indexes[0])

		respFind, err = discosvc.FindManyInstances(ctx, &pb.BatchFindInstancesRequest{
			ConsumerServiceId: "notExistServiceId",
			Services: []*pb.FindService{
				{
					Service: &pb.MicroServiceKey{
						AppId:       "query_instance",
						ServiceName: "query_instance_service",
					},
				},
			},
		})
		assert.NoError(t, err)
		assert.Equal(t, int64(0), respFind.Services.Failed[0].Indexes[0])
		assert.Equal(t, pb.ErrServiceNotExists, respFind.Services.Failed[0].Error.Code)
	})

	t.Run("when batch query instances, should be ok", func(t *testing.T) {
		respFind, err := discosvc.FindManyInstances(ctx, &pb.BatchFindInstancesRequest{
			ConsumerServiceId: serviceId1,
			Services: []*pb.FindService{
				{
					Service: &pb.MicroServiceKey{
						AppId:       "query_instance",
						ServiceName: "query_instance_service",
					},
				},
				{
					Service: &pb.MicroServiceKey{
						AppId:       "query_instance",
						ServiceName: "query_instance_diff_env_service",
					},
				},
				{
					Service: &pb.MicroServiceKey{
						AppId:       "query_instance_diff_app",
						ServiceName: "query_instance_service",
					},
				},
				{
					Service: &pb.MicroServiceKey{
						AppId:       "not-exists",
						ServiceName: "not-exists",
					},
				},
			},
		})
		assert.NoError(t, err)
		assert.Equal(t, int64(0), respFind.Services.Updated[0].Index)
		assertInstanceContain(t, respFind.Services.Updated[0].Instances, instanceId1)
		assertInstanceContain(t, respFind.Services.Updated[0].Instances, instanceId2)
		assert.Equal(t, int64(2), respFind.Services.Updated[1].Index)
		assert.Empty(t, respFind.Services.Updated[1].Instances)
		assert.Equal(t, 2, len(respFind.Services.Failed[0].Indexes))
		assert.Equal(t, pb.ErrServiceNotExists, respFind.Services.Failed[0].Error.Code)

		respFind, err = discosvc.FindManyInstances(ctx, &pb.BatchFindInstancesRequest{
			ConsumerServiceId: serviceId4,
			Services: []*pb.FindService{
				{
					Service: &pb.MicroServiceKey{
						AppId:       "query_instance",
						ServiceName: "query_instance_diff_env_service",
					},
				},
			},
		})
		assert.NoError(t, err)
		assert.Equal(t, 1, len(respFind.Services.Updated[0].Instances), fmt.Sprintf("%v", respFind))
		assert.Equal(t, instanceId4, respFind.Services.Updated[0].Instances[0].InstanceId)

		respFind, err = discosvc.FindManyInstances(ctx, &pb.BatchFindInstancesRequest{
			Services: []*pb.FindService{
				{
					Service: &pb.MicroServiceKey{
						Environment: pb.ENV_PROD,
						AppId:       "query_instance",
						ServiceName: "query_instance_diff_env_service",
					},
				},
			},
		})
		assert.NoError(t, err)
		assert.Equal(t, 1, len(respFind.Services.Updated[0].Instances), fmt.Sprintf("%v", respFind))
		assert.Equal(t, instanceId4, respFind.Services.Updated[0].Instances[0].InstanceId)

		ctx := util.SetContext(ctx, util.CtxNocache, "")
		respFind, err = discosvc.FindManyInstances(ctx, &pb.BatchFindInstancesRequest{
			ConsumerServiceId: serviceId8,
			Services: []*pb.FindService{
				{
					Service: &pb.MicroServiceKey{
						AppId:       "query_instance",
						ServiceName: "query_instance_with_rev",
					},
				},
				{
					Service: &pb.MicroServiceKey{
						AppId:       "query_instance",
						ServiceName: "batch_query_instance_with_rev",
					},
				},
			},
			Instances: []*pb.FindInstance{
				{
					Instance: &pb.HeartbeatSetElement{
						ServiceId:  serviceId9,
						InstanceId: instanceId9,
					},
				},
				{
					Instance: &pb.HeartbeatSetElement{
						ServiceId:  serviceId8,
						InstanceId: instanceId8,
					},
				},
			},
		})
		assert.NoError(t, err)
		rev := respFind.Services.Updated[0].Rev
		assert.Equal(t, int64(0), respFind.Services.Updated[0].Index)
		assert.Equal(t, int64(1), respFind.Services.Updated[1].Index)
		assert.Equal(t, instanceId8, respFind.Services.Updated[0].Instances[0].InstanceId)
		assert.Equal(t, instanceId9, respFind.Services.Updated[1].Instances[0].InstanceId)
		assert.NotEqual(t, 0, len(rev))
		instanceRev := respFind.Instances.Updated[0].Rev
		assert.Equal(t, int64(0), respFind.Instances.Updated[0].Index)
		assert.Equal(t, int64(1), respFind.Instances.Updated[1].Index)
		assert.Equal(t, instanceId9, respFind.Instances.Updated[0].Instances[0].InstanceId)
		assert.Equal(t, instanceId8, respFind.Instances.Updated[1].Instances[0].InstanceId)
		assert.NotEqual(t, 0, len(instanceRev))

		respFind, err = discosvc.FindManyInstances(ctx, &pb.BatchFindInstancesRequest{
			ConsumerServiceId: serviceId8,
			Services: []*pb.FindService{
				{
					Service: &pb.MicroServiceKey{
						AppId:       "query_instance",
						ServiceName: "query_instance_with_rev",
					},
					Rev: "x",
				},
			},
			Instances: []*pb.FindInstance{
				{
					Instance: &pb.HeartbeatSetElement{
						ServiceId:  serviceId9,
						InstanceId: instanceId9,
					},
					Rev: "x",
				},
			},
		})
		assert.NoError(t, err)
		assert.Equal(t, instanceId8, respFind.Services.Updated[0].Instances[0].InstanceId)
		assert.Equal(t, rev, respFind.Services.Updated[0].Rev)
		assert.Equal(t, instanceId9, respFind.Instances.Updated[0].Instances[0].InstanceId)
		assert.Equal(t, instanceRev, respFind.Instances.Updated[0].Rev)

		respFind, err = discosvc.FindManyInstances(ctx, &pb.BatchFindInstancesRequest{
			ConsumerServiceId: serviceId8,
			Services: []*pb.FindService{
				{
					Service: &pb.MicroServiceKey{
						AppId:       "query_instance",
						ServiceName: "query_instance_with_rev",
					},
					Rev: rev,
				},
			},
			Instances: []*pb.FindInstance{
				{
					Instance: &pb.HeartbeatSetElement{
						ServiceId:  serviceId9,
						InstanceId: instanceId9,
					},
					Rev: instanceRev,
				},
			},
		})
		assert.NoError(t, err)
		assert.Equal(t, int64(0), respFind.Services.NotModified[0])
		assert.Equal(t, int64(0), respFind.Instances.NotModified[0])
	})

	t.Run("find should return 200 even if consumer is diff apps, should ok", func(t *testing.T) {
		respFind, err := discosvc.FindManyInstances(ctx, &pb.BatchFindInstancesRequest{
			ConsumerServiceId: serviceId3,
			Services: []*pb.FindService{
				{
					Service: &pb.MicroServiceKey{
						AppId:       "query_instance",
						ServiceName: "query_instance_service",
					},
				},
			},
		})
		assert.NoError(t, err)
		assert.Equal(t, 0, len(respFind.Services.Updated[0].Instances))
	})

	t.Run("find shared service discovery, should ok", func(t *testing.T) {
		config.Server.Config.GlobalVisible = "query_instance_shared_provider"
		core.RegisterGlobalServices()
		core.Service.Environment = pb.ENV_PROD

		respFind, err := discosvc.FindManyInstances(
			util.SetTargetDomainProject(
				util.SetDomainProject(util.CloneContext(ctx), "user", "user"),
				"default", "default"),
			&pb.BatchFindInstancesRequest{
				ConsumerServiceId: serviceId6,
				Services: []*pb.FindService{
					{
						Service: &pb.MicroServiceKey{
							AppId:       "default",
							ServiceName: "query_instance_shared_provider",
						},
					},
				},
			})
		assert.NoError(t, err)
		assert.Equal(t, 1, len(respFind.Services.Updated[0].Instances))
		assert.Equal(t, instanceId5, respFind.Services.Updated[0].Instances[0].InstanceId)

		respFind, err = discosvc.FindManyInstances(ctx, &pb.BatchFindInstancesRequest{
			ConsumerServiceId: serviceId7,
			Services: []*pb.FindService{
				{
					Service: &pb.MicroServiceKey{
						AppId:       "default",
						ServiceName: "query_instance_shared_provider",
					},
				},
			},
		})
		assert.NoError(t, err)
		assert.Equal(t, 1, len(respFind.Services.Updated[0].Instances))
		assert.Equal(t, instanceId5, respFind.Services.Updated[0].Instances[0].InstanceId)

		respFind, err = discosvc.FindManyInstances(
			util.SetTargetDomainProject(
				util.SetDomainProject(util.CloneContext(ctx), "user", "user"),
				"default", "default"),
			&pb.BatchFindInstancesRequest{
				ConsumerServiceId: serviceId6,
				Instances: []*pb.FindInstance{
					{
						Instance: &pb.HeartbeatSetElement{
							ServiceId:  serviceId5,
							InstanceId: instanceId5,
						},
					},
				},
			})
		assert.NoError(t, err)
		assert.Equal(t, pb.ErrServiceNotExists, respFind.Instances.Failed[0].Error.Code)

		respFind, err = discosvc.FindManyInstances(ctx, &pb.BatchFindInstancesRequest{
			ConsumerServiceId: serviceId7,
			Instances: []*pb.FindInstance{
				{
					Instance: &pb.HeartbeatSetElement{
						ServiceId:  serviceId5,
						InstanceId: instanceId5,
					},
				},
			},
		})
		assert.NoError(t, err)
		assert.Equal(t, 1, len(respFind.Instances.Updated[0].Instances))
		assert.Equal(t, instanceId5, respFind.Instances.Updated[0].Instances[0].InstanceId)

		core.Service.Environment = pb.ENV_DEV
	})
}

func TestFindInstances(t *testing.T) {
	var (
		serviceId1  string
		serviceId2  string
		serviceId3  string
		serviceId4  string
		serviceId5  string
		serviceId6  string
		serviceId7  string
		serviceId8  string
		serviceId9  string
		serviceId10 string
	)
	ctx := getContext()
	defer discosvc.UnregisterManyService(ctx, &pb.DelServicesRequest{ServiceIds: []string{
		serviceId1, serviceId2, serviceId3, serviceId4, serviceId5,
		serviceId7, serviceId8, serviceId9, serviceId10,
	}, Force: true})
	defer discosvc.UnregisterService(util.SetDomainProject(util.CloneContext(ctx), "user", "user"), &pb.DeleteServiceRequest{
		ServiceId: serviceId6, Force: true,
	})

	t.Run("prepare data, should be passed", func(t *testing.T) {
		respCreate, err := discosvc.RegisterService(ctx, &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				AppId:       "query_instance",
				ServiceName: "query_instance_service",
				Version:     "1.0.0",
				Level:       "FRONT",
				Status:      pb.MS_UP,
			},
		})
		assert.NoError(t, err)
		serviceId1 = respCreate.ServiceId

		respCreate, err = discosvc.RegisterService(ctx, &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				AppId:       "query_instance",
				ServiceName: "query_instance_service",
				Version:     "1.0.5",
				Level:       "FRONT",
				Status:      pb.MS_UP,
			},
		})
		assert.NoError(t, err)
		serviceId2 = respCreate.ServiceId

		respCreate, err = discosvc.RegisterService(ctx, &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				AppId:       "query_instance_diff_app",
				ServiceName: "query_instance_service",
				Version:     "1.0.0",
				Level:       "FRONT",
				Status:      pb.MS_UP,
			},
		})
		assert.NoError(t, err)
		serviceId3 = respCreate.ServiceId

		respCreate, err = discosvc.RegisterService(ctx, &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				Environment: pb.ENV_PROD,
				AppId:       "query_instance",
				ServiceName: "find_instance_diff_env_service",
				Version:     "1.0.0",
				Level:       "FRONT",
				Status:      pb.MS_UP,
			},
		})
		assert.NoError(t, err)
		serviceId4 = respCreate.ServiceId

		respCreate, err = discosvc.RegisterService(ctx, &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				Environment: pb.ENV_PROD,
				AppId:       "default",
				ServiceName: "query_instance_shared_provider",
				Version:     "1.0.0",
				Level:       "FRONT",
				Status:      pb.MS_UP,
				Properties: map[string]string{
					pb.PropAllowCrossApp: "true",
				},
			},
		})
		assert.NoError(t, err)
		serviceId5 = respCreate.ServiceId

		respCreate, err = discosvc.RegisterService(
			util.SetDomainProject(util.CloneContext(ctx), "user", "user"),
			&pb.CreateServiceRequest{
				Service: &pb.MicroService{
					AppId:       "default",
					ServiceName: "query_instance_diff_domain_consumer",
					Version:     "1.0.0",
					Level:       "FRONT",
					Status:      pb.MS_UP,
				},
			})
		assert.NoError(t, err)
		serviceId6 = respCreate.ServiceId

		respCreate, err = discosvc.RegisterService(ctx, &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				AppId:       "default",
				ServiceName: "query_instance_shared_consumer",
				Version:     "1.0.0",
				Level:       "FRONT",
				Status:      pb.MS_UP,
			},
		})
		assert.NoError(t, err)
		serviceId7 = respCreate.ServiceId

		respCreate, err = discosvc.RegisterService(ctx, &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				AppId:       "query_instance",
				ServiceName: "find_instance_with_rev",
				Version:     "1.0.0",
				Level:       "FRONT",
				Status:      pb.MS_UP,
			},
		})
		assert.NoError(t, err)
		serviceId8 = respCreate.ServiceId

		respCreate, err = discosvc.RegisterService(ctx, &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				AppId:       "query_instance",
				ServiceName: "batch_query_instance_with_rev",
				Version:     "1.0.0",
				Level:       "FRONT",
				Status:      pb.MS_UP,
			},
		})
		assert.NoError(t, err)
		serviceId9 = respCreate.ServiceId

		respCreate, err = discosvc.RegisterService(ctx, &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				AppId:       "query_instance_alias",
				ServiceName: "query_instance_alias",
				Alias:       "query_instance_alias:query_instance_alias",
				Version:     "1.0.0",
				Level:       "FRONT",
				Status:      pb.MS_UP,
			},
		})
		assert.NoError(t, err)
		serviceId10 = respCreate.ServiceId

		_, err = discosvc.RegisterInstance(ctx, &pb.RegisterInstanceRequest{
			Instance: &pb.MicroServiceInstance{
				ServiceId: serviceId1,
				HostName:  "UT-HOST",
				Endpoints: []string{
					"find:127.0.0.1:8080",
				},
				Status: pb.MSI_UP,
			},
		})
		assert.NoError(t, err)

		_, err = discosvc.RegisterInstance(ctx, &pb.RegisterInstanceRequest{
			Instance: &pb.MicroServiceInstance{
				ServiceId: serviceId2,
				HostName:  "UT-HOST",
				Endpoints: []string{
					"find:127.0.0.2:8080",
				},
				Status: pb.MSI_UP,
			},
		})
		assert.NoError(t, err)

		_, err = discosvc.RegisterInstance(ctx, &pb.RegisterInstanceRequest{
			Instance: &pb.MicroServiceInstance{
				ServiceId: serviceId4,
				HostName:  "UT-HOST",
				Endpoints: []string{
					"find:127.0.0.4:8080",
				},
				Status: pb.MSI_UP,
			},
		})
		assert.NoError(t, err)

		_, err = discosvc.RegisterInstance(ctx, &pb.RegisterInstanceRequest{
			Instance: &pb.MicroServiceInstance{
				ServiceId: serviceId5,
				HostName:  "UT-HOST",
				Endpoints: []string{
					"find:127.0.0.5:8080",
				},
				Status: pb.MSI_UP,
			},
		})
		assert.NoError(t, err)

		_, err = discosvc.RegisterInstance(ctx, &pb.RegisterInstanceRequest{
			Instance: &pb.MicroServiceInstance{
				ServiceId: serviceId8,
				HostName:  "UT-HOST",
				Endpoints: []string{
					"find:127.0.0.8:8080",
				},
				Status: pb.MSI_UP,
			},
		})
		assert.NoError(t, err)

		_, err = discosvc.RegisterInstance(ctx, &pb.RegisterInstanceRequest{
			Instance: &pb.MicroServiceInstance{
				ServiceId: serviceId9,
				HostName:  "UT-HOST",
				Endpoints: []string{
					"find:127.0.0.9:8080",
				},
				Status: pb.MSI_UP,
			},
		})
		assert.NoError(t, err)

		_, err = discosvc.RegisterInstance(ctx, &pb.RegisterInstanceRequest{
			Instance: &pb.MicroServiceInstance{
				ServiceId: serviceId10,
				HostName:  "UT-HOST",
				Endpoints: []string{
					"find:127.0.0.9:8080",
				},
				Status: pb.MSI_UP,
			},
		})
		assert.NoError(t, err)
	})

	t.Run("when query invalid parameters, should be failed", func(t *testing.T) {
		_, err := discosvc.FindInstances(ctx, &pb.FindInstancesRequest{
			ConsumerServiceId: serviceId1,
			AppId:             TooLongAppID,
			ServiceName:       "query_instance_service",
		})
		testErr := err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrInvalidParams, testErr.Code)

		_, err = discosvc.FindInstances(ctx, &pb.FindInstancesRequest{
			ConsumerServiceId: serviceId1,
			AppId:             "",
			ServiceName:       "query_instance_service",
		})
		testErr = err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrInvalidParams, testErr.Code)

		_, err = discosvc.FindInstances(ctx, &pb.FindInstancesRequest{
			ConsumerServiceId: serviceId1,
			AppId:             " ",
			ServiceName:       "query_instance_service",
		})
		testErr = err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrInvalidParams, testErr.Code)

		_, err = discosvc.FindInstances(ctx, &pb.FindInstancesRequest{
			ConsumerServiceId: serviceId1,
			AppId:             "query_instance",
			ServiceName:       TooLongExistence,
		})
		testErr = err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrInvalidParams, testErr.Code)

		_, err = discosvc.FindInstances(ctx, &pb.FindInstancesRequest{
			ConsumerServiceId: serviceId1,
			AppId:             "query_instance",
			ServiceName:       "",
		})
		testErr = err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrInvalidParams, testErr.Code)

		_, err = discosvc.FindInstances(ctx, &pb.FindInstancesRequest{
			ConsumerServiceId: serviceId1,
			AppId:             "query_instance",
			ServiceName:       " ",
		})
		testErr = err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrInvalidParams, testErr.Code)

		_, err = discosvc.FindInstances(ctx, &pb.FindInstancesRequest{
			ConsumerServiceId: serviceId1,
			AppId:             "query_instance",
			ServiceName:       "noneservice",
		})
		testErr = err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrServiceNotExists, testErr.Code)

		_, err = discosvc.FindInstances(ctx, &pb.FindInstancesRequest{
			ConsumerServiceId: "notExistServiceId",
			AppId:             "query_instance",
			ServiceName:       "query_instance_service",
		})
		testErr = err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrServiceNotExists, testErr.Code)
	})

	t.Run("when query instances between diff dimensions", func(t *testing.T) {
		UTFunc := func(consumerId string, code int32) {
			_, err := discosvc.ListInstance(ctx, &pb.GetInstancesRequest{
				ConsumerServiceId: consumerId,
				ProviderServiceId: serviceId2,
			})
			if code == pb.ResponseSuccess {
				assert.NoError(t, err)
				return
			}
			testErr := err.(*errsvc.Error)
			assert.Error(t, testErr)
			assert.Equal(t, code, testErr.Code)
		}

		UTFunc(serviceId3, pb.ErrServiceNotExists)

		UTFunc(serviceId1, pb.ResponseSuccess)

		_, err := discosvc.ListInstance(ctx, &pb.GetInstancesRequest{
			ConsumerServiceId: serviceId4,
			ProviderServiceId: serviceId2,
		})
		testErr := err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrServiceNotExists, testErr.Code)
	})
}

func TestAppendFindResponse(t *testing.T) {
	ctx := context.Background()
	var (
		find              pb.FindInstancesResponse
		updatedResult     []*pb.FindResult
		notModifiedResult []int64
		failedResult      *pb.FindFailedResult
	)
	discosvc.AppendFindResponse(ctx, 1, find.Response, find.Instances, &updatedResult, &notModifiedResult, &failedResult)
	assert.NotNil(t, updatedResult)
	assert.Nil(t, notModifiedResult)
	assert.Nil(t, failedResult)
	assert.Equal(t, int64(1), updatedResult[0].Index)

	updatedResult = nil
	cloneCtx := util.WithResponseRev(ctx, "1")
	discosvc.AppendFindResponse(cloneCtx, 1, find.Response, find.Instances, &updatedResult, &notModifiedResult, &failedResult)
	assert.NotNil(t, updatedResult)
	assert.Nil(t, notModifiedResult)
	assert.Nil(t, failedResult)
	assert.Equal(t, int64(1), updatedResult[0].Index)
	assert.Equal(t, "1", updatedResult[0].Rev)

	updatedResult = nil
	cloneCtx = util.WithRequestRev(ctx, "1")
	cloneCtx = util.WithResponseRev(cloneCtx, "1")
	discosvc.AppendFindResponse(cloneCtx, 1, find.Response, find.Instances, &updatedResult, &notModifiedResult, &failedResult)
	if updatedResult != nil || notModifiedResult == nil || failedResult != nil {
		t.Fatal("TestAppendFindResponse failed")
	}
	assert.Nil(t, updatedResult)
	assert.NotNil(t, notModifiedResult)
	assert.Nil(t, failedResult)
	assert.Equal(t, int64(1), notModifiedResult[0])

	notModifiedResult = nil
	find.Response = pb.CreateResponse(pb.ErrInternal, "test")
	discosvc.AppendFindResponse(ctx, 1, find.Response, find.Instances, &updatedResult, &notModifiedResult, &failedResult)
	assert.Nil(t, updatedResult)
	assert.Nil(t, notModifiedResult)
	assert.NotNil(t, failedResult)
	assert.Equal(t, pb.ErrInternal, failedResult.Error.Code)

	find.Response = pb.CreateResponse(pb.ErrInvalidParams, "test")
	discosvc.AppendFindResponse(ctx, 2, find.Response, find.Instances, &updatedResult, &notModifiedResult, &failedResult)
	assert.Nil(t, updatedResult)
	assert.Nil(t, notModifiedResult)
	assert.NotNil(t, failedResult)
	assert.Equal(t, pb.ErrInternal, failedResult.Error.Code)

	failedResult = nil
	find.Response = nil
	discosvc.AppendFindResponse(ctx, 1, find.Response, find.Instances, &updatedResult, &notModifiedResult, &failedResult)
	discosvc.AppendFindResponse(ctx, 2, find.Response, find.Instances, &updatedResult, &notModifiedResult, &failedResult)
	cloneCtx = util.WithRequestRev(ctx, "1")
	cloneCtx = util.WithResponseRev(cloneCtx, "1")
	discosvc.AppendFindResponse(cloneCtx, 3, find.Response, find.Instances, &updatedResult, &notModifiedResult, &failedResult)
	discosvc.AppendFindResponse(cloneCtx, 4, find.Response, find.Instances, &updatedResult, &notModifiedResult, &failedResult)
	find.Response = pb.CreateResponse(pb.ErrInternal, "test")
	discosvc.AppendFindResponse(ctx, 5, find.Response, find.Instances, &updatedResult, &notModifiedResult, &failedResult)
	discosvc.AppendFindResponse(ctx, 6, find.Response, find.Instances, &updatedResult, &notModifiedResult, &failedResult)
	assert.NotNil(t, updatedResult)
	assert.NotNil(t, notModifiedResult)
	assert.NotNil(t, failedResult)
	assert.Equal(t, 2, len(updatedResult))
	assert.Equal(t, 2, len(notModifiedResult))
	assert.Equal(t, 2, len(failedResult.Indexes))
}

func TestGetInstance(t *testing.T) {
	var (
		serviceId1  string
		serviceId2  string
		serviceId3  string
		instanceId2 string
	)
	ctx := getContext()
	defer discosvc.UnregisterManyService(ctx, &pb.DelServicesRequest{ServiceIds: []string{
		serviceId1, serviceId2, serviceId3,
	}, Force: true})

	t.Run("prepare data, should be passed", func(t *testing.T) {
		respCreate, err := discosvc.RegisterService(ctx, &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				AppId:       "get_instance",
				ServiceName: "get_instance_service",
				Version:     "1.0.0",
				Level:       "FRONT",
				Status:      pb.MS_UP,
			},
		})
		assert.NoError(t, err)
		serviceId1 = respCreate.ServiceId

		respCreate, err = discosvc.RegisterService(ctx, &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				AppId:       "get_instance",
				ServiceName: "get_instance_service",
				Version:     "1.0.5",
				Level:       "FRONT",
				Status:      pb.MS_UP,
			},
			Tags: map[string]string{
				"test": "test",
			},
		})
		assert.NoError(t, err)
		serviceId2 = respCreate.ServiceId

		resp, err := discosvc.RegisterInstance(ctx, &pb.RegisterInstanceRequest{
			Instance: &pb.MicroServiceInstance{
				ServiceId: serviceId2,
				HostName:  "UT-HOST",
				Endpoints: []string{
					"get:127.0.0.2:8080",
				},
				Status: pb.MSI_UP,
			},
		})
		assert.NoError(t, err)
		instanceId2 = resp.InstanceId

		respCreate, err = discosvc.RegisterService(ctx, &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				AppId:       "get_instance_cross",
				ServiceName: "get_instance_service",
				Version:     "1.0.0",
				Level:       "FRONT",
				Status:      pb.MS_UP,
			},
		})
		assert.NoError(t, err)
		serviceId3 = respCreate.ServiceId
	})

	t.Run("when get one instance request is invalid, should be failed", func(t *testing.T) {
		_, err := discosvc.GetInstance(ctx, &pb.GetOneInstanceRequest{
			ConsumerServiceId:  serviceId2,
			ProviderServiceId:  serviceId2,
			ProviderInstanceId: instanceId2,
		})
		assert.NoError(t, err)

		_, err = discosvc.GetInstance(ctx, &pb.GetOneInstanceRequest{
			ConsumerServiceId:  serviceId1,
			ProviderServiceId:  "",
			ProviderInstanceId: instanceId2,
		})
		testErr := err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrInvalidParams, testErr.Code)

		_, err = discosvc.GetInstance(ctx, &pb.GetOneInstanceRequest{
			ConsumerServiceId:  serviceId1,
			ProviderServiceId:  serviceId2,
			ProviderInstanceId: "",
		})
		testErr = err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrInvalidParams, testErr.Code)

		_, err = discosvc.GetInstance(ctx, &pb.GetOneInstanceRequest{
			ConsumerServiceId:  "",
			ProviderServiceId:  serviceId2,
			ProviderInstanceId: instanceId2,
		})
		assert.NoError(t, err)

		_, err = discosvc.GetInstance(ctx, &pb.GetOneInstanceRequest{
			ConsumerServiceId:  "not-exist-id",
			ProviderServiceId:  serviceId2,
			ProviderInstanceId: instanceId2,
		})
		testErr = err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrServiceNotExists, testErr.Code)

		_, err = discosvc.GetInstance(ctx, &pb.GetOneInstanceRequest{
			ConsumerServiceId:  serviceId1,
			ProviderServiceId:  serviceId2,
			ProviderInstanceId: instanceId2,
			Tags:               []string{"not-exist-tag"},
		})
		testErr = err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrInstanceNotExists, testErr.Code)

		_, err = discosvc.GetInstance(ctx,
			&pb.GetOneInstanceRequest{
				ConsumerServiceId:  serviceId1,
				ProviderServiceId:  serviceId2,
				ProviderInstanceId: instanceId2,
				Tags:               []string{"test"},
			})
		assert.NoError(t, err)
	})

	t.Run("when get between diff apps, should be failed", func(t *testing.T) {
		_, err := discosvc.GetInstance(ctx, &pb.GetOneInstanceRequest{
			ConsumerServiceId:  serviceId3,
			ProviderServiceId:  serviceId2,
			ProviderInstanceId: instanceId2,
		})
		testErr := err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrInstanceNotExists, testErr.Code)

		_, err = discosvc.ListInstance(ctx, &pb.GetInstancesRequest{
			ConsumerServiceId: serviceId3,
			ProviderServiceId: serviceId2,
		})
		testErr = err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrServiceNotExists, testErr.Code)
	})

	t.Run("when get instances request is invalid, should be failed", func(t *testing.T) {
		_, err := discosvc.ListInstance(ctx, &pb.GetInstancesRequest{
			ConsumerServiceId: "",
			ProviderServiceId: serviceId2,
		})
		assert.NoError(t, err)

		_, err = discosvc.ListInstance(ctx, &pb.GetInstancesRequest{
			ConsumerServiceId: "noneservice",
			ProviderServiceId: serviceId2,
		})
		testErr := err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrServiceNotExists, testErr.Code)

		resp, err := discosvc.ListInstance(ctx, &pb.GetInstancesRequest{
			ConsumerServiceId: serviceId1,
			ProviderServiceId: serviceId2,
		})
		assert.NoError(t, err)
		assert.NotNil(t, resp)
	})
}

func TestUnregisterInstance(t *testing.T) {
	var (
		serviceId  string
		instanceId string
	)
	ctx := getContext()

	t.Run("prepare data, should be passed", func(t *testing.T) {
		respCreate, err := discosvc.RegisterService(ctx, &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				AppId:       "unregister_instance",
				ServiceName: "unregister_instance_service",
				Version:     "1.0.5",
				Level:       "FRONT",
				Status:      pb.MS_UP,
			},
			Tags: map[string]string{
				"test": "test",
			},
		})
		assert.NoError(t, err)
		serviceId = respCreate.ServiceId

		resp, err := discosvc.RegisterInstance(ctx, &pb.RegisterInstanceRequest{
			Instance: &pb.MicroServiceInstance{
				ServiceId: serviceId,
				HostName:  "UT-HOST",
				Endpoints: []string{
					"unregister:127.0.0.2:8080",
				},
				Status: pb.MSI_UP,
			},
		})
		assert.NoError(t, err)
		instanceId = resp.InstanceId
	})

	t.Run("when request is valid, should be passed", func(t *testing.T) {
		err := discosvc.UnregisterInstance(ctx, &pb.UnregisterInstanceRequest{
			ServiceId:  serviceId,
			InstanceId: instanceId,
		})
		assert.NoError(t, err)
	})

	t.Run("when request is invalid, should be failed", func(t *testing.T) {
		err := discosvc.UnregisterInstance(ctx, &pb.UnregisterInstanceRequest{
			ServiceId:  "",
			InstanceId: instanceId,
		})
		testErr := err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrInvalidParams, testErr.Code)

		err = discosvc.UnregisterInstance(ctx, &pb.UnregisterInstanceRequest{
			ServiceId:  TooLongServiceID,
			InstanceId: instanceId,
		})
		testErr = err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrInvalidParams, testErr.Code)

		err = discosvc.UnregisterInstance(ctx, &pb.UnregisterInstanceRequest{
			ServiceId:  "not-exist-id",
			InstanceId: instanceId,
		})
		testErr = err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrInstanceNotExists, testErr.Code)

		err = discosvc.UnregisterInstance(ctx, &pb.UnregisterInstanceRequest{
			ServiceId:  serviceId,
			InstanceId: "",
		})
		testErr = err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrInvalidParams, testErr.Code)

		err = discosvc.UnregisterInstance(ctx, &pb.UnregisterInstanceRequest{
			ServiceId:  serviceId,
			InstanceId: "@",
		})
		testErr = err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrInvalidParams, testErr.Code)

		err = discosvc.UnregisterInstance(ctx, &pb.UnregisterInstanceRequest{
			ServiceId:  serviceId,
			InstanceId: TooLongServiceID,
		})
		testErr = err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrInvalidParams, testErr.Code)

		err = discosvc.UnregisterInstance(ctx, &pb.UnregisterInstanceRequest{
			ServiceId:  serviceId,
			InstanceId: "not-exist-id",
		})
		testErr = err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrInstanceNotExists, testErr.Code)
	})
}

func TestSendHeartbeat(t *testing.T) {
	var (
		serviceId   string
		instanceId1 string
		instanceId2 string
	)
	ctx := getContext()
	defer discosvc.UnregisterService(ctx, &pb.DeleteServiceRequest{ServiceId: serviceId, Force: true})

	t.Run("prepare data, should be passed", func(t *testing.T) {
		respCreate, err := discosvc.RegisterService(ctx, &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				ServiceName: "heartbeat_service",
				AppId:       "heartbeat_service",
				Version:     "1.0.0",
				Level:       "FRONT",
				Status:      pb.MS_UP,
			},
		})
		assert.NoError(t, err)
		serviceId = respCreate.ServiceId

		resp, err := discosvc.RegisterInstance(ctx, &pb.RegisterInstanceRequest{
			Instance: &pb.MicroServiceInstance{
				ServiceId: serviceId,
				HostName:  "UT-HOST",
				Endpoints: []string{
					"heartbeat:127.0.0.1:8080",
				},
				Status: pb.MSI_UP,
			},
		})
		assert.NoError(t, err)

		instanceId1 = resp.InstanceId

		resp, err = discosvc.RegisterInstance(ctx, &pb.RegisterInstanceRequest{
			Instance: &pb.MicroServiceInstance{
				ServiceId: serviceId,
				HostName:  "UT-HOST",
				Endpoints: []string{
					"heartbeat:127.0.0.2:8080",
				},
				Status: pb.MSI_UP,
			},
		})
		assert.NoError(t, err)
		instanceId2 = resp.InstanceId
	})

	t.Run("when update a lease, should be passed", func(t *testing.T) {
		resp, err := discosvc.GetInstance(ctx, &pb.GetOneInstanceRequest{ProviderServiceId: serviceId, ProviderInstanceId: instanceId1})
		assert.NoError(t, err)
		resp.Instance.Properties = nil

		err = discosvc.PutInstance(ctx, &pb.RegisterInstanceRequest{Instance: resp.Instance})
		assert.NoError(t, err)

		err = discosvc.SendHeartbeat(ctx, &pb.HeartbeatRequest{
			ServiceId:  serviceId,
			InstanceId: instanceId1,
		})
		assert.NoError(t, err)

		resp, err = discosvc.GetInstance(ctx, &pb.GetOneInstanceRequest{ProviderServiceId: serviceId, ProviderInstanceId: instanceId1})
		assert.NoError(t, err)
		assert.Equal(t, "test_engineID", resp.Instance.Properties["engineID"])
	})

	t.Run("when update lease with invalid request, should be failed", func(t *testing.T) {
		err := discosvc.SendHeartbeat(ctx, &pb.HeartbeatRequest{
			ServiceId:  "",
			InstanceId: instanceId1,
		})
		testErr := err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrInvalidParams, testErr.Code)

		err = discosvc.SendHeartbeat(ctx, &pb.HeartbeatRequest{
			ServiceId:  TooLongServiceID,
			InstanceId: instanceId1,
		})
		testErr = err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrInvalidParams, testErr.Code)

		err = discosvc.SendHeartbeat(ctx, &pb.HeartbeatRequest{
			ServiceId:  serviceId,
			InstanceId: "",
		})
		testErr = err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrInvalidParams, testErr.Code)

		err = discosvc.SendHeartbeat(ctx, &pb.HeartbeatRequest{
			ServiceId:  serviceId,
			InstanceId: "@",
		})
		testErr = err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrInvalidParams, testErr.Code)

		err = discosvc.SendHeartbeat(ctx, &pb.HeartbeatRequest{
			ServiceId:  serviceId,
			InstanceId: TooLongServiceID,
		})
		testErr = err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrInvalidParams, testErr.Code)

		err = discosvc.SendHeartbeat(ctx, &pb.HeartbeatRequest{
			ServiceId:  "100000000000",
			InstanceId: instanceId1,
		})
		testErr = err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrInstanceNotExists, testErr.Code)

		err = discosvc.SendHeartbeat(ctx, &pb.HeartbeatRequest{
			ServiceId:  serviceId,
			InstanceId: "not-exist-ins",
		})
		testErr = err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrInstanceNotExists, testErr.Code)
	})

	t.Run("when batch update lease, should be passed", func(t *testing.T) {
		resp, err := discosvc.SendManyHeartbeat(ctx, &pb.HeartbeatSetRequest{
			Instances: []*pb.HeartbeatSetElement{},
		})
		testErr := err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrInvalidParams, testErr.Code)

		resp, err = discosvc.SendManyHeartbeat(ctx, &pb.HeartbeatSetRequest{})
		testErr = err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrInvalidParams, testErr.Code)

		resp, err = discosvc.SendManyHeartbeat(ctx, &pb.HeartbeatSetRequest{
			Instances: []*pb.HeartbeatSetElement{
				{
					ServiceId:  serviceId,
					InstanceId: instanceId1,
				},
				{
					ServiceId:  serviceId,
					InstanceId: instanceId2,
				},
			},
		})
		assert.NoError(t, err)
		assert.Equal(t, 2, len(resp.Instances))

		resp, err = discosvc.SendManyHeartbeat(ctx, &pb.HeartbeatSetRequest{
			Instances: []*pb.HeartbeatSetElement{
				{
					ServiceId:  serviceId,
					InstanceId: instanceId1,
				},
				{
					ServiceId:  serviceId,
					InstanceId: "not-exist-instanceId",
				},
			},
		})
		assert.NoError(t, err)
		for _, result := range resp.Instances {
			if result.InstanceId == "not-exist-instanceId" {
				assert.NotEmpty(t, result.ErrMessage)
			}
		}
	})
}

func TestClusterHealth(t *testing.T) {
	ctx := getContext()

	t.Run("prepare data, should be passed", func(t *testing.T) {
		_, err := discosvc.RegisterService(ctx, core.CreateServiceRequest())
		assert.NoError(t, err)
	})

	t.Run("when SC does not exist, should be failed", func(t *testing.T) {
		old := core.Service.ServiceName
		core.Service.ServiceName = "x"
		_, err := discosvc.ClusterHealth(ctx)
		testErr := err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrServiceNotExists, testErr.Code)
		core.Service.ServiceName = old
	})

	t.Run("when SC registered, should be passed", func(t *testing.T) {
		respClusterHealth, err := discosvc.ClusterHealth(ctx)
		assert.NoError(t, err)
		assert.NotNil(t, respClusterHealth)
	})
}

func TestUpdateInstance(t *testing.T) {
	var (
		serviceId  string
		instanceId string
	)
	ctx := getContext()
	defer discosvc.UnregisterService(ctx, &pb.DeleteServiceRequest{ServiceId: serviceId, Force: true})

	respCreate, err := discosvc.RegisterService(ctx, &pb.CreateServiceRequest{
		Service: &pb.MicroService{
			ServiceName: "update_instance_service",
			AppId:       "update_instance_service",
			Version:     "1.0.0",
			Level:       "FRONT",
			Status:      pb.MS_UP,
		},
	})
	assert.NoError(t, err)
	serviceId = respCreate.ServiceId

	registerInstanceRequest := &pb.RegisterInstanceRequest{
		Instance: &pb.MicroServiceInstance{
			ServiceId: serviceId,
			Endpoints: []string{
				"updateInstance:127.0.0.1:8080",
			},
			HostName:   "UT-HOST",
			Status:     pb.MSI_UP,
			Properties: map[string]string{"nodeIP": "test"},
		},
	}

	resp, err := discosvc.RegisterInstance(ctx, registerInstanceRequest)
	assert.NoError(t, err)
	instanceId = resp.InstanceId

	t.Run("when update instance status, should be passed", func(t *testing.T) {
		err := discosvc.PutInstanceStatus(ctx, &pb.UpdateInstanceStatusRequest{
			ServiceId:  serviceId,
			InstanceId: instanceId,
			Status:     pb.MSI_DOWN,
		})
		assert.NoError(t, err)

		err = discosvc.PutInstanceStatus(ctx, &pb.UpdateInstanceStatusRequest{
			ServiceId:  serviceId,
			InstanceId: instanceId,
			Status:     pb.MSI_OUTOFSERVICE,
		})
		assert.NoError(t, err)

		err = discosvc.PutInstanceStatus(ctx, &pb.UpdateInstanceStatusRequest{
			ServiceId:  serviceId,
			InstanceId: instanceId,
			Status:     pb.MSI_STARTING,
		})
		assert.NoError(t, err)

		err = discosvc.PutInstanceStatus(ctx, &pb.UpdateInstanceStatusRequest{
			ServiceId:  serviceId,
			InstanceId: instanceId,
			Status:     pb.MSI_TESTING,
		})
		assert.NoError(t, err)

		err = discosvc.PutInstanceStatus(ctx, &pb.UpdateInstanceStatusRequest{
			ServiceId:  serviceId,
			InstanceId: instanceId,
			Status:     pb.MSI_UP,
		})
		assert.NoError(t, err)

		err = discosvc.PutInstanceStatus(ctx, &pb.UpdateInstanceStatusRequest{
			ServiceId:  "",
			InstanceId: instanceId,
			Status:     pb.MSI_STARTING,
		})
		testErr := err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrInvalidParams, testErr.Code)

		err = discosvc.PutInstanceStatus(ctx, &pb.UpdateInstanceStatusRequest{
			ServiceId:  "notexistservice",
			InstanceId: instanceId,
			Status:     pb.MSI_STARTING,
		})
		testErr = err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrInstanceNotExists, testErr.Code)

		err = discosvc.PutInstanceStatus(ctx, &pb.UpdateInstanceStatusRequest{
			ServiceId:  serviceId,
			InstanceId: "notexistins",
			Status:     pb.MSI_STARTING,
		})
		testErr = err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrInstanceNotExists, testErr.Code)

		err = discosvc.PutInstanceStatus(ctx, &pb.UpdateInstanceStatusRequest{
			ServiceId:  serviceId,
			InstanceId: instanceId,
			Status:     "nonestatus",
		})
		testErr = err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrInvalidParams, testErr.Code)
	})

	t.Run("when update instance properties, should be passed", func(t *testing.T) {
		err := discosvc.PutInstanceProperties(ctx, &pb.UpdateInstancePropsRequest{
			ServiceId:  serviceId,
			InstanceId: instanceId,
			Properties: map[string]string{
				"test": "test",
			},
		})
		assert.NoError(t, err)

		resp, err := discosvc.GetInstance(ctx, &pb.GetOneInstanceRequest{ProviderServiceId: serviceId, ProviderInstanceId: instanceId})
		assert.NoError(t, err)
		assert.Equal(t, "test", resp.Instance.Properties["test"])

		size := 1000
		properties := make(map[string]string, size)
		for i := 0; i < size; i++ {
			s := strconv.Itoa(i) + strings.Repeat("x", 253)
			properties[s] = s
		}
		err = discosvc.PutInstanceProperties(ctx, &pb.UpdateInstancePropsRequest{
			ServiceId:  serviceId,
			InstanceId: instanceId,
			Properties: properties,
		})
		assert.NoError(t, err)

		err = discosvc.PutInstanceProperties(ctx, &pb.UpdateInstancePropsRequest{
			ServiceId:  serviceId,
			InstanceId: instanceId,
		})
		assert.NoError(t, err)

		resp, err = discosvc.GetInstance(ctx, &pb.GetOneInstanceRequest{ProviderServiceId: serviceId, ProviderInstanceId: instanceId})
		assert.NoError(t, err)
		_, ok := resp.Instance.Properties["test"]
		assert.False(t, ok)
	})

	t.Run("when update instance properties with invalid request, should be failed", func(t *testing.T) {
		err = discosvc.PutInstanceProperties(ctx, &pb.UpdateInstancePropsRequest{
			ServiceId:  serviceId,
			InstanceId: "notexistins",
			Properties: map[string]string{
				"test": "test",
			},
		})
		testErr := err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrInstanceNotExists, testErr.Code)

		err = discosvc.PutInstanceProperties(ctx, &pb.UpdateInstancePropsRequest{
			ServiceId:  "",
			InstanceId: instanceId,
			Properties: map[string]string{
				"test": "test",
			},
		})
		testErr = err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrInvalidParams, testErr.Code)

		err = discosvc.PutInstanceProperties(ctx, &pb.UpdateInstancePropsRequest{
			ServiceId:  TooLongServiceID,
			InstanceId: instanceId,
			Properties: map[string]string{
				"test": "test",
			},
		})
		testErr = err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrInvalidParams, testErr.Code)

		err = discosvc.PutInstanceProperties(ctx, &pb.UpdateInstancePropsRequest{
			ServiceId:  serviceId,
			InstanceId: "",
			Properties: map[string]string{
				"test": "test",
			},
		})
		testErr = err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrInvalidParams, testErr.Code)

		err = discosvc.PutInstanceProperties(ctx, &pb.UpdateInstancePropsRequest{
			ServiceId:  serviceId,
			InstanceId: "@",
			Properties: map[string]string{
				"test": "test",
			},
		})
		testErr = err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrInvalidParams, testErr.Code)

		err = discosvc.PutInstanceProperties(ctx, &pb.UpdateInstancePropsRequest{
			ServiceId:  serviceId,
			InstanceId: TooLongServiceID,
			Properties: map[string]string{
				"test": "test",
			},
		})
		testErr = err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrInvalidParams, testErr.Code)

		err = discosvc.PutInstanceProperties(ctx, &pb.UpdateInstancePropsRequest{
			ServiceId:  "notexistservice",
			InstanceId: instanceId,
			Properties: map[string]string{
				"test": "test",
			},
		})
		testErr = err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrInstanceNotExists, testErr.Code)
	})

	t.Run("when update instance with valid request, should be passed", func(t *testing.T) {
		registerInstanceRequest.Instance.HostName = "updated"
		err := discosvc.PutInstance(ctx, registerInstanceRequest)
		assert.NoError(t, err)

		resp, err := discosvc.GetInstance(ctx, &pb.GetOneInstanceRequest{
			ProviderServiceId: serviceId, ProviderInstanceId: instanceId,
		})
		assert.NoError(t, err)
		assert.Equal(t, "updated", resp.Instance.HostName)
	})

	t.Run("when update instance with invalid request, should be failed", func(t *testing.T) {
		err = discosvc.PutInstance(ctx, &pb.RegisterInstanceRequest{})
		testErr := err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrInvalidParams, testErr.Code)

		err = discosvc.PutInstance(ctx, &pb.RegisterInstanceRequest{
			Instance: &pb.MicroServiceInstance{},
		})
		testErr = err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrInvalidParams, testErr.Code)
	})
}

func assertInstanceContain(t *testing.T, instances []*pb.MicroServiceInstance, instanceID string) {
	found := false
	for _, instance := range instances {
		if instance.InstanceId == instanceID {
			found = true
			break
		}
	}
	assert.True(t, found)
}

func TestInstanceUsage(t *testing.T) {
	t.Run("get domain/project without instance usage, should return 0", func(t *testing.T) {
		usage, err := discosvc.InstanceUsage(context.Background(), &pb.GetServiceCountRequest{
			Domain:  "domain_without_service",
			Project: "project_without_service",
		})
		assert.NoError(t, err)
		assert.Equal(t, int64(0), usage)
	})

	t.Run("get domain/project with 1 instance usage, should return 1", func(t *testing.T) {
		ctx := util.SetDomainProject(context.Background(), "domain_with_service", "project_with_service")
		resp, err := discosvc.RegisterService(ctx, &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				ServiceName: "test",
			},
		})
		assert.NoError(t, err)
		defer discosvc.UnregisterService(ctx, &pb.DeleteServiceRequest{ServiceId: resp.ServiceId, Force: true})

		_, err = discosvc.RegisterInstance(ctx, &pb.RegisterInstanceRequest{Instance: &pb.MicroServiceInstance{
			ServiceId: resp.ServiceId,
			HostName:  "test",
		}})
		assert.NoError(t, err)

		usage, err := discosvc.InstanceUsage(context.Background(), &pb.GetServiceCountRequest{
			Domain:  "domain_with_service",
			Project: "project_with_service",
		})
		assert.NoError(t, err)
		assert.Equal(t, int64(1), usage)
	})
}
