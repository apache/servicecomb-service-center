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
	"strconv"
	"strings"
	"testing"

	"github.com/apache/servicecomb-service-center/pkg/util"
	"github.com/apache/servicecomb-service-center/server/core"
	"github.com/apache/servicecomb-service-center/server/service/disco"
	quotasvc "github.com/apache/servicecomb-service-center/server/service/quota"
	pb "github.com/go-chassis/cari/discovery"
	"github.com/go-chassis/cari/pkg/errsvc"
	"github.com/stretchr/testify/assert"
)

var (
	TooLongAppID       = strings.Repeat("x", 161)
	TooLongSchemaID    = strings.Repeat("x", 161)
	TooLongServiceID   = strings.Repeat("x", 65)
	TooLongServiceName = strings.Repeat("x", 129)
	TooLongExistence   = strings.Repeat("x", 128+160+2)
	TooLongAlias       = strings.Repeat("x", 129)
	TooLongFramework   = strings.Repeat("x", 65)
	TooLongDescription = strings.Repeat("x", 257)
)

func TestRegisterService(t *testing.T) {
	ctx := getContext()

	t.Run("when service is nil, should be failed", func(t *testing.T) {
		_, err := disco.RegisterService(ctx, &pb.CreateServiceRequest{
			Service: nil,
		})
		testErr := err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrInvalidParams, testErr.Code)
	})

	t.Run("all max, should be passed", func(t *testing.T) {
		size := int(quotasvc.SchemaQuota()) + 1
		paths := make([]*pb.ServicePath, 0, size)
		properties := make(map[string]string, size)
		for i := 0; i < size; i++ {
			s := strconv.Itoa(i) + strings.Repeat("x", 253)
			paths = append(paths, &pb.ServicePath{Path: s, Property: map[string]string{s: s}})
			properties[s] = s
		}
		r := &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				AppId:       TooLongAppID[:len(TooLongAppID)-1],
				ServiceName: TooLongServiceName[:len(TooLongServiceName)-1],
				Version:     "32767.32767.32767.32767",
				Alias:       TooLongAlias[:len(TooLongAlias)-1],
				Level:       "BACK",
				Status:      "UP",
				Schemas:     []string{TooLongSchemaID[:len(TooLongSchemaID)-1]},
				Paths:       paths,
				Properties:  properties,
				Framework: &pb.FrameWork{
					Name:    TooLongFramework[:len(TooLongFramework)-1],
					Version: TooLongFramework[:len(TooLongFramework)-1],
				},
				RegisterBy: "SDK",
			},
		}
		_, err := disco.RegisterService(ctx, r)
		assert.NoError(t, err)
	})

	t.Run("when service with rules/tags/instances, should be passed", func(t *testing.T) {
		tags := make(map[string]string, 10)
		tags["test"] = "test"
		rules := []*pb.AddOrUpdateServiceRule{
			{
				RuleType:    "BLACK",
				Attribute:   "ServiceName",
				Pattern:     "test",
				Description: "test",
			},
		}
		instances := []*pb.MicroServiceInstance{
			{
				Endpoints: []string{
					"createService:127.0.0.1:8080",
				},
				HostName: "UT-HOST",
				Status:   pb.MSI_UP,
			},
		}

		resp, err := disco.RegisterService(ctx, &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				ServiceName: "create_service_rule_tag",
				AppId:       "default",
				Version:     "1.0.0",
				Level:       "FRONT",
				Schemas: []string{
					"xxxxxxxx",
				},
				Status: "UP",
			},
			Tags:      tags,
			Rules:     rules,
			Instances: instances,
		})
		assert.NoError(t, err)

		delete(tags, "test")
		tags["second"] = "second"
		resp, err = disco.RegisterService(ctx, &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				ServiceName: "create_service_rule_tag",
				AppId:       "default",
				Version:     "1.0.0",
				Level:       "FRONT",
				Schemas: []string{
					"xxxxxxxx",
				},
				Status: "UP",
			},
			Tags: tags,
		})
		assert.NoError(t, err)

		respGetTags, err := disco.ListTag(ctx, &pb.GetServiceTagsRequest{
			ServiceId: resp.ServiceId,
		})
		assert.NoError(t, err)
		assert.Equal(t, "second", respGetTags.Tags["second"])

		respGetInsts, err := disco.ListInstance(ctx, &pb.GetInstancesRequest{
			ConsumerServiceId: resp.ServiceId,
			ProviderServiceId: resp.ServiceId,
		})
		assert.NoError(t, err)
		assert.Equal(t, "UT-HOST", respGetInsts.Instances[0].HostName)

		err = disco.UnregisterService(ctx, &pb.DeleteServiceRequest{
			ServiceId: resp.ServiceId,
			Force:     true,
		})
		assert.NoError(t, err)
	})

	t.Run("when creating the same service, should be failed", func(t *testing.T) {
		resp, err := disco.RegisterService(ctx, &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				ServiceName: "some-relay",
				Alias:       "sr",
				AppId:       "default",
				Version:     "1.0.0",
				Level:       "FRONT",
				Schemas: []string{
					"xxxxxxxx",
				},
				Status: "UP",
			},
		})
		assert.NoError(t, err)
		sameId := resp.ServiceId

		resp, err = disco.RegisterService(ctx, &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				ServiceName: "some-relay",
				Alias:       "sr1",
				AppId:       "default",
				Version:     "1.0.0",
				Level:       "FRONT",
				Schemas: []string{
					"xxxxxxxx",
				},
				Status: "UP",
			},
		})
		assert.NoError(t, err)
		assert.Equal(t, sameId, resp.ServiceId)

		resp, err = disco.RegisterService(ctx, &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				ServiceName: "some-relay1",
				Alias:       "sr",
				AppId:       "default",
				Version:     "1.0.0",
				Level:       "FRONT",
				Schemas: []string{
					"xxxxxxxx",
				},
				Status: "UP",
			},
		})
		assert.NoError(t, err)
		assert.Equal(t, sameId, resp.ServiceId)

		resp, err = disco.RegisterService(ctx, &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				ServiceId:   sameId,
				ServiceName: "some-relay",
				Alias:       "sr1",
				AppId:       "default",
				Version:     "1.0.0",
				Level:       "FRONT",
				Schemas: []string{
					"xxxxxxxx",
				},
				Status: "UP",
			},
		})
		assert.NoError(t, err)
		assert.Equal(t, sameId, resp.ServiceId)

		resp, err = disco.RegisterService(ctx, &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				ServiceId:   sameId,
				ServiceName: "some-relay1",
				Alias:       "sr",
				AppId:       "default",
				Version:     "1.0.0",
				Level:       "FRONT",
				Schemas: []string{
					"xxxxxxxx",
				},
				Status: "UP",
			},
		})
		assert.NoError(t, err)
		assert.Equal(t, sameId, resp.ServiceId)

		resp, err = disco.RegisterService(ctx, &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				ServiceId:   "customId",
				ServiceName: "some-relay",
				Alias:       "sr1",
				AppId:       "default",
				Version:     "1.0.0",
				Level:       "FRONT",
				Schemas: []string{
					"xxxxxxxx",
				},
				Status: "UP",
			},
		})
		testErr := err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrServiceAlreadyExists, testErr.Code)

		resp, err = disco.RegisterService(ctx, &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				ServiceId:   "customId",
				ServiceName: "some-relay1",
				Alias:       "sr",
				AppId:       "default",
				Version:     "1.0.0",
				Level:       "FRONT",
				Schemas: []string{
					"xxxxxxxx",
				},
				Status: "UP",
			},
		})
		testErr = err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrServiceAlreadyExists, testErr.Code)

		resp, err = disco.RegisterService(ctx, &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				ServiceId:   sameId,
				ServiceName: "some-relay2",
				Alias:       "sr2",
				AppId:       "default",
				Version:     "1.0.0",
				Level:       "FRONT",
				Schemas: []string{
					"xxxxxxxx",
				},
				Status: "UP",
			},
		})
		testErr = err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrServiceAlreadyExists, testErr.Code)
	})

	t.Run("same serviceId,different service, can not register again,error is same as the service register twice", func(t *testing.T) {
		resp, err := disco.RegisterService(ctx, &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				ServiceId:   "same_serviceId",
				ServiceName: "serviceA",
				AppId:       "default",
				Version:     "1.0.0",
				Level:       "FRONT",
				Schemas: []string{
					"xxxxxxxx",
				},
				Status: "UP",
			},
		})
		assert.NoError(t, err)
		defer disco.UnregisterService(ctx, &pb.DeleteServiceRequest{ServiceId: resp.ServiceId, Force: true})

		_, err = disco.RegisterService(ctx, &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				ServiceId:   "same_serviceId",
				ServiceName: "serviceB",
				AppId:       "default",
				Version:     "1.0.0",
				Level:       "FRONT",
				Schemas: []string{
					"xxxxxxxx",
				},
				Status: "UP",
			},
		})
		testErr := err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrServiceAlreadyExists, testErr.Code)
	})

	t.Run("when creating a diff env service, should be failed", func(t *testing.T) {
		service := &pb.MicroService{
			ServiceName: "diff_env_service",
			AppId:       "default",
			Version:     "1.0.0",
			Level:       "FRONT",
			Schemas: []string{
				"xxxxxxxx",
			},
			Status: "UP",
		}
		_, err := disco.RegisterService(ctx, &pb.CreateServiceRequest{
			Service: service,
		})
		assert.NoError(t, err)

		service.ServiceId = ""
		service.Environment = pb.ENV_PROD
		_, err = disco.RegisterService(ctx, &pb.CreateServiceRequest{
			Service: service,
		})
		assert.NoError(t, err)
	})

	t.Run("when service with properties, should be passed", func(t *testing.T) {
		r := &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				ServiceName: "some-backend",
				AppId:       "default",
				Version:     "1.0.0",
				Level:       "BACK",
				Schemas: []string{
					"xxxxxxxx",
				},
				Properties: make(map[string]string),
				Status:     "UP",
			},
		}
		r.Service.Properties["project"] = "x"
		resp, err := disco.RegisterService(ctx, r)
		assert.NoError(t, err)

		defer disco.UnregisterService(ctx, &pb.DeleteServiceRequest{ServiceId: resp.ServiceId, Force: true})
	})

	t.Run("when service body is valid, should be passed", func(t *testing.T) {
		r := &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				AppId:       "default",
				ServiceName: "service-validate-alias",
				Alias:       "aA:-_.1",
				Version:     "1.0.0",
				Level:       "BACK",
				Status:      "UP",
			},
		}

		resp, err := disco.RegisterService(ctx, r)
		assert.NoError(t, err)
		disco.UnregisterService(ctx, &pb.DeleteServiceRequest{ServiceId: resp.ServiceId, Force: true})

		r = &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				AppId:       "default",
				ServiceName: "registerBy-test",
				Version:     "1.0.10",
				Level:       "BACK",
				Status:      "UP",
				RegisterBy:  "PLATFORM",
			},
		}
		resp, err = disco.RegisterService(ctx, r)
		assert.NoError(t, err)
		disco.UnregisterService(ctx, &pb.DeleteServiceRequest{ServiceId: resp.ServiceId, Force: true})

		r = &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				AppId:       "default",
				ServiceName: "registerBy-test",
				Version:     "1.0.11",
				Level:       "BACK",
				Status:      "UP",
				RegisterBy:  "SIDECAR",
			},
		}
		resp, err = disco.RegisterService(ctx, r)
		assert.NoError(t, err)
		disco.UnregisterService(ctx, &pb.DeleteServiceRequest{ServiceId: resp.ServiceId, Force: true})
	})

	t.Run("when service body is invalid, should be failed", func(t *testing.T) {
		r := &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				ServiceId:   " ",
				AppId:       "default",
				ServiceName: "service-validate",
				Version:     "1.0.0",
				Level:       "BACK",
				Status:      "UP",
			},
		}
		_, err := disco.RegisterService(ctx, r)
		testErr := err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrInvalidParams, testErr.Code)

		r = &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				AppId:       TooLongAppID,
				ServiceName: "service-validate",
				Version:     "1.0.0",
				Level:       "BACK",
				Status:      "UP",
			},
		}
		_, err = disco.RegisterService(ctx, r)
		testErr = err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrInvalidParams, testErr.Code)

		r = &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				AppId:   "default",
				Version: "1.0.0",
				Level:   "BACK",
				Status:  "UP",
			},
		}
		_, err = disco.RegisterService(ctx, r)
		testErr = err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrInvalidParams, testErr.Code)

		r = &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				AppId:       "default",
				ServiceName: "service-validate",
				Version:     "1.",
				Level:       "BACK",
				Status:      "UP",
			},
		}
		_, err = disco.RegisterService(ctx, r)
		testErr = err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrInvalidParams, testErr.Code)

		r = &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				AppId:       "default",
				ServiceName: "service-validate",
				Version:     "1.a.0",
				Level:       "BACK",
				Status:      "UP",
			},
		}
		_, err = disco.RegisterService(ctx, r)
		testErr = err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrInvalidParams, testErr.Code)

		r = &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				AppId:       "default",
				ServiceName: "service-validate",
				Version:     "1.32768.0",
				Level:       "BACK",
				Status:      "UP",
			},
		}
		_, err = disco.RegisterService(ctx, r)
		testErr = err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrInvalidParams, testErr.Code)

		r = &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				AppId:       "default",
				ServiceName: "service-validate",
				Version:     "1.1.0.0.0",
				Level:       "BACK",
				Status:      "UP",
			},
		}
		_, err = disco.RegisterService(ctx, r)
		testErr = err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrInvalidParams, testErr.Code)

		r = &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				AppId:       "default",
				ServiceName: "service-validate",
				Version:     "1.0.0",
				Level:       "INVALID",
				Status:      "UP",
			},
		}
		_, err = disco.RegisterService(ctx, r)
		testErr = err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrInvalidParams, testErr.Code)

		r = &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				Environment: "notexistenv",
				AppId:       "default",
				ServiceName: "service-invalidate-env",
				Version:     "1.0.0",
				Level:       "BACK",
				Status:      "UP",
			},
		}
		_, err = disco.RegisterService(ctx, r)
		testErr = err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrInvalidParams, testErr.Code)

		r = &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				AppId:       "default",
				ServiceName: "service-validate-alias",
				Alias:       "@$#%",
				Version:     "1.0.0",
				Level:       "BACK",
				Status:      "UP",
			},
		}
		_, err = disco.RegisterService(ctx, r)
		testErr = err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrInvalidParams, testErr.Code)

		r = &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				AppId:       "default",
				ServiceName: "service-validate-alias",
				Alias:       TooLongAlias,
				Version:     "1.0.0",
				Level:       "BACK",
				Status:      "UP",
			},
		}
		_, err = disco.RegisterService(ctx, r)
		testErr = err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrInvalidParams, testErr.Code)

		r = &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				ServiceName: "framework-test",
				AppId:       "default",
				Version:     "1.0.4",
				Level:       "BACK",
				Framework: &pb.FrameWork{
					Version: TooLongFramework,
				},
				Properties: make(map[string]string),
				Status:     "UP",
			},
		}
		_, err = disco.RegisterService(ctx, r)
		testErr = err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrInvalidParams, testErr.Code)

		r = &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				ServiceName: "framework-test",
				AppId:       "default",
				Version:     "1.0.5",
				Level:       "BACK",
				Framework: &pb.FrameWork{
					Name: TooLongFramework,
				},
				Properties: make(map[string]string),
				Status:     "UP",
			},
		}
		_, err = disco.RegisterService(ctx, r)
		testErr = err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrInvalidParams, testErr.Code)

		r = &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				ServiceName: "framework-test",
				AppId:       "default",
				Version:     "1.0.5",
				Level:       "BACK",
				Framework: &pb.FrameWork{
					Name: "test@$",
				},
				Properties: make(map[string]string),
				Status:     "UP",
			},
		}
		_, err = disco.RegisterService(ctx, r)
		testErr = err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrInvalidParams, testErr.Code)

		r = &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				AppId:       "default",
				ServiceName: "framework-test",
				Version:     "1.0.6",
				Level:       "BACK",
				Status:      "UP",
				RegisterBy:  "InValid",
			},
		}
		_, err = disco.RegisterService(ctx, r)
		testErr = err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrInvalidParams, testErr.Code)

		r = &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				AppId:       "default",
				ServiceName: "description-test",
				Version:     "1.0.6",
				Level:       "BACK",
				Status:      "UP",
				Description: TooLongDescription,
			},
		}
		_, err = disco.RegisterService(ctx, r)
		testErr = err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrInvalidParams, testErr.Code)

		r = &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				AppId:       "default",
				ServiceName: "schema-test",
				Level:       "BACK",
				Status:      "UP",
				Schemas:     []string{TooLongSchemaID},
			},
		}
		_, err = disco.RegisterService(ctx, r)
		testErr = err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrInvalidParams, testErr.Code)
	})

	t.Run("when service with framework/registerBy/status, should be passed", func(t *testing.T) {
		r := &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				ServiceName: "framework-test",
				AppId:       "default",
				Version:     "1.0.1",
				Level:       "BACK",
				Framework: &pb.FrameWork{
					Version: "1.0.0",
				},
				Properties: make(map[string]string),
				Status:     "UP",
			},
		}
		resp, err := disco.RegisterService(ctx, r)
		assert.NoError(t, err)
		defer disco.UnregisterService(ctx, &pb.DeleteServiceRequest{ServiceId: resp.ServiceId, Force: true})

		r = &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				ServiceName: "framework-test",
				AppId:       "default",
				Version:     "1.0.2",
				Level:       "BACK",
				Framework: &pb.FrameWork{
					Name: "framework",
				},
				Properties: make(map[string]string),
				Status:     "UP",
			},
		}
		resp, err = disco.RegisterService(ctx, r)
		assert.NoError(t, err)
		defer disco.UnregisterService(ctx, &pb.DeleteServiceRequest{ServiceId: resp.ServiceId, Force: true})

		r = &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				ServiceName: "status-test",
				AppId:       "default",
				Version:     "1.0.3",
				Level:       "BACK",
				Properties:  make(map[string]string),
			},
		}
		resp, err = disco.RegisterService(ctx, r)
		assert.NoError(t, err)
		defer disco.UnregisterService(ctx, &pb.DeleteServiceRequest{ServiceId: resp.ServiceId, Force: true})
	})
}

