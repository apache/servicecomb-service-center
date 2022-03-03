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

package sd

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/apache/servicecomb-service-center/datasource/mongo/model"
	"github.com/apache/servicecomb-service-center/datasource/sdcommon"
)

func TestListWatchConfig_String(t *testing.T) {
	t.Run("TestListWatchConfig_String", func(t *testing.T) {
		config := sdcommon.ListWatchConfig{
			Timeout: 666,
		}
		ret := config.String()
		assert.Equal(t, "{timeout: 666ns}", ret)
	})
	t.Run("when time is nil", func(t *testing.T) {
		config := sdcommon.ListWatchConfig{}
		ret := config.String()
		assert.Equal(t, "{timeout: 0s}", ret)
	})
}

func TestDoParseWatchRspToMongoInfo(t *testing.T) {
	documentID := primitive.NewObjectID()

	mockDocument, _ := bson.Marshal(bson.M{"_id": documentID, "domain": "default", "project": "default", "instance": bson.M{"instance_id": "8064a600438511eb8584fa163e8a81c9", "service_id": "91afbe0faa9dc1594689139f099eb293b0cd048d",
		"hostname": "ecs-hcsadlab-dev-0002", "status": "UP", "timestamp": "1608552622", "mod_timestamp": "1608552622", "version": "0.0.1"}})

	mockServiceDocument, _ := bson.Marshal(bson.M{"_id": documentID, "domain": "default", "project": "default", "service": bson.M{"service_id": "91afbe0faa9dc1594689139f099eb293b0cd048d", "timestamp": "1608552622", "mod_timestamp": "1608552622", "version": "0.0.1"}})

	// case instance insertOp

	mockWatchRsp := &MongoWatchResponse{OperationType: insertOp,
		FullDocument: mockDocument,
		DocumentKey:  MongoDocument{ID: documentID},
	}
	ilw := mongoListWatch{
		Key: instance,
		parseFunc: func(doc bson.Raw) (resource sdcommon.Resource) {
			docID := MongoDocument{}
			err := bson.Unmarshal(doc, &docID)
			if err != nil {
				return
			}
			service := model.Instance{}
			err = bson.Unmarshal(doc, &service)
			if err != nil {
				return
			}
			resource.Value = service
			resource.Key = docID.ID.Hex()
			return
		},
	}
	info := ilw.doParseWatchRspToResource(mockWatchRsp)
	assert.Equal(t, documentID.Hex(), info.Key)

	// case updateOp
	mockWatchRsp.OperationType = updateOp
	info = ilw.doParseWatchRspToResource(mockWatchRsp)
	assert.Equal(t, documentID.Hex(), info.Key)
	assert.Equal(t, "1608552622", info.Value.(model.Instance).Instance.ModTimestamp)

	// case delete
	mockWatchRsp.OperationType = deleteOp
	info = ilw.doParseWatchRspToResource(mockWatchRsp)
	assert.Equal(t, documentID.Hex(), info.Key)

	// case service insertOp
	mockWatchRsp = &MongoWatchResponse{OperationType: insertOp,
		FullDocument: mockServiceDocument,
		DocumentKey:  MongoDocument{ID: primitive.NewObjectID()},
	}
	ilw.Key = service
	info = ilw.doParseWatchRspToResource(mockWatchRsp)
	assert.Equal(t, documentID.Hex(), info.Key)
}

func TestInnerListWatch_ResumeToken(t *testing.T) {
	ilw := mongoListWatch{
		Key:         instance,
		resumeToken: bson.Raw("resumToken"),
	}
	t.Run("get resume token test", func(t *testing.T) {
		res := ilw.ResumeToken()
		assert.NotNil(t, res)
		assert.Equal(t, bson.Raw("resumToken"), res)
	})

	t.Run("set resume token test", func(t *testing.T) {
		ilw.setResumeToken(bson.Raw("resumToken2"))
		assert.Equal(t, ilw.resumeToken, bson.Raw("resumToken2"))
	})
}
