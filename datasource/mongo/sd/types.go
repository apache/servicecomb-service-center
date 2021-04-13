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
	"github.com/go-chassis/cari/discovery"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"sync"

	"github.com/apache/servicecomb-service-center/datasource/sdcommon"
)

const (
	service  = "service"
	instance = "instance"
	rule     = "rule"
	dep      = "dependency"
)

type cacherRegisterInitiallizer func() (cacher *MongoCacher)

var cacherRegisterMutex sync.Mutex
var CacherRegister = make(map[string]cacherRegisterInitiallizer)

func RegisterCacher(t string, f cacherRegisterInitiallizer) {
	cacherRegisterMutex.Lock()
	defer cacherRegisterMutex.Unlock()
	if _, exist := CacherRegister[t]; exist {
		return
	}
	CacherRegister[t] = f
}

type MongoEvent struct {
	DocumentID string
	Value      interface{}
	Type       discovery.EventType
}

type MongoEventFunc func(evt MongoEvent)

type MongoEventHandler interface {
	Type() string
	OnEvent(evt MongoEvent)
}

func NewMongoEventByResource(resource *sdcommon.Resource, action discovery.EventType) MongoEvent {
	return MongoEvent{
		Type:       action,
		Value:      resource.Value,
		DocumentID: resource.Key,
	}
}

func NewMongoEvent(documentID string, action discovery.EventType, v interface{}) MongoEvent {
	event := MongoEvent{}
	event.DocumentID = documentID
	event.Type = action
	event.Value = v
	return event
}

type MongoWatchResponse struct {
	OperationType string
	FullDocument  bson.Raw
	DocumentKey   MongoDocument
}

type MongoDocument struct {
	ID primitive.ObjectID `bson:"_id"`
}

type ResumeToken struct {
	Data []byte `bson:"_data"`
}
