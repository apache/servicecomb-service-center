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
	"github.com/apache/servicecomb-service-center/datasource/mongo/client"
	"github.com/apache/servicecomb-service-center/server/plugin/quota"
	pb "github.com/go-chassis/cari/discovery"
	"github.com/go-chassis/go-chassis/v2/storage"
	"github.com/stretchr/testify/assert"
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

// need mongodb cluster
//func TestServiceDelete(t *testing.T) {
//	t.Run("delete service by mongo, should pass", func(t *testing.T) {
//		request := &pb.CreateServiceRequest{
//			Service: &pb.MicroService{
//				ServiceId:   "ms-service-delete-new-id",
//				ServiceName: "ms-service-delete",
//				AppId:       "default",
//				Version:     "1.0.4",
//				Level:       "BACK",
//				Properties:  make(map[string]string),
//			},
//		}
//
//		resp, err := datasource.Instance().RegisterService(getContext(), request)
//		assert.NoError(t, err)
//		assert.Equal(t, resp.Response.GetCode(), pb.ResponseSuccess)
//
//		res, err := datasource.Instance().UnregisterService(getContext(), &pb.DeleteServiceRequest{
//			ServiceId: "ms-service-delete-new-id",
//			Force:     false,
//		})
//		fmt.Println(res.Response.Message)
//		assert.NoError(t, err)
//		assert.Equal(t, pb.ResponseSuccess, res.Response.GetCode())
//	})
//}

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
	t.Run("create service", func(t *testing.T) {
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
	t.Run("the request is valid", func(t *testing.T) {
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
	t.Run("create service and add tags", func(t *testing.T) {
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
	t.Run("add service and add tags", func(t *testing.T) {
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

	t.Run("the request is valid", func(t *testing.T) {
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
	t.Run("create service and add tags", func(t *testing.T) {
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
	t.Run("the request is valid", func(t *testing.T) {
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
	t.Run("register service and datasource.Instance()", func(t *testing.T) {
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
	t.Run("request is valid", func(t *testing.T) {
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
	t.Run("request rule is already exist", func(t *testing.T) {
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
	t.Run("register service and rules", func(t *testing.T) {
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
	t.Run("get when request is valid", func(t *testing.T) {
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
	t.Run("register service and rules", func(t *testing.T) {
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
	t.Run("delete when request is valid", func(t *testing.T) {
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
	t.Run("create service and rules", func(t *testing.T) {
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
	t.Run("update when request is valid", func(t *testing.T) {
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

// 需要多集群mongo支持
//func TestSchema(t *testing.T) {
//	t.Run("create a schema in production env", func(t *testing.T) {
//		respCreateService, err := datasource.Instance().RegisterService(getContext(), &pb.CreateServiceRequest{
//			Service: &pb.MicroService{
//				ServiceId:   "create_schema_prod_service_ms_id1",
//				AppId:       "create_schema_prod_service_ms",
//				ServiceName: "create_schema_service_service_ms",
//				Version:     "1.0.0",
//				Level:       "FRONT",
//				Status:      pb.MS_UP,
//				Environment: pb.ENV_PROD,
//			},
//		})
//		assert.NoError(t, err)
//		assert.Equal(t, pb.ResponseSuccess, respCreateService.Response.GetCode())
//
//		respCreateService, err = datasource.Instance().RegisterService(getContext(), &pb.CreateServiceRequest{
//			Service: &pb.MicroService{
//				ServiceId:   "create_schema_prod_service_ms_id2",
//				AppId:       "create_schema_prod_service_ms",
//				ServiceName: "create_schema_service_service_ms",
//				Version:     "1.0.1",
//				Level:       "FRONT",
//				Schemas: []string{
//					"first_schemaId_service_ms",
//					"second_schemaId_service_ms",
//				},
//				Status:      pb.MS_UP,
//				Environment: pb.ENV_PROD,
//			},
//		})
//		assert.NoError(t, err)
//		assert.Equal(t, pb.ResponseSuccess, respCreateService.Response.GetCode())
//
//		respModifySchema, err := datasource.Instance().ModifySchema(getContext(), &pb.ModifySchemaRequest{
//			ServiceId: "create_schema_prod_service_ms_id1",
//			SchemaId:  "first_schemaId_service_ms",
//			Schema:    "first_schema_service_ms",
//		})
//		assert.NoError(t, err)
//		assert.Equal(t, pb.ResponseSuccess, respModifySchema.Response.GetCode())
//
//		respModifySchema, err = datasource.Instance().ModifySchema(getContext(), &pb.ModifySchemaRequest{
//			ServiceId: "create_schema_prod_service_ms_id1",
//			SchemaId:  "first_schemaId_service_ms",
//			Schema:    "first_schema_change_service_ms",
//			Summary:   "first0summary1change_service_ms",
//		})
//		assert.NoError(t, err)
//		assert.Equal(t, pb.ResponseSuccess, respModifySchema.Response.GetCode())
//		existRes, err := datasource.Instance().ExistSchema(getContext(), &pb.GetExistenceRequest{
//			ServiceId: "create_schema_prod_service_ms_id1",
//			SchemaId:  "first_schemaId_service_ms",
//		})
//		assert.NoError(t, err)
//		assert.Equal(t, pb.ResponseSuccess, existRes.Response.GetCode())
//		assert.Equal(t, "first0summary1change_service_ms", existRes.Summary)
//
//		resSchemas, err := datasource.Instance().GetAllSchemas(getContext(), &pb.GetAllSchemaRequest{
//			ServiceId:  "create_schema_prod_service_ms_id1",
//			WithSchema: true,
//		})
//		assert.NoError(t, err)
//		assert.Equal(t, 1, len(resSchemas.Schemas))
//	})
//}
