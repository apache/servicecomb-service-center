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
	"fmt"
	pg "plugin"
	"testing"
)

func TestLoader_Init(t *testing.T) {
	defer func() {
		if err := recover(); err != nil {
			fmt.Printf(`TestLoader_Init failed, %v`, err)
			t.FailNow()
		}
	}()
	loader := Loader{}
	loader.Init()
}

func TestLoader_ReloadPlugins(t *testing.T) {
	loader := Loader{}
	loader.Init()
	err := loader.ReloadPlugins()
	if err != nil {
		fmt.Printf(`TestLoader_ReloadPlugins failed, %s`, err.Error())
		t.FailNow()
	}

	loader.Dir = "xxx"
	err = loader.ReloadPlugins()
	if err == nil {
		fmt.Printf(`TestLoader_ReloadPlugins failed`)
		t.FailNow()
	}
}

func TestLoader_Exist(t *testing.T) {
	loader := Loader{}
	loader.Init()
	b := loader.Exist("")
	if b {
		fmt.Printf(`TestLoader_Exist failed`)
		t.FailNow()
	}
}

func TestLoader_Find(t *testing.T) {
	loader := Loader{}
	loader.Init()
	f, err := loader.Find("", "")
	if err == nil || f != nil {
		fmt.Printf(`TestLoader_Find failed`)
		t.FailNow()
	}

	loader.Plugins["a"] = &wrapPlugin{&pg.Plugin{}, make(map[string]pg.Symbol)}
	f, err = loader.Find("a", "")
	if err == nil || f != nil {
		fmt.Printf(`TestLoader_Find failed`)
		t.FailNow()
	}
}

func TestSetPluginDir(t *testing.T) {
	SetPluginDir("")
}

func TestPluginLoader(t *testing.T) {
	loader := PluginLoader()
	if loader == nil {
		fmt.Printf(`TestPluginLoader failed`)
		t.FailNow()
	}

	err := Reload()
	if err != nil {
		fmt.Printf(`TestPluginLoader Reload failed, %s`, err)
		t.FailNow()
	}

	f, err := FindFunc("", "")
	if err == nil || f != nil {
		fmt.Printf(`TestPluginLoader FindFunc failed`)
		t.FailNow()
	}
}
