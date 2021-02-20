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
	"fmt"

	"github.com/apache/servicecomb-service-center/datasource"
	"github.com/apache/servicecomb-service-center/datasource/etcd/client"
	"github.com/apache/servicecomb-service-center/datasource/etcd/path"
	"github.com/apache/servicecomb-service-center/pkg/etcdsync"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/privacy"
	"github.com/apache/servicecomb-service-center/pkg/util"
	"github.com/go-chassis/cari/rbac"
)

func (ds *DataSource) CreateAccount(ctx context.Context, a *rbac.Account) error {
	lock, err := etcdsync.Lock("/account-creating/"+a.Name, -1, false)
	if err != nil {
		return fmt.Errorf("account %s is creating", a.Name)
	}
	defer func() {
		err := lock.Unlock()
		if err != nil {
			log.Errorf(err, "can not release account lock")
		}
	}()
	name := path.GenerateRBACAccountKey(a.Name)
	exist, err := datasource.Instance().AccountExist(ctx, a.Name)
	if err != nil {
		log.Errorf(err, "can not save account info")
		return err
	}
	if exist {
		return datasource.ErrAccountDuplicated
	}
	a.Password, err = privacy.HashPassword(a.Password)
	if err != nil {
		log.Error("pwd hash failed", err)
		return err
	}
	a.ID = util.GenerateUUID()
	value, err := json.Marshal(a)
	if err != nil {
		log.Errorf(err, "account info is invalid")
		return err
	}
	err = client.PutBytes(ctx, name, value)
	if err != nil {
		log.Errorf(err, "can not save account info")
		return err
	}
	log.Info("create new account: " + a.ID)
	return nil
}

func (ds *DataSource) AccountExist(ctx context.Context, name string) (bool, error) {
	resp, err := client.Instance().Do(ctx, client.GET,
		client.WithStrKey(path.GenerateRBACAccountKey(name)))
	if err != nil {
		return false, err
	}
	if resp.Count == 0 {
		return false, nil
	}
	return true, nil
}
func (ds *DataSource) GetAccount(ctx context.Context, name string) (*rbac.Account, error) {
	resp, err := client.Instance().Do(ctx, client.GET,
		client.WithStrKey(path.GenerateRBACAccountKey(name)))
	if err != nil {
		return nil, err
	}
	if resp.Count != 1 {
		return nil, client.ErrNotUnique
	}
	account := &rbac.Account{}
	err = json.Unmarshal(resp.Kvs[0].Value, account)
	if err != nil {
		log.Errorf(err, "account info format invalid")
		return nil, err
	}
	return account, nil
}
func (ds *DataSource) ListAccount(ctx context.Context) ([]*rbac.Account, int64, error) {
	resp, err := client.Instance().Do(ctx, client.GET,
		client.WithStrKey(path.GenerateRBACAccountKey("")), client.WithPrefix())
	if err != nil {
		return nil, 0, err
	}
	accounts := make([]*rbac.Account, 0, resp.Count)
	for _, v := range resp.Kvs {
		a := &rbac.Account{}
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
func (ds *DataSource) DeleteAccount(ctx context.Context, names []string) (bool, error) {
	if len(names) == 0 {
		return false, nil
	}
	resp, err := client.Instance().Do(ctx, client.DEL,
		client.WithStrKey(path.GenerateRBACAccountKey(names[0])))
	if err != nil {
		return false, err
	}
	return resp.Succeeded, nil
}
func (ds *DataSource) UpdateAccount(ctx context.Context, name string, account *rbac.Account) error {
	value, err := json.Marshal(account)
	if err != nil {
		log.Errorf(err, "account info is invalid")
		return err
	}
	_, err = client.Instance().Do(ctx, client.PUT,
		client.WithStrKey(path.GenerateRBACAccountKey(name)),
		client.WithValue(value))
	return err
}
