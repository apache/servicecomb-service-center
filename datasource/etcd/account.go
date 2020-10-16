// Licensed to the Apache Software Foundation (ASF) under one or more
// contributor license agreements.  See the NOTICE file distributed with
// this work for additional information regarding copyright ownership.
// The ASF licenses this file to You under the Apache License, Version 2.0
// (the "License"); you may not use this file except in compliance with
// the License.  You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package etcd

import (
	"context"
	"encoding/json"
	utils "github.com/apache/servicecomb-service-center/datasource/etcd/util"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/rbacframe"
	"github.com/apache/servicecomb-service-center/server/core/backend"
	"github.com/apache/servicecomb-service-center/server/plugin/registry"
)

func (ds *DataSource) AccountExist(ctx context.Context, key string) (bool, error) {
	resp, err := kv.Registry().Do(ctx, registry.GET,
		registry.WithStrKey(GenerateETCDAccountKey(key)))
	if err != nil {
		return false, err
	}
	if resp.Count == 0 {
		return false, nil
	}
	return true, nil
}

func (ds *DataSource) GetAccount(ctx context.Context, key string) (*rbacframe.Account, error) {
	resp, err := kv.Registry().Do(ctx, registry.GET,
		registry.WithStrKey(GenerateETCDAccountKey(key)))
	if err != nil {
		return nil, err
	}
	if resp.Count != 1 {
		return nil, ErrNotUnique
	}
	account := &rbacframe.Account{}
	err = json.Unmarshal(resp.Kvs[0].Value, account)
	if err != nil {
		log.Errorf(err, "account info format invalid")
		return nil, err
	}
	return account, nil
}
func (ds *DataSource) ListAccount(ctx context.Context, key string) ([]*rbacframe.Account, int64, error) {
	resp, err := kv.Registry().Do(ctx, registry.GET,
		registry.WithStrKey(GenerateETCDAccountKey(key)), registry.WithPrefix())
	if err != nil {
		return nil, 0, err
	}
	accounts := make([]*rbacframe.Account, 0, resp.Count)
	for _, v := range resp.Kvs {
		a := &rbacframe.Account{}
		err = json.Unmarshal(v.Value, a)
		if err != nil {
			log.Error("account info format invalid:", err)
			continue //do not fail if some account is invalid
		}
		a.Password = ""
		accounts = append(accounts, a)
	}
	return accounts, resp.Count, nil
}
func (ds *DataSource) DeleteAccount(ctx context.Context, key string) (bool, error) {
	resp, err := kv.Registry().Do(ctx, registry.DEL,
		registry.WithStrKey(GenerateETCDAccountKey(key)))
	if err != nil {
		return false, err
	}
	return resp.Count != 0, nil
}
func (ds *DataSource) UpdateAccount(ctx context.Context, key string, account *rbacframe.Account) error {
	value, err := json.Marshal(account)
	if err != nil {
		log.Errorf(err, "account info is invalid")
		return err
	}
	_, err = kv.Registry().Do(ctx, registry.PUT,
		registry.WithStrKey(GenerateETCDAccountKey(key)),
		registry.WithValue(value))
	return err
}

func (ds *DataSource) AddDomain(ctx context.Context, domain string) (bool, error) {
	ok, err := kv.Registry().PutNoOverride(ctx,
		registry.WithStrKey(GenerateETCDDomainKey(domain)))
	if err != nil {
		return false, err
	}
	return ok, nil
}

func (ds *DataSource) DomainExist(ctx context.Context, domain string) (bool, error) {
	opts := append(utils.FromContext(ctx),
		registry.WithStrKey(GenerateETCDDomainKey(domain)),
		registry.WithCountOnly())
	rsp, err := kv.Store().Domain().Search(ctx, opts...)
	if err != nil {
		return false, err
	}
	return rsp.Count > 0, nil
}

func (ds *DataSource) AddProject(ctx context.Context, domain, project string) (bool, error) {
	ok, err := kv.Registry().PutNoOverride(ctx,
		registry.WithStrKey(GenerateETCDProjectKey(domain, project)))
	if err != nil {
		return ok, err
	}
	return ok, nil
}

func (ds *DataSource) ProjectExist(ctx context.Context, domain, project string) (bool, error) {
	opts := append(utils.FromContext(ctx),
		registry.WithStrKey(GenerateETCDProjectKey(domain, project)),
		registry.WithCountOnly())
	rsp, err := kv.Store().Project().Search(ctx, opts...)
	if err != nil {
		return false, err
	}
	return rsp.Count > 0, nil
}
