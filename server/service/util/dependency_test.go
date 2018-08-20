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
	"github.com/apache/incubator-servicecomb-service-center/server/core/proto"
	"golang.org/x/net/context"
	"testing"
)

func TestDeleteDependencyForService(t *testing.T) {
	_, err := DeleteDependencyForDeleteService("", "", &proto.MicroServiceKey{})
	if err != nil {
		t.Fatalf(`DeleteDependencyForDeleteService failed`)
	}
}

func TestTransferToMicroServiceDependency(t *testing.T) {
	_, err := TransferToMicroServiceDependency(context.Background(), "")
	if err == nil {
		t.Fatalf(`TransferToMicroServiceDependency failed`)
	}
}

func TestEqualServiceDependency(t *testing.T) {
	b := equalServiceDependency(&proto.MicroServiceKey{}, &proto.MicroServiceKey{})
	if !b {
		t.Fatalf(`equalServiceDependency failed`)
	}

	b = equalServiceDependency(&proto.MicroServiceKey{
		AppId: "a",
	}, &proto.MicroServiceKey{
		AppId: "b",
	})
	if b {
		t.Fatalf(`equalServiceDependency failed`)
	}
}

func TestCreateDependencyRule(t *testing.T) {
	err := CreateDependencyRule(context.Background(), &Dependency{
		Consumer: &proto.MicroServiceKey{},
	})
	if err != nil {
		t.Fatalf(`CreateDependencyRule failed`)
	}

	err = AddDependencyRule(context.Background(), &Dependency{
		Consumer: &proto.MicroServiceKey{},
	})
	if err != nil {
		t.Fatalf(`AddDependencyRule failed`)
	}

	err = AddServiceVersionRule(context.Background(), "", &proto.MicroService{}, &proto.MicroServiceKey{})
	if err == nil {
		t.Fatalf(`AddServiceVersionRule failed`)
	}

	b, err := containServiceDependency([]*proto.MicroServiceKey{
		{AppId: "a"},
	}, &proto.MicroServiceKey{
		AppId: "b",
	})
	if b {
		t.Fatalf(`containServiceDependency contain failed`)
	}

	b, err = containServiceDependency([]*proto.MicroServiceKey{
		{AppId: "a"},
	}, &proto.MicroServiceKey{
		AppId: "a",
	})
	if !b {
		t.Fatalf(`containServiceDependency not contain failed`)
	}

	_, err = containServiceDependency(nil, nil)
	if err == nil {
		t.Fatalf(`containServiceDependency invalid failed`)
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
		t.Fatalf(`diffServiceVersion failed`)
	}
}

func TestBadParamsResponse(t *testing.T) {
	p := BadParamsResponse("a")
	if p == nil {
		t.Fatalf(`BadParamsResponse failed`)
	}
}

func TestDependencyRuleExistUtil(t *testing.T) {
	_, err := dependencyRuleExistUtil(context.Background(), "", &proto.MicroServiceKey{})
	if err == nil {
		t.Fatalf(`dependencyRuleExistUtil failed`)
	}
}

func TestParamsChecker(t *testing.T) {
	p := ParamsChecker(nil, nil)
	if p != nil {
		t.Fatalf(`ParamsChecker invalid failed`)
	}

	p = ParamsChecker(&proto.MicroServiceKey{
		AppId:       "a",
		ServiceName: "b",
		Version:     "1.0.0",
	}, nil)
	if p != nil {
		t.Fatalf(`ParamsChecker invalid failed`)
	}

	p = ParamsChecker(&proto.MicroServiceKey{
		AppId:       "a",
		ServiceName: "b",
		Version:     "1.0.0",
	}, []*proto.MicroServiceKey{
		{ServiceName: "*"},
	})
	if p != nil {
		t.Fatalf(`ParamsChecker * failed`)
	}

	p = ParamsChecker(&proto.MicroServiceKey{
		AppId:       "a",
		ServiceName: "b",
		Version:     "1.0.0",
	}, []*proto.MicroServiceKey{
		{},
	})
	if p == nil {
		t.Fatalf(`ParamsChecker invalid provider key failed`)
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
		t.Fatalf(`ParamsChecker duplicate provider key failed`)
	}
}

