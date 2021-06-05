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
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/apache/servicecomb-service-center/datasource"
	"github.com/apache/servicecomb-service-center/datasource/mongo/client"
	"github.com/apache/servicecomb-service-center/datasource/mongo/model"
	"github.com/apache/servicecomb-service-center/pkg/log"
)

func insertRole(ctx context.Context, document interface{}, opts ...*options.InsertOneOptions) error {
	_, err := client.GetMongoClient().Insert(ctx, model.CollectionRole, document, opts...)
	return err
}

func countRole(ctx context.Context, filter interface{}, opts ...*options.CountOptions) (int64, error) {
	return client.GetMongoClient().Count(ctx, model.CollectionRole, filter, opts...)
}

func findRole(ctx context.Context, filter interface{}, opts ...*options.FindOneOptions) (*rbac.Role, error) {
	result, err := client.GetMongoClient().FindOne(ctx, model.CollectionRole, filter, opts...)
	if err != nil {
		log.Error("failed to find role", err)
		return nil, err
	}
	if result.Err() != nil {
		return nil, datasource.ErrRoleNotExist
	}
	var role rbac.Role
	err = result.Decode(&role)
	if err != nil {
		log.Error("failed to decode role", err)
		return nil, err
	}
	return &role, nil
}

func findRoles(ctx context.Context, filter interface{}, opts ...*options.FindOptions) ([]*rbac.Role, int64, error) {
	cursor, err := client.GetMongoClient().Find(ctx, model.CollectionRole, filter, opts...)
	if err != nil {
		log.Error("failed to find roles", err)
		return nil, 0, err
	}
	var roles []*rbac.Role
	defer cursor.Close(ctx)
	for cursor.Next(ctx) {
		var role rbac.Role
		err = cursor.Decode(&role)
		if err != nil {
			log.Error("failed to decode role", err)
			continue
		}
		roles = append(roles, &role)
	}
	return roles, int64(len(roles)), nil
}

func deleteRole(ctx context.Context, filter interface{}, opts ...*options.DeleteOptions) (bool, error) {
	result, err := client.GetMongoClient().Delete(ctx, model.CollectionRole, filter, opts...)
	if err != nil {
		return false, err
	}
	if result.DeletedCount == 0 {
		return false, nil
	}
	return true, nil
}

func updateRole(ctx context.Context, filter interface{}, update interface{}, opts ...*options.UpdateOptions) error {
	result, err := client.GetMongoClient().Update(ctx, model.CollectionRole, filter, update, opts...)
	if err != nil {
		return err
	}
	if result.ModifiedCount == 0 {
		return ErrNoDataToUpdate
	}
	return nil
}
