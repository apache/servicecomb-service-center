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

package event

import (
	"context"
	"fmt"

	"github.com/go-chassis/cari/discovery"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/apache/servicecomb-service-center/datasource"
	"github.com/apache/servicecomb-service-center/datasource/mongo/client"
	"github.com/apache/servicecomb-service-center/datasource/mongo/model"
	"github.com/apache/servicecomb-service-center/pkg/log"
)

func getService(ctx context.Context, filter interface{}, opts ...*options.FindOneOptions) (*model.Service, error) {
	result, err := client.GetMongoClient().FindOne(ctx, model.CollectionService, filter, opts...)
	if err != nil {
		return nil, err
	}
	var svc *model.Service
	if result.Err() != nil {
		return nil, datasource.ErrNoData
	}
	err = result.Decode(&svc)
	if err != nil {
		return nil, err
	}
	return svc, nil
}

func getRules(ctx context.Context, filter interface{}) ([]*model.Rule, error) {
	cursor, err := client.GetMongoClient().Find(ctx, model.CollectionRule, filter)
	if err != nil {
		return nil, err
	}
	if cursor.Err() != nil {
		return nil, cursor.Err()
	}
	var rules []*model.Rule
	defer cursor.Close(ctx)
	for cursor.Next(ctx) {
		var rule model.Rule
		err := cursor.Decode(&rule)
		if err != nil {
			log.Error("type conversion error", err)
			return nil, err
		}
		rules = append(rules, &rule)
	}
	return rules, nil
}

func getMicroServiceDependency(ctx context.Context, filter interface{}, opts ...*options.FindOneOptions) (*discovery.MicroServiceDependency, error) {
	findRes, err := client.GetMongoClient().FindOne(ctx, model.CollectionDep, filter, opts...)
	if err != nil {
		return nil, err
	}
	var depRule *model.DependencyRule
	if findRes.Err() != nil {
		return nil, datasource.ErrNoData
	}
	err = findRes.Decode(&depRule)
	if err != nil {
		return nil, err
	}
	return depRule.Dep, nil
}

func getTags(ctx context.Context, filter interface{}) (tags map[string]string, err error) {
	result, err := client.GetMongoClient().FindOne(ctx, model.CollectionService, filter)
	if err != nil {
		return nil, err
	}
	if result.Err() != nil {
		return nil, result.Err()
	}
	var service model.Service
	err = result.Decode(&service)
	if err != nil {
		log.Error("type conversion error", err)
		return nil, err
	}
	return service.Tags, nil
}

func addDomain(ctx context.Context, domain string) error {
	d := model.Domain{
		Domain: domain,
	}
	result, err := client.GetMongoClient().Insert(ctx, model.CollectionDomain, d)
	if err == nil {
		log.Info(fmt.Sprintf("insert domain to mongodb success %s", result.InsertedID))
	}
	return err
}

func existDomain(ctx context.Context, filter interface{}) (bool, error) {
	return client.GetMongoClient().DocExist(ctx, model.CollectionDomain, filter)
}

func addProject(ctx context.Context, project model.Project) error {
	result, err := client.GetMongoClient().Insert(ctx, model.CollectionProject, project)
	if err == nil {
		log.Info(fmt.Sprintf("insert project to mongodb success %s", result.InsertedID))
	}
	return err
}

func existProject(ctx context.Context, filter interface{}) (bool, error) {
	return client.GetMongoClient().DocExist(ctx, model.CollectionProject, filter)
}
