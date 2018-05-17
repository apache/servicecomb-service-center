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
package backend

import (
	"github.com/apache/incubator-servicecomb-service-center/server/core/proto"
)

type KvEventFunc func(evt KvEvent)

type KvEvent struct {
	Revision int64
	Type     proto.EventType
	Prefix   string
	Object   interface{}
}

type KvEventHandler interface {
	Type() StoreType
	OnEvent(evt KvEvent)
}

// the event handler/func must be good performance, or will block the event bus.
func AddEventHandleFunc(t StoreType, f KvEventFunc) {
	EventProxy(t).AddHandleFunc(f)
}

func AddEventHandler(h KvEventHandler) {
	AddEventHandleFunc(h.Type(), h.OnEvent)
}
