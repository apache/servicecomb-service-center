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
	"github.com/apache/incubator-servicecomb-service-center/pkg/util"
	"io/ioutil"
	"os"
	"path/filepath"
	"plugin"
	"regexp"
	"sync"
)

var (
	loader   Loader
	once     sync.Once
	regex, _ = regexp.Compile(`([A-Za-z0-9_.-]+)_plugin.so$`)
)

type Loader struct {
	Dir     string
	Plugins map[string]*plugin.Plugin
	mux     sync.RWMutex
}

func (pm *Loader) Init() {
	pm.Plugins = make(map[string]*plugin.Plugin)

	err := pm.ReloadPlugins()
	if len(pm.Plugins) == 0 {
		util.Logger().Warnf(err, "no any plugin has been loaded.")
	}
}

func (pm *Loader) ReloadPlugins() error {
	dir := os.ExpandEnv(pm.Dir)
	if len(dir) == 0 {
		dir, _ = os.Getwd()
	}
	if len(dir) == 0 {
		return fmt.Errorf("'plugins_dir' is unset.")
	}

	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return err
	}

	for _, file := range files {
		if !file.Mode().IsRegular() {
			continue
		}

		submatchs := regex.FindStringSubmatch(file.Name())
		if len(submatchs) < 2 {
			continue
		}

		// golang 1.8+ feature
		pluginFileFullPath := filepath.Join(dir, file.Name())
		p, err := plugin.Open(pluginFileFullPath)
		if err != nil {
			return fmt.Errorf("load plugin '%s' error for %s", submatchs[1], err.Error())
		}
		util.Logger().Infof("load plugin '%s' successfully.", submatchs[1])

		pm.mux.Lock()
		pm.Plugins[submatchs[1]] = p
		pm.mux.Unlock()
	}
	return nil
}

func (pm *Loader) Find(pluginName, funcName string) (plugin.Symbol, error) {
	pm.mux.RLock()
	p, ok := pm.Plugins[pluginName]
	pm.mux.RUnlock()
	if !ok {
		return nil, fmt.Errorf("can not find plugin '%s'.", pluginName)
	}
	f, err := p.Lookup(funcName)
	if err != nil {
		return nil, err
	}
	return f, nil
}

func (pm *Loader) Exist(pluginName string) bool {
	pm.mux.RLock()
	_, ok := pm.Plugins[pluginName]
	pm.mux.RUnlock()
	return ok
}

func SetPluginDir(dir string) {
	loader.Dir = dir
}

func PluginLoader() *Loader {
	once.Do(loader.Init)
	return &loader
}

func Reload() error {
	return PluginLoader().ReloadPlugins()
}

func FindFunc(pluginName, funcName string) (plugin.Symbol, error) {
	return PluginLoader().Find(pluginName, funcName)
}
