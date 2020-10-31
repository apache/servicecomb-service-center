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
	"github.com/apache/servicecomb-service-center/datasource"
	"github.com/apache/servicecomb-service-center/datasource/etcd/client"
	"github.com/apache/servicecomb-service-center/pkg/etcdsync"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/rbacframe"
	"github.com/apache/servicecomb-service-center/pkg/util"
	"github.com/apache/servicecomb-service-center/server/core"
	stringutil "github.com/go-chassis/foundation/string"
	"golang.org/x/crypto/bcrypt"
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
	exist, err := datasource.Instance().AccountExist(ctx, a.Name)
	if err != nil {
		log.Errorf(err, "can not save account info")
		return err
	}
	if exist {
		return ErrDuplicated
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(a.Password), 14)
	if err != nil {
		log.Errorf(err, "pwd hash failed")
		return err
	}
	a.Password = stringutil.Bytes2str(hash)
	a.ID = util.GenerateUUID()
	value, err := json.Marshal(a)
	if err != nil {
		log.Errorf(err, "account info is invalid")
		return err
	}
	err = client.PutBytes(ctx, key, value)
	if err != nil {
		log.Errorf(err, "can not save account info")
		return err
	}
	log.Info("create new account: " + a.ID)
	return nil
}

func GetAccount(ctx context.Context, name string) (*rbacframe.Account, error) {
	return datasource.Instance().GetAccount(ctx, name)
}
func ListAccount(ctx context.Context) ([]*rbacframe.Account, int64, error) {
	return datasource.Instance().ListAccount(ctx, "")
}
func AccountExist(ctx context.Context, name string) (bool, error) {
	return datasource.Instance().AccountExist(ctx, name)
}
func DeleteAccount(ctx context.Context, name string) (bool, error) {
	return datasource.Instance().DeleteAccount(ctx, name)
}

//CreateAccount save 2 kv
//1. account info
func EditAccount(ctx context.Context, a *rbacframe.Account) error {
	exist, err := datasource.Instance().AccountExist(ctx, a.Name)
	if err != nil {
		log.Errorf(err, "can not edit account info")
		return err
	}
	if !exist {
		return ErrCanNotEdit
	}

	err = datasource.Instance().UpdateAccount(ctx, a.Name, a)
	if err != nil {
		log.Errorf(err, "can not edit account info")
		return err
	}
	log.Info("account is edit")
	return nil
}
