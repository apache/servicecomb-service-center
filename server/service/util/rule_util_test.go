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
package util_test

import (
	"github.com/apache/incubator-servicecomb-service-center/pkg/util"
	"github.com/apache/incubator-servicecomb-service-center/server/core/proto"
	serviceUtil "github.com/apache/incubator-servicecomb-service-center/server/service/util"
	"golang.org/x/net/context"
	"net/http"
	"testing"
)

func TestRuleFilter_Filter(t *testing.T) {
	rf := serviceUtil.RuleFilter{
		DomainProject: "",
		Provider:      &proto.MicroService{},
		ProviderRules: []*proto.ServiceRule{},
	}
	_, err := rf.Filter(context.Background(), "")
	if err != nil {
		t.Fatalf("RuleFilter Filter failed")
	}
}

func TestGetRulesUtil(t *testing.T) {
	_, err := serviceUtil.GetRulesUtil(util.SetContext(context.Background(), "cacheOnly", "1"), "", "")
	if err != nil {
		t.Fatalf("GetRulesUtil WithCacheOnly failed")
	}

	_, err = serviceUtil.GetRulesUtil(context.Background(), "", "")
	if err == nil {
		t.Fatalf("GetRulesUtil failed")
	}

	_, err = serviceUtil.GetOneRule(util.SetContext(context.Background(), "cacheOnly", "1"), "", "", "")
	if err != nil {
		t.Fatalf("GetOneRule WithCacheOnly failed")
	}

	_, err = serviceUtil.GetOneRule(context.Background(), "", "", "")
	if err == nil {
		t.Fatalf("GetOneRule failed")
	}
}

func TestRuleExist(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("TestRuleExist panic")
		}
	}()
	serviceUtil.RuleExist(util.SetContext(context.Background(), "cacheOnly", "1"), "", "", "", "")
	serviceUtil.RuleExist(context.Background(), "", "", "", "")
}

func TestGetServiceRuleType(t *testing.T) {
	_, _, err := serviceUtil.GetServiceRuleType(util.SetContext(context.Background(), "cacheOnly", "1"), "", "")
	if err != nil {
		t.Fatalf("GetServiceRuleType WithCacheOnly failed")
	}

	_, _, err = serviceUtil.GetServiceRuleType(context.Background(), "", "")
	if err == nil {
		t.Fatalf("GetServiceRuleType failed")
	}
}

func TestAllowAcrossApp(t *testing.T) {
	err := serviceUtil.AllowAcrossDimension(context.Background(), &proto.MicroService{
		AppId: "a",
	}, &proto.MicroService{
		AppId: "a",
	})
	if err != nil {
		t.Fatalf("AllowAcrossApp with the same appId and no property failed")
	}

	err = serviceUtil.AllowAcrossDimension(context.Background(), &proto.MicroService{
		AppId: "a",
	}, &proto.MicroService{
		AppId: "c",
	})
	if err == nil {
		t.Fatalf("AllowAcrossApp with the diff appId and no property failed")
	}

	err = serviceUtil.AllowAcrossDimension(context.Background(), &proto.MicroService{
		AppId: "a",
		Properties: map[string]string{
			proto.PROP_ALLOW_CROSS_APP: "true",
		},
	}, &proto.MicroService{
		AppId: "a",
	})
	if err != nil {
		t.Fatalf("AllowAcrossApp with the same appId and allow property failed")
	}

	err = serviceUtil.AllowAcrossDimension(context.Background(), &proto.MicroService{
		AppId: "a",
		Properties: map[string]string{
			proto.PROP_ALLOW_CROSS_APP: "true",
		},
	}, &proto.MicroService{
		AppId: "b",
	})
	if err != nil {
		t.Fatalf("AllowAcrossApp with the diff appId and allow property failed")
	}

	err = serviceUtil.AllowAcrossDimension(context.Background(), &proto.MicroService{
		AppId: "a",
		Properties: map[string]string{
			proto.PROP_ALLOW_CROSS_APP: "false",
		},
	}, &proto.MicroService{
		AppId: "b",
	})
	if err == nil {
		t.Fatalf("AllowAcrossApp with the diff appId and deny property failed")
	}

	err = serviceUtil.AllowAcrossDimension(context.Background(), &proto.MicroService{
		AppId: "a",
		Properties: map[string]string{
			proto.PROP_ALLOW_CROSS_APP: "",
		},
	}, &proto.MicroService{
		AppId: "b",
	})
	if err == nil {
		t.Fatalf("AllowAcrossApp with the diff appId and empty property failed")
	}
}

