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
package util

import (
	"fmt"
	"github.com/ServiceComb/service-center/pkg/util"
	"github.com/ServiceComb/service-center/server/core/proto"
	"golang.org/x/net/context"
	"testing"
)

func TestRefreshDependencyCache(t *testing.T) {
	err := RefreshDependencyCache(context.Background(), "", "", &proto.MicroService{})
	if err == nil {
		fmt.Printf(`RefreshDependencyCache failed`)
		t.FailNow()
	}
}

func TestDeleteDependencyForService(t *testing.T) {
	_, err := DeleteDependencyForService(context.Background(), &proto.MicroServiceKey{}, "")
	if err == nil {
		fmt.Printf(`DeleteDependencyForService failed`)
		t.FailNow()
	}

	err = deleteDependencyRuleUtil(context.Background(),
		&proto.MicroServiceDependency{
			Dependency: []*proto.MicroServiceKey{
				{AppId: "a"},
			},
		},
		&proto.MicroServiceKey{
			AppId: "a",
		}, "")
	if err == nil {
		fmt.Printf(`deleteDependencyRuleUtil with the same deps failed`)
		t.FailNow()
	}

	err = deleteDependencyRuleUtil(context.Background(),
		&proto.MicroServiceDependency{
			Dependency: []*proto.MicroServiceKey{
				{AppId: "b"},
			},
		},
		&proto.MicroServiceKey{
			AppId: "a",
		}, "")
	if err == nil {
		fmt.Printf(`deleteDependencyRuleUtil failed`)
		t.FailNow()
	}

	_, err = deleteDependencyUtil(context.Background(), "", "", "", map[string]bool{})
	if err == nil {
		fmt.Printf(`deleteDependencyUtil failed`)
		t.FailNow()
	}

	_, err = deleteConsumerDepOfProviderRule(context.Background(), "", &proto.MicroServiceKey{}, &proto.MicroServiceKey{})
	if err == nil {
		fmt.Printf(`deleteConsumerDepOfProviderRule failed`)
		t.FailNow()
	}

	_, err = deleteDepRuleUtil("", &proto.MicroServiceDependency{}, &proto.MicroServiceKey{})
	if err != nil {
		fmt.Printf(`deleteDepRuleUtil failed`)
		t.FailNow()
	}
}

func TestTransferToMicroServiceDependency(t *testing.T) {
	_, err := TransferToMicroServiceDependency(util.SetContext(context.Background(), "cacheOnly", "1"), "")
	if err != nil {
		fmt.Printf(`TransferToMicroServiceDependency WithCacheOnly failed`)
		t.FailNow()
	}

	_, err = TransferToMicroServiceDependency(context.Background(), "")
	if err == nil {
		fmt.Printf(`TransferToMicroServiceDependency failed`)
		t.FailNow()
	}
}

func TestEqualServiceDependency(t *testing.T) {
	b := equalServiceDependency(&proto.MicroServiceKey{}, &proto.MicroServiceKey{})
	if !b {
		fmt.Printf(`equalServiceDependency failed`)
		t.FailNow()
	}

	b = equalServiceDependency(&proto.MicroServiceKey{
		AppId: "a",
	}, &proto.MicroServiceKey{
		AppId: "b",
	})
	if b {
		fmt.Printf(`equalServiceDependency failed`)
		t.FailNow()
	}
}

