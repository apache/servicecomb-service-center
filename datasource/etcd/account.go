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

package etcd

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	crbac "github.com/go-chassis/cari/rbac"
	"github.com/go-chassis/foundation/stringutil"
	"github.com/little-cui/etcdadpt"

	"github.com/apache/servicecomb-service-center/datasource"
	"github.com/apache/servicecomb-service-center/datasource/etcd/path"
	esync "github.com/apache/servicecomb-service-center/datasource/etcd/sync"
	"github.com/apache/servicecomb-service-center/datasource/rbac"
	"github.com/apache/servicecomb-service-center/pkg/log"
)

func init() {
	rbac.Install("etcd", NewRbacDAO)
	rbac.Install("embeded_etcd", NewRbacDAO)
	rbac.Install("embedded_etcd", NewRbacDAO)
}

func NewRbacDAO(opts rbac.Options) (rbac.DAO, error) {
	return &RbacDAO{}, nil
}

type RbacDAO struct {
}

func (ds *RbacDAO) CreateAccount(ctx context.Context, a *crbac.Account) error {
	exist, err := ds.AccountExist(ctx, a.Name)
	if err != nil {
		log.Error("can not save account info", err)
		return err
	}
	if exist {
		return rbac.ErrAccountDuplicated
	}

	opts, err := GenAccountOpts(a, etcdadpt.ActionPut)
	if err != nil {
		log.Error("", err)
		return err
	}
	syncOpts, err := esync.GenCreateOpts(ctx, datasource.ResourceAccount, a)
	if err != nil {
		log.Error("fail to create sync opts", err)
		return err
	}
	opts = append(opts, syncOpts...)
	err = etcdadpt.Txn(ctx, opts)
	if err != nil {
		log.Error("can not save account info", err)
		return err
	}
	log.Info("create new account: " + a.ID)
	return nil
}

func GenAccountOpts(a *crbac.Account, action etcdadpt.Action) ([]etcdadpt.OpOptions, error) {
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

func (ds *RbacDAO) AccountExist(ctx context.Context, name string) (bool, error) {
	return etcdadpt.Exist(ctx, path.GenerateRBACAccountKey(name))
}
func (ds *RbacDAO) GetAccount(ctx context.Context, name string) (*crbac.Account, error) {
	kv, err := etcdadpt.Get(ctx, path.GenerateRBACAccountKey(name))
	if err != nil {
		return nil, err
	}
	if kv == nil {
		return nil, rbac.ErrAccountNotExist
	}
	account := &crbac.Account{}
	err = json.Unmarshal(kv.Value, account)
	if err != nil {
		log.Error("account info format invalid", err)
		return nil, err
	}
	ds.compatibleOldVersionAccount(account)
	return account, nil
}

func (ds *RbacDAO) compatibleOldVersionAccount(a *crbac.Account) {
	// old version use Role, now use Roles
	// Role/Roles will not exist at the same time
	if len(a.Role) == 0 {
		return
	}
	a.Roles = []string{a.Role}
	a.Role = ""
}

func (ds *RbacDAO) ListAccount(ctx context.Context) ([]*crbac.Account, int64, error) {
	kvs, n, err := etcdadpt.List(ctx, path.GenerateRBACAccountKey(""))
	if err != nil {
		return nil, 0, err
	}
	accounts := make([]*crbac.Account, 0, n)
	for _, v := range kvs {
		a := &crbac.Account{}
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
func (ds *RbacDAO) DeleteAccount(ctx context.Context, names []string) (bool, error) {
	if len(names) == 0 {
		return false, nil
	}
	var allOpts []etcdadpt.OpOptions
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
		allOpts = append(allOpts, opts...)
		syncOpts, err := esync.GenDeleteOpts(ctx, datasource.ResourceAccount, a.Name, a)
		if err != nil {
			log.Error("fail to create sync opts", err)
			return false, err
		}
		allOpts = append(allOpts, syncOpts...)
	}
	err := etcdadpt.Txn(ctx, allOpts)
	if err != nil {
		log.Error(rbac.ErrDeleteAccountFailed.Error(), err)
		return false, err
	}
	return true, nil
}

func (ds *RbacDAO) UpdateAccount(ctx context.Context, name string, account *crbac.Account) error {
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
	syncOpts, err := esync.GenUpdateOpts(ctx, datasource.ResourceAccount, account)
	if err != nil {
		log.Error("fail to create sync opts", err)
		return err
	}
	opts = append(opts, syncOpts...)
	err = etcdadpt.Txn(ctx, opts)
	if err != nil {
		log.Error("BatchCommit failed", err)
	}
	return err
}

func hasRole(account *crbac.Account, r string) bool {
	for _, n := range account.Roles {
		if r == n {
			return true
		}
	}
	return false
}
