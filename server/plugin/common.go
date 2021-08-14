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
	pg "plugin"

	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/plugin"
)

const (
	BUILDIN       = "buildin"
	STATIC        = "static"
	DYNAMIC       = "dynamic"
	keyPluginName = "name"
)

// DynamicPluginFunc should be called in buildin implement
func DynamicPluginFunc(pn Name, funcName string) pg.Symbol {
	if wi, ok := Plugins().instances[pn]; ok && !wi.dynamic {
		return nil
	}

	f, err := plugin.FindFunc(pn.String(), funcName)
	if err != nil {
		log.Errorf(err, "plugin '%s': not implemented function '%s'", pn, funcName)
	}
	return f
}
