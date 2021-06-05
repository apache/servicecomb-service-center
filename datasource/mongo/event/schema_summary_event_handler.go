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
	pb "github.com/go-chassis/cari/discovery"

	"github.com/apache/servicecomb-service-center/datasource/mongo/model"
	"github.com/apache/servicecomb-service-center/datasource/mongo/sd"
	"github.com/apache/servicecomb-service-center/server/metrics"
)

type SchemaSummaryEventHandler struct {
}

func NewSchemaSummaryEventHandler() *SchemaSummaryEventHandler {
	return &SchemaSummaryEventHandler{}
}

func (h *SchemaSummaryEventHandler) Type() string {
	return model.ColumnSchema
}

func (h *SchemaSummaryEventHandler) OnEvent(evt sd.MongoEvent) {
	schema := evt.Value.(model.Schema)
	action := evt.Type
	switch action {
	case pb.EVT_INIT, pb.EVT_CREATE:
		metrics.ReportSchemas(schema.Domain, increaseOne)
	case pb.EVT_DELETE:
		metrics.ReportSchemas(schema.Domain, decreaseOne)
	default:
	}
}