func TestServiceDependencyRuleExist(t *testing.T) {
	_, err := DependencyRuleExist(context.Background(), &proto.MicroServiceKey{}, &proto.MicroServiceKey{})
	if err == nil {
		t.Fatalf(`ServiceDependencyRuleExist failed`)
	}
}

func TestUpdateServiceForAddDependency(t *testing.T) {
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
		t.Fatalf(`isNeedUpdate failed`)
	}
}

func TestDependency(t *testing.T) {
	d := &Dependency{
		removedDependencyRuleList: []*proto.MicroServiceKey{
			{ServiceName: "a", Version: "1.0.0"},
		},
		newDependencyRuleList: []*proto.MicroServiceKey{
			{ServiceName: "a", Version: "1.0.0"},
		},
	}
	d.RemoveConsumerOfProviderRule()
	d.AddConsumerOfProviderRule()
	err := d.UpdateProvidersRuleOfConsumer(context.Background(), "")
	if err == nil {
		t.Fatalf(`Dependency_UpdateProvidersRuleOfConsumer failed`)
	}

	dr := &DependencyRelation{
		provider: &proto.MicroService{},
		consumer: &proto.MicroService{},
		ctx:      context.Background(),
	}
	_, err = dr.getDependencyProviderIds([]*proto.MicroServiceKey{
		{ServiceName: "*"},
	})
	if err == nil {
		t.Fatalf(`DependencyRelation_getDependencyProviderIds * failed`)
	}
	_, err = dr.getDependencyProviderIds([]*proto.MicroServiceKey{
		{ServiceName: "a", Version: "1.0.0"},
		{ServiceName: "b", Version: "latest"},
	})
	if err == nil {
		t.Fatalf(`DependencyRelation_getDependencyProviderIds failed`)
	}

	_, err = dr.GetDependencyConsumers()
	if err == nil {
		t.Fatalf(`DependencyRelation_GetDependencyConsumers failed`)
	}

	_, err = dr.getServiceByMicroServiceKey(&proto.MicroServiceKey{})
	if err == nil {
		t.Fatalf(`DependencyRelation_getServiceByMicroServiceKey failed`)
	}

	_, err = dr.getConsumerOfSameServiceNameAndAppId(&proto.MicroServiceKey{})
	if err == nil {
		t.Fatalf(`DependencyRelation_getConsumerOfSameServiceNameAndAppId failed`)
	}

	dr = &DependencyRelation{
		consumer: &proto.MicroService{},
		ctx:      context.Background(),
	}
	_, err = dr.getDependencyConsumersOfProvider()
	if err == nil {
		t.Fatalf(`DependencyRelation_getDependencyConsumersOfProvider failed`)
	}
	_, err = dr.GetDependencyProviders()
	if err == nil {
		t.Fatalf(`DependencyRelation_GetDependencyProviders failed`)
	}

	err = CleanUpDependencyRules(context.Background(), "")
	if err == nil {
		t.Fatalf(`DependencyRelation_CleanUpDependencyRules failed`)
	}

	err = CleanUpDependencyRules(context.Background(), "a/b")
	if err == nil {
		t.Fatalf(`DependencyRelation_CleanUpDependencyRules failed`)
	}

	_, err = removeProviderRuleKeys(context.Background(), "a/b", nil)
	if err == nil {
		t.Fatalf(`DependencyRelation_removeProviderRuleKeys failed`)
	}
}

func TestDependencyRelationFilterOpt(t *testing.T) {
	op := toDependencyRelationFilterOpt(
		WithSameDomainProject(),
		WithoutSelfDependency(),
	)
	if !op.NonSelf || !op.SameDomainProject {
		t.Fatalf(`toDependencyRelationFilterOpt failed`)
	}
}

func TestGetConsumerIdsWithFilter(t *testing.T) {
	_, _, err := getConsumerIdsWithFilter(context.Background(), "", &proto.MicroService{}, nil)
	if err == nil {
		t.Fatalf(`TestGetConsumerIdsWithFilter failed`)
	}
}
