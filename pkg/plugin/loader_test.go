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
	pg "plugin"
	"testing"
)

func TestLoader_Init(t *testing.T) {
	defer func() {
		if err := recover(); err != nil {
			t.Fatalf(`TestLoader_Init failed, %v`, err)
		}
	}()
	loader := Loader{}
	loader.Init()
}

type testLoaderConfigurator struct {
}

func (c *testLoaderConfigurator) GetImplName(_ Kind) string {
	return "test"
}
func (c *testLoaderConfigurator) GetPluginDir() string {
	return "dir"
}

func TestLoader_ReloadPlugins(t *testing.T) {
	loader := Loader{}
	loader.Init()
	err := loader.ReloadPlugins()
	if err != nil {
		t.Fatalf(`TestLoader_ReloadPlugins failed, %s`, err.Error())
	}

	old := GetConfigurator()
	RegisterConfigurator(&testLoaderConfigurator{})

	err = loader.ReloadPlugins()
	if err == nil {
		t.Fatalf(`TestLoader_ReloadPlugins failed`)
	}

	RegisterConfigurator(old)
}

func TestLoader_Exist(t *testing.T) {
	loader := Loader{}
	loader.Init()
	b := loader.Exist("")
	if b {
		t.Fatalf(`TestLoader_Exist failed`)
	}
}

func TestLoader_Find(t *testing.T) {
	loader := Loader{}
	loader.Init()
	f, err := loader.Find("", "")
	if err == nil || f != nil {
		t.Fatalf(`TestLoader_Find failed`)
	}

	loader.Plugins["a"] = &wrapPlugin{&pg.Plugin{}, make(map[string]pg.Symbol)}
	f, err = loader.Find("a", "")
	if err == nil || f != nil {
		t.Fatalf(`TestLoader_Find failed`)
	}
}

func TestPluginLoader(t *testing.T) {
	loader := GetLoader()
	if loader == nil {
		t.Fatalf(`TestPluginLoader failed`)
	}

	err := Reload()
	if err != nil {
		t.Fatalf(`TestPluginLoader Reload failed, %s`, err)
	}

	f, err := FindFunc("", "")
	if err == nil || f != nil {
		t.Fatalf(`TestPluginLoader FindFunc failed`)
	}
}
