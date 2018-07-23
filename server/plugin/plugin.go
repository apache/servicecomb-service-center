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
	"github.com/apache/incubator-servicecomb-service-center/pkg/plugin"
	"github.com/apache/incubator-servicecomb-service-center/pkg/util"
	"github.com/apache/incubator-servicecomb-service-center/server/infra/auditlog"
	"github.com/apache/incubator-servicecomb-service-center/server/infra/auth"
	"github.com/apache/incubator-servicecomb-service-center/server/infra/quota"
	"github.com/apache/incubator-servicecomb-service-center/server/infra/registry"
	"github.com/apache/incubator-servicecomb-service-center/server/infra/security"
	"github.com/apache/incubator-servicecomb-service-center/server/infra/tls"
	"github.com/apache/incubator-servicecomb-service-center/server/infra/tracing"
	"github.com/apache/incubator-servicecomb-service-center/server/infra/uuid"
	"github.com/astaxie/beego"
	pg "plugin"
	"strconv"
	"sync"
)

const BUILDIN = "buildin"

const (
	UUID PluginName = iota
	AUDIT_LOG
	AUTH
	CIPHER
	QUOTA
	REGISTRY
	TRACING
	TLS
	typeEnd
)

var pluginNames = map[PluginName]string{
	UUID:      "uuid",
	AUDIT_LOG: "auditlog",
	AUTH:      "auth",
	CIPHER:    "cipher",
	QUOTA:     "quota",
	REGISTRY:  "registry",
	TRACING:   "trace",
	TLS:       "ssl",
}

var pluginMgr = &PluginManager{}

func init() {
	pluginMgr.Initialize()
}

type PluginName int

func (pn PluginName) String() string {
	if name, ok := pluginNames[pn]; ok {
		return name
	}
	return "PLUGIN" + strconv.Itoa(int(pn))
}

type Plugin struct {
	// plugin class name
	PName PluginName
	// plugin name
	Name string
	New  func() PluginInstance
}

type PluginInstance interface{}

type wrapInstance struct {
	dynamic  bool
	instance PluginInstance
	lock     sync.RWMutex
}

type PluginManager struct {
	plugins   map[PluginName]map[string]*Plugin
	instances map[PluginName]*wrapInstance
}

func (pm *PluginManager) Initialize() {
	pm.plugins = make(map[PluginName]map[string]*Plugin, int(typeEnd))
	pm.instances = make(map[PluginName]*wrapInstance, int(typeEnd))

	for t := PluginName(0); t != typeEnd; t++ {
		pm.instances[t] = &wrapInstance{}
	}
}

func (pm *PluginManager) ReloadAll() {
	for pn := range pm.instances {
		pm.Reload(pn)
	}
}

// unsafe
func (pm *PluginManager) Register(p Plugin) {
	m, ok := pm.plugins[p.PName]
	if !ok {
		m = make(map[string]*Plugin, 5)
	}
	m[p.Name] = &p
	pm.plugins[p.PName] = m
	util.Logger().Infof("load '%s' plugin named '%s'", p.PName, p.Name)
}

func (pm *PluginManager) Get(pn PluginName, name string) *Plugin {
	m, ok := pm.plugins[pn]
	if !ok {
		return nil
	}
	return m[name]
}

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

func (pm *PluginManager) New(pn PluginName) {
	var (
		title = "static"
		f     func() PluginInstance
	)

	wi := pm.instances[pn]
	p := pm.existDynamicPlugin(pn)
	if p != nil {
		wi.dynamic = true
		title = "dynamic"
		f = p.New
	} else {
		wi.dynamic = false
		m, ok := pm.plugins[pn]
		if !ok {
			return
		}

		name := beego.AppConfig.DefaultString(pn.String()+"_plugin", BUILDIN)
		p, ok = m[name]
		if !ok {
			return
		}

		f = p.New
	}
	util.Logger().Infof("call %s '%s' plugin %s(), new a '%s' instance",
		title, p.PName, util.FuncName(f), p.Name)

	wi.instance = f()
}

func (pm *PluginManager) Reload(pn PluginName) {
	wi := pm.instances[pn]
	wi.lock.Lock()
	wi.instance = nil
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

func (pm *PluginManager) Registry() registry.Registry {
	return pm.Instance(REGISTRY).(registry.Registry)
}
func (pm *PluginManager) UUID() uuid.UUID { return pm.Instance(UUID).(uuid.UUID) }
func (pm *PluginManager) AuditLog() auditlog.AuditLogger {
	return pm.Instance(AUDIT_LOG).(auditlog.AuditLogger)
}
func (pm *PluginManager) Auth() auth.Auth              { return pm.Instance(AUTH).(auth.Auth) }
func (pm *PluginManager) Cipher() security.Cipher      { return pm.Instance(CIPHER).(security.Cipher) }
func (pm *PluginManager) Quota() quota.QuotaManager    { return pm.Instance(QUOTA).(quota.QuotaManager) }
func (pm *PluginManager) Tracing() (v tracing.Tracing) { return pm.Instance(TRACING).(tracing.Tracing) }
func (pm *PluginManager) TLS() tls.TLS                 { return pm.Instance(TLS).(tls.TLS) }

func Plugins() *PluginManager {
	return pluginMgr
}

func RegisterPlugin(p Plugin) {
	Plugins().Register(p)
}

// DynamicPluginFunc should be called in buildin implement
func DynamicPluginFunc(pn PluginName, funcName string) pg.Symbol {
	if wi, ok := Plugins().instances[pn]; ok && !wi.dynamic {
		return nil
	}

	f, err := plugin.FindFunc(pn.String(), funcName)
	if err != nil {
		util.Logger().Errorf(err, "plugin '%s': not implemented function '%s'.", pn, funcName)
	}
	return f
}
