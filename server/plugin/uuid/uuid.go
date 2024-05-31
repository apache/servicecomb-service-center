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

package uuid

import (
	"context"

	"github.com/apache/servicecomb-service-center/pkg/plugin"
	"github.com/apache/servicecomb-service-center/pkg/util"
)

const (
	UUID       plugin.Kind = "uuid"
	ContextKey util.CtxKey = "_uuid_key"
)

type IDGenerator interface {
	GetServiceID(ctx context.Context) string
	GetInstanceID(ctx context.Context) string
	GetEnvID(ctx context.Context) string
}

func Generator() IDGenerator {
	return plugin.Plugins().Instance(UUID).(IDGenerator)
}
