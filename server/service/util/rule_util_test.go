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
package util

import (
	"github.com/apache/incubator-servicecomb-service-center/pkg/util"
	"github.com/apache/incubator-servicecomb-service-center/server/core/proto"
	"golang.org/x/net/context"
	"net/http"
	"testing"
)

func TestRuleFilter_Filter(t *testing.T) {
	rf := RuleFilter{
		DomainProject: "",
		ProviderRules: []*proto.ServiceRule{},
	}
	_, err := rf.Filter(context.Background(), "")
	if err != nil {
		t.Fatalf("RuleFilter Filter failed")
	}
	_, _, err = rf.FilterAll(context.Background(), []string{""})
	if err != nil {
		t.Fatalf("RuleFilter FilterAll failed")
	}
	rf.ProviderRules = []*proto.ServiceRule{
		{},
	}
	_, _, err = rf.FilterAll(context.Background(), []string{""})
	if err != nil {
		t.Fatalf("RuleFilter FilterAll failed")
	}
}

func TestGetRulesUtil(t *testing.T) {
	_, err := GetRulesUtil(context.Background(), "", "")
	if err != nil {
		t.Fatalf("GetRulesUtil failed")
	}

	_, err = GetOneRule(context.Background(), "", "", "")
	if err != nil {
		t.Fatalf("GetOneRule failed")
	}
}

func TestRuleExist(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("TestRuleExist panic")
		}
	}()
	RuleExist(util.SetContext(context.Background(), CTX_CACHEONLY, "1"), "", "", "", "")
	RuleExist(context.Background(), "", "", "", "")
}

func TestGetServiceRuleType(t *testing.T) {
	_, _, err := GetServiceRuleType(context.Background(), "", "")
	if err != nil {
		t.Fatalf("GetServiceRuleType failed")
	}
}

func TestAllowAcrossApp(t *testing.T) {
	err := AllowAcrossDimension(context.Background(), &proto.MicroService{
		AppId: "a",
	}, &proto.MicroService{
		AppId: "a",
	})
	if err != nil {
		t.Fatalf("AllowAcrossApp with the same appId and no property failed")
	}

	err = AllowAcrossDimension(context.Background(), &proto.MicroService{
		AppId: "a",
	}, &proto.MicroService{
		AppId: "c",
	})
	if err == nil {
		t.Fatalf("AllowAcrossApp with the diff appId and no property failed")
	}

	err = AllowAcrossDimension(context.Background(), &proto.MicroService{
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

	err = AllowAcrossDimension(context.Background(), &proto.MicroService{
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

	err = AllowAcrossDimension(context.Background(), &proto.MicroService{
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

	err = AllowAcrossDimension(context.Background(), &proto.MicroService{
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
	err := MatchRules([]*proto.ServiceRule{
		{
			RuleType:  "WHITE",
			Attribute: "",
			Pattern:   "",
		},
	}, nil, nil)
	if err == nil {
		t.Fatalf("MatchRules nil failed")
	}

	err = MatchRules([]*proto.ServiceRule{
		{
			RuleType:  "WHITE",
			Attribute: "",
			Pattern:   "",
		},
	}, &proto.MicroService{}, nil)
	if err == nil {
		t.Fatalf("MatchRules invalid WHITE failed")
	}

	err = MatchRules([]*proto.ServiceRule{
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

	err = MatchRules([]*proto.ServiceRule{
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

	err = MatchRules([]*proto.ServiceRule{
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

	err = MatchRules([]*proto.ServiceRule{
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

	err = MatchRules([]*proto.ServiceRule{
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

	err = MatchRules([]*proto.ServiceRule{
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

	err = MatchRules([]*proto.ServiceRule{
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
	_, _, err := GetAllProviderIds(context.Background(), "", &proto.MicroService{})
	if err != nil {
		t.Fatalf("GetConsumerIdsByProvider invalid failed")
	}

	_, _, err = GetAllConsumerIds(context.Background(), "", &proto.MicroService{
		ServiceId: "a",
	})
	if err != nil {
		t.Fatalf("GetConsumerIdsByProvider WithCacheOnly not exist service failed")
	}

	_, err = GetConsumerIds(context.Background(), "",
		&proto.MicroService{
			ServiceId: "a",
		})
	if err != nil {
		t.Fatalf("GetConsumerIds WithCacheOnly failed")
	}
}

func TestGetProvider(t *testing.T) {
	_, err := GetProviderIds(context.Background(), "",
		&proto.MicroService{
			ServiceId: "a",
		})
	if err != nil {
		t.Fatalf("GetProviderIds WithCacheOnly failed")
	}

	_, _, err = GetAllProviderIds(context.Background(), "", &proto.MicroService{})
	if err != nil {
		t.Fatalf("GetAllProviderIds WithCacheOnly failed")
	}
}

func TestAccessible(t *testing.T) {
	err := Accessible(context.Background(), "", "")
	if err.StatusCode() != http.StatusBadRequest {
		t.Fatalf("Accessible invalid failed")
	}
}
