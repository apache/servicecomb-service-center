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
	"github.com/apache/servicecomb-service-center/pkg/rbacframe"
)

func (ds *DataSource) CreateAccount(ctx context.Context, a *rbacframe.Account) error {
	return nil
}

func (ds *DataSource) AccountExist(ctx context.Context, key string) (bool, error) {
	return false, nil
}

func (ds *DataSource) GetAccount(ctx context.Context, key string) (*rbacframe.Account, error) {
	return &rbacframe.Account{}, nil
}

func (ds *DataSource) ListAccount(ctx context.Context, key string) ([]*rbacframe.Account, int64, error) {
	return nil, 0, nil
}

func (ds *DataSource) DeleteAccount(ctx context.Context, key string) (bool, error) {
	return false, nil
}

func (ds *DataSource) UpdateAccount(ctx context.Context, key string, account *rbacframe.Account) error {
	return nil
}
