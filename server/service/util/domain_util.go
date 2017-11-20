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
	"github.com/ServiceComb/service-center/pkg/util"
	apt "github.com/ServiceComb/service-center/server/core"
	"github.com/ServiceComb/service-center/server/core/backend"
	"github.com/ServiceComb/service-center/server/core/backend/store"
	"github.com/ServiceComb/service-center/server/infra/registry"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"golang.org/x/net/context"
	"strings"
)

func GetAllDomainRawData(ctx context.Context, opts ...registry.PluginOpOption) ([]*mvccpb.KeyValue, error) {
	opts = append(opts,
		registry.WithStrKey(apt.GenerateDomainKey("")),
		registry.WithPrefix())
	rsp, err := store.Store().Domain().Search(ctx, opts...)
	if err != nil {
		return nil, err
	}
	return rsp.Kvs, nil

}

func GetAllDomain(ctx context.Context, opts ...registry.PluginOpOption) ([]string, error) {
	insWatherByDomainKeys := []string{}
	kvs, err := GetAllDomainRawData(ctx, opts...)
	if err != nil {
		return nil, err
	}

	if len(kvs) == 0 {
		return insWatherByDomainKeys, err
	}

	domain := ""
	instByDomain := ""
	arrTmp := []string{}
	for _, kv := range kvs {
		arrTmp = strings.Split(util.BytesToStringWithNoCopy(kv.Key), "/")
		domain = arrTmp[len(arrTmp)-1]
		instByDomain = apt.GetInstanceRootKey(domain)
		insWatherByDomainKeys = append(insWatherByDomainKeys, instByDomain)
	}
	return insWatherByDomainKeys, err
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

func ProjectExist(ctx context.Context, domain, project string, opts ...registry.PluginOpOption) (bool, error) {
	opts = append(opts,
		registry.WithStrKey(apt.GenerateProjectKey(domain, project)),
		registry.WithCountOnly())
	rsp, err := store.Store().Project().Search(ctx, opts...)
	if err != nil {
		return false, err
	}
	return rsp.Count > 0, nil
}

func NewDomain(ctx context.Context, domain string) error {
	_, err := backend.Registry().PutNoOverride(ctx,
		registry.WithStrKey(apt.GenerateDomainKey(domain)))
	if err != nil {
		return err
	}
	return nil
}

func NewProject(ctx context.Context, domain, project string) error {
	_, err := backend.Registry().PutNoOverride(ctx,
		registry.WithStrKey(apt.GenerateProjectKey(domain, project)))
	if err != nil {
		return err
	}
	return nil
}

func NewDomainProject(ctx context.Context, domain, project string) error {
	ok, err := DomainExist(ctx, domain, registry.WithCacheOnly())
	if !ok && err == nil {
		err = NewDomain(ctx, domain)
	}
	if err != nil {
		return err
	}
	ok, err = ProjectExist(ctx, domain, project, registry.WithCacheOnly())
	if !ok && err == nil {
		err = NewProject(ctx, domain, project)
	}
	return err
}
