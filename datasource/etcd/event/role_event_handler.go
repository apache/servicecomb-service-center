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
	"github.com/apache/servicecomb-service-center/datasource/etcd/sd"
	"github.com/apache/servicecomb-service-center/server/metrics"
	pb "github.com/go-chassis/cari/discovery"
)

// RoleEventHandler is the handler to handle:
type RoleEventHandler struct {
}

func (h *RoleEventHandler) Type() sd.Type {
	return "ROLE"
}

func (h *RoleEventHandler) OnEvent(evt sd.KvEvent) {
	action := evt.Type
	domainName := "default" //TODO ...

	if action == pb.EVT_INIT || action == pb.EVT_CREATE {
		metrics.ReportRoles(domainName, 1)
		return
	}

	if action == pb.EVT_DELETE {
		metrics.ReportRoles(domainName, -1)
		return
	}
}

func NewRoleEventHandler() *RoleEventHandler {
	return &RoleEventHandler{}
}
