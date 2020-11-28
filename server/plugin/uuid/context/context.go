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

package context

import (
	"context"
	"crypto/sha1"
	"fmt"

	"github.com/apache/servicecomb-service-center/pkg/util"
	mgr "github.com/apache/servicecomb-service-center/server/plugin"
	"github.com/apache/servicecomb-service-center/server/plugin/uuid"
	"github.com/apache/servicecomb-service-center/server/plugin/uuid/buildin"
)

func init() {
	mgr.RegisterPlugin(mgr.Plugin{Kind: uuid.UUID, Name: "context", New: New})
}

func New() mgr.Instance {
	return &UUID{}
}

type UUID struct {
	buildin.UUID
}

func (cu *UUID) fromContext(ctx context.Context) string {
	key, ok := ctx.Value(uuid.ContextKey).(string)
	if !ok {
		return ""
	}
	return key
}

func (cu *UUID) GetServiceID(ctx context.Context) string {
	content := cu.fromContext(ctx)
	if len(content) == 0 {
		return cu.UUID.GetServiceID(ctx)
	}
	return fmt.Sprintf("%x", sha1.Sum(util.StringToBytesWithNoCopy(content)))
}
