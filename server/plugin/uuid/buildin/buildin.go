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

package buildin

import (
	"context"

	"github.com/apache/servicecomb-service-center/pkg/util"
	mgr "github.com/apache/servicecomb-service-center/server/plugin"
)

func init() {
	mgr.RegisterPlugin(mgr.Plugin{PName: mgr.UUID, Name: "buildin", New: New})
}

func New() mgr.Instance {
	return &UUID{}
}

type UUID struct {
}

func (du *UUID) GetServiceID(_ context.Context) string {
	df, ok := mgr.DynamicPluginFunc(mgr.UUID, "GetServiceID").(func() string)
	if ok {
		return df()
	}
	return util.GenerateUUID()
}

func (du *UUID) GetInstanceID(_ context.Context) string {
	df, ok := mgr.DynamicPluginFunc(mgr.UUID, "GetInstanceID").(func() string)
	if ok {
		return df()
	}
	return util.GenerateUUID()
}
