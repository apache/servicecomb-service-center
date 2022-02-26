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
	"errors"
	"fmt"

	"github.com/apache/servicecomb-service-center/datasource/rbac"
	"github.com/apache/servicecomb-service-center/pkg/log"
	v1sync "github.com/apache/servicecomb-service-center/syncer/api/v1"

	rbacmodel "github.com/go-chassis/cari/rbac"
)

const (
	Account = "account"
)

func NewAccount(e *v1sync.Event) Resource {
	a := &account{
		event: e,
	}
	a.manager = a
	return a
}

type accountManager interface {
	CreateAccount(ctx context.Context, a *rbacmodel.Account) error
	GetAccount(ctx context.Context, name string) (*rbacmodel.Account, error)
	UpdateAccount(ctx context.Context, account *rbacmodel.Account) error
	DeleteAccount(ctx context.Context, name string) error
}

type account struct {
	event *v1sync.Event

	input       *rbacmodel.Account
	accountName string

	cur *rbacmodel.Account

	defaultFailHandler

	manager accountManager
}

func (a *account) loadInput() error {
	a.input = new(rbacmodel.Account)
	callback := func() {
		a.accountName = a.input.Name
	}
	param := newInputParam(a.input, callback)

	return newInputLoader(
		a.event,
		param,
		param,
		param,
	).loadInput()
}

func (a *account) LoadCurrentResource(ctx context.Context) *Result {
	err := a.loadInput()
	if err != nil {
		return FailResult(err)
	}

	cur, err := a.manager.GetAccount(ctx, a.accountName)
	if err != nil {
		if errors.Is(err, rbac.ErrAccountNotExist) {
			return nil
		}
		return FailResult(err)
	}
	a.cur = cur
	return nil
}

func (a *account) NeedOperate(ctx context.Context) *Result {
	c := &checker{
		curNotNil: a.cur != nil,
		event:     a.event,
		updateTime: func() (int64, error) {
			return formatUpdateTimeSecond(a.cur.UpdateTime)
		},
		resourceID: a.input.Name,
	}
	c.tombstoneLoader = c
	return c.needOperate(ctx)
}

func (a *account) CreateHandle(ctx context.Context) error {
	if a.cur != nil {
		log.Warn(fmt.Sprintf("create action but account exist, %s",
			a.accountName))
		return a.UpdateHandle(ctx)
	}
	return a.manager.CreateAccount(ctx, a.input)
}

func (a *account) UpdateHandle(ctx context.Context) error {
	if a.cur == nil {
		log.Warn(fmt.Sprintf("update action but account not exist, %s",
			a.accountName))
		return a.CreateHandle(ctx)
	}
	return a.manager.UpdateAccount(ctx, a.input)
}

func (a *account) DeleteHandle(ctx context.Context) error {
	return a.manager.DeleteAccount(ctx, a.input.Name)
}

func (a *account) Operate(ctx context.Context) *Result {
	return newOperator(a).operate(ctx, a.event.Action)
}

func (a *account) CreateAccount(ctx context.Context, at *rbacmodel.Account) error {
	return rbac.Instance().CreateAccount(ctx, at)
}

func (a *account) GetAccount(ctx context.Context, name string) (*rbacmodel.Account, error) {
	return rbac.Instance().GetAccount(ctx, name)
}

func (a *account) UpdateAccount(ctx context.Context, account *rbacmodel.Account) error {
	return rbac.Instance().UpdateAccount(ctx, account.Name, account)
}

func (a *account) DeleteAccount(ctx context.Context, name string) error {
	_, err := rbac.Instance().DeleteAccount(ctx, []string{name})
	return err
}
