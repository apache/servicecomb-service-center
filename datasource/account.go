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

package datasource

import (
	"context"
	"errors"
	"github.com/apache/servicecomb-service-center/pkg/rbacframe"
)

var (
	ErrDuplicated = errors.New("account is duplicated")
	ErrCanNotEdit = errors.New("account can not be edited")
)

// AccountManager contains the RBAC CRUD
type AccountManager interface {
	CreateAccount(ctx context.Context, a *rbacframe.Account) error
	AccountExist(ctx context.Context, key string) (bool, error)
	GetAccount(ctx context.Context, key string) (*rbacframe.Account, error)
	ListAccount(ctx context.Context, key string) ([]*rbacframe.Account, int64, error)
	DeleteAccount(ctx context.Context, key string) (bool, error)
	UpdateAccount(ctx context.Context, key string, account *rbacframe.Account) error
}
