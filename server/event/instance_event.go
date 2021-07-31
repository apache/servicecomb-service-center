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
	"github.com/apache/servicecomb-service-center/pkg/event"
	simple "github.com/apache/servicecomb-service-center/pkg/time"
	pb "github.com/go-chassis/cari/discovery"
)

const QueueSize = 5000

var INSTANCE = event.RegisterType("INSTANCE", QueueSize)

// 状态变化推送
type InstanceEvent struct {
	event.Event
	Revision int64
	Response *pb.WatchInstanceResponse
}

func NewInstanceEvent(serviceID string, rev int64, createAt simple.Time, response *pb.WatchInstanceResponse) *InstanceEvent {
	return &InstanceEvent{
		Event:    event.NewEventWithTime(INSTANCE, response.Key.Tenant, serviceID, createAt),
		Revision: rev,
		Response: response,
	}
}
