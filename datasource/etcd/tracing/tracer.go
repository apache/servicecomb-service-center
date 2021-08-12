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

package tracing

import (
	tracesvc "github.com/apache/servicecomb-service-center/server/plugin/tracing"
	"github.com/little-cui/etcdadpt/middleware/tracing"
)

type Tracer struct{}

func (t *Tracer) Begin(operationName string, r *tracing.Request) (span interface{}) {
	return tracesvc.ClientBegin(operationName, &tracesvc.Request{
		Ctx:      r.Ctx,
		Endpoint: r.Endpoint,
		Method:   r.Options.Action.String(),
		URL:      "/?" + r.Options.URI(),
	})
}

func (t *Tracer) End(span interface{}, response *tracing.Response) {
	tracesvc.ClientEnd(span, response.Code, response.Message)
}

func New() *Tracer {
	return &Tracer{}
}
