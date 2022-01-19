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

package mongo

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/apache/servicecomb-service-center/datasource/rbac"
	dmongo "github.com/go-chassis/cari/db/mongo"
	rbacmodel "github.com/go-chassis/cari/rbac"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/apache/servicecomb-service-center/datasource/mongo/dao"
	"github.com/apache/servicecomb-service-center/datasource/mongo/model"
	mutil "github.com/apache/servicecomb-service-center/datasource/mongo/util"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/privacy"
	"github.com/apache/servicecomb-service-center/pkg/util"
)

func init() {
	rbac.Install("mongo", NewRbacDAO)
}

func NewRbacDAO(opts rbac.Options) (rbac.DAO, error) {
	return &RbacDAO{}, nil
}

type RbacDAO struct {
}

func (ds *RbacDAO) CreateAccount(ctx context.Context, a *rbacmodel.Account) error {
	exist, err := ds.AccountExist(ctx, a.Name)
	if err != nil {
		msg := fmt.Sprintf("failed to query account, account name %s", a.Name)
		log.Error(msg, err)
		return err
	}
	if exist {
		return rbac.ErrAccountDuplicated
	}
	a.Password, err = privacy.ScryptPassword(a.Password)
	if err != nil {
		msg := fmt.Sprintf("failed to hash account pwd, account name %s", a.Name)
		log.Error(msg, err)
		return err
	}
	a.Role = ""
	a.CurrentPassword = ""
	a.ID = util.GenerateUUID()
	a.CreateTime = strconv.FormatInt(time.Now().Unix(), 10)
	a.UpdateTime = a.CreateTime
	_, err = dmongo.GetClient().GetDB().Collection(model.CollectionAccount).InsertOne(ctx, a)
	if err != nil {
		if dao.IsDuplicateKey(err) {
			return rbac.ErrAccountDuplicated
		}
		return err
	}
	log.Info("succeed to create new account: " + a.ID)
	return nil
}

func (ds *RbacDAO) AccountExist(ctx context.Context, name string) (bool, error) {
	filter := mutil.NewFilter(mutil.AccountName(name))
	count, err := dmongo.GetClient().GetDB().Collection(model.CollectionAccount).CountDocuments(ctx, filter)
	if err != nil {
		return false, err
	}
	if count == 0 {
		return false, nil
	}
	return true, nil
}

func (ds *RbacDAO) GetAccount(ctx context.Context, name string) (*rbacmodel.Account, error) {
	filter := mutil.NewFilter(mutil.AccountName(name))
	result := dmongo.GetClient().GetDB().Collection(model.CollectionAccount).FindOne(ctx, filter)
	if err := result.Err(); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, rbac.ErrAccountNotExist
		}
		msg := fmt.Sprintf("failed to query account, account name %s", name)
		log.Error(msg, result.Err())
		return nil, rbac.ErrQueryAccountFailed
	}
	var account rbacmodel.Account
	err := result.Decode(&account)
	if err != nil {
		log.Error("failed to decode account", err)
		return nil, err
	}
	return &account, nil
}

func (ds *RbacDAO) ListAccount(ctx context.Context) ([]*rbacmodel.Account, int64, error) {
	filter := mutil.NewFilter()
	cursor, err := dmongo.GetClient().GetDB().Collection(model.CollectionAccount).Find(ctx, filter)
	if err != nil {
		return nil, 0, err
	}
	var accounts []*rbacmodel.Account
	defer cursor.Close(ctx)
	for cursor.Next(ctx) {
		var account rbacmodel.Account
		err = cursor.Decode(&account)
		if err != nil {
			log.Error("failed to decode account", err)
			continue
		}
		account.Password = ""
		accounts = append(accounts, &account)
	}
	return accounts, int64(len(accounts)), nil
}

func (ds *RbacDAO) DeleteAccount(ctx context.Context, names []string) (bool, error) {
	if len(names) == 0 {
		return false, nil
	}
	inFilter := mutil.NewFilter(mutil.In(names))
	filter := mutil.NewFilter(mutil.AccountName(inFilter))
	result, err := dmongo.GetClient().GetDB().Collection(model.CollectionAccount).DeleteMany(ctx, filter)
	if err != nil {
		return false, err
	}
	if result.DeletedCount == 0 {
		return false, nil
	}
	return true, nil
}

func (ds *RbacDAO) UpdateAccount(ctx context.Context, name string, account *rbacmodel.Account) error {
	filter := mutil.NewFilter(mutil.AccountName(name))
	setFilter := mutil.NewFilter(
		mutil.ID(account.ID),
		mutil.Password(account.Password),
		mutil.Roles(account.Roles),
		mutil.TokenExpirationTime(account.TokenExpirationTime),
		mutil.CurrentPassword(account.CurrentPassword),
		mutil.Status(account.Status),
		mutil.AccountUpdateTime(strconv.FormatInt(time.Now().Unix(), 10)),
	)
	updateFilter := mutil.NewFilter(mutil.Set(setFilter))
	res, err := dmongo.GetClient().GetDB().Collection(model.CollectionAccount).UpdateMany(ctx, filter, updateFilter)
	if err != nil {
		return err
	}
	if res.ModifiedCount == 0 {
		return mutil.ErrNoDataToUpdate
	}
	return nil
}
