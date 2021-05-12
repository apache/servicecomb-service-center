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

	rbacmodel "github.com/go-chassis/cari/rbac"

	"github.com/apache/servicecomb-service-center/datasource"
	"github.com/apache/servicecomb-service-center/pkg/log"
)

//CreateAccount save 2 kv
//1. account info
func CreateAccount(ctx context.Context, a *rbacmodel.Account) error {
	return datasource.Instance().CreateAccount(ctx, a)
}

func GetAccount(ctx context.Context, name string) (*rbacmodel.Account, error) {
	return datasource.Instance().GetAccount(ctx, name)
}
func ListAccount(ctx context.Context) ([]*rbacmodel.Account, int64, error) {
	return datasource.Instance().ListAccount(ctx)
}
func AccountExist(ctx context.Context, name string) (bool, error) {
	return datasource.Instance().AccountExist(ctx, name)
}
func DeleteAccount(ctx context.Context, name string) (bool, error) {
	return datasource.Instance().DeleteAccount(ctx, []string{name})
}

//CreateAccount save 2 kv
//1. account info
func EditAccount(ctx context.Context, a *rbacmodel.Account) error {
	exist, err := datasource.Instance().AccountExist(ctx, a.Name)
	if err != nil {
		log.Errorf(err, "can not edit account info")
		return err
	}
	if !exist {
		return datasource.ErrAccountCanNotEdit
	}

	err = datasource.Instance().UpdateAccount(ctx, a.Name, a)
	if err != nil {
		log.Errorf(err, "can not edit account info")
		return err
	}
	log.Info("account is edit")
	return nil
}
