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
	"github.com/ServiceComb/service-center/server/infra/registry"
	serviceUtil "github.com/ServiceComb/service-center/server/service/util"
	"golang.org/x/net/context"
	"testing"
)

func TestGetDomain(t *testing.T) {
	_, err := serviceUtil.GetAllDomainRawData(context.Background(), registry.WithCacheOnly())
	if err != nil {
		fmt.Printf("GetAllDomainRawData WithCacheOnly failed")
		t.FailNow()
	}

	_, err = serviceUtil.GetAllDomainRawData(context.Background())
	if err == nil {
		fmt.Printf("GetAllDomainRawData failed")
		t.FailNow()
	}

	_, err = serviceUtil.GetAllDomain(context.Background(), registry.WithCacheOnly())
	if err != nil {
		fmt.Printf("GetAllDomain WithCacheOnly failed")
		t.FailNow()
	}

	_, err = serviceUtil.GetAllDomain(context.Background())
	if err == nil {
		fmt.Printf("GetAllDomain failed")
		t.FailNow()
	}
}

func TestDomainExist(t *testing.T) {
	_, err := serviceUtil.DomainExist(context.Background(), "", registry.WithCacheOnly())
	if err != nil {
		fmt.Printf("DomainExist WithCacheOnly failed")
		t.FailNow()
	}

	_, err = serviceUtil.DomainExist(context.Background(), "")
	if err == nil {
		fmt.Printf("DomainExist failed")
		t.FailNow()
	}
}

func TestNewDomain(t *testing.T) {
	err := serviceUtil.NewDomain(context.Background(), "")
	if err == nil {
		fmt.Printf("NewDomain failed")
		t.FailNow()
	}
}

func TestProjectExist(t *testing.T) {
	_, err := serviceUtil.ProjectExist(context.Background(), "", "", registry.WithCacheOnly())
	if err != nil {
		fmt.Printf("DomainExist WithCacheOnly failed")
		t.FailNow()
	}

	_, err = serviceUtil.ProjectExist(context.Background(), "", "")
	if err == nil {
		fmt.Printf("DomainExist failed")
		t.FailNow()
	}
}

func TestNewProject(t *testing.T) {
	err := serviceUtil.NewProject(context.Background(), "", "")
	if err == nil {
		fmt.Printf("NewProject failed")
		t.FailNow()
	}
}

func TestNewDomainProject(t *testing.T) {
	err := serviceUtil.NewDomainProject(context.Background(), "", "")
	if err == nil {
		fmt.Printf("NewDomainProject failed")
		t.FailNow()
	}
}
