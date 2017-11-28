//Copyright 2017 Huawei Technologies Co., Ltd
//
//Licensed under the Apache License, Version 2.0 (the "License");
//you may not use this file except in compliance with the License.
//You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
//Unless required by applicable law or agreed to in writing, software
//distributed under the License is distributed on an "AS IS" BASIS,
//WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//See the License for the specific language governing permissions and
//limitations under the License.
package util_test

import (
	"fmt"
	"github.com/ServiceComb/service-center/pkg/util"
	"github.com/ServiceComb/service-center/server/core/proto"
	serviceUtil "github.com/ServiceComb/service-center/server/service/util"
	"golang.org/x/net/context"
	"testing"
)

func TestRuleErr(t *testing.T) {
	e1 := serviceUtil.NotAllowAcrossDimensionError("a")
	if e1.Error() != "a" {
		fmt.Printf("NotAllowAcrossAppError failed")
		t.FailNow()
	}
	e2 := serviceUtil.NotMatchWhiteListError("a")
	if e2.Error() != "a" {
		fmt.Printf("NotAllowAcrossAppError failed")
		t.FailNow()
	}
	e3 := serviceUtil.MatchBlackListError("a")
	if e3.Error() != "a" {
		fmt.Printf("NotAllowAcrossAppError failed")
		t.FailNow()
	}
}

func TestRuleFilter_Filter(t *testing.T) {
	rf := serviceUtil.RuleFilter{
		DomainProject: "",
		Provider:      &proto.MicroService{},
		ProviderRules: []*proto.ServiceRule{},
	}
	_, err := rf.Filter(context.Background(), "")
	if err != nil {
		fmt.Printf("RuleFilter Filter failed")
		t.FailNow()
	}
}

func TestGetRulesUtil(t *testing.T) {
	_, err := serviceUtil.GetRulesUtil(util.SetContext(context.Background(), "cacheOnly", "1"), "", "")
	if err != nil {
		fmt.Printf("GetRulesUtil WithCacheOnly failed")
		t.FailNow()
	}

	_, err = serviceUtil.GetRulesUtil(context.Background(), "", "")
	if err == nil {
		fmt.Printf("GetRulesUtil failed")
		t.FailNow()
	}

	_, err = serviceUtil.GetOneRule(util.SetContext(context.Background(), "cacheOnly", "1"), "", "", "")
	if err != nil {
		fmt.Printf("GetOneRule WithCacheOnly failed")
		t.FailNow()
	}

	_, err = serviceUtil.GetOneRule(context.Background(), "", "", "")
	if err == nil {
		fmt.Printf("GetOneRule failed")
		t.FailNow()
	}
}

func TestRuleExist(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("TestRuleExist panic")
			t.FailNow()
		}
	}()
	serviceUtil.RuleExist(util.SetContext(context.Background(), "cacheOnly", "1"), "", "", "", "")
	serviceUtil.RuleExist(context.Background(), "", "", "", "")
}

func TestGetServiceRuleType(t *testing.T) {
	_, _, err := serviceUtil.GetServiceRuleType(util.SetContext(context.Background(), "cacheOnly", "1"), "", "")
	if err != nil {
		fmt.Printf("GetServiceRuleType WithCacheOnly failed")
		t.FailNow()
	}

	_, _, err = serviceUtil.GetServiceRuleType(context.Background(), "", "")
	if err == nil {
		fmt.Printf("GetServiceRuleType failed")
		t.FailNow()
	}
}