func TestMatchRules(t *testing.T) {
	err := serviceUtil.MatchRules([]*proto.ServiceRule{
		{
			RuleType:  "WHITE",
			Attribute: "",
			Pattern:   "",
		},
	}, nil, nil)
	if err == nil {
		t.Fatalf("MatchRules nil failed")
	}

	err = serviceUtil.MatchRules([]*proto.ServiceRule{
		{
			RuleType:  "WHITE",
			Attribute: "",
			Pattern:   "",
		},
	}, &proto.MicroService{}, nil)
	if err == nil {
		t.Fatalf("MatchRules invalid WHITE failed")
	}

	err = serviceUtil.MatchRules([]*proto.ServiceRule{
		{
			RuleType:  "WHITE",
			Attribute: "ServiceName",
			Pattern:   "^a.*",
		},
	}, &proto.MicroService{
		ServiceName: "a",
	}, nil)
	if err != nil {
		t.Fatalf("MatchRules WHITE with field ServiceName failed")
	}

	err = serviceUtil.MatchRules([]*proto.ServiceRule{
		{
			RuleType:  "WHITE",
			Attribute: "tag_a",
			Pattern:   "^b.*",
		},
	}, &proto.MicroService{}, map[string]string{
		"a": "b",
	})
	if err != nil {
		t.Fatalf("MatchRules WHITE with tag b failed")
	}

	err = serviceUtil.MatchRules([]*proto.ServiceRule{
		{
			RuleType:  "WHITE",
			Attribute: "tag_a",
			Pattern:   "^b.*",
		},
	}, &proto.MicroService{}, map[string]string{
		"a": "c",
	})
	if err == nil {
		t.Fatalf("MatchRules WHITE with tag c failed")
	}

	err = serviceUtil.MatchRules([]*proto.ServiceRule{
		{
			RuleType:  "BLACK",
			Attribute: "tag_a",
			Pattern:   "^b.*",
		},
	}, &proto.MicroService{}, map[string]string{
		"a": "b",
	})
	if err == nil {
		t.Fatalf("MatchRules BLACK with tag b failed")
	}

	err = serviceUtil.MatchRules([]*proto.ServiceRule{
		{
			RuleType:  "BLACK",
			Attribute: "ServiceName",
			Pattern:   "^a.*",
		},
	}, &proto.MicroService{
		ServiceName: "a",
	}, nil)
	if err == nil {
		t.Fatalf("MatchRules BLACK with field ServiceName failed")
	}

	err = serviceUtil.MatchRules([]*proto.ServiceRule{
		{
			RuleType:  "BLACK",
			Attribute: "tag_a",
			Pattern:   "^b.*",
		},
	}, &proto.MicroService{}, map[string]string{
		"a": "c",
	})
	if err != nil {
		t.Fatalf("MatchRules BLACK with tag c failed")
	}

	err = serviceUtil.MatchRules([]*proto.ServiceRule{
		{
			RuleType:  "BLACK",
			Attribute: "tag_a",
			Pattern:   "^b.*",
		},
	}, &proto.MicroService{}, map[string]string{
		"b": "b",
	})
	if err != nil {
		t.Fatalf("MatchRules with not exist tag failed")
	}
}

func TestGetConsumer(t *testing.T) {
	_, _, err := serviceUtil.GetConsumerIdsByProvider(context.Background(), "", &proto.MicroService{})
	if err == nil {
		t.Fatalf("GetConsumerIdsByProvider invalid failed")
	}

	_, _, err = serviceUtil.GetConsumerIdsByProvider(context.Background(), "", &proto.MicroService{
		ServiceId: "a",
	})
	if err == nil {
		t.Fatalf("GetConsumerIdsByProvider not exist service failed")
	}

	_, err = serviceUtil.GetConsumersInCache(util.SetContext(context.Background(), "cacheOnly", "1"), "",
		&proto.MicroService{
			ServiceId: "a",
		})
	if err != nil {
		t.Fatalf("GetConsumersInCache WithCacheOnly failed")
	}
}

func TestGetProvider(t *testing.T) {
	_, err := serviceUtil.GetProvidersInCache(util.SetContext(context.Background(), "cacheOnly", "1"), "",
		&proto.MicroService{
			ServiceId: "a",
		})
	if err != nil {
		t.Fatalf("GetProvidersInCache WithCacheOnly failed")
	}

	_, _, err = serviceUtil.GetProviderIdsByConsumer(context.Background(), "", &proto.MicroService{})
	if err == nil {
		t.Fatalf("GetProviderIdsByConsumer invalid failed")
	}

	_, _, err = serviceUtil.GetProviderIdsByConsumer(util.SetContext(context.Background(), "cacheOnly", "1"),
		"", &proto.MicroService{})
	if err != nil {
		t.Fatalf("GetProviderIdsByConsumer WithCacheOnly failed")
	}
}

func TestAccessible(t *testing.T) {
	err := serviceUtil.Accessible(context.Background(), "", "")
	if err.StatusCode() != http.StatusInternalServerError {
		t.Fatalf("Accessible invalid failed")
	}

	err = serviceUtil.Accessible(util.SetContext(context.Background(), "cacheOnly", "1"), "", "")
	if err.StatusCode() == http.StatusInternalServerError {
		t.Fatalf("Accessible WithCacheOnly failed")
	}
}
