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

func (ds *DataSource) CreateRole(ctx context.Context, r *rbacframe.Role) error {
	return nil
}

func (ds *DataSource) RoleExist(ctx context.Context, name string) (bool, error) {
	return false, nil
}

func (ds *DataSource) GetRole(ctx context.Context, name string) (*rbacframe.Role, error) {
	return &rbacframe.Role{}, nil
}

func (ds *DataSource) ListRole(ctx context.Context) ([]*rbacframe.Role, int64, error) {
	return nil, 0, nil
}

func (ds *DataSource) DeleteRole(ctx context.Context, name string) (bool, error) {
	return false, nil
}

func (ds *DataSource) UpdateRole(ctx context.Context, name string, role *rbacframe.Role) error {
	return nil
}