func TestUnregisterService(t *testing.T) {
	var (
		serviceContainInstId string
		serviceNoInstId      string
		serviceConsumerId    string
	)
	ctx := getContext()

	t.Run("prepare data, should be passed", func(t *testing.T) {
		instances := []*pb.MicroServiceInstance{
			{
				Endpoints: []string{
					"deleteService:127.0.0.1:8080",
				},
				HostName: "delete-host",
				Status:   pb.MSI_UP,
			},
		}

		respCreate, err := disco.RegisterService(ctx, &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				ServiceName: "delete_service_with_inst",
				AppId:       "delete_service",
				Version:     "1.0.0",
				Level:       "FRONT",
				Status:      "UP",
			},
			Instances: instances,
		})
		assert.NoError(t, err)
		serviceContainInstId = respCreate.ServiceId

		provider := &pb.MicroService{
			ServiceName: "delete_serivce_no_inst",
			AppId:       "delete_service",
			Version:     "1.0.0",
			Level:       "FRONT",
			Status:      "UP",
		}
		respCreate, err = disco.RegisterService(ctx, &pb.CreateServiceRequest{
			Service: provider,
		})
		assert.NoError(t, err)
		serviceNoInstId = respCreate.ServiceId

		respCreate, err = disco.RegisterService(ctx, &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				ServiceName: "delete_serivce_consuemr",
				AppId:       "delete_service",
				Version:     "1.0.0",
				Level:       "FRONT",
				Status:      "UP",
			},
		})
		assert.NoError(t, err)
		serviceConsumerId = respCreate.ServiceId

		_, err = disco.FindInstances(ctx, &pb.FindInstancesRequest{
			ConsumerServiceId: serviceConsumerId,
			AppId:             provider.AppId,
			ServiceName:       provider.ServiceName,
		})
		assert.NoError(t, err)

		DependencyHandle()
	})

	t.Run("when request is invalid, should be failed", func(t *testing.T) {
		err := disco.UnregisterService(ctx, &pb.DeleteServiceRequest{
			ServiceId: "",
			Force:     true,
		})
		testErr := err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrInvalidParams, testErr.Code)

		err = disco.UnregisterService(ctx, &pb.DeleteServiceRequest{
			ServiceId: TooLongServiceID,
			Force:     true,
		})
		testErr = err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrInvalidParams, testErr.Code)

		err = disco.UnregisterService(ctx, &pb.DeleteServiceRequest{
			ServiceId: "notexistservice",
			Force:     true,
		})
		testErr = err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrServiceNotExists, testErr.Code)
	})

	t.Run("when delete sc service, should be failed", func(t *testing.T) {
		serviceID, err := disco.ExistService(ctx, &pb.GetExistenceRequest{
			Type:        "microservice",
			Environment: core.Service.Environment,
			AppId:       core.Service.AppId,
			ServiceName: core.Service.ServiceName,
			Version:     core.Service.Version,
		})
		assert.NoError(t, err)

		err = disco.UnregisterService(ctx, &pb.DeleteServiceRequest{
			ServiceId: serviceID,
			Force:     true,
		})
		testErr := err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrInvalidParams, testErr.Code)
	})

	t.Run("when delete a service contains instances with not force flag, should be failed", func(t *testing.T) {
		err := disco.UnregisterService(ctx, &pb.DeleteServiceRequest{
			ServiceId: serviceContainInstId,
			Force:     false,
		})
		testErr := err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrDeployedInstance, testErr.Code)
	})

	t.Run("when delete a service contains instances with force flag, should be ok", func(t *testing.T) {
		err := disco.UnregisterService(ctx, &pb.DeleteServiceRequest{
			ServiceId: serviceContainInstId,
			Force:     true,
		})
		assert.NoError(t, err)
	})

	t.Run("when delete a service depended on consumer with not force flag, should be failed", func(t *testing.T) {
		err := disco.UnregisterService(ctx, &pb.DeleteServiceRequest{
			ServiceId: serviceNoInstId,
			Force:     false,
		})
		testErr := err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrDependedOnConsumer, testErr.Code)
	})

	t.Run("when delete a service depended on consumer with force flag, should be ok", func(t *testing.T) {
		err := disco.UnregisterService(ctx, &pb.DeleteServiceRequest{
			ServiceId: serviceNoInstId,
			Force:     true,
		})
		assert.NoError(t, err)
	})

	t.Run("clean test data, should be ok", func(t *testing.T) {
		err := disco.UnregisterService(ctx, &pb.DeleteServiceRequest{
			ServiceId: serviceConsumerId,
			Force:     false,
		})
		assert.NoError(t, err)
	})
}

