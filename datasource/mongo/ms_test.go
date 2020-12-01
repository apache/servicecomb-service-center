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
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/apache/servicecomb-service-center/datasource"
	"github.com/apache/servicecomb-service-center/datasource/mongo"
	"github.com/apache/servicecomb-service-center/datasource/mongo/client"
	"github.com/apache/servicecomb-service-center/server/plugin/quota"
	pb "github.com/go-chassis/cari/discovery"
	"github.com/go-chassis/go-chassis/v2/storage"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
)

func init() {
	config := storage.Options{
		URI: "mongodb://localhost:27017",
	}
	client.NewMongoClient(config)
}

func TestServiceRegister(t *testing.T) {
	t.Run("Register service by mongo, should pass", func(t *testing.T) {
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
				ServiceId:   "service-ms-appID_id",
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
		resp, err := datasource.Instance().RegisterService(getContext(), request)
		assert.NotNil(t, resp)
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, resp.Response.GetCode())
	})
	t.Run("Register service with the same id by mongo, should pass", func(t *testing.T) {
		request := &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				ServiceId:   "service-ms-appID_id",
				AppId:       "service-ms-appID11",
				ServiceName: "service-ms-serviceName11",
				Version:     "32767.32767.32767.3276711",
				Alias:       "service-ms-alias11",
			},
		}
		resp, err := datasource.Instance().RegisterService(getContext(), request)
		assert.NotNil(t, resp)
		assert.NoError(t, err)
		assert.Equal(t, pb.ErrServiceAlreadyExists, resp.Response.GetCode())
	})
	t.Run("Register service with the same id by mongo, should pass", func(t *testing.T) {
		request := &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				ServiceId:   "service-ms-appID_id_test",
				AppId:       "service-ms-appID",
				ServiceName: "service-ms-serviceName",
				Version:     "32767.32767.32767.32767",
				Alias:       "service-ms-alias",
			},
		}
		resp, err := datasource.Instance().RegisterService(getContext(), request)
		assert.NotNil(t, resp)
		assert.NoError(t, err)
		assert.Equal(t, pb.ErrServiceAlreadyExists, resp.Response.GetCode())
	})
}

func TestGetService(t *testing.T) {
	t.Run("get a exist service by mongo, should pass", func(t *testing.T) {
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

		resp, err := datasource.Instance().RegisterService(getContext(), request)
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, resp.Response.GetCode())

		request = &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				ServiceId:   "ms-service-query-id1",
				ServiceName: "ms-service-query1",
				AppId:       "default",
				Version:     "1.0.4",
				Level:       "BACK",
				Properties:  make(map[string]string),
			},
		}

		resp, err = datasource.Instance().RegisterService(getContext(), request)
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, resp.Response.GetCode())

		// search service by serviceID
		queryResp, err := datasource.Instance().GetService(getContext(), &pb.GetServiceRequest{
			ServiceId: "ms-service-query-id",
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, queryResp.Response.GetCode())
	})
	t.Run("get all service by mongo, should pass", func(t *testing.T) {
		request := &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				ServiceId:   "ms-service-query-id3",
				ServiceName: "ms-service-query3",
				AppId:       "default",
				Version:     "1.0.4",
				Level:       "BACK",
				Properties:  make(map[string]string),
			},
		}

		resp, err := datasource.Instance().RegisterService(getContext(), request)
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, resp.Response.GetCode())

		// search service by serviceID
		queryResp, err := datasource.Instance().GetServices(getContext(), &pb.GetServicesRequest{})
		assert.NoError(t, err)
		assert.Greater(t, len(queryResp.Services), 0)
	})
	t.Run("get a exist service with id by mongo, should pass", func(t *testing.T) {
		queryResp, err := datasource.Instance().ExistServiceByID(getContext(), &pb.GetExistenceByIDRequest{
			ServiceId: "ms-service-query-id1",
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, queryResp.Response.GetCode())
		assert.Equal(t, true, queryResp.Exist)
	})
	t.Run("query a service by a not existed serviceId, should not pass", func(t *testing.T) {
		// not exist service
		resp, err := datasource.Instance().GetService(getContext(), &pb.GetServiceRequest{
			ServiceId: "no-exist-service",
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ErrServiceNotExists, resp.Response.GetCode())
	})

}

func TestUpdateService(t *testing.T) {
	t.Run("update service by mongo, should pass", func(t *testing.T) {
		request := &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				ServiceId:   "ms-service-update-new-id",
				ServiceName: "ms-service-update",
				AppId:       "default",
				Version:     "1.0.4",
				Level:       "BACK",
				Properties:  make(map[string]string),
			},
		}

		resp, err := datasource.Instance().RegisterService(getContext(), request)
		assert.NoError(t, err)
		assert.Equal(t, resp.Response.GetCode(), pb.ResponseSuccess)

		requestNew := &pb.UpdateServicePropsRequest{
			ServiceId:  "ms-service-update-new-id",
			Properties: make(map[string]string),
		}
		requestNew.Properties["k"] = "v"
		res, err := datasource.Instance().UpdateService(getContext(), requestNew)
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, res.Response.GetCode())

		respGetService, err := datasource.Instance().GetService(getContext(), &pb.GetServiceRequest{ServiceId: "ms-service-update-new-id"})
		assert.NoError(t, err)
		assert.Equal(t, "v", respGetService.Service.Properties["k"])
	})
}