func TestCreateDependencyRule(t *testing.T) {
	err := CreateDependencyRule(context.Background(), &Dependency{
		Consumer: &proto.MicroServiceKey{},
	})
	if err != nil {
		fmt.Printf(`CreateDependencyRule failed`)
		t.FailNow()
	}

	err = AddDependencyRule(context.Background(), &Dependency{
		Consumer: &proto.MicroServiceKey{},
	})
	if err != nil {
		fmt.Printf(`AddDependencyRule failed`)
		t.FailNow()
	}

	err = AddServiceVersionRule(context.Background(), "", &proto.MicroServiceKey{}, &proto.MicroServiceKey{}, "")
	if err == nil {
		fmt.Printf(`AddServiceVersionRule failed`)
		t.FailNow()
	}

	err = CreateDependencyRuleForFind(context.Background(), "", &proto.MicroServiceKey{}, &proto.MicroServiceKey{})
	if err == nil {
		fmt.Printf(`CreateDependencyRuleForFind failed`)
		t.FailNow()
	}

	_, err = addDepRuleUtil("", &proto.MicroServiceDependency{}, &proto.MicroServiceKey{})
	if err != nil {
		fmt.Printf(`addDepRuleUtil failed`)
		t.FailNow()
	}

	b, err := containServiceDependency([]*proto.MicroServiceKey{
		{AppId: "a"},
	}, &proto.MicroServiceKey{
		AppId: "b",
	})
	if b {
		fmt.Printf(`containServiceDependency contain failed`)
		t.FailNow()
	}

	b, err = containServiceDependency([]*proto.MicroServiceKey{
		{AppId: "a"},
	}, &proto.MicroServiceKey{
		AppId: "a",
	})
	if !b {
		fmt.Printf(`containServiceDependency not contain failed`)
		t.FailNow()
	}

	_, err = containServiceDependency(nil, nil)
	if err == nil {
		fmt.Printf(`containServiceDependency invalid failed`)
		t.FailNow()
	}

	ok := diffServiceVersion(&proto.MicroServiceKey{
		AppId:       "a",
		ServiceName: "a",
		Version:     "1",
	}, &proto.MicroServiceKey{
		AppId:       "a",
		ServiceName: "a",
		Version:     "2",
	})
	if !ok {
		fmt.Printf(`diffServiceVersion failed`)
		t.FailNow()
	}

	err = validateMicroServiceKey(&proto.MicroServiceKey{}, false)
	if err == nil {
		fmt.Printf(`validateMicroServiceKey false invalid failed`)
		t.FailNow()
	}

	err = validateMicroServiceKey(&proto.MicroServiceKey{}, true)
	if err == nil {
		fmt.Printf(`validateMicroServiceKey true invalid failed`)
		t.FailNow()
	}

	err = validateMicroServiceKey(&proto.MicroServiceKey{
		AppId:       "a",
		ServiceName: "b",
		Version:     "latest",
	}, true)
	if err != nil {
		fmt.Printf(`validateMicroServiceKey true failed`)
		t.FailNow()
	}

	err = validateMicroServiceKey(&proto.MicroServiceKey{
		AppId:       "a",
		ServiceName: "b",
		Version:     "1.0.0",
	}, false)
	if err != nil {
		fmt.Printf(`validateMicroServiceKey false failed`)
		t.FailNow()
	}

	ok = isDependencyAll(&proto.MicroServiceDependency{})
	if ok {
		fmt.Printf(`isDependencyAll not * failed`)
		t.FailNow()
	}

	ok = isDependencyAll(&proto.MicroServiceDependency{
		Dependency: []*proto.MicroServiceKey{
			{
				ServiceName: "*",
			},
		},
	})
	if !ok {
		fmt.Printf(`isDependencyAll * failed`)
		t.FailNow()
	}

	ok = isExist([]*proto.MicroServiceKey{
		{
			ServiceName: "*",
		},
	}, &proto.MicroServiceKey{
		ServiceName: "*",
	})
	if !ok {
		fmt.Printf(`isExist failed`)
		t.FailNow()
	}
}

func TestBadParamsResponse(t *testing.T) {
	p := BadParamsResponse("a")
	if p == nil {
		fmt.Printf(`BadParamsResponse failed`)
		t.FailNow()
	}
}

