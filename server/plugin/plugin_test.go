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
package plugin

import (
	"net/http"
	"testing"
)

type mockAuthPlugin struct {
	Times int
}

func (*mockAuthPlugin) Identify(r *http.Request) error {
	return nil
}

func TestPluginManager_New(t *testing.T) {
	pm := &Manager{}
	pm.Initialize()

	p := pm.Get(AUTH, "buildin")
	if p != nil {
		t.Fatalf("TestPluginManager_New failed")
	}

	times := 0
	fn := func() Instance {
		times++
		AUTH.ActiveConfigs().Set("a", "a")
		return &mockAuthPlugin{times}
	}
	pm.Register(Plugin{AUTH, "buildin", fn})

	i := pm.Instance(AUTH)
	if i != pm.Instance(AUTH) || AUTH.ActiveConfigs().String("a", "") != "a" {
		t.Fatalf("TestPluginManager_New failed")
	}

	pm.ReloadAll()
	if AUTH.ActiveConfigs().String("a", "") != "" {
		t.Fatalf("TestPluginManager_New failed")
	}

	n := pm.Instance(AUTH)
	if i == n {
		t.Fatalf("TestPluginManager_New failed")
	}

	defer func() {
		if err := recover(); err != nil {
			t.Fatalf("TestPluginManager_New failed")
		}
	}()
	RegisterPlugin(Plugin{Kind(999), "999", nil})
	DynamicPluginFunc(Kind(999), "999")

	LoadPlugins()
}
