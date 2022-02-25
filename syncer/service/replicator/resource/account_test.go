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

package resource

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"testing"
	"time"

	v1sync "github.com/apache/servicecomb-service-center/syncer/api/v1"

	datasourcerbac "github.com/apache/servicecomb-service-center/datasource/rbac"
	"github.com/go-chassis/cari/rbac"
	"github.com/go-chassis/cari/sync"
	"github.com/stretchr/testify/assert"
)

type mockAccount struct {
	accounts map[string]*rbac.Account
}

func (f mockAccount) CreateAccount(_ context.Context, a *rbac.Account) error {
	_, ok := f.accounts[a.Name]
	if ok {
		return datasourcerbac.ErrAccountDuplicated
	}

	f.accounts[a.Name] = a
	return nil
}

func (f mockAccount) GetAccount(_ context.Context, name string) (*rbac.Account, error) {
	result, ok := f.accounts[name]
	if !ok {
		return nil, datasourcerbac.ErrAccountNotExist
	}
	return result, nil
}

func (f mockAccount) UpdateAccount(_ context.Context, account *rbac.Account) error {
	name := account.Name
	_, ok := f.accounts[name]
	if !ok {
		return datasourcerbac.ErrAccountNotExist
	}
	f.accounts[name] = account
	return nil
}

func (f mockAccount) DeleteAccount(_ context.Context, name string) error {
	_, ok := f.accounts[name]
	if !ok {
		return datasourcerbac.ErrAccountNotExist
	}
	delete(f.accounts, name)
	return nil
}

func TestOperateAccount(t *testing.T) {
	t.Run("create update delete case", func(t *testing.T) {
		createTime := strconv.FormatInt(time.Now().Unix(), 10)
		input := &rbac.Account{
			ID:         "xxxx",
			Name:       "test",
			Password:   "xxx",
			Roles:      []string{"admin"},
			Status:     "active",
			CreateTime: createTime,
			UpdateTime: createTime,
		}
		value, _ := json.Marshal(input)
		id, _ := v1sync.NewEventID()
		e := &v1sync.Event{
			Id:        id,
			Action:    sync.CreateAction,
			Subject:   Account,
			Opts:      nil,
			Value:     value,
			Timestamp: v1sync.Timestamp(),
		}
		a := &account{
			event: e,
		}
		a.manager = &mockAccount{
			accounts: make(map[string]*rbac.Account),
		}
		ctx := context.Background()
		result := a.LoadCurrentResource(ctx)
		if assert.Nil(t, result) {
			result = a.NeedOperate(ctx)
			if assert.Nil(t, result) {
				result = a.Operate(ctx)
				if assert.NotNil(t, result) && assert.Equal(t, Success, result.Status) {
					data, err := a.manager.GetAccount(ctx, "test")
					assert.Nil(t, err)
					assert.NotNil(t, data)
				}
			}
		}

		updateTime := strconv.FormatInt(time.Now().Unix(), 10)
		input = &rbac.Account{
			ID:         "xxxx",
			Name:       "test",
			Password:   "xxx",
			Roles:      []string{"developer"},
			Status:     "active",
			CreateTime: createTime,
			UpdateTime: updateTime,
		}
		value, _ = json.Marshal(input)

		id, _ = v1sync.NewEventID()
		e1 := &v1sync.Event{
			Id:        id,
			Action:    sync.UpdateAction,
			Subject:   Account,
			Opts:      nil,
			Value:     value,
			Timestamp: v1sync.Timestamp(),
		}
		a1 := &account{
			event:   e1,
			manager: a.manager,
		}
		result = a1.LoadCurrentResource(ctx)
		if assert.Nil(t, result) {
			result = a1.NeedOperate(ctx)
			if assert.Nil(t, result) {
				result = a1.Operate(ctx)
				if assert.NotNil(t, result) && assert.Equal(t, Success, result.Status) {
					data, err := a1.manager.GetAccount(ctx, "test")
					assert.Nil(t, err)
					assert.NotNil(t, data)
					assert.Equal(t, []string{"developer"}, data.Roles)
				}
			}
		}

		id, _ = v1sync.NewEventID()
		e2 := &v1sync.Event{
			Id:        id,
			Action:    sync.DeleteAction,
			Subject:   Account,
			Opts:      nil,
			Value:     value,
			Timestamp: v1sync.Timestamp(),
		}
		a2 := &account{
			event:   e2,
			manager: a.manager,
		}
		result = a2.LoadCurrentResource(ctx)
		if assert.Nil(t, result) {
			result = a2.NeedOperate(ctx)
			if assert.Nil(t, result) {
				result = a2.Operate(ctx)
				if assert.NotNil(t, result) && assert.Equal(t, Success, result.Status) {
					_, err := a1.manager.GetAccount(ctx, "test")
					assert.NotNil(t, err)
					assert.True(t, errors.Is(err, datasourcerbac.ErrAccountNotExist))
				}
			}
		}
	})
}

func TestNewAccount(t *testing.T) {
	id, _ := v1sync.NewEventID()
	e := &v1sync.Event{
		Id:        id,
		Action:    sync.DeleteAction,
		Subject:   Account,
		Opts:      nil,
		Value:     nil,
		Timestamp: v1sync.Timestamp(),
	}
	a := NewAccount(e)
	assert.NotNil(t, a)
}
