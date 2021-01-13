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
	"context"
	"errors"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"

	"github.com/apache/servicecomb-service-center/datasource"
	. "github.com/apache/servicecomb-service-center/datasource/etcd/util"
	"github.com/apache/servicecomb-service-center/pkg/util"
	"github.com/go-chassis/cari/discovery"
)

func TestRuleFilter_Filter(t *testing.T) {
	rf := RuleFilter{
		DomainProject: "",
		ProviderRules: []*discovery.ServiceRule{},
	}
	var err error
	t.Run("when there is no such a customer in db", func(t *testing.T) {
		_, err = rf.Filter(context.Background(), "")
		if err != nil && !errors.Is(err, datasource.ErrNoData) {
			t.Fatalf("RuleFilter Filter failed")
		}
		assert.Equal(t, datasource.ErrNoData, err, "no data found")
	})
	t.Run("FilterAll when customer not exist", func(t *testing.T) {
		_, _, err = rf.FilterAll(context.Background(), []string{""})
		if err != nil && !errors.Is(err, datasource.ErrNoData) {
			t.Fatalf("RuleFilter FilterAll failed")
		}
		assert.Equal(t, nil, err, "no customer found err is nil")
	})
	rf.ProviderRules = []*discovery.ServiceRule{
		{},
	}
	t.Run("FilterAll when ProviderRules not nil and service not exist", func(t *testing.T) {
		_, _, err = rf.FilterAll(context.Background(), []string{""})
		if err != nil && !errors.Is(err, datasource.ErrNoData) {
			t.Fatalf("RuleFilter FilterAll failed")
		}
		assert.Equal(t, datasource.ErrNoData, err, "no customer found when FilterAll")
	})
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
	RuleExist(util.WithCacheOnly(context.Background()), "", "", "", "")
	RuleExist(context.Background(), "", "", "", "")
}

func TestGetServiceRuleType(t *testing.T) {
	_, _, err := GetServiceRuleType(context.Background(), "", "")
	if err != nil {
		t.Fatalf("GetServiceRuleType failed")
	}
}

func TestAllowAcrossApp(t *testing.T) {
	err := AllowAcrossDimension(context.Background(), &discovery.MicroService{
		AppId: "a",
	}, &discovery.MicroService{
		AppId: "a",
	})
	if err != nil {
		t.Fatalf("AllowAcrossApp with the same appId and no property failed")
	}

	err = AllowAcrossDimension(context.Background(), &discovery.MicroService{
		AppId: "a",
	}, &discovery.MicroService{
		AppId: "c",
	})
	if err == nil {
		t.Fatalf("AllowAcrossApp with the diff appId and no property failed")
	}

	err = AllowAcrossDimension(context.Background(), &discovery.MicroService{
		AppId: "a",
		Properties: map[string]string{
			discovery.PropAllowCrossApp: "true",
		},
	}, &discovery.MicroService{
		AppId: "a",
	})
	if err != nil {
		t.Fatalf("AllowAcrossApp with the same appId and allow property failed")
	}

	err = AllowAcrossDimension(context.Background(), &discovery.MicroService{
		AppId: "a",
		Properties: map[string]string{
			discovery.PropAllowCrossApp: "true",
		},
	}, &discovery.MicroService{
		AppId: "b",
	})
	if err != nil {
		t.Fatalf("AllowAcrossApp with the diff appId and allow property failed")
	}

	err = AllowAcrossDimension(context.Background(), &discovery.MicroService{
		AppId: "a",
		Properties: map[string]string{
			discovery.PropAllowCrossApp: "false",
		},
	}, &discovery.MicroService{
		AppId: "b",
	})
	if err == nil {
		t.Fatalf("AllowAcrossApp with the diff appId and deny property failed")
	}

	err = AllowAcrossDimension(context.Background(), &discovery.MicroService{
		AppId: "a",
		Properties: map[string]string{
			discovery.PropAllowCrossApp: "",
		},
	}, &discovery.MicroService{
		AppId: "b",
	})
	if err == nil {
		t.Fatalf("AllowAcrossApp with the diff appId and empty property failed")
	}
}