func TestUnregisterManyService(t *testing.T) {
	t.Run("unregister exist services, should be passed", func(t *testing.T) {
		var (
			serviceId1 string
			serviceId2 string
		)

		respCreate, err := disco.RegisterService(getContext(), &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				ServiceName: "batch_delete_services_1",
				AppId:       "batch_delete",
				Version:     "1.0.0",
				Level:       "FRONT",
				Status:      "UP",
			},
		})
		assert.NoError(t, err)
		serviceId1 = respCreate.ServiceId

		respCreate, err = disco.RegisterService(getContext(), &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				ServiceName: "batch_delete_services_2",
				AppId:       "batch_delete",
				Version:     "1.0.0",
				Level:       "FRONT",
				Status:      "UP",
			},
		})
		assert.NoError(t, err)
		serviceId2 = respCreate.ServiceId

		resp, err := disco.UnregisterManyService(getContext(), &pb.DelServicesRequest{
			ServiceIds: []string{serviceId1, serviceId2},
			Force:      false,
		})
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, 2, len(resp.Services))
		for _, result := range resp.Services {
			assert.Empty(t, result.ErrMessage)
		}
	})

	t.Run("when batch delete services, should be ok", func(t *testing.T) {
		_, err := disco.UnregisterManyService(getContext(), &pb.DelServicesRequest{
			ServiceIds: []string{},
			Force:      false,
		})
		testErr := err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrInvalidParams, testErr.Code)

		_, err = disco.UnregisterManyService(getContext(), &pb.DelServicesRequest{
			ServiceIds: []string{TooLongServiceID},
			Force:      false,
		})
		testErr = err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrInvalidParams, testErr.Code)
	})

	t.Run("batch delete services contain instances, should be failed", func(t *testing.T) {
		var (
			serviceIdFailed1 string
			serviceIdFailed2 string
		)

		respCreate, err := disco.RegisterService(getContext(), &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				ServiceName: "batch_delete_services_failed_1",
				AppId:       "batch_delete",
				Version:     "1.0.0",
				Level:       "FRONT",
				Status:      "UP",
			},
		})
		assert.NoError(t, err)
		serviceIdFailed1 = respCreate.ServiceId

		instances := []*pb.MicroServiceInstance{
			{
				Endpoints: []string{
					"batchDeleteServices:127.0.0.2:8081",
				},
				HostName: "batch-delete-host",
				Status:   pb.MSI_UP,
			},
		}

		respCreate, err = disco.RegisterService(getContext(), &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				ServiceName: "batch_delete_services_failed_2",
				AppId:       "batch_delete",
				Version:     "1.0.0",
				Level:       "FRONT",
				Status:      "UP",
			},
			Instances: instances,
		})
		assert.NoError(t, err)
		serviceIdFailed2 = respCreate.ServiceId

		resp, err := disco.UnregisterManyService(getContext(), &pb.DelServicesRequest{
			ServiceIds: []string{serviceIdFailed1, serviceIdFailed2},
			Force:      false,
		})
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, 2, len(resp.Services))
		for _, result := range resp.Services {
			if serviceIdFailed1 == result.ServiceId {
				assert.Empty(t, result.ErrMessage)
			}
			if serviceIdFailed2 == result.ServiceId {
				assert.NotEmpty(t, result.ErrMessage)
			}
		}
	})
}

