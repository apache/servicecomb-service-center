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
	"strconv"
	"time"

	"github.com/apache/servicecomb-service-center/datasource"
	"github.com/apache/servicecomb-service-center/datasource/etcd/path"
	"github.com/apache/servicecomb-service-center/pkg/etcdsync"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/privacy"
	"github.com/apache/servicecomb-service-center/pkg/util"
	"github.com/go-chassis/cari/rbac"
	"github.com/go-chassis/foundation/stringutil"
	"github.com/little-cui/etcdadpt"
)

type AccountManager struct {
}

func (ds *AccountManager) CreateAccount(ctx context.Context, a *rbac.Account) error {
	lock, err := etcdsync.Lock("/account-creating/"+a.Name, -1, false)
	if err != nil {
		return fmt.Errorf("account %s is creating", a.Name)
	}
	defer func() {
		err := lock.Unlock()
		if err != nil {
			log.Error("can not release account lock", err)
		}
	}()
	exist, err := ds.AccountExist(ctx, a.Name)
	if err != nil {
		log.Error("can not save account info", err)
		return err
	}
	if exist {
		return datasource.ErrAccountDuplicated
	}
	a.Password, err = privacy.ScryptPassword(a.Password)
	if err != nil {
		log.Error("pwd hash failed", err)
		return err
	}
	a.Role = ""
	a.CurrentPassword = ""
	a.ID = util.GenerateUUID()
	a.CreateTime = strconv.FormatInt(time.Now().Unix(), 10)
	a.UpdateTime = a.CreateTime
	opts, err := GenAccountOpts(a, etcdadpt.ActionPut)
	if err != nil {
		log.Error("", err)
		return err
	}
	err = etcdadpt.Txn(ctx, opts)
	if err != nil {
		log.Error("can not save account info", err)
		return err
	}
	log.Info("create new account: " + a.ID)
	return nil
}
func GenAccountOpts(a *rbac.Account, action etcdadpt.Action) ([]etcdadpt.OpOptions, error) {
	opts := make([]etcdadpt.OpOptions, 0)
	value, err := json.Marshal(a)
	if err != nil {
		log.Error("account info is invalid", err)
		return nil, err
	}
	opts = append(opts, etcdadpt.OpOptions{
		Key:    stringutil.Str2bytes(path.GenerateAccountKey(a.Name)),
		Value:  value,
		Action: action,
	})
	for _, r := range a.Roles {
		opt := etcdadpt.OpOptions{
			Key:    stringutil.Str2bytes(path.GenRoleAccountIdxKey(r, a.Name)),
			Action: action,
		}
		opts = append(opts, opt)
	}

	return opts, nil
}
func (ds *AccountManager) AccountExist(ctx context.Context, name string) (bool, error) {
	return etcdadpt.Exist(ctx, path.GenerateRBACAccountKey(name))
}
func (ds *AccountManager) GetAccount(ctx context.Context, name string) (*rbac.Account, error) {
	kv, err := etcdadpt.Get(ctx, path.GenerateRBACAccountKey(name))
	if err != nil {
		return nil, err
	}
	if kv == nil {
		return nil, datasource.ErrAccountNotExist
	}
	account := &rbac.Account{}
	err = json.Unmarshal(kv.Value, account)
	if err != nil {
		log.Error("account info format invalid", err)
		return nil, err
	}
	ds.compatibleOldVersionAccount(account)
	return account, nil
}

func (ds *AccountManager) compatibleOldVersionAccount(a *rbac.Account) {
	// old version use Role, now use Roles
	// Role/Roles will not exist at the same time
	if len(a.Role) == 0 {
		return
	}
	a.Roles = []string{a.Role}
	a.Role = ""
}

func (ds *AccountManager) ListAccount(ctx context.Context) ([]*rbac.Account, int64, error) {
	kvs, n, err := etcdadpt.List(ctx, path.GenerateRBACAccountKey(""))
	if err != nil {
		return nil, 0, err
	}
	accounts := make([]*rbac.Account, 0, n)
	for _, v := range kvs {
		a := &rbac.Account{}
		err = json.Unmarshal(v.Value, a)
		if err != nil {
			log.Error("account info format invalid:", err)
			continue //do not fail if some account is invalid
		}
		a.Password = ""
		ds.compatibleOldVersionAccount(a)
		accounts = append(accounts, a)
	}
	return accounts, n, nil
}
func (ds *AccountManager) DeleteAccount(ctx context.Context, names []string) (bool, error) {
	if len(names) == 0 {
		return false, nil
	}
	for _, name := range names {
		a, err := ds.GetAccount(ctx, name)
		if err != nil {
			log.Error("", err)
			continue //do not fail if some account is invalid
		}
		if a == nil {
			log.Warn("can not find account")
			continue
		}
		opts, err := GenAccountOpts(a, etcdadpt.ActionDelete)
		if err != nil {
			log.Error("", err)
			continue //do not fail if some account is invalid

		}
		err = etcdadpt.Txn(ctx, opts)
		if err != nil {
			log.Error(datasource.ErrDeleteAccountFailed.Error(), err)
			return false, err
		}
	}
	return true, nil
}
func (ds *AccountManager) UpdateAccount(ctx context.Context, name string, account *rbac.Account) error {
	var (
		opts []etcdadpt.OpOptions
		err  error
	)

	account.UpdateTime = strconv.FormatInt(time.Now().Unix(), 10)

	opts, err = GenAccountOpts(account, etcdadpt.ActionPut)
	if err != nil {
		log.Error("GenAccountOpts failed", err)
		return err
	}

	old, err := ds.GetAccount(ctx, account.Name)
	if err != nil {
		log.Error("GetAccount failed", err)
		return err
	}

	for _, r := range old.Roles {
		if hasRole(account, r) {
			continue
		}
		opt := etcdadpt.OpOptions{
			Key:    stringutil.Str2bytes(path.GenRoleAccountIdxKey(r, old.Name)),
			Action: etcdadpt.ActionDelete,
		}
		opts = append(opts, opt)
	}

	err = etcdadpt.Txn(ctx, opts)
	if err != nil {
		log.Error("BatchCommit failed", err)
	}
	return err
}

func hasRole(account *rbac.Account, r string) bool {
	for _, n := range account.Roles {
		if r == n {
			return true
		}
	}
	return false
}