func TestAllowAcrossApp(t *testing.T) {
	err := serviceUtil.AllowAcrossDimension(&proto.MicroService{
		AppId: "a",
	}, &proto.MicroService{
		AppId: "a",
	})
	if err != nil {
		fmt.Printf("AllowAcrossApp with the same appId and no property failed")
		t.FailNow()
	}

	err = serviceUtil.AllowAcrossDimension(&proto.MicroService{
		AppId: "a",
	}, &proto.MicroService{
		AppId: "c",
	})
	if err == nil {
		fmt.Printf("AllowAcrossApp with the diff appId and no property failed")
		t.FailNow()
	}

	err = serviceUtil.AllowAcrossDimension(&proto.MicroService{
		AppId: "a",
		Properties: map[string]string{
			proto.PROP_ALLOW_CROSS_APP: "true",
		},
	}, &proto.MicroService{
		AppId: "a",
	})
	if err != nil {
		fmt.Printf("AllowAcrossApp with the same appId and allow property failed")
		t.FailNow()
	}

	err = serviceUtil.AllowAcrossDimension(&proto.MicroService{
		AppId: "a",
		Properties: map[string]string{
			proto.PROP_ALLOW_CROSS_APP: "true",
		},
	}, &proto.MicroService{
		AppId: "b",
	})
	if err != nil {
		fmt.Printf("AllowAcrossApp with the diff appId and allow property failed")
		t.FailNow()
	}

	err = serviceUtil.AllowAcrossDimension(&proto.MicroService{
		AppId: "a",
		Properties: map[string]string{
			proto.PROP_ALLOW_CROSS_APP: "false",
		},
	}, &proto.MicroService{
		AppId: "b",
	})
	if err == nil {
		fmt.Printf("AllowAcrossApp with the diff appId and deny property failed")
		t.FailNow()
	}

	err = serviceUtil.AllowAcrossDimension(&proto.MicroService{
		AppId: "a",
		Properties: map[string]string{
			proto.PROP_ALLOW_CROSS_APP: "",
		},
	}, &proto.MicroService{
		AppId: "b",
	})
	if err == nil {
		fmt.Printf("AllowAcrossApp with the diff appId and empty property failed")
		t.FailNow()
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
	if err != nil {
		fmt.Printf("MatchRules nil failed")
		t.FailNow()
	}

	err = serviceUtil.MatchRules([]*proto.ServiceRule{
		{
			RuleType:  "WHITE",
			Attribute: "",
			Pattern:   "",
		},
	}, &proto.MicroService{}, nil)
	if err == nil {
		fmt.Printf("MatchRules invalid WHITE failed")
		t.FailNow()
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
		fmt.Printf("MatchRules WHITE with field ServiceName failed")
		t.FailNow()
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
		fmt.Printf("MatchRules WHITE with tag b failed")
		t.FailNow()
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
		fmt.Printf("MatchRules WHITE with tag c failed")
		t.FailNow()
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
		fmt.Printf("MatchRules BLACK with tag b failed")
		t.FailNow()
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
		fmt.Printf("MatchRules BLACK with field ServiceName failed")
		t.FailNow()
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
		fmt.Printf("MatchRules BLACK with tag c failed")
		t.FailNow()
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
		fmt.Printf("MatchRules with not exist tag failed")
		t.FailNow()
	}
}

func TestGetConsumer(t *testing.T) {
	_, _, err := serviceUtil.GetConsumerIds(context.Background(), "", &proto.MicroService{})
	if err == nil {
		fmt.Printf("GetConsumerIds invalid failed")
		t.FailNow()
	}

	_, _, err = serviceUtil.GetConsumerIds(context.Background(), "", &proto.MicroService{
		ServiceId: "a",
	})
	if err == nil {
		fmt.Printf("GetConsumerIds not exist service failed")
		t.FailNow()
	}

	_, err = serviceUtil.GetConsumersInCache(util.SetContext(context.Background(), "cacheOnly", "1"), "", "",
		&proto.MicroService{
			ServiceId: "a",
		})
	if err != nil {
		fmt.Printf("GetConsumersInCache WithCacheOnly failed")
		t.FailNow()
	}
}

func TestGetProvider(t *testing.T) {
	_, err := serviceUtil.GetProvidersInCache(util.SetContext(context.Background(), "cacheOnly", "1"), "", "",
		&proto.MicroService{
			ServiceId: "a",
		})
	if err != nil {
		fmt.Printf("GetProvidersInCache WithCacheOnly failed")
		t.FailNow()
	}

	_, _, err = serviceUtil.GetProviderIdsByConsumerId(context.Background(), "", "", &proto.MicroService{})
	if err == nil {
		fmt.Printf("GetProviderIdsByConsumerId invalid failed")
		t.FailNow()
	}

	_, _, err = serviceUtil.GetProviderIdsByConsumerId(util.SetContext(context.Background(), "cacheOnly", "1"),
		"", "", &proto.MicroService{})
	if err != nil {
		fmt.Printf("GetProviderIdsByConsumerId WithCacheOnly failed")
		t.FailNow()
	}
}
