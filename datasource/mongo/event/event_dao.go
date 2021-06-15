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

	"github.com/go-chassis/cari/discovery"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/apache/servicecomb-service-center/datasource"
	"github.com/apache/servicecomb-service-center/datasource/mongo/client"
	"github.com/apache/servicecomb-service-center/datasource/mongo/client/model"
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

func getMicroServiceKeysByDependencyRule(ctx context.Context, filter interface{}, opts ...*options.FindOptions) ([]*discovery.MicroServiceKey, error) {
	res, err := client.GetMongoClient().Find(ctx, model.CollectionDep, filter, opts...)
	if err != nil {
		return nil, err
	}
	microServiceKeys := make([]*discovery.MicroServiceKey, 0)
	defer res.Close(ctx)
	for res.Next(ctx) {
		var depRule *model.DependencyRule
		err = res.Decode(&depRule)
		if err != nil {
			log.Error("failed to decode dependencyRule", err)
			continue
		}
		microServiceKeys = append(microServiceKeys, depRule.Dep.Dependency...)
	}
	return microServiceKeys, nil
}