func TestExistService(t *testing.T) {
	var (
		serviceId1 string
		serviceId2 string
	)
	ctx := getContext()
	defer disco.UnregisterManyService(ctx, &pb.DelServicesRequest{
		ServiceIds: []string{serviceId1, serviceId2},
		Force:      true,
	})

	t.Run("prepare data, should be passed", func(t *testing.T) {
		service := &pb.MicroService{
			Alias:       "es",
			ServiceName: "exist_service",
			AppId:       "exist_appId",
			Version:     "1.0.0",
			Level:       "FRONT",
			Schemas: []string{
				"first_schemaId",
			},
			Status: "UP",
		}
		respCreateService, err := disco.RegisterService(ctx, &pb.CreateServiceRequest{
			Service: service,
		})
		assert.NoError(t, err)
		serviceId1 = respCreateService.ServiceId

		service.ServiceId = ""
		service.Environment = pb.ENV_PROD
		respCreateService, err = disco.RegisterService(ctx, &pb.CreateServiceRequest{
			Service: service,
		})
		assert.NoError(t, err)
		serviceId2 = respCreateService.ServiceId
	})

	t.Run("when type is invalid, should be failed", func(t *testing.T) {
		_, err := disco.ExistService(ctx, &pb.GetExistenceRequest{
			Type: "nonetype",
		})
		testErr := err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrInvalidParams, testErr.Code)
	})

	t.Run("when param is invalid, should be failed", func(t *testing.T) {
		_, err := disco.ExistService(ctx, &pb.GetExistenceRequest{
			Type:        "microservice",
			ServiceName: TooLongExistence,
			Version:     "1.0.0",
			AppId:       "exist_appId",
		})
		testErr := err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrInvalidParams, testErr.Code)

		_, err = disco.ExistService(ctx, &pb.GetExistenceRequest{
			Type:        "microservice",
			AppId:       "default",
			ServiceName: "",
			Version:     "3.0.0",
		})
		testErr = err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrInvalidParams, testErr.Code)

		_, err = disco.ExistService(ctx, &pb.GetExistenceRequest{
			Type:        "microservice",
			AppId:       TooLongAppID,
			ServiceName: "exist-invalid-appid",
			Version:     "3.0.0",
		})
		testErr = err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrInvalidParams, testErr.Code)

		_, err = disco.ExistService(ctx, &pb.GetExistenceRequest{
			Type:        "microservice",
			AppId:       "default",
			ServiceName: "exist-invalid-version",
			Version:     "3.32768.0",
		})
		testErr = err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrInvalidParams, testErr.Code)

		_, err = disco.ExistService(ctx, &pb.GetExistenceRequest{
			Type:        "microservice",
			AppId:       "default",
			ServiceName: "exist-invalid-version",
			Version:     "3.0.0.0.0",
		})
		testErr = err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrInvalidParams, testErr.Code)

		_, err = disco.ExistService(ctx, &pb.GetExistenceRequest{
			Type:        "microservice",
			AppId:       "default",
			ServiceName: "exist-empty-version",
			Version:     "",
		})
		testErr = err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrInvalidParams, testErr.Code)

		_, err = disco.ExistService(ctx, &pb.GetExistenceRequest{
			Type:        "microservice",
			AppId:       "exist_appId",
			ServiceName: "exist_service",
			Version:     "0.0.0-1.0.0",
		})
		testErr = err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrInvalidParams, testErr.Code)

		_, err = disco.ExistService(ctx, &pb.GetExistenceRequest{
			Type:        "microservice",
			AppId:       "exist_appId",
			ServiceName: "exist_service",
			Version:     "0.0.0+",
		})
		testErr = err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrInvalidParams, testErr.Code)

		_, err = disco.ExistService(ctx, &pb.GetExistenceRequest{
			Type:        "microservice",
			AppId:       "exist_appId",
			ServiceName: "exist_service",
			Version:     "latest",
		})
		testErr = err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrInvalidParams, testErr.Code)
	})

	t.Run("when service does not exist, should be failed", func(t *testing.T) {
		_, err := disco.ExistService(ctx, &pb.GetExistenceRequest{
			Type:        "microservice",
			AppId:       "exist_appId",
			ServiceName: "notExistService",
			Version:     "1.0.0",
		})
		testErr := err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrServiceNotExists, testErr.Code)

		_, err = disco.ExistService(ctx, &pb.GetExistenceRequest{
			Type:        "microservice",
			Environment: pb.ENV_TEST,
			AppId:       "exist_appId",
			ServiceName: "exist_service",
			Version:     "1.0.0",
		})
		testErr = err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrServiceNotExists, testErr.Code)

		_, err = disco.ExistService(ctx, &pb.GetExistenceRequest{
			Type:        "microservice",
			Environment: pb.ENV_TEST,
			AppId:       "exist_appId",
			ServiceName: "es",
			Version:     "1.0.0",
		})
		testErr = err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrServiceNotExists, testErr.Code)

		_, err = disco.ExistService(ctx, &pb.GetExistenceRequest{
			Type:        "microservice",
			AppId:       "exist_appId",
			ServiceName: "exist_service",
			Version:     "2.0.0",
		})
		testErr = err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrServiceVersionNotExists, testErr.Code)
	})

	t.Run("when service exists, should be failed", func(t *testing.T) {
		serviceID, err := disco.ExistService(ctx, &pb.GetExistenceRequest{
			Type:        "microservice",
			AppId:       "exist_appId",
			ServiceName: "exist_service",
			Version:     "1.0.0",
		})
		assert.NoError(t, err)
		assert.Equal(t, serviceId1, serviceID)

		serviceID, err = disco.ExistService(ctx, &pb.GetExistenceRequest{
			Type:        "microservice",
			Environment: pb.ENV_PROD,
			AppId:       "exist_appId",
			ServiceName: "exist_service",
			Version:     "1.0.0",
		})
		assert.NoError(t, err)
		assert.Equal(t, serviceId2, serviceID)

		serviceID, err = disco.ExistService(ctx, &pb.GetExistenceRequest{
			Type:        "microservice",
			AppId:       "exist_appId",
			ServiceName: "es",
			Version:     "1.0.0",
		})
		assert.NoError(t, err)
		assert.Equal(t, serviceId1, serviceID)

		serviceID, err = disco.ExistService(ctx, &pb.GetExistenceRequest{
			Type:        "microservice",
			Environment: pb.ENV_PROD,
			AppId:       "exist_appId",
			ServiceName: "es",
			Version:     "1.0.0",
		})
		assert.NoError(t, err)
		assert.Equal(t, serviceId2, serviceID)
	})
}

