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
	"github.com/ServiceComb/service-center/pkg/util"
	"github.com/ServiceComb/service-center/server/core/backend/store"
	pb "github.com/ServiceComb/service-center/server/core/proto"
)

type DependencyEventHandler struct {
}

func (h *DependencyEventHandler) Type() store.StoreType {
	return store.SERVICE
}

func (h *DependencyEventHandler) OnEvent(evt *store.KvEvent) {
	action := evt.Action
	if action != pb.EVT_CREATE && action != pb.EVT_INIT {
		return
	}

	kv := evt.KV
	method, _, data := pb.GetInfoFromDependencyKV(kv)
	if data == nil {
		util.Logger().Errorf(nil,
			"unmarshal dependency file failed, method %s [%s] event, data is nil",
			method, action)
		return
	}
	// TODO maintain dependency rules.
}

func NewDependencyEventHandler() *DependencyEventHandler {
	return &DependencyEventHandler{}
}
