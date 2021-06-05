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

package service

import (
	"context"
	"fmt"

	"github.com/go-chassis/cari/rbac"

	"github.com/apache/servicecomb-service-center/datasource"
	mutil "github.com/apache/servicecomb-service-center/datasource/mongo/util"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/privacy"
	"github.com/apache/servicecomb-service-center/pkg/util"
)

func (ds *DataSource) CreateAccount(ctx context.Context, a *rbac.Account) error {
	exist, err := ds.AccountExist(ctx, a.Name)
	if err != nil {
		msg := fmt.Sprintf("failed to query account, account name %s", a.Name)
		log.Error(msg, err)
		return err
	}
	if exist {
		return datasource.ErrAccountDuplicated
	}
	a.Password, err = privacy.ScryptPassword(a.Password)
	if err != nil {
		msg := fmt.Sprintf("failed to hash account pwd, account name %s", a.Name)
		log.Error(msg, err)
		return err
	}
	a.ID = util.GenerateUUID()
	err = insertAccount(ctx, a)
	if err != nil {
		if mutil.IsDuplicateKey(err) {
			return datasource.ErrAccountDuplicated
		}
		return err
	}
	log.Info("succeed to create new account: " + a.ID)
	return nil
}

func (ds *DataSource) AccountExist(ctx context.Context, name string) (bool, error) {
	filter := mutil.NewFilter(mutil.AccountName(name))
	count, err := countAccount(ctx, filter)
	if err != nil {
		return false, err
	}
	if count == 0 {
		return false, nil
	}
	return true, nil
}

func (ds *DataSource) GetAccount(ctx context.Context, name string) (*rbac.Account, error) {
	filter := mutil.NewFilter(mutil.AccountName(name))
	account, err := findAccount(ctx, filter)
	if err != nil {
		msg := fmt.Sprintf("failed to find account, account name %s", name)
		if err == datasource.ErrNoData {
			return nil, datasource.ErrAccountNotExist
		}
		log.Error(msg, err)
		return nil, err
	}
	return account, nil
}

func (ds *DataSource) ListAccount(ctx context.Context) ([]*rbac.Account, int64, error) {
	filter := mutil.NewFilter()
	return findAccounts(ctx, filter)
}

func (ds *DataSource) DeleteAccount(ctx context.Context, names []string) (bool, error) {
	if len(names) == 0 {
		return false, nil
	}
	inFilter := mutil.NewFilter(mutil.In(names))
	filter := mutil.NewFilter(mutil.AccountName(inFilter))
	return deleteAccount(ctx, filter)
}

func (ds *DataSource) UpdateAccount(ctx context.Context, name string, account *rbac.Account) error {
	filter := mutil.NewFilter(mutil.AccountName(name))
	setValue := mutil.NewFilter(
		mutil.ID(account.ID),
		mutil.Password(account.Password), mutil.Roles(account.Roles),
		mutil.TokenExpirationTime(account.TokenExpirationTime),
		mutil.CurrentPassword(account.CurrentPassword),
		mutil.Status(account.Status),
	)
	updateFilter := mutil.NewFilter(mutil.Set(setValue))
	return updateAccount(ctx, filter, updateFilter)
}
