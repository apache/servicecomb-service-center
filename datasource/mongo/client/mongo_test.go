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

package client

import (
	"context"
	"testing"

	_ "github.com/apache/servicecomb-service-center/server/init"
	_ "github.com/apache/servicecomb-service-center/server/plugin/security/cipher/buildin"

	"github.com/go-chassis/go-chassis/v2/storage"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	TESTCOL = "test"
)

func init() {
	config := storage.Options{
		URI: "mongodb://localhost:27017",
	}
	NewMongoClient(config)
}

func TestInsert(t *testing.T) {
	insertRes, err := GetMongoClient().Insert(context.Background(), TESTCOL, bson.M{
		"instance": "instance1",
		"number":   123,
	})
	if err != nil {
		t.Fatalf("TestMongoInsert failed, %#v", err)
	}
	res, err := GetMongoClient().Find(context.Background(), TESTCOL, bson.M{"_id": insertRes.InsertedID})
	if err != nil {
		t.Fatalf("TestMongoInsert check insert result failed, %#v", err)
	}
	var result bson.M
	var flag bool
	for res.Next(context.Background()) {
		err := res.Decode(&result)
		if err != nil {
			t.Fatalf("TestMongoInsert decode result failed, %#v", err)
		}
		flag = true
	}
	if !flag {
		t.Fatalf("TestMongoInsert check res failed, can't get insert doc")
	}
}

func TestBatchInsert(t *testing.T) {
	insertRes, err := GetMongoClient().BatchInsert(context.Background(), TESTCOL, []interface{}{bson.M{"instance": "instance2"}, bson.M{"instance": "instance3"}})
	if err != nil {
		t.Fatalf("TestMongoBatchInsert failed, %#v", err)
	}
	res, err := GetMongoClient().Find(context.Background(), TESTCOL, bson.M{"_id": bson.M{"$in": insertRes.InsertedIDs}})
	if err != nil {
		t.Fatalf("TestBatchMongoInsert query mongdb failed, %#v", err)
	}
	var result []bson.M
	for res.Next(context.Background()) {
		var doc bson.M
		err := res.Decode(&doc)
		if err != nil {
			t.Fatalf("TestBatchInsert decode result failed, %#v", err)
		}
		result = append(result, doc)
	}
	if len(result) != 2 {
		t.Fatalf("TestBatchInsert check result failed")
	}
}

func TestDelete(t *testing.T) {
	insertRes, err := GetMongoClient().Insert(context.Background(), TESTCOL, bson.M{"instance": "instance4"})
	if err != nil {
		t.Fatalf("TestMongoDelete insert failed, %#v", err)
	}
	deleteRes, err := GetMongoClient().Delete(context.Background(), TESTCOL, bson.M{"_id": insertRes.InsertedID})
	if err != nil {
		t.Fatalf("TestMongoDelete delte failed, %#v", err)
	}
	if deleteRes.DeletedCount != 1 {
		t.Fatalf("TestDelete check deleteRes failed")
	}
}

func TestBatchDelete(t *testing.T) {
	insertRes, err := GetMongoClient().BatchInsert(context.Background(), TESTCOL, []interface{}{bson.M{"instance": "instance5"}, bson.M{"instance": "instance6"}})
	if err != nil {
		t.Fatalf("TestMongoBatchDelete insert failed, %#v", err)
	}
	var wm []mongo.WriteModel
	for _, id := range insertRes.InsertedIDs {
		filter := bson.M{"_id": id}
		model := mongo.NewDeleteManyModel().SetFilter(filter)
		wm = append(wm, model)
	}
	batchDelRes, err := GetMongoClient().BatchDelete(context.Background(), TESTCOL, wm)
	if err != nil {
		t.Fatalf("TestBatchDelete batchDelete failed, %#v", err)
	}
	if batchDelRes.DeletedCount != 2 {
		t.Fatalf("TestBatchDelete check result failed")
	}
}

func TestUpdate(t *testing.T) {
	insertRes, err := GetMongoClient().Insert(context.Background(), TESTCOL, bson.M{
		"instance": "instance7",
	})
	if err != nil {
		t.Fatalf("TestMongoUpdate insert failed, %#v", err)
	}
	filter := bson.M{"_id": insertRes.InsertedID}
	update := bson.M{"$set": bson.M{"add": 1}}
	updateRes, err := GetMongoClient().Update(context.Background(), TESTCOL, filter, update)
	if err != nil {
		t.Fatalf("TestMongoUpdate update failed,%#v", err)
	}
	if updateRes.ModifiedCount != 1 {
		t.Fatalf("TestMongoUpdate check result failed,%#v", err)
	}
}