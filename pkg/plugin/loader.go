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
	"io/ioutil"
	"os"
	"path/filepath"
	"plugin"
	"regexp"
	"sync"

	"github.com/apache/servicecomb-service-center/pkg/log"
)

var (
	loader   Loader
	once     sync.Once
	regex, _ = regexp.Compile(`([A-Za-z0-9_.-]+)_plugin.so$`)
)

type wrapPlugin struct {
	p     *plugin.Plugin
	funcs map[string]plugin.Symbol
}

type Loader struct {
	Plugins map[string]*wrapPlugin
	mux     sync.RWMutex
}

func (pm *Loader) Init() {
	pm.Plugins = make(map[string]*wrapPlugin, 10)

	err := pm.ReloadPlugins()
	if len(pm.Plugins) == 0 {
		log.Errorf(err, "no any plugin has been loaded")
	}
}

func (pm *Loader) ReloadPlugins() error {
	dir := os.ExpandEnv(GetConfigurator().GetPluginDir())
	if len(dir) == 0 {
		dir, _ = os.Getwd()
	}
	if len(dir) == 0 {
		return fmt.Errorf("'plugins_dir' is unset")
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
		log.Infof("load plugin '%s' successfully", submatchs[1])

		pm.mux.Lock()
		pm.Plugins[submatchs[1]] = &wrapPlugin{p, make(map[string]plugin.Symbol, 10)}
		pm.mux.Unlock()
	}
	return nil
}

func (pm *Loader) Find(pluginName, funcName string) (plugin.Symbol, error) {
	pm.mux.RLock()
	w, ok := pm.Plugins[pluginName]
	if !ok {
		pm.mux.RUnlock()
		return nil, fmt.Errorf("can not find plugin '%s'", pluginName)
	}

	f, ok := w.funcs[funcName]
	if ok {
		pm.mux.RUnlock()
		return f, nil
	}
	pm.mux.RUnlock()

	pm.mux.Lock()
	var err error
	f, err = w.p.Lookup(funcName)
	if err != nil {
		pm.mux.Unlock()
		return nil, err
	}
	w.funcs[funcName] = f
	pm.mux.Unlock()
	return f, nil
}

func (pm *Loader) Exist(pluginName string) bool {
	pm.mux.RLock()
	_, ok := pm.Plugins[pluginName]
	pm.mux.RUnlock()
	return ok
}

func GetLoader() *Loader {
	once.Do(loader.Init)
	return &loader
}

func Reload() error {
	return GetLoader().ReloadPlugins()
}

func FindFunc(pluginName, funcName string) (plugin.Symbol, error) {
	return GetLoader().Find(pluginName, funcName)
}