func TestMatchRules(t *testing.T) {
	err := MatchRules([]*discovery.ServiceRule{
		{
			RuleType:  "WHITE",
			Attribute: "",
			Pattern:   "",
		},
	}, nil, nil)
	if err == nil {
		t.Fatalf("MatchRules nil failed")
	}

	err = MatchRules([]*discovery.ServiceRule{
		{
			RuleType:  "WHITE",
			Attribute: "",
			Pattern:   "",
		},
	}, &discovery.MicroService{}, nil)
	if err == nil {
		t.Fatalf("MatchRules invalid WHITE failed")
	}

	err = MatchRules([]*discovery.ServiceRule{
		{
			RuleType:  "WHITE",
			Attribute: "ServiceName",
			Pattern:   "^a.*",
		},
	}, &discovery.MicroService{
		ServiceName: "a",
	}, nil)
	if err != nil {
		t.Fatalf("MatchRules WHITE with field ServiceName failed")
	}

	err = MatchRules([]*discovery.ServiceRule{
		{
			RuleType:  "WHITE",
			Attribute: "tag_a",
			Pattern:   "^b.*",
		},
	}, &discovery.MicroService{}, map[string]string{
		"a": "b",
	})
	if err != nil {
		t.Fatalf("MatchRules WHITE with tag b failed")
	}

	err = MatchRules([]*discovery.ServiceRule{
		{
			RuleType:  "WHITE",
			Attribute: "tag_a",
			Pattern:   "^b.*",
		},
	}, &discovery.MicroService{}, map[string]string{
		"a": "c",
	})
	if err == nil {
		t.Fatalf("MatchRules WHITE with tag c failed")
	}

	err = MatchRules([]*discovery.ServiceRule{
		{
			RuleType:  "BLACK",
			Attribute: "tag_a",
			Pattern:   "^b.*",
		},
	}, &discovery.MicroService{}, map[string]string{
		"a": "b",
	})
	if err == nil {
		t.Fatalf("MatchRules BLACK with tag b failed")
	}

	err = MatchRules([]*discovery.ServiceRule{
		{
			RuleType:  "BLACK",
			Attribute: "ServiceName",
			Pattern:   "^a.*",
		},
	}, &discovery.MicroService{
		ServiceName: "a",
	}, nil)
	if err == nil {
		t.Fatalf("MatchRules BLACK with field ServiceName failed")
	}

	err = MatchRules([]*discovery.ServiceRule{
		{
			RuleType:  "BLACK",
			Attribute: "tag_a",
			Pattern:   "^b.*",
		},
	}, &discovery.MicroService{}, map[string]string{
		"a": "c",
	})
	if err != nil {
		t.Fatalf("MatchRules BLACK with tag c failed")
	}

	err = MatchRules([]*discovery.ServiceRule{
		{
			RuleType:  "BLACK",
			Attribute: "tag_a",
			Pattern:   "^b.*",
		},
	}, &discovery.MicroService{}, map[string]string{
		"b": "b",
	})
	if err != nil {
		t.Fatalf("MatchRules with not exist tag failed")
	}
}

func TestGetConsumer(t *testing.T) {
	_, _, err := GetAllProviderIds(context.Background(), "", &discovery.MicroService{})
	if err != nil {
		t.Fatalf("GetConsumerIdsByProvider invalid failed")
	}

	_, _, err = GetAllConsumerIds(context.Background(), "", &discovery.MicroService{
		ServiceId: "a",
	})
	if err != nil {
		t.Fatalf("GetConsumerIdsByProvider WithCacheOnly not exist service failed")
	}

	_, err = GetConsumerIds(context.Background(), "",
		&discovery.MicroService{
			ServiceId: "a",
		})
	if err != nil {
		t.Fatalf("GetConsumerIds WithCacheOnly failed")
	}
}

func TestGetProvider(t *testing.T) {
	_, err := GetProviderIds(context.Background(), "",
		&discovery.MicroService{
			ServiceId: "a",
		})
	if err != nil {
		t.Fatalf("GetProviderIds WithCacheOnly failed")
	}

	_, _, err = GetAllProviderIds(context.Background(), "", &discovery.MicroService{})
	if err != nil {
		t.Fatalf("GetAllProviderIds WithCacheOnly failed")
	}
}

func TestAccessible(t *testing.T) {
	err := Accessible(context.Background(), "xxx", "")
	if err.StatusCode() != http.StatusBadRequest {
		t.Fatalf("Accessible invalid failed")
	}
}
