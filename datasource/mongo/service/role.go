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

	"github.com/go-chassis/cari/rbac"
	"go.mongodb.org/mongo-driver/bson"

	"github.com/apache/servicecomb-service-center/datasource"
	"github.com/apache/servicecomb-service-center/datasource/mongo/client"
	"github.com/apache/servicecomb-service-center/datasource/mongo/model"
	mutil "github.com/apache/servicecomb-service-center/datasource/mongo/util"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/util"
)

func (ds *DataSource) CreateRole(ctx context.Context, r *rbac.Role) error {
	exist, err := ds.RoleExist(ctx, r.Name)
	if err != nil {
		log.Error("failed to query role", err)
		return err
	}
	if exist {
		return datasource.ErrRoleDuplicated
	}
	r.ID = util.GenerateUUID()
	err = insertRole(ctx, r)
	if err != nil {
		if mutil.IsDuplicateKey(err) {
			return datasource.ErrRoleDuplicated
		}
		return err
	}
	log.Info("succeed to create new role: " + r.ID)
	return nil
}

func (ds *DataSource) RoleExist(ctx context.Context, name string) (bool, error) {
	filter := mutil.NewFilter(mutil.RoleName(name))
	count, err := countRole(ctx, filter)
	if err != nil {
		return false, err
	}
	if count == 0 {
		return false, nil
	}
	return true, nil
}

func (ds *DataSource) GetRole(ctx context.Context, name string) (*rbac.Role, error) {
	filter := mutil.NewFilter(mutil.RoleName(name))
	return findRole(ctx, filter)

}

func (ds *DataSource) ListRole(ctx context.Context) ([]*rbac.Role, int64, error) {
	filter := mutil.NewFilter()
	return findRoles(ctx, filter)
}

func (ds *DataSource) DeleteRole(ctx context.Context, name string) (bool, error) {
	n, _ := client.Count(ctx, model.CollectionAccount, bson.M{"roles": bson.M{"$in": []string{name}}})
	if n > 0 {
		return false, datasource.ErrRoleBindingExist
	}
	filter := mutil.NewFilter(mutil.RoleName(name))
	return deleteRole(ctx, filter)
}

func (ds *DataSource) UpdateRole(ctx context.Context, name string, role *rbac.Role) error {
	filter := mutil.NewFilter(mutil.RoleName(name))
	setValue := mutil.NewFilter(
		mutil.ID(role.ID),
		mutil.RoleName(role.Name),
		mutil.Perms(role.Perms),
	)
	updateFilter := mutil.NewFilter(mutil.Set(setValue))
	return updateRole(ctx, filter, updateFilter)
}
