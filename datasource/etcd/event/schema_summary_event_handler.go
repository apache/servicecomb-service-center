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
	"github.com/apache/servicecomb-service-center/datasource/etcd/kv"
	"github.com/apache/servicecomb-service-center/datasource/etcd/sd"
	pb "github.com/apache/servicecomb-service-center/pkg/registry"
	"github.com/apache/servicecomb-service-center/server/core"
	"github.com/apache/servicecomb-service-center/server/service/metrics"
	"strings"
)

// SchemaSummaryEventHandler report schema metrics
type SchemaSummaryEventHandler struct {
}

func (h *SchemaSummaryEventHandler) Type() sd.Type {
	return kv.SchemaSummary
}

func (h *SchemaSummaryEventHandler) OnEvent(evt sd.KvEvent) {
	action := evt.Type
	switch action {
	case pb.EVT_INIT, pb.EVT_CREATE, pb.EVT_DELETE:
		domainProject, _, _ := core.GetInfoFromSchemaSummaryKV(evt.KV.Key)
		idx := strings.Index(domainProject, "/")
		newDomain := domainProject[:idx]
		if pb.EVT_DELETE == action {
			metrics.ReportSchemas(newDomain, -1)
		} else {
			metrics.ReportSchemas(newDomain, 1)
		}
	default:
	}
}

func NewSchemaSummaryEventHandler() *SchemaSummaryEventHandler {
	return &SchemaSummaryEventHandler{}
}