func TestUpdateService(t *testing.T) {
	var (
		serviceId string
	)
	ctx := getContext()

	defer disco.UnregisterService(ctx, &pb.DeleteServiceRequest{ServiceId: serviceId, Force: true})

	t.Run("prepare data, should be passed", func(t *testing.T) {
		respCreateService, err := disco.RegisterService(ctx, &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				Alias:       "es",
				ServiceName: "update_prop_service",
				AppId:       "update_prop_appId",
				Version:     "1.0.0",
				Level:       "FRONT",
				Status:      "UP",
			},
		})
		assert.NoError(t, err)
		serviceId = respCreateService.ServiceId
	})

	t.Run("when property is not nil, should be ok", func(t *testing.T) {
		r := &pb.UpdateServicePropsRequest{
			ServiceId:  serviceId,
			Properties: make(map[string]string),
		}
		r2 := &pb.UpdateServicePropsRequest{
			ServiceId:  serviceId,
			Properties: make(map[string]string),
		}
		r.Properties["test"] = "1"
		r2.Properties["k"] = "v"
		err := disco.PutServiceProperties(ctx, r)
		assert.NoError(t, err)

		service, err := disco.GetService(ctx, &pb.GetServiceRequest{
			ServiceId: serviceId,
		})
		assert.NoError(t, err)
		assert.Equal(t, 1, len(service.Properties))
		assert.Equal(t, "1", service.Properties["test"])

		err = disco.PutServiceProperties(ctx, r2)
		assert.NoError(t, err)

		service, err = disco.GetService(ctx, &pb.GetServiceRequest{
			ServiceId: serviceId,
		})
		assert.NoError(t, err)
		assert.Equal(t, 1, len(service.Properties))
		assert.Equal(t, "v", service.Properties["k"])
	})

	t.Run("when service does not exist, should be failed", func(t *testing.T) {
		r := &pb.UpdateServicePropsRequest{
			ServiceId:  "notexistservice",
			Properties: make(map[string]string),
		}
		err := disco.PutServiceProperties(ctx, r)
		testErr := err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrServiceNotExists, testErr.Code)
	})

	t.Run("when remove the properties, should be ok", func(t *testing.T) {
		r := &pb.UpdateServicePropsRequest{
			ServiceId:  serviceId,
			Properties: nil,
		}
		err := disco.PutServiceProperties(ctx, r)
		assert.NoError(t, err)

		r = &pb.UpdateServicePropsRequest{
			ServiceId:  "",
			Properties: map[string]string{},
		}
		err = disco.PutServiceProperties(ctx, r)
		testErr := err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrInvalidParams, testErr.Code)
	})

	t.Run("when request is invalid, should be failed", func(t *testing.T) {
		r := &pb.UpdateServicePropsRequest{
			ServiceId:  TooLongServiceID,
			Properties: map[string]string{},
		}
		err := disco.PutServiceProperties(ctx, r)
		testErr := err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrInvalidParams, testErr.Code)
	})
}

