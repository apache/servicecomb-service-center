//Copyright 2017 Huawei Technologies Co., Ltd
//
//Licensed under the Apache License, Version 2.0 (the "License");
//you may not use this file except in compliance with the License.
//You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
//Unless required by applicable law or agreed to in writing, software
//distributed under the License is distributed on an "AS IS" BASIS,
//WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//See the License for the specific language governing permissions and
//limitations under the License.
package plugin

import (
	"fmt"
	"github.com/ServiceComb/service-center/pkg/util"
	"github.com/ServiceComb/service-center/server/infra/auditlog"
	"github.com/ServiceComb/service-center/server/infra/auth"
	"github.com/ServiceComb/service-center/server/infra/quota"
	"github.com/ServiceComb/service-center/server/infra/registry"
	"github.com/ServiceComb/service-center/server/infra/security"
	"github.com/ServiceComb/service-center/server/infra/uuid"
	"github.com/astaxie/beego"
	"sync"
)

const (
	STATIC PluginType = iota
	DYNAMIC
)

const (
	UUID PluginName = iota
	AUDIT_LOG
	AUTH
	CIPHER
	QUOTA
	REGISTRY
	typeEnd
)

var pluginNames = map[PluginName]string{
	UUID:      "uuid",
	AUDIT_LOG: "auditlog",
	AUTH:      "auth",
	CIPHER:    "cipher",
	QUOTA:     "quota",
	REGISTRY:  "registry",
}

var pluginMgr = &PluginManager{}

func init() {
	pluginMgr.Initialize()
}

type PluginType int

func (pt PluginType) String() string {
	if pt == DYNAMIC {
		return "dynamic"
	}
	return "static"
}

type PluginName int

func (pn PluginName) String() string {
	if name, ok := pluginNames[pn]; ok {
		return name
	}
	return "PLUGIN" + fmt.Sprint(pn)
}

type Plugin struct {
	Type  PluginType
	PName PluginName
	Name  string
	New   func() PluginInstance
}

type PluginInstance interface{}

type wrapInstance struct {
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
		pluginMgr.instances[t] = &wrapInstance{}
	}
}

func (pm *PluginManager) ReloadAll() {
	for _, i := range pm.instances {
		i.lock.Lock()
		i.instance = nil
		i.lock.Unlock()
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
	util.Logger().Infof("%s load '%s' plugin named '%s'", p.Type, p.PName, p.Name)
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
	wi.instance = pm.New(pn)
	wi.lock.Unlock()

	return wi.instance
}

func (pm *PluginManager) New(pn PluginName) PluginInstance {
	var f func() PluginInstance
	p := pm.existDynamicPlugin(pn)
	if p != nil {
		f = p.New
	} else {
		m, ok := pm.plugins[pn]
		if !ok {
			return nil
		}

		name := beego.AppConfig.DefaultString(pn.String()+"_plugin", "buildin")
		p, ok = m[name]
		if !ok {
			return nil
		}

		f = p.New
	}
	util.Logger().Infof("new '%s' plugin '%s' instance", p.PName, p.Name)
	return f()
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
	for _, p := range m {
		if p.Type == DYNAMIC {
			return p
		}
	}
	return nil
}

func (pm *PluginManager) Registry() registry.Registry {
	return pm.Instance(REGISTRY).(registry.Registry)
}

func (pm *PluginManager) UUID() uuid.UUID {
	return pm.Instance(UUID).(uuid.UUID)
}

func (pm *PluginManager) AuditLog() auditlog.AuditLogger {
	return pm.Instance(AUDIT_LOG).(auditlog.AuditLogger)
}

func (pm *PluginManager) Auth() auth.Auth {
	return pm.Instance(AUTH).(auth.Auth)
}

func (pm *PluginManager) Cipher() security.Cipher {
	return pm.Instance(CIPHER).(security.Cipher)
}

func (pm *PluginManager) Quota() quota.QuotaManager {
	return pm.Instance(QUOTA).(quota.QuotaManager)
}

func Plugins() *PluginManager {
	return pluginMgr
}

func RegisterPlugin(p Plugin) {
	Plugins().Register(p)
}
