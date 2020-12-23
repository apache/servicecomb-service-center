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
	"time"

	"github.com/go-chassis/cari/discovery"
	pb "github.com/go-chassis/cari/discovery"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	service  = "service"
	instance = "instance"
)

var (
	Types []string
)

func RegisterType(name string) {
	Types = append(Types, name)
}

type MongoEvent struct {
	DocumentID string
	BusinessID string
	Value      interface{}
	Type       discovery.EventType
}

type MongoEventFunc func(evt MongoEvent)

type MongoEventHandler interface {
	Type() string
	OnEvent(evt MongoEvent)
}

func NewMongoEventByMongoInf(info MongoInfo, action discovery.EventType) MongoEvent {
	return MongoEvent{BusinessID: info.BusinessID, DocumentID: info.DocumentID, Value: info.Value, Type: action}
}

func NewMongoEvent(id string, documentID string, action discovery.EventType, v interface{}) MongoEvent {
	event := MongoEvent{}
	event.BusinessID = id
	event.DocumentID = documentID
	event.Type = action
	event.Value = v
	return event
}

type MongoListWatchResponse struct {
	OperationType string
	Infos         []MongoInfo
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

type MongoInfo struct {
	DocumentID string
	BusinessID string
	Value      interface{}
}

type Service struct {
	Domain      string
	Project     string
	Tags        map[string]string
	ServiceInfo *pb.MicroService
}

type Instance struct {
	Domain       string
	Project      string
	RefreshTime  time.Time
	InstanceInfo *pb.MicroServiceInstance
}
