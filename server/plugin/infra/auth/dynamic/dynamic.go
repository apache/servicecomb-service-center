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
	mgr "github.com/ServiceComb/service-center/server/plugin"
	"net/http"
)

var authFunc func(r *http.Request) error

func init() {
	f := findAuthFunc("Identify")
	if f == nil {
		return
	}

	authFunc = f
	mgr.RegisterPlugin(mgr.Plugin{mgr.DYNAMIC, mgr.AUTH, "dynamic", New})
}

func findAuthFunc(funcName string) func(r *http.Request) error {
	ff, err := plugin.FindFunc("auth", funcName)
	if err != nil {
		return nil
	}
	f, ok := ff.(func(*http.Request) error)
	if !ok {
		util.Logger().Warnf(nil, "unexpected function '%s' format found in plugin 'auth'.", funcName)
		return nil
	}
	return f
}

func New() mgr.PluginInstance {
	return &DynamicAuth{}
}

type DynamicAuth struct {
}

func (da *DynamicAuth) Identify(r *http.Request) error {
	return authFunc(r)
}
