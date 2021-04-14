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
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	md "go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/apache/servicecomb-service-center/datasource/mongo/client"
	"github.com/apache/servicecomb-service-center/datasource/mongo/client/model"
	"github.com/apache/servicecomb-service-center/datasource/sdcommon"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/util"
)

type mongoListWatch struct {
	Key         string
	resumeToken bson.Raw
}

func (lw *mongoListWatch) List(op sdcommon.ListWatchConfig) (*sdcommon.ListWatchResp, error) {
	otCtx, cancel := context.WithTimeout(op.Context, op.Timeout)
	defer cancel()

	resp, err := client.GetMongoClient().Find(otCtx, lw.Key, bson.M{})
	if err != nil {
		log.Error(fmt.Sprintf("list key %s failed", lw.Key), err)
		return nil, err
	}

	// convert mongoListResponse to ListWatchResp
	lwRsp := &sdcommon.ListWatchResp{}
	lwRsp.Resources = make([]*sdcommon.Resource, 0)

	for resp.Next(context.Background()) {
		info := lw.doParseDocumentToResource(resp.Current)
		lwRsp.Resources = append(lwRsp.Resources, &info)
	}

	return lwRsp, nil
}

func (lw *mongoListWatch) EventBus(op sdcommon.ListWatchConfig) *sdcommon.EventBus {
	return sdcommon.NewEventBus(lw, op)
}

func (lw *mongoListWatch) DoWatch(ctx context.Context, f func(*sdcommon.ListWatchResp)) error {
	csOptions := &options.ChangeStreamOptions{}
	csOptions.SetFullDocument(options.UpdateLookup)

	resumeToken := lw.resumeToken
	if resumeToken != nil {
		csOptions.SetResumeAfter(resumeToken)
	}
	pipline := md.Pipeline{}
	if lw.Key == instance {
		// ignore instance refresh_time change event for avoid meaningless instance push.
		match := bson.D{{"updateDescription.updatedFields.refresh_time", bson.D{{"$exists", false}}}}
		pipline = md.Pipeline{{{"$match", match}}}
	}
	resp, err := client.GetMongoClient().Watch(ctx, lw.Key, pipline, csOptions)

	if err != nil {
		log.Error(fmt.Sprintf("watch table %s failed", lw.Key), err)
		f(nil)
		return err
	}

	for resp.Next(ctx) {
		lw.resumeToken = resp.ResumeToken()

		wRsp := &MongoWatchResponse{}
		err := bson.Unmarshal(resp.Current, &wRsp)

		if err != nil {
			log.Error("error to parse bson raw to mongo watch response", err)
			return err
		}

		// convert mongoWatchResponse to ListWatchResp
		resource := lw.doParseWatchRspToResource(wRsp)

		lwRsp := &sdcommon.ListWatchResp{}
		lwRsp.Resources = append(lwRsp.Resources, &resource)
		switch wRsp.OperationType {
		case insertOp:
			lwRsp.Action = sdcommon.ActionCreate
		case updateOp:
			lwRsp.Action = sdcommon.ActionUpdate
		case deleteOp:
			lwRsp.Action = sdcommon.ActionDelete
		default:
			log.Warn(fmt.Sprintf("unrecognized action:%s", lwRsp.Action))
		}

		f(lwRsp)
	}

	return err
}

func (lw *mongoListWatch) doParseDocumentToResource(fullDocument bson.Raw) (resource sdcommon.Resource) {
	var err error

	documentID := MongoDocument{}
	err = bson.Unmarshal(fullDocument, &documentID)
	if err != nil {
		return
	}

	resource.DocumentID = documentID.ID.Hex()

	switch lw.Key {
	case instance:
		instance := model.Instance{}
		err = bson.Unmarshal(fullDocument, &instance)
		if err != nil {
			log.Error("error to parse bson raw to documentInfo", err)
			return
		}
		if instance.Instance == nil {
			log.Error(fmt.Sprintf("unexpect instance value,the documentID is %s", resource.DocumentID), nil)
			return
		}
		resource.Key = instance.Instance.InstanceId
		resource.Value = instance
		resource.Index = instance.Instance.ServiceId
	case service:
		service := model.Service{}
		err := bson.Unmarshal(fullDocument, &service)
		if err != nil {
			log.Error("error to parse bson raw to documentInfo", err)
			return
		}
		if service.Service == nil {
			log.Error(fmt.Sprintf("unexpect service value,the documentID is %s", resource.DocumentID), nil)
			return
		}
		resource.Key = service.Service.ServiceId
		resource.Value = service
		resource.Index = util.StringJoin([]string{service.Domain, service.Project, service.Service.ServiceName, service.Service.Version, service.Service.AppId, service.Service.Environment}, "/")
	default:
		return
	}

	return
}

func (lw *mongoListWatch) ResumeToken() bson.Raw {
	return lw.resumeToken
}

func (lw *mongoListWatch) setResumeToken(resumeToken bson.Raw) {
	lw.resumeToken = resumeToken
}

func (lw *mongoListWatch) doParseWatchRspToResource(wRsp *MongoWatchResponse) (resource sdcommon.Resource) {
	switch wRsp.OperationType {
	case deleteOp:
		//delete operation has no fullDocumentValue
		resource.DocumentID = wRsp.DocumentKey.ID.Hex()
		return
	case insertOp, updateOp, replaceOp:
		return lw.doParseDocumentToResource(wRsp.FullDocument)
	default:
		log.Warn(fmt.Sprintf("unrecognized operation:%s", wRsp.OperationType))
	}
	return
}
