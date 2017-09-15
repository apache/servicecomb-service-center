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
	apt "github.com/ServiceComb/service-center/server/core"
	"github.com/ServiceComb/service-center/server/core/registry"
	"github.com/ServiceComb/service-center/server/core/registry/store"
	"github.com/ServiceComb/service-center/util"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"golang.org/x/net/context"
	"strings"
)

func GetAllTenantRawData(ctx context.Context, opts ...registry.PluginOpOption) ([]*mvccpb.KeyValue, error) {
	opts = append(opts,
		registry.WithStrKey(apt.GenerateDomainKey("")),
		registry.WithPrefix())
	rsp, err := store.Store().Domain().Search(ctx, opts...)
	if err != nil {
		return nil, err
	}
	return rsp.Kvs, nil

}

func GetAllTenant(ctx context.Context, opts ...registry.PluginOpOption) ([]string, error) {
	insWatherByTenantKeys := []string{}
	kvs, err := GetAllTenantRawData(ctx, opts...)
	if err != nil {
		return nil, err
	}

	if len(kvs) == 0 {
		return insWatherByTenantKeys, err
	}

	tenant := ""
	instByTenant := ""
	arrTmp := []string{}
	for _, kv := range kvs {
		arrTmp = strings.Split(util.BytesToStringWithNoCopy(kv.Key), "/")
		tenant = arrTmp[len(arrTmp)-1]
		instByTenant = apt.GetInstanceRootKey(tenant)
		insWatherByTenantKeys = append(insWatherByTenantKeys, instByTenant)
	}
	return insWatherByTenantKeys, err
}

func DomainExist(ctx context.Context, domain string, opts ...registry.PluginOpOption) (bool, error) {
	opts = append(opts,
		registry.WithStrKey(apt.GenerateDomainKey(domain)),
		registry.WithCountOnly())
	rsp, err := store.Store().Domain().Search(ctx, opts...)
	if err != nil {
		return false, err
	}
	return rsp.Count > 0, nil
}

func NewDomain(ctx context.Context, tenant string) error {
	_, err := registry.GetRegisterCenter().Do(ctx,
		registry.PUT,
		registry.WithStrKey(apt.GenerateDomainKey(tenant)))
	if err != nil {
		return err
	}
	return nil
}
