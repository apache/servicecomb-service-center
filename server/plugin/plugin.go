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

const (
	defaultPluginSize     = 20
	defaultPluginImplSize = 5
)

var pluginMgr = &Manager{}

func init() {
	pluginMgr.Initialize()
}

type wrapInstance struct {
	dynamic  bool
	instance Instance
	lock     sync.RWMutex
}

// Manager manages plugin instance generation.
// Manager keeps the plugin instance currently used by server
// for every plugin interface.
type Manager struct {
	plugins   map[Kind]map[ImplName]*Plugin
	instances map[Kind]*wrapInstance
}

// Initialize initializes the struct
func (pm *Manager) Initialize() {
	pm.plugins = make(map[Kind]map[ImplName]*Plugin, defaultPluginSize)
	pm.instances = make(map[Kind]*wrapInstance, defaultPluginSize)
}

// ReloadAll reloads all the plugin instances
func (pm *Manager) ReloadAll() {
	for pn := range pm.instances {
		pm.Reload(pn)
	}
}

// Register registers a 'Plugin'
// unsafe
func (pm *Manager) Register(p Plugin) {
	t := p.Kind
	m, ok := pm.plugins[t]
	if !ok {
		m = make(map[ImplName]*Plugin, defaultPluginImplSize)
	}
	m[p.Name] = &p
	pm.plugins[t] = m
	pm.instances[t] = &wrapInstance{}
	log.Infof("load '%s' plugin named '%s'", t, p.Name)
}

// Get gets a 'Plugin'
func (pm *Manager) Get(pn Kind, name ImplName) *Plugin {
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
// The go plugin file should be {plugins_dir}/{Kind}_plugin.so.
// ('plugins_dir' must be configured as a valid path in service-center config.)
// The plugin config in service-center config should be:
// {Kind}_plugin = {ImplName}
//
// e.g. For registry plugin, you can set a config in app.conf:
// plugins_dir = /home, and supply a go plugin file: /home/registry_plugin.so;
// or if you want to use etcd as registry, you can set a config in app.conf:
// registry_plugin = etcd.
func (pm *Manager) Instance(pn Kind) Instance {
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
func (pm *Manager) New(pn Kind) {
	var (
		title = Static
		f     func() Instance
	)

	wi := pm.instances[pn]
	p := pm.existDynamicPlugin(pn)
	if p != nil {
		// Dynamic plugin has high priority.
		wi.dynamic = true
		title = Dynamic
		f = p.New
	} else {
		wi.dynamic = false
		m, ok := pm.plugins[pn]
		if !ok {
			return
		}

		name := beego.AppConfig.DefaultString(pn.String()+"_plugin", Buildin)
		p, ok = m[ImplName(name)]
		if !ok {
			return
		}

		f = p.New
	}
	log.Infof("call %s '%s' plugin %s(), new a '%s' instance",
		title, p.Kind, util.FuncName(f), p.Name)

	wi.instance = f()
}

// Reload reloads the instance of the specified plugin interface.
func (pm *Manager) Reload(pn Kind) {
	wi := pm.instances[pn]
	wi.lock.Lock()
	wi.instance = nil
	wi.lock.Unlock()
}

func (pm *Manager) existDynamicPlugin(pn Kind) *Plugin {
	m, ok := pm.plugins[pn]
	if !ok {
		return nil
	}
	// 'buildin' implement of all plugins should call DynamicPluginFunc()
	if plugin.GetLoader().Exist(pn.String()) {
		return m[Buildin]
	}
	return nil
}

func (pm *Manager) IsDynamicPlugin(pn Kind) bool {
	wi, ok := Plugins().instances[pn]
	return ok && wi.dynamic
}

// Plugins returns the 'Manager'.
func Plugins() *Manager {
	return pluginMgr
}

// RegisterPlugin registers a 'Plugin'.
func RegisterPlugin(p Plugin) {
	pluginMgr.Register(p)
}

// LoadPlugins loads and sets all the plugin interfaces's instance.
func LoadPlugins() {
	for p := range pluginMgr.plugins {
		pluginMgr.Instance(p)
	}
}