func TestTagsAdd(t *testing.T) {
	// create service
	t.Run("create service, the request is valid, should pass", func(t *testing.T) {
		svc1 := &pb.MicroService{
			ServiceId:   "service_tag_id",
			AppId:       "create_tag_group_ms",
			ServiceName: "create_tag_service_ms",
			Version:     "1.0.0",
			Level:       "FRONT",
			Status:      pb.MS_UP,
		}
		resp, err := datasource.Instance().RegisterService(getContext(), &pb.CreateServiceRequest{
			Service: svc1,
		})
		assert.NoError(t, err)
		assert.NotEqual(t, "", resp.ServiceId)
	})

	//
	t.Run("create service, the request is valid, should pass", func(t *testing.T) {
		defaultQuota := quota.DefaultTagQuota
		tags := make(map[string]string, defaultQuota)
		for i := 0; i < defaultQuota; i++ {
			s := "tag" + strconv.Itoa(i)
			tags[s] = s
		}
		resp, err := datasource.Instance().AddTags(getContext(), &pb.AddServiceTagsRequest{
			ServiceId: "service_tag_id",
			Tags:      tags,
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, resp.Response.GetCode())
	})

}

func TestTagsGet(t *testing.T) {
	t.Run("create service and add tags, the request is valid, should pass", func(t *testing.T) {
		svc := &pb.MicroService{
			ServiceId:   "get_tag_group_ms_id",
			AppId:       "get_tag_group_ms",
			ServiceName: "get_tag_service_ms",
			Version:     "1.0.0",
			Level:       "FRONT",
			Status:      pb.MS_UP,
		}
		resp, err := datasource.Instance().RegisterService(getContext(), &pb.CreateServiceRequest{
			Service: svc,
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, resp.Response.GetCode())

		respAddTags, err := datasource.Instance().AddTags(getContext(), &pb.AddServiceTagsRequest{
			ServiceId: "get_tag_group_ms_id",
			Tags: map[string]string{
				"a": "test",
				"b": "b",
			},
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respAddTags.Response.GetCode())
	})
	t.Run("the request is valid", func(t *testing.T) {
		resp, err := datasource.Instance().GetTags(getContext(), &pb.GetServiceTagsRequest{
			ServiceId: "get_tag_group_ms_id",
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, resp.Response.GetCode())
		assert.Equal(t, "test", resp.Tags["a"])
	})
}

func TestTagUpdate(t *testing.T) {
	t.Run("add service and add tags, the request is valid, should pass", func(t *testing.T) {
		svc := &pb.MicroService{
			ServiceId:   "update_tag_group_ms_id",
			AppId:       "update_tag_group_ms",
			ServiceName: "update_tag_service_ms",
			Version:     "1.0.0",
			Level:       "FRONT",
			Status:      pb.MS_UP,
		}
		resp, err := datasource.Instance().RegisterService(getContext(), &pb.CreateServiceRequest{
			Service: svc,
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, resp.Response.GetCode())

		respAddTags, err := datasource.Instance().AddTags(getContext(), &pb.AddServiceTagsRequest{
			ServiceId: "update_tag_group_ms_id",
			Tags: map[string]string{
				"a": "test",
				"b": "b",
			},
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respAddTags.Response.GetCode())
	})

	t.Run("the request is valid, should pass", func(t *testing.T) {
		resp, err := datasource.Instance().UpdateTag(getContext(), &pb.UpdateServiceTagRequest{
			ServiceId: "update_tag_group_ms_id",
			Key:       "a",
			Value:     "update",
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, resp.Response.GetCode())
	})
}

func TestTagsDelete(t *testing.T) {
	t.Run("create service and add tags, the request is valid, should pass", func(t *testing.T) {
		resp, err := datasource.Instance().RegisterService(getContext(), &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				ServiceId:   "delete_tag_group_ms_id",
				AppId:       "delete_tag_group_ms",
				ServiceName: "delete_tag_service_ms",
				Version:     "1.0.0",
				Level:       "FRONT",
				Status:      pb.MS_UP,
			},
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, resp.Response.GetCode())

		respAddTages, err := datasource.Instance().AddTags(getContext(), &pb.AddServiceTagsRequest{
			ServiceId: "delete_tag_group_ms_id",
			Tags: map[string]string{
				"a": "test",
				"b": "b",
			},
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respAddTages.Response.GetCode())
	})
	t.Run("the request is valid, should pass", func(t *testing.T) {
		resp, err := datasource.Instance().DeleteTags(getContext(), &pb.DeleteServiceTagsRequest{
			ServiceId: "delete_tag_group_ms_id",
			Keys:      []string{"b"},
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, resp.Response.GetCode())

		respTags, err := datasource.Instance().GetTags(getContext(), &pb.GetServiceTagsRequest{
			ServiceId: "delete_tag_group_ms_id",
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respTags.Response.GetCode())
		assert.Equal(t, "", respTags.Tags["b"])
	})
}

func TestRuleAdd(t *testing.T) {
	t.Run("register service, the request is valid, should pass", func(t *testing.T) {
		respCreateService, err := datasource.Instance().RegisterService(getContext(), &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				ServiceId:   "create_rule_group_ms_id",
				AppId:       "create_rule_group_ms",
				ServiceName: "create_rule_service_ms",
				Version:     "1.0.0",
				Level:       "FRONT",
				Status:      pb.MS_UP,
			},
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respCreateService.Response.GetCode())
	})
	t.Run("register service, the request is valid, should pass", func(t *testing.T) {
		respAddRule, err := datasource.Instance().AddRule(getContext(), &pb.AddServiceRulesRequest{
			ServiceId: "create_rule_group_ms_id",
			Rules: []*pb.AddOrUpdateServiceRule{
				{
					RuleType:    "BLACK",
					Attribute:   "ServiceName",
					Pattern:     "Test*",
					Description: "test black",
				},
			},
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respAddRule.Response.GetCode())
		ruleId := respAddRule.RuleIds[0]
		assert.NotEqual(t, "", ruleId)
	})
	t.Run("request rule is already exist, should pass", func(t *testing.T) {
		respAddRule, err := datasource.Instance().AddRule(getContext(), &pb.AddServiceRulesRequest{
			ServiceId: "create_rule_group_ms_id",
			Rules: []*pb.AddOrUpdateServiceRule{
				{
					RuleType:    "BLACK",
					Attribute:   "ServiceName",
					Pattern:     "Test*",
					Description: "test black",
				},
			},
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respAddRule.Response.GetCode())
		assert.Equal(t, 0, len(respAddRule.RuleIds))
	})
}

func TestRuleGet(t *testing.T) {
	t.Run("register service and rules, the request is valid, should pass", func(t *testing.T) {
		respCreateService, err := datasource.Instance().RegisterService(getContext(), &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				ServiceId:   "get_rule_group_ms_id",
				AppId:       "get_rule_group_ms",
				ServiceName: "get_rule_service_ms",
				Version:     "1.0.0",
				Level:       "FRONT",
				Status:      pb.MS_UP,
			},
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respCreateService.Response.GetCode())

		respAddRule, err := datasource.Instance().AddRule(getContext(), &pb.AddServiceRulesRequest{
			ServiceId: "get_rule_group_ms_id",
			Rules: []*pb.AddOrUpdateServiceRule{
				{
					RuleType:    "BLACK",
					Attribute:   "ServiceName",
					Pattern:     "Test*",
					Description: "test BLACK",
				},
			},
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respAddRule.Response.GetCode())
		ruleId := respAddRule.RuleIds[0]
		assert.NotEqual(t, "", ruleId)
	})
	t.Run("get rule,  when request is valid, should pass", func(t *testing.T) {
		respGetRule, err := datasource.Instance().GetRules(getContext(), &pb.GetServiceRulesRequest{
			ServiceId: "get_rule_group_ms_id",
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respGetRule.Response.GetCode())
		assert.Equal(t, 1, len(respGetRule.Rules))
	})
}

func TestRuleDelete(t *testing.T) {
	var ruleId string
	t.Run("register service and rules, when request is valid, should pass", func(t *testing.T) {
		respCreateService, err := datasource.Instance().RegisterService(getContext(), &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				ServiceId:   "delete_rule_group_ms_id",
				AppId:       "delete_rule_group_ms",
				ServiceName: "delete_rule_service_ms",
				Version:     "1.0.0",
				Level:       "FRONT",
				Status:      pb.MS_UP,
			},
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respCreateService.Response.GetCode())

		respAddRule, err := datasource.Instance().AddRule(getContext(), &pb.AddServiceRulesRequest{
			ServiceId: "delete_rule_group_ms_id",
			Rules: []*pb.AddOrUpdateServiceRule{
				{
					RuleType:    "BLACK",
					Attribute:   "ServiceName",
					Pattern:     "Test*",
					Description: "test BLACK",
				},
			},
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respAddRule.Response.GetCode())
		ruleId = respAddRule.RuleIds[0]
	})
	t.Run("delete rule, when request is valid, should pass", func(t *testing.T) {
		resp, err := datasource.Instance().DeleteRule(getContext(), &pb.DeleteServiceRulesRequest{
			ServiceId: "delete_rule_group_ms_id",
			RuleIds:   []string{ruleId},
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, resp.Response.GetCode())

		respGetRule, err := datasource.Instance().GetRules(getContext(), &pb.GetServiceRulesRequest{
			ServiceId: "delete_rule_group_ms_id",
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, resp.Response.GetCode())
		assert.Equal(t, 1, len(respGetRule.Rules))
	})
}

func TestRuleUpdate(t *testing.T) {
	var ruleId string
	t.Run("create service and rules, when request is valid, should pass", func(t *testing.T) {
		respCreateService, err := datasource.Instance().RegisterService(getContext(), &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				ServiceId:   "update_rule_group_ms_id",
				AppId:       "update_rule_group_ms",
				ServiceName: "update_rule_service_ms",
				Version:     "1.0.0",
				Level:       "FRONT",
				Status:      pb.MS_UP,
			},
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respCreateService.Response.GetCode())

		respAddRule, err := datasource.Instance().AddRule(getContext(), &pb.AddServiceRulesRequest{
			ServiceId: "update_rule_group_ms_id",
			Rules: []*pb.AddOrUpdateServiceRule{
				{
					RuleType:    "BLACK",
					Attribute:   "ServiceName",
					Pattern:     "Test*",
					Description: "test BLACK",
				},
			},
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respAddRule.Response.GetCode())
		assert.NotEqual(t, "", respAddRule.RuleIds[0])
		ruleId = respAddRule.RuleIds[0]
	})
	t.Run("update rule, when request is valid, should pass", func(t *testing.T) {
		resp, err := datasource.Instance().UpdateRule(getContext(), &pb.UpdateServiceRuleRequest{
			ServiceId: "update_rule_group_ms_id",
			RuleId:    ruleId,
			Rule: &pb.AddOrUpdateServiceRule{
				RuleType:    "BLACK",
				Attribute:   "AppId",
				Pattern:     "Test*",
				Description: "test white update",
			},
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, resp.Response.GetCode())
	})
}

func TestInstance_Creat(t *testing.T) {
	var serviceId string

	t.Run("create service, when request is valid, should pass", func(t *testing.T) {
		insertRes, err := datasource.Instance().RegisterService(getContext(), &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				ServiceId:   "service1",
				ServiceName: "create_instance_service_ms",
				AppId:       "create_instance_ms",
				Version:     "1.0.0",
				Level:       "FRONT",
				Status:      pb.MS_UP,
			},
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, insertRes.Response.GetCode())
		serviceId = insertRes.ServiceId
	})

	t.Run("register instance, when request is valid, should pass", func(t *testing.T) {
		respCreateInst, err := datasource.Instance().RegisterInstance(getContext(), &pb.RegisterInstanceRequest{
			Instance: &pb.MicroServiceInstance{
				ServiceId: serviceId,
				Endpoints: []string{
					"createInstance_ms:127.0.0.1:8080",
				},
				HostName: "UT-HOST",
				Status:   pb.MSI_UP,
			},
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respCreateInst.Response.GetCode())
		assert.NotEqual(t, "ins_instance", respCreateInst.InstanceId)

		respCreateInst, err = datasource.Instance().RegisterInstance(getContext(), &pb.RegisterInstanceRequest{
			Instance: &pb.MicroServiceInstance{
				InstanceId: "instance2",
				ServiceId:  serviceId,
				Endpoints: []string{
					"createInstance_ms:127.0.0.1:8080",
				},
				HostName: "UT-HOST",
				Status:   pb.MSI_UP,
			},
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respCreateInst.Response.GetCode())
		assert.Equal(t, "instance2", respCreateInst.InstanceId)
	})

	t.Run("update the same instance, should pass", func(t *testing.T) {
		resp, err := datasource.Instance().RegisterInstance(getContext(), &pb.RegisterInstanceRequest{
			Instance: &pb.MicroServiceInstance{
				ServiceId:  serviceId,
				InstanceId: "instance3",
				Endpoints: []string{
					"sameInstance:127.0.0.1:8080",
				},
				HostName: "UT-HOST",
				Status:   pb.MSI_UP,
			},
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, resp.Response.GetCode())

		resp, err = datasource.Instance().RegisterInstance(getContext(), &pb.RegisterInstanceRequest{
			Instance: &pb.MicroServiceInstance{
				ServiceId:  serviceId,
				InstanceId: "instance4",
				Endpoints: []string{
					"sameInstance:127.0.0.1:8080",
				},
				HostName: "UT-HOST",
				Status:   pb.MSI_UP,
			},
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, resp.Response.GetCode())
		assert.Equal(t, "instance4", resp.InstanceId)
	})

	t.Run("delete test data", func(t *testing.T) {
		_, err := client.GetMongoClient().Delete(getContext(), mongo.CollectionService, bson.M{"domain": "default", "project": "default"})
		assert.NoError(t, err)

		_, err = client.GetMongoClient().Delete(getContext(), mongo.CollectionInstance, bson.M{"domain": "default", "project": "default"})
		assert.NoError(t, err)
	})
}

func TestInstance_update(t *testing.T) {

	var (
		serviceId  string
		instanceId string
	)

	t.Run("register service and instance, when request is valid, should pass", func(t *testing.T) {
		respCreateService, err := datasource.Instance().RegisterService(getContext(), &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				ServiceId:   "service1",
				ServiceName: "update_instance_service_ms",
				AppId:       "update_instance_service_ms",
				Version:     "1.0.0",
				Level:       "FRONT",
				Status:      pb.MS_UP,
			},
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respCreateService.Response.GetCode())
		serviceId = respCreateService.ServiceId

		respCreateInstance, err := datasource.Instance().RegisterInstance(getContext(), &pb.RegisterInstanceRequest{
			Instance: &pb.MicroServiceInstance{
				ServiceId:  serviceId,
				InstanceId: "instance1",
				Endpoints: []string{
					"updateInstance:127.0.0.1:8080",
				},
				HostName:   "UT-HOST-MS",
				Status:     pb.MSI_UP,
				Properties: map[string]string{"nodeIP": "test"},
			},
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respCreateInstance.Response.GetCode())
		instanceId = respCreateInstance.InstanceId
	})

	t.Run("update instance status, should pass", func(t *testing.T) {
		respUpdateStatus, err := datasource.Instance().UpdateInstanceStatus(getContext(), &pb.UpdateInstanceStatusRequest{
			ServiceId:  serviceId,
			InstanceId: instanceId,
			Status:     pb.MSI_DOWN,
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respUpdateStatus.Response.GetCode())

		respUpdateStatus, err = datasource.Instance().UpdateInstanceStatus(getContext(), &pb.UpdateInstanceStatusRequest{
			ServiceId:  serviceId,
			InstanceId: instanceId,
			Status:     pb.MSI_OUTOFSERVICE,
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respUpdateStatus.Response.GetCode())

		respUpdateStatus, err = datasource.Instance().UpdateInstanceStatus(getContext(), &pb.UpdateInstanceStatusRequest{
			ServiceId:  serviceId,
			InstanceId: instanceId,
			Status:     pb.MSI_STARTING,
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respUpdateStatus.Response.GetCode())

		respUpdateStatus, err = datasource.Instance().UpdateInstanceStatus(getContext(), &pb.UpdateInstanceStatusRequest{
			ServiceId:  serviceId,
			InstanceId: instanceId,
			Status:     pb.MSI_TESTING,
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respUpdateStatus.Response.GetCode())

		respUpdateStatus, err = datasource.Instance().UpdateInstanceStatus(getContext(), &pb.UpdateInstanceStatusRequest{
			ServiceId:  serviceId,
			InstanceId: instanceId,
			Status:     pb.MSI_UP,
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respUpdateStatus.Response.GetCode())

		respUpdateStatus, err = datasource.Instance().UpdateInstanceStatus(getContext(), &pb.UpdateInstanceStatusRequest{
			ServiceId:  serviceId,
			InstanceId: "notexistins",
			Status:     pb.MSI_STARTING,
		})
		assert.NoError(t, err)
		assert.NotEqual(t, pb.ResponseSuccess, respUpdateStatus.Response.GetCode())
	})

	t.Run("update instance properties, should pass", func(t *testing.T) {
		respUpdateProperties, err := datasource.Instance().UpdateInstanceProperties(getContext(),
			&pb.UpdateInstancePropsRequest{
				ServiceId:  serviceId,
				InstanceId: instanceId,
				Properties: map[string]string{
					"test": "test",
				},
			})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respUpdateProperties.Response.GetCode())

		size := 1000
		properties := make(map[string]string, size)
		for i := 0; i < size; i++ {
			s := strconv.Itoa(i) + strings.Repeat("x", 253)
			properties[s] = s
		}
		respUpdateProperties, err = datasource.Instance().UpdateInstanceProperties(getContext(),
			&pb.UpdateInstancePropsRequest{
				ServiceId:  serviceId,
				InstanceId: instanceId,
				Properties: properties,
			})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respUpdateProperties.Response.GetCode())

		respUpdateProperties, err = datasource.Instance().UpdateInstanceProperties(getContext(),
			&pb.UpdateInstancePropsRequest{
				ServiceId:  serviceId,
				InstanceId: "not_exist_ins",
				Properties: map[string]string{
					"test": "test",
				},
			})
		assert.NoError(t, err)
		assert.NotEqual(t, pb.ResponseSuccess, respUpdateProperties.Response.GetCode())

		respUpdateProperties, err = datasource.Instance().UpdateInstanceProperties(getContext(),
			&pb.UpdateInstancePropsRequest{
				ServiceId:  serviceId,
				InstanceId: instanceId,
			})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respUpdateProperties.Response.GetCode())

		respUpdateProperties, err = datasource.Instance().UpdateInstanceProperties(getContext(),
			&pb.UpdateInstancePropsRequest{
				ServiceId:  "not_exist_service",
				InstanceId: instanceId,
				Properties: map[string]string{
					"test": "test",
				},
			})
		assert.NoError(t, err)
		assert.NotEqual(t, pb.ResponseSuccess, respUpdateProperties.Response.GetCode())
	})

	t.Run("delete test data", func(t *testing.T) {
		_, err := client.GetMongoClient().Delete(getContext(), mongo.CollectionService, bson.M{"domain": "default", "project": "default"})
		assert.NoError(t, err)

		_, err = client.GetMongoClient().Delete(getContext(), mongo.CollectionInstance, bson.M{"domain": "default", "project": "default"})
		assert.NoError(t, err)
	})
}

func TestInstance_Query(t *testing.T) {

	var (
		serviceId1  string
		instanceId1 string
	)

	t.Run("register services and instance for testInstance_query, when request is invalid, should pass", func(t *testing.T) {
		insertServiceRes, err := datasource.Instance().RegisterService(getContext(), &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				ServiceId:   "service1",
				AppId:       "query_instance_ms",
				ServiceName: "query_instance_service_ms",
				Version:     "1.0.0",
				Level:       "FRONT",
				Status:      pb.MS_UP,
			},
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, insertServiceRes.Response.GetCode())
		serviceId1 = insertServiceRes.ServiceId

		insertInstanceRes, err := datasource.Instance().RegisterInstance(getContext(), &pb.RegisterInstanceRequest{
			Instance: &pb.MicroServiceInstance{
				InstanceId: "instance1",
				ServiceId:  serviceId1,
				HostName:   "UT-HOST-MS",
				Endpoints: []string{
					"find:127.0.0.1:8080",
				},
				Status: pb.MSI_UP,
			},
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, insertInstanceRes.Response.GetCode())
		instanceId1 = insertInstanceRes.InstanceId
	})

	t.Run("query instance, when request is invalid, should pass", func(t *testing.T) {
		findRes, err := datasource.Instance().FindInstances(getContext(), &pb.FindInstancesRequest{
			ConsumerServiceId: serviceId1,
			AppId:             "query_instance_ms",
			ServiceName:       "query_instance_service_ms",
			VersionRule:       "latest",
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, findRes.Response.GetCode())
		assert.Equal(t, instanceId1, findRes.Instances[0].InstanceId)
	})

	t.Run("batch query instance, when request is invalid, should pass", func(t *testing.T) {
		respFind, err := datasource.Instance().BatchFind(getContext(), &pb.BatchFindInstancesRequest{
			ConsumerServiceId: serviceId1,
			Services: []*pb.FindService{
				{
					Service: &pb.MicroServiceKey{
						AppId:       "query_instance_ms",
						ServiceName: "query_instance_service_ms",
						Version:     "latest",
					},
				},
				{
					Service: &pb.MicroServiceKey{
						AppId:       "query_instance_ms",
						ServiceName: "query_instance_service_ms",
						Version:     "1.0.0+",
					},
				},
				{
					Service: &pb.MicroServiceKey{
						AppId:       "query_instance_ms",
						ServiceName: "query_instance_service_ms",
						Version:     "0.0.0",
					},
				},
			},
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respFind.Response.GetCode())
		assert.Equal(t, int64(0), respFind.Services.Updated[0].Index)

	})

	t.Run("delete test data", func(t *testing.T) {
		_, err := client.GetMongoClient().Delete(getContext(), mongo.CollectionService, bson.M{"domain": "default", "project": "default"})
		assert.NoError(t, err)

		_, err = client.GetMongoClient().Delete(getContext(), mongo.CollectionInstance, bson.M{"domain": "default", "project": "default"})
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

	t.Run("register service and instances, when request is invalid, should pass", func(t *testing.T) {
		respCreateService, err := datasource.Instance().RegisterService(getContext(), &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				ServiceId:   "service1",
				AppId:       "get_instance_ms",
				ServiceName: "get_instance_service_ms",
				Version:     "1.0.0",
				Level:       "FRONT",
				Status:      pb.MS_UP,
			},
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respCreateService.Response.GetCode())
		serviceId1 = respCreateService.ServiceId

		respCreateService, err = datasource.Instance().RegisterService(getContext(), &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				ServiceId:   "service2",
				AppId:       "get_instance_ms",
				ServiceName: "get_instance_service_ms",
				Version:     "1.0.5",
				Level:       "FRONT",
				Status:      pb.MS_UP,
			},
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respCreateService.Response.GetCode())
		serviceId2 = respCreateService.ServiceId

		respCreateInstance, err := datasource.Instance().RegisterInstance(getContext(), &pb.RegisterInstanceRequest{
			Instance: &pb.MicroServiceInstance{
				InstanceId: "instance1",
				ServiceId:  serviceId2,
				HostName:   "UT-HOST-MS",
				Endpoints: []string{
					"get:127.0.0.2:8080",
				},
				Status: pb.MSI_UP,
			},
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respCreateService.Response.GetCode())
		instanceId2 = respCreateInstance.InstanceId

		respCreateService, err = datasource.Instance().RegisterService(getContext(), &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				ServiceId:   "service3",
				AppId:       "get_instance_cross_ms",
				ServiceName: "get_instance_service_ms",
				Version:     "1.0.0",
				Level:       "FRONT",
				Status:      pb.MS_UP,
			},
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respCreateService.Response.GetCode())
		serviceId3 = respCreateService.ServiceId
	})

	t.Run("get between diff apps, when request is invalid, should pass", func(t *testing.T) {
		resp, err := datasource.Instance().GetInstance(getContext(), &pb.GetOneInstanceRequest{
			ConsumerServiceId:  serviceId3,
			ProviderServiceId:  serviceId2,
			ProviderInstanceId: instanceId2,
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, resp.Response.GetCode())
	})

	t.Run("get instances, when request is invalid, should pass", func(t *testing.T) {
		resp, err := datasource.Instance().GetInstances(getContext(), &pb.GetInstancesRequest{
			ConsumerServiceId: "not-exist-service-ms",
			ProviderServiceId: serviceId2,
		})
		assert.NoError(t, err)
		assert.NotEqual(t, pb.ResponseSuccess, resp.Response.GetCode())
		resp, err = datasource.Instance().GetInstances(getContext(), &pb.GetInstancesRequest{
			ConsumerServiceId: serviceId1,
			ProviderServiceId: serviceId2,
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, resp.Response.GetCode())
	})

	t.Run("delete test data", func(t *testing.T) {
		_, err := client.GetMongoClient().Delete(getContext(), mongo.CollectionService, bson.M{"domain": "default", "project": "default"})
		assert.NoError(t, err)

		_, err = client.GetMongoClient().Delete(getContext(), mongo.CollectionInstance, bson.M{"domain": "default", "project": "default"})
		assert.NoError(t, err)
	})
}

func TestInstance_Unregister(t *testing.T) {
	var (
		serviceId  string
		instanceId string
	)

	t.Run("register service and instances, when request is invalid, should pass", func(t *testing.T) {
		respCreateService, err := datasource.Instance().RegisterService(getContext(), &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				ServiceId:   "service1",
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
		assert.Equal(t, pb.ResponseSuccess, respCreateService.Response.GetCode())
		serviceId = respCreateService.ServiceId

		respCreateInstance, err := datasource.Instance().RegisterInstance(getContext(), &pb.RegisterInstanceRequest{
			Instance: &pb.MicroServiceInstance{
				InstanceId: "instance1",
				ServiceId:  serviceId,
				HostName:   "UT-HOST-MS",
				Endpoints: []string{
					"unregister:127.0.0.2:8080",
				},
				Status: pb.MSI_UP,
			},
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respCreateInstance.Response.GetCode())
		instanceId = respCreateInstance.InstanceId
	})

	t.Run("unregister instance, when request is invalid, should pass", func(t *testing.T) {
		resp, err := datasource.Instance().UnregisterInstance(getContext(), &pb.UnregisterInstanceRequest{
			ServiceId:  serviceId,
			InstanceId: instanceId,
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, resp.Response.GetCode())
	})

	t.Run("delete test data", func(t *testing.T) {
		_, err := client.GetMongoClient().Delete(getContext(), mongo.CollectionService, bson.M{"domain": "default", "project": "default"})
		assert.NoError(t, err)

		_, err = client.GetMongoClient().Delete(getContext(), mongo.CollectionInstance, bson.M{"domain": "default", "project": "default"})
		assert.NoError(t, err)
	})
}
