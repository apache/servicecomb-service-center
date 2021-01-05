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

	"go.mongodb.org/mongo-driver/bson"

	"github.com/apache/servicecomb-service-center/datasource"
	"github.com/apache/servicecomb-service-center/datasource/mongo/client"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/rbacframe"
	"github.com/apache/servicecomb-service-center/pkg/util"
)

func (ds *DataSource) CreateRole(ctx context.Context, r *rbacframe.Role) error {
	exist, err := ds.RoleExist(ctx, r.Name)
	if err != nil {
		log.Error("failed to query role", err)
		return err
	}
	if exist {
		return datasource.ErrRoleDuplicated
	}
	r.ID = util.GenerateUUID()
	_, err = client.GetMongoClient().Insert(ctx, CollectionRole, r)
	if err != nil {
		if client.IsDuplicateKey(err) {
			return datasource.ErrRoleDuplicated
		}
		return err
	}
	log.Info("create new role: " + r.ID)
	return nil
}

func (ds *DataSource) RoleExist(ctx context.Context, name string) (bool, error) {
	filter := bson.M{
		ColumnRoleName: name,
	}
	count, err := client.GetMongoClient().Count(ctx, CollectionRole, filter)
	if err != nil {
		return false, err
	}
	if count == 0 {
		return false, nil
	}
	return true, nil
}

func (ds *DataSource) GetRole(ctx context.Context, name string) (*rbacframe.Role, error) {
	filter := bson.M{
		ColumnRoleName: name,
	}
	result, err := client.GetMongoClient().FindOne(ctx, CollectionRole, filter)
	if err != nil {
		return nil, err
	}
	if result.Err() != nil {
		return nil, client.ErrNoDocuments
	}
	var role rbacframe.Role
	err = result.Decode(&role)
	if err != nil {
		log.Error("Decode role failed: ", err)
		return nil, err
	}
	return &role, nil
}

func (ds *DataSource) ListRole(ctx context.Context) ([]*rbacframe.Role, int64, error) {
	cursor, err := client.GetMongoClient().Find(ctx, CollectionRole, bson.M{})
	if err != nil {
		return nil, 0, err
	}
	var roles []*rbacframe.Role
	defer cursor.Close(ctx)
	for cursor.Next(ctx) {
		var role rbacframe.Role
		err = cursor.Decode(&role)
		if err != nil {
			log.Error("decode role failed: ", err)
			continue
		}
		roles = append(roles, &role)
	}
	return roles, int64(len(roles)), nil
}

func (ds *DataSource) DeleteRole(ctx context.Context, name string) (bool, error) {
	filter := bson.M{
		ColumnRoleName: name,
	}
	result, err := client.GetMongoClient().Delete(ctx, CollectionRole, filter)
	if err != nil {
		return false, err
	}
	if result.DeletedCount == 0 {
		return false, nil
	}
	return true, nil
}

func (ds *DataSource) UpdateRole(ctx context.Context, name string, role *rbacframe.Role) error {
	filter := bson.M{
		ColumnRoleName: name,
	}
	update := bson.M{
		"$set": bson.M{
			ColumnID:       role.ID,
			ColumnRoleName: role.Name,
			ColumnPerms:    role.Perms,
		},
	}
	_, err := client.GetMongoClient().Update(ctx, CollectionRole, filter, update)
	if err != nil {
		return err
	}
	return nil
}
