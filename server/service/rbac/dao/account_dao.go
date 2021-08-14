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

//Package rbac is dao layer API to help service center manage account, policy and role info
package dao

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/apache/servicecomb-service-center/pkg/etcdsync"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/privacy"
	"github.com/apache/servicecomb-service-center/pkg/rbacframe"
	"github.com/apache/servicecomb-service-center/pkg/util"
	"github.com/apache/servicecomb-service-center/server/core"
	"github.com/apache/servicecomb-service-center/server/service/kv"
)

var ErrDuplicated = errors.New("account is duplicated")
var ErrCanNotEdit = errors.New("account can not be edited")

//CreateAccount save 2 kv
//1. account info
func CreateAccount(ctx context.Context, a *rbacframe.Account) error {
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
	key := core.GenerateAccountKey(a.Name)
	exist, err := kv.Exist(ctx, key)
	if err != nil {
		log.Errorf(err, "can not save account info")
		return err
	}
	if exist {
		return ErrDuplicated
	}
	hash, err := privacy.ScryptPassword(a.Password)
	if err != nil {
		log.Errorf(err, "pwd hash failed")
		return err
	}
	a.Password = hash
	a.ID = util.GenerateUUID()
	value, err := json.Marshal(a)
	if err != nil {
		log.Errorf(err, "account info is invalid")
		return err
	}
	err = kv.PutBytes(ctx, key, value)
	if err != nil {
		log.Errorf(err, "can not save account info")
		return err
	}
	log.Info("create new account: " + a.ID)
	return nil
}

func GetAccount(ctx context.Context, name string) (*rbacframe.Account, error) {
	key := core.GenerateAccountKey(name)
	r, err := kv.Get(ctx, key)
	if err != nil {
		log.Errorf(err, "can not get account info")
		return nil, err
	}
	a := &rbacframe.Account{}
	err = json.Unmarshal(r.Value, a)
	if err != nil {
		log.Errorf(err, "account info format invalid")
		return nil, err
	}
	return a, nil
}
func ListAccount(ctx context.Context) ([]*rbacframe.Account, int64, error) {
	key := core.GenerateAccountKey("")
	r, n, err := kv.List(ctx, key)
	if err != nil {
		log.Errorf(err, "can not get account info")
		return nil, 0, err
	}
	as := make([]*rbacframe.Account, 0, n)
	for _, v := range r {
		a := &rbacframe.Account{}
		err = json.Unmarshal(v.Value, a)
		if err != nil {
			log.Error("account info format invalid:", err)
			continue //do not fail if some account is invalid
		}
		a.Password = ""
		as = append(as, a)
	}
	return as, n, nil
}
func AccountExist(ctx context.Context, name string) (bool, error) {
	exist, err := kv.Exist(ctx, core.GenerateAccountKey(name))
	if err != nil {
		log.Errorf(err, "can not get account info")
		return false, err
	}
	return exist, nil
}
func DeleteAccount(ctx context.Context, name string) (bool, error) {
	exist, err := kv.Delete(ctx, core.GenerateAccountKey(name))
	if err != nil {
		log.Errorf(err, "can not get account info")
		return false, err
	}
	log.Info("account is deleted")
	return exist, nil
}

//CreateAccount save 2 kv
//1. account info
func EditAccount(ctx context.Context, a *rbacframe.Account) error {
	key := core.GenerateAccountKey(a.Name)
	exist, err := kv.Exist(ctx, key)
	if err != nil {
		log.Errorf(err, "can not edit account info")
		return err
	}
	if !exist {
		return ErrCanNotEdit
	}

	value, err := json.Marshal(a)
	if err != nil {
		log.Errorf(err, "account info is invalid")
		return err
	}
	err = kv.PutBytes(ctx, key, value)
	if err != nil {
		log.Errorf(err, "can not edit account info")
		return err
	}
	log.Info("account is edit")
	return nil
}
