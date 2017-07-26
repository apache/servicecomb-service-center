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
package tenant

import (
	"context"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"github.com/ServiceComb/service-center/server/core"
	"github.com/ServiceComb/service-center/server/core/registry"
	"strings"
)

func GetAllTenantRawData() ([]*mvccpb.KeyValue, error) {
	opt := &registry.PluginOp{
		Key:        []byte(core.GenerateTenantKey("")),
		Action:     registry.GET,
		WithPrefix: true,
	}
	rsp, err := registry.GetRegisterCenter().Do(context.TODO(), opt)
	if err != nil {
		return nil, err
	}
	return rsp.Kvs, nil

}

func GetAllTenent() ([]string, error) {
	insWatherByTenantKeys := []string{}
	kvs, err := GetAllTenantRawData()
	if err != nil {
		return nil, err
	}
	if len(kvs) != 0 {
		tenant := ""
		instByTenant := ""
		arrTmp := []string{}
		for _, kv := range kvs {
			arrTmp = strings.Split(string(kv.Key), "/")
			tenant = arrTmp[len(arrTmp)-1]
			instByTenant = core.GetInstanceRootKey(tenant)
			insWatherByTenantKeys = append(insWatherByTenantKeys, instByTenant)
		}
	}
	return insWatherByTenantKeys, err
}
