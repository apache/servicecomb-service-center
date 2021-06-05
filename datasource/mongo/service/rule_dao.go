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

	"github.com/go-chassis/cari/discovery"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/apache/servicecomb-service-center/datasource/mongo/client"
	"github.com/apache/servicecomb-service-center/datasource/mongo/model"
	"github.com/apache/servicecomb-service-center/pkg/log"
)

func getServiceRules(ctx context.Context, filter interface{}, opts ...*options.FindOptions) ([]*discovery.ServiceRule, error) {
	ruleRes, err := client.GetMongoClient().Find(ctx, model.CollectionRule, filter, opts...)
	if err != nil {
		return nil, err
	}
	var rules []*discovery.ServiceRule
	for ruleRes.Next(ctx) {
		var tempRule *model.Rule
		err := ruleRes.Decode(&tempRule)
		if err != nil {
			return nil, err
		}
		rules = append(rules, tempRule.Rule)
	}
	return rules, nil
}

func findRules(ctx context.Context, filter interface{}, opts ...*options.FindOptions) ([]*model.Rule, error) {
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
			log.Error("failed to decode rule", err)
			return nil, err
		}
		rules = append(rules, &rule)
	}
	return rules, nil
}

func existRule(ctx context.Context, filter interface{}) (bool, error) {
	return client.GetMongoClient().DocExist(ctx, model.CollectionRule, filter)
}

func insertRule(ctx context.Context, document interface{}, opts ...*options.InsertOneOptions) error {
	_, err := client.GetMongoClient().Insert(ctx, model.CollectionRule, document, opts...)
	return err
}

func batchDeleteRule(ctx context.Context, models []mongo.WriteModel, opts ...*options.BulkWriteOptions) error {
	_, err := client.GetMongoClient().BatchDelete(ctx, model.CollectionRule, models, opts...)
	return err
}

func updateRule(ctx context.Context, filter interface{}, updateFilter interface{}) error {
	return client.GetMongoClient().DocUpdate(ctx, model.CollectionRule, filter, updateFilter)
}
