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
	"strconv"
	"strings"
	"testing"

	"github.com/apache/servicecomb-service-center/datasource"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/util"
	"github.com/apache/servicecomb-service-center/server/config"
	"github.com/apache/servicecomb-service-center/server/core"
	pb "github.com/go-chassis/cari/discovery"
	"github.com/go-chassis/cari/pkg/errsvc"
	"github.com/stretchr/testify/assert"
)

func TestInstance_Create(t *testing.T) {
	var serviceID string
	ctx := getContext()
	defer datasource.GetMetadataManager().UnregisterService(ctx, &pb.DeleteServiceRequest{ServiceId: serviceID, Force: true})

	t.Run("create service", func(t *testing.T) {
		respCreateService, err := datasource.GetMetadataManager().RegisterService(ctx, &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				ServiceName: "create_instance_service_ms",
				AppId:       "create_instance_ms",
				Version:     "1.0.0",
				Level:       "FRONT",
				Status:      pb.MS_UP,
			},
		})

		assert.NoError(t, err)
		serviceID = respCreateService.ServiceId
	})

	t.Run("register instance", func(t *testing.T) {
		respCreateInst, err := datasource.GetMetadataManager().RegisterInstance(ctx, &pb.RegisterInstanceRequest{
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
		assert.NotEqual(t, "", respCreateInst.InstanceId)

		respCreateInst, err = datasource.GetMetadataManager().RegisterInstance(ctx, &pb.RegisterInstanceRequest{
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
		assert.Equal(t, "customId_ms", respCreateInst.InstanceId)
	})

	t.Run("update the same instance", func(t *testing.T) {
		instance := &pb.MicroServiceInstance{
			ServiceId: serviceID,
			Endpoints: []string{
				"sameInstance:127.0.0.1:8080",
			},
			HostName: "UT-HOST",
			Status:   pb.MSI_UP,
		}
		resp, err := datasource.GetMetadataManager().RegisterInstance(ctx, &pb.RegisterInstanceRequest{
			Instance: instance,
		})
		assert.NoError(t, err)

		resp, err = datasource.GetMetadataManager().RegisterInstance(ctx, &pb.RegisterInstanceRequest{
			Instance: instance,
		})
		assert.NoError(t, err)
		assert.Equal(t, instance.InstanceId, resp.InstanceId)
	})
}

func TestInstance_HeartBeat(t *testing.T) {
	var (
		serviceID   string
		instanceID1 string
		instanceID2 string
	)
	ctx := getContext()
	defer datasource.GetMetadataManager().UnregisterService(ctx, &pb.DeleteServiceRequest{ServiceId: serviceID, Force: true})

	t.Run("register service and instance, should pass", func(t *testing.T) {
		log.Info("register service")
		respCreateService, err := datasource.GetMetadataManager().RegisterService(ctx, &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				ServiceName: "heartbeat_service_ms",
				AppId:       "heartbeat_service_ms",
				Version:     "1.0.0",
				Level:       "FRONT",
				Status:      pb.MS_UP,
			},
		})
		assert.NoError(t, err)
		serviceID = respCreateService.ServiceId

		respCreateInstance, err := datasource.GetMetadataManager().RegisterInstance(ctx, &pb.RegisterInstanceRequest{
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
		instanceID1 = respCreateInstance.InstanceId

		respCreateInstance, err = datasource.GetMetadataManager().RegisterInstance(ctx, &pb.RegisterInstanceRequest{
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
		instanceID2 = respCreateInstance.InstanceId
	})

	t.Run("update a lease", func(t *testing.T) {
		log.Info("valid instance")
		err := datasource.GetMetadataManager().SendHeartbeat(ctx, &pb.HeartbeatRequest{
			ServiceId:  serviceID,
			InstanceId: instanceID1,
		})
		assert.NoError(t, err)

		log.Info("serviceId does not exist")
		err = datasource.GetMetadataManager().SendHeartbeat(ctx, &pb.HeartbeatRequest{
			ServiceId:  "100000000000",
			InstanceId: instanceID1,
		})
		testErr := err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrInstanceNotExists, testErr.Code)

		log.Info("instance does not exist")
		err = datasource.GetMetadataManager().SendHeartbeat(ctx, &pb.HeartbeatRequest{
			ServiceId:  serviceID,
			InstanceId: "not-exist-ins",
		})
		testErr = err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrInstanceNotExists, testErr.Code)
	})

	t.Run("batch update lease", func(t *testing.T) {
		log.Info("request contains at least 1 instances")
		resp, err := datasource.GetMetadataManager().SendManyHeartbeat(ctx, &pb.HeartbeatSetRequest{
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
		assert.Equal(t, 2, len(resp.Instances))
	})
}

func TestInstance_Update(t *testing.T) {
	var (
		serviceID  string
		instanceID string
	)
	ctx := getContext()
	defer datasource.GetMetadataManager().UnregisterService(ctx, &pb.DeleteServiceRequest{ServiceId: serviceID, Force: true})

	t.Run("register service and instance, should pass", func(t *testing.T) {
		log.Info("register service")
		respCreateService, err := datasource.GetMetadataManager().RegisterService(ctx, &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				ServiceName: "update_instance_service_ms",
				AppId:       "update_instance_service_ms",
				Version:     "1.0.0",
				Level:       "FRONT",
				Status:      pb.MS_UP,
			},
		})
		assert.NoError(t, err)
		serviceID = respCreateService.ServiceId

		log.Info("create instance")
		respCreateInstance, err := datasource.GetMetadataManager().RegisterInstance(ctx, &pb.RegisterInstanceRequest{
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
		instanceID = respCreateInstance.InstanceId
	})

	t.Run("update instance status", func(t *testing.T) {
		log.Info("update instance status to DOWN")
		err := datasource.GetMetadataManager().PutInstanceStatus(ctx, &pb.UpdateInstanceStatusRequest{
			ServiceId:  serviceID,
			InstanceId: instanceID,
			Status:     pb.MSI_DOWN,
		})
		assert.NoError(t, err)

		log.Info("update instance status to OUTOFSERVICE")
		err = datasource.GetMetadataManager().PutInstanceStatus(ctx, &pb.UpdateInstanceStatusRequest{
			ServiceId:  serviceID,
			InstanceId: instanceID,
			Status:     pb.MSI_OUTOFSERVICE,
		})
		assert.NoError(t, err)

		log.Info("update instance status to STARTING")
		err = datasource.GetMetadataManager().PutInstanceStatus(ctx, &pb.UpdateInstanceStatusRequest{
			ServiceId:  serviceID,
			InstanceId: instanceID,
			Status:     pb.MSI_STARTING,
		})
		assert.NoError(t, err)

		log.Info("update instance status to TESTING")
		err = datasource.GetMetadataManager().PutInstanceStatus(ctx, &pb.UpdateInstanceStatusRequest{
			ServiceId:  serviceID,
			InstanceId: instanceID,
			Status:     pb.MSI_TESTING,
		})
		assert.NoError(t, err)

		log.Info("update instance status to UP")
		err = datasource.GetMetadataManager().PutInstanceStatus(ctx, &pb.UpdateInstanceStatusRequest{
			ServiceId:  serviceID,
			InstanceId: instanceID,
			Status:     pb.MSI_UP,
		})
		assert.NoError(t, err)

		log.Info("update instance status with a not exist instance")
		err = datasource.GetMetadataManager().PutInstanceStatus(ctx, &pb.UpdateInstanceStatusRequest{
			ServiceId:  serviceID,
			InstanceId: "notexistins",
			Status:     pb.MSI_STARTING,
		})
		testErr := err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrInstanceNotExists, testErr.Code)
	})

	t.Run("update instance properties", func(t *testing.T) {
		log.Info("update one properties")
		err := datasource.GetMetadataManager().PutInstanceProperties(ctx,
			&pb.UpdateInstancePropsRequest{
				ServiceId:  serviceID,
				InstanceId: instanceID,
				Properties: map[string]string{
					"test": "test",
				},
			})
		assert.NoError(t, err)

		log.Info("all max properties updated")
		size := 1000
		properties := make(map[string]string, size)
		for i := 0; i < size; i++ {
			s := strconv.Itoa(i) + strings.Repeat("x", 253)
			properties[s] = s
		}
		err = datasource.GetMetadataManager().PutInstanceProperties(ctx,
			&pb.UpdateInstancePropsRequest{
				ServiceId:  serviceID,
				InstanceId: instanceID,
				Properties: properties,
			})
		assert.NoError(t, err)

		log.Info("update instance that does not exist")
		err = datasource.GetMetadataManager().PutInstanceProperties(ctx,
			&pb.UpdateInstancePropsRequest{
				ServiceId:  serviceID,
				InstanceId: "not_exist_ins",
				Properties: map[string]string{
					"test": "test",
				},
			})
		testErr := err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrInstanceNotExists, testErr.Code)

		log.Info("remove properties")
		err = datasource.GetMetadataManager().PutInstanceProperties(ctx,
			&pb.UpdateInstancePropsRequest{
				ServiceId:  serviceID,
				InstanceId: instanceID,
			})
		assert.NoError(t, err)

		log.Info("update service that does not exist")
		err = datasource.GetMetadataManager().PutInstanceProperties(ctx,
			&pb.UpdateInstancePropsRequest{
				ServiceId:  "not_exist_service",
				InstanceId: instanceID,
				Properties: map[string]string{
					"test": "test",
				},
			})
		testErr = err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrInstanceNotExists, testErr.Code)
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
		instanceID1 string
		instanceID2 string
		instanceID4 string
		instanceID5 string
		instanceID8 string
	)
	ctx := getContext()

	defer func() {
		datasource.GetMetadataManager().UnregisterService(ctx, &pb.DeleteServiceRequest{ServiceId: serviceID1, Force: true})
		datasource.GetMetadataManager().UnregisterService(ctx, &pb.DeleteServiceRequest{ServiceId: serviceID2, Force: true})
		datasource.GetMetadataManager().UnregisterService(ctx, &pb.DeleteServiceRequest{ServiceId: serviceID3, Force: true})
		datasource.GetMetadataManager().UnregisterService(ctx, &pb.DeleteServiceRequest{ServiceId: serviceID4, Force: true})
		datasource.GetMetadataManager().UnregisterService(ctx, &pb.DeleteServiceRequest{ServiceId: serviceID5, Force: true})
		datasource.GetMetadataManager().UnregisterService(ctx, &pb.DeleteServiceRequest{ServiceId: serviceID6, Force: true})
		datasource.GetMetadataManager().UnregisterService(ctx, &pb.DeleteServiceRequest{ServiceId: serviceID7, Force: true})
		datasource.GetMetadataManager().UnregisterService(ctx, &pb.DeleteServiceRequest{ServiceId: serviceID8, Force: true})
	}()

	t.Run("register services and instances for testInstance_query", func(t *testing.T) {
		respCreateService, err := datasource.GetMetadataManager().RegisterService(ctx, &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				AppId:       "query_instance_ms",
				ServiceName: "query_instance_service_ms",
				Version:     "1.0.0",
				Level:       "FRONT",
				Status:      pb.MS_UP,
			},
		})
		assert.NoError(t, err)
		serviceID1 = respCreateService.ServiceId

		respCreateService, err = datasource.GetMetadataManager().RegisterService(ctx, &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				AppId:       "query_instance_ms",
				ServiceName: "query_instance_service_ms",
				Version:     "1.0.5",
				Level:       "FRONT",
				Status:      pb.MS_UP,
			},
		})
		assert.NoError(t, err)
		serviceID2 = respCreateService.ServiceId

		respCreateService, err = datasource.GetMetadataManager().RegisterService(ctx, &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				AppId:       "query_instance_diff_app_ms",
				ServiceName: "query_instance_service_ms",
				Version:     "1.0.0",
				Level:       "FRONT",
				Status:      pb.MS_UP,
			},
		})
		assert.NoError(t, err)
		serviceID3 = respCreateService.ServiceId

		respCreateService, err = datasource.GetMetadataManager().RegisterService(ctx, &pb.CreateServiceRequest{
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
		serviceID4 = respCreateService.ServiceId

		respCreateService, err = datasource.GetMetadataManager().RegisterService(ctx, &pb.CreateServiceRequest{
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
		serviceID5 = respCreateService.ServiceId

		respCreateService, err = datasource.GetMetadataManager().RegisterService(
			util.SetDomainProject(util.CloneContext(ctx), "user", "user"),
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
		serviceID6 = respCreateService.ServiceId

		respCreateService, err = datasource.GetMetadataManager().RegisterService(ctx, &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				AppId:       "default",
				ServiceName: "query_instance_shared_consumer_ms",
				Version:     "1.0.0",
				Level:       "FRONT",
				Status:      pb.MS_UP,
			},
		})
		assert.NoError(t, err)
		serviceID7 = respCreateService.ServiceId

		respCreateService, err = datasource.GetMetadataManager().RegisterService(ctx, &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				AppId:       "query_instance_ms",
				ServiceName: "query_instance_with_rev_ms",
				Version:     "1.0.0",
				Level:       "FRONT",
				Status:      pb.MS_UP,
			},
		})
		assert.NoError(t, err)
		serviceID8 = respCreateService.ServiceId

		respCreateInstance, err := datasource.GetMetadataManager().RegisterInstance(ctx, &pb.RegisterInstanceRequest{
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
		instanceID1 = respCreateInstance.InstanceId

		respCreateInstance, err = datasource.GetMetadataManager().RegisterInstance(ctx, &pb.RegisterInstanceRequest{
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
		instanceID2 = respCreateInstance.InstanceId

		respCreateInstance, err = datasource.GetMetadataManager().RegisterInstance(ctx, &pb.RegisterInstanceRequest{
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
		instanceID4 = respCreateInstance.InstanceId

		respCreateInstance, err = datasource.GetMetadataManager().RegisterInstance(ctx, &pb.RegisterInstanceRequest{
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
		instanceID5 = respCreateInstance.InstanceId

		respCreateInstance, err = datasource.GetMetadataManager().RegisterInstance(ctx, &pb.RegisterInstanceRequest{
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
		instanceID8 = respCreateInstance.InstanceId
	})

	t.Run("query instance, should ok", func(t *testing.T) {
		respFind, err := datasource.GetMetadataManager().FindInstances(ctx, &pb.FindInstancesRequest{
			ConsumerServiceId: serviceID1,
			AppId:             "query_instance_ms",
			ServiceName:       "query_instance_service_ms",
		})
		assert.NoError(t, err)
		assertInstancesContain(t, respFind.Instances, instanceID1)
		assertInstancesContain(t, respFind.Instances, instanceID2)
	})

	t.Run("query not exist service instance, should failed", func(t *testing.T) {
		_, err := datasource.GetMetadataManager().FindInstances(ctx, &pb.FindInstancesRequest{
			ConsumerServiceId: serviceID1,
			AppId:             "query_instance_ms",
			ServiceName:       "not-exist",
		})
		testErr := err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrServiceNotExists, testErr.Code)
	})

	t.Run("query instance when with consumerID or specify env without consumerID, should ok", func(t *testing.T) {
		respFind, err := datasource.GetMetadataManager().FindInstances(ctx, &pb.FindInstancesRequest{
			ConsumerServiceId: serviceID4,
			AppId:             "query_instance_ms",
			ServiceName:       "query_instance_diff_env_service_ms",
		})
		assert.NoError(t, err)
		assert.Equal(t, 1, len(respFind.Instances))
		assertInstancesContain(t, respFind.Instances, instanceID4)

		respFind, err = datasource.GetMetadataManager().FindInstances(ctx, &pb.FindInstancesRequest{
			Environment: pb.ENV_PROD,
			AppId:       "query_instance_ms",
			ServiceName: "query_instance_diff_env_service_ms",
		})
		assert.NoError(t, err)
		assert.Equal(t, 1, len(respFind.Instances))
		assertInstancesContain(t, respFind.Instances, instanceID4)
	})

	t.Run("query instance with revision, should ok", func(t *testing.T) {
		ctx := util.SetContext(ctx, util.CtxNocache, "")
		respFind, err := datasource.GetMetadataManager().FindInstances(ctx, &pb.FindInstancesRequest{
			ConsumerServiceId: serviceID8,
			AppId:             "query_instance_ms",
			ServiceName:       "query_instance_with_rev_ms",
		})
		assert.NoError(t, err)
		rev, _ := ctx.Value(util.CtxResponseRevision).(string)
		assert.NotEqual(t, 0, len(rev))
		assertInstancesContain(t, respFind.Instances, instanceID8)

		util.WithRequestRev(ctx, "x")
		respFind, err = datasource.GetMetadataManager().FindInstances(ctx, &pb.FindInstancesRequest{
			ConsumerServiceId: serviceID8,
			AppId:             "query_instance_ms",
			ServiceName:       "query_instance_with_rev_ms",
		})
		assert.NoError(t, err)
		assert.Equal(t, ctx.Value(util.CtxResponseRevision), rev)
		assertInstancesContain(t, respFind.Instances, instanceID8)
	})

	t.Run("find should return 200 if consumer is diff apps, should ok", func(t *testing.T) {
		respFind, err := datasource.GetMetadataManager().FindInstances(ctx, &pb.FindInstancesRequest{
			ConsumerServiceId: serviceID3,
			AppId:             "query_instance_ms",
			ServiceName:       "query_instance_service_ms",
		})
		assert.NoError(t, err)
		assert.Equal(t, 0, len(respFind.Instances))
	})

	t.Run("find provider instance but specify tag does not exist, should ok", func(t *testing.T) {
		respFind, err := datasource.GetMetadataManager().FindInstances(ctx, &pb.FindInstancesRequest{
			ConsumerServiceId: serviceID1,
			AppId:             "query_instance_ms",
			ServiceName:       "query_instance_service_ms",
			Tags:              []string{"not_exist_tag"},
		})
		assert.NoError(t, err)
		assert.Equal(t, 0, len(respFind.Instances))
	})

	t.Run("find global provider instance, should ok", func(t *testing.T) {
		config.Server.Config.GlobalVisible = "query_instance_shared_provider_ms"
		core.RegisterGlobalServices()
		core.Service.Environment = pb.ENV_PROD
		respFind, err := datasource.GetMetadataManager().FindInstances(
			util.SetTargetDomainProject(
				util.SetDomainProject(util.CloneContext(ctx), "user", "user"),
				"default", "default"),
			&pb.FindInstancesRequest{
				ConsumerServiceId: serviceID6,
				AppId:             "default",
				ServiceName:       "query_instance_shared_provider_ms",
			})
		assert.NoError(t, err)
		assert.Equal(t, 1, len(respFind.Instances))
		assertInstancesContain(t, respFind.Instances, instanceID5)

		respFind, err = datasource.GetMetadataManager().FindInstances(ctx, &pb.FindInstancesRequest{
			ConsumerServiceId: serviceID7,
			AppId:             "default",
			ServiceName:       "query_instance_shared_provider_ms",
		})
		assert.NoError(t, err)
		assert.Equal(t, 1, len(respFind.Instances))
		assertInstancesContain(t, respFind.Instances, instanceID5)

		log.Info("query same domain deps")
		// todo finish ut after implementing ListProviders interface

		core.Service.Environment = pb.ENV_DEV
	})

	t.Run("query instances between diff dimensions", func(t *testing.T) {
		log.Info("diff appId")
		UTFunc := func(consumerId string, code int32) {
			_, err := datasource.GetMetadataManager().ListInstance(ctx, &pb.GetInstancesRequest{
				ConsumerServiceId: consumerId,
				ProviderServiceId: serviceID2,
			})
			if code == pb.ResponseSuccess {
				assert.NoError(t, err)
				return
			}
			testErr := err.(*errsvc.Error)
			assert.Error(t, testErr)
			assert.Equal(t, code, testErr.Code)
		}

		UTFunc(serviceID3, pb.ErrServiceNotExists)

		UTFunc(serviceID1, pb.ResponseSuccess)

		log.Info("diff env")
		_, err := datasource.GetMetadataManager().ListInstance(ctx, &pb.GetInstancesRequest{
			ConsumerServiceId: serviceID4,
			ProviderServiceId: serviceID2,
		})
		testErr := err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrServiceNotExists, testErr.Code)
	})
}

func assertInstancesContain(t *testing.T, instances []*pb.MicroServiceInstance, instanceId string) {
	found := false
	for _, instance := range instances {
		if instance.InstanceId == instanceId {
			found = true
			break
		}
	}
	assert.True(t, found)
}

func TestServicesStatistics_Get(t *testing.T) {
	var ctx context.Context
	var serviceId1 string
	var serviceId2 string

	t.Run("register services and instances for TestServicesStatistics_Get", func(t *testing.T) {
		//service1
		ctx = util.WithNoCache(util.SetDomainProject(context.Background(), "default", "Project1"))
		respCreateService, err := datasource.GetMetadataManager().RegisterService(ctx, &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				AppId:       "query_statistics1",
				ServiceName: "service1",
				Version:     "1.0.0",
				Level:       "FRONT",
				Status:      pb.MS_UP,
			},
		})
		assert.NoError(t, err)
		serviceId1 = respCreateService.ServiceId

		_, err = datasource.GetMetadataManager().RegisterInstance(ctx, &pb.RegisterInstanceRequest{
			Instance: &pb.MicroServiceInstance{
				ServiceId: serviceId1,
				HostName:  "UT-HOST-MS1",
				Endpoints: []string{
					"find:127.0.0.1:8080",
				},
				Status: pb.MSI_UP,
			},
		})
		assert.NoError(t, err)

		_, err = datasource.GetMetadataManager().RegisterInstance(ctx, &pb.RegisterInstanceRequest{
			Instance: &pb.MicroServiceInstance{
				ServiceId: serviceId1,
				HostName:  "UT-HOST-MS2",
				Endpoints: []string{
					"find:127.0.0.1:8080",
				},
				Status: pb.MSI_UP,
			},
		})
		assert.NoError(t, err)

		//service2
		ctx = util.WithNoCache(util.SetDomainProject(context.Background(), "DomainTest1", "Project1"))
		respCreateService, err = datasource.GetMetadataManager().RegisterService(ctx, &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				AppId:       "query_statistics2",
				ServiceName: "service2-DomainTest1",
				Version:     "1.0.5",
				Level:       "FRONT",
				Status:      pb.MS_UP,
			},
		})
		assert.NoError(t, err)
		serviceId2 = respCreateService.ServiceId

		_, err = datasource.GetMetadataManager().RegisterInstance(ctx, &pb.RegisterInstanceRequest{
			Instance: &pb.MicroServiceInstance{
				ServiceId: serviceId2,
				HostName:  "UT-HOST-MS1",
				Endpoints: []string{
					"find:127.0.0.1:8080",
				},
				Status: pb.MSI_UP,
			},
		})
		assert.NoError(t, err)
	})

	t.Run("query services statistics", func(t *testing.T) {
		ctx = util.WithNoCache(util.SetDomainProject(context.Background(), "default", "Project1"))
		log.Info("query services default domain statistics")
		st, err := datasource.GetMetadataManager().GetOverview(ctx, &pb.GetServicesRequest{})
		assert.NoError(t, err)
		assert.Equal(t, int64(2), st.Instances.CountByDomain)

		log.Info("query services domain statistics")
		ctx = util.WithNoCache(util.SetDomainProject(context.Background(), "DomainTest1", "Project1"))
		st, err = datasource.GetMetadataManager().GetOverview(ctx, &pb.GetServicesRequest{})
		assert.NoError(t, err)
		assert.Equal(t, int64(1), st.Instances.CountByDomain)
	})

	t.Run("delete a service ", func(t *testing.T) {
		ctx = util.WithNoCache(util.SetDomainProject(context.Background(), "default", "Project1"))
		err := datasource.GetMetadataManager().UnregisterService(ctx, &pb.DeleteServiceRequest{
			ServiceId: serviceId1,
			Force:     true,
		})
		assert.NoError(t, err)

		ctx = util.WithNoCache(util.SetDomainProject(context.Background(), "DomainTest1", "Project1"))
		err = datasource.GetMetadataManager().UnregisterService(ctx, &pb.DeleteServiceRequest{
			ServiceId: serviceId2,
			Force:     true,
		})
		assert.NoError(t, err)
	})
}

func TestInstance_GetOne(t *testing.T) {
	var (
		serviceId1  string
		serviceId2  string
		serviceId3  string
		instanceId2 string
	)
	ctx := getContext()
	defer datasource.GetMetadataManager().UnregisterService(ctx, &pb.DeleteServiceRequest{ServiceId: serviceId1, Force: true})
	defer datasource.GetMetadataManager().UnregisterService(ctx, &pb.DeleteServiceRequest{ServiceId: serviceId2, Force: true})
	defer datasource.GetMetadataManager().UnregisterService(ctx, &pb.DeleteServiceRequest{ServiceId: serviceId3, Force: true})

	t.Run("register service and instances", func(t *testing.T) {
		respCreateService, err := datasource.GetMetadataManager().RegisterService(ctx, &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				AppId:       "get_instance_ms",
				ServiceName: "get_instance_service_ms",
				Version:     "1.0.0",
				Level:       "FRONT",
				Status:      pb.MS_UP,
			},
		})
		assert.NoError(t, err)
		serviceId1 = respCreateService.ServiceId

		respCreateService, err = datasource.GetMetadataManager().RegisterService(ctx, &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				AppId:       "get_instance_ms",
				ServiceName: "get_instance_service_ms",
				Version:     "1.0.5",
				Level:       "FRONT",
				Status:      pb.MS_UP,
			},
		})
		assert.NoError(t, err)
		serviceId2 = respCreateService.ServiceId

		respCreateInstance, err := datasource.GetMetadataManager().RegisterInstance(ctx, &pb.RegisterInstanceRequest{
			Instance: &pb.MicroServiceInstance{
				ServiceId: serviceId2,
				HostName:  "UT-HOST-MS",
				Endpoints: []string{
					"get:127.0.0.2:8080",
				},
				Status: pb.MSI_UP,
			},
		})
		assert.NoError(t, err)
		instanceId2 = respCreateInstance.InstanceId

		respCreateService, err = datasource.GetMetadataManager().RegisterService(ctx, &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				AppId:       "get_instance_cross_ms",
				ServiceName: "get_instance_service_ms",
				Version:     "1.0.0",
				Level:       "FRONT",
				Status:      pb.MS_UP,
			},
		})
		assert.NoError(t, err)
		serviceId3 = respCreateService.ServiceId
	})

	t.Run("get one instance when invalid request", func(t *testing.T) {
		log.Info("find service itself")
		resp, err := datasource.GetMetadataManager().GetInstance(ctx, &pb.GetOneInstanceRequest{
			ConsumerServiceId:  serviceId2,
			ProviderServiceId:  serviceId2,
			ProviderInstanceId: instanceId2,
		})
		assert.NoError(t, err)
		assert.NotNil(t, resp.Instance)

		log.Info("consumer does not exist")
		_, err = datasource.GetMetadataManager().GetInstance(ctx, &pb.GetOneInstanceRequest{
			ConsumerServiceId:  "not-exist-id-ms",
			ProviderServiceId:  serviceId2,
			ProviderInstanceId: instanceId2,
		})
		testErr := err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrServiceNotExists, testErr.Code)
	})

	t.Run("get between diff apps", func(t *testing.T) {
		_, err := datasource.GetMetadataManager().GetInstance(ctx, &pb.GetOneInstanceRequest{
			ConsumerServiceId:  serviceId3,
			ProviderServiceId:  serviceId2,
			ProviderInstanceId: instanceId2,
		})
		testErr := err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrInstanceNotExists, testErr.Code)

		_, err = datasource.GetMetadataManager().ListInstance(ctx, &pb.GetInstancesRequest{
			ConsumerServiceId: serviceId3,
			ProviderServiceId: serviceId2,
		})
		testErr = err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrServiceNotExists, testErr.Code)
	})

	t.Run("get instances when request is invalid", func(t *testing.T) {
		log.Info("consumer does not exist")
		resp, err := datasource.GetMetadataManager().ListInstance(ctx, &pb.GetInstancesRequest{
			ConsumerServiceId: "not-exist-service-ms",
			ProviderServiceId: serviceId2,
		})
		testErr := err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrServiceNotExists, testErr.Code)

		log.Info("consumer does not exist")
		resp, err = datasource.GetMetadataManager().ListInstance(ctx, &pb.GetInstancesRequest{
			ConsumerServiceId: serviceId1,
			ProviderServiceId: serviceId2,
		})
		assert.NoError(t, err)
		assert.NotNil(t, resp)
	})
}

