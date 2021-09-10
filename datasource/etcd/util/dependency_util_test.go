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
	"testing"

	. "github.com/apache/servicecomb-service-center/datasource/etcd/util"
	"github.com/go-chassis/cari/discovery"
)

func TestDeleteDependencyForService(t *testing.T) {
	_, err := DeleteDependencyForDeleteService("", "", &discovery.MicroServiceKey{})
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
	b := EqualServiceDependency(&discovery.MicroServiceKey{}, &discovery.MicroServiceKey{})
	if !b {
		t.Fatalf(`EqualServiceDependency failed`)
	}

	b = EqualServiceDependency(&discovery.MicroServiceKey{
		AppId: "a",
	}, &discovery.MicroServiceKey{
		AppId: "b",
	})
	if b {
		t.Fatalf(`EqualServiceDependency failed`)
	}
}

func TestAddDependencyRule(t *testing.T) {
	err := AddDependencyRule(context.Background(), &Dependency{
		Consumer: &discovery.MicroServiceKey{},
	})
	if err != nil {
		t.Fatalf(`AddDependencyRule failed`)
	}

	err = AddServiceVersionRule(context.Background(), "", &discovery.MicroService{}, &discovery.MicroServiceKey{})
	if err != nil {
		t.Fatalf(`AddServiceVersionRule failed`)
	}

	b, err := ContainServiceDependency([]*discovery.MicroServiceKey{
		{AppId: "a"},
	}, &discovery.MicroServiceKey{
		AppId: "b",
	})
	if b {
		t.Fatalf(`ContainServiceDependency contain failed`)
	}

	b, err = ContainServiceDependency([]*discovery.MicroServiceKey{
		{AppId: "a"},
	}, &discovery.MicroServiceKey{
		AppId: "a",
	})
	if !b {
		t.Fatalf(`ContainServiceDependency not contain failed`)
	}

	_, err = ContainServiceDependency(nil, nil)
	if err == nil {
		t.Fatalf(`ContainServiceDependency invalid failed`)
	}

	ok := DiffServiceVersion(&discovery.MicroServiceKey{
		AppId:       "a",
		ServiceName: "a",
		Version:     "1",
	}, &discovery.MicroServiceKey{
		AppId:       "a",
		ServiceName: "a",
		Version:     "2",
	})
	if !ok {
		t.Fatalf(`DiffServiceVersion failed`)
	}
}

func TestDependencyRuleExistUtil(t *testing.T) {
	_, err := DependencyRuleExistWithKey(context.Background(), "", &discovery.MicroServiceKey{})
	if err == nil {
		t.Fatalf(`DependencyRuleExistWithKey failed`)
	}
}

func TestServiceDependencyRuleExist(t *testing.T) {
	_, err := DependencyRuleExist(context.Background(), &discovery.MicroServiceKey{}, &discovery.MicroServiceKey{})
	if err != nil {
		t.Fatalf(`ServiceDependencyRuleExist failed`)
	}
}

func TestUpdateServiceForAddDependency(t *testing.T) {
	old := IsNeedUpdate([]*discovery.MicroServiceKey{
		{
			AppId:       "a",
			ServiceName: "a",
			Version:     "1",
		},
	}, &discovery.MicroServiceKey{
		AppId:       "a",
		ServiceName: "a",
		Version:     "2",
	})
	if old == nil {
		t.Fatalf(`IsNeedUpdate failed`)
	}
}

func TestCleanUpDependencyRules(t *testing.T) {
	err := CleanUpDependencyRules(context.Background(), "")
	if err == nil {
		t.Fatalf(`DependencyRelation_CleanUpDependencyRules failed`)
	}

	err = CleanUpDependencyRules(context.Background(), "a/b")
	if err != nil {
		t.Fatalf(`DependencyRelation_CleanUpDependencyRules failed`)
	}

	_, err = RemoveProviderRuleKeys(context.Background(), "a/b", nil)
	if err != nil {
		t.Fatalf(`DependencyRelation_removeProviderRuleKeys failed`)
	}
}