func TestParamsChecker(t *testing.T) {
	p := ParamsChecker(nil, nil)
	if p == nil || p.Response.Code == proto.Response_SUCCESS {
		fmt.Printf(`ParamsChecker invalid failed`)
		t.FailNow()
	}

	p = ParamsChecker(&proto.MicroServiceKey{
		AppId:       "a",
		ServiceName: "b",
		Version:     "1.0.0",
	}, nil)
	if p == nil || p.Response.Code == proto.Response_SUCCESS {
		fmt.Printf(`ParamsChecker invalid failed`)
		t.FailNow()
	}

	p = ParamsChecker(&proto.MicroServiceKey{
		AppId:       "a",
		ServiceName: "b",
		Version:     "1.0.0",
	}, []*proto.MicroServiceKey{
		{ServiceName: "*"},
	})
	if p != nil {
		fmt.Printf(`ParamsChecker * failed`)
		t.FailNow()
	}

	p = ParamsChecker(&proto.MicroServiceKey{
		AppId:       "a",
		ServiceName: "b",
		Version:     "1.0.0",
	}, []*proto.MicroServiceKey{
		{},
	})
	if p == nil {
		fmt.Printf(`ParamsChecker invalid provider key failed`)
		t.FailNow()
	}

	p = ParamsChecker(&proto.MicroServiceKey{
		AppId:       "a",
		ServiceName: "b",
		Version:     "1.0.0",
	}, []*proto.MicroServiceKey{
		{ServiceName: "a", Version: "1"},
		{ServiceName: "a", Version: "1"},
	})
	if p == nil {
		fmt.Printf(`ParamsChecker duplicate provider key failed`)
		t.FailNow()
	}
}

func TestServiceDependencyRuleExist(t *testing.T) {
	_, err := ProviderDependencyRuleExist(util.SetContext(context.Background(), "cacheOnly", "1"), "", &proto.MicroServiceKey{}, &proto.MicroServiceKey{})
	if err != nil {
		fmt.Printf(`ServiceDependencyRuleExist WithCacheOnly failed`)
		t.FailNow()
	}

	_, err = ProviderDependencyRuleExist(context.Background(), "", &proto.MicroServiceKey{}, &proto.MicroServiceKey{})
	if err == nil {
		fmt.Printf(`ServiceDependencyRuleExist failed`)
		t.FailNow()
	}
}

func TestUpdateServiceForAddDependency(t *testing.T) {
	_, _, err := updateDepRuleUtil("", &proto.MicroServiceDependency{}, &proto.MicroServiceKey{})
	if err != nil {
		fmt.Printf(`updateDepRuleUtil failed`)
		t.FailNow()
	}

	old := isNeedUpdate([]*proto.MicroServiceKey{
		{
			AppId:       "a",
			ServiceName: "a",
			Version:     "1",
		},
	}, &proto.MicroServiceKey{
		AppId:       "a",
		ServiceName: "a",
		Version:     "2",
	})
	if old == nil {
		fmt.Printf(`isNeedUpdate failed`)
		t.FailNow()
	}
}

func TestFilter(t *testing.T) {
	_, _, err := getConsumerIdsWithFilter(context.Background(), "", "", &proto.MicroService{}, noFilter)
	if err == nil {
		fmt.Printf(`getConsumerIdsWithFilter failed`)
		t.FailNow()
	}

	_, _, err = filterConsumerIds(context.Background(), []string{}, noFilter)
	if err != nil {
		fmt.Printf(`filterConsumerIds invalid failed`)
		t.FailNow()
	}

	_, _, err = filterConsumerIds(context.Background(), []string{"a"}, noFilter)
	if err != nil {
		fmt.Printf(`filterConsumerIds invalid failed`)
		t.FailNow()
	}

	rf := RuleFilter{
		DomainProject: "",
		Provider:      &proto.MicroService{},
		ProviderRules: []*proto.ServiceRule{},
	}
	_, _, err = filterConsumerIds(context.Background(), []string{"a"}, rf.Filter)
	if err != nil {
		fmt.Printf(`filterConsumerIds invalid failed`)
		t.FailNow()
	}
}