func TestListService(t *testing.T) {
	t.Run("when query all services, should be ok", func(t *testing.T) {
		resp, err := disco.ListService(getContext(), &pb.GetServicesRequest{})
		assert.NoError(t, err)
		assert.NotEqual(t, 0, len(resp.Services))
	})

	t.Run("when query a not exist service by serviceId, should be failed", func(t *testing.T) {
		service, err := disco.GetService(getContext(), &pb.GetServiceRequest{
			ServiceId: "",
		})
		testErr := err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrInvalidParams, testErr.Code)
		assert.Nil(t, service)

		service, err = disco.GetService(getContext(), &pb.GetServiceRequest{
			ServiceId: "notexistservice",
		})
		testErr = err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrServiceNotExists, testErr.Code)
		assert.Nil(t, service)

		service, err = disco.GetService(getContext(), &pb.GetServiceRequest{
			ServiceId: TooLongServiceID,
		})
		testErr = err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrInvalidParams, testErr.Code)
		assert.Nil(t, service)
	})
}

func TestServiceUsage(t *testing.T) {
	t.Run("get domain/project without service usage, should return 0", func(t *testing.T) {
		usage, err := disco.ServiceUsage(context.Background(), &pb.GetServiceCountRequest{
			Domain:  "domain_without_service",
			Project: "project_without_service",
		})
		assert.NoError(t, err)
		assert.Equal(t, int64(0), usage)
	})

	t.Run("get domain/project with 1 service usage, should return 1", func(t *testing.T) {
		ctx := util.SetDomainProject(context.Background(), "domain_with_service", "project_with_service")
		resp, err := disco.RegisterService(ctx, &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				ServiceName: "test",
			},
		})
		assert.NoError(t, err)
		defer disco.UnregisterService(ctx, &pb.DeleteServiceRequest{ServiceId: resp.ServiceId, Force: true})

		usage, err := disco.ServiceUsage(context.Background(), &pb.GetServiceCountRequest{
			Domain:  "domain_with_service",
			Project: "project_with_service",
		})
		assert.NoError(t, err)
		assert.Equal(t, int64(1), usage)
	})
}
