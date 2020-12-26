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
	"context"
	"github.com/apache/servicecomb-service-center/datasource/mongo/client"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"go.mongodb.org/mongo-driver/bson"
	md "go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type innerListWatch struct {
	Client      *client.MongoClient
	Key         string
	resumeToken bson.Raw
}

func (lw *innerListWatch) List(op ListWatchConfig) (*MongoListWatchResponse, error) {
	otCtx, cancel := context.WithTimeout(op.Context, op.Timeout)
	defer cancel()

	resp, err := client.GetMongoClient().Find(otCtx, lw.Key, bson.M{})
	if err != nil {
		log.Errorf(err, "list key %s failed", lw.Key)
		return nil, err
	}

	lwRsp := &MongoListWatchResponse{}
	lwRsp.Infos = make([]MongoInfo, 0)
	for resp.Next(context.Background()) {
		info := lw.doParseDocumentToMongoInfo(resp.Current)
		lwRsp.Infos = append(lwRsp.Infos, info)
	}

	return lwRsp, nil
}

func (lw *innerListWatch) ResumeToken() bson.Raw {
	return lw.resumeToken
}

func (lw *innerListWatch) setResumeToken(resumeToken bson.Raw) {
	lw.resumeToken = resumeToken
}

func (lw *innerListWatch) Watch(op ListWatchConfig) Watcher {
	return newInnerWatcher(lw, op)
}

func (lw *innerListWatch) DoWatch(ctx context.Context, f func(*MongoListWatchResponse)) error {
	csOptions := &options.ChangeStreamOptions{}
	csOptions.SetFullDocument(options.UpdateLookup)

	resumeToken := lw.ResumeToken()
	if resumeToken != nil {
		csOptions.SetResumeAfter(resumeToken)
	}

	resp, err := client.GetMongoClient().Watch(ctx, lw.Key, md.Pipeline{}, csOptions)

	if err != nil {
		log.Errorf(err, "watch table %s failed", lw.Key)
		f(nil)
		return err
	}

	for resp.Next(ctx) {
		lwRsp := &MongoListWatchResponse{}

		lw.setResumeToken(resp.ResumeToken())

		wRsp := &MongoWatchResponse{}
		err := bson.Unmarshal(resp.Current, &wRsp)

		if err != nil {
			log.Errorf(err, "error to parse bson raw to mongo watch response")
			return err
		}

		info := lw.doParseWatchRspToMongoInfo(wRsp)

		lwRsp.OperationType = wRsp.OperationType
		lwRsp.Infos = append(lwRsp.Infos, info)

		f(lwRsp)
	}

	return err
}

func (lw *innerListWatch) doParseDocumentToMongoInfo(fullDocument bson.Raw) (info MongoInfo) {
	var err error

	documentId := MongoDocument{}
	err = bson.Unmarshal(fullDocument, &documentId)
	if err != nil {
		return
	}

	info.DocumentId = documentId.Id.Hex()

	switch lw.Key {
	case INSTANCE:
		instance := Instance{}
		err = bson.Unmarshal(fullDocument, &instance)
		if err != nil {
			log.Errorf(err, "error to parse bson raw to documentInfo")
			return
		}
		info.BusinessId = instance.InstanceInfo.InstanceId
		info.Value = instance
	case SERVICE:
		service := Service{}
		err := bson.Unmarshal(fullDocument, &service)
		if err != nil {
			log.Errorf(err, "error to parse bson raw to documentInfo")
			return
		}
		info.BusinessId = service.ServiceInfo.ServiceId
		info.Value = service
	default:
		return
	}
	return
}

func (lw *innerListWatch) doParseWatchRspToMongoInfo(wRsp *MongoWatchResponse) (info MongoInfo) {
	switch wRsp.OperationType {
	case DELETE:
		//delete operation has no fullDocumentValue
		info.DocumentId = wRsp.DocumentKey.Id.Hex()
		return
	case INSERT, UPDATE, REPLACE:
		return lw.doParseDocumentToMongoInfo(wRsp.FullDocument)
	}
	return
}
