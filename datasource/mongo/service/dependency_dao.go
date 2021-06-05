/*
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements.  See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership.  The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License.  You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

package service

import (
	"context"

	"github.com/go-chassis/cari/discovery"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/apache/servicecomb-service-center/datasource"
	"github.com/apache/servicecomb-service-center/datasource/mongo/client"
	"github.com/apache/servicecomb-service-center/datasource/mongo/model"
	"github.com/apache/servicecomb-service-center/pkg/log"
)

func insertDep(ctx context.Context, document interface{}, opts ...*options.InsertOneOptions) error {
	_, err := client.GetMongoClient().Insert(ctx, model.CollectionDep, document, opts...)
	return err
}

func findDep(ctx context.Context, filter interface{}, opts ...*options.FindOneOptions) (*model.DependencyRule, error) {
	result, err := client.GetMongoClient().FindOne(ctx, model.CollectionDep, filter, opts...)
	if err != nil {
		log.Error("failed to find dependency", err)
		return nil, err
	}
	if result.Err() != nil {
		return nil, datasource.ErrNoData
	}
	var depRule *model.DependencyRule
	err = result.Decode(&depRule)
	if err != nil {
		return nil, err
	}
	return depRule, nil
}

func findDeps(ctx context.Context, filter interface{}, opts ...*options.FindOptions) ([]*model.DependencyRule, error) {
	findRes, err := client.GetMongoClient().Find(ctx, model.CollectionDep, filter, opts...)
	if err != nil {
		return nil, err
	}
	defer findRes.Close(ctx)
	var depRules []*model.DependencyRule
	for findRes.Next(ctx) {
		var tmp *model.DependencyRule
		err := findRes.Decode(&tmp)
		if err != nil {
			return nil, err
		}
		depRules = append(depRules, tmp)
	}
	return depRules, nil
}

func findMicroServiceDependency(ctx context.Context, filter interface{}, opts ...*options.FindOneOptions) (*discovery.MicroServiceDependency, error) {
	microServiceDependency := &discovery.MicroServiceDependency{
		Dependency: []*discovery.MicroServiceKey{},
	}
	findRes, err := client.GetMongoClient().FindOne(ctx, model.CollectionDep, filter, opts...)
	if err != nil {
		return nil, err
	}
	if findRes.Err() == nil {
		var depRule *model.DependencyRule
		err := findRes.Decode(&depRule)
		if err != nil {
			log.Error("failed to decode dependency's rule", err)
			return nil, err
		}
		microServiceDependency.Dependency = append(microServiceDependency.Dependency, depRule.Dep.Dependency...)
		return microServiceDependency, nil
	}
	return microServiceDependency, nil
}

func findMicroServiceDependencies(ctx context.Context, filter interface{}, opts ...*options.FindOptions) ([]*discovery.MicroServiceKey, error) {
	var microServiceKeys []*discovery.MicroServiceKey
	cursor, err := client.GetMongoClient().Find(ctx, model.CollectionDep, filter, opts...)
	if err != nil {
		return nil, err
	}
	for cursor.Next(ctx) {
		var depRule *model.DependencyRule
		err = cursor.Decode(&depRule)
		if err != nil {
			return nil, err
		}
		microServiceKeys = append(microServiceKeys, depRule.Dep.Dependency...)
	}
	return microServiceKeys, nil
}

func updateServiceDependencies(ctx context.Context, filter interface{}, update interface{}, opts ...*options.UpdateOptions) error {
	_, err := client.GetMongoClient().Update(ctx, model.CollectionDep, filter, update, opts...)
	if err != nil {
		log.Error("failed to update dependencies", err)
		return err
	}
	return nil
}

func deleteServiceDependencies(ctx context.Context, filter interface{}, opts ...*options.DeleteOptions) error {
	_, err := client.GetMongoClient().Delete(ctx, model.CollectionDep, filter, opts...)
	return err
}

func getDepRules(ctx context.Context, filter interface{}, opts ...*options.FindOptions) ([]*model.DependencyRule, error) {
	findRes, err := client.GetMongoClient().Find(ctx, model.CollectionDep, filter, opts...)
	if err != nil {
		return nil, err
	}

	var depRules []*model.DependencyRule
	for findRes.Next(ctx) {
		var depRule *model.DependencyRule
		err := findRes.Decode(&depRule)
		if err != nil {
			return nil, err
		}
		depRules = append(depRules, depRule)
	}
	return depRules, nil
}
