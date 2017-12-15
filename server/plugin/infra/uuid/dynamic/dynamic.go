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

package dynamic

import (
	"github.com/ServiceComb/service-center/pkg/plugin"
	"github.com/ServiceComb/service-center/pkg/util"
	"github.com/ServiceComb/service-center/pkg/uuid"
	mgr "github.com/ServiceComb/service-center/server/plugin"
)

var (
	serviceIdFunc  func() string
	instanceIdFunc func() string
)

func init() {
	serviceIdFunc = findUuidFunc("GetServiceId")
	instanceIdFunc = findUuidFunc("GetInstanceId")

	mgr.RegisterPlugin(mgr.Plugin{mgr.DYNAMIC, mgr.UUID, "dynamic", New})
}

func buildinUuidFunc() string {
	return uuid.GenerateUuid()
}

func findUuidFunc(funcName string) func() string {
	ff, err := plugin.FindFunc("uuid", funcName)
	if err != nil {
		return buildinUuidFunc
	}
	f, ok := ff.(func() string)
	if !ok {
		util.Logger().Warnf(nil, "unexpected function '%s' format found in plugin 'uuid'.", funcName)
		return buildinUuidFunc
	}
	return f
}

func New() mgr.PluginInstance {
	return &DynamicUUID{}
}

type DynamicUUID struct {
}

func (du *DynamicUUID) GetServiceId() string {
	return serviceIdFunc()
}

func (du *DynamicUUID) GetInstanceId() string {
	return instanceIdFunc()
}
