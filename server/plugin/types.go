// Licensed to the Apache Software Foundation (ASF) under one or more
// contributor license agreements.  See the NOTICE file distributed with
// this work for additional information regarding copyright ownership.
// The ASF licenses this file to You under the Apache License, Version 2.0
// (the "License"); you may not use this file except in compliance with
// the License.  You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package plugin

import (
	"github.com/apache/servicecomb-service-center/pkg/util"
	"github.com/apache/servicecomb-service-center/server/core"
	"strconv"
)

type PluginInstance interface{}

type PluginName int

func (pn PluginName) String() string {
	if name, ok := pluginNames[pn]; ok {
		return name
	}
	return "PLUGIN" + strconv.Itoa(int(pn))
}

func (pn PluginName) ActiveConfigs() util.JSONObject {
	return core.ServerInfo.Config.Plugins.Object(pn.String())
}

func (pn PluginName) ClearConfigs() {
	core.ServerInfo.Config.Plugins.Set(pn.String(), nil)
}

type Plugin struct {
	// plugin class name
	PName PluginName
	// plugin name
	Name string
	New  func() PluginInstance
}
