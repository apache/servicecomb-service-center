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
	"strconv"
	"time"

	"github.com/apache/servicecomb-service-center/datasource/mongo/client"
	"github.com/apache/servicecomb-service-center/datasource/mongo/client/model"
	mutil "github.com/apache/servicecomb-service-center/datasource/mongo/util"
	"github.com/apache/servicecomb-service-center/datasource/rbac"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/util"
	rbacmodel "github.com/go-chassis/cari/rbac"
	"go.mongodb.org/mongo-driver/bson"
)

func (ds *RbacDAO) CreateRole(ctx context.Context, r *rbacmodel.Role) error {
	exist, err := ds.RoleExist(ctx, r.Name)
	if err != nil {
		log.Error("failed to query role", err)
		return err
	}
	if exist {
		return rbac.ErrRoleDuplicated
	}
	r.ID = util.GenerateUUID()
	r.CreateTime = strconv.FormatInt(time.Now().Unix(), 10)
	r.UpdateTime = r.CreateTime
	_, err = client.GetMongoClient().Insert(ctx, model.CollectionRole, r)
	if err != nil {
		if client.IsDuplicateKey(err) {
			return rbac.ErrRoleDuplicated
		}
		return err
	}
	log.Info("succeed to create new role: " + r.ID)
	return nil
}

func (ds *RbacDAO) RoleExist(ctx context.Context, name string) (bool, error) {
	filter := mutil.NewFilter(mutil.RoleName(name))
	count, err := client.GetMongoClient().Count(ctx, model.CollectionRole, filter)
	if err != nil {
		return false, err
	}
	if count == 0 {
		return false, nil
	}
	return true, nil
}

func (ds *RbacDAO) GetRole(ctx context.Context, name string) (*rbacmodel.Role, error) {
	filter := mutil.NewFilter(mutil.RoleName(name))
	result, err := client.GetMongoClient().FindOne(ctx, model.CollectionRole, filter)
	if err != nil {
		return nil, err
	}
	if result.Err() != nil {
		return nil, rbac.ErrRoleNotExist
	}
	var role rbacmodel.Role
	err = result.Decode(&role)
	if err != nil {
		log.Error("failed to decode role", err)
		return nil, err
	}
	return &role, nil
}

func (ds *RbacDAO) ListRole(ctx context.Context) ([]*rbacmodel.Role, int64, error) {
	filter := mutil.NewFilter()
	cursor, err := client.GetMongoClient().Find(ctx, model.CollectionRole, filter)
	if err != nil {
		return nil, 0, err
	}
	var roles []*rbacmodel.Role
	defer cursor.Close(ctx)
	for cursor.Next(ctx) {
		var role rbacmodel.Role
		err = cursor.Decode(&role)
		if err != nil {
			log.Error("failed to decode role", err)
			continue
		}
		roles = append(roles, &role)
	}
	return roles, int64(len(roles)), nil
}

func (ds *RbacDAO) DeleteRole(ctx context.Context, name string) (bool, error) {
	n, err := client.Count(ctx, model.CollectionAccount, bson.M{"roles": bson.M{"$in": []string{name}}})
	if err != nil {
		return false, err
	}
	if n > 0 {
		return false, rbac.ErrRoleBindingExist
	}
	filter := mutil.NewFilter(mutil.RoleName(name))
	result, err := client.DeleteDoc(ctx, model.CollectionRole, filter)
	if err != nil {
		return false, err
	}
	if result.DeletedCount == 0 {
		return false, nil
	}
	return true, nil
}

func (ds *RbacDAO) UpdateRole(ctx context.Context, name string, role *rbacmodel.Role) error {
	filter := mutil.NewFilter(mutil.RoleName(name))
	setFilter := mutil.NewFilter(
		mutil.ID(role.ID),
		mutil.RoleName(role.Name),
		mutil.Perms(role.Perms),
		mutil.RoleUpdateTime(strconv.FormatInt(time.Now().Unix(), 10)),
	)
	updateFilter := mutil.NewFilter(mutil.Set(setFilter))
	_, err := client.GetMongoClient().Update(ctx, model.CollectionRole, filter, updateFilter)
	if err != nil {
		return err
	}
	return nil
}
