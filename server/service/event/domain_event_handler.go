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
	"github.com/apache/incubator-servicecomb-service-center/server/core/backend"
	pb "github.com/apache/incubator-servicecomb-service-center/server/core/proto"
	"github.com/apache/incubator-servicecomb-service-center/server/infra/discovery"
	"github.com/apache/incubator-servicecomb-service-center/server/service/metrics"
)

type DomainEventHandler struct {
}

func (h *DomainEventHandler) Type() discovery.Type {
	return backend.DOMAIN
}

func (h *DomainEventHandler) OnEvent(evt discovery.KvEvent) {
	action := evt.Type
	switch action {
	case pb.EVT_INIT, pb.EVT_CREATE:
		metrics.ReportDomains(1)
	case pb.EVT_DELETE:
		metrics.ReportDomains(-1)
	}
}

func NewDomainEventHandler() *DomainEventHandler {
	return &DomainEventHandler{}
}
