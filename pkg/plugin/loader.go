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
	"github.com/ServiceComb/service-center/server/core"
	"io/ioutil"
	"os"
	"path/filepath"
	"plugin"
	"regexp"
	"sync"
)

var (
	pluginManager PluginManager
	once          sync.Once
)

type PluginManager struct {
	Plugins map[string]*plugin.Plugin
}

func (pm *PluginManager) Init() {
	pm.Plugins = make(map[string]*plugin.Plugin)

	err := pm.ReloadPlugins()

	if len(pm.Plugins) == 0 {
		util.Logger().Warnf(err, "no any plugin has been loaded.")
	}
}

func (pm *PluginManager) ReloadPlugins() error {
	dir := os.ExpandEnv(core.ServerInfo.Config.PluginsDir)
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

	nameRegex := `([A-Za-z0-9_.-]+)_plugin.so$`
	regex, err := regexp.Compile(nameRegex)
	if err != nil {
		return err
	}

	for _, file := range files {
		if !file.Mode().IsRegular() {
			continue
		}
		submatchs := regex.FindStringSubmatch(file.Name())
		if len(submatchs) >= 2 {
			// golang 1.8+ feature
			pluginFileFullPath := filepath.Join(dir, file.Name())
			p, err := plugin.Open(pluginFileFullPath)
			util.Logger().Debugf("load plugin '%s'. path: %s, result: %s",
				submatchs[1], pluginFileFullPath, err)
			if err != nil {
				return fmt.Errorf("load plugin '%s' error for %s", submatchs[1], err.Error())
			}
			util.Logger().Infof("load plugin '%s' successfully.", submatchs[1])
			pm.Plugins[submatchs[1]] = p
		}
	}
	return nil
}

func (pm *PluginManager) Find(pluginName, funcName string) (plugin.Symbol, error) {
	p, ok := pm.Plugins[pluginName]
	if !ok {
		return nil, fmt.Errorf("can not find plugin '%s'.", pluginName)
	}
	f, err := p.Lookup(funcName)
	if err != nil {
		return nil, err
	}
	return f, nil
}

func GetPluginManager() *PluginManager {
	once.Do(pluginManager.Init)
	return &pluginManager
}

func Reload() error {
	return GetPluginManager().ReloadPlugins()
}

func FindFunc(pluginName, funcName string) (plugin.Symbol, error) {
	return GetPluginManager().Find(pluginName, funcName)
}