func TestInstance_GetAll(t *testing.T) {
	t.Run("register 2 instances, get all instances count should return 2", func(t *testing.T) {
		var (
			serviceId1 string
			serviceId2 string
		)
		ctx := util.WithNoCache(util.SetDomainProject(context.Background(), "TestInstance_GetAll", "1"))

		respCreateService, err := datasource.GetMetadataManager().RegisterService(ctx, &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				AppId:       "get_instance_ms",
				ServiceName: "get_instance_service_ms",
				Version:     "1.0.0",
				Level:       "FRONT",
				Status:      pb.MS_UP,
			},
		})
		assert.NoError(t, err)
		serviceId1 = respCreateService.ServiceId
		defer datasource.GetMetadataManager().UnregisterService(ctx, &pb.DeleteServiceRequest{ServiceId: serviceId1, Force: true})

		respCreateService, err = datasource.GetMetadataManager().RegisterService(ctx, &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				AppId:       "get_instance_ms",
				ServiceName: "get_instance_service_ms",
				Version:     "1.0.5",
				Level:       "FRONT",
				Status:      pb.MS_UP,
			},
		})
		assert.NoError(t, err)
		serviceId2 = respCreateService.ServiceId
		defer datasource.GetMetadataManager().UnregisterService(ctx, &pb.DeleteServiceRequest{ServiceId: serviceId2, Force: true})

		_, err = datasource.GetMetadataManager().RegisterInstance(ctx, &pb.RegisterInstanceRequest{
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

		_, err = datasource.GetMetadataManager().RegisterInstance(ctx, &pb.RegisterInstanceRequest{
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

		respAll, err := datasource.GetMetadataManager().ListManyInstances(ctx, &pb.GetAllInstancesRequest{})
		assert.NoError(t, err)
		assert.Equal(t, 2, len(respAll.Instances))
	})

	t.Run("domain contain no instances, get all instances should be pass, return 0 instance", func(t *testing.T) {
		ctx := util.WithNoCache(util.SetDomainProject(context.Background(), "TestInstance_GetAll", "2"))
		respAll, err := datasource.GetMetadataManager().ListManyInstances(ctx, &pb.GetAllInstancesRequest{})
		assert.NoError(t, err)
		assert.Equal(t, 0, len(respAll.Instances))
	})
}

func TestInstance_Unregister(t *testing.T) {
	var (
		serviceId  string
		instanceId string
	)
	ctx := getContext()
	defer datasource.GetMetadataManager().UnregisterService(ctx, &pb.DeleteServiceRequest{ServiceId: serviceId, Force: true})

	t.Run("register service and instances", func(t *testing.T) {
		respCreateService, err := datasource.GetMetadataManager().RegisterService(ctx, &pb.CreateServiceRequest{
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
		serviceId = respCreateService.ServiceId

		respCreateInstance, err := datasource.GetMetadataManager().RegisterInstance(ctx, &pb.RegisterInstanceRequest{
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
		instanceId = respCreateInstance.InstanceId
	})

	t.Run("unregister instance", func(t *testing.T) {
		err := datasource.GetMetadataManager().UnregisterInstance(ctx, &pb.UnregisterInstanceRequest{
			ServiceId:  serviceId,
			InstanceId: instanceId,
		})
		assert.NoError(t, err)
	})

	t.Run("unregister instance when request is invalid", func(t *testing.T) {
		log.Info("service id does not exist")
		err := datasource.GetMetadataManager().UnregisterInstance(ctx, &pb.UnregisterInstanceRequest{
			ServiceId:  "not-exist-id-ms",
			InstanceId: instanceId,
		})
		testErr := err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrInstanceNotExists, testErr.Code)

		log.Info("instance id does not exist")
		err = datasource.GetMetadataManager().UnregisterInstance(ctx, &pb.UnregisterInstanceRequest{
			ServiceId:  serviceId,
			InstanceId: "not-exist-id-ms",
		})
		testErr = err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrInstanceNotExists, testErr.Code)
	})
}

func TestInstance_Exist(t *testing.T) {
	var (
		serviceId  string
		instanceId string
	)
	ctx := getContext()
	defer datasource.GetMetadataManager().UnregisterService(ctx, &pb.DeleteServiceRequest{ServiceId: serviceId, Force: true})

	t.Run("register service and instance", func(t *testing.T) {
		respCreateService, err := datasource.GetMetadataManager().RegisterService(ctx, &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				AppId:       "check_exist_instance_ms",
				ServiceName: "check_exist_instance_service_ms",
				Version:     "1.0.5",
				Level:       "FRONT",
				Status:      pb.MS_UP,
			},
			Tags: map[string]string{
				"test": "test",
			},
		})
		assert.NoError(t, err)
		serviceId = respCreateService.ServiceId

		respCreateInstance, err := datasource.GetMetadataManager().RegisterInstance(ctx, &pb.RegisterInstanceRequest{
			Instance: &pb.MicroServiceInstance{
				ServiceId: serviceId,
				HostName:  "UT-HOST-MS",
				Endpoints: []string{
					"checkExist:127.0.0.2:8080",
				},
				Status: pb.MSI_UP,
			},
		})
		assert.NoError(t, err)
		instanceId = respCreateInstance.InstanceId
	})

	t.Run("the check instance should exist", func(t *testing.T) {
		respCheckInstance, err := datasource.GetMetadataManager().ExistInstance(ctx, &pb.MicroServiceInstanceKey{
			ServiceId:  serviceId,
			InstanceId: instanceId,
		})
		assert.NoError(t, err)
		assert.True(t, respCheckInstance.Exist)
	})

	t.Run("unregister instance", func(t *testing.T) {
		err := datasource.GetMetadataManager().UnregisterInstance(ctx, &pb.UnregisterInstanceRequest{
			ServiceId:  serviceId,
			InstanceId: instanceId,
		})
		assert.NoError(t, err)
	})

	t.Run("The check instance should not exist", func(t *testing.T) {
		resp, err := datasource.GetMetadataManager().ExistInstance(ctx, &pb.MicroServiceInstanceKey{
			ServiceId:  serviceId,
			InstanceId: instanceId,
		})
		assert.NoError(t, err)
		assert.False(t, resp.Exist)
	})

	t.Run("unregister service", func(t *testing.T) {
		err := datasource.GetMetadataManager().UnregisterService(ctx, &pb.DeleteServiceRequest{
			ServiceId: serviceId,
		})
		assert.NoError(t, err)
	})
}