func TestDependency(t *testing.T) {
	d := &Dependency{
		removedDependencyRuleList: []*proto.MicroServiceKey{
			{ServiceName: "a", Version: "1.0.0"},
		},
		NewDependencyRuleList: []*proto.MicroServiceKey{
			{ServiceName: "a", Version: "1.0.0"},
		},
	}
	d.RemoveConsumerOfProviderRule()
	d.AddConsumerOfProviderRule()
	err := d.UpdateProvidersRuleOfConsumer("")
	if err == nil {
		fmt.Printf(`Dependency_UpdateProvidersRuleOfConsumer failed`)
		t.FailNow()
	}

	dr := &DependencyRelation{
		provider: &proto.MicroService{},
		consumer: &proto.MicroService{},
		ctx:      util.SetContext(context.Background(), "cacheOnly", "1"),
	}
	_, err = dr.GetDependencyProviders()
	if err != nil {
		fmt.Printf(`DependencyRelation_GetDependencyProviders failed`)
		t.FailNow()
	}

	_, err = dr.getDependencyProviderIds([]*proto.MicroServiceKey{
		{ServiceName: "*"},
		{ServiceName: "a", Version: "1.0.0"},
		{ServiceName: "b", Version: "latest"},
	})
	if err != nil {
		fmt.Printf(`DependencyRelation_getDependencyProviderIds * WithCacheOnly failed`)
		t.FailNow()
	}
	_, err = dr.getDependencyProviderIds([]*proto.MicroServiceKey{
		{ServiceName: "a", Version: "1.0.0"},
		{ServiceName: "b", Version: "latest"},
	})
	if err != nil {
		fmt.Printf(`DependencyRelation_getDependencyProviderIds WithCacheOnly failed`)
		t.FailNow()
	}

	_, err = dr.GetDependencyConsumers()
	if err != nil {
		fmt.Printf(`DependencyRelation_GetDependencyConsumers WithCacheOnly failed`)
		t.FailNow()
	}

	_, err = dr.getServiceByMicroServiceKey("", &proto.MicroServiceKey{})
	if err != nil {
		fmt.Printf(`DependencyRelation_getServiceByMicroServiceKey WithCacheOnly failed`)
		t.FailNow()
	}

	_, err = dr.getConsumerOfSameServiceNameAndAppId(&proto.MicroServiceKey{})
	if err != nil {
		fmt.Printf(`DependencyRelation_getConsumerOfSameServiceNameAndAppId WithCacheOnly failed`)
		t.FailNow()
	}

	dr = &DependencyRelation{
		provider: &proto.MicroService{},
		consumer: &proto.MicroService{},
		ctx:      context.Background(),
	}
	_, err = dr.getDependencyProviderIds([]*proto.MicroServiceKey{
		{ServiceName: "*"},
	})
	if err == nil {
		fmt.Printf(`DependencyRelation_getDependencyProviderIds * failed`)
		t.FailNow()
	}
	_, err = dr.getDependencyProviderIds([]*proto.MicroServiceKey{
		{ServiceName: "a", Version: "1.0.0"},
		{ServiceName: "b", Version: "latest"},
	})
	if err == nil {
		fmt.Printf(`DependencyRelation_getDependencyProviderIds failed`)
		t.FailNow()
	}

	_, err = dr.GetDependencyConsumers()
	if err == nil {
		fmt.Printf(`DependencyRelation_GetDependencyConsumers failed`)
		t.FailNow()
	}

	_, err = dr.getServiceByMicroServiceKey("", &proto.MicroServiceKey{})
	if err == nil {
		fmt.Printf(`DependencyRelation_getServiceByMicroServiceKey failed`)
		t.FailNow()
	}

	_, err = dr.getConsumerOfSameServiceNameAndAppId(&proto.MicroServiceKey{})
	if err == nil {
		fmt.Printf(`DependencyRelation_getConsumerOfSameServiceNameAndAppId failed`)
		t.FailNow()
	}

	dr = &DependencyRelation{
		consumer: &proto.MicroService{},
		ctx:      context.Background(),
	}
	_, err = dr.getDependencyConsumersOfProvider()
	if err == nil {
		fmt.Printf(`DependencyRelation_getDependencyConsumersOfProvider failed`)
		t.FailNow()
	}
}
