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
	"sync"

	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/plugin"
	"github.com/apache/servicecomb-service-center/pkg/util"

	"github.com/astaxie/beego"
)

var pluginMgr = &PluginManager{}

func init() {
	pluginMgr.Initialize()
}

type wrapInstance struct {
	dynamic  bool
	instance PluginInstance
	lock     sync.RWMutex
}

// PluginManager manages plugin instance generation.
// PluginManager keeps the plugin instance currently used by server
// for every plugin interface.
type PluginManager struct {
	plugins   map[PluginName]map[PluginImplName]*Plugin
	instances map[PluginName]*wrapInstance
}

// Initialize initializes the struct
func (pm *PluginManager) Initialize() {
	pm.plugins = make(map[PluginName]map[PluginImplName]*Plugin, int(typeEnd))
	pm.instances = make(map[PluginName]*wrapInstance, int(typeEnd))

	for t := PluginName(0); t != typeEnd; t++ {
		pm.instances[t] = &wrapInstance{}
	}
}

// ReloadAll reloads all the plugin instances
func (pm *PluginManager) ReloadAll() {
	for pn := range pm.instances {
		pm.Reload(pn)
	}
}

// Register registers a 'Plugin'
// unsafe
func (pm *PluginManager) Register(p Plugin) {
	m, ok := pm.plugins[p.PName]
	if !ok {
		m = make(map[PluginImplName]*Plugin, 5)
	}
	m[p.Name] = &p
	pm.plugins[p.PName] = m
	log.Infof("load '%s' plugin named '%s'", p.PName, p.Name)
}

// Get gets a 'Plugin'
func (pm *PluginManager) Get(pn PluginName, name PluginImplName) *Plugin {
	m, ok := pm.plugins[pn]
	if !ok {
		return nil
	}
	return m[name]
}

// Instance gets an plugin instance.
// What plugin instance you get is depended on the supplied go plugin files
// (high priority) or the plugin config(low priority)
//
// The go plugin file should be {plugins_dir}/{PluginName}_plugin.so.
// ('plugins_dir' must be configured as a valid path in service-center config.)
// The plugin config in service-center config should be:
// {PluginName}_plugin = {PluginImplName}
//
// e.g. For registry plugin, you can set a config in app.conf:
// plugins_dir = /home, and supply a go plugin file: /home/registry_plugin.so;
// or if you want to use etcd as registry, you can set a config in app.conf:
// registry_plugin = etcd.
func (pm *PluginManager) Instance(pn PluginName) PluginInstance {
	wi := pm.instances[pn]
	wi.lock.RLock()
	if wi.instance != nil {
		wi.lock.RUnlock()
		return wi.instance
	}
	wi.lock.RUnlock()

	wi.lock.Lock()
	if wi.instance != nil {
		wi.lock.Unlock()
		return wi.instance
	}
	pm.New(pn)
	wi.lock.Unlock()

	return wi.instance
}

// New initializes and sets the instance of a plugin interface,
// but not returns it.
// Use 'Instance' if you want to get the plugin instance.
// We suggest you to use 'Instance' instead of 'New'.
func (pm *PluginManager) New(pn PluginName) {
	var (
		title = STATIC
		f     func() PluginInstance
	)

	wi := pm.instances[pn]
	p := pm.existDynamicPlugin(pn)
	if p != nil {
		// Dynamic plugin has high priority.
		wi.dynamic = true
		title = DYNAMIC
		f = p.New
	} else {
		wi.dynamic = false
		m, ok := pm.plugins[pn]
		if !ok {
			return
		}

		name := beego.AppConfig.DefaultString(pn.String()+"_plugin", BUILDIN)
		p, ok = m[PluginImplName(name)]
		if !ok {
			return
		}

		f = p.New
		pn.ActiveConfigs().Set(keyPluginName, name)
	}
	log.Infof("call %s '%s' plugin %s(), new a '%s' instance",
		title, p.PName, util.FuncName(f), p.Name)

	wi.instance = f()
}

// Reload reloads the instance of the specified plugin interface.
func (pm *PluginManager) Reload(pn PluginName) {
	wi := pm.instances[pn]
	wi.lock.Lock()
	wi.instance = nil
	pn.ClearConfigs()
	wi.lock.Unlock()
}

func (pm *PluginManager) existDynamicPlugin(pn PluginName) *Plugin {
	m, ok := pm.plugins[pn]
	if !ok {
		return nil
	}
	// 'buildin' implement of all plugins should call DynamicPluginFunc()
	if plugin.PluginLoader().Exist(pn.String()) {
		return m[BUILDIN]
	}
	return nil
}

// Plugins returns the 'PluginManager'.
func Plugins() *PluginManager {
	return pluginMgr
}

// RegisterPlugin registers a 'Plugin'.
func RegisterPlugin(p Plugin) {
	Plugins().Register(p)
}

// LoadPlugins loads and sets all the plugin interfaces's instance.
func LoadPlugins() {
	for t := PluginName(0); t != typeEnd; t++ {
		Plugins().Instance(t)
	}
}
