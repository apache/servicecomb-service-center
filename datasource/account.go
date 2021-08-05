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

	"github.com/go-chassis/cari/rbac"
)

var (
	ErrAccountDuplicated      = errors.New("account is duplicated")
	ErrAccountCanNotEdit      = errors.New("account can not be edited")
	ErrDLockNotFound          = errors.New("dlock not found")
	ErrCannotReleaseLock      = errors.New("can not release account")
	ErrAccountLockNotExist    = errors.New("account lock not exist")
	ErrDeleteAccountFailed    = errors.New("failed to delete account")
	ErrQueryAccountFailed     = errors.New("failed to query account")
	ErrQueryAccountLockFailed = errors.New("failed to query account lock")
	ErrAccountNotExist        = errors.New("account not exist")
	ErrRoleBindingExist       = errors.New("role is bind to account")
)

const (
	StatusBanned    = "banned"
	StatusAttempted = "attempted"
)

// AccountManager contains the RBAC CRUD
type AccountManager interface {
	CreateAccount(ctx context.Context, a *rbac.Account) error
	AccountExist(ctx context.Context, name string) (bool, error)
	GetAccount(ctx context.Context, name string) (*rbac.Account, error)
	ListAccount(ctx context.Context) ([]*rbac.Account, int64, error)
	DeleteAccount(ctx context.Context, names []string) (bool, error)
	UpdateAccount(ctx context.Context, name string, account *rbac.Account) error
}

// AccountLockManager saves login failure status
type AccountLockManager interface {
	UpsertLock(ctx context.Context, lock *AccountLock) error
	GetLock(ctx context.Context, key string) (*AccountLock, error)
	ListLock(ctx context.Context) ([]*AccountLock, int64, error)
	DeleteLock(ctx context.Context, key string) error
	DeleteLockList(ctx context.Context, keys []string) error
	Ban(ctx context.Context, key string) error
}
type AccountLock struct {
	Key       string `json:"key,omitempty"`
	Status    string `json:"status,omitempty"`
	ReleaseAt int64  `json:"releaseAt,omitempty" bson:"release_at"`
}
type AccountLockResponse struct {
	Total       int64          `json:"total"`
	AccountLock []*AccountLock `json:"data"`
}
