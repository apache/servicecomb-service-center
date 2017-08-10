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
package upgrade

import (
	"fmt"
	"github.com/ServiceComb/service-center/server/core"
	"github.com/ServiceComb/service-center/server/core/mux"
	"github.com/ServiceComb/service-center/server/service/microservice"
	"github.com/ServiceComb/service-center/util"
	"github.com/ServiceComb/service-center/version"
)

var patches = make(map[string]func() error)

func needUpgrade() bool {
	if core.GetSystemConfig() == nil {
		err := core.LoadSystemConfig()
		if err != nil {
			util.LOGGER.Errorf(err, "check version failed, can not load the system config")
			return false
		}
	}
	return !microservice.VersionMatchRule(core.GetSystemConfig().Version,
		fmt.Sprintf("%s+", version.Ver().Version))
}

func mapToList(dict map[string]func() error) []string {
	ret := make([]string, 0, len(dict))
	for k := range dict {
		ret = append(ret, k)
	}
	return ret
}

func Upgrade() error {
	if !needUpgrade() {
		return nil
	}

	lock, err := mux.Lock(mux.GLOBAL_LOCK)
	if err != nil {
		return err
	}
	defer lock.Unlock()

	versions := mapToList(patches)
	curVer := core.GetSystemConfig().Version
	for _, less := range versions {
		if !microservice.VersionMatchRule(curVer, less) {
			continue
		}
		err := patches[less]()
		if err != nil {
			return fmt.Errorf("upgrade %s patch failed for %s", less, err.Error())
		}
	}

	return upgradeSystemConfig()
}

func AddPatch(less string, f func() error) {
	patches[less] = f
}

func upgradeSystemConfig() error {
	cfg := core.GetSystemConfig()
	cfg.Version = version.Ver().Version
	return core.UpgradeSystemConfig()
}
